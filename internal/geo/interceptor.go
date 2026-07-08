package geo

import (
	"math"
	"math/rand"
)

type Interceptor struct {
	Name      string
	X, Y      float64
	Speed     int
	HP        int
	MaxHP     int
	Weapon    string
	Ammo      int
	Range     int
	TargetNode int    // node ID to patrol/intercept at (-1 = none)
	TargetUFO  *UFO   // specific UFO to pursue
	Launching  bool
	RangeLeft  int
}

func NewInterceptor(baseX, baseY int) *Interceptor {
	return &Interceptor{
		Name:       "Interceptor",
		X:          float64(baseX),
		Y:          float64(baseY),
		Speed:      36,
		HP:         60,
		MaxHP:      60,
		Weapon:     "avalanche",
		Ammo:       8,
		Range:      60,
		TargetNode: -1,
	}
}

// LaunchAtNode sends interceptor to a node to patrol/intercept.
func (i *Interceptor) LaunchAtNode(nodeID int, gn *GeoNetwork) {
	node := gn.NodeByID(nodeID)
	if node == nil {
		return
	}
	i.TargetNode = nodeID
	i.TargetUFO = nil
	i.Launching = true
	i.RangeLeft = i.Range * 3
}

// LaunchAtUFO sends interceptor to pursue a specific UFO.
func (i *Interceptor) LaunchAtUFO(ufo *UFO) {
	i.TargetUFO = ufo
	i.TargetNode = -1
	i.Launching = true
	i.RangeLeft = i.Range * 3
}

// Update moves interceptor toward its target. Returns true if reached.
func (i *Interceptor) Update(gn *GeoNetwork) bool {
	if i.TargetUFO != nil {
		if !i.TargetUFO.Active {
			i.TargetUFO = nil
			i.Launching = false
			return false
		}
		// Chase the UFO's current position
		return i.moveTo(i.TargetUFO.X, i.TargetUFO.Y)
	}

	if i.TargetNode >= 0 {
		node := gn.NodeByID(i.TargetNode)
		if node == nil {
			i.Launching = false
			return false
		}
		tx := float64(node.X)
		ty := float64(node.Y)
		reached := i.moveTo(tx, ty)
		if reached {
			// Check if any UFOs are at this node
			for _, ufo := range (&UFOList{}).Active() {
				if ufo.CurrentNode() == i.TargetNode {
					return true
				}
			}
		}
		return false
	}

	return false
}

func (i *Interceptor) moveTo(tx, ty float64) bool {
	dx := tx - i.X
	dy := ty - i.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 1.5 {
		return true
	}

	speed := float64(i.Speed) * 0.015
	if speed > dist {
		speed = dist
	}
	if dist > 0 {
		i.X += (dx / dist) * speed
		i.Y += (dy / dist) * speed
	}

	i.RangeLeft--
	if i.RangeLeft <= 0 {
		i.TargetNode = -1
		i.TargetUFO = nil
		i.Launching = false
		return false
	}

	return false
}

func (i *Interceptor) Disengage() {
	i.TargetNode = -1
	i.TargetUFO = nil
	i.Launching = false
}

func (i *Interceptor) FireAt(ufo *UFO) int {
	if i.Ammo <= 0 {
		return 0
	}
	i.Ammo--
	damage := 15 + rand.Intn(20)
	ufo.Type.Toughness -= damage
	if ufo.Type.Toughness <= 0 {
		ufo.Active = false
		return -1
	}
	return damage
}

type InterceptorList []*Interceptor

func (il InterceptorList) Active() []*Interceptor {
	var active []*Interceptor
	for _, i := range il {
		if i.HP > 0 {
			active = append(active, i)
		}
	}
	return active
}
