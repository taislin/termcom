package battle

import (
	"testing"
	"time"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/soldier"
)

// --- Phase 40: battle test coverage (gaps beyond Phase 34) ---

// newTestBattlescape builds a minimal, fully-wired Battlescape suitable for
// exercising action methods (Grenade, Medikit, Psi, reaction fire, etc.)
// without needing a procedurally generated map or a real screen.
func newTestBattlescape(w, h int) *Battlescape {
	m := NewBattleMap(w, h)
	return &Battlescape{
		Map:       m,
		Gas:       NewGasGrid(w, h),
		Particles: engine.NewParticleSystem(512),
		Camera:    engine.NewCamera(1, 1),
		Phase:     PhasePlayerTurn,
		Status:    StatusPlayerTurn,
		Units:     UnitList{},
		Turn:      1,
	}
}

// testAlienType returns a valid procedurally generated alien type (with a
// weapon present in RuleItems) for constructing alien units in tests.
func testAlienType() *data.AlienType {
	_, types := data.GenerateSpecies(42)
	if len(types) == 0 {
		return nil
	}
	return types[0]
}

func TestLOSThroughWindow(t *testing.T) {
	m := NewBattleMap(15, 15)
	// Open grass: LOS clear across the row.
	if !m.hasLOS(2, 7, 12, 7) {
		t.Fatal("expected clear LOS across open grass")
	}
	// A wall blocks LOS.
	m.Set(7, 7, TileWall)
	if m.hasLOS(2, 7, 12, 7) {
		t.Error("expected LOS blocked by a wall")
	}
	// A window is transparent (not opaque): LOS passes through it.
	m.Set(7, 7, TileWindow)
	if !m.hasLOS(2, 7, 12, 7) {
		t.Error("expected LOS to pass through a window")
	}
}

func TestFOVRadius(t *testing.T) {
	m := NewBattleMap(30, 30)
	ux, uy := 15, 15
	sight := 8
	m.ComputeFOV(ux, uy, sight)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if m.Tiles[y][x].Visible {
				dx := x - ux
				dy := y - uy
				if dx*dx+dy*dy > sight*sight {
					t.Errorf("tile (%d,%d) visible beyond sight range", x, y)
				}
			}
		}
	}
	if m.Tiles[0][0].Visible {
		t.Error("far corner should be outside FOV")
	}
	if !m.Tiles[uy][ux].Visible {
		t.Error("origin should be visible to itself")
	}
}

func TestGrenadeDamage(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	at := testAlienType()
	if at == nil {
		t.Skip("no alien types available")
	}
	sel := NewSoldierUnit(soldier.NewSoldier("Grenadier"))
	sel.X, sel.Y = 10, 10
	sel.TU = 100
	bs.Selected = sel
	bs.Units = UnitList{sel}

	alien := NewAlienUnit(at)
	alien.X, alien.Y = 10, 11
	bs.Units = append(bs.Units, alien)

	bs.CursorX, bs.CursorY = 10, 11
	hpBefore := alien.HP
	bs.Grenade()
	if alien.HP >= hpBefore {
		t.Errorf("expected grenade to damage alien: before=%d after=%d", hpBefore, alien.HP)
	}
	if d, _, ok := bs.Gas.Get(10, 11); !ok || d != 3 {
		t.Error("expected smoke density 3 at grenade impact")
	}
}

func TestMedikitHeals(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	sel := NewSoldierUnit(soldier.NewSoldier("Medic"))
	sel.X, sel.Y = 10, 10
	sel.TU = 100
	ally := NewSoldierUnit(soldier.NewSoldier("Patient"))
	ally.X, ally.Y = 10, 11
	ally.HP = 5
	ally.MaxHP = 20
	bs.Selected = sel
	bs.Units = UnitList{sel, ally}
	bs.CursorX, bs.CursorY = 10, 11

	bs.UseMedikit()
	if ally.HP <= 5 {
		t.Errorf("expected medikit to heal ally: hp=%d", ally.HP)
	}
	if sel.TU != 75 {
		t.Errorf("expected medic TU 75 after medikit, got %d", sel.TU)
	}
}

