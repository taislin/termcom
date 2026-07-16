package battle

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
)

func TestNewAlienAI(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	ai := NewAlienAI(u)
	if ai.State != AIPatrol {
		t.Errorf("expected AIPatrol, got %d", ai.State)
	}
	if ai.Unit != u {
		t.Error("alienAI should reference the unit")
	}
}

func TestAIUpdateDeadUnit(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	u.Alive = false
	ai := NewAlienAI(u)
	actions := ai.Update(nil, nil, nil, nil, nil)
	if len(actions) != 0 {
		t.Error("dead unit should produce no actions")
	}
}

func TestAIUpdatePatrolNoTarget(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	m := NewBattleMap(30, 20)
	m.Set(15, 10, TileFloor)
	u.X = 15
	u.Y = 10
	u.TU = 50
	ai := NewAlienAI(u)
	actions := ai.Update(UnitList{u}, m, UnitList{}, nil, nil)
	if len(actions) != 1 || actions[0].Type != "move" {
		t.Errorf("expected 1 move action, got %d actions", len(actions))
	}
}

func TestAIUpdatePatrolToAttack(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 16, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	m.Set(15, 10, TileFloor)
	m.Set(16, 10, TileFloor)
	alien.X = 15
	alien.Y = 10
	alien.TU = 50
	m.Set(15, 10, TileFloor)
	m.Set(16, 10, TileFloor)
	ai := NewAlienAI(alien)
	ai.State = AIPatrol
	actions := ai.Update(UnitList{alien}, m, UnitList{human}, nil, nil)
	if ai.State != AIAttack {
		t.Errorf("expected AIAttack, got %d", ai.State)
	}
	if len(actions) > 0 && actions[0].Type != "fire" && actions[0].Type != "melee" {
		t.Errorf("expected fire/melee action, got %s", actions[0].Type)
	}
}

func TestAIEvaluateCover(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	ai := NewAlienAI(u)
	m := NewBattleMap(30, 20)
	m.Set(10, 10, TileWall)
	// wall has cover > 0
	cover := ai.evaluateCover(10, 10, m)
	if cover <= 0 {
		t.Error("wall should provide cover > 0")
	}
	// grass has cover = 0
	m.Set(5, 5, TileGrass)
	cover = ai.evaluateCover(5, 5, m)
	if cover != 0 {
		t.Errorf("grass should provide 0 cover, got %d", cover)
	}
}

func TestAICanFireAt(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	u := NewAlienUnit(at)
	u.TU = 50
	u.WeaponAmmo = 10
	ai := NewAlienAI(u)

	// should be able to fire
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	if !ai.canFireAt(human) {
		t.Error("should be able to fire with TU and ammo")
	}

	// no TU
	u.TU = 0
	if ai.canFireAt(human) {
		t.Error("should not fire with 0 TU")
	}

	// no ammo (for weapons with limited ammo like pistol with AmmoMax=6)
	u.TU = 50
	u.Weapon = "rifle"
	u.WeaponAmmo = 0
	if ai.canFireAt(human) {
		t.Error("should not fire with 0 ammo")
	}
	u.Weapon = at.Weapon // restore
	u.WeaponAmmo = 10
}

