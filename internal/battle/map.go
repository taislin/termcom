package battle

import (
	"math"
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/mapgen"
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
	// Human furniture tiles
	TileDesk        // Desk/workstation
	TileChair       // Chair/seating (generic)
	TileChairLeft   // Chair facing left (toward a table)
	TileChairRight  // Chair facing right (toward a table)
	TileComputer    // Computer terminal
	TileBed      // Bed/cot
	TileLocker   // Locker/storage
	TileCabinet  // Cabinet/shelving
	// Vehicle tiles
	TileCar          // Car left half: ▄ (top), º (bottom)
	TileCarMid       // Car middle roof: █ (top), empty bottom
	TileCarRight     // Car right half: ▄ (top), º (bottom)
	TileForklift      // Forklift left half: █ (top), º (bottom)
	TileForkliftRight // Forklift right half: ▄ (top), empty (bottom-right)
	// Urban hazard tiles
	TileFuelPump // Fuel pump — explodes on destruction (5x5 blast)
	// Shipping container tiles (indestructible, full LOS block)
	TileContainerRed
	TileContainerBlue
	TileContainerYellow
	// Biome structure tiles
	TileAdobe    // thick adobe wall (dusty orange, indestructible by plasma)
	TileMetalWall // prefab metallic wall (silver, indestructible)
	TileWreck    // aircraft wreckage (rusty metal, indestructible, full cover)
	TileTimber   // stacked timber (flammable, full cover)
	TileDish     // satellite/comms dish (metal, indestructible)
	TileTruck    // military supply truck (vehicle, full cover)
	TileIce      // frozen lake ice (passable, zero cover)
	TileStreetlamp // streetlamp / floodlight pole (emits light, shootable)
	TileGlass     // broken glass / debris on floor (noisy when stepped on)
	TileDebris    // scattered debris / rubble-strewn floor (noisy when stepped on)
	TileCryoPipe  // cryo-coolant pipe (shootable, vents freezing gas)
	TileSkylight  // translucent glass floor on upper levels (collapses under weight)
	// Farm biome tiles
	TileWheat    // tall wheat/corn crop, passable, 20 cover, flammable
	TileHayBale  // hay bale, solid cover, 60 cover, flammable
	// Coastal biome tiles
	TilePier     // wooden pier over water, passable, 10 cover
	TileDockCrate // dock crate, destructible, 50 cover
	// Mountain biome tiles
	TileCliffFace // impassable cliff, 80 cover, indestructible
	TileScree     // loose scree slope, passable, 10 cover, noisy
	TileBoulder   // large boulder, 70 cover, indestructible
	// Swamp biome tiles
	TileSwampWater // shallow murky water, passable, 5 cover, high TU
	TileCypressTree // cypress tree trunk, 80 cover, destructible
	// Jungle biome tiles
	TileMud       // deep mud, passable, 5 cover, high TU, noisy
	TileVine      // dense hanging vines, passable, 20 cover
	TileBamboo    // bamboo thicket, 60 cover, opaque, destructible
	TileDryBush   // dry coastal scrub, passable, 20 cover, flammable
	// Vehicle tiles
	TileBusEnd    // Bus left/right end, 50 cover
	TileBusMid    // Bus middle roof, 50 cover
	TileHeloBody  // Helicopter fuselage, 50 cover
	TileHeloTail  // Helicopter tail boom, 30 cover
	TileHeloNose  // Helicopter nose/cockpit, 40 cover
	TileHeloRotor // Helicopter rotor (overhead, passable below)
	TileHeloRotorSides // Helicopter rotor blade sides (overhead, passable below)
	TileHeloBodyBack // Helicopter rear/back fuselage, 50 cover
	TileHeloRotorBack // Helicopter rear rotor blade (overhead, passable below)
	TileHeloWindow // Helicopter window/glass, 0 cover
	TileTractorCab  // Tractor cab/hood, 30 cover
	TileTractorBody // Tractor body/rear, 30 cover
	// Alien vehicle tiles
	TileCrawlerLeft  // Alien crawler left end: ◢
	TileCrawlerMid   // Alien crawler middle: █
	TileCrawlerRight // Alien crawler right end: ◣
	TileCrawlerLeg   // Alien crawler leg/undercarriage: ^
	// Wheel tiles
	TileWheel      // Vehicle wheel
	TileWheelSmall // Vehicle wheel (small)
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
	Lit       bool // lamp is currently emitting light
	LitByLamp bool // tile is inside a lamp's light radius
	BaseColor tcell.Color
	Rune      rune
}

// MoveCost returns the TU cost for a unit to step onto the tile at
// (x,y). Base cost varies by terrain: pavement is fastest, grass/indoor
// normal, sand/marsh/water are slow. Rain makes soft ground muddy
// (extra cost). Heavy/special terrain is handled by the caller via per-tile
// tile type; crouch cost is added by the caller.
func (m *BattleMap) MoveCost(x, y int, w *Weather) int {
	t := m.At(x, y)
	base := 4
	switch t.Type {
	case TilePavement, TilePier:
		base = 3
	case TileSand, TileMarsh, TileMud, TileScree, TileVine:
		base = 6
	case TileWater, TileSwampWater:
		base = 8
	case TileTree, TileRock, TileCypressTree, TileBamboo:
		base = 8
	case TileSnow, TileIce:
		base = 5
	case TileBush, TileWheat:
		base = 5
	}
	if w != nil && w.MovePenalty() > 0 {
		switch t.Type {
		case TileGrass, TileSand, TileMarsh, TileMud, TileSwampWater:
			base += w.MovePenalty()
		}
	}
	return base
}

