package geo

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/save"
	"github.com/taislin/termcom/internal/soldier"
)

// Geoscape timing constants.
const (
	baseSpawnRateFloor  = 1200 // minimum ticks between alien base attempts
	missionSpawnFloor   = 300  // minimum ticks between mission spawns
	activityTickRate    = 7200 // ticks between alien activity increases (~2h at speed 1)
	ufoSpawnRateFloor   = 100  // hard floor for UFO spawn interval
	ufoSpawnRateBase    = 600  // base UFO spawn interval
	ufoSpawnRateDecay   = 20   // monthly decay of spawn interval
	ufoSpawnRateFloorSoft = 200 // soft floor for UFO spawn interval
	missionSpawnBase    = 1800 // base mission spawn interval
	missionSpawnActWt   = 15   // alien activity weight in spawn formula
	missionSpawnDecay   = 30   // monthly decay of mission spawn interval
	baseSpawnRateBase   = 3600 // base alien-base spawn interval
	baseSpawnRateDecay  = 60   // monthly decay of alien-base spawn interval
	alienActivityTick   = 2400 // ticks between alien base threat escalation
	missionIntervalBase = 1800 // base interval for alien-base missions
	missionIntervalDecay = 30  // monthly decay of alien-base mission interval
	missionIntervalFloor = 600 // floor for alien-base mission interval
	baseRadarRange      = 24   // base detection range tiles
	perRadarRange       = 10   // extra range per radar facility
	minimapClickRadius  = 25   // click detection radius on minimap
	maxEdgePathDist     = 50.0 // max edge distance for BFS pathfinding
	transportSentinel   = 999999 // large sentinel for "no nearest" in transport routing

	// World map dimensions used for day/night rendering.
	worldMapW = 180
	worldMapH = 90
)

// cityName returns the localized name of the city at the given node ID,
// or "?" if the ID is invalid.
func (gs *Geoscape) cityName(id int) string {
	if c := gs.CityByID(id); c != nil {
		return c.LangName()
	}
	return "?"
}

// resumeRealtime unpauses the game and ensures time is flowing.
func (gs *Geoscape) resumeRealtime() {
	gs.Game.Paused = false
	if gs.Game.TimeSpeed == 0 {
		gs.Game.TimeSpeed = 1
	}
}

// calcAlienPower returns the total alien threat level for autoresolve.
func (gs *Geoscape) calcAlienPower(alienCount int) int {
	diffMult := 1.0
	if gs.Game.Difficulty >= 0 && gs.Game.Difficulty < len(engine.Difficulties) {
		diffMult = engine.Difficulties[gs.Game.Difficulty].AlienScale
	}
	return int(float64(alienCount*(40+gs.MissionsWon*3)) * diffMult)
}

// calcSquadPower returns the total power of the given soldiers (includes perk bonuses).
func calcSquadPower(healthy []*soldier.Soldier) int {
	power := 0
	for _, s := range healthy {
		power += s.HP + s.Accuracy/2 + s.Strength + s.Reactions/2
		if s.HasPerk("marksman") {
			power += 15
		}
		if s.HasPerk("tough") {
			power += 20
		}
		if s.HasPerk("close_combat") {
			power += 10
		}
		if s.HasPerk("overwatch") {
			power += 10
		}
	}
	return power
}

// calcWinChance returns the clamped [10,70] win probability for autoresolve.
func (gs *Geoscape) calcWinChance(healthy []*soldier.Soldier, missionTypeMod int) int {
	alienCount := 5 + gs.MissionsWon/2
	if alienCount > 10 {
		alienCount = 10
	}
	squadPower := calcSquadPower(healthy)
	alienPower := gs.calcAlienPower(alienCount)
	winChance := 30 + (squadPower-alienPower)/5 + missionTypeMod
	if winChance > 70 {
		winChance = 70
	}
	if winChance < 10 {
		winChance = 10
	}
	return winChance
}

// AlienMission describes an active UFO threat that must be responded to.
type AlienMission struct {
	Type      string
	NodeID    int     // target node ID on the world map
	HoursLeft float64 // hours remaining to respond before the UFO departs
}

// AlienBase represents a persistent alien stronghold on the Geoscape.
type AlienBase struct {
	CityID          int
	Threat          int    // 0-100, defense level
	TurnsAlive      int    // ticks since establishment (scales defenses)
	LastMissionTick int    // last tick a mission was spawned from here
	DefendingUFOID  int    // UFO ID of the defensive craft (-1 = none)
	Name            string // e.g. "Alien Base #1"
}

// CrashSite represents a landed UFO that can be explored for loot/tech.
type CrashSite struct {
	UFOName string
	NodeID  int
	Looted  bool
	Seed    int64 // deterministic seed for procedural UFO blueprint
}

// Transport handles the movement of soldiers between bases.
type Transport struct {
	FromNode       int
	ToNode         int
	Progress       float64
	Returning      bool
	CrashSite      *CrashSite
	SourceBaseCity int // city ID of the base that dispatched this transport
}

// Geoscape is the main state controller for the strategic world map.
type Geoscape struct {
	// Game Engine and State
	Game                *engine.Game
	Cities              []*City
	UFOs                UFOList
	Interceptors        InterceptorList
	CrashSites          []*CrashSite
	Transport           *Transport
	Message             string
	MessageTimer        time.Time
	TickCounter         int
	Bases               []*base.Base
	ActiveBase          int
	LastMonth           int
	LastDay             int
	Missions            []*AlienMission
	AlienBases          []*AlienBase
	AlienActivity       int
	MissionsWon         int
	Victory             bool
	Defeated            bool
	LastSpeed           int
	CursorNode          int
	TargetSelectionMode bool
	TargetCursor        int // cursor index in target-selection mode (separate from CursorNode)
	PreBattleStats      map[string][6]int
	ActiveCrashSite     *CrashSite
	ActiveBaseDefense   *base.Base // non-nil if the current battle is defending this base
	ActiveMissionType   string     // mission Type string of the battle in progress (for rewards)
	ActiveFinalMission  bool       // non-nil if the current battle is the Cydonia final mission
	CydoniaTriggered    bool       // ensures the final mission is added only once
	ShowRadarOverlay    bool       // toggle radar coverage circles on minimap

	// Interaction State
	MissionSelectMode  bool
	MissionSelectIdx   int
	MissionOdds        int
	respondedAlienBase *AlienBase // non-nil when responding to an alien base assault

	// Visual Effects
	DogfightVisual *DogfightAnim

	// Cached layout rects from the last Render, used by the click handler so
	// mouse hit-testing matches the (possibly mobile-stacked) on-screen layout.
	tableRect [4]int // x, y, w, h of the region table panel
	mapRect   [4]int // x, y, w, h of the world-map panel
}

// DogfightAnim drives the minimap combat visual for interceptor-vs-UFO engagements.
type DogfightAnim struct {
	Active      bool
	Timer       int // frames remaining in animation
	Interceptor *Interceptor
	UFO         *UFO

	InterHP    int // snapshot for display
	InterMaxHP int
	UFOHP      int
	UFOMaxHP   int

	InterDamage    int  // damage dealt to UFO (0=miss, -1=destroyed, >0=hit)
	UFODamage      int  // damage received from UFO
	UFOAlive       bool // UFO survived the exchange
	InterAlive     bool // interceptor survived
	InterDestroyed bool

	HitFlash    int // >0 = flash interceptor (damage taken)
	UFOHitFlash int // >0 = flash UFO (damage taken)
}

func (gs *Geoscape) FindBaseIndex(cityID int) int {
	for i, b := range gs.Bases {
		if b.CityID == cityID {
			return i
		}
	}
	return -1
}

func (gs *Geoscape) SelectedBase() *base.Base {
	if gs.ActiveBase < 0 || gs.ActiveBase >= len(gs.Bases) {
		return nil
	}
	return gs.Bases[gs.ActiveBase]
}

// Touch-control view helpers (engine defines the geoView interface).
func (gs *Geoscape) UFOCount() int        { return len(gs.UFOs) }
func (gs *Geoscape) MissionCount() int     { return len(gs.Missions) }
func (gs *Geoscape) HasSelectedBase() bool { return gs.SelectedBase() != nil }

// CanConfirm reports whether the Enter key currently performs a meaningful
// action (i.e. confirming an interceptor launch at a selected target).
func (gs *Geoscape) CanConfirm() bool {
	if gs.TargetSelectionMode {
		return len(gs.getTargets()) > 0
	}
	return false
}

// processBattleResult handles battle resolution and pushes the debrief screen.
func (gs *Geoscape) processBattleResult() {
	r := gs.Game.ActiveBattle
	defendingBase := gs.SelectedBase()
	if gs.ActiveBaseDefense != nil {
		defendingBase = gs.ActiveBaseDefense
	}
	if defendingBase == nil {
		gs.Game.ActiveBattle = nil
		return
	}
	// Merge battle-result soldiers into base roster (preserving non-deployed soldiers).
	battleSoldiers := make(map[string]*soldier.Soldier, len(r.Soldiers))
	for _, s := range r.Soldiers {
		battleSoldiers[s.Name] = s
	}
	for _, s := range defendingBase.Soldiers {
		if bs, ok := battleSoldiers[s.Name]; ok {
			*s = *bs // copy stats from battle result
		}
	}
	dead := defendingBase.RemoveDeadSoldiers()

	// Build per-soldier report from PreBattleStats
	statNames := []string{language.String("STAT_HP"), language.String("STAT_ACC"), language.String("STAT_REA"), language.String("STAT_STR"), language.String("STAT_BRA"), language.String("STAT_TU")}
	var soldiers []engine.DebriefSoldier
	if gs.PreBattleStats != nil {
		// Include dead soldiers from the map not in defendingBase.Soldiers
		alive := make(map[string]bool, len(defendingBase.Soldiers))
		for _, s := range defendingBase.Soldiers {
			alive[s.Name] = true
		}
		for name, old := range gs.PreBattleStats {
			died := !alive[name]
			// Find current soldier data if alive
			var cur *soldier.Soldier
			for _, s := range defendingBase.Soldiers {
				if s.Name == name {
					cur = s
					break
				}
			}
			rankStr := "---"
			gains := ""
			if cur != nil {
				rankStr = cur.Rank.String()
				newStats := [6]int{cur.HP, cur.Accuracy, cur.Reactions, cur.Strength, cur.Bravery, cur.TU}
				gainParts := []string{}
				for i := 0; i < 6; i++ {
					if newStats[i] > old[i] {
						gainParts = append(gainParts, fmt.Sprintf(language.String("STAT_GAIN_FORMAT"), statNames[i], newStats[i]-old[i]))
					}
				}
				gains = strings.Join(gainParts, " ")
			}
			soldiers = append(soldiers, engine.DebriefSoldier{
				Name:      name,
				Rank:      rankStr,
				Died:      died,
				StatGains: gains,
			})
		}
		gs.PreBattleStats = nil
	}

	// Build mission name
	missionName := language.String("GEO_MISSION_TACTICAL")
	baseDestroyed := false
	if gs.ActiveFinalMission {
		missionName = language.String("MSG_CYDONIA_ASSAULT")
	} else if gs.ActiveCrashSite != nil {
		missionName = fmt.Sprintf(language.String("GEO_MISSION_CRASH_SITE"), localizeUFOName(gs.ActiveCrashSite.UFOName))
	} else if gs.ActiveBaseDefense != nil {
		missionName = fmt.Sprintf(language.String("GEO_MISSION_BASE_DEFENSE"), defendingBase.Name)
	} else if gs.ActiveMissionType != "" {
		missionName = gs.ActiveMissionType
	}

	if r.Won {
		stunnedCount := 0
		if len(r.StunnedAliens) > 0 {
			capacity := defendingBase.CountFacility(base.FacContainment) * 10
			for _, alien := range r.StunnedAliens {
				if len(defendingBase.LiveAliens) < capacity {
					defendingBase.LiveAliens = append(defendingBase.LiveAliens, alien)
					stunnedCount++
				} else {
					break
				}
			}
		}
		defendingBase.AddLoot(r.LootItems)
		gs.MissionsWon++
		if gs.ActiveFinalMission {
			gs.Victory = true
		} else if gs.ActiveCrashSite != nil {
			cs := gs.ActiveCrashSite
			cs.Looted = true
			loot := generateUFOLoot(cs.UFOName)
			defendingBase.AddLoot(loot)
			r.LootItems = append(r.LootItems, loot...)
		}
		if gs.ActiveMissionType != "" {
			gs.applyMissionRewards(defendingBase)
		}
		// Alien Base Assault victory: destroy the alien base
		if gs.ActiveMissionType == language.String("MISSION_ALIEN_BASE") {
			cityNode := -1
			// Find which city this was at from the crash site or cursor
			if gs.ActiveCrashSite != nil {
				cityNode = gs.ActiveCrashSite.NodeID
			}
			if ab := gs.respondedAlienBase; ab != nil {
				gs.destroyAlienBase(ab)
				gs.respondedAlienBase = nil
				gs.Message = fmt.Sprintf(language.String("GEO_ALIEN_BASE_DESTROYED"), gs.CityByID(ab.CityID).LangName())
				gs.MessageTimer = time.Now()
			} else if cityNode >= 0 {
				if ab := gs.alienBaseAt(cityNode); ab != nil {
					gs.destroyAlienBase(ab)
				}
			}
		}

		dd := &engine.DebriefData{
			Won:            true,
			MissionName:    missionName,
			BaseName:       defendingBase.Name,
			Kills:          r.Kills,
			Casualties:     dead,
			LootItems:      r.LootItems,
			StunnedCount:   stunnedCount,
			FundsEarned:    50000,
			Soldiers:       soldiers,
			CydoniaVictory: gs.ActiveFinalMission,
		}
		gs.Game.SetScreen(engine.StateDebrief, engine.NewDebriefScreen(gs.Game, dd))
		gs.Game.PushState(engine.StateDebrief)
	} else {
		if gs.ActiveBaseDefense != nil {
			gs.destroyBase(defendingBase)
			baseDestroyed = true
		}
		// Remove crash site on loss
		if gs.ActiveCrashSite != nil {
			for i, cs := range gs.CrashSites {
				if cs == gs.ActiveCrashSite {
					gs.CrashSites = append(gs.CrashSites[:i], gs.CrashSites[i+1:]...)
					break
				}
			}
		}
		gs.respondedAlienBase = nil
		dd := &engine.DebriefData{
			Won:           false,
			MissionName:   missionName,
			BaseName:      defendingBase.Name,
			Kills:         r.Kills,
			Casualties:    dead,
			LootItems:     nil,
			StunnedCount:  0,
			FundsEarned:   0,
			Soldiers:      soldiers,
			BaseDestroyed: baseDestroyed,
		}
		gs.Game.SetScreen(engine.StateDebrief, engine.NewDebriefScreen(gs.Game, dd))
		gs.Game.PushState(engine.StateDebrief)
	}
	gs.MessageTimer = time.Now()
	gs.ActiveCrashSite = nil
	gs.ActiveBaseDefense = nil
	gs.ActiveMissionType = ""
	gs.ActiveFinalMission = false
	gs.Game.ActiveBattle = nil
}

