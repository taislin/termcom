package battle

import (
	"log"
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
	// Alien base display runes
	case '⌸', '⊕', 'ꙮ', '☼', '◊', '╨':
		return TileAlienTech
	case '╓', '⊚', '╖':
		return TileConsole
	case '◢', '═', '◣', '║', '⍾', '◥', '◤':
		return TilePowerSource
	case '⊏', '≈', '⊐':
		return TileMachinery
	default:
		return TileUFOFloor
	}
}

// tileRuneToDisplay returns a display rune to store on the tile when the WFC
// rune is a visual character (Unicode BMP > U+00FF) rather than a simple type
// marker. It returns 0 for ASCII type markers so the standard tile glyph
// pipeline handles them.
func tileRuneToDisplay(ch rune) rune {
	switch ch {
	case '#', '.', 'D', 'C', 'M', 'P', 'X', 'S', 'B', 'A', 'T':
		return 0
	}
	if ch > 0xFF {
		return ch
	}
	return 0
}

// fillGaps replaces stray base-terrain tiles that were not covered by WFC
// (due to integer division at the right/bottom borders) with the most common
// neighbor tile type. This fixes uncovered border strips. WfcW and wfcH are the
// boundaries of the compiled WFC grid (gw * stride, gh * stride); tiles inside
// these boundaries are left untouched to preserve WFC layout.
func (m *BattleMap) fillGaps(level int, fillTile TileType, wfcW, wfcH int) {
	for y := 0; y < m.LevelHeight; y++ {
		for x := 0; x < m.Width; x++ {
			// Only fill gaps in the remainder strips outside the WFC grid
			if x < wfcW && y < wfcH {
				continue
			}
			t := m.AtLevel(x, y, level)
			if t.Type != fillTile {
				continue
			}
			counts := map[TileType]int{}
			dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
			for _, d := range dirs {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
					continue
				}
				nt := m.AtLevel(nx, ny, level)
				if nt.Type != fillTile {
					counts[nt.Type]++
				}
			}
			best, bestCount := fillTile, 0
			for tt, c := range counts {
				if c > bestCount {
					best, bestCount = tt, c
				}
			}
			if best != fillTile {
				m.SetLevel(x, y, level, best)
			}
		}
	}
}

// CompileToBattleMap stamps the collapsed wave into the destination BattleMap
// at the given level, offset by (ox, oy). stride is the fixed pixel size of
// each wave grid cell (e.g. 9 for urban, 3 for UFO). If stride is 0, it
// defaults to the tile's own dimensions (legacy behavior for uniform tiles).
func (wv *Wave) CompileToBattleMap(m *BattleMap, ox, oy, level, stride int) {
	for gy := 0; gy < wv.h; gy++ {
		for gx := 0; gx < wv.w; gx++ {
			id := wv.cells[gy][gx].collapsed
			if id < 0 {
				continue
			}
			tile := wv.rules.Tiles[id]
			rows := tile.gridRows()
			cols := tile.gridCols()
			s := stride
			if s <= 0 {
				s = cols
			}
			for ty := 0; ty < rows; ty++ {
				for tx := 0; tx < cols; tx++ {
					ch := tile.RuneGrid[ty][tx]
					mx := ox + gx*s + tx
					my := oy + gy*s + ty
					tt := tileRuneToType(ch)
					m.SetLevel(mx, my, level, tt)
					if dr := tileRuneToDisplay(ch); dr != 0 {
						ary := my + level*m.LevelHeight
						if mx >= 0 && mx < m.Width && ary >= 0 && ary < m.Height {
							m.Tiles[ary][mx].Rune = dr
						}
					}
				}
			}
		}
	}
}

// wfcDirFromKey maps a JSON direction key to the internal direction index.
var wfcDirFromKey = map[string]int{"N": dirN, "E": dirE, "S": dirS, "W": dirW}

// wfcOpposite returns the direction index opposite to d (N<->S, E<->W).
var wfcOpposite = [4]int{dirS, dirW, dirN, dirE}

// wfcEdgeClass maps a tile rune to a simplified class for edge matching.
// '.' and furniture characters count as "open"; '#' counts as "wall".
func wfcEdgeClass(ch rune) byte {
	if ch == '#' {
		return '#'
	}
	return '.'
}

// extractEdge returns the edge string of tile t in direction d.
// N: top row left→right; S: bottom row left→right;
// E: rightmost column top→bottom; W: leftmost column top→bottom.
func extractEdge(t WFCTile, d int) string {
	rows := t.gridRows()
	cols := t.gridCols()
	if rows == 0 || cols == 0 {
		return ""
	}
	buf := make([]byte, 0, rows)
	switch d {
	case dirN:
		for _, ch := range t.RuneGrid[0] {
			buf = append(buf, wfcEdgeClass(ch))
		}
	case dirS:
		for _, ch := range t.RuneGrid[rows-1] {
			buf = append(buf, wfcEdgeClass(ch))
		}
	case dirE:
		for r := 0; r < rows; r++ {
			buf = append(buf, wfcEdgeClass(t.RuneGrid[r][cols-1]))
		}
	case dirW:
		for r := 0; r < rows; r++ {
			buf = append(buf, wfcEdgeClass(t.RuneGrid[r][0]))
		}
	}
	return string(buf)
}

