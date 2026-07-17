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
	LocomSlither
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
	Pixels []string
	// Manip lists every manipulator category this torso is valid for. A torso
	// may be shared across categories (e.g. an armless alien can wear a slim or
	// hollow torso).
	Manip []Manipulators
	// BodyType restricts the torso to a body subtype (e.g. "gaseous"). Empty
	// means it applies to carbon-flesh / silicon / bio_synthetic bodies. Each
	// exotic subtype (gaseous, amorphous, mechanical, nanotech, crystalline)
	// has dedicated core templates.
	BodyType string
}

type TaggedLegs struct {
	Pixels     []string
	Locomotion Locomotion
}

type TaggedWeapon struct {
	Pixels     []string
	DamageType int // DMG_* constant this design represents (-1 = generic)
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
	Heads   []TaggedHead
	Eyes    []TaggedEyes
	Torsos  []TaggedTorso
	Legs    []TaggedLegs
	Weapons []TaggedWeapon
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
				{Pixels: torsoSlim, Manip: []Manipulators{ManipBipedal, ManipNone}},
				{Pixels: torsoWide, Manip: []Manipulators{ManipBipedal, ManipNone}},
				{Pixels: torsoArmored, Manip: []Manipulators{ManipBipedal, ManipNone}},
				{Pixels: torsoHollow, Manip: []Manipulators{ManipBipedal, ManipNone}},
				{Pixels: torsoAsymmetric, Manip: []Manipulators{ManipBipedal, ManipNone}},
				{Pixels: torsoBladed, Manip: []Manipulators{ManipBipedal}},
				{Pixels: torsoTentacle, Manip: []Manipulators{ManipNone}},
				{Pixels: torsoFloating, Manip: []Manipulators{ManipNone}},
				{Pixels: torsoMultiArmed, Manip: []Manipulators{ManipMultiArmed}},
				{Pixels: torsoMultiArmed2, Manip: []Manipulators{ManipMultiArmed}},
				{Pixels: torsoGaseous, Manip: []Manipulators{ManipBipedal, ManipMultiArmed, ManipNone}, BodyType: SubtypeGaseous},
				{Pixels: torsoAmorphous, Manip: []Manipulators{ManipBipedal, ManipMultiArmed, ManipNone}, BodyType: SubtypeAmorphous},
				{Pixels: torsoMechanical, Manip: []Manipulators{ManipBipedal, ManipMultiArmed, ManipNone}, BodyType: SubtypeMechanical},
				{Pixels: torsoNanotech, Manip: []Manipulators{ManipBipedal, ManipMultiArmed, ManipNone}, BodyType: SubtypeNanotech},
				{Pixels: torsoCrystalline, Manip: []Manipulators{ManipBipedal, ManipMultiArmed, ManipNone}, BodyType: SubtypeCrystalline},
			},
			Legs: []TaggedLegs{
				{Pixels: legsBipedal, Locomotion: LocomBipedal},
				{Pixels: legsLong, Locomotion: LocomBipedal},
				{Pixels: legsWide, Locomotion: LocomBipedal},
				{Pixels: legsArachnid, Locomotion: LocomArachnid},
				{Pixels: legsCrab, Locomotion: LocomArachnid},
				{Pixels: legsSerpentine, Locomotion: LocomSlither},
				{Pixels: legsTentacle, Locomotion: LocomFloating},
				{Pixels: legsPillar, Locomotion: LocomFloating},
				{Pixels: legsFloating, Locomotion: LocomFloating},
			},
			Weapons: []TaggedWeapon{
				{Pixels: weaponKinetic, DamageType: DMG_KINETIC},
				{Pixels: weaponPlasma, DamageType: DMG_PLASMA},
				{Pixels: weaponLaser, DamageType: DMG_LASER},
				{Pixels: weaponExplosive, DamageType: DMG_EXPLOSIVE},
				{Pixels: weaponMelee, DamageType: DMG_MELEE},
				{Pixels: weaponPsionic, DamageType: DMG_PSIONIC},
			},
		}
	}
	return defaultSpriteRegistry
}

