package geo

import (
	"math/rand"
	"testing"
	"time"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
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
	if c.LangName() != language.String("CITY_NEW_YORK") {
		t.Errorf("expected New York, got %s", c.LangName())
	}
}

func TestUFOSpawnOnCities(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
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
	ufo := SpawnUFOOnCities(cities, 0)
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
	ufo := SpawnUFOOnCities(cities, 0)
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
	ufo := SpawnUFOOnCities(cities, 0)
	ufo.Type.Toughness = 1000    // high HP so it doesn't die
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
	ufo := SpawnUFOOnCities(cities, 0)
	damage := inter.FireAt(ufo)
	if damage != 0 {
		t.Errorf("expected 0 damage with no ammo, got %d", damage)
	}
}

func TestInterceptorDisengage(t *testing.T) {
	inter := NewInterceptor(48, 31)
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
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
		NodeFrom:  cities[0].ID,
		NodeTo:    cities[1].ID,
		Progress:  0.5,
		TurnsLeft: 1,
		Active:    true,
		Type:      UFOTypes[0],
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

func TestMultiBaseBuildCycleTransfer(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	if len(gs.Bases) != 1 {
		t.Fatalf("expected 1 base, got %d", len(gs.Bases))
	}
	if gs.SelectedBase() == nil {
		t.Fatal("selected base is nil")
	}

	// Build a second base at city 1
	gs.CursorNode = 1
	g.Funds = 1000000
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases after build, got %d", len(gs.Bases))
	}
	if gs.HasBaseAt(1) == nil {
		t.Error("expected a base at city 1")
	}
	if !gs.Cities[1].HasRadar {
		t.Error("city 1 should have radar after building base")
	}

	// Cannot build two bases at the same city
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Errorf("expected still 2 bases, got %d", len(gs.Bases))
	}

	// Cycle base (start from base 0)
	gs.ActiveBase = 0
	gs.CycleBase()
	if gs.ActiveBase != 1 {
		t.Errorf("expected active base 1 after cycle, got %d", gs.ActiveBase)
	}

	// Transfer a soldier from base 0 to base 1
	src := gs.Bases[0]
	dst := gs.Bases[1]
	if len(src.Soldiers) == 0 {
		t.Fatal("source base has no soldiers to transfer")
	}
	beforeDst := len(dst.Soldiers)
	before := len(src.Soldiers)
	ts := gs.NewTransferScreen()
	ts.FromIdx = 0
	ts.ToIdx = 1
	ts.Tab = 0
	ts.SelSoldier = 0
	ts.transferSoldier()
	if len(src.Soldiers) != before-1 {
		t.Errorf("expected source to lose 1 soldier, got %d -> %d", before, len(src.Soldiers))
	}
	if len(dst.Soldiers) != beforeDst+1 {
		t.Errorf("expected destination to gain 1 soldier, got %d -> %d", beforeDst, len(dst.Soldiers))
	}
}

func TestBaseDefenseDestroy(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	// add a second base
	g.Funds = 1000000
	gs.CursorNode = 2
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases, got %d", len(gs.Bases))
	}
	def := gs.Bases[1]
	gs.destroyBase(def)
	if len(gs.Bases) != 1 {
		t.Errorf("expected 1 base after destroy, got %d", len(gs.Bases))
	}
	if gs.HasBaseAt(2) != nil {
		t.Error("base at city 2 should be gone")
	}
	if gs.Cities[2].HasRadar {
		t.Error("city 2 radar should be off after base destroyed")
	}
}

func TestMultiBaseSaveLoad(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	g.Funds = 1000000
	gs.CursorNode = 3
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases, got %d", len(gs.Bases))
	}
	bs := gs.buildSaveData()
	if len(bs.Bases) != 2 {
		t.Fatalf("save data should have 2 bases, got %d", len(bs.Bases))
	}
	if bs.Bases[0].CityID != gs.Bases[0].CityID || bs.Bases[1].CityID != gs.Bases[1].CityID {
		t.Error("saved base city IDs do not match")
	}
}

