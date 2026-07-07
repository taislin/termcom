package data

type Weapon struct {
	Name       string
	ShortName  string
	Damage     int
	Accuracy   int // base accuracy %
	TU         int // time units to fire
	Range      int
	AmmoMax    int
	AmmoCur    int
	Auto       bool // can burst fire
	BurstSize  int
	Strength   int // min strength to use
	IsAmmo     bool
	IsAlien    bool
}

var Weapons = map[string]Weapon{
	"pistol": {
		Name: "Pistol", ShortName: "PIS",
		Damage: 15, Accuracy: 65, TU: 15, Range: 8,
		AmmoMax: 12, AmmoCur: 12, Strength: 5,
	},
	"rifle": {
		Name: "Rifle", ShortName: "RIF",
		Damage: 22, Accuracy: 70, TU: 20, Range: 20,
		AmmoMax: 20, AmmoCur: 20, Strength: 10,
	},
	"heavy": {
		Name: "Heavy Cannon", ShortName: "HVC",
		Damage: 35, Accuracy: 55, TU: 25, Range: 15,
		AmmoMax: 6, AmmoCur: 6, Strength: 18, IsAmmo: true,
	},
	"auto": {
		Name: "Auto Cannon", ShortName: "AUC",
		Damage: 20, Accuracy: 60, TU: 25, Range: 18,
		AmmoMax: 18, AmmoCur: 18, Auto: true, BurstSize: 3, Strength: 16,
	},
	"rocket": {
		Name: "Rocket Launcher", ShortName: "RKT",
		Damage: 80, Accuracy: 45, TU: 30, Range: 30,
		AmmoMax: 1, AmmoCur: 1, Strength: 20, IsAmmo: true,
	},
	"laser_pistol": {
		Name: "Laser Pistol", ShortName: "LSP",
		Damage: 28, Accuracy: 75, TU: 12, Range: 12,
		AmmoMax: 99, AmmoCur: 99, Strength: 5, IsAlien: false,
	},
	"laser_rifle": {
		Name: "Laser Rifle", ShortName: "LSR",
		Damage: 40, Accuracy: 80, TU: 18, Range: 25,
		AmmoMax: 99, AmmoCur: 99, Strength: 12,
	},
	"plasma_rifle": {
		Name: "Plasma Rifle", ShortName: "PLR",
		Damage: 55, Accuracy: 75, TU: 22, Range: 28,
		AmmoMax: 99, AmmoCur: 99, Strength: 14, IsAlien: true,
	},
	"plasma_pistol": {
		Name: "Plasma Pistol", ShortName: "PLP",
		Damage: 30, Accuracy: 70, TU: 14, Range: 10,
		AmmoMax: 99, AmmoCur: 99, Strength: 6, IsAlien: true,
	},
	"stun_rod": {
		Name: "Stun Rod", ShortName: "STR",
		Damage: 10, Accuracy: 90, TU: 20, Range: 1,
		AmmoMax: 99, AmmoCur: 99, Strength: 10,
	},
	"medi_kit": {
		Name: "Medi-Kit", ShortName: "MED",
		Damage: 0, Accuracy: 0, TU: 25, Range: 1,
		AmmoMax: 10, AmmoCur: 10, Strength: 5,
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
