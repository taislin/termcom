package engine

import (
	"github.com/gdamore/tcell/v2"
)

type ScreenRaw struct {
	screen tcell.Screen
	width  int
	height int
}

func NewScreenRaw() (*ScreenRaw, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	w, h := s.Size()
	return &ScreenRaw{screen: s, width: w, height: h}, nil
}

func (s *ScreenRaw) Close() {
	s.screen.Fini()
}

func (s *ScreenRaw) Clear() {
	s.screen.Clear()
}

func (s *ScreenRaw) Flush() {
	s.screen.Show()
}

func (s *ScreenRaw) Size() (int, int) {
	return s.width, s.height
}

func (s *ScreenRaw) SetCell(x, y int, ch rune, style tcell.Style) {
	if x >= 0 && x < s.width && y >= 0 && y < s.height {
		s.screen.SetContent(x, y, ch, nil, style)
	}
}

func (s *ScreenRaw) DrawString(x, y int, str string, style tcell.Style) {
	for i, ch := range str {
		s.SetCell(x+i, y, ch, style)
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

var StyleDefault = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
var StyleHighlight = tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
var StyleRed = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)
var StyleGreen = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
var StyleBlue = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorBlue)
var StyleYellow = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow)
var StyleCyan = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorTeal)
var StyleMagenta = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorFuchsia)
var StyleGray = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGray)

var StyleCyanBold = StyleCyan.Bold(true)
var StyleRedBold = StyleRed.Bold(true)
var StyleGreenBold = StyleGreen.Bold(true)
