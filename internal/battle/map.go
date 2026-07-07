package battle

import (
	"math/rand"
)

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
		bx := rand.Intn(w - bw - 2) + 1
		by := rand.Intn(h - bh - 2) + 1
		for x := 0; x < bw; x++ {
			m.Set(bx+x, by, TileWall)
			m.Set(bx+x, by+bh-1, TileWall)
		}
		for y := 0; y < bh; y++ {
			m.Set(bx, by+y, TileWall)
			m.Set(bx+bw-1, by+y, TileWall)
		}
		// Door
		doorX := bx + 1 + rand.Intn(bw-2)
		m.Set(doorX, by+bh-1, TileDoor)
		// Windows
		if bw > 4 {
			m.Set(bx+bw/2, by, TileWindow)
		}
	}

	return m
}
