package geo

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/civ13/termcom/internal/audio"
	"github.com/civ13/termcom/internal/base"
	"github.com/civ13/termcom/internal/battle"
	"github.com/civ13/termcom/internal/data"
	"github.com/civ13/termcom/internal/engine"
	"github.com/civ13/termcom/internal/language"
	"github.com/civ13/termcom/internal/save"
	"github.com/civ13/termcom/internal/soldier"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type AlienMission struct {
	Type      string
	NodeID    int     // target node ID
	HoursLeft float64 // hours remaining to respond
}

type CrashSite struct {
	UFOName string
	NodeID  int
	Looted  bool
}

type Transport struct {
	FromNode  int
	ToNode    int
	Progress  float64
	Returning bool
	CrashSite *CrashSite
}

type Geoscape struct {
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
	AlienActivity       int
	MissionsWon         int
	Victory             bool
	LastSpeed           int
	CursorNode          int
	TargetSelectionMode bool
	PreBattleStats      map[string][6]int
	ActiveCrashSite     *CrashSite
	ActiveBaseDefense   *base.Base // non-nil if the current battle is defending this base
	ActiveMissionType   string     // mission Type string of the battle in progress (for rewards)
	ActiveFinalMission  bool       // non-nil if the current battle is the Cydonia final mission
	CydoniaTriggered    bool       // ensures the final mission is added only once
	ShowRadarOverlay    bool       // toggle radar coverage circles on minimap
}

