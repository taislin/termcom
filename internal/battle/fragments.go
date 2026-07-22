package battle

import (
	"math/rand"

	"github.com/taislin/termcom/internal/mapgen"
)

// tileTypeByName maps JSON string names to TileType.
var tileTypeByName = map[string]TileType{
	"TileFloor":       TileFloor,
	"TileWall":        TileWall,
	"TileDoor":        TileDoor,
	"TileDoorOpen":    TileDoorOpen,
	"TileWindow":      TileWindow,
	"TileGrass":       TileGrass,
	"TileTree":        TileTree,
	"TileRock":        TileRock,
	"TileWater":       TileWater,
	"TileUFOFloor":    TileUFOFloor,
	"TileUFOWall":     TileUFOWall,
	"TileStairs":      TileStairs,
	"TilePavement":    TilePavement,
	"TileSand":        TileSand,
	"TileSnow":        TileSnow,
	"TileMarsh":       TileMarsh,
	"TileBush":        TileBush,
	"TileFence":       TileFence,
	"TileRubble":      TileRubble,
	"TileObject":      TileObject,
	"TileConsole":     TileConsole,
	"TileMachinery":   TileMachinery,
	"TilePod":         TilePod,
	"TilePowerSource": TilePowerSource,
	"TileStorage":     TileStorage,
	"TileAlienTech":   TileAlienTech,
	"TileStairsDown":  TileStairsDown,
	"TileDesk":        TileDesk,
	"TileChair":       TileChair,
	"TileChairLeft":    TileChairLeft,
	"TileChairRight":   TileChairRight,
	"TileComputer":    TileComputer,
	"TileBed":         TileBed,
	"TileLocker":      TileLocker,
	"TileCabinet":     TileCabinet,
	"TileCar":         TileCar,
	"TileCarMid":      TileCarMid,
	"TileCarRight":    TileCarRight,
	"TileForklift":      TileForklift,
	"TileForkliftRight": TileForkliftRight,
	"TileFuelPump":      TileFuelPump,
	"TileContainerRed":    TileContainerRed,
	"TileContainerBlue":   TileContainerBlue,
	"TileContainerYellow": TileContainerYellow,
	"TileAdobe":       TileAdobe,
	"TileMetalWall":   TileMetalWall,
	"TileWreck":       TileWreck,
	"TileTimber":      TileTimber,
	"TileDish":        TileDish,
	"TileTruck":         TileTruck,
	"TileIce":           TileIce,
	"TileStreetlamp":    TileStreetlamp,
	"TileGlass":         TileGlass,
	"TileDebris":        TileDebris,
	"TileCryoPipe":      TileCryoPipe,
	"TileSkylight":      TileSkylight,
	"TileWheat":         TileWheat,
	"TileHayBale":       TileHayBale,
	"TilePier":          TilePier,
	"TileDockCrate":     TileDockCrate,
	"TileCliffFace":     TileCliffFace,
	"TileScree":         TileScree,
	"TileBoulder":       TileBoulder,
	"TileSwampWater":    TileSwampWater,
	"TileCypressTree":   TileCypressTree,
	"TileSnowTree":      TileSnowTree,
	"TileMud":           TileMud,
	"TileVine":          TileVine,
	"TileBamboo":        TileBamboo,
	"TileDryBush":       TileDryBush,
	"TileBusEnd":        TileBusEnd,
	"TileBusMid":        TileBusMid,
	"TileHeloBody":      TileHeloBody,
	"TileHeloTail":      TileHeloTail,
	"TileHeloNose":      TileHeloNose,
	"TileHeloRotor":     TileHeloRotor,
	"TileHeloRotorSides":TileHeloRotorSides,
	"TileHeloBodyBack":  TileHeloBodyBack,
	"TileHeloRotorBack": TileHeloRotorBack,
	"TileHeloWindow":    TileHeloWindow,
	"TileTractorCab":    TileTractorCab,
	"TileTractorBody":   TileTractorBody,
	"TileCrawlerLeft":   TileCrawlerLeft,
	"TileCrawlerMid":    TileCrawlerMid,
	"TileCrawlerRight":  TileCrawlerRight,
	"TileCrawlerLeg":    TileCrawlerLeg,
	"TileWheel":         TileWheel,
	"TileWheelSmall":    TileWheelSmall,
}

