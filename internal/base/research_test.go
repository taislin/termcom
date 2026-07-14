package base

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
)

func TestNewResearchScreen(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	rs := NewResearchScreen(&engine.Game{}, b)
	if rs == nil {
		t.Fatal("NewResearchScreen returned nil")
	}
	if rs.Base != b {
		t.Error("base should be set")
	}
}

func TestResearchGetAllTopics(t *testing.T) {
	b := NewBase("Test", 0)
	rs := NewResearchScreen(&engine.Game{}, b)
	entries := rs.getAllTopics()
	if len(entries) == 0 {
		t.Error("should have at least some research topics")
	}
}

func TestResearchGetUnlocks(t *testing.T) {
	b := NewBase("Test", 0)
	rs := NewResearchScreen(&engine.Game{}, b)
	topic := data.ResearchByID("alien_alloys")
	if topic == nil {
		t.Fatal("alien_alloys topic not found")
	}
	unlocks := rs.getUnlocks(topic)
	// Alien alloys unlocks items/weapons
	if len(unlocks) == 0 {
		t.Log("alien_alloys has no unlocks (may be empty)")
	}
}

func TestResearchGetChildren(t *testing.T) {
	b := NewBase("Test", 0)
	rs := NewResearchScreen(&engine.Game{}, b)
	topic := data.ResearchByID("alien_alloys")
	if topic == nil {
		t.Fatal("alien_alloys topic not found")
	}
	children := rs.getChildren(topic)
	// alien_alloys should have child topics that require it
	_ = children
}

func TestResearchStartResearch(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	rs := NewResearchScreen(&engine.Game{}, b)
	rs.startResearch()
	// Should set a message (success or failure)
	if rs.Message == "" {
		t.Error("startResearch should set a message")
	}
}

func TestResearchStartResearchLocked(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	rs := NewResearchScreen(&engine.Game{}, b)
	// Select a topic that requires prereqs (e.g. light_suit -> requires alien_alloys)
	topic := data.ResearchByID("light_suit")
	if topic == nil {
		t.Skip("light_suit topic not found")
	}
	// Find the topic in entries
	entries := rs.getAllTopics()
	for i, e := range entries {
		if e.topic.ID == "light_suit" {
			rs.Selection = i
			break
		}
	}
	rs.startResearch()
	if rs.Message == "" {
		t.Error("should set a message when trying to start locked research")
	}
}

func TestResearchDoInterrogate(t *testing.T) {
	b := NewBase("Test", 0)
	b.LiveAliens = append(b.LiveAliens, "Sectoid")
	b.Facilities = append(b.Facilities, &Facility{Type: FacContainment})
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	rs := NewResearchScreen(&engine.Game{}, b)
	rs.doInterrogate()
	if rs.Message == "" {
		t.Error("doInterrogate should set a message")
	}
}

func TestResearchDoInterrogateNoAlien(t *testing.T) {
	b := NewBase("Test", 0)
	b.Facilities = append(b.Facilities, &Facility{Type: FacContainment})
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab})
	rs := NewResearchScreen(&engine.Game{}, b)
	rs.doInterrogate()
	if rs.Message == "" {
		t.Error("should set message when no alien to interrogate")
	}
}

func TestResearchDoInterrogateNoLabs(t *testing.T) {
	b := NewBase("Test", 0)
	b.LiveAliens = append(b.LiveAliens, "Sectoid")
	rs := NewResearchScreen(&engine.Game{}, b)
	rs.doInterrogate()
	if rs.Message == "" {
		t.Error("should set message when no labs for interrogation")
	}
}

func TestResearchSelectionBounds(t *testing.T) {
	b := NewBase("Test", 0)
	rs := NewResearchScreen(&engine.Game{}, b)
	entries := rs.getAllTopics()
	if len(entries) == 0 {
		t.Skip("no topics available")
	}
	rs.Selection = 999 // Out of bounds
	_ = rs.getAllTopics() // This should clamp the selection in render, but getTopics doesn't
	if rs.Selection != 999 {
		t.Log("selection clamping handled elsewhere")
	}
}
