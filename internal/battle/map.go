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

func min(a, b int) int {
	if a < b {
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
	TilePavement
	TileSand
	TileSnow
	TileMarsh
	TileBush
	TileFence
	TileRubble
	TileObject
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
	TilePavement: ':',
	TileSand:     '.',
	TileSnow:     '*',
	TileMarsh:   '~',
	TileBush:     '"',
	TileFence:    '|',
	TileRubble:   ':',
	TileObject:   'o',
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
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs, TilePavement, TileSand:
		return true
	}
	return false
}

func (m *BattleMap) Opaque(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileWall, TileTree, TileRock, TileUFOWall, TileFence:
		return true
	}
	return false
}

// fillRect fills a rectangle with a tile type
func (m *BattleMap) fillRect(x, y, w, h int, t TileType) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			m.Set(x+dx, y+dy, t)
		}
	}
}

// drawRect draws a rectangle outline with a tile type
func (m *BattleMap) drawRect(x, y, w, h int, t TileType) {
	for dx := 0; dx < w; dx++ {
		m.Set(x+dx, y, t)
		m.Set(x+dx, y+h-1, t)
	}
	for dy := 0; dy < h; dy++ {
		m.Set(x, y+dy, t)
		m.Set(x+w-1, y+dy, t)
	}
}

// placeDoor places a door on a wall
func (m *BattleMap) placeDoor(x, y int) {
	m.Set(x, y, TileDoor)
}

// isPassable checks if a tile is passable
func (m *BattleMap) isPassable(x, y int) bool {
	return m.Passable(x, y)
}

// connectDoors connects two rooms with a door
func (m *BattleMap) connectDoors(x1, y1, x2, y2 int) {
	if x1 == x2 {
		start := min(y1, y2)
		end := max(y1, y2)
		for y := start; y <= end; y++ {
			if m.At(x1, y).Type == TileWall {
				m.Set(x1, y, TileDoor)
				return
			}
		}
	} else if y1 == y2 {
		start := min(x1, x2)
		end := max(x1, x2)
		for x := start; x <= end; x++ {
			if m.At(x, y1).Type == TileWall {
				m.Set(x, y1, TileDoor)
				return
			}
		}
	}
}

// generateCorridor creates an L-shaped corridor between two points
func (m *BattleMap) generateCorridor(x1, y1, x2, y2 int, w int) {
	if rand.Intn(2) == 0 {
		// Horizontal first, then vertical
		start := min(x1, x2)
		end := max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.At(x, y1+dy).Type != TileWall {
					m.Set(x, y1+dy, TileFloor)
				}
			}
		}
		start = min(y1, y2)
		end = max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.At(x2+dx, y).Type != TileWall {
					m.Set(x2+dx, y, TileFloor)
				}
			}
		}
	} else {
		// Vertical first, then horizontal
		start := min(y1, y2)
		end := max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.At(x1+dx, y).Type != TileWall {
					m.Set(x1+dx, y, TileFloor)
				}
			}
		}
		start = min(x1, x2)
		end = max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.At(x, y2+dy).Type != TileWall {
					m.Set(x, y2+dy, TileFloor)
				}
			}
		}
	}
}

// GenerateProcedural creates a map based on a terrain biome definition.
func GenerateProcedural(biomeName string, w, h int) *BattleMap {
	biome, ok := Biomes[biomeName]
	if !ok {
		return GenerateCrashSite(w, h)
	}
	m := NewBattleMap(w, h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, biome.DefaultTile)

			r := rand.Intn(100)
			cumulative := 0
			for tileType, prob := range biome.TileProbs {
				cumulative += prob
				if r < cumulative {
					m.Set(x, y, tileType)
					break
				}
			}
		}
	}
	return m
}
	m := NewBattleMap(w, h)

	// Scatter terrain based on OpenXcom forest/jungle patterns
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 3 {
				m.Set(x, y, TileTree)
			} else if r < 5 {
				m.Set(x, y, TileBush)
			} else if r < 7 {
				m.Set(x, y, TileRock)
			} else if r < 8 {
				m.Set(x, y, TileFence)
			}
		}
	}

	// UFO crash site in center (OpenXcom: 8x6 default UFO size)
	ufoX := w/2 - 4
	ufoY := h/2 - 3

	// UFO walls with irregular edges
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

	// UFO door
	m.Set(ufoX+4, ufoY+5, TileDoor)

	// Scatter some debris around crash
	for i := 0; i < 15; i++ {
		dx := rand.Intn(12) - 6
		dy := rand.Intn(10) - 5
		x := ufoX + 4 + dx
		y := ufoY + 3 + dy
		if m.At(x, y).Type == TileGrass || m.At(x, y).Type == TileTree {
			m.Set(x, y, TileRubble)
		}
	}

	return m
}

