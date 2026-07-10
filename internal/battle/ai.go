package battle

import (
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
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
	AISuppress
)

type SquadRole int

const (
	RoleNormal SquadRole = iota
	RoleFlanker
	RoleSuppressor
)

type SquadPlan struct {
	PrimaryTarget   *Unit
	SecondaryTarget *Unit
	Roles           map[*Unit]SquadRole
	Retreat         bool
}

type AlienAI struct {
	Unit       *Unit
	State      AIState
	PatrolX    int
	PatrolY    int
	LastSeenX  int
	LastSeenY  int
	TurnsSince int
	InCover    bool
}

func NewAlienAI(u *Unit) *AlienAI {
	return &AlienAI{
		Unit:  u,
		State: AIPatrol,
	}
}

func (ai *AlienAI) Update(units UnitList, m *BattleMap, humanUnits UnitList, plan *SquadPlan, tactics *engine.PlayerTactics) []AlienAction {
	if !ai.Unit.Alive {
		return nil
	}

	var actions []AlienAction
	nearest, dist := ai.findNearest(humanUnits, m)
	role := RoleNormal
	if plan != nil {
		if r, ok := plan.Roles[ai.Unit]; ok {
			role = r
		}
	}

	if ai.Unit.HP <= 0 {
		return nil
	}

	ai.InCover = ai.evaluateCover(ai.Unit.X, ai.Unit.Y, m) > 0

	// Adapt behavior to player tactics observed across missions
	var tac engine.PlayerTactics
	if tactics != nil {
		tac = *tactics
	}
	tbc := tac.BattleCount
	if tbc < 1 {
		tbc = 1
	}
	avgGrenades := float64(tac.GrenadeUsage) / float64(tbc)
	avgRange := tac.AverageRange
	avgKills := float64(tac.TotalAlienKills) / float64(tbc)
	avgLosses := float64(tac.TotalSoldierLosses) / float64(tbc)

	grenadeHeavy := avgGrenades >= 1.5
	longRange := avgRange >= 8.0
	playerLosing := avgLosses >= 1.0                      // aliens dominating -> more aggressive
	alienLosing := avgKills >= 2.0 && avgLosses < 0.5    // aliens dying fast -> more cautious

	switch ai.State {
	case AIPatrol:
		if nearest != nil && dist < 12 {
			ai.State = AIAttack
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0
		} else {
			px, py := ai.patrolTarget(m)
			if m.Passable(px, py) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: px, ToY: py,
				})
			}
		}

	case AIAttack:
		if nearest == nil {
			ai.TurnsSince++
			if ai.TurnsSince > 2 {
				ai.State = AISearch
			}
			return nil
		}

		ai.LastSeenX = nearest.X
		ai.LastSeenY = nearest.Y
		ai.TurnsSince = 0

		target := ai.selectTarget(nearest, humanUnits, plan, m)

		if target != nil && ai.canFireAt(target) {
			if role == RoleSuppressor && ai.InCover {
				actions = append(actions, AlienAction{
					Type: "fire", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
				ai.State = AISuppress
			} else if role == RoleFlanker && dist > 3 && ai.Unit.TU >= 20 {
				ai.State = AIFlank
			} else if (dist <= 2 || (longRange && dist <= 3) || ai.Unit.AlienType.Aggression > 7) {
				if dist <= 1 {
					actions = append(actions, AlienAction{
						Type: "melee", Unit: ai.Unit, Target: target,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: target.X, ToY: target.Y,
					})
				} else {
					actions = append(actions, AlienAction{
						Type: "fire", Unit: ai.Unit, Target: target,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: target.X, ToY: target.Y,
					})
				}
			}
		}

		if role == RoleFlanker && dist > 3 && ai.Unit.TU >= 20 {
			fx, fy := ai.findFlankPosition(target, nearest, m, humanUnits)
			if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: fx, ToY: fy,
				})
			}
		} else if !ai.InCover && dist > 3 && ai.Unit.TU >= 16 || (longRange && dist > 4 && ai.Unit.TU >= 18) {
			if longRange && dist > 4 {
				// Player snipes from afar: close the distance aggressively
				ax, ay := ai.advanceToward(nearest.X, nearest.Y, m, units)
				if (ax != ai.Unit.X || ay != ai.Unit.Y) && m.Passable(ax, ay) {
					actions = append(actions, AlienAction{
						Type: "move", Unit: ai.Unit,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: ax, ToY: ay,
					})
				}
			} else {
				cx, cy := ai.findCoverTowardTarget(nearest.X, nearest.Y, m, humanUnits)
				if (cx != ai.Unit.X || cy != ai.Unit.Y) && m.Passable(cx, cy) {
					actions = append(actions, AlienAction{
						Type: "move", Unit: ai.Unit,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: cx, ToY: cy,
					})
				}
			}
		} else if grenadeHeavy && ai.Unit.TU >= 14 {
			// Player lobs grenades at clusters: spread out to minimize casualties
			if buddy := ai.nearestAlly(units); buddy != nil {
				fx, fy := ai.disperseFrom(buddy, m, humanUnits)
				if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
					actions = append(actions, AlienAction{
						Type: "move", Unit: ai.Unit,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: fx, ToY: fy,
					})
				}
			}
		} else if ai.Unit.TU >= 20 && ai.Unit.TU < ai.Unit.MaxTU && target != nil && ai.canFireAt(target) {
			actions = append(actions, AlienAction{
				Type: "fire", Unit: ai.Unit, Target: target,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: target.X, ToY: target.Y,
			})
		}

	case AISuppress:
		if nearest == nil {
			ai.State = AIAttack
			break
		}
		target := ai.selectTarget(nearest, humanUnits, plan, m)
		if target != nil && ai.canFireAt(target) {
			actions = append(actions, AlienAction{
				Type: "fire", Unit: ai.Unit, Target: target,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: target.X, ToY: target.Y,
			})
		}
		ai.TurnsSince++
		if ai.TurnsSince > 3 || !ai.InCover {
			ai.State = AIAttack
			ai.TurnsSince = 0
		}

	case AISearch:
		if nearest != nil {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.State = AIAttack
			ai.TurnsSince = 0
		} else {
			nx, ny := ai.moveTowardTargetCover(ai.LastSeenX, ai.LastSeenY, m, humanUnits)
			if (nx != ai.Unit.X || ny != ai.Unit.Y) && m.Passable(nx, ny) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: nx, ToY: ny,
				})
			}
			ai.TurnsSince++
			if ai.TurnsSince > 6 {
				ai.State = AIPatrol
			}
		}

	case AIFlank:
		if nearest == nil {
			ai.State = AIAttack
			break
		}
		fx, fy := ai.findFlankPosition(nearest, nearest, m, humanUnits)
		if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
			actions = append(actions, AlienAction{
				Type: "move", Unit: ai.Unit,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: fx, ToY: fy,
			})
		}
		ai.TurnsSince++
		if ai.TurnsSince > 2 {
			ai.State = AIAttack
			ai.TurnsSince = 0
		}

	case AIRetreat:
		if nearest != nil {
			fx, fy := ai.retreatTarget(nearest, m)
			if m.Passable(fx, fy) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: fx, ToY: fy,
				})
			}
		}
		ai.TurnsSince++
		if ai.TurnsSince > 3 {
			ai.State = AIPatrol
		}
	}

	retreatHP := ai.Unit.MaxHP / 4
	braveryThreshold := 50
	if alienLosing {
		// Aliens dying fast: retreat sooner and more readily
		retreatHP = ai.Unit.MaxHP / 3
		braveryThreshold = 70
	}
	if playerLosing {
		// Aliens dominating: fight on, retreat only the timid
		braveryThreshold = 30
	}

	if ai.Unit.HP < retreatHP && ai.Unit.Alive && ai.Unit.AlienType != nil {
		if ai.Unit.AlienType.Bravery < braveryThreshold {
			ai.State = AIRetreat
			ai.TurnsSince = 0
			return actions
		}
	}

	if plan != nil && plan.Retreat {
		ai.State = AIRetreat
		ai.TurnsSince = 0
		return actions
	}

	return actions
}

