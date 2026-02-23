package sprit

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// TickDelta returns the duration of a single game tick at the current TPS.
// Convenience for passing to Animation.Update() from a game's Update() method.
//
//	anim.Update(sprit.TickDelta())
func TickDelta() time.Duration {
	return time.Second / time.Duration(ebiten.TPS())
}

// DrawCentered draws img centered at (cx, cy) on screen.
func DrawCentered(screen, img *ebiten.Image, cx, cy float64) {
	bounds := img.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(cx-w/2, cy-h/2)
	screen.DrawImage(img, opts)
}

// DrawScaled draws img at (x, y) scaled by the given factor.
func DrawScaled(screen, img *ebiten.Image, x, y, scale float64) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(x, y)
	screen.DrawImage(img, opts)
}

// DrawRotated draws img at (x, y) rotated by angle (radians) around its center.
func DrawRotated(screen, img *ebiten.Image, x, y, angle float64) {
	bounds := img.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	opts := &ebiten.DrawImageOptions{}
	// Move origin to center of image for rotation.
	opts.GeoM.Translate(-w/2, -h/2)
	opts.GeoM.Rotate(angle)
	// Translate to final position (x, y is the top-left of the un-rotated image).
	opts.GeoM.Translate(x+w/2, y+h/2)
	screen.DrawImage(img, opts)
}

// FlipH returns a new image that is a horizontal mirror of img.
func FlipH(img *ebiten.Image) *ebiten.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	flipped := ebiten.NewImage(w, h)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(-1, 1)
	opts.GeoM.Translate(float64(w), 0)
	flipped.DrawImage(img, opts)
	return flipped
}

// FlipV returns a new image that is a vertical mirror of img.
func FlipV(img *ebiten.Image) *ebiten.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	flipped := ebiten.NewImage(w, h)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(1, -1)
	opts.GeoM.Translate(0, float64(h))
	flipped.DrawImage(img, opts)
	return flipped
}
