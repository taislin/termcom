package data

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
)

// AlienSpecies represents a procedurally generated alien species.
// Each species has variants at different ranks, sharing a base morphology and damage affinity.
type AlienSpecies struct {
	Name       string       // The generated species name (e.g. "Vrekt")
	Prefix     string       // Short prefix for rank variants (e.g. "VRK")
	BaseIcon   rune         // shared icon across all variants
	BaseStyle  tcell.Style  // shared style across all variants
	PrimaryDMG int          // species-wide damage affinity
	Lore       string       // species-wide lore
	Morphology *Morphology  // shared physical form across all variants
	Types      []*AlienType // Rank 0..4 variants (may be nil for higher ranks)
}

// Syllable pools for generating exotic-sounding alien names.
// Latin-based (used for en, es, fr, pt)
var (
	prefixSyll = []string{
		"Vr", "Za", "Xo", "Kr", "Th", "Qu", "Sh", "Bl", "Dr", "Gh",
		"Ny", "Py", "Wr", "Sk", "Tr", "Ch", "Ph", "Fl", "St", "Sn",
		"Br", "Gr", "Cr", "Pr", "Fr", "Cl", "Sp", "Sl", "Sm", "Sw",
		"Zr", "Xk", "Ql", "Jx", "Vh", "Ng", "Mk", "Nz", "Rq", "Wx",
	}
	midSyll = []string{
		"ek", "or", "an", "ul", "ix", "az", "en", "on", "ar", "al",
		"is", "us", "ox", "ir", "um", "ak", "el", "id", "os", "ym",
		"iv", "un", "ag", "eb", "og", "iz", "ub", "am", "ol", "ed",
		"ik", "av", "od", "ul", "er", "at", "op", "ev", "im", "yx",
	}
	endSyll = []string{
		"id", "on", "ar", "ex", "us", "ith", "ax", "or", "en", "al",
		"ix", "um", "ak", "oi", "esh", "urr", "oth", "agh", "unn", "izz",
		"eon", "aux", "yne", "oph", "ule", "urg", "ash", "isk", "orn", "uum",
		"exx", "ylt", "oph", "uun", "eir", "aq", "iim", "oxx", "uum", "aaz",
	}
)

// Cyrillic syllable pools for Russian
var (
	ruPrefixSyll = []string{
		"Бра", "Кро", "Хра", "Гро", "Вра", "Ска", "Тра", "Зра", "Дра", "Шра",
		"Пра", "Фра", "Ква", "Кша", "Рта", "Жра", "Цра", "Чра", "Щра", "Хва",
		"Гла", "Бло", "Вло", "Кла", "Пла", "Хла", "Сна", "Бна", "Кна", "Тна",
	}
	ruMidSyll = []string{
		"ек", "ор", "ан", "ул", "ир", "аз", "ен", "он", "ар", "ал",
		"ис", "ус", "окс", "ир", "ум", "ак", "ел", "ид", "ос", "ыл",
		"ив", "ун", "аг", "ев", "ог", "из", "уб", "ам", "ол", "ед",
		"ик", "ав", "од", "ут", "ер", "ат", "оп", "ем", "им", "ых",
	}
	ruEndSyll = []string{
		"ид", "он", "ар", "екс", "ус", "иф", "акс", "ор", "ен", "ал",
		"икс", "ум", "ок", "ой", "еш", "урр", "оф", "аг", "ун", "из",
		"еон", "аус", "ин", "оф", "ул", "ург", "аш", "иск", "орн", "уум",
		"екс", "илт", "оф", "уун", "ейр", "ак", "иим", "окс", "уум", "аз",
	}
)

