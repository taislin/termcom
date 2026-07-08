package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/civ13/ycom/internal/audio"
	"github.com/gdamore/tcell/v3"
)

type BattlePhase int

const (
	PhasePlayerTurn BattlePhase = iota
	PhaseAlienTurn
	PhaseVictory
	PhaseDefeat
)

const sidebarW = 24

type Projectile struct {
	FromX, FromY int
	ToX, ToY     int
	Progress     int
	Length       int
	Symbol       rune
	Style        tcell.Style
}

type AlienAction struct {
	Type   string // "move", "fire", "melee", "patrol"
	Unit   *Unit
	Target *Unit
	FromX, FromY int
	ToX, ToY     int
}

type Battlescape struct {
	Game       *engine.Game
	Map        *BattleMap
	Units      UnitList
	AlienAIs   []*AlienAI
	CivilianAIs []*CivilianAI
	Phase      BattlePhase
	Turn       int
	CursorX    int
	CursorY    int
	Selected   *Unit
	Message    string
	Log        []string
	ScrollX    int
	ScrollY    int
	Squad      []*soldier.Soldier
	UFOName    string
	ExitTimer  int

	AlienTurnQueue  []AlienAction
	AlienTurnIdx    int
	AlienTurnDelay  int
	Projectile      *Projectile
}

func (bs *Battlescape) AddMessage(msg string) {
	bs.Message = msg
	bs.Log = append(bs.Log, msg)
	if len(bs.Log) > 50 {
		bs.Log = bs.Log[len(bs.Log)-50:]
	}
}

func NewBattlescape(g *engine.Game, squad []*soldier.Soldier, ufoName string) *Battlescape {
	var m *BattleMap
	switch ufoName {
	case "Terror":
		m = GenerateTerrorSite(50, 50)
	case "Supply":
		m = GenerateUFOInterior(50, 50)
	case "Alien Base":
		m = GenerateCydonia(50, 50)
	case "Forest":
		m = GenerateForest(50, 50)
	case "Desert":
		m = GenerateDesert(50, 50)
	case "Polar":
		m = GeneratePolar(50, 50)
	default:
		m = GenerateCrashSite(50, 50)
	}

	bs := &Battlescape{
		Game:    g,
		Map:     m,
		Phase:   PhasePlayerTurn,
		Turn:    1,
		CursorX: m.Width / 2,
		CursorY: m.Height / 2,
		Squad:   squad,
		UFOName: ufoName,
	}

	bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_START"), ufoName))

	for i, s := range squad {
		if s.HP <= 0 {
			continue
		}
		u := NewSoldierUnit(s)
		u.X = 3 + i*2
		u.Y = m.Height - 3
		bs.Units = append(bs.Units, u)
	}

	alienRank := 0
	alienTypes := []*data.AlienType{
		data.GetAlienByRank(alienRank),
		data.GetAlienByRank(alienRank),
		data.GetAlienByRank(alienRank),
		data.GetAlienByRank(alienRank + 1),
		data.GetAlienByRank(alienRank + 1),
	}

	for _, at := range alienTypes {
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
		u.X = 10 + rand.Intn(m.Width-14)
		u.Y = 3 + rand.Intn(m.Height/2-4)
		bs.Units = append(bs.Units, u)
		ai := NewAlienAI(u)
		ai.PatrolX = u.X + rand.Intn(6) - 3
		ai.PatrolY = u.Y + rand.Intn(6) - 3
		bs.AlienAIs = append(bs.AlienAIs, ai)
	}

	if ufoName == "Terror" {
		civCount := 8 + rand.Intn(5)
		for i := 0; i < civCount; i++ {
			name := civNames[rand.Intn(len(civNames))]
			u := NewCivilianUnit(name)
			u.X = 5 + rand.Intn(m.Width-10)
			u.Y = m.Height/2 + rand.Intn(m.Height/2-5)
			if m.Passable(u.X, u.Y) {
				bs.Units = append(bs.Units, u)
				bs.CivilianAIs = append(bs.CivilianAIs, NewCivilianAI(u))
			}
		}
	}

	bs.ComputeFOVForTeam()

	return bs
}

func (bs *Battlescape) ComputeFOVForTeam() {
	bs.Map.ClearVisibility()
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive {
			bs.Map.ComputeFOV(u.X, u.Y)
		}
	}
}

