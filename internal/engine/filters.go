package engine

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

// filterRand is a seeded RNG so night-vision noise is non-deterministic across runs.
var filterRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type VisionMode int

const (
	VisionNormal VisionMode = iota
	VisionNight
	VisionThermal
)

func luminance(r, g, b float64) float64 {
	return 0.299*r + 0.587*g + 0.114*b
}

func ApplyNightVision(s *ScreenRaw) {
	scrW, scrH := s.Size()
	fb := s.fb

	for y := 0; y < scrH; y++ {
		for x := 0; x < scrW; x++ {
			cell := fb.Get(x, y)
			if cell.Ch == 0 {
				continue
			}

			fgR, fgG, fgB := colorRGB(cell.Fg)
			lum := luminance(fgR, fgG, fgB)

			var newFg tcell.Color
			if filterRand.Intn(100) < 5 {
				dim := lum * 0.3
				newFg = tcell.NewRGBColor(0, int32(dim), 0)
			} else {
				green := lum * 1.2
				if green > 255 {
					green = 255
				}
				if lum > 128 {
					newFg = tcell.NewRGBColor(0, int32(green), int32(green*0.15))
				} else if lum > 40 {
					newFg = tcell.NewRGBColor(0, int32(green), 0)
				} else {
					newFg = tcell.NewRGBColor(0, int32(green*0.6), 0)
				}
			}

			style := tcell.StyleDefault.Foreground(newFg).Background(DarkenColor(StyleDefault.GetBackground(), 0.25))
			s.SetCell(x, y, cell.Ch, style)
		}
	}
}

type ThermalEntity struct {
	X, Y int
}

func ApplyThermalVision(s *ScreenRaw, entities []ThermalEntity) {
	scrW, scrH := s.Size()
	fb := s.fb

	entityMap := make(map[[2]int]bool, len(entities))
	for _, e := range entities {
		entityMap[[2]int{e.X, e.Y}] = true
	}

	for y := 0; y < scrH; y++ {
		for x := 0; x < scrW; x++ {
			cell := fb.Get(x, y)
			if cell.Ch == 0 {
				continue
			}

			isEntity := entityMap[[2]int{x, y}]

			var newFg, newBg tcell.Color
			if isEntity {
				fgR, fgG, fgB := colorRGB(cell.Fg)
				lum := luminance(fgR, fgG, fgB)
				if lum > 128 {
					newFg = color.XTerm11
					newBg = tcell.NewRGBColor(60, 40, 0)
				} else if lum > 40 {
					newFg = color.Orange
					newBg = tcell.NewRGBColor(40, 20, 0)
				} else {
					newFg = color.XTerm9
					newBg = tcell.NewRGBColor(30, 5, 0)
				}
			} else {
				fgR, fgG, fgB := colorRGB(cell.Fg)
				lum := luminance(fgR, fgG, fgB)
				cold := lum * 0.15
				newFg = tcell.NewRGBColor(int32(cold*0.3), int32(cold*0.4), int32(cold))
				newBg = tcell.NewRGBColor(0, 0, int32(cold*0.5))
			}

			style := tcell.StyleDefault.Foreground(newFg).Background(newBg)
			s.SetCell(x, y, cell.Ch, style)
		}
	}
}

func ApplyVisionFilter(s *ScreenRaw, mode VisionMode, entities []ThermalEntity) {
	switch mode {
	case VisionNight:
		ApplyNightVision(s)
	case VisionThermal:
		ApplyThermalVision(s, entities)
	}
}
