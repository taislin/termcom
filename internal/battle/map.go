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
	// UFO furniture tiles
	TileConsole     // Control panels, navigation consoles
	TileMachinery   // Engines, generators, equipment
	TilePod         // Alien pods, containment units
	TilePowerSource // Power source, fuel cells
	TileStorage     // Storage containers, crates
	TileAlienTech   // Alien technology, artifacts
)

type Tile struct {
	Type      TileType
	Cover     int  // 0-100, damage reduction % from shots passing through
	Destroyed bool
	Visible   bool
	Seen      bool
	Blood     int // 0=none, 1=human(red), 2=alien_green, 3=alien_purple
	Fire      int // 0=none, >0=turns of fire remaining
}

// TileCover returns the base cover value for a tile type.
func TileCover(t TileType) int {
	switch t {
	case TileWall, TileUFOWall:
		return 80
	case TileTree:
		return 60
	case TileRock:
		return 70
	case TileBush:
		return 40
	case TileFence:
		return 30
	case TileObject, TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech:
		return 50
	case TileRubble:
		return 20
	case TileDoor:
		return 0
	default:
		return 0
	}
}

func (t Tile) IsFlammable() bool {
	switch t.Type {
	case TileGrass, TileTree, TileBush, TileFence, TileDoor:
		return true
	}
	return false
}

var bloodRunes = [4]rune{0, ',', '%', ':'}

func (m *BattleMap) SpawnBlood(x, y, bloodType int) {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return
	}
	tile := &m.Tiles[y][x]
	if tile.Type != TileFloor && tile.Type != TileGrass && tile.Type != TilePavement &&
		tile.Type != TileSand && tile.Type != TileUFOFloor && tile.Type != TileSnow {
		return
	}
	if tile.Blood == 0 {
		tile.Blood = bloodType
	}
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range dirs {
		nx, ny := x+d[0], y+d[1]
		if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.Height {
			continue
		}
		nt := &m.Tiles[ny][nx]
		if nt.Type == TileFloor || nt.Type == TileGrass || nt.Type == TilePavement ||
			nt.Type == TileSand || nt.Type == TileUFOFloor || nt.Type == TileSnow {
			if nt.Blood == 0 && rand.Intn(3) == 0 {
				nt.Blood = bloodType
			}
		}
	}
}

func (m *BattleMap) SpreadFire() {
	type fireSpread struct {
		x, y int
	}
	var newFires []fireSpread

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tile := &m.Tiles[y][x]
			if tile.Fire <= 0 {
				continue
			}
			tile.Fire--
			if tile.Fire <= 0 {
				tile.Type = TileFloor
				tile.Cover = TileCover(TileFloor)
				tile.Fire = 0
				continue
			}
			if rand.Intn(100) < 20 {
				dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
				for _, d := range dirs {
					nx, ny := x+d[0], y+d[1]
					if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.Height {
						continue
					}
					nt := &m.Tiles[ny][nx]
					if nt.Fire <= 0 && nt.IsFlammable() {
						newFires = append(newFires, fireSpread{nx, ny})
						break
					}
				}
			}
		}
	}

	for _, f := range newFires {
		tile := &m.Tiles[f.y][f.x]
		if tile.Fire <= 0 && tile.IsFlammable() {
			tile.Type = TileFloor
			tile.Cover = TileCover(TileFloor)
			tile.Fire = 3
		}
	}
}

var tileChars = map[TileType]rune{
	TileFloor:      '.',
	TileWall:       '#',
	TileDoor:       '+',
	TileWindow:     '¤',
	TileGrass:      '·',
	TileTree:       '♣',
	TileRock:       '∩',
	TileWater:      '≈',
	TileUFOFloor:   '≡',
	TileUFOWall:    '█',
	TileStairs:     '▓',
	TilePavement:   '░',
	TileSand:       '·',
	TileSnow:       '∗',
	TileMarsh:     '≋',
	TileBush:       '†',
	TileFence:      '║',
	TileRubble:     '▒',
	TileObject:     '■',
	// UFO furniture characters
	TileConsole:     '░',  // Console panel
	TileMachinery:   '⚙',  // Machinery (U+2699 GEAR - BMP symbol)
	TilePod:         '◈',  // Alien pod
	TilePowerSource: '⌁',  // Power source (U+2301 ELECTRICAL ARC)
	TileStorage:     '▤',  // Storage container
	TileAlienTech:   '⊕',  // Alien technology
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
	Gas    *GasGrid
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
			m.Tiles[y][x] = Tile{Type: TileGrass, Cover: TileCover(TileGrass)}
		}
	}
	return m
}

