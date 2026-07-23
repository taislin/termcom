# termcom Development Manual

Build, test, and development reference for the termcom codebase.

## Prerequisites

- Go 1.25+
- Terminal with true-color support (for VFX, lighting effects)

## Building & Running

```bash
# Run the game
go run ./cmd/termcom

# Build binary
go build -o termcom ./cmd/termcom

# Or via Makefile
make build
make run
make test          # Run all tests (verbose)
make test-cover    # Run tests with coverage report
make lint          # Run go vet + staticcheck
make clean         # Remove binary and coverage
```

### WASM Build (Browser Version)

```bash
# Build WASM binary
cd cmd/termcom_wasm
GOOS=js GOARCH=wasm go build -o ../../web_wasm/termcom.wasm .

# Or use build script
./scripts/build_wasm.sh    # Linux/macOS
.\scripts\build_wasm.ps1   # Windows

# Test locally
cd web_wasm
python -m http.server 8080
# Open http://localhost:8080
```

**Architecture:**
- `cmd/termcom_wasm/main.go` — WASM entry point, registers screens, sets `datafs.Set(embeddedFS())`
- `cmd/termcom_wasm/embed.go` — `//go:embed wasmdata` embeds data files
- `internal/datafs/datafs.go` — Unified FS abstraction (`ReadFile`, `ReadDir`, `Glob`)
- `internal/engine/screen_wasm.go` — tcell.Screen implementation for WASM (binary framebuffer)
- `internal/engine/wasm_bridge.go` — Go↔JS bridge (SetScreen, OnKey, termcomGetFrame)
- `web_wasm/index.html` — Canvas renderer with differential rendering

**Data embedding:** Data files (`data/`, `maps/`) are copied to `cmd/termcom_wasm/wasmdata/` at build time, embedded via `//go:embed`, and served as `fs.FS` through `internal/datafs`. All data loaders use `datafs` first, fall back to `os` for `..` paths.

**Font configuration:** Canvas auto-sizes font based on cell dimensions. Override with `?font=FontName` URL parameter.

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
go run ./cmd/termcom_battle building_assault
```

**Available types:** `crash_site`, `terror`, `supply_raid`, `alien_base`, `alien_research`, `council`, `cydonia`, `abduction`, `forest`, `desert`, `polar`, `building_assault`

**What it does:**
- Creates a Game instance with procedural alien species
- Creates a base with full facilities
- Spawns 6 Sergeants with rifle + personal armour
- Launches the selected map type
- Drops straight into player turn

### Map Viewer (`cmd/test_map`)

Generates a map from any generator and renders it on the terminal with full colour, correct tile characters, and multi-level support.

```bash
# List available generators
go run ./cmd/test_map --list

