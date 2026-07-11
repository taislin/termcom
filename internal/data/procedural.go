package data

import (
	"fmt"
	"math/rand"
	"strings"
)

// AlienSpecies represents a procedurally generated alien species.
// Each species has variants at different ranks.
type AlienSpecies struct {
	Name       string       // e.g. "Vrekt"
	Prefix     string       // short prefix for rank variants, e.g. "VRK"
	BaseIcon   rune         // shared icon across all variants
	PrimaryDMG int          // species-wide damage affinity
	Lore       string       // species-wide lore
	Morphology *Morphology  // shared morphology across all variants
	Types      []*AlienType // Rank 0..4 variants (may be nil for higher ranks)
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

// Morphology lore snippets describing body subtypes.
var morphLoreSnippets = map[string]string{
	SubtypeCarbonFlesh:  "flesh and bone",
	SubtypeSilicon:      "silicon-based crystalline tissue",
	SubtypeGaseous:      "a semi-corporeal gaseous form",
	SubtypeCrystalline:  "dense mineral crystalline structure",
	SubtypeAmorphous:    "an ever-shifting amorphous body",
	SubtypeMechanical:   "precision-forged mechanical components",
	SubtypeBioSynthetic: "a fusion of organic tissue and synthetic armor",
	SubtypeNanotech:     "a swarm of nanoscale machines",
}

// Limb lore descriptions.
func limbLore(m *Morphology) string {
	armDesc := "no arms"
	switch {
	case m.Arms == 1:
		armDesc = "a single manipulative tentacle"
	case m.Arms == 2:
		armDesc = "a pair of arms"
	case m.Arms <= 4:
		armDesc = fmt.Sprintf("%d grasping limbs", m.Arms)
	default:
		armDesc = fmt.Sprintf("a mass of %d limbs", m.Arms)
	}
	legDesc := "it hovers above the ground"
	switch {
	case m.Legs == 1:
		legDesc = "it slithers on a single muscular foot"
	case m.Legs == 2:
		legDesc = "it walks upright"
	case m.Legs == 4:
		legDesc = "it moves on four legs"
	case m.Legs >= 6:
		legDesc = fmt.Sprintf("it scurries on %d legs", m.Legs)
	}
	return fmt.Sprintf("With %s, %s", armDesc, legDesc)
}

// GenerateSpecies creates a full set of procedural alien species from a seed.
// Returns the species list and a combined AlienTypes slice for use in battles.
func GenerateSpecies(seed int64) ([]*AlienSpecies, []*AlienType) {
	rng := rand.New(rand.NewSource(seed))

	speciesCount := 5 + rng.Intn(3) // 5-7 species per run
	allSpecies := make([]*AlienSpecies, 0, speciesCount)
	allTypes := make([]*AlienType, 0, speciesCount*4)

	usedNames := make(map[string]bool)
	usedIcons := UsedHardcodedIcons()

	for i := 0; i < speciesCount; i++ {
		sp := generateOneSpecies(rng, i, usedNames, usedIcons)
		usedNames[sp.Name] = true
		allSpecies = append(allSpecies, sp)
		allTypes = append(allTypes, sp.Types...)
	}

	return allSpecies, allTypes
}

func generateOneSpecies(rng *rand.Rand, idx int, usedNames map[string]bool, usedIcons map[rune]bool) *AlienSpecies {
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

	// Choose icon: draw a distinct glyph from the pool for this species' damage type.
	icon := nextIcon(primaryDMG, usedIcons)

	// Generate morphology
	morph := generateMorphology(rng, primaryDMG)

	sp := &AlienSpecies{
		Name:       name,
		Prefix:     prefix,
		BaseIcon:   icon,
		PrimaryDMG: primaryDMG,
		Morphology: morph,
	}

	// Generate 2-4 rank variants (not all species have all ranks)
	maxRank := 1 + rng.Intn(4) // 1-4 variants
	sp.Types = make([]*AlienType, 0, maxRank)

	for rank := 0; rank < maxRank; rank++ {
		at := generateVariant(rng, sp, rank, usedIcons)
		sp.Types = append(sp.Types, at)
	}

	// Generate species lore
	sp.Lore = generateLore(name, primaryDMG, morph)

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

// generateMorphology creates the physical form of an alien species.
// Body subtype is influenced by damage type for thematic consistency.
func generateMorphology(rng *rand.Rand, dmgType int) *Morphology {
	bodyType := BodyOrganic
	if rng.Intn(3) == 0 { // 33% synthetic
		bodyType = BodySynthetic
	}

	// Pick body subtype influenced by damage type
	var subtype string
	if bodyType == BodyOrganic {
		subtype = pickOrganicSubtype(rng, dmgType)
	} else {
		subtype = pickSyntheticSubtype(rng, dmgType)
	}

	// Generate limb counts — constrained by body subtype
	arms := generateArmCount(rng, subtype)
	legs := generateLegCount(rng, subtype)

	// Generate senses
	eyesight := pickSenseQuality(rng)
	hearing := pickHearingQuality(rng)
	thermal := pickBinarySense(rng)
	psiSense := pickBinarySense(rng)
	chemSense := pickBinarySense(rng)

	// Enforce sense restrictions
	if subtype == SubtypeMechanical || subtype == SubtypeNanotech {
		chemSense = SenseNone
	}
	if subtype == SubtypeGaseous || subtype == SubtypeAmorphous {
		// Amorphous/gaseous: compensate with psi and thermal
		if psiSense == SenseNone {
			psiSense = SenseLow
		}
		thermal = SenseHigh
	}
	if subtype == SubtypeCrystalline {
		// Crystalline: good eyesight (light refraction), poor hearing
		hearing = SensePoor
		if eyesight == SenseNormal || eyesight == SensePoor {
			eyesight = SenseExcellent
		}
	}
	if subtype == SubtypeSilicon {
		// Silicon-based: thermal vision
		thermal = SenseHigh
	}
	if dmgType == DMG_PSIONIC && psiSense == SenseNone {
		psiSense = SenseHigh
	}

	return &Morphology{
		Arms:          arms,
		Legs:          legs,
		BodyType:      bodyType,
		BodySubtype:   subtype,
		Eyesight:      eyesight,
		Hearing:       hearing,
		ThermalSense:  thermal,
		PsionicSense:  psiSense,
		ChemicalSense: chemSense,
	}
}

func pickOrganicSubtype(rng *rand.Rand, dmgType int) string {
	// Damage type biases
	switch dmgType {
	case DMG_MELEE:
		// Melee aliens more likely amorphous or carbon_flesh
		if rng.Intn(3) == 0 {
			return SubtypeAmorphous
		}
	case DMG_PSIONIC:
		// Psionic aliens more likely amorphous or gaseous
		roll := rng.Intn(4)
		if roll == 0 {
			return SubtypeAmorphous
		}
		if roll == 1 {
			return SubtypeGaseous
		}
	case DMG_KINETIC:
		// Kinetic aliens more likely crystalline or silicon
		if rng.Intn(3) == 0 {
			return SubtypeCrystalline
		}
	case DMG_LASER:
		// Laser aliens more likely silicon (light interaction)
		if rng.Intn(3) == 0 {
			return SubtypeSilicon
		}
	}
	// Default weighted roll
	roll := rng.Intn(10)
	switch {
	case roll < 4:
		return SubtypeCarbonFlesh
	case roll < 6:
		return SubtypeSilicon
	case roll < 7:
		return SubtypeGaseous
	case roll < 9:
		return SubtypeCrystalline
	default:
		return SubtypeAmorphous
	}
}

func pickSyntheticSubtype(rng *rand.Rand, dmgType int) string {
	switch dmgType {
	case DMG_MELEE:
		if rng.Intn(3) == 0 {
			return SubtypeBioSynthetic
		}
	case DMG_PSIONIC:
		// Psionic + synthetic = bio_synthetic (organic core)
		return SubtypeBioSynthetic
	case DMG_KINETIC:
		if rng.Intn(3) == 0 {
			return SubtypeNanotech
		}
	}
	roll := rng.Intn(10)
	switch {
	case roll < 5:
		return SubtypeMechanical
	case roll < 8:
		return SubtypeBioSynthetic
	default:
		return SubtypeNanotech
	}
}

func generateArmCount(rng *rand.Rand, subtype string) int {
	switch subtype {
	case SubtypeGaseous, SubtypeAmorphous:
		return 0 // amorphous/gaseous have no arms
	case SubtypeCrystalline:
		return rng.Intn(3) // 0-2 (rigid)
	case SubtypeNanotech:
		return 0 // no fixed limbs
	case SubtypeMechanical:
		return 2 + rng.Intn(3) // 2-4
	case SubtypeSilicon:
		return 2 + rng.Intn(2) // 2-3
	default:
		roll := rng.Intn(10)
		switch {
		case roll < 2:
			return 0
		case roll < 3:
			return 1
		case roll < 7:
			return 2
		case roll < 9:
			return 3 + rng.Intn(2) // 3-4
		default:
			return 5 + rng.Intn(2) // 5-6
		}
	}
}

func generateLegCount(rng *rand.Rand, subtype string) int {
	switch subtype {
	case SubtypeGaseous, SubtypeAmorphous, SubtypeNanotech:
		return 0 // floating
	case SubtypeCrystalline:
		return 2 // rigid biped
	case SubtypeMechanical:
		roll := rng.Intn(6)
		switch {
		case roll < 2:
			return 0 // hovering
		case roll < 5:
			return 2
		default:
			return 4
		}
	case SubtypeSilicon:
		return 2 + rng.Intn(3)*2 // 2 or 4
	default:
		roll := rng.Intn(10)
		switch {
		case roll < 2:
			return 0
		case roll < 6:
			return 2
		case roll < 9:
			return 4
		default:
			return 6 + rng.Intn(3)*2 // 6 or 8
		}
	}
}

func pickSenseQuality(rng *rand.Rand) string {
	roll := rng.Intn(10)
	switch {
	case roll < 1:
		return SenseNone
	case roll < 3:
		return SensePoor
	case roll < 7:
		return SenseNormal
	case roll < 9:
		return SenseExcellent
	default:
		return SenseMultiSpec
	}
}

func pickHearingQuality(rng *rand.Rand) string {
	roll := rng.Intn(10)
	switch {
	case roll < 1:
		return SenseNone
	case roll < 3:
		return SensePoor
	case roll < 7:
		return SenseNormal
	case roll < 9:
		return SenseExcellent
	default:
		return SenseEcholoc
	}
}

func pickBinarySense(rng *rand.Rand) string {
	roll := rng.Intn(10)
	switch {
	case roll < 3:
		return SenseNone
	case roll < 7:
		return SenseLow
	default:
		return SenseHigh
	}
}

// --- Morphology stat modifiers ---

func morphHPMod(m *Morphology) int {
	mod := 0
	if m.IsFloating() {
		mod -= 3
	}
	if m.BodySubtype == SubtypeCrystalline {
		mod += 5
	}
	if m.BodySubtype == SubtypeAmorphous {
		mod -= 2
	}
	if m.BodySubtype == SubtypeNanotech {
		mod -= 4
	}
	return mod
}

func morphTUMod(m *Morphology) int {
	mod := 0
	if m.IsFloating() {
		mod -= 5
	}
	switch m.Legs {
	case 4:
		mod += 10
	case 6, 8:
		mod += 15
	}
	if m.BodySubtype == SubtypeCrystalline {
		mod -= 10
	}
	if m.BodySubtype == SubtypeMechanical {
		mod += 5
	}
	return mod
}

func morphAccMod(m *Morphology) int {
	mod := 0
	switch m.Arms {
	case 0:
		mod -= 15
	case 1:
		mod -= 5
	case 3, 4:
		mod += 5
	case 5, 6:
		mod += 10
	}
	if m.IsFloating() {
		mod -= 5
	}
	if m.Eyesight == SenseNone {
		mod -= 20
	} else if m.Eyesight == SenseExcellent || m.Eyesight == SenseMultiSpec {
		mod += 10
	}
	if m.BodySubtype == SubtypeMechanical {
		mod += 5
	}
	return mod
}

func morphReactMod(m *Morphology) int {
	mod := 0
	if m.IsFloating() {
		mod += 10
	}
	switch m.Legs {
	case 4:
		mod += 5
	case 6, 8:
		mod += 10
	}
	if m.Hearing == SenseEcholoc || m.Hearing == SenseExcellent {
		mod += 5
	}
	if m.ThermalSense == SenseHigh {
		mod += 5
	}
	return mod
}

func morphStrMod(m *Morphology) int {
	mod := 0
	switch m.Arms {
	case 0:
		mod -= 10
	case 3, 4:
		mod += 5
	case 5, 6:
		mod += 10
	}
	switch m.Legs {
	case 4:
		mod += 5
	case 6, 8:
		mod += 5
	}
	if m.BodySubtype == SubtypeCrystalline {
		mod += 5
	}
	return mod
}

func morphPsiMod(m *Morphology) int {
	mod := 0
	if m.Eyesight == SenseNone {
		mod += 10 // blind aliens compensate with psi
	}
	if m.PsionicSense == SenseHigh {
		mod += 15
	} else if m.PsionicSense == SenseLow {
		mod += 5
	}
	if m.BodySubtype == SubtypeAmorphous || m.BodySubtype == SubtypeGaseous {
		mod += 10
	}
	return mod
}

func morphArmMod(m *Morphology) int {
	mod := 0
	if m.BodySubtype == SubtypeCrystalline {
		mod += 5
	}
	if m.BodySubtype == SubtypeMechanical {
		mod += 3
	}
	if m.BodySubtype == SubtypeNanotech {
		mod -= 2
	}
	return mod
}

func morphAggroMod(m *Morphology) int {
	mod := 0
	if m.BodySubtype == SubtypeAmorphous || m.BodySubtype == SubtypeGaseous {
		mod -= 2
	}
	if m.BodySubtype == SubtypeMechanical {
		mod += 1
	}
	return mod
}

// subtypeResistMod returns bonus resistance from body subtype for a damage type.
func subtypeResistMod(subtype string, dmgType int) int {
	switch subtype {
	case SubtypeCarbonFlesh:
		if dmgType == DMG_KINETIC {
			return 10
		}
		if dmgType == DMG_EXPLOSIVE {
			return -10
		}
	case SubtypeSilicon:
		if dmgType == DMG_LASER {
			return 20
		}
		if dmgType == DMG_PLASMA {
			return 10
		}
		if dmgType == DMG_EXPLOSIVE {
			return -15
		}
	case SubtypeGaseous:
		if dmgType == DMG_KINETIC {
			return 80 // nearly immune
		}
		if dmgType == DMG_PLASMA {
			return -20
		}
	case SubtypeCrystalline:
		if dmgType == DMG_EXPLOSIVE {
			return -25
		}
		return 15
	case SubtypeAmorphous:
		if dmgType == DMG_PSIONIC {
			return 10
		}
	case SubtypeMechanical:
		if dmgType == DMG_PSIONIC {
			return 80 // immune
		}
		if dmgType == DMG_PLASMA {
			return 15
		}
		if dmgType == DMG_LASER {
			return -15
		}
	case SubtypeBioSynthetic:
		return 5 // small bonus to everything
	case SubtypeNanotech:
		if dmgType == DMG_KINETIC {
			return 20
		}
		return -10
	}
	return 0
}

func generateVariant(rng *rand.Rand, sp *AlienSpecies, rank int, usedIcons map[rune]bool) *AlienType {
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

	// Apply morphology modifiers
	m := sp.Morphology
	hpBase += morphHPMod(m)
	tuBase += morphTUMod(m)
	accBase += morphAccMod(m)
	reactBase += morphReactMod(m)
	strBase += morphStrMod(m)
	psiBase += morphPsiMod(m)
	armBase += morphArmMod(m)
	aggroBase += morphAggroMod(m)

	// Clamp values
	hpBase = max(hpBase, 4)
	tuBase = max(tuBase, 20)
	accBase = clamp(accBase, 20, 100)
	braveBase = clamp(braveBase, 10, 100)
	reactBase = clamp(reactBase, 15, 100)
	strBase = max(strBase, 2)
	psiBase = clamp(psiBase, 0, 100)
	armBase = max(armBase, 0)
	aggroBase = clamp(aggroBase, 1, 10)

	// Choose weapon based on rank
	weaps := alienWeaponsByRank[rank]
	weapon := weaps[rng.Intn(len(weaps))]

	// Generate base resistances
	resistPlasma := genResist(rng, sp.PrimaryDMG, DMG_PLASMA, rank)
	resistLaser := genResist(rng, sp.PrimaryDMG, DMG_LASER, rank)
	resistExplosive := genResist(rng, sp.PrimaryDMG, DMG_EXPLOSIVE, rank)
	resistMelee := genResist(rng, sp.PrimaryDMG, DMG_MELEE, rank)
	resistKinetic := genResist(rng, sp.PrimaryDMG, DMG_KINETIC, rank)
	resistPsionic := genResist(rng, sp.PrimaryDMG, DMG_PSIONIC, rank)

	// Apply body subtype resistance modifiers
	resistPlasma += subtypeResistMod(m.BodySubtype, DMG_PLASMA)
	resistLaser += subtypeResistMod(m.BodySubtype, DMG_LASER)
	resistExplosive += subtypeResistMod(m.BodySubtype, DMG_EXPLOSIVE)
	resistMelee += subtypeResistMod(m.BodySubtype, DMG_MELEE)
	resistKinetic += subtypeResistMod(m.BodySubtype, DMG_KINETIC)
	resistPsionic += subtypeResistMod(m.BodySubtype, DMG_PSIONIC)

	// Clamp resistances
	resistPlasma = clamp(resistPlasma, -50, 80)
	resistLaser = clamp(resistLaser, -50, 80)
	resistExplosive = clamp(resistExplosive, -50, 80)
	resistMelee = clamp(resistMelee, -50, 80)
	resistKinetic = clamp(resistKinetic, -50, 80)
	resistPsionic = clamp(resistPsionic, -50, 80)

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

	// Icon: a distinct glyph from this species' damage-type pool.
	icon := nextIcon(sp.PrimaryDMG, usedIcons)

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

		Lore:       variantLore,
		Portrait:   generatePortrait(rng, icon, sp.PrimaryDMG, rank, m),
		Morphology: m,
	}
}

// portraitPart holds optional decorative characters that vary per species.
type portraitPart struct {
	crown  rune // head ornament
	chest  rune // chest marking
	weapon rune // held weapon
	sensor rune // sense antenna/sensor decoration
}

func generatePortrait(rng *rand.Rand, icon rune, dmgType int, rank int, m *Morphology) StyledPortrait {
	parts := portraitPart{
		crown:  pickRune(rng, []rune{' ', '°', '*', '+', '÷', '¤', '~', '^'}),
		chest:  pickRune(rng, []rune{' ', '·', ':', 'o', '×', '†', '◊', '≈'}),
		weapon: pickRune(rng, []rune{'/', '\\', '|', '†', '¶', '©', '£', '¥'}),
		sensor: pickSenseSensor(rng, m),
	}
	// Generate a unique palette for this species variant
	palette := [3]int32{
		int32(rng.Intn(150) + 100), // R
		int32(rng.Intn(150) + 100), // G
		int32(rng.Intn(150) + 100), // B
	}
	return assemblePortrait(parts, m, rank, palette)
}

func pickSenseSensor(rng *rand.Rand, m *Morphology) rune {
	if m.PsionicSense == SenseHigh {
		return pickRune(rng, []rune{'Ω', 'Ψ', 'Φ'})
	}
	if m.ThermalSense == SenseHigh {
		return pickRune(rng, []rune{'~', '≋', '^'})
	}
	if m.Hearing == SenseEcholoc {
		return pickRune(rng, []rune{')', '(', '»'})
	}
	if m.ChemicalSense == SenseHigh {
		return pickRune(rng, []rune{'⌁', '¤', '*'})
	}
	return ' '
}

// headShape returns the 2-line head silhouette for a body subtype.
func headShape(subtype string, chest rune) (string, string) {
	switch subtype {
	case SubtypeCarbonFlesh:
		return " (o.o) ", fmt.Sprintf("  |%c|  ", chest)
	case SubtypeSilicon:
		return " <◊◊◊> ", fmt.Sprintf("  |%c|  ", chest)
	case SubtypeGaseous:
		return " ~~~~  ", fmt.Sprintf("  ~%c~  ", chest)
	case SubtypeCrystalline:
		return " /△△\\ ", fmt.Sprintf("  |%c|  ", chest)
	case SubtypeAmorphous:
		return " (~~~) ", fmt.Sprintf("  ~%c~  ", chest)
	case SubtypeMechanical:
		return " [=■=] ", fmt.Sprintf("  |%c|  ", chest)
	case SubtypeBioSynthetic:
		return " (o-o) ", fmt.Sprintf("  |%c|  ", chest)
	case SubtypeNanotech:
		return " ·:.·  ", fmt.Sprintf("  ·%c·  ", chest)
	default:
		return " (o.o) ", fmt.Sprintf("  |%c|  ", chest)
	}
}

// armRow returns the arm visualization row.
func armRow(arms int, chest rune) string {
	switch {
	case arms == 0:
		return fmt.Sprintf("  |%c|  ", chest)
	case arms == 1:
		return fmt.Sprintf(" /-+%c\\ ", chest)
	case arms == 2:
		return fmt.Sprintf("/-%c-+-%c\\", chest, chest)
	case arms <= 4:
		return fmt.Sprintf("†-%c-+-%c†", chest, chest)
	default:
		return fmt.Sprintf("※-%c+-%c※", chest, chest)
	}
}

// legRows returns the 2 bottom rows for leg visualization.
func legRows(legs int) (string, string) {
	switch legs {
	case 0:
		return " ~~~  ", "  ~~~  "
	case 2:
		return " \\_/  ", " |_|_| "
	case 4:
		return " \\_/_/ ", "|_|_|_|"
	case 6:
		return "\\_/_/_/", "|_|_|_|"
	case 8:
		return "\\_/_/_/_/", "|_|_|_|_|"
	default:
		return " \\_/  ", " |_|_| "
	}
}

// sensorRow returns the top decoration row for senses (shown above head).
func sensorRow(p portraitPart, m *Morphology) string {
	s := p.sensor
	if s == ' ' {
		return ""
	}
	return fmt.Sprintf("  %c%c%c  ", s, s, s)
}

// assemblePortrait builds a 7-column by 6-row alien portrait driven by morphology.
// Row layout: [sensor] [head1] [head2] [neck] [torso+arms] [legs1] [legs2]
func assemblePortrait(p portraitPart, m *Morphology, rank int, palette [3]int32) StyledPortrait {
	chest := p.chest

	head1, head2 := headShape(m.BodySubtype, chest)
	neck := "   |   "
	torso := armRow(m.Arms, chest)
	leg1, leg2 := legRows(m.Legs)

	lines := []string{head1, head2, neck, torso, leg1, leg2}

	// Sensor decoration row (replaces head1 if senses warrant it)
	sensorLine := sensorRow(p, m)
	if sensorLine != "" {
		lines[0] = sensorLine
	}

	// Rank crown: rank >= 2 replaces line 0 with crown
	if rank >= 2 {
		crown := p.crown
		lines[0] = fmt.Sprintf("  %c%c%c  ", crown, crown, crown)
	}

	// Pad every line to exactly 7 runes
	for i, l := range lines {
		r := []rune(l)
		for len(r) < 7 {
			r = append(r, ' ')
		}
		lines[i] = string(r[:7])
	}

	// Color zones: head, torso, legs
	headColor := [3]int32{
		int32(clamp(int(palette[0])+40, 0, 255)),
		int32(clamp(int(palette[1])+40, 0, 255)),
		int32(clamp(int(palette[2])+40, 0, 255)),
	}
	torsoColor := palette
	legColor := [3]int32{
		int32(clamp(int(palette[0])-30, 0, 255)),
		int32(clamp(int(palette[1])-30, 0, 255)),
		int32(clamp(int(palette[2])-30, 0, 255)),
	}

	// 6 rows: 0-1 = head, 2-3 = torso, 4-5 = legs
	sections := []int{0, 0, 1, 1, 2, 2}
	styledLines := make([]StyledLine, len(lines))
	for i, l := range lines {
		var color [3]int32
		switch sections[i] {
		case 0:
			color = headColor
		case 1:
			color = torsoColor
		case 2:
			color = legColor
		}
		styledLines[i] = StyledLine{
			Content: l,
			Color:   color,
		}
	}

	return StyledPortrait{Lines: styledLines}
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

func generateLore(name string, dmgType int, m *Morphology) string {
	template := loreTemplates[dmgType%len(loreTemplates)]
	base := strings.Replace(template, "%s", DamageTypeStr(dmgType), 1)

	bodyDesc := morphLoreSnippets[m.BodySubtype]
	if bodyDesc == "" {
		bodyDesc = "an unknown biology"
	}

	limbDesc := limbLore(m)

	return fmt.Sprintf("%s %s. Its body is composed of %s.", limbDesc, base, bodyDesc)
}

func pickRune(rng *rand.Rand, pool []rune) rune {
	return pool[rng.Intn(len(pool))]
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
