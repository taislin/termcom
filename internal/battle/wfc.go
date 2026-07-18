package battle

import (
	"math/rand"
)

// Direction indices for WFC neighbor rules.
const (
	dirN = 0
	dirE = 1
	dirS = 2
	dirW = 3
)

// wfcDirDX/DY give the grid step for each direction.
var wfcDirDX = [4]int{0, 1, 0, -1}
var wfcDirDY = [4]int{-1, 0, 1, 0}

// wfcOpposite returns the opposite direction index.
func wfcOpposite(d int) int { return d ^ 2 }

// WFCTile is a single modular UFO piece. RuneGrid is the 3x3 footprint where
// '.' denotes empty/hull-fill space. Neighbors[d] lists the tile IDs that may
// legally sit in direction d relative to this tile.
type WFCTile struct {
	ID       int
	Name     string
	RuneGrid [3][3]rune
	// Neighbors[d] for d in {N,E,S,W} gives the set of allowed neighbor tile IDs.
	Neighbors [4][]int
}

// superposition is the set of still-possible tile IDs for one wave cell.
// allowed[id] is true when tile id is still valid here.
type superposition struct {
	allowed  []bool
	count    int // number of true entries; cached for entropy
	collapsed int // -1 if uncollapsed, otherwise the chosen tile ID
}

func (s *superposition) entropy() int { return s.count }

func (s *superposition) collapseTo(id int) {
	for i := range s.allowed {
		s.allowed[i] = false
	}
	s.allowed[id] = true
	s.count = 1
	s.collapsed = id
}

// WFCRules holds the tile set and a fast adjacency matrix.
type WFCRules struct {
	Tiles     []WFCTile
	numTiles  int
	// compatible[a][d][b] is true if tile b may sit in direction d of tile a.
	compatible [][4][]bool
}

// NewWFCRules validates and precomputes the compatibility matrix from tiles.
func NewWFCRules(tiles []WFCTile) *WFCRules {
	n := len(tiles)
	r := &WFCRules{Tiles: tiles, numTiles: n}
	r.compatible = make([][4][]bool, n)
	for a := 0; a < n; a++ {
		for d := 0; d < 4; d++ {
			r.compatible[a][d] = make([]bool, n)
			for _, b := range tiles[a].Neighbors[d] {
				if b >= 0 && b < n {
					r.compatible[a][d][b] = true
				}
			}
		}
	}
	return r
}

// Wave is the 2D grid of superpositions being collapsed.
type Wave struct {
	rules  *WFCRules
	w, h   int
	cells  [][]superposition
}

func newWave(rules *WFCRules, w, h int) *Wave {
	cells := make([][]superposition, h)
	for y := 0; y < h; y++ {
		cells[y] = make([]superposition, w)
		for x := 0; x < w; x++ {
			allowed := make([]bool, rules.numTiles)
			for i := range allowed {
				allowed[i] = true
			}
			cells[y][x] = superposition{allowed: allowed, count: rules.numTiles, collapsed: -1}
		}
	}
	return &Wave{rules: rules, w: w, h: h, cells: cells}
}

// minEntropyCell returns the coordinates of the uncollapsed cell with the
// fewest remaining options, plus whether any such cell exists. Deterministic
// tie-breaking (lowest x then y) keeps generation reproducible per seed.
func (wv *Wave) minEntropyCell(rng *rand.Rand) (int, int, bool) {
	best := -1
	var bx, by int
	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			c := &wv.cells[y][x]
			if c.collapsed >= 0 {
				continue
			}
			if best == -1 || c.count < best {
				best = c.count
				bx, by = x, y
			}
		}
	}
	if best == -1 {
		return 0, 0, false
	}
	return bx, by, true
}

// observe collapses the lowest-entropy cell by randomly choosing one of its
// remaining valid tiles. Returns the chosen tile ID.
func (wv *Wave) observe(x, y int, rng *rand.Rand) int {
	c := &wv.cells[y][x]
	options := make([]int, 0, c.count)
	for i := 0; i < len(c.allowed); i++ {
		if c.allowed[i] {
			options = append(options, i)
		}
	}
	chosen := options[rng.Intn(len(options))]
	c.collapseTo(chosen)
	return chosen
}

