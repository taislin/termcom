# Map Chunk JSON Guide for LLMs

You are generating a **map fragment/chunk JSON file** for the X-COM tactical game. The fragment system uses biome-tagged chunks stamped onto an `AssembleMap` outdoor terrain grid. Fragments load automatically from `data/maps/*.json`.

## JSON Schema

```json
{
  "id": "unique_chunk_name",
  "tags": ["biome1", "biome2", ...],
  "width": <int>,
  "height": <int>,
  "weight": <int>,
  "no_rotate": <bool>,
  "rows": [
    "...",
    "..."
  ],
  "terrain": {
    "<glyph>": "TileTypeName",
    ...
  },
  "furniture": {
    "<glyph>": "TileTypeName",
    ...
  }
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique identifier, used as filename without `.json`. Must not collide with other chunks. |
| `tags` | Yes | List of biomes this chunk can appear in. Must include one or more valid biome names. |
| `width` | Yes | Width of the chunk in tiles (columns). |
| `height` | Yes | Height of the chunk in tiles (rows). |
| `weight` | No | Spawning weight (higher = more likely). Defaults to 1. Set to 0 to use as anchor-only. |
| `no_rotate` | No | If true, the chunk will never be rotated when placed. Default false. |
| `rows` | Yes | Array of strings forming the ASCII grid. Each string must be exactly `width` characters. There must be `height` rows. |
| `terrain` | Yes | Maps ASCII glyphs to TileType names for **ground/structural terrain**. |
| `furniture` | No | Maps ASCII glyphs to TileType names for **furniture/objects** placed on top of terrain. Glyphs used in `terrain` must not overlap with `furniture`. If a glyph is in both, `terrain` wins. |

### Rules

1. **rune set:** Use only BMP Unicode (U+0000-U+FFFF). No emoji (U+1F300+).
2. **rows:** All rows must be same length (=width). Use ASCII characters for the grid — the glyphs in `terrain`/`furniture` define the actual tile type mapping.
3. **Weight & anchors:** Higher weight = more likely to spawn. For large signature chunks (anchors) the game picks a chunk with `tags` matching the biome. Weight 0 chunks still appear as anchors.
4. **Tag your chunk properly:** It must be placed in the right biome(s). Multi-tag chunks can appear in multiple biomes.

### Available Biomes

| Biome tag | Description |
|-----------|-------------|
| `urban` | City streets, buildings, pavement |
| `forest` | Dense woodland, logging camps |
| `desert` | Sandy terrain, adobe, cacti |
| `polar` | Snow, ice, frozen structures |
| `rural` | Farmland, country roads, houses |
| `ufo` | UFO interiors, alien structures |
| `alien` | Alien base interiors |
| `farm` | Farm biome — barns, silos, windmills, crops |
| `coastal` | Coastline — piers, docks, boats, lighthouses |
| `mountain` | Rocky highlands — ridges, caves, scree |
| `swamp` | Marshland — cypress groves, huts, sinkholes |
| `jungle` | Dense jungle — bamboo, vines, temple ruins |

## Tile Type Reference

### Core Terrain (used in `terrain` mapping)

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileFloor` | `.` | 95,90,85 | yes | no | no | 0 | Interior floor |
| `TileWall` | `#` | 160,155,150 | no | yes | yes | 80 | Building wall |
| `TileDoor` | `+` | 140,100,50 | yes | no | yes | 0 | Doorway |
| `TileWindow` | `□` | 120,170,220 | no | yes | yes | 20 | Building window |
| `TileGrass` | `·` | 50,110,40 | yes | no | no | 0 | Open ground |
| `TileTree` | `♣` | 35,90,25 | no | yes | yes | 80 | Tree |
| `TileRock` | `∩` | 130,125,120 | no | yes | yes | 70 | Rock formation |
| `TileWater` | `≈` | 40,80,200 | no | no | no | 0 | Deep water (impassable) |
| `TileUFOFloor` | `≡` | 50,75,110 | yes | no | no | 0 | UFO interior floor |
| `TileUFOWall` | `█` | 70,100,150 | no | yes | yes | 80 | UFO hull wall |
| `TileStairs` | `▒` | 110,105,100 | yes | no | no | 0 | Stairs up |
| `TileStairsDown` | `▓` | 80,75,70 | yes | no | no | 0 | Stairs down |
| `TilePavement` | `░` | 120,120,120 | yes | no | no | 0 | Road / sidewalk |
| `TileSand` | `·` | 200,180,120 | yes | no | no | 0 | Sandy ground |
| `TileSnow` | `∗` | 230,235,245 | yes | no | no | 0 | Snowy ground |
| `TileMarsh` | `≋` | 60,100,70 | yes | no | no | 0 | Marsh / bog |
| `TileBush` | `†` | 45,100,35 | yes | no | yes | 40 | Bush / scrub |
| `TileFence` | `│` | 145,120,80 | no | yes | yes | 40 | Wooden fence |
| `TileRubble` | `▒` | 120,115,110 | yes | no | no | 20 | Destroyed wall debris |
| `TileObject` | `•` | 170,170,170 | no | yes | no | 50 | Scatter object |

