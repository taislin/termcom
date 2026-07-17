package battle

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// getAlienByRank returns the closest alien type at or above the given rank
// from a custom list (procedural aliens).
func getAlienByRank(types []*data.AlienType, minRank int) *data.AlienType {
	var best *data.AlienType
	for _, at := range types {
		if at.Rank >= minRank {
			if best == nil || at.Rank < best.Rank {
				best = at
			}
		}
	}
	return best
}

// BattlePhase represents the high-level state of the tactical combat.
type BattlePhase int

const (
	PhasePlayerTurn BattlePhase = iota // Player is actively controlling units
	PhaseAlienTurn                     // AI is executing queued actions
	PhaseVictory                       // Mission objectives achieved
	PhaseDefeat                        // Squad wiped or objective lost
)

// CombatStatus tracks the specific sub-state of a turn.
type CombatStatus int

const (
	StatusPlayerTurn CombatStatus = iota
	StatusAlienTurn
	StatusPlayerOverwatch // Triggered by an alien move
	StatusAlienOverwatch  // Triggered by a player move
)

// Game balance constants for battlescape actions.
const (
	MinReactionTU        = 15
	OverwatchFlashFrames  = 30
	GrenadeTUCost         = 20
	GrenadeBaseDamage     = 40
	GrenadeStrMult        = 2
	GrenadeRange          = 6
	GrenadeSplashSq       = 4
	GrenadeSplashFalloff  = 5
	GrenadeMinSplash      = 5
	MineBaseDamage        = 60
	MineDamageBonus       = 20
	ScannerTUCost         = 10
	ScannerRange          = 15
	ReactionMult          = 2
	ReactionAccDiv        = 3
	ReactionDistPen       = 5
	ReactionMinChance     = 1
	AlienGrenadeStrBonus  = 20
)

// Projectile represents a visual effect of a shot traveling across the map.
type Projectile struct {
	FromX, FromY int
	ToX, ToY     int
	Progress     int
	Length       int
	Symbol       rune
	Style        tcell.Style
}

// AlienAction defines a specific task for the Alien AI to execute during its turn.
type AlienAction struct {
	Type         string // "move", "fire", "melee", "patrol"
	Unit         *Unit
	Target       *Unit
	FromX, FromY int
	ToX, ToY     int
}

// LogEntry stores a single message with the turn it was created on.
type LogEntry struct {
	Text string
	Turn int
}

type Battlescape struct {
	// Game Engine and World State
	Game           *engine.Game
	Base           *base.Base
	Map            *BattleMap
	Units          UnitList
	AlienAIs       []*AlienAI
	CivilianAIs    []*CivilianAI
	Phase          BattlePhase
	Turn           int
	CursorX        int
	CursorY        int
	Selected       *Unit
	Message        string
	Log            []LogEntry
	ScrollX        int
	ScrollY        int
	Squad          []*soldier.Soldier
	UFOName        string
	ExitTimer      int
	IsNight        bool
	Status         CombatStatus
	OverwatchFlash int
	PlayerLock     int

	// Combat Tracking
	PlayerShotDistSum float64
	PlayerShotCount   int
	PlayerFlankShots  int

	// Turn Management
	AlienTurnQueue []AlienAction
	AlienTurnIdx   int
	ActionDelay    int
	Projectile     *Projectile

	// Visual Effects
	Camera         *engine.Camera
	Particles      *engine.ParticleSystem
	HoveredUnit    *Unit
	floaters       []FloatingText
	scannerPings   [][2]int // positions revealed by motion scanner this turn
	SidebarW       int
	ReinforceTimer int
	AbductionCivs  int
	AbductionTotal int

	Gas        *GasGrid
	VisionMode engine.VisionMode
	FrameCount int

	// Tactical Planning
	AlienSquadPlan *SquadPlan
	AlienMemory    *SquadMemory

	// Combat Resources
	PlayerGrenadeCount int

	// Pathfinding and Cache
	MovementCache    map[[2]int]bool
	MovementCacheKey int

	// Mission Objectives
	CustomVictory *CustomVictory

	// Placed mines (proximity)
	Mines []PlacedMine

	// Environmental State
	MissionModifiers      []MissionModifier
	Weather               Weather
	ReinforcementsSpawned bool

	// Crash site interior tiles (for placing alien crew inside UFO)
	CrashInterior []struct{ X, Y int }
	CrashExterior []struct{ X, Y int }

	// Input State
	State          BattleState
	mouseX, mouseY int // last mouse position for edge-scrolling
	QuitConfirm    bool
	mouseActive    bool
}

// The following setters mutate Battlescape fields that are also read by the
// render path (ApplyCursorStyles holds bs.State.mu via RLock). They must NOT
// take bs.State.mu themselves: input handlers already hold it via HandleEvent,
// and Update() takes it for the duration of its mutation pass. Re-locking here
// would deadlock (sync.RWMutex is not reentrant).
func (bs *Battlescape) SetPhase(p BattlePhase) {
	bs.Phase = p
}

func (bs *Battlescape) SetSelected(u *Unit) {
	bs.Selected = u
}

func (bs *Battlescape) SetHovered(u *Unit) {
	bs.HoveredUnit = u
}

func (bs *Battlescape) SetScroll(x, y int) {
	bs.ScrollX = x
	bs.ScrollY = y
}

type CustomVictory struct {
	Condition   string // "eliminate_all", "survive_turns", "reach_point"
	Turns       int    // for survive_turns
	TargetX     int    // for reach_point
	TargetY     int    // for reach_point
	MinSoldiers int    // for reach_point: min soldiers that must reach the point
}

// Touch-control view helper (engine defines the battleView interface).
func (bs *Battlescape) HasSelectedUnit() bool {
	return bs.Selected != nil && bs.Selected.Alive
}

func (bs *Battlescape) AddMessage(msg string) {
	bs.Message = msg
	bs.Log = append(bs.Log, LogEntry{Text: msg, Turn: bs.Turn})
	if len(bs.Log) > 50 {
		bs.Log = bs.Log[len(bs.Log)-50:]
	}
}

// GetMovementRange returns a map of tiles the selected unit can reach
func (bs *Battlescape) GetMovementRange() map[[2]int]bool {
	if bs.Selected == nil || bs.Selected.TU <= 0 {
		return nil
	}

	cacheKey := bs.Selected.X*10000 + bs.Selected.Y*100 + bs.Selected.TU + bs.Selected.Level*1000000
	if bs.MovementCache != nil && bs.MovementCacheKey == cacheKey {
		return bs.MovementCache
	}

	result := make(map[[2]int]bool)
	bs.MovementCache = result
	bs.MovementCacheKey = cacheKey

	startX, startY := bs.Selected.X, bs.Selected.Y
	maxTU := bs.Selected.TU

	// BFS to find reachable tiles
	type node struct {
		x, y, tu int
	}
	queue := []node{{startX, startY, maxTU}}
	visited := make(map[[2]int]bool)
	visited[[2]int{startX, startY}] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Add current tile to reachable set
		result[[2]int{current.x, current.y}] = true

		// Check all 4 directions
		dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
		for _, d := range dirs {
			nx, ny := current.x+d[0], current.y+d[1]
			if nx < 0 || nx >= bs.Map.Width || ny < 0 || ny >= bs.Map.Height {
				continue
			}
			if visited[[2]int{nx, ny}] {
				continue
			}

			tile := bs.Map.At(nx, ny)
			if !bs.Map.Passable(nx, ny) {
				continue
			}

			// TU cost: 4 for normal terrain, 8 for difficult terrain
			cost := 4
			if tile.Type == TileTree || tile.Type == TileRock || tile.Type == TileWater {
				cost = 8
			}

			remainingTU := current.tu - cost
			if remainingTU >= 0 {
				visited[[2]int{nx, ny}] = true
				queue = append(queue, node{nx, ny, remainingTU})
			}
		}
	}

	bs.MovementCache = result
	bs.MovementCacheKey = cacheKey
	return result
}

func (bs *Battlescape) CalculatePath(startX, startY, endX, endY int) [][2]int {
	if startX == endX && startY == endY {
		return [][2]int{{startX, startY}}
	}

	type node struct {
		x, y int
		path [][2]int
	}

	queue := []node{{startX, startY, [][2]int{{startX, startY}}}}
	visited := make(map[[2]int]bool)
	visited[[2]int{startX, startY}] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.x == endX && curr.y == endY {
			return curr.path
		}

		dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
		for _, d := range dirs {
			nx, ny := curr.x+d[0], curr.y+d[1]
			if nx < 0 || nx >= bs.Map.Width || ny < 0 || ny >= bs.Map.Height {
				continue
			}
			if visited[[2]int{nx, ny}] || !bs.Map.Passable(nx, ny) {
				continue
			}

			visited[[2]int{nx, ny}] = true
			newPath := make([][2]int, len(curr.path)+1)
			copy(newPath, curr.path)
			newPath[len(curr.path)] = [2]int{nx, ny}
			queue = append(queue, node{nx, ny, newPath})
		}
	}

	return nil
}

func NewBattlescape(g *engine.Game, b *base.Base, squad []*soldier.Soldier, ufoName string, crashSeed int64) *Battlescape {
	var m *BattleMap
	var crashResult *CrashResult
	switch ufoName {
	case "Terror":
		m = GenerateTerrorSite(50, 50)
	case "Supply Raid":
		m = GenerateUFOInterior(50, 50)
	case "Alien Base Assault":
		m = GenerateAlienBase(50, 50)
	case "Alien Research":
		m = GenerateUFOInterior(50, 50)
	case "Council":
		m = GenerateTerrorSite(50, 50)
	case "Cydonia":
		m = GenerateCydonia(50, 50)
	case "Abduction":
		m = GenerateAbductionSite(50, 50)
	case "Forest":
		m = GenerateForest(50, 50)
	case "Desert":
		m = GenerateDesert(50, 50)
	case "Polar":
		m = GeneratePolar(50, 50)
	default:
		m, crashResult = GenerateCrashSite(50, 50, crashSeed)
	}
	if m != nil && !m.ValidateMap() {
		// Fallback: clear to open ground so combat remains playable.
		m = NewBattleMap(50, 50)
	}

	rng := rand.New(rand.NewSource(int64(g.GameTime.UnixNano())))
	mods := RollModifiers(rng, ufoName)
	weather := RollWeather(rng, ufoName)

	bs := &Battlescape{
		Game:             g,
		Base:             b,
		Map:              m,
		Phase:            PhasePlayerTurn,
		Status:           StatusPlayerTurn,
		Turn:             1,
		CursorX:          3,
		CursorY:          m.Height - 3,
		Squad:            squad,
		UFOName:          ufoName,
		IsNight:          g.GameTime.Hour() < 6 || g.GameTime.Hour() > 18 || HasModifier(mods, ModNightOps),
		Camera:           engine.NewCamera(3, m.Height-3),
		Particles:        engine.NewParticleSystem(512),
		ActionDelay:      g.ActionDelay,
		Gas:              NewGasGrid(m.Width, m.Height),
		MissionModifiers: mods,
		Weather:          weather,
	}

	if ufoName == "Abduction" || ufoName == "Council" {
		bs.AbductionTotal = 8 + rand.Intn(5)
	}
	m.Gas = bs.Gas

	bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_START"), ufoName))

	if HasModifier(mods, ModNightOps) {
		bs.AddMessage(language.String("BATTLE_MOD_NIGHT_OPS"))
	}
	if HasModifier(mods, ModReinforcements) {
		bs.AddMessage(language.String("BATTLE_MOD_REINFORCEMENTS"))
	}
	if HasModifier(mods, ModTimeLimit) {
		bs.AddMessage(language.String("BATTLE_MOD_TIME_LIMIT"))
		bs.ReinforceTimer = 15
	}
	if HasModifier(mods, ModHeavyFog) {
		bs.AddMessage(language.String("BATTLE_MOD_HEAVY_FOG"))
	}
	if HasModifier(mods, ModAlienAmbush) {
		bs.AddMessage(language.String("BATTLE_MOD_ALIEN_AMBUSH"))
	}
	if HasModifier(mods, ModBoobyTrapped) {
		bs.AddMessage(language.String("BATTLE_MOD_BOOBY_TRAPPED"))
		// Place 3-5 random mines on floor tiles away from spawn
		numMines := 3 + rng.Intn(3)
		for placed := 0; placed < numMines; placed++ {
			for attempt := 0; attempt < 20; attempt++ {
				mx := 5 + rng.Intn(m.Width-10)
				my := 5 + rng.Intn(m.Height-10)
				if m.At(mx, my).Type != TileFloor {
					continue
				}
				// Don't place near player spawn
				nearSpawn := false
				for _, s := range squad {
					if abs(mx-s.PosX) < 8 && abs(my-s.PosY) < 8 {
						nearSpawn = true
						break
					}
				}
				if nearSpawn {
					continue
				}
				// Don't double-place
				alreadyMined := false
				for _, pm := range bs.Mines {
					if pm.X == mx && pm.Y == my {
						alreadyMined = true
						break
					}
				}
				if alreadyMined {
					continue
				}
				bs.Mines = append(bs.Mines, PlacedMine{X: mx, Y: my})
				break
			}
		}
	}
	if !weather.IsClear() {
		bs.AddMessage(language.Sprintf("BATTLE_WEATHER", weather.Name()))
	}

	for i, s := range squad {
		if s.HP <= 0 || s.Wounds > 0 {
			continue
		}
		u := NewSoldierUnit(s)
		u.X = 3 + i*2
		u.Y = m.Height - 3
		u.IsNight = bs.IsNight
		bs.Units = append(bs.Units, u)
	}

	alienTypes := g.GetAlienTypes()
	gameMonth := int(g.GameTime.Month()) - 3 + (g.GameTime.Year()-1999)*12
	if gameMonth < 0 {
		gameMonth = 0
	}

	diffMult := 1.0
	if g.Difficulty >= 0 && g.Difficulty < len(engine.Difficulties) {
		diffMult = engine.Difficulties[g.Difficulty].AlienScale
	}

	// Scale base rank with game time: +1 rank per 3 months
	alienRank := gameMonth / 3
	if alienRank > 3 {
		alienRank = 3
	}

	// Scale alien count with game time
	baseCount := 5
	extraCount := gameMonth / 2
	if extraCount > 5 {
		extraCount = 5
	}
	totalAliens := baseCount + extraCount

	// Stat bonus: +2 HP and +3 Accuracy per month (capped), scaled by difficulty
	hpBonus := int(float64(gameMonth*2) * diffMult)
	if hpBonus > 20 {
		hpBonus = 20
	}
	accBonus := int(float64(gameMonth*3) * diffMult)
	if accBonus > 30 {
		accBonus = 30
	}

	// Equipment escalation tier (Phase 30)
	equipTier := data.GetAlienEquipTier(gameMonth)

	spawnAliens := make([]*data.AlienType, 0, totalAliens)
	for i := 0; i < totalAliens; i++ {
		rank := alienRank
		// Upper half gets +1 rank
		if i >= totalAliens/2 {
			rank++
		}
		if rank > 5 {
			rank = 5
		}
		at := getAlienByRank(alienTypes, rank)
		spawnAliens = append(spawnAliens, at)
	}

	// Store crash interior for alien placement inside the UFO, and collect
	// nearby outdoor tiles so some aliens can spawn as perimeter guards.
	if crashResult != nil {
		for _, pt := range crashResult.InteriorTiles {
			bs.CrashInterior = append(bs.CrashInterior, struct{ X, Y int }{pt.X, pt.Y})
		}
		for _, pt := range crashResult.ExteriorTiles {
			bs.CrashExterior = append(bs.CrashExterior, struct{ X, Y int }{pt.X, pt.Y})
		}
	}

	tierHP, tierArmor := data.GetTierStatBonus(equipTier)
	for i, at := range spawnAliens {
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
		// For crash sites, roughly half the crew spawns inside the UFO and the
		// rest deploy as perimeter guards on nearby outdoor tiles, so they are
		// not all clustered in one place.
		if len(bs.CrashInterior) > 0 {
			if i%2 == 0 || len(bs.CrashExterior) == 0 {
				tile := bs.CrashInterior[rand.Intn(len(bs.CrashInterior))]
				u.X = tile.X
				u.Y = tile.Y
			} else {
				tile := bs.CrashExterior[rand.Intn(len(bs.CrashExterior))]
				u.X = tile.X
				u.Y = tile.Y
			}
		} else {
			u.X = 10 + randn(m.Width-14)
			u.Y = 3 + randn(m.Height/2-4)
		}
		u.IsNight = bs.IsNight
		u.HP += hpBonus + tierHP
		u.MaxHP += hpBonus + tierHP
		u.Armour += tierArmor
		u.Accuracy += accBonus
		// Upgrade weapon based on equipment tier
		u.Weapon = data.GetTierWeapon(equipTier, at.Rank)
		if ammo, ok := data.RuleItems[u.Weapon]; ok {
			u.WeaponAmmo = ammo.AmmoMax
		}
		bs.Units = append(bs.Units, u)
		ai := NewAlienAI(u)
		ai.PatrolX = u.X + rand.Intn(6) - 3
		ai.PatrolY = u.Y + rand.Intn(6) - 3
		bs.AlienAIs = append(bs.AlienAIs, ai)
	}

	if ufoName == "Terror" {
		civCount := 8 + rand.Intn(5)
		for i := 0; i < civCount; i++ {
			name := civNames[rand.Intn(len(civNames))]
			u := NewCivilianUnit(name)
			u.X = 5 + randn(m.Width-10)
			u.Y = m.Height/2 + randn(m.Height/2-5)
			if m.Passable(u.X, u.Y) {
				u.IsNight = bs.IsNight
				bs.Units = append(bs.Units, u)
				bs.CivilianAIs = append(bs.CivilianAIs, NewCivilianAI(u))
			}
		}
	}

	if ufoName == "Abduction" {
		for i := 0; i < bs.AbductionTotal; i++ {
			name := civNames[rand.Intn(len(civNames))]
			u := NewCivilianUnit(name)
			u.X = 10 + randn(m.Width-20)
			u.Y = 5 + randn(m.Height-10)
			if m.Passable(u.X, u.Y) {
				u.IsNight = bs.IsNight
				bs.Units = append(bs.Units, u)
				bs.CivilianAIs = append(bs.CivilianAIs, NewCivilianAI(u))
			}
		}
	}

	bs.ComputeFOVForTeam()

	return bs
}

