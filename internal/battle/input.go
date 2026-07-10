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
		bs.ScrollX -= 3
		bs.ScrollY -= 2
		if bs.ScrollX < 0 {
			bs.ScrollX = 0
		}
		if bs.ScrollY < 0 {
			bs.ScrollY = 0
		}
		return
	}
	if buttons&tcell.WheelDown != 0 {
		bs.ScrollX += 3
		bs.ScrollY += 2
		return
	}

	if buttons&tcell.Button1 != 0 {
		if bs.State.CursorState == StateMovePlan {
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			bs.CursorX, bs.CursorY = mx, my
			return
		}
		unit := bs.Units.At(mx, my)
		if unit != nil && unit.Faction == 0 && unit.Alive {
			bs.Selected = unit
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
			bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
		} else {
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
		}
		bs.updateMovePath()
	} else if buttons&tcell.Button3 != 0 {
		if bs.State.CursorState == StateTargeting {
			bs.CursorX, bs.CursorY = mx, my
			bs.FireWeapon()
			return
		}
		if bs.State.CursorState == StateMovePlan {
			bs.MoveSelected()
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			return
		}
		bs.CursorX, bs.CursorY = mx, my
		bs.State.CursorState = StateMovePlan
		bs.updateMovePath()
	}
}

func (bs *Battlescape) updateMovePath() {
	if bs.State.CursorState != StateMovePlan || bs.Selected == nil {
		bs.State.MovePath = nil
		return
	}
	bs.State.MovePath = bs.CalculatePath(bs.Selected.X, bs.Selected.Y, bs.CursorX, bs.CursorY)
}
