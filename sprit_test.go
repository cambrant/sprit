package sprit

import (
	"os"
	"testing"
)

func TestLoad_ValidSprites(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 sprites from sprites.hcl.
	names := atlas.Sprites()
	if len(names) != 3 {
		t.Fatalf("expected 3 sprites, got %d: %v", len(names), names)
	}

	// Verify known sprites are present.
	for _, name := range []string{"player_idle", "tree", "rock"} {
		s := atlas.Sprite(name)
		if s == nil {
			t.Errorf("Sprite(%q) = nil, want non-nil", name)
			continue
		}
		if s.Name != name {
			t.Errorf("Sprite(%q).Name = %q", name, s.Name)
		}
		w, h := s.Bounds()
		if w <= 0 || h <= 0 {
			t.Errorf("Sprite(%q).Bounds() = (%d, %d), want positive", name, w, h)
		}
		if s.Image == nil {
			t.Errorf("Sprite(%q).Image = nil", name)
		}
	}
}

func TestLoad_ValidSpriteWithRect(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// player_idle has rect = [0, 0, 8, 8] from an 8x8 image.
	s := atlas.Sprite("player_idle")
	if s == nil {
		t.Fatal("Sprite(\"player_idle\") = nil")
	}
	w, h := s.Bounds()
	if w != 8 || h != 8 {
		t.Errorf("player_idle.Bounds() = (%d, %d), want (8, 8)", w, h)
	}
}

func TestLoad_SpriteLookupMissing(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := atlas.Sprite("nonexistent")
	if s != nil {
		t.Errorf("Sprite(\"nonexistent\") = %v, want nil", s)
	}
}

func TestLoad_SpritesListSorted(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := atlas.Sprites()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Sprites() not sorted: %v", names)
			break
		}
	}
}

func TestLoad_MinimalSprite(t *testing.T) {
	fsys := os.DirFS("testdata/minimal")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := atlas.Sprites()
	if len(names) != 1 {
		t.Fatalf("expected 1 sprite, got %d: %v", len(names), names)
	}

	s := atlas.Sprite("pixel")
	if s == nil {
		t.Fatal("Sprite(\"pixel\") = nil")
	}
	w, h := s.Bounds()
	if w != 1 || h != 1 {
		t.Errorf("pixel.Bounds() = (%d, %d), want (1, 1)", w, h)
	}
}

func TestLoad_InvalidHCL(t *testing.T) {
	fsys := os.DirFS("testdata/invalid")
	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected error from invalid testdata, got nil")
	}
}

func TestLoad_EmptyFS(t *testing.T) {
	// An empty filesystem should produce an atlas with no sprites.
	fsys := os.DirFS(t.TempDir())
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(atlas.Sprites()) != 0 {
		t.Errorf("expected 0 sprites, got %d", len(atlas.Sprites()))
	}
}

func TestLoad_SpriteFullImage(t *testing.T) {
	// "rock" has no rect, so should use the full 8x8 image.
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := atlas.Sprite("rock")
	if s == nil {
		t.Fatal("Sprite(\"rock\") = nil")
	}
	w, h := s.Bounds()
	if w != 8 || h != 8 {
		t.Errorf("rock.Bounds() = (%d, %d), want (8, 8)", w, h)
	}
}

func TestLoad_ValidAnimations(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 animations from animations.hcl.
	names := atlas.Animations()
	if len(names) != 3 {
		t.Fatalf("expected 3 animations, got %d: %v", len(names), names)
	}

	// Verify known animations.
	for _, name := range []string{"player_walk", "player_die", "player_bounce"} {
		a := atlas.Animation(name)
		if a == nil {
			t.Errorf("Animation(%q) = nil, want non-nil", name)
			continue
		}
		if a.Name != name {
			t.Errorf("Animation(%q).Name = %q", name, a.Name)
		}
		if len(a.Frames) != 4 {
			t.Errorf("Animation(%q) has %d frames, want 4", name, len(a.Frames))
		}
	}
}

func TestLoad_AnimationLookupMissing(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := atlas.Animation("nonexistent")
	if a != nil {
		t.Errorf("Animation(\"nonexistent\") = %v, want nil", a)
	}
}

func TestLoad_AnimationIndependentInstances(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a1 := atlas.Animation("player_walk")
	a2 := atlas.Animation("player_walk")
	if a1 == nil || a2 == nil {
		t.Fatal("expected non-nil animations")
	}

	// They should be separate instances.
	if a1 == a2 {
		t.Error("two calls to Animation() returned the same pointer")
	}

	// Advancing one should not affect the other.
	a1.Update(a1.Speed)
	if a1.Frame() == a2.Frame() {
		t.Error("after advancing a1, a1.Frame() should differ from a2.Frame()")
	}
}

func TestLoad_AnimationModes(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name string
		mode AnimationMode
	}{
		{"player_walk", AnimLoop},
		{"player_die", AnimOnce},
		{"player_bounce", AnimPingPong},
	}
	for _, tt := range tests {
		a := atlas.Animation(tt.name)
		if a == nil {
			t.Errorf("Animation(%q) = nil", tt.name)
			continue
		}
		if a.Mode != tt.mode {
			t.Errorf("Animation(%q).Mode = %d, want %d", tt.name, a.Mode, tt.mode)
		}
	}
}

func TestLoad_AnimationsListSorted(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := atlas.Animations()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Animations() not sorted: %v", names)
			break
		}
	}
}
