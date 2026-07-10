package save

import (
	"os"
	"testing"
	"time"

	"github.com/civ13/ycom/internal/base"
	"github.com/civ13/ycom/internal/soldier"
)

func TestSaveLoadGame(t *testing.T) {
	b := base.NewBase("Test Base", 0)
	b.Scientists = 15
	b.Engineers = 8
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacLab, Building: false})
	s := soldier.NewSoldier("TestGuy")
	s.Rank = soldier.Corporal
	s.HP = 18
	s.MaxHP = 22
	s.Kills = 5
	s.Weapon = "rifle"
	s.Armor = "personal"
	s.Wounds = 3
	b.Soldiers = append(b.Soldiers, s)
	b.AddItem("alloys", 10)
	b.AddItem("rifle", 2)
	b.CompletedResearch = append(b.CompletedResearch, "alien_alloys")
	b.ActiveResearch = &base.ResearchProject{
		TopicID:    "sectoid_autopsy",
		Progress:   20,
		Cost:       40,
		Scientists: 5,
	}
	b.ManufactureQueue = append(b.ManufactureQueue, &base.ManufactureJob{
		ItemKey:   "pistol",
		Count:     3,
		Progress:  2,
		CostDays:  10,
		Engineers: 4,
	})

	sd := &SaveData{
		GameTime:      time.Date(1999, time.June, 15, 12, 0, 0, 0, time.UTC),
		Funds:         500000,
		Paused:        true,
		TimeSpeed:     2,
		AlienActivity: 15,
		Bases:         []*BaseSave{FromBase(b)},
	}

	path := "test_save.json"
	err := SaveGame(path, sd)
	if err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove(path)

	loaded, err := LoadGame(path)
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	loadedBase := ToBase(loaded.Bases[0])

	if loaded.Funds != 500000 {
		t.Errorf("expected funds 500000, got %d", loaded.Funds)
	}
	if loaded.AlienActivity != 15 {
		t.Errorf("expected alien activity 15, got %d", loaded.AlienActivity)
	}
	if loadedBase.Name != "Test Base" {
		t.Errorf("expected Test Base, got %s", loadedBase.Name)
	}
	if len(loadedBase.Soldiers) != 5 {
		t.Fatalf("expected 5 soldiers (4 starting + 1 custom), got %d", len(loadedBase.Soldiers))
	}
	sol := loadedBase.Soldiers[4]
	if sol.Name != "TestGuy" {
		t.Errorf("expected TestGuy, got %s", sol.Name)
	}
	if sol.Rank != soldier.Corporal {
		t.Errorf("expected Corporal (%d), got %d", soldier.Corporal, sol.Rank)
	}
	if sol.Wounds != 3 {
		t.Errorf("expected 3 wounds, got %d", sol.Wounds)
	}
	if loadedBase.CountItem("alloys") != 10 {
		t.Errorf("expected 10 alloys, got %d", loadedBase.CountItem("alloys"))
	}
	if loadedBase.CountItem("rifle") != 2 {
		t.Errorf("expected 2 rifles, got %d", loadedBase.CountItem("rifle"))
	}
	if len(loadedBase.CompletedResearch) != 1 {
		t.Errorf("expected 1 completed research, got %d", len(loadedBase.CompletedResearch))
	}
	if loadedBase.ActiveResearch == nil {
		t.Fatal("expected active research")
	}
	if loadedBase.ActiveResearch.TopicID != "sectoid_autopsy" {
		t.Errorf("expected sectoid_autopsy, got %s", loadedBase.ActiveResearch.TopicID)
	}
	if len(loadedBase.ManufactureQueue) != 1 {
		t.Fatalf("expected 1 manufacture job, got %d", len(loadedBase.ManufactureQueue))
	}
	if loadedBase.ManufactureQueue[0].ItemKey != "pistol" {
		t.Errorf("expected pistol, got %s", loadedBase.ManufactureQueue[0].ItemKey)
	}
}

func TestFromBaseToBase(t *testing.T) {
	b := base.NewBase("Round Trip", 2)
	b.Scientists = 20
	b.Engineers = 12
	b.Facilities = append(b.Facilities, &base.Facility{Type: base.FacStorage, Building: false})
	s := soldier.NewSoldier("Alice")
	s.Rank = soldier.Sergeant
	s.HP = 25
	s.MaxHP = 30
	s.Weapon = "laser_rifle"
	s.Armor = "medium"
	s.Kills = 12
	s.Wounds = 7
	b.Soldiers = append(b.Soldiers, s)
	b.AddItem("elerium", 5)
	b.AddItem("personal", 2)

	bs := FromBase(b)
	b2 := ToBase(bs)

	if b2.Name != "Round Trip" {
		t.Errorf("expected Round Trip, got %s", b2.Name)
	}
	if b2.Scientists != 20 {
		t.Errorf("expected 20 scientists, got %d", b2.Scientists)
	}
	if len(b2.Soldiers) != 5 {
		t.Fatalf("expected 5 soldiers (4 starting + 1 custom), got %d", len(b2.Soldiers))
	}
	s2 := b2.Soldiers[4]
	if s2.Name != "Alice" {
		t.Errorf("expected Alice, got %s", s2.Name)
	}
	if s2.Rank != soldier.Sergeant {
		t.Errorf("expected Sergeant, got %v", s2.Rank)
	}
	if s2.Weapon != "laser_rifle" {
		t.Errorf("expected laser_rifle, got %s", s2.Weapon)
	}
	if s2.Armor != "medium" {
		t.Errorf("expected medium, got %s", s2.Armor)
	}
	if s2.Wounds != 7 {
		t.Errorf("expected 7 wounds, got %d", s2.Wounds)
	}
	if b2.CountItem("elerium") != 5 {
		t.Errorf("expected 5 elerium, got %d", b2.CountItem("elerium"))
	}
	if len(b2.Facilities) != 2 {
		t.Errorf("expected 2 facilities (hangar + storage), got %d", len(b2.Facilities))
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := LoadGame("nonexistent_save_file.json")
	if err == nil {
		t.Error("expected error loading nonexistent file")
	}
}
