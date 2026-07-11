package data

import (
	"math/rand"
)

// --- Trait enums derived from Morphology ---

type Sense int

const (
	SenseStandard    Sense = iota
	SenseEcholocation
	SenseOmni
)

type Manipulators int

const (
	ManipNone     Manipulators = iota
	ManipBipedal
	ManipMultiArmed
)

type Locomotion int

const (
	LocomFloating Locomotion = iota
	LocomBipedal
	LocomArachnid
)

// --- Tagged module types ---

type TaggedHead struct {
	Pixels []string
	Sense  Sense
}

type TaggedTorso struct {
	Pixels       []string
	Manipulators Manipulators
}

type TaggedLegs struct {
	Pixels     []string
	Locomotion Locomotion
}

// --- AlienPixels: body + weapon layers ---

type AlienPixels struct {
	Body   [24][20]bool
	Weapon [24][20]bool
}

// --- Registry ---

type SpriteRegistry struct {
	Heads  []TaggedHead
	Torsos []TaggedTorso
	Legs   []TaggedLegs
}

func NewAlienSpriteRegistry() *SpriteRegistry {
	return &SpriteRegistry{
		Heads: []TaggedHead{
			{Pixels: headStandard, Sense: SenseStandard},
			{Pixels: headEcholocation, Sense: SenseEcholocation},
			{Pixels: headOmni, Sense: SenseOmni},
		},
		Torsos: []TaggedTorso{
			{Pixels: torsoBipedalArms, Manipulators: ManipBipedal},
			{Pixels: torsoMultiArmed, Manipulators: ManipMultiArmed},
			{Pixels: torsoNone, Manipulators: ManipNone},
		},
		Legs: []TaggedLegs{
			{Pixels: legsBipedal, Locomotion: LocomBipedal},
			{Pixels: legsArachnid, Locomotion: LocomArachnid},
			{Pixels: legsFloating, Locomotion: LocomFloating},
		},
	}
}

// --- Head templates (10 rows x 20 wide) ---

var headStandard = []string{
	"....................",
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXX......XXX....",
	"....X...XX...X......",
	"....XXX......XXX....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headEcholocation = []string{
	"....................",
	".XX..............XX.",
	".XXX............XXX.",
	".XXXX..........XXXX.",
	"..XXXX........XXXX..",
	"...XXXXXXXXXXXXXX...",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
}

var headOmni = []string{
	"....................",
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"...XX..XXXX..XX.....",
	"..XX....XX....XX....",
	"...XX..XXXX..XX.....",
	"....XXXXXXXXXXXX....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
}

// --- Torso templates (8 rows x 20 wide) ---
// 'X' = body, 'W' = weapon (rendered in a distinct lighter color).

