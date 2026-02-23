# Roadmap — sprit

## Version Plan

Each version below is a self-contained milestone. Versions are implemented in
order; each builds on the previous one. No version should require more than a
focused session of work.

---

### v0.1.0 — Project skeleton, HCL parsing, and config validation

Set up the module, dependencies, Makefile, and the HCL configuration layer.
No image loading yet — this version proves the config schema works end-to-end.

- [x] Initialize `go.mod` with `github.com/cambrant/sprit`
- [x] Add `hashicorp/hcl/v2` and `hajimehoshi/ebiten/v2` dependencies
- [x] Create `Makefile` with `test`, `test-verbose`, `test-coverage` targets
- [x] Create `config.go`:
  - Unexported HCL config structs (`assetFile`, `spriteConfig`, `animationConfig`)
  - `parseHCL(filename string, data []byte) (*assetFile, error)` — parse raw HCL bytes
  - Validation: required fields present, `rect` has exactly 4 elements, `mode` is one of `once`/`loop`/`pingpong`, `speed > 0`, `transparent` and `background` mutually exclusive
- [x] Create `config_test.go`:
  - Table-driven tests for valid sprite configs, valid animation configs
  - Error cases: missing required fields, bad rect, invalid mode, both transparent and background set, bad HCL syntax
- [x] Create `testdata/valid/sprites.hcl` and `testdata/valid/animations.hcl` with representative fixtures
- [x] Create `testdata/invalid/` fixtures: `bad_syntax.hcl`, `bad_rect.hcl`, `bad_mode.hcl`, `conflict.hcl` (transparent + background)
- [x] Verify: `make test` passes

---

### v0.2.0 — Image loading and processing pipeline

Implement the image layer: PNG decoding from `fs.FS`, sub-rect extraction,
transparency/background processing, and the image cache.

- [ ] Create `image.go`:
  - `loadImage(fsys fs.FS, path string) (image.Image, error)` — decode PNG from filesystem
  - `extractRect(img image.Image, x, y, w, h int) (image.Image, error)` — sub-rect extraction with bounds validation
  - `applyTransparency(img image.Image, transparent bool, background string) *image.NRGBA` — apply transparency rules (preserve alpha / fill with color / fill with white)
  - `parseHexColor(hex string) (color.NRGBA, error)` — parse `#RRGGBB` or `#RGB` hex color strings
  - `toEbitenImage(img image.Image) *ebiten.Image` — convert standard image to ebiten image
  - Image cache: `map[string]image.Image` keyed by file path, shared within a single `Load` call
- [ ] Create `image_test.go`:
  - Test PNG decoding from `os.DirFS("testdata/...")`
  - Test sub-rect extraction: valid rect, out-of-bounds rect (error)
  - Test transparency: transparent=true preserves alpha, background fills correct color, default fills white
  - Test hex color parsing: `#FF0000`, `#f00`, invalid strings
- [ ] Create `testdata/valid/single.png` — small test image (e.g. 8x8 with known pixel values)
- [ ] Create `testdata/valid/sheet.png` — small sprite sheet (e.g. 32x8, four 8x8 frames)
- [ ] Create `testdata/minimal/pixel.png` — 1x1 RGBA pixel for smoke tests
- [ ] Verify: `make test` passes

---

### v0.3.0 — Sprite type and Atlas.Sprite() lookup

Wire up the `Sprite` type with its draw methods and connect it to the Atlas.

- [ ] Create `sprite.go`:
  - `Sprite` struct: `Name`, `Image *ebiten.Image`, `W`, `H`
  - `Draw(screen *ebiten.Image, x, y float64)` — draw at position
  - `DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions)` — draw with full control
  - `Bounds() (w, h int)` — return dimensions
