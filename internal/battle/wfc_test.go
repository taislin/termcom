package battle

import (
	"math/rand"
	"testing"
)

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

	m := NewMultiLevelBattleMap(12, 12, 1)
	wv.CompileToBattleMap(m, 0, 0, 0)

	for gy := 0; gy < wv.h; gy++ {
		for gx := 0; gx < wv.w; gx++ {
			id := wv.cells[gy][gx].collapsed
			if id < 0 {
				t.Fatalf("uncollapsed cell (%d,%d)", gx, gy)
			}
			tile := wv.rules.Tiles[id]
			for ty := 0; ty < 3; ty++ {
				for tx := 0; tx < 3; tx++ {
					mx, my := gx*3+tx, gy*3+ty
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

func TestUrbanWFCTilesHaveVariableSizes(t *testing.T) {
	tiles := urbanWFCTiles()
	sizes := map[int]int{}
	for _, tdef := range tiles {
		sizes[tdef.gridCols()]++
	}
	// Expect both 3x3 small pieces and larger multi-room blocks (6x6, 9x9).
	if sizes[3] == 0 {
		t.Fatal("expected 3x3 urban tiles")
	}
	if sizes[6] == 0 && sizes[9] == 0 {
		t.Fatal("expected large multi-room urban blocks (6x6 or 9x9)")
	}
}
