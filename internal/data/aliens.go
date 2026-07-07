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
	{
		Name: "Sectoid", ShortName: "SEC",
		HP: 10, TU: 50, Accuracy: 55, Bravery: 40, Reactions: 50,
		Strength: 8, Psi: 40, Armour: 5, Weapon: "plasma_pistol",
		Points: 5, Rank: 0, Aggression: 3,
	},
	{
		Name: "Sectoid Leader", ShortName: "SEL",
		HP: 12, TU: 55, Accuracy: 60, Bravery: 50, Reactions: 55,
		Strength: 9, Psi: 60, Armour: 8, Weapon: "plasma_rifle",
		Points: 10, Rank: 1, Aggression: 4,
	},
	{
		Name: "Floater", ShortName: "FLT",
		HP: 15, TU: 55, Accuracy: 60, Bravery: 50, Reactions: 60,
		Strength: 12, Psi: 10, Armour: 10, Weapon: "plasma_rifle",
		Points: 8, Rank: 1, Aggression: 6,
	},
	{
		Name: "Floater Leader", ShortName: "FLL",
		HP: 18, TU: 60, Accuracy: 65, Bravery: 60, Reactions: 65,
		Strength: 14, Psi: 25, Armour: 12, Weapon: "plasma_rifle",
		Points: 14, Rank: 2, Aggression: 7,
	},
	{
		Name: "Muton", ShortName: "MUT",
		HP: 25, TU: 55, Accuracy: 55, Bravery: 70, Reactions: 50,
		Strength: 20, Psi: 0, Armour: 18, Weapon: "plasma_rifle",
		Points: 12, Rank: 2, Aggression: 8,
	},
	{
		Name: "Muton Leader", ShortName: "MUL",
		HP: 28, TU: 60, Accuracy: 60, Bravery: 80, Reactions: 55,
		Strength: 22, Psi: 0, Armour: 20, Weapon: "plasma_rifle",
		Points: 18, Rank: 3, Aggression: 9,
	},
	{
		Name: "Ethereal", ShortName: "ETH",
		HP: 18, TU: 65, Accuracy: 70, Bravery: 100, Reactions: 70,
		Strength: 10, Psi: 80, Armour: 12, Weapon: "plasma_rifle",
		Points: 25, Rank: 4, Aggression: 5,
	},
	{
		Name: "Ethereal Leader", ShortName: "EHL",
		HP: 22, TU: 70, Accuracy: 75, Bravery: 100, Reactions: 75,
		Strength: 12, Psi: 100, Armour: 15, Weapon: "plasma_rifle",
		Points: 40, Rank: 5, Aggression: 4,
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
