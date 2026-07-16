package battle

import (
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
)

// randn returns rand.Intn(n) but safely yields 0 when n <= 0, avoiding a panic
// on degenerate (very small) map dimensions.
func randn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// TileType defines the physical nature of a map cell, determining its
// appearance, movement cost, and defensive value.
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
	TileStairsDown  // stairs leading to lower level
)

// Tile represents a single cell on the tactical map.
type Tile struct {
	Type      TileType
	Level     int // which level this tile is on (0=ground, 1=upper)
	Cover     int // 0-100, damage reduction % from shots passing through
	Destroyed bool
	Visible   bool
	Seen      bool
	Blood     int // 0=none, 1=human(red), 2=alien_green, 3=alien_purple
	Fire      int // 0=none, >0=turns of fire remaining
	BaseColor tcell.Color
	Rune      rune
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
	TileStairsDown: '▓',
	TileStairs:     '▒',
	TilePavement:   '░',
	TileSand:       '·',
	TileSnow:       '∗',
	TileMarsh:      '≋',
	TileBush:       '†',
	TileFence:      '║',
	TileRubble:     '▒',
	TileObject:     '■',
	// UFO furniture characters
	TileConsole:     '░', // Console panel
	TileMachinery:   '⚙', // Machinery (U+2699 GEAR - BMP symbol)
	TilePod:         '◈', // Alien pod
	TilePowerSource: '⌁', // Power source (U+2301 ELECTRICAL ARC)
	TileStorage:     '▤', // Storage container
	TileAlienTech:   '⊕', // Alien technology
}

func TileChar(t TileType) rune {
	ch, ok := tileChars[t]
	if ok {
		return ch
	}
	return '.'
}

// BattleMap represents the tactical grid for a combat mission,
// managing tile data, visibility, and multi-level geometry.
type BattleMap struct {
	Width        int
	Height       int
	NumLevels    int // 1 for most maps, 2 for UFO interiors
	LevelHeight  int // height per level
	CurrentLevel int // 0=ground, 1=upper
	Tiles        [][]Tile
	Gas          *GasGrid
}

func NewBattleMap(w, h int) *BattleMap {
	m := &BattleMap{
		Width:        w,
		Height:       h,
		NumLevels:    1,
		LevelHeight:  h,
		CurrentLevel: 0,
		Tiles:        make([][]Tile, h),
	}
	for y := 0; y < h; y++ {
		m.Tiles[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			m.Tiles[y][x] = Tile{Type: TileGrass, Cover: TileCover(TileGrass), Level: 0}
		}
	}
	return m
}

func NewMultiLevelBattleMap(w, levelH, numLevels int) *BattleMap {
	totalH := levelH * numLevels
	m := &BattleMap{
		Width:        w,
		Height:       totalH,
		NumLevels:    numLevels,
		LevelHeight:  levelH,
		CurrentLevel: 0,
		Tiles:        make([][]Tile, totalH),
	}
	for y := 0; y < totalH; y++ {
		level := y / levelH
		m.Tiles[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			m.Tiles[y][x] = Tile{Type: TileGrass, Cover: TileCover(TileGrass), Level: level}
		}
	}
	return m
}

// tileAt returns the raw tile without level filtering.
func (m *BattleMap) tileAt(x, y int) Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return Tile{Type: TileWall, Cover: TileCover(TileWall), Level: m.CurrentLevel}
	}
	return m.Tiles[y][x]
}

// At returns the tile at (x,y) on the current level.
func (m *BattleMap) At(x, y int) Tile {
	if m.NumLevels <= 1 {
		return m.tileAt(x, y)
	}
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return Tile{Type: TileWall, Cover: TileCover(TileWall), Level: m.CurrentLevel}
	}
	// Convert current-level y to array y
	arrayY := y + m.CurrentLevel*m.LevelHeight
	if arrayY < 0 || arrayY >= m.Height {
		return Tile{Type: TileWall, Cover: TileCover(TileWall), Level: m.CurrentLevel}
	}
	tile := m.Tiles[arrayY][x]
	if tile.Level != m.CurrentLevel {
		return Tile{Type: TileWall, Cover: TileCover(TileWall), Level: m.CurrentLevel}
	}
	return tile
}

