package battle

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
)

func TestStampVehicleOnMap(t *testing.T) {
	bp := data.GenerateScoutUFO()
	m := NewBattleMap(30, 30)

	result := StampVehicleOnMap(bp, 10, 10, m, 0.0)

	if result.PartsDestroyed != 0 {
		t.Errorf("severity 0 should destroy 0 parts, got %d", result.PartsDestroyed)
	}
	totalParts := len(bp.Parts)
	if result.PartsSurvived != totalParts {
		t.Errorf("expected %d survived, got %d", totalParts, result.PartsSurvived)
	}
}

func TestStampVehicleFullDestruction(t *testing.T) {
	bp := data.GenerateSmallScout()
	m := NewBattleMap(30, 30)

	result := StampVehicleOnMap(bp, 10, 10, m, 1.0)

	if result.PartsSurvived > 0 {
		t.Errorf("severity 1.0 should destroy all parts, but %d survived", result.PartsSurvived)
	}
}

func TestStampVehiclePartialDamage(t *testing.T) {
	bp := data.GenerateHeavyUFO()
	m := NewBattleMap(30, 30)

	// Run 100 times to average out randomness
	totalSurvived := 0
	for i := 0; i < 100; i++ {
		result := StampVehicleOnMap(bp, 5, 5, m, 0.5)
		totalSurvived += result.PartsSurvived
	}
	avgSurvived := float64(totalSurvived) / 100.0
	totalParts := float64(len(bp.Parts))
	ratio := avgSurvived / totalParts

	// At severity 0.5, roughly 25-75% should survive
	if ratio < 0.1 || ratio > 0.9 {
		t.Errorf("at severity 0.5, expected ~25-75%% survival, got %.1f%%", ratio*100)
	}
}

func TestStampVehicleOutOfBounds(t *testing.T) {
	bp := data.GenerateScoutUFO()
	m := NewBattleMap(10, 10)

	// Place at the edge — many parts should be clipped
	result := StampVehicleOnMap(bp, 9, 9, m, 0.0)

	if result.PartsSurvived > len(bp.Parts) {
		t.Errorf("cannot survive more than total parts: %d > %d", result.PartsSurvived, len(bp.Parts))
	}
}

func TestStampVehicleLoot(t *testing.T) {
	bp := data.NewVehicleBlueprint("loot_test", 3, 3)
	bp.Place(data.VehiclePartDefs["power_fission"], 0, 0)
	bp.Place(data.VehiclePartDefs["engine_standard"], 1, 0)
	bp.Place(data.VehiclePartDefs["hull_light"], 2, 0)

	m := NewBattleMap(30, 30)
	result := StampVehicleOnMap(bp, 10, 10, m, 0.0)

	if len(result.LootItems) != 2 {
		t.Errorf("expected 2 loot items (power+engine), got %d: %v", len(result.LootItems), result.LootItems)
	}
}

func TestStampVehicleRubble(t *testing.T) {
	bp := data.NewVehicleBlueprint("rubble_test", 1, 1)
	bp.Place(data.VehiclePartDefs["hull_light"], 0, 0)

	m := NewBattleMap(10, 10)
	StampVehicleOnMap(bp, 5, 5, m, 1.0)

	tile := m.Tiles[5][5]
	if tile.Type != TileRubble {
		t.Errorf("expected rubble at (5,5) after full destruction, got %v", tile.Type)
	}
}

func TestStampVehiclePreservesMapOutside(t *testing.T) {
	bp := data.GenerateSmallScout()
	m := NewBattleMap(30, 30)

	// Mark a distant tile
	m.Tiles[0][0] = Tile{Type: TileTree, Cover: 60}

	StampVehicleOnMap(bp, 10, 10, m, 0.0)

	if m.Tiles[0][0].Type != TileTree {
		t.Error("stamp should not modify tiles outside the vehicle area")
	}
}

func TestExplodeTile(t *testing.T) {
	m := NewBattleMap(20, 20)
	// Place some vehicle parts
	m.Tiles[10][10] = Tile{Type: TilePowerSource, Cover: 50, Rune: '⚙'}
	m.Tiles[10][11] = Tile{Type: TileUFOWall, Cover: 80, Rune: '█'}
	m.Tiles[10][12] = Tile{Type: TileMachinery, Cover: 50, Rune: '⌖'}

	ExplodeTile(10, 10, m, 3)

	// Power source should always be destroyed by explosion
	if m.Tiles[10][10].Type != TileRubble {
		t.Error("power source should be destroyed by explosion")
	}
}
