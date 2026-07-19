package battle

import (
	"sync"

	"github.com/taislin/termcom/internal/engine"
	"github.com/gdamore/tcell/v3"
)

type GasType int

const (
	GasSmoke  GasType = iota
	GasPoison
	GasFreeze
)

// Gas density bounds and tuning.
const (
	MaxGasDensity      = 3    // density is clamped to this ceiling
	MinDiffuseDensity  = 1    // cells at or below this density stop spreading
	GasCoverDensity3   = 40   // cover penalty (%) at max density
	GasCoverDensity2   = 20   // cover penalty (%) at medium density
)

// Gas palette colours per density level (smoke foreground/background, poison fg/bg).
var (
	gasSmokeFg = []tcell.Color{
		tcell.NewRGBColor(80, 80, 80),
		tcell.NewRGBColor(128, 128, 128),
		tcell.NewRGBColor(160, 160, 160),
	}
	gasSmokeBg = []tcell.Color{
		tcell.NewRGBColor(0, 0, 0),
		tcell.NewRGBColor(25, 25, 25),
		tcell.NewRGBColor(40, 40, 40),
	}
	gasPoisonFg = []tcell.Color{
		tcell.NewRGBColor(0, 100, 0),
		tcell.NewRGBColor(0, 160, 0),
		tcell.NewRGBColor(0, 200, 0),
	}
	gasPoisonBg = []tcell.Color{
		tcell.NewRGBColor(0, 0, 0),
		tcell.NewRGBColor(0, 25, 0),
		tcell.NewRGBColor(0, 40, 0),
	}
	gasFreezeFg = []tcell.Color{
		tcell.NewRGBColor(150, 200, 230),
		tcell.NewRGBColor(190, 225, 245),
		tcell.NewRGBColor(225, 245, 255),
	}
	gasFreezeBg = []tcell.Color{
		tcell.NewRGBColor(0, 10, 20),
		tcell.NewRGBColor(10, 25, 40),
		tcell.NewRGBColor(20, 40, 60),
	}
	gasRune = []rune{'\u2591', '\u2592', '\u2593'} // light/medium/dark shade by density-1
)

type GasCell struct {
	Density int
	Type    GasType
}

type GasGrid struct {
	mu      sync.RWMutex
	cells   map[[2]int]GasCell
	width   int
	height  int
	Visible func(x, y int) bool
}

func NewGasGrid(w, h int) *GasGrid {
	return &GasGrid{
		cells:  make(map[[2]int]GasCell),
		width:  w,
		height: h,
	}
}

func (g *GasGrid) Set(x, y, density int, gt GasType) {
	if x < 0 || x >= g.width || y < 0 || y >= g.height {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if density <= 0 {
		delete(g.cells, [2]int{x, y})
		return
	}
	if density > MaxGasDensity {
		density = MaxGasDensity
	}
	g.cells[[2]int{x, y}] = GasCell{Density: density, Type: gt}
}

func (g *GasGrid) Get(x, y int) (int, GasType, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, ok := g.cells[[2]int{x, y}]
	if !ok {
		return 0, GasSmoke, false
	}
	return c.Density, c.Type, true
}

func (g *GasGrid) BlocksLOS(x, y int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, ok := g.cells[[2]int{x, y}]
	return ok && c.Density >= MaxGasDensity
}

func (g *GasGrid) CoverPenalty(x, y int) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, ok := g.cells[[2]int{x, y}]
	if !ok {
		return 0
	}
	switch c.Density {
	case MaxGasDensity:
		return GasCoverDensity3
	case MaxGasDensity - 1:
		return GasCoverDensity2
	default:
		return 0
	}
}

func (g *GasGrid) Diffuse() {
	g.mu.Lock()
	defer g.mu.Unlock()

	type spread struct {
		x, y int
		d    int
		t    GasType
	}
	var spreads []spread

	for k, v := range g.cells {
		if v.Density <= MinDiffuseDensity {
			continue
		}
		newD := v.Density - 1
		dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
		for _, d := range dirs {
			nx, ny := k[0]+d[0], k[1]+d[1]
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}
			spreads = append(spreads, spread{nx, ny, newD, v.Type})
		}
	}

	for _, s := range spreads {
		key := [2]int{s.x, s.y}
		existing, ok := g.cells[key]
		if !ok || s.d > existing.Density {
			g.cells[key] = GasCell{Density: s.d, Type: s.t}
		}
	}

	for k, v := range g.cells {
		v.Density--
		if v.Density <= 0 {
			delete(g.cells, k)
		} else {
			g.cells[k] = v
		}
	}
}

func (g *GasGrid) Draw(ctx *engine.ScreenCtx, scrollX, scrollY, viewW, viewH int) {
	g.mu.RLock()
	cellCopy := make(map[[2]int]GasCell, len(g.cells))
	for k, v := range g.cells {
		cellCopy[k] = v
	}
	g.mu.RUnlock()

	for k, v := range cellCopy {
		mx, my := k[0], k[1]
		sx := mx - scrollX + 1
		sy := my - scrollY + 1
		if sx < 1 || sx > viewW || sy < 1 || sy > viewH {
			continue
		}
		g.mu.RLock()
		vis := g.Visible
		g.mu.RUnlock()
		if vis != nil && !vis(mx, my) {
			continue
		}

		var ch rune
		var style tcell.Style

		ch, style = gasStyle(v.Density, v.Type)
		if ch == 0 {
			continue
		}

		ctx.SetCell(sx, sy, ch, style)
	}
}

// gasStyle returns the display rune and colour style for a gas cell of the
// given density and type. Returns ch=0 when density is out of range.
func gasStyle(density int, gt GasType) (rune, tcell.Style) {
	if density < 1 || density > MaxGasDensity {
		return 0, tcell.StyleDefault
	}
	idx := density - 1
	ch := gasRune[idx]
	if gt == GasPoison {
		return ch, tcell.StyleDefault.Foreground(gasPoisonFg[idx]).Background(gasPoisonBg[idx])
	}
	if gt == GasFreeze {
		return ch, tcell.StyleDefault.Foreground(gasFreezeFg[idx]).Background(gasFreezeBg[idx])
	}
	return ch, tcell.StyleDefault.Foreground(gasSmokeFg[idx]).Background(gasSmokeBg[idx])
}

func (g *GasGrid) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cells = make(map[[2]int]GasCell)
}
