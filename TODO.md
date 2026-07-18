# TODO — Game Improvements
Scope: Features and fixes for the battlescape tactical combat system.
---

## Completed

### Map Generation — Wave Function Collapse (WFC) UFO Interiors
- [x] `internal/battle/wfc.go`: `WFCTile` (ID, 3x3 RuneGrid, Neighbors[4]),
      `WFCRules` with precomputed `compatible[a][d][b]` matrix, `Wave` grid of
      `superposition{allowed []bool; count; collapsed}`.
- [x] Min-entropy observation: `minEntropyCell` (deterministic tie-break) +
      `observe` random collapse of lowest-entropy cell.
- [x] Queue-based `propagate`: removes neighbor tiles with no source-compatible
      option (handles both collapsed and superposition sources); returns false
      on contradiction.
- [x] `Solve` loop with restart-on-contradiction (maxRestarts) + best-effort
      fallback; `CompileToBattleMap` stamps 3x3 tiles into multi-level `BattleMap`.
- [x] `ufoWFCTiles()`: 17-piece modular library (floor, 4 walls, 4 corners,
      engine, console/pod/power rooms, 4 door variants) with adjacency rules.
- [x] `GenerateUFOInteriorWFC(w,h,rng)` builds a 2-level UFO; wired with
      benchmark `UFOInteriorWFC` and tests in `wfc_test.go`.
- [x] `urbanWFCTiles()`: small 3x3 urban pieces (floor, 4 walls, 4 corners,
      4 doors, office/bed/storage rooms) PLUS large multi-room blocks —
      `Apartment6` (6x6), `Warehouse6` (6x6), `Office9` (9x9), `Barracks9`
      (9x9); `RuneGrid` generalized to variable size to support big modules.
- [x] `GenerateUrbanBuildingWFC(w,h,rng)` builds an enclosed urban building
      (perimeter wall, pavement base) from the mixed tile library; tests in
      `wfc_test.go` verify enclosure + interior floors/furniture.
- [x] WFC tiles now loaded from JSON: `data/wfc/ufo.json` + `data/wfc/urban.json`
      (new `mapgen.WFCLibrary` schema with per-direction `neighbors`); hardcoded
      libs kept as fallback. `TestWFCJSONMatchesHardcoded` guards drift.
- [x] WFC wired into live missions: `Supply Raid` + `Alien Research` use
      `GenerateUFOInteriorWFC`; new `Building Assault` mission (`MISSION_BUILDING`,
      all 8 languages) uses `GenerateUrbanBuildingWFC`.
- [x] Expanded `data/maps/*.json` urban fragment library: apartment, shop,
      warehouse, park, rubble field, tower (added to existing building/shack).
      `AssembleMap` (terror/abduction/crash) consumes them.