func resolveTileType(name string) TileType {
	if tt, ok := LookupTileType(name); ok {
		return tt
	}
	if tt, ok := tileTypeByName[name]; ok {
		return tt
	}
	return TileFloor
}

// ApplyMapgenChunk stamps a chunk onto a BattleMap at (startX, startY).
// Space characters (' ') are treated as transparent — the underlying tile is
// preserved. This enables nested mapgen (stamping a UFO on top of a forest).
// Nested chunks (place_nested) are stamped recursively at their offsets using
// the first variant of each pool. For weighted random selection from
// multi-variant pools, use ApplyMapgenChunkRNG.
func ApplyMapgenChunk(m *BattleMap, startX, startY int, chunk *mapgen.MapgenChunk) {
	applyMapgenChunkDepth(m, startX, startY, chunk, 0, nil)
}

// ApplyMapgenChunkRNG is like ApplyMapgenChunk but uses rng for weighted
// random selection from multi-variant nested pools (place_nested).
func ApplyMapgenChunkRNG(m *BattleMap, startX, startY int, chunk *mapgen.MapgenChunk, rng *rand.Rand) {
	applyMapgenChunkDepth(m, startX, startY, chunk, 0, rng)
}

func applyMapgenChunkDepth(m *BattleMap, startX, startY int, chunk *mapgen.MapgenChunk, depth int, rng *rand.Rand) {
	if depth > 10 || chunk == nil {
		return
	}
	for dy, row := range chunk.Rows {
		for dx := 0; dx < len(row); dx++ {
			ch := string(row[dx])
			if ch == " " {
				continue
			}
			tx, ty := startX+dx, startY+dy
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.LevelHeight {
				continue
			}
			if tt, ok := chunk.Terrain[ch]; ok {
				tt2 := resolveTileType(tt)
				m.Set(tx, ty, tt2)
				if tt2 == TileStreetlamp {
					m.Tiles[ty][tx].Lit = true
				}
			}
		}
	}
	for _, n := range chunk.Nested {
		var nested *mapgen.MapgenChunk
		if rng != nil {
			nested = mapgen.Pick(n.ID, rng)
		} else {
			nested = mapgen.Get(n.ID)
		}
		if nested != nil {
			applyMapgenChunkDepth(m, startX+n.X, startY+n.Y, nested, depth+1, rng)
		}
	}
}

// ApplyMapgenChunkRotated stamps a chunk with rotation (0-3 quarter-turns).
// Nested chunks (place_nested) are stamped with the same rotation applied to
// their offsets and to the nested chunk itself. The first variant of each
// nested pool is used. For weighted random selection from multi-variant pools,
// use ApplyMapgenChunkRotatedRNG.
func ApplyMapgenChunkRotated(m *BattleMap, startX, startY, rot int, chunk *mapgen.MapgenChunk) {
	applyMapgenChunkRotatedDepth(m, startX, startY, rot, chunk, 0, nil)
}

// ApplyMapgenChunkRotatedRNG is like ApplyMapgenChunkRotated but uses rng for
// weighted random selection from multi-variant nested pools (place_nested).
func ApplyMapgenChunkRotatedRNG(m *BattleMap, startX, startY, rot int, chunk *mapgen.MapgenChunk, rng *rand.Rand) {
	applyMapgenChunkRotatedDepth(m, startX, startY, rot, chunk, 0, rng)
}

