package battle

import (
	"fmt"
	"sync"

	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type CursorState int

const (
	StateInspect CursorState = iota
	StateMovePlan
	StateTargeting
)

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
	if bs.PlayerLock > 0 && bs.Phase == PhasePlayerTurn {
		return
	}
	switch e.Key() {
	case tcell.KeyUp: 
		bs.MoveCursor(0, -1)
		bs.updateMovePath()
	case tcell.KeyDown: 
		bs.MoveCursor(0, 1)
		bs.updateMovePath()
	case tcell.KeyLeft: 
		bs.MoveCursor(-1, 0)
		bs.updateMovePath()
	case tcell.KeyRight: 
		bs.MoveCursor(1, 0)
		bs.updateMovePath()
	case tcell.KeyEnter: 
		bs.LeftClick()
	}
	
	switch e.Str() {
	case " ":
		bs.RightClick()
	case "w", "W":
		bs.Camera.Pan(0, -3)
	case "a", "A":
		bs.Camera.Pan(-3, 0)
	case "s", "S":
		bs.Camera.Pan(0, 3)
	case "d", "D":
		bs.Camera.Pan(3, 0)
	case "q", "Q": bs.cycleUnit(1)
	case "m", "M": 
		bs.State.CursorState = StateMovePlan
		if bs.Selected != nil {
			bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
		}
		bs.updateMovePath()
	case "f", "F": 
		bs.State.CursorState = StateTargeting
	case "e", "E": 
		bs.EndTurn()
	case "c", "C": 
		bs.Crouch()
	case "r", "R": 
		bs.Reload()
	case "g", "G": 
		bs.Grenade()
	case "o", "O": 
		bs.Game.PushState(engine.StateOptions)
	}
}

func (bs *Battlescape) handleMouse(e *tcell.EventMouse) {
	x, y := e.Position()
	mx, my := x+bs.ScrollX-1, y+bs.ScrollY-1

	buttons := e.Buttons()

	if buttons&tcell.WheelUp != 0 {
		bs.Camera.Pan(0, -3)
		return
	}
	if buttons&tcell.WheelDown != 0 {
		bs.Camera.Pan(0, 3)
		return
	}

	if buttons&tcell.Button1 != 0 {
		unit := bs.Units.At(mx, my)

		if unit != nil && unit.Faction == 0 && unit.Alive {
			bs.Selected = unit
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			bs.HoveredUnit = nil
			bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
			return
		}

		if unit != nil && unit.Faction == 1 && unit.Alive {
			bs.HoveredUnit = unit
			if bs.State.CursorState == StateTargeting && bs.CursorX == mx && bs.CursorY == my {
				bs.FireWeapon()
				bs.State.CursorState = StateInspect
			} else {
				bs.CursorX, bs.CursorY = mx, my
				bs.State.CursorState = StateTargeting
				bs.State.TargetUnit = unit
			}
			return
		}

		if bs.State.CursorState == StateMovePlan && bs.CursorX == mx && bs.CursorY == my {
			bs.MoveSelected()
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			return
		}

		if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateMovePlan
			bs.HoveredUnit = nil
			bs.updateMovePath()
		} else {
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			bs.HoveredUnit = nil
		}
	}
}

func (bs *Battlescape) updateMovePath() {
	if bs.State.CursorState != StateMovePlan || bs.Selected == nil {
		bs.State.MovePath = nil
		return
	}
	bs.State.MovePath = bs.CalculatePath(bs.Selected.X, bs.Selected.Y, bs.CursorX, bs.CursorY)
}
