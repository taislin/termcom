package engine

import (
	"github.com/gdamore/tcell/v3"
)

// PixelImage represents a 2D grid of colors.
// Since each terminal cell can show two vertically stacked pixels (foreground and background of ▀),
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

// DrawPixelImage draws the PixelImage onto a tcell.Screen.
// Each cell at (x + col, y + row/2) uses '▀' (U+2580) with
// FG = top pixel (row) and BG = bottom pixel (row+1).
// If the height is odd, the bottom pixel of the last cell row defaults to tcell.ColorBlack.
func DrawPixelImage(screen tcell.Screen, x, y int, img *PixelImage) {
	for row := 0; row < img.Height; row += 2 {
		for col := 0; col < img.Width; col++ {
			topColor := img.Pixels[row][col]
			bottomColor := tcell.ColorBlack
			if row+1 < img.Height {
				bottomColor = img.Pixels[row+1][col]
			}

			// If both are default/transparent, we skip drawing to support transparent images.
			if topColor == tcell.ColorDefault && bottomColor == tcell.ColorDefault {
				continue
			}

			// If one of them is transparent, we handle it by blending or drawing against a black fallback.
			resolvedTop := topColor
			if topColor == tcell.ColorDefault {
				resolvedTop = tcell.ColorBlack
			}
			resolvedBottom := bottomColor
			if bottomColor == tcell.ColorDefault {
				resolvedBottom = tcell.ColorBlack
			}

			style := tcell.StyleDefault.Foreground(resolvedTop).Background(resolvedBottom)
			screen.SetContent(x+col, y+row/2, '▀', nil, style)
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

