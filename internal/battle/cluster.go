package battle

import (
	"math/rand"
)

// Blob generates K seed clusters of tile t, each grown by expanding to
// neighbors with probability Prob until reaching roughly Size tiles.
func (m *BattleMap) Blob(t TileType, seeds, size, prob int, rng *rand.Rand) {
	for s := 0; s < seeds; s++ {
		sx := rng.Intn(max(1, m.Width-2)) + 1
		sy := rng.Intn(max(1, m.LevelHeight-2)) + 1
		cur := 0
		frontier := [][2]int{{sx, sy}}
		visited := map[[2]int]bool{{sx, sy}: true}
		for len(frontier) > 0 && cur < size {
			idx := rng.Intn(len(frontier))
			cx, cy := frontier[idx][0], frontier[idx][1]
			frontier = append(frontier[:idx], frontier[idx+1:]...)
			if m.At(cx, cy).Type != t {
				m.Set(cx, cy, t)
				cur++
			}
			dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
			for _, d := range dirs {
				nx, ny := cx+d[0], cy+d[1]
				if nx < 1 || nx >= m.Width-1 || ny < 1 || ny >= m.LevelHeight-1 {
					continue
				}
				if visited[[2]int{nx, ny}] {
					continue
				}
				if rng.Intn(100) < prob {
					visited[[2]int{nx, ny}] = true
					frontier = append(frontier, [2]int{nx, ny})
				}
			}
		}
	}
}

// Poisson scatters tiles of type t with a minimum spacing of radius, producing
// sparse-but-even cover that avoids both clumping and grid uniformity.
func (m *BattleMap) Poisson(t TileType, radius, count int, rng *rand.Rand) {
	placed := [][2]int{}
	attempts := 0
	for len(placed) < count && attempts < count*20 {
		attempts++
		x := rng.Intn(max(1, m.Width-2)) + 1
		y := rng.Intn(max(1, m.LevelHeight-2)) + 1
		if m.At(x, y).Type != TileGrass && m.At(x, y).Type != TileFloor &&
			m.At(x, y).Type != TileUFOFloor && m.At(x, y).Type != TilePavement &&
			m.At(x, y).Type != TileSand && m.At(x, y).Type != TileSnow {
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

// ValidateMap checks that the map has reachable open space and is not fully
// walled off. Returns true if the map passes basic sanity checks.
// isPassable is a multi-level-aware passability check.
func (m *BattleMap) isPassableLevel(x, y, level int) bool {
	t := m.AtLevel(x, y, level)
	switch t.Type {
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs, TileStairsDown,
		TilePavement, TileSand, TileSnow, TileRubble,
		TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet,
		TileMarsh, TileIce, TileGlass, TileDebris, TileSkylight, TileBush,
		TileWheat, TilePier, TileScree, TileSwampWater, TileMud, TileVine, TileHeloRotor:
		return t.Level == level
	}
	return false
}

// ValidateMap checks the map for basic playability:
// - At least one passable tile exists.
// - At least 75% of passable tiles are reachable via flood-fill.
// Multi-level maps validate stair connectivity between levels.
func (m *BattleMap) ValidateMap() bool {
	open := 0
	type pos struct{ x, y, level int }
	var start *pos
	for level := 0; level < m.NumLevels; level++ {
		h := m.LevelHeight
		if m.NumLevels <= 1 {
			h = m.Height
		}
		for y := 0; y < h; y++ {
			for x := 0; x < m.Width; x++ {
				if m.isPassableLevel(x, y, level) {
					open++
					if start == nil {
						start = &pos{x, y, level}
					}
				}
			}
		}
	}
	if open == 0 {
		return false
	}

	seen := map[pos]bool{}
	stack := []pos{*start}
	reachable := 0
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

	for len(stack) > 0 {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[c] {
			continue
		}
		seen[c] = true
		reachable++

		// Cardinal neighbors on same level.
		for _, d := range dirs {
			nx, ny := c.x+d[0], c.y+d[1]
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
				continue
			}
			if !m.isPassableLevel(nx, ny, c.level) {
				continue
			}
			if !seen[pos{nx, ny, c.level}] {
				stack = append(stack, pos{nx, ny, c.level})
			}
		}

		// Stair traversal between levels.
		t := m.AtLevel(c.x, c.y, c.level)
		if t.Type == TileStairsDown && c.level+1 < m.NumLevels {
			below := m.AtLevel(c.x, c.y, c.level+1)
			if below.Type == TileStairs && !seen[pos{c.x, c.y, c.level + 1}] {
				stack = append(stack, pos{c.x, c.y, c.level + 1})
			}
		}
		if t.Type == TileStairs && c.level-1 >= 0 {
			above := m.AtLevel(c.x, c.y, c.level-1)
			if above.Type == TileStairsDown && !seen[pos{c.x, c.y, c.level - 1}] {
				stack = append(stack, pos{c.x, c.y, c.level - 1})
			}
		}
	}

	return reachable >= open*3/4
}

// RepairConnectivity flood-fills passable tiles from (sx,sy) and carves a
// UFO corridor from the seed to any unreachable floor pocket, guaranteeing the
// map is fully traversable. Used by walled base/interior generators.
func (m *BattleMap) RepairConnectivity(sx, sy int) {
	dirs := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	seen := map[[2]int]bool{}
	stack := [][2]int{{sx, sy}}
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[c] {
			continue
		}
		seen[c] = true
		for _, d := range dirs {
			nx, ny := c[0]+d[0], c[1]+d[1]
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
				continue
			}
			if m.Passable(nx, ny) && !seen[[2]int{nx, ny}] {
				stack = append(stack, [2]int{nx, ny})
			}
		}
	}
	// Find unreachable floor pockets and connect them to the seed.
	for y := 0; y < m.LevelHeight; y++ {
		for x := 0; x < m.Width; x++ {
			if !m.Passable(x, y) {
				continue
			}
			if seen[[2]int{x, y}] {
				continue
			}
			m.generateCorridorUFO(sx, sy, x, y, 1)
			// Re-flood from the new connection point.
			stack = append(stack, [2]int{x, y})
			for len(stack) > 0 {
				c := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if seen[c] {
					continue
				}
				seen[c] = true
				for _, d := range dirs {
					nx, ny := c[0]+d[0], c[1]+d[1]
					if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
						continue
					}
					if m.Passable(nx, ny) && !seen[[2]int{nx, ny}] {
						stack = append(stack, [2]int{nx, ny})
					}
				}
			}
		}
	}
}
