package battle

import (
	"math"
	"math/rand"
	"sync"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
)

// Game balance constants for alien behavior.
const (
	VisualRangeThreshold = 12
	FlankDistThreshold   = 3
	FlankTUThreshold     = 20
	PsiSkillThreshold    = 40
	PsiTUThreshold       = 20
	GrenadeTUThreshold   = 18
	GrenadeRangeMax      = 8
	GrenadeRangeMin      = 1
	FlankManeuverTU      = 16
	FlankLongRangeTU     = 18
	DisperseTU           = 14
	TargetWoundedBonus   = 5
	TargetCrouchPenalty  = 3
	TargetLowTUBonus     = 3
	TargetHeavyPenalty   = 5
	MorphRangeBonus      = 3
	MorphThermalBonus    = 4
	MorphChemBonus       = 5
	ThermalSenseRange    = 10
	EcholocSenseRange    = 6
	LongRangeDist        = 8
	LongRangeAggroDist   = 4
	CoverSeekDist        = 3
	MeleeDist            = 1
	CloseAttackDist      = 2
	FireModeAutoDist     = 4
	FireModeBurstDist    = 8
	FlankMinDist         = 3
	RetreatMinDist       = 4
	CivilianFleeRange    = 10
	CivilianFleeStep     = 3
	PatrolScanRadius     = 12
	PatrolScanOffset     = 6
	RetreatStep          = 4
)
type AIState int

const (
	AIIdle     AIState = iota // Unit is stationary and waiting
	AIPatrol                  // Unit moves between patrol points
	AISearch                  // Unit moves toward the last known player position
	AIAttack                  // Unit actively engages the player
	AIFlank                   // Unit attempts to move to the side of a target
	AIRetreat                 // Unit falls back to a defensive position
	AISuppress                // Unit fires to keep the player in cover
)

// SquadRole defines the tactical function of a unit within a larger squad.
type SquadRole int

const (
	RoleNormal     SquadRole = iota // General combatant
	RoleFlanker                     // Specialized in flanking maneuvers
	RoleSuppressor                  // Specialized in suppressive fire
)

// SquadPlan coordinates multiple aliens to act as a cohesive unit.
type SquadPlan struct {
	PrimaryTarget   *Unit
	SecondaryTarget *Unit
	Roles           map[*Unit]SquadRole
	Retreat         bool
}

// Sighting records the last confirmed or inferred position of an enemy unit.
type Sighting struct {
	X, Y  int
	Turn  int
	Alive bool
}

// SquadMemory is a shared belief map: every alien writes sightings it makes and
// reads the most recent enemy positions reported by the squad, so a unit that
// has lost direct LOS can still converge on where its allies last saw a target.
type SquadMemory struct {
	mu        sync.Mutex
	sightings map[*Unit]Sighting
	turn      int
}

func NewSquadMemory() *SquadMemory {
	return &SquadMemory{sightings: make(map[*Unit]Sighting)}
}

