package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/geo"
	"github.com/taislin/termcom/internal/save"
	"github.com/taislin/termcom/internal/soldier"
	"github.com/taislin/termcom/web"
)

const (
	defaultCols = 220
	defaultRows = 50
)

func main() {
	addr := ":8080"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	log.Printf("TERMCOM Web Server starting on %s", addr)
	log.Printf("Open http://localhost%s in your browser", addr)

	g, ns, err := engine.NewGameWeb(defaultCols, defaultRows)
	if err != nil {
		log.Fatalf("Failed to init game: %v", err)
	}

	wireGame(g)

	srv := web.StartServer(addr)

	wr := engine.NewWebRenderer()

	// Handle synchronous updates exactly when the game engine flushes/shows the screen.
	ns.OnShow = func() {
		ansi := wr.Render(g.WebScreen())
		if ansi != "" {
			srv.SendOutput(ansi)
		}
	}

	// Forward keyboard input from browser to the game.
	web.InputHandler = func(input string) {
		ev := parseKeyEvent(input)
		if ev != nil {
			g.InjectKey(ev)
		}
	}

	// Forward resize events from browser to the game.
	web.ResizeHandler = func(cols, rows int) {
		if cols > 0 && rows > 0 {
			g.InjectResize(cols, rows)
			wr.ForceRepaint()
		}
	}

	// Run the game loop in a goroutine (it blocks until quit).
	go g.Run()

	// Notify the initial connected client about current screen size.
	web.ConnectHandler = func(cols, rows int) {
		wr.ForceRepaint()
		if cols > 0 && rows > 0 {
			g.InjectResize(cols, rows)
		} else {
			g.InjectResize(defaultCols, defaultRows)
		}
	}

	fmt.Println("Press Ctrl+C to stop")
	select {}
}

// wireGame sets up OnNewGame, OnContinue, etc. — same logic as cmd/termcom/main.go.
func wireGame(g *engine.Game) {
	g.OnNewGame = func() {
		picker := engine.NewDifficultyScreen(g, func(difficulty int) {
			gs := geo.NewGeoscape(g)
			g.RegisterScreen(engine.StateGeoscape, gs)
			g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
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
			g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
			g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
			g.SetState(engine.StateGeoscape)
		})
		g.PushScreen(picker)
	}

	g.OnCustomBattle = func() {
		screen := engine.NewCustomBattleScreen(g, func(entry engine.CustomBattleEntry) {
			launchCustomBattle(g, entry.FilePath)
		})
		g.PushScreen(screen)
	}

	g.RegisterScreen(engine.StateHelp, engine.NewHelpScreen(g))
	g.RegisterScreen(engine.StateMenu, engine.NewMenuScreen(g))
}

// parseKeyEvent converts an xterm.js key string into a tcell.EventKey.
// xterm.js sends the raw bytes that would have been sent over a PTY:
// printable chars as-is, special keys as ANSI escape sequences.
func parseKeyEvent(s string) *tcell.EventKey {
	if len(s) == 0 {
		return nil
	}
	// Single printable rune
	if len(s) == 1 {
		ch := rune(s[0])
		switch ch {
		case '\r', '\n':
			return tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)
		case '\x1b':
			return tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
		case '\x08', '\x7f':
			return tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone)
		case '\t':
			return tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)
		default:
			if ch >= 1 && ch <= 26 {
				// Ctrl+A..Z
				key := tcell.Key(int(tcell.KeyCtrlA) + int(ch) - 1)
				return tcell.NewEventKey(key, "", tcell.ModCtrl)
			}
			return tcell.NewEventKey(tcell.KeyRune, string(ch), tcell.ModNone)
		}
	}
	// ANSI escape sequences
	if s[0] == '\x1b' {
		switch s {
		case "\x1b[A", "\x1bOA":
			return tcell.NewEventKey(tcell.KeyUp, "", tcell.ModNone)
		case "\x1b[B", "\x1bOB":
			return tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone)
		case "\x1b[C", "\x1bOC":
			return tcell.NewEventKey(tcell.KeyRight, "", tcell.ModNone)
		case "\x1b[D", "\x1bOD":
			return tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone)
		case "\x1b[H", "\x1bOH":
			return tcell.NewEventKey(tcell.KeyHome, "", tcell.ModNone)
		case "\x1b[F", "\x1bOF":
			return tcell.NewEventKey(tcell.KeyEnd, "", tcell.ModNone)
		case "\x1b[5~":
			return tcell.NewEventKey(tcell.KeyPgUp, "", tcell.ModNone)
		case "\x1b[6~":
			return tcell.NewEventKey(tcell.KeyPgDn, "", tcell.ModNone)
		case "\x1b[2~":
			return tcell.NewEventKey(tcell.KeyInsert, "", tcell.ModNone)
		case "\x1b[3~":
			return tcell.NewEventKey(tcell.KeyDelete, "", tcell.ModNone)
		case "\x1bOP", "\x1b[11~":
			return tcell.NewEventKey(tcell.KeyF1, "", tcell.ModNone)
		case "\x1bOQ", "\x1b[12~":
			return tcell.NewEventKey(tcell.KeyF2, "", tcell.ModNone)
		case "\x1bOR", "\x1b[13~":
			return tcell.NewEventKey(tcell.KeyF3, "", tcell.ModNone)
		case "\x1bOS", "\x1b[14~":
			return tcell.NewEventKey(tcell.KeyF4, "", tcell.ModNone)
		case "\x1b[15~":
			return tcell.NewEventKey(tcell.KeyF5, "", tcell.ModNone)
		case "\x1b[17~":
			return tcell.NewEventKey(tcell.KeyF6, "", tcell.ModNone)
		case "\x1b[18~":
			return tcell.NewEventKey(tcell.KeyF7, "", tcell.ModNone)
		case "\x1b[19~":
			return tcell.NewEventKey(tcell.KeyF8, "", tcell.ModNone)
		case "\x1b[20~":
			return tcell.NewEventKey(tcell.KeyF9, "", tcell.ModNone)
		case "\x1b[21~":
			return tcell.NewEventKey(tcell.KeyF10, "", tcell.ModNone)
		case "\x1b[23~":
			return tcell.NewEventKey(tcell.KeyF11, "", tcell.ModNone)
		case "\x1b[24~":
			return tcell.NewEventKey(tcell.KeyF12, "", tcell.ModNone)
		default:
			// Alt+char: ESC followed by a single character
			if len(s) == 2 {
				return tcell.NewEventKey(tcell.KeyRune, string(rune(s[1])), tcell.ModAlt)
			}
		}
	}
	// Multi-character printable (shouldn't normally happen from xterm.js)
	runes := []rune(s)
	if len(runes) == 1 {
		return tcell.NewEventKey(tcell.KeyRune, string(runes[0]), tcell.ModNone)
	}
	return nil
}

