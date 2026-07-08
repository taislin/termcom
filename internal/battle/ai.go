package battle

import (
	"math"
	"math/rand"
)

type AIState int

const (
	AIIdle AIState = iota
	AIPatrol
	AISearch
	AIAttack
	AIFlee
)

type AlienAI struct {
	Unit       *Unit
	State      AIState
	PatrolX    int
	PatrolY    int
	LastSeenX  int
	LastSeenY  int
	TurnsSince int
}

func NewAlienAI(u *Unit) *AlienAI {
	return &AlienAI{
		Unit:  u,
		State: AIPatrol,
	}
}

func (ai *AlienAI) GenerateActions(units UnitList, m *BattleMap, humanUnits UnitList) []AlienAction {
	if !ai.Unit.Alive {
		return nil
	}

	var actions []AlienAction
	nearest, dist := ai.findNearest(humanUnits, m)

	switch ai.State {
	case AIPatrol:
		if nearest != nil && dist < 15 {
			ai.State = AIAttack
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0
		} else {
			px, py := ai.patrolTarget(m)
			actions = append(actions, AlienAction{
				Type: "patrol", Unit: ai.Unit,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: px, ToY: py,
			})
		}

	case AIAttack:
		if nearest != nil && dist < 20 {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0

			if dist <= 1 {
				actions = append(actions, AlienAction{
					Type: "melee", Unit: ai.Unit, Target: nearest,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: nearest.X, ToY: nearest.Y,
				})
			} else if ai.Unit.TU >= 15 {
				actions = append(actions, AlienAction{
					Type: "fire", Unit: ai.Unit, Target: nearest,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: nearest.X, ToY: nearest.Y,
				})
			}

			if ai.Unit.AlienType != nil && ai.Unit.AlienType.Aggression > 5 && dist > 3 {
				nx, ny := ai.moveTowardTarget(nearest.X, nearest.Y, m)
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: nx, ToY: ny,
				})
			}
		} else {
			ai.TurnsSince++
			if ai.TurnsSince > 3 {
				ai.State = AISearch
				ai.TurnsSince = 0
			}
		}

	case AISearch:
		if nearest != nil {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.State = AIAttack
			ai.TurnsSince = 0
		} else {
			nx, ny := ai.moveTowardTarget(ai.LastSeenX, ai.LastSeenY, m)
			actions = append(actions, AlienAction{
				Type: "move", Unit: ai.Unit,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: nx, ToY: ny,
			})
			ai.TurnsSince++
			if ai.TurnsSince > 5 {
				ai.State = AIPatrol
			}
		}

	case AIFlee:
		if nearest != nil {
			dx := ai.Unit.X - nearest.X
			dy := ai.Unit.Y - nearest.Y
			fx := ai.Unit.X + signum(dx)*2
			fy := ai.Unit.Y + signum(dy)*2
			actions = append(actions, AlienAction{
				Type: "move", Unit: ai.Unit,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: fx, ToY: fy,
			})
		}
		ai.TurnsSince++
		if ai.TurnsSince > 3 {
			ai.State = AIPatrol
		}
	}

	if ai.Unit.HP < ai.Unit.MaxHP/4 && ai.Unit.Alive {
		if ai.Unit.AlienType != nil && ai.Unit.AlienType.Bravery < 50 {
			ai.State = AIFlee
			ai.TurnsSince = 0
		}
	}

	return actions
}

func (ai *AlienAI) Update(units UnitList, m *BattleMap, humanUnits UnitList) {
	if !ai.Unit.Alive {
		return
	}

	nearest, dist := ai.findNearest(humanUnits, m)

	switch ai.State {
	case AIPatrol:
		if nearest != nil && dist < 15 {
			ai.State = AIAttack
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0
		} else {
			ai.patrol(m)
		}

	case AIAttack:
		if nearest != nil && dist < 20 {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0

			if dist <= 1 {
				ai.meleeAttack(nearest)
			} else if ai.Unit.TU >= 15 {
				_, _, _ = ai.Unit.FireAt(nearest)
			}

			if ai.Unit.AlienType != nil && ai.Unit.AlienType.Aggression > 5 && dist > 3 {
				ai.moveToward(nearest.X, nearest.Y, m)
			}
		} else {
			ai.TurnsSince++
			if ai.TurnsSince > 3 {
				ai.State = AISearch
				ai.TurnsSince = 0
			}
		}

	case AISearch:
		if nearest != nil {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.State = AIAttack
			ai.TurnsSince = 0
		} else {
			ai.moveToward(ai.LastSeenX, ai.LastSeenY, m)
			ai.TurnsSince++
			if ai.TurnsSince > 5 {
				ai.State = AIPatrol
			}
		}

	case AIFlee:
		if nearest != nil {
			dx := ai.Unit.X - nearest.X
			dy := ai.Unit.Y - nearest.Y
			fx := ai.Unit.X + signum(dx)*2
			fy := ai.Unit.Y + signum(dy)*2
			ai.Unit.MoveTo(fx, fy, m)
		}
		ai.TurnsSince++
		if ai.TurnsSince > 3 {
			ai.State = AIPatrol
		}
	}

	if ai.Unit.HP < ai.Unit.MaxHP/4 && ai.Unit.Alive {
		if ai.Unit.AlienType != nil && ai.Unit.AlienType.Bravery < 50 {
			ai.State = AIFlee
			ai.TurnsSince = 0
		}
	}
}

