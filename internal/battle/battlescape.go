package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/civ13/ycom/internal/audio"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

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

type BattlePhase int

const (
	PhasePlayerTurn BattlePhase = iota
	PhaseAlienTurn
	PhaseVictory
	PhaseDefeat
)

type Projectile struct {
	FromX, FromY int
	ToX, ToY     int
	Progress     int
	Length       int
	Symbol       rune
	Style        tcell.Style
}

type AlienAction struct {
	Type   string // "move", "fire", "melee", "patrol"
	Unit   *Unit
	Target *Unit
	FromX, FromY int
	ToX, ToY     int
}

type Battlescape struct {
	Game       *engine.Game
	Base       *base.Base
	Map        *BattleMap
	Units      UnitList
	AlienAIs   []*AlienAI
	CivilianAIs []*CivilianAI
	Phase      BattlePhase
	Turn       int
	CursorX    int
	CursorY    int
	Selected   *Unit
	Message    string
	Log        []string
	ScrollX    int
	ScrollY    int
	Squad      []*soldier.Soldier
	UFOName    string
	ExitTimer  int
	IsNight    bool

	AlienTurnQueue  []AlienAction
	AlienTurnIdx    int
	AlienTurnDelay  int
	Projectile      *Projectile

	Camera        *engine.Camera
	Particles     *engine.ParticleSystem
	HoveredUnit   *Unit
	SidebarW      int
	ReinforceTimer int
	AbductionCivs  int
	AbductionTotal int

	Gas         *GasGrid
	VisionMode  engine.VisionMode
	FrameCount  int
	
	// Input State
	State       BattleState
}

func (bs *Battlescape) AddMessage(msg string) {
	bs.Message = msg
	bs.Log = append(bs.Log, msg)
	if len(bs.Log) > 50 {
		bs.Log = bs.Log[len(bs.Log)-50:]
	}
}