func NewGeoscape(g *engine.Game) *Geoscape {
	b := base.NewBase("Base 1", 0)
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})

	cities := GetCities()
	cities[0].HasRadar = true
	cities[0].InterceptorCount = 2

	gs := &Geoscape{
		Game:         g,
		Cities:       cities,
		Bases:        []*base.Base{b},
		ActiveBase:   0,
		CursorNode:   0,
		Message:      language.String("MSG_WELCOME"),
		MessageTimer: time.Now(),
		LastMonth:    int(g.GameTime.Month()),
		LastDay:      g.GameTime.YearDay(),
	}
	return gs
}

func NewGeoscapeFromSave(g *engine.Game, sd *save.SaveData) *Geoscape {
	var bases []*base.Base
	for _, bs := range sd.Bases {
		bases = append(bases, save.ToBase(bs))
	}
	if len(bases) == 0 {
		b := base.NewBase("Base 1", 0)
		b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
		b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
		b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
		b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
		b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})
		bases = append(bases, b)
	}

	cities := GetCities()
	for _, b := range bases {
		if b.CityID >= 0 && b.CityID < len(cities) {
			cities[b.CityID].HasRadar = true
		}
	}

	gs := &Geoscape{
		Game:          g,
		Cities:        cities,
		Bases:         bases,
		ActiveBase:    0,
		CursorNode:    0,
		Message:       language.String("MSG_GAME_LOADED"),
		MessageTimer:  time.Now(),
		LastMonth:     int(sd.GameTime.Month()),
		LastDay:       sd.GameTime.YearDay(),
		AlienActivity: sd.AlienActivity,
	}

	g.GameTime = sd.GameTime
	g.Funds = sd.Funds
	g.Paused = sd.Paused
	g.TimeSpeed = sd.TimeSpeed
	g.Difficulty = sd.Difficulty

	for _, u := range sd.UFOs {
		ufoType := GetUFOTypeByName(u.TypeName)
		if ufoType != nil {
			gs.UFOs = append(gs.UFOs, &UFO{
				Type:     *ufoType,
				X:        u.X,
				Y:        u.Y,
				Progress: u.Progress,
				NodeFrom: u.NodeFrom,
				NodeTo:   u.NodeTo,
				Active:   u.Active,
			})
		}
	}
	for _, m := range sd.Missions {
		gs.Missions = append(gs.Missions, &AlienMission{
			Type:      m.Type,
			NodeID:    int(m.X),
			HoursLeft: m.HoursLeft,
		})
	}

	return gs
}

func (gs *Geoscape) Update() {
	gs.TickCounter++

	// 1. Process battle outcomes (rewards, casualties)
	if gs.Game.ActiveBattle != nil {
		gs.processBattleResult()
	}

	// 2. Defeat condition: alien activity exceeds the threshold
	if gs.AlienActivity >= 100 && !gs.Defeated && !gs.Victory {
		stats := fmt.Sprintf(language.String("GEO_MISSIONS_WON"), gs.MissionsWon)
		gs.Game.GameOver(false, stats)
		gs.Defeated = true
		gs.Game.Paused = true
	}

	// 3. Victory condition: check if enough missions have been completed
	if gs.MissionsWon >= 10 && !gs.Victory && !gs.Defeated {
		// Instead of immediate victory, trigger the final campaign stage (Cydonia)
		gs.triggerCydonia()
	}

	// 4. Final mission check: campaign ends when final mission is resolved
	if gs.Victory && gs.Game.ActiveBattle == nil {
		stats := fmt.Sprintf(language.String("GEO_CAMPAIGN_COMPLETE"), gs.MissionsWon)
		gs.Game.GameOver(true, stats)
	}

	// 5. Real-time world simulation (only when game is not paused and time is moving)
	if !gs.Game.Paused && gs.Game.TimeSpeed > 0 {
		speedMult := []int{0, 1, 5, 20, 60}
		minutes := speedMult[gs.Game.TimeSpeed]

		// UFO Spawning: Frequency increases as the campaign progresses (gameMonth).
		gameMonth := int(gs.Game.GameTime.Month()) - 3 + (gs.Game.GameTime.Year()-1999)*12
		if gameMonth < 0 {
			gameMonth = 0
		}
		ufoSpawnRate := ufoSpawnRateBase - gameMonth*ufoSpawnRateDecay
		if ufoSpawnRate < ufoSpawnRateFloorSoft {
			ufoSpawnRate = ufoSpawnRateFloorSoft
		}
		diffUFOScale := 1.0
		if gs.Game.Difficulty >= 0 && gs.Game.Difficulty < len(engine.Difficulties) {
			diffUFOScale = engine.Difficulties[gs.Game.Difficulty].UFOScale
		}
		ufoSpawnRate = int(float64(ufoSpawnRate) / diffUFOScale)
		if ufoSpawnRate < ufoSpawnRateFloor {
			ufoSpawnRate = ufoSpawnRateFloor
		}
		if gs.TickCounter%ufoSpawnRate == 0 {
			maxUFOs := 5 + gameMonth/2
			if maxUFOs > 12 {
				maxUFOs = 12
			}
			if gs.UFOs.Count() < maxUFOs {
				ufo := SpawnUFOOnCities(gs.Cities, gameMonth)
				gs.UFOs = append(gs.UFOs, ufo)
				gs.Message = fmt.Sprintf(language.String("MSG_UFO_DETECTED"), ufo.Type.DisplayName(), gs.cityName(ufo.CurrentNode()))
				gs.MessageTimer = time.Now()
				audio.PlayAlert()
				if engine.Config.PauseOnAlienDetect {
					gs.Game.Paused = true
				}
			}
		}

		// Mission Spawning: Trigger events based on AlienActivity and game time.
		spawnRate := missionSpawnBase - (gs.AlienActivity * missionSpawnActWt) - gameMonth*missionSpawnDecay
		if spawnRate < missionSpawnFloor {
			spawnRate = missionSpawnFloor
		}
		if gs.TickCounter%spawnRate == 0 {
			gs.spawnMission()
		}

		// Periodic increase in overall alien activity.
		if gs.TickCounter%activityTickRate == 0 {
			gs.AlienActivity++
		}

		// Alien Base Establishment: attempt to build new bases periodically.
		baseSpawnRate := baseSpawnRateBase - gameMonth*baseSpawnRateDecay
		if baseSpawnRate < baseSpawnRateFloor {
			baseSpawnRate = baseSpawnRateFloor
		}
		if gs.TickCounter%baseSpawnRate == 0 {
			gs.tryEstablishBase(gameMonth)
		}

		// Alien Base Ticking: missions, defenders, threat escalation.
		gs.tickAlienBases(gameMonth)

		// Mission Timer Update: decrease remaining time and trigger consequences if expired.
		remaining := make([]*AlienMission, 0, len(gs.Missions))
		for _, m := range gs.Missions {
			m.HoursLeft -= float64(minutes) / 60.0
			if m.HoursLeft <= 0 {
				cityNameStr := gs.cityName(m.NodeID)
				// Base defense mission that expired: the aliens overrun the base
				if defBase := gs.HasBaseAt(m.NodeID); defBase != nil {
					gs.Message = fmt.Sprintf(language.String("MSG_BASE_DESTROYED"), defBase.Name)
					gs.MessageTimer = time.Now()
					gs.destroyBase(defBase)
				} else {
					gs.Message = fmt.Sprintf(language.String("MSG_ATTACK_CITY"), m.Type, cityNameStr)
					gs.MessageTimer = time.Now()
					gs.AlienActivity += 10
				}
			} else {
				remaining = append(remaining, m)
			}
		}
		gs.Missions = remaining

		// Update UFOs along network edges
		for _, u := range gs.UFOs {
			u.Update(gs.Cities)
		}

		// Prune inactive UFOs
		activeUFOs := make(UFOList, 0, len(gs.UFOs))
		for _, u := range gs.UFOs {
			if u.Active {
				activeUFOs = append(activeUFOs, u)
			}
		}
		gs.UFOs = activeUFOs

		// Advance dogfight animation if one is playing
		gs.updateDogfightVisual()

		for _, i := range gs.Interceptors {
			if i.Launching {
				// Skip if this interceptor is already in a dogfight animation
				if gs.DogfightVisual != nil && gs.DogfightVisual.Active && gs.DogfightVisual.Interceptor == i {
					continue
				}
				reached := i.Update(gs.Cities, gs.UFOs)
				if reached {
					gs.dogfight(i)
				}
			}
		}

		// Prune destroyed interceptors
		activeInterceptors := make(InterceptorList, 0, len(gs.Interceptors))
		for _, i := range gs.Interceptors {
			if i.HP > 0 {
				activeInterceptors = append(activeInterceptors, i)
			}
		}
		gs.Interceptors = activeInterceptors

		if gs.Transport != nil {
			t := gs.Transport
			// Move transport along path
			if !t.Returning {
				// Move toward crash site
				path := gs.ShortestPath(t.FromNode, t.ToNode)
				if len(path) > 1 {
					nextCity := gs.CityByID(path[1])
					if nextCity != nil {
						t.Progress += 0.05
						if t.Progress >= 1.0 {
							t.Progress = 0
							t.FromNode = path[1]
							// Check if we arrived at crash site
							if t.FromNode == t.ToNode {
								cs := t.CrashSite
								if cs != nil && !cs.Looted {
									// Start tactical battle
									gs.Transport = nil
									selectedBase := gs.SelectedBase()
									if selectedBase == nil {
										return
									}
									healthy := selectedBase.HealthySoldiers()
									if len(healthy) > 0 {
										gs.Game.Paused = true
										gs.ActiveCrashSite = cs
										gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_RETRIEVED"), localizeUFOName(cs.UFOName), 0)
										gs.MessageTimer = time.Now()
										gs.PreBattleStats = make(map[string][6]int)
										for _, s := range healthy {
											gs.PreBattleStats[s.Name] = [6]int{s.HP, s.Accuracy, s.Reactions, s.Strength, s.Bravery, s.TU}
										}
										city := gs.CityByID(cs.NodeID)
									cx, cy := -1, -1
									if city != nil {
										cx, cy = city.X, city.Y
									}
									bs := battle.NewBattlescape(gs.Game, selectedBase, healthy, localizeUFOName(cs.UFOName), cs.Seed, cx, cy)
										gs.Game.SetScreen(engine.StateBattlescape, bs)
										gs.Game.PushState(engine.StateBattlescape)
										return
									}
									gs.Message = language.String("MSG_NO_SOLDIERS")
									gs.MessageTimer = time.Now()
								}
								t.Returning = true
								t.ToNode = t.SourceBaseCity
							}
						}
					}
				}
			} else {
				// Return to base
				path := gs.ShortestPath(t.FromNode, t.ToNode)
				if len(path) > 1 {
					t.Progress += 0.05
					if t.Progress >= 1.0 {
						t.Progress = 0
						t.FromNode = path[1]
						if t.FromNode == t.ToNode {
							gs.Transport = nil
							gs.Message = language.String("MSG_TRANSPORT_RETURNED")
							gs.MessageTimer = time.Now()
						}
					}
				} else {
					gs.Transport = nil
				}
			}
		}

		gs.Game.GameTime = gs.Game.GameTime.Add(time.Duration(minutes) * time.Minute)

		// Daily AdvanceDay for all bases
		curDay := gs.Game.GameTime.YearDay()
		if curDay != gs.LastDay {
			gs.LastDay = curDay
			for _, b := range gs.Bases {
				b.AdvanceDay()
			}
		}

		// Advance research and manufacturing for ALL bases (not just the selected one)
		if gs.TickCounter%30 == 0 {
			var msgs []string
			for _, b := range gs.Bases {
				done := b.AdvanceResearch()
				for _, name := range done {
					audio.PlayResearchComplete()
					msgs = append(msgs, fmt.Sprintf(language.String("MSG_RESEARCH_COMPLETE"), name))
				}
				crafted := b.AdvanceManufacture()
				for _, item := range crafted {
					audio.PlayManufactureComplete()
					msgs = append(msgs, fmt.Sprintf(language.String("MSG_MANUFACTURE_COMPLETE"), item))
				}
			}
			if len(msgs) > 0 {
				gs.Message = msgs[0]
				gs.MessageTimer = time.Now()
			}
		}
	}
	// Monthly budget check
	curMonth := int(gs.Game.GameTime.Month())
	if curMonth != gs.LastMonth {
		gs.LastMonth = curMonth
		totalFunding := 0
		totalSalary := 0
		for _, b := range gs.Bases {
			b.AlienActivity = gs.AlienActivity
			salary, funding := b.AdvanceMonth()
			totalFunding += funding
			totalSalary += salary
		}
		gs.Game.Funds += int64(totalFunding - totalSalary)
		gs.Message = fmt.Sprintf(language.String("MSG_MONTHLY_REPORT"), totalFunding/1000, totalSalary/1000)
		gs.MessageTimer = time.Now()
		if engine.Config.AutosaveEnabled {
			gs.SaveGameAuto()
		}
	}
}

