package engine

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

/*
func TestDrawPixelImageOddHeight(t *testing.T) {
	img := NewPixelImage(3, 3) // Odd height
	img.Pixels[0][0] = color.Red
	img.Pixels[1][0] = color.Blue
	img.Pixels[2][0] = color.Green // Last pixel, odd row

	scr := newMockScreen()
	
	DrawPixelImage(scr, 0, 0, img)

	// Check top-half cell (row 0 & 1)
	cell1, ok := scr.cells[[2]int{0, 0}]
	if !ok {
		t.Fatalf("Expected cell at 0,0 to be drawn")
	}
	if cell1.Ch != '▀' {
		t.Errorf("Expected '▀', got %c", cell1.Ch)
	}
	fg := cell1.style.GetForeground()
	bg := cell1.style.GetBackground()
	if fg != color.Red || bg != color.Blue {
		t.Errorf("Expected FG=Red, BG=Blue, got FG=%v, BG=%v", fg, bg)
	}

	// Check bottom-half cell (row 2, bottom padded to Black)
	cell2, ok := scr.cells[[2]int{0, 1}]
	if !ok {
		t.Fatalf("Expected cell at 0,1 to be drawn")
	}
	if cell2.Ch != '▀' {
		t.Errorf("Expected '▀', got %c", cell2.Ch)
	}
	fg2 := cell2.style.GetForeground()
	bg2 := cell2.style.GetBackground()
	if fg2 != color.Green || bg2 != color.Black {
		t.Errorf("Expected FG=Green, BG=Black (padded), got FG=%v, BG=%v", fg2, bg2)
	}
}
*/

func TestCompositeImages(t *testing.T) {
	base := NewPixelImage(2, 2)
	base.Pixels[0][0] = color.Red
	base.Pixels[1][1] = color.Blue

	overlay := NewPixelImage(2, 2)
	overlay.Pixels[0][0] = color.Green
	// overlay.Pixels[1][1] is ColorDefault (transparent)

	res := CompositeImages(base, overlay)

	// Check (0,0) was overwritten by overlay
	if res.Pixels[0][0] != color.Green {
		t.Errorf("Expected (0,0) to be Green, got %v", res.Pixels[0][0])
	}
	// Check (1,1) kept base color because overlay was transparent
	if res.Pixels[1][1] != color.Blue {
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
