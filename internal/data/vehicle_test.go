package data

import (
	"image"
	"testing"
)

func TestBlueprintMass(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	vb.Place(VehiclePartDefs["hull_light"], 0, 0)   // 2.0
	vb.Place(VehiclePartDefs["engine_standard"], 1, 0) // 3.0
	mass := vb.CalculateMass()
	if mass != 5.0 {
		t.Errorf("expected mass 5.0, got %f", mass)
	}
}

func TestBlueprintThrust(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	vb.Place(VehiclePartDefs["hull_light"], 0, 0)     // no thrust
	vb.Place(VehiclePartDefs["engine_standard"], 1, 0) // 12.0
	vb.Place(VehiclePartDefs["engine_light"], 2, 0)    // 8.0
	thrust := vb.CalculateThrust()
	if thrust != 20.0 {
		t.Errorf("expected thrust 20.0, got %f", thrust)
	}
}

func TestBlueprintSpeed(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	vb.Place(VehiclePartDefs["hull_light"], 0, 0)     // mass 2.0
	vb.Place(VehiclePartDefs["engine_standard"], 1, 0) // thrust 12.0, mass 3.0 -> total mass 5.0
	speed := vb.GetTopSpeed()
	if speed != 2.4 {
		t.Errorf("expected speed 2.4, got %f", speed)
	}
}

func TestBlueprintSpeedEmpty(t *testing.T) {
	vb := NewVehicleBlueprint("empty", 5, 5)
	speed := vb.GetTopSpeed()
	if speed != 0 {
		t.Errorf("expected speed 0 for empty blueprint, got %f", speed)
	}
}

func TestBlueprintFirepower(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	vb.Place(VehiclePartDefs["weapon_cannon"], 0, 0) // 15.0
	vb.Place(VehiclePartDefs["weapon_laser"], 1, 0)   // 25.0
	fp := vb.GetTotalFirepower()
	if fp != 40.0 {
		t.Errorf("expected firepower 40.0, got %f", fp)
	}
}

func TestBlueprintIsValid(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	if err := vb.IsValid(); err == nil {
		t.Error("empty blueprint should be invalid")
	}

	// Add cockpit only
	vb.Place(VehiclePartDefs["cockpit_standard"], 2, 2)
	if err := vb.IsValid(); err == nil {
		t.Error("blueprint without engine should be invalid")
	}

	// Add engine
	vb.Place(VehiclePartDefs["engine_standard"], 2, 3)
	if err := vb.IsValid(); err != nil {
		t.Errorf("blueprint with cockpit+engine should be valid: %v", err)
	}
}

func TestBlueprintContiguity(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	vb.Place(VehiclePartDefs["cockpit_standard"], 0, 0)
	vb.Place(VehiclePartDefs["engine_standard"], 4, 4) // not connected
	if vb.IsContiguous() {
		t.Error("disconnected parts should not be contiguous")
	}

	// Replace with connected chain
	vb2 := NewVehicleBlueprint("test2", 5, 5)
	vb2.Place(VehiclePartDefs["cockpit_standard"], 0, 0)
	vb2.Place(VehiclePartDefs["hull_light"], 1, 0)
	vb2.Place(VehiclePartDefs["hull_light"], 2, 0)
	vb2.Place(VehiclePartDefs["hull_light"], 3, 0)
	vb2.Place(VehiclePartDefs["engine_standard"], 4, 0)
	if !vb2.IsContiguous() {
		t.Error("connected parts should be contiguous")
	}
}

func TestPlaceAndRemove(t *testing.T) {
	vb := NewVehicleBlueprint("test", 5, 5)
	def := VehiclePartDefs["hull_light"]
	vb.Place(def, 2, 2)
	if vb.Get(2, 2) == nil {
		t.Fatal("expected part at (2,2)")
	}
	if vb.Get(2, 2).Def != def {
		t.Error("wrong part at (2,2)")
	}
	vb.Remove(2, 2)
	if vb.Get(2, 2) != nil {
		t.Error("expected nil after remove")
	}
}

