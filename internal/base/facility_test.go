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
