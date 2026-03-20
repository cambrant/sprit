# sprit

A minimal Go library for managing 2D sprite and animation assets in
[Ebitengine](https://ebitengine.org) games. Define sprites and animations in HCL
metadata files, point them at PNG sprite sheets, and access everything by
name through a single `Atlas` type.

## Install

```
go get github.com/cambrant/sprit
```

## Quick Start

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

func (g *Game) Layout(w, h int) (int, int) { return 320, 240 }

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

Create HCL files alongside images in the embedded directory:

```hcl
# assets/sprites.hcl
sprite "tree" {
  file        = "tileset.png"
  rect        = [128, 64, 32, 32]
  transparent = true
}
```

```hcl
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

## HCL Schema

All `*.hcl` files found in the provided `fs.FS` are parsed. Each file may
contain any mix of `sprite` and `animation` blocks.

### Sprite Block

```hcl
sprite "name" {
  file        = "path/to/image.png"   # required
  rect        = [x, y, w, h]          # optional — sub-rectangle to extract
  transparent = true                   # optional — preserve alpha as-is
  background  = "#RRGGBB"             # optional — fill transparent pixels
}
```

### Animation Block

```hcl
animation "name" {
  file         = "sheet.png"          # required
  frame_width  = 32                   # required
  frame_height = 32                   # required
  frame_count  = 6                    # optional — default: image_width / frame_width
  row          = 0                    # optional — row index for multi-row sheets
  mode         = "loop"               # required — "once", "loop", or "pingpong"
  speed        = 100                  # required — milliseconds per frame
  transparent  = true                 # optional
  background   = "#RRGGBB"           # optional
}
```

### Transparency Rules

Evaluated in order for each sprite or animation frame:

1. `transparent = true` → alpha channel preserved as-is.
2. `background = "#color"` → transparent pixels replaced with the given color.
3. Neither set → transparent pixels replaced with white (`#FFFFFF`).

`transparent` and `background` are mutually exclusive — specifying both is a
validation error.

For the full schema reference, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## API Overview

| Symbol | Description |
|---|---|
| `Load(fsys fs.FS) (*Atlas, error)` | Walk filesystem, parse HCL, load images, return atlas |
| `Atlas.Sprite(name) *Sprite` | Look up a sprite by name (nil if missing) |
| `Atlas.Animation(name) *Animation` | Get an independent animation playback instance (nil if missing) |
| `Atlas.Sprites() []string` | List all sprite names |
| `Atlas.Animations() []string` | List all animation names |
| `Sprite.Draw(screen, x, y)` | Draw at position |
| `Sprite.DrawWithOptions(screen, opts)` | Draw with full `DrawImageOptions` control |
| `Sprite.Bounds() (w, h)` | Get dimensions |
| `Animation.Update(dt)` | Advance playback by dt |
| `Animation.Draw(screen, x, y)` | Draw current frame at position |
| `Animation.DrawWithOptions(screen, opts)` | Draw current frame with options |
| `Animation.Frame() *ebiten.Image` | Get current frame image |
| `Animation.IsFinished() bool` | True when a "once" animation has ended |
| `Animation.Reset()` | Restart from first frame |

### Utility Functions

| Function | Description |
|---|---|
| `TickDelta() time.Duration` | Duration of one game tick at current TPS — pass to `Animation.Update()` |
| `DrawCentered(screen, img, cx, cy)` | Draw image centered at a point |
| `DrawScaled(screen, img, x, y, scale)` | Draw image with uniform scale |
| `DrawRotated(screen, img, x, y, angle)` | Draw image rotated around its center |
| `FlipH(img) *ebiten.Image` | Return horizontally mirrored copy |
| `FlipV(img) *ebiten.Image` | Return vertically mirrored copy |

## License

See [LICENSE](LICENSE) for details.
