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
- `internal/data` — 100%
- `internal/soldier` — 78%
- `internal/geo` — 43%
- `internal/battle` — 28%
- `internal/base` — 15%

### Dependencies
- `github.com/gdamore/tcell/v2` — Terminal rendering, input, colors

### Architecture
```
cmd/ycom/main.go          Entry point
internal/
  engine/game.go           Game state machine, main loop, input dispatch
  engine/screen.go         Low-level screen/cell rendering helpers
  geo/geoscape.go          Geoscape: world map, time, interceptions
  geo/world.go             World map data (equirectangular ASCII)
  geo/ufo.go               UFO spawning, movement
  geo/interceptor.go       Interceptor launch, dogfight
  battle/battlescape.go    Battlescape: turn logic, TU, line-of-sight
  battle/map.go            Tactical map generation (crash sites, terror)
  battle/unit.go           Soldiers and aliens on the tactical map
  battle/ai.go             Alien AI (patrol, seek, attack, flee)
  base/base.go             Base management screen
  base/facility.go         Facility types and construction
  soldier/soldier.go       Soldier stats, ranking, inventory
  data/items.go            Weapons, armor, items
  data/aliens.go           Alien species, stat blocks
  data/research.go         Research tree
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
| Enter | Open door |
| T | Take item |
| Esc | Cancel |
| ? | Help |
