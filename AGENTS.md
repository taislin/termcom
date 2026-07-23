# AGENTS.md

## Project: termcom — ASCII X-COM Demake in Go

### Overview
A faithful demake of X-COM: UFO Defense (1994, MicroProse) rendered entirely in
coloured ASCII on a terminal. Built with Go + tcell. All gameplay screens — Geoscape,
Base, Battlescape — are rendered as colored ASCII art.

### Build & Run
```bash
go run ./cmd/termcom
# or
make run
```

### Test & Lint
```bash
make test           # Run all tests
make test-cover     # Run tests with coverage report
make lint           # Run go vet + staticcheck
make build          # Build binary
make clean          # Remove binary and coverage
```

**Windows (no make):**
```powershell
go test ./... -v                    # Run all tests
go vet ./...                        # Run go vet
staticcheck ./...                   # Run staticcheck
go build -ldflags="-X github.com/taislin/termcom/internal/engine.GameVersion=$(Get-Content VERSION)" -o termcom.exe ./cmd/termcom
```

### Coverage
- `internal/data` — 78%
- `internal/soldier` — 65%
- `internal/save` — 70%
- `internal/geo` — 32%
- `internal/battle` — 38%
- `internal/base` — 28%
- `internal/audio` — 0% (platform-specific, no tests)
- `internal/engine` — 5%

### Dependencies
- `github.com/gdamore/tcell/v3` — Terminal rendering, input, colors
- `github.com/ebitengine/oto/v3` — Cross-platform audio (Windows MIDI synthesis)
- `github.com/gorilla/websocket` — WebSocket for browser version

### Architecture
```
cmd/
  termcom/              Main game entry point (with icon.ico + .syso)
  termcom_battle/       Interactive battle launcher (menu, custom battles)
  test_aliens/          Alien roster viewer (colored console output)
  webserver/            Web server for browser version (xterm.js)
maps/
  *.json                Custom battle definitions (name, author, date, units, victory)
internal/
  engine/game.go           Game state machine, main loop, input dispatch
  engine/screen.go         Low-level screen/cell rendering, FrameBuffer, styles
  engine/custom_battle.go  Custom battle selection screen (split-panel, JSON loading)
  engine/debrief.go        After-action report screen (kills, casualties, stat gains, loot)
  engine/portrait.go       Soldier/alien portrait rendering (half-block PixelImage)
  engine/vfx.go            True-color lighting, alpha blending
  engine/particles.go      Particle system with sync.Pool (explosions, smoke)
  engine/filters.go        Vision filters (night vision, thermal overlay)
  engine/water.go          Animated water with sine-wave color cycling
  engine/camera.go         Screen shake with decay and thread-safe offsets
  engine/menu.go           Title screen with per-character glow effect
  engine/help.go           Help screen system (Geoscape, Base, Battlescape, Research)
  engine/options.go        Options screen
  engine/difficulty.go     Difficulty selection screen
  engine/encyclopedia.go   Encyclopedia/unlocked tech viewer
  engine/slotpicker.go     Save/Load slot picker
  engine/config.go         Config and language integration
  geo/geoscape.go          Geoscape: regional dashboard, time, interceptions, minimap
  geo/world.go             World map data (equirectangular ASCII)
  geo/ufo.go               UFO spawning, movement
  geo/interceptor.go       Interceptor launch, dogfight, weapon systems
  geo/dogfight.go           Dedicated dogfight screen (turn-based, HP/ammo bars, fire/recall)
  geo/transfer.go          Transport movement between bases
  battle/battlescape.go    Battlescape: turn logic, TU, LOS, VFX, custom victory conditions
  battle/map.go            Tactical map generation (crash sites, terror, forest, etc.)
  battle/gas.go            Volumetric smoke/poison gas grid with diffusion
  battle/unit.go           Soldiers and aliens on the tactical map
  battle/ai.go             Alien AI (patrol, seek, attack, flee, flank, retreat, senses)
  battle/input.go          Battlescape input handling (mouse + keyboard)
  base/base.go             Base management screen
  base/facility.go         Facility types, construction, base state, hangars
  base/equip.go            Soldier equipment screen
  base/research.go         Research screen
  base/manufacture.go      Manufacturing screen
  soldier/soldier.go       Soldier stats, ranking, inventory
  data/items.go            Weapons, armor, items (RuleItems map)
  data/aliens.go           Alien species, stat blocks, portraits, morphology types
  data/research.go         Research topic struct and dynamic lookup
  data/techgen.go          Procedural tech tree generator (DAG, tiers, cost variance)
  data/procedural.go       Procedural alien species + morphology + portrait generation
  save/save.go             Save/load system (JSON, version migration v1-v3)
  mapgen/mapgen.go        CDDA-style mapgen chunk loader (JSON fragments, weighted pools, place_nested)
  language/               Multi-language system (en, zh, es, fr, ru, pt, ja, ko)
  audio/audio_common.go    Platform-independent audio dispatch
  audio/audio_windows.go   Windows MIDI-based sound synthesis
  audio/audio_other.go     Linux/macOS stub (oto PCM synthesis)
```

