package battle

import (
	"testing"
)

func TestBattleMapLOS(t *testing.T) {
	// Create a 10x10 map
	m := NewBattleMap(10, 10)
	// Place a wall at (5, 5)
	m.Set(5, 5, TileWall)

	// LOS check: 4,5 to 6,5 (blocked by wall at 5,5)
	if m.hasLOS(4, 5, 6, 5) {
		t.Errorf("Expected LOS to be blocked at (5,5)")
	}

	// LOS check: 4,4 to 6,4 (open ground)
	if !m.hasLOS(4, 4, 6, 4) {
		t.Errorf("Expected LOS to be clear")
	}
}

func TestCoverAlongLine(t *testing.T) {
	m := NewBattleMap(10, 10)
	// Place a tree (60% cover) at (5, 5)
	m.Set(5, 5, TileTree)

	// Check cover along line (4,5) to (6,5)
	cover := m.CoverAlongLine(4, 5, 6, 5)
	if cover != 60 {
		t.Errorf("Expected 60%% cover, got %d%%", cover)
	}

	// Check cover along line (4,4) to (6,6) - should be 0 as it doesn't pass through (5,5)
	// Actually, Bresenham's line from (4,4) to (6,6) goes through (5,5)
	// (4,4) -> (5,5) -> (6,6)
	cover2 := m.CoverAlongLine(4, 4, 6, 6)
	if cover2 != 60 {
		t.Errorf("Expected 60%% cover, got %d%%", cover2)
	}
}

func TestBattleMapBounds(t *testing.T) {
	m := NewBattleMap(10, 10)
	
	// Test At() out of bounds
	t1 := m.At(-1, 0)
	if t1.Type != TileWall {
		t.Errorf("Expected TileWall for out of bounds")
	}

	// Test At() in bounds
	t2 := m.At(0, 0)
	if t2.Type == TileWall {
		t.Errorf("Expected TileGrass, got TileWall")
	}
}
