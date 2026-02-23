package sprit

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png" // register PNG decoder
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// imageCache stores decoded images keyed by file path, shared within a single
// Load call so that multiple sprites referencing the same file only decode once.
type imageCache map[string]image.Image

// loadImage decodes a PNG from the given filesystem path. Results are cached in
// the provided cache; subsequent calls with the same path return the cached copy.
func loadImage(fsys fs.FS, path string, cache imageCache) (image.Image, error) {
	if img, ok := cache[path]; ok {
		return img, nil
	}
	f, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	cache[path] = img
	return img, nil
}

// extractRect extracts a sub-rectangle from img. Returns an error if the
// requested region falls outside the image bounds.
func extractRect(img image.Image, x, y, w, h int) (image.Image, error) {
	bounds := img.Bounds()
	if x < 0 || y < 0 || w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid rect [%d, %d, %d, %d]: values must be non-negative and dimensions positive", x, y, w, h)
	}
	if x+w > bounds.Dx() || y+h > bounds.Dy() {
		return nil, fmt.Errorf("rect [%d, %d, %d, %d] exceeds image bounds %dx%d", x, y, w, h, bounds.Dx(), bounds.Dy())
	}
	sub := image.NewNRGBA(image.Rect(0, 0, w, h))
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			sub.Set(dx, dy, img.At(bounds.Min.X+x+dx, bounds.Min.Y+y+dy))
		}
	}
	return sub, nil
}

// applyTransparency applies the transparency rules to an image:
//  1. transparent=true → alpha channel preserved as-is
//  2. background="#hex" → transparent pixels filled with the given color
//  3. neither → transparent pixels filled with white
func applyTransparency(img image.Image, transparent bool, background string) (*image.NRGBA, error) {
	bounds := img.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	if transparent {
		// Preserve alpha as-is.
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				out.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
			}
		}
		return out, nil
	}

	// Determine fill color: background hex or default white.
	var fill color.NRGBA
	if background != "" {
		c, err := parseHexColor(background)
		if err != nil {
			return nil, err
		}
		fill = c
	} else {
		fill = color.NRGBA{255, 255, 255, 255}
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			ox := x - bounds.Min.X
			oy := y - bounds.Min.Y
			if a == 0 {
				// Fully transparent pixel — replace with fill color.
				out.SetNRGBA(ox, oy, fill)
			} else if a < 0xFFFF {
				// Semi-transparent pixel — replace with fill color.
				out.SetNRGBA(ox, oy, fill)
			} else {
				// Fully opaque — keep original color.
				out.SetNRGBA(ox, oy, color.NRGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: 255,
				})
			}
		}
	}
	return out, nil
}

// parseHexColor parses a hex color string in #RRGGBB or #RGB format and
// returns it as a color.NRGBA with alpha 255.
func parseHexColor(hex string) (color.NRGBA, error) {
	if len(hex) == 0 || hex[0] != '#' {
		return color.NRGBA{}, fmt.Errorf("invalid hex color %q: must start with #", hex)
	}
	hex = hex[1:]
	switch len(hex) {
	case 3:
		r, err := parseHexByte(hex[0], hex[0])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		g, err := parseHexByte(hex[1], hex[1])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		b, err := parseHexByte(hex[2], hex[2])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
	case 6:
		r, err := parseHexByte(hex[0], hex[1])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		g, err := parseHexByte(hex[2], hex[3])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		b, err := parseHexByte(hex[4], hex[5])
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
	default:
		return color.NRGBA{}, fmt.Errorf("invalid hex color #%s: must be 3 or 6 hex digits", hex)
	}
}

// parseHexByte converts two hex character nibbles into a byte value.
func parseHexByte(hi, lo byte) (uint8, error) {
	h, err := hexVal(hi)
	if err != nil {
		return 0, err
	}
	l, err := hexVal(lo)
	if err != nil {
		return 0, err
	}
	return h<<4 | l, nil
}

// hexVal converts a single hex character to its numeric value.
func hexVal(c byte) (uint8, error) {
	c = byte(strings.ToLower(string(c))[0])
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	default:
		return 0, fmt.Errorf("invalid hex character %q", c)
	}
}

// toEbitenImage converts a standard library image to an *ebiten.Image.
func toEbitenImage(img image.Image) *ebiten.Image {
	return ebiten.NewImageFromImage(img)
}
