package engine

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// nullScreen is a minimal tcell.Screen implementation that does no real I/O.
// It is used by the web server to give the game engine a screen to render into
// without needing a real terminal.  All rendering output is captured via the
// FrameBuffer attached to ScreenRaw; see WebRenderer.
type nullScreen struct {
	mu     sync.Mutex
	w, h   int
	eventQ chan tcell.Event
	stopQ  chan struct{}
	OnShow func()

	cursorX       int
	cursorY       int
	cursorVisible bool
}

func newNullScreen(w, h int) *nullScreen {
	return &nullScreen{
		w:      w,
		h:      h,
		eventQ: make(chan tcell.Event, 64),
		stopQ:  make(chan struct{}),
	}
}

func (s *nullScreen) Init() error                                        { return nil }
func (s *nullScreen) Fini()                                              { close(s.stopQ) }
func (s *nullScreen) Clear()                                             {}
func (s *nullScreen) Fill(_ rune, _ tcell.Style)                        {}
func (s *nullScreen) Put(x, y int, str string, style tcell.Style) (string, int) { return "", 0 }
func (s *nullScreen) PutStr(x, y int, str string)                       {}
func (s *nullScreen) PutStrStyled(x, y int, str string, style tcell.Style) {}
func (s *nullScreen) Get(x, y int) (string, tcell.Style, int)           { return " ", tcell.StyleDefault, 1 }
func (s *nullScreen) SetContent(_ int, _ int, _ rune, _ []rune, _ tcell.Style) {}
func (s *nullScreen) SetStyle(_ tcell.Style)                            {}
func (s *nullScreen) ShowCursor(x, y int) {
	s.mu.Lock()
	s.cursorX = x
	s.cursorY = y
	s.cursorVisible = true
	s.mu.Unlock()
}
func (s *nullScreen) HideCursor() {
	s.mu.Lock()
	s.cursorVisible = false
	s.mu.Unlock()
}
func (s *nullScreen) SetCursorStyle(_ tcell.CursorStyle, _ ...color.Color) {}
func (s *nullScreen) Size() (int, int)                                   { s.mu.Lock(); defer s.mu.Unlock(); return s.w, s.h }
func (s *nullScreen) EventQ() chan tcell.Event                           { return s.eventQ }
func (s *nullScreen) EnableMouse(...tcell.MouseFlags)                    {}
func (s *nullScreen) DisableMouse()                                      {}
func (s *nullScreen) EnablePaste()                                       {}
func (s *nullScreen) DisablePaste()                                      {}
func (s *nullScreen) EnableFocus()                                       {}
func (s *nullScreen) DisableFocus()                                      {}
func (s *nullScreen) Colors() int                                        { return 256 }
func (s *nullScreen) Show()                                              { if s.OnShow != nil { s.OnShow() } }
func (s *nullScreen) Sync()                                              { if s.OnShow != nil { s.OnShow() } }
func (s *nullScreen) CharacterSet() string                               { return "UTF-8" }
func (s *nullScreen) RegisterRuneFallback(_ rune, _ string)             {}
func (s *nullScreen) UnregisterRuneFallback(_ rune)                     {}
func (s *nullScreen) Resize(_, _, _, _ int)                             {}
func (s *nullScreen) Suspend() error                                     { return nil }
func (s *nullScreen) Resume() error                                      { return nil }
func (s *nullScreen) Beep() error                                        { return nil }
func (s *nullScreen) SetSize(w, h int)                                  { s.mu.Lock(); s.w = w; s.h = h; s.mu.Unlock() }
func (s *nullScreen) LockRegion(_, _, _, _ int, _ bool)                 {}
func (s *nullScreen) Tty() (tcell.Tty, bool)                            { return nil, false }
func (s *nullScreen) SetTitle(_ string)                                  {}
func (s *nullScreen) SetClipboard(_ []byte)                             {}
func (s *nullScreen) GetClipboard()                                      {}
func (s *nullScreen) HasClipboard() bool                                 { return false }
func (s *nullScreen) ShowNotification(_, _ string)                       {}
func (s *nullScreen) KeyboardProtocol() tcell.KeyProtocol                { return tcell.LegacyKeyboard }
func (s *nullScreen) Terminal() (string, string)                         { return "", "" }

// InjectEvent posts an event into the screen's event queue for the game to consume.
func (s *nullScreen) InjectEvent(ev tcell.Event) {
	select {
	case s.eventQ <- ev:
	default:
	}
}

// NewScreenRawWeb creates a ScreenRaw backed by a nullScreen (no real terminal).
// w and h specify the initial virtual terminal dimensions.
func NewScreenRawWeb(w, h int) (*ScreenRaw, *nullScreen, error) {
	ns := newNullScreen(w, h)
	return &ScreenRaw{screen: ns, width: w, height: h, fb: NewFrameBuffer(w, h)}, ns, nil
}

// WebRenderer tracks the previous frame and emits only changed cells,
// eliminating the full-screen clear that causes flickering.
type WebRenderer struct {
	mu    sync.Mutex
	prev  []cellData
	w, h  int
	force bool // next Render call will do a full repaint
}

// NewWebRenderer creates a renderer. Call ForceRepaint() after a client
// reconnects or resizes to trigger a full repaint on the next frame.
func NewWebRenderer() *WebRenderer {
	return &WebRenderer{force: true}
}

