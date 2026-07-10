package data

import (
	"fmt"
	"math/rand"
)

type techDef struct {
	ID          string
	Name        string
	BaseCost    int
	Tier        int
	Requires    []string
	UnlockItems []string
	UnlockWeap  []string
	UnlockArmor []string
	AlienLore   bool
}

var baseTechs = []techDef{
	{ID: "alien_alloys", Name: "Alien Alloys", BaseCost: 60, Tier: 1, UnlockItems: []string{"alloys"}},
	{ID: "elerium", Name: "Elerium-115", BaseCost: 80, Tier: 1, UnlockItems: []string{"elerium"}},
	{ID: "ufo_nav", Name: "UFO Navigation", BaseCost: 100, Tier: 1, AlienLore: true},
	{ID: "ufo_power", Name: "UFO Power Source", BaseCost: 120, Tier: 1, AlienLore: true},
	{ID: "alien_comm", Name: "Alien Communications", BaseCost: 90, Tier: 1, AlienLore: true},

	{ID: "laser_weapons", Name: "Laser Weapons", BaseCost: 120, Tier: 2, UnlockWeap: []string{"laser_pistol", "laser_rifle"}},
	{ID: "personal_armour", Name: "Personal Armour", BaseCost: 80, Tier: 2, UnlockArmor: []string{"personal"}},

	{ID: "plasma_weapons", Name: "Plasma Weapons", BaseCost: 200, Tier: 3, UnlockWeap: []string{"plasma_pistol", "plasma_rifle"}},
	{ID: "light_suit", Name: "Light Suit", BaseCost: 150, Tier: 3, UnlockArmor: []string{"light"}},
	{ID: "ufo_propulsion", Name: "UFO Propulsion", BaseCost: 110, Tier: 3, AlienLore: true},

	{ID: "heavy_plasma", Name: "Heavy Plasma", BaseCost: 250, Tier: 4, UnlockWeap: []string{"heavy_plasma"}},
	{ID: "medium_suit", Name: "Medium Suit", BaseCost: 200, Tier: 4, UnlockArmor: []string{"medium"}},
	{ID: "mind_control", Name: "Mind Control", BaseCost: 150, Tier: 4, AlienLore: true},

	{ID: "heavy_suit", Name: "Heavy Suit", BaseCost: 280, Tier: 5, UnlockArmor: []string{"heavy"}},
	{ID: "power_suit", Name: "Power Suit", BaseCost: 400, Tier: 5, UnlockArmor: []string{"power_suit"}},
	{ID: "flight_suit", Name: "Flying Suit", BaseCost: 500, Tier: 5, UnlockArmor: []string{"flight_suit"}},
}

func GenerateTechTree(seed int64, aliens []*AlienSpecies) []ResearchTopic {
	rng := rand.New(rand.NewSource(seed))

	defs := make([]techDef, len(baseTechs))
	copy(defs, baseTechs)

	var autopsyIDs []string
	for _, sp := range aliens {
		def := techDef{
			ID:       sp.Name + "_autopsy",
			Name:     sp.Name + " Autopsy",
			BaseCost: 40 + rng.Intn(30),
			Tier:     1,
			AlienLore: true,
		}
		autopsyIDs = append(autopsyIDs, def.ID)
		defs = append(defs, def)
	}
	if len(autopsyIDs) == 0 {
		for _, sp := range fallbackAutopsySpecies {
			def := techDef{
				ID:       sp + "_autopsy",
				Name:     sp + " Autopsy",
				BaseCost: 40 + rng.Intn(30),
				Tier:     1,
				AlienLore: true,
			}
			autopsyIDs = append(autopsyIDs, def.ID)
			defs = append(defs, def)
		}
	}

	for i := range defs {
		if defs[i].Tier > 1 {
			modifier := 0.85 + rng.Float64()*0.30
			defs[i].BaseCost = int(float64(defs[i].BaseCost) * modifier)
			if defs[i].BaseCost < 10 {
				defs[i].BaseCost = 10
			}
		}
	}

	tierDefs := make(map[int][]int)
	for i := range defs {
		tierDefs[defs[i].Tier] = append(tierDefs[defs[i].Tier], i)
	}

	shuffledAutopsies := make([]string, len(autopsyIDs))
	copy(shuffledAutopsies, autopsyIDs)
	rng.Shuffle(len(shuffledAutopsies), func(i, j int) {
		shuffledAutopsies[i], shuffledAutopsies[j] = shuffledAutopsies[j], shuffledAutopsies[i]
	})
	autopsyIdx := 0

	for tier := 2; tier <= 5; tier++ {
		for _, idx := range tierDefs[tier] {
			d := &defs[idx]

			if isWeaponTech(d.ID) && autopsyIdx < len(shuffledAutopsies) {
				d.Requires = append(d.Requires, shuffledAutopsies[autopsyIdx])
				autopsyIdx++
			}

			var lowerTier []int
			for t := 1; t < tier; t++ {
				lowerTier = append(lowerTier, tierDefs[t]...)
			}
			if len(lowerTier) == 0 {
				continue
			}

			rng.Shuffle(len(lowerTier), func(i, j int) {
				lowerTier[i], lowerTier[j] = lowerTier[j], lowerTier[i]
			})

			if len(d.Requires) == 0 {
				d.Requires = append(d.Requires, defs[lowerTier[0]].ID)
				lowerTier = lowerTier[1:]
			}

			remaining := lowerTier
			rng.Shuffle(len(remaining), func(i, j int) {
				remaining[i], remaining[j] = remaining[j], remaining[i]
			})
			n := 1 + rng.Intn(2)
			if n > len(remaining) {
				n = len(remaining)
			}
			for i := 0; i < n; i++ {
				d.Requires = append(d.Requires, defs[remaining[i]].ID)
			}
		}
	}

	for i := range defs {
		defs[i].Requires = dedupStrings(defs[i].Requires)
	}

	result := make([]ResearchTopic, len(defs))
	for i, d := range defs {
		result[i] = ResearchTopic{
			ID:          d.ID,
			Name:        d.Name,
			Cost:        d.BaseCost,
			Tier:        d.Tier,
			Requires:    d.Requires,
			UnlockItems: d.UnlockItems,
			UnlockWeap:  d.UnlockWeap,
			UnlockArmor: d.UnlockArmor,
			AlienLore:   d.AlienLore,
		}
	}

	validateDAG(result)

	return result
}

func isWeaponTech(id string) bool {
	switch id {
	case "laser_weapons", "plasma_weapons", "heavy_plasma", "mind_control":
		return true
	}
	return false
}

func validateDAG(topics []ResearchTopic) {
	byID := make(map[string]*ResearchTopic)
	for i := range topics {
		byID[topics[i].ID] = &topics[i]
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var visit func(id string) bool
	visit = func(id string) bool {
		if visited[id] {
			return false
		}
		if inStack[id] {
			return true
		}
		inStack[id] = true
		t := byID[id]
		if t == nil {
			panic(fmt.Sprintf("dangling prerequisite %q in tech tree", id))
		}
		for _, req := range t.Requires {
			if visit(req) {
				return true
			}
		}
		inStack[id] = false
		visited[id] = true
		return false
	}

	for _, t := range topics {
		if visit(t.ID) {
			panic(fmt.Sprintf("circular dependency detected involving %s", t.ID))
		}
	}
}

var fallbackAutopsySpecies = []string{"Sectoid", "Floater", "Muton", "Ethereal"}

func dedupStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
