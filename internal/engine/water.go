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

func DrawWater(s *ScreenRaw, x, y int, gameTime float64) {
	wave := math.Sin(float64(x)*0.5+gameTime) * math.Cos(float64(y)*0.5+gameTime)
	t := (wave + 1) / 2
	idx := int(t * 3)
	if idx > 3 {
		idx = 3
	}

	c := waterColors[idx]
	bg := tcell.NewRGBColor(int32(c[0]), int32(c[1]), int32(c[2]))

	ch := '~'
	if wave > 0.3 {
		ch = '≈'
	} else if rand.Intn(5) == 0 {
		ch = '≈'
	}

	fg := tcell.NewRGBColor(int32(c[0]+40), int32(c[1]+60), int32(c[2]+40))
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