// --- Head templates (10 rows x 20 wide) ---
// Detail chars (see render loop at headOffset above):
//   X = body (solid head/skull)
//   m = body + Mouth layer (maw; drawn in dark mouth color)
//   a = body + Accent layer (antenna / bright marking)
//   h = body + Highlight layer (bright edge glint)
//   d = body + Shadow layer (dark recessed detail, e.g. eye sockets)
// Eyes are NOT a head char — they come from the separate eye mask below
// and are carved into the body afterward.

var headRound = []string{
	"....................",
	".......XXXXXX.......",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	"....XXXmmmmmmXXX....",
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
	"...XXmmmmmmmmmmXXX..",
	"...XXXXXXXXXXXXXXX..",
	"...XXXXXXXXXXXXXXX..",
	"....XXXXXXXXXXXX....",
}

var headTall = []string{
	".......XXXX.........",
	"......XXXXXX........",
	"......XXXXXX........",
	".....XXXXXXXX.......",
	".....XXXXXXXX.......",
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
	"....XXddXXXXddXX....",
	"....XddXXddXXddX....",
	"....XXXXXXXXXXXX....",
	"....XXXddXdddXXX....",
	".....XXXXddXXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headWide = []string{
	"....................",
	"aa................aa",
	"hhh..............hhh",
	"XXXh............hXXX",
	"XXXXXXXXXXXXXXXXXXXX",
	".XXXXXXXXXXXXXXXXXX.",
	".XXXXXXXXXXXXXXXXXX.",
	".XXXXXXXXXXXXXXXXXX.",
	".dXXXXXmmmmmXXXXXXd.",
	"..dXXXXXXXXXXXXXd...",
}

var headAntennae = []string{
	"..a.............a...",
	"..aa..........aa....",
	"...aa........aa.....",
	"...hhhXXXXXXhhh.....",
	"....XXXXXXXXXX......",
	"....XXXXXXXXXX......",
	"....XXXmmmmXXX......",
	".....XXXXXXXX.......",
	"......XXXXXX........",
	".......XXXX.........",
}

var headVisor = []string{
	"....................",
	"......XXXXXXX.......",
	".....ddXXXXXXdd.....",
	"....dddXhhhhXddd....",
	"....dXXXhhhhXXXd....",
	"....dXXXXXXXXXXd....",
	"....XXXXXXXXXXX.....",
	".....XXXmmmmXXX.....",
	"......XXXXXXXX......",
	".......XXXXXX.......",
}

var headCone = []string{
	".........hh.........",
	"........hXXh........",
	".......hXXXXh.......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	".....XXXXXXXXXX.....",
	"....XXXXmmmmXXXX....",
	"....dXXXXXXXXXXd....",
	".....dXXXXXXXXd.....",
}

// --- Eye templates (10 rows x 20 wide) ---
// These act as masks overlaid on the head: each 'X' carves a hole in the
// head body and marks the Eyes layer (white/glow). Supported chars:
//   X = Eyes (carves body, marks eye)
//   h = Eyes + Highlight (glowing / visor rim)
//   a = Eyes + Accent (colored pupil)
//   d = Eyes + Shadow (recessed eye)

var eyeClassic = []string{
	"....................",
	"....................",
	"....................",
	".....XXX...XXX......",
	"....XaaXX.XaaXX....",
	"....XXaXX.XaXXX....",
	".....XXX...XXX......",
	"....................",
	"....................",
	"....................",
}

var eyeCyclops = []string{
	"....................",
	"....................",
	"....................",
	".......XXXXX........",
	"......XXaaaXX.......",
	".......XXXXX........",
	"....................",
	"....................",
	"....................",
	"....................",
}

var eyeArachnid = []string{
	"....................",
	"......XXXd.dXXX.....",
	".......dd...dd......",
	".....XXdXXdXXdXX....",
	"........dd.dd.......",
	".....XXdXXdXXdXX....",
	"........dd.dd.......",
	"........XXdXX.......",
	".........d..d.......",
	"....................",
}

