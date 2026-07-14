package base

import (
	"testing"

	"github.com/taislin/termcom/internal/engine"
)

func TestNewManufactureScreen(t *testing.T) {
	b := NewBase("Test", 0)
	ms := NewManufactureScreen(&engine.Game{}, b)
	if ms == nil {
		t.Fatal("NewManufactureScreen returned nil")
	}
	if ms.Base != b {
		t.Error("base should be set")
	}
}

func TestGetBuildablePlans(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	b.Facilities = append(b.Facilities, &Facility{Type: FacStorage})
	b.AddItem("alloys", 10)
	b.AddItem("elerium", 5)
	ms := NewManufactureScreen(&engine.Game{}, b)
	plans := ms.getBuildablePlans()
	if len(plans) == 0 {
		t.Error("should have buildable plans with sufficient materials")
	}
}

func TestGetBuildablePlansNoMaterials(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	ms := NewManufactureScreen(&engine.Game{}, b)
	plans := ms.getBuildablePlans()
	_ = plans
}

func TestGetBuildablePlansSorted(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	b.Facilities = append(b.Facilities, &Facility{Type: FacStorage})
	b.AddItem("alloys", 10)
	b.AddItem("elerium", 5)
	ms := NewManufactureScreen(&engine.Game{}, b)
	plans := ms.getBuildablePlans()
	if len(plans) < 2 {
		t.Skip("need at least 2 plans to test sorting")
	}
	for i := 1; i < len(plans); i++ {
		if plans[i].Days < plans[i-1].Days {
			t.Errorf("plans not sorted by days: %s (%d) < %s (%d)",
				plans[i].Name, plans[i].Days, plans[i-1].Name, plans[i-1].Days)
		}
	}
}

func TestMfgScreenStartManufacture(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	b.Facilities = append(b.Facilities, &Facility{Type: FacStorage})
	b.AddItem("alloys", 10)
	ms := NewManufactureScreen(&engine.Game{}, b)
	ms.startManufacture()
	if ms.Message == "" {
		t.Error("startManufacture should set a message")
	}
}

func TestStartManufactureNoPlans(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	ms := NewManufactureScreen(&engine.Game{}, b)
	if len(ms.getBuildablePlans()) == 0 {
		ms.startManufacture()
	} else {
		t.Skip("some plans already buildable")
	}
}

func TestManufactureSelectionBounds(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop})
	ms := NewManufactureScreen(&engine.Game{}, b)
	ms.Selection = 5
	_ = ms.getBuildablePlans()
}

func TestManufacturePlansStatic(t *testing.T) {
	if len(ManufacturePlans) == 0 {
		t.Error("ManufacturePlans should not be empty")
	}
	names := make(map[string]bool)
	for _, plan := range ManufacturePlans {
		if names[plan.Name] {
			t.Errorf("duplicate plan name: %s", plan.Name)
		}
		names[plan.Name] = true
		if plan.ItemKey == "" {
			t.Errorf("plan %s has empty ItemKey", plan.Name)
		}
		if plan.Days <= 0 {
			t.Errorf("plan %s has invalid Days: %d", plan.Name, plan.Days)
		}
	}
}
