package soldier

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
)

type Rank int

const (
	Rookie Rank = iota
	Squaddie
	Corporal
	Sergeant
	Lieutenant
	Captain
	Major
	Colonel
)

var RankNames = []string{
	"Rookie", "Squaddie", "Corporal", "Sergeant",
	"Lieutenant", "Captain", "Major", "Colonel",
}

func (r Rank) String() string {
	switch r {
	case Rookie:
		return language.String("RANK_ROOKIE")
	case Squaddie:
		return language.String("RANK_SQUADDIE")
	case Corporal:
		return language.String("RANK_CORPORAL")
	case Sergeant:
		return language.String("RANK_SERGEANT")
	case Lieutenant:
		return language.String("RANK_LIEUTENANT")
	case Captain:
		return language.String("RANK_CAPTAIN")
	case Major:
		return language.String("RANK_MAJOR")
	case Colonel:
		return language.String("RANK_COLONEL")
	default:
		return language.String("UNKNOWN")
	}
}

type Soldier struct {
	Name       string
	Rank       Rank
	HP         int
	MaxHP      int
	TU         int
	MaxTU      int
	Accuracy   int
	Bravery    int
	Reactions  int
	Strength   int
	PsiSkill   int
	PsiStr     int
	PosX       int
	PosY       int
	Weapon     string
	WeaponAmmo int // Current ammo for the weapon
	Armor      string
	Kills      int
	Missions   int
	Wounds     int // days until healed
	Fatigue    int // days until rested
	Perks      []string
	Inventory  []string // additional carried items

	// Transient per-mission XP counters (reset by PostMission, not persisted)
	ExpFiring    int
	ExpThrowing  int
	ExpReactions int
	ExpBravery   int
	ExpPsiSkill  int
	ExpPsiStr    int
	ExpMelee     int
	GainedXP     bool
}

var StatCaps = struct {
	TU, HP, Acc, React, Brave, Str, Psi, Melee int
}{
	TU:    80,
	HP:    60,
	Acc:   120,
	React: 100,
	Brave: 100,
	Str:   70,
	Psi:   100,
	Melee: 120,
}

// RankOpenings: total roster size required before a rank (index) may exist.
// Ranks 0 (Rookie) and 1 (Squaddie) are unlimited; Corporal opens at 4, etc.
var RankOpenings = []int{0, 0, 4, 8, 14, 22, 30, 40}

func (s *Soldier) CanDeploy() bool {
	return s.HP > 0 && s.Wounds == 0 && s.Fatigue == 0
}

func (s *Soldier) Encumbrance() int {
	w := 0
	if item, ok := data.RuleItems[s.Weapon]; ok {
		w += item.Weight
	}
	for _, item := range s.Inventory {
		if it, ok := data.RuleItems[item]; ok {
			w += it.Weight
		}
	}
	return w
}

func (s *Soldier) TUPenalty() int {
	return s.Encumbrance() / 5
}

func (s *Soldier) AddItem(item string) {
	s.Inventory = append(s.Inventory, item)
}

