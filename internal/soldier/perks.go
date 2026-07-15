package soldier

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/taislin/termcom/internal/language"
)

type Perk struct {
	ID          string
	Name        string
	Description string
	StatBonuses StatBonus
	BattleMod   BattleModifier
}

type StatBonus struct {
	HP        int
	TU        int
	Accuracy  int
	Bravery   int
	Reactions int
	Strength  int
	PsiSkill  int
}

type BattleModifier int

const (
	BModNone BattleModifier = iota
	BModLightningReflexes
	BModMarksman
	BModGrenadier
	BModFieldMedic
	BModIronWill
	BModSteadyAim
	BModCloseCombat
	BModOverwatch
	BModDemolitions
	BModScavenger
	BModTough
)

var AllPerks = []Perk{
	{
		ID: "lightning_reflexes", Name: "Lightning Reflexes",
		Description: "+10 Reactions",
		StatBonuses: StatBonus{Reactions: 10},
	},
	{
		ID: "marksman", Name: "Marksman",
		Description: "+15% accuracy at range > 8 tiles",
		BattleMod:   BModMarksman,
	},
	{
		ID: "grenadier", Name: "Grenadier",
		Description: "+2 grenade splash radius",
		BattleMod:   BModGrenadier,
	},
	{
		ID: "field_medic", Name: "Field Medic",
		Description: "Medikit heals 15 HP instead of 10",
		BattleMod:   BModFieldMedic,
	},
	{
		ID: "iron_will", Name: "Iron Will",
		Description: "+20 Psi Strength",
		StatBonuses: StatBonus{PsiSkill: 10},
		BattleMod:   BModIronWill,
	},
	{
		ID: "steady_aim", Name: "Steady Aim",
		Description: "+10% accuracy when not moving",
		BattleMod:   BModSteadyAim,
	},
	{
		ID: "close_combat", Name: "Close Combat Specialist",
		Description: "+15% accuracy at range ≤ 4 tiles",
		BattleMod:   BModCloseCombat,
	},
	{
		ID: "overwatch", Name: "Overwatch Expert",
		Description: "+20% reaction fire accuracy",
		BattleMod:   BModOverwatch,
	},
	{
		ID: "demolitions", Name: "Demolitions",
		Description: "+50% grenade damage",
		BattleMod:   BModDemolitions,
	},
	{
		ID: "scavenger", Name: "Scavenger",
		Description: "+25% loot from battles",
		BattleMod:   BModScavenger,
	},
	{
		ID: "tough", Name: "Tough",
		Description: "+5 MaxHP",
		StatBonuses: StatBonus{HP: 5},
		BattleMod:   BModTough,
	},
	{
		ID: "quick_learner", Name: "Quick Learner",
		Description: "+50% XP from battles",
		StatBonuses: StatBonus{},
	},
}

func RollPerk(rng *rand.Rand, currentPerks []string) *Perk {
	available := make([]Perk, 0, len(AllPerks))
	for _, p := range AllPerks {
		if !hasPerk(currentPerks, p.ID) {
			available = append(available, p)
		}
	}
	if len(available) == 0 {
		return nil
	}
	return &available[rng.Intn(len(available))]
}

func hasPerk(perks []string, id string) bool {
	for _, p := range perks {
		if p == id {
			return true
		}
	}
	return false
}

func (s *Soldier) ApplyPerk(perk Perk) {
	s.Perks = append(s.Perks, perk.ID)
	s.MaxHP += perk.StatBonuses.HP
	s.HP += perk.StatBonuses.HP
	s.MaxTU += perk.StatBonuses.TU
	s.TU += perk.StatBonuses.TU
	s.Accuracy += perk.StatBonuses.Accuracy
	s.Bravery += perk.StatBonuses.Bravery
	s.Reactions += perk.StatBonuses.Reactions
	s.Strength += perk.StatBonuses.Strength
	s.PsiSkill += perk.StatBonuses.PsiSkill
}

func (s *Soldier) HasPerk(id string) bool {
	for _, p := range s.Perks {
		if p == id {
			return true
		}
	}
	return false
}

func (s *Soldier) HasBattleMod(mod BattleModifier) bool {
	for _, pID := range s.Perks {
		for _, p := range AllPerks {
			if p.ID == pID && p.BattleMod == mod {
				return true
			}
		}
	}
	return false
}

// LangName returns the localized perk name, falling back to the English Name
// when no translation key is registered.
func (p *Perk) LangName() string {
	key := "PERK_" + strings.ToUpper(strings.ReplaceAll(p.ID, "_", " "))
	if s := language.String(key); s != key {
		return s
	}
	return p.Name
}

// LangDesc returns the localized perk description, falling back to the English
// Description when no translation key is registered.
func (p *Perk) LangDesc() string {
	key := "PERK_" + strings.ToUpper(strings.ReplaceAll(p.ID, "_", " ")) + "_DESC"
	if s := language.String(key); s != key {
		return s
	}
	return p.Description
}

func (s *Soldier) PerkNames() []string {
	var names []string
	for _, pID := range s.Perks {
		for _, p := range AllPerks {
			if p.ID == pID {
				names = append(names, p.LangName())
			}
		}
	}
	return names
}

func FormatPerks(perks []string) string {
	if len(perks) == 0 {
		return language.String("NONE")
	}
	names := make([]string, 0, len(perks))
	for _, pID := range perks {
		for _, p := range AllPerks {
			if p.ID == pID {
				names = append(names, p.LangName())
			}
		}
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

func (s *Soldier) FormatPerksShort() string {
	return fmt.Sprintf("[%s]", FormatPerks(s.Perks))
}