func (gs *Geoscape) destroyBase(b *base.Base) {
	idx := -1
	for i, base := range gs.Bases {
		if base == b {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	if len(gs.Bases) <= 1 {
		gs.Message = fmt.Sprintf(language.String("MSG_BASE_DESTROYED"), b.Name)
		gs.Game.GameOver(false, language.String("GEO_LAST_BASE_DESTROYED"))
		gs.Defeated = true
		gs.Game.Paused = true
		return
	}
	gs.Bases = append(gs.Bases[:idx], gs.Bases[idx+1:]...)
	if gs.ActiveBase >= len(gs.Bases) {
		gs.ActiveBase = len(gs.Bases) - 1
	}
	if b.CityID >= 0 && b.CityID < len(gs.Cities) {
		gs.Cities[b.CityID].HasRadar = false
	}
	gs.Message = fmt.Sprintf(language.String("MSG_BASE_DESTROYED"), b.Name)
	gs.MessageTimer = time.Now()
}

// applyMissionRewards grants mission-specific bonus loot and funding when a
// geoscape mission battle is won.
func (gs *Geoscape) applyMissionRewards(b *base.Base) {
	switch gs.ActiveMissionType {
	case language.String("MISSION_COUNCIL"):
		bonus := 100000
		gs.Game.Funds += int64(bonus)
		b.AddLoot([]string{"alloys", "alloys", "elerium"})
		gs.Message = fmt.Sprintf(language.String("MSG_COUNCIL_REWARD"), bonus/1000)
	case language.String("MISSION_SUPPLY"):
		b.AddLoot([]string{"alloys", "alloys", "alloys", "elerium", "ufo_nav"})
		gs.Message = language.String("MSG_SUPPLY_RAID_LOOT")
	case language.String("MISSION_RESEARCH"):
		b.AddLoot([]string{"alloys", "elerium", "ufo_power", "ufo_weapon"})
		gs.Message = language.String("MSG_RESEARCH_LOOT")
	}
}

func (gs *Geoscape) HasBaseAt(cityID int) *base.Base {
	for _, b := range gs.Bases {
		if b.CityID == cityID {
			return b
		}
	}
	return nil
}

func (gs *Geoscape) CycleBase() {
	if len(gs.Bases) <= 1 {
		gs.Message = language.String("MSG_ONLY_ONE_BASE")
		gs.MessageTimer = time.Now()
		return
	}
	gs.ActiveBase = (gs.ActiveBase + 1) % len(gs.Bases)
	if sb := gs.SelectedBase(); sb != nil {
		gs.Message = fmt.Sprintf(language.String("MSG_SWITCHED_BASE"), sb.Name)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) BuildBase() {
	city := gs.CityByID(gs.CursorNode)
	if city == nil {
		return
	}
	if gs.HasBaseAt(gs.CursorNode) != nil {
		gs.Message = fmt.Sprintf(language.String("MSG_BASE_EXISTS"), city.LangName())
		gs.MessageTimer = time.Now()
		return
	}
	cost := int64(500000)
	if gs.Game.Funds < cost {
		gs.Message = language.String("MSG_INSUFFICIENT_FUNDS")
		gs.MessageTimer = time.Now()
		return
	}
	gs.Game.Funds -= cost
	baseNum := len(gs.Bases) + 1
	b := base.NewBase(fmt.Sprintf(language.String("GEO_BASE_NAME_FMT"), baseNum), gs.CursorNode)
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 2})
	gs.Bases = append(gs.Bases, b)
	city.HasRadar = true
	gs.ActiveBase = len(gs.Bases) - 1
	gs.Message = fmt.Sprintf(language.String("MSG_BASE_BUILT"), b.Name, city.LangName())
	gs.MessageTimer = time.Now()
}
func (gs *Geoscape) CityByID(id int) *City {
	for _, c := range gs.Cities {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (gs *Geoscape) ShortestPath(from, to int) []int {
	if from == to {
		return []int{from}
	}
	type item struct {
		id   int
		path []int
	}
	queue := []item{{id: from, path: []int{from}}}
	visited := map[int]bool{from: true}
	maxEdgeDist := maxEdgePathDist

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		curCity := gs.CityByID(cur.id)
		if curCity == nil {
			continue
		}

		for _, n := range gs.Cities {
			if visited[n.ID] {
				continue
			}
			dx := float64(n.X - curCity.X)
			dy := float64(n.Y - curCity.Y)
			if dx*dx+dy*dy > maxEdgeDist*maxEdgeDist {
				continue
			}
			newPath := make([]int, len(cur.path)+1)
			copy(newPath, cur.path)
			newPath[len(cur.path)] = n.ID

			if n.ID == to {
				return newPath
			}
			visited[n.ID] = true
			queue = append(queue, item{id: n.ID, path: newPath})
		}
	}
	return nil
}

func (gs *Geoscape) updateDogfightVisual() {
	dv := gs.DogfightVisual
	if dv == nil || !dv.Active {
		return
	}

	dv.Timer--
	if dv.HitFlash > 0 {
		dv.HitFlash--
	}
	if dv.UFOHitFlash > 0 {
		dv.UFOHitFlash--
	}

	if dv.Timer > 0 {
		return
	}

	// Animation finished — display result
	inter := dv.Interceptor
	ufo := dv.UFO
	damage := dv.InterDamage

	if damage == 0 {
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_MISS"), inter.Weapon.Name)
	} else if !dv.UFOAlive {
		city := gs.CityByID(ufo.CurrentNode())
		if city != nil && GetTile(city.X, city.Y) == 0 {
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_LOST_AT_SEA"), ufo.Type.DisplayName())
		} else {
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_CRASHED"), ufo.Type.DisplayName())
		}
	} else {
		gs.Message = fmt.Sprintf(language.String("MSG_HIT_UFO"), damage)
	}
	if dv.UFODamage > 0 {
		gs.Message += fmt.Sprintf(language.String("MSG_UFO_HIT_INTERCEPTOR"), dv.UFODamage, inter.HP, inter.MaxHP)
	}
	if dv.InterDestroyed {
		gs.Message += language.String("GEO_INTERCEPTOR_DESTROYED")
	}

	gs.MessageTimer = time.Now()

	// Disengage interceptor if destroyed or target down
	if dv.InterDestroyed || !dv.UFOAlive {
		inter.Disengage()
	}

	gs.DogfightVisual = nil
}

func (gs *Geoscape) dogfight(inter *Interceptor) {
	ufo := inter.TargetUFO
	if ufo == nil {
		for _, u := range gs.UFOs {
			if u.Active && u.CurrentNode() == inter.TargetNode {
				ufo = u
				break
			}
		}
		if ufo == nil {
			return
		}
	}

	dist := math.Sqrt(math.Pow(ufo.X-inter.X, 2) + math.Pow(ufo.Y-inter.Y, 2))
	if dist > float64(inter.Range) {
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_CLOSING"), inter.Weapon.Name, inter.Mode.String())
		gs.MessageTimer = time.Now()
		return
	}

	// Resolve combat immediately
	damage := inter.FireAt(ufo)
	audio.PlayShoot()

	var ufoDmg int
	var interDestroyed bool

	if damage == -1 {
		// UFO destroyed — handle crash site / funding immediately
		gs.Game.Funds += int64(ufo.Type.Points * 1000)
		city := gs.CityByID(ufo.CurrentNode())
		if city != nil && GetTile(city.X, city.Y) == 0 {
			// lost at sea
		} else {
			gs.CrashSites = append(gs.CrashSites, &CrashSite{
				UFOName: ufo.Type.Name,
				NodeID:  ufo.CurrentNode(),
				Seed:    rand.Int63(),
			})
		}
		if inter.State != nil {
			inter.State.Status = "available"
			inter.State.HP = inter.HP
		}
	}

	if ufo.Active && inter.HP > 0 {
		ufoDmg = ufo.FireAtInterceptor(inter)
		audio.PlayPlasmaFire()
		if inter.State != nil {
			inter.State.HP = inter.HP
		}
	}

	if inter.HP <= 0 {
		interDestroyed = true
		if inter.State != nil {
			inter.State.HP = 0
			inter.State.Status = "destroyed"
		}
	}

	// Set up animation state (messages are displayed when animation ends)
	ufoMaxHP := ufo.Type.MaxHP
	if ufoMaxHP <= 0 {
		ufoMaxHP = ufo.Type.Toughness
	}
	ufoHP := ufo.Type.Toughness
	if ufoHP < 0 {
		ufoHP = 0
	}
	hitFlash := 0
	if ufoDmg > 0 {
		hitFlash = 8
	}
	ufoHitFlash := 0
	if damage > 0 || damage == -1 {
		ufoHitFlash = 6
	}

	gs.DogfightVisual = &DogfightAnim{
		Active:      true,
		Timer:       28,
		Interceptor: inter,
		UFO:         ufo,

		InterHP:    inter.HP,
		InterMaxHP: inter.MaxHP,
		UFOHP:      ufoHP,
		UFOMaxHP:   ufoMaxHP,

		InterDamage:    damage,
		UFODamage:      ufoDmg,
		UFOAlive:       ufo.Active,
		InterAlive:     inter.HP > 0,
		InterDestroyed: interDestroyed,

		HitFlash:    hitFlash,
		UFOHitFlash: ufoHitFlash,
	}
}

func (gs *Geoscape) spawnMission() {
	if gs.SelectedBase() == nil {
		return
	}
	// Weighted mission pool. Common missions appear more often; alien base
	// assaults, research raids, and council requests are rarer and more
	// significant.
	type weighted struct {
		typ    string
		weight int
	}
	pool := []weighted{
		{language.String("MISSION_TERROR"), 30},
		{language.String("MISSION_SUPPLY"), 22},
		{language.String("MISSION_ABDUCTION"), 22},
		{language.String("MISSION_RESEARCH"), 14},
		{language.String("MISSION_COUNCIL"), 8},
		{language.String("MISSION_BUILDING"), 8},
		{language.String("MISSION_ALIEN_BASE"), 4},
	}
	total := 0
	for _, w := range pool {
		total += w.weight
	}
	pick := rand.Intn(total)
	chosen := pool[0].typ
	for _, w := range pool {
		if pick < w.weight {
			chosen = w.typ
			break
		}
		pick -= w.weight
	}

	var target *City

	if chosen == language.String("MISSION_ALIEN_BASE") && len(gs.AlienBases) > 0 {
		// Target an actual alien base, not a random city
		ab := gs.AlienBases[rand.Intn(len(gs.AlienBases))]
		target = gs.CityByID(ab.CityID)
	} else {
		// Build candidate list. If the player has multiple bases, aliens may
		// directly assault a base (base defense scenario).
		var candidates []*City
		for _, c := range gs.Cities {
			if c.ID == gs.SelectedBase().CityID {
				if len(gs.Bases) > 1 && rand.Intn(100) < 25 {
					candidates = append(candidates, c)
				}
				continue
			}
			candidates = append(candidates, c)
		}
		if len(candidates) == 0 {
			return
		}
		target = candidates[rand.Intn(len(candidates))]
	}
	if target == nil {
		return
	}

	turnsLeft := 24.0 // 24 game hours to respond
	if chosen == language.String("MISSION_ALIEN_BASE") {
		turnsLeft = 12.0 // 12 game hours for base assaults
	} else if chosen == language.String("MISSION_COUNCIL") {
		turnsLeft = 36.0 // council gives more time but offers a bonus
	}
	mission := &AlienMission{
		Type:      chosen,
		NodeID:    target.ID,
		HoursLeft: turnsLeft,
	}
	gs.Missions = append(gs.Missions, mission)
	target.MissionHere = true
	gs.Message = fmt.Sprintf(language.String("MSG_ALERT_MISSION"), chosen, target.LangName())
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayAlert()
}

// tryEstablishBase attempts to create a new alien base at a suitable city.
// Bases become more frequent and tougher as the campaign progresses.
func (gs *Geoscape) tryEstablishBase(gameMonth int) {
	maxBases := 1 + gameMonth/2
	if maxBases > 10 {
		maxBases = 10
	}
	if len(gs.AlienBases) >= maxBases {
		return
	}
	// Find cities without an existing alien base
	var candidates []*City
	for _, c := range gs.Cities {
		if gs.hasAlienBaseAt(c.ID) {
			continue
		}
		candidates = append(candidates, c)
	}
	if len(candidates) == 0 {
		return
	}
	// Prefer higher-threat cities
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Threat > candidates[j].Threat
	})
	maxIdx := len(candidates) / 3
	if maxIdx < 1 {
		maxIdx = 1
	}
	target := candidates[rand.Intn(maxIdx)]
	baseThreat := 20 + gameMonth*5
	if baseThreat > 80 {
		baseThreat = 80
	}
	ab := &AlienBase{
		CityID:          target.ID,
		Threat:          baseThreat,
		TurnsAlive:      0,
		LastMissionTick: gs.TickCounter,
		DefendingUFOID:  -1,
		Name:            fmt.Sprintf(language.String("GEO_ALIEN_BASE_NAME"), len(gs.AlienBases)+1),
	}
	gs.AlienBases = append(gs.AlienBases, ab)
	target.Threat += 15
	if target.Threat > 100 {
		target.Threat = 100
	}
	gs.Message = fmt.Sprintf(language.String("GEO_ALIEN_BASE_ALERT"), ab.Name, target.LangName())
	gs.MessageTimer = time.Now()
	audio.PlayAlert()

	// Spawn a defending UFO for the new base
	gs.spawnBaseDefender(ab)
}

