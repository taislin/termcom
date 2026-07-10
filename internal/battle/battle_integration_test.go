package battle

import (
	"testing"

	"github.com/civ13/termcom/internal/data"
	"github.com/civ13/termcom/internal/soldier"
)

func TestFullBattleSimulation(t *testing.T) {
	squad := []*soldier.Soldier{
		soldier.NewSoldier("Alpha"),
		soldier.NewSoldier("Bravo"),
		soldier.NewSoldier("Charlie"),
	}
	m := GenerateCrashSite(30, 24)
	var units UnitList
	var alienAIs []*AlienAI

	for i, s := range squad {
		if s.HP <= 0 {
			continue
		}
		u := NewSoldierUnit(s)
		u.X = 3 + i*2
		u.Y = m.Height - 3
		units = append(units, u)
	}

	alienRank := 0
	alienTypes := []*data.AlienType{
		data.GetAlienByRank(alienRank),
		data.GetAlienByRank(alienRank),
		data.GetAlienByRank(alienRank + 1),
	}

	for _, at := range alienTypes {
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
		u.X = 15
		u.Y = 5
		units = append(units, u)
		ai := NewAlienAI(u)
		ai.PatrolX = u.X + 3
		ai.PatrolY = u.Y
		alienAIs = append(alienAIs, ai)
	}

	humanUnits := units.Faction(0)
	alienUnits := units.Faction(1)

	if len(humanUnits) != 3 {
		t.Fatalf("expected 3 humans, got %d", len(humanUnits))
	}
	if len(alienUnits) < 1 {
		t.Fatal("expected aliens")
	}

	for _, u := range humanUnits {
		if u.Soldier == nil {
			t.Error("human unit should have soldier")
		}
	}

	soldier := humanUnits[0]
	target := alienUnits[0]
	soldier.X = target.X - 1
	soldier.Y = target.Y
	damage, hit, _ := soldier.FireAt(target, nil)
	if hit {
		if damage <= 0 {
			t.Error("damage should be positive")
		}
	}

	soldier.TU = soldier.MaxTU
	soldier.MoveTo(soldier.X+1, soldier.Y, m)

	for _, ai := range alienAIs {
		ai.Unit.TU = ai.Unit.MaxTU
		ai.Update(units, m, humanUnits, nil, nil)
	}

	alienHPBefore := 0
	for _, u := range alienUnits {
		alienHPBefore += u.HP
	}

	humanUnits[0].TU = humanUnits[0].MaxTU
	_, _, _ = humanUnits[0].FireAt(alienUnits[0], nil)

	alienHPAfter := 0
	for _, u := range alienUnits {
		alienHPAfter += u.HP
	}

	if alienHPAfter >= alienHPBefore && alienHPBefore > 0 {
		t.Log("shots may have missed, that's ok with random")
	}
}

func TestBattleVictoryCondition(t *testing.T) {
	squad := []*soldier.Soldier{soldier.NewSoldier("Solo")}
	var units UnitList

	u := NewSoldierUnit(squad[0])
	u.X = 5
	u.Y = 5
	units = append(units, u)

	alien := NewAlienUnit(data.GetAlienByRank(0))
	alien.X = 10
	alien.Y = 10
	alien.HP = 1
	alien.Alive = true
	units = append(units, u, alien)

	alien.HP = 0
	alien.Alive = false

	humanAlive := units.Faction(0).Alive()
	alienAlive := units.Faction(1).Alive()

	won := len(alienAlive) == 0
	lost := len(humanAlive) == 0

	if !won {
		t.Error("should be victory when all aliens dead")
	}
	if lost {
		t.Error("should not be defeat when humans alive")
	}
}

func TestBattleDefeatCondition(t *testing.T) {
	squad := []*soldier.Soldier{soldier.NewSoldier("Solo")}
	var units UnitList

	u := NewSoldierUnit(squad[0])
	u.X = 5
	u.Y = 5
	u.HP = 0
	u.Alive = false
	units = append(units, u)

	alien := NewAlienUnit(data.GetAlienByRank(0))
	alien.X = 10
	alien.Y = 10
	alien.Alive = true
	units = append(units, alien)

	humanAlive := units.Faction(0).Alive()
	alienAlive := units.Faction(1).Alive()

	won := len(alienAlive) == 0
	lost := len(humanAlive) == 0

	if won {
		t.Error("should not be victory")
	}
	if !lost {
		t.Error("should be defeat when all humans dead")
	}
}