func TestAIFindFlankPosition(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	human.X = 15
	human.Y = 10
	human.Faction = 0
	// Clear a path
	for x := 5; x < 19; x++ {
		for y := 5; y < 15; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	fx, fy := ai.findFlankPosition(human, human, m, UnitList{alien})
	if fx == alien.X && fy == alien.Y {
		t.Error("flank position should be different from current position")
	}
}

func TestAIRetreatTarget(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 15
	alien.Y = 10
	alien.TU = 50
	human.X = 10
	human.Y = 10
	human.Faction = 0
	// Clear area
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	rx, _ := ai.retreatTarget(human, m, UnitList{alien, human})
	// Should retreat away from threat (threat is at x=10, we're at 15, so retreat should go x > 15)
	if rx <= alien.X {
		t.Errorf("retreat should move away from threat, rx=%d, alienX=%d, threatX=%d", rx, alien.X, human.X)
	}
}

func TestAIAdvanceToward(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	ax, ay := ai.advanceToward(20, 10, m, UnitList{alien})
	if ax <= alien.X {
		t.Errorf("advance should move toward target, ax=%d, alienX=%d", ax, alien.X)
	}
	if ay != alien.Y {
		t.Errorf("advance should stay on same row for horizontal target, ay=%d", ay)
	}
}

func TestAIFindNearest(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human1 := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	human2 := &Unit{X: 20, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	ai := NewAlienAI(alien)
	nearest, dist := ai.findNearest(UnitList{human1, human2}, m)
	if nearest != human1 {
		t.Error("nearest should be human1 at distance 2")
	}
	if dist != 2.0 {
		t.Errorf("expected distance 2.0, got %f", dist)
	}
}

func TestAISelectTarget(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	human.X = 12
	human.Y = 10
	human.Faction = 0
	ai := NewAlienAI(alien)
	target := ai.selectTarget(human, UnitList{human}, nil, m)
	if target != human {
		t.Error("should select the nearest human")
	}
}

func TestAIUpdateFleeFromLowHP(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	alien.HP = 1
	alien.MaxHP = 10
	human.X = 12
	human.Y = 10
	human.Faction = 0
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	ai.State = AIAttack
	actions := ai.Update(UnitList{alien}, m, UnitList{human}, nil, nil)
	// Should transition to flee/retreat
	if ai.State != AIRetreat {
		t.Errorf("expected AIRetreat for low HP alien, got %d", ai.State)
	}
	if len(actions) == 0 {
		t.Error("should produce retreat actions")
	}
}

func TestAISuppressStateTransition(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	ai.State = AISuppress
	ai.TurnsSince = 4
	// No target visible → should transition back to attack
	actions := ai.Update(UnitList{alien}, m, UnitList{}, nil, nil)
	if ai.State != AIAttack {
		t.Errorf("expected AIAttack after suppress expires, got %d", ai.State)
	}
	if len(actions) != 0 {
		t.Error("no actions expected with no target")
	}
}

func TestAISearchStateMoveTowardLastSeen(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	ai.LastSeenX = 5
	ai.LastSeenY = 5
	ai.State = AISearch

	// No human to detect
	actions := ai.Update(UnitList{alien}, m, UnitList{}, nil, nil)
	if len(actions) != 1 || actions[0].Type != "move" {
		t.Errorf("expected move action toward last seen, got %d actions", len(actions))
	}
	// After many turns without finding, should return to patrol
	ai.TurnsSince = 7
	ai.Update(UnitList{alien}, m, UnitList{}, nil, nil)
	if ai.State != AIPatrol {
		t.Errorf("expected AIPatrol after search timeout, got %d", ai.State)
	}
}

func TestAIUpdateFlankState(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	human := &Unit{X: 12, Y: 10, HP: 20, MaxHP: 20, TU: 50, Alive: true, Faction: 0, Weapon: "rifle", WeaponAmmo: 10}
	m := NewBattleMap(30, 20)
	alien.X = 10
	alien.Y = 10
	alien.TU = 50
	human.X = 15
	human.Y = 10
	human.Faction = 0
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	ai.State = AIFlank
	actions := ai.Update(UnitList{alien}, m, UnitList{human}, nil, nil)
	if len(actions) != 1 || actions[0].Type != "move" {
		t.Errorf("expected move action during flank, got %d actions", len(actions))
	}
}

func TestCivilianAIScared(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	civ := &Unit{X: 5, Y: 5, HP: 5, MaxHP: 5, Alive: true, Faction: 2, TU: 20}
	m := NewBattleMap(30, 20)
	alien.X = 6
	alien.Y = 5
	alien.Faction = 1
	// Clear
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	cai := NewCivilianAI(civ)
	actions := cai.GenerateActions(UnitList{civ, alien}, m)
	if len(actions) != 1 {
		t.Errorf("expected 1 flee action, got %d", len(actions))
	}
	if actions[0].Type != "move" {
		t.Errorf("expected move action, got %s", actions[0].Type)
	}
}

func TestCivilianAINotScared(t *testing.T) {
	civ := &Unit{X: 5, Y: 5, HP: 5, MaxHP: 5, Alive: true, Faction: 2, TU: 20}
	m := NewBattleMap(30, 20)
	// No threat nearby
	cai := NewCivilianAI(civ)
	actions := cai.GenerateActions(UnitList{civ}, m)
	if len(actions) != 0 {
		t.Error("civilian should not flee with no threat nearby")
	}
}

func TestAIPatrolTargetInBounds(t *testing.T) {
	at := data.GetAlienByName("Sectoid")
	alien := NewAlienUnit(at)
	m := NewBattleMap(30, 20)
	alien.X = 15
	alien.Y = 10
	alien.TU = 50
	for x := 0; x < 30; x++ {
		for y := 0; y < 20; y++ {
			m.Set(x, y, TileFloor)
		}
	}
	ai := NewAlienAI(alien)
	px, py := ai.patrolTarget(m)
	if px < 0 || px >= m.Width || py < 0 || py >= m.Height {
		t.Errorf("patrol target (%d,%d) out of bounds for %dx%d", px, py, m.Width, m.Height)
	}
}

func TestAINearestAlly(t *testing.T) {
	at1 := data.GetAlienByName("Sectoid")
	at2 := data.GetAlienByName("Floater")
	alien1 := NewAlienUnit(at1)
	alien2 := NewAlienUnit(at2)
	alien1.X = 10
	alien1.Y = 10
	alien2.X = 13
	alien2.Y = 10
	ai := NewAlienAI(alien1)
	ally := ai.nearestAlly(UnitList{alien1, alien2})
	if ally != alien2 {
		t.Error("nearest ally should be alien2")
	}
}
