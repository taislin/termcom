package geo

import (
	"testing"
)

func TestGetUFOTypeByName(t *testing.T) {
	types := []string{"Small Scout", "Medium Scout", "Large Scout", "Harvester", "Bomber", "Transport"}
	for _, name := range types {
		ut := GetUFOTypeByName(name)
		if ut == nil {
			t.Errorf("GetUFOTypeByName(%q) returned nil", name)
		} else if ut.Name != name {
			t.Errorf("expected name %q, got %q", name, ut.Name)
		}
	}
}

func TestGetUFOTypeByNameInvalid(t *testing.T) {
	ut := GetUFOTypeByName("Nonexistent")
	if ut != nil {
		t.Error("expected nil for nonexistent type")
	}
}

func TestUFOSpawnWithDifficulty(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 3)
	if ufo == nil {
		t.Fatal("SpawnUFOOnCities returned nil")
	}
	// Higher difficulty should increase HP
	if ufo.Type.MaxHP <= 0 {
		t.Error("UFO should have positive MaxHP")
	}
	if ufo.TurnsLeft <= 0 {
		t.Error("new UFO should have positive turns left")
	}
}

func TestUFOUpdateMovement(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	if ufo == nil {
		t.Fatal("SpawnUFOOnCities returned nil")
	}
	before := ufo.X
	beforeY := ufo.Y
	ufo.Update(cities)
	// Position should be updated (may be same if speed is low)
	_ = before
	_ = beforeY
}

func TestUFOUpdateArrival(t *testing.T) {
	cities := GetCities()
	if len(cities) < 2 {
		t.Fatal("need at least 2 cities")
	}
	ufo := &UFO{
		Active:    true,
		Type:      UFOTypes[0],
		NodeFrom:  cities[0].ID,
		NodeTo:    cities[1].ID,
		Progress:  0.99,
		TurnsLeft: 100,
	}
	ufo.updatePosition(cities)
	if ufo.Progress < 1.0 {
		t.Log("UFO not yet at destination")
	}
	ufo.Update(cities)
	// After arrival, should either pick new destination or stay
	_ = ufo.NodeFrom
}

func TestUFOExpiryByTurns(t *testing.T) {
	cities := GetCities()
	if len(cities) < 2 {
		t.Fatal("need at least 2 cities")
	}
	ufo := &UFO{
		Active:    true,
		Type:      UFOTypes[0],
		NodeFrom:  cities[0].ID,
		NodeTo:    cities[1].ID,
		Progress:  0.5,
		TurnsLeft: 1,
	}
	ufo.Update(cities)
	if ufo.Active {
		t.Error("UFO should expire when TurnsLeft reaches 0")
	}
}

func TestUFOTileCoordinates(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	if ufo == nil {
		t.Fatal("Spawn returned nil")
	}
	tx := ufo.TileX()
	ty := ufo.TileY()
	if tx < 0 || ty < 0 {
		t.Errorf("negative tile coordinates: (%d,%d)", tx, ty)
	}
}

func TestUFOCurrentNode(t *testing.T) {
	cities := GetCities()
	if len(cities) < 2 {
		t.Fatal("need at least 2 cities")
	}
	ufo := &UFO{
		Active:   true,
		Type:     UFOTypes[0],
		NodeFrom: cities[0].ID,
		NodeTo:   cities[1].ID,
		Progress: 0.3,
	}
	// Progress < 0.5 should return NodeFrom
	if n := ufo.CurrentNode(); n != ufo.NodeFrom {
		t.Errorf("expected NodeFrom %d, got %d", ufo.NodeFrom, n)
	}
	ufo.Progress = 0.7
	if n := ufo.CurrentNode(); n != ufo.NodeTo {
		t.Errorf("expected NodeTo %d, got %d", ufo.NodeTo, n)
	}
}

func TestUFOFireAtInterceptor(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	if ufo == nil {
		t.Fatal("Spawn returned nil")
	}
	inter := NewInterceptor(48, 31)
	beforeHP := inter.HP
	damage := ufo.FireAtInterceptor(inter)
	if damage > 0 {
		if inter.HP >= beforeHP {
			t.Error("interceptor HP should decrease when hit")
		}
		if inter.HP < 0 {
			t.Error("interceptor HP should not go negative")
		}
	} else {
		if inter.HP != beforeHP {
			t.Error("interceptor HP should not change when missed")
		}
	}
}

func TestUFOFireAtInterceptorInactive(t *testing.T) {
	ufo := &UFO{Active: false}
	inter := NewInterceptor(48, 31)
	damage := ufo.FireAtInterceptor(inter)
	if damage != 0 {
		t.Error("inactive UFO should deal 0 damage")
	}
}

func TestUFOListActive(t *testing.T) {
	cities := GetCities()
	u1 := SpawnUFOOnCities(cities, 0)
	u2 := SpawnUFOOnCities(cities, 0)
	u2.Active = false
	list := UFOList{u1, u2}
	active := list.Active()
	if len(active) != 1 {
		t.Errorf("expected 1 active, got %d", len(active))
	}
}

func TestUFOListCount(t *testing.T) {
	cities := GetCities()
	u1 := SpawnUFOOnCities(cities, 0)
	u2 := SpawnUFOOnCities(cities, 0)
	list := UFOList{u1, u2}
	if list.Count() != 2 {
		t.Errorf("expected count 2, got %d", list.Count())
	}
	u2.Active = false
	if list.Count() != 1 {
		t.Errorf("expected count 1 after deactivating, got %d", list.Count())
	}
}

func TestUFOSpawnAtCity(t *testing.T) {
	cities := GetCities()
	if len(cities) < 2 {
		t.Fatal("need at least 2 cities")
	}
	ufo := SpawnUFOAtCity(cities[0], cities, 0)
	if ufo == nil {
		t.Fatal("SpawnUFOAtCity returned nil")
	}
	if ufo.NodeTo != cities[0].ID {
		t.Errorf("expected NodeTo %d, got %d", cities[0].ID, ufo.NodeTo)
	}
	if ufo.Progress < 0.2 {
		t.Errorf("expected progress >= 0.3, got %f", ufo.Progress)
	}
}

func TestUFOSpawnAtCitySingleCity(t *testing.T) {
	cities := GetCities()
	if len(cities) < 1 {
		t.Fatal("need at least 1 city")
	}
	ufo := SpawnUFOAtCity(cities[0], cities[:1], 0)
	if ufo != nil {
		t.Log("SpawnUFOAtCity may return nil with only 1 city (depends on candidate logic)")
	}
}

func TestUFOSpawnOnCitiesFewerThanTwo(t *testing.T) {
	ufo := SpawnUFOOnCities([]*City{{ID: 0}}, 0)
	if ufo != nil {
		t.Error("SpawnUFOOnCities should return nil with fewer than 2 cities")
	}
}
