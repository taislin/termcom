//go:build android

package android

import (
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/taislin/termcom/internal/engine"
)

// androidScreen implements tcell.Screen for Android.
// It captures all rendering into the FrameBuffer and signals the Java frontend
// via onShow. Events are injected from Java through InjectEvent.
type androidScreen struct {
	mu     sync.Mutex
	w, h   int
	fb     *engine.FrameBuffer
	eventQ chan tcell.Event
	stopQ  chan struct{}
	onShow func()

	cursorX       int
	cursorY       int
	cursorVisible bool
}

var _ tcell.Screen = (*androidScreen)(nil)

func newAndroidScreen(w, h int) *androidScreen {
	return &androidScreen{
		w:      w,
		h:      h,
		fb:     engine.NewFrameBuffer(w, h),
		eventQ: make(chan tcell.Event, 64),
		stopQ:  make(chan struct{}),
	}
}

func (as *androidScreen) Init() error  { return nil }
func (as *androidScreen) Fini()        { close(as.stopQ) }
func (as *androidScreen) Clear()       { as.fb.Clear() }
func (as *androidScreen) Fill(ch rune, st tcell.Style) {
	var attr tcell.AttrMask
	if st.HasBold() { attr |= tcell.AttrBold }
	if st.HasBlink() { attr |= tcell.AttrBlink }
	if st.HasReverse() { attr |= tcell.AttrReverse }
	if st.HasDim() { attr |= tcell.AttrDim }
	if st.HasItalic() { attr |= tcell.AttrItalic }
	if st.HasStrikeThrough() { attr |= tcell.AttrStrikeThrough }
	for y := 0; y < as.h; y++ {
		for x := 0; x < as.w; x++ {
			as.fb.Set(x, y, ch, st.GetForeground(), st.GetBackground(), attr)
		}
	}
}

func (as *androidScreen) SetContent(x, y int, ch rune, comb []rune, style tcell.Style) {
	var attr tcell.AttrMask
	if style.HasBold() { attr |= tcell.AttrBold }
	if style.HasBlink() { attr |= tcell.AttrBlink }
	if style.HasReverse() { attr |= tcell.AttrReverse }
	if style.HasDim() { attr |= tcell.AttrDim }
	if style.HasItalic() { attr |= tcell.AttrItalic }
	if style.HasStrikeThrough() { attr |= tcell.AttrStrikeThrough }
	as.fb.Set(x, y, ch, style.GetForeground(), style.GetBackground(), attr)
}

func (as *androidScreen) Put(x, y int, str string, style tcell.Style) (string, int) {
	runes := []rune(str)
	if len(runes) == 0 {
		return "", 0
	}
	ch := runes[0]
	as.SetContent(x, y, ch, nil, style)
	return string(runes[1:]), 1
}

func (as *androidScreen) PutStr(x, y int, str string) {
	as.PutStrStyled(x, y, str, tcell.StyleDefault)
}

func (as *androidScreen) PutStrStyled(x, y int, str string, style tcell.Style) {
	currX := x
	for _, ch := range str {
		if currX >= as.w {
			break
		}
		as.SetContent(currX, y, ch, nil, style)
		currX++
	}
}

func (as *androidScreen) Get(x, y int) (string, tcell.Style, int) {
	if x < 0 || x >= as.w || y < 0 || y >= as.h {
		return " ", tcell.StyleDefault, 1
	}
	cd := as.fb.Get(x, y)
	st := tcell.StyleDefault.Foreground(cd.Fg).Background(cd.Bg)
	if cd.Attr&tcell.AttrBold != 0 { st = st.Bold(true) }
	if cd.Attr&tcell.AttrBlink != 0 { st = st.Blink(true) }
	if cd.Attr&tcell.AttrReverse != 0 { st = st.Reverse(true) }
	if cd.Attr&tcell.AttrDim != 0 { st = st.Dim(true) }
	if cd.Attr&tcell.AttrItalic != 0 { st = st.Italic(true) }
	if cd.Attr&tcell.AttrStrikeThrough != 0 { st = st.StrikeThrough(true) }
	return string(cd.Ch), st, 1
}

func (as *androidScreen) Size() (int, int) {
	as.mu.Lock()
	defer as.mu.Unlock()
	return as.w, as.h
}

func (as *androidScreen) SetSize(w, h int) {
	as.mu.Lock()
	as.w = w
	as.h = h
	as.mu.Unlock()
	as.fb.Resize(w, h)
}

func (as *androidScreen) Show() {
	if as.onShow != nil {
		as.onShow()
	}
}

func (as *androidScreen) Sync() {
	as.Show()
}

func (as *androidScreen) EventQ() chan tcell.Event {
	return as.eventQ
}

func (as *androidScreen) InjectEvent(ev tcell.Event) {
	select {
	case as.eventQ <- ev:
	default:
	}
}

func (as *androidScreen) ShowCursor(x, y int) {
	as.mu.Lock()
	as.cursorX = x
	as.cursorY = y
	as.cursorVisible = true
	as.mu.Unlock()
}

func (as *androidScreen) HideCursor() {
	as.mu.Lock()
	as.cursorVisible = false
	as.mu.Unlock()
}

func (as *androidScreen) SetCursorStyle(_ tcell.CursorStyle, _ ...color.Color) {}

func (as *androidScreen) Colors() int { return 256 }

func (as *androidScreen) Beep() error           { return nil }
func (as *androidScreen) Suspend() error         { return nil }
func (as *androidScreen) Resume() error          { return nil }
func (as *androidScreen) CharacterSet() string    { return "UTF-8" }
func (as *androidScreen) EnableMouse(...tcell.MouseFlags) {}
func (as *androidScreen) DisableMouse()           {}
func (as *androidScreen) EnablePaste()            {}
func (as *androidScreen) DisablePaste()           {}
func (as *androidScreen) EnableFocus()            {}
func (as *androidScreen) DisableFocus()           {}
func (as *androidScreen) HasPendingEvent() bool   { return len(as.eventQ) > 0 }
func (as *androidScreen) SetStyle(_ tcell.Style)  {}
func (as *androidScreen) RegisterRuneFallback(_ rune, _ string) {}
func (as *androidScreen) UnregisterRuneFallback(_ rune) {}
func (as *androidScreen) Resize(_, _, _, _ int) {}
func (as *androidScreen) LockRegion(_, _, _, _ int, _ bool) {}
func (as *androidScreen) Tty() (tcell.Tty, bool) { return nil, false }
func (as *androidScreen) SetTitle(_ string) {}
func (as *androidScreen) SetClipboard(_ []byte) {}
func (as *androidScreen) GetClipboard() {}
func (as *androidScreen) HasClipboard() bool { return false }
func (as *androidScreen) ShowNotification(_, _ string) {}
func (as *androidScreen) KeyboardProtocol() tcell.KeyProtocol { return tcell.LegacyKeyboard }
func (as *androidScreen) Terminal() (string, string) { return "", "" }

// FrameBuffer returns the internal framebuffer for Marshalling.
func (as *androidScreen) FrameBuffer() *engine.FrameBuffer {
	return as.fb
}
