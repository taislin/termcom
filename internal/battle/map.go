package battle

import (
	"math/rand"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type TileType int

const (
	TileFloor TileType = iota
	TileWall
	TileDoor
	TileWindow
	TileGrass
	TileTree
	TileRock
	TileWater
	TileUFOFloor
	TileUFOWall
	TileStairs
)

type Tile struct {
	Type    TileType
	Destroyed bool
}

var tileChars = map[TileType]rune{
	TileFloor:    '.',
	TileWall:     '#',
	TileDoor:     '+',
	TileWindow:   'o',
	TileGrass:    ',',
	TileTree:     'T',
	TileRock:     '%',
	TileWater:    '~',
	TileUFOFloor: '=',
	TileUFOWall:  '█',
	TileStairs:   '▓',
}

func TileChar(t TileType) rune {
	ch, ok := tileChars[t]
	if ok {
		return ch
	}
	return '.'
}

type BattleMap struct {
	Width  int
	Height int
	Tiles  [][]Tile
}

func NewBattleMap(w, h int) *BattleMap {
	m := &BattleMap{
		Width:  w,
		Height: h,
		Tiles:  make([][]Tile, h),
	}
	for y := 0; y < h; y++ {
		m.Tiles[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			m.Tiles[y][x] = Tile{Type: TileGrass}
		}
	}
	return m
}

func (m *BattleMap) At(x, y int) Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return Tile{Type: TileWall}
	}
	return m.Tiles[y][x]
}

func (m *BattleMap) Set(x, y int, t TileType) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Tiles[y][x].Type = t
	}
}

func (m *BattleMap) Passable(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs:
		return true
	}
	return false
}

func (m *BattleMap) Opaque(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileWall, TileTree, TileRock, TileUFOWall:
		return true
	}
	return false
}

func GenerateCrashSite(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Scatter terrain
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 5 {
				m.Set(x, y, TileTree)
			} else if r < 8 {
				m.Set(x, y, TileRock)
			}
		}
	}

	// UFO crash site in center
	ufoX := w/2 - 4
	ufoY := h/2 - 3
	// UFO walls
	for x := 0; x < 8; x++ {
		m.Set(ufoX+x, ufoY, TileUFOWall)
		m.Set(ufoX+x, ufoY+5, TileUFOWall)
	}
	for y := 0; y < 6; y++ {
		m.Set(ufoX, ufoY+y, TileUFOWall)
		m.Set(ufoX+7, ufoY+y, TileUFOWall)
	}
	// UFO interior
	for y := 1; y < 5; y++ {
		for x := 1; x < 7; x++ {
			m.Set(ufoX+x, ufoY+y, TileUFOFloor)
		}
	}
	// Door
	m.Set(ufoX+4, ufoY+5, TileDoor)

	return m
}

func GenerateTerrorSite(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Urban terrain
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TileFloor)
		}
	}

	// Buildings
	for i := 0; i < 8; i++ {
		bw := 4 + rand.Intn(6)
		bh := 4 + rand.Intn(6)
		bx := rand.Intn(w-bw-2) + 1
		by := rand.Intn(h-bh-2) + 1
		for x := 0; x < bw; x++ {
			m.Set(bx+x, by, TileWall)
			m.Set(bx+x, by+bh-1, TileWall)
		}
		for y := 0; y < bh; y++ {
			m.Set(bx, by+y, TileWall)
			m.Set(bx+bw-1, by+y, TileWall)
		}
		doorX := bx + 1 + rand.Intn(bw-2)
		m.Set(doorX, by+bh-1, TileDoor)
		if bw > 4 {
			m.Set(bx+bw/2, by, TileWindow)
		}
	}

	return m
}

func GenerateUFOInterior(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TileUFOFloor)
		}
	}

	for x := 0; x < w; x++ {
		m.Set(x, 0, TileUFOWall)
		m.Set(x, h-1, TileUFOWall)
	}
	for y := 0; y < h; y++ {
		m.Set(0, y, TileUFOWall)
		m.Set(w-1, y, TileUFOWall)
	}

	rooms := 3 + rand.Intn(3)
	for i := 0; i < rooms; i++ {
		rw := 4 + rand.Intn(4)
		rh := 3 + rand.Intn(3)
		rx := 2 + rand.Intn(max(1, w-rw-4))
		ry := 2 + rand.Intn(max(1, h-rh-4))
		for x := 0; x < rw; x++ {
			m.Set(rx+x, ry, TileUFOWall)
			m.Set(rx+x, ry+rh-1, TileUFOWall)
		}
		for y := 0; y < rh; y++ {
			m.Set(rx, ry+y, TileUFOWall)
			m.Set(rx+rw-1, ry+y, TileUFOWall)
		}
		doorX := rx + 1 + rand.Intn(rw-2)
		m.Set(doorX, ry+rh-1, TileDoor)
	}

	cx := w/2 - 3
	cy := h/2 - 2
	for x := 0; x < 6; x++ {
		m.Set(cx+x, cy, TileUFOWall)
		m.Set(cx+x, cy+3, TileUFOWall)
	}
	for y := 0; y < 4; y++ {
		m.Set(cx, cy+y, TileUFOWall)
		m.Set(cx+5, cy+y, TileUFOWall)
	}
	m.Set(cx+3, cy+3, TileDoor)

	return m
}

func GenerateCydonia(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 8 {
				m.Set(x, y, TileRock)
			} else if r < 12 {
				m.Set(x, y, TileWater)
			} else {
				m.Set(x, y, TileGrass)
			}
		}
	}

	baseX := w/2 - 5
	baseY := h/2 - 4
	for x := 0; x < 10; x++ {
		m.Set(baseX+x, baseY, TileUFOWall)
		m.Set(baseX+x, baseY+7, TileUFOWall)
	}
	for y := 0; y < 8; y++ {
		m.Set(baseX, baseY+y, TileUFOWall)
		m.Set(baseX+9, baseY+y, TileUFOWall)
	}
	for y := 1; y < 7; y++ {
		for x := 1; x < 9; x++ {
			m.Set(baseX+x, baseY+y, TileUFOFloor)
		}
	}
	m.Set(baseX+4, baseY+3, TileUFOWall)
	m.Set(baseX+5, baseY+3, TileUFOWall)
	m.Set(baseX+4, baseY+4, TileUFOWall)
	m.Set(baseX+5, baseY+4, TileUFOWall)
	m.Set(baseX+5, baseY+7, TileDoor)
	m.Set(baseX+4, baseY, TileDoor)

	for i := 0; i < 20; i++ {
		tx := rand.Intn(w)
		ty := rand.Intn(h)
		if m.At(tx, ty).Type == TileGrass {
			m.Set(tx, ty, TileTree)
		}
	}

	return m
}