// Set sets the tile type at (x,y) on the current level.
func (m *BattleMap) Set(x, y int, t TileType) {
	if m.NumLevels <= 1 {
		if x >= 0 && x < m.Width && y >= 0 && y < m.Height {
			m.Tiles[y][x].Type = t
			m.Tiles[y][x].Cover = TileCover(t)
		}
		return
	}
	arrayY := y + m.CurrentLevel*m.LevelHeight
	if x >= 0 && x < m.Width && arrayY >= 0 && arrayY < m.Height {
		m.Tiles[arrayY][x].Type = t
		m.Tiles[arrayY][x].Cover = TileCover(t)
	}
}

// SetLevel sets a tile at (x,y) on a specific level.
func (m *BattleMap) SetLevel(x, y, level int, t TileType) {
	arrayY := y + level*m.LevelHeight
	if x >= 0 && x < m.Width && arrayY >= 0 && arrayY < m.Height {
		m.Tiles[arrayY][x].Type = t
		m.Tiles[arrayY][x].Cover = TileCover(t)
		m.Tiles[arrayY][x].Level = level
	}
}

// AtLevel returns the tile at (x,y) on a specific level.
func (m *BattleMap) AtLevel(x, y, level int) Tile {
	if m.NumLevels <= 1 {
		return m.tileAt(x, y)
	}
	arrayY := y + level*m.LevelHeight
	if x < 0 || x >= m.Width || arrayY < 0 || arrayY >= m.Height {
		return Tile{Type: TileWall, Cover: TileCover(TileWall), Level: level}
	}
	return m.Tiles[arrayY][x]
}

func (m *BattleMap) Passable(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs, TileStairsDown, TilePavement, TileSand, TileSnow,
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
	if !m.IsDestructible(x, y) {
		return false
	}
	if m.NumLevels <= 1 {
		if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
			return false
		}
		tile := &m.Tiles[y][x]
		tile.Type = TileRubble
		tile.Cover = TileCover(TileRubble)
	} else {
		arrayY := y + m.CurrentLevel*m.LevelHeight
		if x < 0 || x >= m.Width || arrayY < 0 || arrayY >= m.Height {
			return false
		}
		tile := &m.Tiles[arrayY][x]
		tile.Type = TileRubble
		tile.Cover = TileCover(TileRubble)
	}
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
		t := m.At(x, y)
		c := t.Cover
		if c > maxCover {
			maxCover = c
		}
	}
	return maxCover
}

const SightRange = 20

func (m *BattleMap) ClearVisibility() {
	startY := m.CurrentLevel * m.LevelHeight
	endY := startY + m.LevelHeight
	if m.NumLevels <= 1 {
		startY = 0
		endY = m.Height
	}
	for y := startY; y < endY; y++ {
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
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.LevelHeight {
				continue
			}
			if dx*dx+dy*dy > sightRange*sightRange {
				continue
			}
			if m.hasLOS(ux, uy, tx, ty) {
				if m.NumLevels <= 1 {
					m.Tiles[ty][tx].Visible = true
					m.Tiles[ty][tx].Seen = true
				} else {
					arrayY := ty + m.CurrentLevel*m.LevelHeight
					if arrayY >= 0 && arrayY < m.Height {
						m.Tiles[arrayY][tx].Visible = true
						m.Tiles[arrayY][tx].Seen = true
					}
				}
			}
		}
	}
}

// hasLOS determines if there is a clear line-of-sight between two coordinates.
// It uses Bresenham's line algorithm to traverse the map and check for opaque tiles.
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

// IsVisible returns true if the tile at (x, y) is currently within a unit's line-of-sight.
func (m *BattleMap) IsVisible(x, y int) bool {
	t := m.At(x, y)
	return t.Visible
}

// IsSeen returns true if the tile at (x, y) has ever been visited or seen by the player.
func (m *BattleMap) IsSeen(x, y int) bool {
	t := m.At(x, y)
	return t.Seen
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

func (m *BattleMap) fillRectLevel(x, y, w, h, level int, t TileType) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			m.SetLevel(x+dx, y+dy, level, t)
		}
	}
}

func (m *BattleMap) drawRectLevel(x, y, w, h, level int, t TileType) {
	for dx := 0; dx < w; dx++ {
		m.SetLevel(x+dx, y, level, t)
		m.SetLevel(x+dx, y+h-1, level, t)
	}
	for dy := 0; dy < h; dy++ {
		m.SetLevel(x, y+dy, level, t)
		m.SetLevel(x+w-1, y+dy, level, t)
	}
}