func applyMapgenChunkRotatedDepth(m *BattleMap, startX, startY, rot int, chunk *mapgen.MapgenChunk, depth int, rng *rand.Rand) {
	if depth > 10 || chunk == nil {
		return
	}
	rot &= 3
	resolved := make([][]TileType, chunk.Height)
	for dy, row := range chunk.Rows {
		resolved[dy] = make([]TileType, chunk.Width)
		for dx := 0; dx < len(row) && dx < chunk.Width; dx++ {
			ch := string(row[dx])
			if ch == " " {
				resolved[dy][dx] = -1
				continue
			}
			if tt, ok := chunk.Terrain[ch]; ok {
				resolved[dy][dx] = resolveTileType(tt)
			} else {
				resolved[dy][dx] = -1
			}
		}
	}

	cw, ch := chunk.Width, chunk.Height
	for i := 0; i < rot; i++ {
		nw, nh := ch, cw
		rotated := make([][]TileType, nh)
		for y := 0; y < nh; y++ {
			rotated[y] = make([]TileType, nw)
		}
		for y := 0; y < ch; y++ {
			for x := 0; x < cw; x++ {
				rotated[x][ch-1-y] = resolved[y][x]
			}
		}
		resolved = rotated
		cw, ch = nw, nh
	}

	for dy := 0; dy < ch; dy++ {
		for dx := 0; dx < cw; dx++ {
			tt := resolved[dy][dx]
			if tt == -1 {
				continue
			}
			tx, ty := startX+dx, startY+dy
			if tx < 0 || tx >= m.Width || ty < 0 || ty >= m.LevelHeight {
				continue
			}
			m.Set(tx, ty, tt)
			if tt == TileStreetlamp {
				m.Tiles[ty][tx].Lit = true
			}
		}
	}

	for _, n := range chunk.Nested {
		var nested *mapgen.MapgenChunk
		if rng != nil {
			nested = mapgen.Pick(n.ID, rng)
		} else {
			nested = mapgen.Get(n.ID)
		}
		if nested == nil {
			continue
		}
		nx, ny := n.X, n.Y
		switch rot {
		case 1:
			nx, ny = chunk.Height-1-n.Y, n.X
		case 2:
			nx, ny = chunk.Width-1-n.X, chunk.Height-1-n.Y
		case 3:
			nx, ny = n.Y, chunk.Width-1-n.X
		}
		applyMapgenChunkRotatedDepth(m, startX+nx, startY+ny, rot, nested, depth+1, rng)
	}
}