func (bs *Battlescape) Update() {
	if bs.Phase == PhaseAlienTurn {
		if bs.Projectile != nil {
			bs.Projectile.Progress++
			if bs.Projectile.Progress >= bs.Projectile.Length {
				bs.Projectile = nil
			}
			return
		}

		if bs.AlienTurnDelay > 0 {
			bs.AlienTurnDelay--
			return
		}

		if bs.AlienTurnIdx < len(bs.AlienTurnQueue) {
			action := bs.AlienTurnQueue[bs.AlienTurnIdx]
			bs.AlienTurnIdx++
			bs.executeAlienAction(action)
			bs.AlienTurnDelay = 3
		} else {
			for _, cai := range bs.CivilianAIs {
				actions := cai.GenerateActions(bs.Units, bs.Map)
				for _, a := range actions {
					bs.executeAlienAction(a)
				}
			}
			bs.finishAlienTurn()
		}
	}

	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		bs.ExitTimer++
		if bs.ExitTimer > 60 {
			bs.finishBattle()
		}
	}
}

func (bs *Battlescape) executeAlienAction(action AlienAction) {
	switch action.Type {
	case "fire":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		damage, hit, err := action.Unit.FireAt(action.Target)
		if err != nil {
			return
		}
		dx := action.Target.X - action.Unit.X
		dy := action.Target.Y - action.Unit.Y
		length := int(math.Sqrt(float64(dx*dx+dy*dy)))
		if length < 1 {
			length = 1
		}
		symbol := '*'
		if data.RuleItems[action.Unit.Weapon].AmmoMax >= 99 {
			symbol = '|'
		}
		bs.Projectile = &Projectile{
			FromX: action.Unit.X, FromY: action.Unit.Y,
			ToX: action.Target.X, ToY: action.Target.Y,
			Progress: 0, Length: length,
			Symbol: symbol,
			Style:  engine.StyleYellow,
		}
		if hit {
			name := action.Target.Soldier.Name
			if action.Target.AlienType != nil {
				name = action.Target.AlienType.Name
			}
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_HIT"), name, damage))
			if !action.Target.Alive {
				bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_KILL"), name))
			}
		} else {
			bs.AddMessage(language.String("MSG_ALIEN_MISS"))
		}
	case "melee":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		damage := action.Unit.Strength + rand.Intn(10)
		damage -= action.Target.Armour
		if damage < 1 {
			damage = 1
		}
		action.Target.HP -= damage
		if action.Target.HP <= 0 {
			action.Target.Alive = false
		}
		bs.Projectile = &Projectile{
			FromX: action.Unit.X, FromY: action.Unit.Y,
			ToX: action.Target.X, ToY: action.Target.Y,
			Progress: 0, Length: 1,
			Symbol: 'X',
			Style:  engine.StyleRedBold,
		}
		name := action.Target.Soldier.Name
		if action.Target.AlienType != nil {
			name = action.Target.AlienType.Name
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_MELEE"), action.Unit.AlienType.Name, name, damage))
		if !action.Target.Alive {
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_KILL"), name))
		}
	case "move":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
	case "patrol":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
	}
}

func (bs *Battlescape) finishBattle() {
	won := bs.Phase == PhaseVictory

	// Sync soldier HP back to roster
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Soldier != nil {
			u.Soldier.HP = u.HP
			if u.HP <= 0 {
				u.Soldier.HP = 0
				u.Soldier.Wounds = 30
			} else if u.HP < u.MaxHP {
				dmg := u.MaxHP - u.HP
				u.Soldier.Wounds = dmg * 3
				if u.Soldier.Wounds > 30 {
					u.Soldier.Wounds = 30
				}
			}
		}
	}

	// Count alien kills
	alienKills := 0
	for _, u := range bs.Units {
		if u.Faction == 1 && !u.Alive {
			alienKills++
		}
	}

	// Award XP to surviving soldiers
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive && u.Soldier != nil {
			xp := alienKills * 5
			if won {
				xp += 10
			}
			u.Soldier.GainXP(xp)
			u.Soldier.Missions++
		}
	}

	// Collect loot — type-specific corpses
	var loot []string
	if won {
		corpseMap := map[string]string{
			"SEC": "corpse_sect",
			"SEL": "corpse_sect",
			"FLT": "corpse_float",
			"FLL": "corpse_float",
			"MUT": "corpse_muton",
			"MUL": "corpse_muton",
			"ETH": "corpse_ether",
			"EHL": "corpse_ether",
		}
		corpses := make(map[string]bool)
		for _, u := range bs.Units {
			if u.Faction == 1 && !u.Alive && u.AlienType != nil {
				if key, ok := corpseMap[u.AlienType.ShortName]; ok {
					corpses[key] = true
				}
			}
		}
		for key := range corpses {
			loot = append(loot, key)
		}
		if len(loot) == 0 {
			loot = append(loot, "alien_corpse")
		}
		if rand.Intn(100) < 40 {
			loot = append(loot, "alloys")
		}
		if rand.Intn(100) < 25 {
			loot = append(loot, "elerium")
		}
	}

	// Find surviving squad soldiers
	var surviving []*soldier.Soldier
	for _, s := range bs.Squad {
		for _, u := range bs.Units {
			if u.Faction == 0 && u.Soldier == s {
				surviving = append(surviving, u.Soldier)
				break
			}
		}
	}
	// Add soldiers that weren't deployed (HP <= 0 at start)
	for _, s := range bs.Squad {
		found := false
		for _, u := range bs.Units {
			if u.Soldier == s {
				found = true
				break
			}
		}
		if !found {
			surviving = append(surviving, s)
		}
	}

	// Award funds
	if won {
		bs.Game.Funds += 50000
	}

	bs.Game.ActiveBattle = &engine.BattleResult{
		Won:       won,
		Kills:     alienKills,
		Soldiers:  surviving,
		LootItems: loot,
	}
	bs.Game.PopState()
}