// GenerateTerrorSite creates a terror site map (OpenXcom: 50x50 urban)
func GenerateTerrorSite(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Fill with pavement (roads)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TilePavement)
		}
	}

	// Generate roads (OpenXcom urban script pattern)
	// Vertical road
	if rand.Intn(2) == 0 {
		roadX := w/4 + rand.Intn(w/2)
		for y := 0; y < h; y++ {
			m.Set(roadX-1, y, TilePavement)
			m.Set(roadX, y, TilePavement)
			m.Set(roadX+1, y, TilePavement)
		}
	}
	// Horizontal road
	if rand.Intn(2) == 0 {
		roadY := h/4 + rand.Intn(h/2)
		for x := 0; x < w; x++ {
			m.Set(x, roadY-1, TilePavement)
			m.Set(x, roadY, TilePavement)
			m.Set(x, roadY+1, TilePavement)
		}
	}

	// Generate buildings (OpenXcom urban blocks: 10x10 areas)
	buildings := 0
	maxBuildings := 12
	attempts := 0
	for buildings < maxBuildings && attempts < 200 {
		attempts++
		bw := 6 + rand.Intn(8)
		bh := 5 + rand.Intn(7)
		bx := rand.Intn(w-bw-2) + 1
		by := rand.Intn(h-bh-2) + 1

		// Check for overlap
		overlap := false
		for dy := -1; dy <= bh; dy++ {
			for dx := -1; dx <= bw; dx++ {
				if m.At(bx+dx, by+dy).Type == TileWall {
					overlap = true
					break
				}
			}
			if overlap {
				break
			}
		}
		if overlap {
			continue
		}

		// Draw building walls
		m.drawRect(bx, by, bw, bh, TileWall)

		// Fill interior with floor
		m.fillRect(bx+1, by+1, bw-2, bh-2, TileFloor)

		// Place door (usually on south wall)
		doorX := bx + 1 + rand.Intn(bw-2)
		m.Set(doorX, by+bh-1, TileDoor)

		// Place windows
		if bw > 4 {
			m.Set(bx+bw/2, by, TileWindow)
			m.Set(bx+bw/2, by+bh-1, TileWindow)
		}
		if bh > 4 {
			m.Set(bx, by+bh/2, TileWindow)
			m.Set(bx+bw-1, by+bh/2, TileWindow)
		}

		// Add some interior walls for rooms
		if bw >= 8 && bh >= 6 {
			wallX := bx + bw/2
			m.Set(wallX, by+1, TileWall)
			m.Set(wallX, by+2, TileWall)
			m.Set(wallX, by+3, TileDoor)
		}

		buildings++
	}

	// Scatter some objects (furniture, etc.)
	for i := 0; i < 20; i++ {
		x := rand.Intn(w-2) + 1
		y := rand.Intn(h-2) + 1
		if m.At(x, y).Type == TileFloor {
			m.Set(x, y, TileObject)
		}
	}

	return m
}

// GenerateUFOInterior creates a UFO interior map (OpenXcom: 50x50)
func GenerateUFOInterior(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Fill with UFO floor
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TileUFOFloor)
		}
	}

	// Outer walls
	m.drawRect(0, 0, w, h, TileUFOWall)

	// Generate rooms (OpenXcom UFO interior pattern)
	rooms := 6 + rand.Intn(4)
	roomCenters := make([][2]int, 0, rooms)

	attempts := 0
	for i := 0; i < rooms && attempts < 100; i++ {
		attempts++
		rw := 5 + rand.Intn(6)
		rh := 4 + rand.Intn(5)
		rx := 2 + rand.Intn(max(1, w-rw-4))
		ry := 2 + rand.Intn(max(1, h-rh-4))

		// Check for overlap
		overlap := false
		for dy := -1; dy <= rh; dy++ {
			for dx := -1; dx <= rw; dx++ {
				if m.At(rx+dx, ry+dy).Type == TileUFOWall {
					overlap = true
					break
				}
			}
			if overlap {
				break
			}
		}
		if overlap {
			continue
		}

		// Draw room walls
		m.drawRect(rx, ry, rw, rh, TileUFOWall)

		// Fill interior
		m.fillRect(rx+1, ry+1, rw-2, rh-2, TileUFOFloor)

		// Place door
		doorX := rx + 1 + rand.Intn(rw-2)
		m.Set(doorX, ry+rh-1, TileDoor)

		roomCenters = append(roomCenters, [2]int{rx + rw/2, ry + rh/2})
	}

	// Connect rooms with corridors
	for i := 0; i < len(roomCenters)-1; i++ {
		m.generateCorridor(
			roomCenters[i][0], roomCenters[i][1],
			roomCenters[i+1][0], roomCenters[i+1][1],
			1,
		)
	}

	// Command center in the middle
	cx := w/2 - 4
	cy := h/2 - 3
	m.drawRect(cx, cy, 8, 6, TileUFOWall)
	m.fillRect(cx+1, cy+1, 6, 4, TileUFOFloor)
	m.Set(cx+4, cy+5, TileDoor)

	// Add some objects
	for i := 0; i < 15; i++ {
		x := rand.Intn(w-4) + 2
		y := rand.Intn(h-4) + 2
		if m.At(x, y).Type == TileUFOFloor {
			m.Set(x, y, TileObject)
		}
	}

	return m
}