// TileCover returns the base cover value for a tile type.
func TileCover(t TileType) int {
	switch t {
	case TileWall, TileUFOWall, TileContainerRed, TileContainerBlue, TileContainerYellow,
		TileAdobe, TileMetalWall, TileWreck, TileTimber, TileTruck,
		TileCliffFace, TileCypressTree:
		return 80
	case TileTree, TileBoulder, TileBamboo:
		return 60
	case TileRock:
		return 70
	case TileHayBale:
		return 60
	case TileBush:
		return 40
	case TileDryBush:
		return 20
	case TileFence:
		return 30
	case TileCryoPipe:
		return 30
	case TileObject, TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet,
		TileCar, TileCarRight, TileCarMid, TileForklift, TileForkliftRight,
		TileFuelPump,
		TileBusEnd, TileBusMid, TileHeloBody, TileHeloBodyBack, TileDockCrate,
		TileCrawlerLeft, TileCrawlerMid, TileCrawlerRight:
		return 50
	case TileCrawlerLeg:
		return 20
	case TileTractorCab, TileTractorBody, TileVine:
		return 30
	case TileHeloTail:
		return 30
	case TileHeloNose:
		return 40
	case TileRubble:
		return 20
	case TileWheat:
		return 20
	case TileDoor:
		return 0
	case TileGlass, TileDebris:
		return 0
	case TileSkylight:
		return 0
	case TilePier, TileScree, TileSwampWater, TileMud, TileHeloRotor, TileHeloRotorSides, TileHeloRotorBack, TileWheel, TileWheelSmall:
		return 5
	default:
		return 0
	}
}

