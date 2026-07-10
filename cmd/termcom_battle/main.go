package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/civ13/termcom/internal/base"
	"github.com/civ13/termcom/internal/battle"
	"github.com/civ13/termcom/internal/engine"
	"github.com/civ13/termcom/internal/soldier"
)

var battleTypes = []string{
	"crash_site",
	"terror",
	"supply_raid",
	"alien_base",
	"alien_research",
	"council",
	"cydonia",
	"abduction",
	"forest",
	"desert",
	"polar",
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run ./cmd/test_battle [battle_type]\n\n")
	fmt.Fprintf(os.Stderr, "Available battle types:\n")
	for _, bt := range battleTypes {
		fmt.Fprintf(os.Stderr, "  %s\n", bt)
	}
	fmt.Fprintf(os.Stderr, "\nNo argument = random battle type\n")
	os.Exit(1)
}

func ufoNameForType(bt string) string {
	switch bt {
	case "crash_site":
		return ""
	case "terror":
		return "Terror"
	case "supply_raid":
		return "Supply Raid"
	case "alien_base":
		return "Alien Base Assault"
	case "alien_research":
		return "Alien Research"
	case "council":
		return "Council"
	case "cydonia":
		return "Cydonia"
	case "abduction":
		return "Abduction"
	case "forest":
		return "Forest"
	case "desert":
		return "Desert"
	case "polar":
		return "Polar"
	default:
		return ""
	}
}

func makeSquad() []*soldier.Soldier {
	names := []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot"}
	squad := make([]*soldier.Soldier, 0, len(names))
	for _, name := range names {
		s := soldier.NewSoldier(name)
		s.Rank = soldier.Sergeant
		s.Accuracy = 70 + rand.Intn(20)
		s.Reactions = 60 + rand.Intn(20)
		s.Strength = 20 + rand.Intn(10)
		s.HP = 30 + rand.Intn(6)
		s.MaxHP = s.HP
		s.TU = 55 + rand.Intn(10)
		s.MaxTU = s.TU
		s.Weapon = "rifle"
		s.WeaponAmmo = 50
		s.Armor = "personal_armour"
		squad = append(squad, s)
	}
	return squad
}

func main() {
	var bt string
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		valid := false
		for _, v := range battleTypes {
			if arg == v {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Fprintf(os.Stderr, "Unknown battle type: %s\n\n", arg)
			usage()
		}
		bt = arg
	} else {
		bt = battleTypes[rand.Intn(len(battleTypes))]
	}

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

	ufoName := ufoNameForType(bt)
	squad := b.HealthySoldiers()
	if len(squad) == 0 {
		fmt.Fprintf(os.Stderr, "No healthy soldiers!\n")
		os.Exit(1)
	}

	bs := battle.NewBattlescape(g, b, squad, ufoName)
	g.SetScreen(engine.StateBattlescape, bs)
	g.SetState(engine.StateBattlescape)

	fmt.Fprintf(os.Stderr, "Launching battle: %s\n", bt)
	g.Run()
}