package geo

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/civ13/ycom/internal/audio"
	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/battle"
	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/civ13/ycom/internal/save"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/gdamore/tcell/v3"
)

type AlienMission struct {
	Type      string
	NodeID    int     // target node ID
	HoursLeft float64 // hours remaining to respond
}

type CrashSite struct {
	UFOName  string
	NodeID   int
	Looted   bool
}

type Transport struct {
	FromNode int
	ToNode   int
	Progress float64
	Returning bool
}

type Geoscape struct {
	Game          *engine.Game
	Cities        []*City // Renamed from Network
	UFOs          UFOList
	Interceptors  InterceptorList
	CrashSites    []*CrashSite
	Transport     *Transport
	BaseNode      int     // home base node ID
	Message       string
	MessageTimer  time.Time
	TickCounter   int
	Base          *base.Base
	LastMonth     int
	Missions      []*AlienMission
	AlienActivity int
	MissionsWon   int
	Victory       bool
	// Cursor for node selection
	CursorNode    int
}

func NewGeoscape(g *engine.Game) *Geoscape {
	b := base.NewBase("Base 1")
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
		BaseNode:     0,
		CursorNode:   0,
		Message:      language.String("MSG_WELCOME"),
		MessageTimer: time.Now(),
		Base:         b,
		LastMonth:    int(g.GameTime.Month()),
	}
	return gs
}