func (m *BattleMap) At(x, y int) Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return Tile{Type: TileWall, Cover: TileCover(TileWall)}
	}
	return m.Tiles[y][x]
}

func (m *BattleMap) Set(x, y int, t TileType) {
	if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
		m.Tiles[y][x].Type = t
		m.Tiles[y][x].Cover = TileCover(t)
	}
}

func (m *BattleMap) Passable(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs, TilePavement, TileSand,
		TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech:
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
	if m.Gas != nil && m.Gas.BlocksLOS(x, y) {
		return true
	}
	return false
}

func (m *BattleMap) IsDestructible(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileWall, TileUFOWall, TileTree, TileRock, TileFence, TileDoor:
		return true
	}
	return false
}

func (m *BattleMap) DestroyWall(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	tile := &m.Tiles[y][x]
	if !m.IsDestructible(x, y) {
		return false
	}
	tile.Type = TileRubble
	tile.Cover = TileCover(TileRubble)
	tile.Destroyed = true
	return true
}

// CoverAlongLine returns the maximum cover value (%) of tiles between (x1,y1)
// and (x2,y2), exclusive of the endpoints. Uses Bresenham's line.
func (m *BattleMap) CoverAlongLine(x1, y1, x2, y2 int) int {
	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	x, y := x1, y1
	maxCover := 0
	for {
		if x == x2 && y == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
		if x == x2 && y == y2 {
			break
		}
		if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
			c := m.Tiles[y][x].Cover
			if c > maxCover {
				maxCover = c
			}
		}
	}
	return maxCover
}

const SightRange = 20

func (m *BattleMap) ClearVisibility() {
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			m.Tiles[y][x].Visible = false
		}
	}
}

func (m *BattleMap) ComputeFOV(ux, uy int, sightRange int) {
	for dy := -sightRange; dy <= sightRange; dy++ {
		for dx := -sightRange; dx <= sightRange; dx++ {
			tx := ux + dx
			ty := uy + dy
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.Height {
				continue
			}
			if dx*dx+dy*dy > sightRange*sightRange {
				continue
			}
			if m.hasLOS(ux, uy, tx, ty) {
				m.Tiles[ty][tx].Visible = true
				m.Tiles[ty][tx].Seen = true
			}
		}
	}
}

func (m *BattleMap) hasLOS(x1, y1, x2, y2 int) bool {
	dx := x2 - x1
	dy := y2 - y1
	absDx := dx
	absDy := dy
	if absDx < 0 {
		absDx = -absDx
	}
	if absDy < 0 {
		absDy = -absDy
	}
	sx := 1
	if dx < 0 {
		sx = -1
	}
	sy := 1
	if dy < 0 {
		sy = -1
	}
	err := absDx - absDy
	x := x1
	y := y1
	for {
		if x == x2 && y == y2 {
			return true
		}
		if m.Opaque(x, y) && !(x == x1 && y == y1) {
			return false
		}
		e2 := 2 * err
		if e2 > -absDy {
			err -= absDy
			x += sx
		}
		if e2 < absDx {
			err += absDx
			y += sy
		}
	}
}

func (m *BattleMap) IsVisible(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	return m.Tiles[y][x].Visible
}

func (m *BattleMap) IsSeen(x, y int) bool {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}
	return m.Tiles[y][x].Seen
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

type MapCommandType int

const (
	CmdFillRect MapCommandType = iota
	CmdDrawRect
	CmdScatter
	CmdPlaceBuilding
	CmdCorridor
	CmdClearArea
)

