package engine

import (
	"math"
	"math/rand"
	"sync"
)

type Camera struct {
	mu             sync.RWMutex
	X, Y           int
	OffsetX, OffsetY int
	ShakeIntensity float64
	decay          float64
}

func NewCamera(x, y int) *Camera {
	return &Camera{
		X:    x,
		Y:    y,
		decay: 8.0,
	}
}

func (c *Camera) UpdateShake(dt float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ShakeIntensity <= 0 {
		return
	}

	c.ShakeIntensity -= c.decay * dt
	if c.ShakeIntensity < 0 {
		c.ShakeIntensity = 0
		c.OffsetX = 0
		c.OffsetY = 0
		return
	}

	c.OffsetX = int((rand.Float64()*2 - 1) * c.ShakeIntensity)
	c.OffsetY = int((rand.Float64()*2 - 1) * c.ShakeIntensity)
}

func (c *Camera) TriggerShake(intensity float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ShakeIntensity = math.Max(c.ShakeIntensity, intensity)
	c.OffsetX = int((rand.Float64()*2 - 1) * intensity)
	c.OffsetY = int((rand.Float64()*2 - 1) * intensity)
}

func (c *Camera) SetTarget(x, y int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.X = x
	c.Y = y
}

func (c *Camera) ViewX() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.X + c.OffsetX
}

func (c *Camera) ViewY() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Y + c.OffsetY
}

func (c *Camera) Pos() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.X + c.OffsetX, c.Y + c.OffsetY
}

func (c *Camera) Pan(dx, dy int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.X += dx
	c.Y += dy
}
