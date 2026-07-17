package data

import (
	"strings"

	"github.com/taislin/termcom/internal/language"
)

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

// FireMode represents a weapon firing mode.
type FireMode int

const (
	FireModeAimed FireMode = iota // single aimed shot
	FireModeBurst                 // 3-round burst
	FireModeAuto                  // full auto (all remaining ammo)
)

func (fm FireMode) String() string {
	switch fm {
	case FireModeAimed:
		return language.String("FIRE_MODE_AIMED")
	case FireModeBurst:
		return language.String("FIRE_MODE_BURST")
	case FireModeAuto:
		return language.String("FIRE_MODE_AUTO")
	default:
		return "???"
	}
}

// HasMode returns true if the weapon supports the given fire mode.
func (r *RuleItem) HasMode(fm FireMode) bool {
	switch fm {
	case FireModeAimed:
		return true
	case FireModeBurst:
		return r.BurstSize > 0
	case FireModeAuto:
		return r.Auto
	default:
		return false
	}
}

// Modes returns the list of fire modes available for this weapon.
func (r *RuleItem) Modes() []FireMode {
	modes := []FireMode{FireModeAimed}
	if r.BurstSize > 0 {
		modes = append(modes, FireModeBurst)
	}
	if r.Auto {
		modes = append(modes, FireModeAuto)
	}
	return modes
}

// ModeTU returns the TU cost for the given fire mode.
func (r *RuleItem) ModeTU(fm FireMode) int {
	switch fm {
	case FireModeBurst:
		return r.TU * 3 / 2
	case FireModeAuto:
		return r.TU * 2
	default:
		return r.TU
	}
}

// ModeAccuracy returns the accuracy penalty for the given fire mode.
func (r *RuleItem) ModeAccuracy(fm FireMode) int {
	switch fm {
	case FireModeBurst:
		return 10
	case FireModeAuto:
		return 20
	default:
		return 0
	}
}

// ModeRounds returns how many rounds are fired per shot in the given mode.
// Returns -1 for auto (all remaining ammo).
func (r *RuleItem) ModeRounds(fm FireMode) int {
	switch fm {
	case FireModeBurst:
		return 3
	case FireModeAuto:
		return -1
	default:
		return 1
	}
}

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
	TU        int // time units to fire (aimed)
	Range     int
	AmmoMax   int
	AmmoCur   int
	Auto      bool // can full-auto fire
	BurstSize int  // rounds per burst (0 = no burst)
	Strength  int // min strength to use
	IsAmmo    bool
	IsAlien   bool
}

// Weapons is the runtime weapon state map (keyed by item ID).
// AmmoCur tracks current ammunition per weapon type in the global store.
var Weapons = map[string]RuleItem{}

