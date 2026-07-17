package data

import (
	"strings"

	"github.com/taislin/termcom/internal/language"
)

type ResearchTopic struct {
	ID          string
	Name        string
	Cost        int
	Tier        int
	Requires    []string
	UnlockItems []string
	UnlockWeap  []string
	UnlockArmor []string
	AlienLore   bool
}

func (t *ResearchTopic) DisplayName() string {
	key := "RESEARCH_" + strings.ToUpper(strings.ReplaceAll(t.ID, " ", "_"))
	display := language.String(key)
	if display == key {
		return t.Name // fallback to English
	}
	return display
}

var ResearchTree []ResearchTopic

func InitResearchTree(seed int64, aliens []*AlienSpecies) {
	ResearchTree = GenerateTechTree(seed, aliens)
}

// ResearchByID returns a pointer to a copy of the matching topic.
func ResearchByID(id string) *ResearchTopic {
	for i := range ResearchTree {
		if ResearchTree[i].ID == id {
			t := ResearchTree[i]
			return &t
		}
	}
	return nil
}
