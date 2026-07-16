package data

import "fmt"

// WeaponDesign defines a player-designed modular weapon.
type WeaponDesign struct {
	ID        string // unique key, e.g. "custom_0"
	BaseType  string // "pistol", "rifle"
	Barrel    int    // 0=short, 1=standard, 2=long, 3=extended
	Optics    int    // 0=none, 1=iron, 2=scope, 3=advanced
	Auto      bool   // full-auto fire mode
	AmmoType  int    // 0=standard, 1=AP, 2=incendiary, 3=exploding
	Stock     int    // 0=none, 1=light, 2=heavy
}

// Barrel definitions: each affects range, accuracy, TU cost, weight
type BarrelDef struct {
	Name       string
	RangeMod   int
	AccuracyMod int
	TUMod      int
	WeightMod  float64
	CostMod    int
}

var Barrels = []BarrelDef{
	{"Short", -3, -5, -2, -1.0, -200},
	{"Standard", 0, 0, 0, 0.0, 0},
	{"Long", +5, +5, +3, +1.5, +500},
	{"Extended", +10, +8, +6, +3.0, +1200},
}

// Optics definitions
type OpticsDef struct {
	Name       string
	AccuracyMod int
	TUMod      int
	WeightMod  float64
	CostMod    int
}

var OpticsList = []OpticsDef{
	{"None", 0, 0, 0.0, 0},
	{"Iron Sights", +3, +1, +0.2, +200},
	{"Scope", +8, +2, +0.5, +800},
	{"Advanced Optics", +15, +3, +1.0, +2000},
}

// Auto fire mode modifiers
var AutoFireMods = struct {
	AccuracyMod int
	RangeMod    int
	TUMod      int
	WeightMod  float64
	CostMod    int
}{
	AccuracyMod: -15,
	RangeMod:    -5,
	TUMod:      -8, // fires faster (less TU per shot)
	WeightMod:  +1.0,
	CostMod:    +600,
}

// Ammo type definitions
type AmmoTypeDef struct {
	Name      string
	DamageMod int
	TUMod    int
	WeightMod float64
	CostMod   int
	IsAlien   bool
}

var AmmoTypes = []AmmoTypeDef{
	{"Standard", 0, 0, 0.0, 0, false},
	{"AP", +5, +2, +0.5, +400, false},
	{"Incendiary", +8, +3, +0.8, +800, false},
	{"Explosive", +12, +4, +1.2, +1500, false},
}

// Stock definitions
type StockDef struct {
	Name       string
	AccuracyMod int
	TUMod      int
	WeightMod  float64
	CostMod    int
}

var Stocks = []StockDef{
	{"None", -5, -1, -0.5, -200},
	{"Light", 0, 0, 0.0, 0},
	{"Heavy", +5, +2, +1.5, +600},
}

// Base weapon templates
type baseWeaponTemplate struct {
	Name      string
	Damage    int
	Accuracy  int
	TU        int
	Range     int
	AmmoMax   int
	Strength  int
	Cost      int
	Weight    float64
	BattleType int
}

var baseTemplates = map[string]baseWeaponTemplate{
	"pistol": {
		Name: "Pistol", Damage: 15, Accuracy: 65, TU: 15, Range: 8,
		AmmoMax: 12, Strength: 5, Cost: 3000, Weight: 2.0, BattleType: BT_FIREARM,
	},
	"rifle": {
		Name: "Rifle", Damage: 22, Accuracy: 70, TU: 20, Range: 20,
		AmmoMax: 20, Strength: 10, Cost: 5000, Weight: 4.0, BattleType: BT_FIREARM,
	},
}

