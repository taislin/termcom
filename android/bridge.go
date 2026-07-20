//go:build android

package android

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/geo"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/mapgen"
	"github.com/taislin/termcom/internal/save"
	"github.com/taislin/termcom/internal/soldier"
)

type customBattleDef struct {
	Name      string `json:"name"`
	Night     bool   `json:"night"`
	Map       struct {
		Generator string `json:"generator"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"map"`
	Soldiers  []struct {
		Name      string `json:"name"`
		Rank      int    `json:"rank"`
		HP        int    `json:"hp"`
		TU        int    `json:"tu"`
		Accuracy  int    `json:"accuracy"`
		Reactions int    `json:"reactions"`
		Strength  int    `json:"strength"`
		Weapon    string `json:"weapon"`
		Armor     string `json:"armor"`
		X         int    `json:"x"`
		Y         int    `json:"y"`
	} `json:"soldiers"`
	Aliens    []struct {
		Name       string `json:"name"`
		HP         int    `json:"hp"`
		TU         int    `json:"tu"`
		Accuracy   int    `json:"accuracy"`
		Bravery    int    `json:"bravery"`
		Reactions  int    `json:"reactions"`
		Strength   int    `json:"strength"`
		Psi        int    `json:"psi"`
		Armour     int    `json:"armour"`
		Weapon     string `json:"weapon"`
		Rank       int    `json:"rank"`
		DamageType int    `json:"damage_type"`
		Aggression int    `json:"aggression"`
		X          int    `json:"x"`
		Y          int    `json:"y"`
	} `json:"aliens"`
	Civilians []struct {
		Name string `json:"name"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	} `json:"civilians"`
	Victory struct {
		Condition   string `json:"condition"`
		Turns       int    `json:"turns"`
		TargetX     int    `json:"target_x"`
		TargetY     int    `json:"target_y"`
		MinSoldiers int    `json:"min_soldiers"`
	} `json:"victory"`
}

// GameBridge is the exported interface for gomobile bind.
// Java calls these methods after loading the shared library.
type GameBridge struct {
	game    *engine.Game
	as      *androidScreen
	cols    int
	rows    int
	running bool
	done    chan struct{}
}

// NewGame creates a new GameBridge. dataDir is the Android internal storage path
// for config and save files. cols and rows are the initial terminal grid size.
func NewGame(dataDir string, cols, rows int) *GameBridge {
	if dataDir != "" {
		if err := os.Chdir(dataDir); err != nil {
			// Fall back to CWD if dataDir is invalid
		}
	}

	if err := mapgen.Init(); err != nil {
		// Non-fatal: log and continue with hardcoded defaults
		_ = err
	}
	data.NewAlienSpriteRegistry().RebuildFromTemplates(
		mapgen.ToTemplateData("head"),
		mapgen.ToTemplateData("eye"),
		mapgen.ToTemplateData("torso"),
		mapgen.ToTemplateData("leg"),
		mapgen.ToTemplateData("weapon"),
	)

	engine.LoadConfig()
	engine.Config.TouchMode = true
	engine.HideTouchOverlay = true

	as := newAndroidScreen(cols, rows)
	scr := engine.NewScreenRawWithScreen(as, cols, rows)

	initialState := engine.StateMenu
	if _, err := os.Stat(engine.ConfigFile); os.IsNotExist(err) {
		initialState = engine.StateLanguageSelect
	}

	g := engine.NewGameWithScreen(scr, initialState)

	bridge := &GameBridge{
		game:    g,
		as:      as,
		cols:    cols,
		rows:    rows,
		done:    make(chan struct{}),
	}

	// Wire game callbacks
	g.OnNewGame = func() {
		picker := engine.NewDifficultyScreen(g, func(difficulty int) {
			gs := geo.NewGeoscape(g)
			g.RegisterScreen(engine.StateGeoscape, gs)
			g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
			g.SetState(engine.StateGeoscape)
		})
		g.PushScreen(picker)
	}

	g.OnContinue = func() {
		sd, err := save.LoadGame(engine.SaveFile)
		if err != nil {
			return
		}
		gs := geo.NewGeoscapeFromSave(g, sd)
		g.RegisterScreen(engine.StateGeoscape, gs)
		g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
		g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
		g.SetState(engine.StateGeoscape)
	}

	g.OnLoadGame = func() {
		var slots []engine.SlotInfo
		for slot := 1; slot <= 10; slot++ {
			sd, err := save.LoadGame(save.SavePath(slot))
			if err != nil {
				continue
			}
			label := engine.FormatSlotLabel(slot, sd.GameTime.Format("2006 Jan 02"), sd.Funds)
			slots = append(slots, engine.SlotInfo{Slot: slot, Label: label})
		}
		picker := engine.NewSlotPickerScreen(g, engine.SlotPickerLoad, slots, func(slot int) {
			sd, err := save.LoadGame(save.SavePath(slot))
			if err != nil {
				return
			}
			gs := geo.NewGeoscapeFromSave(g, sd)
			g.RegisterScreen(engine.StateGeoscape, gs)
			g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
			g.SetState(engine.StateGeoscape)
		})
		g.PushScreen(picker)
	}

	g.OnCustomBattle = func() {
		screen := engine.NewCustomBattleScreen(g, func(entry engine.CustomBattleEntry) {
			bridge.launchCustomBattle(entry.FilePath)
		})
		g.PushScreen(screen)
	}

	g.RegisterScreen(engine.StateHelp, engine.NewHelpScreen(g, engine.StateGeoscape))
	g.RegisterScreen(engine.StateMenu, engine.NewMenuScreen(g))
	g.RegisterScreen(engine.StateLanguageSelect, engine.NewLanguageSelectScreen(g))

	return bridge
}

// Start runs the game loop in a background goroutine.
func (b *GameBridge) Start() {
	if b.running {
		return
	}
	b.running = true
	go func() {
		b.game.Run()
		close(b.done)
	}()
}

// Stop signals the game loop to exit.
func (b *GameBridge) Stop() {
	if !b.running {
		return
	}
	b.running = false
	b.game.Stop()
	<-b.done
}

// Resize updates the terminal grid dimensions.
func (b *GameBridge) Resize(cols, rows int) {
	if cols <= 0 || rows <= 0 {
		return
	}
	b.cols = cols
	b.rows = rows
	b.game.InjectResize(cols, rows)
}

// InjectTouch sends a mouse event derived from a touch gesture.
func (b *GameBridge) InjectTouch(x, y int, action string) {
	var btn tcell.ButtonMask
	switch action {
	case "long_press":
		btn = tcell.Button2
	case "scroll_up":
		btn = tcell.WheelUp
	case "scroll_down":
		btn = tcell.WheelDown
	default:
		btn = tcell.Button1
	}
	ev := tcell.NewEventMouse(x, y, btn, tcell.ModNone)
	b.game.InjectMouse(ev)
}

// InjectKey sends a keyboard event.
func (b *GameBridge) InjectKey(key string) {
	if len(key) == 0 {
		return
	}
	if len(key) == 1 {
		ch := key[0]
		switch ch {
		case '\r', '\n':
			b.game.InjectKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
		case '\x1b':
			b.game.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
		case '\b':
			b.game.InjectKey(tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone))
		case '\t':
			b.game.InjectKey(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone))
		default:
			b.game.InjectKey(tcell.NewEventKey(tcell.KeyRune, string(ch), tcell.ModNone))
		}
		return
	}
	switch key {
	case "escape":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	case "enter":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	case "space":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	case "backspace":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone))
	case "tab":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone))
	case "up":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyUp, "", tcell.ModNone))
	case "down":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone))
	case "left":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	case "right":
		b.game.InjectKey(tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone))
	case "f1", "F1", "f2", "F2", "f3", "F3", "f4", "F4", "f5", "F5", "f6", "F6",
		"f7", "F7", "f8", "F8", "f9", "F9", "f10", "F10", "f11", "F11", "f12", "F12":
		k := tcell.KeyF1 + tcell.Key(keyCode(key))
		b.game.InjectKey(tcell.NewEventKey(k, "", tcell.ModNone))
	}
}

func keyCode(name string) int {
	switch name {
	case "f1", "F1":
		return 0
	case "f2", "F2":
		return 1
	case "f3", "F3":
		return 2
	case "f4", "F4":
		return 3
	case "f5", "F5":
		return 4
	case "f6", "F6":
		return 5
	case "f7", "F7":
		return 6
	case "f8", "F8":
		return 7
	case "f9", "F9":
		return 8
	case "f10", "F10":
		return 9
	case "f11", "F11":
		return 10
	case "f12", "F12":
		return 11
	}
	return 0
}

// FrameWidth returns the current framebuffer width in cells.
func (b *GameBridge) FrameWidth() int {
	return b.game.WebScreen().FrameBuffer().Width()
}

// FrameHeight returns the current framebuffer height in cells.
func (b *GameBridge) FrameHeight() int {
	return b.game.WebScreen().FrameBuffer().Height()
}

// FrameData returns the current frame as a flat byte array.
// 8 bytes per cell [rune_lo, rune_hi, fg_r, fg_g, fg_b, bg_r, bg_g, attr].
func (b *GameBridge) FrameData() []byte {
	return b.game.WebScreen().FrameBuffer().MarshalBinary()
}

// SetLanguage changes the game language by key (e.g. "en", "zh").
func (b *GameBridge) SetLanguage(lang string) {
	language.SetLanguage(lang)
	engine.Config.Language = lang
}

// FrameListener is called when a new frame is ready to be drawn.
type FrameListener interface {
	OnFrameReady()
}

// SetFrameListener registers a callback to be invoked on every frame flush.
func (b *GameBridge) SetFrameListener(l FrameListener) {
	if b.as != nil {
		b.as.onShow = func() {
			if l != nil {
				l.OnFrameReady()
			}
		}
	}
}

// GetButtonsJSON returns the current active control buttons as a JSON string.
func (b *GameBridge) GetButtonsJSON() string {
	type JavaButton struct {
		Label   string `json:"label"`
		Enabled bool   `json:"enabled"`
		Index   int    `json:"index"`
	}
	var res []JavaButton
	for i, btn := range engine.Menu.Buttons {
		res = append(res, JavaButton{
			Label:   btn.Label,
			Enabled: btn.Enabled,
			Index:   i,
		})
	}
	data, err := json.Marshal(res)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// ClickButton triggers the action of the button at the specified index.
func (b *GameBridge) ClickButton(index int) {
	if index >= 0 && index < len(engine.Menu.Buttons) {
		btn := engine.Menu.Buttons[index]
		if btn.Enabled && btn.Action != nil {
			btn.Action()
		}
	}
}

func (b *GameBridge) launchCustomBattle(path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var def customBattleDef
	if err := json.Unmarshal(raw, &def); err != nil {
		return
	}

	if def.Night {
		b.game.GameTime = time.Date(1999, time.March, 1, 2, 0, 0, 0, time.UTC)
	} else {
		b.game.GameTime = time.Date(1999, time.March, 1, 12, 0, 0, 0, time.UTC)
	}
	b.game.Paused = true

	baseInst := base.NewBase("Test Base", 0)
	baseInst.Facilities = append(baseInst.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	baseInst.Facilities = append(baseInst.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
	baseInst.Facilities = append(baseInst.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
	baseInst.Facilities = append(baseInst.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
	baseInst.Facilities = append(baseInst.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})

	var squad []*soldier.Soldier
	for _, cs := range def.Soldiers {
		s := soldier.NewSoldier(cs.Name)
		s.Rank = soldier.Rank(cs.Rank)
		if cs.HP > 0 {
			s.HP = cs.HP
			s.MaxHP = cs.HP
		}
		if cs.TU > 0 {
			s.TU = cs.TU
			s.MaxTU = cs.TU
		}
		if cs.Accuracy > 0 {
			s.Accuracy = cs.Accuracy
		}
		if cs.Reactions > 0 {
			s.Reactions = cs.Reactions
		}
		if cs.Strength > 0 {
			s.Strength = cs.Strength
		}
		if cs.Weapon != "" {
			s.Weapon = cs.Weapon
			s.WeaponAmmo = data.RuleItems[cs.Weapon].AmmoMax
		}
		if cs.Armor != "" {
			s.Armor = cs.Armor
		}
		squad = append(squad, s)
	}
	if len(squad) == 0 {
		for _, name := range []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot"} {
			s := soldier.NewSoldier(name)
			s.Rank = soldier.Sergeant
			s.Accuracy = 70
			s.Reactions = 60
			s.Strength = 20
			s.HP = 30
			s.MaxHP = 30
			s.TU = 55
			s.MaxTU = 55
			s.Weapon = "rifle"
			s.WeaponAmmo = data.RuleItems["rifle"].AmmoMax
			s.Armor = "personal"
			squad = append(squad, s)
		}
	}
	baseInst.Soldiers = squad

	var m *battle.BattleMap
	gen := def.Map.Generator
	w, h := def.Map.Width, def.Map.Height
	if w <= 0 || h <= 0 {
		w, h = 50, 50
	}
	switch gen {
	case "terror":
		m = battle.GenerateTerrorSite(w, h, time.Now().UnixNano())
	case "supply_raid", "ufo_interior":
		m = battle.GenerateUFOInterior(w, h, time.Now().UnixNano())
	case "alien_base":
		m = battle.GenerateAlienBase(w, h, time.Now().UnixNano())
	case "alien_research":
		m = battle.GenerateUFOInterior(w, h, time.Now().UnixNano())
	case "council":
		m = battle.GenerateTerrorSite(w, h, time.Now().UnixNano())
	case "cydonia":
		m = battle.GenerateCydonia(w, h, time.Now().UnixNano())
	case "abduction":
		m = battle.GenerateAbductionSite(w, h)
	case "forest":
		m = battle.GenerateForest(w, h)
	case "desert":
		m = battle.GenerateDesert(w, h)
	case "polar":
		m = battle.GeneratePolar(w, h)
	case "farm":
		m = battle.GenerateFarm(w, h)
	case "coastal":
		m = battle.GenerateCoastal(w, h)
	case "mountain":
		m = battle.GenerateMountain(w, h)
	case "swamp":
		m = battle.GenerateSwamp(w, h)
	case "jungle":
		m = battle.GenerateJungle(w, h)
	default:
		m, _ = battle.GenerateCrashSite(w, h, 42, -1, -1)
	}

	var units []battle.CustomUnitDef
	for _, cs := range def.Soldiers {
		units = append(units, battle.CustomUnitDef{
			Name:      cs.Name,
			HP:        cs.HP,
			TU:        cs.TU,
			Accuracy:  cs.Accuracy,
			Reactions: cs.Reactions,
			Strength:  cs.Strength,
			Weapon:    cs.Weapon,
			Armor:     cs.Armor,
			Faction:   0,
			X:         cs.X,
			Y:         cs.Y,
		})
	}
	for _, ca := range def.Aliens {
		units = append(units, battle.CustomUnitDef{
			Name:       ca.Name,
			HP:         ca.HP,
			TU:         ca.TU,
			Accuracy:   ca.Accuracy,
			Bravery:    ca.Bravery,
			Reactions:  ca.Reactions,
			Strength:   ca.Strength,
			Psi:        ca.Psi,
			Armour:     ca.Armour,
			Weapon:     ca.Weapon,
			Rank:       ca.Rank,
			DamageType: ca.DamageType,
			Aggression: ca.Aggression,
			Faction:    1,
			X:          ca.X,
			Y:          ca.Y,
		})
	}
	for _, cc := range def.Civilians {
		units = append(units, battle.CustomUnitDef{
			Name:    cc.Name,
			Faction: 2,
			X:       cc.X,
			Y:       cc.Y,
		})
	}

	var cv *battle.CustomVictory
	if def.Victory.Condition != "" {
		cv = &battle.CustomVictory{
			Condition:   def.Victory.Condition,
			Turns:       def.Victory.Turns,
			TargetX:     def.Victory.TargetX,
			TargetY:     def.Victory.TargetY,
			MinSoldiers: def.Victory.MinSoldiers,
		}
	}

	bs := battle.NewCustomBattlescape(b.game, baseInst, squad, m, units, cv, def.Name)
	b.game.SetScreen(engine.StateBattlescape, bs)
	b.game.SetState(engine.StateBattlescape)
}
