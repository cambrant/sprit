package sprit

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// newTestAnimation creates an Animation with n distinct frames for testing.
// Each frame is a 1x1 ebiten image (different objects so we can compare pointers).
func newTestAnimation(n int, mode AnimationMode, speedMs int) *Animation {
	frames := make([]*ebiten.Image, n)
	for i := range frames {
		frames[i] = ebiten.NewImage(1, 1)
	}
	return &Animation{
		Name:      "test",
		Frames:    frames,
		Mode:      mode,
		Speed:     time.Duration(speedMs) * time.Millisecond,
		current:   0,
		elapsed:   0,
		direction: 1,
		finished:  false,
	}
}

func TestAnimation_Loop(t *testing.T) {
	anim := newTestAnimation(3, AnimLoop, 100)
	step := 100 * time.Millisecond

	// Frame 0 initially.
	if anim.Frame() != anim.Frames[0] {
		t.Errorf("initial frame: got frame %d, want 0", anim.current)
	}

	// Advance through all frames and verify wrap-around.
	expected := []int{1, 2, 0, 1, 2, 0}
	for i, want := range expected {
		anim.Update(step)
		if anim.current != want {
			t.Errorf("step %d: got frame %d, want %d", i+1, anim.current, want)
		}
		if anim.IsFinished() {
			t.Errorf("step %d: loop animation should never be finished", i+1)
		}
	}
}

func TestAnimation_Once(t *testing.T) {
	anim := newTestAnimation(3, AnimOnce, 100)
	step := 100 * time.Millisecond

	// Advance to frame 1.
	anim.Update(step)
	if anim.current != 1 {
		t.Errorf("after step 1: got frame %d, want 1", anim.current)
	}
	if anim.IsFinished() {
		t.Error("should not be finished after frame 1")
	}

	// Advance to frame 2 (last) — should finish.
	anim.Update(step)
	if anim.current != 2 {
		t.Errorf("after step 2: got frame %d, want 2", anim.current)
	}
	if !anim.IsFinished() {
		t.Error("should be finished after reaching last frame")
	}

	// Further updates should keep it on the last frame.
	anim.Update(step)
	anim.Update(step)
	if anim.current != 2 {
		t.Errorf("after extra steps: got frame %d, want 2 (stuck on last)", anim.current)
	}
	if !anim.IsFinished() {
		t.Error("should remain finished")
	}
}

func TestAnimation_PingPong(t *testing.T) {
	anim := newTestAnimation(4, AnimPingPong, 100)
	step := 100 * time.Millisecond

	// Forward: 0 → 1 → 2 → 3 → reverse → 2 → 1 → 0 → reverse → 1 → 2 ...
	expected := []int{1, 2, 3, 2, 1, 0, 1, 2, 3, 2}
	for i, want := range expected {
		anim.Update(step)
		if anim.current != want {
			t.Errorf("step %d: got frame %d, want %d", i+1, anim.current, want)
		}
		if anim.IsFinished() {
			t.Errorf("step %d: pingpong animation should never be finished", i+1)
		}
	}
}

func TestAnimation_PingPong_TwoFrames(t *testing.T) {
	anim := newTestAnimation(2, AnimPingPong, 100)
	step := 100 * time.Millisecond

	// 0 → 1 → 0 → 1 → 0
	expected := []int{1, 0, 1, 0, 1}
	for i, want := range expected {
		anim.Update(step)
		if anim.current != want {
			t.Errorf("step %d: got frame %d, want %d", i+1, anim.current, want)
		}
	}
}

func TestAnimation_Reset(t *testing.T) {
	anim := newTestAnimation(3, AnimOnce, 100)
	step := 100 * time.Millisecond

	// Advance to the end.
	anim.Update(step)
	anim.Update(step)
	if !anim.IsFinished() {
		t.Fatal("expected animation to be finished")
	}
	if anim.current != 2 {
		t.Fatalf("expected frame 2, got %d", anim.current)
	}

	// Reset and verify.
	anim.Reset()
	if anim.IsFinished() {
		t.Error("after Reset(), IsFinished() should be false")
	}
	if anim.current != 0 {
		t.Errorf("after Reset(), frame = %d, want 0", anim.current)
	}
	if anim.Frame() != anim.Frames[0] {
		t.Error("after Reset(), Frame() should return first frame")
	}

	// Should be able to play again.
	anim.Update(step)
	if anim.current != 1 {
		t.Errorf("after Reset+Update: got frame %d, want 1", anim.current)
	}
}

