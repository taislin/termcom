package data

type AlienType struct {
	Name       string
	ShortName  string
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
}

var AlienTypes = []AlienType{
	// Rank 0 - Low tier
	{
		Name: "Sectoid", ShortName: "SEC",
		HP: 10, TU: 50, Accuracy: 55, Bravery: 40, Reactions: 50,
		Strength: 8, Psi: 40, Armour: 5, Weapon: "plasma_pistol",
		Points: 5, Rank: 0, Aggression: 3,
	},
	{
		Name: "Sectoid Navigator", ShortName: "SEN",
		HP: 11, TU: 52, Accuracy: 58, Bravery: 45, Reactions: 52,
		Strength: 8, Psi: 50, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 3,
	},
	{
		Name: "Sectoid Commander", ShortName: "SEC2",
		HP: 14, TU: 55, Accuracy: 62, Bravery: 55, Reactions: 58,
		Strength: 9, Psi: 70, Armour: 8, Weapon: "plasma_rifle",
		Points: 15, Rank: 2, Aggression: 4,
	},
	// Rank 1 - Mid tier
	{
		Name: "Floater", ShortName: "FLT",
		HP: 15, TU: 55, Accuracy: 60, Bravery: 50, Reactions: 60,
		Strength: 12, Psi: 10, Armour: 10, Weapon: "plasma_rifle",
		Points: 8, Rank: 1, Aggression: 6,
	},
	{
		Name: "Floater Navigator", ShortName: "FLN",
		HP: 16, TU: 58, Accuracy: 63, Bravery: 55, Reactions: 63,
		Strength: 13, Psi: 18, Armour: 11, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 6,
	},
	{
		Name: "Floater Commander", ShortName: "FLC",
		HP: 20, TU: 62, Accuracy: 68, Bravery: 65, Reactions: 68,
		Strength: 15, Psi: 30, Armour: 14, Weapon: "plasma_rifle",
		Points: 20, Rank: 3, Aggression: 7,
	},
	{
		Name: "Chryssalid", ShortName: "CHR",
		HP: 14, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 75,
		Strength: 18, Psi: 0, Armour: 8, Weapon: "chryssalid_claw",
		Points: 15, Rank: 1, Aggression: 10,
	},
	{
		Name: "Chryssalid Queen", ShortName: "CHQ",
		HP: 35, TU: 60, Accuracy: 75, Bravery: 100, Reactions: 80,
		Strength: 25, Psi: 0, Armour: 12, Weapon: "chryssalid_claw",
		Points: 35, Rank: 4, Aggression: 10,
	},
	{
		Name: "Hyperworm", ShortName: "HYP",
		HP: 8, TU: 70, Accuracy: 50, Bravery: 30, Reactions: 65,
		Strength: 6, Psi: 0, Armour: 3, Weapon: "plasma_pistol",
		Points: 4, Rank: 0, Aggression: 5,
	},
	{
		Name: "Silacoid", ShortName: "SIL",
		HP: 20, TU: 40, Accuracy: 45, Bravery: 80, Reactions: 35,
		Strength: 16, Psi: 0, Armour: 20, Weapon: "plasma_pistol",
		Points: 10, Rank: 1, Aggression: 4,
	},
	// Rank 2 - High tier
	{
		Name: "Muton", ShortName: "MUT",
		HP: 25, TU: 55, Accuracy: 55, Bravery: 70, Reactions: 50,
		Strength: 20, Psi: 0, Armour: 18, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 8,
	},
	{
		Name: "Muton Navigator", ShortName: "MUN",
		HP: 27, TU: 58, Accuracy: 58, Bravery: 75, Reactions: 53,
		Strength: 21, Psi: 0, Armour: 19, Weapon: "plasma_rifle",
		Points: 16, Rank: 3, Aggression: 8,
	},
	{
		Name: "Muton Commander", ShortName: "MUC",
		HP: 30, TU: 62, Accuracy: 62, Bravery: 85, Reactions: 58,
		Strength: 23, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 25, Rank: 4, Aggression: 9,
	},
	{
		Name: "Cyberdisc", ShortName: "CYB",
		HP: 30, TU: 50, Accuracy: 65, Bravery: 100, Reactions: 60,
		Strength: 15, Psi: 0, Armour: 22, Weapon: "heavy_plasma",
		Points: 20, Rank: 2, Aggression: 7,
	},
	{
		Name: "Celatid", ShortName: "CEL",
		HP: 12, TU: 60, Accuracy: 60, Bravery: 50, Reactions: 55,
		Strength: 10, Psi: 0, Armour: 6, Weapon: "plasma_pistol",
		Points: 8, Rank: 1, Aggression: 6,
	},
	// Rank 3 - Elite
	{
		Name: "Ethereal", ShortName: "ETH",
		HP: 18, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 70,
		Strength: 10, Psi: 80, Armour: 12, Weapon: "plasma_rifle",
		Points: 25, Rank: 4, Aggression: 5,
	},
	{
		Name: "Ethereal Navigator", ShortName: "ETN",
		HP: 20, TU: 68, Accuracy: 73, Bravery: 100, Reactions: 73,
		Strength: 11, Psi: 90, Armour: 13, Weapon: "plasma_rifle",
		Points: 30, Rank: 5, Aggression: 4,
	},
	{
		Name: "Ethereal Commander", ShortName: "ETC",
		HP: 24, TU: 72, Accuracy: 78, Bravery: 100, Reactions: 78,
		Strength: 13, Psi: 100, Armour: 16, Weapon: "heavy_plasma",
		Points: 50, Rank: 6, Aggression: 4,
	},
	// Rank 4 - Boss
	{
		Name: "Reaper", ShortName: "REAP",
		HP: 50, TU: 45, Accuracy: 50, Bravery: 100, Reactions: 40,
		Strength: 30, Psi: 0, Armour: 25, Weapon: "reaper_claw",
		Points: 30, Rank: 3, Aggression: 9,
	},
	{
		Name: "Sectopod", ShortName: "SEC3",
		HP: 60, TU: 40, Accuracy: 70, Bravery: 100, Reactions: 50,
		Strength: 25, Psi: 0, Armour: 30, Weapon: "heavy_plasma",
		Points: 50, Rank: 5, Aggression: 8,
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

func GetAlienByRank(minRank int) *AlienType {
	var best *AlienType
	for i := range AlienTypes {
		if AlienTypes[i].Rank >= minRank {
			if best == nil || AlienTypes[i].Rank < best.Rank {
				best = &AlienTypes[i]
			}
		}
	}
	return best
}
