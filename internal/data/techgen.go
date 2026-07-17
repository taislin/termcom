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

const (
	autopsyCostBase = 40
	autopsyCostRange = 30
	studyCostBase = 60
	studyCostRange = 50
	costModMin = 0.85
	costModRange = 0.30
	techCostFloor = 10
	prereqMin = 1
	prereqRange = 2
)

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

var studyNames = []string{
	"Xenobiology", "Behavioral Analysis", "Tactical Study", "Morphology Review",
	"Neural Mapping", "Genetic Analysis", "Combat Doctrine", "Psionic Profile",
}

func generateSpeciesStudyTopic(rng *rand.Rand, sp *AlienSpecies, autopsyID string) techDef {
	nameSuffix := studyNames[rng.Intn(len(studyNames))]
	return techDef{
		ID:       sp.Name + "_study",
		Name:     sp.Name + " " + nameSuffix,
		BaseCost: studyCostBase + rng.Intn(studyCostRange),
		Tier:     2,
		Requires: []string{autopsyID},
		AlienLore: true,
	}
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
			BaseCost: autopsyCostBase + rng.Intn(autopsyCostRange),
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
				BaseCost: autopsyCostBase + rng.Intn(autopsyCostRange),
				Tier:     1,
				AlienLore: true,
			}
			autopsyIDs = append(autopsyIDs, def.ID)
			defs = append(defs, def)
		}
	}

	for i, sp := range aliens {
		if i < len(autopsyIDs) {
			study := generateSpeciesStudyTopic(rng, sp, autopsyIDs[i])
			defs = append(defs, study)
		}
	}

	for i := range defs {
		if defs[i].Tier > 1 {
			modifier := costModMin + rng.Float64()*costModRange
			defs[i].BaseCost = int(float64(defs[i].BaseCost) * modifier)
			if defs[i].BaseCost < techCostFloor {
				defs[i].BaseCost = techCostFloor
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
			n := prereqMin + rng.Intn(prereqRange)
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

// checkTechTreeValidity verifies the generated tech DAG has no dangling
// edges, no cycles, strict tier ordering, and that every topic is reachable
// from a root (a topic with no prerequisites). A "dead-end" tree is one in
// which some topic can never be researched because its prerequisites form an
// isolated or cyclic component.
func buildTechMap(topics []ResearchTopic) map[string]ResearchTopic {
	byID := make(map[string]ResearchTopic, len(topics))
	for _, t := range topics {
		byID[t.ID] = t
	}
	return byID
}

func hasCycle(topics []ResearchTopic, byID map[string]ResearchTopic) error {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var visit func(id string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if inStack[id] {
			return fmt.Errorf("circular dependency involving %s", id)
		}
		t, ok := byID[id]
		if !ok {
			return fmt.Errorf("dangling prerequisite %q in tech tree", id)
		}
		inStack[id] = true
		for _, req := range t.Requires {
			if rt, ok := byID[req]; ok {
				if rt.Tier >= t.Tier {
					return fmt.Errorf("tier violation: %s (tier %d) requires %s (tier %d)",
						t.ID, t.Tier, req, rt.Tier)
				}
			}
			if err := visit(req); err != nil {
				return err
			}
		}
		inStack[id] = false
		visited[id] = true
		return nil
	}

	for _, t := range topics {
		if err := visit(t.ID); err != nil {
			return err
		}
	}
	return nil
}

func collectReachable(topics []ResearchTopic, byID map[string]ResearchTopic) (map[string]bool, error) {
	reachable := make(map[string]bool)
	for _, t := range topics {
		if len(t.Requires) == 0 {
			reachable[t.ID] = true
		}
	}
	changed := true
	for changed {
		changed = false
		for _, t := range topics {
			if reachable[t.ID] {
				continue
			}
			allMet := true
			for _, r := range t.Requires {
				if !reachable[r] {
					allMet = false
					break
				}
			}
			if allMet {
				reachable[t.ID] = true
				changed = true
			}
		}
	}
	for _, t := range topics {
		if !reachable[t.ID] {
			return nil, fmt.Errorf("dead-end topic %s: unreachable from any root", t.ID)
		}
	}
	return reachable, nil
}

func checkTechTreeValidity(topics []ResearchTopic) error {
	byID := buildTechMap(topics)
	if err := hasCycle(topics, byID); err != nil {
		return err
	}
	if _, err := collectReachable(topics, byID); err != nil {
		return err
	}
	return nil
}

// validateDAG is a fail-fast runtime guard that panics if the generated tech
// tree is not a valid, fully-reachable DAG.
func validateDAG(topics []ResearchTopic) {
	if err := checkTechTreeValidity(topics); err != nil {
		panic(err)
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
