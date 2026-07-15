package engine

import (
	"math"

	"github.com/gdamore/tcell/v3"
)

type cellData struct {
	Ch   rune
	Fg   tcell.Color
	Bg   tcell.Color
	Attr tcell.AttrMask
}

type FrameBuffer struct {
	cells []cellData
	w, h  int
}

func NewFrameBuffer(w, h int) *FrameBuffer {
	return &FrameBuffer{
		cells: make([]cellData, w*h),
		w:     w,
		h:     h,
	}
}

func (fb *FrameBuffer) Resize(w, h int) {
	if w == fb.w && h == fb.h {
		return
	}
	newCells := make([]cellData, w*h)
	n := len(fb.cells)
	if n > w*h {
		n = w*h
	}
	copy(newCells, fb.cells[:n])
	fb.cells = newCells
	fb.w = w
	fb.h = h
}

func (fb *FrameBuffer) Set(x, y int, ch rune, fg, bg tcell.Color, attr tcell.AttrMask) {
	if x >= 0 && x < fb.w && y >= 0 && y < fb.h {
		fb.cells[y*fb.w+x] = cellData{Ch: ch, Fg: fg, Bg: bg, Attr: attr}
	}
}

func (fb *FrameBuffer) Get(x, y int) cellData {
	if x >= 0 && x < fb.w && y >= 0 && y < fb.h {
		return fb.cells[y*fb.w+x]
	}
	return cellData{}
}

func (fb *FrameBuffer) Clear() {
	for i := range fb.cells {
		fb.cells[i] = cellData{}
	}
}

// Width returns the framebuffer width in cells.
func (fb *FrameBuffer) Width() int {
	return fb.w
}

// Height returns the framebuffer height in cells.
func (fb *FrameBuffer) Height() int {
	return fb.h
}

// MarshalBinary encodes the framebuffer as 8 bytes per cell:
// [rune_lo, rune_hi, fg_r, fg_g, fg_b, bg_r, bg_g, attr].
// Runes are limited to BMP (U+0000-U+FFFF) per project convention.
// Color components are 0-255. Attr is tcell.AttrMask as a byte.
func (fb *FrameBuffer) MarshalBinary() []byte {
	n := len(fb.cells)
	data := make([]byte, n*8)
	for i, cd := range fb.cells {
		off := i * 8
		r := uint16(cd.Ch)
		data[off+0] = byte(r)
		data[off+1] = byte(r >> 8)
		fr, fg, fbCol := cd.Fg.RGB()
		data[off+2] = byte(fr)
		data[off+3] = byte(fg)
		data[off+4] = byte(fbCol)
		br, bg, _ := cd.Bg.RGB()
		data[off+5] = byte(br)
		data[off+6] = byte(bg)
		data[off+7] = byte(cd.Attr)
	}
	return data
}

func colorRGB(c tcell.Color) (float64, float64, float64) {
	r, g, b := c.RGB()
	return float64(r), float64(g), float64(b)
}

func lerpColor(a, b [3]float64, t float64) [3]float64 {
	return [3]float64{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}

func smoothstep(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func ApplyLightSource(s *ScreenRaw, fb *FrameBuffer, sourceX, sourceY int, radius float64, lightColor tcell.Color) {
	scrW, scrH := s.Size()
	lR, lG, lB := colorRGB(lightColor)
	radiusInt := int(math.Ceil(radius))
	count := 0

	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			x, y := sourceX+dx, sourceY+dy
			if x < 0 || x >= scrW || y < 0 || y >= scrH {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}
			count++
		}
	}

	if count == 0 {
		return
	}

	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			x, y := sourceX+dx, sourceY+dy
			if x < 0 || x >= scrW || y < 0 || y >= scrH {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}
			falloff := smoothstep(1 - dist/radius)
			cell := fb.Get(x, y)
			bgR, bgG, bgB := colorRGB(cell.Bg)
			blended := lerpColor([3]float64{bgR, bgG, bgB}, [3]float64{lR, lG, lB}, falloff)
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			fgR, fgG, fgB := colorRGB(cell.Fg)
			fgBlend := lerpColor([3]float64{fgR, fgG, fgB}, [3]float64{lR, lG, lB}, falloff*0.5)
			newFg := tcell.NewRGBColor(int32(fgBlend[0]), int32(fgBlend[1]), int32(fgBlend[2]))
			style := tcell.StyleDefault.Foreground(newFg).Background(newBg)
			s.SetCell(x, y, cell.Ch, style)
		}
	}
}

func ApplyBloom(s *ScreenRaw, fb *FrameBuffer, centerX, centerY int, bloomColor tcell.Color) {
	radius := 1.5
	lR, lG, lB := colorRGB(bloomColor)
	radiusInt := 2
	scrW, scrH := s.Size()

	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			x, y := centerX+dx, centerY+dy
			if x < 0 || x >= scrW || y < 0 || y >= scrH {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}
			// Gentle falloff for bloom
			falloff := 0.3 * (1 - dist/radius)
			cell := fb.Get(x, y)
			bgR, bgG, bgB := colorRGB(cell.Bg)
			blended := [3]float64{
				bgR + (lR-bgR)*falloff,
				bgG + (lG-bgG)*falloff,
				bgB + (lB-bgB)*falloff,
			}
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			
			// Maintain existing foreground, just blend background
			style := tcell.StyleDefault.Foreground(cell.Fg).Background(newBg)
			s.SetCell(x, y, cell.Ch, style)
		}
	}
}

