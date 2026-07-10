package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/soldier"
)

type Unit struct {
	Type       int
	Soldier    *soldier.Soldier
	AlienType  *data.AlienType
	CivName    string
	X, Y       int
	Level      int
	HP         int
	MaxHP      int
	TU         int
	MaxTU      int
	Accuracy   int
	Bravery    int
	Reactions  int
	Strength   int
	PsiSkill   int
	PsiStr     int
	Armour     int
	Weapon     string
	WeaponAmmo int
	Alive      bool
	Stunned    bool
	StunPoints int
	Crouching  bool
	Panicked   bool
	Faction    int
	IsNight    bool
	ReservedTU int
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
	}
}

var civNames = []string{
	"Alex", "Sam", "Jordan", "Casey", "Morgan", "Taylor", "Riley", "Quinn",
	"Drew", "Jamie", "Robin", "Pat", "Terry", "Leslie", "Sandy", "Dee",
	"Lee", "Kim", "Avery", "Reese", "Dakota", "Skyler", "Blair", "Emery",
}

func NewCivilianUnit(name string) *Unit {
	return &Unit{
		Type:     2,
		CivName:  name,
		HP:       5,
		MaxHP:    5,
		TU:       20,
		MaxTU:    20,
		Accuracy: 0,
		Bravery:  30,
		Reactions: 0,
		Strength: 5,
		Armour:   0,
		Weapon:   "",
		Alive:    true,
		Faction:  2,
	}
}

func (u *Unit) Name() string {
	if u.Soldier != nil {
		return u.Soldier.Name
	}
	if u.AlienType != nil {
		return u.AlienType.Name
	}
	if u.CivName != "" {
		return u.CivName
	}
	return "Civilian"
}

func (u *Unit) FireAt(target *Unit, m *BattleMap) (int, bool, error) {
	w, ok := data.RuleItems[u.Weapon]
	if !ok {
		return 0, false, fmt.Errorf("unknown weapon: %s", u.Weapon)
	}
	if u.TU < w.TU {
		return 0, false, fmt.Errorf("not enough TU")
	}
	if u.WeaponAmmo <= 0 && w.AmmoMax < 99 {
		return 0, false, fmt.Errorf("out of ammo")
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

	if rand.Intn(100) >= hitChance {
		return 0, false, nil
	}

	damage := w.Damage + rand.Intn(w.Damage/3+1)
	
	if u.Weapon == "stun_rod" {
		target.StunPoints += damage
		if target.StunPoints >= target.MaxHP {
			target.Stunned = true
		}
		return damage, true, nil
	}
	
	damage -= target.Armour
	if target.Crouching {
		damage = damage * 7 / 10
	}

	// Apply cover from intervening tiles
	if m != nil {
		cover := m.CoverAlongLine(u.X, u.Y, target.X, target.Y)
		if cover > 0 {
			damage = damage * (100 - cover) / 100
		}
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
	if target.HP <= 0 {
		target.Alive = false
	}
	return damage, true, nil
}

// WeaponDamageType returns the damage type for a given weapon ID.
func WeaponDamageType(weapon string) int {
	switch weapon {
	case "plasma_pistol", "plasma_rifle", "heavy_plasma", "alien_grenade":
		return data.DMG_PLASMA
	case "laser_pistol", "laser_rifle":
		return data.DMG_LASER
	case "rocket":
		return data.DMG_EXPLOSIVE
	case "chryssalid_claw", "reaper_claw", "stun_rod":
		return data.DMG_MELEE
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