// Katakana syllable pools for Japanese.
// Uses voiced consonants (Dakuten: Ga, Za, Da, Ba) for alien/threatening feel, per sci-fi convention.
var (
	jaPrefixSyll = []string{
		"ガ", "ゾ", "ヴァ", "ギ", "ザ", "ゼ", "ド", "バ", "ベ", "ボ",
		"グ", "ズ", "ヅ", "ヂ", "ヴィ", "ヴ", "ゴ", "ダ", "デ", "ブ",
		"グラ", "ズラ", "ブラ", "ドラ", "グレ", "ズレ", "ブレ", "ドレ", "ガル", "ギル",
	}
	jaMidSyll = []string{
		"ル", "ン", "ッ", "ー", "リ", "ラ", "ロ", "レ", "ル", "ン",
		"ル", "ン", "ッ", "ー", "リ", "ラ", "ロ", "レ", "ル", "ン",
		"ズ", "グ", "ザ", "ガ", "バ", "ド", "ブ", "ダ", "ヴ", "ゾ",
	}
	jaEndSyll = []string{
		"ス", "ト", "ク", "ズ", "クス", "グ", "ツ", "ド", "ル", "ン",
		"スク", "スト", "クト", "ズク", "ルク", "ング", "グス", "ドス", "ルト", "ンズ",
		"イクス", "オクス", "アクス", "ウス", "エクス", "ーク", "ート", "ール", "ック", "ーン",
	}
)

// Hangul syllable pools for Korean.
// Uses aspirated (ㅋ/ㅌ/ㅍ/ㅊ) and tense (ㄲ/ㄸ/ㅃ/ㅆ/ㅉ) consonants for alien sound, per sci-fi convention.
var (
	koPrefixSyll = []string{
		"카", "크", "제", "트", "블", "벡", "캬", "터", "포", "프",
		"쵸", "추", "키", "코", "쿠", "타", "토", "투", "파", "페",
		"크라", "트라", "프라", "블라", "그로", "즈로", "크로", "트로", "쁘", "쯔",
	}
	koMidSyll = []string{
		"락", "르", "나", "멘", "로", "란", "른", "넨", "론", "린",
		"락", "르", "나", "멘", "로", "란", "른", "넨", "론", "린",
		"타", "카", "파", "차", "크", "트", "프", "츠", "캬", "터",
	}
	koEndSyll = []string{
		"스", "론", "트", "스크", "즈", "크", "스", "론", "트", "스크",
		"르스", "론트", "크스", "스트", "트론", "크론", "스크", "른", "틀", "즌",
		"이스", "오스", "우스", "아스", "에스", "크스", "트스", "프스", "츠", "크",
	}
)

// Chinese character pools for Chinese — uses only phonetic transliteration characters (no semantic Hanzi).
// This signals "alien entity" to a Chinese reader, per sci-fi convention (Three-Body Problem style).
var (
	zhPrefixSyll = []string{
		"萨", "泽", "克", "尔", "斯", "达", "诺", "塔", "德", "加",
		"卡", "拉", "托", "尼", "罗", "基", "纳", "科", "米", "帕",
		"索", "奇", "迪", "伽", "雷", "法", "曼", "坦", "戈", "伦",
	}
	zhMidSyll = []string{
		"萨", "泽", "克", "尔", "斯", "达", "诺", "塔", "德", "加",
		"卡", "拉", "托", "尼", "罗", "基", "纳", "科", "米", "帕",
		"索", "奇", "迪", "伽", "雷", "法", "曼", "坦", "戈", "伦",
	}
	zhEndSyll = []string{
		"斯", "德", "克", "姆", "恩", "尔", "亚", "安", "乌", "兹",
		"格", "顿", "特", "金", "丁", "普", "兰", "森", "廷", "什",
		"伯", "奇", "弗", "达", "伦", "曼", "特", "斯", "克", "德",
	}
)