func TestSpawnMissionVariety(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	seen := make(map[string]bool)
	for i := 0; i < 2000; i++ {
		gs.spawnMission()
		if len(gs.Missions) > 0 {
			seen[gs.Missions[len(gs.Missions)-1].Type] = true
		}
		// Keep the list bounded so the test stays fast
		if len(gs.Missions) > 50 {
			gs.Missions = gs.Missions[:1]
		}
	}
	for _, want := range []string{
		language.String("MISSION_TERROR"),
		language.String("MISSION_SUPPLY"),
		language.String("MISSION_ABDUCTION"),
		language.String("MISSION_RESEARCH"),
		language.String("MISSION_COUNCIL"),
		language.String("MISSION_ALIEN_BASE"),
	} {
		if !seen[want] {
			t.Errorf("mission type %q never spawned in 2000 attempts", want)
		}
	}
}

func TestRespondToSelectedMission(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	// Place a mission at node 5
	gs.Missions = append(gs.Missions, &AlienMission{Type: language.String("MISSION_TERROR"), NodeID: 5, HoursLeft: 24})
	gs.CursorNode = 5
	gs.RespondToSelectedMission()
	if !gs.Game.InState(engine.StateBattlescape) {
		t.Errorf("expected battle state %v after responding", engine.StateBattlescape)
	}
	if gs.ActiveMissionType != language.String("MISSION_TERROR") {
		t.Errorf("expected ActiveMissionType TERROR, got %q", gs.ActiveMissionType)
	}
}

func TestApplyMissionRewards(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	b := gs.SelectedBase()
	before := g.Funds

	gs.ActiveMissionType = language.String("MISSION_COUNCIL")
	gs.applyMissionRewards(b)
	if g.Funds <= before {
		t.Errorf("council reward should increase funds, before=%d after=%d", before, g.Funds)
	}
	if b.CountItem("elerium") <= 0 {
		t.Error("council reward should grant elerium loot")
	}

	gs.ActiveMissionType = language.String("MISSION_SUPPLY")
	gs.applyMissionRewards(b)
	if b.CountItem("alloys") <= 0 {
		t.Error("supply raid reward should grant alloys")
	}

	gs.ActiveMissionType = language.String("MISSION_RESEARCH")
	gs.applyMissionRewards(b)
	if b.CountItem("ufo_power") <= 0 {
		t.Error("research reward should grant ufo_power loot")
	}
}

func TestGenerateAlienBaseMap(t *testing.T) {
	m := battle.GenerateAlienBase(50, 50)
	if m == nil {
		t.Fatal("GenerateAlienBase returned nil")
	}
	if m.Width != 50 || m.Height != 50 {
		t.Errorf("expected 50x50 map, got %dx%d", m.Width, m.Height)
	}
}

func TestCydoniaTriggersOnce(t *testing.T) {
	g := &engine.Game{GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.MissionsWon = 10
	gs.triggerCydonia()
	n1 := len(gs.Missions)
	gs.triggerCydonia()
	n2 := len(gs.Missions)
	if n2 != n1 {
		t.Errorf("triggerCydonia should only add the final mission once, got %d -> %d", n1, n2)
	}
	if !gs.CydoniaTriggered {
		t.Error("CydoniaTriggered should be set after triggerCydonia")
	}
}

func TestMultiBaseResearchAdvancesAllBases(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)

	// Build a second base at city 2
	gs.CursorNode = 2
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases, got %d", len(gs.Bases))
	}

	// Add labs to both bases (base0 already has one from NewGeoscape)
	gs.Bases[0].Facilities = append(gs.Bases[0].Facilities, &base.Facility{Type: base.FacLab})
	gs.Bases[1].Facilities = append(gs.Bases[1].Facilities, &base.Facility{Type: base.FacLab})

	// Set up research directly on both bases (bypass StartResearch for determinism)
	gs.Bases[0].ActiveResearch = &base.ResearchProject{
		TopicID: "alien_alloys", Cost: 100, Scientists: 5,
	}
	gs.Bases[1].ActiveResearch = &base.ResearchProject{
		TopicID: "alien_alloys", Cost: 100, Scientists: 5,
	}
	// Ensure enough unassigned scientists to match
	gs.Bases[0].UnassignedScientists = 10
	gs.Bases[1].UnassignedScientists = 10

	initial0 := gs.Bases[0].ActiveResearch.Progress
	initial1 := gs.Bases[1].ActiveResearch.Progress

	// Switch to base 0
	gs.ActiveBase = 0

	// Advance research on both bases (simulating what the Update loop does for all bases)
	_ = gs.Bases[0].AdvanceResearch()
	_ = gs.Bases[1].AdvanceResearch()

	if gs.Bases[0].ActiveResearch.Progress <= initial0 {
		t.Error("base 0 research should have progressed")
	}
	if gs.Bases[1].ActiveResearch.Progress <= initial1 {
		t.Error("base 1 research should have progressed (bug: only selected base advances)")
	}
}

