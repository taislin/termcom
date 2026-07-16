package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
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
	Faction     int
	IsNight     bool
	ReservedTU  int
	FatalWounds int
	BleedRate   int
	Morale      int
	HasMoved    bool
	InOverwatch bool
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
	w, ok := data.RuleItems[u.Weapon]
	if !ok {
		return 0, false, false, fmt.Errorf("unknown weapon: %s", u.Weapon)
	}
	if u.TU < w.TU {
		return 0, false, false, fmt.Errorf("not enough TU")
	}
	if u.WeaponAmmo <= 0 && w.AmmoMax < 99 {
		return 0, false, false, fmt.Errorf("out of ammo")
	}
	if w.AmmoMax < 99 {
		u.WeaponAmmo--
	}
	u.TU -= w.TU

	dist := math.Sqrt(float64((target.X-u.X)*(target.X-u.X) + (target.Y-u.Y)*(target.Y-u.Y)))
	accMod := 100 - int(dist*3)
	if accMod < 10 {
		accMod = 10
	}
	hitChance := u.Accuracy * accMod / 100
	if u.Crouching {
		hitChance = hitChance * 110 / 100
	}
	if u.IsNight {
		hitChance = hitChance * 75 / 100
	}
	if weather != nil {
		hitChance -= weather.AccuracyPenalty()
	}

	if u.Soldier != nil {
		if dist > 8 && u.Soldier.HasBattleMod(soldier.BModMarksman) {
			hitChance = hitChance * 115 / 100
		}
		if dist <= 4 && u.Soldier.HasBattleMod(soldier.BModCloseCombat) {
			hitChance = hitChance * 115 / 100
		}
		if !u.HasMoved && u.Soldier.HasBattleMod(soldier.BModSteadyAim) {
			hitChance = hitChance * 110 / 100
		}
		if u.InOverwatch && u.Soldier.HasBattleMod(soldier.BModOverwatch) {
			hitChance = hitChance * 120 / 100
		}
	}

	if rand.Intn(100) >= hitChance {
		return 0, false, false, nil
	}

	// Adjacent cover check: when the target is within 1 tile and standing in
	// partial cover, the cover value acts as a block chance rather than a
	// silent damage reduction. A successful block stops the shot entirely.
	if m != nil && dist <= 1.5 {
		if tc := m.At(target.X, target.Y).Cover; tc > 0 {
			if rand.Intn(100) < tc {
				return 0, false, true, nil
			}
		}
	}

	damage := w.Damage + rand.Intn(w.Damage/3+1)

	if u.Weapon == "stun_rod" {
		target.StunPoints += damage
		if target.StunPoints >= target.MaxHP {
			target.Stunned = true
		}
		if u.Soldier != nil {
			u.Soldier.AddMeleeExp()
		}
		return damage, true, false, nil
	}

	// Apply cover from intervening tiles. Cover acts as a block chance rather
	// than a silent damage reduction: a 100% obstacle always blocks, while
	// partial cover only has a 1/3 chance of stopping the shot (a 30% bush
	// blocks ~10% of the time). A shot that gets through deals full damage.
	cover := 0
	if m != nil {
		cover = m.CoverAlongLine(u.X, u.Y, target.X, target.Y)
	}
	if cover >= 100 {
		return 0, false, true, nil
	}
	if cover > 0 && rand.Intn(100) < cover/3 {
		return 0, false, true, nil
	}

	damage -= target.Armour
	if target.Crouching {
		damage = damage * 7 / 10
	}

	// Apply damage type resistance/weakness from target
	weapDMG := WeaponDamageType(u.Weapon)
	if target.AlienType != nil {
		resist := target.AlienType.Resist(weapDMG)
		if resist > 0 {
			damage = damage * (100 - resist) / 100
		} else if resist < 0 {
			damage = damage * (100 - resist) / 100
		}
	}

	if damage < 1 {
		damage = 1
	}
	target.HP -= damage

	if u.Soldier != nil {
		switch weapDMG {
		case data.DMG_MELEE:
			u.Soldier.AddMeleeExp()
		case data.DMG_EXPLOSIVE:
			u.Soldier.AddThrowingExp()
		default:
			u.Soldier.AddFiringExp()
		}
	}

	if target.HP <= 0 {
		target.Alive = false
	} else if rand.Intn(100) < 15 {
		target.FatalWounds++
		target.BleedRate += damage / 4
		if target.BleedRate > 5 {
			target.BleedRate = 5
		}
	}
	return damage, true, false, nil
}

// WeaponDamageType returns the damage type for a given weapon ID.
func WeaponDamageType(weapon string) int {
	switch weapon {
	case "plasma_pistol", "plasma_rifle", "heavy_plasma", "alien_grenade", "alien_rocket":
		return data.DMG_PLASMA
	case "laser_pistol", "laser_rifle", "alien_laser", "alien_heavy_laser":
		return data.DMG_LASER
	case "rocket":
		return data.DMG_EXPLOSIVE
	case "chryssalid_claw", "reaper_claw", "stun_rod", "alien_claw", "alien_fang":
		return data.DMG_MELEE
	case "alien_psi_bolt":
		return data.DMG_PSIONIC
	default:
		return data.DMG_KINETIC
	}
}

func (u *Unit) MoveTo(x, y int, m *BattleMap) bool {
	dist := math.Abs(float64(x-u.X)) + math.Abs(float64(y-u.Y))
	tuCost := int(dist) * 4
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
