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

type EyeStyle int

const (
	EyeClassic EyeStyle = iota
	EyeCyclops
	EyeArachnid
	EyeVisor
	EyeNone
)

// --- Tagged module types ---

type TaggedHead struct {
	Pixels []string
	Sense  Sense
}

type TaggedEyes struct {
	Pixels []string
	Style  EyeStyle
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
	Mouth     [24][20]bool
	Interior  [24][20]bool
	Belly     [24][20]bool
	Texture   [24][20]bool
}

// --- Registry ---

type SpriteRegistry struct {
	Heads  []TaggedHead
	Eyes   []TaggedEyes
	Torsos []TaggedTorso
	Legs   []TaggedLegs
}

var defaultSpriteRegistry *SpriteRegistry

func NewAlienSpriteRegistry() *SpriteRegistry {
	if defaultSpriteRegistry == nil {
		defaultSpriteRegistry = &SpriteRegistry{
			Heads: []TaggedHead{
				{Pixels: headRound, Sense: SenseStandard},
				{Pixels: headSquare, Sense: SenseStandard},
				{Pixels: headTall, Sense: SenseStandard},
				{Pixels: headSkull, Sense: SenseStandard},
				{Pixels: headWide, Sense: SenseEcholocation},
				{Pixels: headAntennae, Sense: SenseEcholocation},
				{Pixels: headVisor, Sense: SenseOmni},
				{Pixels: headCone, Sense: SenseOmni},
			},
			Eyes: []TaggedEyes{
				{Pixels: eyeClassic, Style: EyeClassic},
				{Pixels: eyeCyclops, Style: EyeCyclops},
				{Pixels: eyeArachnid, Style: EyeArachnid},
				{Pixels: eyeVisor, Style: EyeVisor},
				{Pixels: eyeNone, Style: EyeNone},
			},
			Torsos: []TaggedTorso{
				{Pixels: torsoSlim, Manipulators: ManipBipedal},
				{Pixels: torsoWide, Manipulators: ManipBipedal},
				{Pixels: torsoArmored, Manipulators: ManipBipedal},
				{Pixels: torsoHollow, Manipulators: ManipBipedal},
				{Pixels: torsoAsymmetric, Manipulators: ManipBipedal},
				{Pixels: torsoBladed, Manipulators: ManipBipedal},
				{Pixels: torsoTentacle, Manipulators: ManipNone},
				{Pixels: torsoMultiArmed, Manipulators: ManipMultiArmed},
			},
			Legs: []TaggedLegs{
				{Pixels: legsBipedal, Locomotion: LocomBipedal},
				{Pixels: legsLong, Locomotion: LocomBipedal},
				{Pixels: legsWide, Locomotion: LocomBipedal},
				{Pixels: legsArachnid, Locomotion: LocomArachnid},
				{Pixels: legsSerpentine, Locomotion: LocomFloating},
				{Pixels: legsTentacle, Locomotion: LocomFloating},
				{Pixels: legsPillar, Locomotion: LocomFloating},
			},
		}
	}
	return defaultSpriteRegistry
}

// --- Head templates (10 rows x 20 wide) ---
// Detail chars: e=eye, m=mouth, a=antenna, h=highlight, d=dark

var headRound = []string{
	"....................",
	".......XXXXXX.......",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	"....XXXmmXXXmXXX....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headSquare = []string{
	"....................",
	"....XXXXXXXXXXXX....",
	"...XXXXXXXXXXXXXXX..",
	"...XXXXXXXXXXXXXXX..",
	"...XXXXXXXXXXXXXXX..",
	"...XXXXXXXXXXXXXXX..",
	"...XXmmmmmmmmmmXX...",
	"...XXXXXXXXXXXXXXX..",
	"...XXXXXXXXXXXXXXX..",
	"....XXXXXXXXXXXX....",
}

var headTall = []string{
	".......XXXX.........",
	"......XXXXXX........",
	"......XXXXXX........",
	".....XXXXXXXX.......",
	".....XXXXXXXXX......",
	".....XXXXXXXX.......",
	".....XXmmmmXX.......",
	".....XXXXXXXX.......",
	"......XXXXXX........",
	".......XXXX.........",
}

