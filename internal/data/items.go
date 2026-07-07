package data

// Battle types corresponding to OpenXcom definitions.
const (
	BT_NONE = iota
	BT_FIREARM
	BT_AMMO
	BT_MELEE
	BT_GRENADE
	BT_PROXIMITYGRENADE
	BT_MEDIKIT
	BT_SCANNER
	BT_MINDPROBE
	BT_PSIAMP
	BT_FLARE
	BT_CORPSE
)

// RuleItem defines the attributes of an item based on the original OpenXcom definition.
type RuleItem struct {
	Type       string
	Name       string
	ShortName  string
	Weight     int
	CostBuy    int
	CostSell   int
	BattleType int

	// Weapon specific fields
	Damage    int
	Accuracy  int // base accuracy %
	TU        int // time units to fire
	Range     int
	AmmoMax   int
	AmmoCur   int
	Auto      bool // can burst fire
	BurstSize int
	Strength  int // min strength to use
	IsAmmo    bool
	IsAlien   bool
}

// Weapons is the runtime weapon state map (keyed by item ID).
// AmmoCur tracks current ammunition per weapon type in the global store.
var Weapons = map[string]RuleItem{}

// RuleItems contains the full definition of all items in the game.
var RuleItems = map[string]RuleItem{
	"pistol": {
		Type:       "STR_PISTOL",
		Name:       "Pistol",
		ShortName:  "PIS",
		Weight:     3,
		CostBuy:    800,
		CostSell:   600,
		BattleType: BT_FIREARM,
		Damage:     15,
		Accuracy:   65,
		TU:         15,
		Range:      8,
		AmmoMax:    12,
		Strength:   5,
	},
	"rifle": {
		Type:       "STR_RIFLE",
		Name:       "Rifle",
		ShortName:  "RIF",
		Weight:     6,
		CostBuy:    1500,
		CostSell:   1125,
		BattleType: BT_FIREARM,
		Damage:     22,
		Accuracy:   70,
		TU:         20,
		Range:      20,
		AmmoMax:    20,
		Strength:   10,
	},
	"heavy": {
		Type:       "STR_HEAVY_CANNON",
		Name:       "Heavy Cannon",
		ShortName:  "HVC",
		Weight:     10,
		CostBuy:    2200,
		CostSell:   1650,
		BattleType: BT_FIREARM,
		Damage:     35,
		Accuracy:   55,
		TU:         25,
		Range:      15,
		AmmoMax:    6,
		Strength:   18,
		IsAmmo:     true,
	},
	"auto": {
		Type:       "STR_AUTO_CANNON",
		Name:       "Auto Cannon",
		ShortName:  "AUC",
		Weight:     12,
		CostBuy:    2600,
		CostSell:   1950,
		BattleType: BT_FIREARM,
		Damage:     20,
		Accuracy:   60,
		TU:         25,
		Range:      18,
		AmmoMax:    18,
		Auto:       true,
		BurstSize:  3,
		Strength:   16,
	},
	"rocket": {
		Type:       "STR_ROCKET_LAUNCHER",
		Name:       "Rocket Launcher",
		ShortName:  "RKT",
		Weight:     10,
		CostBuy:    4000,
		CostSell:   3000,
		BattleType: BT_FIREARM,
		Damage:     80,
		Accuracy:   45,
		TU:         30,
		Range:      30,
		AmmoMax:    1,
		Strength:   20,
		IsAmmo:     true,
	},
	"laser_pistol": {
		Type:       "STR_LASER_PISTOL",
		Name:       "Laser Pistol",
		ShortName:  "LSP",
		Weight:     2,
		CostBuy:    6000,
		CostSell:   4500,
		BattleType: BT_FIREARM,
		Damage:     28,
		Accuracy:   75,
		TU:         12,
		Range:      12,
		AmmoMax:    99,
		Strength:   5,
		IsAlien:    false,
	},
	"laser_rifle": {
		Type:       "STR_LASER_RIFLE",
		Name:       "Laser Rifle",
		ShortName:  "LSR",
		Weight:     4,
		CostBuy:    8000,
		CostSell:   6000,
		BattleType: BT_FIREARM,
		Damage:     40,
		Accuracy:   80,
		TU:         18,
		Range:      25,
		AmmoMax:    99,
		Strength:   12,
	},
	"plasma_rifle": {
		Type:       "STR_PLASMA_RIFLE",
		Name:       "Plasma Rifle",
		ShortName:  "PLR",
		Weight:     4,
		CostBuy:    12000,
		CostSell:   9000,
		BattleType: BT_FIREARM,
		Damage:     55,
		Accuracy:   75,
		TU:         22,
		Range:      28,
		AmmoMax:    99,
		Strength:   14,
		IsAlien:    true,
	},
	"plasma_pistol": {
		Type:       "STR_PLASMA_PISTOL",
		Name:       "Plasma Pistol",
		ShortName:  "PLP",
		Weight:     2,
		CostBuy:    9000,
		CostSell:   6750,
		BattleType: BT_FIREARM,
		Damage:     30,
		Accuracy:   70,
		TU:         14,
		Range:      10,
		AmmoMax:    99,
		Strength:   6,
		IsAlien:    true,
	},
	"stun_rod": {
		Type:       "STR_STUN_ROD",
		Name:       "Stun Rod",
		ShortName:  "STR",
		Weight:     2,
		CostBuy:    2000,
		CostSell:   1500,
		BattleType: BT_MELEE,
		Damage:     10,
		Accuracy:   90,
		TU:         20,
		Range:      1,
		AmmoMax:    99,
		Strength:   10,
	},
	"medi_kit": {
		Type:       "STR_MEDI_KIT",
		Name:       "Medi-Kit",
		ShortName:  "MED",
		Weight:     2,
		CostBuy:    6000,
		CostSell:   4500,
		BattleType: BT_MEDIKIT,
		TU:         25,
		Range:      1,
		AmmoMax:    10,
		Strength:   5,
	},
}

