package base

import (
	"testing"
	"time"

	"github.com/taislin/termcom/internal/engine"
)

// --- Phase 41: base test coverage (gaps beyond Phase 34) ---

func TestAdjacencyBonus(t *testing.T) {
	b := NewBase("Adj", 0)
	b.Facilities = nil
	if got := b.AdjacentResearchBonus(); got != 1.0 {
		t.Errorf("expected 1.0 with no labs, got %f", got)
	}
	if got := b.AdjacentManufactureBonus(); got != 1.0 {
		t.Errorf("expected 1.0 with no workshops, got %f", got)
	}

	// Two adjacent labs -> +0.20 multiplier (each counts the other).
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab, Row: 1, Col: 0})
	if got := b.AdjacentResearchBonus(); got != 1.2 {
		t.Errorf("expected 1.2 with two adjacent labs, got %f", got)
	}

	// Two adjacent workshops -> +0.20 multiplier.
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop, Row: 1, Col: 2})
	if got := b.AdjacentManufactureBonus(); got != 1.2 {
		t.Errorf("expected 1.2 with two adjacent workshops, got %f", got)
	}
}

func TestAdvanceMonthSalaryFunding(t *testing.T) {
	b := NewBase("Budget", 0)
	// NewBase: 4 soldiers, 10 scientists, 10 engineers, no radar.
	salary, funding := b.AdvanceMonth()
	// salary = (4 + 10 + 10) * 2000 = 48000
	if salary != 48000 {
		t.Errorf("expected salary 48000, got %d", salary)
	}
	// funding = 300000 base (no radar)
	if funding != 300000 {
		t.Errorf("expected base funding 300000, got %d", funding)
	}

	// A radar adds +75000 government funding.
	b.Facilities = append(b.Facilities, &Facility{Type: FacRadar})
	if got := b.GovernmentFunding(); got != 375000 {
		t.Errorf("expected 375000 funding with one radar, got %d", got)
	}

	// High alien activity reduces government funding (~40% at activity 100).
	b.AlienActivity = 100
	_, funded := b.AdvanceMonth()
	if funded != 150000 {
		t.Errorf("expected funding 150000 at full alien activity, got %d", funded)
	}
}

func TestInterrogateUnknownAlien(t *testing.T) {
	b := NewBase("Interrogate", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	b.LiveAliens = []string{"Sectoid"}
	// Interrogating an alien we do not hold must fail.
	if _, ok := b.InterrogateAlien("Floaters"); ok {
		t.Error("expected false interrogating an alien not in LiveAliens")
	}
}

func TestEquipWeaponInvalidIndex(t *testing.T) {
	b := NewBase("Equip", 0)
	b.AddItem("rifle", 1)
	if b.EquipWeapon(-1, "rifle") {
		t.Error("expected false for negative soldier index")
	}
	if b.EquipWeapon(999, "rifle") {
		t.Error("expected false for out-of-range soldier index")
	}
}

func TestEquipWeaponCapacityFull(t *testing.T) {
	b := NewBase("Cap", 0)
	// Fill storage to capacity; the soldier already holds a rifle (non-pistol),
	// so returning it would overflow and equipping must fail.
	b.UsedStorage = b.StorageCapacity()
	b.Stores["pistol"] = 1
	if b.EquipWeapon(0, "pistol") {
		t.Error("expected equip to fail when storage is full (cannot return old weapon)")
	}
}

func TestSellFacilityRefund(t *testing.T) {
	g := &engine.Game{Funds: 1000000, GameTime: time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC)}
	b := NewBase("Sell", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab, Building: false, Row: 1, Col: 1})
	bs := &BaseScreen{Game: g, Base: b, Tab: 0, Selection: len(b.Facilities) - 1}
	before := g.Funds
	bs.SellFacility()
	if g.Funds <= before {
		t.Errorf("expected funds to increase by refund, before=%d after=%d", before, g.Funds)
	}
}