// AssembleMap builds a map by filling base terrain, applying clustered biome
// terrain, then placing biome-tagged mapgen chunks (loaded from JSON) with
// rotation, spacing, and corridor connectivity. rng controls randomness.
func AssembleMap(biome string, w, h int, rng *rand.Rand) *BattleMap {
	if biome == "urban" && w >= 20 && h >= 20 {
		return GenerateTerrorSite(w, h, rng.Int63())
	}
	m := NewBattleMap(w, h)
	baseTile := TileGrass
	switch biome {
	case "urban":
		baseTile = TileGrass
	case "desert":
		baseTile = TileSand
	case "polar":
		baseTile = TileSnow
	case "rural":
		baseTile = TileGrass
	case "ufo", "alien":
		baseTile = TileUFOFloor
	case "farm":
		baseTile = TileGrass
	case "coastal":
		baseTile = TileSand
	case "mountain":
		baseTile = TileGrass
	case "swamp":
		baseTile = TileMarsh
	case "jungle":
		baseTile = TileGrass
	}
	m.fillRect(0, 0, w, h, baseTile)

	clusterBiome(m, biome, w, h, rng)

	// Coastal: ocean fills the top third.
	waterDepth := 0
	if biome == "coastal" {
		waterDepth = h / 3
		if waterDepth < 4 {
			waterDepth = 4
		}
		m.fillRect(0, 0, w, waterDepth, TileWater)
	}

	chunks := mapgen.ByTag(biome)
	if len(chunks) == 0 {
		return m
	}

	weightedPick := func(candidates []*mapgen.MapgenChunk) *mapgen.MapgenChunk {
		return mapgen.WeightedPick(candidates, rng)
	}

	type placed struct{ x, y, w, h int }
	positions := []placed{}

	// Coastal: place pier and tide pool chunks along the shoreline.
	if biome == "coastal" {
		pierChunk := mapgen.Get("coastal_pier")
		if pierChunk != nil {
			tries := 0
			pierPlaced := 0
			pierTarget := 2 + rng.Intn(2)
			for pierPlaced < pierTarget && tries < 30 {
				tries++
				px := 1 + rng.Intn(max(1, w-pierChunk.Width-2))
				py := waterDepth - 3
				if py < 0 {
					py = 0
				}
				overlap := false
				for _, p := range positions {
					if px < p.x+p.w+1 && px+pierChunk.Width+1 > p.x &&
						py < p.y+p.h+1 && py+pierChunk.Height+1 > p.y {
						overlap = true
						break
					}
				}
				if !overlap {
					ApplyMapgenChunkRotatedRNG(m, px, py, 0, pierChunk, rng)
					positions = append(positions, placed{px, py, pierChunk.Width, pierChunk.Height})
					pierPlaced++
				}
			}
		}
		tideChunk := mapgen.Get("coastal_tide_pool")
		if tideChunk != nil {
			tries := 0
			tidePlaced := 0
			for tidePlaced < 1 && tries < 20 {
				tries++
				px := 1 + rng.Intn(max(1, w-tideChunk.Width-2))
				py := waterDepth - 4
				if py < 0 {
					py = 0
				}
				overlap := false
				for _, p := range positions {
					if px < p.x+p.w+2 && px+tideChunk.Width+2 > p.x &&
						py < p.y+p.h+2 && py+tideChunk.Height+2 > p.y {
						overlap = true
						break
					}
				}
				if !overlap {
					ApplyMapgenChunkRotatedRNG(m, px, py, 0, tideChunk, rng)
					positions = append(positions, placed{px, py, tideChunk.Width, tideChunk.Height})
					tidePlaced++
				}
			}
		}
	}

	// For urban maps, reserve road corridors first (painted as pavement) so the
	// scattered buildings land only in the gaps and roads never cut through a
	// house.
	var roadReserved func(x, y int) bool
	if biome == "urban" {
		var roads []placed
		if rng.Intn(2) == 0 {
			roadX := w/4 + rng.Intn(w/2)
			roads = append(roads, placed{roadX - 1, 0, 3, h})
		}
		if rng.Intn(2) == 0 {
			roadY := h/4 + rng.Intn(h/2)
			roads = append(roads, placed{0, roadY - 1, w, 3})
		}
		roadReserved = func(x, y int) bool {
			for _, r := range roads {
				if x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h {
					return true
				}
			}
			return false
		}
		for _, r := range roads {
			m.fillRect(r.x, r.y, r.w, r.h, TilePavement)
		}
	}

	anchor := weightedPick(chunks)
	coastalExclude := map[string]bool{"coastal_pier": true, "coastal_docks": true, "coastal_boat": true, "coastal_tide_pool": true}
	for coastalExclude[anchor.ID] {
		anchor = weightedPick(chunks)
	}
	rot := 0
	if !anchor.NoRotate {
		rot = rng.Intn(4)
	}
	ax, ay := w/2-anchor.Width/2, h/2-anchor.Height/2
	if (roadReserved == nil || !roadReserved(ax, ay)) &&
		(waterDepth == 0 || ay >= waterDepth) {
		ApplyMapgenChunkRotatedRNG(m, ax, ay, rot, anchor, rng)
		positions = append(positions, placed{ax, ay, anchor.Width, anchor.Height})
	}

	attempts := 0
	target := 10 + rng.Intn(5)
	for len(positions) < target && attempts < 200 {
		attempts++
		c := weightedPick(chunks)
		if coastalExclude[c.ID] {
			continue
		}
		r := 0
		if !c.NoRotate {
			r = rng.Intn(4)
		}
		cw, ch := c.Width, c.Height
		for i := 0; i < r; i++ {
			cw, ch = ch, cw
		}
		fx := rng.Intn(max(1, w-cw-2)) + 1
		fy := rng.Intn(max(1, h-ch-2)) + 1

		overlap := false
		for _, p := range positions {
			if fx < p.x+p.w+1 && fx+cw+1 > p.x && fy < p.y+p.h+1 && fy+ch+1 > p.y {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		if roadReserved != nil {
			hitsRoad := false
			for py := fy; py < fy+ch && !hitsRoad; py++ {
				for px := fx; px < fx+cw; px++ {
					if roadReserved(px, py) {
						hitsRoad = true
						break
					}
				}
			}
			if hitsRoad {
				continue
			}
		}
		if waterDepth > 0 && fy < waterDepth {
			continue
		}
		ApplyMapgenChunkRotatedRNG(m, fx, fy, r, c, rng)
		positions = append(positions, placed{fx, fy, cw, ch})
	}

	// Connect each placed chunk to its nearest already-placed neighbor so the
	// corridor network is compact (minimum-spanning-tree-like) rather than a
	// simple chain that may create long winding paths.
	for i := 1; i < len(positions); i++ {
		pi := positions[i]
		cx1, cy1 := pi.x+pi.w/2, pi.y+pi.h/2
		bestDist := -1
		bestIdx := 0
		for j := 0; j < i; j++ {
			pj := positions[j]
			dx := cx1 - (pj.x + pj.w/2)
			dy := cy1 - (pj.y + pj.h/2)
			d := dx*dx + dy*dy
			if bestDist < 0 || d < bestDist {
				bestDist = d
				bestIdx = j
			}
		}
		pj := positions[bestIdx]
		m.generateCorridorFill(pj.x+pj.w/2, pj.y+pj.h/2, cx1, cy1, 1, baseTile)
	}

	return m
}

// clusterBiome applies biome-aware clustered terrain to m using blob growth
// and poisson spacing. Uses deterministic seeds derived from rng.
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
		m.Blob(TileIce, 5, w*h/60, 50, rng)
		m.Poisson(TileRock, 3, w*h/150, rng)
		m.Poisson(TileSnowTree, 3, w*h/200, rng)
	case "urban":
		m.Poisson(TileObject, 4, w*h/200, rng)
		m.Poisson(TileBush, 5, w*h/150, rng)
		m.Poisson(TileTree, 3, w*h/250, rng)
	case "rural":
		m.Blob(TileRock, 5, w*h/60, 50, rng)
		m.Blob(TileTree, 6, w*h/50, 55, rng)
		m.Poisson(TileObject, 4, w*h/200, rng)
	case "farm":
		m.Blob(TileWheat, 8, w*h/25, 65, rng)
		m.Poisson(TileTree, 4, w*h/150, rng)
		m.Poisson(TileFence, 4, w*h/200, rng)
		clearX := w/4 + rng.Intn(w/2)
		clearY := h/4 + rng.Intn(h/2)
		m.fillRect(clearX-3, clearY-3, 7, 7, TileGrass)
	case "coastal":
		m.Blob(TileSand, 4, w*h/50, 45, rng)
		m.Poisson(TileRock, 3, w*h/120, rng)
		m.Poisson(TileDryBush, 5, w*h/60, rng)
	case "mountain":
		m.Blob(TileRock, 5, w*h/40, 55, rng)
		m.Poisson(TileBoulder, 3, w*h/80, rng)
		m.Poisson(TileCliffFace, 3, w*h/60, rng)
		m.Poisson(TileScree, 5, w*h/50, rng)
		m.Poisson(TileBush, 3, w*h/100, rng)
		m.Poisson(TileTree, 2, w*h/200, rng)
		clearX := w/4 + rng.Intn(w/2)
		clearY := h/4 + rng.Intn(h/2)
		m.fillRect(clearX-2, clearY-2, 5, 5, TileGrass)
	case "swamp":
		m.Blob(TileSwampWater, 6, w*h/25, 60, rng)
		m.Poisson(TileCypressTree, 4, w*h/45, rng)
		m.Blob(TileMud, 4, w*h/80, 50, rng)
		m.Poisson(TileBush, 5, w*h/60, rng)
		m.Poisson(TileVine, 4, w*h/80, rng)
		m.Poisson(TileTree, 2, w*h/120, rng)
		clearX := w/4 + rng.Intn(w/2)
		clearY := h/4 + rng.Intn(h/2)
		m.fillRect(clearX-2, clearY-2, 5, 5, TileMarsh)
	case "jungle":
		m.Blob(TileTree, 8, w*h/20, 70, rng)
		m.Blob(TileBamboo, 6, w*h/35, 55, rng)
		m.Blob(TileVine, 7, w*h/40, 60, rng)
		m.Poisson(TileMud, 5, w*h/60, rng)
		clearX := w/4 + rng.Intn(w/2)
		clearY := h/4 + rng.Intn(h/2)
		m.fillRect(clearX-2, clearY-2, 5, 5, TileGrass)
	}
}