var headSkull = []string{
	"....................",
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXdXXXXXdXXX....",
	"....XdXXXdXXXXdX....",
	"....XXXXXXXXXXXX....",
	"....XXXddXXddXXX....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headWide = []string{
	"....................",
	"XX................XX",
	"XXX..............XXX",
	"XXXX............XXXX",
	"XXXXXXXXXXXXXXXXXXXX",
	".XXXXXXXXXXXXXXXXXX.",
	".XXXXXXmmmmmmXXXXXX.",
	".XXXXXXXXXXXXXXXXXX.",
	".XXXXXXXXXXXXXXXXXX.",
	"..XXXXXXXXXXXXXXX...",
}

var headAntennae = []string{
	"..X...........X.....",
	"..XX.........XX.....",
	"...XX.......XX......",
	"....XXXXXXXXXX......",
	"....XXXXXXXXXX......",
	"....XXXXXXXXXX......",
	"....XXmmmmXXX.......",
	".....XXXXXXXX.......",
	"......XXXXXX........",
	".......XXXX.........",
}

var headVisor = []string{
	"....................",
	"......XXXXXX........",
	".....XXXXXXXXXX.....",
	"....dddXXXXXXddd....",
	"....dXXXXXXXXdX.....",
	"....XXXXXXXXXXXX....",
	"....XXXmmmmXXXX.....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headCone = []string{
	"........XX..........",
	"........XXXX........",
	".......XXXXXX.......",
	"......XXXXXXXX......",
	"......XXeXXeXX......",
	".....XXXXXXXXXX.....",
	".....XXXmmmmXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	".....XXXXXXXXXX.....",
}

// --- Eye templates (10 rows x 20 wide) ---
// These act as masks: 'X' carves a hole in the head silhouette and
// marks the eyes layer for white/glowing rendering.

var eyeClassic = []string{
	"....................",
	"....................",
	"....................",
	".....XXX...XXX......",
	"....XXXX...XXXX.....",
	"....XXXX...XXXX.....",
	".....XX.....XX......",
	"....................",
	"....................",
	"....................",
}

var eyeCyclops = []string{
	"....................",
	"....................",
	"....................",
	".......XXXXX........",
	".......XXXXX........",
	".......XXXXX........",
	"....................",
	"....................",
	"....................",
	"....................",
}

var eyeArachnid = []string{
	"....................",
	"........X...X.......",
	"....................",
	".....X...X.X...X....",
	"....................",
	".....X...X.X...X....",
	"....................",
	"........X...X.......",
	"....................",
	"....................",
}

var eyeVisor = []string{
	"....................",
	"....................",
	"....................",
	"....................",
	"....XXXXXXXXXXXX....",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
}

var eyeNone = []string{
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
	"....................",
}

// --- Torso templates (8 rows x 20 wide) ---
// 'X' = body, 'W' = weapon, 'a' = accent, 'h' = highlight, 'd' = dark

var torsoSlim = []string{
	".......XXXXXX.......",
	".......XXXXXX.......",
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	".......XXXXXX.WWWWWW",
	".......XXXXWWWWWWW..",
	".......XXXXX........",
	".......XXXXXX.......",
}

var torsoWide = []string{
	".....XXXXXXXXXX.....",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	".....XXXXXXXXXWWWWWW",
	".....XXXXXXWWWWWWW..",
	".....XXXXXXXXXX.....",
	".....XXXXXXXXXX.....",
}

var torsoArmored = []string{
	".....dddXXXXdd......",
	"....dXXXXXXXXXXd....",
	"...XXdXXXXXXXXdXX...",
	 "...XX.XX.XX.XXXX...",
	 "..XXXX.XX.XXXXWWWWW",
	 "..XXXX.XXWWWWWWWW..",
	 "..XXXX.WWWWWWWWW...",
	 "...XXX.XXXX.XXX....",
}

var torsoTentacle = []string{
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"X......XXXXXX......X",
	"XX..XXXXXXXXXXXX..XX",
	".XXXX..XXXXXX..XXXX.",
	".......XXXXXX.......",
}

