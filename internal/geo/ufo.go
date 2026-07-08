package geo

import (
	"math/rand"
)

type UFOType struct {
	Name       string
	Short      string
	Speed      int
	Toughness  int // HP
	Weapon     string
	Points     int
}

var UFOTypes = []UFOType{
	{"Small Scout",   "SSC", 28, 10, "plasma_pistol", 5},
	{"Medium Scout",  "MSC", 24, 20, "plasma_rifle", 10},
	{"Large Scout",   "LSC", 20, 35, "plasma_rifle", 15},
	{"Harvester",     "HAR", 16, 50, "plasma_rifle", 20},
	{"Bomber",        "BMB", 12, 80, "plasma_rifle", 30},
	{"Transport",     "TRN", 10, 60, "plasma_rifle", 15},
}

type UFO struct {
	ID       int
	Type     UFOType
	// Network-based movement
	NodeFrom   int     // source node ID
	NodeTo     int     // destination node ID
	Progress   float64 // 0.0 to 1.0 along edge
	// Interpolated screen position (computed from nodes)
	X, Y       float64
	TurnsLeft  int
	Active     bool
}

func GetUFOTypeByName(name string) *UFOType {
	for i := range UFOTypes {
		if UFOTypes[i].Name == name {
			return &UFOTypes[i]
		}
	}
	return nil
}

var ufoIDCounter int

// SpawnUFOOnNetwork creates a UFO at a random edge on the network.
func SpawnUFOOnNetwork(gn *GeoNetwork) *UFO {
	t := UFOTypes[rand.Intn(len(UFOTypes))]
	ufoIDCounter++

	// Pick a random edge to start on
	edge := gn.Edges[rand.Intn(len(gn.Edges))]
	startNode := edge.From
	endNode := edge.To
	// Randomly reverse direction
	if rand.Intn(2) == 0 {
		startNode, endNode = endNode, startNode
	}

	ufo := &UFO{
		ID:         ufoIDCounter,
		Type:       t,
		NodeFrom:   startNode,
		NodeTo:     endNode,
		Progress:   0.0,
		TurnsLeft:  500 + rand.Intn(500),
		Active:     true,
	}
	ufo.updatePosition(gn)
	return ufo
}

// SpawnUFOAtNode creates a UFO arriving at a specific node from a random neighbor.
func SpawnUFOAtNode(gn *GeoNode, network *GeoNetwork) *UFO {
	t := UFOTypes[rand.Intn(len(UFOTypes))]
	ufoIDCounter++

	// Pick a random neighbor to come from
	neighbors := network.Neighbors(gn.ID)
	if len(neighbors) == 0 {
		return nil
	}
	from := neighbors[rand.Intn(len(neighbors))]

	ufo := &UFO{
		ID:         ufoIDCounter,
		Type:       t,
		NodeFrom:   from.ID,
		NodeTo:     gn.ID,
		Progress:   0.3, // partially along
		TurnsLeft:  500 + rand.Intn(500),
		Active:     true,
	}
	ufo.updatePosition(network)
	return ufo
}

func (u *UFO) Update(gn *GeoNetwork) {
	if !u.Active {
		return
	}
	// Speed as progress per tick (higher = faster traversal)
	speed := float64(u.Type.Speed) * 0.002
	u.Progress += speed

	if u.Progress >= 1.0 {
		// Arrived at destination node
		u.Progress = 1.0
		u.NodeFrom = u.NodeTo
		// Pick next destination: random neighbor
		neighbors := gn.Neighbors(u.NodeTo)
		if len(neighbors) > 0 {
			next := neighbors[rand.Intn(len(neighbors))]
			u.NodeFrom = u.NodeTo
			u.NodeTo = next.ID
			u.Progress = 0.0
		}
	}

	u.updatePosition(gn)
	u.TurnsLeft--
	if u.TurnsLeft <= 0 {
		u.Active = false
	}
}

func (u *UFO) updatePosition(gn *GeoNetwork) {
	from := gn.NodeByID(u.NodeFrom)
	to := gn.NodeByID(u.NodeTo)
	if from == nil || to == nil {
		return
	}
	u.X = float64(from.X) + float64(to.X-from.X)*u.Progress
	u.Y = float64(from.Y) + float64(to.Y-from.Y)*u.Progress
}

func (u *UFO) TileX() int {
	return int(u.X)
}

func (u *UFO) TileY() int {
	return int(u.Y)
}

func (u *UFO) CurrentNode() int {
	if u.Progress < 0.5 {
		return u.NodeFrom
	}
	return u.NodeTo
}

func (u *UFO) FireAtInterceptor(inter *Interceptor) int {
	if !u.Active {
		return 0
	}
	accuracy := 30
	damage := 5 + rand.Intn(10)
	if rand.Intn(100) < accuracy {
		inter.HP -= damage
		if inter.HP < 0 {
			inter.HP = 0
		}
		return damage
	}
	return 0
}

type UFOList []*UFO

func (ul UFOList) Active() []*UFO {
	var active []*UFO
	for _, u := range ul {
		if u.Active {
			active = append(active, u)
		}
	}
	return active
}

func (ul UFOList) Count() int {
	n := 0
	for _, u := range ul {
		if u.Active {
			n++
		}
	}
	return n
}
