package data

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
)

// PartCategory classifies a vehicle component.
type PartCategory int

const (
	PartHull     PartCategory = iota // Armor / structure
	PartEngine                        // Thrust
	PartWeapon                        // Geoscape dogfight damage
	PartPowerCore                     // Energy — volatile, explodable
	PartCockpit                       // Navigation / flight control
)

// VehiclePartDef is the static definition of a vehicle component.
type VehiclePartDef struct {
	ID             string
	Name           string
	Category       PartCategory
	Mass           float64     // weight in tons
	Thrust         float64     // engine output
	GeoDamage      float64     // dogfight damage per shot
	BattlescapeHP  int         // hit points on the tactical map
	ExplodesOnDeath bool        // triggers AoE explosion when destroyed
	TacticalRune  rune         // glyph on the Battlescape
	Color          tcell.Color // display color
	CostBuy        int         // purchase cost in funds
	CostAlloys     int         // alien alloy cost
}

// VehiclePartDefs is the master catalogue of all available parts.
var VehiclePartDefs = map[string]*VehiclePartDef{
	"hull_light": {
		ID: "hull_light", Name: "Light Hull", Category: PartHull,
		Mass: 2.0, BattlescapeHP: 30, TacticalRune: '█', Color: tcell.NewRGBColor(140, 140, 140),
		CostBuy: 3000, CostAlloys: 2,
	},
	"hull_heavy": {
		ID: "hull_heavy", Name: "Heavy Hull", Category: PartHull,
		Mass: 5.0, BattlescapeHP: 60, TacticalRune: '█', Color: tcell.NewRGBColor(100, 100, 100),
		CostBuy: 6000, CostAlloys: 5,
	},
	"engine_standard": {
		ID: "engine_standard", Name: "Standard Engine", Category: PartEngine,
		Mass: 3.0, Thrust: 12.0, BattlescapeHP: 20, TacticalRune: '⌁', Color: tcell.NewRGBColor(0, 180, 255),
		CostBuy: 4000, CostAlloys: 3,
	},
	"engine_light": {
		ID: "engine_light", Name: "Light Engine", Category: PartEngine,
		Mass: 1.5, Thrust: 8.0, BattlescapeHP: 15, TacticalRune: '⌁', Color: tcell.NewRGBColor(100, 200, 255),
		CostBuy: 2500, CostAlloys: 2,
	},
	"engine_alien": {
		ID: "engine_alien", Name: "Alien Propulsion", Category: PartEngine,
		Mass: 2.0, Thrust: 20.0, BattlescapeHP: 25, TacticalRune: '▲', Color: tcell.NewRGBColor(0, 255, 180),
		CostBuy: 12000, CostAlloys: 8,
	},
	"weapon_cannon": {
		ID: "weapon_cannon", Name: "Cannon Pod", Category: PartWeapon,
		Mass: 1.5, GeoDamage: 15.0, BattlescapeHP: 15, TacticalRune: '╒', Color: tcell.NewRGBColor(255, 160, 0),
		CostBuy: 2000, CostAlloys: 1,
	},
	"weapon_laser": {
		ID: "weapon_laser", Name: "Laser Turret", Category: PartWeapon,
		Mass: 2.0, GeoDamage: 25.0, BattlescapeHP: 20, TacticalRune: '⌖', Color: tcell.NewRGBColor(255, 80, 80),
		CostBuy: 8000, CostAlloys: 4,
	},
	"weapon_plasma": {
		ID: "weapon_plasma", Name: "Plasma Cannon", Category: PartWeapon,
		Mass: 3.0, GeoDamage: 40.0, BattlescapeHP: 25, TacticalRune: '⌖', Color: tcell.NewRGBColor(200, 0, 255),
		CostBuy: 15000, CostAlloys: 10,
	},
	"power_fission": {
		ID: "power_fission", Name: "Fission Core", Category: PartPowerCore,
		Mass: 2.0, BattlescapeHP: 15, ExplodesOnDeath: true,
		TacticalRune: '⚙', Color: tcell.NewRGBColor(255, 255, 0),
		CostBuy: 5000, CostAlloys: 3,
	},
	"power_fusion": {
		ID: "power_fusion", Name: "Fusion Core", Category: PartPowerCore,
		Mass: 3.0, BattlescapeHP: 20, ExplodesOnDeath: true,
		TacticalRune: '⚙', Color: tcell.NewRGBColor(255, 200, 0),
		CostBuy: 12000, CostAlloys: 8,
	},
	"power_elirium": {
		ID: "power_elirium", Name: "Elerium Reactor", Category: PartPowerCore,
		Mass: 4.0, BattlescapeHP: 25, ExplodesOnDeath: true,
		TacticalRune: '⚙', Color: tcell.NewRGBColor(0, 255, 200),
		CostBuy: 20000, CostAlloys: 12,
	},
	"cockpit_standard": {
		ID: "cockpit_standard", Name: "Standard Cockpit", Category: PartCockpit,
		Mass: 1.0, BattlescapeHP: 10, TacticalRune: '◧', Color: tcell.NewRGBColor(200, 200, 255),
		CostBuy: 2000, CostAlloys: 1,
	},
	"cockpit_armored": {
		ID: "cockpit_armored", Name: "Armored Cockpit", Category: PartCockpit,
		Mass: 2.0, BattlescapeHP: 25, TacticalRune: '◧', Color: tcell.NewRGBColor(160, 160, 200),
		CostBuy: 5000, CostAlloys: 3,
	},
}

