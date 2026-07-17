package engine

import (
	"math"
	"math/rand"

	"github.com/gdamore/tcell/v3"
)

var waterColors = [4][3]float64{
	{0, 0, 100},
	{0, 20, 140},
	{0, 40, 180},
	{10, 60, 200},
}

const (
	waveFreq    = 0.5
	colorScale  = 3
	waveChurn   = 0.3
	randWave    = 5
	fgOffR      = 40
	fgOffG      = 60
	fgOffB      = 40
)

func DrawWater(s *ScreenRaw, x, y int, gameTime float64) {
	wave := math.Sin(float64(x)*waveFreq+gameTime) * math.Cos(float64(y)*waveFreq+gameTime)
	t := (wave + 1) / 2
	idx := int(t * colorScale)
	if idx > colorScale {
		idx = colorScale
	}

	c := waterColors[idx]
	bg := tcell.NewRGBColor(int32(c[0]), int32(c[1]), int32(c[2]))

	ch := '~'
	if wave > waveChurn {
		ch = '≈'
	} else if rand.Intn(randWave) == 0 {
		ch = '≈'
	}

	fg := tcell.NewRGBColor(int32(c[0]+fgOffR), int32(c[1]+fgOffG), int32(c[2]+fgOffB))
	style := tcell.StyleDefault.Foreground(fg).Background(bg)
	s.SetCell(x, y, ch, style)
}

func DrawWaterRect(s *ScreenRaw, x, y, w, h int, gameTime float64) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			DrawWater(s, x+dx, y+dy, gameTime)
		}
	}
}
