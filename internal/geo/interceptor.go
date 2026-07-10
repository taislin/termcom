package geo

import (
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
)

type Interceptor struct {
	Name       string
	X, Y       float64
	Speed      int
	HP         int
	MaxHP      int
	WeaponKey  string
	Weapon     data.InterceptorWeapon
	Ammo       int
	Range      int
	TargetNode int
	TargetUFO  *UFO
	Launching  bool
	RangeLeft  int
	Mode       data.CombatMode
	PilotSkill int // 0-100, affects accuracy
	State      *data.InterceptorState
}

func NewInterceptor(baseX, baseY int) *Interceptor {
	w := data.InterceptorWeapons["avalanche"]
	return &Interceptor{
		Name:       "Interceptor",
		X:          float64(baseX),
		Y:          float64(baseY),
		Speed:      36,
		HP:         60,
		MaxHP:      60,
		WeaponKey:  "avalanche",
		Weapon:     w,
		Ammo:       w.FireRate * 4,
		Range:      w.Range,
		TargetNode: -1,
		Mode:       data.CombatCautious,
		PilotSkill: 50,
	}
}

// NewInterceptorFromState creates an interceptor from a persisted state.
func NewInterceptorFromState(s *data.InterceptorState, baseX, baseY int) *Interceptor {
	w := data.InterceptorWeapons[s.WeaponKey]
	return &Interceptor{
		Name:       s.Name,
		X:          float64(baseX),
		Y:          float64(baseY),
		Speed:      36,
		HP:         s.HP,
		MaxHP:      s.MaxHP,
		WeaponKey:  s.WeaponKey,
		Weapon:     w,
		Ammo:       s.Ammo,
		Range:      w.Range,
		TargetNode: -1,
		Mode:       data.CombatCautious,
		PilotSkill: 50,
		State:      s,
	}
}

// SetWeapon changes the interceptor's loadout and rearms it.
func (i *Interceptor) SetWeapon(key string) {
	w, ok := data.InterceptorWeapons[key]
	if !ok {
		return
	}
	i.WeaponKey = key
	i.Weapon = w
	i.Range = w.Range
	i.Ammo = w.FireRate * 4
}

// SetMode changes the interceptor's combat behavior.
func (i *Interceptor) SetMode(m data.CombatMode) {
	i.Mode = m
}

// EffectiveAccuracy returns accuracy factoring pilot skill.
func (i *Interceptor) EffectiveAccuracy() int {
	base := i.Weapon.Accuracy
	pilotMod := (i.PilotSkill - 50) / 5
	return base + pilotMod
}

// LaunchAtNode sends interceptor to a city to patrol/intercept.
func (i *Interceptor) LaunchAtNode(nodeID int, cities []*City) {
	for _, c := range cities {
		if c.ID == nodeID {
			i.TargetNode = nodeID
			i.TargetUFO = nil
			i.Launching = true
			i.RangeLeft = i.Range * 3
			return
		}
	}
}

// LaunchAtUFO sends interceptor to pursue a specific UFO.
func (i *Interceptor) LaunchAtUFO(ufo *UFO) {
	i.TargetUFO = ufo
	i.TargetNode = -1
	i.Launching = true
	i.RangeLeft = i.Range * 3
}

// Update moves interceptor toward its target. Returns true if reached.
func (i *Interceptor) Update(cities []*City) bool {
	if i.TargetUFO != nil {
		if !i.TargetUFO.Active {
			i.TargetUFO = nil
			i.Launching = false
			return false
		}
		
		// Combat mode behavior
		switch i.Mode {
		case data.CombatAttack:
			// Aggressive: close to optimal firing range (30% of max)
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, 0.3)
		case data.CombatCautious:
			// Balanced: maintain medium distance (50% of max)
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, 0.5)
		case data.CombatBreakoff:
			// Defensive: keep distance, disengage if HP low
			if i.HP < i.MaxHP/3 {
				i.Disengage()
				return false
			}
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, 0.7)
		}
		
		// Default: chase directly
		return i.moveTo(i.TargetUFO.X, i.TargetUFO.Y)
	}

	if i.TargetNode >= 0 {
		var target *City
		for _, c := range cities {
			if c.ID == i.TargetNode {
				target = c
				break
			}
		}
		if target == nil {
			i.Launching = false
			return false
		}
		tx := float64(target.X)
		ty := float64(target.Y)
		reached := i.moveTo(tx, ty)
		if reached {
			// Check if any UFOs are at this city
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

// moveToWithTarget moves toward target but stops at a percentage of max range.
func (i *Interceptor) moveToWithTarget(tx, ty float64, rangeFraction float64) bool {
	dx := tx - i.X
	dy := ty - i.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	
	// Stop at fraction of weapon range
	stopDist := float64(i.Range) * rangeFraction
	
	if dist <= stopDist+0.5 {
		return true
	}

	speed := float64(i.Speed) * 0.015
	if speed > dist-stopDist {
		speed = dist - stopDist
	}
	if speed < 0 {
		speed = 0
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
	if i.State != nil {
		i.State.HP = i.HP
		if i.HP <= 0 {
			i.State.Status = "Destroyed"
		} else {
			i.State.Status = "Available"
		}
	}
}

func (i *Interceptor) FireAt(ufo *UFO) int {
	if i.Ammo <= 0 {
		return 0
	}
	
	// Check range - adjust accuracy based on distance
	dist := math.Sqrt(math.Pow(ufo.X-i.X, 2)+math.Pow(ufo.Y-i.Y, 2))
	rangeRatio := dist / float64(i.Range)
	
	// Accuracy penalty beyond optimal range
	accuracy := i.EffectiveAccuracy()
	if rangeRatio > 0.7 {
		// Beyond 70% of max range, accuracy drops
		accuracy = int(float64(accuracy) * (1.0 - (rangeRatio-0.7)*1.5))
	}
	
	// Combat mode adjustments
	switch i.Mode {
	case data.CombatAttack:
		// Aggressive: +10% accuracy, -5 range
		accuracy += 10
	case data.CombatBreakoff:
		// Defensive: -10% accuracy, but can disengage
		accuracy -= 10
	}
	
	// Clamp accuracy
	if accuracy < 10 {
		accuracy = 10
	}
	if accuracy > 100 {
		accuracy = 100
	}
	
	// Fire
	i.Ammo--
	
	// Hit check
	if rand.Intn(100) >= accuracy {
		return 0 // miss
	}
	
	// Calculate damage with some variance
	damage := i.Weapon.Damage + rand.Intn(i.Weapon.Damage/3+1)
	
	// Critical hit chance (10%)
	if rand.Intn(100) < 10 {
		damage = damage * 3 / 2
	}
	
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
		if i.HP > 0 && i.Launching {
			active = append(active, i)
		}
	}
	return active
}
