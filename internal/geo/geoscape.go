package geo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/civ13/ycom/internal/audio"
	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/battle"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/save"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/gdamore/tcell/v2"
)

type AlienMission struct {
	Type      string
	CityName  string
	TurnsLeft int
	X, Y      int
}

type Geoscape struct {
	Game          *engine.Game
	UFOs          UFOList
	Interceptors  InterceptorList
	BaseX, BaseY  int
	ScrollX       int
	ScrollY       int
	BaseName      string
	Message       string
	MessageTimer  time.Time
	TickCounter   int
	Base          *base.Base
	LastMonth     int
	Missions      []*AlienMission
	AlienActivity int
	MissionsWon   int
	Victory       bool
}

func NewGeoscape(g *engine.Game) *Geoscape {
	b := base.NewBase("Base 1")
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})

	gs := &Geoscape{
		Game:         g,
		BaseX:        28,
		BaseY:        32,
		ScrollX:      28,
		ScrollY:      32,
		BaseName:     "Base 1",
		Message:      "Welcome, Commander. Your mission: defend Earth from alien invasion.",
		MessageTimer: time.Now(),
		Base:         b,
		LastMonth:    int(g.GameTime.Month()),
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
			gs.Message = fmt.Sprintf("Victory! %d aliens killed. Loot: %v", r.Kills, r.LootItems)
		} else {
			gs.Message = fmt.Sprintf("Defeat! Lost: %v", dead)
		}
		gs.MessageTimer = time.Now()
		gs.Game.ActiveBattle = nil
	}

	// Defeat check — alien activity too high
	if gs.AlienActivity >= 100 && !gs.Victory {
		gs.Message = "Earth has fallen! Alien activity is overwhelming!"
		gs.MessageTimer = time.Now()
		gs.Victory = true
		gs.Game.Paused = true
	}

	// Victory check — enough missions completed
	if gs.MissionsWon >= 10 && !gs.Victory {
		gs.Message = "Alien threat neutralized! You have saved Earth!"
		gs.MessageTimer = time.Now()
		gs.Victory = true
		gs.Game.Paused = true
	}

	if !gs.Game.Paused && gs.Game.TimeSpeed > 0 {
		// Spawn UFOs periodically
		if gs.TickCounter%600 == 0 && gs.UFOs.Count() < 5 {
			ufo := SpawnUFO()
			gs.UFOs = append(gs.UFOs, ufo)
			gs.Message = fmt.Sprintf("UFO detected! %s at [%d,%d]", ufo.Type.Name, ufo.TileX(), ufo.TileY())
			gs.MessageTimer = time.Now()
		}

		// Spawn alien missions periodically
		if gs.TickCounter%1800 == 0 {
			gs.spawnMission()
		}

		// Check mission timers — iterate over a copy of the slice
		// to avoid mutating the slice during iteration
		remaining := make([]*AlienMission, 0, len(gs.Missions))
		for _, m := range gs.Missions {
			m.TurnsLeft--
			if m.TurnsLeft <= 0 {
				gs.Message = fmt.Sprintf("%s attack on %s! Panic spreads!", m.Type, m.CityName)
				gs.MessageTimer = time.Now()
				gs.AlienActivity += 10
			} else {
				remaining = append(remaining, m)
			}
		}
		gs.Missions = remaining

		for _, u := range gs.UFOs {
			u.Update()
		}

		for _, i := range gs.Interceptors {
			if i.Launching {
				reached := i.Update()
				if reached {
					gs.dogfight(i)
				}
			}
		}

		speedMult := []int{0, 1, 5, 20, 60}
		minutes := speedMult[gs.Game.TimeSpeed]
		gs.Game.GameTime = gs.Game.GameTime.Add(time.Duration(minutes) * time.Minute)

		// Advance research and manufacturing
		if gs.TickCounter%30 == 0 {
			var msgs []string
			done := gs.Base.AdvanceResearch()
			for _, name := range done {
				msgs = append(msgs, fmt.Sprintf("Research complete: %s", name))
			}
			crafted := gs.Base.AdvanceManufacture()
			for _, item := range crafted {
				msgs = append(msgs, fmt.Sprintf("Manufacturing complete: %s", item))
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
		gs.Message = fmt.Sprintf("MONTHLY REPORT: Funding +$%dK  Salaries -$%dK", funding/1000, salary/1000)
		gs.MessageTimer = time.Now()
	}
}

func (gs *Geoscape) dogfight(inter *Interceptor) {
	if inter.Target == nil {
		return
	}
	ufo := inter.Target
	damage := inter.FireAt(ufo)
	if damage == -1 {
		gs.Game.Funds += int64(ufo.Type.Points * 1000)
		inter.Disengage()
		gs.startBattle(ufo)
	} else {
		gs.Message = fmt.Sprintf("Hit UFO for %d damage!", damage)
		gs.MessageTimer = time.Now()
	}

	// UFO fires back
	if ufo.Active && inter.HP > 0 {
		ufoDmg := ufo.FireAtInterceptor(inter)
		if ufoDmg > 0 {
			gs.Message = fmt.Sprintf("UFO hit interceptor for %d damage! (HP:%d/%d)", ufoDmg, inter.HP, inter.MaxHP)
			gs.MessageTimer = time.Now()
		}
		if inter.HP <= 0 {
			gs.Message = "Interceptor destroyed!"
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
		gs.Message = "No soldiers available to deploy!"
		gs.MessageTimer = time.Now()
		return
	}
	gs.Game.Paused = true
	gs.Message = fmt.Sprintf("%s shot down! Sending squad...", ufo.Type.Name)
	gs.MessageTimer = time.Now()

	bs := battle.NewBattlescape(gs.Game, gs.Base.Soldiers, ufo.Type.Name)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
}

func (gs *Geoscape) spawnMission() {
	types := []string{"Terror", "Supply", "Alien Base Assault"}
	cityNames := []string{"London", "Tokyo", "New York", "Moscow", "Sydney", "Paris", "Berlin"}
	cityX := []int{85, 145, 50, 95, 150, 82, 88}
	cityY := []int{28, 32, 32, 26, 55, 28, 27}

	idx := rand.Intn(len(types))
	cityIdx := rand.Intn(len(cityNames))
	turnsLeft := 5
	if types[idx] == "Alien Base Assault" {
		turnsLeft = 3
	}
	mission := &AlienMission{
		Type:      types[idx],
		CityName:  cityNames[cityIdx],
		TurnsLeft: turnsLeft,
		X:         cityX[cityIdx],
		Y:         cityY[cityIdx],
	}
	gs.Missions = append(gs.Missions, mission)
	gs.Message = fmt.Sprintf("ALERT: %s mission near %s detected!", types[idx], cityNames[cityIdx])
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayAlert()
}

func (gs *Geoscape) RespondToMission(idx int) {
	if idx < 0 || idx >= len(gs.Missions) {
		gs.Message = "Invalid mission index."
		gs.MessageTimer = time.Now()
		return
	}
	// Check for alive soldiers
	aliveCount := 0
	for _, s := range gs.Base.Soldiers {
		if s.HP > 0 {
			aliveCount++
		}
	}
	if aliveCount == 0 {
		gs.Message = "No healthy soldiers available to deploy!"
		gs.MessageTimer = time.Now()
		return
	}
	mission := gs.Missions[idx]
	gs.Missions = append(gs.Missions[:idx], gs.Missions[idx+1:]...)
	gs.Message = fmt.Sprintf("Squad deployed to %s mission at %s!", mission.Type, mission.CityName)
	gs.MessageTimer = time.Now()
	gs.Game.Paused = true

	ufoName := "Crash Site"
	switch mission.Type {
	case "Terror":
		ufoName = "Terror"
	case "Supply":
		ufoName = "Supply"
	case "Alien Base Assault":
		ufoName = "Alien Base"
	}
	bs := battle.NewBattlescape(gs.Game, gs.Base.Soldiers, ufoName)
	gs.Game.SetScreen(engine.StateBattlescape, bs)
	gs.Game.PushState(engine.StateBattlescape)
}

func (gs *Geoscape) Autoresolve() {
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		dx := u.X - float64(gs.BaseX)
		dy := u.Y - float64(gs.BaseY)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			nearest = u
		}
	}
	if nearest == nil {
		gs.Message = "No UFOs to autoresolve."
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
		gs.Message = fmt.Sprintf("Autoresolve: Victory! %s destroyed. +$%dK", nearest.Type.Name, nearest.Type.Points)
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
			gs.Message = fmt.Sprintf("Autoresolve: Defeat! Lost a soldier. %s escaped.", nearest.Type.Name)
		} else {
			gs.Message = "Autoresolve: No soldiers available!"
		}
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SaveGameToFile() {
	ufoSaves := make([]*save.UFOSave, 0)
	for _, u := range gs.UFOs {
		ufoSaves = append(ufoSaves, &save.UFOSave{
			TypeName: u.Type.Name,
			X:        u.X,
			Y:        u.Y,
			Active:   u.Active,
		})
	}
	missionSaves := make([]*save.MissionSave, 0)
	for _, m := range gs.Missions {
		missionSaves = append(missionSaves, &save.MissionSave{
			Type:      m.Type,
			CityName:  m.CityName,
			TurnsLeft: m.TurnsLeft,
			X:         m.X,
			Y:         m.Y,
		})
	}
	sd := &save.SaveData{
		GameTime:      gs.Game.GameTime,
		Funds:         gs.Game.Funds,
		Paused:        gs.Game.Paused,
		TimeSpeed:     gs.Game.TimeSpeed,
		AlienActivity: gs.AlienActivity,
		Base:          save.FromBase(gs.Base),
		UFOs:          ufoSaves,
		Missions:      missionSaves,
	}
	err := save.SaveGame("xcom_save.json", sd)
	if err != nil {
		gs.Message = "Save failed: " + err.Error()
	} else {
		gs.Message = "Game saved to xcom_save.json"
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LoadGameFromFile() {
	sd, err := save.LoadGame("xcom_save.json")
	if err != nil {
		gs.Message = "Load failed: " + err.Error()
		gs.MessageTimer = time.Now()
		return
	}
	gs.Game.GameTime = sd.GameTime
	gs.Game.Funds = sd.Funds
	gs.Game.Paused = sd.Paused
	gs.Game.TimeSpeed = sd.TimeSpeed
	gs.AlienActivity = sd.AlienActivity
	gs.Base = save.ToBase(sd.Base)
	gs.UFOs = nil
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
	gs.Missions = nil
	for _, m := range sd.Missions {
		gs.Missions = append(gs.Missions, &AlienMission{
			Type:      m.Type,
			CityName:  m.CityName,
			TurnsLeft: m.TurnsLeft,
			X:         m.X,
			Y:         m.Y,
		})
	}
	gs.BaseName = gs.Base.Name
	gs.Message = "Game loaded successfully!"
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) TogglePause() {
	gs.Game.Paused = !gs.Game.Paused
	if gs.Game.Paused {
		gs.Message = "TIME PAUSED"
	} else {
		gs.Message = fmt.Sprintf("TIME RUNNING (Speed %dx)", gs.Game.TimeSpeed)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SetSpeed(s int) {
	gs.Game.TimeSpeed = s
	gs.Game.Paused = false
	gs.Message = fmt.Sprintf("Time speed: %dx", s)
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LaunchInterceptor() {
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		dx := u.X - float64(gs.BaseX)
		dy := u.Y - float64(gs.BaseY)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			nearest = u
		}
	}
	if nearest == nil {
		gs.Message = "No UFOs detected on radar."
		gs.MessageTimer = time.Now()
		return
	}

	inter := NewInterceptor(gs.BaseX, gs.BaseY)
	inter.Launch(nearest)
	gs.Interceptors = append(gs.Interceptors, inter)
	gs.Message = fmt.Sprintf("Interceptor launched! Pursuing %s.", nearest.Type.Name)
	gs.MessageTimer = time.Now()
	gs.Game.Bell()
	audio.PlayClick()
}

func (gs *Geoscape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	mw, mh := MapSize()

	viewW := w - 2
	viewH := h - 6

	halfVW := viewW / 2
	halfVH := viewH / 2

	offsetX := gs.ScrollX - halfVW
	offsetY := gs.ScrollY - halfVH

	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}
	if offsetX+viewW > mw {
		offsetX = mw - viewW
	}
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY+viewH > mh {
		offsetY = mh - viewH
	}
	if offsetY < 0 {
		offsetY = 0
	}

	for y := 0; y < viewH && y < mh; y++ {
		for x := 0; x < viewW && x < mw; x++ {
			mx := x + offsetX
			my := y + offsetY
			if mx >= mw || my >= mh {
				continue
			}
			tile := GetTile(mx, my)
			ch := '·'
			style := engine.StyleBlue
			switch tile {
			case 1:
				ch = '.'
				style = engine.StyleGreen
			case 2:
				ch = '○'
				style = engine.StyleYellow
			case 3:
				ch = '▲'
				style = engine.StyleCyan
			case 4:
				ch = '?'
				style = engine.StyleRed
			case 5:
				ch = '▸'
				style = engine.StyleCyan
			}
			ctx.SetCell(x+1, y+1, ch, style)
		}
	}

	for _, c := range cities {
		sx := c.X - offsetX + 1
		sy := c.Y - offsetY + 1
		if sx > 0 && sx < w-1 && sy > 0 && sy < h-6 {
			ctx.SetCell(sx, sy, '●', engine.StyleYellow)
			name := c.Name
			if len(name) > 8 {
				name = name[:8]
			}
			ctx.DrawString(sx+1, sy, name, engine.StyleGray)
		}
	}

	bsx := gs.BaseX - offsetX + 1
	bsy := gs.BaseY - offsetY + 1
	if bsx > 0 && bsx < w-1 && bsy > 0 && bsy < h-6 {
		ctx.SetCell(bsx, bsy, '▲', engine.StyleCyanBold)
	}

	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		ux := int(u.X) - offsetX + 1
		uy := int(u.Y) - offsetY + 1
		if ux > 0 && ux < w-1 && uy > 0 && uy < h-6 {
			ctx.SetCell(ux, uy, '?', engine.StyleRedBold)
		}
	}

	for _, i := range gs.Interceptors {
		if i.HP <= 0 {
			continue
		}
		ix := int(i.X) - offsetX + 1
		iy := int(i.Y) - offsetY + 1
		if ix > 0 && ix < w-1 && iy > 0 && iy < h-6 {
			ctx.SetCell(ix, iy, '▸', engine.StyleCyanBold)
		}
	}

	// Legend
	lx := w - 22
	ly := 2
	ctx.DrawPanel(lx, ly, 21, 10, "LEGEND", engine.StyleDefault)
	ctx.SetCell(lx+1, ly+1, '·', engine.StyleBlue)
	ctx.DrawString(lx+3, ly+1, "Water", engine.StyleBlue)
	ctx.SetCell(lx+1, ly+2, '.', engine.StyleGreen)
	ctx.DrawString(lx+3, ly+2, "Land", engine.StyleGreen)
	ctx.SetCell(lx+1, ly+3, '○', engine.StyleYellow)
	ctx.DrawString(lx+3, ly+3, "City", engine.StyleYellow)
	ctx.SetCell(lx+1, ly+4, '▲', engine.StyleCyan)
	ctx.DrawString(lx+3, ly+4, "Base", engine.StyleCyan)
	ctx.SetCell(lx+1, ly+5, '?', engine.StyleRed)
	ctx.DrawString(lx+3, ly+5, "UFO", engine.StyleRed)
	ctx.SetCell(lx+1, ly+6, '▸', engine.StyleCyanBold)
	ctx.DrawString(lx+3, ly+6, "Interceptor", engine.StyleCyanBold)
	ctx.SetCell(lx+1, ly+7, '★', engine.StyleMagenta)
	ctx.DrawString(lx+3, ly+7, "Mission", engine.StyleMagenta)

	for _, m := range gs.Missions {
		mx := m.X - offsetX + 1
		my := m.Y - offsetY + 1
		if mx > 0 && mx < w-1 && my > 0 && my < h-6 {
			ctx.SetCell(mx, my, '★', engine.StyleMagenta)
		}
	}

	ctx.DrawPanel(0, h-4, w, 3, "GEOSCAPE", engine.StyleDefault)
	fundsStr := fmt.Sprintf("Funds: $%dK", gs.Game.Funds/1000)
	timeStr := fmt.Sprintf("Time: %s", gs.Game.GameTime.Format("02/01/2006 15:04"))
	pauseStr := "RUNNING"
	if gs.Game.Paused {
		pauseStr = "PAUSED"
	}
	ctx.DrawString(2, h-3, fundsStr, engine.StyleGreen)
	ctx.DrawString(w/3, h-3, timeStr, engine.StyleDefault)
	ctx.DrawString(w*2/3, h-3, pauseStr, engine.StyleYellow)

	soldiersStr := fmt.Sprintf("Squad: %d", len(gs.Base.Soldiers))
	alienStr := fmt.Sprintf("Alien Activity: %d%%", gs.AlienActivity)
	missionStr := fmt.Sprintf("Missions: %d Active, %d Won", len(gs.Missions), gs.MissionsWon)

	ctx.DrawString(w-54, h-3, missionStr, engine.StyleMagenta)
	ctx.DrawString(w-28, h-3, alienStr, engine.StyleRed)
	ctx.DrawString(w-9, h-3, soldiersStr, engine.StyleCyan)

	if time.Since(gs.MessageTimer) < 4*time.Second && gs.Message != "" {
		ctx.DrawString(2, h-2, gs.Message, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := "[B]ase  [L]aunch  [A]utoresolve  [M]ission  Space=Pause  1-4=Speed  Q=Quit"
	if gs.Victory {
		help = "VICTORY ACHIEVED!  Q=Quit"
	}
	ctx.DrawString(1, h-1, help, engine.StyleGray)
}

func (gs *Geoscape) scrollMap(dx, dy int) {
	mw, mh := MapSize()
	gs.ScrollX += dx
	gs.ScrollY += dy
	if gs.ScrollX < 0 {
		gs.ScrollX = 0
	}
	if gs.ScrollY < 0 {
		gs.ScrollY = 0
	}
	if gs.ScrollX >= mw {
		gs.ScrollX = mw - 1
	}
	if gs.ScrollY >= mh {
		gs.ScrollY = mh - 1
	}
}

func (gs *Geoscape) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		gs.scrollMap(0, -1)
	case tcell.KeyDown:
		gs.scrollMap(0, 1)
	case tcell.KeyLeft:
		gs.scrollMap(-1, 0)
	case tcell.KeyRight:
		gs.scrollMap(1, 0)
	case tcell.KeyRune:
		switch e.Rune() {
		case 'b', 'B':
			gs.Game.PushState(engine.StateBase)
		case 'l', 'L':
			gs.LaunchInterceptor()
		case 'a', 'A':
			gs.Autoresolve()
		case 'm', 'M':
			gs.RespondToMission(0) // Default to first mission for now, could be improved
		case ' ':
			gs.TogglePause()
		case '1':
			gs.SetSpeed(1)
		case '2':
			gs.SetSpeed(2)
		case '3':
			gs.SetSpeed(3)
		case '4':
			gs.SetSpeed(4)
		case 'q', 'Q':
			gs.Game.Quit()
		}
	case tcell.KeyF5:
		gs.SaveGameToFile()
	case tcell.KeyF9:
		gs.LoadGameFromFile()
	}
}

func (gs *Geoscape) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := gs.Game.ScreenSize()

	if y >= h-4 && y <= h-2 {
		switch {
		case x >= 1 && x <= 8:
			gs.TogglePause()
		case x >= 10 && x <= 20:
			gs.LaunchInterceptor()
		}
		return
	}

	if y > 0 && y < h-4 && x > 0 && x < w-1 {
		mw, mh := MapSize()
		viewW := w - 2
		viewH := h - 6
		halfVW := viewW / 2
		halfVH := viewH / 2
		offsetX := gs.ScrollX - halfVW
		offsetY := gs.ScrollY - halfVH
		if offsetX < 0 {
			offsetX = 0
		}
		if offsetY < 0 {
			offsetY = 0
		}
		if offsetX+viewW > mw {
			offsetX = mw - viewW
		}
		if offsetX < 0 {
			offsetX = 0
		}
		if offsetY+viewH > mh {
			offsetY = mh - viewH
		}
		if offsetY < 0 {
			offsetY = 0
		}
		mx := x - 1 + offsetX
		my := y - 1 + offsetY
		if mx >= 0 && mx < mw && my >= 0 && my < mh {
			gs.Message = fmt.Sprintf("Cursor at [%d,%d]", mx, my)
			gs.MessageTimer = time.Now()
		}
	}
}
