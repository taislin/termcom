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

### Alien Roster Viewer (`cmd/test_aliens`)

Generates the full procedural alien roster and prints each alien to the console with colored portraits, stats, and morphology info.

```bash
# Default seed (42)
go run ./cmd/test_aliens

# Specific seed
go run ./cmd/test_aliens 12345
```

**What it does:**
- Generates 5-7 procedural species (10-28 alien types) via `data.GenerateSpecies(seed)`
- Renders each alien's 7x6 ASCII portrait using half-block characters with true-color ANSI RGB
- Displays all stats (HP, TU, Accuracy, Bravery, Reactions, Strength, Psi, Armour, Aggression)
- Shows resistances (green = resist, red = weak, gray = neutral)
- Lists morphology details (body type, limbs, senses) and lore
- Ends with a summary table of all aliens

### Custom Battles

Create custom battle scenarios by placing JSON files in the `maps/` folder. These appear in both the main menu ("Custom Battle") and the battle test tool.

```bash
# Via main menu
go run ./cmd/termcom
# Select "Custom Battle"

# Via battle test tool (interactive menu or direct)
go run ./cmd/termcom_battle
# Select from the list (custom maps marked with [custom])
```

**Template files in `maps/`:**

| File | Description |
|------|-------------|
| `crash_site_ambush.json` | Crash site with 4 soldiers vs 5 aliens + civilians. Eliminate all. |
| `hold_the_line.json` | Forest defense, night. Survive 10 turns against 7 aliens. |
| `extraction_point.json` | Desert extraction. Reach the exit zone in the southeast. |

**JSON schema:**

```json
{
  "name": "Mission Name",
  "author": "Author",
  "date": "2026-07-11",
  "description": "Brief description shown in the menu.",
  "night": false,
  "map": {
    "type": "generated",
    "generator": "crash_site",
    "width": 50,
    "height": 50
  },
  "soldiers": [
    {
      "name": "Cpl. Alpha", "rank": 2,
      "hp": 28, "tu": 52, "accuracy": 72, "reactions": 60, "strength": 18,
      "weapon": "rifle", "armor": "personal",
      "x": 5, "y": 25
    }
  ],
  "aliens": [
    {
      "name": "Sectoid", "hp": 10, "tu": 50, "accuracy": 55,
      "bravery": 40, "reactions": 50, "strength": 8, "psi": 40, "armour": 5,
      "weapon": "plasma_pistol", "rank": 0, "damage_type": 0, "aggression": 3,
      "x": 25, "y": 8
    }
  ],
  "civilians": [
    { "name": "Survivor", "x": 18, "y": 35 }
  ],
  "victory": {
    "condition": "eliminate_all"
  }
}
```

**Victory conditions:**

| Condition | Fields | Description |
|-----------|--------|-------------|
| `eliminate_all` | _(none)_ | Kill all aliens to win |
| `survive_turns` | `turns` | Survive N turns without squad wipe |
| `reach_point` | `target_x`, `target_y`, `min_soldiers` | Get N soldiers to the target tile |

**Available map generators:** `crash_site`, `terror`, `supply_raid`/`ufo_interior`, `alien_base`, `council`, `cydonia`, `abduction`, `forest`, `desert`, `polar`

**Available weapon IDs:** `pistol`, `rifle`, `heavy`, `auto`, `rocket`, `laser_pistol`, `laser_rifle`, `stun_rod`, `plasma_pistol`, `plasma_rifle`, `heavy_plasma`, `alien_blaster`, `alien_cannon`, `alien_laser`, `alien_heavy_laser`, `alien_grenade`, `alien_rocket`, `alien_psi_bolt`, `chryssalid_claw`, `reaper_claw`, `alien_claw`, `alien_fang`

**Available armor IDs:** `none`, `personal`, `light`, `medium`, `heavy`, `power_suit`, `flight_suit`

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
  termcom_battle/       Test script: interactive battle launcher
  test_aliens/          Alien roster viewer (console output)
  webserver/            Web server (for remote play)
maps/
  *.json                Custom battle definitions
internal/
  engine/
    game.go             Game loop, state machine
    screen.go           Rendering primitives (DrawPanel, DrawString, etc.)
    custom_battle.go    Custom battle selection screen
    portrait.go         Soldier/alien portrait rendering
    ...                 (VFX, particles, camera, menu, help, options, etc.)
  battle/
    battlescape.go      Battlescape: turns, units, AI, victory conditions
    map.go              Tactical map generators (10+ biomes)
    unit.go             Unit creation (soldier, alien, civilian)
    ai.go               Alien AI behavior
    gas.go              Volumetric smoke/gas
    ...                 (input, movement, LOS, etc.)
  geo/                  Geoscape: world map, UFOs, interceptors, missions
  base/                 Base management: facilities, research, manufacture
  soldier/              Soldier stats, ranking, inventory
  data/
    items.go            Weapons, armor, items (RuleItems map)
    aliens.go           Alien species, portraits, morphology types
    procedural.go       Procedural alien species + portraits per run
    procedural_items.go Procedural weapons and armor generation
    techgen.go          Procedural tech tree generation
    research.go         Research topic definitions
  save/                 Save/load system, version migration
  language/             Localization strings
  audio/                Platform-specific audio synthesis
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

