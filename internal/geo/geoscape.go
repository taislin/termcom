package geo

import (
	"fmt"
	"math/rand"
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
	Network       *GeoNetwork
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

	gn := NewRegionalNetwork()
	gn.Nodes[0].HasRadar = true
	gn.Nodes[0].InterceptorCount = 2

	gs := &Geoscape{
		Game:         g,
		Network:      gn,
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

	gn := NewRegionalNetwork()
	gn.Nodes[0].HasRadar = true
	gn.Nodes[0].InterceptorCount = 2

	gs := &Geoscape{
		Game:          g,
		Network:       gn,
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
			ufo := SpawnUFOOnNetwork(gs.Network)
			gs.UFOs = append(gs.UFOs, ufo)
			node := gs.Network.NodeByID(ufo.CurrentNode())
		 nodeName := "?"
			if node != nil {
				nodeName = node.Name
			}
			gs.Message = fmt.Sprintf(language.String("MSG_UFO_DETECTED"), ufo.Type.Name, nodeName)
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
				node := gs.Network.NodeByID(m.NodeID)
				nodeName := "?"
				if node != nil {
					nodeName = node.Name
				}
				gs.Message = fmt.Sprintf(language.String("MSG_ATTACK_CITY"), m.Type, nodeName)
				gs.MessageTimer = time.Now()
				gs.AlienActivity += 10
			} else {
				remaining = append(remaining, m)
			}
		}
		gs.Missions = remaining

		// Update UFOs along network edges
		for _, u := range gs.UFOs {
			u.Update(gs.Network)
		}

		for _, i := range gs.Interceptors {
			if i.Launching {
				reached := i.Update(gs.Network)
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
					nextNode := gs.Network.NodeByID(path[1])
					if nextNode != nil {
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

func (gs *Geoscape) ShortestPath(from, to int) []int {
	return gs.Network.ShortestPath(from, to)
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
	// Pick a random non-base node
	var candidates []*GeoNode
	for _, n := range gs.Network.Nodes {
		if n.ID != gs.BaseNode {
			candidates = append(candidates, n)
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
	node := gs.Network.NodeByID(mission.NodeID)
	nodeName := "?"
	if node != nil {
		nodeName = node.Name
		node.MissionHere = false
	}
	gs.Message = fmt.Sprintf(language.String("MSG_SQUAD_DEPLOYED"), mission.Type, nodeName)
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
		node := gs.Network.NodeByID(u.CurrentNode())
		if node == nil {
			continue
		}
		baseNode := gs.Network.NodeByID(gs.BaseNode)
		if baseNode == nil {
			continue
		}
		dx := float64(node.X - baseNode.X)
		dy := float64(node.Y - baseNode.Y)
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
	// Target the node with the cursor, or nearest UFO if available
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs.Active() {
		node := gs.Network.NodeByID(u.CurrentNode())
		if node == nil {
			continue
		}
		baseNode := gs.Network.NodeByID(gs.BaseNode)
		if baseNode == nil {
			continue
		}
		dx := float64(node.X - baseNode.X)
		dy := float64(node.Y - baseNode.Y)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			nearest = u
		}
	}

	baseNode := gs.Network.NodeByID(gs.BaseNode)
	if baseNode == nil {
		return
	}
	inter := NewInterceptor(baseNode.X, baseNode.Y)

	if nearest != nil {
		// Launch at specific UFO
		inter.LaunchAtUFO(nearest)
		gs.Interceptors = append(gs.Interceptors, inter)
		gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_LAUNCHED"), nearest.Type.Name)
	} else {
		// Launch at cursor node for patrol
		targetNode := gs.Network.NodeByID(gs.CursorNode)
		if targetNode != nil && targetNode.ID != gs.BaseNode {
			inter.LaunchAtNode(gs.CursorNode, gs.Network)
			gs.Interceptors = append(gs.Interceptors, inter)
			gs.Message = fmt.Sprintf(language.String("MSG_INTERCEPTOR_PATROL"), targetNode.Name)
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

	// Clear screen
	for y := 1; y < h-5; y++ {
		for x := 1; x < w-1; x++ {
			ctx.SetCell(x, y, ' ', engine.StyleDefault)
		}
	}

	// Draw the network graph
	gs.Network.Render(ctx, w, h)

	// Draw UFOs on edges
	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		ux := int(u.X) + 1
		uy := int(u.Y) + 1
		if ux > 0 && ux < w-1 && uy > 0 && uy < h-6 {
			ctx.SetCell(ux, uy, '?', engine.StyleRedBold)
		}
	}

	// Draw interceptors
	for _, i := range gs.Interceptors {
		if i.HP <= 0 {
			continue
		}
		ix := int(i.X) + 1
		iy := int(i.Y) + 1
		if ix > 0 && ix < w-1 && iy > 0 && iy < h-6 {
			ctx.SetCell(ix, iy, '\u25B8', engine.StyleCyanBold) // ▸
		}
	}

	// Draw crash sites
	for _, cs := range gs.CrashSites {
		if cs.Looted {
			continue
		}
		node := gs.Network.NodeByID(cs.NodeID)
		if node == nil {
			continue
		}
		csx := node.X + 1
		csy := node.Y + 1
		if csx > 0 && csx < w-1 && csy > 0 && csy < h-6 {
			ctx.SetCell(csx, csy, '\u2297', engine.StyleYellow.Bold(true)) // ⊗
		}
	}

	// Draw transport
	if gs.Transport != nil {
		fromNode := gs.Network.NodeByID(gs.Transport.FromNode)
		toNode := gs.Network.NodeByID(gs.Transport.ToNode)
		if fromNode != nil && toNode != nil {
			tx := float64(fromNode.X) + float64(toNode.X-fromNode.X)*gs.Transport.Progress
			ty := float64(fromNode.Y) + float64(toNode.Y-fromNode.Y)*gs.Transport.Progress
			sx := int(tx) + 1
			sy := int(ty) + 1
			if sx > 0 && sx < w-1 && sy > 0 && sy < h-6 {
				ctx.SetCell(sx, sy, '\u2666', engine.StyleGreenBold) // ♦
			}
		}
	}

	// Draw mission markers
	for _, m := range gs.Missions {
		node := gs.Network.NodeByID(m.NodeID)
		if node == nil {
			continue
		}
		mx := node.X + 1
		my := node.Y - 1
		if mx > 0 && mx < w-1 && my > 0 && my < h-6 {
			ctx.SetCell(mx, my, '\u2605', engine.StyleMagenta) // ★
		}
	}

	// Legend
	lx := w - 22
	ly := 2
	ctx.DrawPanel(lx, ly, 21, 10, language.String("LEGEND"), engine.StyleDefault)
	for y := ly + 1; y < ly+9; y++ {
		for x := lx + 1; x < lx+20; x++ {
			ctx.SetCell(x, y, ' ', engine.StyleDefault)
		}
	}
	ctx.SetCell(lx+1, ly+1, '\u25CB', engine.StyleGreen)
	ctx.DrawString(lx+3, ly+1, language.String("LEGEND_NODE_SAFE"), engine.StyleGreen)
	ctx.SetCell(lx+1, ly+2, '\u25CB', engine.StyleYellow)
	ctx.DrawString(lx+3, ly+2, language.String("LEGEND_NODE_THREAT"), engine.StyleYellow)
	ctx.SetCell(lx+1, ly+3, '\u25CF', engine.StyleRed)
	ctx.DrawString(lx+3, ly+3, language.String("LEGEND_NODE_DANGER"), engine.StyleRed)
	ctx.SetCell(lx+1, ly+4, '\u25C6', engine.StyleCyan)
	ctx.DrawString(lx+3, ly+4, language.String("LEGEND_BASE"), engine.StyleCyan)
	ctx.SetCell(lx+1, ly+5, '?', engine.StyleRedBold)
	ctx.DrawString(lx+3, ly+5, language.String("LEGEND_UFO"), engine.StyleRedBold)
	ctx.SetCell(lx+1, ly+6, '\u25B8', engine.StyleCyanBold)
	ctx.DrawString(lx+3, ly+6, language.String("LEGEND_INTERCEPTOR"), engine.StyleCyanBold)
	ctx.SetCell(lx+1, ly+7, '\u2666', engine.StyleGreenBold)
	ctx.DrawString(lx+3, ly+7, language.String("LEGEND_TRANSPORT"), engine.StyleGreenBold)

	// Bottom status bar
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
	help := language.String("GEOSCAPE_HELP")
	if gs.Victory {
		help = "VICTORY ACHIEVED!  Q=Quit"
	}
	ctx.DrawString(1, h-1, help, engine.StyleGray)
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
	// Move to nearest node in that direction
	curNode := gs.Network.NodeByID(gs.CursorNode)
	if curNode == nil {
		return
	}
	var best *GeoNode
	bestScore := -999999
	for _, n := range gs.Network.Nodes {
		if n.ID == gs.CursorNode {
			continue
		}
		ndx := n.X - curNode.X
		ndy := n.Y - curNode.Y
		// Score: prefer nodes in the requested direction
		score := ndx*dx + ndy*dy
		if dx == 0 && dy == 0 {
			continue
		}
		if score > bestScore {
			bestScore = score
			best = n
		}
	}
	if best != nil && bestScore > 0 {
		gs.CursorNode = best.ID
	}
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

	// Handle help bar clicks (bottom bar)
	if y == h-1 {
		switch {
		case x >= 1 && x <= 3:
			gs.Game.PushState(engine.StateBase)
		case x >= 5 && x <= 12:
			gs.LaunchInterceptor()
		case x >= 14 && x <= 25:
			gs.Autoresolve()
		case x >= 27 && x <= 36:
			gs.RespondToMission(0)
		case x >= 38 && x <= 47:
			gs.sendTransportToNearest()
		case x >= 49 && x <= 57:
			gs.TogglePause()
		case x >= 59 && x <= 64:
			gs.SetSpeed(1)
		case x >= 66 && x <= 70:
			gs.Game.Quit()
		}
		return
	}

	// Click on a node
	if y > 0 && y < h-5 && x > 0 && x < w-1 {
		node := gs.Network.NearestNode(x-1, y-1)
		if node != nil {
			gs.CursorNode = node.ID
			gs.Message = fmt.Sprintf(language.String("GEOSCAPE_NODE_SELECTED"), node.Name, node.Region)
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
