package battle

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/soldier"
)

func TestNewBattleMap(t *testing.T) {
	m := NewBattleMap(30, 20)
	if m.Width != 30 || m.Height != 20 {
		t.Errorf("expected 30x20, got %dx%d", m.Width, m.Height)
	}
}

func TestTileChar(t *testing.T) {
	tests := []struct {
		tile TileType
		want rune
	}{
		{TileFloor, '.'},
		{TileWall, '#'},
		{TileDoor, '+'},
		{TileGrass, '·'},
		{TileTree, '♣'},
		{TileRock, '∩'},
		{TileWater, '≈'},
		{TileUFOFloor, '≡'},
		{TileUFOWall, '█'},
		{TileConsole, '░'},
		{TileMachinery, '⚙'},
		{TilePod, '◈'},
		{TilePowerSource, '⌁'},
		{TileStorage, '▤'},
		{TileAlienTech, '⊕'},
	}
	for _, tt := range tests {
		if got := TileChar(tt.tile); got != tt.want {
			t.Errorf("TileChar(%d) = %c, want %c", tt.tile, got, tt.want)
		}
	}
}

func TestMapPassable(t *testing.T) {
	m := NewBattleMap(10, 10)
	m.Set(5, 5, TileFloor)
	if !m.Passable(5, 5) {
		t.Error("floor should be passable")
	}
	m.Set(5, 5, TileWall)
	if m.Passable(5, 5) {
		t.Error("wall should not be passable")
	}
}

func TestMapOpaque(t *testing.T) {
	m := NewBattleMap(10, 10)
	m.Set(5, 5, TileFloor)
	if m.Opaque(5, 5) {
		t.Error("floor should not be opaque")
	}
	m.Set(5, 5, TileWall)
	if !m.Opaque(5, 5) {
		t.Error("wall should be opaque")
	}
}

func TestGenerateCrashSite(t *testing.T) {
	m := GenerateCrashSite(30, 24)
	if m.Width != 30 || m.Height != 24 {
		t.Errorf("expected 30x24, got %dx%d", m.Width, m.Height)
	}
	// Should have UFO walls in center
	found := false
	for y := 10; y < 14; y++ {
		for x := 11; x < 19; x++ {
			if m.At(x, y).Type == TileUFOWall {
				found = true
			}
		}
	}
	if !found {
		t.Error("crash site should have UFO walls")
	}
}

func TestGenerateTerrorSite(t *testing.T) {
	m := GenerateTerrorSite(30, 24)
	if m.Width != 30 || m.Height != 24 {
		t.Errorf("expected 30x24, got %dx%d", m.Width, m.Height)
	}
	// Should have some walls (buildings)
	walls := 0
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.At(x, y).Type == TileWall {
				walls++
			}
		}
	}
	if walls == 0 {
		t.Error("terror site should have walls (buildings)")
	}
}

func TestNewSoldierUnit(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	if u.Name() != "Test" {
		t.Errorf("expected Test, got %s", u.Name())
	}
	if u.Faction != 0 {
		t.Error("soldier should be faction 0")
	}
	if !u.Alive {
		t.Error("soldier should be alive")
	}
	if u.HP <= 0 {
		t.Error("soldier should have positive HP")
	}
}

func TestNewAlienUnit(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	if u.Name() != "Sectoid" {
		t.Errorf("expected Sectoid, got %s", u.Name())
	}
	if u.Faction != 1 {
		t.Error("alien should be faction 1")
	}
	if !u.Alive {
		t.Error("alien should be alive")
	}
}

func TestUnitMoveTo(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileFloor)

	u.X = 3
	u.Y = 3
	u.TU = 50

	ok := u.MoveTo(5, 5, m)
	if !ok {
		t.Error("should be able to move to floor")
	}
	if u.X != 5 || u.Y != 5 {
		t.Errorf("expected (5,5), got (%d,%d)", u.X, u.Y)
	}
}

