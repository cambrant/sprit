# Architecture — sprit

## Mission Statement

**sprit** is a minimal Go library for managing 2D sprite and animation assets in
Ebitengine games. It reads HCL metadata files that describe sprites and
animations, loads the referenced PNG images from an `fs.FS` (typically
`embed.FS`), and exposes them by name through a single `Atlas` type. The library
handles sprite sheet extraction, transparency modes, background filling, and
frame-based animation playback (once, loop, ping-pong) — the repetitive
plumbing that every small 2D game needs but shouldn't have to rewrite.

The intended consumer is a small, self-contained Ebitengine game that embeds its
assets with `//go:embed` and wants a declarative, file-driven way to define
sprites and animations without writing image-loading boilerplate.

---

## Design Principles

1. **Library, not framework** — sprit provides types and functions. It does not
   own the game loop, impose an entity system, or require a specific project
   layout from the consumer.

2. **Declarative asset definitions** — sprite and animation metadata lives in
   HCL files alongside the images. Adding a new sprite means editing an HCL
   file, not writing Go code.

3. **Minimal public API** — one `Load` function, one `Atlas` type, one `Sprite`
   type, one `Animation` type. Utility helpers are provided for common drawing
   operations but nothing beyond what small games need.

4. **Accept `fs.FS`, not `embed.FS`** — the public API accepts the `io/fs.FS`
   interface, so consumers can pass `embed.FS`, `os.DirFS`, or any other
   filesystem implementation. This makes the library testable without embedding.

5. **Zero global state** — all state lives in the returned `Atlas` and its
   children. Multiple atlases can coexist.

---

## Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/hajimehoshi/ebiten/v2` | `*ebiten.Image` as the sprite/frame type; provides `DrawImageOptions` for rendering |
| `github.com/hashicorp/hcl/v2` | Parse HCL metadata files that describe sprites and animations |
| Go standard library (`image`, `image/png`, `image/color`, `io/fs`, `time`) | Image decoding, filesystem abstraction, timing |

No other external dependencies.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────┐
│  Consumer Game                                           │
│                                                          │
│   //go:embed assets/*                                    │
│   var assets embed.FS                                    │
│                                                          │
│   atlas, err := sprit.Load(assets)                       │
│   player := atlas.Sprite("player_idle")                  │
│   walk   := atlas.Animation("player_walk")               │
└──────────┬───────────────────────────────────────────────┘
           │ fs.FS
           ▼
┌──────────────────────────────────────────────────────────┐
│  sprit.Load(fsys)                                        │
│                                                          │
│  1. Walk fsys, discover *.hcl files                      │
│  2. Parse each HCL file → spriteConfig / animationConfig │
│  3. Load referenced PNG files (cached by path)           │
│  4. Process images (sub-rect, transparency, background)  │
│  5. Build Sprite and Animation objects                   │
│  6. Return populated Atlas                               │
└──────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────┐
│  Atlas                                                   │
│  ├── sprites    map[string]*Sprite                       │
│  │   └── Sprite { Name, Image, W, H }                   │
│  ├── animations map[string]*Animation                    │
│  │   └── Animation { Name, Frames[], Mode, Speed, ... } │
│  └── images     map[string]*ebiten.Image  (cache)        │
└──────────────────────────────────────────────────────────┘
```

---

## HCL Schema

All HCL files found in the provided `fs.FS` are parsed. Each file may contain
any mix of `sprite` and `animation` blocks.

### Sprite Block

```hcl
sprite "player_idle" {
  file        = "player_idle.png"       # required — image file path relative to fs root
  rect        = [0, 0, 32, 32]          # optional — [x, y, width, height] sub-rect
  transparent = true                    # optional — preserve alpha channel as-is
  background  = "#1a1a2e"               # optional — fill transparent pixels with this color
}
```

**Field reference:**

| Field | Type | Required | Description |
|---|---|---|---|
| `file` | string | yes | Path to the PNG file within the fs.FS |
| `rect` | list(number) | no | `[x, y, w, h]` sub-rectangle to extract from the image. If omitted, the entire image is used. |
| `transparent` | bool | no | When `true`, the alpha channel is preserved as-is. |
| `background` | string | no | Hex color (e.g. `"#FF0000"`). Transparent pixels are filled with this color. |

**Transparency rules (evaluated in order):**

1. `transparent = true` → alpha channel preserved. Image is used as-is.
2. `background = "#color"` → transparent pixels replaced with the given color.
3. Neither set → transparent pixels replaced with white (`#FFFFFF`).

`transparent` and `background` are mutually exclusive. Specifying both is a
validation error.

### Animation Block

```hcl
animation "player_walk" {
  file         = "player_walk.png"      # required — sprite sheet file
  frame_width  = 32                     # required — width of each frame in pixels
  frame_height = 32                     # required — height of each frame in pixels
  frame_count  = 6                      # optional — number of frames (default: image_width / frame_width)
  row          = 0                      # optional — which row to read from (default: 0)
  mode         = "loop"                 # required — "once", "loop", or "pingpong"
  speed        = 100                    # required — milliseconds per frame
  transparent  = true                   # optional — same semantics as sprite block
  background   = "#1a1a2e"              # optional — same semantics as sprite block
}
```