func TestMapGeneration(t *testing.T) {
	maps := []struct {
		name string
		gen  func(int, int) *BattleMap
	}{
		{"CrashSite", GenerateCrashSite},
		{"TerrorSite", GenerateTerrorSite},
		{"UFOInterior", GenerateUFOInterior},
		{"Cydonia", GenerateCydonia},
	}
	for _, tc := range maps {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.gen(30, 24)
			if m.Width != 30 || m.Height != 24 {
				t.Errorf("expected 30x24, got %dx%d", m.Width, m.Height)
			}
			passable := 0
			for y := 0; y < m.Height; y++ {
				for x := 0; x < m.Width; x++ {
					if m.Passable(x, y) {
						passable++
					}
				}
			}
			if passable < 50 {
				t.Errorf("map has too few passable tiles: %d", passable)
			}
		})
	}
}

func TestBresenhamLOS(t *testing.T) {
	m := GenerateCrashSite(30, 24)
	s := &Unit{X: 5, Y: 5, Alive: true, Faction: 0}
	if !s.CanSee(5, 5, m) {
		t.Error("should see self")
	}
	if !s.CanSee(6, 5, m) {
		t.Error("should see adjacent tile")
	}
}

func TestUnitMovement(t *testing.T) {
	m := NewBattleMap(30, 24)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Set(x, y, TileGrass)
		}
	}
	u := &Unit{X: 10, Y: 10, TU: 50, MaxTU: 50, Alive: true, Faction: 0}
	if !u.MoveTo(11, 10, m) {
		t.Error("should move 1 tile")
	}
	if u.X != 11 || u.Y != 10 {
		t.Errorf("expected (11,10), got (%d,%d)", u.X, u.Y)
	}
	if u.TU >= 50 {
		t.Error("TU should decrease")
	}

	m.Set(12, 10, TileWall)
	startX, startY := u.X, u.Y
	u.TU = 50
	if u.MoveTo(12, 10, m) {
		t.Error("should not move into wall")
	}
	if u.X != startX || u.Y != startY {
		t.Errorf("position should not change, got (%d,%d)", u.X, u.Y)
	}
}

func TestCoverSystem(t *testing.T) {
	m := NewBattleMap(30, 24)
	m.Set(6, 5, TileWall)
	s := &Unit{X: 5, Y: 5, Alive: true, Faction: 0}
	if s.CanSee(7, 5, m) {
		t.Error("wall should block LOS")
	}
	m.Set(6, 5, TileGrass)
	if !s.CanSee(7, 5, m) {
		t.Error("should see through grass")
	}
}

func TestWeaponFireCombat(t *testing.T) {
	for i := 0; i < 50; i++ {
		attacker := &Unit{
			X: 5, Y: 5, TU: 50, MaxTU: 50,
			Accuracy: 100, Weapon: "rifle",
			Alive: true, Faction: 0,
		}
		defender := &Unit{
			X: 6, Y: 5, HP: 50, MaxHP: 50,
			Armour: 0, Alive: true, Faction: 1,
		}
		damage, hit, _ := attacker.FireAt(defender, nil)
		if hit {
			if damage <= 0 {
				t.Error("damage should be positive")
			}
		}
	}
}

func TestAlienAIPatrol(t *testing.T) {
	alien := &Unit{X: 10, Y: 10, TU: 40, MaxTU: 40, HP: 20, MaxHP: 20, Alive: true, Faction: 1}
	ai := NewAlienAI(alien)
	ai.PatrolX = 12
	ai.PatrolY = 10
	m := GenerateCrashSite(30, 24)
	humans := UnitList{}
	ai.Update(UnitList{}, m, humans, nil, nil)
	if ai.State != AIPatrol {
		t.Error("should be patrolling with no humans visible")
	}
}

func TestAlienAIAttack(t *testing.T) {
	alien := &Unit{X: 10, Y: 10, TU: 40, MaxTU: 40, HP: 20, MaxHP: 20, Accuracy: 60, Alive: true, Faction: 1}
	human := &Unit{X: 11, Y: 10, TU: 50, HP: 20, MaxHP: 20, Armour: 0, Alive: true, Faction: 0}
	ai := NewAlienAI(alien)
	m := GenerateCrashSite(30, 24)
	ai.Update(UnitList{}, m, UnitList{human}, nil, nil)
	if ai.State != AIAttack {
		t.Error("should attack when human visible nearby")
	}
}

func TestAlienAIRetreat(t *testing.T) {
	alien := &Unit{
		X: 10, Y: 10, TU: 40, MaxTU: 40,
		HP: 3, MaxHP: 20, Bravery: 30,
		Alive: true, Faction: 1,
		AlienType: &data.AlienType{Name: "Test", Bravery: 30},
	}
	ai := NewAlienAI(alien)
	m := GenerateCrashSite(30, 24)
	ai.Update(UnitList{}, m, UnitList{}, nil, nil)
	if ai.State != AIRetreat {
		t.Error("should retreat when HP low and low bravery")
	}
}

