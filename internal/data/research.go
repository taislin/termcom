package data

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

var ResearchTree []ResearchTopic

func InitResearchTree(seed int64, aliens []*AlienSpecies) {
	ResearchTree = GenerateTechTree(seed, aliens)
}

func ResearchByID(id string) *ResearchTopic {
	for i := range ResearchTree {
		if ResearchTree[i].ID == id {
			return &ResearchTree[i]
		}
	}
	return nil
}
