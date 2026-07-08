package data

// Damage types used by weapons and resisted/weak to by aliens.
const (
	DMG_PLASMA = iota
	DMG_LASER
	DMG_EXPLOSIVE
	DMG_MELEE
	DMG_KINETIC
	DMG_PSIONIC
)

// AlienType defines a species/variant encountered in tactical combat.
type AlienType struct {
	Name       string
	ShortName  string
	Icon       rune   // Display character for this alien type
	HP         int
	TU         int
	Accuracy   int
	Bravery    int
	Reactions  int
	Strength   int
	Psi        int
	Armour     int
	Weapon     string
	Points     int // kill XP
	Rank       int // 0=lowest
	Aggression int // 1-10 higher = more aggressive
	DamageType int // primary damage type this alien deals

	// Resistance and weakness: positive = resistance (% damage reduced),
	// negative = weakness (% extra damage). Values for each DMG_* type.
	ResistPlasma  int
	ResistLaser   int
	ResistExplosive int
	ResistMelee   int
	ResistKinetic int
	ResistPsionic int

	Lore string // autopsy flavor text
}

var AlienTypes = []AlienType{
	// Rank 0 - Low tier
	{
		Name: "Sectoid", ShortName: "SEC", Icon: '\u03A9',
		HP: 10, TU: 50, Accuracy: 55, Bravery: 40, Reactions: 50,
		Strength: 8, Psi: 40, Armour: 5, Weapon: "plasma_pistol",
		Points: 5, Rank: 0, Aggression: 3, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistKinetic: -10,
		Lore: "Small grey humanoid with an oversized cranium. Relies on psionic potential rather than physical prowess.",
	},
	{
		Name: "Sectoid Navigator", ShortName: "SEN", Icon: '\u03A9',
		HP: 11, TU: 52, Accuracy: 58, Bravery: 45, Reactions: 52,
		Strength: 8, Psi: 50, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 3, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistPsionic: 15,
		Lore: "A sectoid with enhanced psionic sensitivity, coordinating squad movements.",
	},
	{
		Name: "Sectoid Commander", ShortName: "SEC2", Icon: '\u03A9',
		HP: 14, TU: 55, Accuracy: 62, Bravery: 55, Reactions: 58,
		Strength: 9, Psi: 70, Armour: 8, Weapon: "plasma_rifle",
		Points: 15, Rank: 2, Aggression: 4, DamageType: DMG_PLASMA,
		ResistPlasma: 25, ResistPsionic: 30, ResistKinetic: -15,
		Lore: "The dominant mind in a sectoid brood. Can project psionic waves to control lesser species.",
	},
	// Rank 1 - Mid tier
	{
		Name: "Floater", ShortName: "FLT", Icon: '\u221E',
		HP: 15, TU: 55, Accuracy: 60, Bravery: 50, Reactions: 60,
		Strength: 12, Psi: 10, Armour: 10, Weapon: "plasma_rifle",
		Points: 8, Rank: 1, Aggression: 6, DamageType: DMG_PLASMA,
		ResistPlasma: 15, ResistExplosive: -20,
		Lore: "A mutilated humanoid kept alive by cybernetic implants. Hovers above the ground on anti-grav units.",
	},
	{
		Name: "Floater Navigator", ShortName: "FLN", Icon: '\u221E',
		HP: 16, TU: 58, Accuracy: 63, Bravery: 55, Reactions: 63,
		Strength: 13, Psi: 18, Armour: 11, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 6, DamageType: DMG_PLASMA,
		ResistPlasma: 15, ResistExplosive: -15, ResistPsionic: 10,
		Lore: "A floater with enhanced neural links, coordinating air-to-ground operations.",
	},
	{
		Name: "Floater Commander", ShortName: "FLC", Icon: '\u221E',
		HP: 20, TU: 62, Accuracy: 68, Bravery: 65, Reactions: 68,
		Strength: 15, Psi: 30, Armour: 14, Weapon: "plasma_rifle",
		Points: 20, Rank: 3, Aggression: 7, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistExplosive: -25, ResistPsionic: 20,
		Lore: "The most augmented of the floaters. Commands from above, raining plasma fire.",
	},
	{
		Name: "Chryssalid", ShortName: "CHR", Icon: '\u03C8',
		HP: 14, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 75,
		Strength: 18, Psi: 0, Armour: 8, Weapon: "chryssalid_claw",
		Points: 15, Rank: 1, Aggression: 10, DamageType: DMG_MELEE,
		ResistMelee: 30, ResistLaser: -20, ResistPlasma: -10,
		Lore: "Insectoid predator with razor-sharp claws. Victims rise as zombies within hours.",
	},
	{
		Name: "Chryssalid Queen", ShortName: "CHQ", Icon: '\u03C8',
		HP: 35, TU: 60, Accuracy: 75, Bravery: 100, Reactions: 80,
		Strength: 25, Psi: 0, Armour: 12, Weapon: "chryssalid_claw",
		Points: 35, Rank: 4, Aggression: 10, DamageType: DMG_MELEE,
		ResistMelee: 40, ResistExplosive: -30, ResistLaser: -20,
		Lore: "The brood mother. Larger and faster than her spawn, with a thick chitinous carapace.",
	},
	{
		Name: "Hyperworm", ShortName: "HYP", Icon: '\u2248',
		HP: 8, TU: 70, Accuracy: 50, Bravery: 30, Reactions: 65,
		Strength: 6, Psi: 0, Armour: 3, Weapon: "plasma_pistol",
		Points: 4, Rank: 0, Aggression: 5, DamageType: DMG_MELEE,
		ResistMelee: 20, ResistKinetic: -15,
		Lore: "Small parasitic worm-like creature. Swarms are the real threat.",
	},
	{
		Name: "Silacoid", ShortName: "SIL", Icon: '\u2593',
		HP: 20, TU: 40, Accuracy: 45, Bravery: 80, Reactions: 35,
		Strength: 16, Psi: 0, Armour: 20, Weapon: "plasma_pistol",
		Points: 10, Rank: 1, Aggression: 4, DamageType: DMG_KINETIC,
		ResistKinetic: 40, ResistExplosive: 20, ResistLaser: -25,
		Lore: "Crystalline organism with rock-like hide. Absorbs kinetic impacts with ease.",
	},
	// Rank 2 - High tier
	{
		Name: "Muton", ShortName: "MUT", Icon: '\u03A3',
		HP: 25, TU: 55, Accuracy: 55, Bravery: 70, Reactions: 50,
		Strength: 20, Psi: 0, Armour: 18, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 8, DamageType: DMG_PLASMA,
		ResistPlasma: 30, ResistMelee: 20, ResistLaser: -10,
		Lore: "Brutish green-skinned warriors bred for combat. Strong but not bright.",
	},
	{
		Name: "Muton Navigator", ShortName: "MUN", Icon: '\u03A3',
		HP: 27, TU: 58, Accuracy: 58, Bravery: 75, Reactions: 53,
		Strength: 21, Psi: 0, Armour: 19, Weapon: "plasma_rifle",
		Points: 16, Rank: 3, Aggression: 8, DamageType: DMG_PLASMA,
		ResistPlasma: 30, ResistMelee: 25, ResistPsionic: -15,
		Lore: "A muton with tactical awareness. Guides the squad with crude but effective strategies.",
	},
	{
		Name: "Muton Commander", ShortName: "MUC", Icon: '\u03A3',
		HP: 30, TU: 62, Accuracy: 62, Bravery: 85, Reactions: 58,
		Strength: 23, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 25, Rank: 4, Aggression: 9, DamageType: DMG_PLASMA,
		ResistPlasma: 35, ResistMelee: 30, ResistExplosive: -20,
		Lore: "The alpha of the muton pack. Its battle rage is legendary among alien forces.",
	},
	{
		Name: "Cyberdisc", ShortName: "CYB", Icon: '\u25CE',
		HP: 30, TU: 50, Accuracy: 65, Bravery: 100, Reactions: 60,
		Strength: 15, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 20, Rank: 2, Aggression: 7, DamageType: DMG_PLASMA,
		ResistPlasma: 25, ResistLaser: 15, ResistExplosive: -25,
		Lore: "Mechanical disc-shaped unit. Fires plasma in all directions while hovering.",
	},
	{
		Name: "Celatid", ShortName: "CEL", Icon: '\u25C7',
		HP: 12, TU: 60, Accuracy: 60, Bravery: 50, Reactions: 55,
		Strength: 10, Psi: 0, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 6, DamageType: DMG_LASER,
		ResistLaser: 30, ResistPlasma: -15,
		Lore: "Amorphous gelatinous creature that fires concentrated acid bolts.",
	},
	// Rank 3 - Elite
	{
		Name: "Ethereal", ShortName: "ETH", Icon: '\u03A8',
		HP: 18, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 70,
		Strength: 10, Psi: 80, Armour: 12, Weapon: "plasma_rifle",
		Points: 25, Rank: 4, Aggression: 5, DamageType: DMG_PSIONIC,
		ResistPsionic: 50, ResistPlasma: 15, ResistMelee: -20,
		Lore: "The puppet masters. Telepathic beings that control all other alien species.",
	},
	{
		Name: "Ethereal Navigator", ShortName: "ETN", Icon: '\u03A8',
		HP: 20, TU: 68, Accuracy: 73, Bravery: 100, Reactions: 73,
		Strength: 11, Psi: 90, Armour: 13, Weapon: "plasma_rifle",
		Points: 30, Rank: 5, Aggression: 4, DamageType: DMG_PSIONIC,
		ResistPsionic: 60, ResistPlasma: 15, ResistMelee: -25,
		Lore: "An ethereal attuned to the psionic frequencies of the command chain.",
	},
	{
		Name: "Ethereal Commander", ShortName: "ETC", Icon: '\u03A8',
		HP: 24, TU: 72, Accuracy: 78, Bravery: 100, Reactions: 78,
		Strength: 13, Psi: 100, Armour: 16, Weapon: "heavy_plasma",
		Points: 50, Rank: 6, Aggression: 4, DamageType: DMG_PSIONIC,
		ResistPsionic: 70, ResistPlasma: 20, ResistMelee: -30, ResistExplosive: -15,
		Lore: "The supreme psychic intelligence. Its death causes a psionic shockwave across the battlefield.",
	},
	// Rank 4 - Boss
	{
		Name: "Reaper", ShortName: "REAP", Icon: '\u2660',
		HP: 50, TU: 45, Accuracy: 50, Bravery: 100, Reactions: 40,
		Strength: 30, Psi: 0, Armour: 25, Weapon: "reaper_claw",
		Points: 30, Rank: 3, Aggression: 9, DamageType: DMG_MELEE,
		ResistMelee: 50, ResistKinetic: 30, ResistLaser: -20, ResistPlasma: -15,
		Lore: "Massive ambush predator. Can swallow a soldier whole.",
	},
	{
		Name: "Sectopod", ShortName: "SEC3", Icon: '\u229E',
		HP: 60, TU: 40, Accuracy: 70, Bravery: 100, Reactions: 50,
		Strength: 25, Psi: 0, Armour: 30, Weapon: "heavy_plasma",
		Points: 50, Rank: 5, Aggression: 8, DamageType: DMG_PLASMA,
		ResistPlasma: 40, ResistExplosive: 20, ResistLaser: 10, ResistMelee: -30,
		Lore: "Walking tank. The ultimate expression of alien military engineering.",
	},
}

