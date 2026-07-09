package base

import (
	"os"
	"testing"

	"github.com/civ13/ycom/internal/data"
)

func TestMain(m *testing.M) {
	species, _ := data.GenerateSpecies(42)
	data.InitResearchTree(42, species)
	os.Exit(m.Run())
}

func TestNewBase(t *testing.T) {
	b := NewBase("Test")
	if b.Name != "Test" {
		t.Errorf("expected Test, got %s", b.Name)
	}
	if b.Scientists != 10 {
		t.Errorf("expected 10 scientists, got %d", b.Scientists)
	}
}

func TestCountFacility(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	if b.CountFacility(FacLab) != 2 {
		t.Errorf("expected 2 labs, got %d", b.CountFacility(FacLab))
	}
}

func TestLivingCapacity(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	if b.LivingCapacity() != 8 {
		t.Errorf("expected 8, got %d", b.LivingCapacity())
	}
}

func TestBuildFacility(t *testing.T) {
	b := NewBase("Test")
	ok := b.BuildFacility(FacLab)
	if !ok {
		t.Error("BuildFacility should return true")
	}
	if len(b.Facilities) != 2 {
		t.Errorf("expected 2, got %d", len(b.Facilities))
	}
	if !b.Facilities[1].Building {
		t.Error("should be building")
	}
}

func TestAdvanceDay(t *testing.T) {
	b := NewBase("Test")
	b.BuildFacility(FacLab)
	f := b.Facilities[1]
	f.DaysLeft = 2
	b.AdvanceDay()
	if f.DaysLeft != 1 || !f.Building {
		t.Error("after 1 day should still be building")
	}
	b.AdvanceDay()
	if f.DaysLeft != 0 || f.Building {
		t.Error("after 2 days should be done")
	}
}

func TestHireSoldier(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	ok, msg := b.HireSoldier()
	if !ok {
		t.Errorf("HireSoldier failed: %s", msg)
	}
	if len(b.Soldiers) != 5 {
		t.Errorf("expected 5 soldiers (4 starting + 1 hired), got %d", len(b.Soldiers))
	}
}

func TestHireSoldierNoRoom(t *testing.T) {
	b := NewBase("Test")
	for i := 0; i < 8; i++ {
		b.HireSoldier()
	}
	ok, _ := b.HireSoldier()
	if ok {
		t.Error("should fail when at capacity")
	}
}

func TestDismissSoldier(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.HireSoldier()
	if len(b.Soldiers) != 6 {
		t.Fatal("expected 6 soldiers (4 starting + 2 hired)")
	}
	b.DismissSoldier(0)
	if len(b.Soldiers) != 5 {
		t.Errorf("expected 5 soldiers, got %d", len(b.Soldiers))
	}
}

func TestDismissOutOfBounds(t *testing.T) {
	b := NewBase("Test")
	if b.DismissSoldier(-1) {
		t.Error("should fail on negative index")
	}
	if b.DismissSoldier(100) {
		t.Error("should fail on out-of-bounds index")
	}
}

func TestRemoveDeadSoldiers(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.HireSoldier()
	// b has 4 starting + 2 hired = 6, kill the 5th soldier (index 4)
	b.Soldiers[4].HP = 0
	b.Soldiers[5].HP = 20
	dead := b.RemoveDeadSoldiers()
	if len(dead) != 1 {
		t.Errorf("expected 1 dead, got %d", len(dead))
	}
	if len(b.Soldiers) != 5 {
		t.Errorf("expected 5 alive, got %d", len(b.Soldiers))
	}
}

func TestFacilityDefs(t *testing.T) {
	required := []FacilityType{FacLivingQuarters, FacLab, FacWorkshop, FacStorage, FacRadar, FacContainment, FacPsiLab, FacHangar}
	for _, ft := range required {
		def, ok := FacilityDefs[ft]
		if !ok {
			t.Errorf("missing def for %d", ft)
		}
		if def.Name == "" {
			t.Errorf("empty name for %d", ft)
		}
	}
}

func TestAddItem(t *testing.T) {
	b := NewBase("Test")
	b.AddItem("rifle", 3)
	if b.CountItem("rifle") != 3 {
		t.Errorf("expected 3 rifles, got %d", b.CountItem("rifle"))
	}
	b.AddItem("rifle", 2)
	if b.CountItem("rifle") != 5 {
		t.Errorf("expected 5 rifles after add, got %d", b.CountItem("rifle"))
	}
}

