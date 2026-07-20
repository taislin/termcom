package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/geo"
	"github.com/taislin/termcom/internal/mapgen"
	"github.com/taislin/termcom/internal/save"
	"github.com/taislin/termcom/internal/soldier"
)

func main() {
	if err := mapgen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: mapgen init: %v\n", err)
	}
	data.NewAlienSpriteRegistry().RebuildFromTemplates(
		mapgen.ToTemplateData("head"),
		mapgen.ToTemplateData("eye"),
		mapgen.ToTemplateData("torso"),
		mapgen.ToTemplateData("leg"),
		mapgen.ToTemplateData("weapon"),
	)

	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	g.OnNewGame = func() {
		seedScreen := engine.NewSeedScreen(g, func(seed int64) {
			_ = seed
			picker := engine.NewDifficultyScreen(g, func(difficulty int) {
				gs := geo.NewGeoscape(g)
				g.RegisterScreen(engine.StateGeoscape, gs)
				g.RegisterScreen(engine.StateBase, base.NewBaseScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateEquip, base.NewEquipScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateLoadout, base.NewLoadoutScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateResearch, base.NewResearchScreen(g, gs.SelectedBase()))
				g.RegisterScreen(engine.StateManufacture, base.NewManufactureScreen(g, gs.SelectedBase()))
				g.SetState(engine.StateGeoscape)
				if !engine.Config.TutorialShown && !engine.HasSave() {
					g.RegisterScreen(engine.StateTutorial, engine.NewTutorialScreen(g, nil))
					g.PushState(engine.StateTutorial)
				}
			})
			g.PushScreen(picker)
		})
		g.PushScreen(seedScreen)
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
			launchCustomBattle(g, entry.FilePath)
		})
		g.PushScreen(screen)
	}

	g.RegisterScreen(engine.StateHelp, engine.NewHelpScreen(g, engine.StateGeoscape))
	g.RegisterScreen(engine.StateMenu, engine.NewMenuScreen(g))
	g.RegisterScreen(engine.StateLanguageSelect, engine.NewLanguageSelectScreen(g))

	g.Run()
}

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
	Name        string            `json:"name"`
	Author      string            `json:"author"`
	Date        string            `json:"date"`
	Description string            `json:"description"`
	Night       bool              `json:"night"`
	Map         customMapDef      `json:"map"`
	Soldiers    []customSoldierDef `json:"soldiers"`
	Aliens      []customAlienDef   `json:"aliens"`
	Civilians   []customCivilianDef `json:"civilians"`
	Victory     customVictoryDef  `json:"victory"`
}

func launchCustomBattle(g *engine.Game, path string) {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read custom battle: %v\n", err)
		return
	}
	var def customBattleDef
	if err := json.Unmarshal(raw, &def); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse custom battle: %v\n", err)
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

	bs := battle.NewCustomBattlescape(g, b, squad, m, units, cv, def.Name)
	g.SetScreen(engine.StateBattlescape, bs)
	g.SetState(engine.StateBattlescape)
}