// propagate enforces constraints from (x,y) outward using a queue. It returns
// false if any cell is reduced to zero options (contradiction).
func (wv *Wave) propagate(x, y int) bool {
	queue := make([]int, 0, wv.w*wv.h)
	queue = append(queue, x, y)

	for len(queue) > 0 {
		cx := queue[0]
		cy := queue[1]
		queue = queue[2:]

		// The changed cell at (cx,cy) (collapsed or still a superposition that
		// lost options) determines which neighbor tiles remain valid. A neighbor
		// tile b in direction d is valid only if some still-allowed source tile
		// a satisfies compatible[a][d][b].
		src := &wv.cells[cy][cx]

		// compat[d][b] = does b have at least one allowed source neighbor in d.
		compat := [4][]bool{
			make([]bool, wv.rules.numTiles),
			make([]bool, wv.rules.numTiles),
			make([]bool, wv.rules.numTiles),
			make([]bool, wv.rules.numTiles),
		}
		for a := 0; a < wv.rules.numTiles; a++ {
			if !src.allowed[a] {
				continue
			}
			for d := 0; d < 4; d++ {
				for b := 0; b < wv.rules.numTiles; b++ {
					if wv.rules.compatible[a][d][b] {
						compat[d][b] = true
					}
				}
			}
		}

		for d := 0; d < 4; d++ {
			nx, ny := cx+wfcDirDX[d], cy+wfcDirDY[d]
			if nx < 0 || nx >= wv.w || ny < 0 || ny >= wv.h {
				continue
			}
			nb := &wv.cells[ny][nx]
			if nb.collapsed >= 0 {
				continue
			}
			// A neighbor tile b is valid only if compat[d][b] (some allowed
			// source tile is compatible with b in direction d).
			removed := false
			for b := 0; b < len(nb.allowed); b++ {
				if !nb.allowed[b] {
					continue
				}
				if !compat[d][b] {
					nb.allowed[b] = false
					nb.count--
					removed = true
				}
			}
			if nb.count == 0 {
				return false
			}
			if removed {
				queue = append(queue, nx, ny)
			}
		}
	}
	return true
}

// fullyCollapsed reports whether every cell has exactly one tile.
func (wv *Wave) fullyCollapsed() bool {
	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			if wv.cells[y][x].collapsed < 0 {
				return false
			}
		}
	}
	return true
}

// Solve runs the WFC observation/propagation loop with restart-on-contradiction.
// maxRestarts bounds the number of retries; on exhaustion it returns the best
// (most-collapsed) wave found so far.
func (wv *Wave) Solve(rng *rand.Rand, maxRestarts int) *Wave {
	best := wv
	bestCollapsed := 0

	for restart := 0; restart <= maxRestarts; restart++ {
		if restart > 0 {
			wv = newWave(wv.rules, wv.w, wv.h)
		}
		contradiction := false
		for {
			x, y, ok := wv.minEntropyCell(rng)
			if !ok {
				break
			}
			wv.observe(x, y, rng)
			if !wv.propagate(x, y) {
				contradiction = true
				break
			}
		}
		count := 0
		for y := 0; y < wv.h; y++ {
			for x := 0; x < wv.w; x++ {
				if wv.cells[y][x].collapsed >= 0 {
					count++
				}
			}
		}
		if !contradiction && wv.fullyCollapsed() {
			return wv
		}
		if count > bestCollapsed {
			best = wv
			bestCollapsed = count
		}
	}
	return best
}

// tileRuneToType maps a WFC rune to a BattleMap TileType. '.' is treated as
// hull-fill (UFO floor).
func tileRuneToType(ch rune) TileType {
	switch ch {
	case '#':
		return TileUFOWall
	case '.':
		return TileUFOFloor
	case 'D':
		return TileDoor
	case 'C':
		return TileConsole
	case 'M':
		return TileMachinery
	case 'P':
		return TilePod
	case 'X':
		return TilePowerSource
	case 'S':
		return TileStorage
	case 'A':
		return TileAlienTech
	case 'T':
		return TileStairsDown
	default:
		return TileUFOFloor
	}
}

// CompileToBattleMap stamps the collapsed wave (grid of 3x3 tiles) into the
// destination BattleMap at the given level, offset by (ox, oy). The destination
// must be large enough to hold w*3 x h*3.
func (wv *Wave) CompileToBattleMap(m *BattleMap, ox, oy, level int) {
	for gy := 0; gy < wv.h; gy++ {
		for gx := 0; gx < wv.w; gx++ {
			id := wv.cells[gy][gx].collapsed
			if id < 0 {
				continue
			}
			tile := wv.rules.Tiles[id]
			for ty := 0; ty < 3; ty++ {
				for tx := 0; tx < 3; tx++ {
					ch := tile.RuneGrid[ty][tx]
					mx := ox + gx*3 + tx
					my := oy + gy*3 + ty
					m.SetLevel(mx, my, level, tileRuneToType(ch))
				}
			}
		}
	}
}

