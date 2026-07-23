package battle

import (
	"fmt"
	"sync"

	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type CursorState int

const (
	StateSmart CursorState = iota
	StateFire
	StateMove
)

func (s CursorState) String() string {
	switch s {
	case StateSmart:
		return language.String("MODE_SMART")
	case StateFire:
		return language.String("MODE_FIRE")
	case StateMove:
		return language.String("MODE_MOVE")
	}
	return "?"
}

const CamPanStep = 3
const helpBarCol = 1

type BattleState struct {
	mu           sync.RWMutex
	CursorState  CursorState
	MovePath     [][2]int
	TargetUnit   *Unit
}

func (bs *Battlescape) HandleEvent(ev tcell.Event) {
	bs.State.mu.Lock()
	defer bs.State.mu.Unlock()

	switch e := ev.(type) {
	case *tcell.EventKey:
		bs.handleKey(e)
	case *tcell.EventMouse:
		bs.handleMouse(e)
	}
}

func (bs *Battlescape) handleKey(e *tcell.EventKey) {
	if bs.QuitConfirm {
		switch {
		case e.Str() == "y" || e.Str() == "Y" || e.Key() == tcell.KeyEnter:
			bs.QuitConfirm = false
			bs.exitBattle()
		case e.Str() == "n" || e.Str() == "N" || e.Key() == tcell.KeyEscape:
			bs.QuitConfirm = false
		}
		return
	}

	if bs.ShowInventory {
		switch e.Key() {
		case tcell.KeyUp:
			if bs.InvCursor > 0 {
				bs.InvCursor--
			}
		case tcell.KeyDown:
			maxItems := len(bs.currentInvList())
			if bs.InvCursor < maxItems-1 {
				bs.InvCursor++
			}
		case tcell.KeyTab:
			bs.InvTab = 1 - bs.InvTab
			bs.InvCursor = 0
		case tcell.KeyEnter:
			bs.invActivate()
		case tcell.KeyEscape:
			bs.ShowInventory = false
		}
		switch e.Str() {
		case "i", "I":
			bs.ShowInventory = false
		case "u", "U":
			if bs.InvTab == 0 {
				bs.invUse()
			}
		case "d", "D":
			if bs.InvTab == 0 {
				bs.invDrop()
			}
		case "p", "P":
			if bs.InvTab == 1 {
				bs.invPickup()
			}
		}
		return
	}

	if bs.PlayerLock > 0 && bs.Phase == PhasePlayerTurn {
		return
	}

	switch e.Key() {
	case tcell.KeyEscape:
		if bs.State.CursorState == StateFire || bs.State.CursorState == StateMove {
			bs.State.CursorState = StateSmart
			bs.State.MovePath = nil
			bs.AddMessage(language.String("MODE_SMART"))
			return
		}
		if bs.Phase == PhasePlayerTurn || bs.Phase == PhaseAlienTurn {
			bs.QuitConfirm = true
		}
		return

	case tcell.KeyUp:
		bs.MoveCursor(0, -1)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case tcell.KeyDown:
		bs.MoveCursor(0, 1)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case tcell.KeyLeft:
		bs.MoveCursor(-1, 0)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case tcell.KeyRight:
		bs.MoveCursor(1, 0)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}

	case tcell.KeyEnter:
		bs.LeftClick()

	case tcell.KeyTab:
		bs.CycleFireMode()
	}

	switch e.Str() {
	case " ":
		bs.RightClick()

	case "w", "W":
		if bs.State.CursorState == StateSmart {
			bs.Camera.Pan(0, -CamPanStep)
		} else if bs.State.CursorState == StateMove {
			bs.moveUnitOneTile(0, -1)
		} else {
			bs.MoveCursor(0, -1)
			if bs.State.CursorState == StateMove {
				bs.updateMovePath()
			}
		}
	case "a", "A":
		if bs.State.CursorState == StateSmart {
			bs.Camera.Pan(-CamPanStep, 0)
		} else if bs.State.CursorState == StateMove {
			bs.moveUnitOneTile(-1, 0)
		} else {
			bs.MoveCursor(-1, 0)
			if bs.State.CursorState == StateMove {
				bs.updateMovePath()
			}
		}
	case "s", "S":
		if bs.State.CursorState == StateSmart {
			bs.Camera.Pan(0, CamPanStep)
		} else if bs.State.CursorState == StateMove {
			bs.moveUnitOneTile(0, 1)
		} else {
			bs.MoveCursor(0, 1)
			if bs.State.CursorState == StateMove {
				bs.updateMovePath()
			}
		}
	case "d", "D":
		if bs.State.CursorState == StateSmart {
			bs.Camera.Pan(CamPanStep, 0)
		} else if bs.State.CursorState == StateMove {
			bs.moveUnitOneTile(1, 0)
		} else {
			bs.MoveCursor(1, 0)
			if bs.State.CursorState == StateMove {
				bs.updateMovePath()
			}
		}

	case "x", "X":
		bs.State.CursorState = (bs.State.CursorState + 1) % 3
		bs.State.MovePath = nil
		if bs.State.CursorState == StateMove && bs.Selected != nil {
			bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
			bs.updateMovePath()
		}
		bs.AddMessage(bs.State.CursorState.String())

	case "q", "Q":
		bs.cycleUnit(1)

	case "e", "E":
		bs.EndTurn()
	case "c", "C":
		bs.Crouch()
	case "r", "R":
		bs.Reload()
	case "g", "G":
		bs.Grenade()
	case "p", "P":
		bs.PsiAttack()
	case "h", "H":
		bs.UseMedikit()
	case "y", "Y":
		bs.UseMotionScanner()
	case "t", "T":
		bs.PlaceMine()
	case "o", "O":
		bs.Game.PushState(engine.StateOptions)
	case "i", "I":
		bs.openInventory()
	}
}

