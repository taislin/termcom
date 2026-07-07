package battle

import (
	"fmt"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/civ13/ycom/internal/audio"
	"github.com/gdamore/tcell/v2"
)

type BattlePhase int

const (
	PhasePlayerTurn BattlePhase = iota
	PhaseAlienTurn
	PhaseVictory
	PhaseDefeat
)

type Battlescape struct {
	Game       *engine.Game
	Map        *BattleMap
	Units      UnitList
	AlienAIs   []*AlienAI
	Phase      BattlePhase
	Turn       int
	CursorX    int
	CursorY    int
	Selected   *Unit
	Message    string
	ScrollX    int
	ScrollY    int
	Squad      []*soldier.Soldier
	UFOName    string
	ExitTimer  int
}

func NewBattlescape(g *engine.Game, squad []*soldier.Soldier, ufoName string) *Battlescape {
	var m *BattleMap
	switch ufoName {
	case "Terror":
		m = GenerateTerrorSite(30, 24)
	case "Supply":
		m = GenerateUFOInterior(30, 24)
	case "Alien Base":
		m = GenerateCydonia(30, 24)
	default:
		m = GenerateCrashSite(30, 24)
	}

	bs := &Battlescape{
		Game:    g,
		Map:     m,
		Phase:   PhasePlayerTurn,
		Turn:    1,
		CursorX: m.Width / 2,
		CursorY: m.Height / 2,
		Message: fmt.Sprintf("Mission: Eliminate all hostiles! UFO: %s", ufoName),
		Squad:   squad,
		UFOName: ufoName,
	}

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

	return bs
}

func (bs *Battlescape) Update() {
	if bs.Phase == PhaseAlienTurn {
		bs.doAlienTurn()
		bs.Phase = PhasePlayerTurn
		bs.restorePlayerTU()
		bs.Turn++
		bs.checkVictory()
	}

	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		bs.ExitTimer++
		if bs.ExitTimer > 60 {
			bs.finishBattle()
		}
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

func (bs *Battlescape) doAlienTurn() {
	for _, ai := range bs.AlienAIs {
		if !ai.Unit.Alive {
			continue
		}
		ai.Unit.TU = ai.Unit.MaxTU
		humanUnits := bs.Units.Faction(0)
		ai.Update(bs.Units, bs.Map, humanUnits)
	}
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
	if len(aliens) == 0 {
		bs.Phase = PhaseVictory
		bs.Message = "MISSION COMPLETE! All hostiles eliminated."
	} else if len(humans) == 0 {
		bs.Phase = PhaseDefeat
		bs.Message = "MISSION FAILED! All soldiers KIA."
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
	viewW := scrW - 2
	viewH := scrH - 5
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
}

func (bs *Battlescape) SelectUnit() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		bs.Selected = unit
		bs.Message = fmt.Sprintf("Selected %s (HP:%d TU:%d)", unit.Soldier.Name, unit.HP, unit.TU)
	} else if bs.Selected != nil && unit == nil {
		if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
			bs.Message = fmt.Sprintf("Moved %s to [%d,%d]", bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY)
		} else {
			bs.Message = "Cannot move there."
		}
	} else {
		bs.cycleUnit(1)
	}
}

func (bs *Battlescape) FireWeapon() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	target := bs.Units.At(bs.CursorX, bs.CursorY)
	if target == nil || target.Faction == 0 {
		bs.Message = "No target."
		return
	}
	if !bs.Selected.CanSee(target.X, target.Y, bs.Map) {
		bs.Message = "Target not in line of sight."
		return
	}
	damage, hit := bs.Selected.FireAt(target)
	if hit {
		audio.PlayShoot()
		name := "alien"
		if target.AlienType != nil {
			name = target.AlienType.Name
		}
		bs.Message = fmt.Sprintf("HIT! %d damage to %s (HP:%d)", damage, name, target.HP)
	} else {
		audio.PlayShoot()
		bs.Message = "Missed!"
	}
}

func (bs *Battlescape) Reload() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 8 {
		bs.Message = "Not enough TU to reload."
		return
	}
	w := data.Weapons[bs.Selected.Weapon]
	if w.AmmoMax >= 99 {
		bs.Message = "Energy weapon — no reload needed."
		return
	}
	if w.AmmoCur >= w.AmmoMax {
		bs.Message = "Weapon already fully loaded."
		return
	}
	bs.Selected.TU -= 8
	w.AmmoCur = w.AmmoMax
	data.Weapons[bs.Selected.Weapon] = w
	audio.PlayClick()
	bs.Message = fmt.Sprintf("Reloaded %s. (%d/%d)", w.Name, w.AmmoCur, w.AmmoMax)
}

func (bs *Battlescape) EndTurn() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	audio.PlayClick()
	bs.Phase = PhaseAlienTurn
	bs.Message = "Alien turn..."
}

func (bs *Battlescape) Confirm() {
	bs.SelectUnit()
}