// ufoWFCTiles defines the modular UFO piece library for the Tiled Model.
// Each piece is 3x3. '.' = floor, '#' = wall, other letters = furniture.
func ufoWFCTiles() []WFCTile {
	floor := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallN := [3][3]rune{
		{'#', '#', '#'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallE := [3][3]rune{
		{'.', '.', '#'},
		{'.', '.', '#'},
		{'.', '.', '#'},
	}
	wallS := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'#', '#', '#'},
	}
	wallW := [3][3]rune{
		{'#', '.', '.'},
		{'#', '.', '.'},
		{'#', '.', '.'},
	}
	corridorNS := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	_ = corridorNS
	cornerNE := [3][3]rune{
		{'#', '#', '#'},
		{'#', '.', '.'},
		{'.', '.', '.'},
	}
	cornerSE := [3][3]rune{
		{'.', '.', '.'},
		{'#', '.', '.'},
		{'#', '#', '#'},
	}
	cornerSW := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', '#'},
		{'#', '#', '#'},
	}
	cornerNW := [3][3]rune{
		{'#', '#', '#'},
		{'.', '.', '#'},
		{'.', '.', '.'},
	}
	engine := [3][3]rune{
		{'.', '#', '.'},
		{'#', 'M', '#'},
		{'.', '#', '.'},
	}
	consoleRoom := [3][3]rune{
		{'.', '.', '.'},
		{'.', 'C', '.'},
		{'.', '.', '.'},
	}
	podRoom := [3][3]rune{
		{'.', 'P', '.'},
		{'.', '.', '.'},
		{'.', 'P', '.'},
	}
	powerCore := [3][3]rune{
		{'#', '.', '#'},
		{'.', 'X', '.'},
		{'#', '.', '#'},
	}
	doorN := [3][3]rune{
		{'.', 'D', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	doorE := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', 'D'},
		{'.', '.', '.'},
	}
	doorS := [3][3]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', 'D', '.'},
	}
	doorW := [3][3]rune{
		{'.', '.', '.'},
		{'D', '.', '.'},
		{'.', '.', '.'},
	}

	// Adjacency helper: a tile's neighbor set per direction. We allow floor to
	// connect to floor/walls/corners/rooms/doors; walls/corners must meet
	// walls/corners/floor (never open edge to open edge across a gap).
	tiles := []WFCTile{
		{ID: 0, Name: "Floor", RuneGrid: floor, Neighbors: [4][]int{}},
		{ID: 1, Name: "WallN", RuneGrid: wallN, Neighbors: [4][]int{}},
		{ID: 2, Name: "WallE", RuneGrid: wallE, Neighbors: [4][]int{}},
		{ID: 3, Name: "WallS", RuneGrid: wallS, Neighbors: [4][]int{}},
		{ID: 4, Name: "WallW", RuneGrid: wallW, Neighbors: [4][]int{}},
		{ID: 5, Name: "CornerNE", RuneGrid: cornerNE, Neighbors: [4][]int{}},
		{ID: 6, Name: "CornerSE", RuneGrid: cornerSE, Neighbors: [4][]int{}},
		{ID: 7, Name: "CornerSW", RuneGrid: cornerSW, Neighbors: [4][]int{}},
		{ID: 8, Name: "CornerNW", RuneGrid: cornerNW, Neighbors: [4][]int{}},
		{ID: 9, Name: "Engine", RuneGrid: engine, Neighbors: [4][]int{}},
		{ID: 10, Name: "ConsoleRoom", RuneGrid: consoleRoom, Neighbors: [4][]int{}},
		{ID: 11, Name: "PodRoom", RuneGrid: podRoom, Neighbors: [4][]int{}},
		{ID: 12, Name: "PowerCore", RuneGrid: powerCore, Neighbors: [4][]int{}},
		{ID: 13, Name: "DoorN", RuneGrid: doorN, Neighbors: [4][]int{}},
		{ID: 14, Name: "DoorE", RuneGrid: doorE, Neighbors: [4][]int{}},
		{ID: 15, Name: "DoorS", RuneGrid: doorS, Neighbors: [4][]int{}},
		{ID: 16, Name: "DoorW", RuneGrid: doorW, Neighbors: [4][]int{}},
	}

	// Define legal adjacencies.
	// "Open" tiles (floor, rooms, doors, engine) may sit anywhere.
	// Wall pieces only allow their open side to face an open tile or a door,
	// and their solid side to face a wall/corner or the map boundary (handled
	// by propagation ignoring out-of-bounds).
	open := []int{0, 9, 10, 11, 12, 13, 14, 15, 16}

	// Floor neighbors: any open tile, or a wall/corner (so rooms can be enclosed).
	tiles[0].Neighbors = [4][]int{open, open, open, open}

	// WallN: solid (wall) on North, open on South.
	// North must face wall/corner; South must face open.
	tiles[1].Neighbors[dirN] = []int{1, 2, 3, 4, 5, 6, 7, 8}
	tiles[1].Neighbors[dirS] = open
	tiles[1].Neighbors[dirE] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	tiles[1].Neighbors[dirW] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

	// WallE: solid on East, open on West.
	tiles[2].Neighbors[dirE] = []int{1, 2, 3, 4, 5, 6, 7, 8}
	tiles[2].Neighbors[dirW] = open
	tiles[2].Neighbors[dirN] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	tiles[2].Neighbors[dirS] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

	// WallS: solid on South, open on North.
	tiles[3].Neighbors[dirS] = []int{1, 2, 3, 4, 5, 6, 7, 8}
	tiles[3].Neighbors[dirN] = open
	tiles[3].Neighbors[dirE] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	tiles[3].Neighbors[dirW] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

	// WallW: solid on West, open on East.
	tiles[4].Neighbors[dirW] = []int{1, 2, 3, 4, 5, 6, 7, 8}
	tiles[4].Neighbors[dirE] = open
	tiles[4].Neighbors[dirN] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	tiles[4].Neighbors[dirS] = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

	// Corners (solid on two sides). They mostly face walls/corners; open
	// diagonal-to-room. Keep permissive: allow walls/corners on all sides so
	// the hull can close, plus open tiles facing inward.
	cornerNeighbors := func() [4][]int {
		return [4][]int{
			[]int{1, 2, 3, 4, 5, 6, 7, 8},
			[]int{1, 2, 3, 4, 5, 6, 7, 8},
			[]int{1, 2, 3, 4, 5, 6, 7, 8},
			[]int{1, 2, 3, 4, 5, 6, 7, 8},
		}
	}
	tiles[5].Neighbors = cornerNeighbors()
	tiles[6].Neighbors = cornerNeighbors()
	tiles[7].Neighbors = cornerNeighbors()
	tiles[8].Neighbors = cornerNeighbors()

	// Engine, rooms, powercore: open on all sides.
	for _, id := range []int{9, 10, 11, 12} {
		tiles[id].Neighbors = [4][]int{open, open, open, open}
	}

	// Doors: open on the axis through the doorway, walls elsewhere implicitly
	// via being open tiles; allow any open neighbor plus walls facing the
	// solid sides. Keep permissive: open on all sides.
	for _, id := range []int{13, 14, 15, 16} {
		tiles[id].Neighbors = [4][]int{open, open, open, open}
	}

	return tiles
}