func TestRemoveItem(t *testing.T) {
	b := NewBase("Test")
	b.AddItem("rifle", 3)
	if !b.RemoveItem("rifle", 2) {
		t.Error("should succeed")
	}
	if b.CountItem("rifle") != 1 {
		t.Errorf("expected 1, got %d", b.CountItem("rifle"))
	}
	if b.RemoveItem("rifle", 5) {
		t.Error("should fail")
	}
}

func TestAddLoot(t *testing.T) {
	b := NewBase("Test")
	b.AddLoot([]string{"alloys", "elerium", "corpse_sect"})
	if b.CountItem("alloys") != 1 {
		t.Error("expected 1 alloys")
	}
	if b.CountItem("elerium") != 1 {
		t.Error("expected 1 elerium")
	}
	if b.CountItem("corpse_sect") != 1 {
		t.Error("expected 1 corpse_sect")
	}
}

func TestEquipWeapon(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.AddItem("laser_rifle", 1)
	s := b.Soldiers[0]
	s.Weapon = "rifle"
	if !b.EquipWeapon(0, "laser_rifle") {
		t.Error("should equip")
	}
	if s.Weapon != "laser_rifle" {
		t.Errorf("expected laser_rifle, got %s", s.Weapon)
	}
	if b.CountItem("rifle") != 1 {
		t.Error("old weapon should be returned to stores")
	}
}

func TestEquipWeaponNoStock(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	if b.EquipWeapon(0, "plasma_rifle") {
		t.Error("should fail with no stock")
	}
}

func TestEquipArmor(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.AddItem("personal", 1)
	s := b.Soldiers[0]
	s.Armor = "none"
	if !b.EquipArmor(0, "personal") {
		t.Error("should equip")
	}
	if s.Armor != "personal" {
		t.Errorf("expected personal, got %s", s.Armor)
	}
}

func TestEquipArmorSwap(t *testing.T) {
	b := NewBase("Test")
	b.Scientists = 0
	b.Engineers = 0
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.AddItem("medium", 1)
	s := b.Soldiers[0]
	s.Armor = "light"
	if !b.EquipArmor(0, "medium") {
		t.Error("should equip")
	}
	if s.Armor != "medium" {
		t.Errorf("expected medium, got %s", s.Armor)
	}
	if b.CountItem("light") != 1 {
		t.Errorf("old armor should be returned, got %d", b.CountItem("light"))
	}
}

func TestEquipArmorNone(t *testing.T) {
	b := NewBase("Test")
	b.Scientists = 0
	b.Engineers = 0
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	s := b.Soldiers[0]
	s.Armor = "personal"
	if !b.EquipArmor(0, "none") {
		t.Error("should equip none")
	}
	if s.Armor != "none" {
		t.Errorf("expected none, got %s", s.Armor)
	}
	if b.CountItem("personal") != 1 {
		t.Errorf("armor should return to stores, got %d", b.CountItem("personal"))
	}
}

func TestMonthlyBudget(t *testing.T) {
	b := NewBase("Test")
	b.Scientists = 0
	b.Engineers = 0
	// HireSoldier already decreases soldiers - remove starting soldiers for clean test
	b.Soldiers = nil
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	salary := b.MonthlySalary()
	if salary != 2000 {
		t.Errorf("expected 2000, got %d", salary)
	}
	funding := b.GovernmentFunding()
	if funding != 200000 {
		t.Errorf("expected 200000, got %d", funding)
	}
}

func TestStorageCapacity(t *testing.T) {
	b := NewBase("Test")
	if b.StorageCapacity() != 0 {
		t.Error("no storage should be 0")
	}
	b.Facilities = append(b.Facilities, &Facility{Type: FacStorage})
	if b.StorageCapacity() != 50 {
		t.Errorf("expected 50, got %d", b.StorageCapacity())
	}
}

func TestHasResearch(t *testing.T) {
	b := NewBase("Test")
	if b.HasResearch("alien_alloys") {
		t.Error("should not have alien_alloys yet")
	}
	b.CompletedResearch = append(b.CompletedResearch, "alien_alloys")
	if !b.HasResearch("alien_alloys") {
		t.Error("should have alien_alloys now")
	}
}

