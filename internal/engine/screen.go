package engine

import (
	"sync"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func RuneWidth(ch rune) int {
	return displaywidth.Rune(ch)
}

func StringWidth(str string) int {
	return displaywidth.String(str)
}

type ScreenRaw struct {
	screen tcell.Screen
	width  int
	height int
	fb     *FrameBuffer
}

func NewScreenRawWithScreen(s tcell.Screen, w, h int) *ScreenRaw {
	return &ScreenRaw{screen: s, width: w, height: h, fb: NewFrameBuffer(w, h)}
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
	s.SetStyle(StyleDefault)
	w, h := s.Size()
	return &ScreenRaw{screen: s, width: w, height: h, fb: NewFrameBuffer(w, h)}, nil
}

func (s *ScreenRaw) SetTheme(theme string) {
	ApplyTheme(theme)
	s.screen.SetStyle(StyleDefault)
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
	st := tcell.StyleDefault.Foreground(cd.Fg).Background(cd.Bg)
	if cd.Attr&tcell.AttrBold != 0 {
		st = st.Bold(true)
	}
	if cd.Attr&tcell.AttrBlink != 0 {
		st = st.Blink(true)
	}
	if cd.Attr&tcell.AttrReverse != 0 {
		st = st.Reverse(true)
	}
	if cd.Attr&tcell.AttrDim != 0 {
		st = st.Dim(true)
	}
	if cd.Attr&tcell.AttrItalic != 0 {
		st = st.Italic(true)
	}
	if cd.Attr&tcell.AttrStrikeThrough != 0 {
		st = st.StrikeThrough(true)
	}
	return st
}

func (s *ScreenRaw) Peek(x, y int) (rune, tcell.Style) {
	cd := s.fb.Get(x, y)
	return cd.Ch, styleFromCell(cd)
}

func (s *ScreenRaw) FrameBuffer() *FrameBuffer {
	return s.fb
}

func (s *ScreenRaw) DrawString(x, y int, str string, style tcell.Style) {
	currX := x
	for _, ch := range str {
		w := RuneWidth(ch)
		s.SetCell(currX, y, ch, style)
		currX += w
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
		w := RuneWidth(ch)
		s.SetCell(currX, y, ch, style)
		currX += w
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
		titleW := StringWidth(titleStr)
		s.DrawString(x+w/2-titleW/2, y, titleStr, style)
	}
}

// Define theme-agnostic placeholders if needed
var ColorBlack = color.Black
var ColorBlackTcell = color.Black

// Global styles. Written by ApplyTheme (under styleMu) on the config/options
// screen; read from the main goroutine during rendering. Not safe for
// concurrent read/write from different goroutines.
var (
	styleMu       sync.Mutex
	StyleDefault  = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm15)
	StyleHighlight = tcell.StyleDefault.Background(color.DarkBlue).Foreground(color.XTerm15)
	StyleRed      = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm9)
	StyleGreen    = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm2)
	StyleBlue     = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm12)
	StyleYellow   = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm11)
	StyleCyan     = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm6)
	StyleMagenta  = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm13)
	StyleGray     = tcell.StyleDefault.Background(ColorBlack).Foreground(color.XTerm8)
	StyleOrange   = tcell.StyleDefault.Background(ColorBlack).Foreground(color.Orange)
	StyleWater    = tcell.StyleDefault.Background(ColorBlack).Foreground(tcell.NewRGBColor(30, 100, 200))
	StyleCyanBold = StyleCyan.Bold(true)
	StyleRedBold  = StyleRed.Bold(true)
	StyleHotkey   = tcell.StyleDefault.Background(ColorBlack).Foreground(color.Orange)
)

func setThemeStyles(bg, fg, dim tcell.Color, bold bool) {
	StyleDefault = tcell.StyleDefault.Background(bg).Foreground(fg)
	StyleHighlight = tcell.StyleDefault.Background(fg).Foreground(bg)
	StyleGray = tcell.StyleDefault.Background(bg).Foreground(dim)
	StyleCyanBold = StyleCyan
	StyleRedBold = StyleRed
	StyleHotkey = tcell.StyleDefault.Background(bg).Foreground(dim)
	if bold {
		StyleDefault = StyleDefault.Bold(true)
		StyleHighlight = StyleHighlight.Bold(true)
		StyleGray = StyleGray.Bold(true)
		StyleHotkey = StyleHotkey.Bold(true)
	}
}