var torsoMultiArmed = []string{
	".XXXXXXXXXXXXXXXXXX.",
	".XXX..XXXXXXXX..XXX.",
	"..XX...XXXXXX...XX..",
	".......XXXXXX.......",
	".XXXXXXXXXXXXXXXXXX.",
	".XXX...XXXXXX...XXX.",
	"..XX...XXXXXX...XX..",
	".......XXXXXX.......",
}

var torsoHollow = []string{
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXX....XXXX.....",
	"....XX......XXX.....",
	"....XXX....XXXX.....",
	".....XXXXXXXXXX.....",
	"......XXXX.WWW......",
	"..........WWWWW......",
}

var torsoAsymmetric = []string{
	".....XXXXXXX........",
	"....XXXXXXXXXX......",
	"...XXXX.XX.XXXXX....",
	"..XXXXX.XX.XXXXXX...",
	"..XXXX.XX.XXXXXXX...",
	"......XX.WWWW.......",
	".......WWWWWWW......",
	"........WWWWW.......",
}

var torsoBladed = []string{
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXXX.XX.XXXX....",
	"..WWXXXX.XX.XXXXWW..",
	"..WWXXXXXXXXXXXXWW..",
	"......XX.XX.WWW.....",
	".......XX...WW......",
	"..............WW....",
}

// --- Leg templates (6 rows x 20 wide) ---

var legsBipedal = []string{
	".......XXXXXX.......",
	"......XXX..XXX......",
	"......XX....XX......",
	".....XX......XX.....",
	".....XX......XX.....",
	"....XXX......XXX....",
}

var legsLong = []string{
	".......XXXXXX.......",
	".......XXXXXX.......",
	"........XXXX........",
	"........XXXX........",
	"........XXXX........",
	".......XX..XX.......",
}

var legsWide = []string{
	".......XXXXXX.......",
	"......XXXXXXXX......",
	".....XX......XX.....",
	"....XX........XX....",
	"...XX..........XX...",
	"..XX............XX..",
}

var legsArachnid = []string{
	".......XXXXXX.......",
	"......XX.XX.XX......",
	".....XX..XX..XX.....",
	"....XX...XX...XX....",
	"...XX....XX....XX...",
	"..XX.....XX.....XX..",
}

var legsTentacle = []string{
	".......XXXXX........",
	".......X..XX........",
	"......XX..XX........",
	"XX...XX...XX...XX...",
	".XX.XX.....XX.XX....",
	".XXXX.......XXX.....",
}

var legsPillar = []string{
	".......XXXXXX.......",
	"........XXXX........",
	"........XXXX........",
	"........XXXX........",
	"........XXXX........",
	".......XXXXXX.......",
}