// GetMovementRange returns a map of tiles the selected unit can reach
func (bs *Battlescape) GetMovementRange() map[[2]int]bool {
	result := make(map[[2]int]bool)
	if bs.Selected == nil || bs.Selected.TU <= 0 {
		return result
	}

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

func NewBattlescape(g *engine.Game, b *base.Base, squad []*soldier.Soldier, ufoName string) *Battlescape {
	var m *BattleMap
	switch ufoName {
	case "Terror":
		m = GenerateTerrorSite(50, 50)
	case "Supply":
		m = GenerateUFOInterior(50, 50)
	case "Alien Base Assault":
		m = GenerateCydonia(50, 50)
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
		m = GenerateCrashSite(50, 50)
	}

	bs := &Battlescape{
		Game:    g,
		Base:    b,
		Map:     m,
		Phase:   PhasePlayerTurn,
		Turn:    1,
		CursorX: 3,
		CursorY: m.Height - 3,
		Squad:   squad,
		UFOName: ufoName,
		IsNight: g.GameTime.Hour() < 6 || g.GameTime.Hour() > 18,
		Camera:   engine.NewCamera(3, m.Height-3),
		Particles: engine.NewParticleSystem(512),
		Gas:       NewGasGrid(m.Width, m.Height),
	}

	if ufoName == "Abduction" {
		bs.AbductionTotal = 8 + rand.Intn(5)
	}
	m.Gas = bs.Gas

	bs.AddMessage(fmt.Sprintf(language.String("MSG_MISSION_START"), ufoName))

	for i, s := range squad {
		if s.HP <= 0 {
			continue
		}
		u := NewSoldierUnit(s)
		u.X = 3 + i*2
		u.Y = m.Height - 3
		u.IsNight = bs.IsNight
		bs.Units = append(bs.Units, u)
	}

	alienTypes := g.GetAlienTypes()
	alienRank := 0
	spawnAliens := []*data.AlienType{
		getAlienByRank(alienTypes, alienRank),
		getAlienByRank(alienTypes, alienRank),
		getAlienByRank(alienTypes, alienRank),
		getAlienByRank(alienTypes, alienRank+1),
		getAlienByRank(alienTypes, alienRank+1),
	}

	for _, at := range spawnAliens {
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
		u.X = 10 + rand.Intn(m.Width-14)
		u.Y = 3 + rand.Intn(m.Height/2-4)
		u.IsNight = bs.IsNight
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
			u.X = 5 + rand.Intn(m.Width-10)
			u.Y = m.Height/2 + rand.Intn(m.Height/2-5)
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
			u.X = 10 + rand.Intn(m.Width-20)
			u.Y = 5 + rand.Intn(m.Height-10)
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

func (bs *Battlescape) ComputeFOVForTeam() {
	bs.Map.ClearVisibility()
	sightRange := SightRange
	if bs.IsNight {
		sightRange = 10
	}
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive && u.Level == bs.Map.CurrentLevel {
			bs.Map.ComputeFOV(u.X, u.Y, sightRange)
		}
	}
	for _, u := range bs.Units {
		if u.Faction == 1 && u.Alive && u.AlienType != nil && u.Level == bs.Map.CurrentLevel && bs.Map.IsVisible(u.X, u.Y) {
			bs.Game.LearnAlien(u.AlienType.Name, 1)
		}
	}
	bs.Gas.Visible = func(x, y int) bool {
		return bs.Map.IsVisible(x, y)
	}
}

func (bs *Battlescape) Update() {
	dt := 0.016
	bs.FrameCount++
	bs.Camera.UpdateShake(dt)
	bs.Particles.Update(dt)

	if bs.FrameCount%1800 == 0 {
		audio.PlayWind()
	}

	if bs.FrameCount%12 == 0 && bs.Phase != PhaseVictory && bs.Phase != PhaseDefeat {
		switch bs.UFOName {
		case "Polar":
			engine.SpawnSnow(bs.Particles, 0, 0, bs.Map.Width, 1)
		case "Desert":
			engine.SpawnDust(bs.Particles, 0, 0, bs.Map.Width, bs.Map.Height)
		case "Cydonia", "Alien Base Assault":
			engine.SpawnEmbers(bs.Particles, 0, bs.Map.Height/2, bs.Map.Width, bs.Map.Height/2)
		}
	}

	if bs.Phase == PhaseAlienTurn {
		if bs.Projectile != nil {
			bs.Projectile.Progress++
			if bs.Projectile.Progress >= bs.Projectile.Length {
				bs.Projectile = nil
			}
			return
		}

		if bs.AlienTurnDelay > 0 {
			bs.AlienTurnDelay--
			return
		}

		if bs.AlienTurnIdx < len(bs.AlienTurnQueue) {
			action := bs.AlienTurnQueue[bs.AlienTurnIdx]
			bs.AlienTurnIdx++
			bs.executeAlienAction(action)
			bs.AlienTurnDelay = 8  // Increased delay so messages are visible
		} else {
			for _, cai := range bs.CivilianAIs {
				actions := cai.GenerateActions(bs.Units, bs.Map)
				for _, a := range actions {
					bs.executeAlienAction(a)
				}
			}
			bs.finishAlienTurn()
		}
	}

	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		bs.ExitTimer++
		if bs.ExitTimer > 60 {
			bs.finishBattle()
		}
	}
}

func (bs *Battlescape) executeAlienAction(action AlienAction) {
	switch action.Type {
	case "fire":
		if action.Target == nil || !action.Target.Alive {
			return
		}
		damage, hit, err := action.Unit.FireAt(action.Target, bs.Map)
		if err != nil {
			return
		}
		audio.PlayWeaponFire(action.Unit.Weapon)
		dx := action.Target.X - action.Unit.X
		dy := action.Target.Y - action.Unit.Y
		length := int(math.Sqrt(float64(dx*dx+dy*dy)))
		if length < 1 {
			length = 1
		}
		symbol := '*'
		if data.RuleItems[action.Unit.Weapon].AmmoMax >= 99 {
			symbol = '|'
		}
		bs.Projectile = &Projectile{
			FromX: action.Unit.X, FromY: action.Unit.Y,
			ToX: action.Target.X, ToY: action.Target.Y,
			Progress: 0, Length: length,
			Symbol: symbol,
			Style:  engine.StyleYellow,
		}
		engine.SpawnMuzzleFlash(bs.Particles, action.Unit.X-bs.ScrollX+1, action.Unit.Y-bs.ScrollY+1)
		if hit {
			engine.SpawnExplosion(bs.Particles, action.Target.X-bs.ScrollX+1, action.Target.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
			bs.Camera.TriggerShake(0.5)
			bs.SpawnBloodSplatter(action.Target)
			name := action.Target.Soldier.Name
			if action.Target.AlienType != nil {
				name = action.Target.AlienType.Name
			}
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_HIT"), name, damage))
			if !action.Target.Alive {
				bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_KILL"), name))
			}
		} else {
			bs.AddMessage(language.String("MSG_ALIEN_MISS"))
		}
	case "melee":
		if action.Target == nil || !action.Target.Alive {
			return
		}
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
		name := action.Target.Soldier.Name
		if action.Target.AlienType != nil {
			name = action.Target.AlienType.Name
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_MELEE"), action.Unit.AlienType.Name, name, damage))
		if !action.Target.Alive {
			bs.AddMessage(fmt.Sprintf(language.String("MSG_ALIEN_KILL"), name))
		}
	case "move":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
		bs.ComputeFOVForTeam()
		bs.checkHumanReactionFire(action.Unit)
	case "patrol":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
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
		if u.TU < 15 || u.Weapon == "" {
			continue
		}
		if !u.CanSee(movedAlien.X, movedAlien.Y, bs.Map) {
			continue
		}
		dx := movedAlien.X - u.X
		dy := movedAlien.Y - u.Y
		dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist == 0 {
			dist = 1
		}
		chance := u.Reactions*2 + u.Accuracy/3 - dist*5
		if chance < 5 {
			chance = 5
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
		damage, hit, _ := u.FireAt(movedAlien, bs.Map)
		bs.Projectile = &Projectile{
			FromX: u.X, FromY: u.Y,
			ToX: movedAlien.X, ToY: movedAlien.Y,
			Progress: 0, Length: dist,
			Symbol: '*', Style: engine.StyleCyanBold,
		}
		if hit {
			engine.SpawnExplosion(bs.Particles, movedAlien.X-bs.ScrollX+1, movedAlien.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
			bs.Camera.TriggerShake(0.3)
			bs.SpawnBloodSplatter(movedAlien)
			bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_HIT"), damage, movedAlien.Name(), movedAlien.HP))
			if !movedAlien.Alive {
				bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_KILL"), movedAlien.Name()))
			}
		} else {
			bs.AddMessage(language.String("MSG_REACTION_MISS"))
		}
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
		if u.TU < 15 || u.Weapon == "" {
			continue
		}
		if !bs.Map.IsVisible(u.X, u.Y) {
			continue
		}
		if !u.CanSee(movedHuman.X, movedHuman.Y, bs.Map) {
			continue
		}
		dx := movedHuman.X - u.X
		dy := movedHuman.Y - u.Y
		dist := int(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist == 0 {
			dist = 1
		}
		chance := u.Reactions*2 + u.Accuracy/3 - dist*5
		if chance < 5 {
			chance = 5
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
		damage, hit, _ := u.FireAt(movedHuman, bs.Map)
		dist2 := int(math.Sqrt(float64((movedHuman.X-u.X)*(movedHuman.X-u.X) + (movedHuman.Y-u.Y)*(movedHuman.Y-u.Y))))
		if dist2 < 1 {
			dist2 = 1
		}
		bs.Projectile = &Projectile{
			FromX: u.X, FromY: u.Y,
			ToX: movedHuman.X, ToY: movedHuman.Y,
			Progress: 0, Length: dist2,
			Symbol: '*', Style: engine.StyleRedBold,
		}
		if hit {
			engine.SpawnExplosion(bs.Particles, movedHuman.X-bs.ScrollX+1, movedHuman.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
			bs.Camera.TriggerShake(0.3)
			bs.SpawnBloodSplatter(movedHuman)
			bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_HIT"), damage, movedHuman.Name(), movedHuman.HP))
			if !movedHuman.Alive {
				bs.AddMessage(fmt.Sprintf(language.String("MSG_REACTION_KILL"), movedHuman.Name()))
			}
		} else {
			bs.AddMessage(language.String("MSG_REACTION_MISS"))
		}
		return
	}
}

func (bs *Battlescape) finishBattle() {
	won := bs.Phase == PhaseVictory

	// Sync soldier HP back to roster
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Soldier != nil {
			u.Soldier.HP = u.HP
			if u.HP <= 0 {
				u.Soldier.HP = 0
				u.Soldier.Wounds = 30
			} else if u.HP < u.MaxHP {
				dmg := u.MaxHP - u.HP
				u.Soldier.Wounds = dmg * 3
				if u.Soldier.Wounds > 30 {
					u.Soldier.Wounds = 30
				}
			}
		}
	}

	// Count alien kills
	alienKills := 0
	for _, u := range bs.Units {
		if u.Faction == 1 && !u.Alive {
			alienKills++
		}
	}

	// Award XP to surviving soldiers
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive && u.Soldier != nil {
			xp := alienKills * 5
			if won {
				xp += 10
			}
			u.Soldier.GainXP(xp)
			u.Soldier.Missions++
		}
	}

	// Collect loot — type-specific corpses
	var loot []string
	if won {
		corpseMap := map[string]string{
			"SEC": "corpse_sect",
			"SEL": "corpse_sect",
			"FLT": "corpse_float",
			"FLL": "corpse_float",
			"MUT": "corpse_muton",
			"MUL": "corpse_muton",
			"ETH": "corpse_ether",
			"EHL": "corpse_ether",
		}
		corpses := make(map[string]bool)
		for _, u := range bs.Units {
			if u.Faction == 1 && !u.Alive && u.AlienType != nil {
				if key, ok := corpseMap[u.AlienType.ShortName]; ok {
					corpses[key] = true
				}
			}
		}
		for key := range corpses {
			loot = append(loot, key)
		}
		if len(loot) == 0 {
			loot = append(loot, "alien_corpse")
		}
		if rand.Intn(100) < 40 {
			loot = append(loot, "alloys")
		}
		if rand.Intn(100) < 25 {
			loot = append(loot, "elerium")
		}
	}

	// Find surviving squad soldiers
	var surviving []*soldier.Soldier
	for _, s := range bs.Squad {
		for _, u := range bs.Units {
			if u.Faction == 0 && u.Soldier == s {
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

	bs.Game.ActiveBattle = &engine.BattleResult{
		Won:       won,
		Kills:     alienKills,
		Soldiers:  surviving,
		LootItems: loot,
	}
	bs.Game.PopState()
}

func (bs *Battlescape) finishAlienTurn() {
	bs.Phase = PhasePlayerTurn
	bs.restorePlayerTU()
	bs.ComputeFOVForTeam()
	bs.Turn++
	bs.learnFromKills()
	bs.Gas.Diffuse()
	bs.Map.SpreadFire()
	if bs.UFOName == "Abduction" {
		bs.processAbduction()
	}
	bs.checkReinforcements()
	bs.checkVictory()
	bs.Selected = nil
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive && u.HP > 0 && u.Level == bs.Map.CurrentLevel {
			bs.Selected = u
			break
		}
	}
}

func (bs *Battlescape) processAbduction() {
	aliveAliens := 0
	for _, u := range bs.Units {
		if u.Faction == 1 && u.Alive {
			aliveAliens++
		}
	}
	if aliveAliens == 0 {
		return
	}
	// Each turn, aliens abduct 1 civilian if any are alive
	for _, u := range bs.Units {
		if u.Faction == 2 && u.Alive {
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
	// Reinforcements only on terror and alien base missions
	if bs.UFOName != "Terror" && bs.UFOName != "Alien Base Assault" {
		return
	}
	aliveAliens := 0
	for _, u := range bs.Units {
		if u.Faction == 1 && u.Alive {
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

	g := bs.Game
	alienTypes := g.GetAlienTypes()
	alienRank := 0
	if bs.Turn > 6 {
		alienRank = 1
	}
	if bs.Turn > 12 {
		alienRank = 2
	}
	count := 1 + rand.Intn(2)
	for i := 0; i < count; i++ {
		at := getAlienByRank(alienTypes, alienRank)
		if at == nil {
			continue
		}
		u := NewAlienUnit(at)
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
	}
	bs.AddMessage(fmt.Sprintf(language.String("MSG_REINFORCEMENTS"), count))
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
		if u.Faction == 0 && u.Alive {
			u.TU = u.MaxTU
		}
	}
}

func (bs *Battlescape) checkVictory() {
	humans := bs.Units.Faction(0).Alive()
	aliens := bs.Units.Faction(1).Alive()
	civilians := bs.Units.Faction(2).Alive()
	if len(aliens) == 0 {
		bs.Phase = PhaseVictory
		audio.PlayVictory()
		if bs.UFOName == "Terror" {
			totalCiv := 0
			for _, u := range bs.Units {
				if u.Faction == 2 {
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
		bs.Phase = PhaseDefeat
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_MISSION_FAILED"))
	} else if bs.UFOName == "Terror" && len(civilians) == 0 && len(aliens) > 0 {
		bs.Phase = PhaseDefeat
		audio.PlayDefeat()
		bs.AddMessage(language.String("MSG_MISSION_FAILED_CIV"))
	} else if bs.UFOName == "Abduction" && bs.AbductionCivs >= bs.AbductionTotal {
		bs.Phase = PhaseDefeat
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
}

func (bs *Battlescape) SelectUnit() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		audio.PlaySelect()
		bs.Selected = unit
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
	} else {
		bs.cycleUnit(1)
	}
}

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
		bs.Selected = unit
		bs.State.CursorState = StateInspect
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
		return
	}
	bs.State.CursorState = StateInspect
}

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

func (bs *Battlescape) MoveSelected() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
		audio.PlayMove()
		bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY))
		bs.ComputeFOVForTeam()
		bs.checkAlienReactionFire(bs.Selected)
	} else {
		bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
	}
}

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
	damage, hit, err := bs.Selected.FireAt(target, bs.Map)
	if err != nil {
		bs.AddMessage(err.Error())
		return
	}
	audio.PlayWeaponFire(bs.Selected.Weapon)
	engine.SpawnMuzzleFlash(bs.Particles, bs.Selected.X-bs.ScrollX+1, bs.Selected.Y-bs.ScrollY+1)
	if hit {
		audio.PlayHit()
		engine.SpawnExplosion(bs.Particles, target.X-bs.ScrollX+1, target.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
		bs.Camera.TriggerShake(0.5)
		bs.SpawnBloodSplatter(target)
		w := data.RuleItems[bs.Selected.Weapon]
		if w.Type == "plasma" || w.Type == "explosive" {
			bs.SpawnFire(target.X, target.Y, 3)
		}
		name := "alien"
		if target.AlienType != nil {
			name = target.AlienType.Name
		}
		bs.AddMessage(fmt.Sprintf(language.String("MSG_HIT_TARGET"), damage, name, target.HP))
	} else {
		audio.PlayMiss()
		bs.AddMessage(language.String("MSG_MISSED"))
	}
}

func (bs *Battlescape) Reload() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 8 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_RELOAD"))
		return
	}
	w := data.Weapons[bs.Selected.Weapon]
	if w.AmmoMax >= 99 {
		bs.AddMessage(language.String("MSG_ENERGY_WEAPON"))
		return
	}
	if w.AmmoCur >= w.AmmoMax {
		bs.AddMessage(language.String("MSG_WEAPON_LOADED"))
		return
	}
	bs.Selected.TU -= 8
	w.AmmoCur = w.AmmoMax
	data.Weapons[bs.Selected.Weapon] = w
	audio.PlayReload()
	bs.AddMessage(fmt.Sprintf(language.String("MSG_RELOADED"), w.Name, w.AmmoCur, w.AmmoMax))
}

func (bs *Battlescape) EndTurn() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	audio.PlayAlienTurn()
	bs.Phase = PhaseAlienTurn
	bs.AddMessage(language.String("MSG_ALIEN_TURN"))

	bs.AlienTurnQueue = nil
	bs.AlienTurnIdx = 0

	for _, ai := range bs.AlienAIs {
		if !ai.Unit.Alive {
			continue
		}
		ai.Unit.TU = ai.Unit.MaxTU
		humanUnits := bs.Units.Faction(0)
		actions := ai.GenerateActions(bs.Units, bs.Map, humanUnits)
		bs.AlienTurnQueue = append(bs.AlienTurnQueue, actions...)
	}

	if len(bs.AlienTurnQueue) == 0 {
		bs.finishAlienTurn()
	} else {
		bs.AlienTurnDelay = 3
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
		bs.AddMessage("No stairs on this map.")
		return
	}
	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	if tile.Type != TileStairs && tile.Type != TileStairsDown {
		// Check if selected unit is on stairs
		if bs.Selected != nil {
			tile = bs.Map.At(bs.Selected.X, bs.Selected.Y)
		}
		if tile.Type != TileStairs && tile.Type != TileStairsDown {
			bs.AddMessage("Move to stairs first.")
			return
		}
	}

	oldLevel := bs.Map.CurrentLevel
	if oldLevel == 0 && bs.Map.NumLevels > 1 {
		bs.Map.CurrentLevel = 1
	} else if oldLevel > 0 {
		bs.Map.CurrentLevel = 0
	} else {
		bs.AddMessage("No stairs here.")
		return
	}

	// Teleport selected unit to stairs on new level
	if bs.Selected != nil && bs.Selected.TU >= 8 {
		bs.Selected.TU -= 8
		bs.Selected.Level = bs.Map.CurrentLevel
		bs.ComputeFOVForTeam()
		bs.AddMessage(fmt.Sprintf("Descended to level %d.", bs.Map.CurrentLevel+1))
	} else if bs.Selected != nil {
		bs.Map.CurrentLevel = oldLevel
		bs.AddMessage("Not enough TU to use stairs.")
	} else {
		bs.Map.CurrentLevel = oldLevel
		bs.AddMessage("Select a soldier first.")
	}
}

func (bs *Battlescape) Grenade() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.TU < 20 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_GRENADE"))
		return
	}

	grenadeRange := 6
	damage := 40 + bs.Selected.Strength*2
	ax := bs.CursorX
	ay := bs.CursorY
	dx := ax - bs.Selected.X
	dy := ay - bs.Selected.Y
	dist := dx*dx + dy*dy
	if dist > grenadeRange*grenadeRange {
		bs.AddMessage(language.String("MSG_GRENADE_OUT_OF_RANGE"))
		return
	}

	bs.Selected.TU -= 20

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		udx := u.X - ax
		udy := u.Y - ay
		udist := udx*udx + udy*udy
		if udist <= 4 {
			splashDmg := damage - udist*5
			if splashDmg < 5 {
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

	bs.Camera.TriggerShake(3.0)
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

func (bs *Battlescape) PsiAttack() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.Soldier.Weapon != "psi_amp" {
		bs.AddMessage("Need Psi-Amplifier to use Psi.")
		return
	}

	if bs.Selected.TU < 20 {
		bs.AddMessage(language.String("MSG_NOT_ENOUGH_TU_PSI"))
		return
	}

	target := bs.Units.At(bs.CursorX, bs.CursorY)
	if target == nil || target.Faction != 1 {
		bs.AddMessage("Select an alien target.")
		return
	}

	bs.Selected.TU -= 20

	targetPsi := 0
	if target.AlienType != nil {
		targetPsi = target.AlienType.Psi
	} else if target.Soldier != nil {
		targetPsi = target.Soldier.PsiStr
	}

	success := rand.Intn(100) < (bs.Selected.Soldier.PsiSkill - targetPsi/2)

	if success {
		target.TU = 0
		bs.AddMessage(fmt.Sprintf(language.String("MSG_PSI_SUCCESS"), target.Name()))
	} else {
		bs.AddMessage(language.String("MSG_PSI_FAIL"))
	}
}

func (bs *Battlescape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	bs.SidebarW = w / 3
	if bs.SidebarW < 20 {
		bs.SidebarW = 20
	}
	viewW := w - bs.SidebarW - 2
	if viewW < 10 {
		viewW = 10
	}
	viewH := h - 5

	camX, camY := bs.Camera.Pos()
	bs.ScrollX = camX - viewW/2
	bs.ScrollY = camY - viewH/2

	blackStyle := tcell.StyleDefault.Background(color.XTerm0).Foreground(color.XTerm0)

	for y := 0; y < viewH; y++ {
		for x := 0; x < viewW; x++ {
			mx := x + bs.ScrollX
			my := y + bs.ScrollY

			if mx < 0 || mx >= bs.Map.Width || my < 0 || my >= bs.Map.Height {
				ctx.SetCell(x+1, y+1, ' ', blackStyle)
				continue
			}

			tile := bs.Map.At(mx, my)

			if !tile.Seen {
				if bs.IsNight {
					ctx.SetCell(x+1, y+1, ' ', blackStyle)
				} else {
					ctx.SetCell(x+1, y+1, ' ', engine.StyleDefault)
				}
				continue
			}

			ch := TileChar(tile.Type)
			style := engine.StyleGreen

			v := tileVar(mx, my, 6)
			switch tile.Type {
			case TileGrass:
				grassChars := []rune{'·', '·', '\'', ',', '·', '"'}
				ch = grassChars[v]
			case TileTree:
				treeChars := []rune{'♣', '♠', '\u03C8', '\u03A8', '♣', '♠'}
				ch = treeChars[v]
			case TileBush:
				bushChars := []rune{'†', '\u03C8', '‡', '†', '\u03C8', '†'}
				ch = bushChars[v%len(bushChars)]
			}

			if tile.Visible {
				switch tile.Type {
				case TileGrass:
					grassColors := [][3]int32{{42, 130, 42}, {55, 148, 35}, {36, 118, 58}, {62, 142, 30}, {32, 112, 65}, {48, 138, 28}}
					gc := grassColors[v]
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(gc[0], gc[1], gc[2]))
				case TileWall:
					wv := tileVar(mx, my, 4)
					wallColors := [][3]int32{{108, 108, 108}, {92, 92, 95}, {118, 112, 108}, {100, 98, 100}}
					wc := wallColors[wv]
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(wc[0], wc[1], wc[2]))
				case TileDoor:
					style = engine.StyleYellow
				case TileTree:
					treeColors := [][3]int32{{22, 112, 22}, {32, 92, 14}, {16, 122, 42}, {42, 98, 10}, {26, 108, 50}, {38, 88, 18}}
					tc := treeColors[v]
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(tc[0], tc[1], tc[2]))
				case TileBush:
					bushColors := [][3]int32{{52, 132, 28}, {42, 118, 42}, {62, 122, 18}, {36, 128, 52}, {58, 115, 30}, {44, 125, 40}}
					bc := bushColors[v]
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(bc[0], bc[1], bc[2]))
				case TileRock:
					rv := tileVar(mx, my, 4)
					rockColors := [][3]int32{{128, 118, 105}, {105, 98, 92}, {138, 128, 112}, {115, 108, 100}}
					rc := rockColors[rv]
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(rc[0], rc[1], rc[2]))
				case TileWater:
					if bs.IsNight {
						style = tcell.StyleDefault.Foreground(
							tcell.NewRGBColor(20, 60, 140),
						).Background(tcell.NewRGBColor(0, 10, 60))
					} else {
						style = tcell.StyleDefault.Foreground(
							tcell.NewRGBColor(40, 100, 200),
						).Background(tcell.NewRGBColor(0, 30, 120))
					}
				case TileUFOFloor:
					style = engine.StyleCyan
				case TileUFOWall:
					style = engine.StyleCyanBold
				case TileConsole:
					style = engine.StyleYellow
				case TileMachinery:
					style = engine.StyleGray
				case TilePod:
					style = engine.StyleGreenBold
				case TilePowerSource:
					style = engine.StyleRed
				case TileStorage:
					style = engine.StyleYellow
				case TileAlienTech:
					style = engine.StyleMagenta
				}
			} else {
				if bs.IsNight {
					switch tile.Type {
					case TileGrass:
						sgv := tileVar(mx, my, 3)
						seenGrass := [][3]int32{{8, 20, 8}, {10, 22, 6}, {7, 18, 10}}
						sgc := seenGrass[sgv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(sgc[0], sgc[1], sgc[2]))
					case TileTree:
						stv := tileVar(mx, my, 3)
						seenTree := [][3]int32{{5, 18, 5}, {7, 16, 3}, {4, 20, 7}}
						stc := seenTree[stv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(stc[0], stc[1], stc[2]))
					case TileBush:
						sbv := tileVar(mx, my, 3)
						seenBush := [][3]int32{{9, 22, 6}, {7, 18, 9}, {10, 20, 5}}
						sbc := seenBush[sbv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(sbc[0], sbc[1], sbc[2]))
					case TileWall:
						swv := tileVar(mx, my, 3)
						seenWall := [][3]int32{{20, 20, 20}, {17, 17, 18}, {22, 21, 20}}
						swc := seenWall[swv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(swc[0], swc[1], swc[2]))
					case TileRock:
						srv := tileVar(mx, my, 3)
						seenRock := [][3]int32{{22, 20, 18}, {18, 16, 15}, {24, 22, 20}}
						src := seenRock[srv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(src[0], src[1], src[2]))
					case TileWater:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(8, 14, 26))
					case TileUFOFloor:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(10, 20, 22))
					case TileUFOWall:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(12, 23, 26))
					case TileDoor:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(22, 19, 7))
					case TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(30, 30, 30))
					}
				} else {
					switch tile.Type {
					case TileGrass:
						sgv := tileVar(mx, my, 3)
						seenGrass := [][3]int32{{20, 52, 20}, {24, 58, 16}, {17, 48, 26}}
						sgc := seenGrass[sgv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(sgc[0], sgc[1], sgc[2]))
					case TileTree:
						stv := tileVar(mx, my, 3)
						seenTree := [][3]int32{{12, 48, 12}, {18, 42, 8}, {9, 52, 18}}
						stc := seenTree[stv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(stc[0], stc[1], stc[2]))
					case TileBush:
						sbv := tileVar(mx, my, 3)
						seenBush := [][3]int32{{22, 54, 14}, {18, 48, 22}, {26, 50, 12}}
						sbc := seenBush[sbv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(sbc[0], sbc[1], sbc[2]))
					case TileWall:
						swv := tileVar(mx, my, 3)
						seenWall := [][3]int32{{52, 52, 52}, {44, 44, 46}, {58, 56, 52}}
						swc := seenWall[swv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(swc[0], swc[1], swc[2]))
					case TileRock:
						srv := tileVar(mx, my, 3)
						seenRock := [][3]int32{{55, 50, 46}, {46, 42, 40}, {60, 56, 50}}
						src := seenRock[srv]
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(src[0], src[1], src[2]))
					case TileWater:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(20, 35, 65))
					case TileUFOFloor:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(25, 50, 55))
					case TileUFOWall:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(30, 58, 64))
					case TileDoor:
						style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(55, 48, 18))
					case TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech:
						style = engine.StyleGray
					}
				}
			}

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

			tile = bs.Map.At(mx, my)
			if tile.Fire > 0 {
				fireFrame := (bs.FrameCount / 4) % 3
				switch fireFrame {
				case 0:
					ch = '^'
					style = tcell.StyleDefault.Foreground(color.XTerm11).Background(tcell.NewRGBColor(40, 20, 0))
				case 1:
					ch = 'w'
					style = tcell.StyleDefault.Foreground(color.Orange).Background(tcell.NewRGBColor(50, 15, 0))
				case 2:
					ch = '*'
					style = tcell.StyleDefault.Foreground(color.XTerm9).Background(tcell.NewRGBColor(30, 10, 0))
				}
			} else if tile.Blood > 0 {
				ch = bloodRunes[tile.Blood]
				switch tile.Blood {
				case 1:
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(140, 10, 10))
				case 2:
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(20, 180, 20))
				case 3:
					style = tcell.StyleDefault.Foreground(tcell.NewRGBColor(160, 30, 200))
				}
			}

			ctx.SetCell(x+1, y+1, ch, style)
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
					if bs.IsNight {
						engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 3, tcell.NewRGBColor(120, 110, 70))
					} else {
						engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 4, tcell.NewRGBColor(180, 160, 100))
					}
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
						if bs.IsNight {
							engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 1, tcell.NewRGBColor(60, 80, 150))
						} else {
							engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 2, tcell.NewRGBColor(100, 140, 255))
						}
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
		if u.Faction == 1 && !bs.Map.IsVisible(u.X, u.Y) {
			continue
		}
		ch := '\u263B' // ☻ human face
		style := engine.StyleCyanBold
		if u.Faction == 1 {
			ch = '\u03A9' // Ω alien (default)
			if u.AlienType != nil {
				ch = u.AlienType.Icon
			}
			style = engine.StyleRedBold
			if engine.Config.BloomEnabled {
				engine.ApplyBloom(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, tcell.NewRGBColor(255, 50, 50))
			}
		} else if u.Faction == 2 {
			ch = 'c'
			style = engine.StyleGreen
		}
		if u == bs.Selected {
			style = style.Reverse(true)
			if engine.Config.LightingEnabled {
				engine.ApplyDirectionalLight(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 0, -1, 5, tcell.NewRGBColor(200, 200, 150), func(x, y int) bool { return !bs.Map.Opaque(x+bs.ScrollX-1, y+bs.ScrollY-1) })
			}
		}
		if u.X == bs.CursorX && u.Y == bs.CursorY {
			style = style.Reverse(true)
		}
		ctx.SetCell(sx, sy, ch, style)
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

	// Draw sidebar border
	sidebarX := viewW + 2
	for y := 0; y < viewH; y++ {
		ctx.SetCell(sidebarX-1, y+1, '|', engine.StyleGray)
	}

	// If hovering an enemy, show ONLY target info in sidebar
	if bs.HoveredUnit != nil && bs.HoveredUnit != bs.Selected {
		sy := 1
		u := bs.HoveredUnit
		ctx.DrawString(sidebarX, sy, language.String("SIDE_TARGET_INFO"), engine.StyleRedBold)
		sy++
		name := ""
		if u.Faction == 1 && u.AlienType != nil {
			name = u.AlienType.Name
		} else if u.Faction == 0 && u.Soldier != nil {
			name = u.Soldier.Name
		} else if u.Faction == 2 {
			name = u.CivName
		}
		if len(name) > bs.SidebarW-1 {
			name = name[:bs.SidebarW-1]
		}
		ctx.DrawString(sidebarX, sy, name, engine.StyleDefault.Bold(true))
		sy++
		weaponName := data.RuleItems[u.Weapon].ShortName
		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_WPN_TARGET"), weaponName), engine.StyleDefault)
		sy++

		hasAutopsy := u.Faction != 1 || u.AlienType == nil
		if !hasAutopsy && u.AlienType != nil {
			for _, id := range bs.Base.CompletedResearch {
				if id == u.AlienType.AutopsyID {
					hasAutopsy = true
					break
				}
			}
		}

		if hasAutopsy {
			ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_HP"), u.HP, u.MaxHP), engine.StyleDefault)
			sy++
			ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_ACC"), u.Accuracy), engine.StyleDefault)
			sy++
			ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_STR_TU"), u.Strength, u.TU), engine.StyleDefault)
			sy++
		}

		if u.Faction == 1 && u.AlienType != nil {
			portrait := u.AlienType.GetPortrait()
			sy++
			for _, sl := range portrait.Lines {
				pl := sl.Content
				if len(pl) > bs.SidebarW-1 {
					pl = pl[:bs.SidebarW-1]
				}
				// Use the color from the StyledLine
				style := tcell.StyleDefault.Foreground(tcell.NewRGBColor(sl.Color[0], sl.Color[1], sl.Color[2]))
				ctx.DrawString(sidebarX, sy, pl, style)
				sy++
			}
		}
		return
	}

	// Draw unit info in sidebar
	sy := 1
	if bs.Selected != nil {
		ctx.DrawString(sidebarX, sy, language.String("SIDE_UNIT_INFO"), engine.StyleCyanBold)
		sy++

		name := bs.Selected.Soldier.Name
		if len(name) > bs.SidebarW-1 {
			name = name[:bs.SidebarW-1]
		}
		ctx.DrawString(sidebarX, sy, name, engine.StyleDefault.Bold(true))
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_HP"), bs.Selected.HP, bs.Selected.MaxHP), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_TU"), bs.Selected.TU, bs.Selected.MaxTU), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_ACC"), bs.Selected.Accuracy), engine.StyleDefault)
		sy++

		weaponName := data.RuleItems[bs.Selected.Weapon].ShortName
		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_WEAPON"), weaponName, bs.Selected.WeaponAmmo), engine.StyleDefault)
		sy++

		armourName := "None"
		if bs.Selected.Armour > 0 {
			for k, v := range data.Armors {
				if v.Undersuit == bs.Selected.Armour {
					armourName = k
					break
				}
			}
		}
		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_ARMOR"), armourName), engine.StyleDefault)
		sy++

		ctx.DrawString(sidebarX, sy, fmt.Sprintf(language.String("SIDE_POS"), bs.Selected.X, bs.Selected.Y), engine.StyleGray)
		sy++

		if bs.Selected.Crouching {
			ctx.DrawString(sidebarX, sy, language.String("SIDE_CROUCH"), engine.StyleYellow)
			sy++
		}
		sy++
	}

	// Draw log in sidebar
	logTitle := language.String("BATTLE_LOG")
	ctx.DrawString(sidebarX, sy, logTitle, engine.StyleCyanBold)
	sy++

	availableLines := viewH - sy
	logEntries := len(bs.Log)
	startIdx := 0
	if logEntries > availableLines {
		startIdx = logEntries - availableLines
	}
	for i := 0; i < availableLines && startIdx+i < logEntries; i++ {
		msg := bs.Log[startIdx+i]
		if len(msg) > bs.SidebarW-1 {
			msg = msg[:bs.SidebarW-1]
		}
		ctx.DrawString(sidebarX, sy+i, msg, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-4, w, 3, language.String("BATTLESCAPE"), engine.StyleDefault)
	lightStr := language.String("LIGHT_DAY")
	if bs.IsNight {
		lightStr = language.String("LIGHT_NIGHT")
	}
	turnStr := fmt.Sprintf(language.String("STATUS_TURN"), bs.Turn, bs.phaseStr()+" ("+lightStr+")")
	if bs.Map.NumLevels > 1 {
		turnStr += fmt.Sprintf(" [L%d]", bs.Map.CurrentLevel+1)
	}
	ctx.DrawString(2, h-3, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf(language.String("STATUS_SELECTED"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-3, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	cursorStr := fmt.Sprintf(language.String("STATUS_CURSOR"), bs.CursorX, bs.CursorY, tileTypeName(tile.Type))
	cursorX := w - len(cursorStr) - 2
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

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}

	// Draw help bar
	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_BATTLESCAPE")
	if bs.Map.NumLevels > 1 {
		help += language.String("HELP_STAIRS_SUFFIX")
	}
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)
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
		bs.AddMessage("NIGHT VISION ON")
	case engine.VisionNight:
		bs.VisionMode = engine.VisionThermal
		bs.AddMessage("THERMAL VISION ON")
	case engine.VisionThermal:
		bs.VisionMode = engine.VisionNormal
		bs.AddMessage("NORMAL VISION")
	}
}

func (bs *Battlescape) HandleMouse(e *tcell.EventMouse) {
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
	bs.Selected = humans[idx]
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
	}
	return language.String("TILE_UNKNOWN")
}

func (ul UnitList) Faction(f int) UnitList {
	var result UnitList
	for _, u := range ul {
		if u.Faction == f {
			result = append(result, u)
		}
	}
	return result
}

func tileVar(mx, my, n int) int {
	h := mx*2654435761 ^ my*2246822519
	if h < 0 {
		h = -h
	}
	return h % n
}

func (bs *Battlescape) ApplyCursorStyles(x, y int, style tcell.Style) tcell.Style {
	bs.State.mu.RLock()
	defer bs.State.mu.RUnlock()

	if x == bs.CursorX && y == bs.CursorY {
		switch bs.State.CursorState {
		case StateInspect:
			return style.Background(color.White).Foreground(color.Black)
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
