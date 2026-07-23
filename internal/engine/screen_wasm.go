//go:build js && wasm

package engine

import (
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type wasmScreen struct {
	mu     sync.Mutex
	w, h   int
	eventQ chan tcell.Event
	stopQ  chan struct{}

	cursorX       int
	cursorY       int
	cursorVisible bool
}

func newWasmScreen(w, h int) *wasmScreen {
	return &wasmScreen{
		w:      w,
		h:      h,
		eventQ: make(chan tcell.Event, eventQueueSize),
		stopQ:  make(chan struct{}),
	}
}

func (s *wasmScreen) Init() error                                              { return nil }
func (s *wasmScreen) Fini()                                                    { close(s.stopQ) }
func (s *wasmScreen) Clear()                                                   {}
func (s *wasmScreen) Fill(_ rune, _ tcell.Style)                               {}
func (s *wasmScreen) Put(x, y int, str string, style tcell.Style) (string, int) { return "", 0 }
func (s *wasmScreen) PutStr(x, y int, str string)                              {}
func (s *wasmScreen) PutStrStyled(x, y int, str string, style tcell.Style)     {}
func (s *wasmScreen) Get(x, y int) (string, tcell.Style, int)                  { return " ", tcell.StyleDefault, 1 }
func (s *wasmScreen) SetContent(_ int, _ int, _ rune, _ []rune, _ tcell.Style) {}
func (s *wasmScreen) SetStyle(_ tcell.Style)                                   {}
func (s *wasmScreen) ShowCursor(x, y int) {
	s.mu.Lock()
	s.cursorX, s.cursorY = x, y
	s.cursorVisible = true
	s.mu.Unlock()
}
func (s *wasmScreen) HideCursor() {
	s.mu.Lock()
	s.cursorVisible = false
	s.mu.Unlock()
}
func (s *wasmScreen) SetCursorStyle(_ tcell.CursorStyle, _ ...color.Color) {}
func (s *wasmScreen) Size() (int, int)                                     { s.mu.Lock(); defer s.mu.Unlock(); return s.w, s.h }
func (s *wasmScreen) EventQ() chan tcell.Event                             { return s.eventQ }
func (s *wasmScreen) EnableMouse(...tcell.MouseFlags)                      {}
func (s *wasmScreen) DisableMouse()                                        {}
func (s *wasmScreen) EnablePaste()                                         {}
func (s *wasmScreen) DisablePaste()                                        {}
func (s *wasmScreen) EnableFocus()                                         {}
func (s *wasmScreen) DisableFocus()                                        {}
func (s *wasmScreen) Colors() int                                          { return 256 }
func (s *wasmScreen) Show()                                                {}
func (s *wasmScreen) Sync()                                                {}
func (s *wasmScreen) CharacterSet() string                                 { return "UTF-8" }
func (s *wasmScreen) RegisterRuneFallback(_ rune, _ string)                {}
func (s *wasmScreen) UnregisterRuneFallback(_ rune)                        {}
func (s *wasmScreen) Resize(_, _, _, _ int)                                {}
func (s *wasmScreen) Suspend() error                                       { return nil }
func (s *wasmScreen) Resume() error                                        { return nil }
func (s *wasmScreen) Beep() error                                          { return nil }
func (s *wasmScreen) SetSize(w, h int)                                     { s.mu.Lock(); s.w = w; s.h = h; s.mu.Unlock() }
func (s *wasmScreen) LockRegion(_, _, _, _ int, _ bool)                    {}
func (s *wasmScreen) Tty() (tcell.Tty, bool)                               { return nil, false }
func (s *wasmScreen) SetTitle(_ string)                                    {}
func (s *wasmScreen) SetClipboard(_ []byte)                                {}
func (s *wasmScreen) GetClipboard()                                        {}
func (s *wasmScreen) HasClipboard() bool                                   { return false }
func (s *wasmScreen) ShowNotification(_, _ string)                         {}
func (s *wasmScreen) KeyboardProtocol() tcell.KeyProtocol                  { return tcell.LegacyKeyboard }
func (s *wasmScreen) Terminal() (string, string)                           { return "", "" }

func (s *wasmScreen) InjectEvent(ev tcell.Event) {
	select {
	case s.eventQ <- ev:
	default:
	}
}

func NewScreenRawWASM(w, h int) *ScreenRaw {
	ws := newWasmScreen(w, h)
	return &ScreenRaw{screen: ws, width: w, height: h, fb: NewFrameBuffer(w, h)}
}
