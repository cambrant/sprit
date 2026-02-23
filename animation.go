package sprit

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// AnimationMode defines how an animation plays back.
type AnimationMode int

const (
	// AnimOnce plays the animation once and stops on the last frame.
	AnimOnce AnimationMode = iota
	// AnimLoop loops back to the first frame after the last.
	AnimLoop
	// AnimPingPong reverses direction at each end.
	AnimPingPong
)

// Animation represents a sequence of frames with a playback mode.
type Animation struct {
	Name   string
	Frames []*ebiten.Image
	Mode   AnimationMode
	Speed  time.Duration // duration per frame

	current   int           // current frame index
	elapsed   time.Duration // time accumulated since last frame change
	direction int           // +1 or -1, used for pingpong
	finished  bool          // true when a "once" animation has ended
}

// Update advances the animation by dt. Call once per game tick.
func (a *Animation) Update(dt time.Duration) {
	if a.finished || len(a.Frames) <= 1 {
		return
	}

	a.elapsed += dt
	for a.elapsed >= a.Speed {
		a.elapsed -= a.Speed
		a.advance()
		if a.finished {
			break
		}
	}
}

// advance moves the animation forward by one frame according to its mode.
func (a *Animation) advance() {
	switch a.Mode {
	case AnimOnce:
		if a.current < len(a.Frames)-1 {
			a.current++
		}
		if a.current == len(a.Frames)-1 {
			a.finished = true
		}

	case AnimLoop:
		a.current = (a.current + 1) % len(a.Frames)

	case AnimPingPong:
		next := a.current + a.direction
		if next >= len(a.Frames) {
			// Reverse at the end — step back from the last frame.
			a.direction = -1
			a.current = len(a.Frames) - 2
		} else if next < 0 {
			// Reverse at the start — step forward from the first frame.
			a.direction = 1
			a.current = 1
		} else {
			a.current = next
		}
	}
}

// Draw draws the current frame at the given screen position (top-left corner).
func (a *Animation) Draw(screen *ebiten.Image, x, y float64) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x, y)
	screen.DrawImage(a.Frame(), opts)
}

// DrawWithOptions draws the current frame using custom DrawImageOptions.
func (a *Animation) DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions) {
	screen.DrawImage(a.Frame(), opts)
}

// Frame returns the current frame image.
func (a *Animation) Frame() *ebiten.Image {
	return a.Frames[a.current]
}

// IsFinished returns true if a "once" mode animation has played to completion.
// Always returns false for "loop" and "pingpong" modes.
func (a *Animation) IsFinished() bool {
	return a.finished
}

// Reset restarts the animation from the first frame.
func (a *Animation) Reset() {
	a.current = 0
	a.elapsed = 0
	a.direction = 1
	a.finished = false
}

// parseMode converts a mode string from HCL to an AnimationMode constant.
func parseMode(s string) (AnimationMode, error) {
	switch s {
	case "once":
		return AnimOnce, nil
	case "loop":
		return AnimLoop, nil
	case "pingpong":
		return AnimPingPong, nil
	default:
		return 0, fmt.Errorf("unknown animation mode %q", s)
	}
}

// clone creates an independent copy of the animation with its own playback
// state. The frame images are shared (not copied).
func (a *Animation) clone() *Animation {
	return &Animation{
		Name:      a.Name,
		Frames:    a.Frames, // shared slice — frames are immutable ebiten images
		Mode:      a.Mode,
		Speed:     a.Speed,
		current:   0,
		elapsed:   0,
		direction: 1,
		finished:  false,
	}
}
