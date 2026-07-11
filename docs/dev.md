# termcom Development Manual

Build, test, and development reference for the termcom codebase.

## Prerequisites

- Go 1.21+
- Terminal with true-color support (for VFX, lighting effects)

## Building & Running

```bash
# Run the game
go run ./cmd/termcom

# Build binary
go build -o termcom.exe ./cmd/termcom

# Or via Makefile
make build
make run
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Verbose output
go test -v ./...

# Run specific package tests
go test ./internal/battle/...
go test ./internal/geo/...
go test ./internal/data/...

# With race detector
go test -race ./...
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
go vet ./...
```

## Test Scripts

Quick-launch scripts for testing specific game states without going through the main menu.

### Battle Test (`cmd/termcom_battle`)

Launches directly into a battlescape with a 6-soldier squad.

```bash
# Random battle type
go run ./cmd/termcom_battle

# Specific battle type
go run ./cmd/termcom_battle crash_site
go run ./cmd/termcom_battle terror
go run ./cmd/termcom_battle cydonia
go run ./cmd/termcom_battle forest
go run ./cmd/termcom_battle desert
go run ./cmd/termcom_battle polar
go run ./cmd/termcom_battle supply_raid
go run ./cmd/termcom_battle alien_base
go run ./cmd/termcom_battle alien_research
go run ./cmd/termcom_battle abduction
go run ./cmd/termcom_battle council
```

**Available types:** `crash_site`, `terror`, `supply_raid`, `alien_base`, `alien_research`, `council`, `cydonia`, `abduction`, `forest`, `desert`, `polar`

**What it does:**
- Creates a Game instance with procedural alien species
- Creates a base with full facilities
- Spawns 6 Sergeants with rifle + personal armour
- Launches the selected map type
- Drops straight into player turn

## Map Types

| Type | Generator | Description |
|------|-----------|-------------|
| `crash_site` | `GenerateCrashSite` | Standard UFO crash recovery |
| `terror` | `GenerateTerrorSite` | Urban terror mission with civilians |
| `supply_raid` | `GenerateUFOInterior` | Interior UFO combat |
| `alien_base` | `GenerateAlienBase` | Alien base assault |
| `alien_research` | `GenerateUFOInterior` | Research facility interior |
| `council` | `GenerateTerrorSite` | Council mission |
| `cydonia` | `GenerateCydonia` | Final mission on Mars |
| `abduction` | `GenerateAbductionSite` | Abduction with timer |
| `forest` | `GenerateForest` | Wooded terrain |
| `desert` | `GenerateDesert` | Desert terrain |
| `polar` | `GeneratePolar` | Snow terrain |

## Project Structure

```
cmd/
  termcom/              Main game entry point
  termcom_battle/       Test script: direct battle launch
  webserver/         Web server (for remote play)
internal/
  engine/            Core engine: game loop, rendering, VFX, particles, camera
  battle/            Battlescape: maps, units, AI, turns, line-of-sight
  geo/               Geoscape: world map, UFOs, interceptors, missions
  base/              Base management: facilities, research, manufacture
  soldier/           Soldier stats, ranking, inventory
  data/              Game data: items, aliens, research, tech tree
  save/              Save/load system, version migration
  language/          Localization strings
  audio/             Platform-specific audio synthesis
docs/
  manual.md          Player manual
  dev.md             This file
  tables.md          Data tables reference
```

## Architecture Notes

- **Rendering:** All graphics via tcell; true-color RGB for lighting, VFX, particles
- **Coordinate system:** (x, y) where x=col, y=row
- **Map tiles:** rune-based with tcell.Style coloring
- **AI:** Behavior tree pattern with patrol, seek, attack, flee, flank, retreat
- **Save system:** JSON-based with version migration (current: v3)
- **Audio:** oto-based PCM synthesis on Windows, terminal BEL fallback on Linux/macOS

## Common Development Tasks

### Adding a new alien type

**Hardcoded aliens** (in `var AlienTypes` in `internal/data/aliens.go`):
1. Add struct to `internal/data/aliens.go`
2. Add weapon to `internal/data/items.go` if needed
3. Add lore text to `internal/language/en.go`
4. Add resistances appropriate to the damage type
5. Portrait defaults to carbon_flesh morphology (set `Morphology: nil`)

**Procedural aliens** (generated from seed in `internal/data/procedural.go`):
- Morphology is auto-generated from damage type and random rolls
- To add a new body subtype: add constant to `aliens.go`, add entry to
  `pickOrganicSubtype`/`pickSyntheticSubtype`, add head shape to `headShape()`,
  add resistance modifiers to `subtypeResistMod()`, add stat modifiers to the
  `morphXMod()` functions, and add lore snippet to `morphLoreSnippets`
- To add a new sense: add constant to `aliens.go`, update `pickSenseQuality`/
  `pickHearingQuality`/`pickBinarySense`, add effect to `canSense()` in `ai.go`,
  add targeting bonus to `selectTarget()` in `ai.go`, and add portrait decoration
  to `pickSenseSensor()`

### Adding a new map type

1. Add generator function to `internal/battle/map.go`
2. Add case to `NewBattlescape` switch in `internal/battle/battlescape.go`
3. Add entry to `cmd/test_battle/main.go` for testing

### Adding a new item

1. Add struct to `internal/data/items.go` with all fields
2. Add language strings to `internal/language/en.go`
3. Add to base stores if purchasable

### Modifying balance

- Soldier stats: `internal/soldier/soldier.go` (NewSoldier defaults)
- Alien stats: `internal/data/aliens.go` (AlienTypes array for hardcoded aliens)
- Morphology stat modifiers: `internal/data/procedural.go` (`morphXMod` functions)
- Body subtype resistances: `internal/data/procedural.go` (`subtypeResistMod`)
- Weapon damage/accuracy/TU: `internal/data/items.go` (RuleItems map)
- Difficulty scaling: `internal/engine/difficulty.go`

## Save Versioning

Current version: **3**

| Version | Changes |
|---------|---------|
| 1 | Initial save format |
| 2 | WeaponAmmo field added |
| 3 | AlienKnowledge map added |

Migration functions are in `internal/save/save.go`. Saves below v2 are rejected.