func (s *Soldier) RemoveItem(item string) bool {
	for i, it := range s.Inventory {
		if it == item {
			s.Inventory = append(s.Inventory[:i], s.Inventory[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Soldier) HasItem(item string) bool {
	for _, it := range s.Inventory {
		if it == item {
			return true
		}
	}
	return false
}

func (s *Soldier) CountItem(item string) int {
	n := 0
	for _, it := range s.Inventory {
		if it == item {
			n++
		}
	}
	return n
}

func NewSoldier(name string) *Soldier {
	hp := 20 + rand.Intn(6)    // 20..25
	tu := 45 + rand.Intn(11)   // 45..55
	return &Soldier{
		Name:       name,
		Rank:       Rookie,
		HP:         hp,
		MaxHP:      hp,
		TU:         tu,
		MaxTU:      tu,
		Accuracy:   40 + rand.Intn(21), // 40..60
		Bravery:    30 + rand.Intn(41), // 30..70
		Reactions:  30 + rand.Intn(21), // 30..50
		Strength:   10 + rand.Intn(11), // 10..20
		PsiSkill:   0,
		PsiStr:     rand.Intn(40), // 0..39
		Weapon:     "rifle",
		WeaponAmmo: data.RuleItems["rifle"].AmmoMax,
		Armor:      "none",
	}
}

func (s *Soldier) FireWeapon(target *Soldier) (int, bool) {
	w := data.RuleItems[s.Weapon]
	if s.WeaponAmmo <= 0 {
		return 0, false
	}
	s.WeaponAmmo--

	// Accuracy roll
	hit := rand.Intn(100) < s.Accuracy
	if !hit {
		return 0, false
	}

	dmg := w.Damage + rand.Intn(w.Damage/2+1)
	// Armour reduction
	tgtArmor := data.Armors[target.Armor]
	dmg -= tgtArmor.Undersuit
	if dmg < 1 {
		dmg = 1
	}
	target.HP -= dmg
	return dmg, true
}

func (s *Soldier) AddFiringExp() {
	s.ExpFiring++
	s.GainedXP = true
}

func (s *Soldier) AddThrowingExp() {
	s.ExpThrowing++
	s.GainedXP = true
}

func (s *Soldier) AddReactionsExp() {
	s.ExpReactions++
	s.GainedXP = true
}

func (s *Soldier) AddBraveryExp() {
	s.ExpBravery++
	s.GainedXP = true
}

func (s *Soldier) AddPsiSkillExp() {
	s.ExpPsiSkill++
	s.GainedXP = true
}

func (s *Soldier) AddPsiStrExp() {
	s.ExpPsiStr++
	s.GainedXP = true
}

func (s *Soldier) AddMeleeExp() {
	s.ExpMelee++
	s.GainedXP = true
}

// improveStat returns a stat gain based on experience points earned in a mission.
// Thresholds: >10→2+d4, >5→1+d4, >2→1+d3, >0→d2.
func improveStat(exp int) int {
	switch {
	case exp > 10:
		return 2 + rand.Intn(5)
	case exp > 5:
		return 1 + rand.Intn(4)
	case exp > 2:
		return 1 + rand.Intn(3)
	case exp > 0:
		return rand.Intn(2)
	default:
		return 0
	}
}

func calcStatGain(exp int, xpMult float64, current, cap int) int {
	if exp <= 0 || current >= cap {
		return 0
	}
	gain := int(float64(improveStat(exp)) * xpMult)
	if current+gain > cap {
		return cap - current
	}
	return gain
}

func (s *Soldier) PostMission() {
	xpMult := 1.0
	if s.HasPerk("quick_learner") {
		xpMult = 1.5
	}

	s.Accuracy += calcStatGain(s.ExpFiring, xpMult, s.Accuracy, StatCaps.Acc)
	s.Strength += calcStatGain(s.ExpThrowing, xpMult, s.Strength, StatCaps.Str)
	s.Strength += calcStatGain(s.ExpMelee, xpMult, s.Strength, StatCaps.Str)
	s.Reactions += calcStatGain(s.ExpReactions, xpMult, s.Reactions, StatCaps.React)
	s.PsiSkill += calcStatGain(s.ExpPsiSkill, xpMult, s.PsiSkill, StatCaps.Psi)
	s.PsiStr += calcStatGain(s.ExpPsiStr, xpMult, s.PsiStr, StatCaps.Psi)
	// Bravery gain: 50% chance, +10 per trigger
	if s.ExpBravery > rand.Intn(11) && s.Bravery < 100 {
		s.Bravery += int(10 * xpMult)
		if s.Bravery > 100 {
			s.Bravery = 100
		}
	}
	if s.GainedXP {
		if s.Rank == Rookie {
			s.Rank = Squaddie
		}
		s.TU += calcStatGain(rand.Intn((StatCaps.TU-s.TU)/10+2), xpMult, s.TU, StatCaps.TU)
		s.HP += calcStatGain(rand.Intn((StatCaps.HP-s.HP)/10+2), xpMult, s.HP, StatCaps.HP)
		s.Strength += calcStatGain(rand.Intn((StatCaps.Str-s.Strength)/10+2), xpMult, s.Strength, StatCaps.Str)
	}
	s.ExpFiring = 0
	s.ExpThrowing = 0
	s.ExpReactions = 0
	s.ExpBravery = 0
	s.ExpPsiSkill = 0
	s.ExpPsiStr = 0
	s.ExpMelee = 0
	s.GainedXP = false
}

// HandlePromotions promotes soldiers up an 8-rank ladder based on total
// roster size. Only one rank is gained per eligible soldier per call, and
// each promotion rolls a perk (preserving rank-on-perk behaviour).
func HandlePromotions(roster []*Soldier) {
	total := len(roster)
	if total == 0 {
		return
	}
	maxRank := 0
	for r := len(RankOpenings) - 1; r >= 0; r-- {
		if total >= RankOpenings[r] {
			maxRank = r
			break
		}
	}
	if maxRank < 1 {
		return
	}
	// Seed from the roster's actual composition plus a time-varying component so
	// that equal-sized rosters (and repeated calls) yield different, per-roster
	// promotion sequences instead of a fixed one.
	var seed int64
	for _, s := range roster {
		seed = seed*31 + int64(len(s.Name)) + int64(s.Rank)*131 + int64(s.Missions)*17 + int64(s.Kills)*7
	}
	seed ^= time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	for target := 1; target <= maxRank; target++ {
		capCount := RankOpenings[target]
		atTarget := 0
		for _, s := range roster {
			if int(s.Rank) == target {
				atTarget++
			}
		}
		openings := capCount - atTarget
		if openings <= 0 {
			continue
		}
		var cands []*Soldier
		for _, s := range roster {
			if int(s.Rank) == target-1 {
				cands = append(cands, s)
			}
		}
		sort.Slice(cands, func(i, j int) bool {
			if cands[i].Kills != cands[j].Kills {
				return cands[i].Kills > cands[j].Kills
			}
			return cands[i].Missions > cands[j].Missions
		})
		for i := 0; i < openings && i < len(cands); i++ {
			c := cands[i]
			c.Rank++
			if perk := RollPerk(rng, c.Perks); perk != nil {
				c.ApplyPerk(*perk)
			}
		}
	}
}

type Squad []*Soldier

func (sq Squad) Alive() []*Soldier {
	var alive []*Soldier
	for _, s := range sq {
		if s.HP > 0 {
			alive = append(alive, s)
		}
	}
	return alive
}

func (sq Squad) AllDead() bool {
	return len(sq.Alive()) == 0
}

var Names = []string{
	"Abbot", "Adams", "Adder", "Akerman", "Allard", "Allen", "Alvarez",
	"Anderson", "Andre", "Angel", "Archer", "Armstrong", "Arnold", "Ash",
	"Atkinson", "Austin", "Avery", "Bailey", "Baker", "Baldwin", "Ball",
	"Banks", "Barker", "Barlow", "Barnes", "Barrett", "Barton", "Bass",
	"Bates", "Batman", "Baxter", "Beach", "Bell", "Bennett", "Berry",
	"Bishop", "Black", "Blake", "Bond", "Booth", "Bowen", "Boyd",
	"Brady", "Branch", "Brennan", "Brooks", "Brown", "Bryant", "Buckland",
	"Burgess", "Burke", "Burns", "Burton", "Butler", "Byrne", "Cameron",
	"Campbell", "Cannon", "Carter", "Casey", "Chambers", "Chapman", "Clark",
	"Clarke", "Cleveland", "Clifford", "Cline", "Cobb", "Cohen", "Cole",
	"Coleman", "Collins", "Colt", "Connor", "Cook", "Cooper", "Cox",
	"Crane", "Crawford", "Cross", "Cruz", "Curtis", "Cyrus", "Dale",
	"Dallas", "Dalton", "Damon", "Dane", "Daniels", "Davis", "Dawson",
	"Day", "Dean", "Decker", "Delgado", "Dempsey", "Dennis", "Diaz",
	"Dickson", "Donovan", "Doors", "Doran", "Dunn", "Dyer", "Eames",
	"Edwards", "Eggins", "Eliot", "Ellis", "Elwood", "Emerson", "Evans",
	"Falcon", "Farrow", "Ferguson", "Finn", "Fischer", "Fisher", "Fitzgerald",
	"Flynn", "Foley", "Ford", "Foreman", "Forman", "Forrest", "Foster",
	"Fox", "Franks", "Frazier", "Fuller", "Gable", "Gallagher", "Garcia",
	"Garfield", "Garner", "Garrison", "Gates", "Gay", "George", "Gibson",
	"Gilbert", "Gill", "Glover", "Goddard", "Gold", "Goodwin", "Grant",
	"Graves", "Gray", "Green", "Griffin", "Griffiths", "Grimes", "Gross",
	"Guest", "Guile", "Hague", "Hall", "Hamilton", "Hammond", "Hampton",
	"Hancock", "Hanover", "Hansen", "Harding", "Harper", "Harris", "Harrison",
	"Hart", "Hastings", "Hawkins", "Hayden", "Hayes", "Haywood", "Heath",
	"Henderson", "Henry", "Herring", "Hicks", "Higgins", "Hill", "Hobbs",
	"Holden", "Holland", "Holt", "Homer", "Hoover", "Hopkins", "Horn",
	"Horton", "Howard", "Howe", "Howell", "Hubbard", "Hudson", "Hughes",
	"Hunt", "Hunter", "Hurley", "Hyde", "Irons", "Irving", "Jackson",
	"James", "Jenkins", "Jennings", "Jensen", "Jeter", "Johnson", "Johnston",
	"Jones", "Jordan", "Joy", "Joyce", "Judah", "Kane", "Kaufman",
	"Keane", "Keating", "Keller", "Kelly", "Kenyon", "Kerr", "Kim",
	"King", "Kirk", "Knight", "Lambert", "Lane", "Lang", "Larson",
	"Lawrence", "Lawson", "Leach", "Lee", "Leon", "Lewis", "Lilly",
	"Lindley", "Lloyd", "Locke", "Long", "Lopez", "Loren", "Love",
	"Lynch", "Lyons", "Maddox", "Magnus", "Malone", "Manning", "Marquez",
	"Marsh", "Marshall", "Martin", "Mason", "Masters", "Mathis", "Maxwell",
	"May", "McBride", "McCain", "McCarthy", "McClure", "McConnell", "McCoy",
	"McCullough", "McDowell", "McGee", "McIntosh", "McKay", "McLean", "McMahon",
	"McMillan", "McNamara", "Mead", "Mendez", "Mercer", "Merrill", "Metcalf",
	"Meyer", "Miles", "Miller", "Mitchell", "Morgan", "Morris", "Morrow",
	"Morton", "Mosley", "Mueller", "Murphy", "Murray", "Myers", "Nash",
	"Neal", "Nelson", "Newman", "Newton", "Nichols", "Nickols", "Norris",
	"Norton", "Nurse", "O'Brien", "O'Connor", "O'Neill", "Oates", "Obrian",
	"O'Casey", "O'Connor", "O'Donnell", "O'Hara", "O'Neil", "O'Neill", "O'Reilly",
	"O'Rourke", "O'Toole", "Oakley", "Ogle", "Oliver", "Osborne", "Otis",
	"Pace", "Palmer", "Parker", "Parrish", "Parsons", "Patterson", "Patton",
	"Pearce", "Pearson", "Peck", "Perry", "Peters", "Peterson", "Phillips",
	"Pierce", "Pike", "Pitts", "Polk", "Porter", "Potter", "Powell",
	"Powers", "Pratt", "Price", "Prince", "Pullman", "Quinn", "Rader",
	"Ramirez", "Randolph", "Ranger", "Ransom", "Ray", "Raymond", "Reed",
	"Reese", "Reeves", "Reid", "Reilly", "Reynolds", "Rhodes", "Rice",
	"Rich", "Richards", "Richardson", "Richmond", "Riggs", "Riley", "Ringo",
	"Rivas", "Rivera", "Roach", "Roberts", "Robertson", "Robinson", "Rock",
	"Rodgers", "Rogers", "Rollins", "Romero", "Roosevelt", "Root", "Rose",
	"Ross", "Rousseau", "Rowe", "Rubin", "Rush", "Russo", "Ruth",
	"Ryan", "Salazar", "Sanders", "Sargent", "Saunders", "Schmidt", "Schneider",
	"Schroeder", "Scott", "Sears", "Selby", "Shaw", "Shelton", "Shepard",
	"Shepherd", "Sherman", "Shields", "Short", "Shultz", "Simmons", "Simon",
	"Simpson", "Sims", "Singer", "Slater", "Sloan", "Small", "Smart",
	"Smith", "Snyder", "Solis", "Sosa", "Sparks", "Spector", "Spencer",
	"Stacy", "Stanley", "Stanton", "Stark", "Starr", "Steele", "Stephens",
	"Stephenson", "Sterling", "Stevens", "Stevenson", "Stewart", "Stills", "Stokes",
	"Stone", "Storm", "Strange", "Stratton", "Stubbs", "Sullivan", "Summers",
	"Swan", "Sweeney", "Sweet", "Sykes", "Talbot", "Tate", "Taylor",
	"Teague", "Terry", "Thatcher", "Thomas", "Thompson", "Thornton", "Tiffany",
	"Tilden", "Timms", "Tobin", "Todd", "Torres", "Townsend", "Travis",
	"Turner", "Tyler", "Underwood", "Upshaw", "Vance", "Vega", "Vogel",
	"Wade", "Wagner", "Walker", "Wall", "Wallace", "Walls", "Walters",
	"Walton", "Ward", "Warner", "Warren", "Washington", "Waters", "Watkins",
	"Watson", "Watts", "Weaver", "Webb", "Weber", "Webster", "Wells",
	"Welsh", "West", "Wheeler", "Whelan", "Whitaker", "White", "Whitesell",
	"Wickham", "Wilder", "Wilkins", "Wilkinson", "Williams", "Williamson", "Willis",
	"Wills", "Wilson", "Wilton", "Windsor", "Winter", "Withers", "Wolfe",
	"Wong", "Wood", "Woods", "Woodward", "Wright", "Wyatt", "York",
	"Young", "Zimmerman",
}

func RandomName() string {
	return Names[rand.Intn(len(Names))]
}

func FormatSoldier(s *Soldier) string {
	return fmt.Sprintf(language.String("SOLDIER_FORMAT"),
		s.Name, s.Rank, s.HP, s.MaxHP, s.TU, s.Accuracy, s.Bravery, s.Strength,
		data.RuleItems[s.Weapon].ShortName, data.Armors[s.Armor].ShortName, s.Kills)
}
