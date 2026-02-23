package sprit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Atlas is the top-level container returned by Load.
// It holds all sprites and animations parsed from HCL files.
type Atlas struct {
	sprites map[string]*Sprite
	images  imageCache
}

// Load walks the given filesystem, parses all *.hcl files, loads referenced
// images, and returns a populated Atlas. Returns an error if any HCL file is
// invalid or any referenced image cannot be loaded.
func Load(fsys fs.FS) (*Atlas, error) {
	atlas := &Atlas{
		sprites: make(map[string]*Sprite),
		images:  make(imageCache),
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
