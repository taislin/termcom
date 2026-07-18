package battle

import (
	"math/rand"
	"testing"

	"github.com/taislin/termcom/internal/mapgen"
)

func TestApplyMapgenChunkRotation(t *testing.T) {
	mapgen.Reset()
	if err := mapgen.LoadDir("../../data/maps"); err != nil {
		t.Fatalf("load maps: %v", err)
	}
	chunk := mapgen.Get("ruined_shack")
	if chunk == nil {
		t.Fatalf("ruined_shack chunk not loaded")
	}
	for rot := 0; rot < 4; rot++ {
		m := NewBattleMap(20, 20)
		ApplyMapgenChunkRotated(m, 2, 2, rot, chunk)
		foundFloor := false
		for y := 0; y < m.LevelHeight; y++ {
			for x := 0; x < m.Width; x++ {
				if m.At(x, y).Type == TileFloor {
					foundFloor = true
				}
			}
		}
		if !foundFloor {
			t.Errorf("rot %d: no floor stamped by chunk", rot)
		}
	}
}

func TestApplyMapgenChunkOverlapPreservesUnderlying(t *testing.T) {
	mapgen.Reset()
	if err := mapgen.LoadDir("../../data/maps"); err != nil {
		t.Fatalf("load maps: %v", err)
	}
	chunk := mapgen.Get("bus_stop_cover")
	if chunk == nil {
		t.Fatalf("bus_stop_cover chunk not loaded")
	}
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileWall)
	ApplyMapgenChunk(m, 4, 4, chunk)
	if m.At(5, 5).Type != TileFloor {
		t.Errorf("expected chunk floor to overwrite underlying wall, got %v", m.At(5, 5).Type)
	}
}

func TestAssembleMapBiomes(t *testing.T) {
	mapgen.Reset()
	if err := mapgen.LoadDir("../../data/maps"); err != nil {
		t.Fatalf("load maps: %v", err)
	}
	rng := rand.New(rand.NewSource(12345))
	for _, biome := range []string{"urban", "forest", "ufo", "alien", "desert", "polar"} {
		m := AssembleMap(biome, 40, 40, rng)
		if !m.ValidateMap() {
			t.Errorf("biome %s: assembled map failed validation", biome)
		}
	}
}

func TestBlobClusters(t *testing.T) {
	m := NewBattleMap(30, 30)
	rng := rand.New(rand.NewSource(7))
	m.Blob(TileTree, 3, 10, 50, rng)
	count := 0
	for y := 0; y < m.LevelHeight; y++ {
		for x := 0; x < m.Width; x++ {
			if m.At(x, y).Type == TileTree {
				count++
			}
		}
	}
	if count == 0 {
		t.Errorf("Blob produced no tiles")
	}
	if count > 3*10+30 {
		t.Errorf("Blob produced too many tiles: %d", count)
	}
}

func TestPoissonSpacing(t *testing.T) {
	m := NewBattleMap(30, 30)
	rng := rand.New(rand.NewSource(99))
	m.Poisson(TileRock, 3, 15, rng)
	var pts [][2]int
	for y := 0; y < m.LevelHeight; y++ {
		for x := 0; x < m.Width; x++ {
			if m.At(x, y).Type == TileRock {
				pts = append(pts, [2]int{x, y})
			}
		}
	}
	for i := 0; i < len(pts); i++ {
		for j := i + 1; j < len(pts); j++ {
			dx, dy := pts[i][0]-pts[j][0], pts[i][1]-pts[j][1]
			if dx*dx+dy*dy < 3*3 {
				t.Errorf("Poisson points too close: %v %v", pts[i], pts[j])
			}
		}
	}
}

func TestValidateMapFullyWalled(t *testing.T) {
	m := NewBattleMap(10, 10)
	m.fillRect(0, 0, 10, 10, TileWall)
	if m.ValidateMap() {
		t.Errorf("expected fully-walled map to fail validation")
	}
}

func TestAllGeneratorsValid(t *testing.T) {
	generators := []func() *BattleMap{
		func() *BattleMap { return GenerateForest(50, 50) },
		func() *BattleMap { return GenerateDesert(50, 50) },
		func() *BattleMap { return GeneratePolar(50, 50) },
		func() *BattleMap { return GenerateAbductionSite(50, 50) },
		func() *BattleMap { return GenerateTerrorSite(50, 50) },
		func() *BattleMap { return GenerateAlienBaseWFC(50, 50, rand.New(rand.NewSource(42))) },
		func() *BattleMap { return GenerateCydonia(50, 50) },
		func() *BattleMap { return GenerateUFOInterior(50, 50) },
	}
	for i, gen := range generators {
		m := gen()
		if m == nil {
			t.Errorf("generator %d returned nil", i)
			continue
		}
		if !m.ValidateMap() {
			t.Errorf("generator %d produced an invalid (unreachable) map", i)
		}
	}
}
