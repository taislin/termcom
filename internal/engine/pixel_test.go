package engine

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// mockScreen implements a minimal tcell.Screen for test verification of SetContent.
type mockScreen struct {
	tcell.Screen
	cells map[[2]int]struct {
		ch    rune
		style tcell.Style
	}
}

func newMockScreen() *mockScreen {
	return &mockScreen{
		cells: make(map[[2]int]struct {
			ch    rune
			style tcell.Style
		}),
	}
}

func (m *mockScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	m.cells[[2]int{x, y}] = struct {
		ch    rune
		style tcell.Style
	}{ch: primary, style: style}
}

func TestDrawPixelImageOddHeight(t *testing.T) {
	img := NewPixelImage(3, 3) // Odd height
	img.Pixels[0][0] = tcell.ColorRed
	img.Pixels[1][0] = tcell.ColorBlue
	img.Pixels[2][0] = tcell.ColorGreen // Last pixel, odd row

	scr := newMockScreen()
	DrawPixelImage(scr, 0, 0, img)

	// Check top-half cell (row 0 & 1)
	cell1, ok := scr.cells[[2]int{0, 0}]
	if !ok {
		t.Fatalf("Expected cell at 0,0 to be drawn")
	}
	if cell1.ch != '▀' {
		t.Errorf("Expected '▀', got %c", cell1.ch)
	}
	fg := cell1.style.GetForeground()
	bg := cell1.style.GetBackground()
	if fg != tcell.ColorRed || bg != tcell.ColorBlue {
		t.Errorf("Expected FG=Red, BG=Blue, got FG=%v, BG=%v", fg, bg)
	}

	// Check bottom-half cell (row 2, bottom padded to Black)
	cell2, ok := scr.cells[[2]int{0, 1}]
	if !ok {
		t.Fatalf("Expected cell at 0,1 to be drawn")
	}
	if cell2.ch != '▀' {
		t.Errorf("Expected '▀', got %c", cell2.ch)
	}
	fg2 := cell2.style.GetForeground()
	bg2 := cell2.style.GetBackground()
	if fg2 != tcell.ColorGreen || bg2 != tcell.ColorBlack {
		t.Errorf("Expected FG=Green, BG=Black (padded), got FG=%v, BG=%v", fg2, bg2)
	}
}

func TestCompositeImages(t *testing.T) {
	base := NewPixelImage(2, 2)
	base.Pixels[0][0] = tcell.ColorRed
	base.Pixels[1][1] = tcell.ColorBlue

	overlay := NewPixelImage(2, 2)
	overlay.Pixels[0][0] = tcell.ColorGreen
	// overlay.Pixels[1][1] is ColorDefault (transparent)

	res := CompositeImages(base, overlay)

	// Check (0,0) was overwritten by overlay
	if res.Pixels[0][0] != tcell.ColorGreen {
		t.Errorf("Expected (0,0) to be Green, got %v", res.Pixels[0][0])
	}
	// Check (1,1) kept base color because overlay was transparent
	if res.Pixels[1][1] != tcell.ColorBlue {
		t.Errorf("Expected (1,1) to be Blue, got %v", res.Pixels[1][1])
	}
}

func TestDarkenColor(t *testing.T) {
	c := tcell.NewRGBColor(100, 200, 50)
	darkened := DarkenColor(c, 0.5)

	r, g, b := darkened.RGB()
	if r != 50 || g != 100 || b != 25 {
		t.Errorf("Expected RGB(50, 100, 25), got RGB(%d, %d, %d)", r, g, b)
	}

	// Test boundary clamping
	darkenedZero := DarkenColor(c, -0.5)
	r, g, b = darkenedZero.RGB()
	if r != 0 || g != 0 || b != 0 {
		t.Errorf("Expected black (0,0,0) for negative factor, got RGB(%d, %d, %d)", r, g, b)
	}
}

func TestDarkenColor_Transparent(t *testing.T) {
	c := tcell.ColorDefault
	res := DarkenColor(c, 0.5)
	if res != tcell.ColorDefault {
		t.Errorf("Expected transparent ColorDefault to pass through unchanged, got %v", res)
	}
}