func (bs *Battlescape) finishAlienTurn() {
	bs.Phase = PhasePlayerTurn
	bs.restorePlayerTU()
	bs.ComputeFOVForTeam()
	bs.Turn++
	bs.checkVictory()
}

func (bs *Battlescape) restorePlayerTU() {
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive {
			u.TU = u.MaxTU
		}
	}
}

func (bs *Battlescape) checkVictory() {
	humans := bs.Units.Faction(0).Alive()
	aliens := bs.Units.Faction(1).Alive()
	civilians := bs.Units.Faction(2).Alive()
	if len(aliens) == 0 {
		bs.Phase = PhaseVictory
		if bs.UFOName == "Terror" {
			totalCiv := 0
			for _, u := range bs.Units {
				if u.Faction == 2 {
					totalCiv++
				}
			}
			saved := len(civilians)
			bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_COMPLETE_CIV"), saved, totalCiv))
		} else {
			bs.AddMessage(language.String("MSG_MISSION_COMPLETE"))
		}
	} else if len(humans) == 0 {
		bs.Phase = PhaseDefeat
		bs.AddMessage(language.String("MSG_MISSION_FAILED"))
	} else if bs.UFOName == "Terror" && len(civilians) == 0 && len(aliens) > 0 {
		bs.Phase = PhaseDefeat
		bs.AddMessage(language.String("MSG_MISSION_FAILED_CIV"))
	}
}

func (bs *Battlescape) MoveCursor(dx, dy int) {
	bs.CursorX += dx
	bs.CursorY += dy
	if bs.CursorX < 0 {
		bs.CursorX = 0
	}
	if bs.CursorY < 0 {
		bs.CursorY = 0
	}
	if bs.CursorX >= bs.Map.Width {
		bs.CursorX = bs.Map.Width - 1
	}
	if bs.CursorY >= bs.Map.Height {
		bs.CursorY = bs.Map.Height - 1
	}

	scrW, scrH := bs.Game.ScreenSize()
	viewW := scrW - sidebarW - 2
	viewH := scrH - 5
	if viewW < 10 {
		viewW = 10
	}
	if bs.CursorX < bs.ScrollX+2 {
		bs.ScrollX = bs.CursorX - 2
	}
	if bs.CursorX > bs.ScrollX+viewW-3 {
		bs.ScrollX = bs.CursorX - viewW + 3
	}
	if bs.CursorY < bs.ScrollY+2 {
		bs.ScrollY = bs.CursorY - 2
	}
	if bs.CursorY > bs.ScrollY+viewH-3 {
		bs.ScrollY = bs.CursorY - viewH + 3
	}
	if bs.ScrollX < 0 {
		bs.ScrollX = 0
	}
	if bs.ScrollY < 0 {
		bs.ScrollY = 0
	}
	maxScrollX := bs.Map.Width - viewW
	if maxScrollX < 0 {
		maxScrollX = 0
	}
	maxScrollY := bs.Map.Height - viewH
	if maxScrollY < 0 {
		maxScrollY = 0
	}
	if bs.ScrollX > maxScrollX {
		bs.ScrollX = maxScrollX
	}
	if bs.ScrollY > maxScrollY {
		bs.ScrollY = maxScrollY
	}
}

