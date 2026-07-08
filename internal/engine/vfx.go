package engine

import (
	"math"
	"sync"

	"github.com/gdamore/tcell/v3"
)

type cellData struct {
	ch   rune
	fg   tcell.Color
	bg   tcell.Color
	attr tcell.AttrMask
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
		fb.cells[y*fb.w+x] = cellData{ch: ch, fg: fg, bg: bg, attr: attr}
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

var vfxPool = sync.Pool{
	New: func() interface{} {
		return &vfxBuffer{}
	},
}

type vfxBuffer struct {
	rgbs [][3]float64
}

func getVfxBuffer(size int) *vfxBuffer {
	buf := vfxPool.Get().(*vfxBuffer)
	if cap(buf.rgbs) < size {
		buf.rgbs = make([][3]float64, size)
	} else {
		buf.rgbs = buf.rgbs[:size]
	}
	return buf
}

func putVfxBuffer(buf *vfxBuffer) {
	vfxPool.Put(buf)
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

	buf := getVfxBuffer(count)
	defer putVfxBuffer(buf)

	idx := 0
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
			bgR, bgG, bgB := colorRGB(cell.bg)
			blended := lerpColor([3]float64{bgR, bgG, bgB}, [3]float64{lR, lG, lB}, falloff)
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			fgR, fgG, fgB := colorRGB(cell.fg)
			fgBlend := lerpColor([3]float64{fgR, fgG, fgB}, [3]float64{lR, lG, lB}, falloff*0.5)
			newFg := tcell.NewRGBColor(int32(fgBlend[0]), int32(fgBlend[1]), int32(fgBlend[2]))
			style := tcell.StyleDefault.Foreground(newFg).Background(newBg)
			s.SetCell(x, y, cell.ch, style)
			idx++
		}
	}
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
			bgR, bgG, bgB := colorRGB(cell.bg)
			blended := [3]float64{
				oR*alpha + bgR*invAlpha,
				oG*alpha + bgG*invAlpha,
				oB*alpha + bgB*invAlpha,
			}
			newBg := tcell.NewRGBColor(int32(blended[0]), int32(blended[1]), int32(blended[2]))
			style := tcell.StyleDefault.Foreground(cell.fg).Background(newBg)
			s.screen.SetContent(cx, cy, cell.ch, nil, style)
		}
	}
}