func (bs *Battlescape) Crouch() {
	if bs.Selected == nil {
		return
	}
	if bs.Selected.Crouching {
		bs.Selected.Crouching = false
		bs.Message = "Standing up."
	} else if bs.Selected.TU >= 4 {
		bs.Selected.Crouching = true
		bs.Selected.TU -= 4
		bs.Message = "Crouching."
	}
}

func (bs *Battlescape) Grenade() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 20 {
		bs.Message = "Not enough TU to throw grenade."
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
		bs.Message = "Target out of grenade range!"
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

	bs.Message = fmt.Sprintf("Grenade detonated at [%d,%d]!", ax, ay)
}

func (bs *Battlescape) UseMedikit() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 25 {
		bs.Message = "Not enough TU to use medikit."
		return
	}

	healAmount := 10
	mx := bs.CursorX
	my := bs.CursorY

	target := bs.Units.At(mx, my)
	if target == nil || target.Faction != 0 {
		bs.Message = "Select a friendly unit to heal."
		return
	}
	if target.HP >= target.MaxHP {
		bs.Message = "Soldier is already at full health."
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
	bs.Message = fmt.Sprintf("Healed %s for %d HP. (HP:%d/%d)", name, healAmount, target.HP, target.MaxHP)
}

func (bs *Battlescape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	viewW := w - 2
	viewH := h - 5

	for y := 0; y < viewH; y++ {
		for x := 0; x < viewW; x++ {
			mx := x + bs.ScrollX
			my := y + bs.ScrollY
			tile := bs.Map.At(mx, my)
			ch := TileChar(tile.Type)
			style := engine.StyleGreen

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
		if sx < 1 || sx >= w-1 || sy < 1 || sy >= viewH+1 {
			continue
		}
		ch := '@'
		style := engine.StyleGreenBold
		if u.Faction == 1 {
			ch = 'E'
			style = engine.StyleRedBold
			if u.AlienType != nil {
				ch = rune(u.AlienType.ShortName[0])
			}
		}
		if u == bs.Selected {
			style = style.Reverse(true)
		}
		if u.X == bs.CursorX && u.Y == bs.CursorY {
			style = style.Reverse(true)
		}
		ctx.SetCell(sx, sy, ch, style)
	}

	ctx.DrawPanel(0, h-4, w, 3, "BATTLESCAPE", engine.StyleDefault)
	turnStr := fmt.Sprintf("Turn: %d | %s", bs.Turn, bs.phaseStr())
	ctx.DrawString(2, h-3, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf("Sel: %s HP:%d/%d TU:%d/%d W:%s",
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.Weapons[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-3, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	cursorStr := fmt.Sprintf("[%d,%d] %s", bs.CursorX, bs.CursorY, tileTypeName(tile.Type))
	ctx.DrawString(w-30, h-3, cursorStr, engine.StyleGray)

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	ctx.DrawString(1, h-1, " hjkl=Move s=Sel f=Fire r=Reload g=Grenade m=Medikit e=End c=Crouch ?=Help", engine.StyleGray)
}

func (bs *Battlescape) phaseStr() string {
	switch bs.Phase {
	case PhasePlayerTurn:
		return "YOUR TURN"
	case PhaseAlienTurn:
		return "ALIEN TURN"
	case PhaseVictory:
		return "VICTORY"
	case PhaseDefeat:
		return "DEFEAT"
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
	case tcell.KeyRune:
		switch e.Rune() {
		case 'f', 'F':
			bs.FireWeapon()
		case 'r', 'R':
			bs.Reload()
		case 'e', 'E':
			bs.EndTurn()
		case 'h', 'H':
			bs.MoveCursor(-1, 0)
		case 'j', 'J':
			bs.MoveCursor(0, 1)
		case 'k', 'K':
			bs.MoveCursor(0, -1)
		case 'l', 'L':
			bs.MoveCursor(1, 0)
		case 's', 'S':
			bs.SelectUnit()
		case 'c', 'C':
			bs.Crouch()
		case 'g', 'G':
			bs.Grenade()
		case 'm', 'M':
			bs.UseMedikit()
		case 'n', 'N':
			bs.EndTurn()
		}
	}
}

func (bs *Battlescape) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()

	mx := x - 1 + bs.ScrollX
	my := y - 1 + bs.ScrollY
	if mx >= 0 && mx < bs.Map.Width && my >= 0 && my < bs.Map.Height {
		bs.CursorX = mx
		bs.CursorY = my
		if buttons&tcell.Button1 != 0 {
			bs.SelectUnit()
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
	bs.Message = fmt.Sprintf("Selected %s (HP:%d TU:%d)", bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU)
}

func tileTypeName(t TileType) string {
	switch t {
	case TileFloor:
		return "Floor"
	case TileWall:
		return "Wall"
	case TileDoor:
		return "Door"
	case TileGrass:
		return "Grass"
	case TileTree:
		return "Tree"
	case TileRock:
		return "Rock"
	case TileWater:
		return "Water"
	case TileUFOFloor:
		return "UFO Floor"
	case TileUFOWall:
		return "UFO Wall"
	}
	return "Unknown"
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