func (ai *AlienAI) selectTarget(nearest *Unit, humanUnits UnitList, plan *SquadPlan, m *BattleMap) *Unit {
	if plan != nil && plan.PrimaryTarget != nil && plan.PrimaryTarget.Alive {
		if ai.Unit.CanSee(plan.PrimaryTarget.X, plan.PrimaryTarget.Y, m) {
			return plan.PrimaryTarget
		}
		if plan.SecondaryTarget != nil && plan.SecondaryTarget.Alive {
			if ai.Unit.CanSee(plan.SecondaryTarget.X, plan.SecondaryTarget.Y, m) {
				return plan.SecondaryTarget
			}
		}
	}

	best := nearest
	bestScore := -999.0
	for _, h := range humanUnits {
		if !h.Alive || !ai.Unit.CanSee(h.X, h.Y, m) {
			continue
		}
		dx := float64(h.X - ai.Unit.X)
		dy := float64(h.Y - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		score := -dist
		if h.HP < h.MaxHP/2 {
			score += 5
		}
		if h.Crouching {
			score -= 3
		}
		if h.Weapon == "rocket" || h.Weapon == "heavy_plasma" {
			score -= 5
		}
		if h.TU < 20 {
			score += 3
		}
		if score > bestScore {
			bestScore = score
			best = h
		}
	}
	return best
}

func (ai *AlienAI) canFireAt(target *Unit) bool {
	if target == nil || !target.Alive {
		return false
	}
	if ai.Unit.TU < 15 {
		return false
	}
	w, ok := data.RuleItems[ai.Unit.Weapon]
	if !ok {
		return false
	}
	if w.AmmoMax < 99 && ai.Unit.WeaponAmmo <= 0 {
		return false
	}
	return true
}

func (ai *AlienAI) evaluateCover(x, y int, m *BattleMap) int {
	t := m.At(x, y)
	return t.Cover
}

func (ai *AlienAI) findCoverTowardTarget(tx, ty int, m *BattleMap, units UnitList) (int, int) {
	bestX, bestY := ai.Unit.X, ai.Unit.Y
	bestScore := -999.0

	dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	dx := tx - ai.Unit.X
	dy := ty - ai.Unit.Y
	dirX := 0
	if dx > 0 {
		dirX = 1
	} else if dx < 0 {
		dirX = -1
	}
	dirY := 0
	if dy > 0 {
		dirY = 1
	} else if dy < 0 {
		dirY = -1
	}

	for _, d := range dirs {
		nx := ai.Unit.X + d[0]
		ny := ai.Unit.Y + d[1]
		if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
			continue
		}
		if !m.Passable(nx, ny) {
			continue
		}
		unitAt := units.At(nx, ny)
		if unitAt != nil && unitAt != ai.Unit {
			continue
		}

		cover := ai.evaluateCover(nx, ny, m)
		tdx := tx - nx
		tdy := ty - ny
		tDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))
		moveDist := math.Abs(float64(d[0])) + math.Abs(float64(d[1]))

		score := float64(cover)*3.0 - tDist*2.0 - moveDist*2.0
		if cover > 0 {
			score += 10
		}
		if (d[0] == dirX || d[1] == dirY) && cover > 0 {
			score += 5
		}

		if score > bestScore {
			bestScore = score
			bestX = nx
			bestY = ny
		}
	}

	return bestX, bestY
}