// PlacedPart is a part instance on a vehicle blueprint grid.
type PlacedPart struct {
	Def       *VehiclePartDef
	CurrentHP int
}

// VehicleBlueprint is a 2D grid of placed components.
// Coordinates are local to the ship (0,0 is top-left).
type VehicleBlueprint struct {
	Name  string
	Width int
	Height int
	Parts map[image.Point]*PlacedPart
}

// NewVehicleBlueprint creates an empty blueprint of the given size.
func NewVehicleBlueprint(name string, w, h int) *VehicleBlueprint {
	return &VehicleBlueprint{
		Name:   name,
		Width:  w,
		Height: h,
		Parts:  make(map[image.Point]*PlacedPart),
	}
}

// Place adds a part at (x,y). Overwrites any existing part at that position.
func (vb *VehicleBlueprint) Place(def *VehiclePartDef, x, y int) {
	vb.Parts[image.Point{X: x, Y: y}] = &PlacedPart{
		Def:       def,
		CurrentHP: def.BattlescapeHP,
	}
}

// Remove clears the part at (x,y).
func (vb *VehicleBlueprint) Remove(x, y int) {
	delete(vb.Parts, image.Point{X: x, Y: y})
}

// Get returns the part at (x,y), or nil if empty.
func (vb *VehicleBlueprint) Get(x, y int) *PlacedPart {
	return vb.Parts[image.Point{X: x, Y: y}]
}

// CalculateMass returns the total mass of all placed parts.
func (vb *VehicleBlueprint) CalculateMass() float64 {
	total := 0.0
	for _, p := range vb.Parts {
		total += p.Def.Mass
	}
	return total
}

// CalculateThrust returns the total thrust from all engine parts.
func (vb *VehicleBlueprint) CalculateThrust() float64 {
	total := 0.0
	for _, p := range vb.Parts {
		if p.Def.Category == PartEngine {
			total += p.Def.Thrust
		}
	}
	return total
}

// GetTopSpeed returns Thrust / Mass (0 if mass is 0).
func (vb *VehicleBlueprint) GetTopSpeed() float64 {
	mass := vb.CalculateMass()
	if mass <= 0 {
		return 0
	}
	return vb.CalculateThrust() / mass
}

// GetTotalFirepower sums GeoDamage of all weapon parts.
func (vb *VehicleBlueprint) GetTotalFirepower() float64 {
	total := 0.0
	for _, p := range vb.Parts {
		if p.Def.Category == PartWeapon {
			total += p.Def.GeoDamage
		}
	}
	return total
}

