package battle

import (
	"sync"

	"github.com/civ13/termcom/internal/engine"
	"github.com/gdamore/tcell/v3"
)

type GasType int

const (
	GasSmoke  GasType = iota
	GasPoison
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
	if density > 3 {
		density = 3
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
	return ok && c.Density >= 3
}

func (g *GasGrid) CoverPenalty(x, y int) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	c, ok := g.cells[[2]int{x, y}]
	if !ok {
		return 0
	}
	switch c.Density {
	case 3:
		return 40
	case 2:
		return 20
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
		if v.Density <= 1 {
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
	defer g.mu.RUnlock()

	for k, v := range g.cells {
		mx, my := k[0], k[1]
		sx := mx - scrollX + 1
		sy := my - scrollY + 1
		if sx < 1 || sx > viewW || sy < 1 || sy > viewH {
			continue
		}
		if g.Visible != nil && !g.Visible(mx, my) {
			continue
		}

		var ch rune
		var style tcell.Style

		switch v.Density {
		case 3:
			ch = '\u2593'
			if v.Type == GasPoison {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 200, 0)).Background(tcell.NewRGBColor(0, 40, 0))
			} else {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(160, 160, 160)).Background(tcell.NewRGBColor(40, 40, 40))
			}
		case 2:
			ch = '\u2592'
			if v.Type == GasPoison {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 160, 0)).Background(tcell.NewRGBColor(0, 25, 0))
			} else {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(128, 128, 128)).Background(tcell.NewRGBColor(25, 25, 25))
			}
		case 1:
			ch = '\u2591'
			if v.Type == GasPoison {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(0, 100, 0))
			} else {
				style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(80, 80, 80))
			}
		default:
			continue
		}

		ctx.SetCell(sx, sy, ch, style)
	}
}

func (g *GasGrid) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cells = make(map[[2]int]GasCell)
}