func (bs *Battlescape) handleMouse(e *tcell.EventMouse) {
	x, y := e.Position()
	bs.mouseX, bs.mouseY = x, y
	bs.mouseActive = true

	_, h := bs.Game.ScreenSize()
	if y == h-1 {
		bs.clickHelpBar(x)
		return
	}

	mx, my := x+bs.ScrollX-1, y+bs.ScrollY-1

	buttons := e.Buttons()
	mods := e.Modifiers()

	if buttons&tcell.WheelUp != 0 {
		if mods&tcell.ModShift != 0 {
			bs.Camera.Pan(-CamPanStep, 0)
		} else {
			bs.Camera.Pan(0, -CamPanStep)
		}
		return
	}
	if buttons&tcell.WheelDown != 0 {
		if mods&tcell.ModShift != 0 {
			bs.Camera.Pan(CamPanStep, 0)
		} else {
			bs.Camera.Pan(0, CamPanStep)
		}
		return
	}
	if buttons&tcell.WheelLeft != 0 {
		bs.Camera.Pan(-CamPanStep, 0)
		return
	}
	if buttons&tcell.WheelRight != 0 {
		bs.Camera.Pan(CamPanStep, 0)
		return
	}

	if buttons&tcell.Button1 != 0 {
		unit := bs.Units.At(mx, my)
		mode := bs.State.CursorState

		switch mode {
		case StateSmart:
			bs.handleSmartClick(mx, my, unit)
		case StateFire:
			bs.handleFireClick(mx, my, unit)
		case StateMove:
			bs.handleMoveClick(mx, my, unit)
		}
		return
	}

	if buttons == 0 {
		unit := bs.Units.At(mx, my)
		if unit != nil && unit.Faction == FactionAlien && unit.Alive {
			bs.HoveredUnit = unit
		} else {
			bs.HoveredUnit = nil
		}
	}
}

func (bs *Battlescape) handleSmartClick(mx, my int, unit *Unit) {
	if unit != nil && unit.Faction == FactionHuman && unit.Alive {
		bs.Selected = unit
		bs.CursorX, bs.CursorY = mx, my
		bs.State.CursorState = StateSmart
		bs.State.MovePath = nil
		bs.HoveredUnit = nil
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
		return
	}

	if unit != nil && unit.Faction == FactionAlien && unit.Alive {
		bs.CursorX, bs.CursorY = mx, my
		bs.FireWeapon()
		return
	}

	if bs.Map.IsDestructible(mx, my) && unit == nil {
		if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
			bs.CursorX, bs.CursorY = mx, my
			bs.FireWeapon()
			return
		}
	}

	if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
		bs.CursorX, bs.CursorY = mx, my
		bs.State.CursorState = StateMove
		bs.updateMovePath()
		return
	}

	bs.CursorX, bs.CursorY = mx, my
	bs.State.CursorState = StateSmart
	bs.State.MovePath = nil
	bs.HoveredUnit = nil
}

func (bs *Battlescape) handleFireClick(mx, my int, unit *Unit) {
	bs.CursorX, bs.CursorY = mx, my
	if unit != nil && unit.Faction == FactionAlien && unit.Alive {
		bs.FireWeapon()
		return
	}
	if bs.Map.IsDestructible(mx, my) {
		if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
			bs.FireWeapon()
			return
		}
	}
}

