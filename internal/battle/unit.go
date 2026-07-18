package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
)

// Faction defines the team allegiance of a unit.
type Faction int

const (
	FactionHuman    Faction = 0
	FactionAlien    Faction = 1
	FactionCivilian Faction = 2
)

// Unit combat / battle tuning constants.
const (
	InfAmmoThreshold    = 99    // AmmoMax at/above this means effectively infinite ammo
	distAccPenalty      = 3     // accuracy loss per tile of distance
	minAccMod           = 10    // floor for distance-based accuracy modifier
	minHitChance        = 5     // floor for final hit chance
	crouchAimBonus      = 110   // % aim multiplier when crouched
	nightAimMult        = 75    // % aim multiplier at night
	marksmanDist        = 8     // min range for marksman bonus
	marksmanAimBonus    = 115   // % aim multiplier for marksman
	closeCombatDist     = 4     // max range for close-combat bonus
	closeCombatAimBonus = 115   // % aim multiplier for close combat
	steadyAimBonus      = 110   // % aim multiplier for steady aim
	overwatchAimBonus   = 120   // % aim multiplier for overwatch
	meleeCoverDist      = 1.5   // distance threshold for melee cover check
	crouchDmgReduce     = 7     // crouching damage reduction numerator (of 10)
	fatalWoundChance    = 15    // % chance per hit to inflict a fatal wound
	bleedDivisor        = 4     // bleed rate = damage / bleedDivisor
	maxBleedRate        = 5     // cap on bleed rate
)

type Unit struct {
	Type        int
	Soldier     *soldier.Soldier
	AlienType   *data.AlienType
	CivName     string
	X, Y        int
	Level       int
	HP          int
	MaxHP       int
	TU          int
	MaxTU       int
	Accuracy    int
	Bravery     int
	Reactions   int
	Strength    int
	PsiSkill    int
	PsiStr      int
	Armour      int
	Weapon      string
	WeaponAmmo  int
	Alive       bool
	Stunned     bool
	StunPoints  int
	Crouching   bool
	Panicked    bool
	Faction     Faction
	IsNight     bool
	ReservedTU  int
	FatalWounds int
	BleedRate   int
	Morale      int
	HasMoved    bool
	InOverwatch bool
	FireMode    data.FireMode
}

func NewSoldierUnit(s *soldier.Soldier) *Unit {
	return &Unit{
		Type:       0,
		Soldier:    s,
		HP:         s.HP,
		MaxHP:      s.MaxHP,
		TU:         s.TU,
		MaxTU:      s.MaxTU,
		Accuracy:   s.Accuracy,
		Bravery:    s.Bravery,
		Reactions:  s.Reactions,
		Strength:   s.Strength,
		PsiSkill:   s.PsiSkill,
		PsiStr:     s.PsiStr,
		Armour:     data.Armors[s.Armor].Undersuit,
		Weapon:     s.Weapon,
		WeaponAmmo: s.WeaponAmmo,
		Alive:      true,
		Faction:    0,
		Morale:     100,
		FireMode:   data.FireModeAimed,
	}
}

func NewAlienUnit(at *data.AlienType) *Unit {
	return &Unit{
		Type:       1,
		AlienType:  at,
		HP:         at.HP,
		MaxHP:      at.HP,
		TU:         at.TU,
		MaxTU:      at.TU,
		Accuracy:   at.Accuracy,
		Bravery:    at.Bravery,
		Reactions:  at.Reactions,
		Strength:   at.Strength,
		PsiSkill:   at.Psi,
		PsiStr:     at.ResistPsionic,
		Armour:     at.Armour,
		Weapon:     at.Weapon,
		WeaponAmmo: data.RuleItems[at.Weapon].AmmoMax,
		Alive:      true,
		Faction:    1,
		Morale:     100,
	}
}

var civNames = []string{
	"Alex", "Sam", "Jordan", "Casey", "Morgan", "Taylor", "Riley", "Quinn",
	"Drew", "Jamie", "Robin", "Pat", "Terry", "Leslie", "Sandy", "Dee",
	"Lee", "Kim", "Avery", "Reese", "Dakota", "Skyler", "Blair", "Emery",
}

func NewCivilianUnit(name string) *Unit {
	return &Unit{
		Type:      2,
		CivName:   name,
		HP:        5,
		MaxHP:     5,
		TU:        20,
		MaxTU:     20,
		Accuracy:  0,
		Bravery:   30,
		Reactions: 0,
		Strength:  5,
		Armour:    0,
		Weapon:    "",
		Alive:     true,
		Faction:   2,
		Morale:    100,
	}
}