var torsoNone = []string{
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var torsoBipedalArms = []string{
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXXX.XX.XXXX....",
	"...XXXX..XX..XXXX...",
	"..XXXXX..XX..XXXXX..",
	".........XX.WWWWW...",
	".........XX.WWWW....",
	".........WWWWWWW....",
}

var torsoMultiArmed = []string{
	".XXXX.XXXXXXXX.XXXX.",
	".XXXX.XXXXXXXX.XXXX.",
	"..XX...XXXXXX...XX..",
	".......XXXXXX.......",
	".XXXX..XXXXXX..XXXX.",
	".XXXX..XXXXXX..XXXX.",
	"..XX...XXXXXX...XX..",
	".......XXXXXX.......",
}

// --- Leg templates (6 rows x 20 wide) ---

var legsFloating = []string{
	".......XXXXXX.......",
	".......XXXXXX.......",
	"........XXXX........",
	"........XXXX........",
	".........XX.........",
	".........XX.........",
}

var legsBipedal = []string{
	".......XXXXXX.......",
	"......XXX..XXX......",
	"......XX....XX......",
	".....XX......XX.....",
	".....XX......XX.....",
	"....XXX......XXX....",
}

var legsArachnid = []string{
	".......XXXXXX.......",
	"......XX.XX.XX......",
	".....XX..XX..XX.....",
	"....XX...XX...XX....",
	"...XX....XX....XX...",
	"..XX.....XX.....XX..",
}

// --- Morphology -> trait mapping ---

func SenseFromMorphology(m *Morphology) Sense {
	if m.Hearing == "echolocation" {
		return SenseEcholocation
	}
	exotic := 0
	if m.PsionicSense == "high" {
		exotic++
	}
	if m.ThermalSense == "high" {
		exotic++
	}
	if m.ChemicalSense == "high" {
		exotic++
	}
	if exotic >= 2 {
		return SenseOmni
	}
	return SenseStandard
}

func ManipulatorsFromMorphology(m *Morphology) Manipulators {
	if m.Arms == 0 {
		return ManipNone
	}
	if m.Arms <= 2 {
		return ManipBipedal
	}
	return ManipMultiArmed
}

func LocomotionFromMorphology(m *Morphology) Locomotion {
	if m.Legs == 0 {
		return LocomFloating
	}
	if m.Legs <= 2 {
		return LocomBipedal
	}
	return LocomArachnid
}

// --- Generation ---

// GenerateAlienPixels produces a 24x20 pixel grid (body + weapon layers)
// by assembling trait-matched head, torso, and leg modules.
func GenerateAlienPixels(seed int64, m *Morphology) AlienPixels {
	if m == nil {
		m = &Morphology{
			Arms: 2, Legs: 2, BodyType: BodyOrganic, BodySubtype: SubtypeCarbonFlesh,
			Eyesight: "normal", Hearing: "normal",
		}
	}

	reg := NewAlienSpriteRegistry()
	rng := rand.New(rand.NewSource(seed))

	sense := SenseFromMorphology(m)
	manip := ManipulatorsFromMorphology(m)
	loco := LocomotionFromMorphology(m)

	head := pickHead(reg.Heads, sense, rng)
	torso := pickTorso(reg.Torsos, manip, rng)
	legs := pickLegs(reg.Legs, loco, rng)

	var result AlienPixels

	for y, row := range head {
		if y >= 10 {
			break
		}
		for x, ch := range row {
			if ch == 'X' {
				result.Body[y][x] = true
			}
		}
	}

	for y, row := range torso {
		ty := 10 + y
		if ty >= 18 {
			break
		}
		for x, ch := range row {
			switch ch {
			case 'X':
				result.Body[ty][x] = true
			case 'W':
				result.Weapon[ty][x] = true
			}
		}
	}

	for y, row := range legs {
		ly := 18 + y
		if ly >= 24 {
			break
		}
		for x, ch := range row {
			if ch == 'X' {
				result.Body[ly][x] = true
			}
		}
	}

	return result
}

func pickHead(candidates []TaggedHead, sense Sense, rng *rand.Rand) []string {
	var filtered [][]string
	for _, h := range candidates {
		if h.Sense == sense {
			filtered = append(filtered, h.Pixels)
		}
	}
	if len(filtered) == 0 {
		return headStandard
	}
	return filtered[rng.Intn(len(filtered))]
}

func pickTorso(candidates []TaggedTorso, manip Manipulators, rng *rand.Rand) []string {
	var filtered [][]string
	for _, t := range candidates {
		if t.Manipulators == manip {
			filtered = append(filtered, t.Pixels)
		}
	}
	if len(filtered) == 0 {
		return torsoBipedalArms
	}
	return filtered[rng.Intn(len(filtered))]
}

func pickLegs(candidates []TaggedLegs, loco Locomotion, rng *rand.Rand) []string {
	var filtered [][]string
	for _, l := range candidates {
		if l.Locomotion == loco {
			filtered = append(filtered, l.Pixels)
		}
	}
	if len(filtered) == 0 {
		return legsBipedal
	}
	return filtered[rng.Intn(len(filtered))]
}

// AlienColorFromSeed returns a deterministic RGB color derived from seed.
func AlienColorFromSeed(seed int64) (r, g, b int32) {
	rng := rand.New(rand.NewSource(seed))
	return int32(rng.Intn(150) + 100),
		int32(rng.Intn(150) + 100),
		int32(rng.Intn(150) + 100)
}

// AlienWeaponColor returns a lighter variant of the body color for the weapon.
func AlienWeaponColor(bodyR, bodyG, bodyB int32) (r, g, b int32) {
	boost := int32(80)
	r = bodyR + boost
	g = bodyG + boost
	b = bodyB + boost
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	return r, g, b
}

// CompressToHalfBlocks converts a 24x20 boolean pixel grid to a 12x10
// rune grid using half-block characters.
func CompressToHalfBlocks(pixels [24][20]bool) [12][10]rune {
	var result [12][10]rune
	for y := 0; y < 12; y++ {
		for x := 0; x < 10; x++ {
			top := pixels[y*2][x]
			bottom := pixels[y*2+1][x]
			switch {
			case top && bottom:
				result[y][x] = '\u2588'
			case top:
				result[y][x] = '\u2580'
			case bottom:
				result[y][x] = '\u2584'
			default:
				result[y][x] = ' '
			}
		}
	}
	return result
}
