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

// DefaultPlanePreview renders a top-down ASCII preview of a plane config.
// Returns a slice of (x, y, rune) for each character to draw.
type PlaneCell struct {
	X, Y int
	Rune rune
}

// Plane preview parameters
const (
	previewLen    = 7
	previewWing   = 4
	previewCenter = previewWing // Y offset for center line
)

// RenderPlanePreview generates the Unicode art for a plane configuration.
// The plane faces right (nose on the right).
func RenderPlanePreview(cfg PlaneConfig) []PlaneCell {
	var cells []PlaneCell
	length := cfg.Length
	wingspan := cfg.Wingspan

	// Normalize to preview scale
	fuseLen := length
	if fuseLen > previewLen {
		fuseLen = previewLen
	}
	if fuseLen < 3 {
		fuseLen = 3
	}
	wingSpan := wingspan
	if wingSpan > previewWing {
		wingSpan = previewWing
	}
	if wingSpan < 1 {
		wingSpan = 1
	}

	centerY := previewCenter
	noseX := 1
	tailX := noseX + fuseLen - 1
	wingStart := noseX + 2
	wingEnd := tailX - 1
	if wingEnd < wingStart {
		wingEnd = wingStart
	}

	// Fuselage (center line)
	for x := noseX; x <= tailX; x++ {
		cells = append(cells, PlaneCell{X: x, Y: centerY, Rune: '\u25A0'})
	}

	// Nose cone
	cells = append(cells, PlaneCell{X: noseX - 1, Y: centerY, Rune: '\u25B6'})

	// Cockpit window (second cell from nose)
	if fuseLen > 2 {
		cells = append(cells, PlaneCell{X: noseX + 1, Y: centerY, Rune: '\u25C6'})
	}

	// Wings (extending above and below center)
	for wy := 1; wy <= wingSpan; wy++ {
		for x := wingStart; x <= wingEnd; x++ {
			// Taper: outer wings are shorter
			if wy == wingSpan && (x == wingStart || x == wingEnd) {
				continue
			}
			cells = append(cells, PlaneCell{X: x, Y: centerY - wy, Rune: '\u25E3'})
			cells = append(cells, PlaneCell{X: x, Y: centerY + wy, Rune: '\u25E2'})
		}
		// Wing tips
		if wy < wingSpan {
			cells = append(cells, PlaneCell{X: wingStart - 1, Y: centerY - wy, Rune: '\u25E5'})
			cells = append(cells, PlaneCell{X: wingStart - 1, Y: centerY + wy, Rune: '\u25E4'})
		}
	}

	// Engines (at tail, offset from center)
	if cfg.Engines >= 1 {
		cells = append(cells, PlaneCell{X: tailX + 1, Y: centerY, Rune: '\u25CE'})
	}
	if cfg.Engines >= 2 {
		cells = append(cells, PlaneCell{X: tailX + 1, Y: centerY - 1, Rune: '\u25CE'})
		cells = append(cells, PlaneCell{X: tailX + 1, Y: centerY + 1, Rune: '\u25CE'})
	}
	if cfg.Engines >= 3 {
		cells = append(cells, PlaneCell{X: tailX + 2, Y: centerY, Rune: '\u25CE'})
	}

	// Tail fins
	cells = append(cells, PlaneCell{X: tailX - 1, Y: centerY - wingSpan - 1, Rune: '\u25BC'})
	cells = append(cells, PlaneCell{X: tailX - 1, Y: centerY + wingSpan + 1, Rune: '\u25B2'})

	// Weapons (under wings)
	if cfg.Weapon >= 0 && cfg.Weapon < len(PlaneWeapons) {
		weaponX := wingStart + 1
		if weaponX > wingEnd {
			weaponX = wingEnd
		}
		cells = append(cells, PlaneCell{X: weaponX, Y: centerY - wingSpan - 1, Rune: '\u25B2'})
		cells = append(cells, PlaneCell{X: weaponX, Y: centerY + wingSpan + 1, Rune: '\u25BC'})
	}

	return cells
}
