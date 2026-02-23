package sprit

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// assetFile represents the top-level structure of an HCL asset file.
// Each file may contain any mix of sprite and animation blocks.
type assetFile struct {
	Sprites    []spriteConfig    `hcl:"sprite,block"`
	Animations []animationConfig `hcl:"animation,block"`
}

// spriteConfig holds the parsed configuration for a single sprite block.
type spriteConfig struct {
	Name        string `hcl:"name,label"`
	File        string `hcl:"file"`
	Rect        []int  `hcl:"rect,optional"`
	Transparent bool   `hcl:"transparent,optional"`
	Background  string `hcl:"background,optional"`
}

// animationConfig holds the parsed configuration for a single animation block.
type animationConfig struct {
	Name        string `hcl:"name,label"`
	File        string `hcl:"file"`
	FrameWidth  int    `hcl:"frame_width"`
	FrameHeight int    `hcl:"frame_height"`
	FrameCount  int    `hcl:"frame_count,optional"`
	Row         int    `hcl:"row,optional"`
	Mode        string `hcl:"mode"`
	Speed       int    `hcl:"speed"`
	Transparent bool   `hcl:"transparent,optional"`
	Background  string `hcl:"background,optional"`
}

// parseHCL parses raw HCL bytes into an assetFile. The filename is used for
// error messages only.
func parseHCL(filename string, data []byte) (*assetFile, error) {
	var af assetFile
	if err := hclsimple.Decode(filename, data, nil, &af); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}
	if err := validateAssetFile(&af); err != nil {
		return nil, fmt.Errorf("validating %s: %w", filename, err)
	}
	return &af, nil
}

// validateAssetFile checks all sprite and animation configs for semantic
// correctness beyond what HCL struct tags enforce.
func validateAssetFile(af *assetFile) error {
	for _, s := range af.Sprites {
		if err := validateSpriteConfig(&s); err != nil {
			return fmt.Errorf("sprite %q: %w", s.Name, err)
		}
	}
	for _, a := range af.Animations {
		if err := validateAnimationConfig(&a); err != nil {
			return fmt.Errorf("animation %q: %w", a.Name, err)
		}
	}
	return nil
}

// validateSpriteConfig validates a single sprite configuration block.
func validateSpriteConfig(s *spriteConfig) error {
	if s.File == "" {
		return fmt.Errorf("file is required")
	}
	if len(s.Rect) != 0 && len(s.Rect) != 4 {
		return fmt.Errorf("rect must have exactly 4 elements [x, y, w, h], got %d", len(s.Rect))
	}
	if s.Transparent && s.Background != "" {
		return fmt.Errorf("transparent and background are mutually exclusive")
	}
	return nil
}

// validateAnimationConfig validates a single animation configuration block.
func validateAnimationConfig(a *animationConfig) error {
	if a.File == "" {
		return fmt.Errorf("file is required")
	}
	if a.FrameWidth <= 0 {
		return fmt.Errorf("frame_width must be > 0")
	}
	if a.FrameHeight <= 0 {
		return fmt.Errorf("frame_height must be > 0")
	}
	if a.Speed <= 0 {
		return fmt.Errorf("speed must be > 0")
	}
	switch a.Mode {
	case "once", "loop", "pingpong":
		// valid
	default:
		return fmt.Errorf("mode must be one of once/loop/pingpong, got %q", a.Mode)
	}
	if a.Transparent && a.Background != "" {
		return fmt.Errorf("transparent and background are mutually exclusive")
	}
	return nil
}