func (bs *Battlescape) SelectUnit() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		bs.Selected = unit
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
	} else {
		bs.cycleUnit(1)
	}
}

func (bs *Battlescape) Confirm() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)

	// Click on friendly unit → select it
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		bs.Selected = unit
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
		return
	}

	// Click on enemy → fire at it (if selected soldier can see it)
	if unit != nil && unit.Faction == 1 && unit.Alive {
		if bs.Selected == nil {
			bs.AddMessage(language.String("MSG_NO_SOLDIER_SELECTED"))
			return
		}
		if !bs.Selected.CanSee(unit.X, unit.Y, bs.Map) {
			bs.AddMessage(language.String("MSG_TARGET_NO_LOS"))
			return
		}
		damage, hit, err := bs.Selected.FireAt(unit)
		if err != nil {
			bs.AddMessage(err.Error())
			return
		}
		if hit {
			audio.PlayShoot()
			name := "alien"
			if unit.AlienType != nil {
				name = unit.AlienType.Name
			}
			bs.AddMessage(fmt.Sprintf(language.String("MSG_HIT_TARGET"), damage, name, unit.HP))
		} else {
			audio.PlayShoot()
			bs.AddMessage(language.String("MSG_MISSED"))
		}
		return
	}

	// Click on empty ground → move selected soldier there
	if unit == nil {
		if bs.Selected == nil {
			bs.AddMessage(language.String("MSG_NO_SOLDIER_SELECTED"))
			return
		}
		if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
			bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY))
			bs.ComputeFOVForTeam()
		} else {
			bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		}
		return
	}
}

func (bs *Battlescape) MoveSelected() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
		bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY))
		bs.ComputeFOVForTeam()
	} else {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
	}
}

func (bs *Battlescape) FireWeapon() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	target := bs.Units.At(bs.CursorX, bs.CursorY)
	if target == nil || target.Faction == 0 {
		bs.AddMessage(language.String("MSG_NO_TARGET"))
		return
	}
	if !bs.Selected.CanSee(target.X, target.Y, bs.Map) {
		bs.AddMessage(language.String("MSG_TARGET_NO_LOS"))
		return
	}
	damage, hit, err := bs.Selected.FireAt(target)
	if err != nil {
		bs.AddMessage(err.Error())
		return
	}
	if hit {
		audio.PlayShoot()
		name := "alien"
		if target.AlienType != nil {
			name = target.AlienType.Name
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_HIT_TARGET"), damage, name, target.HP))
	} else {
		audio.PlayShoot()
		bs.AddMessage(language.String("MSG_MISSED"))
	}
}

func (bs *Battlescape) Reload() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 8 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_RELOAD"))
		return
	}
	w := data.Weapons[bs.Selected.Weapon]
	if w.AmmoMax >= 99 {
		bs.AddMessage(language.String("MSG_ENERGY_WEAPON"))
		return
	}
	if w.AmmoCur >= w.AmmoMax {
		bs.AddMessage(language.String("MSG_WEAPON_LOADED"))
		return
	}
	bs.Selected.TU -= 8
	w.AmmoCur = w.AmmoMax
	data.Weapons[bs.Selected.Weapon] = w
	audio.PlayClick()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_RELOADED"), w.Name, w.AmmoCur, w.AmmoMax))
}

