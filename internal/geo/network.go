package geo

import (
	"math"

	"github.com/civ13/ycom/internal/engine"
	"github.com/gdamore/tcell/v3"
)

type GeoNode struct {
	ID          int
	Name        string
	X, Y        int    // screen coordinates (0-79, 0-39)
	Region      string // continent/region name
	Threat      int    // 0-100, alien activity level
	HasRadar    bool
	InterceptorCount int
	MissionHere bool
}

type GeoEdge struct {
	From, To int // node IDs
	Length   int // travel time in ticks
}

type GeoNetwork struct {
	Nodes      []*GeoNode
	Edges      []GeoEdge
	Selected   int // selected node ID (-1 = none)
	Hovered    int // hovered node ID
}

// NewRegionalNetwork creates a 15-20 node network of regional hubs.
func NewRegionalNetwork() *GeoNetwork {
	gn := &GeoNetwork{
		Nodes: []*GeoNode{
			// North America
			{ID: 0, Name: "New York", X: 18, Y: 12, Region: "NA East"},
			{ID: 1, Name: "Los Angeles", X: 8, Y: 14, Region: "NA West"},
			{ID: 2, Name: "Chicago", X: 14, Y: 11, Region: "NA Central"},
			{ID: 3, Name: "Mexico City", X: 12, Y: 18, Region: "Central Am"},
			// South America
			{ID: 4, Name: "Bogota", X: 16, Y: 22, Region: "SA North"},
			{ID: 5, Name: "Brasilia", X: 22, Y: 26, Region: "SA East"},
			{ID: 6, Name: "Buenos Aires", X: 19, Y: 33, Region: "SA South"},
			// Europe
			{ID: 7, Name: "London", X: 38, Y: 8, Region: "Europe W"},
			{ID: 8, Name: "Paris", X: 39, Y: 10, Region: "Europe W"},
			{ID: 9, Name: "Berlin", X: 42, Y: 8, Region: "Europe C"},
			{ID: 10, Name: "Moscow", X: 48, Y: 7, Region: "Europe E"},
			// Africa
			{ID: 11, Name: "Cairo", X: 46, Y: 15, Region: "Africa N"},
			{ID: 12, Name: "Lagos", X: 38, Y: 20, Region: "Africa W"},
			{ID: 13, Name: "Nairobi", X: 48, Y: 22, Region: "Africa E"},
			// Asia
			{ID: 14, Name: "Delhi", X: 56, Y: 14, Region: "South Asia"},
			{ID: 15, Name: "Beijing", X: 62, Y: 11, Region: "East Asia"},
			{ID: 16, Name: "Tokyo", X: 68, Y: 11, Region: "East Asia"},
			{ID: 17, Name: "Singapore", X: 60, Y: 22, Region: "SE Asia"},
			// Oceania
			{ID: 18, Name: "Sydney", X: 66, Y: 30, Region: "Oceania"},
		},
		Edges: []GeoEdge{
			// NA connections
			{From: 0, To: 2, Length: 8},
			{From: 2, To: 1, Length: 10},
			{From: 0, To: 7, Length: 20},   // transatlantic
			{From: 1, To: 3, Length: 8},
			{From: 3, To: 4, Length: 6},
			// SA connections
			{From: 4, To: 5, Length: 8},
			{From: 5, To: 6, Length: 8},
			{From: 4, To: 12, Length: 12},
			// Europe connections
			{From: 7, To: 8, Length: 3},
			{From: 8, To: 9, Length: 4},
			{From: 9, To: 10, Length: 8},
			{From: 7, To: 10, Length: 14},
			// Africa connections
			{From: 11, To: 12, Length: 10},
			{From: 12, To: 13, Length: 10},
			{From: 11, To: 13, Length: 8},
			{From: 11, To: 14, Length: 10},
			// Asia connections
			{From: 14, To: 15, Length: 10},
			{From: 15, To: 16, Length: 6},
			{From: 15, To: 17, Length: 10},
			{From: 16, To: 18, Length: 12},
			{From: 17, To: 18, Length: 10},
			// Cross-region
			{From: 10, To: 15, Length: 16},
			{From: 13, To: 17, Length: 14},
		},
		Selected: -1,
		Hovered:  -1,
	}
	return gn
}

