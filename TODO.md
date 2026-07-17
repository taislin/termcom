# TODO — Game Improvements
Scope: Features and fixes for the battlescape tactical combat system.
---

## Map Generation Improvements

### 1. Modular Map Fragments
- [x] Define `MapFragment` struct:
      `{ W, H int; Tiles [][]TileType; Anchor [2]int; Tags []string; DoorSides []int }`
- [x] Add optional rotation support (0/90/180/270) in placement.
- [x] Create `internal/battle/fragments.go` with a hand-authored library of
      ~10–15 reusable pieces, each tagged by biome (`urban`, `forest`, `ufo`, `alien`):
      - ruined shack, bus-stop cover, junction, corridor-elbow,
        UFO pod-room, alien altar, etc.
- [x] Implement `PlaceFragment(m, frag, x, y, rot)` with:
      - rotation transform,
      - overlap rejection (preserve existing walls/floors unless `Overwrite` set),
      - door-side stamping from `frag.DoorSides`.
- [x] Implement `AssembleMap(biome, w, h, rng)` that:
      1. fills base terrain (existing scatter),
      2. places 1 anchor fragment (UFO hull / command core),
      3. greedily places N fragments with spacing + connectivity (flood-fill) check,
      4. stamps corridors (`generateCorridor`) between adjacent fragment doors.
- [x] Keep `GenerateCrashSite` / `GenerateTerrorSite` / `GenerateForest` / etc. as
      thin wrappers calling `AssembleMap` with biome-specific fragment sets
      (preserves existing tests and map contracts).

### 2. Clustered Terrain Logic
- [x] Blob growth: seed K centers of a tile type, expand to neighbors with
      probability `p` until target cluster size — produces thickets, rock fields,
      rubble piles instead of even sprinkling.
- [x] Poisson disc sampling: for sparse-but-even cover (e.g. trees in
      abduction site) to avoid both clumping and grid uniformity.
- [x] Biome-aware clustering: group terrain by region
      (forest = tree blobs + bush halo; desert = sand dunes + rock islands;
      polar = marsh patches in snow). Replace per-tile `CmdScatter` calls in
      `GenerateForest` / `GenerateDesert` / `GeneratePolar`.
- [x] New `MapCommand` variants:
      - `CmdBlob{Type, Seeds, Size, Prob}`
      - `CmdPoisson{Type, Radius, Count}`
      so clustering composes with the existing command queue.
- [x] Connectivity guard: after clustering, flood-fill to ensure no cluster
      fully walls off spawn/objective areas; carve fallback corridors if needed.

### 3. Cross-Cutting
- [x] Determinism: thread a `rand.Source` / `seed` through all generators so
      fragments + clusters are reproducible (important for save/load and the
      seeded crash-site path in `GenerateCrashSite`).
- [x] Add `ValidateMap(m)`: reachability + minimum open-space check, reused by
      tests to catch generator regressions.
- [x] Tests:
      - `fragments_test.go` (rotation correctness, overlap rejection),
      - `cluster_test.go` (blob size bounds, poisson spacing),
      alongside existing `map_test.go`.
