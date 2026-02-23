package sprit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Atlas is the top-level container returned by Load.
// It holds all sprites and animations parsed from HCL files.
type Atlas struct {
	sprites    map[string]*Sprite
	animations map[string]*Animation
	images     imageCache
}

// Load walks the given filesystem, parses all *.hcl files, loads referenced
// images, and returns a populated Atlas. Returns an error if any HCL file is
// invalid or any referenced image cannot be loaded.
func Load(fsys fs.FS) (*Atlas, error) {
	atlas := &Atlas{
		sprites:    make(map[string]*Sprite),
		animations: make(map[string]*Animation),
		images:     make(imageCache),
	}

	// Discover and parse all HCL files.
	var configs []*assetFile
	var hclPaths []string

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".hcl") {
			return nil
		}
		hclPaths = append(hclPaths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	// Sort for deterministic processing order.
	sort.Strings(hclPaths)

	for _, path := range hclPaths {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		af, err := parseHCL(path, data)
		if err != nil {
			return nil, err
		}
		configs = append(configs, af)
	}

	// Build sprites from all parsed configs.
	for i, af := range configs {
		dir := filepath.Dir(hclPaths[i])
		for _, sc := range af.Sprites {
			if _, exists := atlas.sprites[sc.Name]; exists {
				return nil, fmt.Errorf("duplicate sprite name %q", sc.Name)
			}
			sprite, err := buildSprite(fsys, dir, sc, atlas.images)
			if err != nil {
				return nil, fmt.Errorf("building sprite %q: %w", sc.Name, err)
			}
			atlas.sprites[sc.Name] = sprite
		}
	}

	// Build animations from all parsed configs.
	for i, af := range configs {
		dir := filepath.Dir(hclPaths[i])
		for _, ac := range af.Animations {
			if _, exists := atlas.animations[ac.Name]; exists {
				return nil, fmt.Errorf("duplicate animation name %q", ac.Name)
			}
			anim, err := buildAnimation(fsys, dir, ac, atlas.images)
			if err != nil {
				return nil, fmt.Errorf("building animation %q: %w", ac.Name, err)
			}
			atlas.animations[ac.Name] = anim
		}
	}

	return atlas, nil
}

// Sprite returns the named sprite, or nil if not found.
func (a *Atlas) Sprite(name string) *Sprite {
	return a.sprites[name]
}

// Sprites returns all sprite names registered in the atlas, sorted
// alphabetically.
func (a *Atlas) Sprites() []string {
	names := make([]string, 0, len(a.sprites))
	for name := range a.sprites {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Animation returns a new independent playback instance of the named animation.
// Returns nil if the animation name is not found. Each call returns a separate
// instance so multiple game objects can play the same animation independently.
func (a *Atlas) Animation(name string) *Animation {
	base, ok := a.animations[name]
	if !ok {
		return nil
	}
	return base.clone()
}

// Animations returns all animation names registered in the atlas, sorted
// alphabetically.
func (a *Atlas) Animations() []string {
	names := make([]string, 0, len(a.animations))
	for name := range a.animations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// buildSprite orchestrates image loading and processing for a single sprite
// configuration. The dir parameter is the directory of the HCL file that
// defined this sprite, used to resolve relative image paths.
func buildSprite(fsys fs.FS, dir string, cfg spriteConfig, cache imageCache) (*Sprite, error) {
	// Resolve image path relative to the HCL file's directory.
	imgPath := cfg.File
	if dir != "." {
		imgPath = dir + "/" + cfg.File
	}

	img, err := loadImage(fsys, imgPath, cache)
	if err != nil {
		return nil, err
	}

	// Extract sub-rect if specified.
	var processed = img
	if len(cfg.Rect) == 4 {
		sub, err := extractRect(img, cfg.Rect[0], cfg.Rect[1], cfg.Rect[2], cfg.Rect[3])
		if err != nil {
			return nil, fmt.Errorf("extracting rect: %w", err)
		}
		processed = sub
	}

	// Apply transparency rules.
	nrgba, err := applyTransparency(processed, cfg.Transparent, cfg.Background)
	if err != nil {
		return nil, fmt.Errorf("applying transparency: %w", err)
	}

	eImg := toEbitenImage(nrgba)
	bounds := nrgba.Bounds()

	return &Sprite{
		Name:  cfg.Name,
		Image: eImg,
		W:     bounds.Dx(),
		H:     bounds.Dy(),
	}, nil
}

// buildAnimation orchestrates image loading, frame extraction, and processing
// for a single animation configuration.
func buildAnimation(fsys fs.FS, dir string, cfg animationConfig, cache imageCache) (*Animation, error) {
	// Resolve image path relative to the HCL file's directory.
	imgPath := cfg.File
	if dir != "." {
		imgPath = dir + "/" + cfg.File
	}

	img, err := loadImage(fsys, imgPath, cache)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()

	// Determine frame count.
	frameCount := cfg.FrameCount
	if frameCount == 0 {
		frameCount = bounds.Dx() / cfg.FrameWidth
	}
	if frameCount <= 0 {
		return nil, fmt.Errorf("computed frame_count is %d (image width %d, frame_width %d)",
			frameCount, bounds.Dx(), cfg.FrameWidth)
	}

	// Validate that the sheet is large enough.
	requiredWidth := frameCount * cfg.FrameWidth
	if requiredWidth > bounds.Dx() {
		return nil, fmt.Errorf("frame_count %d * frame_width %d = %d exceeds image width %d",
			frameCount, cfg.FrameWidth, requiredWidth, bounds.Dx())
	}
	rowY := cfg.Row * cfg.FrameHeight
	if rowY+cfg.FrameHeight > bounds.Dy() {
		return nil, fmt.Errorf("row %d * frame_height %d = %d exceeds image height %d",
			cfg.Row, cfg.FrameHeight, rowY+cfg.FrameHeight, bounds.Dy())
	}

	mode, err := parseMode(cfg.Mode)
	if err != nil {
		return nil, err
	}

	// Extract frames left-to-right.
	var frames []*ebiten.Image
	for i := 0; i < frameCount; i++ {
		x := i * cfg.FrameWidth
		sub, err := extractRect(img, x, rowY, cfg.FrameWidth, cfg.FrameHeight)
		if err != nil {
			return nil, fmt.Errorf("extracting frame %d: %w", i, err)
		}
		nrgba, err := applyTransparency(sub, cfg.Transparent, cfg.Background)
		if err != nil {
			return nil, fmt.Errorf("applying transparency to frame %d: %w", i, err)
		}
		frames = append(frames, toEbitenImage(nrgba))
	}

	return &Animation{
		Name:      cfg.Name,
		Frames:    frames,
		Mode:      mode,
		Speed:     time.Duration(cfg.Speed) * time.Millisecond,
		direction: 1,
	}, nil
}
