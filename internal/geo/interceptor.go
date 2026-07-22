package geo

import (
	"math"
	"math/rand"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
)

const (
	defaultInterceptorSpeed = 36
	defaultInterceptorHP    = 60
	defaultPilotSkill       = 50

	ammoPerFireRate = 4

	fuelRangeMultiplier = 3

	rangeFractionAttack   = 0.3
	rangeFractionCautious = 0.5
	rangeFractionBreakoff = 0.7

	breakoffHPDivisor = 3

	interceptorSpeedScale = 0.015

	maxTrailLength  = 30
	arrivalThreshold = 1.5

	modeAccuracyAttackBonus  = 10
	modeAccuracyBreakoffPenalty = 10
	accuracyMin              = 10
	accuracyMax              = 100

	damageVarianceDivisor = 3

	critChancePct    = 10
	critMultiplierNum = 3
	critMultiplierDen = 2

	effectiveRangeRatioThreshold = 0.7
	rangeFalloffMultiplier       = 1.5
)

type TrailPoint struct {
	X, Y float64
}

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
	Trail      []TrailPoint
	TrailTick  int // throttle counter to avoid dense trails
}

func NewInterceptor(baseX, baseY int) *Interceptor {
	w := data.InterceptorWeapons["avalanche"]
	return &Interceptor{
		Name:       language.String("INTERCEPTOR_DEFAULT_NAME"),
		X:          float64(baseX),
		Y:          float64(baseY),
		Speed:       defaultInterceptorSpeed,
		HP:          defaultInterceptorHP,
		MaxHP:       defaultInterceptorHP,
		WeaponKey:  "avalanche",
		Weapon:     w,
		Ammo:       w.FireRate * ammoPerFireRate,
		Range:      w.Range,
		TargetNode: -1,
		Mode:       data.CombatCautious,
		PilotSkill: defaultPilotSkill,
	}
}

// NewInterceptorFromState creates an interceptor from a persisted state.
func NewInterceptorFromState(s *data.InterceptorState, baseX, baseY int) *Interceptor {
	w, ok := data.InterceptorWeapons[s.WeaponKey]
	if !ok {
		w = data.InterceptorWeapons["avalanche"]
	}
	return &Interceptor{
		Name:       s.Name,
		X:          float64(baseX),
		Y:          float64(baseY),
		Speed:       defaultInterceptorSpeed,
		HP:         s.HP,
		MaxHP:      s.MaxHP,
		WeaponKey:  s.WeaponKey,
		Weapon:     w,
		Ammo:       s.Ammo,
		Range:      w.Range,
		TargetNode: -1,
		Mode:       data.CombatCautious,
		PilotSkill: defaultPilotSkill,
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
	i.Ammo = w.FireRate * ammoPerFireRate
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
		i.RangeLeft = i.Range * fuelRangeMultiplier
		return
	}
}
}

// LaunchAtUFO sends interceptor to pursue a specific UFO.
func (i *Interceptor) LaunchAtUFO(ufo *UFO) {
	i.TargetUFO = ufo
	i.TargetNode = -1
	i.Launching = true
	i.RangeLeft = i.Range * fuelRangeMultiplier
}

