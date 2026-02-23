package sprit

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestTickDelta(t *testing.T) {
	dt := TickDelta()
	if dt <= 0 {
		t.Errorf("TickDelta() = %v, want positive duration", dt)
	}
	// At the default 60 TPS, TickDelta should be ~16.6ms.
	expected := time.Second / time.Duration(ebiten.TPS())
	if dt != expected {
		t.Errorf("TickDelta() = %v, want %v", dt, expected)
	}
}

func TestFlipH(t *testing.T) {
	// Create a 4x2 image with a known pixel at (0,0).
	src := ebiten.NewImage(4, 2)
	src.Set(0, 0, color.NRGBA{255, 0, 0, 255})

	flipped := FlipH(src)
	bounds := flipped.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 2 {
		t.Fatalf("FlipH size = %dx%d, want 4x2", bounds.Dx(), bounds.Dy())
	}
}

func TestFlipV(t *testing.T) {
	// Create a 2x4 image with a known pixel at (0,0).
	src := ebiten.NewImage(2, 4)
	src.Set(0, 0, color.NRGBA{0, 255, 0, 255})

	flipped := FlipV(src)
	bounds := flipped.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 4 {
		t.Fatalf("FlipV size = %dx%d, want 2x4", bounds.Dx(), bounds.Dy())
	}
}

func TestDrawCentered(t *testing.T) {
	screen := ebiten.NewImage(20, 20)
	img := ebiten.NewImage(4, 4)
	img.Fill(color.NRGBA{255, 0, 0, 255})

	// Should not panic.
	DrawCentered(screen, img, 10, 10)
}

func TestDrawScaled(t *testing.T) {
	screen := ebiten.NewImage(20, 20)
	img := ebiten.NewImage(2, 2)
	img.Fill(color.NRGBA{0, 0, 255, 255})

	// Should not panic.
	DrawScaled(screen, img, 0, 0, 3.0)
}

func TestDrawRotated(t *testing.T) {
	screen := ebiten.NewImage(20, 20)
	img := ebiten.NewImage(4, 4)
	img.Fill(color.NRGBA{255, 255, 0, 255})

	// Rotate by 0 — should behave like a normal draw at (8, 8).
	DrawRotated(screen, img, 8, 8, 0)

	// Rotate by π (180°) — should not panic.
	screen2 := ebiten.NewImage(20, 20)
	DrawRotated(screen2, img, 8, 8, math.Pi)

	// Rotate by π/4 (45°) — should not panic.
	screen3 := ebiten.NewImage(20, 20)
	DrawRotated(screen3, img, 8, 8, math.Pi/4)
}

func TestFlipH_PreservesDimensions(t *testing.T) {
	src := ebiten.NewImage(8, 4)
	flipped := FlipH(src)
	bounds := flipped.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 4 {
		t.Errorf("FlipH dimensions = %dx%d, want 8x4", bounds.Dx(), bounds.Dy())
	}
}

func TestFlipV_PreservesDimensions(t *testing.T) {
	src := ebiten.NewImage(8, 4)
	flipped := FlipV(src)
	bounds := flipped.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 4 {
		t.Errorf("FlipV dimensions = %dx%d, want 8x4", bounds.Dx(), bounds.Dy())
	}
}

func TestDrawScaled_ZeroScale(t *testing.T) {
	screen := ebiten.NewImage(10, 10)
	img := ebiten.NewImage(4, 4)
	img.Fill(color.NRGBA{255, 0, 0, 255})

	// Zero scale should not panic.
	DrawScaled(screen, img, 0, 0, 0)
}
