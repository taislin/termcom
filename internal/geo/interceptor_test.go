package geo

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
)

func TestNewInterceptor(t *testing.T) {
	inter := NewInterceptor(48, 31)
	if inter.HP != 60 {
		t.Errorf("expected 60 HP, got %d", inter.HP)
	}
	if inter.Ammo != 4 {
		t.Errorf("expected 4 ammo, got %d", inter.Ammo)
	}
	if inter.Speed != 36 {
		t.Errorf("expected speed 36, got %d", inter.Speed)
	}
}

func TestSetWeapon(t *testing.T) {
	inter := NewInterceptor(48, 31)
	inter.SetWeapon("cannon")
	if inter.WeaponKey != "cannon" {
		t.Errorf("expected cannon, got %s", inter.WeaponKey)
	}
}

func TestSetWeaponInvalid(t *testing.T) {
	inter := NewInterceptor(48, 31)
	original := inter.WeaponKey
	inter.SetWeapon("invalid_weapon")
	if inter.WeaponKey != original {
		t.Error("weapon should not change for invalid key")
	}
}

func TestSetMode(t *testing.T) {
	inter := NewInterceptor(48, 31)
	inter.SetMode(1) // CombatAttack
	if inter.Mode != 1 {
		t.Errorf("expected mode 1, got %d", inter.Mode)
	}
	inter.SetMode(0) // CombatCautious
	if inter.Mode != 0 {
		t.Errorf("expected mode 0, got %d", inter.Mode)
	}
}

func TestEffectiveAccuracy(t *testing.T) {
	inter := NewInterceptor(48, 31)
	acc := inter.EffectiveAccuracy()
	if acc != 55 {
		t.Errorf("expected accuracy 55 (avalanche base), got %d", acc)
	}
	inter.PilotSkill = 100
	acc = inter.EffectiveAccuracy()
	// avalanche accuracy=55 + (100-50)/5 = 55+10 = 65
	if acc != 65 {
		t.Errorf("expected accuracy 65 (55 base + 10 pilot), got %d", acc)
	}
}

func TestLaunchAtNode(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	inter.LaunchAtNode(cities[0].ID, cities)
	if !inter.Launching {
		t.Error("should be launching")
	}
	if inter.TargetNode != cities[0].ID {
		t.Errorf("expected target node %d, got %d", cities[0].ID, inter.TargetNode)
	}
	if inter.RangeLeft <= 0 {
		t.Error("RangeLeft should be positive after launch")
	}
}

func TestLaunchAtInvalidNode(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	inter.LaunchAtNode(9999, cities)
	if inter.Launching {
		t.Error("should not launch at invalid node")
	}
}

func TestLaunchAtUFO(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	inter := NewInterceptor(48, 31)
	inter.LaunchAtUFO(ufo)
	if !inter.Launching {
		t.Error("should be launching")
	}
	if inter.TargetUFO != ufo {
		t.Error("TargetUFO should be set")
	}
	if inter.TargetNode != -1 {
		t.Error("TargetNode should be -1 when pursuing UFO")
	}
}

func TestIntFireAtUFO(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	ufo := SpawnUFOOnCities(cities, 0)
	ufo.Type.Toughness = 1000
	ufo.Type.MaxHP = 1000
	ufo.X = 50
	ufo.Y = 31

	beforeAmmo := inter.Ammo
	damage := inter.FireAt(ufo)
	if inter.Ammo >= beforeAmmo {
		t.Error("ammo should decrease after firing")
	}
	if damage > 0 {
		t.Log("hit registered")
	}
}

func TestInterceptorFireAtUFONoAmmo(t *testing.T) {
	inter := NewInterceptor(48, 31)
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	inter.Ammo = 0
	damage := inter.FireAt(ufo)
	if damage != 0 {
		t.Errorf("expected 0 damage with no ammo, got %d", damage)
	}
}

func TestInterceptorUpdateChasesUFO(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	ufo := SpawnUFOOnCities(cities, 0)
	ufo.X = 50
	ufo.Y = 40
	inter.LaunchAtUFO(ufo)
	inter.X = 48
	inter.Y = 31
	inter.Mode = 1 // CombatAttack
	reached := inter.Update(cities, UFOList{ufo})
	_ = reached
}

func TestInterceptorUpdateChasesNode(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	inter.LaunchAtNode(cities[5].ID, cities)
	reached := inter.Update(cities, UFOList{})
	_ = reached
}

func TestInterceptorUpdateTargetUFODestroyed(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31)
	ufo := SpawnUFOOnCities(cities, 0)
	inter.LaunchAtUFO(ufo)
	ufo.Active = false
	reached := inter.Update(cities, UFOList{ufo})
	if reached {
		t.Error("should not reach destroyed UFO")
	}
	if inter.Launching {
		t.Error("should not be launching after target destroyed")
	}
}

func TestInterceptorUpdateNoTarget(t *testing.T) {
	inter := NewInterceptor(48, 31)
	reached := inter.Update(GetCities(), UFOList{})
	if reached {
		t.Error("should not reach with no target")
	}
}

func TestInterceptorSetWeaponRearms(t *testing.T) {
	inter := NewInterceptor(48, 31)
	inter.Ammo = 0
	inter.SetWeapon("avalanche")
	if inter.Ammo != 4 {
		t.Errorf("expected 4 ammo after rearming, got %d", inter.Ammo)
	}
}

func TestNewInterceptorFromState(t *testing.T) {
	s := &data.InterceptorState{
		Name: "TestCraft", HP: 50, MaxHP: 60, WeaponKey: "avalanche", Ammo: 10,
	}
	inter := NewInterceptorFromState(s, 48, 31)
	if inter.Name != "TestCraft" {
		t.Errorf("expected TestCraft, got %s", inter.Name)
	}
	if inter.HP != 50 {
		t.Errorf("expected 50 HP, got %d", inter.HP)
	}
	if inter.Ammo != 10 {
		t.Errorf("expected 10 ammo, got %d", inter.Ammo)
	}
}