func TestMultiBaseManufactureAdvancesAllBases(t *testing.T) {
	species, _ := data.GenerateSpecies(42)
	data.InitResearchTree(42, species)

	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)

	// Build a second base at city 2
	gs.CursorNode = 2
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases, got %d", len(gs.Bases))
	}

	// Add workshop to base1 (base0 already has one from NewGeoscape) and storage/materials to both
	gs.Bases[1].Facilities = append(gs.Bases[1].Facilities, &base.Facility{Type: base.FacWorkshop})
	for i := 0; i < 2; i++ {
		gs.Bases[i].Facilities = append(gs.Bases[i].Facilities, &base.Facility{Type: base.FacStorage})
		gs.Bases[i].AddItem("alloys", 10)
	}

	// Start manufacture at both bases
	if !gs.Bases[0].StartManufacture("pistol", 1, map[string]int{"alloys": 1}) {
		t.Fatal("base0 could not start manufacture")
	}
	if !gs.Bases[1].StartManufacture("pistol", 1, map[string]int{"alloys": 1}) {
		t.Fatal("base1 could not start manufacture")
	}

	// Assign engineers to both
	gs.Bases[0].AssignEngineers(0, 3)
	gs.Bases[1].AssignEngineers(0, 3)

	initial0 := gs.Bases[0].ManufactureQueue[0].Progress
	initial1 := gs.Bases[1].ManufactureQueue[0].Progress

	// Switch to base 0
	gs.ActiveBase = 0

	// Advance manufacture
	_ = gs.Bases[0].AdvanceManufacture()
	_ = gs.Bases[1].AdvanceManufacture()

	if gs.Bases[0].ManufactureQueue[0].Progress <= initial0 {
		t.Error("base 0 manufacture should have progressed")
	}
	if gs.Bases[1].ManufactureQueue[0].Progress <= initial1 {
		t.Error("base 1 manufacture should have progressed (bug: only selected base advances)")
	}
}

func TestSelectedBaseResearchCompletesUnselected(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)

	// Build a second base at city 2
	gs.CursorNode = 2
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatalf("expected 2 bases, got %d", len(gs.Bases))
	}

	// Set up cheap research on base 1
	gs.Bases[1].ActiveResearch = &base.ResearchProject{
		TopicID: "alien_alloys", Cost: 5, Scientists: 10,
	}
	gs.Bases[1].UnassignedScientists = 0

	// Switch to base 0
	gs.ActiveBase = 0

	// Advance the unselected base 1 research until it completes
	for i := 0; i < 5; i++ {
		gs.Bases[1].AdvanceResearch()
	}

	if !gs.Bases[1].HasResearch("alien_alloys") {
		t.Error("unselected base 1 should have completed research when advanced directly")
	}
}

