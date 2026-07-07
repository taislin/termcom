package geo

import (
	"math"
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
	X, Y     float64
	DX, DY   float64
	TurnsLeft int
	Active   bool
}

var ufoIDCounter int

func NewUFO(x, y float64) *UFO {
	t := UFOTypes[rand.Intn(len(UFOTypes))]
	ufoIDCounter++
	// Random heading toward a city
	target := cities[rand.Intn(len(cities))]
	angle := math.Atan2(float64(target.Y)-y, float64(target.X)-x)
	speed := float64(t.Speed) * 0.02
	return &UFO{
		ID:    ufoIDCounter,
		Type:  t,
		X:     x,
		Y:     y,
		DX:    math.Cos(angle) * speed,
		DY:    math.Sin(angle) * speed,
		TurnsLeft: 500 + rand.Intn(500),
		Active: true,
	}
}

func (u *UFO) Update() {
	if !u.Active {
		return
	}
	u.X += u.DX
	u.Y += u.DY
	// Wrap around map
	if u.X < 0 {
		u.X += mapW
	}
	if u.X >= float64(mapW) {
		u.X -= mapW
	}
	if u.Y < 0 {
		u.Y = 0
		u.DY = -u.DY
	}
	if u.Y >= float64(mapH) {
		u.Y = float64(mapH) - 1
		u.DY = -u.DY
	}
	u.TurnsLeft--
	if u.TurnsLeft <= 0 {
		u.Active = false
	}
}

func (u *UFO) TileX() int {
	return int(u.X)
}

func (u *UFO) TileY() int {
	return int(u.Y)
}

func SpawnUFO() *UFO {
	// Spawn on edge of map at a random land tile
	side := rand.Intn(4)
	var x, y float64
	switch side {
	case 0: // top
		x = float64(rand.Intn(mapW))
		y = 0
	case 1: // bottom
		x = float64(rand.Intn(mapW))
		y = float64(mapH - 1)
	case 2: // left
		x = 0
		y = float64(rand.Intn(mapH))
	case 3: // right
		x = float64(mapW - 1)
		y = float64(rand.Intn(mapH))
	}
	return NewUFO(x, y)
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
