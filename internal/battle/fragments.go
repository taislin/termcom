package battle

import (
	"math/rand"
)

// MapFragment is a reusable, hand-authored map piece that can be rotated and
// stamped onto a BattleMap during procedural assembly.
type MapFragment struct {
	W, H      int
	Tiles     [][]TileType
	Anchor    [2]int  // logical center used for placement / connection
	Tags      []string // biome tags: "urban", "forest", "ufo", "alien"
	DoorSides []int    // sides with doors: 0=south, 1=east, 2=north, 3=west
	Overwrite bool     // if true, replaces existing tiles instead of skipping
}

// rotateTiles returns the fragment tiles rotated by rot quarter-turns (0,1,2,3).
func (f *MapFragment) rotateTiles(rot int) ([][]TileType, int, int) {
	rot &= 3
	tiles := f.Tiles
	cw, ch := f.W, f.H
	for i := 0; i < rot; i++ {
		nw, nh := ch, cw
		rotated := make([][]TileType, nh)
		for y := 0; y < nh; y++ {
			rotated[y] = make([]TileType, nw)
		}
		// 90deg clockwise: dst[y][x] = src[ch-1-y][x]
		for y := 0; y < ch; y++ {
			for x := 0; x < cw; x++ {
				rotated[x][ch-1-y] = tiles[y][x]
			}
		}
		tiles = rotated
		cw, ch = nw, nh
	}
	return tiles, cw, ch
}

// PlaceFragment stamps frag onto m at (x,y) (top-left), rotating by rot.
// Existing non-floor tiles are preserved unless Overwrite is set. Doors from
// DoorSides are stamped on the rotated edges.
func (m *BattleMap) PlaceFragment(frag *MapFragment, x, y, rot int) {
	tiles, w, h := frag.rotateTiles(rot)
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tx, ty := x+dx, y+dy
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.LevelHeight {
				continue
			}
			tt := tiles[dy][dx]
			cur := m.At(tx, ty).Type
			if !frag.Overwrite {
				// Preserve existing solid structures; allow floors/ground to overlay.
				if cur != TileGrass && cur != TileFloor && cur != TileUFOFloor &&
					cur != TilePavement && cur != TileSand && cur != TileSnow {
					continue
				}
			}
			m.Set(tx, ty, tt)
		}
	}
	for _, side := range frag.DoorSides {
		ds := (side + rot) & 3
		switch ds {
		case 0:
			m.Set(x+w/2, y+h-1, TileDoor)
		case 1:
			m.Set(x+w-1, y+h/2, TileDoor)
		case 2:
			m.Set(x+w/2, y, TileDoor)
		case 3:
			m.Set(x, y+h/2, TileDoor)
		}
	}
}

// fragmentLibrary holds the built-in reusable fragments keyed by name.
var fragmentLibrary = map[string]*MapFragment{}

func registerFragment(name string, f *MapFragment) {
	fragmentLibrary[name] = f
}

func init() {
	registerFragment("ruined_shack", &MapFragment{
		W: 5, H: 4, Tags: []string{"urban", "forest", "rural"},
		DoorSides: []int{0},
		Tiles: [][]TileType{
			{TileWall, TileWall, TileWall, TileWall, TileWall},
			{TileWall, TileFloor, TileFloor, TileFloor, TileWall},
			{TileWall, TileFloor, TileFloor, TileFloor, TileWall},
			{TileWall, TileWall, TileFloor, TileWall, TileWall},
		},
	})
	registerFragment("bus_stop_cover", &MapFragment{
		W: 3, H: 2, Tags: []string{"urban"},
		Tiles: [][]TileType{
			{TileWall, TileWall, TileWall},
			{TileFloor, TileFloor, TileFloor},
		},
	})
	registerFragment("urban_building", &MapFragment{
		W: 7, H: 6, Tags: []string{"urban"},
		DoorSides: []int{0},
		Tiles: [][]TileType{
			{TileWall, TileWall, TileWall, TileWall, TileWall, TileWall, TileWall},
			{TileWall, TileFloor, TileFloor, TileFloor, TileFloor, TileFloor, TileWall},
			{TileWall, TileFloor, TileChair, TileFloor, TileComputer, TileFloor, TileWall},
			{TileWall, TileFloor, TileFloor, TileDesk, TileFloor, TileFloor, TileWall},
			{TileWall, TileFloor, TileLocker, TileFloor, TileFloor, TileCabinet, TileWall},
			{TileWall, TileWall, TileWall, TileWall, TileWall, TileWall, TileWall},
		},
	})
	registerFragment("junction", &MapFragment{
		W: 3, H: 3, Tags: []string{"ufo", "alien"},
		DoorSides: []int{0, 1, 2, 3},
		Tiles: [][]TileType{
			{TileUFOWall, TileUFOFloor, TileUFOWall},
			{TileUFOFloor, TileUFOFloor, TileUFOFloor},
			{TileUFOWall, TileUFOFloor, TileUFOWall},
		},
	})
	registerFragment("corridor_elbow", &MapFragment{
		W: 3, H: 3, Tags: []string{"ufo", "alien"},
		DoorSides: []int{3, 1},
		Tiles: [][]TileType{
			{TileUFOWall, TileUFOWall, TileUFOWall},
			{TileUFOFloor, TileUFOFloor, TileUFOFloor},
			{TileUFOWall, TileUFOWall, TileUFOWall},
		},
	})
	registerFragment("ufo_pod_room", &MapFragment{
		W: 5, H: 5, Tags: []string{"ufo"},
		DoorSides: []int{0},
		Tiles: [][]TileType{
			{TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TileUFOFloor, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TilePod, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TileUFOFloor, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall},
		},
	})
	registerFragment("alien_altar", &MapFragment{
		W: 5, H: 5, Tags: []string{"alien"},
		DoorSides: []int{0},
		Tiles: [][]TileType{
			{TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TileUFOFloor, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TileAlienTech, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOFloor, TileUFOFloor, TileUFOFloor, TileUFOWall},
			{TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall, TileUFOWall},
		},
	})
}