// CustomUnitDef defines a unit placement for a custom battle.
type CustomUnitDef struct {
	Name       string
	HP         int
	TU         int
	Accuracy   int
	Bravery    int
	Reactions  int
	Strength   int
	Psi        int
	Armour     int
	Weapon     string
	Rank       int
	DamageType int
	Aggression int
	Faction    int // 0=player(soldier), 1=alien, 2=civilian
	X, Y       int
	Armor      string // for soldiers
}

// NewCustomBattlescape creates a battlescape with explicit unit placements for custom battles.
func NewCustomBattlescape(g *engine.Game, b *base.Base, squad []*soldier.Soldier, m *BattleMap, units []CustomUnitDef, cv *CustomVictory, ufoName string) *Battlescape {
	bs := &Battlescape{
		Game:          g,
		Base:          b,
		Map:           m,
		Phase:         PhasePlayerTurn,
		Status:        StatusPlayerTurn,
		Turn:          1,
		CursorX:       3,
		CursorY:       m.Height - 3,
		Squad:         squad,
		UFOName:       ufoName,
		IsNight:       g.GameTime.Hour() < 6 || g.GameTime.Hour() > 18,
		Camera:        engine.NewCamera(3, m.Height-3),
		Particles:     engine.NewParticleSystem(512),
		ActionDelay:   g.ActionDelay,
		Gas:           NewGasGrid(m.Width, m.Height),
		CustomVictory: cv,
	}
	m.Gas = bs.Gas

	for _, def := range units {
		switch def.Faction {
		case 0: // soldier
			if def.Name != "" {
				s := soldier.NewSoldier(def.Name)
				s.Rank = soldier.Rank(def.Rank)
				if def.HP > 0 {
					s.HP = def.HP
					s.MaxHP = def.HP
				}
				if def.TU > 0 {
					s.TU = def.TU
					s.MaxTU = def.TU
				}
				if def.Accuracy > 0 {
					s.Accuracy = def.Accuracy
				}
				if def.Reactions > 0 {
					s.Reactions = def.Reactions
				}
				if def.Strength > 0 {
					s.Strength = def.Strength
				}
				if def.Weapon != "" {
					s.Weapon = def.Weapon
					s.WeaponAmmo = data.RuleItems[def.Weapon].AmmoMax
				}
				if def.Armor != "" {
					s.Armor = def.Armor
				}
				u := NewSoldierUnit(s)
				u.X = def.X
				u.Y = def.Y
				u.IsNight = bs.IsNight
				bs.Units = append(bs.Units, u)
			}
		case 1: // alien
			at := &data.AlienType{
				Name:       def.Name,
				HP:         def.HP,
				TU:         def.TU,
				Accuracy:   def.Accuracy,
				Bravery:    def.Bravery,
				Reactions:  def.Reactions,
				Strength:   def.Strength,
				Psi:        def.Psi,
				Armour:     def.Armour,
				Weapon:     def.Weapon,
				Rank:       def.Rank,
				DamageType: def.DamageType,
				Aggression: def.Aggression,
			}
			u := NewAlienUnit(at)
			u.X = def.X
			u.Y = def.Y
			u.IsNight = bs.IsNight
			if m.Passable(u.X, u.Y) {
				bs.Units = append(bs.Units, u)
				ai := NewAlienAI(u)
				ai.PatrolX = u.X + rand.Intn(6) - 3
				ai.PatrolY = u.Y + rand.Intn(6) - 3
				bs.AlienAIs = append(bs.AlienAIs, ai)
			}
		case 2: // civilian
			u := NewCivilianUnit(def.Name)
			u.X = def.X
			u.Y = def.Y
			u.IsNight = bs.IsNight
			if m.Passable(u.X, u.Y) {
				bs.Units = append(bs.Units, u)
				bs.CivilianAIs = append(bs.CivilianAIs, NewCivilianAI(u))
			}
		}
	}

	if bs.CustomVictory != nil && bs.CustomVictory.Condition == "survive_turns" {
		bs.AddMessage(language.Sprintf("BATTLE_SURVIVE_TURNS", bs.CustomVictory.Turns))
	} else {
		bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_START"), ufoName))
	}

	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Alive {
			bs.CursorX = u.X
			bs.CursorY = u.Y
			bs.Camera.SetTarget(u.X, u.Y)
			break
		}
	}
	bs.ComputeFOVForTeam()
	return bs
}

func (bs *Battlescape) ComputeFOVForTeam() {
	bs.Map.ClearVisibility()
	sightRange := SightRange
	if bs.IsNight {
		sightRange = 10
	}
	sightRange -= bs.Weather.SightReduction()
	if sightRange < 3 {
		sightRange = 3
	}
	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Alive && u.Level == bs.Map.CurrentLevel {
			bs.Map.ComputeFOV(u.X, u.Y, sightRange)
		}
	}
	for _, u := range bs.Units {
		if u.Faction == FactionAlien && u.Alive && u.AlienType != nil && u.Level == bs.Map.CurrentLevel && bs.Map.IsVisible(u.X, u.Y) {
			bs.Game.LearnAlien(u.AlienType.Name, 1)
		}
	}
	bs.Gas.Visible = func(x, y int) bool {
		return bs.Map.IsVisible(x, y)
	}
}

func (bs *Battlescape) Update() {
	// Serialize mutations of shared Battlescape state (Phase, Selected, etc.)
	// with the render path's RLock in ApplyCursorStyles. Input handlers hold
	// the same lock via HandleEvent; this keeps all writers consistent without
	// re-locking (sync.RWMutex is not reentrant).
	bs.State.mu.Lock()
	defer bs.State.mu.Unlock()

	dt := 0.016 // Fixed delta time for visual updates
	bs.FrameCount++
	bs.Camera.UpdateShake(dt)
	bs.Particles.Update(dt)
	bs.updateFloaters(dt)

	// Handle overwatch flash effect: when an alien is triggered,
	// they flash briefly before performing their action.
	if bs.OverwatchFlash > 0 {
		bs.OverwatchFlash--
		if bs.OverwatchFlash == 0 {
			if bs.Phase == PhaseAlienTurn {
				bs.Status = StatusAlienTurn
			} else {
				bs.Status = StatusPlayerTurn
			}
		}
	}

	// Handle input locking (e.g., during animations or dialogs)
	if bs.PlayerLock > 0 {
		bs.PlayerLock--
	}

	// Edge-scrolling: auto-pan when mouse is near screen borders
	const edgeMargin = 3
	if bs.mouseActive && bs.Phase == PhasePlayerTurn && bs.PlayerLock == 0 && bs.FrameCount%2 == 0 {
		w, h := bs.Game.ScreenSize()
		dx, dy := 0, 0
		if bs.mouseX <= edgeMargin {
			dx = -2
		} else if bs.mouseX >= w-1-edgeMargin {
			dx = 2
		}
		if bs.mouseY <= edgeMargin {
			dy = -2
		} else if bs.mouseY >= h-1-edgeMargin {
			dy = 2
		}
		if dx != 0 || dy != 0 {
			bs.Camera.Pan(dx, dy)
		}
	}

	// Periodic ambient sound effects
	if bs.FrameCount%1800 == 0 {
		audio.PlayWind()
	}

	if bs.FrameCount%12 == 0 && bs.Phase != PhaseVictory && bs.Phase != PhaseDefeat {
		w, h := bs.Game.ScreenSize()
		viewW := engine.Layout.BattleViewWidth(w)
		viewH := engine.Layout.BattleViewHeight(h)
		switch bs.UFOName {
		case "Polar":
			engine.SpawnSnow(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, 1)
		case "Desert":
			engine.SpawnDust(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, viewH)
		case "Cydonia", "Alien Base Assault":
			engine.SpawnEmbers(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, viewH)
		}
	}

	// If a projectile is mid-flight, advance it every frame regardless of phase
	// (alien reaction fire is spawned during the player's own turn, so gating this
	// on PhaseAlienTurn would leave the tracer frozen on the map).
	if bs.Projectile != nil {
		bs.Projectile.Progress++
		if bs.Projectile.Progress >= bs.Projectile.Length {
			bs.Projectile = nil
		}
		return
	}

	// Process Alien Turn: execute actions from the queue with a delay between them.
	if bs.Phase == PhaseAlienTurn {
		// Delay between individual alien actions for visual pacing.
		if bs.ActionDelay > 0 {
			bs.ActionDelay--
			return
		}

		// Execute the next action in the queue.
		if bs.AlienTurnIdx < len(bs.AlienTurnQueue) {
			action := bs.AlienTurnQueue[bs.AlienTurnIdx]
			bs.AlienTurnIdx++
			bs.executeAlienAction(action)
			bs.checkMineTriggers()
			bs.ActionDelay = bs.Game.ActionDelay // Use configured action delay
		} else {
			// After all queued actions, process civilian behaviors and finish turn.
			for _, cai := range bs.CivilianAIs {
				actions := cai.GenerateActions(bs.Units, bs.Map)
				for _, a := range actions {
					bs.executeAlienAction(a)
				}
			}
			bs.finishAlienTurn()
		}
	}

	// Handle mission end timers.
	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		bs.ExitTimer++
		if bs.ExitTimer > 60 {
			bs.finishBattle()
		}
	}
}

// spawnShotFX creates the visual projectile and hit/miss effects for a shot.
// shooter is the unit firing, target is the unit being shot at, damage/hit/coverHit
// are the FireAt results, and missMsg/killMsg are the message keys for misses and kills.
func (bs *Battlescape) spawnShotFX(shooter, target *Unit, damage int, hit, coverHit bool, style tcell.Style, missMsg, hitMsg, killMsg string) {
	dx := target.X - shooter.X
	dy := target.Y - shooter.Y
	length := int(math.Sqrt(float64(dx*dx + dy*dy)))
	if length < 1 {
		length = 1
	}
	symbol := '*'
	if data.RuleItems[shooter.Weapon].AmmoMax >= 99 {
		symbol = '|'
	}
	bs.Projectile = &Projectile{
		FromX: shooter.X, FromY: shooter.Y,
		ToX: target.X, ToY: target.Y,
		Progress: 0, Length: length,
		Symbol: symbol,
		Style:  style,
	}
	engine.SpawnMuzzleFlash(bs.Particles, shooter.X-bs.ScrollX+1, shooter.Y-bs.ScrollY+1)
	if hit {
		engine.SpawnExplosion(bs.Particles, target.X-bs.ScrollX+1, target.Y-bs.ScrollY+1, color.NewRGBColor(255, 80, 30), 8)
		if engine.Config.ScreenShake {
			bs.Camera.TriggerShake(0.5)
		}
		bs.SpawnBloodSplatter(target)
		bs.spawnFloater(target.X, target.Y, fmt.Sprintf("-%d", damage), color.XTerm9)
		name := target.Name()
		bs.AddMessage(fmt.Sprintf(language.String(hitMsg), damage, name, target.HP))
		if !target.Alive {
			bs.AddMessage(fmt.Sprintf(language.String(killMsg), name))
		}
	} else {
		if coverHit {
			bs.AddMessage(language.String("MSG_HIT_COVER"))
		} else {
			bs.AddMessage(language.String(missMsg))
		}
	}
}

