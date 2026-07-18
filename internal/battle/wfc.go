package battle

import (
	"math/rand"

	"github.com/taislin/termcom/internal/mapgen"
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

// WFCTile is a single modular building piece. RuneGrid is the footprint
// (rows x cols) where '.' denotes empty/floor-fill space. Tiles may be any
// size (small 3x3 pieces up to large multi-room blocks). Neighbors[d] lists
// the tile IDs that may legally sit in direction d relative to this tile.
type WFCTile struct {
	ID       int
	Name     string
	RuneGrid [][]rune
	// Neighbors[d] for d in {N,E,S,W} gives the set of allowed neighbor tile IDs.
	Neighbors [4][]int
}

// gridRows/cols return the footprint dimensions of a tile.
func (t WFCTile) gridRows() int { return len(t.RuneGrid) }
func (t WFCTile) gridCols() int {
	if len(t.RuneGrid) == 0 {
		return 0
	}
	return len(t.RuneGrid[0])
}

// superposition is the set of still-possible tile IDs for one wave cell.
// allowed[id] is true when tile id is still valid here.
type superposition struct {
	allowed  []bool
	count    int // number of true entries; cached for entropy
	collapsed int // -1 if uncollapsed, otherwise the chosen tile ID
}

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
	compat [4][]bool // reusable across propagate calls
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
	compat := [4][]bool{
		make([]bool, rules.numTiles),
		make([]bool, rules.numTiles),
		make([]bool, rules.numTiles),
		make([]bool, rules.numTiles),
	}
	return &Wave{rules: rules, w: w, h: h, cells: cells, compat: compat}
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
		for d := 0; d < 4; d++ {
			for b := 0; b < wv.rules.numTiles; b++ {
				wv.compat[d][b] = false
			}
		}
		for a := 0; a < wv.rules.numTiles; a++ {
			if !src.allowed[a] {
				continue
			}
			for d := 0; d < 4; d++ {
				for b := 0; b < wv.rules.numTiles; b++ {
					if wv.rules.compatible[a][d][b] {
						wv.compat[d][b] = true
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
				if !wv.compat[d][b] {
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

// wfcCheckpoint stores a full superposition snapshot for a single cell.
type wfcCheckpoint struct {
	allowed  []bool
	count    int
	collapsed int
}

// wfcSnapshot stores the entire wave at one point in time.
type wfcSnapshot struct {
	cells [][]wfcCheckpoint
}

func (wv *Wave) saveSnapshot() wfcSnapshot {
	s := wfcSnapshot{cells: make([][]wfcCheckpoint, wv.h)}
	for y := 0; y < wv.h; y++ {
		s.cells[y] = make([]wfcCheckpoint, wv.w)
		for x := 0; x < wv.w; x++ {
			src := &wv.cells[y][x]
			cp := wfcCheckpoint{
				allowed:   append([]bool{}, src.allowed...),
				count:     src.count,
				collapsed: src.collapsed,
			}
			s.cells[y][x] = cp
		}
	}
	return s
}

func (wv *Wave) restoreSnapshot(s wfcSnapshot) {
	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			src := &s.cells[y][x]
			dst := &wv.cells[y][x]
			copy(dst.allowed, src.allowed)
			dst.count = src.count
			dst.collapsed = src.collapsed
		}
	}
}

// Solve runs the WFC observation/propagation loop with backtracking.
// Before each observation a snapshot is saved. If propagation hits a
// contradiction the snapshot is restored and a different tile is chosen for
// that cell. If no tile remains valid at that cell we backtrack one level.
// maxRestarts is a safety cap for full resets.
func (wv *Wave) Solve(rng *rand.Rand, maxRestarts int) *Wave {
	best := wv
	bestCollapsed := 0

	type btFrame struct {
		snap    wfcSnapshot
		x, y    int
		tried   []int
	}

	for restart := 0; restart <= maxRestarts; restart++ {
		if restart > 0 {
			wv = newWave(wv.rules, wv.w, wv.h)
		}

		var stack []btFrame
		contradict := false

	mainLoop:
		for {
			if contradict {
				// Backtrack: pop frames until we find a cell with an untried tile.
				for len(stack) > 0 {
					top := &stack[len(stack)-1]
					wv.restoreSnapshot(top.snap)

					// Find a valid tile at top's cell not yet tried.
					cell := &wv.cells[top.y][top.x]
				scan:
					for i := 0; i < len(cell.allowed); i++ {
						if !cell.allowed[i] {
							continue
						}
						for _, t := range top.tried {
							if i == t {
								continue scan
							}
						}
						// Found an untried valid tile.
						top.tried = append(top.tried, i)
						cell.collapseTo(i)
						if wv.propagate(top.x, top.y) {
							// Backtrack succeeded — resume forward.
							contradict = false
							continue mainLoop
						}
						// Still a contradiction — try next tile at same cell.
					}

					// No untried tile at this cell — pop frame and backtrack further.
					stack = stack[:len(stack)-1]
				}

				if len(stack) == 0 {
					contradict = false
				}
				break
			}

			x, y, ok := wv.minEntropyCell(rng)
			if !ok {
				break
			}

			snap := wv.saveSnapshot()
			wv.observe(x, y, rng)
			if !wv.propagate(x, y) {
				// Contradiction: push a backtrack frame and retry.
				stack = append(stack, btFrame{
					snap:  snap,
					x:     x,
					y:     y,
					tried: []int{wv.cells[y][x].collapsed},
				})
				contradict = true
				continue
			}
			// Successful observation — push frame so future steps can backtrack here.
			stack = append(stack, btFrame{
				snap:  snap,
				x:     x,
				y:     y,
				tried: []int{wv.cells[y][x].collapsed},
			})
		}

		count := 0
		for y := 0; y < wv.h; y++ {
			for x := 0; x < wv.w; x++ {
				if wv.cells[y][x].collapsed >= 0 {
					count++
				}
			}
		}
		if !contradict && wv.fullyCollapsed() {
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
	case 'B':
		return TileBed
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
			rows := tile.gridRows()
			cols := tile.gridCols()
			for ty := 0; ty < rows; ty++ {
				for tx := 0; tx < cols; tx++ {
					ch := tile.RuneGrid[ty][tx]
					mx := ox + gx*cols + tx
					my := oy + gy*rows + ty
					m.SetLevel(mx, my, level, tileRuneToType(ch))
				}
			}
		}
	}
}

// ufoWFCTiles defines the modular UFO piece library for the Tiled Model.
// Each piece is 3x3. '.' = floor, '#' = wall, other letters = furniture.
func hardcodedUFOTiles() []WFCTile {
		floor := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallN := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallE := [][]rune{
		{'.', '.', '#'},
		{'.', '.', '#'},
		{'.', '.', '#'},
	}
	wallS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'#', '#', '#'},
	}
	wallW := [][]rune{
		{'#', '.', '.'},
		{'#', '.', '.'},
		{'#', '.', '.'},
	}
	corridorNS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	_ = corridorNS
	cornerNE := [][]rune{
		{'#', '#', '#'},
		{'#', '.', '.'},
		{'.', '.', '.'},
	}
	cornerSE := [][]rune{
		{'.', '.', '.'},
		{'#', '.', '.'},
		{'#', '#', '#'},
	}
	cornerSW := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '#'},
		{'#', '#', '#'},
	}
	cornerNW := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '#'},
		{'.', '.', '.'},
	}
	engine := [][]rune{
		{'.', '#', '.'},
		{'#', 'M', '#'},
		{'.', '#', '.'},
	}
	consoleRoom := [][]rune{
		{'.', '.', '.'},
		{'.', 'C', '.'},
		{'.', '.', '.'},
	}
	podRoom := [][]rune{
		{'.', 'P', '.'},
		{'.', '.', '.'},
		{'.', 'P', '.'},
	}
	powerCore := [][]rune{
		{'#', '.', '#'},
		{'.', 'X', '.'},
		{'#', '.', '#'},
	}
	doorN := [][]rune{
		{'.', 'D', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	doorE := [][]rune{
		{'.', '.', '.'},
		{'.', '.', 'D'},
		{'.', '.', '.'},
	}
	doorS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', 'D', '.'},
	}
	doorW := [][]rune{
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

// wfcDirFromKey maps a JSON direction key to the internal direction index.
var wfcDirFromKey = map[string]int{"N": dirN, "E": dirE, "S": dirS, "W": dirW}

// wfcTilesFromLib converts a parsed JSON WFC library into engine tiles.
func wfcTilesFromLib(lib *mapgen.WFCLibrary) []WFCTile {
	tiles := make([]WFCTile, 0, len(lib.Tiles))
	for _, d := range lib.Tiles {
		grid := make([][]rune, len(d.Rows))
		for i, row := range d.Rows {
			grid[i] = []rune(row)
		}
		nb := [4][]int{}
		for key, ids := range d.Neighbors {
			dir := wfcDirFromKey[key]
			nb[dir] = append([]int{}, ids...)
		}
		tiles = append(tiles, WFCTile{ID: d.ID, Name: d.Name, RuneGrid: grid, Neighbors: nb})
	}
	return tiles
}

// wfcSearchPaths lists where WFC library JSON may live (binary working dir or
// repo-root relative). The first existing file wins.
var wfcSearchPaths = []string{"data/wfc", "../data/wfc", "../../data/wfc"}

// loadWFCLibrary tries to load a WFC tile library by stem (e.g. "ufo") from
// data/wfc/<stem>.json. On any error it returns nil so callers can fall back
// to a hardcoded library, keeping generation working without the JSON files.
func loadWFCLibrary(stem string) []WFCTile {
	for _, dir := range wfcSearchPaths {
		path := dir + "/" + stem + ".json"
		if lib, err := mapgen.LoadWFCLibrary(path); err == nil {
			return wfcTilesFromLib(lib)
		}
	}
	return nil
}

// ufoWFCTiles loads the UFO WFC tile library from data/wfc/ufo.json, falling
// back to the hardcoded library if the file is unavailable.
func ufoWFCTiles() []WFCTile {
	if t := loadWFCLibrary("ufo"); t != nil {
		return t
	}
	return hardcodedUFOTiles()
}

// urbanWFCTiles loads the urban WFC tile library from data/wfc/urban.json,
// falling back to the hardcoded library if the file is unavailable.
func urbanWFCTiles() []WFCTile {
	if t := loadWFCLibrary("urban"); t != nil {
		return t
	}
	return hardcodedUrbanTiles()
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

// urbanWFCTiles defines a modular tile library for procedural urban buildings.
// It mixes small 3x3 pieces (rooms, walls, corners, doors, furniture) with a
// few LARGE multi-room blocks (6x6 and 9x9) so the solver can emit whole
// building wings in a single collapsed cell. '.' = floor, '#' = wall,
// letters = furniture, 'D' = door.
func hardcodedUrbanTiles() []WFCTile {
	// --- Small 3x3 pieces ---
	floor := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallN := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallE := [][]rune{
		{'.', '.', '#'},
		{'.', '.', '#'},
		{'.', '.', '#'},
	}
	wallS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'#', '#', '#'},
	}
	wallW := [][]rune{
		{'#', '.', '.'},
		{'#', '.', '.'},
		{'#', '.', '.'},
	}
	cornerNE := [][]rune{
		{'#', '#', '#'},
		{'#', '.', '.'},
		{'.', '.', '.'},
	}
	cornerSE := [][]rune{
		{'.', '.', '.'},
		{'#', '.', '.'},
		{'#', '#', '#'},
	}
	cornerSW := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '#'},
		{'#', '#', '#'},
	}
	cornerNW := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '#'},
		{'.', '.', '.'},
	}
	doorN := [][]rune{
		{'.', 'D', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	doorE := [][]rune{
		{'.', '.', '.'},
		{'.', '.', 'D'},
		{'.', '.', '.'},
	}
	doorS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', 'D', '.'},
	}
	doorW := [][]rune{
		{'.', '.', '.'},
		{'D', '.', '.'},
		{'.', '.', '.'},
	}
	roomOffice := [][]rune{
		{'.', '.', '.'},
		{'.', 'C', '.'},
		{'.', '.', '.'},
	}
	roomBed := [][]rune{
		{'.', '.', '.'},
		{'.', 'B', '.'},
		{'.', '.', '.'},
	}
	roomStorage := [][]rune{
		{'.', '.', '.'},
		{'.', 'S', '.'},
		{'.', '.', '.'},
	}

	// --- Large 6x6 two-room apartment block (interior walls split it). ---
	apartment6 := [][]rune{
		{'#', '#', '#', '#', '#', '#'},
		{'#', '.', '.', '.', '.', '#'},
		{'#', '.', '#', '#', '.', '#'},
		{'#', '.', '#', '#', '.', '#'},
		{'#', '.', '.', '.', '.', '#'},
		{'#', '#', '#', '#', '#', '#'},
	}
	// --- Large 6x6 warehouse with machinery and a door on the south wall. ---
	warehouse6 := [][]rune{
		{'#', '#', '#', '#', '#', '#'},
		{'#', 'M', '.', '.', 'M', '#'},
		{'#', '.', '.', '.', '.', '#'},
		{'#', '.', '.', '.', '.', '#'},
		{'#', 'M', '.', '.', 'M', '#'},
		{'#', '#', '.', '.', '#', '#'},
	}
	// --- Large 9x9 office wing: open-plan floor with a central meeting room
	// and furniture clusters, walled on the perimeter. ---
	office9 := [][]rune{
		{'#', '#', '#', '#', '#', '#', '#', '#', '#'},
		{'#', '.', '.', '.', '.', '.', '.', '.', '#'},
		{'#', '.', 'C', '.', '.', '.', 'C', '.', '#'},
		{'#', '.', '.', '.', '#', '.', '.', '.', '#'},
		{'#', '.', '.', '.', '#', '.', '.', '.', '#'},
		{'#', '.', 'C', '.', '#', '.', 'C', '.', '#'},
		{'#', '.', '.', '.', '.', '.', '.', '.', '#'},
		{'#', '.', '.', '.', '.', '.', '.', 'D', '#'},
		{'#', '#', '#', '#', '#', '#', '#', '#', '#'},
	}
	// --- Large 9x9 barracks: four small bunk rooms around a corridor. ---
	barracks9 := [][]rune{
		{'#', '#', '#', '#', '#', '#', '#', '#', '#'},
		{'#', 'B', '.', '#', '.', '#', 'B', '.', '#'},
		{'#', '.', '.', '#', '.', '#', '.', '.', '#'},
		{'#', '#', '.', '#', '.', '#', '#', '.', '#'},
		{'.', '.', '.', '.', '.', '.', '.', '.', '.'},
		{'#', '#', '.', '#', '.', '#', '#', '.', '#'},
		{'#', '.', '.', '#', '.', '#', '.', '.', '#'},
		{'#', 'B', '.', '#', '.', '#', 'B', '.', '#'},
		{'#', '#', '#', '#', '#', '#', '#', '#', '#'},
	}

	tiles := []WFCTile{
		{ID: 0, Name: "Floor", RuneGrid: floor},
		{ID: 1, Name: "WallN", RuneGrid: wallN},
		{ID: 2, Name: "WallE", RuneGrid: wallE},
		{ID: 3, Name: "WallS", RuneGrid: wallS},
		{ID: 4, Name: "WallW", RuneGrid: wallW},
		{ID: 5, Name: "CornerNE", RuneGrid: cornerNE},
		{ID: 6, Name: "CornerSE", RuneGrid: cornerSE},
		{ID: 7, Name: "CornerSW", RuneGrid: cornerSW},
		{ID: 8, Name: "CornerNW", RuneGrid: cornerNW},
		{ID: 9, Name: "DoorN", RuneGrid: doorN},
		{ID: 10, Name: "DoorE", RuneGrid: doorE},
		{ID: 11, Name: "DoorS", RuneGrid: doorS},
		{ID: 12, Name: "DoorW", RuneGrid: doorW},
		{ID: 13, Name: "RoomOffice", RuneGrid: roomOffice},
		{ID: 14, Name: "RoomBed", RuneGrid: roomBed},
		{ID: 15, Name: "RoomStorage", RuneGrid: roomStorage},
		// Large multi-room blocks.
		{ID: 16, Name: "Apartment6", RuneGrid: apartment6},
		{ID: 17, Name: "Warehouse6", RuneGrid: warehouse6},
		{ID: 18, Name: "Office9", RuneGrid: office9},
		{ID: 19, Name: "Barracks9", RuneGrid: barracks9},
	}

	// Open tiles may sit adjacent to anything (they present floor/walls on
	// their perimeter, so the enclosing/border logic handles closure).
	open := []int{0, 9, 10, 11, 12, 13, 14, 15}
	// Wall/corner pieces (solid perimeter) — used to close building edges.
	structural := []int{1, 2, 3, 4, 5, 6, 7, 8, 16, 17, 18, 19}

	// Floor: open on all sides (adjacent to walls, corners, rooms, doors).
	for d := 0; d < 4; d++ {
		tiles[0].Neighbors[d] = append([]int{}, open...)
	}
	// Walls: solid side faces structural, open side faces open.
	tiles[1].Neighbors[dirN] = append([]int{}, structural...)
	tiles[1].Neighbors[dirS] = append([]int{}, open...)
	tiles[1].Neighbors[dirE] = append([]int{}, structural...)
	tiles[1].Neighbors[dirW] = append([]int{}, structural...)
	tiles[2].Neighbors[dirE] = append([]int{}, structural...)
	tiles[2].Neighbors[dirW] = append([]int{}, open...)
	tiles[2].Neighbors[dirN] = append([]int{}, structural...)
	tiles[2].Neighbors[dirS] = append([]int{}, structural...)
	tiles[3].Neighbors[dirS] = append([]int{}, structural...)
	tiles[3].Neighbors[dirN] = append([]int{}, open...)
	tiles[3].Neighbors[dirE] = append([]int{}, structural...)
	tiles[3].Neighbors[dirW] = append([]int{}, structural...)
	tiles[4].Neighbors[dirW] = append([]int{}, structural...)
	tiles[4].Neighbors[dirE] = append([]int{}, open...)
	tiles[4].Neighbors[dirN] = append([]int{}, structural...)
	tiles[4].Neighbors[dirS] = append([]int{}, structural...)

	cornerNbrs := func() [4][]int {
		return [4][]int{
			append([]int{}, structural...),
			append([]int{}, structural...),
			append([]int{}, structural...),
			append([]int{}, structural...),
		}
	}
	tiles[5].Neighbors = cornerNbrs()
	tiles[6].Neighbors = cornerNbrs()
	tiles[7].Neighbors = cornerNbrs()
	tiles[8].Neighbors = cornerNbrs()

	// Doors and rooms: open on all sides.
	for _, id := range []int{9, 10, 11, 12, 13, 14, 15} {
		for d := 0; d < 4; d++ {
			tiles[id].Neighbors[d] = append([]int{}, open...)
		}
	}
	// Large multi-room blocks: their perimeter is wall, so they connect to
	// structural pieces (walls/corners/other blocks) on all sides. This keeps
	// the building closed while letting big wings tile together.
	for _, id := range []int{16, 17, 18, 19} {
		for d := 0; d < 4; d++ {
			tiles[id].Neighbors[d] = append([]int{}, structural...)
		}
	}

	return tiles
}

