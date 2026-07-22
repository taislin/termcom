# termcom Tactical Map & Chunk Reference

This guide covers the tactical map system for map and level designers: how chunks work, assembling maps from fragments, tile type reference, and best practices.

## Table of Contents

1. [Map Architecture](#map-architecture)
2. [Fragment (Chunk) System](#fragment-chunk-system)
3. [JSON Fragment Format](#json-fragment-format)
4. [AssembleMap Logic](#assemblemap-logic)
5. [Tile Type Reference](#tile-type-reference)
6. [Biome Reference](#biome-reference)
7. [Examples](#examples)
8. [Best Practices](#best-practices)

---

## Map Architecture

Tactical maps are generated in two phases:

1. **Base terrain** — `AssembleMap` fills an outdoor grid with a biome-appropriate ground tile (grass, sand, snow, pavement, etc.) and stamps building footprints, roads, water, and terrain features using a combination of noise and biome-specific rules.

2. **Fragment stamping** — Pre-authored JSON chunk files are loaded from `data/maps/*.json` and stamped onto the base terrain where they match the biome tags. Fragments provide structured content (buildings, vehicles, UFOs, terrain features) with interior tiles, furniture, and walls.

Each map has a `LevelHeight` and `Width` (both tile counts). A `BattleMap` stores tiles as a 2D grid of `TileType` values.

## Fragment (Chunk) System

Fragments are rectangular pieces of authored map content. Each belongs to one or more biomes via `tags`. The system:

- Groups fragments by biome tag
- Selects a random **anchor** chunk (weighted, min size requirement per biome)
- Stamps additional **scatter** chunks around the map
- Applies 0-3 random quarter-turn rotations unless `no_rotate` is true

Scatter chunks may partially overlap; terrain tiles that are already placed take precedence over new ones. Fragments with `weight: 0` are anchor-only and never scatter.

### Key Behaviours

| Attribute | Effect |
|-----------|--------|
| `tags` | Biomes where this chunk can appear. Multi-tag chunks appear in all matching biomes. |
| `weight` | Relative spawn probability. Higher = more likely. Default 1. |
| `no_rotate` | Prevents random rotation. Use for vehicles, directional objects, text. |
| Anchor chunk | A weighted random selection — must produce a valid placement. |
| Scatter chunks | Randomly placed across the map, may be rejected if they don't fit. |

## JSON Fragment Format

Fragments are JSON files stored in `data/maps/`. Each file contains a single fragment definition.

```json
{
  "id": "unique_chunk_name",
  "tags": ["urban", "rural"],
  "width": 9,
  "height": 7,
  "weight": 2,
  "no_rotate": false,
  "rows": [
    "#######..",
    "#.....#..",
    "#.....#..",
    "#..+..#..",
    "#.....#..",
    "#.....#..",
    "#######.."
  ],
  "terrain": {
    "#": "TileWall",
    ".": "TileFloor",
    "+": "TileDoor"
  },
  "furniture": {}
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier. Used as the filename (without `.json`). Must not collide with other chunk IDs. |
| `tags` | Yes | One or more biome tags from the biome reference below. |
| `width` | Yes | Width in tiles (columns). |
| `height` | Yes | Height in tiles (rows). |
| `weight` | No | Spawning weight. Higher = more likely. Default 1. 0 = anchor-only. |
| `no_rotate` | No | If true, chunk is never rotated when placed. Default false. |
| `rows` | Yes | Array of strings forming the ASCII grid. Each string must be exactly `width` characters. Must have `height` rows. |
| `terrain` | Yes | Maps ASCII glyphs to TileType names for ground/structural tiles. |
| `furniture` | No | Maps ASCII glyphs to TileType names for furniture/objects on top of terrain. If a glyph appears in both `terrain` and `furniture`, `terrain` wins. |

### Glyph Rules

- Use only BMP Unicode (U+0000–U+FFFF). No emoji (U+1F300+).
- Box drawing, technical symbols, and miscellaneous symbols are fine.
- Each glyph in `rows` must have a mapping in either `terrain` or `furniture`.
- Space (` `) is reserved for empty/untouched tiles and is not mapped.

## AssembleMap Logic

The `AssembleMap` function (`internal/battle/map.go`) builds tactical maps through these steps:

1. **Biome selection** based on mission type (crash site, terror, base assault, etc.)
2. **Base fill** — fills the entire grid with the biome's ground tile
3. **Terrain features** — roads, rivers, fences, hedges, fields, craters, ponds, walls, cliffs, and other biome-specific elements are placed procedurally
4. **Anchor stamping** — selects and places anchor fragments
5. **Scatter stamping** — places additional fragments in remaining open areas
6. **Post-processing** — interior floor tiles, furniture placement

### Biome Placement Rules

| Biome | Ground | Features |
|-------|--------|----------|
| `urban` | Pavement | Roads, buildings, streetlamps, walls, fences |
| `forest` | Grass | Trees, bushes, rocks, logging structures |
| `desert` | Sand | Adobe buildings, cacti, rocks, wrecks |
| `polar` | Snow | Ice patches, prefab structures, containers |
| `rural` | Grass | Farmhouses, barns, fences, dirt roads |
| `ufo` | UFOFloor | UFO hull interior (from crash landing) |
| `farm` | Grass | Wheat fields, barns, silos, windmills |
| `coastal` | Sand | Piers, docks, boats, lighthouses |
| `mountain` | Scree | Ridges, caves, boulders, cliff faces |
| `swamp` | Marsh | Cypress trees, huts, sinkholes |
| `jungle` | Mud | Bamboo, vines, temple ruins |

## Tile Type Reference

### Core Terrain

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileFloor` | `.` | yes | no | no | 0 | Interior floor |
| `TileWall` | `#` | no | yes | yes | 80 | Building wall |
| `TileDoor` | `+` | yes | no | yes | 0 | Doorway |
| `TileWindow` | `□` | no | yes | yes | 20 | Building window |
| `TileGrass` | `·` | yes | no | no | 0 | Open ground |
| `TileTree` | `♣` | no | yes | yes | 80 | Tree |
| `TileRock` | `∩` | no | yes | yes | 70 | Rock formation |
| `TileWater` | `≈` | no | no | no | 0 | Deep water (impassable) |
| `TileUFOFloor` | `≡` | yes | no | no | 0 | UFO interior floor |
| `TileUFOWall` | `█` | no | yes | yes | 80 | UFO hull wall |
| `TileStairs` | `▒` | yes | no | no | 0 | Stairs up |
| `TileStairsDown` | `▓` | yes | no | no | 0 | Stairs down |
| `TilePavement` | `░` | yes | no | no | 0 | Road / sidewalk |
| `TileSand` | `·` | yes | no | no | 0 | Sandy ground |
| `TileSnow` | `∗` | yes | no | no | 0 | Snowy ground |
| `TileMarsh` | `≋` | yes | no | no | 0 | Marsh / bog |
| `TileBush` | `†` | yes | no | yes | 40 | Bush / scrub |
| `TileFence` | `│` | no | yes | yes | 40 | Wooden fence |
| `TileRubble` | `▒` | yes | no | no | 20 | Destroyed wall debris |
| `TileObject` | `•` | no | yes | no | 50 | Scatter object |

### Human Furniture

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileDesk` | `◊` | yes | no | yes | 30 | Desk/workstation |
| `TileChair` | `⊟` | yes | no | yes | 10 | Chair |
| `TileChairLeft` | `▦` | yes | no | yes | 10 | Chair (non-directional) |
| `TileChairRight` | `▦` | yes | no | yes | 10 | Chair (non-directional) |
| `TileComputer` | `⌸` | yes | no | yes | 10 | Computer terminal |
| `TileBed` | `□` | yes | no | yes | 20 | Bed/cot |
| `TileLocker` | `◫` | yes | no | yes | 30 | Locker/storage |
| `TileCabinet` | `⊞` | yes | no | yes | 30 | Cabinet/shelving |

### UFO / Alien Furniture

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileConsole` | `⌸` | yes | no | yes | 10 | Control panel |
| `TileMachinery` | `⊛` | yes | no | yes | 30 | Engine/generator |
| `TilePod` | `◈` | yes | no | yes | 30 | Alien pod |
| `TilePowerSource` | `⌁` | yes | no | yes | 20 | Power core |
| `TileStorage` | `▤` | yes | no | yes | 30 | Storage crate |
| `TileAlienTech` | `⊕` | yes | no | yes | 20 | Alien artifact |

### Vehicle Tiles

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileCar` | `▄` | no | yes | yes | 50 | Car left half |
| `TileCarMid` | `█` | no | yes | yes | 50 | Car middle roof |
| `TileCarRight` | `▄` | no | yes | yes | 50 | Car right half |
| `TileForklift` | `█` | no | yes | yes | 50 | Forklift left |
| `TileForkliftRight` | `⊏` | no | yes | yes | 50 | Forklift right |
| `TileFuelPump` | `8` | no | yes | yes | 30 | Fuel pump (explodes 5x5) |
| `TileContainerRed` | `█` | no | yes | no | 80 | Red shipping container |
| `TileContainerBlue` | `█` | no | yes | no | 80 | Blue shipping container |
| `TileContainerYellow` | `█` | no | yes | no | 80 | Yellow shipping container |
| `TileBusEnd` | `▄` | no | yes | yes | 50 | Bus left/right end |
| `TileBusMid` | `█` | no | yes | yes | 50 | Bus middle roof |
| `TileHeloBody` | `█` | no | yes | yes | 50 | Helicopter fuselage |
| `TileHeloTail` | `▄` | no | yes | yes | 30 | Helicopter tail boom |
| `TileHeloNose` | `▷` | no | yes | yes | 40 | Helicopter nose/cockpit |
| `TileHeloRotor` | `+` | yes | no | yes | 0 | Helicopter rotor (overhead) |
| `TileHeloRotorSides` | `-` | yes | no | yes | 0 | Rotor blade sides |
| `TileHeloBodyBack` | `█` | no | yes | yes | 50 | Rear fuselage |
| `TileHeloRotorBack` | `x` | yes | no | yes | 0 | Rear rotor |
| `TileHeloWindow` | `◣` | no | yes | yes | 0 | Helicopter window |
| `TileTractorCab` | `◣` | no | yes | yes | 30 | Tractor cab |
| `TileTractorBody` | `█` | no | yes | yes | 30 | Tractor body |
| `TileCrawlerLeft` | `◢` | no | yes | yes | 50 | Alien crawler left |
| `TileCrawlerMid` | `█` | no | yes | yes | 50 | Alien crawler middle |
| `TileCrawlerRight` | `◣` | no | yes | yes | 50 | Alien crawler right |
| `TileCrawlerLeg` | `^` | no | yes | yes | 20 | Alien crawler leg |
| `TileWheel` | `O` | no | no | yes | 10 | Vehicle wheel |
| `TileWheelSmall` | `o` | no | no | yes | 10 | Small wheel |

### Biome-Specific Structures

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileAdobe` | `█` | no | yes | no | 80 | Thick adobe wall |
| `TileMetalWall` | `█` | no | yes | no | 80 | Prefab metal wall |
| `TileWreck` | `▤` | no | yes | no | 80 | Aircraft wreckage |
| `TileTimber` | `≡` | no | yes | yes | 80 | Stacked timber |
| `TileDish` | `◗` | no | yes | no | 50 | Satellite dish |
| `TileTruck` | `▄` | no | yes | yes | 80 | Military truck |
| `TileIce` | `≈` | yes | no | no | 0 | Frozen lake ice |
| `TileStreetlamp` | `⌖` | no | yes | yes | 10 | Lamp (emits light) |
| `TileGlass` | `,` | yes | no | yes | 0 | Broken glass (noisy) |
| `TileDebris` | `` ` `` | yes | no | yes | 0 | Scattered debris (noisy) |
| `TileCryoPipe` | `╪` | no | yes | yes | 20 | Cryo pipe (vents gas) |
| `TileSkylight` | `⊙` | yes | no | yes | 0 | Glass floor (collapses) |

### Farm Biome

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileWheat` | `ψ` | yes | no | yes | 20 | Tall wheat/corn |
| `TileHayBale` | `█` | no | yes | yes | 60 | Hay bale (flammable) |

### Coastal Biome

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TilePier` | `═` | yes | no | no | 10 | Wooden pier |
| `TileDockCrate` | `▣` | no | yes | yes | 50 | Dock crate |
| `TileDryBush` | `*` | yes | no | yes | 20 | Dry coastal scrub |

### Mountain Biome

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileCliffFace` | `░` | no | yes | no | 80 | Impassable cliff |
| `TileScree` | `·` | yes | no | yes | 10 | Loose scree (noisy) |
| `TileBoulder` | `∩` | no | yes | no | 70 | Large boulder |

### Swamp Biome

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileSwampWater` | `≋` | yes | no | no | 5 | Murky water (high TU) |
| `TileCypressTree` | `♣` | no | yes | yes | 80 | Cypress tree |

### Jungle Biome

| TileType | Glyph | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|----------|--------|--------------|-------|-------------|
| `TileMud` | `≋` | yes | no | no | 5 | Deep mud (high TU) |
| `TileVine` | `‡` | yes | no | yes | 20 | Hanging vines |
| `TileBamboo` | `♣` | no | yes | yes | 60 | Bamboo thicket |

## Biome Reference

| Biome Tag | Ground Tile | Typical Structures | Notes |
|-----------|-------------|-------------------|-------|
| `urban` | TilePavement | Buildings, fences, streetlamps, cars | Dense placement, road grid |
| `forest` | TileGrass | Logging structures, trees, rocks | Tree clusters, open clearings |
| `desert` | TileSand | Adobe buildings, cacti, wrecks | Open sight lines, few obstacles |
| `polar` | TileSnow | Prefab structures, ice patches, containers | White-on-white visibility |
| `rural` | TileGrass | Farmhouses, barns, fences | Dirt roads, open fields |
| `ufo` | TileUFOFloor | Alien walls, pods, power sources | Interior-only, no rotation |
| `farm` | TileGrass | Barns, silos, windmills, wheat fields | Flammable crops, open terrain |
| `coastal` | TileSand | Piers, docks, boats, lighthouses | Water edges, limited approaches |
| `mountain` | TileScree | Caves, ridges, boulders | Verticality, cover-heavy |
| `swamp` | TileMarsh | Huts, cypress groves, sinkholes | Movement penalties, low vis |
| `jungle` | TileMud | Temple ruins, bamboo thickets | Dense cover, slow movement |

## Examples

### Simple building (farm_barn.json)
```json
{
  "id": "farm_barn",
  "tags": ["farm"],
  "width": 9,
  "height": 9,
  "weight": 2,
  "rows": [
    ".........",
    ".WWWWWWW.",
    ".WH...HW.",
    ".WH...HW.",
    ".WH...HW.",
    ".WH...HW.",
    ".WWW+WWW.",
    ".........",
    "........."
  ],
  "terrain": {
    ".": "TileGrass",
    "W": "TileWall",
    "+": "TileDoor"
  },
  "furniture": {
    "H": "TileHayBale"
  }
}
```

### Multi-tile vehicle (rural_helicopter.json)
```json
{
  "id": "rural_helicopter",
  "tags": ["rural", "urban", "coastal", "farm"],
  "width": 7,
  "height": 4,
  "weight": 1,
  "no_rotate": true,
  "rows": [
    " --+-- ",
    "  ===W ",
    "x=====F",
    "  o  o "
  ],
  "terrain": {
    "-": "TileHeloRotorSides",
    "+": "TileHeloRotor",
    "x": "TileHeloRotorBack",
    "=": "TileHeloBody",
    "F": "TileHeloNose",
    "W": "TileHeloWindow",
    "o": "TileWheelSmall"
  },
  "furniture": {}
}
```

## Best Practices

### Chunk Design

- **Keep chunks modular** — buildings should be 5-15 tiles per side for reliable placement.
- **Ensure access** — every building needs at least one door (`+`) or open wall. Units must be able to pathfind in.
- **Avoid interior dead zones** — rooms smaller than 2x2 can trap units and block AI navigation.
- **Tag appropriately** — a chunk tagged `["urban", "rural"]` can appear in either biome. Multi-tag is useful for generic structures (sheds, vehicles, containers) that fit anywhere.
- **Use `no_rotate: true`** for vehicles, directional objects, and any chunk where rotation would break the design (e.g. text, asymmetric vehicles).
- **Weight 0 for landmarks** — use `weight: 0` for signature structures that should only appear as anchors, never as random scatter.

### Tile Usage

- **Terrain vs furniture** — terrain tiles replace the ground layer; furniture tiles overlay it. Use `terrain` for walls, floors, doors. Use `furniture` for objects on the ground (desks, lockers, hay bales).
- **Passable perimeter** — faction spawn points are placed on open passable tiles. Ensure your chunk has passable tiles around its edges.
- **Destructible tiles** — walls (`TileWall`), fences, and furniture can be destroyed. Use indestructible tiles (`"de": false`) sparingly.
- **Cover values** — walls (80), rocks (70), trees (80), bushes (40). Chairs and computers provide minimal cover (10).
- **Explosive tiles** — `TileFuelPump` explodes in a 5x5 radius when destroyed. Use carefully near spawn areas.

### Performance

- Keep chunk dimensions under 20x20. Larger chunks may overlap poorly and reduce tactical variety.
- Avoid excessive furniture density — every object is independently tracked and textured.
- Flammable tiles (`TileWheat`, `TileHayBale`, `TileTimber`, `TileBush`, `TileDryBush`, `TileVine`) trigger fire propagation. Use them deliberately rather than filling the map.

### Fire & Destruction

- Flammable tiles can catch fire from nearby explosions or incendiary weapons.
- Fire spreads to adjacent flammable tiles each turn.
- Destroyed walls produce `TileRubble` which is passable at increased movement cost.
- Glass (`TileGlass`) and debris (`TileDebris`) tiles are noisy — units stepping on them alert nearby enemies.
- Cryo pipes (`TileCryoPipe`) release poisonous gas when destroyed.

### Creating New Chunks

1. Choose a biome tag for your chunk.
2. Sketch the layout on a grid, keeping glyphs distinct.
3. Define terrain and furniture mappings. Ensure each glyph used in `rows` has a mapping.
4. Set `weight` and whether the chunk can rotate.
5. Save as `data/maps/<your_chunk_name>.json`.
6. Test by running a mission in the target biome — verify placement, pathing, and visual appearance.

---

## Browser-Based Tools

Four HTML editor tools are available in `tools/` for creating and editing tile definitions, map chunks, full tactical maps, and mission files. Open them directly in a browser (no server required).

| Tool | File | Purpose |
|------|------|---------|
| **Tile Type Creator** | `tools/tile_creator.html` | Create and edit individual tile type definitions (glyph, color, properties, category). Export as JSONC. |
| **Area Editor** | `tools/area_editor.html` | Visually paint map fragment chunks on a grid. Assign IDs, biome tags, terrain/furniture mappings. Export as chunk JSON. |
| **Map Editor** | `tools/map_editor.html` | Create full tactical maps tile-by-tile with a palette sidebar. Includes a chunk panel for managing and stamping fragments. Load/save map JSON. |
| **Mission Editor** | `tools/mission_editor.html` | Assemble mission definitions: load a map, place alien units from a roster, set mission metadata (name, author, description, night toggle). Export mission JSON. |

All four load tile data from `tools/jsonc.js` (which bundles the `data/tiles/*.jsonc` files) and respect the same tile property system used by the Go engine. After editing tiles with the Tile Type Creator, re-run `tools/bundle_tiles.js` to regenerate `tools/tile_data.js` for the other editors.