func NewGeoscapeFromSave(g *engine.Game, sd *save.SaveData) *Geoscape {
	b := save.ToBase(sd.Base)

	cities := GetCities()
	cities[0].HasRadar = true
	cities[0].InterceptorCount = 2

	gs := &Geoscape{
		Game:          g,
		Cities:        cities,
		BaseNode:      0,
		CursorNode:    0,
		Message:       language.String("MSG_GAME_LOADED"),
		MessageTimer:  time.Now(),
		Base:          b,
		LastMonth:     int(sd.GameTime.Month()),
		AlienActivity: sd.AlienActivity,
	}

	g.GameTime = sd.GameTime
	g.Funds = sd.Funds
	g.Paused = sd.Paused
	g.TimeSpeed = sd.TimeSpeed

	for _, u := range sd.UFOs {
		ufoType := GetUFOTypeByName(u.TypeName)
		if ufoType != nil {
			gs.UFOs = append(gs.UFOs, &UFO{
				Type:   *ufoType,
				X:      u.X,
				Y:      u.Y,
				Active: u.Active,
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
		gs.Base.Soldiers = r.Soldiers
		dead := gs.Base.RemoveDeadSoldiers()

		if r.Won {
			gs.Base.AddLoot(r.LootItems)
			gs.MissionsWon++
			gs.Message = fmt.Sprintf(language.String("MSG_VICTORY_LOOT"), r.Kills, r.LootItems)
		} else {
			gs.Message = fmt.Sprintf(language.String("MSG_DEFEAT_LOST"), dead)
		}
		gs.MessageTimer = time.Now()
		gs.Game.ActiveBattle = nil
	}

	// Defeat check — alien activity too high
	if gs.AlienActivity >= 100 && !gs.Victory {
		gs.Message = language.String("MSG_GAME_OVER_ACTIVITY")
		gs.MessageTimer = time.Now()
		gs.Victory = true
		gs.Game.Paused = true
	}

	// Victory check — enough missions completed
	if gs.MissionsWon >= 10 && !gs.Victory {
		gs.Message = language.String("MSG_GAME_WON")
		gs.MessageTimer = time.Now()
		gs.Victory = true
		gs.Game.Paused = true
	}

	if !gs.Game.Paused && gs.Game.TimeSpeed > 0 {
		speedMult := []int{0, 1, 5, 20, 60}
		minutes := speedMult[gs.Game.TimeSpeed]

		// Spawn UFOs periodically
		if gs.TickCounter%600 == 0 && gs.UFOs.Count() < 5 {
			ufo := SpawnUFOOnCities(gs.Cities)
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

		// Spawn alien missions periodically
		if gs.TickCounter%1800 == 0 {
			gs.spawnMission()
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
				gs.Message = fmt.Sprintf(language.String("MSG_ATTACK_CITY"), m.Type, cityName)
				gs.MessageTimer = time.Now()
				gs.AlienActivity += 10
			} else {
				remaining = append(remaining, m)
			}
		}
		gs.Missions = remaining

		// Update UFOs along network edges
		for _, u := range gs.UFOs {
			u.Update(gs.Cities)
		}

		for _, i := range gs.Interceptors {
			if i.Launching {
				reached := i.Update(gs.Cities)
				if reached {
					gs.dogfight(i)
				}
			}
		}

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
							// Check if we arrived
							if t.FromNode == t.ToNode {
								for _, cs := range gs.CrashSites {
									if !cs.Looted && cs.NodeID == t.ToNode {
										cs.Looted = true
										loot := generateUFOLoot(cs.UFOName)
										gs.Base.AddLoot(loot)
										gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_RETRIEVED"), cs.UFOName, len(loot))
										gs.MessageTimer = time.Now()
										break
									}
								}
								t.Returning = true
								t.ToNode = gs.BaseNode
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

		// Advance research and manufacturing
		if gs.TickCounter%30 == 0 {
		var msgs []string
		done := gs.Base.AdvanceResearch()
		for _, name := range done {
			msgs = append(msgs, fmt.Sprintf(language.String("MSG_RESEARCH_COMPLETE"), name))
		}
		crafted := gs.Base.AdvanceManufacture()
		for _, item := range crafted {
			msgs = append(msgs, fmt.Sprintf(language.String("MSG_MANUFACTURE_COMPLETE"), item))
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
		salary, funding := gs.Base.AdvanceMonth()
		gs.Game.Funds += int64(funding - salary)
		gs.Base.AdvanceDay()
		gs.Base.AdvanceDay()
		gs.Base.AdvanceDay()
		gs.Message = fmt.Sprintf(language.String("MSG_MONTHLY_REPORT"), funding/1000, salary/1000)
		gs.MessageTimer = time.Now()
	}
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

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, n := range gs.Cities {
			if visited[n.ID] {
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
	damage := inter.FireAt(ufo)
	if damage == -1 {
		gs.Game.Funds += int64(ufo.Type.Points * 1000)
		gs.CrashSites = append(gs.CrashSites, &CrashSite{
			UFOName: ufo.Type.Name,
			NodeID:  ufo.CurrentNode(),
		})
		gs.Message = fmt.Sprintf(language.String("MSG_UFO_CRASHED"), ufo.Type.Name)
		gs.MessageTimer = time.Now()
		inter.Disengage()
		gs.startBattle(ufo)
	} else {
		gs.Message = fmt.Sprintf(language.String("MSG_HIT_UFO"), damage)
		gs.MessageTimer = time.Now()
	}

	// UFO fires back
	if ufo.Active && inter.HP > 0 {
		ufoDmg := ufo.FireAtInterceptor(inter)
		if ufoDmg > 0 {
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_HIT_INTERCEPTOR"), ufoDmg, inter.HP, inter.MaxHP)
			gs.MessageTimer = time.Now()
		}
		if inter.HP <= 0 {
			gs.Message = language.String("MSG_INTERCEPTOR_DESTROYED")
			gs.MessageTimer = time.Now()
			inter.Disengage()
		}
	}
}

func (gs *Geoscape) startBattle(ufo *UFO) {
	aliveCount := 0
	for _, s := range gs.Base.Soldiers {
		if s.HP > 0 {
			aliveCount++
		}
	}
	if aliveCount == 0 {
		gs.Message = language.String("MSG_NO_SOLDIERS")
		gs.MessageTimer = time.Now()
		return
	}
	gs.Game.Paused = true
	gs.Message = fmt.Sprintf(language.String("MSG_SHOT_DOWN"), ufo.Type.Name)
	gs.MessageTimer = time.Now()

	bs := battle.NewBattlescape(gs.Game, gs.Base.Soldiers, ufo.Type.Name)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
}

func (gs *Geoscape) spawnMission() {
	types := []string{language.String("MISSION_TERROR"), language.String("MISSION_SUPPLY"), language.String("MISSION_ALIEN_BASE")}
	// Pick a random non-base city
	var candidates []*City
	for _, c := range gs.Cities {
		if c.ID != gs.BaseNode {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return
	}
	target := candidates[rand.Intn(len(candidates))]

	idx := rand.Intn(len(types))
	turnsLeft := 24.0 // 24 game hours to respond
	if types[idx] == language.String("MISSION_ALIEN_BASE") {
		turnsLeft = 12.0 // 12 game hours for base assaults
	}
	mission := &AlienMission{
		Type:      types[idx],
		NodeID:    target.ID,
		HoursLeft: turnsLeft,
	}
	gs.Missions = append(gs.Missions, mission)
	target.MissionHere = true
	gs.Message = fmt.Sprintf(language.String("MSG_ALERT_MISSION"), types[idx], target.Name)
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayAlert()
}

func (gs *Geoscape) RespondToMission(idx int) {
	if idx < 0 || idx >= len(gs.Missions) {
		gs.Message = language.String("MSG_INVALID_MISSION")
		gs.MessageTimer = time.Now()
		return
	}
	aliveCount := 0
	for _, s := range gs.Base.Soldiers {
		if s.HP > 0 {
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
	gs.Message = fmt.Sprintf(language.String("MSG_SQUAD_DEPLOYED"), mission.Type, cityName)
	gs.MessageTimer = time.Now()
	gs.Game.Paused = true

	ufoName := language.String("MISSION_CRASH_SITE")
	switch mission.Type {
	case language.String("MISSION_TERROR"):
		ufoName = language.String("MISSION_TYPE_TERROR")
	case language.String("MISSION_SUPPLY"):
		ufoName = language.String("MISSION_TYPE_SUPPLY")
	case language.String("MISSION_ALIEN_BASE"):
		ufoName = language.String("MISSION_TYPE_BASE")
	}
	bs := battle.NewBattlescape(gs.Game, gs.Base.Soldiers, ufoName)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
}

func (gs *Geoscape) Autoresolve() {
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs.Active() {
		city := gs.CityByID(u.CurrentNode())
		if city == nil {
			continue
		}
		baseCity := gs.CityByID(gs.BaseNode)
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
	for _, s := range gs.Base.Soldiers {
		if s.HP > 0 {
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
			for _, s := range gs.Base.Soldiers {
				if s.HP > 0 {
					alive = append(alive, s)
				}
			}
			idx := rand.Intn(len(alive))
			alive[idx].HP = 0
			gs.Base.RemoveDeadSoldiers()
			gs.Message = fmt.Sprintf(language.String("MSG_AUTO_DEFEAT"), nearest.Type.Name)
		} else {
			gs.Message = language.String("MSG_AUTO_NO_SOLDIERS")
		}
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SaveGameToFile() {
	ufoSaves := make([]*save.UFOSave, 0)
	for _, u := range gs.UFOs {
		ufoSaves = append(ufoSaves, &save.UFOSave{
			TypeName: u.Type.Name,
			X:        float64(u.NodeFrom),
			Y:        float64(u.NodeTo),
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
	sd := &save.SaveData{
		GameTime:       gs.Game.GameTime,
		Funds:          gs.Game.Funds,
		Paused:         gs.Game.Paused,
		TimeSpeed:      gs.Game.TimeSpeed,
		AlienActivity:  gs.AlienActivity,
		SpeciesSeed:    gs.Game.SpeciesSeed,
		AlienKnowledge: gs.Game.AlienKnowledge,
		Base:           save.FromBase(gs.Base),
		UFOs:           ufoSaves,
		Missions:       missionSaves,
	}
	err := save.SaveGame("xcom_save.json", sd)
	if err != nil {
		gs.Message = language.String("MSG_SAVE_FAILED") + err.Error()
	} else {
		gs.Message = language.String("MSG_GAME_SAVED")
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LoadGameFromFile() {
	sd, err := save.LoadGame("xcom_save.json")
	if err != nil {
		gs.Message = language.String("MSG_LOAD_FAILED") + err.Error()
		gs.MessageTimer = time.Now()
		return
	}
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
	gs.Base = save.ToBase(sd.Base)
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
	gs.Message = language.String("MSG_GAME_LOADED")
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) TogglePause() {
	gs.Game.Paused = !gs.Game.Paused
	if gs.Game.Paused {
		gs.Message = language.String("GEOSCAPE_TIME_PAUSED")
	} else {
		gs.Message = fmt.Sprintf(language.String("GEOSCAPE_TIME_RUNNING"), gs.Game.TimeSpeed)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SetSpeed(s int) {
	gs.Game.TimeSpeed = s
	gs.Game.Paused = false
	gs.Message = fmt.Sprintf(language.String("GEOSCAPE_TIME_SPEED"), s)
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LaunchInterceptor() {
	// Target the city with the cursor, or nearest UFO if available
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs.Active() {
		city := gs.CityByID(u.CurrentNode())
		if city == nil {
			continue
		}
		baseCity := gs.CityByID(gs.BaseNode)
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

	baseCity := gs.CityByID(gs.BaseNode)
	if baseCity == nil {
		return
	}
	inter := NewInterceptor(baseCity.X, baseCity.Y)

	if nearest != nil {
		// Launch at specific UFO
		inter.LaunchAtUFO(nearest)
		gs.Interceptors = append(gs.Interceptors, inter)
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_LAUNCHED"), nearest.Type.Name)
	} else {
		// Launch at cursor city for patrol
		targetCity := gs.CityByID(gs.CursorNode)
		if targetCity != nil && targetCity.ID != gs.BaseNode {
			inter.LaunchAtNode(gs.CursorNode, gs.Cities)
			gs.Interceptors = append(gs.Interceptors, inter)
			gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_PATROL"), targetCity.Name)
		} else {
			gs.Message = language.String("GEOSCAPE_NO_UFO")
			gs.MessageTimer = time.Now()
			return
		}
	}
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayClick()
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
	timeStr := fmt.Sprintf(language.String("GEOSCAPE_TIME"), gs.Game.GameTime.Format("02/01/2006 15:04"))
	pauseStr := language.String("GEOSCAPE_RUNNING")
	if gs.Game.Paused {
		pauseStr = language.String("GEOSCAPE_PAUSED")
	}
	ctx.DrawString(2, h-5, fundsStr, engine.StyleGreen)
	ctx.DrawString(w/3, h-5, timeStr, engine.StyleDefault)
	ctx.DrawString(w*2/3, h-5, pauseStr, engine.StyleYellow)

	soldiersStr := fmt.Sprintf(language.String("GEOSCAPE_SQUAD"), len(gs.Base.Soldiers))
	alienStr := fmt.Sprintf(language.String("GEOSCAPE_ACTIVITY"), gs.AlienActivity)
	missionStr := fmt.Sprintf(language.String("GEOSCAPE_MISSIONS"), len(gs.Missions), gs.MissionsWon)

	ctx.DrawString(2, h-4, missionStr, engine.StyleMagenta)
	ctx.DrawString(w/3, h-4, alienStr, engine.StyleRed)
	ctx.DrawString(w*2/3, h-4, soldiersStr, engine.StyleCyan)

	if time.Since(gs.MessageTimer) < 4*time.Second && gs.Message != "" {
		ctx.DrawString(2, h-3, gs.Message, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	// Example hotkey highlighting
	help := "[j]/[k]=Select [L]=Launch [A]=Autoresolve [M]=Mission [B]=Base [R]=Transport [Space]=Pause [Q]=Quit"
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)
}

func (gs *Geoscape) renderRegionTable(ctx *engine.ScreenCtx, x, y, w, h int) {
	// Header
	hdr := " REGION          THREAT  RADAR  SQD  STATUS"
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
			ctx.DrawString(rx, ry, " R ", engine.StyleCyan)
		} else {
			ctx.DrawString(rx, ry, " - ", engine.StyleGray)
		}

		// Interceptor count
		ix := x + int(float64(w)*0.75)
		if c.InterceptorCount > 0 {
			ctx.DrawString(ix, ry, fmt.Sprintf(" %d ", c.InterceptorCount), engine.StyleGreen)
		} else {
			ctx.DrawString(ix, ry, " - ", engine.StyleGray)
		}

		// Status
		sx := x + int(float64(w)*0.85)
		if c.MissionHere {
			ctx.DrawString(sx, ry, "MISSION", engine.StyleMagenta)
		} else if c.ID == 0 {
			ctx.DrawString(sx, ry, "BASE", engine.StyleCyanBold)
		} else if c.Threat > 50 {
			ctx.DrawString(sx, ry, "DANGER", engine.StyleRedBold)
		} else if c.Threat > 0 {
			ctx.DrawString(sx, ry, "ALERT", engine.StyleYellow)
		} else {
			ctx.DrawString(sx, ry, "clear", engine.StyleGray)
		}

		row++
	}

	// Legend at bottom of table
	ly := y + h - 2
	if ly > y+3 {
		ctx.DrawPanel(x, ly, w, 2, "", engine.StyleGray)
		ctx.DrawString(x+1, ly+1, "j/k=Select L=Launch A=Auto M=Mission B=Base", engine.StyleGray)
	}
}

func (gs *Geoscape) renderMinimap(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawPanel(x, y, w, h, "MAP", engine.StyleGray)

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
	if c.ID == 0 {
		return '\u25C6', engine.StyleCyanBold.Bold(true)
	}
	if c.Threat > 50 {
		return '\u25CF', engine.StyleRedBold // ●
	}
	if c.Threat > 0 {
		return '\u25CB', engine.StyleYellow // ○
	}
	return '\u25CB', engine.StyleGreen // ○
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
	case tcell.KeyF5:
		gs.SaveGameToFile()
	case tcell.KeyF9:
		gs.LoadGameFromFile()
	}
	switch e.Str() {
	case "b", "B":
		gs.Game.PushState(engine.StateBase)
	case "l", "L":
		gs.LaunchInterceptor()
	case "a", "A":
		gs.Autoresolve()
	case "m", "M":
		gs.RespondToMission(0)
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
		gs.sendTransportToNearest()
	case "e", "E":
		gs.Game.OpenEncyclopedia(gs.Base.CompletedResearch, gs.Base.UnlockedWeapons, gs.Base.UnlockedArmor)
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
	var nearest *CrashSite
	bestDist := 999999
	for _, cs := range gs.CrashSites {
		if cs.Looted {
			continue
		}
		path := gs.ShortestPath(gs.BaseNode, cs.NodeID)
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
		help := "[j]/[k]=Select [L]=Launch [A]=Autoresolve [M]=Mission [B]=Base [R]=Transport [Space]=Pause [Q]=Quit"
		helpActions := []string{"=Select", "=Launch", "=Autoresolve", "=Mission", "=Base", "=Transport", "=Pause", "=Quit"}
		helpFuncs := []func(){
			func() { gs.moveCursor(0, 1) },
			func() { gs.LaunchInterceptor() },
			func() { gs.Autoresolve() },
			func() { gs.RespondToMission(0) },
			func() { gs.Game.PushState(engine.StateBase) },
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
	loot = append(loot, "alloys", "alloys")
	if rand.Intn(100) < 60 {
		loot = append(loot, "elerium")
	}
	switch ufoName {
	case "Small Scout":
		loot = append(loot, "ufo_nav")
	case "Medium Scout":
		loot = append(loot, "ufo_nav", "ufo_weapon")
	case "Large Scout":
		loot = append(loot, "ufo_nav", "ufo_weapon", "ufo_armor")
	case "Harvester":
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_weapon")
	case "Bomber":
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_weapon", "ufo_armor")
	case "Transport":
		loot = append(loot, "ufo_nav", "ufo_power", "ufo_armor")
	}
	return loot
}

func (gs *Geoscape) DispatchTransport(cs *CrashSite) {
	if gs.Transport != nil {
		gs.Message = language.String("MSG_TRANSPORT_BUSY")
		gs.MessageTimer = time.Now()
		return
	}
	gs.Transport = &Transport{
		FromNode: gs.BaseNode,
		ToNode:   cs.NodeID,
		Progress: 0,
	}
	gs.Message = fmt.Sprintf(language.String("MSG_TRANSPORT_DISPATCHED"), cs.UFOName)
	gs.MessageTimer = time.Now()
}
