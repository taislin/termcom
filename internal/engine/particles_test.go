package engine

import (
	"testing"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func TestParticleSystem(t *testing.T) {
	maxCount := 10
	ps := NewParticleSystem(maxCount)

	// Test Spawn
	style := tcell.StyleDefault.Foreground(color.Red)
	ps.Spawn(0, 0, 1, 1, '*', style, 1.0, 0.1)

	ps.mu.RLock()
	if len(ps.particles) != 1 {
		t.Errorf("Expected 1 particle, got %d", len(ps.particles))
	}
	p := ps.particles[0]
	if p.X != 0 || p.Y != 0 || p.VX != 1 || p.VY != 1 || !p.Active {
		t.Errorf("Particle state incorrect: %+v", p)
	}
	ps.mu.RUnlock()

	// Test Update
	ps.Update(0.5) // Life should be 0.5

	ps.mu.RLock()
	if ps.particles[0].Life != 0.5 {
		t.Errorf("Expected life 0.5, got %f", ps.particles[0].Life)
	}
	ps.mu.RUnlock()

	// Test Update with decay
	ps.Update(0.6) // Life should be <= 0

	ps.mu.RLock()
	if len(ps.particles) != 0 {
		t.Errorf("Expected 0 particles, got %d", len(ps.particles))
	}
	ps.mu.RUnlock()
}

func TestParticleSystemClear(t *testing.T) {
	ps := NewParticleSystem(10)
	style := tcell.StyleDefault
	ps.Spawn(0, 0, 0, 0, '*', style, 1.0, 0.1)

	ps.Clear()

	ps.mu.RLock()
	if len(ps.particles) != 0 {
		t.Errorf("Expected 0 particles after Clear, got %d", len(ps.particles))
	}
	ps.mu.RUnlock()
}