### Code Conventions
- Package names are short, lowercase, single-word
- Exported types/functions use PascalCase
- Unexported helpers use camelCase
- No comments in code unless non-obvious logic
- Error handling: log.Fatal for unrecoverable, return error otherwise
- All rendering via tcell; no raw fmt.Print in game code
- Coordinates: (x, y) where x=col, y=row (screen convention)
- Map tiles: rune-based, colored via tcell.Style
- **NO EMOJIS**: Never use emoji characters (U+1F300-U+1F9FF, U+FE00-U+FE0F, U+200D, etc.) in tileChars or anywhere in game code. Use Unicode symbols from BMP only (U+0000-U+FFFF) — box drawing, technical symbols, miscellaneous symbols (⚙, ⌁, etc.) are fine.
- Always update `docs/manual.md` whenever game data, balance, or mechanics change.
- **Translations**: Every language file in `internal/language/` must be kept in sync. When adding or changing a string key, add/change it in ALL language files (en, zh, es, fr, ru, pt, ja, ko). The language files are: `en.go`, `zh.go`, `es.go`, `fr.go`, `ru.go`, `pt.go`, `ja.go`, `ko.go`. Never add a key to only one file.
- **Version bumps**: Before every push (unless told otherwise), increment `VERSION` by +0.0.1 (e.g. 0.45 → 0.45.1). Only change the middle field when explicitly told. `GameVersion` in `internal/engine/config.go` reads from the `VERSION` file at startup via `init()`; the menu displays it as the centered subtitle. There is no per-language version string to update.
- **Pre-commit checks**: Before committing, always run `make lint` (go vet + staticcheck). On Windows, run `go vet ./... && staticcheck ./...`. staticcheck MUST pass. Never commit code that staticcheck reports warnings on.

### Game Conventions (faithful to original X-COM)
- Time Units (TU) for all actions in Battlescape
- Geoscape runs in real-time with pause (time compression)
- Research requires scientists + time
- Manufacturing requires engineers + time + materials
- Soldiers gain stats from combat experience
- Line-of-sight uses Bresenham raycasting
- Cover system: tiles have 0-100% damage reduction (walls 80%, rocks 70%, trees 60%, bushes 40%)
- Procedural alien species + morphology + portraits per run
- Weighted variant pools: multiple JSON fragments with the same `id` form a weighted pool; `Get(id)` returns a random variant respecting `weight`.
- Chunk nesting (`place_nested`): chunks can stamp sub-chunks at offsets; nested offsets rotate with the parent. Max recursion depth 10.

### Key Bindings (Geoscape)
| Key | Action |
|-----|--------|
| Space | Pause / unpause time |
| 1-4 | Time compression |
| B | Open base |
| L | Launch interceptor |
| A | Autoresolve nearest UFO |
| M | Respond to mission |
| R | Dispatch transport |
| E | Open encyclopedia |
| F5 / F9 | Save / Load |
| Q | Quit |

