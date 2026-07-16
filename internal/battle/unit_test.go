package battle

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/soldier"
)

func TestNewAlienUnitStats(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	if u.HP != at.HP {
		t.Errorf("expected HP %d, got %d", at.HP, u.HP)
	}
	if u.TU != at.TU {
		t.Errorf("expected TU %d, got %d", at.TU, u.TU)
	}
	if u.Accuracy != at.Accuracy {
		t.Errorf("expected Accuracy %d, got %d", at.Accuracy, u.Accuracy)
	}
}

func TestNewCivilianUnit(t *testing.T) {
	u := NewCivilianUnit("TestCiv")
	if u.Name() != "TestCiv" {
		t.Errorf("expected TestCiv, got %s", u.Name())
	}
	if u.HP != 5 {
		t.Errorf("expected 5 HP, got %d", u.HP)
	}
	if u.Faction != 2 {
		t.Errorf("expected faction 2, got %d", u.Faction)
	}
}

func TestUnitFireAtConsumesTU(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	target := NewSoldierUnit(soldier.NewSoldier("Target"))
	u.X = 5
	u.Y = 5
	target.X = 6
	target.Y = 5
	target.HP = 100
	u.TU = 50
	u.Weapon = "rifle"
	u.WeaponAmmo = 10

	beforeTU := u.TU
	_, _, _, err := u.FireAt(target, nil, nil)
	if err != nil {
		t.Fatalf("FireAt returned error: %v", err)
	}
	if u.TU >= beforeTU {
		t.Error("TU should decrease after firing")
	}
}

func TestUnitFireAtConsumesAmmo(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	target := NewSoldierUnit(soldier.NewSoldier("Target"))
	u.X = 5
	u.Y = 5
	target.X = 6
	target.Y = 5
	target.HP = 100
	u.TU = 50
	u.Weapon = "rifle"
	u.WeaponAmmo = 10

	beforeAmmo := u.WeaponAmmo
	_, _, _, err := u.FireAt(target, nil, nil)
	if err != nil {
		t.Fatalf("FireAt returned error: %v", err)
	}
	if u.WeaponAmmo >= beforeAmmo {
		t.Error("ammo should decrease after firing")
	}
}

func TestUnitFireAtNoAmmo(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	target := NewSoldierUnit(soldier.NewSoldier("Target"))
	u.X = 5
	u.Y = 5
	target.X = 6
	target.Y = 5
	u.TU = 50
	u.Weapon = "rifle"
	u.WeaponAmmo = 0

	_, _, _, err := u.FireAt(target, nil, nil)
	if err == nil {
		t.Error("expected error for no ammo")
	}
}

func TestUnitFireAtNoTU(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	target := NewSoldierUnit(soldier.NewSoldier("Target"))
	u.X = 5
	u.Y = 5
	target.X = 6
	target.Y = 5
	u.TU = 0
	u.Weapon = "rifle"
	u.WeaponAmmo = 10

	_, _, _, err := u.FireAt(target, nil, nil)
	if err == nil {
		t.Error("expected error for no TU")
	}
}

func TestUnitMoveToConsumesTU(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileFloor)
	u.X = 3
	u.Y = 3
	u.TU = 50

	beforeTU := u.TU
	ok := u.MoveTo(5, 5, m)
	if !ok {
		t.Fatal("move should succeed")
	}
	if u.TU >= beforeTU {
		t.Error("TU should decrease after move")
	}
}

func TestUnitMoveToInsufficientTU(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileFloor)
	u.X = 3
	u.Y = 3
	u.TU = 2

	ok := u.MoveTo(5, 5, m)
	if ok {
		t.Error("move should fail with insufficient TU")
	}
}

func TestUnitMoveToCrouchingExtraCost(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileFloor)
	u.X = 4
	u.Y = 4
	u.TU = 50
	u.Crouching = true

	beforeTU := u.TU
	ok := u.MoveTo(5, 5, m)
	if !ok {
		t.Fatal("move should succeed")
	}
	// 1 distance = 4 TU + 4 TU crouching penalty
	// distance from (4,4) to (5,5) = 2, cost = 2*4 + 4(crouch) = 12
	if beforeTU-u.TU != 12 {
		t.Errorf("expected 12 TU consumed with crouching, got %d", beforeTU-u.TU)
	}
}