// CalcDesignStats computes the final stats for a weapon design.
func CalcDesignStats(d WeaponDesign) (damage, accuracy, tu, rng, ammoMax, strength int, weight float64, cost int) {
	base, ok := baseTemplates[d.BaseType]
	if !ok {
		base = baseTemplates["rifle"]
	}

	damage = base.Damage
	accuracy = base.Accuracy
	tu = base.TU
	rng = base.Range
	ammoMax = base.AmmoMax
	strength = base.Strength
	weight = base.Weight
	cost = base.Cost

	// Apply barrel
	if d.Barrel >= 0 && d.Barrel < len(Barrels) {
		b := Barrels[d.Barrel]
		rng += b.RangeMod
		accuracy += b.AccuracyMod
		tu += b.TUMod
		weight += b.WeightMod
		cost += b.CostMod
	}

	// Apply optics
	if d.Optics >= 0 && d.Optics < len(OpticsList) {
		o := OpticsList[d.Optics]
		accuracy += o.AccuracyMod
		tu += o.TUMod
		weight += o.WeightMod
		cost += o.CostMod
	}

	// Apply auto fire
	if d.Auto {
		accuracy += AutoFireMods.AccuracyMod
		rng += AutoFireMods.RangeMod
		tu += AutoFireMods.TUMod
		weight += AutoFireMods.WeightMod
		cost += AutoFireMods.CostMod
	}

	// Apply ammo type
	if d.AmmoType >= 0 && d.AmmoType < len(AmmoTypes) {
		a := AmmoTypes[d.AmmoType]
		damage += a.DamageMod
		tu += a.TUMod
		weight += a.WeightMod
		cost += a.CostMod
	}

	// Apply stock
	if d.Stock >= 0 && d.Stock < len(Stocks) {
		s := Stocks[d.Stock]
		accuracy += s.AccuracyMod
		tu += s.TUMod
		weight += s.WeightMod
		cost += s.CostMod
	}

	// Clamp minimums
	if damage < 1 {
		damage = 1
	}
	if accuracy < 10 {
		accuracy = 10
	}
	if tu < 5 {
		tu = 5
	}
	if rng < 1 {
		rng = 1
	}
	if ammoMax < 1 {
		ammoMax = 1
	}
	// Strength requirement scales with weight
	strength = int(weight * 2.5)
	if strength < 5 {
		strength = 5
	}

	return
}

// MakeDesignItem creates a RuleItem from a weapon design.
func MakeDesignItem(d WeaponDesign) RuleItem {
	damage, accuracy, tu, rng, ammoMax, strength, weight, cost := CalcDesignStats(d)
	return RuleItem{
		Type:       d.ID,
		Name:       WeaponDesignName(d),
		ShortName:  "CUS",
		Weight:     int(weight),
		CostBuy:    cost,
		CostSell:   cost / 2,
		BattleType: BT_FIREARM,
		Damage:     damage,
		Accuracy:   accuracy,
		TU:         tu,
		Range:      rng,
		AmmoMax:    ammoMax,
		AmmoCur:    ammoMax,
		Auto:       d.Auto,
		BurstSize:  1,
		Strength:   strength,
		IsAlien:    false,
	}
}

// WeaponDesignName generates a display name for a weapon design.
func WeaponDesignName(d WeaponDesign) string {
	base, ok := baseTemplates[d.BaseType]
	if !ok {
		base = baseTemplates["rifle"]
	}

	var parts []string

	// Barrel prefix
	if d.Barrel >= 0 && d.Barrel < len(Barrels) && Barrels[d.Barrel].Name != "Standard" {
		parts = append(parts, Barrels[d.Barrel].Name)
	}

	// Base weapon name
	parts = append(parts, base.Name)

	// Suffixes
	var suffix []string
	if d.Optics >= 2 {
		suffix = append(suffix, "Scoped")
	}
	if d.Auto {
		suffix = append(suffix, "Auto")
	}
	if d.AmmoType >= 1 && d.AmmoType < len(AmmoTypes) {
		suffix = append(suffix, AmmoTypes[d.AmmoType].Name)
	}
	if len(suffix) > 0 {
		parts = append(parts, "("+fmt.Sprintf("%s", suffix[0])+")")
	}

	return joinParts(parts)
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}

// WeaponDesignBarLabels returns display strings for each parameter's options.
func WeaponDesignBarLabels(d WeaponDesign) []string {
	barrel := "---"
	if d.Barrel >= 0 && d.Barrel < len(Barrels) {
		barrel = Barrels[d.Barrel].Name
	}
	optics := "---"
	if d.Optics >= 0 && d.Optics < len(OpticsList) {
		optics = OpticsList[d.Optics].Name
	}
	auto := "Semi"
	if d.Auto {
		auto = "Auto"
	}
	ammo := "---"
	if d.AmmoType >= 0 && d.AmmoType < len(AmmoTypes) {
		ammo = AmmoTypes[d.AmmoType].Name
	}
	stock := "---"
	if d.Stock >= 0 && d.Stock < len(Stocks) {
		stock = Stocks[d.Stock].Name
	}
	return []string{barrel, optics, auto, ammo, stock}
}
