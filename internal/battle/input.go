package battle

import (
	"fmt"
	"sync"

	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
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
			bs.Camera.Pan(-3, 0) // Shift+WheelUp → pan left
		} else {
			bs.Camera.Pan(0, -3)
		}
		return
	}
	if buttons&tcell.WheelDown != 0 {
		if mods&tcell.ModShift != 0 {
			bs.Camera.Pan(3, 0) // Shift+WheelDown → pan right
		} else {
			bs.Camera.Pan(0, 3)
		}
		return
	}
	if buttons&tcell.WheelLeft != 0 {
		bs.Camera.Pan(-3, 0)
		return
	}
	if buttons&tcell.WheelRight != 0 {
		bs.Camera.Pan(3, 0)
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

func (bs *Battlescape) clickHelpBar(x int) {
	help := language.String("HELP_BATTLESCAPE")
	if bs.Map.NumLevels > 1 {
		help += language.String("HELP_STAIRS_SUFFIX")
	}
	col := 1
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
	case "f", "F":
		if bs.Selected != nil && bs.Selected.Alive {
			bs.State.CursorState = StateTargeting
		}
	case "r", "R":
		bs.Reload()
	case "g", "G":
		bs.Grenade()
	case "m", "M":
		bs.State.CursorState = StateMovePlan
		if bs.Selected != nil {
			bs.CursorX, bs.CursorY = bs.Selected.X, bs.Selected.Y
		}
		bs.updateMovePath()
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
		bs.updateMovePath()
		if bs.Selected != nil && bs.State.CursorState == StateMovePlan {
			bs.updateMovePath()
		}
	case "\u2193":
		bs.MoveCursor(0, 1)
		bs.updateMovePath()
	case "\u2190":
		bs.MoveCursor(-1, 0)
		bs.updateMovePath()
	case "\u2192":
		bs.MoveCursor(1, 0)
		bs.updateMovePath()
	}
}