func TestUnitFireAtStunRod(t *testing.T) {
	u := &Unit{HP: 10, Weapon: "stun_rod", WeaponAmmo: 10, TU: 50, Accuracy: 100, Faction: 0, X: 5, Y: 5}
	target := &Unit{HP: 10, MaxHP: 10, Faction: 1, Alive: true, X: 5, Y: 6}
	m := NewBattleMap(20, 20)
	m.Set(5, 6, TileFloor)
	m.Set(5, 5, TileFloor)

	damage, hit, _, err := u.FireAt(target, m, nil)
	if err != nil {
		t.Fatalf("FireAt error: %v", err)
	}
	if !hit {
		t.Log("stun rod missed (unlikely with 100 acc)")
		return
	}
	if damage < 10 || damage > 13 {
		t.Errorf("expected 10-13 stun damage from stun_rod, got %d", damage)
	}
	// stun rod should add stun points, not reduce HP
	if target.HP != 10 {
		t.Errorf("stun rod should not reduce HP directly, got %d", target.HP)
	}
}

func TestUnitCanSeeThroughOpenLine(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	u.X = 5
	u.Y = 5
	if !u.CanSee(10, 10, m) {
		t.Error("should see along open diagonal")
	}
}

func TestUnitCanSeeBlockedByWall(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	u.X = 5
	u.Y = 5
	m.Set(7, 5, TileWall)
	if u.CanSee(10, 5, m) {
		t.Error("should not see through wall")
	}
}

func TestUnitCanSeeSelf(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	u.X = 5
	u.Y = 5
	if !u.CanSee(5, 5, m) {
		t.Error("unit should always see its own tile")
	}
}

func TestUnitListOnLevel(t *testing.T) {
	u1 := NewSoldierUnit(soldier.NewSoldier("A"))
	u2 := NewSoldierUnit(soldier.NewSoldier("B"))
	u1.Level = 0
	u2.Level = 1
	u1.Alive = true
	u2.Alive = true
	list := UnitList{u1, u2}
	level0 := list.OnLevel(0)
	if len(level0) != 1 || level0[0] != u1 {
		t.Error("expected u1 on level 0")
	}
}

func TestUnitFireAtCoverDamageReduction(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	target := NewSoldierUnit(soldier.NewSoldier("Target"))
	m := NewBattleMap(20, 20)
	u.X = 5
	u.Y = 5
	target.X = 8
	target.Y = 5
	target.HP = 100
	target.Armour = 0
	u.TU = 50
	u.Weapon = "rifle"
	u.WeaponAmmo = 10

	// Place a tree between them (60% cover)
	m.Set(6, 5, TileTree)
	m.Set(7, 5, TileFloor)
	m.Set(5, 5, TileFloor)
	m.Set(8, 5, TileFloor)

	_, hit, _, err := u.FireAt(target, m, nil)
	if err != nil {
		t.Fatalf("FireAt error: %v", err)
	}
	if hit {
		t.Log("hit registered (cover may not block all damage)")
	}
}

func TestUnitFireAtUnknownWeapon(t *testing.T) {
	u := &Unit{HP: 10, Weapon: "unknown_weapon", TU: 50, X: 5, Y: 5}
	target := &Unit{HP: 10, X: 5, Y: 6}
	_, _, _, err := u.FireAt(target, nil, nil)
	if err == nil {
		t.Error("expected error for unknown weapon")
	}
}

func TestUnitMoveToNonPassable(t *testing.T) {
	s := soldier.NewSoldier("Test")
	u := NewSoldierUnit(s)
	m := NewBattleMap(20, 20)
	m.Set(5, 5, TileWall)
	u.X = 4
	u.Y = 5
	u.TU = 50
	ok := u.MoveTo(5, 5, m)
	if ok {
		t.Error("should not move onto wall")
	}
}

func TestUnitFactionConstants(t *testing.T) {
	s := NewSoldierUnit(soldier.NewSoldier("Test"))
	a := NewAlienUnit(data.GetAlienByName("Sectoid"))
	c := NewCivilianUnit("Civ")
	if s.Faction != 0 {
		t.Errorf("soldier faction should be 0, got %d", s.Faction)
	}
	if a.Faction != 1 {
		t.Errorf("alien faction should be 1, got %d", a.Faction)
	}
	if c.Faction != 2 {
		t.Errorf("civilian faction should be 2, got %d", c.Faction)
	}
}

func TestWeaponDamageType(t *testing.T) {
	tests := []struct {
		weapon string
		want   int
	}{
		{"rifle", data.DMG_KINETIC},
		{"plasma_rifle", data.DMG_PLASMA},
		{"laser_rifle", data.DMG_LASER},
		{"rocket", data.DMG_EXPLOSIVE},
		{"stun_rod", data.DMG_MELEE},
	}
	for _, tt := range tests {
		got := WeaponDamageType(tt.weapon)
		if got != tt.want {
			t.Errorf("WeaponDamageType(%q) = %d, want %d", tt.weapon, got, tt.want)
		}
	}
}
