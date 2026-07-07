package base

import "testing"

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
	if len(b.Facilities) != 1 {
		t.Errorf("expected 1, got %d", len(b.Facilities))
	}
	if !b.Facilities[0].Building {
		t.Error("should be building")
	}
}

func TestAdvanceDay(t *testing.T) {
	b := NewBase("Test")
	b.BuildFacility(FacLab)
	f := b.Facilities[0]
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
	if len(b.Soldiers) != 1 {
		t.Errorf("expected 1 soldier, got %d", len(b.Soldiers))
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
	if len(b.Soldiers) != 2 {
		t.Fatal("expected 2 soldiers")
	}
	b.DismissSoldier(0)
	if len(b.Soldiers) != 1 {
		t.Errorf("expected 1 soldier, got %d", len(b.Soldiers))
	}
}

func TestDismissOutOfBounds(t *testing.T) {
	b := NewBase("Test")
	if b.DismissSoldier(0) {
		t.Error("should fail on empty roster")
	}
	if b.DismissSoldier(-1) {
		t.Error("should fail on negative index")
	}
}

func TestRemoveDeadSoldiers(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	b.HireSoldier()
	b.HireSoldier()
	b.Soldiers[0].HP = 0
	b.Soldiers[1].HP = 20
	dead := b.RemoveDeadSoldiers()
	if len(dead) != 1 {
		t.Errorf("expected 1 dead, got %d", len(dead))
	}
	if len(b.Soldiers) != 1 {
		t.Errorf("expected 1 alive, got %d", len(b.Soldiers))
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
