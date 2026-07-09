package data

import "testing"

func initTree() {
	species, _ := GenerateSpecies(42)
	InitResearchTree(42, species)
}

func TestResearchTreePopulated(t *testing.T) {
	initTree()
	if len(ResearchTree) == 0 {
		t.Fatal("no research topics defined")
	}
}

func TestResearchCostsPositive(t *testing.T) {
	initTree()
	for _, rt := range ResearchTree {
		if rt.Cost <= 0 {
			t.Errorf("%s: invalid cost %d", rt.Name, rt.Cost)
		}
		if rt.ID == "" {
			t.Error("research topic with empty ID")
		}
	}
}

func TestResearchByID(t *testing.T) {
	initTree()
	r := ResearchByID("alien_alloys")
	if r == nil {
		t.Fatal("ResearchByID(alien_alloys) returned nil")
	}
	if r.Name != "Alien Alloys" {
		t.Errorf("expected Alien Alloys, got %s", r.Name)
	}
	if ResearchByID("nonexistent") != nil {
		t.Error("expected nil for nonexistent topic")
	}
}

func TestResearchPrerequisites(t *testing.T) {
	initTree()
	for _, rt := range ResearchTree {
		for _, req := range rt.Requires {
			if ResearchByID(req) == nil {
				t.Errorf("%s: missing prerequisite %s", rt.Name, req)
			}
		}
	}
}

func TestNoCircularDeps(t *testing.T) {
	initTree()
	byID := make(map[string]*ResearchTopic)
	for i := range ResearchTree {
		byID[ResearchTree[i].ID] = &ResearchTree[i]
	}
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var dfs func(string) bool
	dfs = func(id string) bool {
		if visited[id] {
			return false
		}
		if inStack[id] {
			return true
		}
		inStack[id] = true
		if topic, ok := byID[id]; ok {
			for _, req := range topic.Requires {
				if dfs(req) {
					return true
				}
			}
		}
		inStack[id] = false
		visited[id] = true
		return false
	}
	for _, topic := range ResearchTree {
		if dfs(topic.ID) {
			t.Fatalf("circular dependency detected at %s", topic.ID)
		}
	}
}

func TestDeterministicTree(t *testing.T) {
	species1, _ := GenerateSpecies(123)
	tree1 := GenerateTechTree(123, species1)
	species2, _ := GenerateSpecies(123)
	tree2 := GenerateTechTree(123, species2)
	if len(tree1) != len(tree2) {
		t.Fatalf("different tree sizes: %d vs %d", len(tree1), len(tree2))
	}
	for i := range tree1 {
		if tree1[i].ID != tree2[i].ID {
			t.Errorf("topic %d: ID mismatch %s vs %s", i, tree1[i].ID, tree2[i].ID)
		}
		if tree1[i].Cost != tree2[i].Cost {
			t.Errorf("topic %d: cost mismatch %d vs %d", i, tree1[i].Cost, tree2[i].Cost)
		}
	}
}

func TestDifferentSeedsProduceDifferentCosts(t *testing.T) {
	species1, _ := GenerateSpecies(100)
	tree1 := GenerateTechTree(100, species1)
	species2, _ := GenerateSpecies(200)
	tree2 := GenerateTechTree(200, species2)
	sameCosts := true
	for i := range tree1 {
		if i < len(tree2) && tree1[i].Cost != tree2[i].Cost {
			sameCosts = false
			break
		}
	}
	if sameCosts {
		t.Error("different seeds produced identical costs — variance not working")
	}
}