// GetTotalHP sums BattlescapeHP of all parts.
func (vb *VehicleBlueprint) GetTotalHP() int {
	total := 0
	for _, p := range vb.Parts {
		total += p.Def.BattlescapeHP
	}
	return total
}

// HasCockpit returns true if at least one cockpit part is placed.
func (vb *VehicleBlueprint) HasCockpit() bool {
	for _, p := range vb.Parts {
		if p.Def.Category == PartCockpit {
			return true
		}
	}
	return false
}

// HasEngine returns true if at least one engine part is placed.
func (vb *VehicleBlueprint) HasEngine() bool {
	for _, p := range vb.Parts {
		if p.Def.Category == PartEngine {
			return true
		}
	}
	return false
}

// IsContiguous checks that all placed parts are orthogonally connected.
func (vb *VehicleBlueprint) IsContiguous() bool {
	if len(vb.Parts) == 0 {
		return false
	}
	visited := make(map[image.Point]bool)
	start := image.Point{}
	for pt := range vb.Parts {
		start = pt
		break
	}
	queue := []image.Point{start}
	visited[start] = true
	dirs := []image.Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nb := image.Point{X: cur.X + d.X, Y: cur.Y + d.Y}
			if !visited[nb] {
				if _, ok := vb.Parts[nb]; ok {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
	}
	return len(visited) == len(vb.Parts)
}

// HasWeapon returns true if at least one weapon part is placed.
func (vb *VehicleBlueprint) HasWeapon() bool {
	for _, p := range vb.Parts {
		if p.Def.Category == PartWeapon {
			return true
		}
	}
	return false
}

// IsValid checks minimum requirements: at least 1 cockpit, 1 engine, 1 weapon, all parts connected.
func (vb *VehicleBlueprint) IsValid() error {
	if len(vb.Parts) == 0 {
		return fmt.Errorf("blueprint is empty")
	}
	if !vb.HasCockpit() {
		return fmt.Errorf("missing cockpit")
	}
	if !vb.HasEngine() {
		return fmt.Errorf("missing engine")
	}
	if !vb.IsContiguous() {
		return fmt.Errorf("parts are not contiguous")
	}
	return nil
}

// ExplosionRadius returns the blast radius in tiles for a part that explodes on death.
// Returns 0 for non-explosive parts.
func (pp *PlacedPart) ExplosionRadius() int {
	if !pp.Def.ExplodesOnDeath {
		return 0
	}
	switch pp.Def.ID {
	case "power_elirium":
		return 5
	case "power_fusion":
		return 4
	default:
		return 3
	}
}

// LootID returns the item ID dropped when this part is intact at mission end.
func (pp *PlacedPart) LootID() string {
	switch pp.Def.Category {
	case PartPowerCore:
		return "ufo_power_source"
	case PartEngine:
		return "ufo_engine"
	case PartWeapon:
		return "ufo_weapon_module"
	case PartCockpit:
		return "ufo_navigator"
	default:
		return ""
	}
}

// DisplayName returns "UFO (Unknown Type)" for unscouted craft,
// or the blueprint name once identified.
func (vb *VehicleBlueprint) DisplayName(scouted bool) string {
	if !scouted {
		return language.String("UFO_UNKNOWN")
	}
	return vb.Name
}

// ── Preset templates ────────────────────────────────────────────────

// GenerateScoutUFO builds a 5x5 alien Scout-class UFO blueprint.
func GenerateScoutUFO() *VehicleBlueprint {
	vb := NewVehicleBlueprint("Scout", 5, 5)

	vb.Place(VehiclePartDefs["hull_light"], 1, 0)
	vb.Place(VehiclePartDefs["hull_light"], 2, 0)
	vb.Place(VehiclePartDefs["hull_light"], 3, 0)

	vb.Place(VehiclePartDefs["hull_light"], 0, 1)
	vb.Place(VehiclePartDefs["cockpit_standard"], 1, 1)
	vb.Place(VehiclePartDefs["hull_light"], 2, 1)
	vb.Place(VehiclePartDefs["hull_light"], 3, 1)
	vb.Place(VehiclePartDefs["hull_light"], 4, 1)

	vb.Place(VehiclePartDefs["power_fission"], 0, 2)
	vb.Place(VehiclePartDefs["hull_light"], 1, 2)
	vb.Place(VehiclePartDefs["hull_light"], 2, 2)
	vb.Place(VehiclePartDefs["hull_light"], 3, 2)
	vb.Place(VehiclePartDefs["power_fission"], 4, 2)

	vb.Place(VehiclePartDefs["weapon_cannon"], 1, 3)
	vb.Place(VehiclePartDefs["weapon_cannon"], 3, 3)

	vb.Place(VehiclePartDefs["engine_light"], 1, 4)
	vb.Place(VehiclePartDefs["engine_light"], 3, 4)

	return vb
}

// GenerateSmallScout builds a tiny 3x3 alien scout drone.
func GenerateSmallScout() *VehicleBlueprint {
	vb := NewVehicleBlueprint("Small Scout", 3, 3)

	vb.Place(VehiclePartDefs["cockpit_standard"], 1, 0)
	vb.Place(VehiclePartDefs["hull_light"], 0, 1)
	vb.Place(VehiclePartDefs["power_fission"], 1, 1)
	vb.Place(VehiclePartDefs["hull_light"], 2, 1)
	vb.Place(VehiclePartDefs["engine_light"], 1, 2)

	return vb
}

// GenerateHeavyUFO builds a larger 7x5 Heavy-class UFO.
func GenerateHeavyUFO() *VehicleBlueprint {
	vb := NewVehicleBlueprint("Heavy", 7, 5)

	vb.Place(VehiclePartDefs["hull_heavy"], 2, 0)
	vb.Place(VehiclePartDefs["hull_heavy"], 3, 0)
	vb.Place(VehiclePartDefs["hull_heavy"], 4, 0)

	for x := 0; x < 7; x++ {
		if x == 3 {
			vb.Place(VehiclePartDefs["cockpit_armored"], x, 1)
		} else {
			vb.Place(VehiclePartDefs["hull_heavy"], x, 1)
		}
	}

	vb.Place(VehiclePartDefs["power_elirium"], 0, 2)
	vb.Place(VehiclePartDefs["hull_heavy"], 1, 2)
	vb.Place(VehiclePartDefs["weapon_plasma"], 2, 2)
	vb.Place(VehiclePartDefs["hull_heavy"], 3, 2)
	vb.Place(VehiclePartDefs["weapon_plasma"], 4, 2)
	vb.Place(VehiclePartDefs["hull_heavy"], 5, 2)
	vb.Place(VehiclePartDefs["power_elirium"], 6, 2)

	vb.Place(VehiclePartDefs["engine_alien"], 1, 3)
	vb.Place(VehiclePartDefs["hull_heavy"], 2, 3)
	vb.Place(VehiclePartDefs["hull_heavy"], 3, 3)
	vb.Place(VehiclePartDefs["hull_heavy"], 4, 3)
	vb.Place(VehiclePartDefs["engine_alien"], 5, 3)

	vb.Place(VehiclePartDefs["engine_alien"], 2, 4)
	vb.Place(VehiclePartDefs["engine_alien"], 4, 4)

	return vb
}

// GenerateInterceptor builds a player Interceptor blueprint.
func GenerateInterceptor() *VehicleBlueprint {
	vb := NewVehicleBlueprint("Interceptor", 7, 5)

	vb.Place(VehiclePartDefs["hull_light"], 3, 0)

	vb.Place(VehiclePartDefs["hull_light"], 2, 1)
	vb.Place(VehiclePartDefs["cockpit_armored"], 3, 1)
	vb.Place(VehiclePartDefs["hull_light"], 4, 1)

	vb.Place(VehiclePartDefs["weapon_laser"], 1, 2)
	vb.Place(VehiclePartDefs["hull_heavy"], 2, 2)
	vb.Place(VehiclePartDefs["power_fusion"], 3, 2)
	vb.Place(VehiclePartDefs["hull_heavy"], 4, 2)
	vb.Place(VehiclePartDefs["weapon_laser"], 5, 2)

	vb.Place(VehiclePartDefs["engine_standard"], 2, 3)
	vb.Place(VehiclePartDefs["engine_standard"], 4, 3)

	vb.Place(VehiclePartDefs["hull_light"], 2, 4)
	vb.Place(VehiclePartDefs["hull_light"], 4, 4)

	return vb
}

// ── Procedural UFO generation ───────────────────────────────────────

// UFOTier categorizes alien craft by threat level.
type UFOTier int

const (
	TierDrone   UFOTier = iota // 3x3, light armament
	TierScout                 // 5x5, moderate
	TierInterceptor           // 5x7, heavy
	TierBomber                // 7x5, devastating
	TierCarrier               // 7x7, endgame
)

// UFOTierLangName returns the translated display name for a UFO tier.
func UFOTierLangName(tier UFOTier) string {
	switch tier {
	case TierDrone:
		return language.String("UFO_CLASS_DRONE")
	case TierScout:
		return language.String("UFO_CLASS_SCOUT")
	case TierInterceptor:
		return language.String("UFO_CLASS_INTERCEPTOR")
	case TierBomber:
		return language.String("UFO_CLASS_BOMBER")
	case TierCarrier:
		return language.String("UFO_CLASS_CARRIER")
	default:
		return language.String("UFO_UNKNOWN")
	}
}

// UFOClassNames maps tiers to internal names.
var UFOClassNames = map[UFOTier]string{
	TierDrone:       "Drone",
	TierScout:       "Scout",
	TierInterceptor: "Interceptor",
	TierBomber:      "Bomber",
	TierCarrier:     "Carrier",
}

// tierParts defines the part pools available for each tier.
type tierConfig struct {
	width, height int
	hulls         []string // part IDs to pick hulls from
	engines       []string
	weapons       []string
	powers        []string
	cockpits      []string
	minWeapons    int
	minEngines    int
}

var tierConfigs = map[UFOTier]tierConfig{
	TierDrone: {
		width: 3, height: 3,
		hulls:    []string{"hull_light"},
		engines:  []string{"engine_light"},
		weapons:  []string{"weapon_cannon"},
		powers:   []string{"power_fission"},
		cockpits: []string{"cockpit_standard"},
		minWeapons: 1, minEngines: 1,
	},
	TierScout: {
		width: 5, height: 5,
		hulls:    []string{"hull_light", "hull_light", "hull_heavy"},
		engines:  []string{"engine_light", "engine_standard"},
		weapons:  []string{"weapon_cannon", "weapon_cannon", "weapon_laser"},
		powers:   []string{"power_fission", "power_fission", "power_fusion"},
		cockpits: []string{"cockpit_standard"},
		minWeapons: 1, minEngines: 1,
	},
	TierInterceptor: {
		width: 7, height: 5,
		hulls:    []string{"hull_light", "hull_heavy", "hull_heavy"},
		engines:  []string{"engine_standard", "engine_standard", "engine_alien"},
		weapons:  []string{"weapon_laser", "weapon_laser", "weapon_plasma"},
		powers:   []string{"power_fusion", "power_fusion", "power_elirium"},
		cockpits: []string{"cockpit_standard", "cockpit_armored"},
		minWeapons: 2, minEngines: 2,
	},
	TierBomber: {
		width: 7, height: 5,
		hulls:    []string{"hull_heavy", "hull_heavy", "hull_heavy"},
		engines:  []string{"engine_alien", "engine_standard"},
		weapons:  []string{"weapon_plasma", "weapon_plasma", "weapon_plasma"},
		powers:   []string{"power_elirium", "power_fusion"},
		cockpits: []string{"cockpit_armored"},
		minWeapons: 3, minEngines: 2,
	},
	TierCarrier: {
		width: 7, height: 7,
		hulls:    []string{"hull_heavy", "hull_heavy", "hull_heavy"},
		engines:  []string{"engine_alien", "engine_alien"},
		weapons:  []string{"weapon_plasma", "weapon_plasma"},
		powers:   []string{"power_elirium", "power_elirium", "power_elirium"},
		cockpits: []string{"cockpit_armored"},
		minWeapons: 2, minEngines: 3,
	},
}

// GenerateProceduralUFO creates a randomized UFO blueprint from a seed and tier.
// Each (seed, tier) pair produces a deterministic but unique layout.
func GenerateProceduralUFO(seed int64, tier UFOTier) *VehicleBlueprint {
	for i := 0; i < 10; i++ {
		rng := newRand(seed + int64(i))
		cfg := tierConfigs[tier]
		name := UFOTierLangName(tier)

		vb := NewVehicleBlueprint(name, cfg.width, cfg.height)

		// Step 1: fill the interior ring with hull
		hullPool := cfg.hulls
		for y := 0; y < cfg.height; y++ {
			for x := 0; x < cfg.width; x++ {
				if isEdge(x, y, cfg.width, cfg.height) {
					continue // skip edges — fill later
				}
				def := VehiclePartDefs[pickFrom(rng, hullPool)]
				vb.Place(def, x, y)
			}
		}

		// Step 2: place cockpit near center
		cx, cy := cfg.width/2, cfg.height/2
		vb.Place(VehiclePartDefs[pickFrom(rng, cfg.cockpits)], cx, cy)

		// Step 3: place engines along bottom edge
		engineSlots := edgeSlotsBottom(cfg.width, cfg.height)
		placed := 0
		for _, slot := range engineSlots {
			if placed >= cfg.minEngines {
				break
			}
			if vb.Get(slot.x, slot.y) != nil {
				continue
			}
			vb.Place(VehiclePartDefs[pickFrom(rng, cfg.engines)], slot.x, slot.y)
			placed++
		}

		// Step 4: place weapons along middle or top edges
		weaponSlots := edgeSlotsSides(cfg.width, cfg.height)
		placed = 0
		for _, slot := range weaponSlots {
			if placed >= cfg.minWeapons {
				break
			}
			if vb.Get(slot.x, slot.y) != nil {
				continue
			}
			vb.Place(VehiclePartDefs[pickFrom(rng, cfg.weapons)], slot.x, slot.y)
			placed++
		}

		// Step 5: place power cores — prefer side edges
		powerSlots := append(edgeSlotsLeft(cfg.width, cfg.height), edgeSlotsRight(cfg.width, cfg.height)...)
		placed = 0
		for _, slot := range powerSlots {
			if placed >= 1 {
				break
			}
			if vb.Get(slot.x, slot.y) != nil {
				continue
			}
			vb.Place(VehiclePartDefs[pickFrom(rng, cfg.powers)], slot.x, slot.y)
			placed++
		}

		if vb.IsValid() == nil && vb.HasWeapon() {
			return vb
		}
	}
	// Fallback to minimal valid blueprint
	return GenerateScoutUFO()
}

// ── RNG + layout helpers ────────────────────────────────────────────

// newRand creates a deterministic RNG source for vehicle generation.
func newRand(seed int64) *rand.Rand {
	// 0 is a special sentinel seed value, remapped to 42 for consistency.
	if seed == 0 {
		seed = 42
	}
	return rand.New(rand.NewSource(seed))
}

func pickFrom(rng *rand.Rand, pool []string) string {
	return pool[rng.Intn(len(pool))]
}

func isEdge(x, y, w, h int) bool {
	return x == 0 || x == w-1 || y == 0 || y == h-1
}

type slot struct{ x, y int }

func edgeSlotsBottom(w, h int) []slot {
	var s []slot
	for x := 1; x < w-1; x += 2 {
		s = append(s, slot{x, h - 1})
	}
	return s
}

func edgeSlotsSides(w, h int) []slot {
	var s []slot
	for y := 1; y < h-1; y += 2 {
		s = append(s, slot{0, y})
		s = append(s, slot{w - 1, y})
	}
	return s
}

func edgeSlotsLeft(w, h int) []slot {
	var s []slot
	for y := 1; y < h-1; y++ {
		s = append(s, slot{0, y})
	}
	return s
}

func edgeSlotsRight(w, h int) []slot {
	var s []slot
	for y := 1; y < h-1; y++ {
		s = append(s, slot{w - 1, y})
	}
	return s
}
