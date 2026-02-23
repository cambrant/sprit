package sprit

import (
	"image"
	"image/color"
	"os"
	"testing"
)

func TestLoadImage_ValidPNG(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)

	img, err := loadImage(fsys, "single.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Errorf("bounds = %dx%d, want 8x8", bounds.Dx(), bounds.Dy())
	}

	// Verify caching: second call returns same object.
	img2, err := loadImage(fsys, "single.png", cache)
	if err != nil {
		t.Fatalf("unexpected error on cached load: %v", err)
	}
	if img != img2 {
		t.Error("expected cached image to be the same object")
	}
}

func TestLoadImage_SpriteSheet(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)

	img, err := loadImage(fsys, "sheet.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 8 {
		t.Errorf("bounds = %dx%d, want 32x8", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_MinimalPixel(t *testing.T) {
	fsys := os.DirFS("testdata/minimal")
	cache := make(imageCache)

	img, err := loadImage(fsys, "pixel.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Errorf("bounds = %dx%d, want 1x1", bounds.Dx(), bounds.Dy())
	}
}

func TestLoadImage_NotFound(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)

	_, err := loadImage(fsys, "nonexistent.png", cache)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestExtractRect_Valid(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)
	img, err := loadImage(fsys, "single.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Extract top-left 4x4 quadrant (should be red).
	sub, err := extractRect(img, 0, 0, 4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bounds := sub.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Errorf("sub bounds = %dx%d, want 4x4", bounds.Dx(), bounds.Dy())
	}

	// Check that pixels are red.
	r, g, b, a := sub.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("pixel (0,0) = (%d,%d,%d,%d), want red (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestExtractRect_FullImage(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)
	img, err := loadImage(fsys, "single.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sub, err := extractRect(img, 0, 0, 8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Bounds().Dx() != 8 || sub.Bounds().Dy() != 8 {
		t.Errorf("sub bounds = %dx%d, want 8x8", sub.Bounds().Dx(), sub.Bounds().Dy())
	}
}

func TestExtractRect_OutOfBounds(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)
	img, err := loadImage(fsys, "single.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name       string
		x, y, w, h int
	}{
		{"exceeds width", 4, 0, 8, 4},
		{"exceeds height", 0, 4, 4, 8},
		{"starts outside", 10, 10, 1, 1},
		{"negative x", -1, 0, 4, 4},
		{"zero width", 0, 0, 0, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractRect(img, tt.x, tt.y, tt.w, tt.h)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestExtractRect_SpriteSheetFrames(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache := make(imageCache)
	img, err := loadImage(fsys, "sheet.png", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Extract frame 0 (red), frame 2 (blue).
	frame0, err := extractRect(img, 0, 0, 8, 8)
	if err != nil {
		t.Fatalf("unexpected error extracting frame 0: %v", err)
	}
	r, g, b, _ := frame0.At(4, 4).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("frame 0 pixel = (%d,%d,%d), want red (255,0,0)", r>>8, g>>8, b>>8)
	}

	frame2, err := extractRect(img, 16, 0, 8, 8)
	if err != nil {
		t.Fatalf("unexpected error extracting frame 2: %v", err)
	}
	r, g, b, _ = frame2.At(4, 4).RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 255 {
		t.Errorf("frame 2 pixel = (%d,%d,%d), want blue (0,0,255)", r>>8, g>>8, b>>8)
	}
}

func TestApplyTransparency_Transparent(t *testing.T) {
	// Create a 2x2 image with a transparent pixel.
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})
	img.SetNRGBA(1, 0, color.NRGBA{0, 255, 0, 128}) // semi-transparent
	img.SetNRGBA(0, 1, color.NRGBA{0, 0, 255, 0})   // fully transparent
	img.SetNRGBA(1, 1, color.NRGBA{255, 255, 0, 255})

	out, err := applyTransparency(img, true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// transparent=true preserves alpha as-is.
	_, _, _, a := out.At(1, 0).RGBA()
	if a>>8 != 128 {
		t.Errorf("semi-transparent pixel alpha = %d, want 128", a>>8)
	}
	_, _, _, a = out.At(0, 1).RGBA()
	if a != 0 {
		t.Errorf("transparent pixel alpha = %d, want 0", a)
	}
}

func TestApplyTransparency_Background(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})     // opaque red
	img.SetNRGBA(1, 0, color.NRGBA{0, 255, 0, 128})     // semi-transparent green
	img.SetNRGBA(0, 1, color.NRGBA{0, 0, 0, 0})         // fully transparent
	img.SetNRGBA(1, 1, color.NRGBA{255, 255, 255, 255}) // opaque white

	out, err := applyTransparency(img, false, "#FF0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fully transparent pixel should be filled with red.
	r, g, b, a := out.NRGBAAt(0, 1).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("transparent pixel = (%d,%d,%d,%d), want red fill (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
	}

	// Semi-transparent pixel should also be filled with background.
	r, g, b, a = out.NRGBAAt(1, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("semi-transparent pixel = (%d,%d,%d,%d), want red fill (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
	}

	// Opaque red pixel should be preserved.
	r, g, b, a = out.NRGBAAt(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("opaque pixel = (%d,%d,%d,%d), want (255,0,0,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestApplyTransparency_Default(t *testing.T) {
	// Neither transparent nor background → fill with white.
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{0, 0, 0, 0}) // fully transparent

	out, err := applyTransparency(img, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, g, b, a := out.NRGBAAt(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 || a>>8 != 255 {
		t.Errorf("default fill = (%d,%d,%d,%d), want white (255,255,255,255)", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestParseHexColor_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  color.NRGBA
	}{
		{"#FF0000", color.NRGBA{255, 0, 0, 255}},
		{"#ff0000", color.NRGBA{255, 0, 0, 255}},
		{"#00FF00", color.NRGBA{0, 255, 0, 255}},
		{"#0000FF", color.NRGBA{0, 0, 255, 255}},
		{"#FFFFFF", color.NRGBA{255, 255, 255, 255}},
		{"#000000", color.NRGBA{0, 0, 0, 255}},
		{"#1a1a2e", color.NRGBA{26, 26, 46, 255}},
		// Short form #RGB.
		{"#f00", color.NRGBA{255, 0, 0, 255}},
		{"#0f0", color.NRGBA{0, 255, 0, 255}},
		{"#00f", color.NRGBA{0, 0, 255, 255}},
		{"#fff", color.NRGBA{255, 255, 255, 255}},
		{"#000", color.NRGBA{0, 0, 0, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseHexColor(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseHexColor(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseHexColor_Invalid(t *testing.T) {
	tests := []string{
		"",
		"FF0000",    // missing #
		"#GG0000",   // invalid hex char
		"#FF00",     // wrong length (4)
		"#FF0000FF", // wrong length (8)
		"#",         // just hash
		"#12",       // wrong length (2)
		"#1234567",  // wrong length (7)
		"not a color",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseHexColor(input)
			if err == nil {
				t.Errorf("expected error for %q, got nil", input)
			}
		})
	}
}

func TestApplyTransparency_InvalidBackground(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{0, 0, 0, 0})

	_, err := applyTransparency(img, false, "not-a-color")
	if err == nil {
		t.Fatal("expected error for invalid background color, got nil")
	}
}

func TestLoadImage_CacheIndependence(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	cache1 := make(imageCache)
	cache2 := make(imageCache)

	img1, _ := loadImage(fsys, "single.png", cache1)
	img2, _ := loadImage(fsys, "single.png", cache2)

	// Different caches should produce different objects.
	if img1 == img2 {
		t.Error("expected different objects from different caches")
	}
}
