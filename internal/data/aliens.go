package data

import (
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
)

// Damage types used by weapons and resisted/weak to by aliens.
const (
	DMG_PLASMA = iota
	DMG_LASER
	DMG_EXPLOSIVE
	DMG_MELEE
	DMG_KINETIC
	DMG_PSIONIC
)

const ColorDefaultAlien = tcell.Color(9)

// AlienIconsByDamage maps each damage type to a pool of display glyphs (BMP runes).
// Glyphs are themed by damage type so the unit's look hints at its threat:
//
//	psionic  -> Greek        (psychic puppet masters)
//	melee    -> Runic        (primal/ancient predators)
//	kinetic  -> Cyrillic     (crystalline brutes)
//	plasma   -> Geometric    (energy-tech units)
//	laser    -> Misc Technical
//	explosive-> Starburst     (detonation, not hazard symbols)
//
// All glyphs are within the BMP (U+0000-U+FFFF) and contain no emoji.
var AlienIconsByDamage = map[int][]rune{
	DMG_PSIONIC:   {'Ω', 'Ψ', 'Σ', 'Φ', 'Θ', 'Ξ', 'Λ', 'Δ', 'Π', 'Ϙ', 'Ϡ', 'Ϟ'},
	DMG_MELEE:     {'ᚠ', 'ᚢ', 'ᚦ', 'ᚨ', 'ᚱ', 'ᚲ', 'ᚷ', 'ᚹ', 'ᚺ', 'ᛁ', 'ᛄ', 'ᛇ', 'ᛈ', 'ᛊ', 'ᛏ', 'ᛒ', 'ᛖ', 'ᛚ', 'ᛞ', 'ᛟ'},
	DMG_KINETIC:   {'Ф', 'Ц', 'Ч', 'Ш', 'Щ', 'Ъ', 'Ы', 'Э', 'Ю', 'Я', 'Ѱ', 'Ѡ', 'Ѫ', 'Ѭ'},
	DMG_PLASMA:    {'◈', '◎', '◆', '◇', '◊', '⊚', '⊛', '⊜', '◆', '◉', '◇', '❖', '◆', '◆', '⊞', '⊠', '⊡'},
	DMG_LASER:     {'⌖', '⌬', '⌭', '⌮', '⌑', '◊', '⎕', '⏃', '⏄', '⏚', '⏛', '⌒'},
	DMG_EXPLOSIVE: {'✸', '✺', '❋', '❂', '✶', '✷', '✴', '✦', '✧', '❉', '✱', '✲'},
}

var (
	fallbackIconCounter int // counter for deterministic fallback when pools are exhausted
)

// nextIcon returns the next unused glyph for the given damage type, recording it in
// used so each alien in a game gets a distinct on-map character. used should be seeded
// with the hardcoded alien icons (see UsedHardcodedIcons). If the damage-type pool is
// exhausted it falls back to any unused glyph from the other pools, keeping icons unique
// as long as the total pool capacity exceeds the aliens generated in a game.
func nextIcon(dmg int, used map[rune]bool) rune {
	if pool, ok := AlienIconsByDamage[dmg]; ok {
		for _, r := range pool {
			if !used[r] {
				used[r] = true
				return r
			}
		}
	}
	for _, pool := range AlienIconsByDamage {
		for _, r := range pool {
			if !used[r] {
				used[r] = true
				return r
			}
		}
	}
	// Truly exhausted (should not happen with current pool sizes): deterministic reuse
	// using a counter rather than map length to avoid mid-run collisions.
	r := '?'
	for _, pool := range AlienIconsByDamage {
		if len(pool) > 0 {
			r = pool[fallbackIconCounter%len(pool)]
			fallbackIconCounter++
			break
		}
	}
	used[r] = true
	return r
}

// UsedHardcodedIcons returns a set of glyphs already taken by the hardcoded alien roster,
// to seed per-game procedural icon assignment.
func UsedHardcodedIcons() map[rune]bool {
	used := make(map[rune]bool)
	for _, at := range AlienTypes {
		used[at.Icon] = true
	}
	return used
}

// Body type constants.
const (
	BodyOrganic   = "organic"
	BodySynthetic = "synthetic"
)

// Body subtype constants.
const (
	SubtypeCarbonFlesh  = "carbon_flesh"
	SubtypeSilicon      = "silicon_based"
	SubtypeGaseous      = "gaseous"
	SubtypeCrystalline  = "crystalline"
	SubtypeAmorphous    = "amorphous"
	SubtypeMechanical   = "mechanical"
	SubtypeBioSynthetic = "bio_synthetic"
	SubtypeNanotech     = "nanotech"
)