type MapCommand struct {
	Type     MapCommandType
	X, Y     int
	W, H     int
	Tile     TileType
	Prob     int     // for Scatter: probability 0-100
	Count    int     // for Scatter: number of attempts
	X2, Y2   int     // for Corridor: endpoint
	DoorSide int     // for PlaceBuilding: 0=south, 1=east, 2=north, 3=west
}

func (m *BattleMap) ApplyCommand(cmd MapCommand) {
	switch cmd.Type {
	case CmdFillRect:
		m.fillRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Tile)
	case CmdDrawRect:
		m.drawRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Tile)
	case CmdScatter:
		for i := 0; i < cmd.Count; i++ {
			x := cmd.X + rand.Intn(cmd.W)
			y := cmd.Y + rand.Intn(cmd.H)
			if rand.Intn(100) < cmd.Prob {
				m.Set(x, y, cmd.Tile)
			}
		}
	case CmdPlaceBuilding:
		m.placeBuilding(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.DoorSide)
	case CmdCorridor:
		m.generateCorridor(cmd.X, cmd.Y, cmd.X2, cmd.Y2, max(1, cmd.W))
	case CmdClearArea:
		m.fillRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Tile)
	}
}

func (m *BattleMap) placeBuilding(bx, by, bw, bh, doorSide int) {
	m.drawRect(bx, by, bw, bh, TileWall)
	m.fillRect(bx+1, by+1, bw-2, bh-2, TileFloor)

	switch doorSide {
	case 0:
		m.Set(bx+1+rand.Intn(max(1, bw-2)), by+bh-1, TileDoor)
	case 1:
		m.Set(bx+bw-1, by+1+rand.Intn(max(1, bh-2)), TileDoor)
	case 2:
		m.Set(bx+1+rand.Intn(max(1, bw-2)), by, TileDoor)
	case 3:
		m.Set(bx, by+1+rand.Intn(max(1, bh-2)), TileDoor)
	}
}

func ApplyCommands(m *BattleMap, cmds []MapCommand) {
	for _, cmd := range cmds {
		m.ApplyCommand(cmd)
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

// GenerateCrashSite creates a crash site map (OpenXcom: 50x50)
func GenerateCrashSite(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileTree, Prob: 3, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileBush, Prob: 2, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileRock, Prob: 2, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileFence, Prob: 1, Count: w * h},
	})

	ufoX := w/2 - 4
	ufoY := h/2 - 3

	ApplyCommands(m, []MapCommand{
		{Type: CmdDrawRect, X: ufoX, Y: ufoY, W: 8, H: 6, Tile: TileUFOWall},
		{Type: CmdFillRect, X: ufoX + 1, Y: ufoY + 1, W: 6, H: 4, Tile: TileUFOFloor},
	})

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
	m.fillRect(0, 0, w, h, TilePavement)

	// Generate roads (OpenXcom urban script pattern)
	if rand.Intn(2) == 0 {
		roadX := w/4 + rand.Intn(w/2)
		m.fillRect(roadX-1, 0, 3, h, TilePavement)
	}
	if rand.Intn(2) == 0 {
		roadY := h/4 + rand.Intn(h/2)
		m.fillRect(0, roadY-1, w, 3, TilePavement)
	}

	// Generate buildings
	buildings := 0
	maxBuildings := 12
	attempts := 0
	for buildings < maxBuildings && attempts < 200 {
		attempts++
		bw := 6 + rand.Intn(8)
		bh := 5 + rand.Intn(7)
		bx := rand.Intn(w-bw-2) + 1
		by := rand.Intn(h-bh-2) + 1

		// Check for overlap (simple check)
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

		m.ApplyCommand(MapCommand{Type: CmdPlaceBuilding, X: bx, Y: by, W: bw, H: bh, DoorSide: 0})
		buildings++
	}

	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 1, Y: 1, W: w - 2, H: h - 2, Tile: TileObject, Prob: 10, Count: 20},
	})

	return m
}