var eyeVisor = []string{
	"....................",
	"....................",
	"....................",
	"........hhhh........",
	"....XXXhaaaahXXX....",
	"........hhhh........",
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

// --- Weapon mask templates (8 rows x 20 wide, drawn over torso rows 10-17) ---
// These define the held weapon silhouette per damage type. 'X' = weapon body,
// 'h' = highlight, 'a' = accent, 'd' = dark. Positioned on the right side
// where the torso 'W' grip marker sits.
//
// DMG_KINETIC: a rail/slug rifle with a long barrel.
var weaponKinetic = []string{
	"....................",
	"....................",
	"...............h....",
	"...XXXXXXXXXXXXXXh..",
	".ddXXXXXXXXXXXXXXhh.",
	".dddX.......XX......",
	"....................",
	"....................",
}

// DMG_PLASMA: a bulbous plasma projector with a glowing core.
var weaponPlasma = []string{
	"....................",
	"....................",
	"....................",
	"........XXXXXXXhhh..",
	".......XXXXXXXXhhhd.",
	"......dXXXdddXXX....",
	"......dd......dd....",
	"......dd......dd....",
}

// DMG_LASER: a slim beam emitter with a barrel lens.
var weaponLaser = []string{
	"....................",
	"....................",
	"...............hhhhd",
	".......XXXXXXXXXXXd.",
	".......XXXXXXXXXXXd.",
	"......XXX...........",
	"......dd............",
	"....................",
}

// DMG_EXPLOSIVE: a bulky launcher with a wide muzzle.
var weaponExplosive = []string{
	"....................",
	"............XdddddX.",
	"............XXXXXXXX",
	"...........XXXXXXXXX",
	"............XX...XX.",
	"...........XX.....XX",
	"....................",
	"....................",
}

// DMG_MELEE: a bladed claw/edge held forward.
var weaponMelee = []string{
	"..................hh",
	".................XXh",
	"...............XXh..",
	".............XXXh...",
	"............XXXh....",
	"..........XXXh......",
	"........dXXXd.......",
	"......ddddd.........",
}

// DMG_PSIONIC: a hovering staff/orb with no solid barrel.
var weaponPsionic = []string{
	"....................",
	".............dXXXd..",
	".............XhahX..",
	".............XXXXX..",
	"..............XXX...",
	"..............XXX...",
	"..............dXd...",
	"...............d....",
}

var torsoSlim = []string{
	".......XXXXXX.......",
	".......XXXXXX.......",
	"......XXXXXXXX......",
	"......XXXXXXXXX.....",
	".......XXXXXXXX.....",
	".......XXXXXXXX.....",
	".......XXXXXX.......",
	".......XXXXXX.......",
}

// --- Torso templates (8 rows x 20 wide) ---
// Detail chars:
//   X = body (solid torso)
//   W = Weapon layer — only used by melee/claw torsos (e.g. torsoBladed)
//       as integral blades. Ranged weapons come from the separate weapon mask.
//   a = body + Accent layer (armor trim)
//   h = body + Highlight layer (polished plate edge)
//   d = body + Shadow layer (recessed seam)

var torsoWide = []string{
	".....XXXXXXXXXX.....",
	".....XXXXXXXXXX.....",
	"....XXXXXXXXXXXX....",
	"....XXXXXXXXXXXX....",
	".....XXXXXXXXXXX....",
	".....XXXXXXXXXXX....",
	".....XXXXXXXXXX.....",
	".....XXXXXXXXXX.....",
}

var torsoArmored = []string{
	".....dddXXXXdd......",
	"....dXXXXXXXXXXd....",
	"...XXdXXXXXXXXdXX...",
	"...XXdXXdXXdXXXX....",
	"..XXXXdXXdXXXXXXX...",
	"..XXXXdXXXXXXXXX....",
	"..XXXXdXXXXXXXXX....",
	"...XXXdXXXXdXXX.....",
}

var torsoTentacle = []string{
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"......XXXXXXXX......",
	"h......XXXXXX......h",
	"XX..XXXXXXXXXXXX..XX",
	".XXXX..XXXXXX..XXXX.",
	".......XXXXXX.......",
}

var torsoMultiArmed = []string{
	".XXXXXXXXXXXXXXXXXX.",
	".XXX..XXXXXXXX..XXX.",
	"..hh...XXXXXX...hh..",
	".......XXXXXX.......",
	".XXXXXXXXXXXXXXXXXX.",
	".XXX...XXXXXX...XXX.",
	"..hh..XaaaaaaX..hh..",
	".....XXXXXXXXXX.....",
}

var torsoMultiArmed2 = []string{
	"a..aa..XXXXXX..aa..a",
	"XX.XX..XXXXXX..XX.XX",
	".XXXX...XXXX...XXXX.",
	"...XXXXXXXXXXXXXX...",
	".a...XXXXXXXXXX...a.",
	".aXXXXXXXXXXXXXXXXa.",
	".a..XXXXXXXXXXX...a.",
	".....XXXXXXXXX......",
}

// torsoFloating is a compact hovering core with no legs/arms. Paired with
// LocomFloating aliens (Legs == 0) via legsFloating for a levitating look.
// It carries no weapon pixels (armless aliens are unarmed).
var torsoFloating = []string{
	".......dXXXXd.......",
	".....dXXXXXXXXd.....",
	"....XXXhhXXhhXXX....",
	"....XXhhXXXXhhXX....",
	"....XXXXaaaaXXXX....",
	".....XXaaaaaaXX.....",
	"......dXXXXXXd......",
	".......dhhhhd.......",
}

// Exotic body-subtype cores. Each silhouette is used only for its matching
// BodySubtype so e.g. gaseous aliens never get drawn as flesh torsos. They
// carry no weapon pixels; arms/legs are drawn as separate modules on top.

// torsoGaseous: a drifting cloud/vortex core.
var torsoGaseous = []string{
	".....X..X..X..X....",
	"....X.X.XXX.X.X....",
	"....X.X.XXX.X.X.X..",
	"..X.X.X.X.X.X.X.X..",
	"....X.X.XXX.X.X....",
	".....X.XXXXX.X.....",
	"......X.X.X.X.X....",
	".......X.X.X.X.....",
}

// torsoAmorphous: a shifting blob with no fixed shape.
var torsoAmorphous = []string{
	"....XX.XXXX.XX.....",
	"..XX..XXXXXX..XX...",
	".XXXXXXXXXXXXXXXX..",
	"..XXXXXXXXXXXXX....",
	"....XXXXXXXXXX.....",
	"..XXXXXXXXXXXXXX...",
	".XXXXXXXXXXXXXXXX..",
	"....XX.XXXX.XX.....",
}

// torsoMechanical: a rigid plated chassis with panel seams.
var torsoMechanical = []string{
	"..XXXXXXXXXXXXXXXX..",
	"..XhhWWWWWWWWWhhX..",
	"..XWWaaaaaaaaWWX...",
	"..XaaXXXXXXXXaaX...",
	"..XWWXXXXXXXXWWX...",
	"..XhhWWWWWWWWWhhX..",
	"..XaaXXXXXXXXaaX...",
	"..XXXXXXXXXXXXXXXX..",
}

// torsoNanotech: a swarm-of-cells core with scattered nodes.
var torsoNanotech = []string{
	".dXdXhXhXhXhXhXhX..",
	"X.X.XhX.X.XhXhXhXh.",
	".XhXhXhXhXhXhXhXhXh",
	"X.X.XaX.X.XaXhXhXhX",
	".XhXhXhXhXhXhX.XhX.",
	"X.XhXhXhXhXhXhXhXh.",
	".X.X.XaX.X.XaXhXhXh",
	"..XhXhXhXhXhXhXhX..",
}

// torsoCrystalline: a faceted gem-like lattice.
var torsoCrystalline = []string{
	".......X....X......",
	"......XXX..XXX.....",
	".....XXXXXaXXXXX...",
	"....XXXXXXXXXXXXX..",
	"....XXXXXhhXXXXX...",
	".....XXXXXaXXXXX...",
	"......XXX..XXX.....",
	".......X....X......",
}

var torsoHollow = []string{
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXX....XXXX.....",
	"....XX......XXX.....",
	"....XXX....XXXX.....",
	".....XXXXXXXXXX.....",
	"......XXXXXXXXXX....",
	"........XXXXXXXX....",
}

var torsoAsymmetric = []string{
	".....XXXXXXX........",
	"....XXXXXXXXXX......",
	"...XXXX.XX.XXXXX....",
	"..XXXXX.XX.XXXXXX...",
	"..XXXX.XX.XXXXXXX...",
	"...XX.XX.XXXXXXXXXX.",
	"......XXXXXXXXXX.....",
	"........XXXXX.......",
}

var torsoBladed = []string{
	"......XXXXXXXX......",
	".....XXXXXXXXXX.....",
	"....XXXX.XX.XXXX....",
	"..WWXXXX.XX.XXXXWW..",
	"..WWXXXXXXXXXXXXWW..",
	"......XX.XX.WWW.....",
	".......XX.XXWW......",
	".......XXXXX..WW....",
}

// --- Leg templates (6 rows x 20 wide) ---
// Detail chars (same semantics as torso):
//   X = body (solid limb)
//   W = Weapon layer (built-in leg weapon; rarely used)
//   a = body + Accent layer (limb trim)
//   h = body + Highlight layer (limb edge glint)
//   d = body + Shadow layer (recessed joint)

var legsBipedal = []string{
	".......XXXXXX.......",
	"......XXX..XXX......",
	"......XX....XX......",
	".....XX......XX.....",
	".....XX......XX.....",
	"....ddd......ddd....",
}

var legsLong = []string{
	".......XXXXXX.......",
	".......XX.XXX.......",
	".......XX..XX.......",
	".......XX..XX.......",
	".......XX..XX.......",
	".......dd..dd.......",
}

var legsWide = []string{
	".......XXXXXX.......",
	"......XXXXXXXX......",
	".....XX......XX.....",
	"....XX........XX....",
	"...XX..........XX...",
	"..dd............dd..",
}

var legsArachnid = []string{
	".......XXXXXX.......",
	"......XX.XX.XX......",
	".....XX..XX..XX.....",
	"....XX...XX...XX....",
	"...XX....XX....XX...",
	"..dd.....dd.....dd..",
}

var legsTentacle = []string{
	"......XXXXXX........",
	"......XX..XX........",
	"......XX..XX........",
	"hh...XX...XX...hh...",
	".XX.XX.....XX.XX....",
	".XXXX.......XXX.....",
}

var legsPillar = []string{
	".......hXXXXh.......",
	"........XXXX........",
	"........XXXX........",
	"........XXXX........",
	"........hXXh........",
	".......dddddd.......",
}

var legsFloating = []string{
	".....XXXXXXXXXX.....",
	".......hXXXXh.......",
	"........haah........",
	"....................",
	"....................",
	"....................",
}

var legsSerpentine = []string{
	".........XXXXXXXX...",
	".........XXXXXXXX...",
	"........XXXXXXXX....",
	"......XXXXXXXXX.....",
	"..ddXXXXXXdd........",
	"dddddddddd..........",
}

var legsCrab = []string{
	"hh...XXXXXXXX...hh..",
	".XX.XXXXXXXXXX.XX...",
	"..XXX.XXXXXX.XXX....",
	".....XXXXXXXX.......",
	"...XX........XX.....",
	"..hh..........hh....",
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
		// 0 legs splits into two cases:
		//  - slithering aliens (flesh/silicon) crawl on a tail/coil
		//  - floaters (gas/amorphous/nanotech, or hovering mechanical)
		//    levitate with no ground contact.
		switch m.BodySubtype {
		case SubtypeGaseous, SubtypeAmorphous, SubtypeNanotech, SubtypeMechanical:
			return LocomFloating
		default:
			return LocomSlither
		}
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
			Eyesight: "normal", Hearing: "normal", DamageType: DMG_KINETIC,
		}
	}

	reg := NewAlienSpriteRegistry()
	rng := rand.New(rand.NewSource(seed))

	sense := SenseFromMorphology(m)
	manip := ManipulatorsFromMorphology(m)
	loco := LocomotionFromMorphology(m)

	head := pickHead(reg.Heads, sense, rng)
	eyes := pickEyes(reg.Eyes, m, rng)
	torso := pickTorso(reg.Torsos, manip, m.BodySubtype, rng)
	legs := pickLegs(reg.Legs, loco, rng)
	weapon := pickWeapon(reg.Weapons, m.DamageType, rng)

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
				if eyes[y][x] == '.' {
					continue
				}
			ex := x + eyeOffset
			ey := y
			if ex < 0 || ex >= 20 || ey < 0 || ey >= 10 {
				continue
			}
			result.Body[ey][ex] = false
			result.Eyes[ey][ex] = true
			switch eyes[y][x] {
			case 'h':
				result.Highlight[ey][ex] = true
			case 'a':
				result.Accent[ey][ex] = true
			case 'd':
				result.Shadow[ey][ex] = true
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
				if manip != ManipNone {
					result.Weapon[ty][x+torsoOffset] = true
				}
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

	// Weapon mask: a separate silhouette drawn over the torso's right side.
	// Only armed aliens (manip != none) carry a weapon.
	if manip != ManipNone {
		for y, row := range weapon {
			ty := 10 + y
			if ty >= 18 || y >= len(weapon) {
				break
			}
			for x, ch := range row {
				switch ch {
				case 'X':
					result.Weapon[ty][x+torsoOffset] = true
				case 'h':
					result.Weapon[ty][x+torsoOffset] = true
					result.Highlight[ty][x+torsoOffset] = true
				case 'a':
					result.Weapon[ty][x+torsoOffset] = true
					result.Accent[ty][x+torsoOffset] = true
				case 'd':
					result.Weapon[ty][x+torsoOffset] = true
					result.Shadow[ty][x+torsoOffset] = true
				}
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
			lx := x + legsOffset
			if lx < 0 || lx >= 20 {
				continue
			}
			switch ch {
			case 'X':
				result.Body[ly][lx] = true
			case 'W':
				result.Weapon[ly][lx] = true
			case 'a':
				result.Body[ly][lx] = true
				result.Accent[ly][lx] = true
			case 'h':
				result.Body[ly][lx] = true
				result.Highlight[ly][lx] = true
			case 'd':
				result.Body[ly][lx] = true
				result.Shadow[ly][lx] = true
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
			// Skip weapon pixels — they carry their own highlight/shadow from
			// the weapon mask and must not be overwritten by body shading.
			if result.Weapon[y][x] {
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
			if y >= 11 && y <= 16 && x >= 7 && x <= 12 && !result.Weapon[y][x] {
				result.Belly[y][x] = true
			}

			// Texture speckle
			if !result.Highlight[y][x] && !result.Shadow[y][x] && !result.Accent[y][x] && !result.Mouth[y][x] && !result.Weapon[y][x] {
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
				if result.Body[y][x] && !result.Weapon[y][x] && texRng.Intn(100) < 40 {
					result.Body[y][x] = false
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeCrystalline:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Weapon[y][x] && !result.Highlight[y][x] && !result.Shadow[y][x] {
					if texRng.Intn(100) < 15 {
						result.Accent[y][x] = true
					}
				}
			}
		}
	case SubtypeMechanical, SubtypeSilicon:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Weapon[y][x] && texRng.Intn(100) < 10 {
					result.Highlight[y][x] = true
				}
			}
		}
	case SubtypeAmorphous:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Weapon[y][x] && texRng.Intn(100) < 25 {
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeNanotech:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Weapon[y][x] && texRng.Intn(100) < 30 {
					result.Highlight[y][x] = true
				}
				if result.Body[y][x] && !result.Weapon[y][x] && texRng.Intn(100) < 15 {
					result.Texture[y][x] = true
				}
			}
		}
	case SubtypeBioSynthetic:
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if result.Body[y][x] && !result.Weapon[y][x] && !result.Highlight[y][x] && !result.Shadow[y][x] {
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

func pickTorso(candidates []TaggedTorso, manip Manipulators, subtype string, rng *rand.Rand) []string {
	var generic [][]string
	var specific [][]string
	for _, t := range candidates {
		okManip := false
		for _, mm := range t.Manip {
			if mm == manip {
				okManip = true
				break
			}
		}
		if !okManip {
			continue
		}
		if t.BodyType == "" {
			generic = append(generic, t.Pixels)
		} else if t.BodyType == subtype {
			specific = append(specific, t.Pixels)
		}
	}
	pool := generic
	if len(specific) > 0 {
		pool = specific
	}
	if len(pool) == 0 {
		return torsoSlim
	}
	return pool[rng.Intn(len(pool))]
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

// pickWeapon selects a weapon-mask design by damage type, falling back to the
// kinetic design when no specific match exists.
func pickWeapon(candidates []TaggedWeapon, dmgType int, rng *rand.Rand) []string {
	var specific, generic [][]string
	for _, w := range candidates {
		if w.DamageType == dmgType {
			specific = append(specific, w.Pixels)
		}
		if w.DamageType == DMG_KINETIC {
			generic = append(generic, w.Pixels)
		}
	}
	pool := generic
	if len(specific) > 0 {
		pool = specific
	}
	if len(pool) == 0 {
		return weaponKinetic
	}
	return pool[rng.Intn(len(pool))]
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
	return 120, 130, 140
}