// Rank title pools — each damage type gets its own themed set of titles.
// Index 0 is empty for Rank 0 (Rookies).
var rankTitlePools = map[int][]string{
	DMG_PLASMA:    {"", language.String("RANK_TITLE_NAVIGATOR"), language.String("RANK_TITLE_COMMANDER"), language.String("RANK_TITLE_ELITE"), language.String("RANK_TITLE_OVERLORD")},
	DMG_LASER:     {"", language.String("RANK_TITLE_SENTINEL"), language.String("RANK_TITLE_ARBITER"), language.String("RANK_TITLE_WARDEN"), language.String("RANK_TITLE_SOVEREIGN")},
	DMG_EXPLOSIVE: {"", language.String("RANK_TITLE_DEMOLISHER"), language.String("RANK_TITLE_VANGUARD"), language.String("RANK_TITLE_BERSERKER"), language.String("RANK_TITLE_ANNIHILATOR")},
	DMG_MELEE:     {"", language.String("RANK_TITLE_STALKER"), language.String("RANK_TITLE_PREDATOR"), language.String("RANK_TITLE_REAPER"), language.String("RANK_TITLE_APEX")},
	DMG_KINETIC:   {"", language.String("RANK_TITLE_SCOUT"), language.String("RANK_TITLE_STRIKER"), language.String("RANK_TITLE_WARLORD"), language.String("RANK_TITLE_TITAN")},
	DMG_PSIONIC:   {"", language.String("RANK_TITLE_ACOLYTE"), language.String("RANK_TITLE_HIEROPHANT"), language.String("RANK_TITLE_ARCHON"), language.String("RANK_TITLE_NEXUS")},
}

// rankTitle returns the appropriate rank title for a given damage type and rank.
func rankTitle(dmgType, rank int) string {
	if rank <= 0 {
		return ""
	}
	pool, ok := rankTitlePools[dmgType]
	if !ok {
		pool = rankTitlePools[DMG_PLASMA]
	}
	idx := rank
	if idx >= len(pool) {
		idx = len(pool) - 1
	}
	return pool[idx]
}

// Weapons available to procedural aliens, keyed by damage type, ordered by rank.
// Each inner slice maps rank 0..4 to a pool of weapons for that rank.
var alienWeaponsByDmgType = map[int][][]string{
	DMG_PLASMA: {
		{"plasma_pistol"},
		{"plasma_pistol", "plasma_rifle"},
		{"plasma_rifle"},
		{"plasma_rifle", "heavy_plasma"},
		{"heavy_plasma"},
	},
	DMG_LASER: {
		{"alien_laser"},
		{"alien_laser"},
		{"alien_laser", "alien_heavy_laser"},
		{"alien_heavy_laser"},
		{"alien_heavy_laser"},
	},
	DMG_MELEE: {
		{"alien_claw"},
		{"alien_claw", "alien_fang"},
		{"alien_fang"},
		{"alien_fang"},
		{"alien_fang"},
	},
	DMG_KINETIC: {
		{"alien_blaster"},
		{"alien_blaster"},
		{"alien_blaster", "alien_cannon"},
		{"alien_cannon"},
		{"alien_cannon"},
	},
	DMG_EXPLOSIVE: {
		{"alien_grenade"},
		{"alien_grenade"},
		{"alien_grenade", "alien_rocket"},
		{"alien_rocket"},
		{"alien_rocket"},
	},
	DMG_PSIONIC: {
		{"alien_psi_bolt"},
		{"alien_psi_bolt", "plasma_pistol"},
		{"alien_psi_bolt"},
		{"alien_psi_bolt", "heavy_plasma"},
		{"alien_psi_bolt"},
	},
}

// Lore templates filled with species name and traits.
var loreTemplates = []string{
	language.String("LORE_TEMPLATE_1"),
	language.String("LORE_TEMPLATE_2"),
	language.String("LORE_TEMPLATE_3"),
	language.String("LORE_TEMPLATE_4"),
	language.String("LORE_TEMPLATE_5"),
	language.String("LORE_TEMPLATE_6"),
	language.String("LORE_TEMPLATE_7"),
	language.String("LORE_TEMPLATE_8"),
	language.String("LORE_TEMPLATE_9"),
	language.String("LORE_TEMPLATE_10"),
	language.String("LORE_TEMPLATE_11"),
	language.String("LORE_TEMPLATE_12"),
	language.String("LORE_TEMPLATE_13"),
	language.String("LORE_TEMPLATE_14"),
	language.String("LORE_TEMPLATE_15"),
}