# Generate and view a specific map type
go run ./cmd/test_map crash               # Crash site
go run ./cmd/test_map ufo_wfc 60 60       # WFC UFO interior (60×60)
go run ./cmd/test_map building2           # 2-level urban building (WFC)
go run ./cmd/test_map terror 50 50 42     # Terror site with seed 42
go run ./cmd/test_map all                 # All-biome assembled map
```

**Available map types:** `crash`, `terror`, `abduction`, `ufo_interior`, `ufo_wfc`, `alien_base`, `alien_base_wfc`, `building`, `building2`, `cydonia`, `forest`, `desert`, `polar`, `all`

**Options:**
| Flag | Description |
|------|-------------|
| `--list` / `-l` | List available generators |
| `--dump` / `-d` | Print map as ASCII to stdout (non-interactive) |

**Controls:**
| Key | Action |
|-----|--------|
| `q` | Quit |
| Tab / Shift+Tab | Next / previous level (multi-level maps) |
| Arrows | Scroll when map exceeds terminal size |

**What it does:**
- Creates a tcell screen and renders the selected map using the same `RenderTile` / `TileChar` / `tilePalette` pipeline as the game
- Computes proper 3×3 tile context for geometry-aware glyphs (UFO hull corners, building corners, etc.)
- Supports multi-level maps with keyboard level switching
- Accepts optional width, height, and seed arguments for reproducible viewing

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

**Available map generators:** `crash_site`, `terror`, `supply_raid`/`ufo_interior`, `alien_base`, `council`, `cydonia`, `abduction`, `forest`, `desert`, `polar`, `building_assault`

**Available weapon IDs:** `pistol`, `rifle`, `heavy`, `auto`, `rocket`, `laser_pistol`, `laser_rifle`, `stun_rod`, `plasma_pistol`, `plasma_rifle`, `heavy_plasma`, `alien_blaster`, `alien_cannon`, `alien_laser`, `alien_heavy_laser`, `alien_grenade`, `alien_rocket`, `alien_psi_bolt`, `chryssalid_claw`, `reaper_claw`, `alien_claw`, `alien_fang`

**Available armor IDs:** `none`, `personal`, `light`, `medium`, `heavy`, `power_suit`, `flight_suit`

## Map Types

| Type | Generator | Description |
|------|-----------|-------------|
| `crash_site` | `GenerateCrashSite` | Standard UFO crash recovery |
| `terror` | `GenerateTerrorSite` | Urban terror mission with civilians |
| `supply_raid` | `GenerateUFOInterior` (legacy) or `GenerateUFOInteriorWFC` (live) | Interior UFO combat |
| `alien_base` | `GenerateAlienBaseWFC` | Alien base assault (WFC) |
| `alien_research` | `GenerateUFOInterior` (legacy) or `GenerateUFOInteriorWFC` (live) | Research facility interior |
| `building_assault` | `GenerateUrbanBuildingWFC` | Procedural urban building (WFC) |
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
  termcom_wasm/         WASM entry point (browser, no backend)
    main.go             Registers screens, sets up datafs
    embed.go            //go:embed wasmdata (embedded data files)
  test_aliens/          Alien roster viewer (console output)
  test_map/             Map visualiser (tcell render of any generator)
  scenario_creator/     Interactive scenario JSON creator
  wfcdiag/              WFC tile library diagnostic tool
  webserver/            Web server (for remote play)
maps/
  *.json                Custom battle definitions
web_wasm/
  index.html            Canvas renderer for WASM version
  termcom.wasm          Built WASM binary (gitignored)
tools/
  map_editor.html       Browser-based full map builder (chunk stamping + tile painting)
  area_editor.html      Browser-based chunk editor (small tile-grid fragments)
  mission_editor.html   Browser-based mission editor (place units, set victory conditions)
  tile_creator.html     Browser-based tile type creator (custom tiles)
  bundle_tiles.js       Build script: bundles tools/tile_data/*.jsonc → tile_data.js
  tile_data.js          Generated tile library used by HTML tools
  tile_data/            Source JSONC files (categorized, with comments)
    terrain.jsonc       Floor, wall, ground, pavement, sand, snow, etc.
    nature.jsonc        Trees, rocks, bushes, wheat, vines, etc.
    water.jsonc         Water, marsh, swamp, ice, pier
    furniture.jsonc     Desks, chairs, beds, lockers, consoles, etc.
    vehicles.jsonc      Cars, helos, tractors, buses, crawlers, wheels
    alien.jsonc         UFO floors, pods, power sources, alien tech
    urban.jsonc         Adobe, containers, rubble, streetlamps, timber
    hazards.jsonc       Glass, debris
data/
  maps/
    *.json              Map fragment library (biome-tagged terrain chunks)
  tiles/
    *.jsonc             Custom tile definitions (dynamically loaded at startup)
  wfc/
    ufo.json            WFC tile library: UFO interiors
    urban.json          WFC tile library: urban buildings
    alien_base.json     WFC tile library: alien base interiors
  aliens/
    *.json              Procedural alien sprite part definitions
internal/
  datafs/
    datafs.go           Unified FS abstraction (os.DirFS native, embedded WASM)
  engine/
    game.go             Game loop, state machine
    screen.go           Rendering primitives (DrawPanel, DrawString, etc.)
    screen_wasm.go      WASM tcell.Screen implementation (binary framebuffer)
    wasm_bridge.go      Go↔JS bridge (SetScreen, OnKey, termcomGetFrame)
    custom_battle.go    Custom battle selection screen
    portrait.go         Soldier/alien portrait rendering
    ...                 (VFX, particles, camera, menu, help, options, etc.)
  battle/
    battlescape.go      Battlescape: turns, units, AI, victory conditions
    map.go              Tactical map generators (10+ biomes)
    wfc.go              Wave Function Collapse solver (Tiled Model)
    fragments.go        Map fragment stamping, chunk placement, AssembleMap
    unit.go             Unit creation (soldier, alien, civilian)
    ai.go               Alien AI behavior
    gas.go              Volumetric smoke/gas
    cluster.go          Clustered terrain (blob growth, poisson sampling)
    tiledef.go          Tile definitions (lazy-loads JSONC, datafs-first)
    ...                 (input, movement, LOS, etc.)
  mapgen/
    mapgen.go           CDDA-style mapgen chunk loader (JSON fragments)
    wfctile.go          WFC tile library loader (JSON with adjacency rules)
    alien_templates.go  Alien template sprite loader
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
scripts/
  build_wasm.sh         WASM build script (Linux/macOS)
  build_wasm.ps1        WASM build script (Windows)
```