func TestPsiAttackConsumesTU(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	at := testAlienType()
	if at == nil {
		t.Skip("no alien types available")
	}
	sel := NewSoldierUnit(soldier.NewSoldier("Psi"))
	sel.X, sel.Y = 10, 10
	sel.TU = 100
	sel.Soldier.Weapon = "psi_amp"
	bs.Selected = sel
	alien := NewAlienUnit(at)
	alien.X, alien.Y = 10, 11
	bs.Units = UnitList{sel, alien}
	bs.CursorX, bs.CursorY = 10, 11

	bs.PsiAttack()
	if sel.TU != 80 {
		t.Errorf("expected PsiAttack to consume 20 TU, got %d", sel.TU)
	}
}

func TestPsiAttackSuccess(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	at := testAlienType()
	if at == nil {
		t.Skip("no alien types available")
	}
	at.Psi = 0
	sel := NewSoldierUnit(soldier.NewSoldier("Psi"))
	sel.X, sel.Y = 10, 10
	sel.TU = 100
	sel.Soldier.Weapon = "psi_amp"
	sel.Soldier.PsiSkill = 100
	bs.Selected = sel
	alien := NewAlienUnit(at)
	alien.X, alien.Y = 10, 11
	bs.Units = UnitList{sel, alien}
	bs.CursorX, bs.CursorY = 10, 11

	bs.PsiAttack()
	if !alien.Panicked {
		t.Error("expected alien panicked on successful psi attack")
	}
	if alien.TU != 0 {
		t.Errorf("expected alien TU zeroed on successful psi attack, got %d", alien.TU)
	}
}

func TestReactionFireHumanTriggers(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	at := testAlienType()
	if at == nil {
		t.Skip("no alien types available")
	}
	human := NewSoldierUnit(soldier.NewSoldier("Spotter"))
	human.X, human.Y = 10, 10
	human.TU = 100
	human.Reactions = 100
	human.Accuracy = 100
	human.Weapon = "rifle"
	human.WeaponAmmo = 99
	alien := NewAlienUnit(at)
	alien.X, alien.Y = 10, 11
	alien.Alive = true
	bs.Units = UnitList{human, alien}
	bs.Map.Tiles[11][10].Visible = true

	bs.checkHumanReactionFire(alien)
	if bs.Status != StatusPlayerOverwatch {
		t.Error("expected human reaction fire to trigger (StatusPlayerOverwatch)")
	}
}

func TestReactionFireAlienTriggers(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	at := testAlienType()
	if at == nil {
		t.Skip("no alien types available")
	}
	human := NewSoldierUnit(soldier.NewSoldier("Runner"))
	human.X, human.Y = 10, 10
	human.Alive = true
	alien := NewAlienUnit(at)
	alien.X, alien.Y = 10, 11
	alien.TU = 100
	alien.Reactions = 100
	alien.Accuracy = 100
	alien.WeaponAmmo = 99
	bs.Units = UnitList{human, alien}
	ai := NewAlienAI(alien)
	bs.AlienAIs = []*AlienAI{ai}
	bs.Map.Tiles[10][10].Visible = true

	bs.checkAlienReactionFire(human)
	if bs.Status != StatusAlienOverwatch {
		t.Error("expected alien reaction fire to trigger (StatusAlienOverwatch)")
	}
}

func TestAICanSense(t *testing.T) {
	m := NewBattleMap(20, 20)
	at := &data.AlienType{
		Name: "TestAlien", HP: 20, TU: 30, Accuracy: 50, Bravery: 30,
		Reactions: 30, Strength: 10, Psi: 0, ResistPsionic: 0, Armour: 0, Weapon: "rifle",
	}
	u := NewAlienUnit(at)
	u.X, u.Y = 5, 5
	ai := NewAlienAI(u)
	// Visible target in open terrain is sensed.
	if !ai.canSense(5, 10, m) {
		t.Error("expected to sense a visible target")
	}
	// A wall between unit and target blocks sight; with no special senses it is not sensed.
	m.Set(5, 8, TileWall)
	if ai.canSense(5, 12, m) {
		t.Error("expected NOT to sense through a wall without special senses")
	}
}