### UFO Furniture (used in `terrain` or `furniture` mapping)

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileConsole` | `⌸` | 70,210,130 | yes | no | yes | 10 | Control panel |
| `TileMachinery` | `⊛` | 180,180,180 | yes | no | yes | 30 | Engine/generator |
| `TilePod` | `◈` | 130,70,190 | yes | no | yes | 30 | Alien pod |
| `TilePowerSource` | `⌁` | 240,200,60 | yes | no | yes | 20 | Power core |
| `TileStorage` | `▤` | 180,140,90 | yes | no | yes | 30 | Storage crate |
| `TileAlienTech` | `⊕` | 230,70,70 | yes | no | yes | 20 | Alien artifact |

### Human Furniture (used in `terrain` or `furniture` mapping)

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileDesk` | `◊` | 160,120,80 | yes | no | yes | 30 | Desk/workstation |
| `TileChair` | `⊟` | 150,100,60 | yes | no | yes | 10 | Chair |
| `TileChairLeft` | `⅃` | 150,100,60 | yes | no | yes | 10 | Chair facing left |
| `TileChairRight` | `L` | 150,100,60 | yes | no | yes | 10 | Chair facing right |
| `TileComputer` | `⌸` | 70,180,210 | yes | no | yes | 10 | Computer terminal |
| `TileBed` | `□` | 200,200,200 | yes | no | yes | 20 | Bed/cot |
| `TileLocker` | `◫` | 140,160,180 | yes | no | yes | 30 | Locker/storage |
| `TileCabinet` | `⊞` | 170,130,90 | yes | no | yes | 30 | Cabinet/shelving |

### Vehicle Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileCar` | `▄` | 50,100,180 | no | yes | yes | 50 | Car left half |
| `TileCarMid` | `█` | 50,100,180 | no | yes | yes | 50 | Car middle roof |
| `TileCarRight` | `▄` | 50,100,180 | no | yes | yes | 50 | Car right half |
| `TileForklift` | `█` | 200,160,40 | no | yes | yes | 50 | Forklift left |
| `TileForkliftRight` | `⊏` | 200,160,40 | no | yes | yes | 50 | Forklift right |
| `TileFuelPump` | `8` | 200,60,40 | no | yes | yes | 30 | Fuel pump (explodes 5x5) |
| `TileContainerRed` | `█` | 180,50,40 | no | yes | no | 80 | Red shipping container |
| `TileContainerBlue` | `█` | 50,80,180 | no | yes | no | 80 | Blue shipping container |
| `TileContainerYellow` | `█` | 200,170,40 | no | yes | no | 80 | Yellow shipping container |
| `TileBusEnd` | `▄` | 200,180,60 | no | yes | yes | 50 | Bus left/right end |
| `TileBusMid` | `█` | 200,180,60 | no | yes | yes | 50 | Bus middle roof |
| `TileHeloBody` | `█` | 60,70,85 | no | yes | yes | 50 | Helicopter fuselage |
| `TileHeloTail` | `▄` | 60,70,85 | no | yes | yes | 30 | Helicopter tail boom |
| `TileHeloNose` | `▷` | 130,200,230 | no | yes | yes | 40 | Helicopter nose/cockpit |
| `TileHeloRotor` | `+` | 180,180,180 | yes | no | yes | 0 | Helicopter rotor (overhead) |
| `TileHeloRotorSides` | `-` | 180,180,180 | yes | no | yes | 0 | Rotor blade sides |
| `TileHeloBodyBack` | `█` | 60,70,85 | no | yes | yes | 50 | Rear fuselage |
| `TileHeloRotorBack` | `x` | 180,180,180 | yes | no | yes | 0 | Rear rotor |
| `TileHeloWindow` | `◣` | 130,200,230 | no | yes | yes | 0 | Helicopter window |
| `TileTractorCab` | `◣` | 130,200,230 | no | yes | yes | 30 | Tractor cab |
| `TileTractorBody` | `█` | 180,60,40 | no | yes | yes | 30 | Tractor body |
| `TileCrawlerLeft` | `◢` | 130,70,190 | no | yes | yes | 50 | Alien crawler left |
| `TileCrawlerMid` | `█` | 130,70,190 | no | yes | yes | 50 | Alien crawler middle |
| `TileCrawlerRight` | `◣` | 130,70,190 | no | yes | yes | 50 | Alien crawler right |
| `TileCrawlerLeg` | `^` | 100,50,160 | no | yes | yes | 20 | Alien crawler leg |
| `TileWheel` | `O` | 60,60,60 | no | no | yes | 10 | Vehicle wheel |
| `TileWheelSmall` | `o` | 60,60,60 | no | no | yes | 10 | Small wheel |

