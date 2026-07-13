package engine

import (
	"testing"
)

func TestCameraShake(t *testing.T) {
	c := NewCamera(10, 10)
	
	// Test TriggerShake
	c.TriggerShake(5.0)
	if c.ShakeIntensity != 5.0 {
		t.Errorf("Expected ShakeIntensity 5.0, got %f", c.ShakeIntensity)
	}

	// Test UpdateShake with dt=0.5. Intensity = 5.0 - (8.0 * 0.5) = 1.0
	c.UpdateShake(0.5)
	if c.ShakeIntensity != 1.0 {
		t.Errorf("Expected ShakeIntensity 1.0, got %f", c.ShakeIntensity)
	}

	// Test decay to 0. Intensity = 1.0 - (8.0 * 0.2) = -0.6 -> 0
	c.UpdateShake(0.2)
	if c.ShakeIntensity != 0.0 {
		t.Errorf("Expected ShakeIntensity 0.0, got %f", c.ShakeIntensity)
	}
	if c.OffsetX != 0 || c.OffsetY != 0 {
		t.Errorf("Expected offsets 0,0, got %d,%d", c.OffsetX, c.OffsetY)
	}
}

func TestCameraPos(t *testing.T) {
	c := NewCamera(10, 20)
	c.OffsetX = 2
	c.OffsetY = -3
	
	x, y := c.Pos()
	if x != 12 || y != 17 {
		t.Errorf("Expected Pos 12,17, got %d,%d", x, y)
	}
}
