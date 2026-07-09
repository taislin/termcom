package geo

import (
	"math/rand"
)

type UFOType struct {
	Name       string
	Short      string
	Speed      int
	Toughness  int // current HP
	MaxHP      int // original HP
	Weapon     string
	Points     int
}

var UFOTypes = []UFOType{
	{"Small Scout",   "SSC", 28, 10, 10, "plasma_pistol", 5},
	{"Medium Scout",  "MSC", 24, 20, 20, "plasma_rifle", 10},
	{"Large Scout",   "LSC", 20, 35, 35, "plasma_rifle", 15},
	{"Harvester",     "HAR", 16, 50, 50, "plasma_rifle", 20},
	{"Bomber",        "BMB", 12, 80, 80, "plasma_rifle", 30},
	{"Transport",     "TRN", 10, 60, 60, "plasma_rifle", 15},
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

// SpawnUFOOnCities creates a UFO at a random position moving between two cities.
func SpawnUFOOnCities(cities []*City) *UFO {
	t := UFOTypes[rand.Intn(len(UFOTypes))]
	ufoIDCounter++

	if len(cities) < 2 {
		return nil
	}

	// Pick two random different cities
	idx1 := rand.Intn(len(cities))
	idx2 := rand.Intn(len(cities)-1)
	if idx2 >= idx1 {
		idx2++
	}

	ufo := &UFO{
		ID:         ufoIDCounter,
		Type:       t,
		NodeFrom:   cities[idx1].ID,
		NodeTo:     cities[idx2].ID,
		Progress:   0.0,
		TurnsLeft:  500 + rand.Intn(500),
		Active:     true,
	}
	ufo.updatePosition(cities)
	return ufo
}

// SpawnUFOAtCity creates a UFO arriving at a specific city from a random other city.
func SpawnUFOAtCity(target *City, cities []*City) *UFO {
	t := UFOTypes[rand.Intn(len(UFOTypes))]
	ufoIDCounter++

	// Pick a random different city to come from
	var candidates []*City
	for _, c := range cities {
		if c.ID != target.ID {
			candidates = append(candidates, c)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	from := candidates[rand.Intn(len(candidates))]

	ufo := &UFO{
		ID:         ufoIDCounter,
		Type:       t,
		NodeFrom:   from.ID,
		NodeTo:     target.ID,
		Progress:   0.3,
		TurnsLeft:  500 + rand.Intn(500),
		Active:     true,
	}
	ufo.updatePosition(cities)
	return ufo
}

func (u *UFO) Update(cities []*City) {
	if !u.Active {
		return
	}
	// Speed as progress per tick (higher = faster traversal)
	speed := float64(u.Type.Speed) * 0.002
	u.Progress += speed

	if u.Progress >= 1.0 {
		// Arrived at destination city
		u.Progress = 1.0
		u.NodeFrom = u.NodeTo
		// Pick next destination: random city (not current)
		var candidates []*City
		for _, c := range cities {
			if c.ID != u.NodeTo {
				candidates = append(candidates, c)
			}
		}
		if len(candidates) > 0 {
			next := candidates[rand.Intn(len(candidates))]
			u.NodeFrom = u.NodeTo
			u.NodeTo = next.ID
			u.Progress = 0.0
		}
	}

	u.updatePosition(cities)
	u.TurnsLeft--
	if u.TurnsLeft <= 0 {
		u.Active = false
	}
}

func (u *UFO) updatePosition(cities []*City) {
	var from, to *City
	for _, c := range cities {
		if c.ID == u.NodeFrom {
			from = c
		}
		if c.ID == u.NodeTo {
			to = c
		}
	}
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