### Procedural systems

Each playthrough generates unique content from a seed:

**Alien species** (`internal/data/procedural.go`):
- 5-7 species per run with unique names, morphology, stats, resistances
- Rank variants (0-4) with scaled stats
- ASCII portraits driven by morphology
- Called via `GenerateSpecies(seed)`

**Research tree** (`internal/data/techgen.go`):
- 16 base techs + autopsies per species
- DAG prerequisites randomized each run
- Species study topics unlocked after autopsies
- Called via `GenerateTechTree(seed, aliens)`

**Items** (`internal/data/procedural_items.go`):
- 2-3 procedural weapons based on species damage types
- 1-2 procedural armor pieces based on species
- Registered via `RegisterProceduralItems(seed, aliens)`
- Called during game init and save load

**Maps** (`internal/battle/map.go`):
- 10 procedural map generators
- Biome-based tile probability distributions
- Seeded for reproducible layouts

### Mission modifiers

Random modifiers are rolled per mission in `internal/battle/modifiers.go`:
- `RollModifiers(rng, missionType)` returns a slice of modifiers
- Applied in `NewBattlescape` — affects alien count, visibility, win conditions
- `ModReinforcements` spawns extra aliens on turn 4
- `ModTimeLimit` causes defeat if battle exceeds 15 turns
- `ModNightOps` forces night battle
- `ModHeavyFog` reduces sight range

### Weather system

Weather is generated per mission in `internal/battle/modifiers.go`:
- `RollWeather(rng, biome)` returns a `Weather` struct
- Affects accuracy penalties, sight range, fire spread chance
- Applied in `ComputeFOVForTeam` (sight) and `FireAt` (accuracy)

### Soldier perks

Perks are defined in `internal/soldier/perks.go`:
- 12 perks with stat bonuses and battle modifiers
- Random perk awarded on each rank-up via `GainXP`
- Perks saved/loaded via `Perks []string` field
- Battle effects checked via `HasPerk` / `HasBattleMod`

### Fatigue system

- Soldiers gain 1-5 days fatigue after surviving a battle
- Fatigue heals 1 per day alongside wounds
- `CanDeploy()` checks HP > 0, Wounds == 0, Fatigue == 0
- `HealthySoldiers()` uses `CanDeploy()` for deployment lists

### Mission auto-resolve

Tactical battles can be auto-resolved from the geoscape via `AutoresolveMission()`:

- Player presses `M` → mission select overlay appears with odds calculation
- Win chance: `30 + (squadPower - alienPower) / 5`, capped at 10-70%
- Squad power: HP + Accuracy/2 + Strength + Reactions/2 + perk bonuses
- Alien power: `alienCount * (40 + missionsWon*3) * difficultyScale`
- Mission type modifiers: Terror -10%, Council +10%, Alien Base -15%

**Rewards (vs tactical):**
- XP: 50% of tactical
- Corpses: None
- Weapon drops: 25% chance per alien (vs 15-55% tactical)
- Alloys/elerium: Full
- Fatigue: 2-3 days (vs 1-5 tactical)

**Casualties:**
- Win: 33% chance of 1 soldier wounded
- Loss: 1-3 soldiers killed (permanent death)

### Modifying balance

- Soldier stats: `internal/soldier/soldier.go` (NewSoldier defaults)
- Alien stats: `internal/data/aliens.go` (AlienTypes array for hardcoded aliens)
- Morphology stat modifiers: `internal/data/procedural.go` (`morphXMod` functions)
- Body subtype resistances: `internal/data/procedural.go` (`subtypeResistMod`)
- Weapon damage/accuracy/TU: `internal/data/items.go` (RuleItems map)
- Difficulty scaling: `internal/engine/difficulty.go`
- Mission modifiers: `internal/battle/modifiers.go` (RollModifiers)
- Weather: `internal/battle/modifiers.go` (RollWeather)
- Perks: `internal/soldier/perks.go` (AllPerks, RollPerk)

## Save Versioning

Current version: **3**

| Version | Changes |
|---------|---------|
| 1 | Initial save format |
| 2 | WeaponAmmo field added |
| 3 | AlienKnowledge map added |

Migration functions are in `internal/save/save.go`. Saves below v2 are rejected.