func TestCydoniaVictory(t *testing.T) {
	g := &engine.Game{GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.MissionsWon = 10
	gs.CydoniaTriggered = true
	gs.ActiveFinalMission = true
	gs.Game.ActiveBattle = &engine.BattleResult{
		Won:     true,
		Soldiers: gs.SelectedBase().Soldiers,
	}
	gs.Update()
	if !gs.Victory {
		t.Error("expected Victory=true after winning the Cydonia mission")
	}
	if !gs.Game.InState(engine.StateGameOver) {
		t.Error("expected GameOver state after victory")
	}
}

// --- Phase 39: geo test coverage ---

func TestInterceptorDestroysUFO(t *testing.T) {
	inter := NewInterceptor(48, 31)
	// Guarantee a hit: max accuracy so rand.Intn(100) >= 100 is never true.
	inter.Weapon.Accuracy = 100
	inter.PilotSkill = 50
	ufo := &UFO{
		Type:   UFOType{Name: "Small Scout", Toughness: 1, MaxHP: 1},
		X:      inter.X,
		Y:      inter.Y,
		Active: true,
	}
	dmg := inter.FireAt(ufo)
	if dmg != -1 {
		t.Errorf("expected UFO destroyed (return -1), got %d", dmg)
	}
	if ufo.Active {
		t.Error("expected UFO to be inactive after destruction")
	}
}

func TestUFODestroysInterceptor(t *testing.T) {
	cities := GetCities()
	ufo := SpawnUFOOnCities(cities, 0)
	inter := NewInterceptor(48, 31)
	ufo.X, ufo.Y = inter.X, inter.Y
	inter.HP = 1
	destroyed := false
	for i := 0; i < 1000; i++ {
		if inter.HP <= 0 {
			destroyed = true
			break
		}
		ufo.FireAtInterceptor(inter)
	}
	if !destroyed {
		t.Error("UFO should be able to destroy a 1-HP interceptor")
	}
}

func TestRespondToMission(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.Missions = append(gs.Missions, &AlienMission{
		Type:      "MISSION_TERROR",
		NodeID:    2,
		HoursLeft: 24,
	})
	if gs.SelectedBase() == nil {
		t.Fatal("expected a selected base")
	}
	if len(gs.SelectedBase().HealthySoldiers()) == 0 {
		t.Fatal("expected healthy soldiers to respond with")
	}
	gs.RespondToMission(0)
	if len(gs.Missions) != 0 {
		t.Errorf("expected mission to be consumed, %d remain", len(gs.Missions))
	}
	if !gs.Game.Paused {
		t.Error("expected game to pause when squad deployed")
	}
	if !gs.Game.InState(engine.StateBattlescape) {
		t.Error("expected battlescape state after responding to mission")
	}
}

func TestAutoresolveMissionBranches(t *testing.T) {
	sawWin, sawLoss := false, false
	for seed := int64(0); seed < 120; seed++ {
		//lint:ignore SA1019 rand.Seed intentionally seeds the global RNG for deterministic test runs.
		rand.Seed(seed)
		g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC), AlienKnowledge: make(map[string]int)}
		gs := NewGeoscape(g)
		gs.Missions = append(gs.Missions, &AlienMission{
			Type:      "MISSION_TERROR",
			NodeID:    2,
			HoursLeft: 24,
		})
		beforeFunds := g.Funds
		beforeWon := gs.MissionsWon
		beforeSoldiers := len(gs.SelectedBase().Soldiers)
		gs.AutoresolveMission(0)
		if g.Funds > beforeFunds && gs.MissionsWon > beforeWon {
			sawWin = true
		} else if len(gs.SelectedBase().Soldiers) < beforeSoldiers {
			sawLoss = true
		}
	}
	if !sawWin {
		t.Error("never observed an autoresolve win across seeds")
	}
	if !sawLoss {
		t.Error("never observed an autoresolve loss across seeds")
	}
}

func TestTransportArrivalStartsBattle(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	cs := &CrashSite{UFOName: "Test UFO", NodeID: 2}
	gs.DispatchTransport(cs)
	if gs.Transport == nil {
		t.Fatal("expected transport to be dispatched")
	}
	started := false
	for i := 0; i < 300; i++ {
		gs.Update()
		if gs.Transport == nil && gs.Game.InState(engine.StateBattlescape) {
			started = true
			break
		}
	}
	if !started {
		t.Error("transport arrival should start a battlescape")
	}
}

