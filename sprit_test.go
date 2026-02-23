package sprit

import (
	"os"
	"testing"
)

func TestLoad_ValidSprites(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 sprites from sprites.hcl.
	names := atlas.Sprites()
	if len(names) != 3 {
		t.Fatalf("expected 3 sprites, got %d: %v", len(names), names)
	}

	// Verify known sprites are present.
	for _, name := range []string{"player_idle", "tree", "rock"} {
		s := atlas.Sprite(name)
		if s == nil {
			t.Errorf("Sprite(%q) = nil, want non-nil", name)
			continue
		}
		if s.Name != name {
			t.Errorf("Sprite(%q).Name = %q", name, s.Name)
		}
		w, h := s.Bounds()
		if w <= 0 || h <= 0 {
			t.Errorf("Sprite(%q).Bounds() = (%d, %d), want positive", name, w, h)
		}
		if s.Image == nil {
			t.Errorf("Sprite(%q).Image = nil", name)
		}
	}
}

func TestLoad_ValidSpriteWithRect(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// player_idle has rect = [0, 0, 8, 8] from an 8x8 image.
	s := atlas.Sprite("player_idle")
	if s == nil {
		t.Fatal("Sprite(\"player_idle\") = nil")
	}
	w, h := s.Bounds()
	if w != 8 || h != 8 {
		t.Errorf("player_idle.Bounds() = (%d, %d), want (8, 8)", w, h)
	}
}

func TestLoad_SpriteLookupMissing(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := atlas.Sprite("nonexistent")
	if s != nil {
		t.Errorf("Sprite(\"nonexistent\") = %v, want nil", s)
	}
}

func TestLoad_SpritesListSorted(t *testing.T) {
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := atlas.Sprites()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Sprites() not sorted: %v", names)
			break
		}
	}
}

func TestLoad_MinimalSprite(t *testing.T) {
	fsys := os.DirFS("testdata/minimal")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := atlas.Sprites()
	if len(names) != 1 {
		t.Fatalf("expected 1 sprite, got %d: %v", len(names), names)
	}

	s := atlas.Sprite("pixel")
	if s == nil {
		t.Fatal("Sprite(\"pixel\") = nil")
	}
	w, h := s.Bounds()
	if w != 1 || h != 1 {
		t.Errorf("pixel.Bounds() = (%d, %d), want (1, 1)", w, h)
	}
}

func TestLoad_InvalidHCL(t *testing.T) {
	fsys := os.DirFS("testdata/invalid")
	_, err := Load(fsys)
	if err == nil {
		t.Fatal("expected error from invalid testdata, got nil")
	}
}

func TestLoad_EmptyFS(t *testing.T) {
	// An empty filesystem should produce an atlas with no sprites.
	fsys := os.DirFS(t.TempDir())
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(atlas.Sprites()) != 0 {
		t.Errorf("expected 0 sprites, got %d", len(atlas.Sprites()))
	}
}

func TestLoad_SpriteFullImage(t *testing.T) {
	// "rock" has no rect, so should use the full 8x8 image.
	fsys := os.DirFS("testdata/valid")
	atlas, err := Load(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := atlas.Sprite("rock")
	if s == nil {
		t.Fatal("Sprite(\"rock\") = nil")
	}
	w, h := s.Bounds()
	if w != 8 || h != 8 {
		t.Errorf("rock.Bounds() = (%d, %d), want (8, 8)", w, h)
	}
}
