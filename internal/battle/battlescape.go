package battle

import (
	"fmt"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/soldier"
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
}

func NewBattlescape(g *engine.Game) *Battlescape {
	var m *BattleMap
	if rand.Intn(2) == 0 {
		m = GenerateCrashSite(30, 24)
	} else {
		m = GenerateTerrorSite(30, 24)
	}

	bs := &Battlescape{
		Game:    g,
		Map:     m,
		Phase:   PhasePlayerTurn,
		Turn:    1,
		CursorX: m.Width / 2,
		CursorY: m.Height / 2,
		Message: "Mission start! Eliminate all hostiles.",
	}

	squad := []*soldier.Soldier{
		soldier.NewSoldier("Rookie A"),
		soldier.NewSoldier("Rookie B"),
		soldier.NewSoldier("Rookie C"),
		soldier.NewSoldier("Rookie D"),
		soldier.NewSoldier("Rookie E"),
		soldier.NewSoldier("Rookie F"),
	}

	for i, s := range squad {
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
		data.GetAlienByRank(alienRank),
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
	viewH := scrH - 4
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
	if unit != nil && unit.Faction == 0 && unit.Alive {
		bs.Selected = unit
		bs.Message = fmt.Sprintf("Selected %s (HP:%d TU:%d)", unit.Soldier.Name, unit.HP, unit.TU)
	} else if bs.Selected != nil {
		if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
			bs.Message = fmt.Sprintf("Moved %s to [%d,%d]", bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY)
		} else {
			bs.Message = "Cannot move there."
		}
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
		name := "alien"
		if target.AlienType != nil {
			name = target.AlienType.Name
		}
		bs.Message = fmt.Sprintf("HIT! %d damage to %s (HP:%d)", damage, name, target.HP)
	} else {
		bs.Message = "Missed!"
	}
}

func (bs *Battlescape) Reload() {
	if bs.Selected == nil {
		return
	}
	w := data.Weapons[bs.Selected.Weapon]
	if w.AmmoMax < 99 {
		bs.Message = "Reloaded."
	}
}

func (bs *Battlescape) EndTurn() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
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
	if bs.Selected == nil {
		return
	}
	bs.Message = "Grenade thrown!"
}

func (bs *Battlescape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	viewW := w - 2
	viewH := h - 4

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
		ctx.SetCell(sx, sy, ch, style)
	}

	ctx.DrawPanel(0, h-3, w, 2, "BATTLESCAPE", engine.StyleDefault)
	turnStr := fmt.Sprintf("Turn: %d | %s", bs.Turn, bs.phaseStr())
	ctx.DrawString(2, h-2, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf("Sel: %s HP:%d/%d TU:%d/%d W:%s",
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.Weapons[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-2, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	cursorStr := fmt.Sprintf("[%d,%d] %s", bs.CursorX, bs.CursorY, tileTypeName(tile.Type))
	ctx.DrawString(w-30, h-2, cursorStr, engine.StyleGray)

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	ctx.DrawString(1, h-1, " hjkl=Move  s=Select  f=Fire  r=Reload  e=End Turn  c=Crouch", engine.StyleGray)

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}
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
		case 'h':
			bs.MoveCursor(-1, 0)
		case 'j':
			bs.MoveCursor(0, 1)
		case 'k':
			bs.MoveCursor(0, -1)
		case 'l':
			bs.MoveCursor(1, 0)
		case 's':
			bs.SelectUnit()
		case 'c':
			bs.Crouch()
		case 'g':
			bs.Grenade()
		case 'n':
			bs.EndTurn()
		}
	}
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
