package engine

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type ScreenRaw struct {
	screen tcell.Screen
	width  int
	height int
	fb     *FrameBuffer
}

func NewScreenRaw() (*ScreenRaw, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.EnableMouse()
	s.SetStyle(tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm15))
	w, h := s.Size()
	return &ScreenRaw{screen: s, width: w, height: h, fb: NewFrameBuffer(w, h)}, nil
}

func (s *ScreenRaw) Close() {
	s.screen.Fini()
}

func (s *ScreenRaw) Clear() {
	s.screen.Clear()
	s.fb.Clear()
}

func (s *ScreenRaw) Flush() {
	s.screen.Show()
}

func (s *ScreenRaw) Size() (int, int) {
	return s.width, s.height
}

func (s *ScreenRaw) UpdateSize() {
	w, h := s.screen.Size()
	s.width = w
	s.height = h
	s.fb.Resize(w, h)
}

func (s *ScreenRaw) SetCell(x, y int, ch rune, style tcell.Style) {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		s.screen.SetContent(x, y, ch, nil, style)
		fg := style.GetForeground()
		bg := style.GetBackground()
		s.fb.Set(x, y, ch, fg, bg, attrMaskFromStyle(style))
	}
}

func attrMaskFromStyle(style tcell.Style) tcell.AttrMask {
	var attr tcell.AttrMask
	if style.HasBold() {
		attr |= tcell.AttrBold
	}
	if style.HasBlink() {
		attr |= tcell.AttrBlink
	}
	if style.HasReverse() {
		attr |= tcell.AttrReverse
	}
	if style.HasDim() {
		attr |= tcell.AttrDim
	}
	if style.HasItalic() {
		attr |= tcell.AttrItalic
	}
	if style.HasStrikeThrough() {
		attr |= tcell.AttrStrikeThrough
	}
	return attr
}

func styleFromCell(cd cellData) tcell.Style {
	st := tcell.StyleDefault.Foreground(cd.fg).Background(cd.bg)
	if cd.attr&tcell.AttrBold != 0 {
		st = st.Bold(true)
	}
	if cd.attr&tcell.AttrBlink != 0 {
		st = st.Blink(true)
	}
	if cd.attr&tcell.AttrReverse != 0 {
		st = st.Reverse(true)
	}
	if cd.attr&tcell.AttrDim != 0 {
		st = st.Dim(true)
	}
	if cd.attr&tcell.AttrItalic != 0 {
		st = st.Italic(true)
	}
	if cd.attr&tcell.AttrStrikeThrough != 0 {
		st = st.StrikeThrough(true)
	}
	return st
}

func (s *ScreenRaw) Peek(x, y int) (rune, tcell.Style) {
	cd := s.fb.Get(x, y)
	return cd.ch, styleFromCell(cd)
}

func (s *ScreenRaw) FrameBuffer() *FrameBuffer {
	return s.fb
}

func (s *ScreenRaw) DrawString(x, y int, str string, style tcell.Style) {
	for i, ch := range str {
		s.SetCell(x+i, y, ch, style)
	}
}

func (s *ScreenRaw) DrawMarkupString(x, y int, str string, normalStyle, highlightStyle tcell.Style) {
	currX := x
	highlight := false
	for _, ch := range str {
		if ch == '[' {
			highlight = true
			continue
		} else if ch == ']' {
			highlight = false
			continue
		}

		style := normalStyle
		if highlight {
			style = highlightStyle
		}
		s.SetCell(currX, y, ch, style)
		currX++
	}
}

func (s *ScreenRaw) DrawRect(x, y, w, h int, ch rune, style tcell.Style) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			s.SetCell(x+dx, y+dy, ch, style)
		}
	}
}

func (s *ScreenRaw) DrawBorder(x, y, w, h int, style tcell.Style) {
	s.SetCell(x, y, '┌', style)
	s.SetCell(x+w-1, y, '┐', style)
	s.SetCell(x, y+h-1, '└', style)
	s.SetCell(x+w-1, y+h-1, '┘', style)

	for i := 1; i < w-1; i++ {
		s.SetCell(x+i, y, '─', style)
		s.SetCell(x+i, y+h-1, '─', style)
	}
	for i := 1; i < h-1; i++ {
		s.SetCell(x, y+i, '│', style)
		s.SetCell(x+w-1, y+i, '│', style)
	}
}

func (s *ScreenRaw) DrawPanel(x, y, w, h int, title string, style tcell.Style) {
	s.DrawBorder(x, y, w, h, style)
	if title != "" {
		titleStr := "┤ " + title + " ├"
		s.DrawString(x+w/2-len(titleStr)/2, y, titleStr, style)
	}
}

var StyleDefault = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm15)
var StyleHighlight = tcell.StyleDefault.Background(color.DarkBlue).Foreground(color.XTerm15)
var StyleRed = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm9)
var StyleGreen = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm2)
var StyleBlue = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm12)
var StyleYellow = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm11)
var StyleCyan = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm6)
var StyleMagenta = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm13)
var StyleGray = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm8)

var StyleCyanBold = StyleCyan.Bold(true)
var StyleRedBold = StyleRed.Bold(true)
var StyleGreenBold = StyleGreen.Bold(true)
var StyleHotkey = tcell.StyleDefault.Background(color.XTerm0).Foreground(color.Orange)