// NodeByID returns the node with the given ID, or nil.
func (gn *GeoNetwork) NodeByID(id int) *GeoNode {
	for _, n := range gn.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// NearestNode returns the node closest to screen coordinates (x,y).
func (gn *GeoNetwork) NearestNode(x, y int) *GeoNode {
	var best *GeoNode
	bestDist := math.MaxInt64
	for _, n := range gn.Nodes {
		dx := n.X - x
		dy := n.Y - y
		d := dx*dx + dy*dy
		if d < bestDist {
			bestDist = d
			best = n
		}
	}
	if bestDist > 25 { // max 5 tiles away
		return nil
	}
	return best
}

// Neighbors returns nodes directly connected to the given node.
func (gn *GeoNetwork) Neighbors(nodeID int) []*GeoNode {
	var result []*GeoNode
	for _, e := range gn.Edges {
		if e.From == nodeID {
			if n := gn.NodeByID(e.To); n != nil {
				result = append(result, n)
			}
		} else if e.To == nodeID {
			if n := gn.NodeByID(e.From); n != nil {
				result = append(result, n)
			}
		}
	}
	return result
}

// EdgeBetween returns the edge connecting two nodes, or nil.
func (gn *GeoNetwork) EdgeBetween(a, b int) *GeoEdge {
	for i := range gn.Edges {
		e := &gn.Edges[i]
		if (e.From == a && e.To == b) || (e.From == b && e.To == a) {
			return e
		}
	}
	return nil
}

// ShortestPath returns node IDs from start to end using BFS.
func (gn *GeoNetwork) ShortestPath(start, end int) []int {
	if start == end {
		return []int{start}
	}
	type item struct {
		id   int
		path []int
	}
	queue := []item{{id: start, path: []int{start}}}
	visited := map[int]bool{start: true}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, n := range gn.Neighbors(cur.id) {
			if visited[n.ID] {
				continue
			}
			newPath := make([]int, len(cur.path)+1)
			copy(newPath, cur.path)
			newPath[len(cur.path)] = n.ID

			if n.ID == end {
				return newPath
			}
			visited[n.ID] = true
			queue = append(queue, item{id: n.ID, path: newPath})
		}
	}
	return nil
}

// Render draws the network graph on screen.
func (gn *GeoNetwork) Render(ctx *engine.ScreenCtx, w, h int) {
	// Draw edges
	for _, e := range gn.Edges {
		from := gn.NodeByID(e.From)
		to := gn.NodeByID(e.To)
		if from == nil || to == nil {
			continue
		}
		gn.drawLine(ctx, from.X+1, from.Y+1, to.X+1, to.Y+1, w, h)
	}

	// Draw nodes
	for _, n := range gn.Nodes {
		sx := n.X + 1
		sy := n.Y + 1
		if sx < 1 || sx >= w-1 || sy < 1 || sy >= h-6 {
			continue
		}

		ch, style := gn.nodeStyle(n)
		ctx.SetCell(sx, sy, ch, style)

		// Draw name and stats below node
		name := n.Name
		if len(name) > 10 {
			name = name[:10]
		}
		ctx.DrawString(sx-len(name)/2, sy+1, name, engine.StyleGray)

		// Threat indicator
		if n.Threat > 0 {
			threatCh := '!'
			threatStyle := engine.StyleYellow
			if n.Threat > 50 {
				threatCh = '!'
				threatStyle = engine.StyleRedBold
			}
			ctx.SetCell(sx+2, sy, threatCh, threatStyle)
		}

		// Radar indicator
		if n.HasRadar {
			ctx.SetCell(sx-2, sy, 'R', engine.StyleCyan)
		}

		// Interceptor count
		if n.InterceptorCount > 0 {
			ctx.SetCell(sx, sy-1, rune('0'+n.InterceptorCount), engine.StyleGreen)
		}

		// Mission indicator
		if n.MissionHere {
			ctx.SetCell(sx-1, sy, '*', engine.StyleMagenta)
		}
	}
}

func (gn *GeoNetwork) nodeStyle(n *GeoNode) (rune, tcell.Style) {
	if n.ID == 0 { // Home base
		return '\u25C6', engine.StyleCyanBold // ◆
	}
	if n.Threat > 50 {
		return '\u25CF', engine.StyleRedBold // ●
	}
	if n.Threat > 0 {
		return '\u25CB', engine.StyleYellow // ○
	}
	return '\u25CB', engine.StyleGreen // ○
}

func (gn *GeoNetwork) drawLine(ctx *engine.ScreenCtx, x1, y1, x2, y2, w, h int) {
	dx := int(math.Abs(float64(x2 - x1)))
	dy := int(math.Abs(float64(y2 - y1)))
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 > 0 && x1 < w-1 && y1 > 0 && y1 < h-6 {
			ctx.SetCell(x1, y1, '\u2500', engine.StyleGray) // ─
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}