// ---- Custom battle support (copied from cmd/termcom/main.go) ----

type customMapDef struct {
	Type      string `json:"type"`
	Generator string `json:"generator"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type customSoldierDef struct {
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
}

type customAlienDef struct {
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
}

type customCivilianDef struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type customVictoryDef struct {
	Condition   string `json:"condition"`
	Turns       int    `json:"turns"`
	TargetX     int    `json:"target_x"`
	TargetY     int    `json:"target_y"`
	MinSoldiers int    `json:"min_soldiers"`
}

type customBattleDef struct {
	Name        string             `json:"name"`
	Author      string             `json:"author"`
	Date        string             `json:"date"`
	Description string             `json:"description"`
	Night       bool               `json:"night"`
	Map         customMapDef       `json:"map"`
	Soldiers    []customSoldierDef `json:"soldiers"`
	Aliens      []customAlienDef   `json:"aliens"`
	Civilians   []customCivilianDef `json:"civilians"`
	Victory     customVictoryDef   `json:"victory"`
}

func launchCustomBattle(g *engine.Game, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read custom battle: %v", err)
		return
	}
	var def customBattleDef
	if err := json.Unmarshal(raw, &def); err != nil {
		log.Printf("Failed to parse custom battle: %v", err)
		return
	}

	if def.Night {
		g.GameTime = time.Date(1999, time.March, 1, 2, 0, 0, 0, time.UTC)
	} else {
		g.GameTime = time.Date(1999, time.March, 1, 12, 0, 0, 0, time.UTC)
	}
	g.Paused = true

	b := base.NewBase("Test Base", 0)
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})

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
	b.Soldiers = squad

	var m *battle.BattleMap
	gen := def.Map.Generator
	w, h := def.Map.Width, def.Map.Height
	if w <= 0 || h <= 0 {
		w, h = 50, 50
	}
	switch gen {
	case "terror":
		m = battle.GenerateTerrorSite(w, h)
	case "supply_raid", "ufo_interior":
		m = battle.GenerateUFOInterior(w, h)
	case "alien_base":
		m = battle.GenerateAlienBase(w, h)
	case "alien_research":
		m = battle.GenerateUFOInterior(w, h)
	case "council":
		m = battle.GenerateTerrorSite(w, h)
	case "cydonia":
		m = battle.GenerateCydonia(w, h)
	case "abduction":
		m = battle.GenerateAbductionSite(w, h)
	case "forest":
		m = battle.GenerateForest(w, h)
	case "desert":
		m = battle.GenerateDesert(w, h)
	case "polar":
		m = battle.GeneratePolar(w, h)
	default:
		m = battle.GenerateCrashSite(w, h)
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

	bs := battle.NewCustomBattlescape(g, b, squad, m, units, cv, def.Name)
	g.SetScreen(engine.StateBattlescape, bs)
	g.SetState(engine.StateBattlescape)
}