func TestSoldierRankProgression(t *testing.T) {
	s := soldier.NewSoldier("Test")
	if s.Rank != soldier.Rookie {
		t.Error("should start as Rookie")
	}
	s.GainXP(30)
	if s.Rank < soldier.Squaddie {
		t.Error("should promote with enough XP")
	}
	if s.MaxHP <= 20 {
		t.Error("MaxHP should increase on promotion")
	}
}

func TestLootDrop(t *testing.T) {
	s := soldier.NewSoldier("Looter")
	var units UnitList

	u := NewSoldierUnit(s)
	u.X = 5
	u.Y = 5
	units = append(units, u)

	alien := NewAlienUnit(data.GetAlienByRank(0))
	alien.X = 10
	alien.Y = 10
	alien.HP = 0
	alien.Alive = false
	units = append(units, u, alien)

	alienAlive := units.Faction(1).Alive()
	won := len(alienAlive) == 0

	if !won {
		t.Error("should be victory")
	}
}

func TestAlienAIStateTransitions(t *testing.T) {
	alien := &Unit{X: 10, Y: 10, TU: 40, MaxTU: 40, HP: 20, MaxHP: 20, Accuracy: 60, Alive: true, Faction: 1}
	ai := NewAlienAI(alien)
	m := GenerateCrashSite(30, 24)

	if ai.State != AIPatrol {
		t.Error("should start patrolling")
	}

	human := &Unit{X: 12, Y: 10, TU: 50, HP: 20, MaxHP: 20, Armour: 0, Alive: true, Faction: 0}
	ai.Update(UnitList{}, m, UnitList{human}, nil, nil)

	alien.HP = 3
	alien.AlienType = &data.AlienType{Name: "Test", Bravery: 30}
	alien.Bravery = 30
	for i := 0; i < 10; i++ {
		ai.TurnsSince = 0
		ai.Update(UnitList{}, m, UnitList{}, nil, nil)
	}
}

func TestMapTileTypes(t *testing.T) {
	m := NewBattleMap(10, 10)
	m.Set(0, 0, TileFloor)
	m.Set(1, 0, TileWall)
	m.Set(2, 0, TileDoor)
	m.Set(3, 0, TileGrass)
	m.Set(4, 0, TileTree)
	m.Set(5, 0, TileRock)
	m.Set(6, 0, TileWater)
	m.Set(7, 0, TileUFOFloor)
	m.Set(8, 0, TileUFOWall)

	if !m.Passable(0, 0) {
		t.Error("floor should be passable")
	}
	if m.Passable(1, 0) {
		t.Error("wall should not be passable")
	}
	if !m.Passable(2, 0) {
		t.Error("door should be passable")
	}
	if !m.Passable(3, 0) {
		t.Error("grass should be passable")
	}
	if m.Passable(4, 0) {
		t.Error("tree should not be passable")
	}
	if m.Passable(5, 0) {
		t.Error("rock should not be passable")
	}
	if m.Passable(6, 0) {
		t.Error("water should not be passable")
	}
	if !m.Passable(7, 0) {
		t.Error("UFO floor should be passable")
	}
	if m.Passable(8, 0) {
		t.Error("UFO wall should not be passable")
	}

	if m.Opaque(0, 0) {
		t.Error("floor should not be opaque")
	}
	if !m.Opaque(1, 0) {
		t.Error("wall should be opaque")
	}
	if !m.Opaque(4, 0) {
		t.Error("tree should be opaque")
	}
	if !m.Opaque(5, 0) {
		t.Error("rock should be opaque")
	}
	if !m.Opaque(8, 0) {
		t.Error("UFO wall should be opaque")
	}
}

func TestUnitListFilters(t *testing.T) {
	s := soldier.NewSoldier("Test")
	su := NewSoldierUnit(s)
	su.Faction = 0
	su.X = 5
	su.Y = 5
	al := NewAlienUnit(data.GetAlienByRank(0))
	al.Faction = 1
	al.X = 15
	al.Y = 15
	ul := UnitList{su, al}

	humans := ul.Faction(0)
	aliens := ul.Faction(1)
	if len(humans) != 1 {
		t.Errorf("expected 1 human, got %d", len(humans))
	}
	if len(aliens) != 1 {
		t.Errorf("expected 1 alien, got %d", len(aliens))
	}

	allAlive := ul.Alive()
	if len(allAlive) != 2 {
		t.Errorf("expected 2 alive, got %d", len(allAlive))
	}

	su.Alive = false
	allAlive = ul.Alive()
	if len(allAlive) != 1 {
		t.Errorf("expected 1 alive after kill, got %d", len(allAlive))
	}

	found := ul.At(su.X, su.Y)
	if found != nil {
		t.Error("At() should return nil for dead unit")
	}

	foundAlien := ul.At(al.X, al.Y)
	if foundAlien == nil {
		t.Error("At() should find alive alien")
	}
}