// GenerateCydonia creates an alien base map (OpenXcom: 50x50, 2 levels)
func GenerateCydonia(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Fill with alien floor
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TileUFOFloor)
		}
	}

	// Outer walls
	m.drawRect(0, 0, w, h, TileUFOWall)

	// Generate entrance area
	m.fillRect(0, h-6, 10, 6, TileUFOFloor)
	m.Set(4, h-1, TileDoor)

	// Generate command center (20x20 area in center)
	ccX := w/2 - 10
	ccY := h/2 - 10
	m.drawRect(ccX, ccY, 20, 20, TileUFOWall)
	m.fillRect(ccX+1, ccY+1, 18, 18, TileUFOFloor)

	// Brain in center
	m.fillRect(ccX+8, ccY+8, 4, 4, TileObject)
	m.Set(ccX+9, ccY+11, TileDoor)

	// Generate pods (5x5 rooms) around command center
	podPositions := [][2]int{
		{ccX - 8, ccY - 8},
		{ccX + 20, ccY - 8},
		{ccX - 8, ccY + 20},
		{ccX + 20, ccY + 20},
	}

	for _, pos := range podPositions {
		if pos[0] > 1 && pos[0] < w-7 && pos[1] > 1 && pos[1] < h-7 {
			m.drawRect(pos[0], pos[1], 6, 6, TileUFOWall)
			m.fillRect(pos[1]+1, pos[1]+1, 4, 4, TileUFOFloor)
			m.Set(pos[0]+3, pos[1]+5, TileDoor)
		}
	}

	// Connect command center to pods with tunnels
	for _, pos := range podPositions {
		if pos[0] > 1 && pos[0] < w-7 && pos[1] > 1 && pos[1] < h-7 {
			m.generateCorridor(
				ccX+10, ccY+10,
				pos[0]+3, pos[1]+3,
				2,
			)
		}
	}

	// Connect entrance to command center
	m.generateCorridor(5, h-6, ccX+10, ccY+20, 2)

	// Add some Alien Alloy objects
	for i := 0; i < 20; i++ {
		x := rand.Intn(w-4) + 2
		y := rand.Intn(h-4) + 2
		if m.At(x, y).Type == TileUFOFloor {
			m.Set(x, y, TileObject)
		}
	}

	return m
}

// GenerateForest creates a forest map (OpenXcom: 50x50)
func GenerateForest(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Dense forest
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 15 {
				m.Set(x, y, TileTree)
			} else if r < 20 {
				m.Set(x, y, TileBush)
			} else if r < 22 {
				m.Set(x, y, TileRock)
			}
		}
	}

	// Small clearing
	clearX := w/4 + rand.Intn(w/2)
	clearY := h/4 + rand.Intn(h/2)
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			if dx*dx+dy*dy <= 9 {
				m.Set(clearX+dx, clearY+dy, TileGrass)
			}
		}
	}

	return m
}

// GenerateDesert creates a desert map (OpenXcom: 50x50)
func GenerateDesert(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Sandy terrain
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 5 {
				m.Set(x, y, TileRock)
			} else if r < 8 {
				m.Set(x, y, TileSand)
			} else if r < 10 {
				m.Set(x, y, TileBush)
			}
		}
	}

	return m
}

// GeneratePolar creates a polar map (OpenXcom: 50x50)
func GeneratePolar(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Snowy terrain
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := rand.Intn(100)
			if r < 3 {
				m.Set(x, y, TileRock)
			} else if r < 10 {
				m.Set(x, y, TileSnow)
			} else if r < 12 {
				m.Set(x, y, TileMarsh)
			}
		}
	}

	return m
}
