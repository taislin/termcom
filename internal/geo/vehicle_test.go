package geo

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
)

func TestGenerateScoutUFO(t *testing.T) {
	bp := data.GenerateScoutUFO()
	if err := bp.IsValid(); err != nil {
		t.Fatalf("Scout UFO should be valid: %v", err)
	}
	if bp.Name != "Scout" {
		t.Errorf("expected name 'Scout', got %q", bp.Name)
	}
	if bp.GetTopSpeed() <= 0 {
		t.Error("Scout should have positive speed")
	}
	if bp.GetTotalFirepower() <= 0 {
		t.Error("Scout should have weapons")
	}
	t.Logf("Scout: mass=%.1f thrust=%.1f speed=%.2f firepower=%.1f",
		bp.CalculateMass(), bp.CalculateThrust(), bp.GetTopSpeed(), bp.GetTotalFirepower())
}

func TestGenerateHeavyUFO(t *testing.T) {
	bp := data.GenerateHeavyUFO()
	if err := bp.IsValid(); err != nil {
		t.Fatalf("Heavy UFO should be valid: %v", err)
	}
	scout := data.GenerateScoutUFO()
	if bp.CalculateMass() <= scout.CalculateMass() {
		t.Error("Heavy should be heavier than Scout")
	}
	if bp.GetTotalFirepower() <= scout.GetTotalFirepower() {
		t.Error("Heavy should have more firepower than Scout")
	}
	t.Logf("Heavy: mass=%.1f thrust=%.1f speed=%.2f firepower=%.1f",
		bp.CalculateMass(), bp.CalculateThrust(), bp.GetTopSpeed(), bp.GetTotalFirepower())
}

func TestGenerateInterceptor(t *testing.T) {
	bp := data.GenerateInterceptor()
	if err := bp.IsValid(); err != nil {
		t.Fatalf("Interceptor should be valid: %v", err)
	}
	if !bp.HasCockpit() {
		t.Error("Interceptor should have a cockpit")
	}
	t.Logf("Interceptor: mass=%.1f thrust=%.1f speed=%.2f firepower=%.1f",
		bp.CalculateMass(), bp.CalculateThrust(), bp.GetTopSpeed(), bp.GetTotalFirepower())
}

func TestGenerateSmallScout(t *testing.T) {
	bp := data.GenerateSmallScout()
	if err := bp.IsValid(); err != nil {
		t.Fatalf("Small Scout should be valid: %v", err)
	}
	scout := data.GenerateScoutUFO()
	if bp.CalculateMass() >= scout.CalculateMass() {
		t.Error("Small Scout should be lighter than Scout")
	}
	t.Logf("Small Scout: mass=%.1f thrust=%.1f speed=%.2f firepower=%.1f",
		bp.CalculateMass(), bp.CalculateThrust(), bp.GetTopSpeed(), bp.GetTotalFirepower())
}

func TestBlueprintSpeedRelationship(t *testing.T) {
	scout := data.GenerateScoutUFO()
	small := data.GenerateSmallScout()
	if small.GetTopSpeed() <= scout.GetTopSpeed() {
		t.Error("smaller ships should be faster")
	}
}

func TestPartDefsLoaded(t *testing.T) {
	if len(data.VehiclePartDefs) < 10 {
		t.Errorf("expected at least 10 part definitions, got %d", len(data.VehiclePartDefs))
	}
}

func TestProceduralUFOGeneration(t *testing.T) {
	tiers := []data.UFOTier{data.TierDrone, data.TierScout, data.TierInterceptor, data.TierBomber, data.TierCarrier}
	for _, tier := range tiers {
		bp := data.GenerateProceduralUFO(12345, tier)
		if bp == nil {
			t.Fatalf("tier %d returned nil blueprint", tier)
		}
		if err := bp.IsValid(); err != nil {
			t.Errorf("tier %d blueprint invalid: %v", tier, err)
		}
		t.Logf("tier %d: %s mass=%.1f speed=%.2f firepower=%.1f",
			tier, bp.Name, bp.CalculateMass(), bp.GetTopSpeed(), bp.GetTotalFirepower())
	}
}

func TestProceduralUFODeterministic(t *testing.T) {
	a := data.GenerateProceduralUFO(999, data.TierScout)
	b := data.GenerateProceduralUFO(999, data.TierScout)
	if a.CalculateMass() != b.CalculateMass() || a.GetTopSpeed() != b.GetTopSpeed() {
		t.Error("same seed should produce identical blueprints")
	}
}

func TestProceduralUFODifferentSeeds(t *testing.T) {
	a := data.GenerateProceduralUFO(1, data.TierScout)
	b := data.GenerateProceduralUFO(2, data.TierScout)
	if a.CalculateMass() == b.CalculateMass() && a.GetTotalFirepower() == b.GetTotalFirepower() {
		t.Error("different seeds should produce different blueprints")
	}
}
