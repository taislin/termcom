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
	Target    *UFO
	Launching bool
	RangeLeft int
}

func NewInterceptor(baseX, baseY int) *Interceptor {
	return &Interceptor{
		Name:   "Interceptor",
		X:      float64(baseX),
		Y:      float64(baseY),
		Speed:  36,
		HP:     60,
		MaxHP:  60,
		Weapon: "avalanche",
		Ammo:   8,
		Range:  60,
	}
}

func (i *Interceptor) Launch(target *UFO) {
	i.Target = target
	i.Launching = true
	i.RangeLeft = i.Range * 3
}

func (i *Interceptor) Update() bool {
	if i.Target == nil || !i.Target.Active {
		i.Target = nil
		i.Launching = false
		return false
	}

	dx := i.Target.X - i.X
	dy := i.Target.Y - i.Y
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
		i.Target = nil
		i.Launching = false
		return false
	}

	return false
}

func (i *Interceptor) Disengage() {
	i.Target = nil
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