## Architecture Notes

- **Rendering:** All graphics via tcell; true-color RGB for lighting, VFX, particles
- **Coordinate system:** (x, y) where x=col, y=row
- **Map tiles:** rune-based with tcell.Style coloring
- **AI:** Behavior tree pattern with patrol, seek, attack, flee, flank, retreat
- **Save system:** JSON-based with version migration (current: v3)
- **Audio:** MIDI synthesis on Windows, oto-based PCM synthesis on Linux/macOS
- **Map generation:** Two complementary systems — `AssembleMap` for outdoor/biome maps (scatter fragments from `data/maps/*.json`) and WFC solver for enclosed interiors (tile library from `data/wfc/*.json`)
- **All map generators are deterministic** — seeded RNG ensures reproducible layouts (critical for replay and save verification)

## Map Generation Systems

### 1. Fragment-based assembly (`AssembleMap`)

Used for outdoor & mixed maps: terror, abduction, crash site terrain, forest, desert, polar.

- Reads biome-tagged fragments from `data/maps/*.json`
- Each fragment defines an ASCII tile grid with terrain/furniture glyph mappings
- `AssembleMap(biome, w, h, rng)` fills base terrain, applies clustered terrain (blob growth, poisson sampling), stamps a random anchor fragment, then greedily places additional fragments with spacing + connectivity (flood-fill) checks, stamps corridors between fragment doors
- Biomes: `urban`, `forest`, `desert`, `polar`, `rural`, `ufo`, `alien`
- Generator wrappers: `GenerateCrashSite`, `GenerateTerrorSite`, `GenerateAbductionSite`, `GenerateForest`, `GenerateDesert`, `GeneratePolar`

**Adding a new fragment:**

1. Create a `.json` file in `data/maps/` with fields: `id`, `tags`, `width`, `height`, `rows` (ASCII art), `terrain` (glyph → TileType name mapping), `furniture` (optional)
2. Tag it with the biome(s) you want it to appear in (e.g. `["urban", "forest"]`)
3. Tile glyphs: `W`=TileWall, `.`=TileFloor, `g`=TileGrass, `t`=TileTree, `r`=TileRock, `s`=TileSand, `n`=TileSnow, `~`=TileWater, `R`=TileRubble, etc. Furniture glyphs per-chunk (defined in furniture map)
4. Fragments load automatically via `mapgen.Init()` at game start

**Current fragment count:** ~84 fragments across all biomes.

### 2. Wave Function Collapse solver (`wfc.go`)

Used for enclosed interiors: UFO hulls and urban building interiors.

- **Tiled Model** — defines a tile library of small 3x3 pieces and larger blocks (6x6, 9x9)
- **Superposition grid** starts with all tiles possible in every cell
- **Observation** picks the lowest-entropy cell and randomly collapses it
- **Propagation** eliminates incompatible tile options from neighbors via queue-based constraint propagation (AC-3 variant)
- **Restart-on-contradiction** retries with a fresh wave if a cell reaches 0 options
- Output compiles to `BattleMap` by stamping each tile's rune grid into the terrain grid

**Tile library format** (`data/wfc/*.json`):

```json
{
  "tiles": [
    {
      "id": 0,
      "name": "Floor",
      "rows": ["...", "...", "..."],
      "neighbors": {
        "N": [0,1,2,3,4,5,6,7,8,9],
        "E": [0,1,2,3,4,5,6,7,8,9],
        "S": [0,1,2,3,4,5,6,7,8,9],
        "W": [0,1,2,3,4,5,6,7,8,9]
      }
    }
  ]
}
```

- `rows` are equal-length strings; characters map to `TileType` via `tileRuneToType`
- `neighbors.N/E/S/W` list the tile IDs allowed to sit in each cardinal direction relative to this tile
- Tile size is variable — 3x3 small pieces and larger multi-room blocks share the grid
- Libraries load at runtime from `data/wfc/ufo.json`, `data/wfc/urban.json`, and `data/wfc/alien_base.json`, with hardcoded fallback

**Available WFC generators:**

| Function | Library | Mission |
|----------|---------|---------|
| `GenerateUFOInteriorWFC` | `data/wfc/ufo.json` (17 tiles) | `Supply Raid`, `Alien Research` |
| `GenerateUrbanBuildingWFC` | `data/wfc/urban.json` (20 tiles) | `Building Assault` |
| `GenerateAlienBaseWFC` | `data/wfc/alien_base.json` | `Alien Base Assault` |

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

