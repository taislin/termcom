package battle

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/civ13/ycom/internal/audio"
	"github.com/gdamore/tcell/v3"
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

const sidebarW = 24

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

	Camera   *engine.Camera
	Particles *engine.ParticleSystem
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

func NewBattlescape(g *engine.Game, squad []*soldier.Soldier, ufoName string) *Battlescape {
	var m *BattleMap
	switch ufoName {
	case "Terror":
		m = GenerateTerrorSite(50, 50)
	case "Supply":
		m = GenerateUFOInterior(50, 50)
	case "Alien Base":
		m = GenerateCydonia(50, 50)
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
		Map:     m,
		Phase:   PhasePlayerTurn,
		Turn:    1,
		CursorX: m.Width / 2,
		CursorY: m.Height / 2,
		Squad:   squad,
		UFOName: ufoName,
		IsNight: g.GameTime.Hour() < 6 || g.GameTime.Hour() > 18,
		Camera:   engine.NewCamera(m.Width/2, m.Height/2),
		Particles: engine.NewParticleSystem(512),
	}

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
		if u.Faction == 0 && u.Alive {
			bs.Map.ComputeFOV(u.X, u.Y, sightRange)
		}
	}
	for _, u := range bs.Units {
		if u.Faction == 1 && u.Alive && u.AlienType != nil && bs.Map.IsVisible(u.X, u.Y) {
			bs.Game.LearnAlien(u.AlienType.Name, 1)
		}
	}
}

func (bs *Battlescape) Update() {
	dt := 0.016
	bs.Camera.UpdateShake(dt)
	bs.Particles.Update(dt)

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
		engine.SpawnExplosion(bs.Particles, action.Unit.X-bs.ScrollX+1, action.Unit.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 200, 50), 4)
		if hit {
			engine.SpawnExplosion(bs.Particles, action.Target.X-bs.ScrollX+1, action.Target.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
			bs.Camera.TriggerShake(0.5)
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
		damage := action.Unit.Strength + rand.Intn(10)
		damage -= action.Target.Armour
		if damage < 1 {
			damage = 1
		}
		action.Target.HP -= damage
		if action.Target.HP <= 0 {
			action.Target.Alive = false
		}
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
	case "patrol":
		action.Unit.MoveTo(action.ToX, action.ToY, bs.Map)
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
	bs.checkVictory()
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

	scrW, scrH := bs.Game.ScreenSize()
	viewW := scrW - sidebarW - 2
	viewH := scrH - 5
	if viewW < 10 {
		viewW = 10
	}
	if bs.CursorX < bs.ScrollX+2 {
		bs.ScrollX = bs.CursorX - 2
	}
	if bs.CursorX > bs.ScrollX+viewW-3 {
		bs.ScrollX = bs.CursorX - viewW + 3
	}
	if bs.CursorY < bs.ScrollY+2 {
		bs.ScrollY = bs.CursorY - 2
	}
	if bs.CursorY > bs.ScrollY+viewH-3 {
		bs.ScrollY = bs.CursorY - viewH + 3
	}
	if bs.ScrollX < 0 {
		bs.ScrollX = 0
	}
	if bs.ScrollY < 0 {
		bs.ScrollY = 0
	}
	maxScrollX := bs.Map.Width - viewW
	if maxScrollX < 0 {
		maxScrollX = 0
	}
	maxScrollY := bs.Map.Height - viewH
	if maxScrollY < 0 {
		maxScrollY = 0
	}
	if bs.ScrollX > maxScrollX {
		bs.ScrollX = maxScrollX
	}
	if bs.ScrollY > maxScrollY {
		bs.ScrollY = maxScrollY
	}
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

func (bs *Battlescape) Confirm() {
	if bs.Phase != PhasePlayerTurn {
		return
	}
	unit := bs.Units.At(bs.CursorX, bs.CursorY)

	// Click on friendly unit → select it
	if unit != nil && unit.Faction == 0 && unit.Alive && unit.Soldier != nil {
		bs.Selected = unit
		bs.AddMessage(fmt.Sprintf(language.String("MSG_UNIT_SELECTED"), unit.Soldier.Name, unit.HP, unit.TU))
		return
	}

	// Click on enemy → fire at it (if selected soldier can see it)
	if unit != nil && unit.Faction == 1 && unit.Alive {
		if bs.Selected == nil {
			bs.AddMessage(language.String("MSG_NO_SOLDIER_SELECTED"))
			return
		}
		if !bs.Selected.CanSee(unit.X, unit.Y, bs.Map) {
			bs.AddMessage(language.String("MSG_TARGET_NO_LOS"))
			return
		}
		damage, hit, err := bs.Selected.FireAt(unit, bs.Map)
		if err != nil {
			bs.AddMessage(err.Error())
			return
		}
		if hit {
			audio.PlayHit()
			name := "alien"
			if unit.AlienType != nil {
				name = unit.AlienType.Name
			}
			bs.AddMessage(fmt.Sprintf(language.String("MSG_HIT_TARGET"), damage, name, unit.HP))
		} else {
			audio.PlayMiss()
			bs.AddMessage(language.String("MSG_MISSED"))
		}
		return
	}

	// Click on empty ground → move selected soldier there
	if unit == nil {
		if bs.Selected == nil {
			bs.AddMessage(language.String("MSG_NO_SOLDIER_SELECTED"))
			return
		}
		if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
			audio.PlayMove()
			bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY))
			bs.ComputeFOVForTeam()
		} else {
			bs.AddMessage(language.String("MSG_CANNOT_MOVE"))
		}
		return
	}
}