// RuleItems contains the full definition of all items in the game.
// Note: ShortName values (MSC, PRM, PSI) are shared with Items map below.
// Code that indexes by ShortName across both maps will find both entries.
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
		BurstSize:  3,
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
		BurstSize:  3,
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
	"heavy_plasma": {
		Type:       "STR_HEAVY_PLASMA",
		Name:       "Heavy Plasma",
		ShortName:  "HPL",
		Weight:     6,
		CostBuy:    15000,
		CostSell:   11250,
		BattleType: BT_FIREARM,
		Damage:     70,
		Accuracy:   65,
		TU:         28,
		Range:      30,
		AmmoMax:    99,
		Strength:   18,
		IsAlien:    true,
	},
	"chryssalid_claw": {
		Type:       "STR_CHRYSSALID_CLAW",
		Name:       "Chryssalid Claw",
		ShortName:  "CHC",
		Weight:     0,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_MELEE,
		Damage:     35,
		Accuracy:   85,
		TU:         18,
		Range:      1,
		AmmoMax:    99,
		Strength:   15,
		IsAlien:    true,
	},
	"reaper_claw": {
		Type:       "STR_REAPER_CLAW",
		Name:       "Reaper Claw",
		ShortName:  "RCL",
		Weight:     0,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_MELEE,
		Damage:     50,
		Accuracy:   80,
		TU:         22,
		Range:      1,
		AmmoMax:    99,
		Strength:   25,
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
	"motion_scanner": {
		Type:       "STR_MOTION_SCANNER",
		Name:       "Motion Scanner",
		ShortName:  "MSC",
		Weight:     3,
		CostBuy:    5000,
		CostSell:   3750,
		BattleType: BT_SCANNER,
		TU:         10,
		Range:      15,
	},
	"proximity_mine": {
		Type:       "STR_PROXIMITY_MINE",
		Name:       "Proximity Mine",
		ShortName:  "PRM",
		Weight:     3,
		CostBuy:    4000,
		CostSell:   3000,
		BattleType: BT_PROXIMITYGRENADE,
		Damage:     60,
		TU:         20,
		Range:      1,
	},
	"psi_amp": {
		Type:       "STR_PSI_AMP",
		Name:       "Psi-Amplifier",
		ShortName:  "PSI",
		Weight:     2,
		CostBuy:    30000,
		CostSell:   22500,
		BattleType: BT_PSIAMP,
		TU:         20,
	},
	"alien_claw": {
		Type:       "STR_ALIEN_CLAW",
		Name:       "Alien Claw",
		ShortName:  "ACL",
		Weight:     0,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_MELEE,
		Damage:     25,
		Accuracy:   80,
		TU:         16,
		Range:      1,
		AmmoMax:    99,
		Strength:   10,
		IsAlien:    true,
	},
	"alien_fang": {
		Type:       "STR_ALIEN_FANG",
		Name:       "Alien Fang",
		ShortName:  "AFG",
		Weight:     0,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_MELEE,
		Damage:     45,
		Accuracy:   85,
		TU:         20,
		Range:      1,
		AmmoMax:    99,
		Strength:   18,
		IsAlien:    true,
	},
	"alien_blaster": {
		Type:       "STR_ALIEN_BLASTER",
		Name:       "Alien Blaster",
		ShortName:  "ABL",
		Weight:     3,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     30,
		Accuracy:   70,
		TU:         16,
		Range:      15,
		AmmoMax:    99,
		Strength:   8,
		IsAlien:    true,
	},
	"alien_cannon": {
		Type:       "STR_ALIEN_CANNON",
		Name:       "Alien Cannon",
		ShortName:  "ACN",
		Weight:     8,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     60,
		Accuracy:   65,
		TU:         26,
		Range:      25,
		AmmoMax:    99,
		Strength:   16,
		IsAlien:    true,
	},
	"alien_laser": {
		Type:       "STR_ALIEN_LASER",
		Name:       "Alien Laser",
		ShortName:  "ALZ",
		Weight:     3,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     35,
		Accuracy:   75,
		TU:         15,
		Range:      18,
		AmmoMax:    99,
		Strength:   8,
		IsAlien:    true,
	},
	"alien_heavy_laser": {
		Type:       "STR_ALIEN_HEAVY_LASER",
		Name:       "Alien Heavy Laser",
		ShortName:  "AHL",
		Weight:     6,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     65,
		Accuracy:   70,
		TU:         24,
		Range:      28,
		AmmoMax:    99,
		Strength:   14,
		IsAlien:    true,
	},
	"alien_grenade": {
		Type:       "STR_ALIEN_GRENADE",
		Name:       "Alien Grenade",
		ShortName:  "AGR",
		Weight:     2,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     45,
		Accuracy:   60,
		TU:         18,
		Range:      8,
		AmmoMax:    99,
		Strength:   6,
		IsAlien:    true,
	},
	"alien_rocket": {
		Type:       "STR_ALIEN_ROCKET",
		Name:       "Alien Rocket",
		ShortName:  "ARK",
		Weight:     8,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_FIREARM,
		Damage:     80,
		Accuracy:   55,
		TU:         28,
		Range:      30,
		AmmoMax:    1,
		Strength:   16,
		IsAlien:    true,
	},
	"alien_psi_bolt": {
		Type:       "STR_ALIEN_PSI_BOLT",
		Name:       "Psionic Bolt",
		ShortName:  "APB",
		Weight:     0,
		CostBuy:    0,
		CostSell:   0,
		BattleType: BT_PSIAMP,
		TU:         20,
		Range:      15,
		AmmoMax:    99,
		IsAlien:    true,
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
	"alien_grenade_item":{Name: "Alien Grenade", ShortName: "AGR", Weight: 1, Value: 4000, Alien: true},
	"medikit":     {Name: "Medi-Kit", ShortName: "MDK", Weight: 2, Value: 6000},
	"motion_scanner":{Name: "Motion Scanner", ShortName: "MSC", Weight: 3, Value: 5000},
	"proximity_mine":{Name: "Proximity Mine", ShortName: "PRM", Weight: 3, Value: 4000},
	"psi_amplifier":{Name: "Psi-Amplifier", ShortName: "PSI", Weight: 2, Value: 30000},
	"ufo_nav":     {Name: "UFO Navigation", ShortName: "NAV", Weight: 5, Value: 15000, Alien: true},
	"ufo_power":   {Name: "UFO Power Source", ShortName: "PWR", Weight: 8, Value: 20000, Alien: true},
	"ufo_weapon":  {Name: "UFO Weapon System", ShortName: "UFW", Weight: 6, Value: 18000, Alien: true},
	"ufo_armor":   {Name: "UFO Hull Plating", ShortName: "UHL", Weight: 7, Value: 16000, Alien: true},
}

// InterceptorWeapon defines interceptor weapon stats.
type InterceptorWeapon struct {
	Name       string
	Damage     int
	Accuracy   int // base accuracy %
	Range      int // max range in geoscape units
	FireRate   int // shots per engagement (TU equivalent)
	Cost       int
	RearmCost  int
}

var interceptorWeaponKeys = map[string]string{
	"avalanche": "WPN_INTERCEPTOR_AVALANCHE",
	"stingray":  "WPN_INTERCEPTOR_STINGRAY",
	"cannon":    "WPN_INTERCEPTOR_CANNON",
}

func (w InterceptorWeapon) DisplayName(key string) string {
	if langKey, ok := interceptorWeaponKeys[key]; ok {
		if s := language.String(langKey); s != langKey {
			return s
		}
	}
	return w.Name
}

var InterceptorWeapons = map[string]InterceptorWeapon{
	"avalanche": {
		Name:       "Avalanche Launchers",
		Damage:     40,
		Accuracy:   55,
		Range:      60,
		FireRate:   1,
		Cost:       12000,
		RearmCost:  1500,
	},
	"stingray": {
		Name:       "Stingray Missiles",
		Damage:     25,
		Accuracy:   70,
		Range:      45,
		FireRate:   2,
		Cost:       8000,
		RearmCost:  1000,
	},
	"cannon": {
		Name:       "Cannon (DEF-7)",
		Damage:     15,
		Accuracy:   85,
		Range:      25,
		FireRate:   3,
		Cost:       5000,
		RearmCost:  500,
	},
}

// CombatMode defines interceptor engagement behavior.
type CombatMode int

const (
	CombatAttack    CombatMode = iota // Aggressive: close range, max damage
	CombatCautious                    // Balanced: maintain distance
	CombatBreakoff                    // Defensive: disengage if damaged
)

func (cm CombatMode) String() string {
	switch cm {
	case CombatAttack:
		return language.String("COMBAT_MODE_ATTACK")
	case CombatCautious:
		return language.String("COMBAT_MODE_CAUTIOUS")
	case CombatBreakoff:
		return language.String("COMBAT_MODE_BREAKOFF")
	default:
		return language.String("UNKNOWN")
	}
}

// InterceptorState defines the persisted state of an interceptor.
type InterceptorState struct {
	ID         int
	Name       string
	WeaponKey  string
	HP         int
	MaxHP      int
	Ammo       int
	Status     string // raw key: "available", "active", "rearming", "damaged", "destroyed"
	PlaneConfig *PlaneConfig // modular design (nil = default)
}


var armorNameKeys = map[string]string{
	"none":        "ARMOR_NONE",
	"personal":    "ARMOR_PERSONAL",
	"light":       "ARMOR_LIGHT",
	"medium":      "ARMOR_MEDIUM",
	"heavy":       "ARMOR_HEAVY",
	"power_suit":  "ARMOR_POWER",
	"flight_suit": "ARMOR_FLIGHT",
}

var itemNameKeys = map[string]string{
	"alloys":          "ITEM_ALLOYS",
	"elerium":         "ITEM_ELERIUM",
	"alien_corpse":    "ITEM_ALIEN_CORPSE",
	"corpse_sect":     "ITEM_SECTOID_CORPSE",
	"corpse_float":    "ITEM_FLOATER_CORPSE",
	"corpse_muton":    "ITEM_MUTON_CORPSE",
	"corpse_ether":    "ITEM_ETHEREAL_CORPSE",
	"alien_grenade_item":   "ITEM_ALIEN_GRENADE",
	"medikit":         "ITEM_MEDI_KIT",
	"motion_scanner":  "ITEM_MOTION_SCANNER",
	"proximity_mine":  "ITEM_PROXIMITY_MINE",
	"psi_amplifier":   "ITEM_PSI_AMPLIFIER",
	"ufo_nav":         "ITEM_UFO_NAV",
	"ufo_power":       "ITEM_UFO_POWER",
	"ufo_weapon":      "ITEM_UFO_WEAPON",
	"ufo_armor":       "ITEM_UFO_ARMOR",
}

func (r RuleItem) DisplayName() string {
	key := "WPN_" + strings.TrimPrefix(r.Type, "STR_")
	if s := language.String(key); s != key {
		return s
	}
	return r.Name
}

func (a Armor) DisplayNameByKey(key string) string {
	if langKey, ok := armorNameKeys[key]; ok {
		if s := language.String(langKey); s != langKey {
			return s
		}
	}
	return a.Name
}

func ItemDisplayName(key string) string {
	if langKey, ok := itemNameKeys[key]; ok {
		if s := language.String(langKey); s != langKey {
			return s
		}
	}
	if item, ok := Items[key]; ok {
		return item.Name
	}
	return key
}

func init() {
	for k, v := range RuleItems {
		v.AmmoCur = v.AmmoMax
		Weapons[k] = v
	}
}