type Armor struct {
	Name      string
	ShortName string
	Undersuit int // base armour value
	Health    int // bonus HP
	TUMod     int // TU penalty (%)
	Value     int // sell value
}

var Armors = map[string]Armor{
	"none": {
		Name: "None", ShortName: "---",
		Undersuit: 0, Health: 0, TUMod: 0,
	},
	"personal": {
		Name: "Personal Armour", ShortName: "PSA",
		Undersuit: 10, Health: 0, TUMod: 0, Value: 15000,
	},
	"light": {
		Name: "Light Suit", ShortName: "LIS",
		Undersuit: 20, Health: 0, TUMod: -5, Value: 35000,
	},
	"medium": {
		Name: "Medium Suit", ShortName: "MDS",
		Undersuit: 30, Health: 0, TUMod: -10, Value: 55000,
	},
	"heavy": {
		Name: "Heavy Suit", ShortName: "HVS",
		Undersuit: 40, Health: 0, TUMod: -15, Value: 75000,
	},
	"power_suit": {
		Name: "Power Suit", ShortName: "PWS",
		Undersuit: 50, Health: 0, TUMod: -10, Value: 100000,
	},
	"flight_suit": {
		Name: "Flying Suit", ShortName: "FLS",
		Undersuit: 45, Health: 0, TUMod: -5, Value: 140000,
	},
}

type Item struct {
	Name      string
	ShortName string
	Weight    int
	Value     int
	Alien     bool
}

var Items = map[string]Item{
	"alloys":      {Name: "Aluminium Alloys", ShortName: "ALY", Weight: 2, Value: 8000, Alien: true},
	"elerium":     {Name: "Elerium-115", ShortName: "ELR", Weight: 3, Value: 12000, Alien: true},
	"alien_corpse":{Name: "Alien Corpse", ShortName: "ALC", Weight: 10, Value: 2000, Alien: true},
	"corpse_sect": {Name: "Sectoid Corpse", ShortName: "SEC", Weight: 10, Value: 3000, Alien: true},
	"corpse_float":{Name: "Floater Corpse", ShortName: "FLT", Weight: 10, Value: 4000, Alien: true},
	"corpse_muton":{Name: "Muton Corpse", ShortName: "MUT", Weight: 15, Value: 6000, Alien: true},
	"corpse_ether":{Name: "Ethereal Corpse", ShortName: "ETH", Weight: 10, Value: 8000, Alien: true},
	"alien_grenade":{Name: "Alien Grenade", ShortName: "AGR", Weight: 1, Value: 4000, Alien: true},
	"medikit":     {Name: "Medi-Kit", ShortName: "MDK", Weight: 2, Value: 6000},
	"motion_scanner":{Name: "Motion Scanner", ShortName: "MSC", Weight: 3, Value: 5000},
	"psi_amplifier":{Name: "Psi-Amplifier", ShortName: "PSI", Weight: 2, Value: 30000},
}

func init() {
	for k, v := range RuleItems {
		v.AmmoCur = v.AmmoMax
		Weapons[k] = v
	}
}