### Adding a new map generator

1. For **fragment-based** maps: add fragment JSONs to `data/maps/` with the desired biome tag, or write a new generator function calling `AssembleMap` with a new biome
2. For **WFC-based** maps: add tiles to `data/wfc/<name>.json` with adjacency rules, write a generator function using `NewWFCRules` + `newWave` + `Solve` + `CompileToBattleMap`
3. Add case to `NewBattlescape` switch in `internal/battle/battlescape.go`
4. Add entry to `cmd/termcom_battle/main.go` for testing

### Adding a new item

1. Add struct to `internal/data/items.go` with all fields
2. Add language strings to `internal/language/en.go`
3. Add to base stores if purchasable

### Adding a new mission type

1. Add language keys `MISSION_<NAME>` and `MISSION_TYPE_<NAME>` to **all 8 language files** (`en.go`, `zh.go`, `es.go`, `fr.go`, `ru.go`, `pt.go`, `ja.go`, `ko.go`)
2. Add to geoscape mission pool (`internal/geo/geoscape.go` — `rollMission` weighted pool)
3. Add to `ufoName` switch in `respondToMission` mapping the type to its display key
4. Add map generator case in `NewBattlescape` (`internal/battle/battlescape.go`)
5. Add command-line alias in `cmd/termcom_battle/main.go`

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

**Maps** (`internal/battle/map.go`, `internal/battle/fragments.go`, `internal/battle/wfc.go`):
- Fragment-based maps (12+ biome wrappers)
- WFC-based interior maps (3 libraries, single-level and multi-level)
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

## WFC Tile Library: Adding New Tiles

Each tile in `data/wfc/*.json` needs:

1. **Unique `id`** (0..N-1, no gaps or duplicates)
2. **`name`** — descriptive string
3. **`rows`** — N strings of equal length (the tile footprint)
4. **`neighbors`** — per-direction (`"N"`, `"E"`, `"S"`, `"W"`) lists of tile IDs that may touch this tile in that direction

Rune meaning (shared by all WFC libraries):

| Glyph | TileType | Description |
|-------|----------|-------------|
| `.` | UFOFloor / Floor | Walkable floor |
| `#` | UFOWall / Wall | Solid wall |
| `D` | Door | Passable door |
| `C` | Console / Computer | Furniture |
| `M` | Machinery | Industrial equipment |
| `P` | Pod | Alien pod |
| `X` | PowerSource | Power core |
| `S` | Storage | Storage unit |
| `B` | Bed | Bunk/bed |
| `A` | AlienTech | Alien technology |

**Neighbor rule tips:**
- A floor tile (`.` on all edges) should list all interior tiles as neighbors in all 4 directions
- Wall tiles (`#` on one side) should list structural tiles as neighbors on the solid side and open/interior tiles on the open side
- Multi-room blocks (walled on all sides) should list only structural tiles as neighbors, so they close the building envelope

## Save Versioning

Current version: **3**

| Version | Changes |
|---------|---------|
| 1 | Initial save format |
| 2 | WeaponAmmo field added |
| 3 | AlienKnowledge map added |

Migration functions are in `internal/save/save.go`. Saves below v2 are rejected.

## Tile Library System

All 87 built-in tile definitions live in `tools/tile_data/*.jsonc` — categorized JSONC files with comments showing each glyph's character next to its Unicode code point. These files are the single source of truth for tile data across both the HTML editor tools and the Go game engine.

| File | Category | Tiles |
|------|----------|-------|
| `terrain.jsonc` | Floors, walls, ground, pavement, sand, snow, scree, mud, stairs, cliff | 13 |
| `nature.jsonc` | Trees, rocks, bushes, wheat, vines, bamboo, hay, boulders | 11 |
| `water.jsonc` | Water, marsh, swamp water, ice, pier | 5 |
| `furniture.jsonc` | Desks, chairs, beds, lockers, cabinets, consoles, machinery, storage | 12 |
| `vehicles.jsonc` | Cars, helos, tractors, forklifts, buses, trucks, crawlers, wheels | 24 |
| `alien.jsonc` | UFO floors/walls, pods, power sources, alien tech, wreckage, dish | 10 |
| `urban.jsonc` | Adobe, metal walls, containers, rubble, fences, streetlamps, timber | 10 |
| `hazards.jsonc` | Glass, debris | 2 |