// Report records a confirmed sighting of an enemy at (x, y) this turn.
func (sm *SquadMemory) Report(target *Unit, x, y, turn int) {
	if sm == nil || target == nil {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.sightings[target]; !ok || turn >= s.Turn {
		sm.sightings[target] = Sighting{X: x, Y: y, Turn: turn, Alive: true}
	}
}

// Turn returns the current turn counter under lock.
func (sm *SquadMemory) Turn() int {
	if sm == nil {
		return 0
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.turn
}

// Forget marks a target as no longer confirmed alive (e.g. killed).
func (sm *SquadMemory) Forget(target *Unit) {
	if sm == nil || target == nil {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sightings, target)
}

// Latest returns the most recent sighting across all known enemies, preferring
// the freshest report. Returns false if the squad has no belief state.
func (sm *SquadMemory) Latest() (Sighting, bool) {
	if sm == nil {
		return Sighting{}, false
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	best := Sighting{Turn: -1}
	found := false
	for _, s := range sm.sightings {
		if s.Turn > best.Turn {
			best = s
			found = true
		}
	}
	return best, found
}

// AlienAI manages the decision-making process for a single alien unit.
type AlienAI struct {
	Unit       *Unit
	State      AIState
	PatrolX    int
	PatrolY    int
	LastSeenX  int
	LastSeenY  int
	TurnsSince int
	InCover    bool
	Memory     *SquadMemory // shared squad belief map (nil if not used)
	rng        *rand.Rand   // per-AI seeded RNG for reproducible behaviour
}

func NewAlienAI(u *Unit) *AlienAI {
	var seed int64
	seed = int64(u.X)*73856093 + int64(u.Y)*19349663
	if u.AlienType != nil {
		seed += int64(u.AlienType.Rank) * 131
	}
	if u.Soldier != nil {
		for _, r := range u.Soldier.Name {
			seed = seed*31 + int64(r)
		}
	}
	seed ^= 0x9e3779b9
	return &AlienAI{
		Unit:  u,
		State: AIPatrol,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Update evaluates the world state and returns a sequence of actions for the alien to take.
func (ai *AlienAI) Update(units UnitList, m *BattleMap, humanUnits UnitList, plan *SquadPlan, tactics *engine.PlayerTactics) []AlienAction {
	if !ai.Unit.Alive || ai.Unit.HP <= 0 {
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

	// Spatial memory: report any human we can currently sense so the squad can
	// share target positions even after an individual alien loses line of sight.
	if ai.Memory != nil && nearest != nil {
		ai.Memory.Report(nearest, nearest.X, nearest.Y, ai.Memory.Turn())
	}

	ai.InCover = ai.evaluateCover(ai.Unit.X, ai.Unit.Y, m) > 0

	// Adaptive AI: Adjust behavior based on global player tactics observed across the campaign.
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

	// Heuristics to determine tactical leanings of the player.
	grenadeHeavy := avgGrenades >= 1.5
	longRange := avgRange >= 8.0
	playerLosing := avgLosses >= 1.0                      // aliens dominating -> more aggressive
	alienLosing := avgKills >= 2.0 && avgLosses < 0.5    // aliens dying fast -> more cautious

	switch ai.State {
	case AIPatrol:
		if nearest != nil && dist < VisualRangeThreshold {
			ai.State = AIAttack
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
			ai.TurnsSince = 0
		} else {
			actions = append(actions, ai.handlePatrol(m)...)
		}


	case AIAttack:
		// Attack: The core combat state. The alien targets a player, fires, or maneuvers.
		if nearest == nil {
			// Transition to Search if the target is lost.
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
		fired := false

		if target != nil && ai.canFireAt(target) {
			tdx := float64(target.X - ai.Unit.X)
			tdy := float64(target.Y - ai.Unit.Y)
			targetDist := math.Sqrt(tdx*tdx + tdy*tdy)
			ai.selectFireMode(int(targetDist))
			// 1. Suppress: If specialized role and in cover, maintain suppressive fire.
			if role == RoleSuppressor && ai.InCover {
				actions = append(actions, AlienAction{
					Type: "fire", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
				ai.State = AISuppress
				fired = true
			} else if role == RoleFlanker && targetDist > FlankDistThreshold && ai.Unit.TU >= FlankTUThreshold {
				// 2. Flank: Move to a side position if specialized and distance allows.
				ai.State = AIFlank
			} else if ai.Unit.AlienType != nil && ai.Unit.AlienType.Psi > PsiSkillThreshold && ai.Unit.TU >= PsiTUThreshold && ai.rng.Intn(3) == 0 {
				// 3. Psi Attack: Use psionic abilities if strong enough and by chance.
				actions = append(actions, AlienAction{
					Type: "psi", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
				fired = true
			} else if ai.Unit.Weapon == "alien_grenade" && ai.Unit.TU >= GrenadeTUThreshold && targetDist <= GrenadeRangeMax && targetDist > GrenadeRangeMin {
				// 4. Grenade: Throw alien grenade at the target's position (AoE).
				actions = append(actions, AlienAction{
					Type: "grenade", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
				fired = true
			} else {
				// 5. Standard Attack: Fire at target (melee if adjacent).
				if targetDist <= MeleeDist {
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
				fired = true
			}
		}

		// Maneuvering: Move to cover or advance based on player tactics.
		if role == RoleFlanker && dist > FlankDistThreshold && ai.Unit.TU >= FlankTUThreshold {
			fx, fy := ai.findFlankPosition(target, nearest, m, humanUnits)
			if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
				actions = ai.appendMove(actions, fx, fy)
			}
		} else if !ai.InCover && dist > FlankDistThreshold && (ai.Unit.TU >= FlankManeuverTU || (longRange && dist > LongRangeAggroDist && ai.Unit.TU >= FlankLongRangeTU)) {
			if longRange && dist > LongRangeAggroDist {
				// Adapt to long-range players: close distance aggressively.
				ax, ay := ai.advanceToward(nearest.X, nearest.Y, m, units)
				if (ax != ai.Unit.X || ay != ai.Unit.Y) && m.Passable(ax, ay) {
					actions = ai.appendMove(actions, ax, ay)
				}
			} else {
				// Standard maneuver: seek cover while facing the target.
				cx, cy := ai.findCoverTowardTarget(nearest.X, nearest.Y, m, humanUnits)
				if (cx != ai.Unit.X || cy != ai.Unit.Y) && m.Passable(cx, cy) {
					actions = ai.appendMove(actions, cx, cy)
				}
			}
		} else if grenadeHeavy && ai.Unit.TU >= DisperseTU {
			// Adapt to grenade-heavy players: disperse from allies to minimize blast damage.
			if buddy := ai.nearestAlly(units); buddy != nil {
				fx, fy := ai.disperseFrom(buddy, m, humanUnits)
				if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
					actions = ai.appendMove(actions, fx, fy)
				}
			}
		} else if !fired && ai.Unit.TU >= 20 && target != nil && ai.canFireAt(target) {
			// Last-resort fire if no combat action was taken yet and enough TU remains.
			ai.selectFireMode(int(dist))
			actions = append(actions, AlienAction{
				Type: "fire", Unit: ai.Unit, Target: target,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: target.X, ToY: target.Y,
			})
			fired = true
		}

	case AISuppress:
		actions = ai.handleSuppress(nearest, humanUnits, plan, m, dist)


	case AISearch:
		actions = ai.handleSearch(nearest, m, humanUnits)

	case AIFlank:
		actions = ai.handleFlank(nearest, m, humanUnits)


	case AIRetreat:
		actions = ai.handleRetreat(nearest, m, units)

	}

	// Dynamic retreat logic: determine if the alien should flee based on HP and bravery.
	retreatHP := ai.Unit.MaxHP / 4
	braveryThreshold := 50
	if alienLosing {
		// Aliens dying fast: retreat sooner and more readily.
		retreatHP = ai.Unit.MaxHP / 3
		braveryThreshold = 70
	}
	if playerLosing {
		// Aliens dominating: fight on, retreat only the most timid units.
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
		if ai.canSense(plan.PrimaryTarget.X, plan.PrimaryTarget.Y, m) {
			return plan.PrimaryTarget
		}
		if plan.SecondaryTarget != nil && plan.SecondaryTarget.Alive {
			if ai.canSense(plan.SecondaryTarget.X, plan.SecondaryTarget.Y, m) {
				return plan.SecondaryTarget
			}
		}
	}

	best := nearest
	if nearest != nil && nearest.Level != ai.Unit.Level {
		best = nil
	}
	bestScore := -999.0
	for _, h := range humanUnits {
		if !h.Alive || h.Level != ai.Unit.Level || !ai.canSense(h.X, h.Y, m) {
			continue
		}
		dx := float64(h.X - ai.Unit.X)
		dy := float64(h.Y - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		score := -dist
		if h.HP < h.MaxHP/2 {
			score += TargetWoundedBonus
		}
		if h.Crouching {
			score -= TargetCrouchPenalty
		}
		if h.Weapon == "rocket" || h.Weapon == "heavy_plasma" {
			score -= TargetHeavyPenalty
		}
		if h.TU < FlankTUThreshold {
			score += TargetLowTUBonus
		}

		if at := ai.Unit.AlienType; at != nil && at.Morphology != nil {
			morph := at.Morphology
			if morph.ThermalSense == data.SenseHigh && h.Crouching {
				score += MorphThermalBonus
			}
			if morph.ChemicalSense == data.SenseHigh && h.HP < h.MaxHP/2 {
				score += MorphChemBonus
			}
			if (morph.Eyesight == data.SenseExcellent || morph.Eyesight == data.SenseMultiSpec) && dist > LongRangeDist {
				score += MorphRangeBonus
			}
		}

		if score > bestScore {
			bestScore = score
			best = h
		}
	}
	return best
}

func (ai *AlienAI) canSense(tx, ty int, m *BattleMap) bool {
	if ai.Unit.CanSee(tx, ty, m) {
		return true
	}
	if ai.Unit.AlienType == nil || ai.Unit.AlienType.Morphology == nil {
		return false
	}
	morph := ai.Unit.AlienType.Morphology
	if morph.ThermalSense == data.SenseHigh {
		dx := float64(tx - ai.Unit.X)
		dy := float64(ty - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= ThermalSenseRange {
			return true
		}
	}
	if morph.Hearing == data.SenseEcholoc {
		dx := float64(tx - ai.Unit.X)
		dy := float64(ty - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= EcholocSenseRange {
			return true
		}
	}
	return false
}

func (ai *AlienAI) canFireAt(target *Unit) bool {
	if target == nil || !target.Alive {
		return false
	}
	w, ok := data.RuleItems[ai.Unit.Weapon]
	if !ok {
		return false
	}
	if ai.Unit.TU < w.TU {
		return false
	}
	if w.AmmoMax < 99 && ai.Unit.WeaponAmmo <= 0 {
		return false
	}
	dx := float64(ai.Unit.X - target.X)
	dy := float64(ai.Unit.Y - target.Y)
	dist := math.Sqrt(dx*dx + dy*dy)
	if w.Range > 0 && dist > float64(w.Range) {
		return false
	}
	return true
}

func humanFrom(units UnitList, level int) UnitList {
	var humans UnitList
	for _, u := range units {
		if u.Alive && u.Faction == FactionHuman && u.Level == level {
			humans = append(humans, u)
		}
	}
	return humans
}

func (ai *AlienAI) evaluateCover(x, y int, m *BattleMap) int {
	t := m.At(x, y)
	return t.Cover
}

// evaluateCoverVsThreats scores how well the tile (x, y) is protected against
// incoming fire from visible human units. Unlike evaluateCover (which only
// reports the tile's own cover value), this considers the geometry of cover
// along the line of fire from each threat, rewarding positions that actually
// block LOS to the units most likely to shoot back.
func (ai *AlienAI) evaluateCoverVsThreats(x, y int, m *BattleMap, humanUnits UnitList) float64 {
	var totalProtection float64
	threatCount := 0
	for _, h := range humanUnits {
		if !h.Alive {
			continue
		}
		// A threat only matters if it can currently see the candidate tile.
		if !h.CanSee(x, y, m) {
			// Full cover: this position is completely hidden from this threat.
			totalProtection += 100
			threatCount++
			continue
		}
		// Cover value of the highest obstacle between the threat and the tile.
		lineCover := m.CoverAlongLine(h.X, h.Y, x, y)
		// Closer threats are more dangerous; weight their protection higher.
		dx := float64(h.X - x)
		dy := float64(h.Y - y)
		dist := math.Sqrt(dx*dx + dy*dy)
		weight := 1.0
		if dist < CoverSeekDist {
			weight = 1.5
		}
		totalProtection += float64(lineCover) * weight
		threatCount++
	}
	if threatCount == 0 {
		return 0
	}
	return totalProtection / float64(threatCount)
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
		// Directional cover: reward positions that actually block LOS to threats.
		score += ai.evaluateCoverVsThreats(nx, ny, m, humanFrom(units, ai.Unit.Level)) * 0.5
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
		if tDist < FlankMinDist {
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

func (ai *AlienAI) retreatTarget(threat *Unit, m *BattleMap, units UnitList) (int, int) {
	dx := ai.Unit.X - threat.X
	dy := ai.Unit.Y - threat.Y

	mag := math.Sqrt(float64(dx*dx + dy*dy))
	if mag < 1 {
		mag = 1
	}
	fx := ai.Unit.X + int(float64(dx)/mag*RetreatStep)
	fy := ai.Unit.Y + int(float64(dy)/mag*RetreatStep)

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
	bestProtection := 0.0

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
			protection := ai.evaluateCoverVsThreats(cx, cy, m, humanFrom(units, ai.Unit.Level))
			total := math.Max(protection, float64(cover))
			tdx := threat.X - cx
			tdy := threat.Y - cy
			threatDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))
			if total > bestProtection && threatDist > RetreatMinDist {
				bestProtection = total
				bestX = cx
				bestY = cy
			}
		}
	}

	return bestX, bestY
}

func (ai *AlienAI) advanceToward(tx, ty int, m *BattleMap, units UnitList) (int, int) {
	return ai.GetNextPathStep(tx, ty, m, units)
}

func (ai *AlienAI) nearestAlly(units UnitList) *Unit {
	var best *Unit
	bestDist := 999.0
	for _, u := range units {
		if u == ai.Unit || !u.Alive || u.Faction != FactionAlien {
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

		score := -tDist + float64(cover)*8 - ai.reactionFirePenalty(nx, ny, m, units) + ai.evaluateCoverVsThreats(nx, ny, m, humanFrom(units, ai.Unit.Level))*0.5

		if score > bestScore {
			bestScore = score
			bestX = nx
			bestY = ny
		}
	}

	return ai.GetNextPathStep(bestX, bestY, m, units)
}

func (ai *AlienAI) patrolTarget(m *BattleMap) (int, int) {
	if ai.PatrolX == 0 && ai.PatrolY == 0 {
		for attempt := 0; attempt < 10; attempt++ {
			px := ai.Unit.X + ai.rng.Intn(PatrolScanRadius) - PatrolScanOffset
			py := ai.Unit.Y + ai.rng.Intn(PatrolScanRadius) - PatrolScanOffset
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
		if !ai.canSense(h.X, h.Y, m) {
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
		if !u.Alive || u.Faction == FactionCivilian || u.Level != cai.Unit.Level {
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

	if nearestThreat != nil && bestDist < CivilianFleeRange {
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
	fx := cai.Unit.X + int(dx/dist*CivilianFleeStep)
	fy := cai.Unit.Y + int(dy/dist*CivilianFleeStep)

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

	if !m.Passable(fx, fy) || units.At(fx, fy) != nil {
		return nil
	}
	return []AlienAction{{
		Type: "move", Unit: cai.Unit,
		FromX: cai.Unit.X, FromY: cai.Unit.Y,
		ToX: fx, ToY: fy,
	}}
}

func (ai *AlienAI) appendMove(actions []AlienAction, toX, toY int) []AlienAction {
	return append(actions, AlienAction{
		Type:  "move",
		Unit:  ai.Unit,
		FromX: ai.Unit.X,
		FromY: ai.Unit.Y,
		ToX:   toX,
		ToY:   toY,
	})
}

func (ai *AlienAI) handlePatrol(m *BattleMap) []AlienAction {
	var actions []AlienAction
	px, py := ai.patrolTarget(m)
	if m.Passable(px, py) {
		actions = ai.appendMove(actions, px, py)
	}
	return actions
}

func (ai *AlienAI) handleSearch(nearest *Unit, m *BattleMap, humanUnits UnitList) []AlienAction {
	var actions []AlienAction
	if nearest != nil {
		ai.LastSeenX = nearest.X
		ai.LastSeenY = nearest.Y
		ai.State = AIAttack
		ai.TurnsSince = 0
	} else {
		// Prefer the freshest squad-shared sighting over our own stale memory.
		sx, sy := ai.LastSeenX, ai.LastSeenY
		if ai.Memory != nil {
			if s, ok := ai.Memory.Latest(); ok {
				if s.Turn >= ai.Memory.Turn()-3 {
					sx, sy = s.X, s.Y
				}
			}
		}
		// If no sighting is available (both zero), fall back to patrol.
		if sx == 0 && sy == 0 {
			ai.State = AIPatrol
			return ai.handlePatrol(m)
		}
		nx, ny := ai.moveTowardTargetCover(sx, sy, m, humanUnits)
		if (nx != ai.Unit.X || ny != ai.Unit.Y) && m.Passable(nx, ny) {
			actions = ai.appendMove(actions, nx, ny)
		}
		ai.TurnsSince++
		if ai.TurnsSince > 6 {
			// Transition back to Patrol if no one is found after several turns.
			ai.State = AIPatrol
		}
	}
	return actions
}

func (ai *AlienAI) handleRetreat(nearest *Unit, m *BattleMap, units UnitList) []AlienAction {
	var actions []AlienAction
	if nearest != nil {
		fx, fy := ai.retreatTarget(nearest, m, units)
		if m.Passable(fx, fy) {
			actions = ai.appendMove(actions, fx, fy)
		}
	}
	ai.TurnsSince++
	if ai.TurnsSince > 3 {
		// Stop retreating and return to patrol after some distance is gained.
		ai.State = AIPatrol
	}
	return actions
}

func (ai *AlienAI) handleSuppress(nearest *Unit, humanUnits UnitList, plan *SquadPlan, m *BattleMap, dist float64) []AlienAction {
	var actions []AlienAction
	// Suppress: Fire at the target while staying in cover to keep the player pinned down.
	if nearest == nil {
		ai.State = AIAttack
		return actions
	}
	target := ai.selectTarget(nearest, humanUnits, plan, m)
	if target != nil && ai.canFireAt(target) {
		ai.selectFireMode(int(dist))
		actions = append(actions, AlienAction{
			Type: "fire", Unit: ai.Unit, Target: target,
			FromX: ai.Unit.X, FromY: ai.Unit.Y,
			ToX: target.X, ToY: target.Y,
		})
	}
	ai.TurnsSince++
	if ai.TurnsSince > 3 || !ai.InCover {
		// Transition back to Attack if target is gone, duration elapsed, or cover is lost.
		ai.State = AIAttack
		ai.TurnsSince = 0
	}
	return actions
}

func (ai *AlienAI) handleFlank(nearest *Unit, m *BattleMap, humanUnits UnitList) []AlienAction {
	var actions []AlienAction
	// Flank: Move toward a position that provides a side-on angle to the target.
	if nearest == nil {
		ai.State = AIAttack
		return actions
	}
	fx, fy := ai.findFlankPosition(nearest, nearest, m, humanUnits)
	if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
		actions = ai.appendMove(actions, fx, fy)
	}
	// After moving, fire if the target is still reachable.
	if ai.canFireAt(nearest) {
		dx := float64(ai.Unit.X - nearest.X)
		dy := float64(ai.Unit.Y - nearest.Y)
		fdist := math.Sqrt(dx*dx + dy*dy)
		ai.selectFireMode(int(fdist))
		actions = append(actions, AlienAction{
			Type: "fire", Unit: ai.Unit, Target: nearest,
			FromX: ai.Unit.X, FromY: ai.Unit.Y,
			ToX: nearest.X, ToY: nearest.Y,
		})
	}
	ai.TurnsSince++
	if ai.TurnsSince > 2 {
		// Return to Attack state after a short flanking maneuver.
		ai.State = AIAttack
		ai.TurnsSince = 0
	}
	return actions
}

func (ai *AlienAI) selectFireMode(dist int) {
	w, ok := data.RuleItems[ai.Unit.Weapon]
	if !ok {
		return
	}
	modes := w.Modes()
	if len(modes) <= 1 {
		return
	}
	// Preference order: Auto > Burst > Aimed, fall back if TU is insufficient.
	tryMode := func(mode data.FireMode) bool {
		if !w.HasMode(mode) {
			return false
		}
		if ai.Unit.TU >= w.ModeTU(mode) {
			ai.Unit.FireMode = mode
			return true
		}
		return false
	}
	if dist <= FireModeAutoDist {
		if tryMode(data.FireModeAuto) || tryMode(data.FireModeBurst) || tryMode(data.FireModeAimed) {
			return
		}
	} else if dist <= FireModeBurstDist {
		if tryMode(data.FireModeBurst) || tryMode(data.FireModeAimed) {
			return
		}
	}
	tryMode(data.FireModeAimed)
}
