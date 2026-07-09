package battle

import (
	"sync"
	"github.com/gdamore/tcell/v3"
	"github.com/civ13/ycom/internal/engine"
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
	case tcell.KeyUp: bs.MoveCursor(0, -1)
	case tcell.KeyDown: bs.MoveCursor(0, 1)
	case tcell.KeyLeft: bs.MoveCursor(-1, 0)
	case tcell.KeyRight: bs.MoveCursor(1, 0)
	case tcell.KeyEnter: bs.Confirm()
	}
	
	switch e.Str() {
	case "q", "Q": bs.cycleUnit(1)
	case "m", "M": bs.State.CursorState = StateMovePlan; bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
	case "f", "F": bs.State.CursorState = StateTargeting
	case "e", "E": bs.EndTurn()
	case "c", "C": bs.Crouch()
	case "r", "R": bs.Reload()
	case "g", "G": bs.Grenade()
	case "o", "O": bs.Game.PushState(engine.StateOptions)
	}
}

func (bs *Battlescape) handleMouse(e *tcell.EventMouse) {
	x, y := e.Position()
	// Adjust for scroll and sidebar
	mx, my := x + bs.ScrollX - 1, y + bs.ScrollY - 1
	
	buttons := e.Buttons()
	if buttons == tcell.WheelUp || buttons == tcell.WheelDown {
		bs.cycleUnit(1)
		return
	}

	if buttons&tcell.Button1 != 0 {
		bs.State.CursorState = StateInspect
		bs.CursorX, bs.CursorY = mx, my
		unit := bs.Units.At(mx, my)
		if unit != nil && unit.Faction == 0 {
			bs.Selected = unit
		}
	} else if buttons&tcell.Button3 != 0 {
		unit := bs.Units.At(mx, my)
		if unit != nil && unit.Faction == 1 {
			bs.State.CursorState = StateTargeting
			bs.State.TargetUnit = unit
		} else if bs.State.CursorState == StateInspect {
			bs.State.CursorState = StateMovePlan
			bs.CursorX, bs.CursorY = mx, my
		} else if bs.State.CursorState == StateMovePlan {
			bs.MoveSelected()
			bs.State.CursorState = StateInspect
		}
	}
}