func (m *BattleMap) corridorLevel(x1, y1, x2, y2, w, level int) {
	if rand.Intn(2) == 0 {
		start := min(x1, x2)
		end := max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.AtLevel(x, y1+dy, level).Type != TileUFOWall {
					m.SetLevel(x, y1+dy, level, TileUFOFloor)
				}
			}
		}
		start = min(y1, y2)
		end = max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.AtLevel(x2+dx, y, level).Type != TileUFOWall {
					m.SetLevel(x2+dx, y, level, TileUFOFloor)
				}
			}
		}
	} else {
		start := min(y1, y2)
		end := max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.AtLevel(x1+dx, y, level).Type != TileUFOWall {
					m.SetLevel(x1+dx, y, level, TileUFOFloor)
				}
			}
		}
		start = min(x1, x2)
		end = max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.AtLevel(x, y2+dy, level).Type != TileUFOWall {
					m.SetLevel(x, y2+dy, level, TileUFOFloor)
				}
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
	Prob     int // for Scatter: probability 0-100
	Count    int // for Scatter: number of attempts
	X2, Y2   int // for Corridor: endpoint
	DoorSide int // for PlaceBuilding: 0=south, 1=east, 2=north, 3=west
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

// Biome defines a terrain generation preset.
type Biome struct {
	DefaultTile TileType
	TileProbs   map[TileType]int // percentage chances that sum to ~100
}

// Biomes is the registry of named biomes for procedural map generation.
var Biomes = map[string]*Biome{
	"forest": {
		DefaultTile: TileGrass,
		TileProbs: map[TileType]int{
			TileTree:  35,
			TileBush:  15,
			TileRock:  5,
			TileFence: 2,
		},
	},
	"urban": {
		DefaultTile: TilePavement,
		TileProbs: map[TileType]int{
			TileWall:    20,
			TileRubble:  5,
			TileFence:   3,
		},
	},
	"desert": {
		DefaultTile: TileSand,
		TileProbs: map[TileType]int{
			TileRock: 12,
			TileBush: 3,
		},
	},
	"snow": {
		DefaultTile: TileSnow,
		TileProbs: map[TileType]int{
			TileRock: 8,
			TileTree: 5,
		},
	},
	"marsh": {
		DefaultTile: TileMarsh,
		TileProbs: map[TileType]int{
			TileWater: 15,
			TileBush:  10,
			TileTree:  5,
		},
	},
}

// GenerateProcedural creates a map based on a terrain biome definition.
func GenerateProcedural(biomeName string, w, h int) *BattleMap {
	biome, ok := Biomes[biomeName]
	if !ok {
		m, _ := GenerateCrashSite(w, h, 42)
		return m
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

// GenerateCrashSite creates a crash site map with a procedural UFO blueprint
// stamped onto the terrain. Returns both the map and crash result.
func GenerateCrashSite(w, h int, seed int64) (*BattleMap, *CrashResult) {
	m := NewBattleMap(w, h)

	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileTree, Prob: 3, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileBush, Prob: 2, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileRock, Prob: 2, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileFence, Prob: 1, Count: w * h},
	})

	// Pick a UFO tier based on seed (deterministic)
	seed16 := seed
	if seed16 == 0 {
		seed16 = 42
	}
	tiers := []data.UFOTier{
		data.TierDrone, data.TierScout, data.TierInterceptor,
		data.TierBomber, data.TierCarrier,
	}
	tier := tiers[seed16%int64(len(tiers))]

	bp := data.GenerateProceduralUFO(seed, tier)

	// Center the UFO on the map
	ufoX := w/2 - bp.Width/2
	ufoY := h/2 - bp.Height/2
	crashSeverity := 0.1 + float64(seed16%80)/100.0 // 0.1–0.9

	result := StampVehicleOnMap(bp, ufoX, ufoY, m, crashSeverity)

	// Add a door on the bottom edge
	doorX := ufoX + bp.Width/2
	doorY := ufoY + bp.Height - 1
	if doorX >= 0 && doorX < w && doorY >= 0 && doorY < h {
		m.Set(doorX, doorY, TileDoor)
	}

	// Scatter debris around crash site
	for i := 0; i < 15; i++ {
		dx := rand.Intn(12) - 6
		dy := rand.Intn(10) - 5
		x := ufoX + bp.Width/2 + dx
		y := ufoY + bp.Height/2 + dy
		if m.At(x, y).Type == TileGrass || m.At(x, y).Type == TileTree {
			m.Set(x, y, TileRubble)
		}
	}

	return m, &result
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
		bx := randn(w-bw-2) + 1
		by := randn(h-bh-2) + 1

		// Ensure building doesn't overlap existing structures or map boundaries
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