func GetAlienByName(name string) *AlienType {
	for i := range AlienTypes {
		if AlienTypes[i].Name == name {
			return &AlienTypes[i]
		}
	}
	return nil
}

func GetAlienByRank(minRank int) *AlienType {
	var best *AlienType
	for i := range AlienTypes {
		if AlienTypes[i].Rank >= minRank {
			if best == nil || AlienTypes[i].Rank < best.Rank {
				best = &AlienTypes[i]
			}
		}
	}
	return best
}

// DamageTypeStr returns a human-readable name for a damage type constant.
func DamageTypeStr(dt int) string {
	switch dt {
	case DMG_PLASMA:
		return "Plasma"
	case DMG_LASER:
		return "Laser"
	case DMG_EXPLOSIVE:
		return "Explosive"
	case DMG_MELEE:
		return "Melee"
	case DMG_KINETIC:
		return "Kinetic"
	case DMG_PSIONIC:
		return "Psionic"
	default:
		return "Unknown"
	}
}

// Resist returns the resistance value for the given damage type.
func (at *AlienType) Resist(dmgType int) int {
	switch dmgType {
	case DMG_PLASMA:
		return at.ResistPlasma
	case DMG_LASER:
		return at.ResistLaser
	case DMG_EXPLOSIVE:
		return at.ResistExplosive
	case DMG_MELEE:
		return at.ResistMelee
	case DMG_KINETIC:
		return at.ResistKinetic
	case DMG_PSIONIC:
		return at.ResistPsionic
	default:
		return 0
	}
}

// AlienTypesByRank returns all alien types with the given rank.
func AlienTypesByRank(rank int) []*AlienType {
	var result []*AlienType
	for i := range AlienTypes {
		if AlienTypes[i].Rank == rank {
			result = append(result, &AlienTypes[i])
		}
	}
	return result
}