func (bs *Battlescape) EndTurn() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	audio.PlayClick()
	bs.Phase = PhaseAlienTurn
	bs.AddMessage(language.String("MSG_ALIEN_TURN"))

	bs.AlienTurnQueue = nil
	bs.AlienTurnIdx = 0

	for _, ai := range bs.AlienAIs {
		if !ai.Unit.Alive {
			continue
		}
		ai.Unit.TU = ai.Unit.MaxTU
		humanUnits := bs.Units.Faction(0)
		actions := ai.GenerateActions(bs.Units, bs.Map, humanUnits)
		bs.AlienTurnQueue = append(bs.AlienTurnQueue, actions...)
	}

	if len(bs.AlienTurnQueue) == 0 {
		bs.finishAlienTurn()
	} else {
		bs.AlienTurnDelay = 3
	}
}

func (bs *Battlescape) Crouch() {
	if bs.Selected == nil {
		return
	}
	if bs.Selected.Crouching {
		bs.Selected.Crouching = false
		bs.AddMessage(language.String("MSG_STANDING_UP"))
	} else if bs.Selected.TU >= 4 {
		bs.Selected.Crouching = true
		bs.Selected.TU -= 4
		bs.AddMessage(language.String("MSG_CROUCHING"))
	}
}

func (bs *Battlescape) Grenade() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 20 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_GRENADE"))
		return
	}

	grenadeRange := 6
	damage := 40 + bs.Selected.Strength*2
	ax := bs.CursorX
	ay := bs.CursorY
	dx := ax - bs.Selected.X
	dy := ay - bs.Selected.Y
	dist := dx*dx + dy*dy
	if dist > grenadeRange*grenadeRange {
		bs.AddMessage(language.String("MSG_GRENADE_OUT_OF_RANGE"))
		return
	}

	bs.Selected.TU -= 20

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		udx := u.X - ax
		udy := u.Y - ay
		udist := udx*udx + udy*udy
		if udist <= 4 {
			splashDmg := damage - udist*5
			if splashDmg < 5 {
				splashDmg = 5
			}
			u.HP -= splashDmg
			if u.HP <= 0 {
				u.HP = 0
				u.Alive = false
			}
		}
	}

	bs.AddMessage(fmt.Sprintf(language.String("MSG_GRENADE_DETONATED"), ax, ay))
}

func (bs *Battlescape) UseMedikit() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 25 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_MEDIKIT"))
		return
	}

	healAmount := 10
	mx := bs.CursorX
	my := bs.CursorY

	target := bs.Units.At(mx, my)
	if target == nil || target.Faction != 0 {
		bs.AddMessage(language.String("MSG_SELECT_FRIENDLY"))
		return
	}
	if target.HP >= target.MaxHP {
		bs.AddMessage(language.String("MSG_ALREADY_FULL_HP"))
		return
	}

	bs.Selected.TU -= 25
	target.HP += healAmount
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	name := "ally"
	if target.Soldier != nil {
		name = target.Soldier.Name
	}
	bs.AddMessage(fmt.Sprintf(language.String("MSG_HEALED"), name, healAmount, target.HP, target.MaxHP))
}