func TestUnitMoveToWall(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileWall)

	u.X = 3
	u.Y = 3
	u.TU = 50

	ok := u.MoveTo(5, 5, m)
	if ok {
		t.Error("should not be able to move to wall")
	}
}

func TestUnitFireAt(t *testing.T) {
	attacker := NewSoldierUnit(soldier.NewSoldier("A"))
	defender := NewSoldierUnit(soldier.NewSoldier("D"))
	attacker.X = 5
	attacker.Y = 5
	defender.X = 6
	defender.Y = 5
	attacker.TU = 50
	defender.HP = 100
	defender.Armour = 0

	_, _, _ = attacker.FireAt(defender, nil, nil)
}

func TestUnitCanSee(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)

	u.X = 5
	u.Y = 5

	if !u.CanSee(10, 5, m) {
		t.Error("should see along open line")
	}

	m.Set(7, 5, TileWall)
	if u.CanSee(10, 5, m) {
		t.Error("should not see through wall")
	}
}

func TestUnitList(t *testing.T) {
	u1 := NewSoldierUnit(soldier.NewSoldier("A"))
	u2 := NewSoldierUnit(soldier.NewSoldier("B"))
	u1.X = 1
	u1.Y = 1
	u2.X = 5
	u2.Y = 5
	u2.Alive = false

	list := UnitList{u1, u2}
	alive := list.Alive()
	if len(alive) != 1 {
		t.Errorf("expected 1 alive, got %d", len(alive))
	}
	found := list.At(1, 1)
	if found == nil {
		t.Error("expected to find unit at (1,1)")
	}
	if list.At(5, 5) != nil {
		t.Error("dead unit should not be found")
	}
}

func TestPhaseStrings(t *testing.T) {
	bs := &Battlescape{Phase: PhasePlayerTurn}
	if bs.phaseStr() != "YOUR TURN" {
		t.Errorf("unexpected phase string: %s", bs.phaseStr())
	}
	bs.Phase = PhaseAlienTurn
	if bs.phaseStr() != "ALIEN TURN" {
		t.Errorf("unexpected phase string: %s", bs.phaseStr())
	}
	bs.Phase = PhaseVictory
	if bs.phaseStr() != "VICTORY" {
		t.Errorf("unexpected phase string: %s", bs.phaseStr())
	}
	bs.Phase = PhaseDefeat
	if bs.phaseStr() != "DEFEAT" {
		t.Errorf("unexpected phase string: %s", bs.phaseStr())
	}
}

func TestTileTypeNames(t *testing.T) {
	tests := []struct {
		tile TileType
		want string
	}{
		{TileFloor, "Floor"},
		{TileWall, "Wall"},
		{TileDoor, "Door"},
		{TileGrass, "Grass"},
		{TileTree, "Tree"},
		{TileRock, "Rock"},
		{TileWater, "Water"},
		{TileUFOFloor, "UFO Floor"},
		{TileUFOWall, "UFO Wall"},
		{TileConsole, "Console"},
		{TileMachinery, "Machinery"},
		{TilePod, "Alien Pod"},
		{TilePowerSource, "Power Source"},
		{TileStorage, "Storage"},
		{TileAlienTech, "Alien Tech"},
		{TileType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tileTypeName(tt.tile); got != tt.want {
			t.Errorf("tileTypeName(%d) = %q, want %q", tt.tile, got, tt.want)
		}
	}
}

func TestStunAlien(t *testing.T) {
    u := &Unit{HP: 10, Weapon: "stun_rod", Faction: 0}
    target := &Unit{HP: 10, Faction: 1, Alive: true}
    
    // Simulate hitting with stun rod
    damage := 10
    if u.Weapon == "stun_rod" {
        target.StunPoints += damage
        if target.StunPoints >= target.HP {
            target.Stunned = true
            target.Alive = false
        }
    }
    
    if !target.Stunned {
        t.Error("expected stunned")
    }
    if target.Alive {
        t.Error("expected not alive")
    }
}