func ApplyDistortion(s *ScreenRaw, fb *FrameBuffer, timeVal float64) {
	scrW, scrH := s.Size()
	
	// Create a temporary buffer to hold the distorted image
	tmp := NewFrameBuffer(scrW, scrH)
	
	for y := 0; y < scrH; y++ {
		// Calculate distortion offset for this row
		offsetX := int(math.Sin(timeVal*0.05+float64(y)*0.1) * 2.0)
		
		for x := 0; x < scrW; x++ {
			srcX := x + offsetX
			if srcX < 0 {
				srcX = 0
			} else if srcX >= scrW {
				srcX = scrW - 1
			}
			tmp.cells[y*scrW+x] = fb.Get(srcX, y)
		}
	}
	
	// Copy back to screen
	for y := 0; y < scrH; y++ {
		for x := 0; x < scrW; x++ {
			cell := tmp.cells[y*scrW+x]
			s.SetCell(x, y, cell.Ch, tcell.StyleDefault.Foreground(cell.Fg).Background(cell.Bg))
		}
	}
}

func ApplyDirectionalLight(s *ScreenRaw, fb *FrameBuffer, sourceX, sourceY int, dirX, dirY float64, radius float64, lightColor tcell.Color, isVisible func(x, y int) bool) {
	lR, lG, lB := colorRGB(lightColor)
	radiusInt := int(math.Ceil(radius))
	scrW, scrH := s.Size()

	mag := math.Sqrt(dirX*dirX + dirY*dirY)
	if mag > 0 {
		dirX /= mag
		dirY /= mag
	}

	for dy := -radiusInt; dy <= radiusInt; dy++ {
		for dx := -radiusInt; dx <= radiusInt; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > radius {
				continue
			}
			if dist == 0 {
				continue
			}

			// Cone filter: dot product against direction vector
			if mag > 0 {
				dot := (float64(dx)*dirX + float64(dy)*dirY) / dist
				if dot < 0.7 {
					continue
				}
			}

			x, y := sourceX+dx, sourceY+dy
			if x < 0 || x >= scrW || y < 0 || y >= scrH {
				continue
			}

			// Bresenham shadow raycast — skip if any intermediate cell blocks LOS
			if isVisible != nil && !raycastClear(sourceX, sourceY, x, y, isVisible) {
				continue
			}

			falloff := smoothstep(1 - dist/radius)
			cell := fb.Get(x, y)
			bgR, bgG, bgB := colorRGB(cell.Bg)
			blended := lerpColor([3]float64{bgR, bgG, bgB}, [3]float64{lR, lG, lB}, falloff*0.4)
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			style := tcell.StyleDefault.Foreground(cell.Fg).Background(newBg)
			s.SetCell(x, y, cell.Ch, style)
		}
	}
}

// raycastClear walks a Bresenham line from (x1,y1) to (x2,y2) and returns true
// only if every intermediate cell (excluding source and destination) passes isVisible.
func raycastClear(x1, y1, x2, y2 int, isVisible func(x, y int) bool) bool {
	dx := x2 - x1
	dy := y2 - y1
	ax, ay := dx, dy
	if ax < 0 {
		ax = -ax
	}
	if ay < 0 {
		ay = -ay
	}
	sx, sy := 1, 1
	if dx < 0 {
		sx = -1
	}
	if dy < 0 {
		sy = -1
	}
	x, y := x1, y1
	var err int
	if ax >= ay {
		err = ax / 2
		for x != x2 {
			x += sx
			err -= ay
			if err < 0 {
				y += sy
				err += ax
			}
			if x == x2 && y == y2 {
				break
			}
			if !isVisible(x, y) {
				return false
			}
		}
	} else {
		err = ay / 2
		for y != y2 {
			y += sy
			err -= ax
			if err < 0 {
				x += sx
				err += ay
			}
			if x == x2 && y == y2 {
				break
			}
			if !isVisible(x, y) {
				return false
			}
		}
	}
	return true
}

func DrawPixel(s *ScreenRaw, x, y int, upperColor, lowerColor tcell.Color) {
	style := tcell.StyleDefault.Foreground(upperColor).Background(lowerColor)
	s.SetCell(x, y, '▀', style)
}

func DrawPixelSingle(s *ScreenRaw, x, y int, color tcell.Color) {
	s.SetCell(x, y, '▀', tcell.StyleDefault.Foreground(color).Background(color))
}

func DrawTransparentRect(s *ScreenRaw, fb *FrameBuffer, x, y, width, height int, overlayColor tcell.Color, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}

	oR, oG, oB := colorRGB(overlayColor)
	invAlpha := 1 - alpha
	scrW, scrH := s.Size()

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			cx, cy := x+dx, y+dy
			if cx < 0 || cx >= scrW || cy < 0 || cy >= scrH {
				continue
			}
			cell := fb.Get(cx, cy)
			bgR, bgG, bgB := colorRGB(cell.Bg)
			blended := [3]float64{
				oR*alpha + bgR*invAlpha,
				oG*alpha + bgG*invAlpha,
				oB*alpha + bgB*invAlpha,
			}
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			style := tcell.StyleDefault.Foreground(cell.Fg).Background(newBg)
			s.screen.SetContent(cx, cy, cell.Ch, nil, style)
		}
	}
}
