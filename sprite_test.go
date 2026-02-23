package sprit

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	colorRed   = color.NRGBA{255, 0, 0, 255}
	colorGreen = color.NRGBA{0, 255, 0, 255}
)

func TestSprite_Bounds(t *testing.T) {
	sprite := &Sprite{
		Name:  "test",
		Image: ebiten.NewImage(16, 24),
		W:     16,
		H:     24,
	}
	w, h := sprite.Bounds()
	if w != 16 || h != 24 {
		t.Errorf("Bounds() = (%d, %d), want (16, 24)", w, h)
	}
}

func TestSprite_Draw(t *testing.T) {
	// Create a small sprite with a known pixel.
	spriteImg := ebiten.NewImage(2, 2)
	spriteImg.Fill(colorRed)

	sprite := &Sprite{
		Name:  "red_box",
		Image: spriteImg,
		W:     2,
		H:     2,
	}

	// Draw onto a screen and verify no panic.
	screen := ebiten.NewImage(10, 10)
	sprite.Draw(screen, 3, 4)
}

func TestSprite_DrawWithOptions(t *testing.T) {
	spriteImg := ebiten.NewImage(2, 2)
	spriteImg.Fill(colorGreen)

	sprite := &Sprite{
		Name:  "green_box",
		Image: spriteImg,
		W:     2,
		H:     2,
	}

	screen := ebiten.NewImage(10, 10)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(5, 5)
	sprite.DrawWithOptions(screen, opts)
}

func TestSprite_Bounds_SinglePixel(t *testing.T) {
	sprite := &Sprite{
		Name:  "pixel",
		Image: ebiten.NewImage(1, 1),
		W:     1,
		H:     1,
	}
	w, h := sprite.Bounds()
	if w != 1 || h != 1 {
		t.Errorf("Bounds() = (%d, %d), want (1, 1)", w, h)
	}
}