// GenerateUrbanBuildingWFC builds an urban building interior map using the WFC
// Tiled Model. Small 3x3 pieces combine with large multi-room blocks (6x6 and
// 9x9) to produce varied building layouts. rng must be seeded for reproducibility.
func GenerateUrbanBuildingWFC(w, h int, rng *rand.Rand) *BattleMap {
	return GenerateUrbanBuildingWFCLevels(w, h, 1, rng)
}

// GenerateUrbanBuildingWFCLevels builds an urban building with the specified
// number of floors. Floors are generated independently using the same WFC tile
// set; stairs connect each pair of adjacent levels.
func GenerateUrbanBuildingWFCLevels(w, h, numLevels int, rng *rand.Rand) *BattleMap {
	if numLevels < 1 {
		numLevels = 1
	}

	m := NewMultiLevelBattleMap(w, h, numLevels)

	rules := NewWFCRules(urbanWFCTiles())

	cell := 9
	gw := w / cell
	gh := h / cell
	if gw < 1 {
		gw = 1
	}
	if gh < 1 {
		gh = 1
	}

	for level := 0; level < numLevels; level++ {
		// Fill with pavement (y is level-relative).
		m.fillRectLevel(0, 0, w, h, level, TilePavement)

		// Run WFC with a seeded offset so each level gets a different layout.
		levelRng := rand.New(rand.NewSource(int64(rng.Int())))
		wv := newWave(rules, gw, gh)
		wv = wv.Solve(levelRng, 30)
		wv.CompileToBattleMap(m, 0, 0, level)

		// Enclose with a perimeter wall (y is level-relative).
		m.drawRectLevel(0, 0, w, h, level, TileWall)
	}

	// Add stairs between levels at a random interior position.
	for level := 0; level < numLevels-1; level++ {
		sx := 1 + rng.Intn(w-2)
		sy := 1 + rng.Intn(h-2)
		m.SetLevel(sx, sy, level, TileStairsDown)
		m.SetLevel(sx, sy, level+1, TileStairs)
	}

	return m
}

