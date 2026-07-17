package battle

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/soldier"
)

// TestHandleEventNoDeadlock verifies that input handlers which mutate shared
// Battlescape state via the Set* helpers do not deadlock on bs.State.mu.
// Regression test for the turn-passing freeze (Enter/Q/E all hit Set* under
// the lock already held by HandleEvent).
func TestHandleEventNoDeadlock(t *testing.T) {
	g, _, err := engine.NewGameWeb(80, 24)
	if err != nil {
		t.Skipf("skipping: cannot create game: %v", err)
	}
	squad := []*soldier.Soldier{
		soldier.NewSoldier("Alpha"),
		soldier.NewSoldier("Bravo"),
	}
	bs := NewBattlescape(g, nil, squad, "Crash", 777)

	// Place the cursor on the first friendly soldier so Enter triggers SetSelected.
	var su *Unit
	for _, u := range bs.Units {
		if u.Faction == 0 && u.Alive && u.Soldier != nil {
			su = u
			break
		}
	}
	if su == nil {
		t.Fatal("no friendly unit spawned")
	}
	bs.CursorX = su.X
	bs.CursorY = su.Y

	keys := []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyEnter, " ", tcell.ModNone),  // LeftClick -> SetSelected (order)
		tcell.NewEventKey(tcell.KeyRune, "q", tcell.ModNone),   // cycleUnit -> SetSelected
		tcell.NewEventKey(tcell.KeyRune, "e", tcell.ModNone),   // EndTurn -> SetPhase
	}

	for i, k := range keys {
		done := make(chan struct{})
		go func() {
			bs.HandleEvent(k)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("HandleEvent deadlocked on key #%d (%v) — re-entrant bs.State.mu lock", i, k)
		}
	}

	// Drive the alien turn to completion so control returns to the player.
	for i := 0; i < 20000; i++ {
		bs.Update()
		if bs.Phase == PhasePlayerTurn && bs.Turn > 1 {
			break
		}
	}
	if bs.Phase != PhasePlayerTurn {
		t.Fatalf("alien turn never returned control to player (phase=%v, turn=%d)", bs.Phase, bs.Turn)
	}
}
