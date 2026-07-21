package battle

import (
	"fmt"
	"log"
	"math"
	"math/bits"
	"math/rand"
	"sort"

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
// Weight controls selection frequency during observation (1.0 = default).
type WFCTile struct {
	ID       int
	Name     string
	RuneGrid [][]rune
	Weight   float64
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
// allowed is a bitmask where bit n is set when tile id n is still valid here.
type superposition struct {
	allowed   uint64
	count     int    // number of set bits; cached for entropy
	collapsed int    // -1 if uncollapsed, otherwise the chosen tile ID
}

func (s *superposition) collapseTo(id int) {
	s.allowed = 1 << uint(id)
	s.count = 1
	s.collapsed = id
}

// WFCRules holds the tile set and a fast adjacency matrix.
// FloorTile and WallTile specify which TileType to use for '.' and '#' runes
// when compiling to a BattleMap (defaults TileUFOFloor/TileUFOWall).
type WFCRules struct {
	Tiles     []WFCTile
	numTiles  int
	FloorTile TileType
	WallTile  TileType
	// compatible[a][d] is a bitmask where bit b is set if tile b may sit in
	// direction d of tile a.
	compatible [][4]uint64
}

// NewWFCRules validates and precomputes the compatibility matrix from tiles.
// Tile IDs must be contiguous 0..n-1 because the compat matrix is indexed by
// ID directly. A mismatch panics.
// produces a silently incorrect adjacency matrix.
func NewWFCRules(tiles []WFCTile) *WFCRules {
	n := len(tiles)
	for i, t := range tiles {
		if t.ID != i {
			panic(fmt.Sprintf("wfc: tile %q has id %d but expected %d; adjacency matrix would be mis-indexed", t.Name, t.ID, i))
		}
	}
	r := &WFCRules{Tiles: tiles, numTiles: n, FloorTile: TileUFOFloor, WallTile: TileUFOWall}
	r.compatible = make([][4]uint64, n)
	for a := 0; a < n; a++ {
		for d := 0; d < 4; d++ {
			var mask uint64
			for _, b := range tiles[a].Neighbors[d] {
				if b >= 0 && b < n {
					mask |= 1 << uint(b)
				}
			}
			r.compatible[a][d] = mask
		}
	}
	return r
}

// Wave is the 2D grid of superpositions being collapsed.
type Wave struct {
	rules  *WFCRules
	w, h   int
	cells  [][]superposition
	compat [4]uint64 // reusable scratch across propagate calls
}

func newWave(rules *WFCRules, w, h int) *Wave {
	cells := make([][]superposition, h)
	numTiles := rules.numTiles
	var fullMask uint64
	if numTiles >= 64 {
		fullMask = ^uint64(0)
	} else {
		fullMask = (1 << uint(numTiles)) - 1
	}
	for y := 0; y < h; y++ {
		cells[y] = make([]superposition, w)
		for x := 0; x < w; x++ {
			cells[y][x] = superposition{allowed: fullMask, count: numTiles, collapsed: -1}
		}
	}
	return &Wave{rules: rules, w: w, h: h, cells: cells}
}

// cellEntropy returns the weighted Shannon entropy for an uncollapsed cell.
// H = -Σ p_i · log(p_i) where p_i = weight_i / totalWeight.
// Returns 0 for cells with count <= 1.
func (wv *Wave) cellEntropy(s *superposition) float64 {
	if s.count <= 1 {
		return 0
	}
	var totalWeight float64
	mask := s.allowed
	for mask != 0 {
		b := bits.TrailingZeros64(mask)
		w := wv.rules.Tiles[b].Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
		mask &^= 1 << uint(b)
	}
	if totalWeight <= 0 {
		return 0
	}
	var h float64
	mask = s.allowed
	for mask != 0 {
		b := bits.TrailingZeros64(mask)
		w := wv.rules.Tiles[b].Weight
		if w <= 0 {
			w = 1
		}
		p := w / totalWeight
		h -= p * math.Log(p)
		mask &^= 1 << uint(b)
	}
	return h
}

// minEntropyCell returns the coordinates of the uncollapsed cell with the
// lowest weighted Shannon entropy, plus whether any such cell exists and the
// wave is contradiction-free. A cell with count==0 means an impossible
// constraint, so we return false to signal external backtracking.
func (wv *Wave) minEntropyCell(rng *rand.Rand) (int, int, bool) {
	type entry struct {
		x, y    int
		entropy float64
	}
	var cells []entry

	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			c := &wv.cells[y][x]
			if c.collapsed >= 0 {
				continue
			}
			if c.count == 0 {
				return 0, 0, false
			}
			e := wv.cellEntropy(c)
			if len(cells) == 0 || e < cells[0].entropy {
				cells = append(cells[:0], entry{x, y, e})
			} else if e == cells[0].entropy {
				cells = append(cells, entry{x, y, e})
			}
		}
	}
	if len(cells) == 0 {
		return 0, 0, false
	}
	chosen := cells[rng.Intn(len(cells))]
	return chosen.x, chosen.y, true
}