**Field reference:**

| Field | Type | Required | Description |
|---|---|---|---|
| `file` | string | yes | Path to the sprite sheet PNG |
| `frame_width` | number | yes | Width of a single frame in pixels |
| `frame_height` | number | yes | Height of a single frame in pixels |
| `frame_count` | number | no | Number of frames to extract. Default: `image_width / frame_width` |
| `row` | number | no | Row index (0-based) for multi-row sheets. Default: `0` |
| `mode` | string | yes | Playback mode: `"once"`, `"loop"`, or `"pingpong"` |
| `speed` | number | yes | Milliseconds per frame |
| `transparent` | bool | no | Same rules as sprite block |
| `background` | string | no | Same rules as sprite block |

Frames are extracted left-to-right from the sprite sheet at the given row.

---

## Data Structures

### Core Types

```go
// Atlas is the top-level container returned by Load.
// It holds all sprites and animations parsed from HCL files.
type Atlas struct {
    sprites    map[string]*Sprite
    animations map[string]*Animation
    images     map[string]*ebiten.Image // loaded image cache, keyed by file path
}

// Sprite represents a single static image, accessible by name.
type Sprite struct {
    Name  string
    Image *ebiten.Image
    W     int
    H     int
}

// Animation represents a sequence of frames with a playback mode.
type Animation struct {
    Name      string
    Frames    []*ebiten.Image
    Mode      AnimationMode
    Speed     time.Duration   // duration per frame
    current   int             // current frame index
    elapsed   time.Duration   // time accumulated since last frame change
    direction int             // +1 or -1, used for pingpong
    finished  bool            // true when a "once" animation has ended
}

// AnimationMode defines how an animation plays back.
type AnimationMode int

const (
    AnimOnce     AnimationMode = iota // play once and stop on last frame
    AnimLoop                          // loop back to first frame
    AnimPingPong                      // reverse direction at each end
)
```

### HCL Configuration Types (unexported)

```go
type assetFile struct {
    Sprites    []spriteConfig    `hcl:"sprite,block"`
    Animations []animationConfig `hcl:"animation,block"`
}

type spriteConfig struct {
    Name        string `hcl:"name,label"`
    File        string `hcl:"file"`
    Rect        []int  `hcl:"rect,optional"`
    Transparent bool   `hcl:"transparent,optional"`
    Background  string `hcl:"background,optional"`
}

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
```

---

## Public API

### Loading

```go
// Load walks the given filesystem, parses all *.hcl files, loads referenced
// images, and returns a populated Atlas. Returns an error if any HCL file is
// invalid or any referenced image cannot be loaded.
func Load(fsys fs.FS) (*Atlas, error)
```

### Atlas Methods

```go
// Sprite returns the named sprite, or nil if not found.
func (a *Atlas) Sprite(name string) *Sprite

// Animation returns a new independent playback instance of the named animation.
// Returns nil if the animation name is not found. Each call returns a separate
// instance so multiple game objects can play the same animation independently.
func (a *Atlas) Animation(name string) *Animation

// Sprites returns all sprite names registered in the atlas.
func (a *Atlas) Sprites() []string

// Animations returns all animation names registered in the atlas.
func (a *Atlas) Animations() []string
```

### Sprite Methods

```go
// Draw draws the sprite at the given screen position (top-left corner).
func (s *Sprite) Draw(screen *ebiten.Image, x, y float64)

// DrawWithOptions draws the sprite using the provided DrawImageOptions,
// giving the caller full control over transform and compositing.
func (s *Sprite) DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions)

// Bounds returns the width and height of the sprite.
func (s *Sprite) Bounds() (w, h int)
```

### Animation Methods

```go
// Update advances the animation by dt. Call once per game tick.
// Typical usage: anim.Update(sprit.TickDelta())
func (a *Animation) Update(dt time.Duration)

// Draw draws the current frame at the given screen position.
func (a *Animation) Draw(screen *ebiten.Image, x, y float64)

// DrawWithOptions draws the current frame using custom DrawImageOptions.
func (a *Animation) DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions)

// Frame returns the current frame image.
func (a *Animation) Frame() *ebiten.Image

// IsFinished returns true if a "once" mode animation has played to completion.
// Always returns false for "loop" and "pingpong" modes.
func (a *Animation) IsFinished() bool

// Reset restarts the animation from the first frame.
func (a *Animation) Reset()
```

### Utility Functions

```go
// TickDelta returns the duration of a single game tick at the current TPS.
// Convenience for passing to Animation.Update() from a game's Update() method.
//   anim.Update(sprit.TickDelta())
func TickDelta() time.Duration

// DrawCentered draws img centered at (cx, cy) on screen.
func DrawCentered(screen, img *ebiten.Image, cx, cy float64)

// DrawScaled draws img at (x, y) scaled by the given factor.
func DrawScaled(screen, img *ebiten.Image, x, y, scale float64)

// DrawRotated draws img at (x, y) rotated by angle (radians) around its center.
func DrawRotated(screen, img *ebiten.Image, x, y, angle float64)

// FlipH returns a new image that is a horizontal mirror of img.
func FlipH(img *ebiten.Image) *ebiten.Image

// FlipV returns a new image that is a vertical mirror of img.
func FlipV(img *ebiten.Image) *ebiten.Image
```