func TestAnimation_SingleFrame(t *testing.T) {
	anim := newTestAnimation(1, AnimLoop, 100)
	step := 100 * time.Millisecond

	// Single frame should stay on frame 0 regardless of updates.
	for i := 0; i < 5; i++ {
		anim.Update(step)
		if anim.current != 0 {
			t.Errorf("step %d: single-frame animation moved to frame %d", i+1, anim.current)
		}
	}
}

func TestAnimation_Frame_ReturnsCorrectImage(t *testing.T) {
	anim := newTestAnimation(3, AnimLoop, 100)
	step := 100 * time.Millisecond

	for i := 0; i < 3; i++ {
		got := anim.Frame()
		if got != anim.Frames[i] {
			t.Errorf("at frame %d: Frame() returned wrong image", i)
		}
		anim.Update(step)
	}
}

func TestAnimation_Draw_NoPanic(t *testing.T) {
	anim := newTestAnimation(2, AnimLoop, 100)
	screen := ebiten.NewImage(10, 10)
	anim.Draw(screen, 1, 2)
}

func TestAnimation_DrawWithOptions_NoPanic(t *testing.T) {
	anim := newTestAnimation(2, AnimLoop, 100)
	screen := ebiten.NewImage(10, 10)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(3, 4)
	anim.DrawWithOptions(screen, opts)
}

func TestAnimation_PartialTicks(t *testing.T) {
	anim := newTestAnimation(3, AnimLoop, 100)

	// 50ms is not enough to advance.
	anim.Update(50 * time.Millisecond)
	if anim.current != 0 {
		t.Errorf("after 50ms: got frame %d, want 0", anim.current)
	}

	// Another 50ms should now advance.
	anim.Update(50 * time.Millisecond)
	if anim.current != 1 {
		t.Errorf("after 100ms total: got frame %d, want 1", anim.current)
	}
}

func TestAnimation_LargeDelta(t *testing.T) {
	anim := newTestAnimation(3, AnimLoop, 100)

	// 250ms should advance 2 frames (100 + 100, 50ms leftover).
	anim.Update(250 * time.Millisecond)
	if anim.current != 2 {
		t.Errorf("after 250ms: got frame %d, want 2", anim.current)
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input string
		want  AnimationMode
		err   bool
	}{
		{"once", AnimOnce, false},
		{"loop", AnimLoop, false},
		{"pingpong", AnimPingPong, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"LOOP", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseMode(tt.input)
			if tt.err && err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
			if !tt.err && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.input, err)
			}
			if !tt.err && got != tt.want {
				t.Errorf("parseMode(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnimation_Clone(t *testing.T) {
	anim := newTestAnimation(3, AnimLoop, 100)
	step := 100 * time.Millisecond

	// Advance original.
	anim.Update(step)
	if anim.current != 1 {
		t.Fatalf("original frame = %d, want 1", anim.current)
	}

	// Clone should start at frame 0.
	clone := anim.clone()
	if clone.current != 0 {
		t.Errorf("clone frame = %d, want 0", clone.current)
	}
	if clone.Name != anim.Name {
		t.Errorf("clone name = %q, want %q", clone.Name, anim.Name)
	}
	if clone.Mode != anim.Mode {
		t.Errorf("clone mode = %d, want %d", clone.Mode, anim.Mode)
	}

	// Advancing clone should not affect original.
	clone.Update(step)
	clone.Update(step)
	if anim.current != 1 {
		t.Errorf("original frame changed to %d after clone update", anim.current)
	}
	if clone.current != 2 {
		t.Errorf("clone frame = %d, want 2", clone.current)
	}
}