func (u *Unit) Name() string {
	if u.Soldier != nil {
		return u.Soldier.Name
	}
	if u.AlienType != nil {
		return u.AlienType.LangName()
	}
	if u.CivName != "" {
		return u.CivName
	}
	return language.String("MSG_CIVILIAN")
}

func (u *Unit) FireAt(target *Unit, m *BattleMap, weather *Weather) (int, bool, bool, error) {
	if target == nil {
		return 0, false, false, fmt.Errorf("no target")
	}
	if m != nil && (target.X < 0 || target.X >= m.Width || target.Y < 0 || target.Y >= m.Height) {
		return 0, false, false, fmt.Errorf("target out of bounds")
	}
	w, ok := data.RuleItems[u.Weapon]
	if !ok {
		return 0, false, false, fmt.Errorf("unknown weapon: %s", u.Weapon)
	}
	tuCost := w.ModeTU(u.FireMode)
	if u.TU < tuCost {
		return 0, false, false, fmt.Errorf("not enough TU")
	}
	rounds := w.ModeRounds(u.FireMode)
	if rounds < 0 {
		rounds = u.WeaponAmmo
	}
	if rounds <= 0 {
		rounds = 1
	}
	if w.AmmoMax < InfAmmoThreshold && u.WeaponAmmo < rounds {
		return 0, false, false, fmt.Errorf("out of ammo")
	}
	if w.AmmoMax < InfAmmoThreshold {
		u.WeaponAmmo -= rounds
	}
	u.TU -= tuCost

	dist := math.Sqrt(float64((target.X-u.X)*(target.X-u.X) + (target.Y-u.Y)*(target.Y-u.Y)))
	accMod := 100 - int(dist*distAccPenalty)
	if accMod < minAccMod {
		accMod = minAccMod
	}
	hitChance := u.Accuracy * accMod / 100
	hitChance -= w.ModeAccuracy(u.FireMode)
	if hitChance < minHitChance {
		hitChance = minHitChance
	}
	if u.Crouching {
		hitChance = hitChance * crouchAimBonus / 100
	}
	if u.IsNight {
		hitChance = hitChance * nightAimMult / 100
	}
	if weather != nil {
		hitChance -= weather.AccuracyPenalty()
	}

	if u.Soldier != nil {
		if dist > marksmanDist && u.Soldier.HasBattleMod(soldier.BModMarksman) {
			hitChance = hitChance * marksmanAimBonus / 100
		}
		if dist <= closeCombatDist && u.Soldier.HasBattleMod(soldier.BModCloseCombat) {
			hitChance = hitChance * closeCombatAimBonus / 100
		}
		if !u.HasMoved && u.Soldier.HasBattleMod(soldier.BModSteadyAim) {
			hitChance = hitChance * steadyAimBonus / 100
		}
		if u.InOverwatch && u.Soldier.HasBattleMod(soldier.BModOverwatch) {
			hitChance = hitChance * overwatchAimBonus / 100
		}
	}

	totalDamage, stunDamage, anyHit, anyCover := u.resolveHits(rounds, hitChance, w, target, m, dist)
	if u.Soldier != nil && anyHit {
		weapDMG := WeaponDamageType(u.Weapon)
		switch weapDMG {
		case data.DMG_MELEE:
			u.Soldier.AddMeleeExp()
		case data.DMG_EXPLOSIVE:
			u.Soldier.AddThrowingExp()
		default:
			u.Soldier.AddFiringExp()
		}
	}
	return totalDamage + stunDamage, anyHit, anyCover, nil
}

func (u *Unit) resolveHits(rounds, hitChance int, w data.RuleItem, target *Unit, m *BattleMap, dist float64) (totalDamage, stunDamage int, anyHit, anyCover bool) {
	for i := 0; i < rounds; i++ {
		if rand.Intn(100) >= hitChance {
			continue
		}
		if m != nil && dist <= meleeCoverDist {
			if tc := m.At(target.X, target.Y).Cover; tc > 0 {
				if rand.Intn(100) < tc {
					anyCover = true
					continue
				}
			}
		}
		dmg := w.Damage + rand.Intn(w.Damage/3+1)
		if u.Weapon == "stun_rod" {
			target.StunPoints += dmg
			if target.StunPoints >= target.MaxHP {
				target.Stunned = true
			}
			stunDamage += dmg
			anyHit = true
			continue
		}
		cover := 0
		if m != nil {
			cover = m.CoverAlongLine(u.X, u.Y, target.X, target.Y)
		}
		if cover >= 100 {
			anyCover = true
			continue
		}
		if cover > 0 && rand.Intn(100) < cover/3 {
			anyCover = true
			continue
		}
		dmg -= target.Armour
		if target.Crouching {
			dmg = dmg * crouchDmgReduce / 10
		}
		weapDMG := WeaponDamageType(u.Weapon)
		if target.AlienType != nil {
			resist := target.AlienType.Resist(weapDMG)
			if resist > 0 {
				dmg = dmg * (100 - resist) / 100
			} else if resist < 0 {
				dmg = dmg * (100 - resist) / 100
			}
		}
		if dmg < 1 {
			dmg = 1
		}
		totalDamage += dmg
		anyHit = true
		if rand.Intn(100) < fatalWoundChance {
			target.FatalWounds++
			target.BleedRate += dmg / bleedDivisor
			if target.BleedRate > maxBleedRate {
				target.BleedRate = maxBleedRate
			}
		}
	}
	if totalDamage > 0 {
		target.HP -= totalDamage
	}
	if target.HP <= 0 {
		target.HP = 0
		target.Alive = false
	}
	return
}

