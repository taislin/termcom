package data

import (
	"math/rand"
	"strings"
)

// AlienSpecies represents a procedurally generated alien species.
// Each species has variants at different ranks.
type AlienSpecies struct {
	Name      string  // e.g. "Vrekt"
	Prefix    string  // short prefix for rank variants, e.g. "VRK"
	BaseIcon  rune    // shared icon across all variants
	PrimaryDMG int    // species-wide damage affinity
	Lore      string  // species-wide lore
	Types     []*AlienType // Rank 0..4 variants (may be nil for higher ranks)
}

// Syllable pools for alien name generation.
var (
	prefixSyll = []string{"Vr", "Za", "Xo", "Kr", "Th", "Qu", "Sh", "Bl", "Dr", "Gh",
		"Ny", "Py", "Wr", "Sk", "Tr", "Ch", "Ph", "Fl", "St", "Sn"}
	midSyll = []string{"ek", "or", "an", "ul", "ix", "az", "en", "on", "ar", "al",
		"is", "us", "ox", "ir", "um", "ak", "el", "id", "os", "ym"}
	endSyll = []string{"id", "on", "ar", "ex", "us", "ith", "ax", "or", "en", "al",
		"ix", "um", "ak", "oi", "esh", "urr", "oth", "agh", "unn", "izz"}
)

// Rank titles used as suffixes for variant names.
var rankTitles = []string{"", "Navigator", "Commander", "Elite", "Overlord"}

// Weapons available to procedural aliens, ordered by rank suitability.
var alienWeaponsByRank = [][]string{
	{"plasma_pistol"},
	{"plasma_pistol", "plasma_rifle"},
	{"plasma_rifle"},
	{"plasma_rifle", "heavy_plasma"},
	{"heavy_plasma"},
}

// Lore templates filled with species name and traits.
var loreTemplates = []string{
	"A hive species from the outer void. Their %s affinity makes them dangerous at any range.",
	"Bioengineered warriors with natural %s resistance. They adapt quickly to new threats.",
	"Silent hunters who strike from the shadows. Their %s attacks leave no survivors.",
	"A ancient race augmented by unknown technology. %s runs through their veins.",
	"Parasitic organisms that absorb the traits of their prey. %s is their signature.",
}

// GenerateSpecies creates a full set of procedural alien species from a seed.
// Returns the species list and a combined AlienTypes slice for use in battles.
func GenerateSpecies(seed int64) ([]*AlienSpecies, []*AlienType) {
	rng := rand.New(rand.NewSource(seed))

	speciesCount := 5 + rng.Intn(3) // 5-7 species per run
	allSpecies := make([]*AlienSpecies, 0, speciesCount)
	allTypes := make([]*AlienType, 0, speciesCount*4)

	usedNames := make(map[string]bool)

	for i := 0; i < speciesCount; i++ {
		sp := generateOneSpecies(rng, i, usedNames)
		usedNames[sp.Name] = true
		allSpecies = append(allSpecies, sp)
		allTypes = append(allTypes, sp.Types...)
	}

	return allSpecies, allTypes
}

func generateOneSpecies(rng *rand.Rand, idx int, usedNames map[string]bool) *AlienSpecies {
	// Generate name
	var name string
	for {
		name = generateName(rng)
		if !usedNames[name] {
			break
		}
	}

	prefix := strings.ToUpper(name[:min(3, len(name))])

	primaryDMG := rng.Intn(6) // DMG_PLASMA..DMG_PSIONIC

	// Choose icon: use a unique rune per species from available symbols.
	// Species icons are distinct from hardcoded aliens (rank 0x100+ block).
	icon := rune(0x100 + idx*7)

	sp := &AlienSpecies{
		Name:       name,
		Prefix:     prefix,
		BaseIcon:   icon,
		PrimaryDMG: primaryDMG,
	}

	// Generate 2-4 rank variants (not all species have all ranks)
	maxRank := 1 + rng.Intn(4) // 2-5 variants
	sp.Types = make([]*AlienType, 0, maxRank)

	for rank := 0; rank < maxRank; rank++ {
		at := generateVariant(rng, sp, rank)
		sp.Types = append(sp.Types, at)
	}

	// Generate species lore
	sp.Lore = generateLore(name, primaryDMG)

	return sp
}