func GenerateAbductionSite(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Fill with grass
	m.fillRect(0, 0, w, h, TileGrass)

	// Scatter rocks
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 1, Y: 1, W: w - 2, H: h - 2, Tile: TileRock, Prob: 5, Count: 30},
	})

	// Scatter trees
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 1, Y: 1, W: w - 2, H: h - 2, Tile: TileTree, Prob: 8, Count: 40},
	})

	// A few small structures (rural buildings)
	buildings := 3 + rand.Intn(3)
	attempts := 0
	for i := 0; i < buildings && attempts < 100; i++ {
		attempts++
		bw := 4 + rand.Intn(4)
		bh := 3 + rand.Intn(3)
		bx := randn(w-bw-2) + 1
		by := randn(h-bh-2) + 1
		overlap := false
		for dy := -1; dy <= bh; dy++ {
			for dx := -1; dx <= bw; dx++ {
				if m.At(bx+dx, by+dy).Type != TileGrass {
					overlap = true
				}
			}
		}
		if overlap {
			continue
		}
		m.ApplyCommand(MapCommand{Type: CmdPlaceBuilding, X: bx, Y: by, W: bw, H: bh, DoorSide: rand.Intn(4)})
	}

	// Scatter objects inside buildings
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 1, Y: 1, W: w - 2, H: h - 2, Tile: TileObject, Prob: 3, Count: 15},
	})

	return m
}