func (ai *AlienAI) findFlankPosition(target, nearest *Unit, m *BattleMap, units UnitList) (int, int) {
	if target == nil {
		return ai.Unit.X, ai.Unit.Y
	}

	dx := target.X - ai.Unit.X
	dy := target.Y - ai.Unit.Y

	absDx := dx
	if absDx < 0 {
		absDx = -absDx
	}
	absDy := dy
	if absDy < 0 {
		absDy = -absDy
	}

	var flankDirs [][2]int
	if absDx > absDy {
		flankDirs = [][2]int{{0, 3}, {0, -3}, {0, 2}, {0, -2}}
	} else {
		flankDirs = [][2]int{{3, 0}, {-3, 0}, {2, 0}, {-2, 0}}
	}

	bestX, bestY := ai.Unit.X, ai.Unit.Y
	bestScore := -999.0

	for _, fd := range flankDirs {
		fx := ai.Unit.X + fd[0]
		fy := ai.Unit.Y + fd[1]
		if fx < 1 || fx >= m.Width-1 || fy < 1 || fy >= m.LevelHeight-1 {
			continue
		}
		if !m.Passable(fx, fy) {
			continue
		}
		unitAt := units.At(fx, fy)
		if unitAt != nil && unitAt != ai.Unit {
			continue
		}

		tdx := target.X - fx
		tdy := target.Y - fy
		tDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))
		if tDist < 3 {
			continue
		}

		mdx := ai.Unit.X - fx
		mdy := ai.Unit.Y - fy
		mDist := math.Abs(float64(mdx)) + math.Abs(float64(mdy))

		cover := ai.evaluateCover(fx, fy, m)
		score := -mDist + float64(cover)*5 + tDist*0.5

		if !ai.Unit.CanSee(fx, fy, m) {
			score -= 8
		}

		if m.CoverAlongLine(ai.Unit.X, ai.Unit.Y, fx, fy) > 20 {
			score += 5
		}

		if score > bestScore {
			bestScore = score
			bestX = fx
			bestY = fy
		}
	}

	return bestX, bestY
}

