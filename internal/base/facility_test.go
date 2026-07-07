package base

import "testing"

func TestNewBase(t *testing.T) {
	b := NewBase("Test Base")
	if b.Name != "Test Base" {
		t.Errorf("expected Test Base, got %s", b.Name)
	}
	if b.Scientists != 10 {
		t.Errorf("expected 10 scientists, got %d", b.Scientists)
	}
	if b.Engineers != 10 {
		t.Errorf("expected 10 engineers, got %d", b.Engineers)
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
	if b.CountFacility(FacWorkshop) != 1 {
		t.Errorf("expected 1 workshop, got %d", b.CountFacility(FacWorkshop))
	}
	if b.CountFacility(FacRadar) != 0 {
		t.Errorf("expected 0 radars, got %d", b.CountFacility(FacRadar))
	}
}

func TestLivingCapacity(t *testing.T) {
	b := NewBase("Test")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	if b.LivingCapacity() != 8 {
		t.Errorf("expected 8 capacity, got %d", b.LivingCapacity())
	}
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters})
	if b.LivingCapacity() != 16 {
		t.Errorf("expected 16 capacity, got %d", b.LivingCapacity())
	}
}

func TestBuildFacility(t *testing.T) {
	b := NewBase("Test")
	ok := b.BuildFacility(FacLab)
	if !ok {
		t.Error("BuildFacility should return true")
	}
	if len(b.Facilities) != 1 {
		t.Errorf("expected 1 facility, got %d", len(b.Facilities))
	}
	if !b.Facilities[0].Building {
		t.Error("new facility should be building")
	}
}

func TestAdvanceDay(t *testing.T) {
	b := NewBase("Test")
	b.BuildFacility(FacLab)
	f := b.Facilities[0]
	f.DaysLeft = 2

	b.AdvanceDay()
	if f.DaysLeft != 1 {
		t.Errorf("expected 1 day left, got %d", f.DaysLeft)
	}
	if !f.Building {
		t.Error("should still be building")
	}

	b.AdvanceDay()
	if f.DaysLeft != 0 {
		t.Errorf("expected 0 days left, got %d", f.DaysLeft)
	}
	if f.Building {
		t.Error("should be done building")
	}
}

func TestFacilityDefs(t *testing.T) {
	required := []FacilityType{FacLivingQuarters, FacLab, FacWorkshop, FacStorage, FacRadar, FacContainment, FacPsiLab, FacHangar}
	for _, ft := range required {
		def, ok := FacilityDefs[ft]
		if !ok {
			t.Errorf("missing facility def for type %d", ft)
		}
		if def.Name == "" {
			t.Errorf("facility type %d has empty name", ft)
		}
		if def.Cost <= 0 {
			t.Errorf("facility type %d has invalid cost %d", ft, def.Cost)
		}
	}
}