func (ai *AlienAI) patrolTarget(m *BattleMap) (int, int) {
	if ai.PatrolX == 0 && ai.PatrolY == 0 {
		ai.PatrolX = ai.Unit.X + rand.Intn(10) - 5
		ai.PatrolY = ai.Unit.Y + rand.Intn(10) - 5
		if ai.PatrolX < 1 {
			ai.PatrolX = 1
		}
		if ai.PatrolY < 1 {
			ai.PatrolY = 1
		}
		if ai.PatrolX >= m.Width-1 {
			ai.PatrolX = m.Width - 2
		}
		if ai.PatrolY >= m.Height-1 {
			ai.PatrolY = m.Height - 2
		}
	}
	return ai.PatrolX, ai.PatrolY
}

func (ai *AlienAI) patrol(m *BattleMap) {
	if ai.PatrolX == 0 && ai.PatrolY == 0 {
		ai.PatrolX = ai.Unit.X + rand.Intn(10) - 5
		ai.PatrolY = ai.Unit.Y + rand.Intn(10) - 5
		if ai.PatrolX < 1 {
			ai.PatrolX = 1
		}
		if ai.PatrolY < 1 {
			ai.PatrolY = 1
		}
		if ai.PatrolX >= m.Width-1 {
			ai.PatrolX = m.Width - 2
		}
		if ai.PatrolY >= m.Height-1 {
			ai.PatrolY = m.Height - 2
		}
	}

	if !ai.Unit.MoveTo(ai.PatrolX, ai.PatrolY, m) {
		ai.PatrolX = 0
		ai.PatrolY = 0
	}

	dx := ai.PatrolX - ai.Unit.X
	dy := ai.PatrolY - ai.Unit.Y
	if dx*dx+dy*dy < 4 {
		ai.PatrolX = 0
		ai.PatrolY = 0
	}
}

func (ai *AlienAI) moveTowardTarget(tx, ty int, m *BattleMap) (int, int) {
	dx := tx - ai.Unit.X
	dy := ty - ai.Unit.Y
	nx := ai.Unit.X
	ny := ai.Unit.Y
	if dx > 0 {
		nx++
	} else if dx < 0 {
		nx--
	}
	if dy > 0 {
		ny++
	} else if dy < 0 {
		ny--
	}
	return nx, ny
}

func (ai *AlienAI) moveToward(tx, ty int, m *BattleMap) {
	dx := tx - ai.Unit.X
	dy := ty - ai.Unit.Y
	nx := ai.Unit.X
	ny := ai.Unit.Y
	if dx > 0 {
		nx++
	} else if dx < 0 {
		nx--
	}
	if dy > 0 {
		ny++
	} else if dy < 0 {
		ny--
	}
	ai.Unit.MoveTo(nx, ny, m)
}

func (ai *AlienAI) meleeAttack(target *Unit) {
	if ai.Unit.TU < 10 {
		return
	}
	ai.Unit.TU -= 10
	damage := ai.Unit.Strength + rand.Intn(10)
	damage -= target.Armour
	if damage < 1 {
		damage = 1
	}
	target.HP -= damage
	if target.HP <= 0 {
		target.Alive = false
	}
}

func (ai *AlienAI) findNearest(humanUnits UnitList, m *BattleMap) (*Unit, float64) {
	var nearest *Unit
	bestDist := 999.0
	for _, h := range humanUnits {
		if !h.Alive {
			continue
		}
		if !ai.Unit.CanSee(h.X, h.Y, m) {
			continue
		}
		dx := float64(h.X - ai.Unit.X)
		dy := float64(h.Y - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < bestDist {
			bestDist = dist
			nearest = h
		}
	}
	return nearest, bestDist
}

func signum(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