var legsSerpentine = []string{
	".........XXXXXXXX...",
	".........XXXXXXXX...",
	"........XXXXXXXX....",
	"......XXXXXXXXX.....",
	"..XXXXXXXXXX........",
	"XXXXXXXXXX..........",
}

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
	eyes := pickEyes(reg.Eyes, m, rng)
	torso := pickTorso(reg.Torsos, manip, rng)
	legs := pickLegs(reg.Legs, loco, rng)

	var result AlienPixels

	headOffset := centerOffset(head[0], 20)
	for y, row := range head {
		if y >= 10 {
			break
		}
		for x, ch := range row {
			if ch == 'X' || ch == 'm' || ch == 'a' || ch == 'h' || ch == 'd' {
				result.Body[y][x+headOffset] = true
			}
			if ch == 'm' {
				result.Mouth[y][x+headOffset] = true
			}
			if ch == 'a' {
				result.Accent[y][x+headOffset] = true
			}
			if ch == 'h' {
				result.Highlight[y][x+headOffset] = true
			}
			if ch == 'd' {
				result.Shadow[y][x+headOffset] = true
			}
		}
	}

	// Eye masking: overlay eyes template on the head area.
	// For each 'X' in the eye template, carve a hole in the body
	// and mark it on the Eyes layer for white/glowing rendering.
	if eyes != nil {
		eyeOffset := centerOffset(eyes[0], 20)
		for y := 0; y < 10 && y < len(eyes); y++ {
			for x := 0; x < len(eyes[y]) && x < 20; x++ {
				if eyes[y][x] != 'X' {
					continue
				}
				ex := x + eyeOffset
				ey := y
				if ex < 0 || ex >= 20 || ey < 0 || ey >= 10 {
					continue
				}
				result.Body[ey][ex] = false
				result.Eyes[ey][ex] = true
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
			case 'a':
				result.Body[ty][x+torsoOffset] = true
				result.Accent[ty][x+torsoOffset] = true
			case 'h':
				result.Body[ty][x+torsoOffset] = true
				result.Highlight[ty][x+torsoOffset] = true
			case 'd':
				result.Body[ty][x+torsoOffset] = true
				result.Shadow[ty][x+torsoOffset] = true
			}
		}
	}

	legsOffset := centerOffset(legs[0], 20)
	if legsOffset < 0 {
		legsOffset = 0
	}
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

	// Edge detection for 3D shading
	texRng := rand.New(rand.NewSource(seed ^ 0xF0F0F0F0F0))

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
				if !left && right && !result.Highlight[y][x] {
					result.Highlight[y][x] = true
				}
				if !right && left && !result.Shadow[y][x] {
					result.Shadow[y][x] = true
				}
				if !up && down && !result.Accent[y][x] {
					result.Accent[y][x] = true
				}
				if left && right && up && down {
					result.Interior[y][x] = true
				}
			}
			if y == 0 || y == 23 || x == 0 || x == 19 {
				if !result.Shadow[y][x] {
					result.Shadow[y][x] = true
				}
			}

			// Belly patch: central torso area
			if y >= 11 && y <= 16 && x >= 7 && x <= 12 {
				result.Belly[y][x] = true
			}

			// Texture speckle
			if !result.Highlight[y][x] && !result.Shadow[y][x] && !result.Accent[y][x] && !result.Mouth[y][x] {
				if texRng.Intn(100) < 20 {
					result.Texture[y][x] = true
				}
			}
		}
	}

	// Biology-specific rendering pass
	switch m.BodySubtype {
	case SubtypeGaseous:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && texRng.Intn(100) < 40 {
					result.Body[y][x] = false
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeCrystalline:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Highlight[y][x] && !result.Shadow[y][x] {
					if texRng.Intn(100) < 15 {
						result.Accent[y][x] = true
					}
				}
			}
		}
	case SubtypeMechanical, SubtypeSilicon:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && texRng.Intn(100) < 10 {
					result.Highlight[y][x] = true
				}
			}
		}
	case SubtypeAmorphous:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && texRng.Intn(100) < 25 {
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeNanotech:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && texRng.Intn(100) < 30 {
					result.Highlight[y][x] = true
				}
				if result.Body[y][x] && texRng.Intn(100) < 15 {
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeBioSynthetic:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Highlight[y][x] && !result.Shadow[y][x] {
					if texRng.Intn(100) < 10 {
						result.Accent[y][x] = true
					}
				}
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
		return headRound
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
		return torsoSlim
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

// EyeTypeFromMorphology selects an eye style based on the alien's eyesight.
func EyeTypeFromMorphology(m *Morphology, rng *rand.Rand) EyeStyle {
	if m == nil {
		return EyeClassic
	}
	switch m.Eyesight {
	case "none", "blind":
		return EyeNone
	case "echolocation":
		return EyeVisor
	case "excellent":
		if rng.Intn(2) == 0 {
			return EyeArachnid
		}
		return EyeClassic
	default: // "normal" or anything else
		if rng.Intn(2) == 0 {
			return EyeCyclops
		}
		return EyeClassic
	}
}

func pickEyes(candidates []TaggedEyes, m *Morphology, rng *rand.Rand) []string {
	style := EyeTypeFromMorphology(m, rng)
	for _, e := range candidates {
		if e.Style == style {
			return e.Pixels
		}
	}
	return eyeClassic
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
