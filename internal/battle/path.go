package battle

import (
	"container/heap"
	"math"
)

type pathNode struct {
	x, y int
	g    int
	f    int
	idx  int
	prev *pathNode
}

type pathPriorityQueue []*pathNode

func (pq pathPriorityQueue) Len() int { return len(pq) }

func (pq pathPriorityQueue) Less(i, j int) bool {
	if pq[i].f == pq[j].f {
		return pq[i].g > pq[j].g
	}
	return pq[i].f < pq[j].f
}

func (pq pathPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].idx = i
	pq[j].idx = j
}

func (pq *pathPriorityQueue) Push(x interface{}) {
	n := x.(*pathNode)
	n.idx = len(*pq)
	*pq = append(*pq, n)
}

func (pq *pathPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.idx = -1
	*pq = old[:n-1]
	return item
}

const pathMoveCost = 4

func heuristic(ax, ay, bx, by int) int {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	return (dx + dy) * pathMoveCost
}

func passableFor(m *BattleMap, units UnitList, ignore *Unit, x, y, level int) bool {
	if x < 0 || y < 0 || x >= m.Width || y >= m.LevelHeight {
		return false
	}
	t := m.AtLevel(x, y, level)
	switch t.Type {
	case TileFloor, TileDoor, TileGrass, TileUFOFloor, TileStairs, TileStairsDown, TilePavement, TileSand, TileSnow,
		TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech,
		TileDesk, TileChair, TileChairLeft, TileChairRight, TileComputer, TileBed, TileLocker, TileCabinet, TileRubble:
	default:
		return false
	}
	if u := units.At(x, y); u != nil && u != ignore {
		return false
	}
	return true
}

func AStar(sx, sy, ex, ey int, level int, m *BattleMap, units UnitList, ignore *Unit) [][2]int {
	if sx == ex && sy == ey {
		return [][2]int{{sx, sy}}
	}
	if !passableFor(m, units, ignore, ex, ey, level) {
		return nil
	}

	open := &pathPriorityQueue{}
	heap.Init(open)
	start := &pathNode{x: sx, y: sy, g: 0, f: heuristic(sx, sy, ex, ey)}
	heap.Push(open, start)

	best := map[int]*pathNode{}
	key := func(x, y int) int { return y*m.Width + x }
	best[key(sx, sy)] = start

	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for open.Len() > 0 {
		cur := heap.Pop(open).(*pathNode)
		if cur.x == ex && cur.y == ey {
			var path [][2]int
			for n := cur; n != nil; n = n.prev {
				path = append([][2]int{{n.x, n.y}}, path...)
			}
			return path
		}

		for _, d := range dirs {
			nx, ny := cur.x+d[0], cur.y+d[1]
			if !passableFor(m, units, ignore, nx, ny, level) {
				continue
			}
			ng := cur.g + pathMoveCost
			k := key(nx, ny)
			if existing, ok := best[k]; ok && existing.g <= ng {
				continue
			}
			nn := &pathNode{x: nx, y: ny, g: ng, f: ng + heuristic(nx, ny, ex, ey), prev: cur}
			best[k] = nn
			heap.Push(open, nn)
		}
	}
	return nil
}

func (ai *AlienAI) GetNextPathStep(tx, ty int, m *BattleMap, units UnitList) (int, int) {
	path := AStar(ai.Unit.X, ai.Unit.Y, tx, ty, ai.Unit.Level, m, units, ai.Unit)
	if len(path) < 2 {
		return ai.Unit.X, ai.Unit.Y
	}
	return path[1][0], path[1][1]
}

func (ai *AlienAI) reactionFirePenalty(x, y int, m *BattleMap, units UnitList) float64 {
	var penalty float64
	for _, u := range units {
		if !u.Alive || u.Faction != 0 || u == ai.Unit {
			continue
		}
		if u.TU < MinReactionTU || u.Weapon == "" {
			continue
		}
		dx := float64(x - u.X)
		dy := float64(y - u.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > float64(SightRange) {
			continue
		}
		if !u.CanSee(x, y, m) {
			continue
		}
		chance := u.Reactions*ReactionMult + u.Accuracy/ReactionAccDiv - int(dist)*ReactionDistPen
		if chance < ReactionMinChance {
			chance = 1
		}
		penalty += float64(chance)
	}
	return penalty
}