func (ai *AlienAI) retreatTarget(threat *Unit, m *BattleMap) (int, int) {
	dx := ai.Unit.X - threat.X
	dy := ai.Unit.Y - threat.Y

	mag := math.Sqrt(float64(dx*dx + dy*dy))
	if mag < 1 {
		mag = 1
	}
	fx := ai.Unit.X + int(float64(dx)/mag*4)
	fy := ai.Unit.Y + int(float64(dy)/mag*4)

	if fx < 1 {
		fx = 1
	}
	if fy < 1 {
		fy = 1
	}
	if fx >= m.Width-1 {
		fx = m.Width - 2
	}
	if fy >= m.LevelHeight-1 {
		fy = m.LevelHeight - 2
	}

	bestX, bestY := fx, fy
	bestCover := 0

	for ox := -1; ox <= 1; ox++ {
		for oy := -1; oy <= 1; oy++ {
			cx := fx + ox
			cy := fy + oy
			if cx < 1 || cx >= m.Width-1 || cy < 1 || cy >= m.LevelHeight-1 {
				continue
			}
			if !m.Passable(cx, cy) {
				continue
			}
			cover := ai.evaluateCover(cx, cy, m)
			tdx := threat.X - cx
			tdy := threat.Y - cy
			threatDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))
			if cover > bestCover && threatDist > 4 {
				bestCover = cover
				bestX = cx
				bestY = cy
			}
		}
	}

	return bestX, bestY
}

