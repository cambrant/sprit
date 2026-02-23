package sprit

import (
	"os"
	"path/filepath"
	"testing"
)

// helper to load a fixture file from testdata.
func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", path))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return data
}

func TestParseHCL_ValidSprites(t *testing.T) {
	data := loadFixture(t, "valid/sprites.hcl")
	af, err := parseHCL("sprites.hcl", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(af.Sprites) != 3 {
		t.Fatalf("expected 3 sprites, got %d", len(af.Sprites))
	}

	tests := []struct {
		name        string
		file        string
		rectLen     int
		transparent bool
		background  string
	}{
		{"player_idle", "single.png", 4, true, ""},
		{"tree", "single.png", 0, false, "#1a1a2e"},
		{"rock", "single.png", 0, false, ""},
	}

	for i, tt := range tests {
		s := af.Sprites[i]
		if s.Name != tt.name {
			t.Errorf("sprite[%d] name = %q, want %q", i, s.Name, tt.name)
		}
		if s.File != tt.file {
			t.Errorf("sprite[%d] file = %q, want %q", i, s.File, tt.file)
		}
		if len(s.Rect) != tt.rectLen {
			t.Errorf("sprite[%d] rect len = %d, want %d", i, len(s.Rect), tt.rectLen)
		}
		if s.Transparent != tt.transparent {
			t.Errorf("sprite[%d] transparent = %v, want %v", i, s.Transparent, tt.transparent)
		}
		if s.Background != tt.background {
			t.Errorf("sprite[%d] background = %q, want %q", i, s.Background, tt.background)
		}
	}
}

func TestParseHCL_ValidAnimations(t *testing.T) {
	data := loadFixture(t, "valid/animations.hcl")
	af, err := parseHCL("animations.hcl", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(af.Animations) != 3 {
		t.Fatalf("expected 3 animations, got %d", len(af.Animations))
	}

	tests := []struct {
		name        string
		file        string
		frameWidth  int
		frameHeight int
		frameCount  int
		mode        string
		speed       int
		transparent bool
		background  string
	}{
		{"player_walk", "sheet.png", 8, 8, 4, "loop", 100, true, ""},
		{"player_die", "sheet.png", 8, 8, 4, "once", 150, false, ""},
		{"player_bounce", "sheet.png", 8, 8, 4, "pingpong", 80, false, "#FF0000"},
	}

	for i, tt := range tests {
		a := af.Animations[i]
		if a.Name != tt.name {
			t.Errorf("animation[%d] name = %q, want %q", i, a.Name, tt.name)
		}
		if a.File != tt.file {
			t.Errorf("animation[%d] file = %q, want %q", i, a.File, tt.file)
		}
		if a.FrameWidth != tt.frameWidth {
			t.Errorf("animation[%d] frame_width = %d, want %d", i, a.FrameWidth, tt.frameWidth)
		}
		if a.FrameHeight != tt.frameHeight {
			t.Errorf("animation[%d] frame_height = %d, want %d", i, a.FrameHeight, tt.frameHeight)
		}
		if a.FrameCount != tt.frameCount {
			t.Errorf("animation[%d] frame_count = %d, want %d", i, a.FrameCount, tt.frameCount)
		}
		if a.Mode != tt.mode {
			t.Errorf("animation[%d] mode = %q, want %q", i, a.Mode, tt.mode)
		}
		if a.Speed != tt.speed {
			t.Errorf("animation[%d] speed = %d, want %d", i, a.Speed, tt.speed)
		}
		if a.Transparent != tt.transparent {
			t.Errorf("animation[%d] transparent = %v, want %v", i, a.Transparent, tt.transparent)
		}
		if a.Background != tt.background {
			t.Errorf("animation[%d] background = %q, want %q", i, a.Background, tt.background)
		}
	}
}

func TestParseHCL_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantErr string
	}{
		{
			name:    "bad HCL syntax",
			fixture: "invalid/bad_syntax.hcl",
			wantErr: "parsing",
		},
		{
			name:    "bad rect (not 4 elements)",
			fixture: "invalid/bad_rect.hcl",
			wantErr: "rect must have exactly 4 elements",
		},
		{
			name:    "invalid animation mode",
			fixture: "invalid/bad_mode.hcl",
			wantErr: "mode must be one of once/loop/pingpong",
		},
		{
			name:    "transparent and background conflict",
			fixture: "invalid/conflict.hcl",
			wantErr: "transparent and background are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadFixture(t, tt.fixture)
			_, err := parseHCL(tt.fixture, data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsSubstring(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseHCL_InlineValid(t *testing.T) {
	// Test a minimal valid sprite config from inline HCL.
	hcl := []byte(`
sprite "minimal" {
  file = "test.png"
}
`)
	af, err := parseHCL("inline.hcl", hcl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(af.Sprites) != 1 {
		t.Fatalf("expected 1 sprite, got %d", len(af.Sprites))
	}
	if af.Sprites[0].Name != "minimal" {
		t.Errorf("name = %q, want %q", af.Sprites[0].Name, "minimal")
	}
}

func TestParseHCL_InlineMissingFile(t *testing.T) {
	// sprite block missing required "file" field should fail HCL decode.
	hcl := []byte(`
sprite "no_file" {
}
`)
	_, err := parseHCL("inline.hcl", hcl)
	if err == nil {
		t.Fatal("expected error for missing file field, got nil")
	}
}

func TestParseHCL_InlineBadSpeed(t *testing.T) {
	hcl := []byte(`
animation "bad_speed" {
  file         = "sheet.png"
  frame_width  = 32
  frame_height = 32
  mode         = "loop"
  speed        = 0
}
`)
	_, err := parseHCL("inline.hcl", hcl)
	if err == nil {
		t.Fatal("expected error for speed=0, got nil")
	}
	if !containsSubstring(err.Error(), "speed must be > 0") {
		t.Errorf("error = %q, want substring %q", err.Error(), "speed must be > 0")
	}
}

func TestParseHCL_InlineAnimConflict(t *testing.T) {
	hcl := []byte(`
animation "conflict" {
  file         = "sheet.png"
  frame_width  = 32
  frame_height = 32
  mode         = "once"
  speed        = 100
  transparent  = true
  background   = "#FF0000"
}
`)
	_, err := parseHCL("inline.hcl", hcl)
	if err == nil {
		t.Fatal("expected error for transparent+background conflict, got nil")
	}
	if !containsSubstring(err.Error(), "mutually exclusive") {
		t.Errorf("error = %q, want substring %q", err.Error(), "mutually exclusive")
	}
}

func TestParseHCL_EmptyFile(t *testing.T) {
	// An empty HCL file is valid — no sprites or animations.
	af, err := parseHCL("empty.hcl", []byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(af.Sprites) != 0 {
		t.Errorf("expected 0 sprites, got %d", len(af.Sprites))
	}
	if len(af.Animations) != 0 {
		t.Errorf("expected 0 animations, got %d", len(af.Animations))
	}
}

func TestParseHCL_MixedFile(t *testing.T) {
	hcl := []byte(`
sprite "hero" {
  file        = "hero.png"
  transparent = true
}

animation "hero_run" {
  file         = "hero_run.png"
  frame_width  = 16
  frame_height = 16
  mode         = "loop"
  speed        = 80
}
`)
	af, err := parseHCL("mixed.hcl", hcl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(af.Sprites) != 1 {
		t.Errorf("expected 1 sprite, got %d", len(af.Sprites))
	}
	if len(af.Animations) != 1 {
		t.Errorf("expected 1 animation, got %d", len(af.Animations))
	}
}

func TestValidateSprite_RectExactlyFour(t *testing.T) {
	tests := []struct {
		name    string
		rect    []int
		wantErr bool
	}{
		{"nil rect", nil, false},
		{"empty rect", []int{}, false},
		{"valid rect", []int{0, 0, 32, 32}, false},
		{"too few", []int{0, 0, 32}, true},
		{"too many", []int{0, 0, 32, 32, 1}, true},
		{"one element", []int{0}, true},
		{"two elements", []int{0, 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &spriteConfig{
				Name: "test",
				File: "test.png",
				Rect: tt.rect,
			}
			err := validateSpriteConfig(s)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAnimation_Modes(t *testing.T) {
	base := func(mode string) *animationConfig {
		return &animationConfig{
			Name:        "test",
			File:        "test.png",
			FrameWidth:  32,
			FrameHeight: 32,
			Mode:        mode,
			Speed:       100,
		}
	}

	validModes := []string{"once", "loop", "pingpong"}
	for _, mode := range validModes {
		t.Run("valid_"+mode, func(t *testing.T) {
			if err := validateAnimationConfig(base(mode)); err != nil {
				t.Errorf("unexpected error for mode %q: %v", mode, err)
			}
		})
	}

	invalidModes := []string{"", "reverse", "bounce", "LOOP", "Once"}
	for _, mode := range invalidModes {
		t.Run("invalid_"+mode, func(t *testing.T) {
			if err := validateAnimationConfig(base(mode)); err == nil {
				t.Errorf("expected error for mode %q, got nil", mode)
			}
		})
	}
}

// containsSubstring reports whether s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
