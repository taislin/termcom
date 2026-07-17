package engine

import (
	"math"
	"math/rand"
	"sync"

	"github.com/gdamore/tcell/v3"
)

type Particle struct {
	X, Y       float64
	VX, VY     float64
	Rune       rune
	Style      tcell.Style
	Life       float64
	FadeSpeed  float64
	Active     bool
	r, g, b    float64
	fadeTarget [3]float64
}

const PixelGravity = 9.8 // pixels/frame²; used as downward velocity increment per tick

var particlePool = sync.Pool{
	New: func() interface{} {
		return &Particle{}
	},
}

func getParticle() *Particle {
	return particlePool.Get().(*Particle)
}

func putParticle(p *Particle) {
	p.Active = false
	particlePool.Put(p)
}

type ParticleSystem struct {
	mu        sync.RWMutex
	particles []*Particle
	maxCount  int
}

func NewParticleSystem(maxCount int) *ParticleSystem {
	return &ParticleSystem{
		particles: make([]*Particle, 0, maxCount),
		maxCount:  maxCount,
	}
}

func (ps *ParticleSystem) Spawn(x, y, vx, vy float64, ch rune, style tcell.Style, life, fadeSpeed float64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.particles) >= ps.maxCount {
		return
	}

	p := getParticle()
	p.X = x
	p.Y = y
	p.VX = vx
	p.VY = vy
	p.Rune = ch
	p.Style = style
	p.Life = life
	p.FadeSpeed = fadeSpeed
	p.Active = true

	fg := style.GetForeground()
	p.r, p.g, p.b = colorRGB(fg)
	p.fadeTarget = [3]float64{32, 32, 32}

	ps.particles = append(ps.particles, p)
}

func (ps *ParticleSystem) Update(dt float64) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	n := 0
	for _, p := range ps.particles {
		if !p.Active {
			continue
		}

		p.Life -= dt
		if p.Life <= 0 {
			putParticle(p)
			continue
		}

		p.VY += PixelGravity * dt
		p.X += p.VX * dt
		p.Y += p.VY * dt

		factor := 1 - p.FadeSpeed*dt
		if factor < 0 {
			factor = 0
		}
		p.r = p.r*factor + p.fadeTarget[0]*(1-factor)
		p.g = p.g*factor + p.fadeTarget[1]*(1-factor)
		p.b = p.b*factor + p.fadeTarget[2]*(1-factor)

		newColor := tcell.NewRGBColor(int32(p.r), int32(p.g), int32(p.b))
		p.Style = p.Style.Foreground(newColor)

		ps.particles[n] = p
		n++
	}
	ps.particles = ps.particles[:n]
}

func (ps *ParticleSystem) Draw(s *ScreenRaw) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for _, p := range ps.particles {
		if !p.Active {
			continue
		}
		ix, iy := int(math.Round(p.X)), int(math.Round(p.Y))
		s.SetCell(ix, iy, p.Rune, p.Style)
	}
}

func (ps *ParticleSystem) Clear() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	for _, p := range ps.particles {
		putParticle(p)
	}
	ps.particles = ps.particles[:0]
}

func SpawnExplosion(ps *ParticleSystem, x, y int, color tcell.Color, count int) {
	r, g, b := colorRGB(color)
	fg := tcell.NewRGBColor(int32(r), int32(g), int32(b))

	for i := 0; i < count; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := 2 + rand.Float64()*6
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed - 4

		ch := '*'
		if rand.Intn(2) == 0 {
			ch = '+'
		}

		style := StyleDefault.Foreground(fg)
		ps.Spawn(float64(x), float64(y), vx, vy, ch, style, 0.4+rand.Float64()*0.6, 0.8)
	}
}

func SpawnRain(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < 3; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y - 1)
		style := StyleDefault.Foreground(tcell.NewRGBColor(100, 150, 255))
		ps.Spawn(rx, ry, -0.5, 12, '|', style, 1.5, 0.1)
	}
}

func SpawnSmoke(ps *ParticleSystem, x, y int, count int) {
	for i := 0; i < count; i++ {
		vx := (rand.Float64() - 0.5) * 2
		vy := -1 - rand.Float64()*3
		ch := '~'
		if rand.Intn(3) == 0 {
			ch = ':'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(128, 128, 128))
		ps.Spawn(float64(x)+rand.Float64()*2-1, float64(y), vx, vy, ch, style, 0.8+rand.Float64()*1.2, 0.5)
	}
}

func SpawnMenuDrift(ps *ParticleSystem, x, y, side int) {
	driftRunes := []rune{'°', '.', '+'}
	driftColors := [][3]int32{
		{192, 64, 255},
		{96, 96, 255},
		{255, 64, 192},
	}
	pick := rand.Intn(3)
	col := driftColors[pick]
	ch := driftRunes[rand.Intn(len(driftRunes))]
	fg := tcell.NewRGBColor(col[0], col[1], col[2])
	style := StyleDefault.Foreground(fg)
	vx := float64(side) * (0.5 + rand.Float64()*1.0)
	vy := -(6.0 + rand.Float64()*4.0)
	life := 1.0 + rand.Float64()*1.0
	ps.Spawn(float64(x), float64(y), vx, vy, ch, style, life, 0.6)
}

func SpawnMuzzleFlash(ps *ParticleSystem, x, y int) {
	flashRunes := []rune{'.', '*', '°', '+'}
	for i := 0; i < 6; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := 1 + rand.Float64()*3
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		ch := flashRunes[rand.Intn(len(flashRunes))]
		brightness := 200 + rand.Intn(55)
		fg := tcell.NewRGBColor(int32(brightness), int32(brightness-30), int32(50+rand.Intn(80)))
		style := StyleDefault.Foreground(fg)
		ps.Spawn(float64(x), float64(y), vx, vy, ch, style, 0.15+rand.Float64()*0.2, 2.0)
	}
}

func SpawnSnow(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < 2; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y - 1)
		ch := '*'
		if rand.Intn(3) == 0 {
			ch = '.'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(200, 210, 255))
		ps.Spawn(rx, ry, (rand.Float64()-0.5)*0.8, 2+rand.Float64()*2, ch, style, 2.0+rand.Float64()*1.0, 0.1)
	}
}

func SpawnDust(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < 2; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + rand.Intn(h))
		ch := '.'
		if rand.Intn(3) == 0 {
			ch = '~'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(180, 160, 120))
		ps.Spawn(rx, ry, 1+rand.Float64()*2, (rand.Float64()-0.5)*0.5, ch, style, 1.5+rand.Float64()*1.0, 0.3)
	}
}

func SpawnEmbers(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < 2; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + rand.Intn(h))
		ch := '.'
		if rand.Intn(3) == 0 {
			ch = '*'
		}
		r := 200 + rand.Intn(55)
		g := 80 + rand.Intn(100)
		style := StyleDefault.Foreground(tcell.NewRGBColor(int32(r), int32(g), 20))
		ps.Spawn(rx, ry, (rand.Float64()-0.5)*1.5, -1.5-rand.Float64()*2, ch, style, 1.0+rand.Float64()*1.0, 0.8)
	}
}
