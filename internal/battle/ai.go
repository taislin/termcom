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
	AIFlank
	AIRetreat
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

	case AIFlank:
		if nearest != nil {
			// Move laterally
			dx := nearest.X - ai.Unit.X
			dy := nearest.Y - ai.Unit.Y
			// Lateral move: swap dx/dy and negate one
			nx := ai.Unit.X + signum(-dy)
			ny := ai.Unit.Y + signum(dx)
			actions = append(actions, AlienAction{
				Type: "move", Unit: ai.Unit,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: nx, ToY: ny,
			})
		}
		ai.TurnsSince++
		if ai.TurnsSince > 2 {
			ai.State = AIAttack
		}

	case AIRetreat:
		if nearest != nil {
			// Move away from nearest human
			dx := ai.Unit.X - nearest.X
			dy := ai.Unit.Y - nearest.Y
			fx := ai.Unit.X + signum(dx)*3
			fy := ai.Unit.Y + signum(dy)*3
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
			ai.State = AIRetreat
			ai.TurnsSince = 0
		}
	}

	// Check for flank opportunities (e.g. target is crouched)
	if ai.State == AIAttack && nearest != nil && nearest.Crouching {
		if ai.Unit.AlienType != nil && ai.Unit.AlienType.Aggression > 7 {
			ai.State = AIFlank
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

	// Check for flank opportunities (e.g. target is crouched)
	if ai.State == AIAttack && nearest != nil && nearest.Crouching {
		if ai.Unit.AlienType != nil && ai.Unit.AlienType.Aggression > 7 {
			ai.State = AIFlank
			ai.TurnsSince = 0
		}
	}

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
				_, _, _ = ai.Unit.FireAt(nearest, m)
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

	case AIFlank:
		if nearest != nil {
			dx := nearest.X - ai.Unit.X
			dy := nearest.Y - ai.Unit.Y
			nx := ai.Unit.X + signum(-dy)
			ny := ai.Unit.Y + signum(dx)
			ai.Unit.MoveTo(nx, ny, m)
		}
		ai.TurnsSince++
		if ai.TurnsSince > 2 {
			ai.State = AIAttack
		}

	case AIRetreat:
		if nearest != nil {
			dx := ai.Unit.X - nearest.X
			dy := ai.Unit.Y - nearest.Y
			fx := ai.Unit.X + signum(dx)*3
			fy := ai.Unit.Y + signum(dy)*3
			ai.Unit.MoveTo(fx, fy, m)
		}
		ai.TurnsSince++
		if ai.TurnsSince > 3 {
			ai.State = AIPatrol
		}
	}

	if ai.Unit.HP < ai.Unit.MaxHP/4 && ai.Unit.Alive {
		if ai.Unit.AlienType != nil && ai.Unit.AlienType.Bravery < 50 {
			ai.State = AIRetreat
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
		boundY := m.Height - 1
		if m.NumLevels > 1 {
			boundY = m.LevelHeight - 1
		}
		if ai.PatrolY >= boundY {
			ai.PatrolY = boundY - 1
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
		boundY := m.Height - 1
		if m.NumLevels > 1 {
			boundY = m.LevelHeight - 1
		}
		if ai.PatrolY >= boundY {
			ai.PatrolY = boundY - 1
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
		if !h.Alive || h.Level != ai.Unit.Level {
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

type CivilianAI struct {
	Unit   *Unit
	Scared bool
}

func NewCivilianAI(u *Unit) *CivilianAI {
	return &CivilianAI{Unit: u}
}

func (cai *CivilianAI) GenerateActions(units UnitList, m *BattleMap) []AlienAction {
	if !cai.Unit.Alive {
		return nil
	}

	var nearestThreat *Unit
	bestDist := 999.0
	for _, u := range units {
		if !u.Alive || u.Faction == 2 || u.Level != cai.Unit.Level {
			continue
		}
		dx := float64(u.X - cai.Unit.X)
		dy := float64(u.Y - cai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < bestDist {
			bestDist = dist
			nearestThreat = u
		}
	}

	if nearestThreat != nil && bestDist < 10 {
		cai.Scared = true
	}

	if !cai.Scared {
		return nil
	}

	if nearestThreat == nil {
		cai.Scared = false
		return nil
	}

	dx := float64(cai.Unit.X - nearestThreat.X)
	dy := float64(cai.Unit.Y - nearestThreat.Y)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		dist = 1
	}
	fx := cai.Unit.X + int(dx/dist*2)
	fy := cai.Unit.Y + int(dy/dist*2)

	if fx < 0 {
		fx = 0
	}
	if fy < 0 {
		fy = 0
	}
	if fx >= m.Width {
		fx = m.Width - 1
	}
	if fy >= m.LevelHeight {
		fy = m.LevelHeight - 1
	}

	if !m.Passable(fx, fy) {
		return nil
	}

	return []AlienAction{{
		Type: "move", Unit: cai.Unit,
		FromX: cai.Unit.X, FromY: cai.Unit.Y,
		ToX: fx, ToY: fy,
	}}
}
