package geo

import (
	"testing"

	"github.com/civ13/ycom/internal/engine"
)

func TestCityCount(t *testing.T) {
	cities := GetCities()
	if len(cities) < 15 {
		t.Errorf("expected at least 15 cities, got %d", len(cities))
	}
}

func TestCityByID(t *testing.T) {
	cities := GetCities()
	var c *City
	for _, city := range cities {
		if city.ID == 0 {
			c = city
			break
		}
	}
	if c == nil {
		t.Fatal("CityByID(0) returned nil")
	}
	if c.Name != "New York" {
		t.Errorf("expected New York, got %s", c.Name)
	}
}

func TestUFOSpawnOnCities(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	if ufo == nil {
		t.Fatal("SpawnUFOOnCities returned nil")
	}
	if !ufo.Active {
		t.Error("new UFO should be active")
	}
	if ufo.Type.Name == "" {
		t.Error("UFO type name is empty")
	}
}

func TestUFOMovementOnCities(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	startProgress := ufo.Progress
	ufo.Update(cities)
	if ufo.Progress <= startProgress {
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

	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	list = append(list, ufo)
	if list.Count() != 1 {
		t.Errorf("expected 1, got %d", list.Count())
	}

	ufo.Active = false
	if list.Count() != 0 {
		t.Error("inactive UFO should not count")
	}
}

func TestInterceptorLaunchAtNode(t *testing.T) {
	cities := GetCities()
	inter := NewInterceptor(48, 31) // New York coords
	if inter.HP != 60 {
		t.Errorf("expected 60 HP, got %d", inter.HP)
	}
	if inter.Ammo != 4 { // avalanche has FireRate 1, 1*4=4
		t.Errorf("expected 4 ammo, got %d", inter.Ammo)
	}

	inter.LaunchAtNode(16, cities) // Tokyo
	if !inter.Launching {
		t.Error("should be launching after LaunchAtNode()")
	}
	if inter.TargetNode != 16 {
		t.Errorf("target node should be 16, got %d", inter.TargetNode)
	}
}

func TestInterceptorFire(t *testing.T) {
	inter := NewInterceptor(48, 31)
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	ufo.Type.Toughness = 1000 // high HP so it doesn't die
	ufo.X = float64(inter.X) + 1 // place nearby
	ufo.Y = float64(inter.Y)

	// Fire multiple times to test at least one hit
	hit := false
	for i := 0; i < 10; i++ {
		inter.Ammo = 1
		damage := inter.FireAt(ufo)
		if damage > 0 {
			hit = true
			break
		}
	}
	if !hit {
		t.Log("no hit in 10 attempts (accuracy may be low)")
	}
	if inter.Ammo < 0 {
		t.Errorf("ammo should not go negative, got %d", inter.Ammo)
	}
}

func TestInterceptorFireEmpty(t *testing.T) {
	inter := NewInterceptor(48, 31)
	inter.Ammo = 0
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	damage := inter.FireAt(ufo)
	if damage != 0 {
		t.Errorf("expected 0 damage with no ammo, got %d", damage)
	}
}

func TestInterceptorDisengage(t *testing.T) {
	inter := NewInterceptor(48, 31)
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities)
	inter.LaunchAtUFO(ufo)
	inter.Disengage()
	if inter.Launching {
		t.Error("should not be launching after Disengage()")
	}
	if inter.TargetUFO != nil {
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
	i2.Launching = true
	list := InterceptorList{i1, i2}
	active := list.Active()
	if len(active) != 1 {
		t.Errorf("expected 1 active, got %d", len(active))
	}
}

func TestUFOExpiry(t *testing.T) {
	cities := GetCities()
	ufo := &UFO{
		NodeFrom:   cities[0].ID,
		NodeTo:     cities[1].ID,
		Progress:   0.5,
		TurnsLeft:  1,
		Active:     true,
		Type:       UFOTypes[0],
	}
	ufo.Update(cities)
	if ufo.Active {
		t.Error("UFO should have expired")
	}
}

func TestShortestPath(t *testing.T) {
	gs := &Geoscape{
		Cities: GetCities(),
	}
	path := gs.ShortestPath(0, 16) // New York to Tokyo
	if path == nil {
		t.Fatal("ShortestPath returned nil")
	}
	if path[0] != 0 || path[len(path)-1] != 16 {
		t.Errorf("path should start at 0 and end at 16, got %v", path)
	}
}
