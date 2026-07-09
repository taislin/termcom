# AGENTS.md

## Project: ycom — ASCII X-COM Demake in Go

### Overview
A faithful demake of X-COM: UFO Defense (1994) rendered entirely in ASCII on a
terminal. Built with Go + tcell. All gameplay screens — Geoscape, Base, Battlescape
— are rendered as colored ASCII art.

### Build & Run
```bash
export PATH=/home/civ13/go/bin:$PATH
cd /home/civ13/gamedev/ycom
go run ./cmd/ycom
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

### Coverage
- `internal/data` — 14%
- `internal/soldier` — 92%
- `internal/geo` — 15%
- `internal/battle` — 25%
- `internal/base` — 19%
- `internal/engine` — 1%
- `internal/save` — 81%

### Dependencies
- `github.com/gdamore/tcell/v3` — Terminal rendering, input, colors

### Architecture
```
cmd/ycom/main.go          Entry point
internal/
  engine/game.go           Game state machine, main loop, input dispatch
  engine/screen.go         Low-level screen/cell rendering helpers + FrameBuffer
  engine/vfx.go            True-color lighting, alpha blending, half-block rendering
  engine/particles.go      Particle system with sync.Pool (explosions, smoke, rain)
  engine/filters.go        Vision filters (night vision, thermal overlay)
  engine/water.go          Animated water with sine-wave color cycling
  engine/camera.go         Screen shake with decay and thread-safe offsets
  engine/menu.go           Title screen with per-character glow effect
  geo/geoscape.go          Geoscape: regional dashboard, time, interceptions, minimap
  geo/world.go             World map data (equirectangular ASCII)
  geo/ufo.go               UFO spawning, movement
  geo/interceptor.go       Interceptor launch, dogfight, weapon systems
  battle/battlescape.go    Battlescape: turn logic, TU, line-of-sight, VFX integration
  battle/map.go            Tactical map generation (crash sites, terror)
  battle/gas.go            Volumetric smoke/poison gas grid with diffusion
  battle/unit.go           Soldiers and aliens on the tactical map
  battle/ai.go             Alien AI (patrol, seek, attack, flee, flank, retreat)
  base/base.go             Base management screen, hangar management
  base/facility.go         Facility types and construction, base state
  soldier/soldier.go       Soldier stats, ranking, inventory
  data/items.go            Weapons, armor, items
  data/aliens.go           Alien species, stat blocks
  data/research.go         Research topic struct and dynamic lookup
  data/techgen.go          Procedural tech tree generator (DAG, tiers, cost variance)
  data/procedural.go       Procedural alien species generation
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

### Game Conventions (faithful to original X-COM)
- Time Units (TU) for all actions in Battlescape
- Geoscape runs in real-time with pause (time compression)
- Research requires scientists + time
- Manufacturing requires engineers + time + materials
- Soldiers gain stats from combat experience
- Line-of-sight uses Bresenham raycasting

### Key Bindings (Geoscape)
| Key | Action |
|-----|--------|
| Space | Pause/unpause time |
| 1-4 | Time compression |
| B | Open base |
| L | Launch interceptor |
| M | Open manufacture |
| R | Open research |
| Esc | Quit |

### Key Bindings (Battlescape)
| Key | Action |
|-----|--------|
| Arrow keys / hjkl | Move cursor / move unit |
| Space | Select/confirm |
| F | Fire weapon |
| R | Reload |
| E | End turn |
| G | Grenade |
| M | Medikit |
| C | Crouch |
| V | Toggle vision mode |
| Esc | Cancel |
| ? | Help |