// Update moves interceptor toward its target. Returns true if reached.
func (i *Interceptor) Update(cities []*City, ufos UFOList) bool {
	if i.TargetUFO != nil {
		if !i.TargetUFO.Active {
			// Target UFO is gone (destroyed by another craft, escaped, or
			// despawned). Return to base and free the hangar slot, otherwise
			// the interceptor would remain stuck in "active" status forever.
			i.Disengage()
			return false
		}
		
		// Combat mode behavior
		switch i.Mode {
		case data.CombatAttack:
			// Aggressive: close to optimal firing range (30% of max)
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, rangeFractionAttack)
		case data.CombatCautious:
			// Balanced: maintain medium distance (50% of max)
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, rangeFractionCautious)
		case data.CombatBreakoff:
			// Defensive: keep distance, disengage if HP low
			if i.HP < i.MaxHP/breakoffHPDivisor {
				i.Disengage()
				return false
			}
			return i.moveToWithTarget(i.TargetUFO.X, i.TargetUFO.Y, rangeFractionBreakoff)
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
			i.Disengage()
			return false
		}
		tx := float64(target.X)
		ty := float64(target.Y)
		reached := i.moveTo(tx, ty)
		if reached {
			// Check if any UFOs are at this city
			for _, ufo := range ufos.Active() {
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
func (i *Interceptor) moveStep(tx, ty, maxDist float64) (dx, dy, dist float64, reached bool) {
	dx = tx - i.X
	dy = ty - i.Y
	dist = math.Sqrt(dx*dx + dy*dy)
	if maxDist >= 0 && dist <= maxDist+0.5 {
		return dx, dy, dist, true
	}
	speed := float64(i.Speed) * interceptorSpeedScale
	if maxDist >= 0 && speed > dist-maxDist {
		speed = dist - maxDist
	}
	if speed < 0 {
		speed = 0
	}
	if dist > 0 {
		i.X += (dx / dist) * speed
		i.Y += (dy / dist) * speed
		i.recordTrail()
	}
	i.RangeLeft--
	if i.RangeLeft <= 0 {
		// Out of fuel: break off and return to base so the craft becomes
		// available again instead of being stranded in "active" status.
		i.Disengage()
		return dx, dy, dist, false
	}
	return dx, dy, dist, false
}

func (i *Interceptor) moveToWithTarget(tx, ty float64, rangeFraction float64) bool {
	stopDist := float64(i.Range) * rangeFraction
	_, _, _, reached := i.moveStep(tx, ty, stopDist)
	return reached
}

func (i *Interceptor) recordTrail() {
	i.TrailTick++
	if i.TrailTick%3 != 0 {
		return
	}
	pt := TrailPoint{X: i.X, Y: i.Y}
	if len(i.Trail) > 0 {
		last := i.Trail[len(i.Trail)-1]
		dx := pt.X - last.X
		dy := pt.Y - last.Y
		if dx*dx+dy*dy < 0.5 {
			return
		}
	}
	i.Trail = append(i.Trail, pt)
	if len(i.Trail) > maxTrailLength {
		i.Trail = i.Trail[1:]
	}
}

func (i *Interceptor) moveTo(tx, ty float64) bool {
	_, _, _, reached := i.moveStep(tx, ty, arrivalThreshold)
	return reached
}

func (i *Interceptor) Disengage() {
	i.TargetNode = -1
	i.TargetUFO = nil
	i.Launching = false
	i.Trail = nil
	i.TrailTick = 0
	if i.State != nil {
		i.State.HP = i.HP
		// Auto-rearm on return to base
		i.State.Ammo = i.Weapon.FireRate * ammoPerFireRate
		if i.HP <= 0 {
			i.State.Status = "destroyed"
		} else {
			i.State.Status = "available"
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
	if rangeRatio > effectiveRangeRatioThreshold {
		// Beyond 70% of max range, accuracy drops
		accuracy = int(float64(accuracy) * (1.0 - (rangeRatio-effectiveRangeRatioThreshold)*rangeFalloffMultiplier))
	}

	// Combat mode adjustments
	switch i.Mode {
	case data.CombatAttack:
		// Aggressive: +10% accuracy, -5 range
		accuracy += modeAccuracyAttackBonus
	case data.CombatBreakoff:
		// Defensive: -10% accuracy, but can disengage
		accuracy -= modeAccuracyBreakoffPenalty
	}

	// Clamp accuracy
	if accuracy < accuracyMin {
		accuracy = accuracyMin
	}
	if accuracy > accuracyMax {
		accuracy = accuracyMax
	}
	
	// Fire
	i.Ammo--
	
	// Hit check
	if rand.Intn(100) >= accuracy {
		return 0 // miss
	}
	
	// Calculate damage with some variance
	damage := i.Weapon.Damage + rand.Intn(i.Weapon.Damage/damageVarianceDivisor+1)

	// Critical hit chance (10%)
	if rand.Intn(100) < critChancePct {
		damage = damage * critMultiplierNum / critMultiplierDen
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