func (gs *Geoscape) SelectedBase() *base.Base {
	if gs.ActiveBase < 0 || gs.ActiveBase >= len(gs.Bases) {
		return nil
	}
	return gs.Bases[gs.ActiveBase]
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

	// Check for battle results
	if gs.Game.ActiveBattle != nil {
		r := gs.Game.ActiveBattle
		defendingBase := gs.SelectedBase()
		if gs.ActiveBaseDefense != nil {
			defendingBase = gs.ActiveBaseDefense
		}
		if defendingBase == nil {
			gs.Game.ActiveBattle = nil
			return
		}
		defendingBase.Soldiers = r.Soldiers
		dead := defendingBase.RemoveDeadSoldiers()

		if r.Won {
			if len(r.StunnedAliens) > 0 {
				capacity := defendingBase.CountFacility(base.FacContainment) * 10
				captured := 0
				for _, alien := range r.StunnedAliens {
					if len(defendingBase.LiveAliens) < capacity {
						defendingBase.LiveAliens = append(defendingBase.LiveAliens, alien)
						captured++
					} else {
						break
					}
				}
				if captured > 0 {
					gs.Message += fmt.Sprintf(language.String("MSG_ALIENS_CAPTURED"), captured)
				}
				if len(r.StunnedAliens) > captured {
					gs.Message += language.String("MSG_ALIEN_NO_SPACE")
				}
			}

			defendingBase.AddLoot(r.LootItems)
			gs.MissionsWon++
			if gs.ActiveFinalMission {
				// Winning the Cydonia assault ends the campaign in victory.
				gs.Victory = true
				gs.Message = language.String("MSG_CYDONIA_WON")
			} else if gs.ActiveCrashSite != nil {
				cs := gs.ActiveCrashSite
				cs.Looted = true
				loot := generateUFOLoot(cs.UFOName)
				defendingBase.AddLoot(loot)
				gs.Message = fmt.Sprintf(language.String("MSG_VICTORY_LOOT"), r.Kills, append(r.LootItems, loot...))
			} else if gs.ActiveBaseDefense != nil {
				gs.Message = fmt.Sprintf(language.String("MSG_BASE_DEFENDED"), defendingBase.Name, r.Kills)
			} else {
				gs.Message = fmt.Sprintf(language.String("MSG_VICTORY_LOOT"), r.Kills, r.LootItems)
			}
			// Mission-specific bonus rewards (non-crash, non-base-defense)
			if gs.ActiveMissionType != "" {
				gs.applyMissionRewards(defendingBase)
			}
		} else {
			if gs.ActiveBaseDefense != nil {
				gs.destroyBase(defendingBase)
			} else {
				gs.Message = fmt.Sprintf(language.String("MSG_DEFEAT_LOST"), dead)
			}
		}
		gs.MessageTimer = time.Now()
		gs.ActiveCrashSite = nil
		gs.ActiveBaseDefense = nil
		gs.ActiveMissionType = ""
		gs.ActiveFinalMission = false

		if gs.PreBattleStats != nil {
			statNames := []string{"HP", "ACC", "REA", "STR", "BRA", "TU"}
			for _, s := range defendingBase.Soldiers {
				old, ok := gs.PreBattleStats[s.Name]
				if !ok {
					continue
				}
				newStats := [6]int{s.HP, s.Accuracy, s.Reactions, s.Strength, s.Bravery, s.TU}
				gains := []string{}
				for i := 0; i < 6; i++ {
					if newStats[i] > old[i] {
						gains = append(gains, fmt.Sprintf("%s+%d", statNames[i], newStats[i]-old[i]))
					}
				}
				if len(gains) > 0 {
					gs.Message = fmt.Sprintf("%s improved: %s", s.Name, strings.Join(gains, " "))
					gs.MessageTimer = time.Now()
				}
			}
			gs.PreBattleStats = nil
		}

		gs.Game.ActiveBattle = nil
	}

	// Defeat check — alien activity too high
	if gs.AlienActivity >= 100 && !gs.Victory {
		stats := fmt.Sprintf("Missions Won: %d", gs.MissionsWon)
		gs.Game.GameOver(false, stats)
		gs.Victory = true
		gs.Game.Paused = true
	}

	// Victory check — enough missions completed
	if gs.MissionsWon >= 10 && !gs.Victory {
		// Instead of immediate victory, trigger Cydonia
		gs.triggerCydonia()
	}

	// Final mission check
	if gs.Victory && gs.Game.ActiveBattle == nil {
		stats := fmt.Sprintf("Campaign Complete. Missions Won: %d", gs.MissionsWon)
		gs.Game.GameOver(true, stats)
	}

	if !gs.Game.Paused && gs.Game.TimeSpeed > 0 {
		speedMult := []int{0, 1, 5, 20, 60}
		minutes := speedMult[gs.Game.TimeSpeed]

		// Spawn UFOs periodically, scaled by game time
		gameMonth := int(gs.Game.GameTime.Month()) - 3 + (gs.Game.GameTime.Year()-1999)*12
		if gameMonth < 0 {
			gameMonth = 0
		}
		ufoSpawnRate := 600 - gameMonth*20
		if ufoSpawnRate < 200 {
			ufoSpawnRate = 200
		}
		diffUFOScale := 1.0
		if gs.Game.Difficulty >= 0 && gs.Game.Difficulty < len(engine.Difficulties) {
			diffUFOScale = engine.Difficulties[gs.Game.Difficulty].UFOScale
		}
		ufoSpawnRate = int(float64(ufoSpawnRate) / diffUFOScale)
		if ufoSpawnRate < 100 {
			ufoSpawnRate = 100
		}
		if gs.TickCounter%ufoSpawnRate == 0 {
			maxUFOs := 5 + gameMonth/2
			if maxUFOs > 12 {
				maxUFOs = 12
			}
			if gs.UFOs.Count() < maxUFOs {
				ufo := SpawnUFOOnCities(gs.Cities, gameMonth)
				gs.UFOs = append(gs.UFOs, ufo)
				city := gs.CityByID(ufo.CurrentNode())
				cityName := "?"
				if city != nil {
					cityName = city.Name
				}
				gs.Message = fmt.Sprintf(language.String("MSG_UFO_DETECTED"), ufo.Type.Name, cityName)
				gs.MessageTimer = time.Now()
				audio.PlayAlert()
			}
		}

		// Spawn alien missions periodically
		// Increase frequency based on AlienActivity and game time:
		spawnRate := 1800 - (gs.AlienActivity * 15) - gameMonth*30
		if spawnRate < 300 {
			spawnRate = 300
		}
		if gs.TickCounter%spawnRate == 0 {
			gs.spawnMission()
		}

		// Gradually increase activity over time
		if gs.TickCounter%7200 == 0 { // ~2 hours at speed 1
			gs.AlienActivity++
		}

		// Check mission timers
		remaining := make([]*AlienMission, 0, len(gs.Missions))
		for _, m := range gs.Missions {
			m.HoursLeft -= float64(minutes) / 60.0
			if m.HoursLeft <= 0 {
				city := gs.CityByID(m.NodeID)
				cityName := "?"
				if city != nil {
					cityName = city.Name
				}
				// Base defense mission that expired: the aliens overrun the base
				if defBase := gs.HasBaseAt(m.NodeID); defBase != nil {
					gs.Message = fmt.Sprintf(language.String("MSG_BASE_DESTROYED"), defBase.Name)
					gs.MessageTimer = time.Now()
					gs.destroyBase(defBase)
				} else {
					gs.Message = fmt.Sprintf(language.String("MSG_ATTACK_CITY"), m.Type, cityName)
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

		for _, i := range gs.Interceptors {
			if i.Launching {
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
										gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_RETRIEVED"), cs.UFOName, 0)
										gs.MessageTimer = time.Now()
										gs.PreBattleStats = make(map[string][6]int)
										for _, s := range healthy {
											gs.PreBattleStats[s.Name] = [6]int{s.HP, s.Accuracy, s.Reactions, s.Strength, s.Bravery, s.TU}
										}
										bs := battle.NewBattlescape(gs.Game, selectedBase, healthy, cs.UFOName)
										gs.Game.SetScreen(engine.StateBattlescape, bs)
										gs.Game.PushState(engine.StateBattlescape)
										return
									}
									gs.Message = language.String("MSG_NO_SOLDIERS")
									gs.MessageTimer = time.Now()
								}
								t.Returning = true
								if sb := gs.SelectedBase(); sb != nil {
									t.ToNode = sb.CityID
								}
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

		// Advance research and manufacturing
		if gs.TickCounter%30 == 0 {
			if sb := gs.SelectedBase(); sb != nil {
				var msgs []string
				done := sb.AdvanceResearch()
				for _, name := range done {
					audio.PlayResearchComplete()
					msgs = append(msgs, fmt.Sprintf(language.String("MSG_RESEARCH_COMPLETE"), name))
				}
				crafted := sb.AdvanceManufacture()
				for _, item := range crafted {
					audio.PlayManufactureComplete()
					msgs = append(msgs, fmt.Sprintf(language.String("MSG_MANUFACTURE_COMPLETE"), item))
				}
				if len(msgs) > 0 {
					gs.Message = msgs[0]
					gs.MessageTimer = time.Now()
				}
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
		gs.SaveGameAuto()
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
		gs.Game.GameOver(false, "Last base destroyed!")
		gs.Victory = true
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
		gs.Message = fmt.Sprintf(language.String("MSG_BASE_EXISTS"), city.Name)
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
	b := base.NewBase(fmt.Sprintf("Base %d", baseNum), gs.CursorNode)
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 2})
	gs.Bases = append(gs.Bases, b)
	city.HasRadar = true
	gs.ActiveBase = len(gs.Bases) - 1
	gs.Message = fmt.Sprintf(language.String("MSG_BASE_BUILT"), b.Name, city.Name)
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
	maxEdgeDist := 50.0

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

func (gs *Geoscape) dogfight(inter *Interceptor) {
	ufo := inter.TargetUFO
	if ufo == nil {
		// Interceptor reached patrol node, look for UFOs here
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

	// Check if interceptor is in range
	dist := math.Sqrt(math.Pow(ufo.X-inter.X, 2) + math.Pow(ufo.Y-inter.Y, 2))
	if dist > float64(inter.Range) {
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_CLOSING"), inter.Weapon.Name, inter.Mode.String())
		gs.MessageTimer = time.Now()
		return
	}

	damage := inter.FireAt(ufo)
	audio.PlayShoot()
	if damage == 0 {
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_MISS"), inter.Weapon.Name)
		gs.MessageTimer = time.Now()
	} else if damage == -1 {
		gs.Game.Funds += int64(ufo.Type.Points * 1000)

		// Check if over water
		city := gs.CityByID(ufo.CurrentNode())
		if city != nil && GetTile(city.X, city.Y) == 0 { // 0 is water
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_LOST_AT_SEA"), ufo.Type.Name)
		} else {
			gs.CrashSites = append(gs.CrashSites, &CrashSite{
				UFOName: ufo.Type.Name,
				NodeID:  ufo.CurrentNode(),
			})
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_CRASHED"), ufo.Type.Name)
		}

		gs.MessageTimer = time.Now()
		inter.Disengage()
	} else {

		gs.Message = fmt.Sprintf(language.String("MSG_HIT_UFO"), damage)
		gs.MessageTimer = time.Now()
	}

	// UFO fires back
	if ufo.Active && inter.HP > 0 {
		ufoDmg := ufo.FireAtInterceptor(inter)
		audio.PlayPlasmaFire()
		if ufoDmg > 0 {
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_HIT_INTERCEPTOR"), ufoDmg, inter.HP, inter.MaxHP)
			gs.MessageTimer = time.Now()
			if inter.State != nil {
				inter.State.HP = inter.HP
			}
		}
		if inter.HP <= 0 {
			gs.Message = language.String("MSG_INTERCEPTOR_DESTROYED")
			gs.MessageTimer = time.Now()
			if inter.State != nil {
				inter.State.HP = 0
				inter.State.Status = "Destroyed"
			}
			inter.Disengage()
		}
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

	// Build candidate list. If the player has multiple bases, aliens may
	// directly assault a base (base defense scenario).
	var candidates []*City
	for _, c := range gs.Cities {
		if c.ID == gs.SelectedBase().CityID {
			// Only allow the home (selected) base as a target occasionally,
			// so base defense missions can occur.
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
	target := candidates[rand.Intn(len(candidates))]

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
	gs.Message = fmt.Sprintf(language.String("MSG_ALERT_MISSION"), chosen, target.Name)
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayAlert()
}

func (gs *Geoscape) triggerCydonia() {
	if gs.CydoniaTriggered {
		return
	}
	gs.CydoniaTriggered = true
	gs.Message = "Cydonia location detected! Final mission ready."
	gs.MessageTimer = time.Now()

	// Add Cydonia as a special mission
	mission := &AlienMission{
		Type:      "Alien Base Assault", // Reuse for Cydonia
		NodeID:    0,                    // Special node for Cydonia
		HoursLeft: 9999.0,               // Indefinite
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
		if s.HP > 0 && s.Wounds == 0 {
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
	city := gs.CityByID(mission.NodeID)
	cityName := "?"
	if city != nil {
		cityName = city.Name
		city.MissionHere = false
	}

	// Base defense mission if the target city hosts one of our bases
	if defBase := gs.HasBaseAt(mission.NodeID); defBase != nil {
		gs.ActiveBaseDefense = defBase
	}
	gs.Message = fmt.Sprintf(language.String("MSG_SQUAD_DEPLOYED"), mission.Type, cityName)
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
	case language.String("MISSION_ABDUCTION"):
		ufoName = language.String("MISSION_TYPE_ABDUCTION")
	case language.String("MISSION_RESEARCH"):
		ufoName = language.String("MISSION_TYPE_RESEARCH")
	case language.String("MISSION_COUNCIL"):
		ufoName = language.String("MISSION_TYPE_COUNCIL")
	}
	if mission.NodeID == 0 {
		ufoName = "Cydonia"
		gs.ActiveFinalMission = true
	}
	if gs.ActiveBaseDefense != nil {
		ufoName = language.String("MISSION_TYPE_BASE")
	}
	gs.ActiveMissionType = mission.Type
	bs := battle.NewBattlescape(gs.Game, defBase, healthy, ufoName)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
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
	bestDist := 9999.0
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
		if s.HP > 0 && s.Wounds == 0 {
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
		gs.Message = fmt.Sprintf(language.String("MSG_AUTO_VICTORY"), nearest.Type.Name, nearest.Type.Points)
	} else {
		if squadSize > 0 {
			// build list of alive soldiers
			var alive []*soldier.Soldier
			for _, s := range gs.SelectedBase().Soldiers {
				if s.HP > 0 {
					alive = append(alive, s)
				}
			}
			idx := rand.Intn(len(alive))
			alive[idx].HP = 0
			gs.SelectedBase().RemoveDeadSoldiers()
			gs.Message = fmt.Sprintf(language.String("MSG_AUTO_DEFEAT"), nearest.Type.Name)
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
			TypeName: u.Type.Name,
			X:        u.X,
			Y:        u.Y,
			Progress: u.Progress,
			NodeFrom: u.NodeFrom,
			NodeTo:   u.NodeTo,
			Active:   u.Active,
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
			gs.UFOs = append(gs.UFOs, &UFO{
				Type:     *ufoType,
				NodeFrom: int(u.X),
				NodeTo:   int(u.Y),
				Progress: 0.5,
				Active:   u.Active,
			})
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
		gs.Message = "Select target (UFO or Crash Site)."
	} else {
		gs.Message = "Launch cancelled."
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
		interState.Status = "Active"
		inter := NewInterceptorFromState(interState, baseCity.X, baseCity.Y)
		inter.LaunchAtUFO(t)
		gs.Interceptors = append(gs.Interceptors, inter)
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_LAUNCHED"), t.Type.Name)
	case *CrashSite:
		gs.DispatchTransport(t)
		gs.Message = fmt.Sprintf("Transport dispatched to crash site at %s", gs.CityByID(t.NodeID).Name)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	// Layout: left=region table, right=minimap
	tableW := w * 60 / 100
	if tableW < 30 {
		tableW = 30
	}
	mapW := w - tableW - 2
	mapX := tableW + 2

	// Clear
	for y := 1; y < h-5; y++ {
		for x := 1; x < w-1; x++ {
			ctx.SetCell(x, y, ' ', engine.StyleDefault)
		}
	}

	gs.renderRegionTable(ctx, 1, 1, tableW-1, h-7)
	gs.renderMinimap(ctx, mapX, 1, mapW-1, h-7)

	// Bottom status
	ctx.DrawPanel(0, h-6, w, 5, language.String("GEOSCAPE"), engine.StyleDefault)
	fundsStr := fmt.Sprintf(language.String("GEOSCAPE_FUNDS"), gs.Game.Funds/1000)
	if gs.Game.Difficulty > 0 && gs.Game.Difficulty < len(engine.Difficulties) {
		fundsStr += fmt.Sprintf("  [%s]", engine.Difficulties[gs.Game.Difficulty].Name)
	}
	timeStr := fmt.Sprintf(language.String("GEOSCAPE_TIME"), gs.Game.GameTime.Format("02/01/2006 15:04"))
	pauseStr := language.String("GEOSCAPE_RUNNING")
	if gs.Game.Paused {
		pauseStr = language.String("GEOSCAPE_PAUSED")
	}
	ctx.DrawString(2, h-5, fundsStr, engine.StyleGreen)
	ctx.DrawString(w/3, h-5, timeStr, engine.StyleDefault)
	ctx.DrawString(w*2/3, h-5, pauseStr, engine.StyleYellow)

	soldiersStr := ""
	if sb := gs.SelectedBase(); sb != nil {
		soldiersStr = fmt.Sprintf("[%s] ", sb.Name) + fmt.Sprintf(language.String("GEOSCAPE_SQUAD"), len(sb.Soldiers))
	}
	alienStr := fmt.Sprintf(language.String("GEOSCAPE_ACTIVITY"), gs.AlienActivity)
	missionStr := fmt.Sprintf(language.String("GEOSCAPE_MISSIONS"), len(gs.Missions), gs.MissionsWon)

	ctx.DrawString(2, h-4, missionStr, engine.StyleMagenta)
	ctx.DrawString(w/3, h-4, alienStr, engine.StyleRed)
	ctx.DrawString(w*2/3, h-4, soldiersStr, engine.StyleCyan)

	if time.Since(gs.MessageTimer) < 4*time.Second && gs.Message != "" {
		ctx.DrawString(2, h-3, gs.Message, engine.StyleDefault)
	}

	help := language.String("HELP_GEOSCAPE")
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)
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
		hdr = " SELECT TARGET"
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
			sel := row == gs.CursorNode%len(targets)
			baseStyle := engine.StyleDefault
			if sel {
				baseStyle = engine.StyleHighlight
			}

			var name string
			switch target := t.(type) {
			case *UFO:
				name = "UFO: " + target.Type.Name
			case *CrashSite:
				name = "Crash: " + target.UFOName
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
			name := c.Name
			if len(name) > 14 {
				name = name[:14]
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
	worldW := 180
	worldH := 90

	// Draw World Map Background
	for dy := 0; dy < innerH; dy++ {
		for dx := 0; dx < innerW; dx++ {
			worldX := (dx * worldW) / innerW
			worldY := (dy * worldH) / innerH

			tile := GetTile(worldX, worldY)
			var ch rune
			var style tcell.Style

			if tile == 1 {
				ch = '░'
				style = engine.StyleGray
			} else {
				ch = ' '
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 20, 60))
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
			radarRange := 24 + radarCount*10
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
		worldX := int(c.X)
		worldY := int(c.Y)
		if worldX >= 0 && worldX < worldW && worldY >= 0 && worldY < worldH && GetTile(worldX, worldY) == 1 {
			style = style.Background(color.XTerm0)
		}
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
		ctx.SetCell(sx, sy, '!', engine.StyleRedBold)
	}

	// Draw interceptors
	for _, i := range gs.Interceptors.Active() {
		sx := x + 1 + int(i.X*float64(innerW)/float64(worldW))
		sy := y + 1 + int(i.Y*float64(innerH)/float64(worldH))
		if sx <= x || sx >= x+w-1 || sy <= y || sy >= y+h-1 {
			continue
		}
		ctx.SetCell(sx, sy, '>', engine.StyleCyanBold)
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
}

func (gs *Geoscape) cityStyle(c *City) (rune, tcell.Style) {
	if gs.HasBaseAt(c.ID) != nil {
		return '\u25C6', tcell.StyleDefault.Background(color.XTerm8).Foreground(color.XTerm6).Bold(true)
	}
	if c.Threat > 50 {
		return '\u25CF', tcell.StyleDefault.Background(color.XTerm8).Foreground(color.XTerm9)
	}
	if c.Threat > 0 {
		return '\u25CB', tcell.StyleDefault.Background(color.XTerm8).Foreground(color.XTerm11)
	}
	return '\u25CB', tcell.StyleDefault.Background(color.XTerm8).Foreground(color.XTerm2)
}

func (gs *Geoscape) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		gs.moveCursor(0, -1)
	case tcell.KeyDown:
		gs.moveCursor(0, 1)
	case tcell.KeyLeft:
		gs.moveCursor(-1, 0)
	case tcell.KeyRight:
		gs.moveCursor(1, 0)
	case tcell.KeyEnter:
		if gs.TargetSelectionMode {
			targets := gs.getTargets()
			if len(targets) > 0 {
				idx := gs.CursorNode % len(targets)
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
		gs.RespondToSelectedMission()
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
			gs.Message = "RADAR OVERLAY: ON"
		} else {
			gs.Message = "RADAR OVERLAY: OFF"
		}
		gs.MessageTimer = time.Now()
	}
}

func (gs *Geoscape) moveCursor(dx, dy int) {
	// Move based on list index instead of spatial position
	// dx is ignored as we move linearly through the Cities list

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

	newIdx := curIdx + dy
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
	bestDist := 999999
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
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := gs.Game.ScreenSize()

	if y == h-1 {
		help := language.String("HELP_GEOSCAPE")
		helpActions := []string{"=Select", "=Launch", "=Autoresolve", "=Mission", "=Base", "=Transport", "=Pause", "=Quit"}
		helpFuncs := []func(){
			func() { gs.moveCursor(0, 1) },
			func() { gs.LaunchInterceptor() },
			func() { gs.Autoresolve() },
			func() { gs.RespondToSelectedMission() },
			func() {
				if sb := gs.SelectedBase(); sb != nil {
					gs.Game.SetScreen(engine.StateBase, base.NewBaseScreen(gs.Game, sb))
					gs.Game.SetScreen(engine.StateEquip, base.NewEquipScreen(gs.Game, sb))
					gs.Game.SetScreen(engine.StateResearch, base.NewResearchScreen(gs.Game, sb))
					gs.Game.SetScreen(engine.StateManufacture, base.NewManufactureScreen(gs.Game, sb))
					gs.Game.PushState(engine.StateBase)
				}
			},
			func() { gs.sendTransportToNearest() },
			func() { gs.TogglePause() },
			func() { gs.Game.Quit() },
		}
		off := 1
		for i, action := range helpActions {
			pos := strings.Index(help, action)
			if pos < 0 {
				continue
			}
			start := off + pos
			end := off + pos + len(action)
			if x >= start && x <= end {
				helpFuncs[i]()
				return
			}
		}
		return
	}

	// Click in table region (left pane)
	tableW := w * 60 / 100
	if x > 1 && x < tableW && y > 2 && y < h-7 {
		row := y - 3
		if row >= 0 && row < len(gs.Cities) {
			gs.CursorNode = gs.Cities[row].ID
			gs.Message = fmt.Sprintf(language.String("GEOSCAPE_NODE_SELECTED"), gs.Cities[row].Name, gs.Cities[row].Region)
			gs.MessageTimer = time.Now()
		}
	}

	// Click in minimap region (right pane)
	mapX := tableW + 2
	mWidth := w - tableW - 2
	innerW := mWidth - 3
	innerH := h - 9
	if innerW > 0 && innerH > 0 && x >= mapX+1 && x < mapX+1+innerW && y >= 2 && y < 2+innerH {
		worldW := 180
		worldH := 90
		worldX := (x - mapX - 1) * worldW / innerW
		worldY := (y - 2) * worldH / innerH
		// Find nearest city to clicked position
		var bestCity *City
		bestDist := 999999
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
			gs.Message = fmt.Sprintf(language.String("GEOSCAPE_NODE_SELECTED"), bestCity.Name, bestCity.Region)
			gs.MessageTimer = time.Now()
		}
	}
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
		FromNode:  gs.SelectedBase().CityID,
		ToNode:    cs.NodeID,
		Progress:  0,
		CrashSite: cs,
	}
	gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_DISPATCHED"), cs.UFOName)
	gs.MessageTimer = time.Now()
}
