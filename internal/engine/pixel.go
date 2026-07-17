package engine

import (
	"github.com/gdamore/tcell/v3"
)


const halfBlockRune = '▀'

// PixelImage represents a 2D grid of colors.
// Since each terminal cell can show two vertically stacked pixels (foreground and background of halfBlockRune),
// a PixelImage of Width x Height pixels requires Width x (Height/2) terminal cells.
type PixelImage struct {
	Width  int
	Height int
	Pixels [][]tcell.Color
}

// NewPixelImage creates a new PixelImage initialized to transparent (tcell.ColorDefault).
func NewPixelImage(w, h int) *PixelImage {
	pixels := make([][]tcell.Color, h)
	for i := range pixels {
		pixels[i] = make([]tcell.Color, w)
		for j := range pixels[i] {
			pixels[i][j] = tcell.ColorDefault
		}
	}
	return &PixelImage{
		Width:  w,
		Height: h,
		Pixels: pixels,
	}
}

func drawPixelCell(screen tcell.Screen, x, col, y, row int, img *PixelImage) (tcell.Color, tcell.Color, bool) {
	if x+col < 0 || x+col >= screenW(screen) || y+row/2 < 0 || y+row/2 >= screenH(screen) {
		return 0, 0, true
	}
	topColor := img.Pixels[row][col]
	bottomColor := ColorBlackTcell
	if row+1 < img.Height {
		bottomColor = img.Pixels[row+1][col]
	}
	if topColor == tcell.ColorDefault && bottomColor == tcell.ColorDefault {
		return 0, 0, true
	}
	resolvedTop := topColor
	if topColor == tcell.ColorDefault {
		resolvedTop = ColorBlackTcell
	}
	resolvedBottom := bottomColor
	if bottomColor == tcell.ColorDefault {
		resolvedBottom = ColorBlackTcell
	}
	return resolvedTop, resolvedBottom, false
}

func screenW(s tcell.Screen) int { w, _ := s.Size(); return w }
func screenH(s tcell.Screen) int { _, h := s.Size(); return h }

// DrawPixelImage draws the PixelImage onto a tcell.Screen.
func DrawPixelImage(screen tcell.Screen, x, y int, img *PixelImage) {
	for row := 0; row < img.Height; row += 2 {
		for col := 0; col < img.Width; col++ {
			resolvedTop, resolvedBottom, skip := drawPixelCell(screen, x, col, y, row, img)
			if skip {
				continue
			}
			style := tcell.StyleDefault.Foreground(resolvedTop).Background(resolvedBottom)
			screen.SetContent(x+col, y+row/2, halfBlockRune, nil, style)
		}
	}
}

// CompositeImages overlays the 'overlay' image onto 'base' (if overlay pixel is not transparent).
// Returns a new PixelImage of the same dimensions as base.
func CompositeImages(base, overlay *PixelImage) *PixelImage {
	w, h := base.Width, base.Height
	res := NewPixelImage(w, h)

	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			// Get base color
			baseCol := base.Pixels[r][c]
			
			// Overlay color
			overlayCol := tcell.ColorDefault
			if r < overlay.Height && c < overlay.Width {
				overlayCol = overlay.Pixels[r][c]
			}

			if overlayCol != tcell.ColorDefault {
				res.Pixels[r][c] = overlayCol
			} else {
				res.Pixels[r][c] = baseCol
			}
		}
	}
	return res
}

// DarkenColor reduces the brightness of a color by a factor (0.0 to 1.0).
// If c is tcell.ColorDefault, it returns c unmodified.
func DarkenColor(c tcell.Color, factor float64) tcell.Color {
	if c == tcell.ColorDefault {
		return c
	}
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	r, g, b := c.RGB()
	nr := int32(float64(r) * factor)
	ng := int32(float64(g) * factor)
	nb := int32(float64(b) * factor)
	return tcell.NewRGBColor(nr, ng, nb)
}

// LightenColor increases the brightness of a color towards white by a factor (e.g. 1.0 to 2.0).
// If c is tcell.ColorDefault, it returns c unmodified.
func LightenColor(c tcell.Color, factor float64) tcell.Color {
	if c == tcell.ColorDefault {
		return c
	}
	if factor < 1.0 {
		return DarkenColor(c, factor)
	}
	r, g, b := c.RGB()
	
	// Blend towards white (255)
	blend := factor - 1.0
	if blend > 1.0 {
		blend = 1.0
	}
	
	nr := int32(float64(r) + (255-float64(r))*blend)
	ng := int32(float64(g) + (255-float64(g))*blend)
	nb := int32(float64(b) + (255-float64(b))*blend)
	
	if nr > 255 { nr = 255 }
	if ng > 255 { ng = 255 }
	if nb > 255 { nb = 255 }
	
	return tcell.NewRGBColor(nr, ng, nb)
}

// DrawPixelImage helper for drawing inside the engine on ScreenRaw.
func (s *ScreenRaw) DrawPixelImage(x, y int, img *PixelImage) {
	DrawPixelImage(s.screen, x, y, img)
}

// DrawPixelImageFramed draws a PixelImage with a box-drawing frame around it.
func DrawPixelImageFramed(screen tcell.Screen, x, y int, img *PixelImage, frameStyle tcell.Style) {
	fw := img.Width + 2
	fh := img.Height/2 + 2

	drawFrameEdge(screen, x, y, fw, fh, '┌', '┐', '└', '┘', '─', '│', frameStyle)

	for row := 0; row < img.Height; row += 2 {
		cellRow := y + 1 + row/2
		for col := 0; col < img.Width; col++ {
			resolvedTop, resolvedBottom, skip := drawPixelCell(screen, x+1+col, 0, cellRow, row, img)
			if skip {
				continue
			}
			style := tcell.StyleDefault.Foreground(resolvedTop).Background(resolvedBottom)
			screen.SetContent(x+1+col, cellRow, halfBlockRune, nil, style)
		}
	}
}

func drawFrameEdge(s tcell.Screen, x, y, w, h int, tl, tr, bl, br, hRune, vRune rune, style tcell.Style) {
	s.SetContent(x, y, tl, nil, style)
	s.SetContent(x+w-1, y, tr, nil, style)
	s.SetContent(x, y+h-1, bl, nil, style)
	s.SetContent(x+w-1, y+h-1, br, nil, style)
	for i := 1; i < w-1; i++ {
		s.SetContent(x+i, y, hRune, nil, style)
		s.SetContent(x+i, y+h-1, hRune, nil, style)
	}
	for i := 1; i < h-1; i++ {
		s.SetContent(x, y+i, vRune, nil, style)
		s.SetContent(x+w-1, y+i, vRune, nil, style)
	}
}

// DrawPixelImageFramed helper for drawing inside the engine on ScreenRaw.
func (s *ScreenRaw) DrawPixelImageFramed(x, y int, img *PixelImage, frameStyle tcell.Style) {
	DrawPixelImageFramed(s.screen, x, y, img, frameStyle)
}