func (bs *Battlescape) MoveSelected() {
	if bs.Selected == nil || bs.Phase != PhasePlayerTurn {
		return
	}
	if bs.Selected.MoveTo(bs.CursorX, bs.CursorY, bs.Map) {
		audio.PlayMove()
		bs.AddMessage(fmt.Sprintf(language.String("MSG_MOVED"), bs.Selected.Soldier.Name, bs.CursorX, bs.CursorY))
		bs.ComputeFOVForTeam()
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
	engine.SpawnExplosion(bs.Particles, bs.Selected.X-bs.ScrollX+1, bs.Selected.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 200, 50), 4)
	if hit {
		audio.PlayHit()
		engine.SpawnExplosion(bs.Particles, target.X-bs.ScrollX+1, target.Y-bs.ScrollY+1, tcell.NewRGBColor(255, 80, 30), 8)
		bs.Camera.TriggerShake(0.5)
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
		}
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
	viewW := w - sidebarW - 2
	if viewW < 10 {
		viewW = 10
	}
	viewH := h - 5

	camX, camY := bs.Camera.Pos()
	scrollX := camX - viewW/2
	scrollY := camY - viewH/2
	if scrollX < 0 {
		scrollX = 0
	}
	if scrollY < 0 {
		scrollY = 0
	}
	bs.ScrollX = scrollX
	bs.ScrollY = scrollY

	for y := 0; y < viewH; y++ {
		for x := 0; x < viewW; x++ {
			mx := x + bs.ScrollX
			my := y + bs.ScrollY
			tile := bs.Map.At(mx, my)

			if !tile.Seen {
				ctx.SetCell(x+1, y+1, ' ', engine.StyleDefault)
				continue
			}

			ch := TileChar(tile.Type)
			style := engine.StyleGreen

			if tile.Visible {
				switch tile.Type {
				case TileGrass:
					style = engine.StyleGreen
				case TileWall:
					style = engine.StyleGray
				case TileDoor:
					style = engine.StyleYellow
				case TileTree:
					style = engine.StyleGreen
				case TileRock:
					style = engine.StyleGray
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
				switch tile.Type {
				case TileGrass:
					style = engine.StyleGray
				case TileWall:
					style = engine.StyleGray
				case TileDoor:
					style = engine.StyleGray
				case TileTree:
					style = engine.StyleGray
				case TileRock:
					style = engine.StyleGray
				case TileWater:
					style = engine.StyleGray
				case TileUFOFloor:
					style = engine.StyleGray
				case TileUFOWall:
					style = engine.StyleGray
				case TileConsole, TileMachinery, TilePod, TilePowerSource, TileStorage, TileAlienTech:
					style = engine.StyleGray
				}
			}

			if bs.Selected != nil && bs.Phase == PhasePlayerTurn {
				movementRange := bs.GetMovementRange()
				if movementRange[[2]int{mx, my}] {
					style = style.Background(tcell.NewRGBColor(20, 40, 80))
				}
			}

			if mx == bs.CursorX && my == bs.CursorY {
				style = style.Reverse(true)
			}

			ctx.SetCell(x+1, y+1, ch, style)
		}
	}

	if bs.IsNight {
		for _, u := range bs.Units {
			if !u.Alive || u.Faction != 0 {
				continue
			}
			sx := u.X - bs.ScrollX + 1
			sy := u.Y - bs.ScrollY + 1
			if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
				engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 4, tcell.NewRGBColor(180, 160, 100))
			}
		}
		for _, u := range bs.Units {
			if !u.Alive || u.Faction != 1 {
				continue
			}
			sx := u.X - bs.ScrollX + 1
			sy := u.Y - bs.ScrollY + 1
			if sx >= 1 && sx < viewW+1 && sy >= 1 && sy < viewH+1 {
				engine.ApplyLightSource(ctx.ScreenRaw, ctx.FrameBuffer(), sx, sy, 2, tcell.NewRGBColor(100, 140, 255))
			}
		}
	}

	for _, u := range bs.Units {
		if !u.Alive {
			continue
		}
		sx := u.X - bs.ScrollX + 1
		sy := u.Y - bs.ScrollY + 1
		if sx < 1 || sx >= viewW+1 || sy < 1 || sy < viewH+1 {
			if u.Faction == 1 && !bs.Map.IsVisible(u.X, u.Y) {
				continue
			}
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
		} else if u.Faction == 2 {
			ch = 'c'
			style = engine.StyleGreen
		}
		if u == bs.Selected {
			style = style.Reverse(true)
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
		}
	}

	bs.Particles.Draw(ctx.ScreenRaw)

	// Draw sidebar border
	sidebarX := viewW + 2
	for y := 0; y < viewH; y++ {
		ctx.SetCell(sidebarX-1, y+1, '|', engine.StyleGray)
	}

	// Draw unit info in sidebar
	sy := 1
	if bs.Selected != nil {
		ctx.DrawString(sidebarX, sy, language.String("SIDE_UNIT_INFO"), engine.StyleCyanBold)
		sy++

		name := bs.Selected.Soldier.Name
		if len(name) > sidebarW-1 {
			name = name[:sidebarW-1]
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
		if len(msg) > sidebarW-1 {
			msg = msg[:sidebarW-1]
		}
		ctx.DrawString(sidebarX, sy+i, msg, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-4, w, 3, language.String("BATTLESCAPE"), engine.StyleDefault)
	lightStr := language.String("LIGHT_DAY")
	if bs.IsNight {
		lightStr = language.String("LIGHT_NIGHT")
	}
	turnStr := fmt.Sprintf(language.String("STATUS_TURN"), bs.Turn, bs.phaseStr()+" ("+lightStr+")")
	ctx.DrawString(2, h-3, turnStr, engine.StyleDefault)

	if bs.Selected != nil {
		selStr := fmt.Sprintf(language.String("STATUS_SELECTED"),
			bs.Selected.Soldier.Name, bs.Selected.HP, bs.Selected.MaxHP,
			bs.Selected.TU, bs.Selected.MaxTU, data.RuleItems[bs.Selected.Weapon].ShortName)
		ctx.DrawString(w/2, h-3, selStr, engine.StyleCyan)
	}

	tile := bs.Map.At(bs.CursorX, bs.CursorY)
	cursorStr := fmt.Sprintf(language.String("STATUS_CURSOR"), bs.CursorX, bs.CursorY, tileTypeName(tile.Type))
	ctx.DrawString(w-30, h-3, cursorStr, engine.StyleGray)

	if bs.Message != "" {
		ctx.DrawString(2, h-2, bs.Message, engine.StyleYellow)
	}

	// Draw help bar
	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := "[hjkl]/[WSAD]=Move [Space]/[Enter]=Act [q]=Cycle [f]=Fire [r]=Reload [g]=Grenade [m]=Medikit [e]=End [c]=Crouch"
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
	if bs.Phase == PhaseVictory || bs.Phase == PhaseDefeat {
		return
	}
	switch e.Key() {
	case tcell.KeyUp:
		bs.MoveCursor(0, -1)
	case tcell.KeyDown:
		bs.MoveCursor(0, 1)
	case tcell.KeyLeft:
		bs.MoveCursor(-1, 0)
	case tcell.KeyRight:
		bs.MoveCursor(1, 0)
	case tcell.KeyEnter:
		bs.Confirm()
	}
	switch e.Str() {
	case " ":
		bs.Confirm()
	case "f", "F":
		bs.FireWeapon()
	case "r", "R":
		bs.Reload()
	case "e", "E":
		bs.EndTurn()
	case "h", "H":
		bs.MoveCursor(-1, 0)
	case "j", "J":
		bs.MoveCursor(0, 1)
	case "k", "K":
		bs.MoveCursor(0, -1)
	case "l", "L":
		bs.MoveCursor(1, 0)
	case "q", "Q":
		bs.cycleUnit(1)
	case "w", "W":
		bs.MoveCursor(0, -1)
	case "a", "A":
		bs.MoveCursor(-1, 0)
	case "s", "S":
		bs.MoveCursor(0, 1)
	case "d", "D":
		bs.MoveCursor(1, 0)
	case "c", "C":
		bs.Crouch()
	case "g", "G":
		bs.Grenade()
	case "m", "M":
		bs.UseMedikit()
	case "p", "P":
		bs.PsiAttack()
	case "n", "N":
		bs.EndTurn()
	case ".":
		bs.MoveSelected()
	}
}

func (bs *Battlescape) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, scrH := bs.Game.ScreenSize()

	// Handle help bar clicks (bottom bar)
	if y == scrH-1 {
		// Help bar: " hjkl/WSAD=Move Space/Enter=Act q=Cycle f=Fire r=Reload g=Grenade m=Medikit e=End c=Crouch"
		switch {
		case x >= 22 && x <= 25: // q=Cycle
			bs.cycleUnit(1)
		case x >= 27 && x <= 31: // f=Fire
			bs.FireWeapon()
		case x >= 33 && x <= 39: // r=Reload
			bs.Reload()
		case x >= 41 && x <= 51: // g=Grenade
			bs.Grenade()
		case x >= 53 && x <= 63: // m=Medikit
			bs.UseMedikit()
		case x >= 65 && x <= 69: // e=End
			bs.EndTurn()
		case x >= 71 && x <= 77: // c=Crouch
			bs.Crouch()
		}
		return
	}

	// Don't process clicks on the sidebar
	scrW, _ := bs.Game.ScreenSize()
	viewW := scrW - sidebarW - 2
	if x >= viewW+2 {
		return
	}

	mx := x - 1 + bs.ScrollX
	my := y - 1 + bs.ScrollY
	if mx >= 0 && mx < bs.Map.Width && my >= 0 && my < bs.Map.Height {
		bs.CursorX = mx
		bs.CursorY = my
		if buttons&tcell.Button1 != 0 {
			bs.Confirm()
		}
		if buttons&tcell.Button3 != 0 {
			bs.FireWeapon()
		}
	}

	if buttons&tcell.WheelUp != 0 {
		bs.cycleUnit(1)
	}
	if buttons&tcell.WheelDown != 0 {
		bs.cycleUnit(-1)
	}
}

func (bs *Battlescape) cycleUnit(dir int) {
	humans := bs.Units.Faction(0).Alive()
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

	// Center screen on the soldier
	scrW, scrH := bs.Game.ScreenSize()
	viewW := scrW - sidebarW - 2
	viewH := scrH - 5
	if viewW < 10 {
		viewW = 10
	}
	bs.ScrollX = bs.Selected.X - viewW/2
	bs.ScrollY = bs.Selected.Y - viewH/2

	// Clamp scroll to map bounds
	if bs.ScrollX < 0 {
		bs.ScrollX = 0
	}
	if bs.ScrollY < 0 {
		bs.ScrollY = 0
	}
	maxScrollX := bs.Map.Width - viewW
	if maxScrollX < 0 {
		maxScrollX = 0
	}
	maxScrollY := bs.Map.Height - viewH
	if maxScrollY < 0 {
		maxScrollY = 0
	}
	if bs.ScrollX > maxScrollX {
		bs.ScrollX = maxScrollX
	}
	if bs.ScrollY > maxScrollY {
		bs.ScrollY = maxScrollY
	}

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