// GenerateUFOInterior creates a UFO interior map (OpenXcom: 50x50)
func GenerateUFOInterior(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Fill with UFO floor
	m.fillRect(0, 0, w, h, TileUFOFloor)

	// Outer hull walls
	m.drawRect(0, 0, w, h, TileUFOWall)

	// Generate rooms
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

		m.ApplyCommand(MapCommand{Type: CmdPlaceBuilding, X: rx, Y: ry, W: rw, H: rh, DoorSide: 0})
		roomCenters = append(roomCenters, [2]int{rx + rw/2, ry + rh/2})
	}

	// Connect rooms with corridors
	for i := 0; i < len(roomCenters)-1; i++ {
		m.ApplyCommand(MapCommand{
			Type: CmdCorridor,
			X:    roomCenters[i][0], Y: roomCenters[i][1],
			X2:   roomCenters[i+1][0], Y2: roomCenters[i+1][1],
			W: 1,
		})
	}

	// Command center in the middle
	cx := w/2 - 4
	cy := h/2 - 3
	m.ApplyCommand(MapCommand{Type: CmdPlaceBuilding, X: cx, Y: cy, W: 8, H: 6, DoorSide: 0})

	// Add furniture and machinery to rooms
	for _, room := range roomCenters {
		rx, ry := room[0], room[1]
		roomType := rand.Intn(4)
		switch roomType {
		case 0: // Control room - consoles along walls
			for dx := -2; dx <= 2; dx++ {
				if m.At(rx+dx, ry-1).Type == TileUFOFloor {
					m.Set(rx+dx, ry-1, TileConsole)
				}
				if m.At(rx+dx, ry+1).Type == TileUFOFloor {
					m.Set(rx+dx, ry+1, TileConsole)
				}
			}
		case 1: // Engine room - machinery in center
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					if m.At(rx+dx, ry+dy).Type == TileUFOFloor {
						m.Set(rx+dx, ry+dy, TileMachinery)
					}
				}
			}
		case 2: // Pod room - alien pods
			for dx := -2; dx <= 2; dx += 2 {
				if m.At(rx+dx, ry).Type == TileUFOFloor {
					m.Set(rx+dx, ry, TilePod)
				}
			}
		case 3: // Storage room - crates
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 0; dy++ {
					if m.At(rx+dx, ry+dy).Type == TileUFOFloor {
						m.Set(rx+dx, ry+dy, TileStorage)
					}
				}
			}
		}
	}

	// Add power source in command center
	if m.At(cx+4, cy+3).Type == TileUFOFloor {
		m.Set(cx+4, cy+3, TilePowerSource)
	}

	// Add alien tech scattered around
	for i := 0; i < 8; i++ {
		x := rand.Intn(w-4) + 2
		y := rand.Intn(h-4) + 2
		if m.At(x, y).Type == TileUFOFloor {
			m.Set(x, y, TileAlienTech)
		}
	}

	// Add some regular objects
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 2, Y: 2, W: w - 4, H: h - 4, Tile: TileObject, Prob: 10, Count: 10},
	})

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
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileTree, Prob: 15, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileBush, Prob: 5, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileRock, Prob: 2, Count: w * h},
	})
	clearX := w/4 + rand.Intn(w/2)
	clearY := h/4 + rand.Intn(h/2)
	ApplyCommands(m, []MapCommand{
		{Type: CmdClearArea, X: clearX - 3, Y: clearY - 3, W: 7, H: 7, Tile: TileGrass},
	})
	return m
}

// GenerateDesert creates a desert map (OpenXcom: 50x50)
func GenerateDesert(w, h int) *BattleMap {
	m := NewBattleMap(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TilePavement)
		}
	}
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileRock, Prob: 5, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileSand, Prob: 3, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileBush, Prob: 2, Count: w * h},
	})
	return m
}

// GeneratePolar creates a polar map (OpenXcom: 50x50)
func GeneratePolar(w, h int) *BattleMap {
	m := NewBattleMap(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m.Set(x, y, TileSnow)
		}
	}
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileRock, Prob: 3, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileMarsh, Prob: 2, Count: w * h},
	})
	return m
}