func (t Tile) IsFlammable() bool {
	switch t.Type {
	case TileGrass, TileTree, TileBush, TileDryBush, TileFence, TileDoor, TileTimber,
		TileWheat, TileHayBale, TileVine, TileBamboo, TileCypressTree:
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
			tile.Rune = tileChars[TileFloor]
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
	TileFence:      '│',
	TileRubble:     '▒',
	TileObject:     '•',
	// UFO furniture characters
	TileConsole:     '⌸', // Console panel (U+2338 QUAD MINUS)
	TileMachinery:   '⊛', // Machinery (U+229B CIRCLED ASTERISK)
	TilePod:         '◈', // Alien pod
	TilePowerSource: '⌁', // Power source (U+2301 ELECTRICAL ARC)
	TileStorage:     '▤', // Storage container
	TileAlienTech:   '⊕', // Alien technology
	// Human furniture characters
	TileDesk:     '◊', // Desk (U+25CA LOZENGE)
	TileChair:    '⊟', // Chair (U+229F)
	TileChairLeft:  '⅃', // Chair facing left (toward a table)
	TileChairRight: 'L', // Chair facing right (toward a table)
	TileComputer: '⌸', // Console (U+2338 QUAD MINUS)
	TileBed:          '□', // Bed (U+25A1)
	TileLocker:       '◫', // Locker (U+25EB)
	TileCabinet:      '⊞', // Cabinet (U+229E)
	TileCar:          '▄', // Car left half (top)
	TileCarMid:       '█', // Car middle roof (top only)
	TileCarRight:     '▄', // Car right half (top)
	TileForklift:     '█', // Forklift left half (top)
	TileForkliftRight: '⊏', // Forklift right half (top)
	// Urban hazard characters
	TileFuelPump: '8', // Fuel pump (looks like nozzle)
	// Shipping container characters
	TileContainerRed:    '█',
	TileContainerBlue:   '█',
	TileContainerYellow: '█',
	// Biome structure characters
	TileAdobe:    '█', // adobe wall (dusty orange)
	TileMetalWall: '█', // prefab metallic wall (silver)
	TileWreck:    '▤', // aircraft wreckage (rusty)
	TileTimber:   '≡', // stacked timber
	TileDish:     '◗', // satellite dish
	TileTruck:    '▄', // military truck (top half)
	TileIce:      '≈', // frozen lake ice
	TileStreetlamp: '⌖', // lamp/floodlight fixture
	TileGlass:     ',', // broken glass / debris (noisy step)
	TileDebris:    '`', // scattered debris (noisy step)
	TileCryoPipe:  '╪', // cryo-coolant pipe
	TileSkylight:  '⊙', // glass skylight floor
	TileWheat:     '▓', // tall wheat (dense crop pattern)
	TileHayBale:   '█', // hay bale
	TilePier:      '═', // wooden pier planks
	TileDockCrate: '▣', // dock crate (square with center dot)
	TileCliffFace: '░', // cliff face (stippled rock)
	TileScree:     '·', // loose scree (like sand/grit dots)
	TileBoulder:   '∩', // large boulder (same as rock but bigger shape)
	TileSwampWater: '≋', // swamp water (wavy)
	TileCypressTree: '♣', // cypress tree (same char as tree, different color)
	TileMud:       '≋', // mud (wavy, like marsh)
	TileVine:      '‡', // vines (dense cross)
	TileBamboo:    '♣', // bamboo (same char as tree, different color)
	TileDryBush:   '*', // dry coastal scrub
	TileBusEnd:    '▄', // Bus end (top)
	TileBusMid:    '█', // Bus middle roof (top)
	TileHeloBody:  '█', // Helicopter fuselage (top)
	TileHeloTail:  '▄', // Helicopter tail (top)
	TileHeloNose:  '▷', // Helicopter nose (top)
	TileHeloRotor: '+', // Helicopter rotor (overhead)
	TileHeloRotorSides: '-', // Helicopter rotor sides (overhead)
	TileHeloBodyBack:  '█', // Helicopter rear fuselage
	TileHeloRotorBack: 'x', // Helicopter rear rotor
	TileHeloWindow:    '◣', // Helicopter window glass
	TileTractorCab:  '◣', // Tractor cab (top)
	TileTractorBody: '█', // Tractor body (top)
	TileCrawlerLeft:  '◢', // Crawler left end
	TileCrawlerMid:   '█', // Crawler middle body
	TileCrawlerRight: '◣', // Crawler right end
	TileCrawlerLeg:   '^', // Crawler leg
	TileWheel:        'O', // Vehicle wheel
	TileWheelSmall:   'o', // Vehicle wheel (small)
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
	GroundLoot   map[[2]int][]string // items on the ground, keyed by tile position
}

func NewBattleMap(w, h int) *BattleMap {
	m := &BattleMap{
		Width:        w,
		Height:       h,
		NumLevels:    1,
		LevelHeight:  h,
		CurrentLevel: 0,
		Tiles:        make([][]Tile, h),
		GroundLoot:   make(map[[2]int][]string),
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
		GroundLoot:   make(map[[2]int][]string),
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
		TileMarsh,
		TileIce, TileGlass, TileDebris,
		TileSkylight, TileBush, TileDryBush,
		TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet,
		TileRubble,
		TileWheat, TilePier, TileScree, TileSwampWater, TileMud, TileVine, TileHeloRotor, TileHeloRotorSides, TileHeloRotorBack:
		return true
	}
	return false
}

func (m *BattleMap) Opaque(x, y int) bool {
	t := m.At(x, y)
	switch t.Type {
	case TileWall, TileTree, TileRock, TileUFOWall, TileFence,
		TileContainerRed, TileContainerBlue, TileContainerYellow,
		TileAdobe, TileMetalWall, TileWreck, TileTruck, TileDish,
		TileHayBale, TileDockCrate, TileCliffFace, TileBoulder,
		TileCypressTree, TileBamboo,
		TileBusEnd, TileBusMid,
		TileHeloBody, TileHeloTail, TileHeloNose, TileHeloBodyBack,
		TileTractorCab, TileTractorBody,
		TileCrawlerLeft, TileCrawlerMid, TileCrawlerRight, TileCrawlerLeg:
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
	case TileWall, TileUFOWall, TileTree, TileRock, TileFence, TileDoor,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet,
		TileCar, TileCarRight, TileCarMid, TileForklift, TileForkliftRight,
		TileFuelPump, TileStreetlamp, TileCryoPipe, TileSkylight,
		TileWheat, TileHayBale,
		TileDockCrate, TileScree, TileDryBush,
		TileCypressTree, TileVine, TileBamboo,
		TileBusEnd, TileBusMid,
		TileHeloBody, TileHeloTail, TileHeloNose, TileHeloBodyBack, TileHeloRotor, TileHeloRotorSides, TileHeloRotorBack, TileHeloWindow, TileWheel, TileWheelSmall,
		TileTractorCab, TileTractorBody,
		TileCrawlerLeft, TileCrawlerMid, TileCrawlerRight, TileCrawlerLeg:
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
		tile.Lit = false
		tile.Type = TileRubble
		tile.Cover = TileCover(TileRubble)
	} else {
		arrayY := y + m.CurrentLevel*m.LevelHeight
		if x < 0 || x >= m.Width || arrayY < 0 || arrayY >= m.Height {
			return false
		}
		tile := &m.Tiles[arrayY][x]
		tile.Lit = false
		tile.Type = TileRubble
		tile.Cover = TileCover(TileRubble)
	}
	return true
}

// CollapseSkylight destroys the skylight at (x,y) on the current level and
// returns true if a collapse occurred. Only acts on upper levels (not ground).
func (m *BattleMap) CollapseSkylight(x, y int) bool {
	if m.NumLevels <= 1 || m.CurrentLevel == 0 {
		return false
	}
	arrayY := y + m.CurrentLevel*m.LevelHeight
	if x < 0 || x >= m.Width || arrayY < 0 || arrayY >= m.LevelHeight*m.NumLevels {
		return false
	}
	tile := &m.Tiles[arrayY][x]
	if tile.Type != TileSkylight {
		return false
	}
	tile.Type = TileRubble
	tile.Cover = TileCover(TileRubble)
	tile.Rune = tileChars[TileRubble]
	return true
}

// FuelPumpExplosionRadius returns the blast radius for fuel pump explosions.
const FuelPumpExplosionRadius = 5

// ExplodesOnDestruction returns the explosion radius if the tile at (x,y)
// should chain-explode when destroyed, or 0 if it does not explode.
func (m *BattleMap) ExplodesOnDestruction(x, y int) int {
	tile := m.At(x, y)
	switch tile.Type {
	case TileFuelPump:
		return FuelPumpExplosionRadius
	}
	return 0
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
			m.Tiles[y][x].Seen = false
			m.Tiles[y][x].LitByLamp = false
		}
	}
}

