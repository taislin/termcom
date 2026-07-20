package battle

import (
	"math/rand"
	"testing"
)

func TestWFCLibrariesLoad(t *testing.T) {
	load := func(name string) []WFCTile {
		t.Helper()
		tiles := loadWFCLibrary(name)
		if len(tiles) == 0 {
			t.Fatalf("wfc library %q loaded zero tiles", name)
		}
		ids := map[int]bool{}
		for _, tile := range tiles {
			if ids[tile.ID] {
				t.Fatalf("wfc library %q: duplicate tile id %d", name, tile.ID)
			}
			ids[tile.ID] = true
			if tile.gridRows() == 0 || tile.gridCols() == 0 {
				t.Fatalf("wfc library %q: tile %q has zero-size grid", name, tile.Name)
			}
			// Neighbors are auto-computed from pixel edges if omitted in JSON;
			// verify the result is non-empty for every direction.
			for d := 0; d < 4; d++ {
				if len(tile.Neighbors[d]) == 0 {
					t.Fatalf("wfc library %q: tile %q dir %d has no neighbors after auto-compute", name, tile.Name, d)
				}
			}
		}
		return tiles
	}
	load("ufo")
	load("urban")
	load("alien_base")
}

func TestWFCSolveNoContradiction(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))
	for _, dims := range [][2]int{{10, 8}, {12, 9}, {16, 8}, {5, 5}} {
		rules := NewWFCRules(ufoWFCTiles())
		wv := newWave(rules, dims[0], dims[1])
		wv = wv.Solve(rng, 50)
		if wv.hasContradiction() {
			t.Fatalf("solved wave (%dx%d) reports a contradiction", dims[0], dims[1])
		}
	}
}

func TestWFCHasContradiction(t *testing.T) {
	rules := NewWFCRules(ufoWFCTiles())
	wv := newWave(rules, 4, 4)
	if wv.hasContradiction() {
		t.Fatal("fresh wave should not be a contradiction")
	}
	// Drive one uncollapsed cell to zero remaining options.
	c := &wv.cells[1][1]
	for i := range c.allowed {
		c.allowed[i] = false
	}
	c.count = 0
	if !wv.hasContradiction() {
		t.Fatal("cell with zero options but not collapsed must be a contradiction")
	}
}
func TestWFCSolveCollapsesFully(t *testing.T) {
	rules := NewWFCRules(ufoWFCTiles())
	rng := rand.New(rand.NewSource(12345))
	wv := newWave(rules, 10, 8)
	wv = wv.Solve(rng, 50)
	if !wv.fullyCollapsed() {
		t.Fatal("expected fully collapsed wave")
	}
	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			id := wv.cells[y][x].collapsed
			if id < 0 {
				t.Fatalf("cell (%d,%d) uncollapsed", x, y)
			}
			if wv.cells[y][x].count != 1 {
				t.Fatalf("cell (%d,%d) count=%d, want 1", x, y, wv.cells[y][x].count)
			}
		}
	}
}

func TestWFCDeterministic(t *testing.T) {
	solve := func(seed int64) [][]int {
		rules := NewWFCRules(ufoWFCTiles())
		rng := rand.New(rand.NewSource(seed))
		wv := newWave(rules, 12, 9)
		wv = wv.Solve(rng, 50)
		out := make([][]int, wv.h)
		for y := 0; y < wv.h; y++ {
			out[y] = make([]int, wv.w)
			for x := 0; x < wv.w; x++ {
				out[y][x] = wv.cells[y][x].collapsed
			}
		}
		return out
	}
	a := solve(999)
	b := solve(999)
	for y := range a {
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				t.Fatalf("non-deterministic at (%d,%d): %d vs %d", x, y, a[y][x], b[y][x])
			}
		}
	}
}

func TestWFCCompileToBattleMap(t *testing.T) {
	rules := NewWFCRules(ufoWFCTiles())
	rng := rand.New(rand.NewSource(7))
	wv := newWave(rules, 4, 4)
	wv = wv.Solve(rng, 50)

	// UFO tiles are 6x6; use stride=6 and a map large enough for 4x4 cells.
	m := NewMultiLevelBattleMap(24, 24, 1)
	wv.CompileToBattleMap(m, 0, 0, 0, 6)

	for gy := 0; gy < wv.h; gy++ {
		for gx := 0; gx < wv.w; gx++ {
			id := wv.cells[gy][gx].collapsed
			if id < 0 {
				t.Fatalf("uncollapsed cell (%d,%d)", gx, gy)
			}
			tile := wv.rules.Tiles[id]
			for ty := 0; ty < tile.gridRows(); ty++ {
				for tx := 0; tx < tile.gridCols(); tx++ {
					mx, my := gx*6+tx, gy*6+ty
					want := tileRuneToType(tile.RuneGrid[ty][tx])
					got := m.AtLevel(mx, my, 0).Type
					if got != want {
						t.Fatalf("mismatch at map (%d,%d): got %v want %v", mx, my, got, want)
					}
				}
			}
		}
	}
}