### Biome Structure Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileAdobe` | `█` | 200,130,70 | no | yes | no | 80 | Thick adobe wall |
| `TileMetalWall` | `█` | 180,185,195 | no | yes | no | 80 | Prefab metal wall |
| `TileWreck` | `▤` | 150,95,60 | no | yes | no | 80 | Aircraft wreckage |
| `TileTimber` | `≡` | 150,110,60 | no | yes | yes | 80 | Stacked timber |
| `TileDish` | `◗` | 170,175,185 | no | yes | no | 50 | Satellite dish |
| `TileTruck` | `▄` | 90,110,70 | no | yes | yes | 80 | Military truck |
| `TileIce` | `≈` | 180,220,235 | yes | no | no | 0 | Frozen lake ice |
| `TileStreetlamp` | `⌖` | 220,210,120 | no | yes | yes | 10 | Lamp (emits light) |
| `TileGlass` | `,` | 190,200,210 | yes | no | yes | 0 | Broken glass (noisy) |
| `TileDebris` | `` ` `` | 150,140,130 | yes | no | yes | 0 | Scattered debris (noisy) |
| `TileCryoPipe` | `╪` | 140,200,230 | no | yes | yes | 20 | Cryo pipe (vents gas) |
| `TileSkylight` | `⊙` | 180,210,240 | yes | no | yes | 0 | Glass floor (collapses) |

### Farm Biome Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileWheat` | `ψ` | 200,180,60 | yes | no | yes | 20 | Tall wheat/corn |
| `TileHayBale` | `█` | 160,140,60 | no | yes | yes | 60 | Hay bale (flammable) |

### Coastal Biome Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TilePier` | `═` | 140,100,60 | yes | no | no | 10 | Wooden pier |
| `TileDockCrate` | `▣` | 150,120,80 | no | yes | yes | 50 | Dock crate |
| `TileDryBush` | `*` | 170,140,60 | yes | no | yes | 20 | Dry coastal scrub |

### Mountain Biome Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileCliffFace` | `░` | 140,120,100 | no | yes | no | 80 | Impassable cliff |
| `TileScree` | `·` | 160,150,130 | yes | no | yes | 10 | Loose scree (noisy) |
| `TileBoulder` | `∩` | 130,125,120 | no | yes | no | 70 | Large boulder |

### Swamp Biome Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileSwampWater` | `≋` | 50,100,80 | yes | no | no | 5 | Murky water (high TU) |
| `TileCypressTree` | `♣` | 40,85,50 | no | yes | yes | 80 | Cypress tree |

### Jungle Biome Tiles

| TileType | Glyph | Color (RGB) | Passable | Opaque | Destructible | Cover | Description |
|----------|-------|-------------|----------|--------|--------------|-------|-------------|
| `TileMud` | `≋` | 110,80,50 | yes | no | no | 5 | Deep mud (high TU) |
| `TileVine` | `‡` | 50,130,50 | yes | no | yes | 20 | Hanging vines |
| `TileBamboo` | `♣` | 80,150,60 | no | yes | yes | 60 | Bamboo thicket |

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

## Important Notes

- **No emoji:** Use only BMP Unicode (U+0000-U+FFFF) for glyphs in `terrain` and `furniture` mappings. Box drawing and symbols are fine.
- **Fragments are deterministic:** Tile layout must be reproducible from the JSON alone.
- **Pathing:** Ensure your chunk has a passable perimeter or doorway connection so the pathfinding can reach it.
- **Vehicle chunks:** Use `"no_rotate": true` for vehicles (helicopters, trucks, etc.) that must maintain orientation.
- **Weight 0:** Fragments with weight 0 still appear as anchors but not as random scatter.
- **Furniture** is placed on top of terrain. If a glyph is in both `terrain` and `furniture`, terrain wins.
- **Fire/smoke interaction:** Flammable tiles (TileWheat, TileHayBale, TileTimber, TileBush, TileDryBush, TileVine) can catch fire.