// spawnAlienWave creates a group of alien units spawned from map edges with
// difficulty-scaled stats. Returns the number of aliens actually spawned.
func (bs *Battlescape) spawnAlienWave(count int) int {
	g := bs.Game
	alienTypes := g.GetAlienTypes()
	gameMonth := int(g.GameTime.Month()) - 3 + (g.GameTime.Year()-1999)*12
	if gameMonth < 0 {
		gameMonth = 0
	}
	diffMult := 1.0
	if g.Difficulty >= 0 && g.Difficulty < len(engine.Difficulties) {
		diffMult = engine.Difficulties[g.Difficulty].AlienScale
	}
	alienRank := gameMonth / 3
	if alienRank > 3 {
		alienRank = 3
	}
	if bs.Turn > 6 {
		alienRank++
	}
	if bs.Turn > 12 {
		alienRank++
	}
	if alienRank > 5 {
		alienRank = 5
	}
	equipTier := data.GetAlienEquipTier(gameMonth)
	tierHP, tierArmor := data.GetTierStatBonus(equipTier)
	spawned := 0
	for i := 0; i < count; i++ {
		at := getAlienByRank(alienTypes, alienRank)
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
		u.HP += int(float64(gameMonth*2)*diffMult) + tierHP
		u.MaxHP += int(float64(gameMonth*2)*diffMult) + tierHP
		u.Armour += tierArmor
		u.Accuracy += int(float64(gameMonth*3) * diffMult)
		u.Weapon = data.GetTierWeapon(equipTier, at.Rank)
		if ammo, ok := data.RuleItems[u.Weapon]; ok {
			u.WeaponAmmo = ammo.AmmoMax
		}
		side := rand.Intn(4)
		switch side {
		case 0:
			u.X = 0
			u.Y = rand.Intn(bs.Map.Height)
		case 1:
			u.X = bs.Map.Width - 1
			u.Y = rand.Intn(bs.Map.Height)
		case 2:
			u.X = rand.Intn(bs.Map.Width)
			u.Y = 0
		case 3:
			u.X = rand.Intn(bs.Map.Width)
			u.Y = bs.Map.Height - 1
		}
		if !bs.Map.Passable(u.X, u.Y) {
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					nx, ny := u.X+dx, u.Y+dy
					if nx >= 0 && nx < bs.Map.Width && ny >= 0 && ny < bs.Map.Height && bs.Map.Passable(nx, ny) {
						u.X, u.Y = nx, ny
						dy = 3
						dx = 3
					}
				}
			}
		}
		u.IsNight = bs.IsNight
		bs.Units = append(bs.Units, u)
		ai := NewAlienAI(u)
		ai.State = AIAttack
		nearest, _ := ai.findNearest(bs.Units.Faction(0), bs.Map)
		if nearest != nil {
			ai.LastSeenX = nearest.X
			ai.LastSeenY = nearest.Y
		}
		bs.AlienAIs = append(bs.AlienAIs, ai)
		spawned++
	}
	return spawned
}

func (bs *Battlescape) executeAlienAction(action AlienAction) {
	if action.Unit.TU <= 0 || !action.Unit.Alive {
		return
	}
	switch action.Type {
	case "fire":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		damage, hit, coverHit, err := action.Unit.FireAt(action.Target, bs.Map, &bs.Weather)
		if err != nil {
			return
		}
		audio.PlayWeaponFire(action.Unit.Weapon)
		bs.spawnShotFX(action.Unit, action.Target, damage, hit, coverHit, engine.StyleYellow, "MSG_ALIEN_MISS", "MSG_ALIEN_HIT", "MSG_ALIEN_KILL")
	case "melee":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		if action.Unit.TU < 20 {
			return
		}
		action.Unit.TU -= 20
		audio.PlayMeleeFire()
		damage := action.Unit.Strength + rand.Intn(10)
		damage -= action.Target.Armour
		if damage < 1 {
			damage = 1
		}
		action.Target.HP -= damage
		if action.Target.HP <= 0 {
			action.Target.Alive = false
		}
		bs.SpawnBloodSplatter(action.Target)
		bs.Projectile = &Projectile{
			FromX: action.Unit.X, FromY: action.Unit.Y,
			ToX: action.Target.X, ToY: action.Target.Y,
			Progress: 0, Length: 1,
			Symbol: 'X',
			Style:  engine.StyleRedBold,
		}
		name := action.Target.Name()
		bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_MELEE"), action.Unit.AlienType.LangName(), name, damage))
		if !action.Target.Alive {
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_KILL"), name))
		}
	case "move":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
		bs.ComputeFOVForTeam()
		bs.checkHumanReactionFire(action.Unit)
	case "patrol":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
	case "psi":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		if action.Unit.TU < 20 {
			return
		}
		action.Unit.TU -= 20
		audio.PlayLaserFire()
		engine.SpawnExplosion(bs.Particles, action.Target.X-bs.ScrollX+1, action.Target.Y-bs.ScrollY+1, color.NewRGBColor(120, 0, 200), 12)
		if engine.Config.ScreenShake {
			bs.Camera.TriggerShake(0.3)
		}
		attackerPsi := 0
		if action.Unit.AlienType != nil {
			attackerPsi = action.Unit.AlienType.Psi
		}
		defenderPsiStr := 0
		if action.Target.Soldier != nil {
			defenderPsiStr = action.Target.Soldier.PsiStr
		} else {
			defenderPsiStr = action.Target.PsiStr
		}
		successChance := attackerPsi - defenderPsiStr/3
		if successChance < 5 {
			successChance = 5
		}
		success := rand.Intn(100) < successChance
		if success {
			action.Target.TU = 0
			action.Target.Panicked = true
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_PSI_PANIC"), action.Unit.AlienType.LangName(), action.Target.Name()))
		} else {
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_PSI_RESIST"), action.Unit.AlienType.LangName(), action.Target.Name()))
		}
		bs.ComputeFOVForTeam()
		bs.checkHumanReactionFire(action.Unit)
	case "grenade":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		ax := action.ToX
		ay := action.ToY
		damage := GrenadeBaseDamage
		if action.Unit.AlienType != nil {
			damage = action.Unit.AlienType.Strength*GrenadeStrMult + AlienGrenadeStrBonus
		}
		for _, u := range bs.Units {
			if !u.Alive {
				continue
			}
			udx := u.X - ax
			udy := u.Y - ay
			udist := udx*udx + udy*udy
			if udist <= GrenadeSplashSq {
				splashDmg := damage - udist*GrenadeSplashFalloff
				if splashDmg < GrenadeMinSplash {
					splashDmg = 5
				}
				if u.Faction == action.Unit.Faction && u != action.Unit {
					splashDmg /= 3
				}
				u.HP -= splashDmg
				if u.HP <= 0 {
					u.HP = 0
					u.Alive = false
				}
				bs.SpawnBloodSplatter(u)
				bs.spawnFloater(u.X, u.Y, fmt.Sprintf("-%d", splashDmg), color.XTerm9)
			}
		}
		splashRadius := 2
		for dy := -splashRadius; dy <= splashRadius; dy++ {
			for dx := -splashRadius; dx <= splashRadius; dx++ {
				if dx*dx+dy*dy > splashRadius*splashRadius {
					continue
				}
				tx, ty := ax+dx, ay+dy
				if bs.Map.IsDestructible(tx, ty) {
					if bs.Map.DestroyWall(tx, ty) {
						SpawnRubble(bs.Particles, tx-bs.ScrollX+1, ty-bs.ScrollY+1)
					}
				}
				if tx >= 0 && tx < bs.Map.Width && ty >= 0 && ty < bs.Map.Height {
					tile := bs.Map.At(tx, ty)
					if tile.IsFlammable() && tile.Fire <= 0 && rand.Intn(3) == 0 {
						bs.SpawnFire(tx, ty, 3)
					}
				}
			}
		}
		bs.Gas.Set(ax, ay, 3, GasSmoke)
		for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			bs.Gas.Set(ax+d[0], ay+d[1], 2, GasSmoke)
		}
		audio.PlayGrenade()
		bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_GRENADE"), action.Unit.AlienType.LangName(), ax, ay))
		if engine.Config.ScreenShake {
			bs.Camera.TriggerShake(2.5)
		}
		engine.SpawnExplosion(bs.Particles, ax-bs.ScrollX+1, ay-bs.ScrollY+1, tcell.NewRGBColor(255, 180, 50), 24)
		engine.SpawnSmoke(bs.Particles, ax-bs.ScrollX+1, ay-bs.ScrollY+1, 8)
		bs.ComputeFOVForTeam()
		bs.checkHumanReactionFire(action.Unit)
	}
}

func (bs *Battlescape) checkHumanReactionFire(movedAlien *Unit) {
	if !movedAlien.Alive {
		return
	}
	for _, u := range bs.Units {
		if u.Faction != 0 || !u.Alive {
			continue
		}
		if u.TU < MinReactionTU || u.Weapon == "" {
			continue
		}
		dx := movedAlien.X - u.X
		dy := movedAlien.Y - u.Y
		dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist == 0 {
			dist = 1
		}
		if dist > SightRange {
			continue
		}
		if !bs.Map.IsVisible(movedAlien.X, movedAlien.Y) {
			continue
		}
		if !u.CanSee(movedAlien.X, movedAlien.Y, bs.Map) {
			continue
		}
		chance := u.Reactions*ReactionMult + u.Accuracy/ReactionAccDiv - dist*ReactionDistPen
		if chance < ReactionMinChance {
			chance = ReactionMinChance
		}
		if rand.Intn(100) >= chance {
			continue
		}
		w, ok := data.RuleItems[u.Weapon]
		if !ok || w.AmmoMax < 99 && u.WeaponAmmo <= 0 {
			continue
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_FIRE"), u.Name(), movedAlien.Name()))
		audio.PlayWeaponFire(u.Weapon)
		bs.Status = StatusPlayerOverwatch
		bs.OverwatchFlash = OverwatchFlashFrames
		bs.Camera.SetTarget(u.X, u.Y)
		u.InOverwatch = true
		damage, hit, coverHit, _ := u.FireAt(movedAlien, bs.Map, &bs.Weather)
		u.InOverwatch = false
		bs.recordPlayerShot(u, movedAlien)
		if hit && u.Soldier != nil {
			u.Soldier.AddReactionsExp()
		}
		bs.spawnShotFX(u, movedAlien, damage, hit, coverHit, engine.StyleCyanBold, "MSG_REACTION_MISS", "MSG_REACTION_HIT", "MSG_REACTION_KILL")
		return
	}
}

func (bs *Battlescape) checkAlienReactionFire(movedHuman *Unit) {
	if !movedHuman.Alive {
		return
	}
	for _, ai := range bs.AlienAIs {
		u := ai.Unit
		if !u.Alive {
			continue
		}
		if u.TU < MinReactionTU || u.Weapon == "" {
			continue
		}
		dx := movedHuman.X - u.X
		dy := movedHuman.Y - u.Y
		dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist == 0 {
			dist = 1
		}
		if dist > SightRange {
			continue
		}
		if !u.CanSee(movedHuman.X, movedHuman.Y, bs.Map) {
			continue
		}
		chance := u.Reactions*ReactionMult + u.Accuracy/ReactionAccDiv - dist*ReactionDistPen
		if chance < ReactionMinChance {
			chance = ReactionMinChance
		}
		if rand.Intn(100) >= chance {
			continue
		}
		w, ok := data.RuleItems[u.Weapon]
		if !ok || w.AmmoMax < 99 && u.WeaponAmmo <= 0 {
			continue
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_FIRE"), u.Name(), movedHuman.Name()))
		audio.PlayWeaponFire(u.Weapon)
		bs.Status = StatusAlienOverwatch
		bs.OverwatchFlash = OverwatchFlashFrames
		bs.Camera.SetTarget(u.X, u.Y)
		u.InOverwatch = true
		damage, hit, coverHit, _ := u.FireAt(movedHuman, bs.Map, &bs.Weather)
		u.InOverwatch = false
		if hit && u.Soldier != nil {
			u.Soldier.AddReactionsExp()
		}
		bs.spawnShotFX(u, movedHuman, damage, hit, coverHit, engine.StyleRedBold, "MSG_REACTION_MISS", "MSG_REACTION_HIT", "MSG_REACTION_KILL")
		return
	}
}

// exitBattle aborts the mission — all human units are killed and the battle ends in defeat.
func (bs *Battlescape) exitBattle() {
	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Alive {
			u.HP = 0
			u.Alive = false
		}
	}
	bs.SetPhase(PhaseDefeat)
	bs.AddMessage(language.String("MSG_BATTLE_EXITED"))
}

func (bs *Battlescape) finishBattle() {
	won := bs.Phase == PhaseVictory

	// Sync soldier HP back to roster
	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Soldier != nil {
			u.Soldier.HP = u.HP
			if u.HP <= 0 {
				u.Soldier.HP = 0
				u.Soldier.Wounds = 30
				alreadyMemorial := false
				for _, m := range bs.Game.Memorial {
					if m == u.Soldier {
						alreadyMemorial = true
						break
					}
				}
				if !alreadyMemorial {
					bs.Game.Memorial = append(bs.Game.Memorial, u.Soldier)
				}
			} else {
				dmg := u.MaxHP - u.HP
				if dmg < 0 {
					dmg = 0
				}
				wounds := dmg*3 + u.FatalWounds*2
				if wounds > 30 {
					wounds = 30
				}
				u.Soldier.Wounds = wounds
			}
			// Fatigue: survivors need rest
			if u.HP > 0 {
				fatigueDays := 1 + bs.Turn/3
				if fatigueDays > 5 {
					fatigueDays = 5
				}
				u.Soldier.Fatigue += fatigueDays
			}
		}
	}

	// Count alien kills
	alienKills := 0
	for _, u := range bs.Units {
		if u.Faction == FactionAlien && !u.Alive {
			alienKills++
		}
	}

	// Award XP to surviving soldiers
	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Soldier != nil {
			u.Soldier.Kills += alienKills
			u.Soldier.PostMission()
			u.Soldier.Missions++
		}
	}

	if bs.Base != nil {
		soldier.HandlePromotions(bs.Base.Soldiers)
	}

	// Collect loot — type-specific corpses + weapon drops
	var loot []string
	var stunnedAliens []string
	if won {
		corpseMap := map[string]string{
			"SEC": "corpse_sect",
			"SEL": "corpse_sect",
			"SEN": "corpse_sect",
			"FLT": "corpse_float",
			"FLL": "corpse_float",
			"FLN": "corpse_float",
			"MUT": "corpse_muton",
			"MUL": "corpse_muton",
			"MUN": "corpse_muton",
			"ETH": "corpse_ether",
			"EHL": "corpse_ether",
			"ETN": "corpse_ether",
		}
		corpses := make(map[string]bool)
		weaponDrops := make(map[string]bool)
		for _, u := range bs.Units {
			if u.Faction == FactionAlien {
				if u.Stunned {
					stunnedAliens = append(stunnedAliens, u.AlienType.Name)
				} else if !u.Alive && u.AlienType != nil {
					if key, ok := corpseMap[u.AlienType.ShortName]; ok {
						corpses[key] = true
					}
					// Higher rank aliens drop weapons more often
					dropChance := 15 + u.AlienType.Rank*10
					if rand.Intn(100) < dropChance {
						weaponDrops[u.AlienType.Weapon] = true
					}
				}
			}
		}
		for key := range corpses {
			loot = append(loot, key)
		}
		for wpn := range weaponDrops {
			if item, ok := data.RuleItems[wpn]; ok && item.CostSell > 0 && !item.IsAlien {
				loot = append(loot, wpn)
			}
		}
		if len(loot) == 0 && len(stunnedAliens) == 0 {
			loot = append(loot, "alien_corpse")
		}
		if rand.Intn(100) < 35 {
			loot = append(loot, "alloys")
		}
		if rand.Intn(100) < 20 {
			loot = append(loot, "elerium")
		}
	}

	// Find surviving squad soldiers
	var surviving []*soldier.Soldier
	for _, s := range bs.Squad {
		for _, u := range bs.Units {
			if u.Faction == FactionHuman && u.Soldier == s {
				surviving = append(surviving, u.Soldier)
				break
			}
		}
	}
	// Add soldiers that weren't deployed (HP <= 0 at start)
	for _, s := range bs.Squad {
		found := false
		for _, u := range bs.Units {
			if u.Soldier == s {
				found = true
				break
			}
		}
		if !found {
			surviving = append(surviving, s)
		}
	}

	// Award funds
	if won {
		bs.Game.Funds += 50000
	}

	// Update tactical tracking (running averages weighted by prior battles)
	t := &bs.Game.Tactics
	prev := t.BattleCount
	if bs.PlayerShotCount > 0 {
		thisRange := bs.PlayerShotDistSum / float64(bs.PlayerShotCount)
		if prev > 0 {
			t.AverageRange = (t.AverageRange*float64(prev) + thisRange) / float64(prev+1)
		} else {
			t.AverageRange = thisRange
		}
	}
	if bs.PlayerFlankShots > 0 {
		thisFlank := float64(bs.PlayerFlankShots)
		if prev > 0 {
			t.FlankingObserved = int((float64(t.FlankingObserved)*float64(prev) + thisFlank) / float64(prev+1))
		} else {
			t.FlankingObserved = bs.PlayerFlankShots
		}
	}
	t.BattleCount++
	t.TotalAlienKills += alienKills
	soldierLosses := 0
	for _, s := range surviving {
		if s.HP <= 0 {
			soldierLosses++
		}
	}
	t.TotalSoldierLosses += soldierLosses
	t.GrenadeUsage += bs.PlayerGrenadeCount

	bs.Game.ActiveBattle = &engine.BattleResult{
		Won:           won,
		Kills:         alienKills,
		Soldiers:      surviving,
		LootItems:     loot,
		StunnedAliens: stunnedAliens,
	}
	bs.Game.PopState()
}

