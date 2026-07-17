package data

// PlaneConfig defines a modular interceptor design.
// Players configure parameters; the game renders the plane procedurally.
type PlaneConfig struct {
	Length   int // 3-7, fuselage cells (nose + tail)
	Wingspan int // 1-4, wing cells each side of center
	Engines  int // 1-3
	Fuel     int // 20-100 (range in geoscape units)
	Weapon   int // 0-3 index into PlaneWeapons
	Armor    int // 0-3 index into PlaneArmors
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ClampPlaneConfig clamps all fields to their valid ranges.
func ClampPlaneConfig(cfg *PlaneConfig) {
	cfg.Length = clampInt(cfg.Length, 3, 7)
	cfg.Wingspan = clampInt(cfg.Wingspan, 1, 4)
	cfg.Engines = clampInt(cfg.Engines, 1, 3)
	cfg.Fuel = clampInt(cfg.Fuel, 20, 100)
	cfg.Weapon = clampInt(cfg.Weapon, 0, len(PlaneWeapons)-1)
	cfg.Armor = clampInt(cfg.Armor, 0, len(PlaneArmors)-1)
}

// DefaultPlaneConfig returns a balanced starter design.
func DefaultPlaneConfig() PlaneConfig {
	return PlaneConfig{
		Length:   5,
		Wingspan: 2,
		Engines:  2,
		Fuel:     50,
		Weapon:   0,
		Armor:    0,
	}
}

// PlaneWeaponDef defines a weapon that can be mounted on a player plane.
type PlaneWeaponDef struct {
	Name      string
	Damage    int
	Accuracy  int // base accuracy %
	Range     int // max range in geoscape units
	FireRate  int // shots per engagement
	Weight    int // adds to total mass
	Mass      float64
	Cost      int
	RearmCost int
}

// PlaneWeapons lists all available player plane weapons.
var PlaneWeapons = []PlaneWeaponDef{
	{Name: "Cannon", Damage: 15, Accuracy: 85, Range: 25, FireRate: 3, Weight: 2, Mass: 2.0, Cost: 5000, RearmCost: 500},
	{Name: "Stingray", Damage: 25, Accuracy: 70, Range: 45, FireRate: 2, Weight: 4, Mass: 4.0, Cost: 8000, RearmCost: 1000},
	{Name: "Avalanche", Damage: 40, Accuracy: 55, Range: 60, FireRate: 1, Weight: 6, Mass: 6.0, Cost: 12000, RearmCost: 1500},
	{Name: "Plasma", Damage: 60, Accuracy: 50, Range: 50, FireRate: 1, Weight: 8, Mass: 8.0, Cost: 20000, RearmCost: 3000},
}

// PlaneArmorDef defines hull plating for a player plane.
type PlaneArmorDef struct {
	Name      string
	HP        int // bonus hull points
	DR        int // damage reduction %
	Weight    int // adds to total mass
	Mass      float64
	Cost      int
}

// PlaneArmors lists all available player plane armors.
var PlaneArmors = []PlaneArmorDef{
	{Name: "None", HP: 0, DR: 0, Weight: 0, Mass: 0.0, Cost: 0},
	{Name: "Light Alloy", HP: 10, DR: 10, Weight: 2, Mass: 2.0, Cost: 8000},
	{Name: "Heavy Alloy", HP: 25, DR: 25, Weight: 5, Mass: 5.0, Cost: 18000},
	{Name: "Alien Plating", HP: 40, DR: 40, Weight: 7, Mass: 7.0, Cost: 35000},
}

// planePartMass defines base mass per fuselage cell.
const planePartMass = 3.0

// planeEngineMass defines mass per engine.
const planeEngineMass = 4.0

// planeWingMass defines mass per wing cell.
const planeWingMass = 1.5

// planeFuelMass defines mass per unit of fuel.
const planeFuelMass = 0.1

// PlaneStats holds derived stats computed from a PlaneConfig.
type PlaneStats struct {
	Speed      float64
	Firepower  float64
	Hull       int
	Mass       float64
	Thrust     float64
	Range      int
	FuelLeft   int
}

// CalcPlaneStats computes derived stats for a plane configuration.
func CalcPlaneStats(cfg PlaneConfig) PlaneStats {
	// Mass = fuselage + engines + wings + fuel + weapon + armor
	fuselageMass := float64(cfg.Length) * planePartMass
	engineMass := float64(cfg.Engines) * planeEngineMass
	wingMass := float64(cfg.Wingspan*2) * planeWingMass
	fuelMass := float64(cfg.Fuel) * planeFuelMass

	wpnMass := 0.0
	armorMass := 0.0
	if cfg.Weapon >= 0 && cfg.Weapon < len(PlaneWeapons) {
		wpnMass = PlaneWeapons[cfg.Weapon].Mass
	}
	if cfg.Armor >= 0 && cfg.Armor < len(PlaneArmors) {
		armorMass = PlaneArmors[cfg.Armor].Mass
	}

	totalMass := fuselageMass + engineMass + wingMass + fuelMass + wpnMass + armorMass
	if totalMass < 1 {
		totalMass = 1
	}

	// Thrust = engines * base thrust per engine
	thrust := float64(cfg.Engines) * 20.0

	// Speed = thrust / mass (higher is faster)
	speed := thrust / totalMass

	// Hull = base + length bonus + armor bonus
	hull := 30 + cfg.Length*5
	if cfg.Armor >= 0 && cfg.Armor < len(PlaneArmors) {
		hull += PlaneArmors[cfg.Armor].HP
	}

	// Firepower = weapon damage * fire rate
	firepower := 0.0
	if cfg.Weapon >= 0 && cfg.Weapon < len(PlaneWeapons) {
		w := PlaneWeapons[cfg.Weapon]
		firepower = float64(w.Damage * w.FireRate)
	}

	return PlaneStats{
		Speed:     speed,
		Firepower: firepower,
		Hull:      hull,
		Mass:      totalMass,
		Thrust:    thrust,
		Range:     cfg.Fuel,
		FuelLeft:  cfg.Fuel,
	}
}

// PlaneCell is a single character in the plane ASCII preview.
type PlaneCell struct {
	X, Y int
	Rune rune
}

// RenderPlanePreview generates block-art for a plane configuration.
// The plane faces right (nose on the right). All characters are BMP block/box
// drawing symbols — no diagonal triangle chars.
func RenderPlanePreview(cfg PlaneConfig) []PlaneCell {
	var cells []PlaneCell

	// Clamp parameters to renderable range.
	fuseLen := cfg.Length
	if fuseLen < 3 {
		fuseLen = 3
	}
	if fuseLen > 9 {
		fuseLen = 9
	}
	wingSpan := cfg.Wingspan
	if wingSpan < 1 {
		wingSpan = 1
	}
	if wingSpan > 5 {
		wingSpan = 5
	}

	// Layout: plane is drawn on a grid where row 0 is the fuselage centre.
	// Positive Y = below, negative Y = above.
	// Caller re-centres the bounding box, so we just use natural coordinates.

	noseX := 0         // leftmost X (nose tip)
	tailX := fuseLen   // rightmost fuselage cell (engines attach at tailX+1)
	wingMidX := fuseLen/2 + 1 // X column where wings are widest

	// ── Nose ──────────────────────────────────────────────────────────────────
	// Nose: left half-block pointing right
	cells = append(cells, PlaneCell{X: noseX, Y: 0, Rune: '\u258C'}) // ▌

	// ── Fuselage ──────────────────────────────────────────────────────────────
	for x := noseX + 1; x <= tailX; x++ {
		r := '\u2588' // █ solid block
		if x == noseX+2 && fuseLen > 3 {
			r = '\u25A3' // ▣ cockpit window
		}
		cells = append(cells, PlaneCell{X: x, Y: 0, Rune: r})
	}

	// ── Wings ─────────────────────────────────────────────────────────────────
	// Each wing row: a horizontal bar of ▄ (upper wing) / ▀ (lower wing).
	// Widest at wingMidX, tapering by 1 cell per row outward.
	for wy := 1; wy <= wingSpan; wy++ {
		// Wing span at this distance from fuselage tapers inward.
		wingLeft := noseX + 2
		wingRight := tailX - 1
		// Taper: remove one cell from each end per wing row after the first.
		taper := wy - 1
		wingLeft += taper
		wingRight -= taper
		if wingLeft > wingRight {
			// Draw at least the centre column as a stub.
			wingLeft = wingMidX
			wingRight = wingMidX
		}
		for x := wingLeft; x <= wingRight; x++ {
			cells = append(cells, PlaneCell{X: x, Y: -wy, Rune: '\u2580'}) // ▀ upper half-block
			cells = append(cells, PlaneCell{X: x, Y: wy, Rune: '\u2584'})  // ▄ lower half-block
		}
		// Wing leading-edge cap
		cells = append(cells, PlaneCell{X: wingLeft - 1, Y: -wy, Rune: '\u258C'}) // ▌
		cells = append(cells, PlaneCell{X: wingLeft - 1, Y: wy, Rune: '\u258C'})  // ▌
	}

	// ── Tail fins ─────────────────────────────────────────────────────────────
	// Small vertical fin stubs at the tail using half-blocks.
	finX := tailX - 1
	if finX < noseX+1 {
		finX = noseX + 1
	}
	cells = append(cells, PlaneCell{X: finX, Y: -(wingSpan + 1), Rune: '\u2590'}) // ▐
	cells = append(cells, PlaneCell{X: finX, Y: wingSpan + 1, Rune: '\u2590'})    // ▐

	// ── Engines ───────────────────────────────────────────────────────────────
	// Engines are solid blocks attached at the tail.
	engX := tailX + 1
	if cfg.Engines >= 1 {
		cells = append(cells, PlaneCell{X: engX, Y: 0, Rune: '\u25A0'})  // ■
	}
	if cfg.Engines >= 2 {
		cells = append(cells, PlaneCell{X: engX, Y: -1, Rune: '\u25A0'}) // ■
		cells = append(cells, PlaneCell{X: engX, Y: 1, Rune: '\u25A0'})  // ■
	}
	if cfg.Engines >= 3 {
		cells = append(cells, PlaneCell{X: engX + 1, Y: -1, Rune: '\u25A0'}) // ■
		cells = append(cells, PlaneCell{X: engX + 1, Y: 1, Rune: '\u25A0'})  // ■
	}

	// ── Weapon hardpoints ─────────────────────────────────────────────────────
	if cfg.Weapon >= 0 && cfg.Weapon < len(PlaneWeapons) {
		hpX := noseX + 3
		if hpX > tailX-1 {
			hpX = tailX - 1
		}
		// Weapon pod: a short dash on the wing tips.
		cells = append(cells, PlaneCell{X: hpX, Y: -(wingSpan + 1), Rune: '\u2501'}) // ━
		cells = append(cells, PlaneCell{X: hpX, Y: wingSpan + 1, Rune: '\u2501'})    // ━
	}

	return cells
}