// observe collapses the lowest-entropy cell by randomly choosing one of its
// remaining valid tiles, weighted by each tile's Weight field. Returns the
// chosen tile ID, or -1 if no options exist.
func (wv *Wave) observe(x, y int, rng *rand.Rand) int {
	c := &wv.cells[y][x]
	options := make([]int, 0, c.count)
	weights := make([]float64, 0, c.count)
	mask := c.allowed
	for mask != 0 {
		b := bits.TrailingZeros64(mask)
		options = append(options, b)
		w := wv.rules.Tiles[b].Weight
		if w <= 0 {
			w = 1
		}
		weights = append(weights, w)
		mask &^= 1 << uint(b)
	}
	if len(options) == 0 {
		return -1
	}
	var chosen int
	if len(options) == 1 {
		chosen = options[0]
	} else {
		total := 0.0
		for _, w := range weights {
			total += w
		}
		pick := rng.Float64() * total
		accum := 0.0
		for i, id := range options {
			accum += weights[i]
			if pick < accum {
				chosen = id
				break
			}
		}
	}
	c.collapseTo(chosen)
	return chosen
}

// propagate enforces constraints from (x,y) outward using a queue. It returns
// false if any cell is reduced to zero options (contradiction), plus the list
// of cell diffs needed to revert propagation's changes on backtrack.
func (wv *Wave) propagate(x, y int) (bool, []cellDiff) {
	queue := make([]int, 0, wv.w*wv.h)
	queue = append(queue, x, y)
	var diffs []cellDiff

	for len(queue) > 0 {
		cx := queue[0]
		cy := queue[1]
		queue = queue[2:]

		src := &wv.cells[cy][cx]

		if src.collapsed >= 0 {
			a := src.collapsed
			for d := 0; d < 4; d++ {
				wv.compat[d] = wv.rules.compatible[a][d]
			}
		} else {
			for d := 0; d < 4; d++ {
				wv.compat[d] = 0
			}
			aMask := src.allowed
			for aMask != 0 {
				a := bits.TrailingZeros64(aMask)
				for d := 0; d < 4; d++ {
					wv.compat[d] |= wv.rules.compatible[a][d]
				}
				aMask &^= 1 << uint(a)
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
			oldAllowed := nb.allowed
			nb.allowed &= wv.compat[d]
			if nb.allowed != oldAllowed {
				diffs = append(diffs, cellDiff{
					x: nx, y: ny,
					oldAllowed:   oldAllowed,
					oldCount:     nb.count,
					oldCollapsed: nb.collapsed,
				})
				nb.count = bits.OnesCount64(nb.allowed)
				if nb.count == 0 {
					return false, diffs
				}
				queue = append(queue, nx, ny)
			}
		}
	}
	return true, diffs
}

// hasContradiction reports whether any uncollapsed cell has zero remaining
// options. Such a wave cannot be completed.
func (wv *Wave) hasContradiction() bool {
	for y := 0; y < wv.h; y++ {
		for x := 0; x < wv.w; x++ {
			if wv.cells[y][x].collapsed < 0 && wv.cells[y][x].count == 0 {
				return true
			}
		}
	}
	return false
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

// cellDiff records a single cell modification so it can be reverted on
// backtrack without copying the full wave grid.
type cellDiff struct {
	x, y         int
	oldAllowed   uint64
	oldCount     int
	oldCollapsed int
}

// wfcSnapshot stores the state of a single cell before observation. The
// corresponding cellDiff list from propagate is kept alongside in the
// backtrack frame and applied together to restore the wave.
type wfcSnapshot struct {
	x, y int
	cell  superposition
}

func (wv *Wave) saveSnapshot(x, y int) wfcSnapshot {
	return wfcSnapshot{x: x, y: y, cell: wv.cells[y][x]}
}

func (wv *Wave) restoreSnapshot(s wfcSnapshot, diffs []cellDiff) {
	wv.cells[s.y][s.x] = s.cell
	for _, d := range diffs {
		wv.cells[d.y][d.x].allowed = d.oldAllowed
		wv.cells[d.y][d.x].count = d.oldCount
		wv.cells[d.y][d.x].collapsed = d.oldCollapsed
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
		snap  wfcSnapshot
		x, y  int
		tried []int
		diffs []cellDiff
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

				scan:
					for mask := top.snap.cell.allowed; mask != 0; {
						i := bits.TrailingZeros64(mask)
						mask &^= 1 << uint(i)
						wv.restoreSnapshot(top.snap, top.diffs)
						cell := &wv.cells[top.y][top.x]
						for _, t := range top.tried {
							if i == t {
								continue scan
							}
						}
						top.tried = append(top.tried, i)
						cell.collapseTo(i)
						ok, _ := wv.propagate(top.x, top.y)
						if ok {
							contradict = false
							continue mainLoop
						}
					}

					stack = stack[:len(stack)-1]
				}
				break
			}

			x, y, ok := wv.minEntropyCell(rng)
			if !ok {
				break
			}

			snap := wv.saveSnapshot(x, y)
			tileID := wv.observe(x, y, rng)
			if tileID < 0 {
				// No options — contradiction without a chosen tile.
				stack = append(stack, btFrame{
					snap: snap,
					x:    x,
					y:    y,
				})
				contradict = true
				continue
			}
			ok, diffs := wv.propagate(x, y)
			frame := btFrame{
				snap:  snap,
				x:     x,
				y:     y,
				tried: []int{tileID},
				diffs: diffs,
			}
			if !ok {
				stack = append(stack, frame)
				contradict = true
				continue
			}
			// Successful observation — push frame so future steps can backtrack here.
			stack = append(stack, frame)
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
		if !contradict && !wv.hasContradiction() && count > bestCollapsed {
			best = wv
			bestCollapsed = count
		}
	}
	return best
}

