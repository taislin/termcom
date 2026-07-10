package soldier

import (
	"fmt"
	"math/rand"

	"github.com/civ13/ycom/internal/data"
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
	if int(r) < len(RankNames) {
		return RankNames[r]
	}
	return "Unknown"
}

type Soldier struct {
	Name      string
	Rank      Rank
	HP        int
	MaxHP     int
	TU        int
	MaxTU     int
	Accuracy  int
	Bravery   int
	Reactions int
	Strength  int
	PsiSkill  int
	PsiStr    int
	PosX      int
	PosY      int
	Weapon    string
	WeaponAmmo int // Current ammo for the weapon
	Armor     string
	Kills     int
	Missions  int
	Wounds    int // days until healed
}

func NewSoldier(name string) *Soldier {
	hp := 20 + rand.Intn(6)
	tu := 45 + rand.Intn(11)
	return &Soldier{
		Name:      name,
		Rank:      Rookie,
		HP:        hp,
		MaxHP:     hp,
		TU:        tu,
		MaxTU:     tu,
		Accuracy:  40 + rand.Intn(21),
		Bravery:   30 + rand.Intn(41),
		Reactions: 30 + rand.Intn(21),
		Strength:  10 + rand.Intn(11),
		PsiSkill:  0,
		PsiStr:    rand.Intn(40),
		Weapon:    "rifle",
		WeaponAmmo: data.RuleItems["rifle"].AmmoMax,
		Armor:     "none",
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

func (s *Soldier) GainXP(kills int) {
	s.Kills += kills
	xpThreshold := []int{0, 10, 25, 50, 80, 120, 170, 230}
	for int(s.Rank) < len(xpThreshold)-1 && s.Kills >= xpThreshold[int(s.Rank)+1] {
		s.Rank++
		s.MaxHP += 2
		s.HP += 2
		s.MaxTU += 1
		s.TU += 1
		s.Accuracy += 2
		s.Strength += 1
		s.Reactions += 1
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
	return fmt.Sprintf("%-12s %s  HP:%d/%d TU:%d ACC:%d BRA:%d STR:%d W:%s A:%s Kills:%d",
		s.Name, s.Rank, s.HP, s.MaxHP, s.TU, s.Accuracy, s.Bravery, s.Strength,
		data.RuleItems[s.Weapon].ShortName, data.Armors[s.Armor].ShortName, s.Kills)
}