func TestMapGeneratorsAll(t *testing.T) {
	generators := []struct {
		name string
		gen  func(int, int) *BattleMap
	}{
		{"CrashSite", GenerateCrashSite},
		{"TerrorSite", GenerateTerrorSite},
		{"AbductionSite", GenerateAbductionSite},
		{"UFOInterior", GenerateUFOInterior},
		{"Cydonia", GenerateCydonia},
		{"AlienBase", GenerateAlienBase},
		{"Forest", GenerateForest},
		{"Desert", GenerateDesert},
		{"Polar", GeneratePolar},
	}
	for _, g := range generators {
		m := g.gen(40, 40)
		if m == nil {
			t.Errorf("%s: returned nil map", g.name)
			continue
		}
		if m.Width != 40 || m.Height != 40 {
			t.Errorf("%s: expected 40x40, got %dx%d", g.name, m.Width, m.Height)
		}
		passable := 0
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				if m.Passable(x, y) {
					passable++
				}
			}
		}
		if passable == 0 {
			t.Errorf("%s: no passable tiles (no valid spawns)", g.name)
		}
	}
}

func TestSmokeBlocksLOS(t *testing.T) {
	g := NewGasGrid(10, 10)
	g.Set(5, 5, 3, GasSmoke)
	if !g.BlocksLOS(5, 5) {
		t.Error("expected density-3 smoke to block LOS")
	}
	g.Set(5, 5, 2, GasSmoke)
	if g.BlocksLOS(5, 5) {
		t.Error("expected density-2 smoke not to block LOS")
	}
}

func TestGasDiffusionSpreads(t *testing.T) {
	g := NewGasGrid(10, 10)
	g.Set(5, 5, 3, GasSmoke)
	if d, _, ok := g.Get(5, 5); !ok || d != 3 {
		t.Fatal("expected center smoke density 3")
	}
	if _, _, ok := g.Get(5, 6); ok {
		t.Fatal("neighbor should be empty before diffusion")
	}
	g.Diffuse()
	if d, _, ok := g.Get(5, 6); !ok || d < 1 {
		t.Errorf("expected neighbor (5,6) to receive diffused smoke, got density=%d ok=%v", d, ok)
	}
	if d, _, ok := g.Get(5, 5); !ok || d != 2 {
		t.Errorf("expected center to decay to density 2, got density=%d ok=%v", d, ok)
	}
}

func TestCustomVictorySurviveTurns(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	human := NewSoldierUnit(soldier.NewSoldier("Survivor"))
	human.X, human.Y = 5, 5
	human.Alive = true
	bs.Units = UnitList{human}
	bs.CustomVictory = &CustomVictory{Condition: "survive_turns", Turns: 3}
	bs.Turn = 3
	bs.checkVictory()
	if bs.Phase != PhaseVictory {
		t.Error("expected victory when survive_turns turn is reached")
	}
}

func TestCustomVictoryReachPoint(t *testing.T) {
	bs := newTestBattlescape(30, 30)
	human := NewSoldierUnit(soldier.NewSoldier("Runner"))
	human.X, human.Y = 7, 7
	human.Alive = true
	bs.Units = UnitList{human}
	bs.CustomVictory = &CustomVictory{Condition: "reach_point", TargetX: 7, TargetY: 7, MinSoldiers: 1}
	bs.checkVictory()
	if bs.Phase != PhaseVictory {
		t.Error("expected victory when a soldier reaches the extraction point")
	}
}

func TestReinforcementWaveSpawns(t *testing.T) {
	g := &engine.Game{GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	_, types := data.GenerateSpecies(42)
	g.AlienTypes = types
	bs := newTestBattlescape(30, 30)
	bs.Game = g
	before := len(bs.Units)
	bs.spawnReinforcementWave(2)
	if len(bs.Units) <= before {
		t.Error("expected reinforcement wave to add alien units")
	}
}