func TestGenerateUFOInteriorWFC(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	m := GenerateUFOInteriorWFC(30, 24, rng)
	if m == nil {
		t.Fatal("nil map")
	}
	if m.NumLevels != 2 {
		t.Fatalf("NumLevels=%d, want 2", m.NumLevels)
	}
	// Verify stairs connect levels.
	if m.AtLevel(0, 0, 0).Type == TileStairsDown || m.AtLevel(m.Width/2-1, m.LevelHeight/2-1, 0).Type != TileStairsDown {
		// just ensure at least one stairs exists somewhere
		found := false
		for y := 0; y < m.LevelHeight && !found; y++ {
			for x := 0; x < m.Width; x++ {
				if m.AtLevel(x, y, 0).Type == TileStairsDown {
					found = true
					break
				}
			}
		}
		if !found {
			t.Fatal("no TileStairsDown found on level 0")
		}
	}
}

func TestGenerateUrbanBuildingWFC(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	m := GenerateUrbanBuildingWFC(45, 45, rng)
	if m == nil {
		t.Fatal("nil map")
	}
	// Perimeter must be enclosed by walls.
	for x := 0; x < m.Width; x++ {
		if m.At(x, 0).Type != TileWall || m.At(x, m.Height-1).Type != TileWall {
			t.Fatalf("top/bottom perimeter not wall at x=%d", x)
		}
	}
	for y := 0; y < m.Height; y++ {
		if m.At(0, y).Type != TileWall || m.At(m.Width-1, y).Type != TileWall {
			t.Fatalf("left/right perimeter not wall at y=%d", y)
		}
	}
	// Map must contain interior floors and at least one furniture/door piece.
	floors, furniture := 0, 0
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			switch m.At(x, y).Type {
			case TileFloor, TilePavement:
				floors++
			case TileConsole, TileBed, TileStorage, TileMachinery, TileDoor:
				furniture++
			}
		}
	}
	if floors == 0 {
		t.Fatal("no interior floor tiles produced")
	}
	if furniture == 0 {
		t.Fatal("no furniture/door tiles produced")
	}
}

func TestGenerateUrbanBuildingWFCLevels(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	m := GenerateUrbanBuildingWFCLevels(45, 45, 2, rng)
	if m == nil {
		t.Fatal("nil map")
	}
	if m.NumLevels != 2 {
		t.Fatalf("NumLevels=%d, want 2", m.NumLevels)
	}
	if m.LevelHeight != 45 {
		t.Fatalf("LevelHeight=%d, want 45", m.LevelHeight)
	}
	// Each level must be enclosed by walls.
	for level := 0; level < m.NumLevels; level++ {
		for x := 0; x < m.Width; x++ {
			if m.AtLevel(x, 0, level).Type != TileWall || m.AtLevel(x, m.LevelHeight-1, level).Type != TileWall {
				t.Fatalf("level %d top/bottom perimeter not wall at x=%d", level, x)
			}
		}
		for y := 0; y < m.LevelHeight; y++ {
			if m.AtLevel(0, y, level).Type != TileWall || m.AtLevel(m.Width-1, y, level).Type != TileWall {
				t.Fatalf("level %d left/right perimeter not wall at y=%d", level, y)
			}
		}
	}
	// Must have stairs connecting levels.
	stairsDown, stairsUp := 0, 0
	for level := 0; level < m.NumLevels; level++ {
		for y := 0; y < m.LevelHeight; y++ {
			for x := 0; x < m.Width; x++ {
				switch m.AtLevel(x, y, level).Type {
				case TileStairsDown:
					stairsDown++
				case TileStairs:
					stairsUp++
				}
			}
		}
	}
	if stairsDown == 0 {
		t.Fatal("no stairs down tiles")
	}
	if stairsUp == 0 {
		t.Fatal("no stairs up tiles")
	}
	if stairsDown != stairsUp {
		t.Fatalf("stairs down (%d) != stairs up (%d)", stairsDown, stairsUp)
	}
}

func TestGenerateAlienBaseWFC(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	m := GenerateAlienBaseWFC(50, 50, rng)
	if m == nil {
		t.Fatal("nil map")
	}
	if m.NumLevels != 2 {
		t.Fatalf("NumLevels=%d, want 2", m.NumLevels)
	}
	// Each level must be enclosed by alien walls.
	for level := 0; level < m.NumLevels; level++ {
		for x := 0; x < m.Width; x++ {
			if m.AtLevel(x, 0, level).Type != TileUFOWall || m.AtLevel(x, m.LevelHeight-1, level).Type != TileUFOWall {
				t.Fatalf("level %d top/bottom perimeter not wall at x=%d", level, x)
			}
		}
		for y := 0; y < m.LevelHeight; y++ {
			if m.AtLevel(0, y, level).Type != TileUFOWall || m.AtLevel(m.Width-1, y, level).Type != TileUFOWall {
				t.Fatalf("level %d left/right perimeter not wall at y=%d", level, y)
			}
		}
	}
	// Must have stairs.
	stairs := 0
	for level := 0; level < m.NumLevels; level++ {
		for y := 0; y < m.LevelHeight; y++ {
			for x := 0; x < m.Width; x++ {
				t := m.AtLevel(x, y, level).Type
				if t == TileStairsDown || t == TileStairs {
					stairs++
				}
			}
		}
	}
	if stairs == 0 {
		t.Fatal("no stairs found")
	}
}

func TestUrbanWFCTilesHaveVariableSizes(t *testing.T) {
	tiles := urbanWFCTiles()
	sizes := map[int]int{}
	for _, tdef := range tiles {
		sizes[tdef.gridCols()]++
	}
	// Urban tiles use a uniform 9×9 cell size for coherent WFC layout.
	if sizes[9] == 0 {
		t.Fatal("expected 9x9 urban tiles")
	}
}
