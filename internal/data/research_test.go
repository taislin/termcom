package data

import "testing"

func TestResearchTreePopulated(t *testing.T) {
	if len(ResearchTree) == 0 {
		t.Fatal("no research topics defined")
	}
}

func TestResearchCostsPositive(t *testing.T) {
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
	for _, rt := range ResearchTree {
		for _, req := range rt.Requires {
			if ResearchByID(req) == nil {
				t.Errorf("%s: missing prerequisite %s", rt.Name, req)
			}
		}
	}
}