var weaponDamageMap = map[string]int{
	"plasma_pistol": data.DMG_PLASMA, "plasma_rifle": data.DMG_PLASMA, "heavy_plasma": data.DMG_PLASMA, "alien_grenade": data.DMG_PLASMA, "alien_rocket": data.DMG_PLASMA,
	"laser_pistol": data.DMG_LASER, "laser_rifle": data.DMG_LASER, "alien_laser": data.DMG_LASER, "alien_heavy_laser": data.DMG_LASER,
	"rocket":        data.DMG_EXPLOSIVE,
	"chryssalid_claw": data.DMG_MELEE, "reaper_claw": data.DMG_MELEE, "stun_rod": data.DMG_MELEE, "alien_claw": data.DMG_MELEE, "alien_fang": data.DMG_MELEE,
	"alien_psi_bolt": data.DMG_PSIONIC,
}

// WeaponDamageType returns the damage type for a given weapon ID.
func WeaponDamageType(weapon string) int {
	if dmg, ok := weaponDamageMap[weapon]; ok {
		return dmg
	}
	return data.DMG_KINETIC
}

func (u *Unit) MoveTo(x, y int, m *BattleMap) bool {
	// Per-step TU cost varies by terrain; sum along a Manhattan path.
	dx := x - u.X
	if dx < 0 {
		dx = -dx
	}
	dy := y - u.Y
	if dy < 0 {
		dy = -dy
	}
	sx, sy := 1, 1
	if x < u.X {
		sx = -1
	}
	if y < u.Y {
		sy = -1
	}
	// Walk the longer axis first, then the shorter (Manhattan order).
	cx, cy := u.X, u.Y
	stepCost := 0
	for i := 0; i < dx; i++ {
		cx += sx
		stepCost += m.MoveCost(cx, cy, nil)
	}
	for i := 0; i < dy; i++ {
		cy += sy
		stepCost += m.MoveCost(cx, cy, nil)
	}
	tuCost := stepCost
	if u.Crouching {
		tuCost += 4
	}
	if tuCost > u.TU {
		return false
	}
	if !m.Passable(x, y) {
		return false
	}
	u.X = x
	u.Y = y
	u.TU -= tuCost
	u.HasMoved = true
	return true
}

func (u *Unit) CanSee(tx, ty int, m *BattleMap) bool {
	if u.Level != m.CurrentLevel {
		return false
	}
	dx := tx - u.X
	dy := ty - u.Y
	absDx := dx
	absDy := dy
	if absDx < 0 {
		absDx = -absDx
	}
	if absDy < 0 {
		absDy = -absDy
	}

	sx := 1
	if dx < 0 {
		sx = -1
	}
	sy := 1
	if dy < 0 {
		sy = -1
	}
	err := absDx - absDy

	x := u.X
	y := u.Y
	for {
		if x == tx && y == ty {
			return true
		}
		if m != nil && m.Opaque(x, y) && !(x == u.X && y == u.Y) {
			return false
		}
		e2 := 2 * err
		if e2 > -absDy {
			err -= absDy
			x += sx
		}
		if e2 < absDx {
			err += absDx
			y += sy
		}
	}
}

type UnitList []*Unit

func (ul UnitList) Alive() UnitList {
	var alive UnitList
	for _, u := range ul {
		if u.Alive {
			alive = append(alive, u)
		}
	}
	return alive
}

func (ul UnitList) At(x, y int) *Unit {
	for _, u := range ul {
		if u.Alive && u.X == x && u.Y == y {
			return u
		}
	}
	return nil
}

func (ul UnitList) OnLevel(level int) []*Unit {
	var result []*Unit
	for _, u := range ul {
		if u.Alive && u.Level == level {
			result = append(result, u)
		}
	}
	return result
}