// tickAlienBases processes per-tick alien base actions: mission spawning and defender UFOs.
func (gs *Geoscape) tickAlienBases(gameMonth int) {
	for _, ab := range gs.AlienBases {
		if ab == nil {
			continue
		}
		ab.TurnsAlive++

		// Periodically escalate base threat
		if ab.TurnsAlive%alienActivityTick == 0 && ab.Threat < 100 {
			ab.Threat += 5
		}

		// Spawn missions from this base
		missionInterval := missionIntervalBase - gameMonth*missionIntervalDecay
		if missionInterval < missionIntervalFloor {
			missionInterval = missionIntervalFloor
		}
		if gs.TickCounter-ab.LastMissionTick >= missionInterval {
			gs.spawnMissionFromBase(ab)
			ab.LastMissionTick = gs.TickCounter
		}

		// Maintain a defending UFO near the base
		if ab.DefendingUFOID >= 0 {
			ufo := gs.findUFOByID(ab.DefendingUFOID)
			if ufo == nil || !ufo.Active {
				ab.DefendingUFOID = -1
			}
		}
		if ab.DefendingUFOID < 0 && gs.UFOs.Count() < 12 {
			gs.spawnBaseDefender(ab)
		}
	}
}

// spawnBaseDefender creates a UFO that patrols near the alien base.
func (gs *Geoscape) spawnBaseDefender(ab *AlienBase) {
	city := gs.CityByID(ab.CityID)
	if city == nil {
		return
	}
	diff := gs.Game.Difficulty
	if diff < 0 {
		diff = 0
	}
	ufo := SpawnUFOAtCity(city, gs.Cities, diff)
	if ufo == nil {
		return
	}
	ufo.TurnsLeft = 9999 // indefinite patrol
	gs.UFOs = append(gs.UFOs, ufo)
	ab.DefendingUFOID = ufo.ID
}