// Sense quality constants.
const (
	SenseNone      = "none"
	SensePoor      = "poor"
	SenseNormal    = "normal"
	SenseExcellent = "excellent"
	SenseMultiSpec = "multi_spectrum"
	SenseEcholoc   = "echolocation"
	SenseLow       = "low"
	SenseHigh      = "high"
)

// Morphology describes the physical form of an alien species.
type Morphology struct {
	Arms          int    // 0-6 manipulative limbs
	Legs          int    // 0-8 locomotive limbs (0=floating)
	BodyType      string // "organic" | "synthetic"
	BodySubtype   string // e.g. "carbon_flesh", "mechanical"
	Eyesight      string // sense quality
	Hearing       string // sense quality
	ThermalSense  string // "none" | "low" | "high"
	PsionicSense  string // "none" | "low" | "high"
	ChemicalSense string // "none" | "low" | "high"
	DamageType    int    // preferred damage affinity; drives procedural weapon-mask styling
}

// OrganicSubtypes lists valid organic body subtypes.
var OrganicSubtypes = []string{
	SubtypeCarbonFlesh, SubtypeSilicon, SubtypeGaseous,
	SubtypeCrystalline, SubtypeAmorphous,
}

// SyntheticSubtypes lists valid synthetic body subtypes.
var SyntheticSubtypes = []string{
	SubtypeMechanical, SubtypeBioSynthetic, SubtypeNanotech,
}

// IsFloating returns true if the alien has no legs (levitates/slithers).
func (m *Morphology) IsFloating() bool { return m.Legs == 0 }

// IsLarge returns true if the alien has 4+ legs (big silhouette).
func (m *Morphology) IsLarge() bool { return m.Legs >= 4 }

// MultiArmed returns true if the alien has 3+ arms.
func (m *Morphology) MultiArmed() bool { return m.Arms >= 3 }

// AlienType defines a species/variant encountered in tactical combat.
type AlienType struct {
	Name       string
	ShortName  string
	Icon       rune // Display character for this alien type
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
	ResistPlasma    int
	ResistLaser     int
	ResistExplosive int
	ResistMelee     int
	ResistKinetic   int
	ResistPsionic   int

	AutopsyID  string      // research ID that unlocks this alien's stats in battlescape sidebar
	Lore       string      // autopsy flavor text
	Morphology *Morphology // physical form (nil for hardcoded aliens)
	Style      tcell.Style // Visual style on battlescape map
	FgColor    tcell.Color // Foreground color for bloom / VFX (mirrors Style fg)
}

// LangName returns the localized alien name, falling back to the English Name
// when no translation key is registered (e.g. procedurally generated aliens).
func (a *AlienType) LangName() string {
	key := "ALIEN_" + strings.ToUpper(strings.ReplaceAll(a.Name, " ", "_"))
	if s := language.String(key); s != key {
		return s
	}
	return a.Name
}