func (bs *Battlescape) finishAlienTurn() {
	bs.SetPhase(PhasePlayerTurn)
	bs.Status = StatusPlayerTurn
	bs.restorePlayerTU()
	bs.ComputeFOVForTeam()
	bs.Turn++
	bs.scannerPings = nil
	bs.learnFromKills()
	bs.Gas.Diffuse()
	bs.Map.SpreadFire()
	if bs.UFOName == "Abduction" {
		bs.processAbduction()
	}
	bs.checkReinforcements()
	bs.checkVictory()

	oldSelected := bs.Selected
	bs.SetSelected(nil)

	// Decay stun: lose 2 stun points per turn (recovery)
	for _, u := range bs.Units {
		if u.Alive && u.StunPoints > 0 {
			u.StunPoints -= 2
			if u.StunPoints < 0 {
				u.StunPoints = 0
			}
		}
	}

	// Bleed tick, morale recovery and panic checks for the new player turn.
	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		if u.BleedRate > 0 {
			u.HP -= u.BleedRate
			if u.HP <= 0 {
				u.HP = 0
				u.Alive = false
			}
		}
		if u.Faction == FactionHuman && u.HP > 0 {
			if u.Morale < 100 {
				u.Morale += 10
				if u.Morale > 100 {
					u.Morale = 100
				}
			}
			if u.HP < u.MaxHP {
				if rand.Intn(100) < 100-2*u.Morale {
					u.Panicked = true
				} else if u.Soldier != nil {
					u.Soldier.AddBraveryExp()
				}
			}
		}
	}

	// Preserve the previously selected human across turns when still valid.
	if oldSelected != nil && oldSelected.Alive && oldSelected.HP > 0 &&
		oldSelected.Faction == 0 && oldSelected.Level == bs.Map.CurrentLevel {
		bs.SetSelected(oldSelected)
	} else {
		for _, u := range bs.Units {
			if u.Faction == FactionHuman && u.Alive && u.HP > 0 && u.Level == bs.Map.CurrentLevel {
				bs.SetSelected(u)
				break
			}
		}
	}
}

func (bs *Battlescape) processAbduction() {
	aliveAliens := 0
	for _, u := range bs.Units {
		if u.Faction == FactionAlien && u.Alive {
			aliveAliens++
		}
	}
	if aliveAliens == 0 {
		return
	}
	// Each turn, aliens abduct 1 civilian if any are alive
	for _, u := range bs.Units {
		if u.Faction == FactionCivilian && u.Alive {
			if rand.Intn(100) < 30+aliveAliens*10 {
				u.Alive = false
				bs.AbductionCivs++
				bs.AddMessage(fmt.Sprintf(language.String("MSG_ABDUCTION_TIMER"), bs.AbductionCivs, bs.AbductionTotal, bs.Turn))
				break
			}
		}
	}
}

func (bs *Battlescape) checkReinforcements() {
	if HasModifier(bs.MissionModifiers, ModReinforcements) && bs.Turn == 4 && !bs.ReinforcementsSpawned {
		bs.ReinforcementsSpawned = true
		spawned := bs.spawnAlienWave(2)
		if spawned > 0 {
			bs.AddMessage(language.Sprintf("BATTLE_REINFORCEMENTS", spawned))
		}
		return
	}

	if bs.UFOName != "Terror" && bs.UFOName != "Alien Base Assault" {
		return
	}
	aliveAliens := 0
	for _, u := range bs.Units {
		if u.Faction == FactionAlien && u.Alive {
			aliveAliens++
		}
	}
	if aliveAliens >= 3 {
		bs.ReinforceTimer = 0
		return
	}
	bs.ReinforceTimer++
	if bs.ReinforceTimer < 4 {
		return
	}
	bs.ReinforceTimer = 0

	count := 1 + rand.Intn(2)
	spawned := bs.spawnAlienWave(count)
	if spawned > 0 {
		bs.AddMessage(fmt.Sprintf(language.String("MSG_REINFORCEMENTS"), spawned))
	}
}

// spawnReinforcementWave creates alien reinforcements and posts a message.
func (bs *Battlescape) spawnReinforcementWave(count int) {
	spawned := bs.spawnAlienWave(count)
	if spawned > 0 {
		bs.AddMessage(language.Sprintf("BATTLE_REINFORCEMENTS", spawned))
	}
}

// learnFromKills scans dead aliens and increases knowledge level for each.
func (bs *Battlescape) learnFromKills() {
	for _, u := range bs.Units {
		if u.Faction != 1 || u.Alive || u.AlienType == nil {
			continue
		}
		bs.Game.LearnAlien(u.AlienType.Name, 2)
	}
}

func (bs *Battlescape) restorePlayerTU() {
	for _, u := range bs.Units {
		if u.Faction == FactionHuman && u.Alive {
			u.HasMoved = false
			if u.Panicked {
				u.Panicked = false
				continue
			}
			u.TU = u.MaxTU
		}
	}
}

func (bs *Battlescape) checkVictory() {
	humans := bs.Units.Faction(0).Alive()
	aliens := bs.Units.Faction(1).Alive()
	civilians := bs.Units.Faction(2).Alive()

	if bs.CustomVictory != nil {
		cv := bs.CustomVictory
		switch cv.Condition {
		case "survive_turns":
			if bs.Turn >= cv.Turns {
				bs.SetPhase(PhaseVictory)
				audio.PlayVictory()
				bs.AddMessage(language.Sprintf("BATTLE_SURVIVED_TURNS", cv.Turns))
			} else if len(humans) == 0 {
				bs.SetPhase(PhaseDefeat)
				audio.PlayDefeat()
				bs.AddMessage(language.String("MSG_MISSION_FAILED"))
			}
			return
		case "reach_point":
			safe := 0
			for _, u := range humans {
				if u.X == cv.TargetX && u.Y == cv.TargetY {
					safe++
				}
			}
			minReq := cv.MinSoldiers
			if minReq <= 0 {
				minReq = 1
			}
			if safe >= minReq {
				bs.SetPhase(PhaseVictory)
				audio.PlayVictory()
				bs.AddMessage(language.Sprintf("MSG_MISSION_EXTRACT", safe))
			} else if len(humans) == 0 {

				bs.SetPhase(PhaseDefeat)
				audio.PlayDefeat()
				bs.AddMessage(language.String("MSG_MISSION_FAILED"))
			}
			return
		default: // "eliminate_all"
			if len(aliens) == 0 {
				bs.SetPhase(PhaseVictory)
				audio.PlayVictory()
				bs.AddMessage(language.String("MSG_MISSION_COMPLETE"))
			} else if len(humans) == 0 {
				bs.SetPhase(PhaseDefeat)
				audio.PlayDefeat()
				bs.AddMessage(language.String("MSG_MISSION_FAILED"))
			}
			return
		}
	}

	// TimeLimit modifier: defeat if turns exceed limit and aliens remain
	if HasModifier(bs.MissionModifiers, ModTimeLimit) && bs.Turn > 15 && len(aliens) > 0 {
		bs.SetPhase(PhaseDefeat)
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_TIME_LIMIT_EXCEEDED"))
		return
	}

	if len(aliens) == 0 {
		bs.SetPhase(PhaseVictory)
		audio.PlayVictory()
		if bs.UFOName == "Terror" {
			totalCiv := 0
			for _, u := range bs.Units {
				if u.Faction == FactionCivilian {
					totalCiv++
				}
			}
			saved := len(civilians)
			bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_COMPLETE_CIV"), saved, totalCiv))
		} else if bs.UFOName == "Abduction" {
			saved := len(civilians)
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ABDUCTION_COMPLETE"), saved))
		} else {
			bs.AddMessage(language.String("MSG_MISSION_COMPLETE"))
		}
	} else if len(humans) == 0 {
		bs.SetPhase(PhaseDefeat)
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_MISSION_FAILED"))
	} else if bs.UFOName == "Terror" && len(civilians) == 0 && len(aliens) > 0 {
		bs.SetPhase(PhaseDefeat)
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_MISSION_FAILED_CIV"))
	} else if bs.UFOName == "Abduction" && bs.AbductionCivs >= bs.AbductionTotal {
		bs.SetPhase(PhaseDefeat)
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_MISSION_FAILED_CIV"))
	}
}

func (bs *Battlescape) MoveCursor(dx, dy int) {
	bs.CursorX += dx
	bs.CursorY += dy
	if bs.CursorX < 0 {
		bs.CursorX = 0
	}
	if bs.CursorY < 0 {
		bs.CursorY = 0
	}
	if bs.CursorX >= bs.Map.Width {
		bs.CursorX = bs.Map.Width - 1
	}
	if bs.CursorY >= bs.Map.Height {
		bs.CursorY = bs.Map.Height - 1
	}

	bs.Camera.SetTarget(bs.CursorX, bs.CursorY)
	bs.updateHoverFromCursor()
}

// updateHoverFromCursor sets HoveredUnit to the alien under the cursor (keyboard or mouse).
func (bs *Battlescape) updateHoverFromCursor() {
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == FactionAlien && unit.Alive {
		bs.HoveredUnit = unit
	} else if unit == nil || unit.Faction != FactionAlien {
		bs.HoveredUnit = nil
	}
}

// SelectUnit selects the unit at the cursor position or cycles to the next one.
func (bs *Battlescape) SelectUnit() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		audio.PlaySelect()
		bs.SetSelected(unit)
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
	} else {
		bs.cycleUnit(1)
	}
}

// LeftClick handles a left-click on the battlescape, selecting units or interacting with the UI.
func (bs *Battlescape) LeftClick() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.State.CursorState == StateMovePlan {
		bs.State.CursorState = StateInspect
		bs.State.MovePath = nil
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		bs.SetSelected(unit)
		bs.State.CursorState = StateInspect
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
		return
	}
	bs.State.CursorState = StateInspect
}

