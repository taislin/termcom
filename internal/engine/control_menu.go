package engine

import "github.com/gdamore/tcell/v3"

type Rect struct {
	X, Y, W, H int
}

type ControlButton struct {
	Label   string
	Action  func() // called when tapped
	Enabled bool
	Hotkey  string // keyboard equivalent (displayed on button)
}

type ControlMenu struct {
	Visible    bool
	Buttons    []ControlButton
	ScrollOff  int
	TouchFirst bool // auto-showed on first touch
	screenW    int
	screenH    int
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
	for i := range cm.Buttons {
		if cm.Buttons[i].Action != nil {
			cm.Buttons[i].Enabled = true
		}
	}
}

func (cm *ControlMenu) SetScreenSize(w, h int) {
	cm.screenW = w
	cm.screenH = h
}

func (cm *ControlMenu) buttonRects() []Rect {
	if len(cm.Buttons) == 0 {
		return nil
	}
	w, h := cm.screenW, cm.screenH
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

func (cm *ControlMenu) Render(s *ScreenRaw) {
	if HideTouchOverlay || !cm.Visible || !Config.TouchMode {
		return
	}
	w, h := s.Size()
	cm.SetScreenSize(w, h)

	// Hamburger button (always visible in touch mode)
	s.DrawString(w-4, 0, "[=]", StyleHighlight)

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
	if !cm.Visible || !Config.TouchMode {
		return false
	}
	if ev.Buttons() == tcell.ButtonNone {
		return false
	}
	x, y := ev.Position()

	// Check hamburger
	if x >= cm.screenW-4 && x <= cm.screenW-1 && y == 0 {
		cm.Toggle()
		return true
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
	return !HideTouchOverlay && cm.Visible && Config.TouchMode && x >= cm.screenW-4 && x <= cm.screenW-1 && y == 0
}