// applyLampLight forces visibility on tiles within the light radius of any
// lit streetlamp. Called after unit-based FOV so lamplit areas stay
// visible even without a unit having direct line-of-sight. Radius 3 = 7x7.
func (m *BattleMap) applyLampLight() {
	const r = 3
	startY := 0
	endY := m.Height
	if m.NumLevels > 1 {
		startY = m.CurrentLevel * m.LevelHeight
		endY = startY + m.LevelHeight
	}
	for y := startY; y < endY; y++ {
		for x := 0; x < m.Width; x++ {
			t := &m.Tiles[y][x]
			if t.Type != TileStreetlamp || !t.Lit {
				continue
			}
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					if dx*dx+dy*dy > r*r {
						continue
					}
					tx, ty := x+dx, y+dy
					if tx < 0 || tx >= m.Width || ty < startY || ty >= endY {
						continue
					}
					m.Tiles[ty][tx].Visible = true
					m.Tiles[ty][tx].Seen = true
					m.Tiles[ty][tx].LitByLamp = true
				}
			}
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
		if m.Opaque(x, y) && !((x == x1 && y == y1) || (x == x2 && y == y2)) {
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

// fillRect fills a rectangle with a tile type on the current level.
func (m *BattleMap) fillRect(x, y, w, h int, t TileType) {
	m.fillRectLevel(x, y, w, h, m.CurrentLevel, t)
}

// drawRect draws a rectangle outline with a tile type on the current level.
func (m *BattleMap) drawRect(x, y, w, h int, t TileType) {
	m.drawRectLevel(x, y, w, h, m.CurrentLevel, t)
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

// corridorImpl is the shared L-shaped corridor implementation.
// guardTile is the tile type that prevents carving; fillTile is the replacement.
func (m *BattleMap) corridorImplProtected(x1, y1, x2, y2, w, level int, fillTile TileType) {
	corridorX := func(x, y int) {
		for dy := 0; dy < w; dy++ {
			t := m.AtLevel(x, y+dy, level)
			if !isUrbanProtected(t.Type) {
				m.SetLevel(x, y+dy, level, fillTile)
			}
		}
	}
	corridorY := func(x, y int) {
		for dx := 0; dx < w; dx++ {
			t := m.AtLevel(x+dx, y, level)
			if !isUrbanProtected(t.Type) {
				m.SetLevel(x+dx, y, level, fillTile)
			}
		}
	}
	if rand.Intn(2) == 0 {
		start := min(x1, x2)
		end := max(x1, x2)
		for x := start; x <= end; x++ {
			corridorX(x, y1)
		}
		start = min(y1, y2)
		end = max(y1, y2)
		for y := start; y <= end; y++ {
			corridorY(x2, y)
		}
	} else {
		start := min(y1, y2)
		end := max(y1, y2)
		for y := start; y <= end; y++ {
			corridorY(x1, y)
		}
		start = min(x1, x2)
		end = max(x1, x2)
		for x := start; x <= end; x++ {
			corridorX(x, y2)
		}
	}
}

func (m *BattleMap) corridorImpl(x1, y1, x2, y2, w, level int, guardTile, fillTile TileType) {
	if rand.Intn(2) == 0 {
		start := min(x1, x2)
		end := max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.AtLevel(x, y1+dy, level).Type != guardTile {
					m.SetLevel(x, y1+dy, level, fillTile)
				}
			}
		}
		start = min(y1, y2)
		end = max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.AtLevel(x2+dx, y, level).Type != guardTile {
					m.SetLevel(x2+dx, y, level, fillTile)
				}
			}
		}
	} else {
		start := min(y1, y2)
		end := max(y1, y2)
		for y := start; y <= end; y++ {
			for dx := 0; dx < w; dx++ {
				if m.AtLevel(x1+dx, y, level).Type != guardTile {
					m.SetLevel(x1+dx, y, level, fillTile)
				}
			}
		}
		start = min(x1, x2)
		end = max(x1, x2)
		for x := start; x <= end; x++ {
			for dy := 0; dy < w; dy++ {
				if m.AtLevel(x, y2+dy, level).Type != guardTile {
					m.SetLevel(x, y2+dy, level, fillTile)
				}
			}
		}
	}
}

func (m *BattleMap) corridorLevel(x1, y1, x2, y2, w, level int) {
	m.corridorImpl(x1, y1, x2, y2, w, level, TileUFOWall, TileUFOFloor)
}

// generateCorridor creates an L-shaped corridor between two points on the current level.
func (m *BattleMap) generateCorridor(x1, y1, x2, y2 int, w int) {
	m.corridorImpl(x1, y1, x2, y2, w, m.CurrentLevel, TileWall, TileFloor)
}