func (bs *Battlescape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	viewW := w - sidebarW - 2
	if viewW < 10 {
		viewW = 10
	}
	viewH := h - 5

	for y := 0; y < viewH; y++ {
		for x := 0; x < viewW; x++ {
			mx := x + bs.ScrollX
			my := y + bs.ScrollY
			tile := bs.Map.At(mx, my)

			if !tile.Seen {
				ctx.SetCell(x+1, y+1, ' ', engine.StyleDefault)
				continue
			}

			ch := TileChar(tile.Type)
			style := engine.StyleGreen

			if tile.Visible {
				switch tile.Type {
				case TileGrass:
					style = engine.StyleGreen
				case TileWall:
					style = engine.StyleGray
				case TileDoor:
					style = engine.StyleYellow
				case TileTree:
					style = engine.StyleGreen
				case TileRock:
					style = engine.StyleGray
				case TileWater:
					style = engine.StyleBlue
				case TileUFOFloor:
					style = engine.StyleCyan
				case TileUFOWall:
					style = engine.StyleCyanBold
				}
			} else {
				switch tile.Type {
				case TileGrass:
					style = engine.StyleGray
				case TileWall:
					style = engine.StyleGray
				case TileDoor:
					style = engine.StyleGray
				case TileTree:
					style = engine.StyleGray
				case TileRock:
					style = engine.StyleGray
				case TileWater:
					style = engine.StyleGray
				case TileUFOFloor:
					style = engine.StyleGray
				case TileUFOWall:
					style = engine.StyleGray
				}
			}

			if mx == bs.CursorX && my == bs.CursorY {
				style = style.Reverse(true)
			}

			ctx.SetCell(x+1, y+1, ch, style)
		}
	}

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		sx := u.X - bs.ScrollX + 1
		sy := u.Y - bs.ScrollY + 1
		if sx < 1 || sx >= viewW+1 || sy < 1 || sy < viewH+1 {
			if u.Faction == 1 && !bs.Map.IsVisible(u.X, u.Y) {
				continue
			}
		}
		if u.Faction == 1 && !bs.Map.IsVisible(u.X, u.Y) {
			continue
		}
		ch := '\u263B' // ☻ human face
		style := engine.StyleCyanBold
		if u.Faction == 1 {
			ch = '\u03A9' // Ω alien (Omega)
			style = engine.StyleRedBold
		} else if u.Faction == 2 {
			ch = 'c'
			style = engine.StyleGreen
		}
		if u == bs.Selected {
			style = style.Reverse(true)
		}
		if u.X == bs.CursorX && u.Y == bs.CursorY {
			style = style.Reverse(true)
		}
		ctx.SetCell(sx, sy, ch, style)
	}

	// Draw projectile
	if bs.Projectile != nil {
		p := bs.Projectile
		progress := float64(p.Progress) / float64(p.Length)
		px := float64(p.FromX) + float64(p.ToX-p.FromX)*progress
		py := float64(p.FromY) + float64(p.ToY-p.FromY)*progress
		sx := int(px) - bs.ScrollX + 1
		sy := int(py) - bs.ScrollY + 1
		if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
			ctx.SetCell(sx, sy, p.Symbol, p.Style)
		}
	}

	// Draw sidebar border
	sidebarX := viewW + 2
	for y := 0; y < viewH; y++ {
		ctx.SetCell(sidebarX-1, y+1, '|', engine.StyleGray)
	}

	// Draw unit info in sidebar
	sy := 1
	if bs.Selected != nil {
		ctx.DrawString(sidebarX, sy, language.String("SIDE_UNIT_INFO"), engine.StyleCyanBold)
		sy++

		name := bs.Selected.Soldier.Name
		if len(name) > sidebarW-1 {
			name = name[:sidebarW-1]
		}
		ctx.DrawString(sidebarX, sy, name, engine.StyleDefault.Bold(true))
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_HP"), bs.Selected.HP, bs.Selected.MaxHP), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_TU"), bs.Selected.TU, bs.Selected.MaxTU), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_ACC"), bs.Selected.Accuracy), engine.StyleDefault)
		sy++

		weaponName := data.RuleItems[bs.Selected.Weapon].ShortName
		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_WEAPON"), weaponName, bs.Selected.WeaponAmmo), engine.StyleDefault)
		sy++

		armourName := "None"
		if bs.Selected.Armour > 0 {
			for k, v := range data.Armors {
				if v.Undersuit == bs.Selected.Armour {
					armourName = k
					break
				}
			}
		}
		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_ARMOR"), armourName), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_POS"), bs.Selected.X, bs.Selected.Y), engine.StyleGray)
		sy++

		if bs.Selected.Crouching {
			ctx.DrawString(sidebarX, sy, language.String("SIDE_CROUCH"), engine.StyleYellow)
			sy++
		}
		sy++
	}

	// Draw log in sidebar
	logTitle := language.String("BATTLE_LOG")
	ctx.DrawString(sidebarX, sy, logTitle, engine.StyleCyanBold)
	sy++

	availableLines := viewH - sy
	logEntries := len(bs.Log)
	startIdx := 0
	if logEntries > availableLines {
		startIdx = logEntries - availableLines
	}
	for i := 0; i < availableLines && startIdx+i < logEntries; i++ {
		msg := bs.Log[startIdx+i]
		if len(msg) > sidebarW-1 {
			msg = msg[:sidebarW-1]
		}
		ctx.DrawString(sidebarX, sy+i, msg, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-4, w, 3, language.String("BATTLESCAPE"), engine.StyleDefault)
	turnStr := fmt.Sprintf(language.String("STATUS_TURN"), bs.Turn, bs.phaseStr())
	ctx.DrawString(2, h-3, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf(language.String("STATUS_SELECTED"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-3, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	cursorStr := fmt.Sprintf(language.String("STATUS_CURSOR"), bs.CursorX, bs.CursorY, tileTypeName(tile.Type))
	ctx.DrawString(w-30, h-3, cursorStr, engine.StyleGray)

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	ctx.DrawString(1, h-1, language.String("BATTLE_HELP"), engine.StyleGray)
}

func (bs *Battlescape) phaseStr() string {
	switch bs.Phase {
	case PhasePlayerTurn:
		return language.String("PHASE_YOUR_TURN")
	case PhaseAlienTurn:
		return language.String("PHASE_ALIEN_TURN")
	case PhaseVictory:
		return language.String("PHASE_VICTORY")
	case PhaseDefeat:
		return language.String("PHASE_DEFEAT")
	}
	return ""
}

func (bs *Battlescape) HandleKey(e *tcell.EventKey) {
	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		return
	}
	switch e.Key() {
	case tcell.KeyUp:
		bs.MoveCursor(0, -1)
	case tcell.KeyDown:
		bs.MoveCursor(0, 1)
	case tcell.KeyLeft:
		bs.MoveCursor(-1, 0)
	case tcell.KeyRight:
		bs.MoveCursor(1, 0)
	case tcell.KeyEnter:
		bs.Confirm()
	}
	switch e.Str() {
	case " ":
		bs.Confirm()
	case "f", "F":
		bs.FireWeapon()
	case "r", "R":
		bs.Reload()
	case "e", "E":
		bs.EndTurn()
	case "h", "H":
		bs.MoveCursor(-1, 0)
	case "j", "J":
		bs.MoveCursor(0, 1)
	case "k", "K":
		bs.MoveCursor(0, -1)
	case "l", "L":
		bs.MoveCursor(1, 0)
	case "s", "S":
		bs.cycleUnit(1)
	case "c", "C":
		bs.Crouch()
	case "g", "G":
		bs.Grenade()
	case "m", "M":
		bs.UseMedikit()
	case "n", "N":
		bs.EndTurn()
	case ".":
		bs.MoveSelected()
	}
}

func (bs *Battlescape) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()

	// Don't process clicks on the sidebar
	scrW, _ := bs.Game.ScreenSize()
	viewW := scrW - sidebarW - 2
	if x >= viewW+2 {
		return
	}

	mx := x - 1 + bs.ScrollX
	my := y - 1 + bs.ScrollY
	if mx >= 0 && mx < bs.Map.Width && my >= 0 && my < bs.Map.Height {
		bs.CursorX = mx
		bs.CursorY = my
		if buttons&tcell.Button1 != 0 {
			bs.Confirm()
		}
		if buttons&tcell.Button3 != 0 {
			bs.FireWeapon()
		}
	}

	if buttons&tcell.WheelUp != 0 {
		bs.cycleUnit(1)
	}
	if buttons&tcell.WheelDown != 0 {
		bs.cycleUnit(-1)
	}
}

func (bs *Battlescape) cycleUnit(dir int) {
	humans := bs.Units.Faction(0).Alive()
	if len(humans) == 0 {
		return
	}
	idx := -1
	for i, u := range humans {
		if u == bs.Selected {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(humans) - 1
	}
	if idx >= len(humans) {
		idx = 0
	}
	bs.Selected = humans[idx]
	bs.CursorX = bs.Selected.X
	bs.CursorY = bs.Selected.Y
	bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
}

func tileTypeName(t TileType) string {
	switch t {
	case TileFloor:
		return language.String("TILE_FLOOR")
	case TileWall:
		return language.String("TILE_WALL")
	case TileDoor:
		return language.String("TILE_DOOR")
	case TileGrass:
		return language.String("TILE_GRASS")
	case TileTree:
		return language.String("TILE_TREE")
	case TileRock:
		return language.String("TILE_ROCK")
	case TileWater:
		return language.String("TILE_WATER")
	case TileUFOFloor:
		return language.String("TILE_UFO_FLOOR")
	case TileUFOWall:
		return language.String("TILE_UFO_WALL")
	}
	return language.String("TILE_UNKNOWN")
}

func (ul UnitList) Faction(f int) UnitList {
	var result UnitList
	for _, u := range ul {
		if u.Faction == f {
			result = append(result, u)
		}
	}
	return result
}