func (bs *Battlescape) handleMoveClick(mx, my int, unit *Unit) {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		bs.CursorX, bs.CursorY = mx, my
		return
	}

	bs.CursorX, bs.CursorY = mx, my

	// If target tile is passable, move normally
	if bs.Map.Passable(mx, my) && bs.Units.At(mx, my) == nil {
		bs.MoveSelected()
		bs.State.MovePath = nil
		if bs.Selected != nil {
			bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
		}
		return
	}

	// Target is blocked — find nearest passable tile
	nx, ny := bs.findNearestPassable(mx, my, 10)
	if nx == mx && ny == my {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		return
	}
	bs.CursorX, bs.CursorY = nx, ny
	bs.MoveSelected()
	bs.State.MovePath = nil
	if bs.Selected != nil {
		bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
	}
}

func (bs *Battlescape) findNearestPassable(tx, ty, maxDist int) (int, int) {
	for d := 1; d <= maxDist; d++ {
		for x := tx - d; x <= tx+d; x++ {
			for y := ty - d; y <= ty+d; y++ {
				if x != tx-d && x != tx+d && y != ty-d && y != ty+d {
					continue
				}
				if x < 0 || x >= bs.Map.Width || y < 0 || y >= bs.Map.Height {
					continue
				}
				if bs.Map.Passable(x, y) && bs.Units.At(x, y) == nil {
					return x, y
				}
			}
		}
	}
	return tx, ty
}

func (bs *Battlescape) moveUnitOneTile(dx, dy int) {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	nx, ny := bs.Selected.X+dx, bs.Selected.Y+dy
	if nx < 0 || nx >= bs.Map.Width || ny < 0 || ny >= bs.Map.Height {
		return
	}
	if !bs.Map.Passable(nx, ny) || bs.Units.At(nx, ny) != nil {
		return
	}
	cost := bs.Map.MoveCost(nx, ny, &bs.Weather)
	crouchExtra := 0
	if bs.Selected.Crouching {
		crouchExtra = 4
	}
	if cost+crouchExtra > bs.Selected.TU {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		return
	}
	u := bs.Selected
	u.X, u.Y = nx, ny
	u.TU -= cost + crouchExtra
	bs.CursorX, bs.CursorY = nx, ny
	audio.PlayMove()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), u.Soldier.Name, u.X, u.Y))
	if t := bs.Map.At(nx, ny).Type; t == TileGlass || t == TileDebris {
		bs.EmitNoise(nx, ny, noiseAlertRadius)
	}
	if bs.Map.CollapseSkylight(nx, ny) {
		bs.UnitFallsThroughSkylight(u)
	}
	bs.ComputeFOVForTeam()
	bs.checkAlienReactionFire(u)
}

func (bs *Battlescape) updateMovePath() {
	if bs.State.CursorState != StateMove || bs.Selected == nil {
		bs.State.MovePath = nil
		return
	}
	bs.State.MovePath = bs.CalculatePath(bs.Selected.X, bs.Selected.Y, bs.CursorX, bs.CursorY)
}

func (bs *Battlescape) clickHelpBar(x int) {
	help := language.String("HELP_BATTLESCAPE")
	if bs.Map.NumLevels > 1 {
		help += language.String("HELP_STAIRS_SUFFIX")
	}
	col := helpBarCol
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
			key := string(runes[i+1 : end])
			bs.dispatchHelpKey(key)
			return
		}
		col = segEnd
		i = end + 1
	}
}

func (bs *Battlescape) dispatchHelpKey(key string) {
	switch key {
	case "q", "Q":
		bs.cycleUnit(1)
	case "x", "X":
		bs.State.CursorState = (bs.State.CursorState + 1) % 3
		bs.State.MovePath = nil
		if bs.State.CursorState == StateMove && bs.Selected != nil {
			bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
			bs.updateMovePath()
		}
		bs.AddMessage(bs.State.CursorState.String())
	case "r", "R":
		bs.Reload()
	case "g", "G":
		bs.Grenade()
	case "e", "E":
		bs.EndTurn()
	case "c", "C":
		bs.Crouch()
	case "Enter":
		bs.LeftClick()
	case "Space":
		bs.RightClick()
	case "\u2191":
		bs.MoveCursor(0, -1)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case "\u2193":
		bs.MoveCursor(0, 1)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case "\u2190":
		bs.MoveCursor(-1, 0)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	case "\u2192":
		bs.MoveCursor(1, 0)
		if bs.State.CursorState == StateMove {
			bs.updateMovePath()
		}
	}
}