// ForceRepaint schedules a full repaint on the next Render call.
// Call this when a new client connects or the terminal is resized.
func (wr *WebRenderer) ForceRepaint() {
	wr.mu.Lock()
	wr.force = true
	wr.mu.Unlock()
}

// Render compares scr's FrameBuffer against the last sent frame and returns
// an ANSI string containing only the differences.  On the first call, or
// after ForceRepaint(), it emits a full repaint (ESC[2J + all cells).
// Subsequent calls emit only cursor-positioned SGR sequences for dirty cells.
func (wr *WebRenderer) Render(scr *ScreenRaw) string {
	wr.mu.Lock()
	force := wr.force
	wr.mu.Unlock()

	w, h := scr.width, scr.height

	// Reinitialise on size change.
	if w != wr.w || h != wr.h {
		wr.w = w
		wr.h = h
		wr.prev = make([]cellData, w*h)
		force = true
	}

	var sb strings.Builder
	sb.Grow(w * h * 20) // generous pre-allocation

	// Hide cursor for the duration of the update.
	sb.WriteString("\x1b[?25l")

	if force {
		// Full repaint: clear screen, home, then write every cell.
		sb.WriteString("\x1b[2J\x1b[H\x1b[0m")

		var prevFg, prevBg tcell.Color
		var prevAttr tcell.AttrMask
		first := true

		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				cd := scr.fb.Get(col, row)
				if col < w && row < h && col+row*w < len(wr.prev) {
					wr.prev[row*w+col] = cd
				}
				ch := cd.ch
				if ch == 0 {
					ch = ' '
				}
				if first || cd.fg != prevFg || cd.bg != prevBg || cd.attr != prevAttr {
					sb.WriteString(sgrCode(cd.fg, cd.bg, cd.attr))
					prevFg, prevBg, prevAttr = cd.fg, cd.bg, cd.attr
					first = false
				}
				sb.WriteRune(ch)
			}
			if row < h-1 {
				sb.WriteString("\r\n")
			}
		}

		wr.mu.Lock()
		wr.force = false
		wr.mu.Unlock()
	} else {
		// Differential update: only emit cells that changed.
		var prevFg, prevBg tcell.Color
		var prevAttr tcell.AttrMask
		sgrSet := false
		prevRow, prevCol := -1, -1

		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				idx := row*w + col
				if idx >= len(wr.prev) {
					continue
				}
				cd := scr.fb.Get(col, row)
				if cd == wr.prev[idx] {
					continue // unchanged
				}
				wr.prev[idx] = cd

				ch := cd.ch
				if ch == 0 {
					ch = ' '
				}

				// Only move cursor if not already in position.
				if row != prevRow || col != prevCol {
					fmt.Fprintf(&sb, "\x1b[%d;%dH", row+1, col+1)
					prevRow, prevCol = row, col
				}

				// Emit SGR only when style changes.
				if !sgrSet || cd.fg != prevFg || cd.bg != prevBg || cd.attr != prevAttr {
					sb.WriteString(sgrCode(cd.fg, cd.bg, cd.attr))
					prevFg, prevBg, prevAttr = cd.fg, cd.bg, cd.attr
					sgrSet = true
				}

				sb.WriteRune(ch)
				prevCol++ // cursor advances after writing
			}
		}
	}

	// Restore cursor position and visibility at the end of frame.
	if ns, ok := scr.screen.(*nullScreen); ok {
		ns.mu.Lock()
		visible := ns.cursorVisible
		cx, cy := ns.cursorX, ns.cursorY
		ns.mu.Unlock()

		if visible && cx >= 0 && cy >= 0 {
			fmt.Fprintf(&sb, "\x1b[%d;%dH\x1b[?25h", cy+1, cx+1)
		} else {
			sb.WriteString("\x1b[?25l")
		}
	} else {
		sb.WriteString("\x1b[0m\x1b[?25h")
	}

	return sb.String()
}

// sgrCode returns the ANSI SGR escape sequence for the given fg/bg/attr.
func sgrCode(fg, bg tcell.Color, attr tcell.AttrMask) string {
	var parts []string

	parts = append(parts, "0") // reset

	if attr&tcell.AttrBold != 0 {
		parts = append(parts, "1")
	}
	if attr&tcell.AttrDim != 0 {
		parts = append(parts, "2")
	}
	if attr&tcell.AttrItalic != 0 {
		parts = append(parts, "3")
	}
	if attr&tcell.AttrBlink != 0 {
		parts = append(parts, "5")
	}
	if attr&tcell.AttrReverse != 0 {
		parts = append(parts, "7")
	}
	if attr&tcell.AttrStrikeThrough != 0 {
		parts = append(parts, "9")
	}

	parts = append(parts, colorSGR(fg, true))
	parts = append(parts, colorSGR(bg, false))

	return "\x1b[" + strings.Join(parts, ";") + "m"
}

// colorSGR returns the SGR parameter(s) for a tcell.Color.
// isFg=true for foreground, false for background.
func colorSGR(c tcell.Color, isFg bool) string {
	if !c.Valid() || c == tcell.ColorNone {
		if isFg {
			return "39"
		}
		return "49"
	}

	// Resolve named/indexed colours to true RGB values.
	tc := c.TrueColor()
	r, g, b := tc.RGB()
	if r < 0 {
		// Unknown colour — use terminal default.
		if isFg {
			return "39"
		}
		return "49"
	}

	if isFg {
		return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	}
	return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
}

