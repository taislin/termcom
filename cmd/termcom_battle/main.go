package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/mapgen"
	"github.com/taislin/termcom/internal/soldier"
	"golang.org/x/term"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiUnderline = "\033[4m"
	ansiClear     = "\033[2J\033[H"

	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[37m"
	ansiGray   = "\033[90m"
)

type menuEntry struct {
	kind     string // "builtin" or "custom"
	label    string
	ufoName  string // for builtins
	filePath string // for customs
}

// readKey reads a single keypress from raw stdin and returns a descriptive string.
func readKey() string {
	var buf [3]byte
	n, _ := os.Stdin.Read(buf[:])
	if n == 0 {
		return ""
	}
	if buf[0] == '\x1b' && n >= 3 && buf[1] == '[' {
		switch buf[2] {
		case 'A':
			return "up"
		case 'B':
			return "down"
		case 'C':
			return "right"
		case 'D':
			return "left"
		}
	}
	if buf[0] == '\r' || buf[0] == '\n' {
		return "enter"
	}
	return string(buf[:n])
}

type customBattle struct {
	Name        string            `json:"name"`
	Author      string            `json:"author"`
	Date        string            `json:"date"`
	Description string            `json:"description"`
	Night       bool              `json:"night"`
	MapDef      customMapDef      `json:"map"`
	Soldiers    []customSoldier   `json:"soldiers"`
	Aliens      []customAlien     `json:"aliens"`
	Civilians   []customCivilian  `json:"civilians"`
	Victory     customVictoryDef  `json:"victory"`
}