func generateName(rng *rand.Rand) string {
	p := prefixSyll[midSyllIdx(rng, len(prefixSyll))]
	m := midSyll[rng.Intn(len(midSyll))]
	e := endSyll[rng.Intn(len(endSyll))]
	return p + m + e
}

func midSyllIdx(rng *rand.Rand, max int) int {
	return rng.Intn(max)
}

func generateVariant(rng *rand.Rand, sp *AlienSpecies, rank int) *AlienType {
	// Base stats scale with rank
	hpBase := 8 + rank*5 + rng.Intn(4)
	tuBase := 45 + rank*5 + rng.Intn(6)
	accBase := 50 + rank*5 + rng.Intn(8)
	braveBase := 35 + rank*10 + rng.Intn(10)
	reactBase := 45 + rank*6 + rng.Intn(8)
	strBase := 6 + rank*4 + rng.Intn(5)
	psiBase := rng.Intn(30 + rank*15)
	armBase := 3 + rank*4 + rng.Intn(4)
	ptsBase := 4 + rank*6 + rng.Intn(5)
	aggroBase := 3 + rank + rng.Intn(3)

	// Clamp values
	psiBase = clamp(psiBase, 0, 100)
	aggroBase = clamp(aggroBase, 1, 10)

	// Choose weapon based on rank
	weaps := alienWeaponsByRank[rank]
	weapon := weaps[rng.Intn(len(weaps))]

	// Generate resistances: species gets a spread, with PrimaryDMG as resistance
	resistPlasma := genResist(rng, sp.PrimaryDMG, DMG_PLASMA, rank)
	resistLaser := genResist(rng, sp.PrimaryDMG, DMG_LASER, rank)
	resistExplosive := genResist(rng, sp.PrimaryDMG, DMG_EXPLOSIVE, rank)
	resistMelee := genResist(rng, sp.PrimaryDMG, DMG_MELEE, rank)
	resistKinetic := genResist(rng, sp.PrimaryDMG, DMG_KINETIC, rank)
	resistPsionic := genResist(rng, sp.PrimaryDMG, DMG_PSIONIC, rank)

	// Build variant name
	varName := sp.Name
	if rank > 0 && rank <= len(rankTitles) {
		varName = sp.Name + " " + rankTitles[rank-1]
	}

	// Build short name
	shortName := sp.Prefix
	if rank > 0 {
		shortName += string(rune('A' + rank - 1))
	}

	// Icon: base icon + rank offset
	icon := sp.BaseIcon + rune(rank)

	// Lore per variant
	variantLore := sp.Lore
	if rank > 0 {
		variantLore = rankTitles[rank-1] + " of the " + sp.Name + " species. " + variantLore
	}

	return &AlienType{
		Name:       varName,
		ShortName:  shortName,
		Icon:       icon,
		HP:         hpBase,
		TU:         tuBase,
		Accuracy:   accBase,
		Bravery:    braveBase,
		Reactions:  reactBase,
		Strength:   strBase,
		Psi:        psiBase,
		Armour:     armBase,
		Weapon:     weapon,
		Points:     ptsBase,
		Rank:       rank,
		Aggression: aggroBase,
		DamageType: sp.PrimaryDMG,

		ResistPlasma:    resistPlasma,
		ResistLaser:     resistLaser,
		ResistExplosive: resistExplosive,
		ResistMelee:     resistMelee,
		ResistKinetic:   resistKinetic,
		ResistPsionic:   resistPsionic,

		Lore: variantLore,
	}
}

// genResist generates a resistance value for a specific damage type.
// If dmgType matches the species affinity, guaranteed resistance.
// Otherwise random chance of resistance or weakness.
func genResist(rng *rand.Rand, affinity int, dmgType int, rank int) int {
	if dmgType == affinity {
		// Species affinity: guaranteed resistance, scales with rank
		return 15 + rank*8 + rng.Intn(10)
	}
	// Random: 40% chance resistance, 30% chance weakness, 30% neutral
	roll := rng.Intn(100)
	if roll < 40 {
		return 5 + rng.Intn(10+rank*3)
	} else if roll < 70 {
		return -(10 + rng.Intn(10+rank*2))
	}
	return 0
}

func generateLore(name string, dmgType int) string {
	template := loreTemplates[dmgType%len(loreTemplates)]
	return strings.Replace(template, "%s", DamageTypeStr(dmgType), 1)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