// RightClick handles a right-click: in targeting mode fires, in move mode confirms movement,
// otherwise enters move-planning mode.
func (bs *Battlescape) RightClick() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.State.CursorState == StateTargeting {
		bs.FireWeapon()
		return
	}
	if bs.State.CursorState == StateMovePlan {
		bs.MoveSelected()
		bs.State.CursorState = StateInspect
		bs.State.MovePath = nil
		return
	}
	bs.State.CursorState = StateMovePlan
	bs.updateMovePath()
}

// MoveSelected moves the selected unit along the calculated path toward the cursor.
func (bs *Battlescape) MoveSelected() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.X == bs.CursorX && bs.Selected.Y == bs.CursorY {
		return
	}
	path := bs.CalculatePath(bs.Selected.X, bs.Selected.Y, bs.CursorX, bs.CursorY)
	if len(path) < 2 {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		return
	}
	u := bs.Selected
	crouchExtra := 0
	if u.Crouching {
		crouchExtra = 4
	}
	// Walk along the path, consuming TU, until we can't afford the next step.
	best := 0
	totalCost := 0
	for i := 1; i < len(path); i++ {
		tile := bs.Map.At(path[i][0], path[i][1])
		stepCost := 4
		if tile.Type == TileTree || tile.Type == TileRock || tile.Type == TileWater {
			stepCost = 8
		}
		totalCost += stepCost
		if totalCost+crouchExtra <= u.TU {
			best = i
		} else {
			break
		}
	}
	if best == 0 {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		return
	}
	totalCost = 0
	for i := 1; i <= best; i++ {
		tile := bs.Map.At(path[i][0], path[i][1])
		stepCost := 4
		if tile.Type == TileTree || tile.Type == TileRock || tile.Type == TileWater {
			stepCost = 8
		}
		totalCost += stepCost
	}
	dest := path[best]
	u.X, u.Y = dest[0], dest[1]
	u.TU -= totalCost + crouchExtra
	audio.PlayMove()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), u.Soldier.Name, u.X, u.Y))
	bs.ComputeFOVForTeam()
	bs.checkAlienReactionFire(u)
}

// FireWeapon fires the selected unit's weapon at the target under the cursor.
func (bs *Battlescape) FireWeapon() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	target := bs.Units.At(bs.CursorX, bs.CursorY)
	if target == nil || target.Faction == 0 {
		bs.AddMessage(language.String("MSG_NO_TARGET"))
		return
	}
	if !bs.Selected.CanSee(target.X, target.Y, bs.Map) {
		bs.AddMessage(language.String("MSG_TARGET_NO_LOS"))
		return
	}
	damage, hit, coverHit, err := bs.Selected.FireAt(target, bs.Map, &bs.Weather)
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "not enough TU"):
			msg = language.String("MSG_NOT_ENOUGH_TU")
		case strings.Contains(msg, "out of ammo"):
			msg = language.String("MSG_OUT_OF_AMMO")
		default:
			msg = language.String("MSG_WEAPON_ERROR")
		}
		bs.AddMessage(msg)
		return
	}
	bs.recordPlayerShot(bs.Selected, target)
	audio.PlayWeaponFire(bs.Selected.Weapon)
	engine.SpawnMuzzleFlash(bs.Particles, bs.Selected.X-bs.ScrollX+1, bs.Selected.Y-bs.ScrollY+1)
	if hit {
		audio.PlayHit()
		engine.SpawnExplosion(bs.Particles, target.X-bs.ScrollX+1, target.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
		if engine.Config.ScreenShake {
			bs.Camera.TriggerShake(0.5)
		}
		bs.SpawnBloodSplatter(target)
		w := data.RuleItems[bs.Selected.Weapon]
		if w.Type == "plasma" || w.Type == "explosive" {
			bs.SpawnFire(target.X, target.Y, 3)
		}
		name := "alien"
		if target.AlienType != nil {
			name = target.AlienType.LangName()
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_HIT_TARGET"), damage, name, target.HP))
		bs.spawnFloater(target.X, target.Y, fmt.Sprintf("-%d", damage), color.XTerm9)
	} else {
		audio.PlayMiss()
		if coverHit {
			bs.AddMessage(language.String("MSG_HIT_COVER"))
			bs.spawnFloater(target.X, target.Y, language.String("FLOATER_COVER"), color.XTerm8)
		} else {
			bs.AddMessage(language.String("MSG_MISSED"))
			bs.spawnFloater(target.X, target.Y, language.String("FLOATER_MISS"), color.XTerm8)
		}
	}
	bs.PlayerLock = bs.Game.ActionDelay / 2
}

func (bs *Battlescape) recordPlayerShot(shooter, target *Unit) {
	if target == nil || target.Faction != 1 || shooter == nil {
		return
	}
	dx := shooter.X - target.X
	dy := shooter.Y - target.Y
	dist := math.Sqrt(float64(dx*dx + dy*dy))
	bs.PlayerShotDistSum += dist
	bs.PlayerShotCount++
	// Flank/ambush: shooter fires from outside the target alien's line of sight
	if !target.CanSee(shooter.X, shooter.Y, bs.Map) {
		bs.PlayerFlankShots++
	}
}

// Reload reloads the selected unit's weapon to full ammo.
func (bs *Battlescape) Reload() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 8 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_RELOAD"))
		return
	}
	w := data.RuleItems[bs.Selected.Weapon]
	if w.AmmoMax >= 99 {
		bs.AddMessage(language.String("MSG_ENERGY_WEAPON"))
		return
	}
	if bs.Selected.WeaponAmmo >= w.AmmoMax {
		bs.AddMessage(language.String("MSG_WEAPON_LOADED"))
		return
	}
	bs.Selected.TU -= 8
	bs.Selected.WeaponAmmo = w.AmmoMax
	audio.PlayReload()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_RELOADED"), w.DisplayName(), bs.Selected.WeaponAmmo, w.AmmoMax))
}

// CycleFireMode cycles the selected unit's weapon between available fire modes.
func (bs *Battlescape) CycleFireMode() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	w, ok := data.RuleItems[bs.Selected.Weapon]
	if !ok {
		return
	}
	modes := w.Modes()
	if len(modes) <= 1 {
		return
	}
	cur := -1
	for i, m := range modes {
		if m == bs.Selected.FireMode {
			cur = i
			break
		}
	}
	next := (cur + 1) % len(modes)
	bs.Selected.FireMode = modes[next]
	bs.AddMessage(fmt.Sprintf(language.String("MSG_FIRE_MODE"), modes[next].String()))
}

func (bs *Battlescape) EndTurn() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	audio.PlayAlienTurn()
	bs.SetPhase(PhaseAlienTurn)
	bs.Status = StatusAlienTurn
	bs.AddMessage(language.String("MSG_ALIEN_TURN"))

	bs.AlienTurnQueue = nil
	bs.AlienTurnIdx = 0

	bs.AlienSquadPlan = bs.planSquadActions()

	if bs.AlienMemory == nil {
		bs.AlienMemory = NewSquadMemory()
	}
	bs.AlienMemory.turn = bs.Turn
	for _, h := range bs.Units.Faction(0) {
		if !h.Alive {
			bs.AlienMemory.Forget(h)
		}
	}

	for _, ai := range bs.AlienAIs {
		if !ai.Unit.Alive {
			continue
		}
		if ai.Unit.Panicked {
			ai.Unit.Panicked = false
			continue
		}
		ai.Unit.TU = ai.Unit.MaxTU
		ai.Unit.HasMoved = false
		ai.Memory = bs.AlienMemory
		humanUnits := bs.Units.Faction(0)
		actions := ai.Update(bs.Units, bs.Map, humanUnits, bs.AlienSquadPlan, &bs.Game.Tactics)
		bs.AlienTurnQueue = append(bs.AlienTurnQueue, actions...)
	}

	if len(bs.AlienTurnQueue) == 0 {
		bs.finishAlienTurn()
	} else {
		bs.ActionDelay = bs.Game.ActionDelay / 3 // Quicker delay for sub-actions
	}
}

func (bs *Battlescape) planSquadActions() *SquadPlan {
	humanUnits := bs.Units.Faction(0)
	aliens := bs.Units.Faction(1)

	var aliveAliens []*Unit
	for _, u := range aliens {
		if u.Alive {
			aliveAliens = append(aliveAliens, u)
		}
	}
	if len(aliveAliens) == 0 {
		return nil
	}

	var primary *Unit
	var secondary *Unit
	bestScore := -999.0

	for _, h := range humanUnits {
		if !h.Alive {
			continue
		}
		score := 0.0
		if h.HP < h.MaxHP/2 {
			score += 8
		}
		switch h.Weapon {
		case "rocket", "heavy_plasma", "laser_rifle":
			score += 6
		case "rifle", "heavy":
			score += 3
		}
		if h.Crouching {
			score -= 4
		}
		if !h.Crouching {
			score += 2
		}
		for _, a := range aliveAliens {
			dx := float64(h.X - a.X)
			dy := float64(h.Y - a.Y)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 6 {
				score += 2
			}
			if a.CanSee(h.X, h.Y, bs.Map) {
				score += 3
			}
		}
		if score > bestScore {
			secondary = primary
			bestScore = score
			primary = h
		}
	}

	roles := make(map[*Unit]SquadRole)
	suppressorCount := len(aliveAliens) / 3
	if suppressorCount < 1 {
		suppressorCount = 0
	}
	flankerCount := len(aliveAliens) / 3

	// Adapt squad composition to observed player tactics
	t := bs.Game.Tactics
	bc := t.BattleCount
	if bc < 1 {
		bc = 1
	}
	avgGrenades := float64(t.GrenadeUsage) / float64(bc)
	avgRange := t.AverageRange
	avgFlank := float64(t.FlankingObserved) / float64(bc)
	flankHeavy := avgFlank >= 1.0
	longRange := avgRange >= 8.0
	grenadeHeavy := avgGrenades >= 1.5
	if flankHeavy {
		// Player flanks often: post more suppressors to pin them down
		suppressorCount = len(aliveAliens) / 2
	}
	if longRange {
		// Player snipes from afar: send more flankers to close distance
		flankerCount = len(aliveAliens) / 2
	}
	if grenadeHeavy {
		// Player grenades clustered aliens: keep units dispersed, fewer strict roles
		suppressorCount = len(aliveAliens) / 4
	}

	// Pre-classify aliens by suitability so roles can be reassigned dynamically.
	type candidate struct {
		u        *Unit
		brave    int
		aggro    int
		hasLOS   bool
		angle    int // approach angle quadrant toward primary target
		role     SquadRole
	}
	var cands []candidate
	for _, u := range aliveAliens {
		brave := 100
		if u.AlienType != nil {
			brave = u.AlienType.Bravery
		}
		aggro := 5
		if u.AlienType != nil {
			aggro = u.AlienType.Aggression
		}
		hasLOS := primary != nil && u.CanSee(primary.X, primary.Y, bs.Map)
		angle := -1
		if primary != nil {
			dx := u.X - primary.X
			dy := u.Y - primary.Y
			// Quantize direction into 4 quadrants (N/E/S/W) for flank spread.
			if dx >= 0 && dy < dx {
				angle = 0
			} else if dy >= 0 && dx <= dy {
				angle = 1
			} else if dx < 0 && dy >= dx {
				angle = 2
			} else {
				angle = 3
			}
		}
		cands = append(cands, candidate{u: u, brave: brave, aggro: aggro, hasLOS: hasLOS, angle: angle, role: RoleNormal})
	}

	suppressors := 0
	flankers := 0
	usedAngles := map[int]bool{}

	// Pass 1: assign suppressors first (units with LOS make the best pinners).
	for i := range cands {
		if suppressors >= suppressorCount {
			break
		}
		c := &cands[i]
		if c.role != RoleNormal || !c.hasLOS || c.brave <= 60 || c.u.TU <= 20 {
			continue
		}
		c.role = RoleSuppressor
		suppressors++
	}

	// Pass 2: assign flankers, spreading them across distinct approach angles so
	// they envelop the target instead of bunching up on one side.
	for i := range cands {
		if flankers >= flankerCount {
			break
		}
		c := &cands[i]
		if c.role != RoleNormal || c.aggro <= 6 || c.u.TU <= 30 {
			continue
		}
		if c.angle >= 0 && usedAngles[c.angle] {
			continue
		}
		c.role = RoleFlanker
		flankers++
		if c.angle >= 0 {
			usedAngles[c.angle] = true
		}
	}

	// Pass 3: dynamic reassignment. If a role quota is unfilled (e.g. no LOS
	// unit was available for suppressor), promote a capable RoleNormal unit so
	// the squad still fields that role. Prefer units that are healthy and mobile.
	for i := range cands {
		if suppressors >= suppressorCount {
			break
		}
		c := &cands[i]
		if c.role != RoleNormal || c.brave <= 60 || c.u.TU <= 20 {
			continue
		}
		c.role = RoleSuppressor
		suppressors++
	}
	for i := range cands {
		if flankers >= flankerCount {
			break
		}
		c := &cands[i]
		if c.role != RoleNormal || c.aggro <= 6 || c.u.TU <= 30 {
			continue
		}
		c.role = RoleFlanker
		flankers++
	}

	for _, c := range cands {
		roles[c.u] = c.role
	}

	retreat := false
	allyCount := len(aliveAliens)
	enemyCount := 0
	for _, h := range humanUnits {
		if h.Alive {
			enemyCount++
		}
	}
	if allyCount > 0 && enemyCount > 0 && allyCount < enemyCount/2 {
		retreat = true
	}

	return &SquadPlan{
		PrimaryTarget:   primary,
		SecondaryTarget: secondary,
		Roles:           roles,
		Retreat:         retreat,
	}
}

func (bs *Battlescape) Crouch() {
	if bs.Selected == nil {
		return
	}
	if bs.Selected.Crouching {
		bs.Selected.Crouching = false
		bs.AddMessage(language.String("MSG_STANDING_UP"))
	} else if bs.Selected.TU >= 4 {
		bs.Selected.Crouching = true
		bs.Selected.TU -= 4
		bs.AddMessage(language.String("MSG_CROUCHING"))
	}
}