// tileRuneToType maps a WFC rune to a BattleMap TileType. floorType and
// wallType specify the base types for '.' and '#' respectively.
func tileRuneToType(ch rune, floorType, wallType TileType) TileType {
	switch ch {
	case '#':
		return wallType
	case '.':
		return floorType
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
					tt := tileRuneToType(ch, wv.rules.FloorTile, wv.rules.WallTile)
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

// findNearestFloor searches outward from (sx, sy) at the given level for a tile
// whose type matches one of the desired floor types. Returns the nearest match
// or the original coordinate if none is found within maxDist.
func findNearestFloor(m *BattleMap, sx, sy, level int, maxDist int, floors ...TileType) (int, int) {
	if sx < 0 || sx >= m.Width || sy < 0 || sy >= m.LevelHeight {
		return sx, sy
	}
	t := m.AtLevel(sx, sy, level).Type
	for _, ft := range floors {
		if t == ft {
			return sx, sy
		}
	}
	for r := 1; r <= maxDist; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if dx > -r && dx < r && dy > -r && dy < r {
					continue
				}
				nx, ny := sx+dx, sy+dy
				if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
					continue
				}
				nt := m.AtLevel(nx, ny, level).Type
				for _, ft := range floors {
					if nt == ft {
						return nx, ny
					}
				}
			}
		}
	}
	return sx, sy
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
			dir, ok := wfcDirFromKey[key]
			if !ok {
				continue
			}
			nb[dir] = append([]int{}, ids...)
		}
		weight := d.Weight
		if weight <= 0 {
			weight = 1
		}
		tiles = append(tiles, WFCTile{ID: d.ID, Name: d.Name, RuneGrid: grid, Neighbors: nb, Weight: weight})
	}
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].ID < tiles[j].ID })
	for i, t := range tiles {
		if t.ID != i {
			panic(fmt.Sprintf("wfc: tile %q has id %d but consecutive ids 0..%d required", t.Name, t.ID, len(tiles)-1))
		}
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

	// Connect levels with stairs. Search outward if the default central
	// position fell on a WFC-placed wall.
	stairsX := (w / 2) & ^1
	stairsY := (levelH / 2) & ^1
	if stairsX+1 >= w {
		stairsX = w - 2
	}
	if stairsY+1 >= levelH {
		stairsY = levelH - 2
	}
	sx0, sy0 := findNearestFloor(m, stairsX, stairsY, 0, 6, TileUFOFloor)
	sx1, sy1 := findNearestFloor(m, stairsX, stairsY, 1, 6, TileUFOFloor)
	sx := min(sx0, sx1)
	sy := min(sy0, sy1)
	m.SetLevel(sx, sy, 0, TileStairsDown)
	m.SetLevel(sx+1, sy, 0, TileUFOFloor)
	m.SetLevel(sx, sy+1, 0, TileUFOFloor)
	m.SetLevel(sx+1, sy+1, 0, TileUFOFloor)
	m.SetLevel(sx, sy, 1, TileStairs)
	m.SetLevel(sx+1, sy, 1, TileUFOFloor)
	m.SetLevel(sx, sy+1, 1, TileUFOFloor)
	m.SetLevel(sx+1, sy+1, 1, TileUFOFloor)

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
	rules.FloorTile = TileFloor
	rules.WallTile = TileWall

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
		m.fillGaps(level, TilePavement, gw*9, gh*9)

		// Enclose with a perimeter wall (y is level-relative).
		m.drawRectLevel(0, 0, w, h, level, TileWall)
	}

	// Add stairs between levels at a random interior position. If the
	// random position landed on a WFC wall, search outward for floor.
	for level := 0; level < numLevels-1; level++ {
		rx := 1 + rng.Intn(w-2)
		ry := 1 + rng.Intn(h-2)
		sx, sy := findNearestFloor(m, rx, ry, level, 6, TileFloor)
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

	// Connect levels with stairs. Search outward if the default central
	// position fell on a WFC-placed wall.
	stairsX := (w / 2) & ^1
	stairsY := (levelH / 2) & ^1
	if stairsX+1 >= w {
		stairsX = w - 2
	}
	if stairsY+1 >= levelH {
		stairsY = levelH - 2
	}
	sx0, sy0 := findNearestFloor(m, stairsX, stairsY, 0, 6, TileUFOFloor)
	sx1, sy1 := findNearestFloor(m, stairsX, stairsY, 1, 6, TileUFOFloor)
	sx := min(sx0, sx1)
	sy := min(sy0, sy1)
	m.SetLevel(sx, sy, 0, TileStairsDown)
	m.SetLevel(sx+1, sy, 0, TileUFOFloor)
	m.SetLevel(sx, sy+1, 0, TileUFOFloor)
	m.SetLevel(sx+1, sy+1, 0, TileUFOFloor)
	m.SetLevel(sx, sy, 1, TileStairs)
	m.SetLevel(sx+1, sy, 1, TileUFOFloor)
	m.SetLevel(sx, sy+1, 1, TileUFOFloor)
	m.SetLevel(sx+1, sy+1, 1, TileUFOFloor)

	return m
}