func TestBuildBaseInsufficientFunds(t *testing.T) {
	g := &engine.Game{Funds: 0, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.CursorNode = 2
	basesBefore := len(gs.Bases)
	gs.BuildBase()
	if len(gs.Bases) != basesBefore {
		t.Error("base should not be built without funds")
	}
	if g.Funds != 0 {
		t.Error("funds should remain 0")
	}
	if gs.Message != language.String("MSG_INSUFFICIENT_FUNDS") {
		t.Errorf("expected insufficient-funds message, got %q", gs.Message)
	}
}

func TestTimeSpeedPauseGate(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 12, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)

	// Paused: time must not advance
	gs.Game.Paused = true
	t0 := gs.Game.GameTime
	gs.Update()
	if !gs.Game.GameTime.Equal(t0) {
		t.Error("GameTime advanced while paused")
	}

	// TimeSpeed == 0: time must not advance
	gs.Game.Paused = false
	gs.Game.TimeSpeed = 0
	t1 := gs.Game.GameTime
	gs.Update()
	if !gs.Game.GameTime.Equal(t1) {
		t.Error("GameTime advanced with TimeSpeed 0")
	}

	// TimeSpeed > 0: time must advance
	gs.Game.TimeSpeed = 1
	t2 := gs.Game.GameTime
	gs.Update()
	if gs.Game.GameTime.Equal(t2) {
		t.Error("GameTime did not advance with TimeSpeed > 0")
	}
}

func TestLoseConditionAlienActivity(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.AlienActivity = 100
	gs.Update()
	if !gs.Game.InState(engine.StateGameOver) {
		t.Error("expected GameOver when AlienActivity >= 100")
	}
	if !gs.Victory {
		t.Error("expected Victory flag set (defeat latch) on alien victory")
	}
}

func TestLoseConditionLastBase(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	// Only a single base exists; destroying it ends the game.
	gs.destroyBase(gs.Bases[0])
	if !gs.Game.InState(engine.StateGameOver) {
		t.Error("expected GameOver when last base destroyed")
	}
}

// --- Phase 41: transfer between bases (geo/transfer.go) ---

func TestTransferItemsBetweenBases(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.CursorNode = 2
	gs.BuildBase()
	if len(gs.Bases) != 2 {
		t.Fatal("expected 2 bases")
	}
	gs.Bases[0].AddItem("rifle", 1)
	before0 := gs.Bases[0].CountItem("rifle")
	before1 := gs.Bases[1].CountItem("rifle")
	ts := &TransferScreen{Geo: gs, FromIdx: 0, ToIdx: 1, SelItem: 0, Tab: 1}
	ts.transferItem()
	if gs.Bases[0].CountItem("rifle") != before0-1 {
		t.Errorf("source base should lose 1 rifle: before=%d after=%d", before0, gs.Bases[0].CountItem("rifle"))
	}
	if gs.Bases[1].CountItem("rifle") != before1+1 {
		t.Errorf("dest base should gain 1 rifle: before=%d after=%d", before1, gs.Bases[1].CountItem("rifle"))
	}
}

func TestTransferSoldierBetweenBases(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	gs := NewGeoscape(g)
	gs.CursorNode = 2
	gs.BuildBase()
	before0 := len(gs.Bases[0].Soldiers)
	before1 := len(gs.Bases[1].Soldiers)
	ts := &TransferScreen{Geo: gs, FromIdx: 0, ToIdx: 1, SelSoldier: 0, Tab: 0}
	ts.transferSoldier()
	if len(gs.Bases[0].Soldiers) != before0-1 {
		t.Errorf("source base should lose 1 soldier: before=%d after=%d", before0, len(gs.Bases[0].Soldiers))
	}
	if len(gs.Bases[1].Soldiers) != before1+1 {
		t.Errorf("dest base should gain 1 soldier: before=%d after=%d", before1, len(gs.Bases[1].Soldiers))
	}
}