func (bs *Battlescape) UseStairs() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Map.NumLevels <= 1 {
		bs.AddMessage(language.String("MSG_NO_STAIRS_MAP"))
		return
	}
	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	if tile.Type != TileStairs && tile.Type != TileStairsDown {
		// Check if selected unit is on stairs
		if bs.Selected != nil {
			tile = bs.Map.At(bs.Selected.X, bs.Selected.Y)
		}
		if tile.Type != TileStairs && tile.Type != TileStairsDown {
			bs.AddMessage(language.String("MSG_MOVE_TO_STAIRS"))
			return
		}
	}

	oldLevel := bs.Map.CurrentLevel
	if oldLevel == 0 && bs.Map.NumLevels > 1 {
		bs.Map.CurrentLevel = 1
	} else if oldLevel > 0 {
		bs.Map.CurrentLevel = 0
	} else {
		bs.AddMessage(language.String("MSG_NO_STAIRS_HERE"))
		return
	}

	// Teleport selected unit to stairs on new level
	if bs.Selected != nil && bs.Selected.TU >= 8 {
		bs.Selected.TU -= 8
		bs.Selected.Level = bs.Map.CurrentLevel
		bs.ComputeFOVForTeam()
		bs.AddMessage(language.Sprintf("MSG_DESCENDED_LEVEL", bs.Map.CurrentLevel+1))
	} else if bs.Selected != nil {
		bs.Map.CurrentLevel = oldLevel
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_STAIRS"))
	} else {
		bs.Map.CurrentLevel = oldLevel
		bs.AddMessage(language.String("MSG_NO_SOLDIER_SELECTED"))
	}
}

func (bs *Battlescape) Grenade() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < GrenadeTUCost {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_GRENADE"))
		return
	}

	grenadeRange := GrenadeRange
	damage := GrenadeBaseDamage + bs.Selected.Strength*GrenadeStrMult
	if bs.Selected.Soldier != nil && bs.Selected.Soldier.HasBattleMod(soldier.BModDemolitions) {
		damage = damage * 3 / 2
	}
	ax := bs.CursorX
	ay := bs.CursorY
	dx := ax - bs.Selected.X
	dy := ay - bs.Selected.Y
	dist := dx*dx + dy*dy
	if dist > grenadeRange*grenadeRange {
		bs.AddMessage(language.String("MSG_GRENADE_OUT_OF_RANGE"))
		return
	}

	bs.PlayerGrenadeCount++
	bs.Selected.TU -= GrenadeTUCost

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		udx := u.X - ax
		udy := u.Y - ay
		udist := udx*udx + udy*udy
		if udist <= GrenadeSplashSq {
			splashDmg := damage - udist*GrenadeSplashFalloff
			if splashDmg < GrenadeMinSplash {
				splashDmg = 5
			}
			u.HP -= splashDmg
			if u.HP <= 0 {
				u.HP = 0
				u.Alive = false
			}
			bs.SpawnBloodSplatter(u)
		}
	}

	splashRadius := 2
	if bs.Selected.Soldier != nil && bs.Selected.Soldier.HasBattleMod(soldier.BModGrenadier) {
		splashRadius += 2
	}
	for dy := -splashRadius; dy <= splashRadius; dy++ {
		for dx := -splashRadius; dx <= splashRadius; dx++ {
			if dx*dx+dy*dy > splashRadius*splashRadius {
				continue
			}
			tx, ty := ax+dx, ay+dy
			if bs.Map.IsDestructible(tx, ty) {
				if bs.Map.DestroyWall(tx, ty) {
					SpawnRubble(bs.Particles, tx-bs.ScrollX+1, ty-bs.ScrollY+1)
				}
			}
			if tx >= 0 && tx < bs.Map.Width && ty >= 0 && ty < bs.Map.Height {
				tile := bs.Map.At(tx, ty)
				if tile.IsFlammable() && tile.Fire <= 0 && rand.Intn(3) == 0 {
					bs.SpawnFire(tx, ty, 3)
				}
			}
		}
	}

	bs.Gas.Set(ax, ay, 3, GasSmoke)
	for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		bs.Gas.Set(ax+d[0], ay+d[1], 2, GasSmoke)
	}

	audio.PlayGrenade()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_GRENADE_DETONATED"), ax, ay))

	if engine.Config.ScreenShake {
		bs.Camera.TriggerShake(3.0)
	}
	engine.SpawnExplosion(bs.Particles, ax-bs.ScrollX+1, ay-bs.ScrollY+1, tcell.NewRGBColor(255, 180, 50), 24)
	engine.SpawnSmoke(bs.Particles, ax-bs.ScrollX+1, ay-bs.ScrollY+1, 8)
}

func (bs *Battlescape) UseMedikit() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 25 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_MEDIKIT"))
		return
	}

	healAmount := 10
	if bs.Selected.Soldier != nil && bs.Selected.Soldier.HasBattleMod(soldier.BModFieldMedic) {
		healAmount = 15
	}
	mx := bs.CursorX
	my := bs.CursorY

	target := bs.Units.At(mx, my)
	if target == nil || target.Faction != 0 {
		bs.AddMessage(language.String("MSG_SELECT_FRIENDLY"))
		return
	}
	if target.HP >= target.MaxHP {
		bs.AddMessage(language.String("MSG_ALREADY_FULL_HP"))
		return
	}

	bs.Selected.TU -= 25
	target.HP += healAmount
	bs.spawnFloater(target.X, target.Y, fmt.Sprintf("+%d", healAmount), color.XTerm2)
	if target.HP > target.MaxHP {
		target.HP = target.MaxHP
	}
	name := "ally"
	if target.Soldier != nil {
		name = target.Soldier.Name
	}
	audio.PlayMedikit()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_HEALED"), name, healAmount, target.HP, target.MaxHP))
}

func SpawnRubble(ps *engine.ParticleSystem, x, y int) {
	rubbleChars := []rune{'.', '*', ',', '\''}
	styles := []tcell.Style{
		tcell.StyleDefault.Foreground(tcell.NewRGBColor(120, 100, 80)),
		tcell.StyleDefault.Foreground(tcell.NewRGBColor(90, 80, 70)),
		tcell.StyleDefault.Foreground(tcell.NewRGBColor(150, 130, 100)),
		tcell.StyleDefault.Foreground(tcell.NewRGBColor(100, 90, 75)),
	}
	for i := 0; i < 4; i++ {
		vx := (rand.Float64() - 0.5) * 3
		vy := -2 - rand.Float64()*3
		ps.Spawn(
			float64(x)+rand.Float64()*0.4-0.2,
			float64(y)+rand.Float64()*0.4-0.2,
			vx, vy,
			rubbleChars[i],
			styles[i],
			0.8+rand.Float64()*0.4,
			0.6,
		)
	}
}

func (bs *Battlescape) UseMotionScanner() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.Soldier == nil || bs.Selected.Soldier.Weapon != "motion_scanner" {
		bs.AddMessage(language.String("MSG_NEED_SCANNER"))
		return
	}
	if bs.Selected.TU < ScannerTUCost {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_SCANNER"))
		return
	}
	bs.Selected.TU -= ScannerTUCost
	scanRange := ScannerRange
	bs.scannerPings = nil

	for _, u := range bs.Units {
		if u.Faction == FactionAlien && u.Alive {
			dx := u.X - bs.Selected.X
			dy := u.Y - bs.Selected.Y
			dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
			if dist <= scanRange {
				bs.scannerPings = append(bs.scannerPings, [2]int{u.X, u.Y})
			}
		}
	}
	audio.PlayClick()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_SCANNER_RESULT"), len(bs.scannerPings)))
}

func (bs *Battlescape) PlaceMine() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.Soldier == nil || bs.Selected.Soldier.Weapon != "proximity_mine" {
		bs.AddMessage(language.String("MSG_NEED_MINE"))
		return
	}
	if bs.Selected.TU < 20 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_MINE"))
		return
	}
	mx := bs.CursorX
	my := bs.CursorY
	dx := mx - bs.Selected.X
	dy := my - bs.Selected.Y
	dist := dx*dx + dy*dy
	if dist > 1 {
		bs.AddMessage(language.String("MSG_MINE_TOO_FAR"))
		return
	}
	if mx < 0 || mx >= bs.Map.Width || my < 0 || my >= bs.Map.Height || bs.Map.At(mx, my).Type != TileFloor {
		bs.AddMessage(language.String("MSG_MINE_BAD_TILE"))
		return
	}
	for _, m := range bs.Mines {
		if m.X == mx && m.Y == my {
			bs.AddMessage(language.String("MSG_MINE_EXISTS"))
			return
		}
	}

	bs.Selected.TU -= 20
	bs.Mines = append(bs.Mines, PlacedMine{X: mx, Y: my})
	bs.AddMessage(fmt.Sprintf(language.String("MSG_MINE_PLACED"), mx, my))
	audio.PlayClick()
}

// checkMineTriggers checks if any alien unit triggered a mine.
func (bs *Battlescape) checkMineTriggers() {
	for _, u := range bs.Units {
		if u.Faction != 1 || !u.Alive {
			continue
		}
		for mi := 0; mi < len(bs.Mines); mi++ {
			m := bs.Mines[mi]
			dx := u.X - m.X
			dy := u.Y - m.Y
			if dx*dx+dy*dy <= 1 {
				// Detonate mine
				damage := MineBaseDamage + rand.Intn(MineDamageBonus)
				u.HP -= damage
				if u.HP <= 0 {
					u.HP = 0
					u.Alive = false
					bs.AddMessage(fmt.Sprintf(language.String("MSG_MINE_KILL"), u.Name(), m.X, m.Y))
				} else {
					bs.AddMessage(fmt.Sprintf(language.String("MSG_MINE_HIT"), u.Name(), damage, m.X, m.Y))
				}
				audio.PlayExplosion()
				engine.SpawnExplosion(bs.Particles, m.X-bs.ScrollX+1, m.Y-bs.ScrollY+1,
					tcell.NewRGBColor(255, 180, 50), 20)
				if engine.Config.ScreenShake {
					bs.Camera.TriggerShake(3.0)
				}
				bs.spawnFloater(m.X, m.Y, fmt.Sprintf("%d!", damage), tcell.NewRGBColor(255, 80, 0))
				// Remove the triggered mine
				bs.Mines = append(bs.Mines[:mi], bs.Mines[mi+1:]...)
				mi--
			}
		}
	}
}

func (bs *Battlescape) PsiAttack() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.Soldier == nil || bs.Selected.Soldier.Weapon != "psi_amp" {
		bs.AddMessage(language.String("MSG_NEED_PSI_AMP"))
		return
	}

	if bs.Selected.TU < 20 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_PSI"))
		return
	}

	target := bs.Units.At(bs.CursorX, bs.CursorY)
	if target == nil || target.Faction != 1 || !target.Alive {
		bs.AddMessage(language.String("MSG_SELECT_ALIEN_TARGET"))
		return
	}

	bs.Selected.TU -= 20

	targetPsi := 0
	if target.AlienType != nil {
		targetPsi = target.AlienType.Psi
	} else if target.Soldier != nil {
		targetPsi = target.Soldier.PsiStr
	}

	attackerSkill := bs.Selected.Soldier.PsiSkill
	if attackerSkill < 1 {
		bs.AddMessage(language.String("MSG_NO_PSI_TRAINING"))
		return
	}

	successChance := attackerSkill - targetPsi/3
	if successChance < 5 {
		successChance = 5
	}
	success := rand.Intn(100) < successChance

	if success {
		if engine.Config.ScreenShake {
			bs.Camera.TriggerShake(0.3)
		}
		engine.SpawnExplosion(bs.Particles, target.X-bs.ScrollX+1, target.Y-bs.ScrollY+1, tcell.NewRGBColor(120, 0, 200), 12)
		target.TU = 0
		target.Panicked = true
		if bs.Selected.Soldier != nil {
			bs.Selected.Soldier.AddPsiSkillExp()
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_PSI_SUCCESS"), target.Name()))
	} else {
		bs.AddMessage(language.String("MSG_PSI_FAIL"))
	}
}

func (bs *Battlescape) DrawCombatStatusBar(ctx *engine.ScreenCtx, w int) {
	label := ""
	style := tcell.StyleDefault
	switch bs.Status {
	case StatusPlayerTurn:
		label = language.String("STATUS_PLAYER_TURN")
		style = engine.StyleBlue.Bold(true)
	case StatusAlienTurn:
		label = language.String("STATUS_ALIEN_TURN")
		style = engine.StyleRed.Bold(true)
	case StatusPlayerOverwatch:
		label = language.String("STATUS_PLAYER_OVERWATCH")
		style = engine.StyleCyan.Bold(true)
	case StatusAlienOverwatch:
		label = language.String("STATUS_ALIEN_OVERWATCH")
		style = engine.StyleOrange.Bold(true)
	}

	for x := 0; x < w; x++ {
		ctx.SetCell(x, 0, ' ', style)
	}

	text := " " + label + " "
	tw := engine.StringWidth(text)
	if tw > w {
		text = text[:w]
		tw = engine.StringWidth(text)
	}
	start := (w - tw) / 2
	if start < 0 {
		start = 0
	}
	ctx.DrawString(start, 0, text, style)
}

