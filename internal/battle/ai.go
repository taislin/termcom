package battle

import (
	"math"
	"math/rand"
	"sync"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
)

var globalRNG = rand.New(rand.NewSource(42))

// AIState defines the current behavioral mode of an alien unit.
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

	// Spatial memory: report any human we can currently sense so the squad can
	// share target positions even after an individual alien loses line of sight.
	if ai.Memory != nil && nearest != nil {
		ai.Memory.Report(nearest, nearest.X, nearest.Y, ai.Memory.turn)
	}

	if ai.Unit.HP <= 0 {
		return nil
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
		// Patrol: Unit moves between random patrol points until a target is detected.
		if nearest != nil && dist < 12 {
			// Transition to Attack if a player is spotted within visual range.
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

		if target != nil && ai.canFireAt(target) {
			ai.selectFireMode(int(dist))
			// 1. Suppress: If specialized role and in cover, maintain suppressive fire.
			if role == RoleSuppressor && ai.InCover {
				actions = append(actions, AlienAction{
					Type: "fire", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
				ai.State = AISuppress
			} else if role == RoleFlanker && dist > 3 && ai.Unit.TU >= 20 {
				// 2. Flank: Move to a side position if specialized and distance allows.
				ai.State = AIFlank
			} else if ai.Unit.AlienType != nil && ai.Unit.AlienType.Psi > 40 && ai.Unit.TU >= 20 && ai.rng.Intn(3) == 0 {
				// 3. Psi Attack: Use psionic abilities if strong enough and by chance.
				actions = append(actions, AlienAction{
					Type: "psi", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
			} else if ai.Unit.Weapon == "alien_grenade" && ai.Unit.TU >= 18 && dist <= 8 && dist > 1 {
				// 4. Grenade: Throw alien grenade at the target's position (AoE).
				actions = append(actions, AlienAction{
					Type: "grenade", Unit: ai.Unit, Target: target,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: target.X, ToY: target.Y,
				})
			} else if (dist <= 2 || (longRange && dist <= 3) || (ai.Unit.AlienType != nil && ai.Unit.AlienType.Aggression > 7)) {
				// 5. Standard Attack: Melee if adjacent, otherwise ranged fire.
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

		// Maneuvering: Move to cover or advance based on player tactics.
		if role == RoleFlanker && dist > 3 && ai.Unit.TU >= 20 {
			fx, fy := ai.findFlankPosition(target, nearest, m, humanUnits)
			if (fx != ai.Unit.X || fy != ai.Unit.Y) && m.Passable(fx, fy) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: fx, ToY: fy,
				})
			}
		} else if !ai.InCover && dist > 3 && (ai.Unit.TU >= 16 || (longRange && dist > 4 && ai.Unit.TU >= 18)) {
			if longRange && dist > 4 {
				// Adapt to long-range players: close distance aggressively.
				ax, ay := ai.advanceToward(nearest.X, nearest.Y, m, units)
				if (ax != ai.Unit.X || ay != ai.Unit.Y) && m.Passable(ax, ay) {
					actions = append(actions, AlienAction{
						Type: "move", Unit: ai.Unit,
						FromX: ai.Unit.X, FromY: ai.Unit.Y,
						ToX: ax, ToY: ay,
					})
				}
			} else {
				// Standard maneuver: seek cover while facing the target.
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
			// Adapt to grenade-heavy players: disperse from allies to minimize blast damage.
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
			// Last-resort fire if enough TU remains.
			ai.selectFireMode(int(dist))
			actions = append(actions, AlienAction{
				Type: "fire", Unit: ai.Unit, Target: target,
				FromX: ai.Unit.X, FromY: ai.Unit.Y,
				ToX: target.X, ToY: target.Y,
			})
		}

	case AISuppress:
		// Suppress: Fire at the target while staying in cover to keep the player pinned down.
		if nearest == nil {
			ai.State = AIAttack
			break
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

	case AISearch:
		// Search: Move toward the last known player position to re-establish contact.
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
					if s.Turn >= ai.Memory.turn-3 {
						sx, sy = s.X, s.Y
					}
				}
			}
			nx, ny := ai.moveTowardTargetCover(sx, sy, m, humanUnits)
			if (nx != ai.Unit.X || ny != ai.Unit.Y) && m.Passable(nx, ny) {
				actions = append(actions, AlienAction{
					Type: "move", Unit: ai.Unit,
					FromX: ai.Unit.X, FromY: ai.Unit.Y,
					ToX: nx, ToY: ny,
				})
			}
			ai.TurnsSince++
			if ai.TurnsSince > 6 {
				// Transition back to Patrol if no one is found after several turns.
				ai.State = AIPatrol
			}
		}

	case AIFlank:
		// Flank: Move toward a position that provides a side-on angle to the target.
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
			// Return to Attack state after a short flanking maneuver.
			ai.State = AIAttack
			ai.TurnsSince = 0
		}

	case AIRetreat:
		// Retreat: Move away from the nearest threat to safer ground.
		if nearest != nil {
			fx, fy := ai.retreatTarget(nearest, m, units)
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
			// Stop retreating and return to patrol after some distance is gained.
			ai.State = AIPatrol
		}
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
	bestScore := -999.0
	for _, h := range humanUnits {
		if !h.Alive || !ai.canSense(h.X, h.Y, m) {
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

		if at := ai.Unit.AlienType; at != nil && at.Morphology != nil {
			morph := at.Morphology
			if morph.ThermalSense == data.SenseHigh && h.Crouching {
				score += 4
			}
			if morph.ChemicalSense == data.SenseHigh && h.HP < h.MaxHP/2 {
				score += 5
			}
			if (morph.Eyesight == data.SenseExcellent || morph.Eyesight == data.SenseMultiSpec) && dist > 8 {
				score += 3
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
		if dist <= 10 {
			return true
		}
	}
	if morph.Hearing == data.SenseEcholoc {
		dx := float64(tx - ai.Unit.X)
		dy := float64(ty - ai.Unit.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= 6 {
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
	return true
}

func humanFrom(units UnitList) UnitList {
	var humans UnitList
	for _, u := range units {
		if u.Alive && u.Faction == 0 {
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
		if dist < 6 {
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
		score += ai.evaluateCoverVsThreats(nx, ny, m, humanFrom(units)) * 0.5
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

func (ai *AlienAI) retreatTarget(threat *Unit, m *BattleMap, units UnitList) (int, int) {
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
			protection := ai.evaluateCoverVsThreats(cx, cy, m, humanFrom(units))
			total := math.Max(protection, float64(cover))
			tdx := threat.X - cx
			tdy := threat.Y - cy
			threatDist := math.Sqrt(float64(tdx*tdx + tdy*tdy))
			if total > bestProtection && threatDist > 4 {
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

		score := -tDist + float64(cover)*8 - ai.reactionFirePenalty(nx, ny, m, units) + ai.evaluateCoverVsThreats(nx, ny, m, humanFrom(units))*0.5

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
			px := ai.Unit.X + ai.rng.Intn(12) - 6
			py := ai.Unit.Y + ai.rng.Intn(12) - 6
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

func (ai *AlienAI) selectFireMode(dist int) {
	w, ok := data.RuleItems[ai.Unit.Weapon]
	if !ok {
		return
	}
	modes := w.Modes()
	if len(modes) <= 1 {
		return
	}
	if dist <= 4 {
		for _, m := range modes {
			if m == data.FireModeAuto {
				ai.Unit.FireMode = data.FireModeAuto
				return
			}
		}
		for _, m := range modes {
			if m == data.FireModeBurst {
				ai.Unit.FireMode = data.FireModeBurst
				return
			}
		}
	} else if dist <= 8 {
		for _, m := range modes {
			if m == data.FireModeBurst {
				ai.Unit.FireMode = data.FireModeBurst
				return
			}
		}
	}
	ai.Unit.FireMode = data.FireModeAimed
}