// GenerateUFOInteriorWFC builds a UFO interior map using the Wave Function
// Collapse Tiled Model on a grid of 3x3 modular pieces. The resulting puzzle
// of tiles is stamped into a multi-level BattleMap. rng must be seeded for
// reproducibility.
func GenerateUFOInteriorWFC(w, h int, rng *rand.Rand) *BattleMap {
	levelH := h / 2
	if levelH < 12 {
		levelH = 12
	}
	m := NewMultiLevelBattleMap(w, levelH, 2)

	rules := NewWFCRules(ufoWFCTiles())

	// Number of 3x3 tile-grid cells that fit on one level.
	gw := w / 3
	gh := levelH / 3
	if gw < 1 {
		gw = 1
	}
	if gh < 1 {
		gh = 1
	}

	buildLevel := func(level int) {
		// Fill level with UFO floor base, then stamp collapsed WFC tiles.
		m.fillRectLevel(0, 0, w, levelH, level, TileUFOFloor)

		wv := newWave(rules, gw, gh)
		wv = wv.Solve(rng, 20)
		wv.CompileToBattleMap(m, 0, 0, level)

		// Enclose: draw outer hull wall border so the ship reads as closed.
		m.drawRectLevel(0, 0, w, levelH, level, TileUFOWall)
	}

	buildLevel(0)
	buildLevel(1)

	// Connect levels with stairs.
	stairsX := (w / 2) & ^1
	stairsY := (levelH / 2) & ^1
	if stairsX+1 >= w {
		stairsX = w - 2
	}
	if stairsY+1 >= levelH {
		stairsY = levelH - 2
	}
	m.SetLevel(stairsX, stairsY, 0, TileStairsDown)
	m.SetLevel(stairsX+1, stairsY, 0, TileUFOFloor)
	m.SetLevel(stairsX, stairsY+1, 0, TileUFOFloor)
	m.SetLevel(stairsX+1, stairsY+1, 0, TileUFOFloor)
	m.SetLevel(stairsX, stairsY, 1, TileStairs)
	m.SetLevel(stairsX+1, stairsY, 1, TileUFOFloor)
	m.SetLevel(stairsX, stairsY+1, 1, TileUFOFloor)
	m.SetLevel(stairsX+1, stairsY+1, 1, TileUFOFloor)

	return m
}