// fragmentsForBiome returns fragments tagged with the given biome.
func fragmentsForBiome(biome string) []*MapFragment {
	var out []*MapFragment
	for _, f := range fragmentLibrary {
		for _, t := range f.Tags {
			if t == biome {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

// AssembleMap builds a map by scattering base terrain, placing an anchor
// fragment, then greedily placing biome fragments with spacing and
// connectivity checks. rng controls randomness for reproducibility.
func AssembleMap(biome string, w, h int, rng *rand.Rand) *BattleMap {
	m := NewBattleMap(w, h)
	baseTile := TileGrass
	switch biome {
	case "urban":
		baseTile = TilePavement
	case "desert":
		baseTile = TileSand
	case "polar":
		baseTile = TileSnow
	case "rural":
		baseTile = TileGrass
	case "ufo", "alien":
		baseTile = TileUFOFloor
	}
	m.fillRect(0, 0, w, h, baseTile)

	// Apply biome-specific clustered terrain (blob growth + poisson spacing).
	clusterBiome(m, biome, w, h, rng)

	frags := fragmentsForBiome(biome)
	if len(frags) == 0 {
		return m
	}

	// Anchor fragment near center.
	anchor := frags[rng.Intn(len(frags))]
	ax := w/2 - anchor.W/2
	ay := h/2 - anchor.H/2
	arot := rng.Intn(4)
	m.PlaceFragment(anchor, ax, ay, arot)

	placed := []struct{ x, y, w, h int }{{ax, ay, anchor.W, anchor.H}}
	attempts := 0
	target := 6 + rng.Intn(4)
	for len(placed) < target && attempts < 200 {
		attempts++
		f := frags[rng.Intn(len(frags))]
		rot := rng.Intn(4)
		_, fw, fh := f.rotateTiles(rot)
		fx := rng.Intn(max(1, w-fw-2)) + 1
		fy := rng.Intn(max(1, h-fh-2)) + 1

		overlap := false
		for _, p := range placed {
			if fx < p.x+p.w+1 && fx+fw+1 > p.x && fy < p.y+p.h+1 && fy+fh+1 > p.y {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		m.PlaceFragment(f, fx, fy, rot)
		placed = append(placed, struct{ x, y, w, h int }{fx, fy, fw, fh})
	}

	// Connect consecutive placed fragments with corridors.
	for i := 0; i < len(placed)-1; i++ {
		c1 := placed[i]
		c2 := placed[i+1]
		x1, y1 := c1.x+c1.w/2, c1.y+c1.h/2
		x2, y2 := c2.x+c2.w/2, c2.y+c2.h/2
		m.generateCorridor(x1, y1, x2, y2, 1)
	}

	return m
}

// clusterBiome applies biome-aware clustered terrain to m, replacing the old
// uniform scatter. Uses deterministic seeds derived from w/h/rng so output is
// reproducible for a given (biome, size, rng) combination.
func clusterBiome(m *BattleMap, biome string, w, h int, rng *rand.Rand) {
	switch biome {
	case "forest":
		m.Blob(TileTree, 6, w*h/40, 55, rng)
		m.Blob(TileBush, 8, w*h/60, 60, rng)
		m.Poisson(TileRock, 3, w*h/120, rng)
		clearX := w/4 + rng.Intn(w/2)
		clearY := h/4 + rng.Intn(h/2)
		m.fillRect(clearX-3, clearY-3, 7, 7, TileGrass)
	case "desert":
		m.Blob(TileSand, 5, w*h/50, 50, rng)
		m.Blob(TileRock, 4, w*h/80, 45, rng)
		m.Poisson(TileBush, 4, w*h/200, rng)
	case "polar":
		m.Blob(TileMarsh, 5, w*h/60, 50, rng)
		m.Poisson(TileRock, 3, w*h/150, rng)
	case "urban":
		m.Poisson(TileObject, 4, w*h/200, rng)
	case "rural":
		m.Blob(TileRock, 5, w*h/60, 50, rng)
		m.Blob(TileTree, 6, w*h/50, 55, rng)
		m.Poisson(TileObject, 4, w*h/200, rng)
	}
}