- [ ] Create `sprite_test.go`:
  - Test `Bounds()` returns correct dimensions
  - Test `Draw` and `DrawWithOptions` produce expected output (verify draw call doesn't panic, spot-check pixel placement on a small offscreen image)
- [ ] Create `sprit.go` (partial — sprite loading only):
  - `Atlas` struct with `sprites map[string]*Sprite` and `images` cache
  - `Load(fsys fs.FS) (*Atlas, error)` — walk fs, discover HCL, parse, load sprites only (animations deferred to next version)
  - `Sprite(name string) *Sprite` — lookup by name
  - `Sprites() []string` — list all names
  - Internal: `buildSprite(cfg spriteConfig, cache) (*Sprite, error)` — orchestrate image loading + processing for one sprite config
- [ ] Create `sprit_test.go` (partial):
  - Load atlas from `testdata/valid/` — verify sprite count, lookup by name, nil for missing name
  - Load from `testdata/invalid/` — verify errors
- [ ] Create `testdata/minimal/one.hcl` — minimal sprite definition pointing at `pixel.png`
- [ ] Verify: `make test` passes

---

### v0.4.0 — Animation type and playback modes

Implement the `Animation` type with frame extraction and the three playback
modes (once, loop, ping-pong).

- [ ] Create `animation.go`:
  - `AnimationMode` type and constants: `AnimOnce`, `AnimLoop`, `AnimPingPong`
  - `Animation` struct: `Name`, `Frames`, `Mode`, `Speed`, `current`, `elapsed`, `direction`, `finished`
  - `Update(dt time.Duration)` — advance animation state
  - `Draw(screen *ebiten.Image, x, y float64)` — draw current frame
  - `DrawWithOptions(screen *ebiten.Image, opts *ebiten.DrawImageOptions)`
  - `Frame() *ebiten.Image` — return current frame
  - `IsFinished() bool` — true only for `AnimOnce` after last frame
  - `Reset()` — restart from first frame
  - Internal: `parseMode(s string) (AnimationMode, error)`
- [ ] Create `animation_test.go`:
  - Test `AnimLoop`: cycles through all frames and wraps back to 0
  - Test `AnimOnce`: advances to last frame, `IsFinished()` returns true, stays on last frame
  - Test `AnimPingPong`: reaches last frame, reverses, reaches first frame, reverses again
  - Test `Reset()`: restores frame 0, clears finished state
  - Test boundary: single-frame animation, `speed=0` is rejected during config validation
  - Test `Frame()` returns correct image at each step
- [ ] Update `sprit.go`:
  - Add `animations map[string]*Animation` to `Atlas`
  - Add animation building in `Load()`: extract frames from sprite sheet, create `Animation` objects
  - `Animation(name string) *Animation` — returns a new independent playback instance (clone)
  - `Animations() []string` — list all names
  - Internal: `buildAnimation(cfg animationConfig, cache) (*Animation, error)` — extract frames, validate frame_count vs image width
- [ ] Update `sprit_test.go`:
  - Load atlas with both sprites and animations
  - Verify animation lookup, independent instances (two calls return separate state)
  - Verify frame count matches expectations from sprite sheet dimensions
- [ ] Verify: `make test` passes

---

### v0.5.0 — Utility functions

Add the convenience drawing helpers and `TickDelta`.

- [ ] Create `util.go`:
  - `TickDelta() time.Duration` — `time.Second / time.Duration(ebiten.TPS())`
  - `DrawCentered(screen, img *ebiten.Image, cx, cy float64)`
  - `DrawScaled(screen, img *ebiten.Image, x, y, scale float64)`
  - `DrawRotated(screen, img *ebiten.Image, x, y, angle float64)`
  - `FlipH(img *ebiten.Image) *ebiten.Image`
  - `FlipV(img *ebiten.Image) *ebiten.Image`
- [ ] Create `util_test.go`:
  - Test `FlipH`: verify pixel at (0,0) moves to (w-1,0)
  - Test `FlipV`: verify pixel at (0,0) moves to (0,h-1)
  - Test `DrawCentered`, `DrawScaled`, `DrawRotated`: verify draw calls complete without panic, spot-check that output image is non-empty
  - Test `TickDelta`: returns a positive duration
- [ ] Verify: `make test` passes

---

### v0.6.0 — README, documentation, and polish

Final documentation pass. Ensure the library is ready for use in a real game
project.

- [ ] Create `README.md`:
  - Project purpose and one-sentence description
  - Install: `go get github.com/cambrant/sprit`
  - Quick start example (embed assets, load atlas, draw sprite, animate)
  - HCL schema reference (summary with link to ARCHITECTURE.md)
  - API overview table
  - Transparency rules summary
  - Utility functions list
- [ ] Review and finalize `docs/ARCHITECTURE.md` — ensure it matches the actual implementation
- [ ] Review all exported types and functions have GoDoc comments
- [ ] Run `go vet ./...` and fix any warnings
- [ ] Run `make test-coverage` and verify reasonable coverage (target: >80%)
- [ ] Tag `v0.6.0` (or `v0.1.0` public if ready for consumers)
- [ ] Verify: `make test` passes

---

## Future Ideas

These are potential enhancements that are **not** planned for the initial
implementation. They may be added if real game projects reveal a need.

- **Sprite atlas packing** — accept many individual PNGs and pack them into a
  single texture atlas at load time for GPU efficiency.
- **Animation events** — callback hooks at specific frames (e.g. footstep sound
  at frame 3).
- **Animation blending** — cross-fade between two animations over N frames.
- **Aseprite/JSON import** — parse Aseprite's JSON export format as an
  alternative to HCL.
- **Hot reload** — watch the filesystem for HCL/PNG changes and reload assets
  without restarting the game (development mode only).
- **Tiled map integration** — load TMX/TSX tile maps and expose tile layers as
  drawable objects.
- **9-slice sprite** — define stretchable regions for UI panels and borders.
- **Animation state machine** — define named states with transitions and
  conditions in HCL.

---

## Active Bug List

_No known bugs. This section will be updated as issues are discovered._