// tilesCompatible returns true if tile a can legally sit in direction d
// relative to tile b (i.e. a's d-edge matches b's opposite edge).
// Tiles of different sizes are never compatible.
func tilesCompatible(a, b WFCTile, d int) bool {
	if a.gridRows() != b.gridRows() || a.gridCols() != b.gridCols() {
		return false
	}
	return extractEdge(a, d) == extractEdge(b, wfcOpposite[d])
}

// autoComputeNeighbors fills the Neighbors field for every tile in the slice
// that has an empty neighbor list. It uses pixel-level edge matching so that
// the JSON data files do not need hand-authored neighbor lists.
func autoComputeNeighbors(tiles []WFCTile) []WFCTile {
	for i := range tiles {
		needsAuto := false
		for d := 0; d < 4; d++ {
			if len(tiles[i].Neighbors[d]) == 0 {
				needsAuto = true
				break
			}
		}
		if !needsAuto {
			continue
		}
		for d := 0; d < 4; d++ {
			if len(tiles[i].Neighbors[d]) > 0 {
				continue
			}
			var nb []int
			for j := range tiles {
				if tilesCompatible(tiles[i], tiles[j], d) {
					nb = append(nb, tiles[j].ID)
				}
			}
			tiles[i].Neighbors[d] = nb
		}
	}
	return tiles
}

// wfcTilesFromLib converts a parsed JSON WFC library into engine tiles.
// If any tile has empty neighbor lists, autoComputeNeighbors fills them via
// pixel-level edge matching.
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
	tiles = autoComputeNeighbors(tiles)
	return tiles
}

// wfcSearchPaths lists where WFC library JSON may live (binary working dir or
// repo-root relative). The first existing file wins.
var wfcSearchPaths = []string{"data/wfc", "../data/wfc", "../../data/wfc"}

// loadWFCLibrary loads a WFC tile library by stem (e.g. "ufo") from
// data/wfc/<stem>.json. It panics if the file cannot be found or parsed.
func loadWFCLibrary(stem string) []WFCTile {
	for _, dir := range wfcSearchPaths {
		path := dir + "/" + stem + ".json"
		if lib, err := mapgen.LoadWFCLibrary(path); err == nil {
			return wfcTilesFromLib(lib)
		}
	}
	log.Fatalf("wfc: required library %q not found in any search path: %v", stem, wfcSearchPaths)
	return nil
}

// ufoWFCTiles loads the UFO WFC tile library from data/wfc/ufo.json.
func ufoWFCTiles() []WFCTile {
	return loadWFCLibrary("ufo")
}

// urbanWFCTiles loads the urban WFC tile library from data/wfc/urban.json.
func urbanWFCTiles() []WFCTile {
	return loadWFCLibrary("urban")
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

	// Number of tile-grid cells that fit on one level.
	// Use 6x6 cells to accommodate both 3x3 and 6x6 tiles.
	cell := 6
	gw := w / cell
	gh := levelH / cell
	if gw < 1 {
		gw = 1
	}
	if gh < 1 {
		gh = 1
	}

	buildLevel := func(level int) {
		// Fill level with UFO floor base, then stamp collapsed WFC tiles.
		m.fillRectLevel(0, 0, w, levelH, level, TileUFOFloor)

		levelRng := rand.New(rand.NewSource(int64(rng.Int())))
		wv := newWave(rules, gw, gh)
		wv = wv.Solve(levelRng, 20)
		wv.CompileToBattleMap(m, 0, 0, level, 6)
		m.fillGaps(level, TileUFOFloor, gw*6, gh*6)

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

// GenerateUrbanBuildingWFC builds an urban building interior map using the WFC
// Tiled Model on a grid of uniform 9x9 modular pieces whose neighbor rules are
// auto-computed from pixel-edge matching. rng must be seeded for reproducibility.
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
		// Fill with pavement as base; WFC will overwrite the interior.
		m.fillRectLevel(0, 0, w, h, level, TilePavement)

		// Run WFC with a seeded offset so each level gets a different layout.
		levelRng := rand.New(rand.NewSource(int64(rng.Int())))
		wv := newWave(rules, gw, gh)
		wv = wv.Solve(levelRng, 50)
		wv.CompileToBattleMap(m, 0, 0, level, 9)

		// The WFC rune decoder uses UFO types for '.' and '#' by default.
		// Convert them to building-appropriate types for urban interiors.
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				switch m.AtLevel(x, y, level).Type {
				case TileUFOFloor:
					m.SetLevel(x, y, level, TileFloor)
				case TileUFOWall:
					m.SetLevel(x, y, level, TileWall)
				}
			}
		}

		m.fillGaps(level, TilePavement, gw*9, gh*9)

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
// data/wfc/alien_base.json.
func alienBaseWFCTiles() []WFCTile {
	return loadWFCLibrary("alien_base")
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
		wv.CompileToBattleMap(m, 0, 0, level, 3)
		m.fillGaps(level, TileUFOFloor, gw*3, gh*3)

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

