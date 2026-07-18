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
	"TileComputer":    TileComputer,
	"TileBed":         TileBed,
	"TileLocker":      TileLocker,
	"TileCabinet":     TileCabinet,
}

func resolveTileType(name string) TileType {
	if tt, ok := tileTypeByName[name]; ok {
		return tt
	}
	return TileFloor
}

// ApplyMapgenChunk stamps a chunk onto a BattleMap at (startX, startY).
// Space characters (' ') are treated as transparent — the underlying tile is
// preserved. This enables nested mapgen (stamping a UFO on top of a forest).
func ApplyMapgenChunk(m *BattleMap, startX, startY int, chunk *mapgen.MapgenChunk) {
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
				m.Set(tx, ty, resolveTileType(tt))
			}
			if ft, ok := chunk.Furniture[ch]; ok {
				m.Set(tx, ty, resolveTileType(ft))
			}
		}
	}
}

// ApplyMapgenChunkRotated stamps a chunk with rotation (0-3 quarter-turns).
func ApplyMapgenChunkRotated(m *BattleMap, startX, startY, rot int, chunk *mapgen.MapgenChunk) {
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
			if ft, ok := chunk.Furniture[ch]; ok {
				resolved[dy][dx] = resolveTileType(ft)
			} else if tt, ok := chunk.Terrain[ch]; ok {
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
		}
	}
}

// AssembleMap builds a map by filling base terrain, applying clustered biome
// terrain, then placing biome-tagged mapgen chunks (loaded from JSON) with
// rotation, spacing, and corridor connectivity. rng controls randomness.
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

	clusterBiome(m, biome, w, h, rng)

	chunks := mapgen.ByTag(biome)
	if len(chunks) == 0 {
		return m
	}

	anchor := chunks[rng.Intn(len(chunks))]
	rot := rng.Intn(4)
	ax, ay := w/2-anchor.Width/2, h/2-anchor.Height/2
	ApplyMapgenChunkRotated(m, ax, ay, rot, anchor)

	type placed struct{ x, y, w, h int }
	positions := []placed{{ax, ay, anchor.Width, anchor.Height}}

	attempts := 0
	target := 6 + rng.Intn(4)
	for len(positions) < target && attempts < 200 {
		attempts++
		c := chunks[rng.Intn(len(chunks))]
		r := rng.Intn(4)
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
		ApplyMapgenChunkRotated(m, fx, fy, r, c)
		positions = append(positions, placed{fx, fy, cw, ch})
	}

	for i := 0; i < len(positions)-1; i++ {
		p1 := positions[i]
		p2 := positions[i+1]
		m.generateCorridor(p1.x+p1.w/2, p1.y+p1.h/2, p2.x+p2.w/2, p2.y+p2.h/2, 1)
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