**Bundle script:** `tools/bundle_tiles.js` reads all `*.jsonc` files and generates `tools/tile_data.js` (a JS file with the combined `TILE_DATA` array). The HTML tools include this script. Re-run after editing any JSONC file:
```bash
node tools/bundle_tiles.js
```

**Go loading:** The game engine reads the same JSONC files at startup via `stripJSONCComments()` in `internal/battle/tiledef.go`. When the JSONC files aren't found (e.g., deployed binary), the game falls back to hardcoded values.

## Developer Tools

Browser-based HTML tools for level design and tile creation. Open directly in any browser — no build step required. Tile definitions are loaded from `tools/tile_data.js` (generated from `tools/tile_data/*.jsonc`).

### Area Editor (`tools/area_editor.html`)

Visual tile-grid editor for designing map fragments (chunks). Paint tiles on a grid, export to chunk JSON for use in `data/maps/`.

- **Left sidebar:** tile palette with all tile types, glyph previews, and color swatches
- **Center:** interactive grid — left-click to paint, right-click to erase (sets to TileGrass)
- **Terminal Preview:** click "Preview" to open a side panel showing how the map looks in a monospace terminal (correct glyph aspect ratio, game colors)
- **Export:** generates a chunk JSON file. Save the exported JSON to `data/maps/<name>.json` and tag it with the appropriate biome(s). The chunk will be picked up automatically by `mapgen.Init()` at startup.
- **Load:** import existing chunk JSONs from `data/maps/` to edit them

**Exported file location:** `data/maps/<name>.json`

### Map Editor (`tools/map_editor.html`)

Full map builder — design complete tactical maps by stamping chunks or painting individual tiles on a large grid.

- **Left sidebar:** tile palette (same tiles as Area Editor)
- **Center:** large grid (up to 100x100) — left-click to paint tile / stamp chunk, right-click to erase
- **Right panel: Chunk Library** — load any number of chunk JSONs from `data/maps/` via multi-file picker. Each loaded chunk can be selected and stamped onto the map with a single click. Active chunks are listed with mini previews.
- **Mode toggle:** switch between Tile Mode (paint individual tiles) and Chunk Mode (stamp entire chunks)
- **Export:** exports the full map as a JSON file with width, height, and a 2D tile grid array
- **Load Map:** re-import previously exported map JSON for editing

**Exported file location:** standalone map JSON (loadable by the editor). To use in-game, wrap in a custom battle definition or add a loader in `NewBattlescape`.

**Usage:**
1. Set W/H for the overall map (default 50×50)
2. Load chunk JSONs from `data/maps/` using the Chunk Library panel
3. Switch to Chunk Mode, click a chunk in the library, then click on the map to stamp it
4. Switch to Tile Mode to paint individual tiles or touch up chunk edges
5. Export the finished map to JSON

### Tile Creator (`tools/tile_creator.html`)

Create new tile types with custom glyphs, colors, and gameplay properties. Export to JSON for use as dynamic tiles loaded at startup.

- **Left sidebar:** tile properties — ID, name key, glyph, color picker, passable/opaque/destructible/flammable/explodes/noisy toggles, cover %, TU cost, move cost, category, description
- **Center:** live preview — large tile render + terminal-style preview
- **Registered Tiles:** list of all tiles you've added in this session. Click to re-edit, X to remove.
- **Export:** generates a `custom_tiles.json` file. Save to `data/tiles/<name>.json` and the game loads it automatically on startup via `InitCustomTiles()`. You can have multiple files in `data/tiles/` — all `.json` files are loaded.
- **Load:** import existing tile JSONs from `data/tiles/` to continue editing

**Exported file location:** `data/tiles/<name>.json` (any filename, must end in `.json`)

**Custom tile JSON format:**

```json
{
  "tiles": [
    {
      "id": "TileNeonSign",
      "glyph": "n",
      "color": "#00ff88",
      "cover": 0,
      "passable": true,
      "opaque": false,
      "destructible": true,
      "flammable": false,
      "explodes": 0,
      "noisy": false,
      "move_cost": 4,
      "name": "TILE_NEON_SIGN"
    }
  ]
}
```

Custom tiles are assigned dynamic `TileType` IDs starting at 1000. They work in:
- Map fragment JSONs (add `"TileNeonSign": "n"` to the `terrain` map)
- WFC tile libraries
- Fragment assembly
- All gameplay systems (LOS, pathfinding, cover, fire, explosions)

**TU cost reference:** 3=pavement, 4=floor, 5=snow/bush/wheat, 6=sand/mud/scree, 8=water/forest
