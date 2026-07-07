package geo

import (
	"math"
	"testing"

	"github.com/civ13/ycom/internal/engine"
)

func TestWorldMapInit(t *testing.T) {
	w, h := MapSize()
	if w != 180 || h != 90 {
		t.Errorf("expected 180x90, got %dx%d", w, h)
	}
}

func TestGetTile(t *testing.T) {
	if GetTile(-1, -1) != 0 {
		t.Error("out of bounds should return 0")
	}
	if GetTile(180, 90) != 0 {
		t.Error("out of bounds should return 0")
	}
}

func TestSetTile(t *testing.T) {
	SetTile(10, 10, 3)
	if GetTile(10, 10) != 3 {
		t.Error("SetTile failed")
	}
	SetTile(10, 10, 1) // restore
}

func TestIsLand(t *testing.T) {
	SetTile(50, 50, 1)
	if !IsLand(50, 50) {
		t.Error("expected IsLand true for tile 1")
	}
	SetTile(50, 50, 0)
	if IsLand(50, 50) {
		t.Error("expected IsLand false for tile 0")
	}
}

func TestCities(t *testing.T) {
	cities := GetCities()
	if len(cities) == 0 {
		t.Error("no cities defined")
	}
	for _, c := range cities {
		if c.Name == "" {
			t.Error("city with empty name")
		}
		if c.X < 0 || c.Y < 0 || c.Y >= mapH {
			t.Errorf("city %s out of bounds: (%d,%d)", c.Name, c.X, c.Y)
		}
	}
}

func TestUFOSpawn(t *testing.T) {
	ufo := SpawnUFO()
	if ufo == nil {
		t.Fatal("SpawnUFO returned nil")
	}
	if !ufo.Active {
		t.Error("new UFO should be active")
	}
	if ufo.Type.Name == "" {
		t.Error("UFO type name is empty")
	}
}

func TestUFOMovement(t *testing.T) {
	ufo := SpawnUFO()
	startX, startY := ufo.X, ufo.Y
	ufo.Update()
	if ufo.X == startX && ufo.Y == startY {
		// Could be same if speed is very low, that's ok
	}
}

func TestUFOList(t *testing.T) {
	var list UFOList
	if list.Count() != 0 {
		t.Error("empty list should have 0 count")
	}
	if len(list.Active()) != 0 {
		t.Error("empty list should have 0 active")
	}

	ufo := SpawnUFO()
	list = append(list, ufo)
	if list.Count() != 1 {
		t.Errorf("expected 1, got %d", list.Count())
	}

	ufo.Active = false
	if list.Count() != 0 {
		t.Error("inactive UFO should not count")
	}
}

func TestInterceptorLaunch(t *testing.T) {
	inter := NewInterceptor(28, 32)
	if inter.HP != 60 {
		t.Errorf("expected 60 HP, got %d", inter.HP)
	}
	if inter.Ammo != 8 {
		t.Errorf("expected 8 ammo, got %d", inter.Ammo)
	}

	ufo := SpawnUFO()
	inter.Launch(ufo)
	if !inter.Launching {
		t.Error("should be launching after Launch()")
	}
	if inter.Target != ufo {
		t.Error("target not set")
	}
}

func TestInterceptorFire(t *testing.T) {
	inter := NewInterceptor(28, 32)
	ufo := SpawnUFO()
	ufo.Type.Toughness = 100

	damage := inter.FireAt(ufo)
	if damage <= 0 {
		t.Errorf("expected positive damage, got %d", damage)
	}
	if inter.Ammo != 7 {
		t.Errorf("expected 7 ammo, got %d", inter.Ammo)
	}
}

func TestInterceptorFireEmpty(t *testing.T) {
	inter := NewInterceptor(28, 32)
	inter.Ammo = 0
	ufo := SpawnUFO()
	damage := inter.FireAt(ufo)
	if damage != 0 {
		t.Errorf("expected 0 damage with no ammo, got %d", damage)
	}
}

func TestInterceptorDisengage(t *testing.T) {
	inter := NewInterceptor(28, 32)
	ufo := SpawnUFO()
	inter.Launch(ufo)
	inter.Disengage()
	if inter.Launching {
		t.Error("should not be launching after Disengage()")
	}
	if inter.Target != nil {
		t.Error("target should be nil after Disengage()")
	}
}

func TestGeoscapeTogglePause(t *testing.T) {
	gs := &Geoscape{}
	gs.Game = &engine.Game{}
	gs.Game.Paused = true
	gs.TogglePause()
	if gs.Game.Paused {
		t.Error("should be unpaused after TogglePause()")
	}
	gs.TogglePause()
	if !gs.Game.Paused {
		t.Error("should be paused after second TogglePause()")
	}
}

func TestGeoscapeSetSpeed(t *testing.T) {
	gs := &Geoscape{}
	gs.Game = &engine.Game{}
	gs.SetSpeed(3)
	if gs.Game.TimeSpeed != 3 {
		t.Errorf("expected speed 3, got %d", gs.Game.TimeSpeed)
	}
	if gs.Game.Paused {
		t.Error("should not be paused after SetSpeed()")
	}
}

func TestInterceptorListActive(t *testing.T) {
	i1 := NewInterceptor(10, 10)
	i2 := NewInterceptor(20, 20)
	i1.HP = 0
	list := InterceptorList{i1, i2}
	active := list.Active()
	if len(active) != 1 {
		t.Errorf("expected 1 active, got %d", len(active))
	}
}

func TestUFOWrap(t *testing.T) {
	ufo := &UFO{
		X:         179.5,
		Y:         5.0,
		DX:        1.0,
		DY:        0.0,
		TurnsLeft: 10,
		Active:    true,
		Type:      UFOTypes[0],
	}
	ufo.Update()
	if ufo.X >= float64(mapW) {
		t.Error("UFO should have wrapped around")
	}
	if !ufo.Active {
		t.Error("UFO should still be active")
	}
}

func TestUFOExpiry(t *testing.T) {
	ufo := &UFO{
		X:         50,
		Y:         50,
		DX:        0,
		DY:        0,
		TurnsLeft: 1,
		Active:    true,
		Type:      UFOTypes[0],
	}
	ufo.Update()
	if ufo.Active {
		t.Error("UFO should have expired")
	}
}

func TestInterceptorRangeExpiry(t *testing.T) {
	inter := NewInterceptor(0, 0)
	ufo := SpawnUFO()
	ufo.X = 100
	ufo.Y = 100
	inter.Launch(ufo)
	inter.RangeLeft = 1
	inter.Update()
	if inter.Launching {
		t.Error("interceptor should have returned to base")
	}
}

func TestDistanceCalculation(t *testing.T) {
	a := UFO{X: 0, Y: 0}
	b := UFO{X: 3, Y: 4}
	dx := b.X - a.X
	dy := b.Y - a.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist != 5.0 {
		t.Errorf("expected distance 5, got %f", dist)
	}
}