func (ai *AlienAI) advanceToward(tx, ty int, m *BattleMap, units UnitList) (int, int) {
	dx := tx - ai.Unit.X
	dy := ty - ai.Unit.Y

	var cands [][2]int
	if dx > 0 {
		cands = append(cands, [2]int{ai.Unit.X + 1, ai.Unit.Y})
	}
	if dx < 0 {
		cands = append(cands, [2]int{ai.Unit.X - 1, ai.Unit.Y})
	}
	if dy > 0 {
		cands = append(cands, [2]int{ai.Unit.X, ai.Unit.Y + 1})
	}
	if dy < 0 {
		cands = append(cands, [2]int{ai.Unit.X, ai.Unit.Y - 1})
	}

	bestX, bestY := ai.Unit.X, ai.Unit.Y
	bestDist := math.Sqrt(float64(dx*dx + dy*dy))
	for _, c := range cands {
		nx, ny := c[0], c[1]
		if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
			continue
		}
		if !m.Passable(nx, ny) {
			continue
		}
		if u := units.At(nx, ny); u != nil && u != ai.Unit {
			continue
		}
		ndx := tx - nx
		ndy := ty - ny
		nd := math.Sqrt(float64(ndx*ndx + ndy*ndy))
		if nd < bestDist {
			bestDist = nd
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}

func (ai *AlienAI) nearestAlly(units UnitList) *Unit {
	var best *Unit
	bestDist := 999.0
	for _, u := range units {
		if u == ai.Unit || !u.Alive || u.Faction != 1 {
			continue
		}
		dx := float64(u.X - ai.Unit.X)
		dy := float64(u.Y - ai.Unit.Y)
		d := math.Sqrt(dx*dx + dy*dy)
		if d < bestDist {
			bestDist = d
			best = u
		}
	}
	return best
}

func (ai *AlienAI) disperseFrom(buddy *Unit, m *BattleMap, units UnitList) (int, int) {
	dx := ai.Unit.X - buddy.X
	dy := ai.Unit.Y - buddy.Y

	var cands [][2]int
	if dx > 0 {
		cands = append(cands, [2]int{ai.Unit.X + 1, ai.Unit.Y})
	}
	if dx < 0 {
		cands = append(cands, [2]int{ai.Unit.X - 1, ai.Unit.Y})
	}
	if dy > 0 {
		cands = append(cands, [2]int{ai.Unit.X, ai.Unit.Y + 1})
	}
	if dy < 0 {
		cands = append(cands, [2]int{ai.Unit.X, ai.Unit.Y - 1})
	}
	if dx > 0 && dy > 0 {
		cands = append(cands, [2]int{ai.Unit.X + 1, ai.Unit.Y + 1})
	}
	if dx > 0 && dy < 0 {
		cands = append(cands, [2]int{ai.Unit.X + 1, ai.Unit.Y - 1})
	}
	if dx < 0 && dy > 0 {
		cands = append(cands, [2]int{ai.Unit.X - 1, ai.Unit.Y + 1})
	}
	if dx < 0 && dy < 0 {
		cands = append(cands, [2]int{ai.Unit.X - 1, ai.Unit.Y - 1})
	}

	bestX, bestY := ai.Unit.X, ai.Unit.Y
	bestScore := -999.0
	for _, c := range cands {
		nx, ny := c[0], c[1]
		if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
			continue
		}
		if !m.Passable(nx, ny) {
			continue
		}
		if u := units.At(nx, ny); u != nil && u != ai.Unit {
			continue
		}
		bdx := nx - buddy.X
		bdy := ny - buddy.Y
		bDist := math.Sqrt(float64(bdx*bdx + bdy*bdy))
		score := bDist + float64(ai.evaluateCover(nx, ny, m))*4
		if score > bestScore {
			bestScore = score
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}

func (ai *AlienAI) moveTowardTargetCover(tx, ty int, m *BattleMap, units UnitList) (int, int) {
	dx := tx - ai.Unit.X
	dy := ty - ai.Unit.Y

	var candidates [][2]int
	if dx > 0 {
		candidates = append(candidates, [2]int{ai.Unit.X + 1, ai.Unit.Y})
	}
	if dx < 0 {
		candidates = append(candidates, [2]int{ai.Unit.X - 1, ai.Unit.Y})
	}
	if dy > 0 {
		candidates = append(candidates, [2]int{ai.Unit.X, ai.Unit.Y + 1})
	}
	if dy < 0 {
		candidates = append(candidates, [2]int{ai.Unit.X, ai.Unit.Y - 1})
	}

	bestX, bestY := ai.Unit.X, ai.Unit.Y
	bestScore := -999.0

	for _, c := range candidates {
		nx, ny := c[0], c[1]
		if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
			continue
		}
		if !m.Passable(nx, ny) {
			continue
		}
		unitAt := units.At(nx, ny)
		if unitAt != nil && unitAt != ai.Unit {
			continue
		}

		cover := ai.evaluateCover(nx, ny, m)
		tdx := tx - nx
		tdy := ty - ny
		tDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))

		score := -tDist + float64(cover)*8

		if score > bestScore {
			bestScore = score
			bestX = nx
			bestY = ny
		}
	}

	return bestX, bestY
}

func (ai *AlienAI) patrolTarget(m *BattleMap) (int, int) {
	if ai.PatrolX == 0 && ai.PatrolY == 0 {
		for attempt := 0; attempt < 10; attempt++ {
			px := ai.Unit.X + rand.Intn(12) - 6
			py := ai.Unit.Y + rand.Intn(12) - 6
			if px < 1 {
				px = 1
			}
			if py < 1 {
				py = 1
			}
			if px >= m.Width-1 {
				px = m.Width - 2
			}
			boundY := m.Height - 1
			if m.NumLevels > 1 {
				boundY = m.LevelHeight - 1
			}
			if py >= boundY {
				py = boundY - 1
			}
			if m.Passable(px, py) && ai.evaluateCover(px, py, m) > 0 {
				ai.PatrolX = px
				ai.PatrolY = py
				break
			}
			if attempt == 9 {
				ai.PatrolX = px
				ai.PatrolY = py
			}
		}
	}
	return ai.PatrolX, ai.PatrolY
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
	fx := cai.Unit.X + int(dx/dist*3)
	fy := cai.Unit.Y + int(dy/dist*3)

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