### Key Bindings (Battlescape)
| Key | Action |
|-----|--------|
| Arrow keys | Move cursor |
| W/A/S/D (Smart mode) | Pan map view |
| W/A/S/D (Fire mode) | Move cursor |
| W/A/S/D (Move mode) | Move selected unit 1 tile |
| Enter | Select / confirm |
| Space (Smart mode) | Enter Move mode |
| Space (Fire mode) | Cancel to Smart |
| Space (Move mode) | Confirm movement |
| Q | Cycle soldiers |
| R | Reload |
| E / N | End turn |
| G | Grenade |
| X | Cycle input mode (Smart / Fire / Move) |
| H | Use medikit |
| C | Crouch |
| V | Toggle vision mode (Normal / Night / Thermal) |
| P | Psi attack |
| Esc (Smart mode) | Quit confirm |
| Esc (Fire / Move mode) | Return to Smart mode |
| ? | Help |

### Key Bindings (Equipment)
| Key | Action |
|-----|--------|
| ↑/↓ | Select soldier |
| 1 / 2 | Weapon / Armor slot |
| Tab | Cycle available items |
| Space | Equip selected item |
| A | Auto-equip all soldiers (best weapon + armor) |
| Esc | Back |

### Key Bindings (Dogfight Screen)
| Key | Action |
|-----|--------|
| F | Fire weapon at UFO |
| [ / ← | Close distance |
| ] / → | Increase distance |
| - / ← | Close distance (alt) |
| = / + | Increase distance (alt) |
| M | Cycle combat mode (Attack/Cautious/Breakoff) |
| B / Esc | Break off / Recall interceptor |
| Any key | Dismiss result (UFO destroyed / interceptor destroyed / disengaged) |

### Key Bindings (Debrief Screen)
| Key | Action |
|-----|--------|
| Enter / Space / Esc | Dismiss after-action report |

### Key Bindings (Battlescape Mouse)
| Action | Input |
|--------|-------|
| Select/Move | Left click |
| Target/Attack | Left click on enemy |
| Cancel | Right click |
| Scroll | Mouse wheel |

**Mouse behaviour per mode:**
| Mode | Left click on empty | Left click on enemy | Left click on friendly | WASD |
|------|---------------------|---------------------|------------------------|------|
| Smart | Move selected unit | Fire weapon | Select unit | Pan map |
| Fire | (move cursor) | Fire weapon | (no-op) | Move cursor |
| Move | Move to tile (pathfinding) | Move as close as possible | (no-op) | Move cursor |

### Mobile Touch Controls
Mobile layout activates automatically when the browser connects with `cols < 100`.
`TouchMode` can also be set manually in config.json (`"touch_mode": true`).

**Touch gestures:**
| Gesture | Action |
|---------|--------|
| Tap | Left click (select, move, fire) |
| Long press (500ms) | Right click (cancel) |
| Vertical drag | Scroll (mouse wheel) |

**On-screen control menu:**
The `[=]` hamburger button in the top-right corner opens a touch-friendly button overlay.
The menu auto-shows on first touch of each screen. Context-sensitive buttons per screen:

- **Geoscape**: Pause, Speed 1-4, Base, Launch, Save, Load, Help
- **Battlescape**: Select, Move, Fire, Reload, End Turn, Grenade, Medikit, Crouch, Cycle, Help
- **Base**: Facilities, Soldiers, Research, Manufacture, Transfer, Hangars, Back, Help
- **Dogfight**: Fire, Mode, Break Off, Back, Help
- **Other screens**: Back, Help

**Responsive layouts (cols < 100):**
- Battlescape: sidebar collapses, full-width viewport with compact unit banner
- Geoscape: minimap hidden, region table full-width
- Encyclopedia/CustomBattle: panels stacked vertically