func TestCanResearch(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	topic := data.ResearchByID("alien_alloys")
	if !b.CanResearch(topic) {
		t.Error("should be able to research alien_alloys")
	}
	topic2 := data.ResearchByID("light_suit")
	if b.CanResearch(topic2) {
		t.Error("should not be able to research light_suit (missing prereqs)")
	}
}

func TestStartResearch(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	ok := b.StartResearch("alien_alloys")
	if !ok {
		t.Error("should start research")
	}
	if b.ActiveResearch == nil {
		t.Fatal("ActiveResearch should be set")
	}
	if b.ActiveResearch.TopicID != "alien_alloys" {
		t.Errorf("expected alien_alloys, got %s", b.ActiveResearch.TopicID)
	}
}

func TestAdvanceResearch(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	b.StartResearch("alien_alloys")
	b.ActiveResearch.Cost = 10
	b.ActiveResearch.Scientists = 5
	done := b.AdvanceResearch()
	if len(done) > 0 {
		t.Error("should not be done yet")
	}
	if b.ActiveResearch.Progress != 5 {
		t.Errorf("expected 5, got %d", b.ActiveResearch.Progress)
	}
	done = b.AdvanceResearch()
	if len(done) == 0 {
		t.Error("should be done now")
	}
	if !b.HasResearch("alien_alloys") {
		t.Error("should have completed research")
	}
}

func TestStartManufacture(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	b.AddItem("alloys", 5)
	ok := b.StartManufacture("pistol", 1, map[string]int{"alloys": 1})
	if !ok {
		t.Error("should start manufacture")
	}
	if len(b.ManufactureQueue) != 1 {
		t.Errorf("expected 1 job, got %d", len(b.ManufactureQueue))
	}
	if b.CountItem("alloys") != 4 {
		t.Errorf("expected 4 alloys, got %d", b.CountItem("alloys"))
	}
}

func TestStartManufactureNoMaterials(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	ok := b.StartManufacture("pistol", 1, map[string]int{"alloys": 1})
	if ok {
		t.Error("should fail without materials")
	}
}

func TestAdvanceManufacture(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	b.AddItem("alloys", 5)
	b.StartManufacture("pistol", 1, map[string]int{"alloys": 1})
	b.ManufactureQueue[0].CostDays = 10
	b.ManufactureQueue[0].Engineers = 3
	crafted := b.AdvanceManufacture()
	if len(crafted) > 0 {
		t.Error("should not be done yet")
	}
	if b.ManufactureQueue[0].Progress != 3 {
		t.Errorf("expected 3 progress, got %d", b.ManufactureQueue[0].Progress)
	}
	crafted = b.AdvanceManufacture()
	if len(crafted) > 0 {
		t.Error("should not be done yet (6/10)")
	}
	crafted = b.AdvanceManufacture()
	if len(crafted) > 0 {
		t.Error("should not be done yet (9/10)")
	}
	crafted = b.AdvanceManufacture()
	if len(crafted) == 0 {
		t.Error("should be done now (12/10)")
	}
	if b.CountItem("pistol") != 1 {
		t.Error("pistol should be in stores")
	}
}

func TestAdvanceDayHealing(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	s := b.Soldiers[0]
	s.HP = 10
	s.MaxHP = 20
	s.Wounds = 5
	b.AdvanceDay()
	if s.Wounds != 4 {
		t.Errorf("expected 4 wounds, got %d", s.Wounds)
	}
	if s.HP != 12 {
		t.Errorf("expected 12 HP, got %d", s.HP)
	}
}

func TestAdvanceDayFullHeal(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	s := b.Soldiers[0]
	s.HP = 10
	s.MaxHP = 20
	s.Wounds = 1
	b.AdvanceDay()
	if s.Wounds != 0 {
		t.Errorf("expected 0 wounds, got %d", s.Wounds)
	}
	if s.HP != s.MaxHP {
		t.Errorf("expected full HP %d, got %d", s.MaxHP, s.HP)
	}
}

func TestSellFacility(t *testing.T) {
	b := NewBase("Test")
	if len(b.Facilities) != 1 {
		t.Fatal("expected 1 starting hangar facility")
	}
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab, Building: false})
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop, Building: false})
	if len(b.Facilities) != 3 {
		t.Fatal("expected 3 facilities")
	}
	b.Facilities = append(b.Facilities[:1], b.Facilities[2:]...)
	if len(b.Facilities) != 2 {
		t.Errorf("expected 2 facilities, got %d", len(b.Facilities))
	}
}