// GenerateUFOInterior creates a UFO interior map (OpenXcom: 50x50)
func GenerateUFOInterior(w, h int) *BattleMap {
	levelH := h / 2
	if levelH < 12 {
		levelH = 12
	}
	m := NewMultiLevelBattleMap(w, levelH, 2)

	type room struct {
		x, y, w, h int
	}

	generateLevel := func(level int) []room {
		m.fillRectLevel(0, 0, w, levelH, level, TileUFOFloor)
		m.drawRectLevel(0, 0, w, levelH, level, TileUFOWall)

		var rooms []room
		attempts := 0
		numRooms := 5 + rand.Intn(3)
		for i := 0; i < numRooms && attempts < 100; i++ {
			attempts++
			rw := 5 + rand.Intn(5)
			rh := 4 + rand.Intn(4)
			rx := 2 + rand.Intn(max(1, w-rw-4))
			ry := 2 + rand.Intn(max(1, levelH-rh-4))

			overlap := false
			for _, existing := range rooms {
				if rx < existing.x+existing.w+1 && rx+rw+1 > existing.x &&
					ry < existing.y+existing.h+1 && ry+rh+1 > existing.y {
					overlap = true
					break
				}
			}
			if overlap {
				continue
			}

			m.fillRectLevel(rx+1, ry+1, rw-1, rh-1, level, TileUFOFloor)
			m.drawRectLevel(rx, ry, rw, rh, level, TileUFOWall)

			doorSide := rand.Intn(4)
			switch doorSide {
			case 0:
				m.SetLevel(rx+rw/2, ry, level, TileDoor)
			case 1:
				m.SetLevel(rx+rw-1, ry+rh/2, level, TileDoor)
			case 2:
				m.SetLevel(rx+rw/2, ry+rh-1, level, TileDoor)
			case 3:
				m.SetLevel(rx, ry+rh/2, level, TileDoor)
			}

			rooms = append(rooms, room{rx, ry, rw, rh})
		}

		for i := 0; i < len(rooms)-1; i++ {
			cx1 := rooms[i].x + rooms[i].w/2
			cy1 := rooms[i].y + rooms[i].h/2
			cx2 := rooms[i+1].x + rooms[i+1].w/2
			cy2 := rooms[i+1].y + rooms[i+1].h/2
			m.corridorLevel(cx1, cy1, cx2, cy2, 1, level)
		}

		return rooms
	}

	level0Rooms := generateLevel(0)
	level1Rooms := generateLevel(1)

	stairsX := w / 2
	stairsY := levelH / 2

	m.SetLevel(stairsX, stairsY, 0, TileStairsDown)
	m.SetLevel(stairsX+1, stairsY, 0, TileUFOFloor)
	m.SetLevel(stairsX, stairsY+1, 0, TileUFOFloor)
	m.SetLevel(stairsX+1, stairsY+1, 0, TileUFOFloor)

	m.SetLevel(stairsX, stairsY, 1, TileStairs)
	m.SetLevel(stairsX+1, stairsY, 1, TileUFOFloor)
	m.SetLevel(stairsX, stairsY+1, 1, TileUFOFloor)
	m.SetLevel(stairsX+1, stairsY+1, 1, TileUFOFloor)

	furnishRoom := func(rooms []room, level int) {
		for _, rm := range rooms {
			rx := rm.x + rm.w/2
			ry := rm.y + rm.h/2
			roomType := rand.Intn(5)
			switch roomType {
			case 0:
				for dx := -2; dx <= 2; dx++ {
					if m.AtLevel(rx+dx, ry-1, level).Type == TileUFOFloor {
						m.SetLevel(rx+dx, ry-1, level, TileConsole)
					}
					if m.AtLevel(rx+dx, ry+1, level).Type == TileUFOFloor {
						m.SetLevel(rx+dx, ry+1, level, TileConsole)
					}
				}
			case 1:
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if m.AtLevel(rx+dx, ry+dy, level).Type == TileUFOFloor {
							m.SetLevel(rx+dx, ry+dy, level, TileMachinery)
						}
					}
				}
			case 2:
				for dx := -2; dx <= 2; dx += 2 {
					if m.AtLevel(rx+dx, ry, level).Type == TileUFOFloor {
						m.SetLevel(rx+dx, ry, level, TilePod)
					}
				}
			case 3:
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 0; dy++ {
						if m.AtLevel(rx+dx, ry+dy, level).Type == TileUFOFloor {
							m.SetLevel(rx+dx, ry+dy, level, TileStorage)
						}
					}
				}
			case 4:
				if m.AtLevel(rx, ry, level).Type == TileUFOFloor {
					m.SetLevel(rx, ry, level, TileAlienTech)
				}
			}
		}
	}

	furnishRoom(level0Rooms, 0)
	furnishRoom(level1Rooms, 1)

	m.SetLevel(stairsX+2, stairsY+2, 0, TilePowerSource)

	for i := 0; i < 6; i++ {
		x := randn(w-4) + 2
		y := randn(levelH-4) + 2
		if m.AtLevel(x, y, 0).Type == TileUFOFloor {
			m.SetLevel(x, y, 0, TileAlienTech)
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

	// Generate entrance area on a random side
	entranceSide := rand.Intn(4)
	switch entranceSide {
	case 0: // top
		m.fillRect(w/2-5, 0, 10, 6, TileUFOFloor)
		m.Set(w/2, 0, TileDoor)
	case 1: // bottom
		m.fillRect(w/2-5, h-6, 10, 6, TileUFOFloor)
		m.Set(w/2, h-1, TileDoor)
	case 2: // left
		m.fillRect(0, h/2-5, 6, 10, TileUFOFloor)
		m.Set(0, h/2, TileDoor)
	case 3: // right
		m.fillRect(w-6, h/2-5, 6, 10, TileUFOFloor)
		m.Set(w-1, h/2, TileDoor)
	}

	// Command center (size varies 16-24)
	ccSize := 16 + rand.Intn(9)
	ccX := w/2 - ccSize/2
	ccY := h/2 - ccSize/2
	m.drawRect(ccX, ccY, ccSize, ccSize, TileUFOWall)
	m.fillRect(ccX+1, ccY+1, ccSize-2, ccSize-2, TileUFOFloor)

	// Brain in center (size varies 3-5)
	brainSize := 3 + rand.Intn(3)
	brainX := ccX + ccSize/2 - brainSize/2
	brainY := ccY + ccSize/2 - brainSize/2
	m.fillRect(brainX, brainY, brainSize, brainSize, TileObject)
	brainDoorSide := rand.Intn(4)
	switch brainDoorSide {
	case 0:
		m.Set(brainX+brainSize/2, brainY-1, TileDoor)
	case 1:
		m.Set(brainX+brainSize/2, brainY+brainSize, TileDoor)
	case 2:
		m.Set(brainX-1, brainY+brainSize/2, TileDoor)
	case 3:
		m.Set(brainX+brainSize, brainY+brainSize/2, TileDoor)
	}

	// Pods: 3-6 rooms around command center
	podCount := 3 + rand.Intn(4)
	podSize := 5 + rand.Intn(2) // 5-6
	var podPositions [][2]int
	for i := 0; i < podCount; i++ {
		angle := float64(i) * (2 * 3.14159 / float64(podCount))
		radius := float64(ccSize/2+podSize+2)
		px := w/2 + int(angle*radius*0.7) - podSize/2
		py := h/2 + int(angle*radius*0.5) - podSize/2
		// Clamp to map bounds
		px = max(2, min(px, w-podSize-2))
		py = max(2, min(py, h-podSize-2))
		m.drawRect(px, py, podSize, podSize, TileUFOWall)
		m.fillRect(px+1, py+1, podSize-2, podSize-2, TileUFOFloor)
		// Door on a random side
		doorSide := rand.Intn(4)
		switch doorSide {
		case 0:
			m.Set(px+podSize/2, py-1, TileDoor)
		case 1:
			m.Set(px+podSize/2, py+podSize, TileDoor)
		case 2:
			m.Set(px-1, py+podSize/2, TileDoor)
		case 3:
			m.Set(px+podSize, py+podSize/2, TileDoor)
		}
		// Random furniture
		furniture := rand.Intn(4)
		switch furniture {
		case 0:
			m.Set(px+podSize/2, py+podSize/2, TilePod)
		case 1:
			m.Set(px+podSize/2, py+podSize/2, TileConsole)
		case 2:
			m.Set(px+podSize/2, py+podSize/2, TileMachinery)
		case 3:
			m.Set(px+podSize/2, py+podSize/2, TilePowerSource)
		}
		podPositions = append(podPositions, [2]int{px + podSize/2, py + podSize/2})
	}

	// Connect command center to pods
	for _, pos := range podPositions {
		m.generateCorridor(ccX+ccSize/2, ccY+ccSize/2, pos[0], pos[1], 2)
	}

	// Connect entrance to command center
	switch entranceSide {
	case 0:
		m.generateCorridor(w/2, 6, ccX+ccSize/2, ccY, 2)
	case 1:
		m.generateCorridor(w/2, h-6, ccX+ccSize/2, ccY+ccSize, 2)
	case 2:
		m.generateCorridor(6, h/2, ccX, ccY+ccSize/2, 2)
	case 3:
		m.generateCorridor(w-6, h/2, ccX+ccSize, ccY+ccSize/2, 2)
	}

	// Scatter objects
	scatterCount := 15 + rand.Intn(15)
	for i := 0; i < scatterCount; i++ {
		x := randn(w-4) + 2
		y := randn(h-4) + 2
		if m.At(x, y).Type == TileUFOFloor {
			m.Set(x, y, TileObject)
		}
	}

	return m
}

// GenerateAlienBase creates an alien base assault map: rocky terrain with a
// central alien structure (distinct from the final Cydonia mission).
func GenerateAlienBase(w, h int) *BattleMap {
	m := NewBattleMap(w, h)

	// Rocky outdoor terrain
	m.fillRect(0, 0, w, h, TileRock)
	ApplyCommands(m, []MapCommand{
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileSand, Prob: 6, Count: w * h},
		{Type: CmdScatter, X: 0, Y: 0, W: w, H: h, Tile: TileBush, Prob: 2, Count: w * h},
	})

	// Vary base structure size (18-28)
	baseSize := 18 + rand.Intn(11)
	bx := w/2 - baseSize/2
	by := h/2 - baseSize/2
	m.drawRect(bx, by, baseSize, baseSize, TileUFOWall)
	m.fillRect(bx+1, by+1, baseSize-2, baseSize-2, TileUFOFloor)

	// Command core (size varies 4-8)
	coreSize := 4 + rand.Intn(5)
	cx := w/2 - coreSize/2
	cy := h/2 - coreSize/2
	m.drawRect(cx, cy, coreSize, coreSize, TileUFOWall)
	m.fillRect(cx+1, cy+1, coreSize-2, coreSize-2, TileUFOFloor)
	m.Set(cx+coreSize/2, cy+coreSize-1, TileDoor)
	coreItems := 1 + rand.Intn(3)
	for i := 0; i < coreItems; i++ {
		ix := cx + 1 + rand.Intn(coreSize-2)
		iy := cy + 1 + rand.Intn(coreSize-2)
		if m.At(ix, iy).Type == TileUFOFloor {
			m.Set(ix, iy, TileAlienTech)
		}
	}

	// Side pods: 2-6 pods placed around the perimeter
	podCount := 2 + rand.Intn(5)
	podSize := 4 + rand.Intn(3) // 4-6
	var pods [][2]int
	margin := podSize + 1
	for i := 0; i < podCount; i++ {
		// Place pods at varying positions around the structure interior
		var px, py int
		switch i % 4 {
		case 0: // top-left area
			px = bx + margin + rand.Intn(max(1, baseSize/2-margin*2))
			py = by + margin + rand.Intn(max(1, baseSize/2-margin*2))
		case 1: // top-right area
			px = bx + baseSize/2 + rand.Intn(max(1, baseSize/2-margin*2))
			py = by + margin + rand.Intn(max(1, baseSize/2-margin*2))
		case 2: // bottom-left area
			px = bx + margin + rand.Intn(max(1, baseSize/2-margin*2))
			py = by + baseSize/2 + rand.Intn(max(1, baseSize/2-margin*2))
		case 3: // bottom-right area
			px = bx + baseSize/2 + rand.Intn(max(1, baseSize/2-margin*2))
			py = by + baseSize/2 + rand.Intn(max(1, baseSize/2-margin*2))
		}
		// Ensure pod fits within the base
		if px+podSize < bx+baseSize-1 && py+podSize < by+baseSize-1 {
			m.drawRect(px, py, podSize, podSize, TileUFOWall)
			m.fillRect(px+1, py+1, podSize-2, podSize-2, TileUFOFloor)
			m.Set(px+podSize/2, py+podSize-1, TileDoor)
			// Random furniture in pod
			furniture := rand.Intn(3)
			switch furniture {
			case 0:
				m.Set(px+podSize/2, py+podSize/2, TilePod)
			case 1:
				m.Set(px+podSize/2, py+podSize/2, TileAlienTech)
			case 2:
				m.Set(px+podSize/2, py+podSize/2, TileMachinery)
			}
			pods = append(pods, [2]int{px + podSize/2, py + podSize/2})
		}
	}

	// Connect core to pods
	coreCx := cx + coreSize/2
	coreCy := cy + coreSize/2
	for _, p := range pods {
		m.generateCorridor(coreCx, coreCy, p[0], p[1], 2)
	}

	// Entrance on a random side
	entranceSide := rand.Intn(4)
	switch entranceSide {
	case 0: // top
		ex := bx + 2 + rand.Intn(max(1, baseSize-4))
		m.fillRect(ex-1, by-2, 3, 3, TileUFOFloor)
		m.Set(ex, by-1, TileDoor)
	case 1: // bottom
		ex := bx + 2 + rand.Intn(max(1, baseSize-4))
		m.fillRect(ex-1, by+baseSize-1, 3, 3, TileUFOFloor)
		m.Set(ex, by+baseSize, TileDoor)
	case 2: // left
		ey := by + 2 + rand.Intn(max(1, baseSize-4))
		m.fillRect(bx-2, ey-1, 3, 3, TileUFOFloor)
		m.Set(bx-1, ey, TileDoor)
	case 3: // right
		ey := by + 2 + rand.Intn(max(1, baseSize-4))
		m.fillRect(bx+baseSize-1, ey-1, 3, 3, TileUFOFloor)
		m.Set(bx+baseSize, ey, TileDoor)
	}

	// Scatter alien tech loot inside the structure
	scatterCount := 8 + rand.Intn(12)
	for i := 0; i < scatterCount; i++ {
		x := bx + 2 + rand.Intn(max(1, baseSize-4))
		y := by + 2 + rand.Intn(max(1, baseSize-4))
		if m.At(x, y).Type == TileUFOFloor {
			m.Set(x, y, TileAlienTech)
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

func (m *BattleMap) neighbourhood(x, y int) [3][3]TileType {
	var res [3][3]TileType
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
				res[dy+1][dx+1] = TileGrass
			} else {
				res[dy+1][dx+1] = m.At(nx, ny).Type
			}
		}
	}
	return res
}