// spawnMissionFromBase creates a mission originating from this alien base.
func (gs *Geoscape) spawnMissionFromBase(ab *AlienBase) {
	missionTypes := []string{
		language.String("MISSION_SUPPLY"),
		language.String("MISSION_TERROR"),
		language.String("MISSION_ABDUCTION"),
		language.String("MISSION_RESEARCH"),
		language.String("MISSION_BUILDING"),
	}
	chosen := missionTypes[rand.Intn(len(missionTypes))]
	// Target a different city from the base city (or a random one)
	var candidates []*City
	for _, c := range gs.Cities {
		if c.ID != ab.CityID {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return
	}
	target := candidates[rand.Intn(len(candidates))]
	mission := &AlienMission{
		Type:      chosen,
		NodeID:    target.ID,
		HoursLeft: 24.0,
	}
	gs.Missions = append(gs.Missions, mission)
	target.MissionHere = true
	gs.Message = fmt.Sprintf(language.String("GEO_ALIEN_BASE_LAUNCH"),
		ab.Name, chosen, gs.CityByID(ab.CityID).LangName(), target.LangName())
	gs.MessageTimer = time.Now()
}

// hasAlienBaseAt checks if an alien base exists at the given city node.
func (gs *Geoscape) hasAlienBaseAt(cityID int) bool {
	for _, ab := range gs.AlienBases {
		if ab.CityID == cityID {
			return true
		}
	}
	return false
}

// alienBaseAt returns the alien base at a city, or nil.
func (gs *Geoscape) alienBaseAt(cityID int) *AlienBase {
	for _, ab := range gs.AlienBases {
		if ab.CityID == cityID {
			return ab
		}
	}
	return nil
}

// destroyAlienBase removes an alien base and reduces regional threat.
func (gs *Geoscape) destroyAlienBase(ab *AlienBase) {
	city := gs.CityByID(ab.CityID)
	if city != nil {
		city.Threat -= 20
		if city.Threat < 0 {
			city.Threat = 0
		}
	}
	// Find and remove from slice
	for i, b := range gs.AlienBases {
		if b == ab {
			gs.AlienBases = append(gs.AlienBases[:i], gs.AlienBases[i+1:]...)
			break
		}
	}
}

// findUFOByID returns a UFO with the given ID, or nil.
func (gs *Geoscape) findUFOByID(id int) *UFO {
	for _, u := range gs.UFOs {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (gs *Geoscape) triggerCydonia() {
	if gs.CydoniaTriggered {
		return
	}
	gs.CydoniaTriggered = true
	gs.Message = language.String("GEO_CYDONIA_DETECTED")
	gs.MessageTimer = time.Now()

	// Add Cydonia as a special mission
	mission := &AlienMission{
		Type:      language.String("GEO_CYDONIA"), // Reuse for Cydonia
		NodeID:    0,                              // Special node for Cydonia
		HoursLeft: 9999.0,                         // Indefinite
	}
	gs.Missions = append(gs.Missions, mission)
	gs.Game.Bell()
}

func (gs *Geoscape) RespondToMission(idx int) {
	if idx < 0 || idx >= len(gs.Missions) {
		gs.Message = language.String("MSG_INVALID_MISSION")
		gs.MessageTimer = time.Now()
		return
	}
	if gs.SelectedBase() == nil {
		return
	}
	aliveCount := 0
	for _, s := range gs.SelectedBase().Soldiers {
		if s.CanDeploy() {
			aliveCount++
		}
	}
	if aliveCount == 0 {
		gs.Message = language.String("MSG_NO_HEALTHY_SOLDIERS")
		gs.MessageTimer = time.Now()
		return
	}
	mission := gs.Missions[idx]
	gs.Missions = append(gs.Missions[:idx], gs.Missions[idx+1:]...)
	if city := gs.CityByID(mission.NodeID); city != nil {
		city.MissionHere = false
	}

	// Base defense mission if the target city hosts one of our bases
	if defBase := gs.HasBaseAt(mission.NodeID); defBase != nil {
		gs.ActiveBaseDefense = defBase
	}
	gs.Message = fmt.Sprintf(language.String("MSG_SQUAD_DEPLOYED"), mission.Type, gs.cityName(mission.NodeID))
	gs.MessageTimer = time.Now()
	gs.Game.Paused = true

	defBase := gs.SelectedBase()
	if gs.ActiveBaseDefense != nil {
		defBase = gs.ActiveBaseDefense
	}

	healthy := defBase.HealthySoldiers()
	if len(healthy) == 0 {
		gs.Message = language.String("MSG_NO_HEALTHY_SOLDIERS")
		gs.MessageTimer = time.Now()
		return
	}

	gs.PreBattleStats = make(map[string][6]int)
	for _, s := range healthy {
		gs.PreBattleStats[s.Name] = [6]int{s.HP, s.Accuracy, s.Reactions, s.Strength, s.Bravery, s.TU}
	}

	ufoName := language.String("MISSION_CRASH_SITE")
	switch mission.Type {
	case language.String("MISSION_TERROR"):
		ufoName = language.String("MISSION_TYPE_TERROR")
	case language.String("MISSION_SUPPLY"):
		ufoName = language.String("MISSION_TYPE_SUPPLY")
	case language.String("MISSION_ALIEN_BASE"):
		ufoName = language.String("MISSION_TYPE_BASE")
		gs.respondedAlienBase = gs.alienBaseAt(mission.NodeID)
	case language.String("MISSION_ABDUCTION"):
		ufoName = language.String("MISSION_TYPE_ABDUCTION")
	case language.String("MISSION_RESEARCH"):
		ufoName = language.String("MISSION_TYPE_RESEARCH")
	case language.String("MISSION_COUNCIL"):
		ufoName = language.String("MISSION_TYPE_COUNCIL")
	case language.String("MISSION_BUILDING"):
		ufoName = language.String("MISSION_TYPE_BUILDING")
	}
	if mission.NodeID == 0 {
		ufoName = language.String("GEO_CYDONIA")
		gs.ActiveFinalMission = true
	}
	if gs.ActiveBaseDefense != nil {
		ufoName = language.String("MISSION_TYPE_BASE")
	}
	gs.ActiveMissionType = mission.Type
	bs := battle.NewBattlescape(gs.Game, defBase, healthy, ufoName, 0, -1, -1)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
}

func (gs *Geoscape) AutoresolveMission(idx int) {
	if idx < 0 || idx >= len(gs.Missions) {
		return
	}
	if gs.SelectedBase() == nil {
		return
	}
	mission := gs.Missions[idx]
	gs.Missions = append(gs.Missions[:idx], gs.Missions[idx+1:]...)

	defBase := gs.SelectedBase()
	if gs.ActiveBaseDefense != nil {
		defBase = gs.ActiveBaseDefense
	}
	healthy := defBase.HealthySoldiers()
	if len(healthy) == 0 {
		gs.Message = language.String("MSG_NO_HEALTHY_SOLDIERS")
		gs.MessageTimer = time.Now()
		return
	}

	if city := gs.CityByID(mission.NodeID); city != nil {
		city.MissionHere = false
	}

	alienCount := 5 + gs.MissionsWon/2
	if alienCount > 10 {
		alienCount = 10
	}

	missionTypeMod := 0
	switch mission.Type {
	case language.String("MISSION_TERROR"):
		missionTypeMod = -10
	case language.String("MISSION_COUNCIL"):
		missionTypeMod = 10
	case language.String("MISSION_ALIEN_BASE"):
		missionTypeMod = -15
	}
	winChance := gs.calcWinChance(healthy, missionTypeMod)

	won := rand.Intn(100) < winChance

	if won {
		reward := int64(25000 + gs.MissionsWon*2000)
		gs.Game.Funds += reward
		gs.MissionsWon++

		for _, s := range healthy {
			s.PostMission()
			s.Missions++
			s.Fatigue += 2
		}
		soldier.HandlePromotions(defBase.Soldiers)

		weaponDrops := make(map[string]bool)
		deadAliens := make(map[string]bool)
		for i := 0; i < alienCount; i++ {
			if rand.Intn(100) < 25 {
				alienTypes := gs.Game.GetAlienTypes()
				if len(alienTypes) > 0 {
					at := alienTypes[rand.Intn(len(alienTypes))]
					weaponDrops[at.Weapon] = true
					deadAliens[at.Name] = true
				}
			}
		}
		for wpn := range weaponDrops {
			if _, ok := data.RuleItems[wpn]; ok {
				defBase.AddItem(wpn, 1)
			}
		}
		for name := range deadAliens {
			gs.Game.LearnAlien(name, 2)
		}

		gs.Message = fmt.Sprintf(language.String("GEO_AUTORESOLVE_VICTORY"),
			gs.cityName(mission.NodeID), reward/1000, alienCount)
	} else {
		casualtyCount := 1 + rand.Intn(min(3, len(healthy)))
		if casualtyCount > len(healthy) {
			casualtyCount = len(healthy)
		}
		killed := make([]string, 0, casualtyCount)
		shuffled := make([]*soldier.Soldier, len(healthy))
		copy(shuffled, healthy)
		rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
		for i := 0; i < casualtyCount && i < len(shuffled); i++ {
			shuffled[i].HP = 0
			shuffled[i].Wounds = 30
			killed = append(killed, shuffled[i].Name)
		}
		defBase.RemoveDeadSoldiers()

		for _, s := range healthy {
			if s.HP > 0 {
				s.Fatigue += 3
			}
		}

		gs.Message = fmt.Sprintf(language.String("GEO_AUTORESOLVE_DEFEAT"),
			gs.cityName(mission.NodeID), casualtyCount, strings.Join(killed, ", "))
	}

	gs.ActiveMissionType = ""
	gs.Game.Paused = false
	gs.MessageTimer = time.Now()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (gs *Geoscape) enterMissionSelectMode() {
	idx := gs.missionIndexAtCursor()
	if idx < 0 {
		gs.Message = language.String("GEO_NO_MISSION_HERE")
		gs.MessageTimer = time.Now()
		return
	}

	defBase := gs.SelectedBase()
	if defBase == nil {
		return
	}
	healthy := defBase.HealthySoldiers()
	if len(healthy) == 0 {
		gs.Message = language.String("MSG_NO_HEALTHY_SOLDIERS")
		gs.MessageTimer = time.Now()
		return
	}

	winChance := gs.calcWinChance(healthy, 0)

	gs.MissionSelectMode = true
	gs.MissionSelectIdx = 0
	gs.MissionOdds = winChance
	gs.Game.Paused = true
}

// RespondToSelectedMission responds to the mission at the cursor's node if one
// exists, otherwise the first available mission.
func (gs *Geoscape) RespondToSelectedMission() {
	idx := gs.missionIndexAtCursor()
	if idx < 0 {
		idx = 0
	}
	gs.RespondToMission(idx)
}

func (gs *Geoscape) missionIndexAtCursor() int {
	for i, m := range gs.Missions {
		if m.NodeID == gs.CursorNode {
			return i
		}
	}
	return -1
}

func (gs *Geoscape) Autoresolve() {
	if gs.SelectedBase() == nil {
		return
	}
	var nearest *UFO
	bestDist := math.MaxFloat64
	for _, u := range gs.UFOs.Active() {
		city := gs.CityByID(u.CurrentNode())
		if city == nil {
			continue
		}
		baseCity := gs.CityByID(gs.SelectedBase().CityID)
		if baseCity == nil {
			continue
		}
		dx := float64(city.X - baseCity.X)
		dy := float64(city.Y - baseCity.Y)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			nearest = u
		}
	}
	if nearest == nil {
		gs.Message = language.String("MSG_NO_UFO_AUTO")
		gs.MessageTimer = time.Now()
		return
	}

	aliveCount := 0
	for _, s := range gs.SelectedBase().Soldiers {
		if s.CanDeploy() {
			aliveCount++
		}
	}
	squadSize := aliveCount
	chance := 30 + squadSize*10
	if chance > 85 {
		chance = 85
	}
	won := rand.Intn(100) < chance

	nearest.Active = false
	if won {
		gs.Game.Funds += int64(nearest.Type.Points * 1000)
		gs.Message = fmt.Sprintf(language.String("MSG_AUTO_VICTORY"), nearest.Type.DisplayName(), nearest.Type.Points)
		} else {
			if squadSize > 0 {
				// build list of alive soldiers
				var alive []*soldier.Soldier
				for _, s := range gs.SelectedBase().Soldiers {
					if s.HP > 0 {
						alive = append(alive, s)
					}
				}
				if len(alive) == 0 {
					gs.Message = fmt.Sprintf(language.String("MSG_AUTO_DEFEAT"), nearest.Type.DisplayName())
					gs.MessageTimer = time.Now()
					nearest.Active = false
					return
				}
				idx := rand.Intn(len(alive))
			alive[idx].HP = 0
			gs.SelectedBase().RemoveDeadSoldiers()
			gs.Message = fmt.Sprintf(language.String("MSG_AUTO_DEFEAT"), nearest.Type.DisplayName())
		} else {
			gs.Message = language.String("MSG_AUTO_NO_SOLDIERS")
		}
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) buildSaveData() *save.SaveData {
	ufoSaves := make([]*save.UFOSave, 0)
	for _, u := range gs.UFOs {
		ufoSaves = append(ufoSaves, &save.UFOSave{
			ID:        u.ID,
			TypeName:  u.Type.Name,
			X:         u.X,
			Y:         u.Y,
			Progress:  u.Progress,
			NodeFrom:  u.NodeFrom,
			NodeTo:    u.NodeTo,
			TurnsLeft: u.TurnsLeft,
			Active:    u.Active,
		})
	}
	missionSaves := make([]*save.MissionSave, 0)
	for _, m := range gs.Missions {
		missionSaves = append(missionSaves, &save.MissionSave{
			Type:      m.Type,
			CityName:  "",
			HoursLeft: m.HoursLeft,
			X:         m.NodeID,
			Y:         0,
		})
	}
	var baseSaves []*save.BaseSave
	for _, b := range gs.Bases {
		baseSaves = append(baseSaves, save.FromBase(b))
	}
	alienBaseSaves := make([]*save.AlienBaseSave, 0)
	for _, ab := range gs.AlienBases {
		alienBaseSaves = append(alienBaseSaves, &save.AlienBaseSave{
			CityID:          ab.CityID,
			Threat:          ab.Threat,
			TurnsAlive:      ab.TurnsAlive,
			LastMissionTick: ab.LastMissionTick,
			DefendingUFOID:  ab.DefendingUFOID,
			Name:            ab.Name,
		})
	}
	return &save.SaveData{
		GameTime:       gs.Game.GameTime,
		Funds:          gs.Game.Funds,
		Paused:         gs.Game.Paused,
		TimeSpeed:      gs.Game.TimeSpeed,
		Difficulty:     gs.Game.Difficulty,
		AlienActivity:  gs.AlienActivity,
		SpeciesSeed:    gs.Game.SpeciesSeed,
		AlienKnowledge: gs.Game.AlienKnowledge,
		Bases:          baseSaves,
		UFOs:           ufoSaves,
		Missions:       missionSaves,
		MissionsWon:    gs.MissionsWon,
		AlienBases:     alienBaseSaves,
	}
}

func (gs *Geoscape) SaveGameToFile() {
	sd := gs.buildSaveData()
	err := save.SaveGame("xcom_save.json", sd)
	if err != nil {
		gs.Message = language.String("MSG_SAVE_FAILED") + err.Error()
	} else {
		gs.Message = language.String("MSG_GAME_SAVED_AUTO")
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SaveGameToSlot(slot int) {
	sd := gs.buildSaveData()
	err := save.SaveGameToSlot(slot, sd)
	if err != nil {
		gs.Message = language.String("MSG_SAVE_FAILED") + err.Error()
	} else {
		gs.Message = fmt.Sprintf(language.String("SLOT_PICKER_SAVED"), slot)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SaveGameAuto() {
	sd := gs.buildSaveData()
	if err := save.SaveGame(save.AutoSavePath(), sd); err != nil {
		gs.Message = language.String("MSG_SAVE_FAILED") + err.Error()
		gs.MessageTimer = time.Now()
	}
}

func (gs *Geoscape) LoadGameFromFile() {
	sd, err := save.LoadGame("xcom_save.json")
	if err != nil {
		gs.Message = language.String("MSG_LOAD_FAILED") + err.Error()
		gs.MessageTimer = time.Now()
		return
	}
	gs.loadFromSaveData(sd)
}

func (gs *Geoscape) LoadGameFromSlot(slot int) {
	sd, err := save.LoadGame(save.SavePath(slot))
	if err != nil {
		gs.Message = language.String("MSG_LOAD_FAILED") + err.Error()
		gs.MessageTimer = time.Now()
		return
	}
	gs.loadFromSaveData(sd)
}

func (gs *Geoscape) loadFromSaveData(sd *save.SaveData) {
	gs.Game.GameTime = sd.GameTime
	gs.Game.Funds = sd.Funds
	gs.Game.Paused = sd.Paused
	gs.Game.TimeSpeed = sd.TimeSpeed
	gs.AlienActivity = sd.AlienActivity
	gs.Game.SpeciesSeed = sd.SpeciesSeed
	if sd.AlienKnowledge != nil {
		gs.Game.AlienKnowledge = sd.AlienKnowledge
	}
	gs.Game.AlienSpecies, gs.Game.AlienTypes = data.GenerateSpecies(sd.SpeciesSeed)
	data.InitResearchTree(sd.SpeciesSeed, gs.Game.AlienSpecies)
	data.RegisterProceduralItems(sd.SpeciesSeed, gs.Game.AlienSpecies)
	gs.Bases = nil
	for _, bs := range sd.Bases {
		gs.Bases = append(gs.Bases, save.ToBase(bs))
	}
	if len(gs.Bases) == 0 {
		b := base.NewBase("Base 1", 0)
		gs.Bases = append(gs.Bases, b)
	}
	gs.ActiveBase = 0
	gs.UFOs = nil
	for _, u := range sd.UFOs {
		ufoType := GetUFOTypeByName(u.TypeName)
		if ufoType != nil {
			ufo := &UFO{
			Type:     *ufoType,
			ID:       u.ID,
			X:        u.X,
			Y:        u.Y,
			Progress: u.Progress,
			NodeFrom: u.NodeFrom,
			NodeTo:   u.NodeTo,
			TurnsLeft: u.TurnsLeft,
			Active:   u.Active,
		}
		if u.ID > ufoIDCounter {
			ufoIDCounter = u.ID
		}
		gs.UFOs = append(gs.UFOs, ufo)
		}
	}
	gs.Missions = nil
	for _, m := range sd.Missions {
		gs.Missions = append(gs.Missions, &AlienMission{
			Type:      m.Type,
			NodeID:    int(m.X),
			HoursLeft: m.HoursLeft,
		})
	}
	gs.MissionsWon = sd.MissionsWon
	// A Cydonia final mission already in progress (or effectively won) must not
	// be re-triggered when the save is loaded.
	if sd.MissionsWon >= 10 {
		gs.CydoniaTriggered = true
	}
	for _, m := range gs.Missions {
		if m.NodeID == 0 {
			gs.CydoniaTriggered = true
		}
	}
	// Restore alien bases
	gs.AlienBases = nil
	for _, abs := range sd.AlienBases {
		gs.AlienBases = append(gs.AlienBases, &AlienBase{
			CityID:          abs.CityID,
			Threat:          abs.Threat,
			TurnsAlive:      abs.TurnsAlive,
			LastMissionTick: abs.LastMissionTick,
			DefendingUFOID:  abs.DefendingUFOID,
			Name:            abs.Name,
		})
	}
	gs.Message = language.String("MSG_GAME_LOADED")
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) listSaveSlots() []engine.SlotInfo {
	var slots []engine.SlotInfo
	for slot := 1; slot <= 10; slot++ {
		sd, err := save.LoadGame(save.SavePath(slot))
		if err != nil {
			continue
		}
		label := engine.FormatSlotLabel(slot, sd.GameTime.Format("2006 Jan 02"), sd.Funds)
		slots = append(slots, engine.SlotInfo{Slot: slot, Label: label})
	}
	return slots
}

func (gs *Geoscape) openSaveSlotPicker() {
	slots := gs.listSaveSlots()
	picker := engine.NewSlotPickerScreen(gs.Game, engine.SlotPickerSave, slots, func(slot int) {
		gs.SaveGameToSlot(slot)
	})
	gs.Game.PushScreen(picker)
}

func (gs *Geoscape) openLoadSlotPicker() {
	slots := gs.listSaveSlots()
	picker := engine.NewSlotPickerScreen(gs.Game, engine.SlotPickerLoad, slots, func(slot int) {
		gs.LoadGameFromSlot(slot)
	})
	gs.Game.PushScreen(picker)
}

func (gs *Geoscape) TogglePause() {
	gs.Game.Paused = !gs.Game.Paused
	if gs.Game.Paused {
		gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
	} else {
		if gs.Game.TimeSpeed == 0 {
			if gs.LastSpeed > 0 {
				gs.Game.TimeSpeed = gs.LastSpeed
			} else {
				gs.Game.TimeSpeed = 1
			}
		}
		gs.Message = fmt.Sprintf(language.String("GEOSCAPE_TIME_RUNNING"), gs.Game.TimeSpeed)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SetSpeed(s int) {
	gs.LastSpeed = s
	gs.Game.TimeSpeed = s
	gs.Game.Paused = false
	gs.Message = fmt.Sprintf(language.String("GEOSCAPE_TIME_SPEED"), s)
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LaunchInterceptor() {
	gs.TargetSelectionMode = !gs.TargetSelectionMode
	if gs.TargetSelectionMode {
		gs.TargetCursor = 0
		gs.Message = language.String("GEO_SELECT_TARGET")
	} else {
		gs.Message = language.String("GEO_LAUNCH_CANCELLED")
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) confirmLaunch(target interface{}) {
	gs.TargetSelectionMode = false

	if gs.SelectedBase() == nil {
		return
	}

	switch t := target.(type) {
	case *UFO:
		available := gs.SelectedBase().GetAvailableInterceptors()
		if len(available) == 0 {
			gs.Message = language.String("MSG_NO_INTERCEPTORS_AVAILABLE")
			gs.MessageTimer = time.Now()
			return
		}
		baseCity := gs.CityByID(gs.SelectedBase().CityID)
		interState := available[0]
		interState.Status = "active"
		inter := NewInterceptorFromState(interState, baseCity.X, baseCity.Y)
		inter.LaunchAtUFO(t)
		gs.Interceptors = append(gs.Interceptors, inter)
		// Resume real-time simulation so the interceptor actually flies to the UFO.
		gs.resumeRealtime()
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_LAUNCHED"), t.Type.DisplayName())
	case *CrashSite:
		gs.DispatchTransport(t)
		if city := gs.CityByID(t.NodeID); city != nil {
			gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_DISPATCHED"), city.LangName())
		}
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	// Layout: left=region table, right=minimap
	engine.Layout.UpdateMode(w, h)
	tableW := engine.Layout.GeoTableWidth(w)
	mapW := engine.Layout.GeoMapWidth(w)
	mapX := engine.Layout.GeoMapX(w)

	mobile := engine.Layout.IsMobile()
	tableX, tableY, tableH := 1, 1, h-7
	miniX, miniY, miniH := mapX, 1, h-7
	// statusY is the top row of the 5-row bottom status panel.
	statusY := h - 6
	if mobile {
		// On mobile the touch control bar is pinned to the bottom. Lay out the
		// content (region table + world map + status panel) to fill all the
		// vertical space above it instead of the fixed h-7 desktop reservation.
		reserved := engine.Menu.ReservedBottom(w, h)
		contentBottom := h - reserved - 1
		if contentBottom < 16 {
			contentBottom = 16
		}
		if contentBottom > h-1 {
			contentBottom = h - 1
		}
		// Place the 5-row status panel just above the control bar.
		statusY = contentBottom - 5
		usableH := statusY - 1 // rows 1..statusY-1 for table + map

		// Stack vertically: region table on top, world map below it.
		half := usableH / 2
		tableX, tableY, tableH = 1, 1, half
		miniX, miniY, miniH = 1, half+2, usableH-half-1
		mapW = w - 2
		mapX = 1
	}

	// Clear
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			ctx.SetCell(x, y, ' ', engine.StyleDefault)
		}
	}

	gs.renderRegionTable(ctx, tableX, tableY, tableW-1, tableH)
	if mapW > 0 {
		gs.renderMinimap(ctx, miniX, miniY, mapW-1, miniH)
	}
	// Cache layout rects for mouse hit-testing in HandleMouse.
	gs.tableRect = [4]int{tableX, tableY, tableW - 1, tableH}
	gs.mapRect = [4]int{miniX, miniY, mapW - 1, miniH}

	// Bottom status
	ctx.DrawPanel(0, statusY, w, 5, language.String("GEOSCAPE"), engine.StyleDefault)
	fundsStr := fmt.Sprintf(language.String("GEOSCAPE_FUNDS"), gs.Game.Funds/1000)
	if gs.Game.Difficulty > 0 && gs.Game.Difficulty < len(engine.Difficulties) {
		fundsStr += fmt.Sprintf(language.String("GEOSCAPE_DIFF_SUFFIX"), engine.Difficulties[gs.Game.Difficulty].LangName())
	}
	timeStr := fmt.Sprintf(language.String("GEOSCAPE_TIME"), gs.Game.GameTime.Format("02/01/2006 15:04"))
	pauseStr := language.String("GEOSCAPE_RUNNING")
	if gs.Game.Paused {
		pauseStr = language.String("GEOSCAPE_PAUSED")
	}
	ctx.DrawString(2, statusY+1, fundsStr, engine.StyleGreen)
	ctx.DrawString(w/3, statusY+1, timeStr, engine.StyleDefault)
	ctx.DrawString(w*2/3, statusY+1, pauseStr, engine.StyleYellow)

	soldiersStr := ""
	if sb := gs.SelectedBase(); sb != nil {
		soldiersStr = fmt.Sprintf("[%s] ", sb.Name) + fmt.Sprintf(language.String("GEOSCAPE_SQUAD"), len(sb.Soldiers))
	}
	alienStr := fmt.Sprintf(language.String("GEOSCAPE_ACTIVITY"), gs.AlienActivity)
	missionStr := fmt.Sprintf(language.String("GEOSCAPE_MISSIONS"), len(gs.Missions), gs.MissionsWon)

	ctx.DrawString(2, statusY+2, missionStr, engine.StyleMagenta)
	ctx.DrawString(w/3, statusY+2, alienStr, engine.StyleRed)
	ctx.DrawString(w*2/3, statusY+2, soldiersStr, engine.StyleCyan)

	if time.Since(gs.MessageTimer) < 4*time.Second && gs.Message != "" {
		ctx.DrawString(2, statusY+3, gs.Message, engine.StyleDefault)
	}

	if !engine.Config.TouchMode {
		help := language.String("HELP_GEOSCAPE")
		ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)
	}

	if gs.MissionSelectMode {
		gs.renderMissionSelect(ctx, w, h)
	}
}

func (gs *Geoscape) renderMissionSelect(ctx *engine.ScreenCtx, w, h int) {
	missionFmt := language.String("GEO_MISSION_FMT")
	oddsFmt := language.String("GEO_AUTORESOLVE_ODDS")
	help := language.String("GEO_MISSION_HELP")
	title := language.String("GEO_MISSION_RESPONSE")
	opt1 := language.String("GEO_OPTION_TACTICAL")
	opt2 := language.String("GEO_OPTION_AUTORESOLVE")
	opt3 := language.String("GEO_OPTION_IGNORE")

	idx := gs.missionIndexAtCursor()
	missionType := ""
	missionNode := -1
	if idx >= 0 && idx < len(gs.Missions) {
		mission := gs.Missions[idx]
		missionType = mission.Type
		missionNode = mission.NodeID
	}

	fmt1 := fmt.Sprintf(missionFmt, missionType, gs.cityName(missionNode))
	fmt2 := fmt.Sprintf(oddsFmt, gs.MissionOdds)

	// Compute required width from all text lines
	minW := 30
	neededW := minW
	for _, s := range []string{fmt1, fmt2, opt1, opt2, opt3, help, title} {
		sw := engine.StringWidth(s) + 8
		if sw > neededW {
			neededW = sw
		}
	}
	// Cap at screen width minus 4, min 50
	overlayW := neededW + 4
	if overlayW < 50 {
		overlayW = 50
	}
	if overlayW > w-4 {
		overlayW = w - 4
	}
	overlayH := 14
	ox := (w - overlayW) / 2
	oy := (h - overlayH) / 2

	for dy := 0; dy < overlayH; dy++ {
		for dx := 0; dx < overlayW; dx++ {
			ctx.SetCell(ox+dx, oy+dy, ' ', engine.StyleDefault.Background(tcell.NewRGBColor(20, 20, 40)))
		}
	}

	ctx.DrawPanel(ox, oy, overlayW, overlayH, title, engine.StyleCyanBold)

	ctx.DrawString(ox+2, oy+2, fmt1, engine.StyleDefault)
	ctx.DrawString(ox+2, oy+3, fmt2, engine.StyleYellow)

	options := []string{opt1, opt2, opt3}
	for i, opt := range options {
		style := engine.StyleDefault
		if i == gs.MissionSelectIdx {
			style = engine.StyleGreen
		}
		ctx.DrawString(ox+2, oy+5+i, opt, style)
	}

	ctx.DrawString(ox+2, oy+9, help, engine.StyleGray)
}

func (gs *Geoscape) getTargets() []interface{} {
	var targets []interface{}
	for _, u := range gs.UFOs.Active() {
		targets = append(targets, u)
	}
	for _, cs := range gs.CrashSites {
		if !cs.Looted {
			targets = append(targets, cs)
		}
	}
	return targets
}

func (gs *Geoscape) renderRegionTable(ctx *engine.ScreenCtx, x, y, w, h int) {
	// Header
	var hdr string
	if gs.TargetSelectionMode {
		hdr = language.String("GEO_HEADER_TARGET")
	} else {
		hdr = language.String("GEO_HEADER_REGION")
	}
	if len(hdr) > w {
		hdr = hdr[:w]
	}
	ctx.DrawString(x, y, hdr, engine.StyleCyanBold)

	sep := ""
	for i := 0; i < w; i++ {
		sep += "\u2500"
	}
	ctx.DrawString(x, y+1, sep, engine.StyleGray)

	row := 0
	if gs.TargetSelectionMode {
		targets := gs.getTargets()
		for _, t := range targets {
			if row >= h-2 {
				break
			}
			ry := y + 2 + row

			// Highlight selected (reuse cursor for selection index)
			sel := row == gs.TargetCursor%len(targets)
			baseStyle := engine.StyleDefault
			if sel {
				baseStyle = engine.StyleHighlight
			}

			var name string
			switch target := t.(type) {
			case *UFO:
				name = fmt.Sprintf(language.String("GEO_UFO_LABEL"), target.Type.DisplayName())
			case *CrashSite:
				name = fmt.Sprintf(language.String("GEO_CRASH_LABEL"), localizeUFOName(target.UFOName))
			}

			prefix := "  "
			if sel {
				prefix = "> "
			}
			ctx.DrawString(x, ry, prefix+name, baseStyle)
			row++
		}
	} else {
		for _, c := range gs.Cities {
			if row >= h-2 {
				break
			}
			ry := y + 2 + row

			// Highlight selected
			sel := c.ID == gs.CursorNode
			baseStyle := engine.StyleDefault
			if sel {
				baseStyle = engine.StyleHighlight
			}

			// City name (truncated)
			name := c.LangName()
			if engine.StringWidth(name) > 14 {
				runes := []rune(name)
				for len(runes) > 0 && engine.StringWidth(string(runes)) > 14 {
					runes = runes[:len(runes)-1]
				}
				name = string(runes)
			}
			prefix := "  "
			if sel {
				prefix = "> "
			}
			ctx.DrawString(x, ry, prefix+name, baseStyle)

			// Threat bar
			tx := x + int(float64(w)*0.4)
			if c.Threat > 0 {
				barLen := c.Threat * 6 / 100
				if barLen < 1 {
					barLen = 1
				}
				threatStyle := engine.StyleYellow
				if c.Threat > 50 {
					threatStyle = engine.StyleRedBold
				}
				bar := ""
				for i := 0; i < 6; i++ {
					if i < barLen {
						bar += "\u2588"
					} else {
						bar += "\u2591"
					}
				}
				ctx.DrawString(tx, ry, bar, threatStyle)
			} else {
				ctx.DrawString(tx, ry, "\u2591\u2591\u2591\u2591\u2591\u2591", engine.StyleGray)
			}

			// Radar
			rx := x + int(float64(w)*0.6)
			if c.HasRadar {
				ctx.DrawString(rx, ry, language.String("GEO_RADAR_ON"), engine.StyleCyan)
			} else {
				ctx.DrawString(rx, ry, language.String("GEO_RADAR_OFF"), engine.StyleGray)
			}

			// Interceptor count
			ix := x + int(float64(w)*0.75)
			if c.InterceptorCount > 0 {
				ctx.DrawString(ix, ry, fmt.Sprintf(" %d ", c.InterceptorCount), engine.StyleGreen)
			} else {
				ctx.DrawString(ix, ry, language.String("GEO_RADAR_OFF"), engine.StyleGray)
			}

			// Status
			sx := x + int(float64(w)*0.85)
			if c.MissionHere {
				ctx.DrawString(sx, ry, language.String("GEO_STATUS_MISSION"), engine.StyleMagenta)
			} else if gs.HasBaseAt(c.ID) != nil {
				ctx.DrawString(sx, ry, language.String("GEO_STATUS_BASE"), engine.StyleCyanBold)
			} else if c.Threat > 50 {
				ctx.DrawString(sx, ry, language.String("GEO_STATUS_DANGER"), engine.StyleRedBold)
			} else if c.Threat > 0 {
				ctx.DrawString(sx, ry, language.String("GEO_STATUS_ALERT"), engine.StyleYellow)
			} else {
				ctx.DrawString(sx, ry, language.String("GEO_STATUS_CLEAR"), engine.StyleGray)
			}

			row++
		}
	}
}

func (gs *Geoscape) renderMinimap(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawPanel(x, y, w, h, language.String("GEO_MAP_PANEL"), engine.StyleGray)

	innerW := w - 2
	innerH := h - 2
	if innerW < 10 || innerH < 5 {
		return
	}

	// World map is 180x90
	worldW := worldMapW
	worldH := worldMapH

	// Day/night terminator calculation
	totalMin := float64(gs.Game.GameTime.Hour())*60 + float64(gs.Game.GameTime.Minute())
	dayFraction := totalMin / (24 * 60)
	sunX := int(dayFraction * float64(worldW)) // sub-solar point in world coords
	seasonOff := 10 * math.Sin(float64(gs.Game.GameTime.Month())*math.Pi/6)

	// Draw World Map Background
	for dy := 0; dy < innerH; dy++ {
		for dx := 0; dx < innerW; dx++ {
			worldX := (dx * worldW) / innerW
			worldY := (dy * worldH) / innerH

			tile := GetTile(worldX, worldY)
			var ch rune
			var style tcell.Style

			if tile == 1 {
				ch = ' '
				style = engine.StyleDefault
			} else {
				ch = '░'
				style = engine.StyleWater
			}

			// Night side: darken with a blue tint
			relX := (worldX - sunX + worldW) % worldW
			latRad := float64(worldY-worldH/2) / float64(worldH/2) * math.Pi / 2
			wobble := int(seasonOff * math.Sin(latRad))
			nightBoundary := worldW/4 + wobble
			if nightBoundary < 0 {
				nightBoundary = 0
			}
			if nightBoundary > worldW/2 {
				nightBoundary = worldW / 2
			}
			isNight := relX >= nightBoundary && relX <= worldW-nightBoundary
			if isNight {
				if engine.Config.Theme == "paper" {
					darkPaper := tcell.NewRGBColor(70, 66, 59)
					if tile == 1 {
						style = tcell.StyleDefault.Background(darkPaper).Foreground(darkPaper)
					} else {
						style = tcell.StyleDefault.Background(darkPaper).Foreground(tcell.NewRGBColor(30, 45, 60))
					}
				} else {
					if tile == 1 {
						style = engine.StyleDefault.Background(engine.DarkenColor(engine.StyleDefault.GetBackground(), 0.35)).Foreground(engine.DarkenColor(engine.StyleDefault.GetBackground(), 0.35))
					} else {
						style = engine.StyleDefault.Background(tcell.StyleDefault.GetBackground()).Foreground(tcell.NewRGBColor(0, 0, 25))
					}
				}
			}

			// Terminator line at the boundary
			termWidth := 2
			if !isNight && ((relX >= nightBoundary-termWidth && relX <= nightBoundary+termWidth) ||
				(relX >= worldW-nightBoundary-termWidth && relX <= worldW-nightBoundary+termWidth)) {
				ch = '·'
				style = style.Foreground(tcell.NewRGBColor(200, 100, 0))
			}

			ctx.SetCell(x+1+dx, y+1+dy, ch, style)
		}
	}

	// Draw regional radar coverage around each base (toggle with V)
	if gs.ShowRadarOverlay {
		radarStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 40, 70))
		for _, b := range gs.Bases {
			city := gs.CityByID(b.CityID)
			if city == nil {
				continue
			}
			radarCount := b.CountFacility(base.FacRadar)
			radarRange := baseRadarRange + radarCount*perRadarRange
			for dy := -radarRange; dy <= radarRange; dy++ {
				for dx := -radarRange; dx <= radarRange; dx++ {
					if dx*dx+dy*dy > radarRange*radarRange {
						continue
					}
					wx := city.X + dx
					wy := city.Y + dy
					if wx < 0 || wx >= worldW || wy < 0 || wy >= worldH {
						continue
					}
					sx := x + 1 + (wx * innerW / worldW)
					sy := y + 1 + (wy * innerH / worldH)
					if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
						continue
					}
					cur, _ := ctx.Peek(sx, sy)
					if cur == ' ' || cur == 0 {
						ctx.SetCell(sx, sy, '·', radarStyle)
					}
				}
			}
		}
	}

	// Draw cities
	for _, c := range gs.Cities {
		sx := x + 1 + (c.X * innerW / worldW)
		sy := y + 1 + (c.Y * innerH / worldH)

		if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
			continue
		}

		ch, style := gs.cityStyle(c)
		if c.ID == gs.CursorNode {
			ch = '◉'
			style = engine.StyleDefault.Bold(true)
		}
		ctx.SetCell(sx, sy, ch, style)
	}

	// Draw crash sites
	for _, cs := range gs.CrashSites {
		c := gs.CityByID(cs.NodeID)
		if c == nil {
			continue
		}
		sx := x + 1 + (c.X * innerW / worldW)
		sy := y + 1 + (c.Y * innerH / worldH)
		if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
			continue
		}
		if cs.Looted {
			ctx.SetCell(sx, sy, '*', engine.StyleGray)
		} else {
			ctx.SetCell(sx, sy, '*', engine.StyleYellow.Bold(true))
		}
	}

	// Draw UFOs
	for _, u := range gs.UFOs.Active() {
		sx := x + 1 + int(u.X*float64(innerW)/float64(worldW))
		sy := y + 1 + int(u.Y*float64(innerH)/float64(worldH))
		if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
			continue
		}

		ch := '!'
		style := engine.StyleRedBold

		// Pulsing radar blip (only when not in a dogfight)
		if gs.DogfightVisual == nil || !gs.DogfightVisual.Active || gs.DogfightVisual.UFO != u {
			if gs.TickCounter%24 < 12 {
				ch = '!'
				style = engine.StyleRed
			} else {
				ch = '◉'
				style = engine.StyleRedBold
			}
		}

		// Dogfight animation effects
		if gs.DogfightVisual != nil && gs.DogfightVisual.Active && gs.DogfightVisual.UFO == u {
			_, prevStyle := ctx.Peek(sx, sy)
			bg := prevStyle.GetBackground()
			if gs.DogfightVisual.UFOHitFlash > 0 {
				ch = '◉'
				style = tcell.StyleDefault.Background(bg).Foreground(color.XTerm11).Bold(true)
			} else if gs.DogfightVisual.Timer%6 < 3 {
				style = engine.StyleRed
			} else {
				style = engine.StyleRedBold
			}
			if !gs.DogfightVisual.UFOAlive {
				// Destroyed / crashing
				if gs.DogfightVisual.Timer%4 < 2 {
					ch = '✕'
					style = tcell.StyleDefault.Background(bg).Foreground(color.XTerm11).Bold(true)
				} else {
					ch = '✕'
					style = tcell.StyleDefault.Background(bg).Foreground(color.XTerm9).Bold(true)
				}
			}
		}
		ctx.SetCell(sx, sy, ch, style)
	}

	// Draw interceptor trails
	for _, in := range gs.Interceptors.Active() {
		for ti, pt := range in.Trail {
			tsx := x + 1 + int(pt.X*float64(innerW)/float64(worldW))
			tsy := y + 1 + int(pt.Y*float64(innerH)/float64(worldH))
			if tsx <= x || tsx >= x+w-1 || tsy <= y || tsy >= y+h-1 {
				continue
			}
			// Fade: older trail points are dimmer
			fade := float64(ti) / float64(len(in.Trail))
			r := int32(0*fade + 30*(1-fade))
			g := int32(80*fade + 180*(1-fade))
			b := int32(80*fade + 180*(1-fade))
			trailStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(r, g, b))
			ctx.SetCell(tsx, tsy, '·', trailStyle)
		}
	}

	// Draw interceptors
	for _, i := range gs.Interceptors.Active() {
		sx := x + 1 + int(i.X*float64(innerW)/float64(worldW))
		sy := y + 1 + int(i.Y*float64(innerH)/float64(worldH))
		if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
			continue
		}

		ch := '>'
		style := engine.StyleCyanBold

		// Engaging a UFO vs patrolling
		if i.TargetUFO != nil && i.TargetUFO.Active {
			ch = '►'
		}

		// Dogfight animation effects
		if gs.DogfightVisual != nil && gs.DogfightVisual.Active && gs.DogfightVisual.Interceptor == i {
			_, prevStyle := ctx.Peek(sx, sy)
			bg := prevStyle.GetBackground()
			if gs.DogfightVisual.HitFlash > 0 {
				// Flashing red when taking damage
				ch = '◄'
				style = tcell.StyleDefault.Background(bg).Foreground(color.XTerm9).Bold(true)
			} else if gs.DogfightVisual.Timer%6 < 3 {
				ch = '►'
				style = engine.StyleCyan
			} else {
				ch = '►'
				style = engine.StyleCyanBold
			}
			if !gs.DogfightVisual.InterAlive {
				ch = '✕'
				style = tcell.StyleDefault.Background(bg).Foreground(color.XTerm9).Bold(true)
			}
		}
		ctx.SetCell(sx, sy, ch, style)
	}

	// Draw transport
	if gs.Transport != nil {
		t := gs.Transport
		fromCity := gs.CityByID(t.FromNode)
		toCity := gs.CityByID(t.ToNode)
		if fromCity != nil && toCity != nil {
			tx := float64(fromCity.X) + float64(toCity.X-fromCity.X)*t.Progress
			ty := float64(fromCity.Y) + float64(toCity.Y-fromCity.Y)*t.Progress
			sx := x + 1 + int(tx*float64(innerW)/float64(worldW))
			sy := y + 1 + int(ty*float64(innerH)/float64(worldH))
			if sx > x && sx < x+w-1 && sy > y && sy < y+h-1 {
				ctx.SetCell(sx, sy, '≈', engine.StyleGreen)
			}
		}
	}

	// Draw dogfight combat info overlay on the minimap
	if gs.DogfightVisual != nil && gs.DogfightVisual.Active {
		dv := gs.DogfightVisual
		pX := x + w - 22
		pY := y + h - 5
		if pX > x && pY > y {
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 21; dx++ {
					ctx.SetCell(pX+dx, pY+dy, ' ', tcell.StyleDefault.Background(tcell.NewRGBColor(10, 10, 30)))
				}
			}
			interPct := float64(dv.InterHP) / float64(dv.InterMaxHP)
			if interPct < 0 {
				interPct = 0
			}
			barLen := 10
			filled := int(interPct * float64(barLen))
			barStr := ""
			for b := 0; b < barLen; b++ {
				if b < filled {
					barStr += "█"
				} else {
					barStr += "░"
				}
			}
			barColor := engine.StyleGreen
			if interPct < 0.3 {
				barColor = engine.StyleRed
			} else if interPct < 0.6 {
				barColor = engine.StyleYellow
			}
			ctx.DrawString(pX+1, pY, fmt.Sprintf(language.String("DOGFIGHT_INTER_BAR"), barStr, dv.InterHP, dv.InterMaxHP), barColor)

			ufoPct := float64(dv.UFOHP) / float64(dv.UFOMaxHP)
			if ufoPct < 0 {
				ufoPct = 0
			}
			filled = int(ufoPct * float64(barLen))
			barStr = ""
			for b := 0; b < barLen; b++ {
				if b < filled {
					barStr += "█"
				} else {
					barStr += "░"
				}
			}
			ctx.DrawString(pX+1, pY+1, fmt.Sprintf(language.String("DOGFIGHT_UFO_BAR"), barStr, dv.UFOHP, dv.UFOMaxHP), engine.StyleRed)

			dmgStr := ""
			if dv.InterDamage > 0 {
				dmgStr = fmt.Sprintf(language.String("DOGFIGHT_HIT"), dv.InterDamage)
			} else if dv.InterDamage == -1 {
				dmgStr = language.String("DOGFIGHT_UFO_DESTROYED")
			} else {
				dmgStr = language.String("DOGFIGHT_MISS")
			}
			if dv.UFODamage > 0 {
				if dmgStr != "" {
					dmgStr += fmt.Sprintf(language.String("DOGFIGHT_SEPARATOR"), dv.UFODamage)
				} else {
					dmgStr = fmt.Sprintf(language.String("DOGFIGHT_HIT"), dv.UFODamage)
				}
			}
			if dmgStr != "" {
				ctx.DrawString(pX+1, pY+2, dmgStr, engine.StyleYellow)
			}
		}
	}
}

func (gs *Geoscape) cityStyle(c *City) (rune, tcell.Style) {
	if gs.HasBaseAt(c.ID) != nil {
		return '\u25C6', engine.StyleCyanBold
	}
	if gs.hasAlienBaseAt(c.ID) {
		return '\u25B2', engine.StyleRed // ▲ = alien base
	}
	if c.Threat > 50 {
		return '\u25CF', engine.StyleRed
	}
	if c.Threat > 0 {
		return '\u25CB', engine.StyleYellow
	}
	return '\u25CB', engine.StyleGreen
}

func (gs *Geoscape) HandleKey(e *tcell.EventKey) {
	if gs.MissionSelectMode {
		switch e.Key() {
		case tcell.KeyUp:
			gs.MissionSelectIdx--
			if gs.MissionSelectIdx < 0 {
				gs.MissionSelectIdx = 2
			}
		case tcell.KeyDown:
			gs.MissionSelectIdx++
			if gs.MissionSelectIdx > 2 {
				gs.MissionSelectIdx = 0
			}
		case tcell.KeyEnter:
			switch gs.MissionSelectIdx {
			case 0:
				gs.MissionSelectMode = false
				gs.RespondToSelectedMission()
			case 1:
				gs.MissionSelectMode = false
				idx := gs.missionIndexAtCursor()
				if idx < 0 {
					idx = 0
				}
				gs.AutoresolveMission(idx)
			case 2:
				gs.MissionSelectMode = false
				gs.Message = language.String("MSG_MISSION_IGNORED")
				gs.MessageTimer = time.Now()
			}
		case tcell.KeyEscape:
			gs.MissionSelectMode = false
			gs.Message = language.String("MSG_MISSION_SELECT_CANCELLED")
			gs.MessageTimer = time.Now()
		}
		return
	}
	switch e.Key() {
	case tcell.KeyUp:
		if gs.TargetSelectionMode {
			gs.TargetCursor--
		} else {
			gs.moveCursor(0, -1)
		}
	case tcell.KeyDown:
		if gs.TargetSelectionMode {
			gs.TargetCursor++
		} else {
			gs.moveCursor(0, 1)
		}
	case tcell.KeyLeft:
		gs.moveCursor(-1, 0)
	case tcell.KeyRight:
		gs.moveCursor(1, 0)
	case tcell.KeyEnter:
		if gs.TargetSelectionMode {
			targets := gs.getTargets()
			if len(targets) > 0 {
				idx := gs.TargetCursor % len(targets)
				gs.confirmLaunch(targets[idx])
			}
		}
	case tcell.KeyF5:
		gs.openSaveSlotPicker()
	case tcell.KeyF9:
		gs.openLoadSlotPicker()
	}
	switch e.Str() {
	case "b", "B":
		if sb := gs.SelectedBase(); sb != nil {
			gs.Game.SetScreen(engine.StateBase, base.NewBaseScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateEquip, base.NewEquipScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateResearch, base.NewResearchScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateManufacture, base.NewManufactureScreen(gs.Game, sb))
			gs.Game.PushState(engine.StateBase)
		}
	case "l", "L":
		if !gs.Game.Paused {
			gs.Game.Paused = true
			gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
			gs.MessageTimer = time.Now()
		}
		gs.LaunchInterceptor()
	case "a", "A":
		gs.Autoresolve()
	case "m", "M":
		if gs.MissionSelectMode {
			switch gs.MissionSelectIdx {
			case 0:
				gs.MissionSelectMode = false
				gs.RespondToSelectedMission()
			case 1:
				gs.MissionSelectMode = false
				idx := gs.missionIndexAtCursor()
				if idx < 0 {
					idx = 0
				}
				gs.AutoresolveMission(idx)
			case 2:
				gs.MissionSelectMode = false
				gs.Message = language.String("MSG_MISSION_IGNORED")
				gs.MessageTimer = time.Now()
			}
		} else {
			gs.enterMissionSelectMode()
		}
	case " ":
		gs.TogglePause()
	case "1":
		gs.SetSpeed(1)
	case "2":
		gs.SetSpeed(2)
	case "3":
		gs.SetSpeed(3)
	case "4":
		gs.SetSpeed(4)
	case "q", "Q":
		gs.Game.Quit()
	case "r", "R":
		if !gs.Game.Paused {
			gs.Game.Paused = true
			gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
			gs.MessageTimer = time.Now()
		}
		gs.sendTransportToNearest()
	case "e", "E":
		if sb := gs.SelectedBase(); sb != nil {
			gs.Game.OpenEncyclopedia(sb.CompletedResearch, sb.UnlockedWeapons, sb.UnlockedArmor)
		}
	case "c", "C":
		gs.CycleBase()
	case "n", "N":
		gs.BuildBase()
	case "t", "T":
		if len(gs.Bases) < 2 {
			gs.Message = language.String("MSG_NEED_TWO_BASES")
			gs.MessageTimer = time.Now()
			break
		}
		gs.Game.PushScreen(gs.NewTransferScreen())
	case "v", "V":
		gs.ShowRadarOverlay = !gs.ShowRadarOverlay
		if gs.ShowRadarOverlay {
			gs.Message = language.String("MSG_RADAR_ON")
		} else {
			gs.Message = language.String("MSG_RADAR_OFF")
		}
		gs.MessageTimer = time.Now()
	}
}

func (gs *Geoscape) moveCursor(dx, dy int) {
	curIdx := -1
	for i, c := range gs.Cities {
		if c.ID == gs.CursorNode {
			curIdx = i
			break
		}
	}

	if curIdx == -1 {
		return
	}

	delta := dy
	if dx != 0 {
		delta = dx * 10 // page left/right by 10
	}
	newIdx := curIdx + delta
	if newIdx < 0 {
		newIdx = len(gs.Cities) - 1
	} else if newIdx >= len(gs.Cities) {
		newIdx = 0
	}

	gs.CursorNode = gs.Cities[newIdx].ID
}

func (gs *Geoscape) sendTransportToNearest() {
	if gs.SelectedBase() == nil {
		return
	}
	var nearest *CrashSite
	bestDist := transportSentinel
	for _, cs := range gs.CrashSites {
		if cs.Looted {
			continue
		}
		path := gs.ShortestPath(gs.SelectedBase().CityID, cs.NodeID)
		dist := len(path)
		if dist < bestDist {
			bestDist = dist
			nearest = cs
		}
	}
	if nearest == nil {
		gs.Message = language.String("MSG_NO_CRASH_SITES")
		gs.MessageTimer = time.Now()
		return
	}
	gs.DispatchTransport(nearest)
}

func (gs *Geoscape) HandleMouse(e *tcell.EventMouse) {
	if !engine.Config.MouseEnabled {
		return
	}
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := gs.Game.ScreenSize()

	if y == h-1 {
		help := language.String("HELP_GEOSCAPE")
		col := 1
		runes := []rune(help)
		for i := 0; i < len(runes); {
			if runes[i] != '[' {
				col += engine.StringWidth(string(runes[i]))
				i++
				continue
			}
			segStart := col
			end := i + 1
			for end < len(runes) && runes[end] != ']' {
				end++
			}
			if end >= len(runes) {
				break
			}
			segEnd := col + engine.StringWidth(string(runes[i:end+1]))
			if x >= segStart && x <= segEnd {
				gs.clickHelpKey(string(runes[i+1 : end]))
				return
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	// Click in table region (left/top pane)
	tX, tY, tW, tH := gs.tableRect[0], gs.tableRect[1], gs.tableRect[2], gs.tableRect[3]
	if x > tX && x < tX+tW && y > tY+1 && y < tY+tH {
		row := y - (tY + 2)
		if row >= 0 && row < len(gs.Cities) {
			gs.CursorNode = gs.Cities[row].ID
			gs.Message = fmt.Sprintf(language.String("GEOSCAPE_NODE_SELECTED"), gs.Cities[row].LangName(), gs.Cities[row].LangRegion())
			gs.MessageTimer = time.Now()
		}
	}

	// Click in minimap region (right/bottom pane)
	mX, mY, mW, mH := gs.mapRect[0], gs.mapRect[1], gs.mapRect[2], gs.mapRect[3]
	innerW := mW - 2
	innerH := mH - 2
	if innerW > 0 && innerH > 0 && x >= mX+1 && x < mX+1+innerW && y >= mY+1 && y < mY+1+innerH {
		worldW := 180
		worldH := 90
		worldX := (x - mX - 1) * worldW / innerW
		worldY := (y - mY - 1) * worldH / innerH
		// Find nearest city, UFO, or CrashSite
		var bestCity *City
		bestDist := 25 // Click radius
		for _, c := range gs.Cities {
			dx := c.X - worldX
			dy := c.Y - worldY
			dist := dx*dx + dy*dy
			if dist < bestDist {
				bestDist = dist
				bestCity = c
			}
		}

		if bestCity != nil {
			gs.CursorNode = bestCity.ID
			idx := gs.FindBaseIndex(bestCity.ID)
			if idx != -1 {
				gs.ActiveBase = idx
			}
			gs.Message = fmt.Sprintf(language.String("GEOSCAPE_NODE_SELECTED"), bestCity.LangName(), bestCity.LangRegion())
			gs.MessageTimer = time.Now()
		} else {
			// Check for UFOs
			for _, u := range gs.UFOs.Active() {
				dx := int(u.X) - worldX
				dy := int(u.Y) - worldY
				if dx*dx+dy*dy < minimapClickRadius {
					gs.Message = fmt.Sprintf(language.String("GEOSCAPE_UFO_SELECTED"), localizeUFOName(u.Type.Name))
					gs.MessageTimer = time.Now()
					break
				}
			}
		}
	}
}

func localizeUFOName(englishName string) string {
	key := "UFO_" + strings.ToUpper(strings.ReplaceAll(englishName, " ", "_"))
	return language.String(key)
}

func generateUFOLoot(ufoName string) []string {
	var loot []string
	switch ufoName {
	case "Small Scout":
		loot = append(loot, "alloys")
		if rand.Intn(100) < 30 {
			loot = append(loot, "elerium")
		}
	case "Medium Scout":
		loot = append(loot, "alloys", "alloys")
		if rand.Intn(100) < 50 {
			loot = append(loot, "elerium")
		}
		loot = append(loot, "ufo_nav")
	case "Large Scout":
		loot = append(loot, "alloys", "alloys", "alloys")
		if rand.Intn(100) < 60 {
			loot = append(loot, "elerium")
		}
		loot = append(loot, "ufo_nav", "ufo_weapon")
	case "Harvester":
		loot = append(loot, "alloys", "alloys", "alloys", "alloys")
		if rand.Intn(100) < 70 {
			loot = append(loot, "elerium", "elerium")
		}
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_weapon")
	case "Bomber":
		loot = append(loot, "alloys", "alloys", "alloys", "alloys", "alloys")
		if rand.Intn(100) < 80 {
			loot = append(loot, "elerium", "elerium")
		}
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_weapon", "ufo_armor")
	case "Transport":
		loot = append(loot, "alloys", "alloys", "alloys", "alloys")
		if rand.Intn(100) < 75 {
			loot = append(loot, "elerium")
		}
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_armor")
	default:
		loot = append(loot, "alloys", "alloys")
	}
	return loot
}

func (gs *Geoscape) clickHelpKey(key string) {
	switch {
	case key == "L" || key == "\u041f":
		if !gs.Game.Paused {
			gs.Game.Paused = true
			gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
			gs.MessageTimer = time.Now()
		}
		gs.LaunchInterceptor()
	case key == "B" || key == "\u0411":
		if sb := gs.SelectedBase(); sb != nil {
			gs.Game.SetScreen(engine.StateBase, base.NewBaseScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateEquip, base.NewEquipScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateResearch, base.NewResearchScreen(gs.Game, sb))
			gs.Game.SetScreen(engine.StateManufacture, base.NewManufactureScreen(gs.Game, sb))
			gs.Game.PushState(engine.StateBase)
		}
	case key == "R":
		if !gs.Game.Paused {
			gs.Game.Paused = true
			gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
			gs.MessageTimer = time.Now()
		}
		gs.sendTransportToNearest()
	case key == "?":
		gs.Game.PushState(engine.StateHelp)
	case isPauseKeyLabel(key):
		gs.TogglePause()
	case key == "1-4":
		nxt := gs.Game.TimeSpeed + 1
		if nxt > 4 {
			nxt = 1
		}
		gs.SetSpeed(nxt)
	default:
		if strings.ContainsAny(key, "\u2190\u2191\u2192\u2193") {
			gs.moveCursor(0, 1)
		}
	}
}

func isPauseKeyLabel(key string) bool {
	switch key {
	case "Space", "Espace", "Espacio", "Espa\u00e7o", "\u041f\u0440\u043e\u0431\u0435\u043b", "\u7a7a\u683c":
		return true
	}
	return false
}

func (gs *Geoscape) DispatchTransport(cs *CrashSite) {
	if gs.Transport != nil {
		gs.Message = language.String("MSG_TRANSPORT_BUSY")
		gs.MessageTimer = time.Now()
		return
	}
	if gs.SelectedBase() == nil {
		return
	}
	gs.Transport = &Transport{
		FromNode:       gs.SelectedBase().CityID,
		ToNode:         cs.NodeID,
		Progress:       0,
		CrashSite:      cs,
		SourceBaseCity: gs.SelectedBase().CityID,
	}
	// Resume real-time simulation so the transport actually travels to the site.
	gs.resumeRealtime()
	gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_DISPATCHED"), localizeUFOName(cs.UFOName))
	gs.MessageTimer = time.Now()
}