func ApplyTheme(theme string) {
	styleMu.Lock()
	defer styleMu.Unlock()
	switch theme {
	case "high_contrast":
		ColorBlack = color.Black
		ColorBlackTcell = color.Black
		setThemeStyles(ColorBlack, color.White, color.Silver, true)
		StyleRed = StyleDefault.Foreground(color.Red).Bold(true)
		StyleGreen = StyleDefault.Foreground(color.Lime).Bold(true)
		StyleBlue = StyleDefault.Foreground(color.Navy).Bold(true)
		StyleYellow = StyleDefault.Foreground(color.Yellow).Bold(true)
		StyleCyan = StyleDefault.Foreground(color.Aqua).Bold(true)
		StyleMagenta = StyleDefault.Foreground(color.Fuchsia).Bold(true)
		StyleOrange = StyleDefault.Foreground(color.Orange).Bold(true)
		StyleWater = StyleDefault.Foreground(tcell.NewRGBColor(0, 60, 180)).Bold(true)
		StyleCyanBold = StyleCyan
		StyleRedBold = StyleRed
		StyleHotkey = StyleDefault.Foreground(color.Orange).Bold(true)

	case "amber":
		ColorBlack = tcell.NewRGBColor(12, 8, 0)
		ColorBlackTcell = ColorBlack
		bg := ColorBlack
		fg := tcell.NewRGBColor(255, 190, 0)
		dim := tcell.NewRGBColor(120, 90, 0)
		setThemeStyles(bg, fg, dim, false)
		StyleRed = StyleDefault.Foreground(tcell.NewRGBColor(255, 80, 0))
		StyleGreen = StyleDefault.Foreground(tcell.NewRGBColor(180, 200, 0))
		StyleBlue = StyleDefault.Foreground(tcell.NewRGBColor(160, 170, 120))
		StyleYellow = StyleDefault.Foreground(fg)
		StyleCyan = StyleDefault.Foreground(tcell.NewRGBColor(200, 180, 60))
		StyleMagenta = StyleDefault.Foreground(tcell.NewRGBColor(200, 100, 60))
		StyleOrange = StyleDefault.Foreground(tcell.NewRGBColor(255, 130, 0))
		StyleWater = StyleDefault.Foreground(tcell.NewRGBColor(60, 40, 0))
		StyleCyanBold = StyleCyan.Bold(true)
		StyleRedBold = StyleRed.Bold(true)
		StyleHotkey = StyleDefault.Foreground(tcell.NewRGBColor(255, 210, 80)).Bold(true)

	case "green":
		ColorBlack = tcell.NewRGBColor(0, 12, 0)
		ColorBlackTcell = ColorBlack
		bg := ColorBlack
		fg := tcell.NewRGBColor(0, 220, 0)
		dim := tcell.NewRGBColor(0, 90, 0)
		setThemeStyles(bg, fg, dim, false)
		StyleRed = StyleDefault.Foreground(tcell.NewRGBColor(0, 255, 0))
		StyleGreen = StyleDefault.Foreground(tcell.NewRGBColor(0, 200, 80))
		StyleBlue = StyleDefault.Foreground(tcell.NewRGBColor(50, 180, 130))
		StyleYellow = StyleDefault.Foreground(tcell.NewRGBColor(160, 240, 0))
		StyleCyan = StyleDefault.Foreground(tcell.NewRGBColor(0, 200, 160))
		StyleMagenta = StyleDefault.Foreground(tcell.NewRGBColor(0, 200, 100))
		StyleOrange = StyleDefault.Foreground(tcell.NewRGBColor(100, 240, 0))
		StyleWater = StyleDefault.Foreground(tcell.NewRGBColor(0, 80, 40))
		StyleCyanBold = StyleCyan.Bold(true)
		StyleRedBold = StyleRed.Bold(true)
		StyleHotkey = StyleDefault.Foreground(tcell.NewRGBColor(100, 255, 100)).Bold(true)

	case "paper":
		ColorBlack = tcell.NewRGBColor(200, 190, 170)
		ColorBlackTcell = ColorBlack
		bg := ColorBlack
		fg := tcell.NewRGBColor(10, 10, 10)
		dim := tcell.NewRGBColor(120, 110, 100)
		setThemeStyles(bg, fg, dim, false)
		StyleRed = StyleDefault.Foreground(tcell.NewRGBColor(160, 20, 20))
		StyleGreen = StyleDefault.Foreground(tcell.NewRGBColor(30, 100, 30))
		StyleBlue = StyleDefault.Foreground(tcell.NewRGBColor(20, 60, 140))
		StyleYellow = StyleDefault.Foreground(tcell.NewRGBColor(140, 110, 0))
		StyleCyan = StyleDefault.Foreground(tcell.NewRGBColor(30, 100, 120))
		StyleMagenta = StyleDefault.Foreground(tcell.NewRGBColor(120, 40, 120))
		StyleOrange = StyleDefault.Foreground(tcell.NewRGBColor(180, 80, 20))
		StyleWater = StyleDefault.Foreground(tcell.NewRGBColor(20, 60, 100))
		StyleCyanBold = StyleCyan.Bold(true)
		StyleRedBold = StyleRed.Bold(true)
		StyleHotkey = StyleDefault.Foreground(tcell.NewRGBColor(180, 80, 20)).Bold(true)

	default:
		ColorBlack = color.XTerm0
		ColorBlackTcell = color.XTerm0
		setThemeStyles(ColorBlack, color.XTerm15, color.XTerm8, false)
		StyleHighlight = StyleDefault.Background(color.Blue).Foreground(color.XTerm15)
		StyleRed = StyleDefault.Foreground(color.XTerm9)
		StyleGreen = StyleDefault.Foreground(color.XTerm2)
		StyleBlue = StyleDefault.Foreground(color.XTerm12)
		StyleYellow = StyleDefault.Foreground(color.XTerm11)
		StyleCyan = StyleDefault.Foreground(color.XTerm6)
		StyleMagenta = StyleDefault.Foreground(color.XTerm13)
		StyleOrange = StyleDefault.Foreground(color.Orange)
		StyleWater = StyleDefault.Foreground(tcell.NewRGBColor(30, 100, 200))
		StyleCyanBold = StyleCyan.Bold(true)
		StyleRedBold = StyleRed.Bold(true)
		StyleHotkey = StyleDefault.Foreground(color.Orange)
	}
}