// Morphology lore snippets describing body subtypes.
var morphLoreSnippets = map[string]string{
	SubtypeCarbonFlesh:  language.String("MORPH_LORE_CARBON"),
	SubtypeSilicon:      language.String("MORPH_LORE_SILICON"),
	SubtypeGaseous:      language.String("MORPH_LORE_GAS"),
	SubtypeCrystalline:  language.String("MORPH_LORE_CRYSTAL"),
	SubtypeAmorphous:    language.String("MORPH_LORE_AMORPHOUS"),
	SubtypeMechanical:   language.String("MORPH_LORE_MECHANICAL"),
	SubtypeBioSynthetic: language.String("MORPH_LORE_BIOSYNTH"),
	SubtypeNanotech:     language.String("MORPH_LORE_NANOTECH"),
}

// Sense lore snippets for flavor text.
var senseLoreSnippets = map[string]string{
	SenseNone:      language.String("SENSE_LORE_NONE"),
	SensePoor:      language.String("SENSE_LORE_POOR"),
	SenseNormal:    language.String("SENSE_LORE_NORMAL"),
	SenseExcellent: language.String("SENSE_LORE_EXCELLENT"),
	SenseMultiSpec: language.String("SENSE_LORE_MULTISPEC"),
	SenseEcholoc:   language.String("SENSE_LORE_ECHOLOC"),
}

