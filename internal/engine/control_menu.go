package engine

import "github.com/gdamore/tcell/v3"

type Rect struct {
	X, Y, W, H int
}

type ControlButton struct {
	Label   string
	Action  func() // called when tapped (may be nil for disabled buttons)
	Enabled bool
	Hotkey  string // keyboard equivalent (displayed on button)
}

type ControlMenu struct {
	Visible    bool
	Buttons    []ControlButton
	TouchFirst bool // auto-showed on first touch
	AlwaysShow bool // pin bar to bottom edge in touch mode
	screenW    int
	screenH    int
	game       *Game // optional back-reference for live screen size
}

// SetGame wires the menu to the active game so it can read the current screen
// size on demand (e.g. during input handling) instead of relying solely on a
// cached value that may be stale after a resize.
func (cm *ControlMenu) SetGame(g *Game) { cm.game = g }

// screenSize returns the live screen dimensions when a game is wired, falling
// back to the last value supplied via SetScreenSize otherwise.
func (cm *ControlMenu) screenSize() (int, int) {
	if cm.game != nil {
		return cm.game.ScreenSize()
	}
	return cm.screenW, cm.screenH
}

var HideTouchOverlay = false

var Menu = &ControlMenu{}

func (cm *ControlMenu) Toggle() {
	cm.Visible = !cm.Visible
}

func (cm *ControlMenu) Show() {
	cm.Visible = true
}

func (cm *ControlMenu) Hide() {
	cm.Visible = false
}

func (cm *ControlMenu) SetButtons(btns []ControlButton) {
	cm.Buttons = btns
}

func (cm *ControlMenu) SetScreenSize(w, h int) {
	cm.screenW = w
	cm.screenH = h
}

func (cm *ControlMenu) buttonRects() []Rect {
	if len(cm.Buttons) == 0 {
		return nil
	}
	w, h := cm.screenSize()
	if w == 0 || h == 0 {
		return nil
	}
	btnH := 3
	btnMinW := 10
	padX := 1
	padY := 1

	cols := 3
	if w < 60 {
		cols = 2
	}
	if w < 40 {
		cols = 1
	}

	btnW := (w - padX*(cols+1)) / cols
	if btnW < btnMinW {
		btnW = btnMinW
	}

	totalRows := (len(cm.Buttons) + cols - 1) / cols
	panelH := totalRows*btnH + padY*2
	startY := h - panelH - 1
	startX := padX

	rects := make([]Rect, len(cm.Buttons))
	for i := range cm.Buttons {
		row := i / cols
		col := i % cols
		x := startX + col*(btnW+padX)
		y := startY + row*btnH
		rects[i] = Rect{X: x, Y: y, W: btnW, H: btnH}
	}
	return rects
}

// ReservedBottom returns the number of rows the touch control panel occupies at
// the bottom of the screen, so screens can lay out their content above it. It
// returns 0 when the control bar is not shown (non-touch or hidden).
func (cm *ControlMenu) ReservedBottom(w, h int) int {
	if HideTouchOverlay || !Config.TouchMode {
		return 0
	}
	if !cm.AlwaysShow && !cm.Visible {
		return 0
	}
	cm.SetScreenSize(w, h)
	rects := cm.buttonRects()
	if len(rects) == 0 {
		return 0
	}
	first := rects[0]
	// Panel border adds one row above the first button row; include the trailing
	// screen edge margin as well.
	return h - (first.Y - 1)
}

func (cm *ControlMenu) Render(s *ScreenRaw) {
	if HideTouchOverlay || !Config.TouchMode {
		return
	}
	if !cm.AlwaysShow && !cm.Visible {
		return
	}
	w, h := s.Size()
	cm.SetScreenSize(w, h)

	// Hamburger button (only when the bar is not pinned to the bottom).
	if !cm.AlwaysShow {
		s.DrawString(w-4, 0, "[=]", StyleHighlight)
	}

	// Draw panel background
	btns := cm.Buttons
	if len(btns) == 0 {
		return
	}
	rects := cm.buttonRects()
	if len(rects) == 0 {
		return
	}

	first := rects[0]
	last := rects[len(rects)-1]
	panelX := first.X - 1
	panelY := first.Y - 1
	panelW := last.X + last.W - panelX + 1
	panelH := last.Y + last.H - panelY + 1

	for fy := panelY; fy < panelY+panelH; fy++ {
		for fx := panelX; fx < panelX+panelW; fx++ {
			s.SetCell(fx, fy, ' ', StyleGray)
		}
	}
	s.DrawBorder(panelX, panelY, panelW, panelH, StyleGray)

	for i, btn := range btns {
		r := rects[i]
		style := StyleDefault
		if !btn.Enabled {
			style = StyleGray
		}
		for dy := 0; dy < r.H; dy++ {
			for dx := 0; dx < r.W; dx++ {
				s.SetCell(r.X+dx, r.Y+dy, ' ', style)
			}
		}
		s.DrawBorder(r.X, r.Y, r.W, r.H, style)

		label := btn.Label
		lw := StringWidth(label)
		if lw > r.W-4 {
			runes := []rune(label)
			for len(runes) > 0 && StringWidth(string(runes)) > r.W-4 {
				runes = runes[:len(runes)-1]
			}
			label = string(runes)
			lw = StringWidth(label)
		}
		lx := r.X + (r.W-lw)/2
		ly := r.Y + 1
		s.DrawString(lx, ly, label, style)

		if btn.Hotkey != "" {
			hk := "[" + btn.Hotkey + "]"
			hw := StringWidth(hk)
			hx := r.X + (r.W-hw)/2
			s.DrawString(hx, r.Y+2, hk, StyleHotkey)
		}
	}
}

func (cm *ControlMenu) HandleMouse(ev *tcell.EventMouse) bool {
	if !Config.TouchMode {
		return false
	}
	if ev.Buttons() == tcell.ButtonNone {
		return false
	}
	x, y := ev.Position()

	// Hamburger toggles the bar only when it is not pinned (always-show mode
	// keeps the bar permanently visible at the bottom).
	if !cm.AlwaysShow {
		w, _ := cm.screenSize()
		if x >= w-4 && x <= w-1 && y == 0 {
			cm.Toggle()
			return true
		}
	}

	if HideTouchOverlay {
		return false
	}
	if !cm.AlwaysShow && !cm.Visible {
		return false
	}

	// Check buttons
	rects := cm.buttonRects()
	for i, r := range rects {
		if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			if i < len(cm.Buttons) && cm.Buttons[i].Enabled && cm.Buttons[i].Action != nil {
				cm.Buttons[i].Action()
			}
			return true
		}
	}
	return false
}

func (cm *ControlMenu) HamburgerHit(x, y int) bool {
	w, _ := cm.screenSize()
	return !HideTouchOverlay && !cm.AlwaysShow && Config.TouchMode && x >= w-4 && x <= w-1 && y == 0
}
