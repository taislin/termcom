package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/soldier"
)

type Unit struct {
	Type      int
	Soldier   *soldier.Soldier
	AlienType *data.AlienType
	X, Y      int
	HP        int
	MaxHP     int
	TU        int
	MaxTU     int
	Accuracy  int
	Bravery   int
	Reactions int
	Strength  int
	Armour    int
	Weapon    string
	WeaponAmmo int
	Alive     bool
	Crouching bool
	Faction   int
}

func NewSoldierUnit(s *soldier.Soldier) *Unit {
	return &Unit{
		Type:      0,
		Soldier:   s,
		HP:        s.HP,
		MaxHP:     s.MaxHP,
		TU:        s.TU,
		MaxTU:     s.MaxTU,
		Accuracy:  s.Accuracy,
		Bravery:   s.Bravery,
		Reactions: s.Reactions,
		Strength:  s.Strength,
		Armour:    data.Armors[s.Armor].Undersuit,
		Weapon:    s.Weapon,
		WeaponAmmo: s.WeaponAmmo,
		Alive:     true,
		Faction:   0,
	}
}

func NewAlienUnit(at *data.AlienType) *Unit {
	return &Unit{
		Type:      1,
		AlienType: at,
		HP:        at.HP,
		MaxHP:     at.HP,
		TU:        at.TU,
		MaxTU:     at.TU,
		Accuracy:  at.Accuracy,
		Bravery:   at.Bravery,
		Reactions: at.Reactions,
		Strength:  at.Strength,
		Armour:    at.Armour,
		Weapon:    at.Weapon,
		WeaponAmmo: data.RuleItems[at.Weapon].AmmoMax,
		Alive:     true,
		Faction:   1,
	}
}

func (u *Unit) Name() string {
	if u.Soldier != nil {
		return u.Soldier.Name
	}
	if u.AlienType != nil {
		return u.AlienType.Name
	}
	return "Unknown"
}

func (u *Unit) FireAt(target *Unit) (int, bool, error) {
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

	if rand.Intn(100) >= hitChance {
		return 0, false, nil
	}

	damage := w.Damage + rand.Intn(w.Damage/3+1)
	damage -= target.Armour
	if target.Crouching {
		damage = damage * 7 / 10
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

func (ul UnitList) Alive() []*Unit {
	var alive []*Unit
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
