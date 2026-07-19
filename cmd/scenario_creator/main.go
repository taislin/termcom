package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type mapGenDef struct {
	label string
	id    string
}

var generators = []mapGenDef{
	{"Crash Site", "crash_site"},
	{"Terror Site", "terror"},
	{"Abduction Site", "abduction"},
	{"UFO Interior", "ufo_interior"},
	{"Alien Base", "alien_base"},
	{"Cydonia", "cydonia"},
	{"Forest", "forest"},
	{"Desert", "desert"},
	{"Polar", "polar"},
	{"Farm", "farm"},
	{"Coastal", "coastal"},
	{"Mountain", "mountain"},
	{"Swamp", "swamp"},
	{"Jungle", "jungle"},
}

type scenarioBuilder struct {
	Name        string
	Author      string
	Date        string
	Description string
	Night       bool
	Generator   string
	Width       int
	Height      int
	Soldiers    []soldierDef
	Aliens      []alienDef
	Civilians   []civilianDef
	Victory     victoryDef
}

type soldierDef struct {
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

type alienDef struct {
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

type civilianDef struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type victoryDef struct {
	Condition   string `json:"condition"`
	Turns       int    `json:"turns"`
	TargetX     int    `json:"target_x"`
	TargetY     int    `json:"target_y"`
	MinSoldiers int    `json:"min_soldiers"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	s := &scenarioBuilder{
		Date:  time.Now().Format("2006-01-02"),
		Width: 50,
		Height: 50,
	}

	fmt.Print("\033[2J\033[H")
	fmt.Println("\033[1;36m╔══════════════════════════════════════════╗")
	fmt.Println("║   TERMCOM — Scenario Creator            ║")
	fmt.Println("╚══════════════════════════════════════════╝\033[0m")
	fmt.Println()

	s.Name = prompt(reader, "Mission name", "My Custom Mission")
	s.Author = prompt(reader, "Author", os.Getenv("USER"))
	s.Description = prompt(reader, "Description", "A custom mission.")

	night := prompt(reader, "Night mission? (y/n)", "n")
	s.Night = night == "y" || night == "Y"

	fmt.Println("\nMap generator:")
	for i, g := range generators {
		fmt.Printf("  %d. %s\n", i+1, g.label)
	}
	genChoice := promptInt(reader, "Select map", 1, 1, len(generators))
	s.Generator = generators[genChoice-1].id
	s.Width = promptInt(reader, "Map width", 50, 20, 100)
	s.Height = promptInt(reader, "Map height", 50, 20, 100)

	// Soldiers
	fmt.Println("\n\033[1;33m── Soldiers ──\033[0m")
	for {
		add := prompt(reader, "Add soldier? (y/n)", "n")
		if add != "y" && add != "Y" {
			break
		}
		sd := soldierDef{
			Rank:      1,
			HP:        28,
			TU:        52,
			Accuracy:  70,
			Reactions: 60,
			Strength:  18,
			Weapon:    "rifle",
			Armor:     "personal",
		}
		sd.Name = prompt(reader, "  Name", fmt.Sprintf("Soldier %d", len(s.Soldiers)+1))
		sd.Rank = promptInt(reader, "  Rank (1-8)", sd.Rank, 1, 8)
		sd.HP = promptInt(reader, "  HP", sd.HP, 1, 80)
		sd.TU = promptInt(reader, "  TU", sd.TU, 1, 100)
		sd.Accuracy = promptInt(reader, "  Accuracy", sd.Accuracy, 1, 120)
		sd.Reactions = promptInt(reader, "  Reactions", sd.Reactions, 1, 100)
		sd.Strength = promptInt(reader, "  Strength", sd.Strength, 1, 100)
		sd.Weapon = prompt(reader, "  Weapon (pistol/rifle/heavy/auto/laser_pistol/laser_rifle/plasma_pistol/plasma_rifle)", sd.Weapon)
		sd.Armor = prompt(reader, "  Armor (none/personal/light/medium/heavy/power_suit/flight_suit)", sd.Armor)
		sd.X = promptInt(reader, "  X position", 5, 0, s.Width-1)
		sd.Y = promptInt(reader, "  Y position", 10+len(s.Soldiers)*3, 0, s.Height-1)
		s.Soldiers = append(s.Soldiers, sd)
	}

	// Aliens
	fmt.Println("\n\033[1;31m── Aliens ──\033[0m")
	alienWeapons := []string{"plasma_pistol", "plasma_rifle", "heavy_plasma"}
	for {
		add := prompt(reader, "Add alien? (y/n)", "n")
		if add != "y" && add != "Y" {
			break
		}
		ad := alienDef{
			HP:        10,
			TU:        50,
			Accuracy:  55,
			Bravery:   40,
			Reactions: 50,
			Strength:  8,
			Psi:       40,
			Armour:    5,
			Weapon:    "plasma_pistol",
			Aggression: 3,
		}
		ad.Name = prompt(reader, "  Name", fmt.Sprintf("Alien %d", len(s.Aliens)+1))
		ad.HP = promptInt(reader, "  HP", ad.HP, 1, 100)
		ad.TU = promptInt(reader, "  TU", ad.TU, 1, 100)
		ad.Accuracy = promptInt(reader, "  Accuracy", ad.Accuracy, 1, 120)
		ad.Bravery = promptInt(reader, "  Bravery", ad.Bravery, 1, 110)
		ad.Reactions = promptInt(reader, "  Reactions", ad.Reactions, 1, 100)
		ad.Strength = promptInt(reader, "  Strength", ad.Strength, 1, 100)
		ad.Psi = promptInt(reader, "  Psi", ad.Psi, 0, 100)
		ad.Armour = promptInt(reader, "  Armour", ad.Armour, 0, 50)
		ad.Aggression = promptInt(reader, "  Aggression (1-10)", ad.Aggression, 1, 10)
		ad.Rank = promptInt(reader, "  Rank (0-2)", 0, 0, 2)
		fmt.Println("  Weapon options: " + strings.Join(alienWeapons, ", "))
		ad.Weapon = prompt(reader, "  Weapon", ad.Weapon)
		ad.X = promptInt(reader, "  X position", 35, 0, s.Width-1)
		ad.Y = promptInt(reader, "  Y position", 10+len(s.Aliens)*3, 0, s.Height-1)
		s.Aliens = append(s.Aliens, ad)
	}

	// Civilians
	fmt.Println("\n\033[1;32m── Civilians ──\033[0m")
	for {
		add := prompt(reader, "Add civilian? (y/n)", "n")
		if add != "y" && add != "Y" {
			break
		}
		cd := civilianDef{
			Name: fmt.Sprintf("Civilian %d", len(s.Civilians)+1),
		}
		cd.Name = prompt(reader, "  Name", cd.Name)
		cd.X = promptInt(reader, "  X position", 20, 0, s.Width-1)
		cd.Y = promptInt(reader, "  Y position", 30, 0, s.Height-1)
		s.Civilians = append(s.Civilians, cd)
	}

	// Victory condition
	fmt.Println("\n\033[1;35m── Victory Condition ──\033[0m")
	fmt.Println("  1. eliminate_all — Kill all enemies")
	fmt.Println("  2. survive_turns — Survive for N turns")
	fmt.Println("  3. reach_point  — Reach an extraction point")
	vicChoice := promptInt(reader, "  Select victory type", 1, 1, 3)
	switch vicChoice {
	case 1:
		s.Victory.Condition = "eliminate_all"
	case 2:
		s.Victory.Condition = "survive_turns"
		s.Victory.Turns = promptInt(reader, "  Turns to survive", 10, 1, 50)
	case 3:
		s.Victory.Condition = "reach_point"
		s.Victory.TargetX = promptInt(reader, "  Extraction X", 45, 0, s.Width-1)
		s.Victory.TargetY = promptInt(reader, "  Extraction Y", 45, 0, s.Height-1)
		s.Victory.MinSoldiers = promptInt(reader, "  Min soldiers to extract", 1, 1, 10)
	}

	// Output
	output := buildJSON(s)
	filename := strings.ReplaceAll(strings.ToLower(s.Name), " ", "_") + ".json"
	outputPath := filepath.Join("maps", filename)

	if err := os.MkdirAll("maps", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create maps/ dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\033[1;32m✓ Scenario saved to: %s\033[0m\n", outputPath)
	fmt.Printf("\n  Launch with: go run ./cmd/termcom_battle %s\n", outputPath)
}

func prompt(reader *bufio.Reader, label, def string) string {
	fmt.Printf("  \033[90m%s\033[0m [%s]: ", label, def)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

func promptInt(reader *bufio.Reader, label string, def, min, max int) int {
	fmt.Printf("  \033[90m%s\033[0m [%d]: ", label, def)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	n, err := strconv.Atoi(input)
	if err != nil || n < min || n > max {
		return def
	}
	return n
}

func buildJSON(s *scenarioBuilder) []byte {
	type outputSoldier struct {
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
	type outputAlien struct {
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
	type outputCivilian struct {
		Name string `json:"name"`
		X    int    `json:"x"`
		Y    int    `json:"y"`
	}
	type outputMap struct {
		Type      string `json:"type"`
		Generator string `json:"generator"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	}
	type outputVictory struct {
		Condition   string `json:"condition"`
		Turns       int    `json:"turns"`
		TargetX     int    `json:"target_x"`
		TargetY     int    `json:"target_y"`
		MinSoldiers int    `json:"min_soldiers"`
	}
	type output struct {
		Name        string           `json:"name"`
		Author      string           `json:"author"`
		Date        string           `json:"date"`
		Description string           `json:"description"`
		Night       bool             `json:"night"`
		Map         outputMap        `json:"map"`
		Soldiers    []outputSoldier  `json:"soldiers"`
		Aliens      []outputAlien    `json:"aliens"`
		Civilians   []outputCivilian `json:"civilians"`
		Victory     outputVictory    `json:"victory"`
	}

	out := output{
		Name:        s.Name,
		Author:      s.Author,
		Date:        s.Date,
		Description: s.Description,
		Night:       s.Night,
		Map: outputMap{
			Type:      "generated",
			Generator: s.Generator,
			Width:     s.Width,
			Height:    s.Height,
		},
		Victory: outputVictory{
			Condition:   s.Victory.Condition,
			Turns:       s.Victory.Turns,
			TargetX:     s.Victory.TargetX,
			TargetY:     s.Victory.TargetY,
			MinSoldiers: s.Victory.MinSoldiers,
		},
	}
	for _, sd := range s.Soldiers {
		out.Soldiers = append(out.Soldiers, outputSoldier(sd))
	}
	for _, ad := range s.Aliens {
		out.Aliens = append(out.Aliens, outputAlien(ad))
	}
	for _, cd := range s.Civilians {
		out.Civilians = append(out.Civilians, outputCivilian(cd))
	}

	data, _ := json.MarshalIndent(out, "", "  ")
	return data
}