func TestLootID(t *testing.T) {
	tests := []struct {
		partID string
		want   string
	}{
		{"power_fission", "ufo_power_source"},
		{"power_elirium", "ufo_power_source"},
		{"engine_standard", "ufo_engine"},
		{"weapon_cannon", "ufo_weapon_module"},
		{"cockpit_standard", "ufo_navigator"},
		{"hull_light", ""},
	}
	for _, tt := range tests {
		pp := &PlacedPart{Def: VehiclePartDefs[tt.partID]}
		got := pp.LootID()
		if got != tt.want {
			t.Errorf("LootID(%s) = %q, want %q", tt.partID, got, tt.want)
		}
	}
}

func TestExplosionRadius(t *testing.T) {
	tests := []struct {
		partID string
		want   int
	}{
		{"hull_light", 0},
		{"power_fission", 3},
		{"power_fusion", 4},
		{"power_elirium", 5},
	}
	for _, tt := range tests {
		pp := &PlacedPart{Def: VehiclePartDefs[tt.partID]}
		got := pp.ExplosionRadius()
		if got != tt.want {
			t.Errorf("ExplosionRadius(%s) = %d, want %d", tt.partID, got, tt.want)
		}
	}
}

func TestPartDefsExist(t *testing.T) {
	expected := []string{
		"hull_light", "hull_heavy",
		"engine_standard", "engine_light", "engine_alien",
		"weapon_cannon", "weapon_laser", "weapon_plasma",
		"power_fission", "power_fusion", "power_elirium",
		"cockpit_standard", "cockpit_armored",
	}
	for _, id := range expected {
		if _, ok := VehiclePartDefs[id]; !ok {
			t.Errorf("missing part definition: %s", id)
		}
	}
}

func TestPartCategories(t *testing.T) {
	categories := map[PartCategory][]string{}
	for id, def := range VehiclePartDefs {
		categories[def.Category] = append(categories[def.Category], id)
	}
	if len(categories[PartHull]) < 2 {
		t.Error("expected at least 2 hull parts")
	}
	if len(categories[PartEngine]) < 2 {
		t.Error("expected at least 2 engine parts")
	}
	if len(categories[PartWeapon]) < 2 {
		t.Error("expected at least 2 weapon parts")
	}
	if len(categories[PartPowerCore]) < 2 {
		t.Error("expected at least 2 power core parts")
	}
	if len(categories[PartCockpit]) < 2 {
		t.Error("expected at least 2 cockpit parts")
	}
}

func TestBlueprintGridBounds(t *testing.T) {
	vb := NewVehicleBlueprint("test", 3, 3)
	// Place at edge
	vb.Place(VehiclePartDefs["hull_light"], 0, 0)
	vb.Place(VehiclePartDefs["hull_light"], 2, 2)
	if len(vb.Parts) != 2 {
		t.Errorf("expected 2 parts, got %d", len(vb.Parts))
	}
	// Get out of bounds
	if vb.Get(-1, 0) != nil {
		t.Error("expected nil for out-of-bounds get")
	}
	if vb.Get(0, 3) != nil {
		t.Error("expected nil for out-of-bounds get")
	}
}

func TestOverwritePart(t *testing.T) {
	vb := NewVehicleBlueprint("test", 3, 3)
	vb.Place(VehiclePartDefs["hull_light"], 1, 1)
	vb.Place(VehiclePartDefs["weapon_cannon"], 1, 1)
	p := vb.Get(1, 1)
	if p == nil || p.Def.ID != "weapon_cannon" {
		t.Error("expected overwrite to place weapon_cannon")
	}
}

func TestImagePointKeys(t *testing.T) {
	vb := NewVehicleBlueprint("test", 3, 3)
	vb.Place(VehiclePartDefs["hull_light"], 1, 2)
	pt := image.Point{X: 1, Y: 2}
	if _, ok := vb.Parts[pt]; !ok {
		t.Error("expected part at image.Point{1,2}")
	}
}