var AlienTypes = []AlienType{
	// Rank 0 - Low tier
	{
		Name: "Sectoid", ShortName: "SEC", Icon: '◈',
		HP: 10, TU: 50, Accuracy: 55, Bravery: 40, Reactions: 50,
		Strength: 8, Psi: 40, Armour: 5, Weapon: "plasma_pistol",
		Points: 5, Rank: 0, Aggression: 3, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistKinetic: -10,
		AutopsyID: "sectoid_autopsy",
		Lore:      "Small grey humanoid with an oversized cranium. Relies on psionic potential rather than physical prowess.",
	},
	{
		Name: "Sectoid Navigator", ShortName: "SEN", Icon: '◎',
		HP: 11, TU: 52, Accuracy: 58, Bravery: 45, Reactions: 52,
		Strength: 8, Psi: 50, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 3, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistPsionic: 15,
		AutopsyID: "sectoid_autopsy",
		Lore:      "A sectoid with enhanced psionic sensitivity, coordinating squad movements.",
	},
	{
		Name: "Sectoid Commander", ShortName: "SEC2", Icon: '◆',
		HP: 14, TU: 55, Accuracy: 62, Bravery: 55, Reactions: 58,
		Strength: 9, Psi: 70, Armour: 8, Weapon: "plasma_rifle",
		Points: 15, Rank: 2, Aggression: 4, DamageType: DMG_PLASMA,
		ResistPlasma: 25, ResistPsionic: 30, ResistKinetic: -15,
		AutopsyID: "sectoid_autopsy",
		Lore:      "The dominant mind in a sectoid brood. Can project psionic waves to control lesser species.",
	},
	// Rank 1 - Mid tier
	{
		Name: "Floater", ShortName: "FLT", Icon: '◇',
		HP: 15, TU: 55, Accuracy: 60, Bravery: 50, Reactions: 60,
		Strength: 12, Psi: 10, Armour: 10, Weapon: "plasma_rifle",
		Points: 8, Rank: 1, Aggression: 6, DamageType: DMG_PLASMA,
		ResistPlasma: 15, ResistExplosive: -20,
		AutopsyID: "floater_autopsy",
		Lore:      "A mutilated humanoid kept alive by cybernetic implants. Hovers above the ground on anti-grav units.",
	},
	{
		Name: "Floater Navigator", ShortName: "FLN", Icon: '⬢',
		HP: 16, TU: 58, Accuracy: 63, Bravery: 55, Reactions: 63,
		Strength: 13, Psi: 18, Armour: 11, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 6, DamageType: DMG_PLASMA,
		ResistPlasma: 15, ResistExplosive: -15, ResistPsionic: 10,
		AutopsyID: "floater_autopsy",
		Lore:      "A floater with enhanced neural links, coordinating air-to-ground operations.",
	},
	{
		Name: "Floater Commander", ShortName: "FLC", Icon: '⊚',
		HP: 20, TU: 62, Accuracy: 68, Bravery: 65, Reactions: 68,
		Strength: 15, Psi: 30, Armour: 14, Weapon: "plasma_rifle",
		Points: 20, Rank: 3, Aggression: 7, DamageType: DMG_PLASMA,
		ResistPlasma: 20, ResistExplosive: -25, ResistPsionic: 20,
		AutopsyID: "floater_autopsy",
		Lore:      "The most augmented of the floaters. Commands from above, raining plasma fire.",
	},
	{
		Name: "Chryssalid", ShortName: "CHR", Icon: 'ᚠ',
		HP: 14, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 75,
		Strength: 18, Psi: 0, Armour: 8, Weapon: "chryssalid_claw",
		Points: 15, Rank: 1, Aggression: 10, DamageType: DMG_MELEE,
		ResistMelee: 30, ResistLaser: -20, ResistPlasma: -10,
		Lore: "Insectoid predator with razor-sharp claws. Victims rise as zombies within hours.",
	},
	{
		Name: "Chryssalid Queen", ShortName: "CHQ", Icon: 'ᚦ',
		HP: 35, TU: 60, Accuracy: 75, Bravery: 100, Reactions: 80,
		Strength: 25, Psi: 0, Armour: 12, Weapon: "chryssalid_claw",
		Points: 35, Rank: 4, Aggression: 10, DamageType: DMG_MELEE,
		ResistMelee: 40, ResistExplosive: -30, ResistLaser: -20,
		Lore: "The brood mother. Larger and faster than her spawn, with a thick chitinous carapace.",
	},
	{Name: "Hyperworm", ShortName: "HYP", Icon: 'ᚨ',
		HP: 8, TU: 70, Accuracy: 50, Bravery: 30, Reactions: 65,
		Strength: 6, Psi: 0, Armour: 3, Weapon: "plasma_pistol",
		Points: 4, Rank: 0, Aggression: 5, DamageType: DMG_PLASMA,
		ResistMelee: 20, ResistKinetic: -15,
		Lore: "Small parasitic worm-like creature. Swarms are the real threat.",
	},
	{
		Name: "Silacoid", ShortName: "SIL", Icon: 'Ф',
		HP: 20, TU: 40, Accuracy: 45, Bravery: 80, Reactions: 35,
		Strength: 16, Psi: 0, Armour: 20, Weapon: "plasma_pistol",
		Points: 10, Rank: 1, Aggression: 4, DamageType: DMG_KINETIC,
		ResistKinetic: 40, ResistExplosive: 20, ResistLaser: -25,
		Lore: "Crystalline organism with rock-like hide. Absorbs kinetic impacts with ease.",
	},
	// Rank 2 - High tier
	{
		Name: "Muton", ShortName: "MUT", Icon: '⊛',
		HP: 25, TU: 55, Accuracy: 55, Bravery: 70, Reactions: 50,
		Strength: 20, Psi: 0, Armour: 18, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 8, DamageType: DMG_PLASMA,
		ResistPlasma: 30, ResistMelee: 20, ResistLaser: -10,
		AutopsyID: "muton_autopsy",
		Lore:      "Brutish green-skinned warriors bred for combat. Strong but not bright.",
	},
	{
		Name: "Muton Navigator", ShortName: "MUN", Icon: '⊜',
		HP: 27, TU: 58, Accuracy: 58, Bravery: 75, Reactions: 53,
		Strength: 21, Psi: 0, Armour: 19, Weapon: "plasma_rifle",
		Points: 16, Rank: 3, Aggression: 8, DamageType: DMG_PLASMA,
		ResistPlasma: 30, ResistMelee: 25, ResistPsionic: -15,
		AutopsyID: "muton_autopsy",
		Lore:      "A muton with tactical awareness. Guides the squad with crude but effective strategies.",
	},
	{
		Name: "Muton Commander", ShortName: "MUC", Icon: '⬣',
		HP: 30, TU: 62, Accuracy: 62, Bravery: 85, Reactions: 58,
		Strength: 23, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 25, Rank: 4, Aggression: 9, DamageType: DMG_PLASMA,
		ResistPlasma: 35, ResistMelee: 30, ResistExplosive: -20,
		AutopsyID: "muton_autopsy",
		Lore:      "The alpha of the muton pack. Its battle rage is legendary among alien forces.",
	},
	{
		Name: "Cyberdisc", ShortName: "CYB", Icon: '◉',
		HP: 30, TU: 50, Accuracy: 65, Bravery: 100, Reactions: 60,
		Strength: 15, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 20, Rank: 2, Aggression: 7, DamageType: DMG_PLASMA,
		ResistPlasma: 25, ResistLaser: 15, ResistExplosive: -25,
		Lore: "Mechanical disc-shaped unit. Fires plasma in all directions while hovering.",
	},
	{
		Name: "Celatid", ShortName: "CEL", Icon: '⌖',
		HP: 12, TU: 60, Accuracy: 60, Bravery: 50, Reactions: 55,
		Strength: 10, Psi: 0, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 6, DamageType: DMG_LASER,
		ResistLaser: 30, ResistPlasma: -15,
		Lore: "Amorphous gelatinous creature that fires concentrated acid bolts.",
	},
	// Rank 3 - Elite
	{
		Name: "Ethereal", ShortName: "ETH", Icon: 'Ψ',
		HP: 18, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 70,
		Strength: 10, Psi: 80, Armour: 12, Weapon: "plasma_rifle",
		Points: 25, Rank: 4, Aggression: 5, DamageType: DMG_PSIONIC,
		ResistPsionic: 50, ResistPlasma: 15, ResistMelee: -20,
		AutopsyID: "ethereal_autopsy",
		Lore:      "The puppet masters. Telepathic beings that control all other alien species.",
	},
	{
		Name: "Ethereal Navigator", ShortName: "ETN", Icon: 'Φ',
		HP: 20, TU: 68, Accuracy: 73, Bravery: 100, Reactions: 73,
		Strength: 11, Psi: 90, Armour: 13, Weapon: "plasma_rifle",
		Points: 30, Rank: 5, Aggression: 4, DamageType: DMG_PSIONIC,
		ResistPsionic: 60, ResistPlasma: 15, ResistMelee: -25,
		AutopsyID: "ethereal_autopsy",
		Lore:      "An ethereal attuned to the psionic frequencies of the command chain.",
	},
	{
		Name: "Ethereal Commander", ShortName: "ETC", Icon: 'Ϙ',
		HP: 24, TU: 72, Accuracy: 78, Bravery: 100, Reactions: 78,
		Strength: 13, Psi: 100, Armour: 16, Weapon: "heavy_plasma",
		Points: 50, Rank: 6, Aggression: 4, DamageType: DMG_PSIONIC,
		ResistPsionic: 70, ResistPlasma: 20, ResistMelee: -30, ResistExplosive: -15,
		AutopsyID: "ethereal_autopsy",
		Lore:      "The supreme psychic intelligence. Its death causes a psionic shockwave across the battlefield.",
	},
	// Rank 4 - Boss
	{
		Name: "Reaper", ShortName: "REAP", Icon: 'ᚱ',
		HP: 50, TU: 45, Accuracy: 50, Bravery: 100, Reactions: 40,
		Strength: 30, Psi: 0, Armour: 25, Weapon: "reaper_claw",
		Points: 30, Rank: 3, Aggression: 9, DamageType: DMG_MELEE,
		ResistMelee: 50, ResistKinetic: 30, ResistLaser: -20, ResistPlasma: -15,
		Lore: "Massive ambush predator. Can swallow a soldier whole.",
	},
	{
		Name: "Sectopod", ShortName: "SEC3", Icon: '⬡',
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

// GetAlienByRank returns the highest-rank alien ≥ minRank.
func GetAlienByRank(minRank int) *AlienType {
	var best *AlienType
	for i := range AlienTypes {
		if AlienTypes[i].Rank >= minRank {
			if best == nil || AlienTypes[i].Rank > best.Rank {
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
		return language.String("DTYPE_PLASMA")
	case DMG_LASER:
		return language.String("DTYPE_LASER")
	case DMG_EXPLOSIVE:
		return language.String("DTYPE_EXPLOSIVE")
	case DMG_MELEE:
		return language.String("DTYPE_MELEE")
	case DMG_KINETIC:
		return language.String("DTYPE_KINETIC")
	case DMG_PSIONIC:
		return language.String("DTYPE_PSIONIC")
	default:
		return language.String("UNKNOWN")
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

func init() {
	for i := range AlienTypes {
		if AlienTypes[i].Style == tcell.StyleDefault {
			AlienTypes[i].Style = tcell.StyleDefault.Foreground(ColorDefaultAlien).Bold(true)
			AlienTypes[i].FgColor = ColorDefaultAlien
		}
	}
}

func pickColor(rng *rand.Rand, colors ...tcell.Color) tcell.Color {
	if len(colors) == 0 {
		return ColorDefaultAlien
	}
	return colors[rng.Intn(len(colors))]
}

// DetermineProceduralIconAndStyle selects the custom Unicode character/rune, tcell.Style,
// and foreground color based on the alien's morphology and body subtype.
// The chosen rune is marked in usedIcons to guarantee uniqueness within a run.
func DetermineProceduralIconAndStyle(m *Morphology, rng *rand.Rand, usedIcons map[rune]bool) (rune, tcell.Style, tcell.Color) {
	if m == nil {
		c := ColorDefaultAlien
		return '?', tcell.StyleDefault.Foreground(c).Bold(true), c
	}

	// 1. Determine base rune based purely on Limbs
	var categoryRunes []rune
	// Categorize based on arms & legs
	if m.Legs >= 4 {
		// Category E. Arachnid / Swarm (Many Legs, 4–8 Legs)
		categoryRunes = []rune{'ᛤ', 'ሐ', 'ሗ', 'Ж', 'ዷ', '✲'}
	} else if m.Legs > 2 {
		if m.Arms >= 4 {
			// Category D. Multi-Armed (4–6 Arms, 2 Legs)
			categoryRunes = []rune{'ቿ', 'ቹ', 'ᚼ'}
		} else if m.Arms == 3 {
			// Category F. 3 Arms, 2 Legs
			categoryRunes = []rune{'ቱ', 'ቲ', 'ቴ', 'պ'}
		} else if m.Arms == 2 {
			// Category C. 2 Arms, 2 Legs (Standard Bipedal)
			categoryRunes = []rune{'ቆ', '☥', 'ቋ', '፹', 'ቐ', 'ቀ', 'ፗ'}
		} else if m.Arms == 1 {
			// Category G. 1 Arm, 2 Legs
			categoryRunes = []rune{'ቲ', 'ኟ', 'ዧ', 'ϝ'}
		} else { // m.Arms == 0
			// Category H. 0 Arms, 2 Legs (Bipedal but no manipulators)
			categoryRunes = []rune{'ሸ', 'Ѿ', 'ኗ', 'ደ', '፬'}
		}
	} else { // m.Legs == 0
		if m.Arms >= 2 {
			// Category B. 2 Arms, 0 Legs (Hovering or Slithering)
			categoryRunes = []rune{'ϯ', 'ቻ', 'ቓ', 'ዎ'}
		} else {
			// Category A. 0 Arms, 0 Legs (Floating, Slithering, or Blob)
			categoryRunes = []rune{'Ѿ', 'Ω', 'Ѻ', 'ዋ'}
		}
	}

	// 2. Biology & Composition (Material / Texture)
	var fgColor tcell.Color
	var style tcell.Style = tcell.StyleDefault.Bold(true)
	var runePool []rune

	switch m.BodySubtype {
	case SubtypeCarbonFlesh:
		// Color: Red, Dark Red, or Pink.
		fgColor = pickColor(rng, tcell.GetColor("red"), tcell.GetColor("darkred"), tcell.GetColor("pink"))
		style = style.Foreground(fgColor)
		runePool = categoryRunes

	case SubtypeSilicon:
		// Color: Dark Grey, Brown, or Dark Orange.
		fgColor = pickColor(rng, tcell.GetColor("darkgray"), tcell.GetColor("brown"), tcell.GetColor("darkorange"))
		style = style.Foreground(fgColor)

		// Override Rune: If it has low limbs (Arms + Legs <= 2), use ⬢ (Solid Hexagon)
		if m.Arms+m.Legs <= 2 {
			runePool = []rune{'⬢'}
		} else {
			runePool = categoryRunes
		}

	case SubtypeGaseous:
		// Color: Magenta, Green, or Purple.
		fgColor = pickColor(rng, tcell.GetColor("magenta"), tcell.GetColor("green"), tcell.GetColor("purple"))
		style = style.Foreground(fgColor)

		// Override Rune: Override the shape with a more ethereal glyph
		runePool = []rune{'◍', '⁂', '⛆', '≋'}

	case SubtypeCrystalline:
		// Color: Bright Cyan, White, or Aqua.
		fgColor = pickColor(rng, tcell.GetColor("cyan"), tcell.GetColor("white"), tcell.GetColor("aqua"))
		style = style.Foreground(fgColor)

		// Override Rune: Overwrite normal stick figures (Category C) with sharp geometry: ♦, ❖, ᛟ
		isStickFigureCategory := m.Legs == 2 && m.Arms < 4
		if isStickFigureCategory {
			runePool = []rune{'♦', '❖', 'ᛟ', '⊻', '⊼', '⊽', 'ᛝ'}
		} else {
			runePool = categoryRunes
		}

	case SubtypeAmorphous:
		// Color: Slime Green or Deep Purple.
		fgColor = pickColor(rng, tcell.GetColor("lime"), tcell.GetColor("darkmagenta"))
		style = style.Foreground(fgColor)

		// Override Rune
		runePool = []rune{'♨', '⚇', '◒', 'ꙮ'}

	case SubtypeMechanical:
		// Color: Silver, Steel Blue, or Yellow.
		fgColor = pickColor(rng, tcell.GetColor("silver"), tcell.GetColor("steelblue"), tcell.GetColor("yellow"))
		style = style.Foreground(fgColor)

		// Override Rune: Use strict, boxy characters regardless of limbs: ⊞
		runePool = []rune{'⊞', '⌺', '⌸'}

	case SubtypeBioSynthetic:
		// Color: Half-Flesh, Half-Neon (e.g., Dark Red with style.Blink(true)).
		fgColor = pickColor(rng, tcell.GetColor("darkred"), tcell.GetColor("lime"), tcell.GetColor("pink"))
		style = style.Foreground(fgColor).Blink(true)

		// Override Rune: Φ or ⍾
		runePool = []rune{'Φ', '⍾', '፸', '⍝'}

	case SubtypeNanotech:
		// Color: Matrix Green or Pitch Black on a Bright White background.
		if rng.Intn(2) == 0 {
			fgColor = tcell.GetColor("lime")
			style = style.Foreground(fgColor)
		} else {
			fgColor = tcell.GetColor("black")
			style = tcell.StyleDefault.Foreground(fgColor).Background(tcell.GetColor("white")).Bold(true)
		}

		// Override Rune
		runePool = []rune{'፨', '⠪', '✜', '⛬', '⡳'}

	default:
		fgColor = ColorDefaultAlien
		style = style.Foreground(fgColor)
		runePool = categoryRunes
	}

	// Filter runePool to find unused ones. When the small morphology pool is
	// exhausted (common for species with many rank-variants sharing one morphology),
	// fall back to nextIcon which iterates the full damage-type pools and guarantees
	// uniqueness across the entire run.  The biology-derived style is preserved.
	var unused []rune
	for _, r := range runePool {
		if !usedIcons[r] {
			unused = append(unused, r)
		}
	}

	var r rune
	if len(unused) > 0 {
		r = unused[rng.Intn(len(unused))]
	} else {
		// All morphology-pool runes are taken; pick any still-unused glyph from the
		// broader damage-type pools so the icon stays unique.
		r = nextIcon(-1, usedIcons) // -1 → skip affinity pool, scan all pools
	}

	// Mark as used so subsequent calls within the same species don't re-pick it.
	usedIcons[r] = true
	return r, style, fgColor
}
