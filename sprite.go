package sprit

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite represents a single static image, accessible by name.
type Sprite struct {
	Name  string
	Image *ebiten.Image
	W     int
	H     int
}

// Draw draws the sprite at the given screen position (top-left corner).
func (s *Sprite) Draw(screen *ebiten.Image, x, y float64) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x, y)
	screen.DrawImage(s.Image, opts)
}

// DrawWithOptions draws the sprite using the provided DrawImageOptions,
// giving the caller full control over transform and compositing.
func (s *Sprite) DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions) {
	screen.DrawImage(s.Image, opts)
}

// Bounds returns the width and height of the sprite.
func (s *Sprite) Bounds() (w, h int) {
	return s.W, s.H
}