// isUrbanProtected returns true for tiles that should not be overwritten
// by corridor generation (walls, furniture, vehicles, street objects).
func isUrbanProtected(t TileType) bool {
	switch t {
	case TileWall, TileDoor, TileWindow,
		TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet,
		TileCar, TileCarMid, TileCarRight, TileForklift, TileForkliftRight, TileWheel, TileWheelSmall,
		TileObject, TileTree, TileBush, TileFence,
		TileFuelPump, TileContainerRed, TileContainerBlue, TileContainerYellow,
		TileAdobe, TileMetalWall, TileWreck, TileTimber, TileDish, TileTruck,
		TileStreetlamp, TileCryoPipe, TileGlass, TileDebris, TileSkylight,
		TileWheat, TileHayBale,
		TilePier, TileDockCrate,
		TileCliffFace, TileScree, TileBoulder,
		TileWater, TileSwampWater, TileCypressTree,
		TileMud, TileVine, TileBamboo, TileDryBush,
		TileBusEnd, TileBusMid,
		TileHeloBody, TileHeloTail, TileHeloNose, TileHeloBodyBack, TileHeloRotor, TileHeloRotorSides, TileHeloRotorBack, TileHeloWindow,
		TileTractorCab, TileTractorBody,
		TileCrawlerLeft, TileCrawlerMid, TileCrawlerRight, TileCrawlerLeg:
		return true
	}
	return false
}

// generateCorridorFill creates an L-shaped corridor using a specific fill tile.
// Corridors respect walls, furniture, vegetation, and vehicles.
func (m *BattleMap) generateCorridorFill(x1, y1, x2, y2, w int, fill TileType) {
	m.corridorImplProtected(x1, y1, x2, y2, w, m.CurrentLevel, fill)
}

// generateCorridorUFO creates an L-shaped corridor that carves through
// UFO walls (used by alien base / Cydonia-style maps).
func (m *BattleMap) generateCorridorUFO(x1, y1, x2, y2 int, w int) {
	m.corridorImpl(x1, y1, x2, y2, w, m.CurrentLevel, TileUFOWall, TileUFOFloor)
}

type MapCommandType int

const (
	CmdFillRect MapCommandType = iota
	CmdDrawRect
	CmdScatter
	CmdPlaceBuilding
	CmdCorridor
	CmdClearArea
	CmdBlob
	CmdPoisson
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
	Seeds    int // for Blob: number of clusters
	Size     int // for Blob: target tiles per cluster
	Radius   int // for Poisson: min spacing
	Seed     int64 // for commands needing an RNG seed
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
		m.furnishBuilding(cmd.X, cmd.Y, cmd.W, cmd.H)
	case CmdCorridor:
		m.generateCorridor(cmd.X, cmd.Y, cmd.X2, cmd.Y2, max(1, cmd.W))
	case CmdClearArea:
		// identical to CmdFillRect; kept for semantic clarity only
		m.fillRect(cmd.X, cmd.Y, cmd.W, cmd.H, cmd.Tile)
	case CmdBlob:
		rng := rand.New(rand.NewSource(cmd.Seed + int64(cmd.X*73856093+cmd.Y*19349663)))
		m.Blob(cmd.Tile, cmd.Seeds, cmd.Size, cmd.Prob, rng)
	case CmdPoisson:
		rng := rand.New(rand.NewSource(cmd.Seed + int64(cmd.X*83492791+cmd.Y*2654435761)))
		m.Poisson(cmd.Tile, cmd.Radius, cmd.Count, rng)
	}
}