// alienBaseWFCTiles loads the alien base WFC tile library from
// data/wfc/alien_base.json, falling back to hardcoded tiles.
func alienBaseWFCTiles() []WFCTile {
	if t := loadWFCLibrary("alien_base"); t != nil {
		return t
	}
	return hardcodedAlienBaseTiles()
}

// GenerateAlienBaseWFC builds an alien base interior map using WFC.
// The base has 2 levels with alien-themed rooms (consoles, machinery,
// containment pods, power sources, alien tech). Stairs connect the levels.
func GenerateAlienBaseWFC(w, h int, rng *rand.Rand) *BattleMap {
	levelH := h / 2
	if levelH < 12 {
		levelH = 12
	}
	m := NewMultiLevelBattleMap(w, levelH, 2)

	rules := NewWFCRules(alienBaseWFCTiles())

	gw := w / 3
	gh := levelH / 3
	if gw < 1 {
		gw = 1
	}
	if gh < 1 {
		gh = 1
	}

	buildLevel := func(level int) {
		m.fillRectLevel(0, 0, w, levelH, level, TileUFOFloor)

		levelRng := rand.New(rand.NewSource(int64(rng.Int())))
		wv := newWave(rules, gw, gh)
		wv = wv.Solve(levelRng, 20)
		wv.CompileToBattleMap(m, 0, 0, level)

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

// hardcodedAlienBaseTiles provides an embedded alien base tile library
// used when data/wfc/alien_base.json cannot be loaded.
func hardcodedAlienBaseTiles() []WFCTile {
	floor := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallN := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '.'},
		{'.', '.', '.'},
	}
	wallE := [][]rune{
		{'.', '.', '#'},
		{'.', '.', '#'},
		{'.', '.', '#'},
	}
	wallS := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '.'},
		{'#', '#', '#'},
	}
	wallW := [][]rune{
		{'#', '.', '.'},
		{'#', '.', '.'},
		{'#', '.', '.'},
	}
	cornerNE := [][]rune{
		{'#', '#', '#'},
		{'#', '.', '.'},
		{'.', '.', '.'},
	}
	cornerSE := [][]rune{
		{'.', '.', '.'},
		{'#', '.', '.'},
		{'#', '#', '#'},
	}
	cornerSW := [][]rune{
		{'.', '.', '.'},
		{'.', '.', '#'},
		{'#', '#', '#'},
	}
	cornerNW := [][]rune{
		{'#', '#', '#'},
		{'.', '.', '#'},
		{'.', '.', '.'},
	}
	consoleRoom := [][]rune{
		{'.', 'C', '.'},
		{'C', 'C', 'C'},
		{'.', '.', '.'},
	}
	consoleRoom90 := [][]rune{
		{'.', 'C', '.'},
		{'.', 'C', '.'},
		{'.', 'C', '.'},
	}
	machineryNW := [][]rune{
		{'M', '.', '.'},
		{'.', '#', '.'},
		{'.', '.', '.'},
	}
	machinerySE := [][]rune{
		{'.', '.', '.'},
		{'.', '#', '.'},
		{'.', '.', 'M'},
	}
	podRoom := [][]rune{
		{'P', 'P', 'P'},
		{'.', '.', '.'},
		{'P', 'P', 'P'},
	}
	powerRoom := [][]rune{
		{'.', 'S', '.'},
		{'S', 'S', 'S'},
		{'.', '#', '.'},
	}
	alienTechRoom := [][]rune{
		{'T', '.', 'T'},
		{'.', '.', '.'},
		{'T', '.', 'T'},
	}
	storageRoom := [][]rune{
		{'S', '.', '.'},
		{'.', '.', '.'},
		{'.', '.', 'S'},
	}
	corridorT := [][]rune{
		{'#', '#', '#'},
		{'.', '#', '.'},
		{'.', '#', '.'},
	}
	corridorB := [][]rune{
		{'.', '#', '.'},
		{'.', '#', '.'},
		{'#', '#', '#'},
	}
	corridorL := [][]rune{
		{'.', '#', '.'},
		{'#', '#', '#'},
		{'.', '#', '.'},
	}
	corridorR := [][]rune{
		{'#', '.', '.'},
		{'#', '.', '.'},
		{'#', '.', '.'},
	}

	return []WFCTile{
		{ID: 0, Name: "Floor", RuneGrid: floor},
		{ID: 1, Name: "WallN", RuneGrid: wallN},
		{ID: 2, Name: "WallE", RuneGrid: wallE},
		{ID: 3, Name: "WallS", RuneGrid: wallS},
		{ID: 4, Name: "WallW", RuneGrid: wallW},
		{ID: 5, Name: "CornerNE", RuneGrid: cornerNE},
		{ID: 6, Name: "CornerSE", RuneGrid: cornerSE},
		{ID: 7, Name: "CornerSW", RuneGrid: cornerSW},
		{ID: 8, Name: "CornerNW", RuneGrid: cornerNW},
		{ID: 9, Name: "ConsoleRoom", RuneGrid: consoleRoom},
		{ID: 10, Name: "ConsoleRoom90", RuneGrid: consoleRoom90},
		{ID: 11, Name: "MachineryNW", RuneGrid: machineryNW},
		{ID: 12, Name: "MachinerySE", RuneGrid: machinerySE},
		{ID: 13, Name: "PodRoom", RuneGrid: podRoom},
		{ID: 14, Name: "PowerRoom", RuneGrid: powerRoom},
		{ID: 15, Name: "AlienTechRoom", RuneGrid: alienTechRoom},
		{ID: 16, Name: "StorageRoom", RuneGrid: storageRoom},
		{ID: 17, Name: "CorridorT", RuneGrid: corridorT},
		{ID: 18, Name: "CorridorB", RuneGrid: corridorB},
		{ID: 19, Name: "CorridorL", RuneGrid: corridorL},
		{ID: 20, Name: "CorridorR", RuneGrid: corridorR},
	}
}