// Limb lore descriptions.
func limbLore(m *Morphology) string {
	var armDesc string
	switch {
	case m.Arms == 0:
		armDesc = language.String("MORPH_ARMS_NONE")
	case m.Arms == 1:
		armDesc = language.String("MORPH_ARMS_1")
	case m.Arms == 2:
		armDesc = language.String("MORPH_ARMS_2")
	case m.Arms <= 4:
		armDesc = language.Sprintf("MORPH_ARMS_MULTI", m.Arms)
	default:
		armDesc = language.Sprintf("MORPH_ARMS_MASS", m.Arms)
	}
	legDesc := language.String("MORPH_LEGS_HOVER")
	switch {
	case m.Legs == 1:
		legDesc = language.String("MORPH_LEGS_1")
	case m.Legs == 2:
		legDesc = language.String("MORPH_LEGS_2")
	case m.Legs == 4:
		legDesc = language.String("MORPH_LEGS_4")
	case m.Legs >= 6:
		legDesc = language.Sprintf("MORPH_LEGS_SCURRY", m.Legs)
	}
	return language.Sprintf("MORPH_WRAP", armDesc, legDesc)
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

	runes := []rune(name)
	prefix := strings.ToUpper(string(runes[:min(3, len(runes))]))

	primaryDMG := rng.Intn(6) // DMG_PLASMA..DMG_PSIONIC

	// Generate morphology
	morph := generateMorphology(rng, primaryDMG)

	sp := &AlienSpecies{
		Name:       name,
		Prefix:     prefix,
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

	if len(sp.Types) > 0 {
		sp.BaseIcon = sp.Types[0].Icon
		sp.BaseStyle = sp.Types[0].Style
	}

	// Generate species lore
	sp.Lore = generateLore(name, primaryDMG, morph)

	return sp
}

func generateName(rng *rand.Rand) string {
	p, m, e := poolsForLang(language.Current())
	pi := p[midSyllIdx(rng, len(p))]
	mi := m[rng.Intn(len(m))]
	ei := e[rng.Intn(len(e))]
	return pi + mi + ei
}

func poolsForLang(lang string) (prefix, mid, end []string) {
	switch lang {
	case "ru":
		return ruPrefixSyll, ruMidSyll, ruEndSyll
	case "ja":
		return jaPrefixSyll, jaMidSyll, jaEndSyll
	case "ko":
		return koPrefixSyll, koMidSyll, koEndSyll
	case "zh":
		return zhPrefixSyll, zhMidSyll, zhEndSyll
	default:
		return prefixSyll, midSyll, endSyll
	}
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

	// 1. Determine body subtype based on body type and damage affinity.
	var subtype string
	if bodyType == BodyOrganic {
		subtype = pickOrganicSubtype(rng, dmgType)
	} else {
		subtype = pickSyntheticSubtype(rng, dmgType)
	}

	// 2. Generate limb counts — constrained by body subtype (e.g. amorphous have no legs).
	arms := generateArmCount(rng, subtype)
	legs := generateLegCount(rng, subtype)

	// 3. Generate base senses.
	eyesight := pickSenseQuality(rng)
	hearing := pickHearingQuality(rng)
	thermal := pickBinarySense(rng)
	psiSense := pickBinarySense(rng)
	chemSense := pickBinarySense(rng)

	// 4. Apply subtype-specific sense restrictions and bonuses.
	if subtype == SubtypeMechanical || subtype == SubtypeNanotech {
		chemSense = SenseNone
	}
	if subtype == SubtypeGaseous || subtype == SubtypeAmorphous {
		// Amorphous/gaseous: compensate for lack of physical form with psi and thermal.
		if psiSense == SenseNone {
			psiSense = SenseLow
		}
		thermal = SenseHigh
	}
	if subtype == SubtypeCrystalline {
		// Crystalline: light refraction grants excellent eyesight, but poor hearing.
		hearing = SensePoor
		if eyesight == SenseNormal || eyesight == SensePoor {
			eyesight = SenseExcellent
		}
	}
	if subtype == SubtypeSilicon {
		// Silicon-based: usually possess thermal vision.
		thermal = SenseHigh
	}
	if dmgType == DMG_PSIONIC && psiSense == SenseNone {
		// Psionic species must have at least some psionic sense.
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
		DamageType:    dmgType,
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
		// Mostly limbless floaters, but tentacled variants get 2 arms.
		if rng.Intn(3) == 0 {
			return 2
		}
		return 0
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
		// Mostly floating, but tentacled variants slither on 2 legs.
		if subtype != SubtypeNanotech && rng.Intn(3) == 0 {
			return 2
		}
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

// pickAlienWeapon selects a weapon for a procedural alien based on its damage type and rank.
func pickAlienWeapon(rng *rand.Rand, dmgType int, rank int) string {
	pool, ok := alienWeaponsByDmgType[dmgType]
	if !ok || rank >= len(pool) {
		return "plasma_pistol"
	}
	weaps := pool[rank]
	return weaps[rng.Intn(len(weaps))]
}

// --- Morphology stat modifiers ---
// Each modifier applies tradeoffs: strong advantages come with compensating weaknesses.

func morphHPMod(m *Morphology) int {
	mod := 0
	if m.IsFloating() {
		mod -= 3
	}
	switch m.BodySubtype {
	case SubtypeCrystalline:
		mod += 5
	case SubtypeAmorphous:
		mod -= 2
	case SubtypeNanotech:
		mod -= 4
	case SubtypeGaseous:
		mod -= 1 // gaseous: low HP compensated by kinetic immunity
	case SubtypeMechanical:
		mod += 2 // mechanical: slightly tougher
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
	switch m.BodySubtype {
	case SubtypeCrystalline:
		mod -= 10
	case SubtypeMechanical:
		mod += 5
	case SubtypeAmorphous:
		mod -= 3 // amorphous: slow movement
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
	switch m.BodySubtype {
	case SubtypeMechanical:
		mod += 5
	case SubtypeGaseous:
		mod -= 5 // gaseous: unfocused form, harder to aim
	case SubtypeCrystalline:
		mod -= 3 // crystalline: rigid, less precise
	case SubtypeAmorphous:
		mod -= 3 // amorphous: shifting form
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
	// Tradeoff: floating + high reactions → lower bravery (fragile confidence)
	if m.IsFloating() && m.BodySubtype == SubtypeMechanical {
		mod -= 3 // mechanical hovering: predictable movement patterns
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
	switch m.BodySubtype {
	case SubtypeCrystalline:
		mod += 5
	case SubtypeGaseous:
		mod -= 5 // gaseous: no physical strength
	case SubtypeAmorphous:
		mod -= 3
	case SubtypeNanotech:
		mod -= 2
	}
	return mod
}

func morphPsiMod(m *Morphology) int {
	mod := 0
	if m.Eyesight == SenseNone {
		mod += 10
	}
	if m.PsionicSense == SenseHigh {
		mod += 15
	} else if m.PsionicSense == SenseLow {
		mod += 5
	}
	if m.BodySubtype == SubtypeAmorphous || m.BodySubtype == SubtypeGaseous {
		mod += 10
	}
	// Tradeoff: mechanical/nanotech can't develop psi
	if m.BodySubtype == SubtypeMechanical || m.BodySubtype == SubtypeNanotech {
		mod -= 20
	}
	return mod
}

func morphBraveMod(m *Morphology) int {
	mod := 0
	// Floating aliens: more evasive but less resolute
	if m.IsFloating() {
		mod -= 5
	}
	// Mechanical: fearless (no self-preservation instinct)
	if m.BodySubtype == SubtypeMechanical {
		mod += 15
	}
	// Crystalline: slow to panic
	if m.BodySubtype == SubtypeCrystalline {
		mod += 10
	}
	// Amorphous/gaseous: hard to intimidate
	if m.BodySubtype == SubtypeAmorphous || m.BodySubtype == SubtypeGaseous {
		mod += 5
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
	braveBase += morphBraveMod(m)
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

	// Choose weapon based on damage type and rank
	weapon := pickAlienWeapon(rng, sp.PrimaryDMG, rank)

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
	if title := rankTitle(sp.PrimaryDMG, rank); title != "" {
		varName = sp.Name + " " + title
	}

	// Build short name
	shortName := sp.Prefix
	if rank > 0 {
		shortName += string(rune('A' + rank - 1))
	}

	// Icon and Style: selected based on morphology and biology.
	// DetermineProceduralIconAndStyle marks the chosen rune in usedIcons internally.
	icon, style, fgColor := DetermineProceduralIconAndStyle(m, rng, usedIcons)

	// Lore per variant
	variantLore := sp.Lore
	if title := rankTitle(sp.PrimaryDMG, rank); title != "" {
		variantLore = title + " of the " + sp.Name + " species. " + variantLore
	}

	return &AlienType{
		Name:       varName,
		ShortName:  shortName,
		Icon:       icon,
		Style:      style,
		FgColor:    fgColor,
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
		Morphology: m,
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

func generateLore(name string, dmgType int, m *Morphology) string {
	// Mix damage type and body subtype for template variety
	subtypeHash := 0
	for _, c := range m.BodySubtype {
		subtypeHash += int(c)
	}
	idx := (dmgType*7 + subtypeHash) % len(loreTemplates)
	template := loreTemplates[idx]
	base := strings.Replace(template, "%s", DamageTypeStr(dmgType), 1)

	bodyDesc := morphLoreSnippets[m.BodySubtype]
	if bodyDesc == "" {
		bodyDesc = language.String("LORE_UNKNOWN_BIOLOGY")
	}

	limbDesc := limbLore(m)

	senseDesc := ""
	if m.Eyesight == SenseNone {
		senseDesc = language.String("LORE_SENSE_PREFIX") + senseLoreSnippets[SenseNone] + language.String("LORE_SENSE_SUFFIX")
	} else if m.Eyesight == SenseMultiSpec {
		senseDesc = language.String("LORE_SENSE_PREFIX") + senseLoreSnippets[SenseMultiSpec] + language.String("LORE_SENSE_SUFFIX")
	} else if m.Hearing == SenseEcholoc {
		senseDesc = language.String("LORE_SENSE_PREFIX") + senseLoreSnippets[SenseEcholoc] + language.String("LORE_SENSE_SUFFIX")
	}

	return fmt.Sprintf("%s %s. "+language.String("LORE_BODY_FORMAT")+"%s", limbDesc, base, bodyDesc, senseDesc)
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