func (bs *Battlescape) Render(ctx *engine.ScreenCtx) {
	// 1. Calculate view dimensions and layout
	w, h := ctx.Size()
	engine.Layout.UpdateMode(w, h)
	bs.SidebarW = engine.Layout.BattleSidebarWidth(w)
	viewW := engine.Layout.BattleViewWidth(w)
	viewH := engine.Layout.BattleViewHeight(h)

	// 2. Top-level UI elements
	bs.DrawCombatStatusBar(ctx, w)

	// 3. Camera and scrolling
	camX, camY := bs.Camera.Pos()
	bs.SetScroll(camX-viewW/2, camY-viewH/2)

	if bs.FrameCount%12 == 0 && bs.Phase != PhaseVictory && bs.Phase != PhaseDefeat {
		switch bs.UFOName {
		case "Polar":
			engine.SpawnSnow(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, viewH)
		case "Desert":
			engine.SpawnDust(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, viewH)
		case "Cydonia", "Alien Base Assault":
			engine.SpawnEmbers(bs.Particles, bs.ScrollX, bs.ScrollY, viewW, viewH)
		}
	}

	blackStyle := engine.StyleDefault

	for y := 0; y < viewH; y++ {
		for x := 0; x < viewW; x++ {
			mx := x + bs.ScrollX
			my := y + bs.ScrollY

			if mx < 0 || mx >= bs.Map.Width || my < 0 || my >= bs.Map.Height {
				ctx.SetCell(x+1, y+1, ' ', blackStyle)
				continue
			}

			tile := bs.Map.At(mx, my)

			ctx2 := bs.Map.neighbourhood(mx, my)
			ch, style := RenderTile(tile, ctx2, tile.Visible, tile.Seen, bs.FrameCount, mx, my)

			if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
				movementRange := bs.GetMovementRange()
				if movementRange[[2]int{mx, my}] {
					if bs.IsNight {
						style = style.Background(tcell.NewRGBColor(10, 25, 50))
					} else {
						style = style.Background(tcell.NewRGBColor(20, 40, 80))
					}
				}
			}

			style = bs.ApplyCursorStyles(mx, my, style)
			ctx.SetCell(x+1, y+1, ch, style)
		}
	}

	if engine.Config.GridLines {
		gridStyle := engine.StyleDefault.Foreground(color.Gray)
		for y := 0; y < viewH; y++ {
			for x := 0; x < viewW; x++ {
				mx := x + bs.ScrollX
				my := y + bs.ScrollY
				if mx%4 == 0 && my%4 == 0 {
					ctx.SetCell(x+1, y+1, '·', gridStyle)
				}
			}
		}
	}

	bs.Gas.Draw(ctx, bs.ScrollX, bs.ScrollY, viewW, viewH)

	if bs.IsNight {
		for _, u := range bs.Units {
			if !u.Alive || u.Faction != 0 || u.Level != bs.Map.CurrentLevel {
				continue
			}
			sx := u.X - bs.ScrollX + 1
			sy := u.Y - bs.ScrollY + 1
			if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
				engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 3, tcell.NewRGBColor(120, 110, 70))
			}
		}
		for _, u := range bs.Units {
			if !u.Alive || u.Faction != 1 || u.Level != bs.Map.CurrentLevel {
				continue
			}
			sx := u.X - bs.ScrollX + 1
			sy := u.Y - bs.ScrollY + 1
			if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
				if bs.Map.IsSeen(u.X, u.Y) {
				engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 1, tcell.NewRGBColor(60, 80, 150))
				}
			}
		}
	}

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		if u.Level != bs.Map.CurrentLevel {
			continue
		}
		sx := u.X - bs.ScrollX + 1
		sy := u.Y - bs.ScrollY + 1
		if sx < 1 || sx >= viewW+1 || sy < 1 || sy >= viewH+1 {
			continue
		}
		if u.Faction == FactionAlien && !bs.Map.IsVisible(u.X, u.Y) {
			continue
		}
		ch := '\u127E' // human
		style := engine.StyleCyanBold
		if u.Faction == FactionAlien {
			ch = '\u03A9' // Ω alien (default)
			if u.AlienType != nil {
				ch = u.AlienType.Icon
			}
			style = engine.StyleRedBold
			if u.AlienType != nil {
				style = u.AlienType.Style
				// A transparent/default foreground renders black on most
				// terminals; fall back to the red alien style so the unit
				// stays visible.
				if style.GetForeground() == tcell.ColorDefault {
					style = engine.StyleRedBold
				}
			}
			if engine.Config.BloomEnabled {
				bloomColor := tcell.NewRGBColor(255, 50, 50)
				if u.AlienType != nil && u.AlienType.FgColor != tcell.ColorDefault {
					r32, g32, b32 := u.AlienType.FgColor.RGB()
					if r32 >= 0 && g32 >= 0 && b32 >= 0 {
						bloomColor = tcell.NewRGBColor(r32, g32, b32)
					}
				}
				engine.ApplyBloom(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, bloomColor)
			}
		} else if u.Faction == FactionCivilian {
			ch = '\u1276' // civilian
			style = engine.StyleGreen
		}
		if u == bs.Selected {
			style = style.Reverse(true)
			if engine.Config.LightingEnabled {
				engine.ApplyDirectionalLight(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 0, -1, 5, tcell.NewRGBColor(200, 200, 150), func(x, y int) bool {
					return !bs.Map.Opaque(x+bs.ScrollX-1, y+bs.ScrollY-1)
				})
			}
		}
		if u.X == bs.CursorX && u.Y == bs.CursorY {
			style = style.Reverse(true)
		}
		ctx.SetCell(sx, sy, ch, style)

		if u == bs.Selected || u == bs.HoveredUnit {
			// Ground the sprite with a dim selection shadow.
			if sy+1 <= viewH {
				ctx.SetCell(sx, sy+1, '▄', engine.StyleGray)
			}
			// Compact 3-cell HP pip bar on the tile above (green→yellow→red).
			if sy-1 >= 1 {
				ratio := float64(u.HP) / float64(u.MaxHP)
				filled := int(ratio*3 + 0.5)
				if filled > 3 {
					filled = 3
				}
				if filled < 0 {
					filled = 0
				}
				for i := 0; i < 3; i++ {
					var pc tcell.Color
					switch {
					case i >= filled:
						pc = color.XTerm8
					case ratio < 0.5:
						pc = color.XTerm9
					case ratio < 0.75:
						pc = color.XTerm3
					default:
						pc = color.XTerm2
					}
					ctx.SetCell(sx-1+i, sy-1, '▬', engine.StyleDefault.Background(color.Black).Foreground(pc))
				}
			}
		}
	}

	// Draw projectile
	if bs.Projectile != nil {
		p := bs.Projectile
		progress := float64(p.Progress) / float64(p.Length)
		px := float64(p.FromX) + float64(p.ToX-p.FromX)*progress
		py := float64(p.FromY) + float64(p.ToY-p.FromY)*progress
		sx := int(px) - bs.ScrollX + 1
		sy := int(py) - bs.ScrollY + 1
		if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
			ctx.SetCell(sx, sy, p.Symbol, p.Style)
			if engine.Config.BloomEnabled {
				engine.ApplyBloom(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, p.Style.GetForeground())
			}
		}
	}

	bs.Particles.Draw(ctx.ScreenRaw)

	if bs.VisionMode != engine.VisionNormal {
		var entities []engine.ThermalEntity
		if bs.VisionMode == engine.VisionThermal {
			for _, u := range bs.Units {
				if u.Alive && u.Level == bs.Map.CurrentLevel {
					sx := u.X - bs.ScrollX + 1
					sy := u.Y - bs.ScrollY + 1
					if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
						entities = append(entities, engine.ThermalEntity{X: sx, Y: sy})
					}
				}
			}
		}
		engine.ApplyVisionFilter(ctx.ScreenRaw, bs.VisionMode, entities)
	}

	// Draw placed mines
	mineStyle := engine.StyleDefault.Foreground(tcell.NewRGBColor(255, 80, 80)).Bold(true)
	for _, m := range bs.Mines {
		sx := m.X - bs.ScrollX + 1
		sy := m.Y - bs.ScrollY + 1
		if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
			ctx.SetCell(sx, sy, '◉', mineStyle)
		}
	}

	// Draw motion scanner pings
	pingStyle := engine.StyleDefault.Foreground(tcell.NewRGBColor(0, 255, 100)).Bold(true)
	for _, p := range bs.scannerPings {
		sx := p[0] - bs.ScrollX + 1
		sy := p[1] - bs.ScrollY + 1
		if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
			blink := (bs.FrameCount / 6) % 2
			if blink == 0 {
				ctx.SetCell(sx, sy, '⚡', pingStyle)
			}
		}
	}

	bs.drawFloaters(ctx)

	bs.drawSidebar(ctx, viewW, viewH, w, h)

	ctx.DrawPanel(0, h-4, w, 3, language.String("BATTLESCAPE"), engine.StyleDefault)
	lightStr := language.String("LIGHT_DAY")
	if bs.IsNight {
		lightStr = language.String("LIGHT_NIGHT")
	}
	turnStr := fmt.Sprintf(language.String("STATUS_TURN"), bs.Turn, bs.phaseStr()+" ("+lightStr+")")
	if bs.Map.NumLevels > 1 {
		turnStr += language.Sprintf("BATTLE_LEVEL_FMT", bs.Map.CurrentLevel+1)
	}
	ctx.DrawString(2, h-3, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf(language.String("STATUS_SELECTED"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-3, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	tileName := tileTypeName(tile.Type)
	cursorStr := fmt.Sprintf(language.String("STATUS_CURSOR"), bs.CursorX, bs.CursorY, tileName)
	coverStr := ""
	if tile.Cover > 0 {
		coverStr = language.Sprintf("BATTLE_COVER_FMT", tile.Cover)
	}
	cursorX := w - len(cursorStr) - len(coverStr) - 2
	if bs.Selected != nil {
		selX := w / 2
		selEnd := selX + len(fmt.Sprintf(language.String("STATUS_SELECTED"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName))
		if cursorX < selEnd+2 {
			cursorX = selEnd + 2
		}
	}
	if cursorX < 2 {
		cursorX = 2
	}
	ctx.DrawString(cursorX, h-3, cursorStr, engine.StyleGray)
	if coverStr != "" {
		ctx.DrawString(cursorX+len(cursorStr), h-3, coverStr, engine.StyleOrange)
	}

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}

	// Draw help bar
	if !engine.Config.TouchMode {
		ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
		help := language.String("HELP_BATTLESCAPE")
		if bs.Map.NumLevels > 1 {
			help += language.String("HELP_STAIRS_SUFFIX")
		}
		ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)
	}

	// Quit confirmation dialog
	if bs.QuitConfirm {
		bs.renderQuitConfirm(ctx, w, h)
	}
}

func (bs *Battlescape) renderQuitConfirm(ctx *engine.ScreenCtx, w, h int) {
	boxW := 46
	boxH := 5
	x := (w - boxW) / 2
	y := (h - boxH) / 2
	for fy := y; fy < y+boxH; fy++ {
		for fx := x; fx < x+boxW; fx++ {
			ctx.SetCell(fx, fy, ' ', engine.StyleGray)
		}
	}
	ctx.DrawPanel(x, y, boxW, boxH, "", engine.StyleGray)
	msg := language.String("CONFIRM_BATTLE_EXIT")
	ctx.DrawString(x+(boxW-engine.StringWidth(msg))/2, y+2, msg, engine.StyleYellow)
	hint := language.String("CONFIRM_BATTLE_EXIT_HINT")
	ctx.DrawString(x+(boxW-engine.StringWidth(hint))/2, y+3, hint, engine.StyleHotkey)
}

func (bs *Battlescape) phaseStr() string {
	switch bs.Phase {
	case PhasePlayerTurn:
		return language.String("PHASE_YOUR_TURN")
	case PhaseAlienTurn:
		return language.String("PHASE_ALIEN_TURN")
	case PhaseVictory:
		return language.String("PHASE_VICTORY")
	case PhaseDefeat:
		return language.String("PHASE_DEFEAT")
	}
	return ""
}

func (bs *Battlescape) HandleKey(e *tcell.EventKey) {
	bs.HandleEvent(e)
}

// PlacedMine represents an armed proximity mine on the map.
type PlacedMine struct {
	X, Y int
}

// FloatingText is a rising, fading combat label (damage numbers, MISS, heals).
type FloatingText struct {
	X, Y    float64
	Text    string
	Color   tcell.Color
	Life    float64
	MaxLife float64
}

func (bs *Battlescape) spawnFloater(mx, my int, text string, col tcell.Color) {
	bs.floaters = append(bs.floaters, FloatingText{
		X:       float64(mx - bs.ScrollX + 1),
		Y:       float64(my - bs.ScrollY + 1),
		Text:    text,
		Color:   col,
		Life:    1.2,
		MaxLife: 1.2,
	})
}

func (bs *Battlescape) updateFloaters(dt float64) {
	n := 0
	for i := range bs.floaters {
		p := &bs.floaters[i]
		p.Life -= dt
		if p.Life <= 0 {
			continue
		}
		p.Y -= 5 * dt // rise upward
		bs.floaters[n] = bs.floaters[i]
		n++
	}
	bs.floaters = bs.floaters[:n]
}

func (bs *Battlescape) drawFloaters(ctx *engine.ScreenCtx) {
	for i := range bs.floaters {
		p := &bs.floaters[i]
		alpha := p.Life / p.MaxLife
		cr, cg, cb := p.Color.RGB()
		tr, tg, tb := 40.0, 40.0, 48.0
		r := int32(float64(cr)*alpha + tr*(1-alpha))
		g := int32(float64(cg)*alpha + tg*(1-alpha))
		b := int32(float64(cb)*alpha + tb*(1-alpha))
		style := tcell.StyleDefault.Foreground(tcell.NewRGBColor(r, g, b))
		ctx.DrawString(int(p.X), int(p.Y), p.Text, style)
	}
}

func (bs *Battlescape) drawSidebar(ctx *engine.ScreenCtx, viewW, viewH, w, h int) {
	if bs.SidebarW <= 0 {
		bs.drawCompactBanner(ctx, w)
		return
	}
	sideX, sideY0, sideH := bs.sidebarLayout(ctx, viewW, viewH, w, h)
	if bs.HoveredUnit != nil && bs.HoveredUnit != bs.Selected {
		bs.drawTargetInfo(ctx, bs.HoveredUnit, sideX, sideY0, sideH)
		return
	}
	if bs.Selected != nil {
		bs.drawUnitInfo(ctx, sideX, sideY0, sideH)
	}
}

func (bs *Battlescape) sidebarLayout(ctx *engine.ScreenCtx, viewW, viewH, w, h int) (sideX, sideY0, sideH int) {
	if engine.Layout.IsMobile() {
		sideX = 1
		sideY0 = engine.Layout.BattleSidebarY(h)
		reserved := engine.Menu.ReservedBottom(w, h)
		sideH = (h - reserved - 4) - sideY0
		if sideH < 3 {
			sideH = 3
		}
		for x := 0; x < w; x++ {
			ctx.SetCell(x, sideY0-1, '─', engine.StyleGray)
		}
	} else {
		sideX = viewW + 2
		sideY0 = 1
		sideH = viewH
		for y := 0; y < viewH; y++ {
			ctx.SetCell(sideX-1, y+1, '|', engine.StyleGray)
		}
	}
	return
}

func (bs *Battlescape) drawTargetInfo(ctx *engine.ScreenCtx, u *Unit, sideX, sideY0, sideH int) {
	sy := sideY0
	halfSide := bs.SidebarW / 2

	ctx.DrawString(sideX, sy, language.String("SIDE_TARGET_INFO"), engine.StyleRedBold)
	sy++
	name := ""
	if u.Faction == FactionAlien && u.AlienType != nil {
		name = u.AlienType.LangName()
	} else if u.Faction == FactionHuman && u.Soldier != nil {
		name = u.Soldier.Name
	} else if u.Faction == FactionCivilian {
		name = u.CivName
	}
	if engine.StringWidth(name) > halfSide-1 {
		rs := []rune(name)
		for len(rs) > 0 && engine.StringWidth(string(rs)) > halfSide-1 {
			rs = rs[:len(rs)-1]
		}
		name = string(rs)
	}
	ctx.DrawString(sideX, sy, name, engine.StyleDefault.Bold(true))
	sy++
	weaponName := data.RuleItems[u.Weapon].ShortName
	ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_WPN_TARGET"), weaponName), engine.StyleDefault)
	sy++

	hasAutopsy := u.Faction != 1 || u.AlienType == nil
	if !hasAutopsy {
		for _, id := range bs.Base.CompletedResearch {
			if id == u.AlienType.AutopsyID {
				hasAutopsy = true
				break
			}
		}
	}

	if hasAutopsy {
		hpColor := engine.StyleGreen
		if u.HP*3 < u.MaxHP {
			hpColor = engine.StyleRed
		} else if u.HP*2 < u.MaxHP {
			hpColor = engine.StyleYellow
		}
		ctx.DrawString(sideX, sy, language.Sprintf("SIDE_HP_BAR", barString(u.HP, u.MaxHP, 8), u.HP, u.MaxHP), hpColor)
		sy++
		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_ACC"), u.Accuracy), engine.StyleDefault)
		sy++
		tuColor := engine.StyleCyan
		if u.TU < u.MaxTU/3 {
			tuColor = engine.StyleRed
		} else if u.TU < u.MaxTU/2 {
			tuColor = engine.StyleYellow
		}
		ctx.DrawString(sideX, sy, language.Sprintf("SIDE_TU_BAR", barString(u.TU, u.MaxTU, 8), u.TU), tuColor)
		sy++
	} else {
		ctx.DrawString(sideX, sy, language.String("SIDE_HP_UNKNOWN"), engine.StyleGray)
		sy++
		ctx.DrawString(sideX, sy, language.String("SIDE_ACC_UNKNOWN"), engine.StyleGray)
		sy++
		ctx.DrawString(sideX, sy, language.String("SIDE_TU_UNKNOWN"), engine.StyleGray)
		sy++
	}

	if u.Faction == FactionAlien && u.AlienType != nil {
		bgColor := tcell.NewRGBColor(20, 20, 28)
		alienImg := engine.GenerateAlienSpriteFromSeed(int64(u.AlienType.Icon), u.AlienType.Morphology, bgColor)
		portW := alienImg.Width + 2
		portX := sideX + bs.SidebarW - portW
		ctx.DrawPixelImageFramed(portX, sideY0, alienImg, engine.StyleRed)
		if sy < 14 {
			sy = 14
		}
	}
	if sy < 2 {
		sy = 2
	}
	sy++
	bs.drawBattleLog(ctx, sideX, sy, sideY0, sideH)
}

func (bs *Battlescape) drawUnitInfo(ctx *engine.ScreenCtx, sideX, sideY0, sideH int) {
	sy := sideY0
	halfSide := bs.SidebarW / 2

	if bs.Selected != nil {
		ctx.DrawString(sideX, sy, language.String("SIDE_UNIT_INFO"), engine.StyleCyanBold)
		sy++
		name := bs.Selected.Soldier.Name
		if engine.StringWidth(name) > halfSide-1 {
			rs := []rune(name)
			for len(rs) > 0 && engine.StringWidth(string(rs)) > halfSide-1 {
				rs = rs[:len(rs)-1]
			}
			name = string(rs)
		}
		ctx.DrawString(sideX, sy, name, engine.StyleDefault.Bold(true))
		sy++

		hpColor := engine.StyleGreen
		if bs.Selected.HP*3 < bs.Selected.MaxHP {
			hpColor = engine.StyleRed
		} else if bs.Selected.HP*2 < bs.Selected.MaxHP {
			hpColor = engine.StyleYellow
		}
		ctx.DrawString(sideX, sy, language.Sprintf("SIDE_HP_BAR", barString(bs.Selected.HP, bs.Selected.MaxHP, 8), bs.Selected.HP, bs.Selected.MaxHP), hpColor)
		sy++

		tuColor := engine.StyleCyan
		if bs.Selected.TU < bs.Selected.MaxTU/3 {
			tuColor = engine.StyleRed
		} else if bs.Selected.TU < bs.Selected.MaxTU/2 {
			tuColor = engine.StyleYellow
		}
		ctx.DrawString(sideX, sy, language.Sprintf("SIDE_TU_BAR_FULL", barString(bs.Selected.TU, bs.Selected.MaxTU, 8), bs.Selected.TU, bs.Selected.MaxTU), tuColor)
		sy++

		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_ACC"), bs.Selected.Accuracy), engine.StyleDefault)
		sy++

		weaponName := data.RuleItems[bs.Selected.Weapon].DisplayName()
		if engine.StringWidth(weaponName) > bs.SidebarW-4 {
			rs := []rune(weaponName)
			for len(rs) > 0 && engine.StringWidth(string(rs)) > bs.SidebarW-4 {
				rs = rs[:len(rs)-1]
			}
			weaponName = string(rs)
		}
		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_WEAPON"), weaponName), engine.StyleDefault)
		sy++

		w := data.RuleItems[bs.Selected.Weapon]
		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_AMMO"), bs.Selected.WeaponAmmo, w.AmmoMax), engine.StyleDefault)
		sy++

		if len(w.Modes()) > 1 {
			ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_FIRE_MODE"), bs.Selected.FireMode.String()), engine.StyleYellow)
			sy++
		}

		armourName := language.String("NONE")
		if bs.Selected.Armour > 0 {
			for k, v := range data.Armors {
				if v.Undersuit == bs.Selected.Armour {
					armourName = v.DisplayNameByKey(k)
					break
				}
			}
		}
		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_ARMOR"), armourName), engine.StyleDefault)
		sy++

		ctx.DrawString(sideX, sy, fmt.Sprintf(language.String("SIDE_POS"), bs.Selected.X, bs.Selected.Y), engine.StyleGray)
		sy++

		if len(bs.Selected.Soldier.Inventory) > 0 {
			ctx.DrawString(sideX, sy, language.String("SIDE_INVENTORY"), engine.StyleDefault)
			sy++
			for _, item := range bs.Selected.Soldier.Inventory {
				invName := data.ItemDisplayName(item)
				if engine.StringWidth(invName) > bs.SidebarW-4 {
					rs := []rune(invName)
					for len(rs) > 0 && engine.StringWidth(string(rs)) > bs.SidebarW-4 {
						rs = rs[:len(rs)-1]
					}
					invName = string(rs)
				}
				ctx.DrawString(sideX+2, sy, invName, engine.StyleGray)
				sy++
			}
		}

		if bs.Selected.Crouching {
			ctx.DrawString(sideX, sy, language.String("SIDE_CROUCH"), engine.StyleYellow)
			sy++
		}
		sy++

		portraitImg := engine.MakeSoldierPortrait(bs.Selected.Soldier.Name, 20, 24)
		portW := portraitImg.Width + 2
		portX := sideX + bs.SidebarW - portW
		ctx.DrawPixelImageFramed(portX, sideY0, portraitImg, engine.StyleCyan)
		sy = 1 + portraitImg.Height/2 + 2
	}

	bs.drawBattleLog(ctx, sideX, sy, sideY0, sideH)
}

func (bs *Battlescape) drawBattleLog(ctx *engine.ScreenCtx, sideX, sy, sideY0, sideH int) {
	logTitle := language.String("BATTLE_LOG")
	ctx.DrawString(sideX, sy, logTitle, engine.StyleCyanBold)
	sy++
	availableLines := sideH - (sy - sideY0)
	logEntries := len(bs.Log)
	startIdx := 0
	if logEntries > availableLines {
		startIdx = logEntries - availableLines
	}
	for i := 0; i < availableLines && startIdx+i < logEntries; i++ {
		entry := bs.Log[startIdx+i]
		text := entry.Text
		runes := []rune(text)
		for len(runes) > 0 && engine.StringWidth(string(runes)) > bs.SidebarW-3 {
			runes = runes[:len(runes)-1]
		}
		text = string(runes)
		style := engine.StyleDefault
		if entry.Turn < bs.Turn {
			style = engine.StyleGray
		}
		ctx.DrawString(sideX, sy+i, ">", engine.StyleHotkey)
		ctx.DrawString(sideX+2, sy+i, text, style)
	}
}

func (bs *Battlescape) drawCompactBanner(ctx *engine.ScreenCtx, w int) {
	if bs.Selected != nil {
		banner := fmt.Sprintf(language.String("BATTLE_COMPACT_BANNER"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName)
		runes := []rune(banner)
		maxW := w - 2
		for len(runes) > 0 && engine.StringWidth(string(runes)) > maxW {
			runes = runes[:len(runes)-1]
		}
		ctx.DrawString(1, 1, string(runes), engine.StyleCyan)
	}
}

func (bs *Battlescape) SpawnBloodSplatter(target *Unit) {
	bloodType := 1
	if target.Faction == 1 {
		if target.AlienType != nil {
			switch target.AlienType.Name {
			case "Muton", "Muton Navigator", "Muton Commander",
				"Chryssalid", "Chryssalid Queen", "Hyperworm":
				bloodType = 2
			default:
				bloodType = 3
			}
		} else {
			bloodType = 3
		}
	}
	bs.Map.SpawnBlood(target.X, target.Y, bloodType)
}

func (bs *Battlescape) SpawnFire(x, y, turns int) {
	if x < 0 || x >= bs.Map.Width || y < 0 || y >= bs.Map.Height {
		return
	}
	tile := &bs.Map.Tiles[y][x]
	if tile.Fire > 0 {
		return
	}
	tile.Fire = turns
}

func (bs *Battlescape) ToggleVision() {
	switch bs.VisionMode {
	case engine.VisionNormal:
		bs.VisionMode = engine.VisionNight
		bs.AddMessage(language.String("MSG_VISION_NIGHT"))
	case engine.VisionNight:
		bs.VisionMode = engine.VisionThermal
		bs.AddMessage(language.String("MSG_VISION_THERMAL"))
	case engine.VisionThermal:
		bs.VisionMode = engine.VisionNormal
		bs.AddMessage(language.String("MSG_VISION_NORMAL"))
	}
}

func (bs *Battlescape) HandleMouse(e *tcell.EventMouse) {
	if !engine.Config.MouseEnabled {
		return
	}
	bs.HandleEvent(e)
}

func (bs *Battlescape) cycleUnit(dir int) {
	humans := bs.Units.Faction(0).Alive().OnLevel(bs.Map.CurrentLevel)
	if len(humans) == 0 {
		return
	}
	idx := -1
	for i, u := range humans {
		if u == bs.Selected {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(humans) - 1
	}
	if idx >= len(humans) {
		idx = 0
	}
	bs.SetSelected(humans[idx])
	bs.CursorX = bs.Selected.X
	bs.CursorY = bs.Selected.Y

	bs.Camera.SetTarget(bs.Selected.X, bs.Selected.Y)

	bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.TU))
}

func tileTypeName(t TileType) string {
	switch t {
	case TileFloor:
		return language.String("TILE_FLOOR")
	case TileWall:
		return language.String("TILE_WALL")
	case TileDoor:
		return language.String("TILE_DOOR")
	case TileGrass:
		return language.String("TILE_GRASS")
	case TileTree:
		return language.String("TILE_TREE")
	case TileRock:
		return language.String("TILE_ROCK")
	case TileWater:
		return language.String("TILE_WATER")
	case TileUFOFloor:
		return language.String("TILE_UFO_FLOOR")
	case TileUFOWall:
		return language.String("TILE_UFO_WALL")
	case TileConsole:
		return language.String("TILE_CONSOLE")
	case TileMachinery:
		return language.String("TILE_MACHINERY")
	case TilePod:
		return language.String("TILE_POD")
	case TilePowerSource:
		return language.String("TILE_POWER_SOURCE")
	case TileStorage:
		return language.String("TILE_STORAGE")
	case TileAlienTech:
		return language.String("TILE_ALIEN_TECH")
	case TileStairs:
		return language.String("TILE_STAIRS")
	case TileStairsDown:
		return language.String("TILE_STAIRS_DOWN")
	case TileBush:
		return language.String("TILE_BUSH")
	case TileFence:
		return language.String("TILE_FENCE")
	case TileSand:
		return language.String("TILE_SAND")
	case TileSnow:
		return language.String("TILE_SNOW")
	case TileMarsh:
		return language.String("TILE_MARSH")
	case TilePavement:
		return language.String("TILE_PAVEMENT")
	case TileRubble:
		return language.String("TILE_RUBBLE")
	case TileWindow:
		return language.String("TILE_WINDOW")
	case TileDesk:
		return language.String("TILE_DESK")
	case TileChair:
		return language.String("TILE_CHAIR")
	case TileComputer:
		return language.String("TILE_COMPUTER")
	case TileBed:
		return language.String("TILE_BED")
	case TileLocker:
		return language.String("TILE_LOCKER")
	case TileCabinet:
		return language.String("TILE_CABINET")
	}
	return language.String("TILE_UNKNOWN")
}

func (ul UnitList) Faction(f Faction) UnitList {
	var result UnitList
	for _, u := range ul {
		if u.Faction == f {
			result = append(result, u)
		}
	}
	return result
}

func (bs *Battlescape) ApplyCursorStyles(x, y int, style tcell.Style) tcell.Style {
	bs.State.mu.RLock()
	defer bs.State.mu.RUnlock()

	if x == bs.CursorX && y == bs.CursorY {
		switch bs.State.CursorState {
		case StateInspect:
			return engine.StyleHighlight
		case StateTargeting:
			return style.Background(color.Red).Blink(true)
		}
	}

	if bs.State.CursorState == StateMovePlan {
		for _, p := range bs.State.MovePath {
			if p[0] == x && p[1] == y {
				return style.Background(color.DarkBlue)
			}
		}
	}

	return style
}

// barString returns a filled/empty bar string for the HP/TU display.
func barString(current, max, length int) string {
	if max <= 0 {
		max = 1
	}
	filled := current * length / max
	if filled > length {
		filled = length
	}
	if filled < 0 {
		filled = 0
	}
	out := make([]rune, length)
	for i := 0; i < length; i++ {
		if i < filled {
			out[i] = '█'
		} else {
			out[i] = '░'
		}
	}
	return string(out)
}
