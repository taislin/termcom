package data

import (
	"math/rand"
)

// --- Trait enums derived from Morphology ---

type Sense int

const (
	SenseStandard Sense = iota
	SenseEcholocation
	SenseOmni
)

type Manipulators int

const (
	ManipNone Manipulators = iota
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
	Body      [24][20]bool
	Weapon    [24][20]bool
	Shadow    [24][20]bool
	Highlight [24][20]bool
	Accent    [24][20]bool
	Eyes      [24][20]bool
}

// --- Registry ---

type SpriteRegistry struct {
	Heads  []TaggedHead
	Torsos []TaggedTorso
	Legs   []TaggedLegs
}

var defaultSpriteRegistry *SpriteRegistry

func NewAlienSpriteRegistry() *SpriteRegistry {
	if defaultSpriteRegistry == nil {
		defaultSpriteRegistry = &SpriteRegistry{
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
	return defaultSpriteRegistry
}

// --- Head templates (10 rows x 20 wide) ---

var headStandard = []string{
	"....................",
	"......XXXXXX......",
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

	headOffset := centerOffset(head[0], 20)
	for y, row := range head {
		if y >= 10 {
			break
		}
		for x, ch := range row {
			if ch == 'X' {
				result.Body[y][x+headOffset] = true
			}
		}
	}

	if m.Eyesight != SenseNone && m.Eyesight != "none" && m.Eyesight != "blind" {
		for y := 0; y < 10; y++ {
			for x := 0; x < 20; x++ {
				if y >= len(head) || x+headOffset >= len(head[y]) {
					continue
				}
				ch := rune(head[y][x+headOffset])
				if ch == 'X' {
					if (x == 6 || x == 13) && y >= 3 && y <= 6 {
						result.Eyes[y][x+headOffset] = true
					}
					if y == 4 && (x == 7 || x == 12) {
						result.Eyes[y][x+headOffset] = true
					}
				}
			}
		}
	}

	torsoOffset := centerOffset(torso[0], 20)
	for y, row := range torso {
		ty := 10 + y
		if ty >= 18 {
			break
		}
		for x, ch := range row {
			switch ch {
			case 'X':
				result.Body[ty][x+torsoOffset] = true
			case 'W':
				result.Weapon[ty][x+torsoOffset] = true
			}
		}
	}

	legsOffset := centerOffset(legs[0], 20)
	for y, row := range legs {
		ly := 18 + y
		if ly >= 24 {
			break
		}
		for x, ch := range row {
			if ch == 'X' {
				result.Body[ly][x+legsOffset] = true
			}
		}
	}

	for y := 0; y < 24; y++ {
		for x := 0; x < 20; x++ {
			if !result.Body[y][x] {
				continue
			}
			if x > 0 && x < 19 && y > 0 && y < 23 {
				left := result.Body[y][x-1]
				right := result.Body[y][x+1]
				up := result.Body[y-1][x]
				down := result.Body[y+1][x]
				if !left && right {
					result.Highlight[y][x] = true
				}
				if !right && left {
					result.Shadow[y][x] = true
				}
				if !up && down {
					result.Accent[y][x] = true
				}
			}
			if y == 0 || y == 23 || x == 0 || x == 19 {
				result.Shadow[y][x] = true
			}
		}
	}

	return result
}

func centerOffset(row string, width int) int {
	left := 0
	right := len(row) - 1
	for left < len(row) && row[left] == '.' {
		left++
	}
	for right >= 0 && row[right] == '.' {
		right--
	}
	if left > right {
		return 0
	}
	trimmed := right - left + 1
	if trimmed >= width {
		return 0
	}
	return (width-trimmed)/2 - left
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

// AlienWeaponColor returns a metallic-grey variant for weapons.
func AlienWeaponColor() (r, g, b int32) {
	return 170, 180, 190
}