type customMapDef struct {
	Type      string `json:"type"`
	Generator string `json:"generator"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type customSoldier struct {
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

type customAlien struct {
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

type customCivilian struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type customVictoryDef struct {
	Condition  string `json:"condition"`
	Turns      int    `json:"turns"`
	TargetX    int    `json:"target_x"`
	TargetY    int    `json:"target_y"`
	MinSoldiers int  `json:"min_soldiers"`
}

var builtinTypes = []struct {
	label   string
	ufoName string
}{
	{"Crash Site", ""},
	{"Terror", "Terror"},
	{"Supply Raid", "Supply Raid"},
	{"Alien Base Assault", "Alien Base Assault"},
	{"Alien Research", "Alien Research"},
	{"Council Mission", "Council"},
	{"Cydonia (Final)", "Cydonia"},
	{"Abduction", "Abduction"},
	{"Forest Patrol", "Forest"},
	{"Desert Raid", "Desert"},
	{"Polar Ops", "Polar"},
	{"Jungle Patrol", "Jungle"},
	{"Urban Combat", "Urban"},
	{"Coastal Assault", "Coastal"},
	{"Mountain Ops", "Mountain"},
	{"Swamp Raid", "Swamp"},
	{"Farm Defense", "Farm"},
	{"Rural Patrol", "Rural"},
}

func scanCustomMaps() []string {
	mapsDir := "maps"
	if _, err := os.Stat(mapsDir); os.IsNotExist(err) {
		return nil
	}
	entries, err := os.ReadDir(mapsDir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			files = append(files, filepath.Join(mapsDir, e.Name()))
		}
	}
	return files
}

func loadCustomBattle(path string) (*customBattle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cb customBattle
	if err := json.Unmarshal(data, &cb); err != nil {
		return nil, err
	}
	return &cb, nil
}

func printMenu(left []menuEntry, selected int) {
	fmt.Print(ansiClear)
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%s%s  TERMCOM — Battle Test Launcher%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Println()

	// Left panel: list
	fmt.Printf("%s%s  %-40s│  Details%s\n", ansiBold, ansiWhite, "Mission Select", ansiReset)
	fmt.Printf("%s  %-40s│%s\n", ansiGray, strings.Repeat("─", 40), ansiReset)

	for i, entry := range left {
		color := ansiWhite
		kindTag := ""
		if entry.kind == "custom" {
			kindTag = fmt.Sprintf(" %s[custom]%s", ansiCyan, ansiReset)
		}

		if i == selected {
			fmt.Printf("  %s%s%-2d. %s%s%s%s\n", ansiBold, ansiGreen, i+1, color, entry.label, ansiReset, kindTag)
		} else {
			fmt.Printf("  %s%-2d. %s%s%s\n", ansiDim, i+1, ansiReset, entry.label, kindTag)
		}
	}

	fmt.Printf("  %s%-2d. %s%s\n", ansiDim, len(left)+1, "Quit", ansiReset)

	// Right panel: details of selected
	fmt.Printf("\033[%dA", len(left)+3) // move cursor up to header row level
	rightX := 43
	fmt.Printf("\033[%dC", rightX)

	if selected >= 0 && selected < len(left) {
		entry := left[selected]
		fmt.Printf("%s%s%s\n", ansiBold, entry.label, ansiReset)
		fmt.Printf("\033[%dC", rightX)

		if entry.kind == "builtin" {
			fmt.Printf("%sType:%s %s\n", ansiDim, ansiReset, entry.ufoName)
			fmt.Printf("\033[%dC", rightX)
			fmt.Printf("%sMap:%s %s tiles\n", ansiDim, ansiReset, "50x50")
			fmt.Printf("\033[%dC", rightX)
			fmt.Printf("%sAliens:%s auto-scaled\n", ansiDim, ansiReset)
			fmt.Printf("\033[%dC", rightX)
			fmt.Printf("%sVictory:%s eliminate all\n", ansiDim, ansiReset)
		} else if entry.filePath != "" {
			cb, err := loadCustomBattle(entry.filePath)
			if err != nil {
				fmt.Printf("%s%sError loading:%s %s\n", ansiRed, ansiBold, ansiReset, err)
			} else {
				if cb.Author != "" {
					fmt.Printf("%sAuthor:%s %s\n", ansiDim, ansiReset, cb.Author)
					fmt.Printf("\033[%dC", rightX)
				}
				if cb.Date != "" {
					fmt.Printf("%sDate:%s   %s\n", ansiDim, ansiReset, cb.Date)
					fmt.Printf("\033[%dC", rightX)
				}
				if cb.Description != "" {
					// Word wrap description at ~35 chars
					words := strings.Fields(cb.Description)
					line := ""
					first := true
					for _, w := range words {
						if len(line)+len(w)+1 > 35 {
							if first {
								fmt.Printf("%sDesc:%s    %s\n", ansiDim, ansiReset, line)
								first = false
							} else {
								fmt.Printf("\033[%dC%s      %s\n", rightX, ansiDim, line)
							}
							line = w
						} else {
							if line != "" {
								line += " "
							}
							line += w
						}
					}
					if line != "" {
						if first {
							fmt.Printf("%sDesc:%s    %s\n", ansiDim, ansiReset, line)
						} else {
							fmt.Printf("\033[%dC%s      %s\n", rightX, ansiDim, line)
						}
					}
					fmt.Printf("\033[%dC", rightX)
				}
				fmt.Printf("%sMap:%s     %s %dx%d\n", ansiDim, ansiReset, cb.MapDef.Generator, cb.MapDef.Width, cb.MapDef.Height)
				fmt.Printf("\033[%dC", rightX)
				fmt.Printf("%sSoldiers:%s %d\n", ansiDim, ansiReset, len(cb.Soldiers))
				fmt.Printf("\033[%dC", rightX)
				fmt.Printf("%sAliens:%s   %d\n", ansiDim, ansiReset, len(cb.Aliens))
				fmt.Printf("\033[%dC", rightX)
				fmt.Printf("%sCivs:%s     %d\n", ansiDim, ansiReset, len(cb.Civilians))
				fmt.Printf("\033[%dC", rightX)

				vicLabel := cb.Victory.Condition
				switch cb.Victory.Condition {
				case "survive_turns":
					vicLabel = fmt.Sprintf("survive %d turns", cb.Victory.Turns)
				case "reach_point":
					vicLabel = fmt.Sprintf("reach (%d,%d)", cb.Victory.TargetX, cb.Victory.TargetY)
				}
				fmt.Printf("%sVictory:%s  %s\n", ansiDim, ansiReset, vicLabel)
				fmt.Printf("\033[%dC", rightX)

				if cb.Night {
					fmt.Printf("%sTime:%s     %snight%s\n", ansiDim, ansiReset, ansiBlue, ansiReset)
				} else {
					fmt.Printf("%sTime:%s     day\n", ansiDim, ansiReset)
				}
			}
		}
	}

	// Move cursor back down to prompt
	down := len(left) + 7
	fmt.Printf("\033[%dB", down)
	fmt.Printf("\r%s  Select mission [1-%d]: %s", ansiBold, len(left)+1, ansiReset)
}

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

	// Build menu entries
	var entries []menuEntry

	// Built-in battles
	for _, bt := range builtinTypes {
		entries = append(entries, menuEntry{
			kind:    "builtin",
			label:   bt.label,
			ufoName: bt.ufoName,
		})
	}

	// Custom maps from maps/ folder
	customFiles := scanCustomMaps()
	for _, f := range customFiles {
		cb, err := loadCustomBattle(f)
		label := filepath.Base(f)
		if err == nil && cb.Name != "" {
			label = cb.Name
		}
		entries = append(entries, menuEntry{
			kind:     "custom",
			label:    label,
			filePath: f,
		})
	}

	// Also accept command-line arg to skip menu
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		// Check builtins
		for _, bt := range builtinTypes {
			if strings.ReplaceAll(strings.ToLower(bt.label), " ", "_") == arg ||
				strings.ToLower(bt.ufoName) == arg {
				launchBuiltin(bt.ufoName, "")
				return
			}
		}
		// Check if it's a file path
		if _, err := os.Stat(os.Args[1]); err == nil {
			launchCustom(os.Args[1])
			return
		}
		fmt.Fprintf(os.Stderr, "Unknown battle type or file: %s\n", os.Args[1])
		os.Exit(1)
	}

	// Interactive menu
	selected := 0

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to enable raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	for {
		printMenu(entries, selected)

		key := readKey()

		switch key {
		case "q", "Q":
			fmt.Println("Goodbye.")
			return
		case "up", "k":
			if selected > 0 {
				selected--
			}
			continue
		case "down", "j":
			if selected < len(entries)-1 {
				selected++
			}
			continue
		case "enter":
			// fall through to launch below
		default:
			n, err := strconv.Atoi(key)
			if err != nil || n < 1 || n > len(entries)+1 {
				continue
			}
			if n == len(entries)+1 {
				fmt.Println("Goodbye.")
				return
			}
			selected = n - 1
		}

		entry := entries[selected]
		if entry.kind == "builtin" {
			launchBuiltin(entry.ufoName, "")
		} else {
			launchCustom(entry.filePath)
		}
		return
	}
}

func launchBuiltin(ufoName, extra string) {
	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	g.GameTime = time.Date(1999, time.March, 1, 12, 0, 0, 0, time.UTC)
	g.Paused = true

	b := base.NewBase("Test Base", 0)
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Row: 0, Col: 3})
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacRadar, Row: 0, Col: 4})
	b.Soldiers = makeSquad()

	squad := b.HealthySoldiers()
	if len(squad) == 0 {
		fmt.Fprintf(os.Stderr, "No healthy soldiers!\n")
		os.Exit(1)
	}

	bs := battle.NewBattlescape(g, b, squad, ufoName, 42, -1, -1)
	g.SetScreen(engine.StateBattlescape, bs)
	g.SetState(engine.StateBattlescape)

	fmt.Fprintf(os.Stderr, "Launching battle: %s\n", ufoName)
	g.Run()
}

func launchCustom(path string) {
	cb, err := loadCustomBattle(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load custom battle: %v\n", err)
		os.Exit(1)
	}

	g, err := engine.NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	if cb.Night {
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

	// Build squad from JSON or use defaults
	var squad []*soldier.Soldier
	for _, cs := range cb.Soldiers {
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
		squad = makeSquad()
	}
	b.Soldiers = squad

	// Generate map
	var m *battle.BattleMap
	gen := cb.MapDef.Generator
	w, h := cb.MapDef.Width, cb.MapDef.Height
	if w <= 0 || h <= 0 {
		w, h = 50, 50
	}
	switch gen {
	case "terror":
		m = battle.GenerateTerrorSite(w, h, time.Now().UnixNano())
	case "supply_raid", "ufo_interior":
		m = battle.GenerateUFOInterior(w, h, time.Now().UnixNano())
	case "building_assault":
		m = battle.GenerateUrbanBuildingWFC(w, h, rand.New(rand.NewSource(42)))
	case "alien_base":
		m = battle.GenerateAlienBase(w, h)
	case "alien_research":
		m = battle.GenerateUFOInterior(w, h, time.Now().UnixNano())
	case "council":
		m = battle.GenerateTerrorSite(w, h, time.Now().UnixNano())
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
		m, _ = battle.GenerateCrashSite(w, h, 42, -1, -1)
	}

	// Build unit definitions
	var units []battle.CustomUnitDef
	for _, cs := range cb.Soldiers {
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
	for _, ca := range cb.Aliens {
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
	for _, cc := range cb.Civilians {
		units = append(units, battle.CustomUnitDef{
			Name:    cc.Name,
			Faction: 2,
			X:       cc.X,
			Y:       cc.Y,
		})
	}

	// Build victory condition
	var cv *battle.CustomVictory
	if cb.Victory.Condition != "" {
		cv = &battle.CustomVictory{
			Condition:   cb.Victory.Condition,
			Turns:       cb.Victory.Turns,
			TargetX:     cb.Victory.TargetX,
			TargetY:     cb.Victory.TargetY,
			MinSoldiers: cb.Victory.MinSoldiers,
		}
	}

	bs := battle.NewCustomBattlescape(g, b, squad, m, units, cv, cb.Name)
	g.SetScreen(engine.StateBattlescape, bs)
	g.SetState(engine.StateBattlescape)

	fmt.Fprintf(os.Stderr, "Launching custom battle: %s\n", cb.Name)
	g.Run()
}

func makeSquad() []*soldier.Soldier {
	names := []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot"}
	squad := make([]*soldier.Soldier, 0, len(names))
	for _, name := range names {
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
	return squad
}
