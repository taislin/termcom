package engine

import "testing"

func TestStateStack(t *testing.T) {
	g := &Game{state: StateGeoscape}
	
	g.PushState(StateBase)
	if g.state != StateBase {
		t.Errorf("Expected state StateBase, got %v", g.state)
	}
	
	g.PopState()
	if g.state != StateGeoscape {
		t.Errorf("Expected state StateGeoscape, got %v", g.state)
	}
}
