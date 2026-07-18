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