---

## Image Processing Pipeline

When a sprite or animation frame is loaded, the following steps are applied:

```
  Read PNG from fs.FS
        │
        ▼
  Decode to image.Image
        │
        ▼
  Extract sub-rect (if rect specified)
        │
        ▼
  Apply transparency mode:
  ┌─────────────────────────────────────────────────┐
  │ transparent=true  → keep alpha as-is            │
  │ background="#hex" → fill transparent → color     │
  │ neither           → fill transparent → white     │
  └─────────────────────────────────────────────────┘
        │
        ▼
  Convert to *ebiten.Image
        │
        ▼
  Store in Atlas (Sprite) or frame slice (Animation)
```

Loaded base images (pre-extraction) are cached in the Atlas by file path. If
multiple sprites reference the same file, the PNG is decoded only once.

---

## File Listing

```
sprit/
├── go.mod                  # module github.com/cambrant/sprit
├── go.sum
├── Makefile                # test, test-verbose, test-coverage targets
├── README.md               # Quick start, usage examples, API overview
├── project-structure.md    # LLM coding style specification
├── docs/
│   ├── ARCHITECTURE.md     # This document
│   └── ROADMAP.md          # Planned features and version history
│
│   # Library source — all files are package sprit
├── sprit.go                # Atlas type, Load() entry point, fs walking
├── sprite.go               # Sprite type, Draw methods, sub-rect extraction
├── animation.go            # Animation type, Update/Draw, playback modes
├── config.go               # HCL config structs, parsing, validation
├── image.go                # Image loading, transparency/background processing, cache
├── util.go                 # Drawing helpers: DrawCentered, FlipH, TickDelta, etc.
│
│   # Tests
├── sprit_test.go           # Atlas loading integration tests
├── sprite_test.go          # Sprite extraction and draw tests
├── animation_test.go       # Animation playback logic tests
├── config_test.go          # HCL parsing and validation tests
├── image_test.go           # Image processing tests
├── util_test.go            # Utility function tests
│
│   # Test fixtures
└── testdata/
    ├── valid/
    │   ├── sprites.hcl     # Valid sprite definitions
    │   ├── animations.hcl  # Valid animation definitions
    │   ├── single.png      # Single sprite image
    │   └── sheet.png       # Sprite sheet with multiple frames
    ├── invalid/
    │   ├── bad_mode.hcl    # Invalid animation mode
    │   ├── bad_rect.hcl    # Invalid rect dimensions
    │   ├── bad_syntax.hcl  # Malformed HCL
    │   └── conflict.hcl    # transparent + background conflict
    └── minimal/
        ├── one.hcl         # Minimal valid config for smoke tests
        └── pixel.png       # 1x1 or small test image
```

### File Responsibilities

| File | Responsibility |
|---|---|
| `sprit.go` | `Atlas` struct, `Load()` function, `fs.FS` walking and HCL discovery, `Sprite()`/`Animation()` lookups |
| `sprite.go` | `Sprite` struct, `Draw`, `DrawWithOptions`, `Bounds` methods |
| `animation.go` | `Animation` struct, `AnimationMode` constants, `Update`, `Draw`, `Frame`, `Reset`, `IsFinished` |
| `config.go` | Unexported HCL config structs, `parseHCL()` function, config validation |
| `image.go` | PNG decoding, sub-rect extraction, transparency/background processing, image cache |
| `util.go` | `TickDelta`, `DrawCentered`, `DrawScaled`, `DrawRotated`, `FlipH`, `FlipV` |

---

## Consumer Usage Example

```go
package main

import (
    "embed"
    "log"

    "github.com/cambrant/sprit"
    "github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/*
var assets embed.FS

type Game struct {
    atlas *sprit.Atlas
    tree  *sprit.Sprite
    walk  *sprit.Animation
}

func NewGame() (*Game, error) {
    atlas, err := sprit.Load(assets)
    if err != nil {
        return nil, err
    }
    return &Game{
        atlas: atlas,
        tree:  atlas.Sprite("tree"),
        walk:  atlas.Animation("player_walk"),
    }, nil
}

func (g *Game) Update() error {
    g.walk.Update(sprit.TickDelta())
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.tree.Draw(screen, 100, 50)
    g.walk.Draw(screen, 200, 150)
}

func (g *Game) Layout(w, h int) (int, int) {
    return 320, 240
}

func main() {
    game, err := NewGame()
    if err != nil {
        log.Fatal(err)
    }
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
```

With the corresponding HCL in `assets/`:

```hcl
# assets/sprites.hcl
sprite "tree" {
  file        = "tileset.png"
  rect        = [128, 64, 32, 32]
  transparent = true
}

# assets/animations.hcl
animation "player_walk" {
  file         = "player_walk.png"
  frame_width  = 32
  frame_height = 32
  frame_count  = 6
  mode         = "loop"
  speed        = 100
  transparent  = true
}
```