func (m *BattleMap) furnishBuilding(bx, by, bw, bh int) {
	var interior [][2]int
	for y := by + 1; y < by+bh-1; y++ {
		for x := bx + 1; x < bx+bw-1; x++ {
			if m.At(x, y).Type == TileFloor {
				interior = append(interior, [2]int{x, y})
			}
		}
	}
	if len(interior) == 0 {
		return
	}
	numFurniture := 2 + rand.Intn(min(3, len(interior)))
	if numFurniture > len(interior) {
		numFurniture = len(interior)
	}
	furnitureTypes := []TileType{TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet}
	for i := 0; i < numFurniture; i++ {
		idx := rand.Intn(len(interior))
		x, y := interior[idx][0], interior[idx][1]
		ft := furnitureTypes[rand.Intn(len(furnitureTypes))]
		m.Set(x, y, ft)
		interior = append(interior[:idx], interior[idx+1:]...)
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

// crashBiomeFromCoords maps world-map pixel coordinates (0-179, 0-89) to a
// biome. The mapping is a deterministic function of position so that the same
// location always gets the same terrain, giving the world a sense of place.
func crashBiomeFromCoords(worldX, worldY int) string {
	biomes := []string{
		"forest", "desert", "polar", "rural",
		"farm", "coastal", "mountain", "swamp", "jungle",
	}
	idx := (worldX*73856093 + worldY*19349663) % len(biomes)
	if idx < 0 {
		idx = -idx
	}
	return biomes[idx]
}

// GenerateCrashSite creates a crash site map with a procedural UFO blueprint
// stamped onto clustered terrain. Returns both the map and crash result.
// worldX/worldY are the crash location on the world map (0-179, 0-89) and
// determine the biome. Pass -1,-1 to default to forest.
func GenerateCrashSite(w, h int, seed int64, worldX, worldY int) (*BattleMap, *CrashResult) {
	biome := "forest"
	if worldX >= 0 && worldY >= 0 {
		biome = crashBiomeFromCoords(worldX, worldY)
	}
	rng := rand.New(rand.NewSource(seed))
	m := AssembleMap(biome, w, h, rng)

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
		dx := rng.Intn(12) - 6
		dy := rng.Intn(10) - 5
		x := ufoX + bp.Width/2 + dx
		y := ufoY + bp.Height/2 + dy
		if m.At(x, y).Type == TileGrass || m.At(x, y).Type == TileTree {
			m.Set(x, y, TileRubble)
		}
	}

	return m, &result
}

// GenerateTerrorSite creates a terror site map (OpenXcom: 50x50 urban) using a
// strict "Grid and Zoning" (plots and parcels) pipeline so the result reads as a
// planned human city rather than scattered fragments:
//
//  1. Road skeleton: one main vertical avenue + 1-2 horizontal cross-streets.
//  2. Sidewalks: grass tiles adjacent to roads become paved walkways.
//  3. City blocks: the roads divide the map into rectangular Lots. Buildings are
//     sidewalk-anchored (entrance facing the road) and never overlap roads or
//     each other.
//  4. Contextual scatter (masking): cars only on road/sidewalk; trees confined
//     to empty grass Lots (parks).
//
// All randomness is seeded so each mission layout is reproducible.
func GenerateTerrorSite(w, h int, seed int64) *BattleMap {
	rng := rand.New(rand.NewSource(seed))
	m := NewBattleMap(w, h)
	m.fillRect(0, 0, w, h, TileGrass)

	// ---- 1. Road skeleton (bounding-box fills, no organic pathfinding) ----
	roadW := 4 + rng.Intn(2) // 4-5
	roadX := w/2 - roadW/2 + rng.Intn(3) - 1
	if roadX < 1 {
		roadX = 1
	}
	if roadX+roadW > w-1 {
		roadX = w - roadW - 1
	}
	m.fillRect(roadX, 0, roadW, h, TilePavement)

	numCross := 1 + rng.Intn(2) // 1-2 horizontal streets
	crossWidths := []int{3 + rng.Intn(2)} // 3-4
	if numCross == 2 {
		crossWidths = append(crossWidths, 3+rng.Intn(2))
	}
	// Evenly distribute cross-streets down the map.
	for i := 0; i < numCross; i++ {
		cw := crossWidths[i]
		cy := (h * (i + 1)) / (numCross + 1)
		cy -= cw / 2
		if cy < 1 {
			cy = 1
		}
		if cy+cw > h-1 {
			cy = h - cw - 1
		}
		m.fillRect(0, cy, w, cw, TilePavement)
	}

	// ---- 2. Sidewalks: grass N/S/E/W adjacent to a road becomes walkway ----
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if m.At(x, y).Type != TileGrass {
				continue
			}
			adjRoad := false
			for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || nx >= w || ny < 0 || ny >= h {
					continue
				}
				if m.At(nx, ny).Type == TilePavement {
					adjRoad = true
					break
				}
			}
			if adjRoad {
				m.Set(x, y, TileFloor)
			}
		}
	}

	// ---- 3. City blocks & sidewalk-anchored building placement ----
	type lot struct{ x, y, w, h int }
	var lots []lot
	// Decompose free (non-road) space into maximal rectangles.
	used := make([][]bool, h)
	for y := 0; y < h; y++ {
		used[y] = make([]bool, w)
	}
	for gy := 0; gy < h; gy++ {
		for gx := 0; gx < w; gx++ {
			if used[gy][gx] || m.At(gx, gy).Type == TilePavement {
				if m.At(gx, gy).Type == TilePavement {
					used[gy][gx] = true
				}
				continue
			}
			rw := 1
			for gx+rw < w && !used[gy][gx+rw] && m.At(gx+rw, gy).Type != TilePavement {
				rw++
			}
			rh := 1
		grow:
			for gy+rh < h {
				for cx := gx; cx < gx+rw; cx++ {
					if used[gy+rh][cx] || m.At(cx, gy+rh).Type == TilePavement {
						break grow
					}
				}
				rh++
			}
			lots = append(lots, lot{gx, gy, rw, rh})
			for yy := gy; yy < gy+rh; yy++ {
				for xx := gx; xx < gx+rw; xx++ {
					used[yy][xx] = true
				}
			}
		}
	}

	buildingIDs := []string{
		"urban_building", "urban_apartment", "urban_shop", "urban_corner_store",
		"urban_warehouse", "urban_tower", "urban_parking_lot", "urban_rooftop",
		"urban_rubble", "bus_stop_cover", "ruined_shack",
	}

	// Fill each lot with one or more buildings (a block holds a cluster of
	// houses) or leave it as a park.
	type placed struct{ x, y, w, h int }
	var placedBuildings []placed
	parkLots := map[int]bool{}
	for li, l := range lots {
		if l.w < 3 || l.h < 3 {
			continue
		}
		// Reserve ~1 in 5 sizeable lots as parks (green space) so the city is
		// not wall-to-wall buildings.
		if l.w >= 6 && l.h >= 6 && rng.Intn(5) == 0 {
			parkLots[li] = true
			continue
		}
		// Subdivide the block and pack several fitting buildings with spacing.
		// setback = gap from lot edge to first building; gap = spacing between
		// neighbouring buildings (keeps yards/alleys between houses).
		const setback = 1
		const gap = 1
		attempts := 0
		placedInLot := 0
		for attempts < 60 {
			attempts++
			// Randomly choose a building that could fit somewhere in the lot.
			id := buildingIDs[rng.Intn(len(buildingIDs))]
			c := mapgen.Get(id)
			if c == nil {
				continue
			}
			if c.Width+2*setback > l.w || c.Height+2*setback > l.h {
				continue
			}
			// Random free anchor inside the lot (with setback).
			maxX := l.x + l.w - c.Width - setback
			maxY := l.y + l.h - c.Height - setback
			if maxX < l.x+setback || maxY < l.y+setback {
				continue
			}
			ax := l.x + setback + rng.Intn(maxX-(l.x+setback)+1)
			ay := l.y + setback + rng.Intn(maxY-(l.y+setback)+1)
			// Reject if it overlaps an already-placed building (incl. gap).
			overlap := false
			for _, p := range placedBuildings {
				if ax < p.x+p.w+gap && ax+c.Width+gap > p.x &&
					ay < p.y+p.h+gap && ay+c.Height+gap > p.y {
					overlap = true
					break
				}
			}
			if overlap {
				continue
			}
			// Reject if it overlaps a road (shouldn't happen inside a lot, but
			// be safe at lot edges).
			hitRoad := false
			for dy := 0; dy < c.Height && !hitRoad; dy++ {
				for dx := 0; dx < c.Width; dx++ {
					if m.At(ax+dx, ay+dy).Type == TilePavement {
						hitRoad = true
						break
					}
				}
			}
			if hitRoad {
				continue
			}
			ApplyMapgenChunkRotated(m, ax, ay, 0, c)
			placedBuildings = append(placedBuildings, placed{ax, ay, c.Width, c.Height})
			placedInLot++
			// Stop early once the block is reasonably filled.
			if placedInLot >= 6 {
				break
			}
		}
	}

	// ---- 4a. Cars: only on road/sidewalk tiles ----
	carChunk := mapgen.Get("urban_car")
	if carChunk != nil {
		carTries := 0
		carsPlaced := 0
		for carsPlaced < 6+w/12 && carTries < 200 {
			carTries++
			cx := 1 + rng.Intn(max(1, w-carChunk.Width-1))
			cy := 1 + rng.Intn(max(1, h-carChunk.Height-1))
			ok := true
			for dy := 0; dy < carChunk.Height && ok; dy++ {
				for dx := 0; dx < carChunk.Width; dx++ {
					tt := m.At(cx+dx, cy+dy).Type
					if tt != TilePavement && tt != TileFloor {
						ok = false
						break
					}
				}
			}
			if !ok {
				continue
			}
			ApplyMapgenChunkRotated(m, cx, cy, 0, carChunk)
			carsPlaced++
		}
	}

	// ---- 4b. Parks: reserved grass lots get trees via masked Poisson ----
	for li, l := range lots {
		if !parkLots[li] {
			continue
		}
		// Confine tree scatter to this lot rectangle (masking).
		maskPoissonInRect(m, TileTree, 2, (l.w*l.h)/10+1, l.x, l.y, l.w, l.h, rng)
	}

	// ---- 4c. Streetlamps along sidewalks/roads ----
	{
		lampCount := w * h / 150
		placed := [][2]int{}
		attempts := 0
		for len(placed) < lampCount && attempts < lampCount*20 {
			attempts++
			x := 1 + rng.Intn(max(1, w-2))
			y := 1 + rng.Intn(max(1, h-2))
			tt := m.At(x, y).Type
			if tt != TilePavement {
				continue
			}
			ok := true
			for _, p := range placed {
				dx, dy := p[0]-x, p[1]-y
				if dx*dx+dy*dy < 25 {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			m.Set(x, y, TileStreetlamp)
			m.Tiles[y][x].Lit = true
			placed = append(placed, [2]int{x, y})
		}
	}

	// Street furniture on sidewalks/roads.
	maskPoissonInRect(m, TileObject, 3, w*h/250, 0, 0, w, h, rng)

	return m
}

// maskPoissonInRect scatters t using Poisson spacing but only onto tiles whose
// current type is one of the allowed passable ground types, confined to the
// given rectangle. Used to keep trees in parks and cars/furniture on pavement.
func maskPoissonInRect(m *BattleMap, t TileType, radius, count, x0, y0, rw, rh int, rng *rand.Rand) {
	placed := [][2]int{}
	attempts := 0
	for len(placed) < count && attempts < count*20 {
		attempts++
		x := x0 + rng.Intn(max(1, rw-2)) + 1
		y := y0 + rng.Intn(max(1, rh-2)) + 1
		if x >= m.Width || y >= m.LevelHeight {
			continue
		}
		switch m.At(x, y).Type {
		case TileGrass, TileFloor, TilePavement, TileSand, TileSnow, TileIce:
		default:
			continue
		}
		ok := true
		for _, p := range placed {
			dx, dy := p[0]-x, p[1]-y
			if dx*dx+dy*dy < radius*radius {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		m.Set(x, y, t)
		placed = append(placed, [2]int{x, y})
	}
}

func GenerateAbductionSite(w, h int) *BattleMap {
	return AssembleMap("rural", w, h, rand.New(rand.NewSource(int64(w*32452843+h*456789))))
}

// GenerateUFOInterior creates a UFO interior map (OpenXcom: 50x50)
func GenerateUFOInterior(w, h int, seed int64) *BattleMap {
	rng := rand.New(rand.NewSource(seed))
	rn := func(n int) int {
		if n <= 0 {
			return 0
		}
		return rng.Intn(n)
	}
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
		numRooms := 5 + rng.Intn(3)
		for i := 0; i < numRooms && attempts < 100; i++ {
			attempts++
			rw := 5 + rng.Intn(5)
			rh := 4 + rng.Intn(4)
			rx := 2 + rng.Intn(max(1, w-rw-4))
			ry := 2 + rng.Intn(max(1, levelH-rh-4))

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

			doorSide := rng.Intn(4)
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
			roomType := rng.Intn(5)
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
		x := rn(w-4) + 2
		y := rn(levelH-4) + 2
		if m.AtLevel(x, y, 0).Type == TileUFOFloor {
			m.SetLevel(x, y, 0, TileAlienTech)
		}
	}

	return m
}

// placeDoor places a door on side `side` (0=south, 1=east, 2=north, 3=west)
// of an axis-aligned rectangle at (x,y) with the given size.
func (m *BattleMap) placeDoor(x, y, size, side int) {
	switch side {
	case 0:
		m.Set(x+size/2, y-1, TileDoor)
	case 1:
		m.Set(x+size/2, y+size, TileDoor)
	case 2:
		m.Set(x-1, y+size/2, TileDoor)
	case 3:
		m.Set(x+size, y+size/2, TileDoor)
	}
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
	m.placeDoor(brainX, brainY, brainSize, rand.Intn(4))

	// Pods: 3-6 rooms around command center
	podCount := 3 + rand.Intn(4)
	podSize := 5 + rand.Intn(2) // 5-6
	var podPositions [][2]int
	for i := 0; i < podCount; i++ {
		angle := float64(i) * (2 * math.Pi / float64(podCount))
		radius := float64(ccSize/2+podSize+2)
		px := w/2 + int(angle*radius*0.7) - podSize/2
		py := h/2 + int(angle*radius*0.5) - podSize/2
		// Clamp to map bounds
		px = max(2, min(px, w-podSize-2))
		py = max(2, min(py, h-podSize-2))
		m.drawRect(px, py, podSize, podSize, TileUFOWall)
		m.fillRect(px+1, py+1, podSize-2, podSize-2, TileUFOFloor)
		m.placeDoor(px, py, podSize, rand.Intn(4))
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
		m.generateCorridorUFO(ccX+ccSize/2, ccY+ccSize/2, pos[0], pos[1], 2)
	}

	// Connect entrance to command center
	switch entranceSide {
	case 0:
		m.generateCorridorUFO(w/2, 6, ccX+ccSize/2, ccY, 2)
	case 1:
		m.generateCorridorUFO(w/2, h-6, ccX+ccSize/2, ccY+ccSize, 2)
	case 2:
		m.generateCorridorUFO(6, h/2, ccX, ccY+ccSize/2, 2)
	case 3:
		m.generateCorridorUFO(w-6, h/2, ccX+ccSize, ccY+ccSize/2, 2)
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

	// Guarantee the whole interior is reachable from the command center.
	m.RepairConnectivity(ccX+ccSize/2, ccY+ccSize/2)

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
		m.generateCorridorUFO(coreCx, coreCy, p[0], p[1], 2)
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

	// Guarantee the whole interior and entrance area is reachable from the core.
	m.RepairConnectivity(coreCx, coreCy)

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

// GenerateForest creates a forest map (OpenXcom: 50x50) via AssembleMap.
func GenerateForest(w, h int) *BattleMap {
	return AssembleMap("forest", w, h, rand.New(rand.NewSource(int64(w*73856093+h*19349663))))
}

// GenerateDesert creates a desert map (OpenXcom: 50x50) via AssembleMap.
func GenerateDesert(w, h int) *BattleMap {
	return AssembleMap("desert", w, h, rand.New(rand.NewSource(int64(w*19349663+h*83492791))))
}

// GeneratePolar creates a polar map (OpenXcom: 50x50) via AssembleMap.
func GeneratePolar(w, h int) *BattleMap {
	return AssembleMap("polar", w, h, rand.New(rand.NewSource(int64(w*2654435761+h*19349663))))
}

// GenerateFarm creates a farm map via AssembleMap.
func GenerateFarm(w, h int) *BattleMap {
	return AssembleMap("farm", w, h, rand.New(rand.NewSource(int64(w*92743891+h*52938467))))
}

// GenerateCoastal creates a coastal map via AssembleMap.
func GenerateCoastal(w, h int) *BattleMap {
	return AssembleMap("coastal", w, h, rand.New(rand.NewSource(int64(w*38291457+h*65102834))))
}

// GenerateMountain creates a mountain map via AssembleMap.
func GenerateMountain(w, h int) *BattleMap {
	return AssembleMap("mountain", w, h, rand.New(rand.NewSource(int64(w*46739281+h*81927364))))
}

// GenerateSwamp creates a swamp map via AssembleMap.
func GenerateSwamp(w, h int) *BattleMap {
	return AssembleMap("swamp", w, h, rand.New(rand.NewSource(int64(w*59384721+h*17283946))))
}

// GenerateJungle creates a jungle map via AssembleMap.
func GenerateJungle(w, h int) *BattleMap {
	return AssembleMap("jungle", w, h, rand.New(rand.NewSource(int64(w*73829164+h*38475621))))
}

// GenerateRural creates a rural map via AssembleMap.
func GenerateRural(w, h int) *BattleMap {
	return AssembleMap("rural", w, h, rand.New(rand.NewSource(int64(w*41827364+h*92837465))))
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

