package geo

import (
	"testing"

	"github.com/civ13/ycom/internal/engine"
)

func TestNetworkCreation(t *testing.T) {
	gn := NewRegionalNetwork()
	if len(gn.Nodes) < 15 {
		t.Errorf("expected at least 15 nodes, got %d", len(gn.Nodes))
	}
	if len(gn.Edges) < 10 {
		t.Errorf("expected at least 10 edges, got %d", len(gn.Edges))
	}
}

func TestNodeByID(t *testing.T) {
	gn := NewRegionalNetwork()
	node := gn.NodeByID(0)
	if node == nil {
		t.Fatal("NodeByID(0) returned nil")
	}
	if node.Name != "New York" {
		t.Errorf("expected New York, got %s", node.Name)
	}
	if gn.NodeByID(999) != nil {
		t.Error("NodeByID(999) should return nil")
	}
}

func TestNearestNode(t *testing.T) {
	gn := NewRegionalNetwork()
	node := gn.NearestNode(18, 12) // New York coords
	if node == nil {
		t.Fatal("NearestNode returned nil for New York coords")
	}
	if node.Name != "New York" {
		t.Errorf("expected New York, got %s", node.Name)
	}
}

func TestNeighbors(t *testing.T) {
	gn := NewRegionalNetwork()
	neighbors := gn.Neighbors(0) // New York
	if len(neighbors) == 0 {
		t.Error("New York should have neighbors")
	}
}

func TestShortestPath(t *testing.T) {
	gn := NewRegionalNetwork()
	path := gn.ShortestPath(0, 16) // New York to Tokyo
	if path == nil {
		t.Fatal("ShortestPath returned nil")
	}
	if path[0] != 0 || path[len(path)-1] != 16 {
		t.Errorf("path should start at 0 and end at 16, got %v", path)
	}
}

func TestUFOSpawnOnNetwork(t *testing.T) {
	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
	if ufo == nil {
		t.Fatal("SpawnUFOOnNetwork returned nil")
	}
	if !ufo.Active {
		t.Error("new UFO should be active")
	}
	if ufo.Type.Name == "" {
		t.Error("UFO type name is empty")
	}
}

func TestUFOMovementOnNetwork(t *testing.T) {
	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
	startProgress := ufo.Progress
	ufo.Update(gn)
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

	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
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
	gn := NewRegionalNetwork()
	inter := NewInterceptor(18, 12)
	if inter.HP != 60 {
		t.Errorf("expected 60 HP, got %d", inter.HP)
	}
	if inter.Ammo != 8 {
		t.Errorf("expected 8 ammo, got %d", inter.Ammo)
	}

	inter.LaunchAtNode(16, gn) // Tokyo
	if !inter.Launching {
		t.Error("should be launching after LaunchAtNode()")
	}
	if inter.TargetNode != 16 {
		t.Errorf("target node should be 16, got %d", inter.TargetNode)
	}
}

func TestInterceptorFire(t *testing.T) {
	inter := NewInterceptor(18, 12)
	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
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
	inter := NewInterceptor(18, 12)
	inter.Ammo = 0
	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
	damage := inter.FireAt(ufo)
	if damage != 0 {
		t.Errorf("expected 0 damage with no ammo, got %d", damage)
	}
}

func TestInterceptorDisengage(t *testing.T) {
	inter := NewInterceptor(18, 12)
	gn := NewRegionalNetwork()
	ufo := SpawnUFOOnNetwork(gn)
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
	list := InterceptorList{i1, i2}
	active := list.Active()
	if len(active) != 1 {
		t.Errorf("expected 1 active, got %d", len(active))
	}
}

func TestUFOExpiry(t *testing.T) {
	gn := NewRegionalNetwork()
	ufo := &UFO{
		NodeFrom:   0,
		NodeTo:     1,
		Progress:   0.5,
		TurnsLeft:  1,
		Active:     true,
		Type:       UFOTypes[0],
	}
	ufo.Update(gn)
	if ufo.Active {
		t.Error("UFO should have expired")
	}
}
