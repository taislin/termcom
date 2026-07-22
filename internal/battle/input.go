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

// Camera pan step (tiles) for keyboard and wheel scrolling.
const CamPanStep = 3

// Help-bar start column for hotkey hit-testing.
const helpBarCol = 1

type BattleState struct {
	mu           sync.RWMutex
	CursorState  CursorState
	MovePath     [][2]int
	TargetUnit   *Unit
}

// HandleEvent handles keyboard/mouse input for the battlescape.
// IMPORTANT: Called methods (handleKey, handleMouse, etc.) must NOT acquire
// bs.State.mu to avoid deadlock — this function already holds the lock.
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
	// Quit confirmation intercept
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

	// Inventory overlay intercept
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
		if bs.Phase == PhasePlayerTurn || bs.Phase == PhaseAlienTurn {
			bs.QuitConfirm = true
		}
		return
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
	case tcell.KeyTab:
		bs.CycleFireMode()
	}
	
	switch e.Str() {
	case " ":
		bs.RightClick()
	case "w", "W":
		bs.Camera.Pan(0, -CamPanStep)
	case "a", "A":
		bs.Camera.Pan(-CamPanStep, 0)
	case "s", "S":
		bs.Camera.Pan(0, CamPanStep)
	case "d", "D":
		bs.Camera.Pan(CamPanStep, 0)
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
			bs.Camera.Pan(-CamPanStep, 0) // Shift+WheelUp → pan left
		} else {
			bs.Camera.Pan(0, -CamPanStep)
		}
		return
	}
	if buttons&tcell.WheelDown != 0 {
		if mods&tcell.ModShift != 0 {
			bs.Camera.Pan(CamPanStep, 0) // Shift+WheelDown → pan right
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

		if unit != nil && unit.Faction == FactionHuman && unit.Alive {
			bs.Selected = unit
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			bs.HoveredUnit = nil
			bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
			return
		}

		if unit != nil && unit.Faction == FactionAlien && unit.Alive {
			bs.HoveredUnit = unit
			// Fire mode or move mode with enemy: fire immediately
			if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
				bs.CursorX, bs.CursorY = mx, my
				bs.State.TargetUnit = unit
				bs.State.CursorState = StateTargeting
				bs.FireWeapon()
				bs.State.CursorState = StateInspect
			} else {
				bs.CursorX, bs.CursorY = mx, my
				bs.State.CursorState = StateTargeting
				bs.State.TargetUnit = unit
			}
			return
		}

		// Target destructible terrain (streetlamps, pipes, fuel pumps…)
		if bs.Selected != nil && bs.Phase == PhasePlayerTurn &&
			bs.Map.IsDestructible(mx, my) && unit == nil {
			bs.CursorX, bs.CursorY = mx, my
			bs.State.TargetUnit = nil
			bs.State.CursorState = StateTargeting
			bs.FireWeapon()
			bs.State.CursorState = StateInspect
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
			// Fire mode: don't switch to move on empty space
			if bs.State.CursorState != StateTargeting {
				bs.State.CursorState = StateMovePlan
			}
			bs.HoveredUnit = nil
			bs.updateMovePath()
		} else {
			bs.CursorX, bs.CursorY = mx, my
			bs.State.CursorState = StateInspect
			bs.State.MovePath = nil
			bs.HoveredUnit = nil
		}
		return
	}

	// Mouse hover (no button): show enemy info when hovering over aliens
	if buttons == 0 {
		unit := bs.Units.At(mx, my)
		if unit != nil && unit.Faction == FactionAlien && unit.Alive {
			bs.HoveredUnit = unit
		} else {
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
