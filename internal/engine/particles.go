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

const (
	explosionUpBias   = 4.0
	explosionSpeedMin = 2.0
	explosionSpeedRng = 6.0
	explosionLifeMin  = 0.4
	explosionLifeRng  = 0.6
	explosionFade     = 0.8

	rainCount    = 3
	rainVX       = -0.5
	rainVY       = 12.0
	rainLife     = 1.5
	rainFade     = 0.1
	rainStartRow = -1

	smokeVXRng   = 2.0
	smokeVYMin   = -1.0
	smokeVYRng   = 3.0
	smokeLifeMin = 0.8
	smokeLifeRng = 1.2
	smokeFade    = 0.5

	menuDriftVXMin   = 0.5
	menuDriftVXRng   = 1.0
	menuDriftVYMin   = 6.0
	menuDriftVYRng   = 4.0
	menuDriftLifeMin = 1.0
	menuDriftLifeRng = 1.0
	menuDriftFade    = 0.6

	muzzleCount         = 6
	muzzleSpeedMin      = 1.0
	muzzleSpeedRng      = 3.0
	muzzleLifeMin       = 0.15
	muzzleLifeRng       = 0.2
	muzzleFade          = 2.0
	muzzleBrightnessMin = 200
	muzzleBrightnessRng = 55

	snowCount    = 2
	snowVXRng    = 0.8
	snowVYMin    = 2.0
	snowVYRng    = 2.0
	snowLifeMin  = 2.0
	snowLifeRng  = 1.0
	snowFade     = 0.1
	snowStartRow = -1

	dustCount   = 2
	dustVXMin   = 1.0
	dustVXRng   = 2.0
	dustVYRng   = 0.5
	dustLifeMin = 1.5
	dustLifeRng = 1.0
	dustFade    = 0.3

	emberCount   = 2
	emberVXRng   = 1.5
	emberVYMin   = 1.5
	emberVYRng   = 2.0
	emberLifeMin = 1.0
	emberLifeRng = 1.0
	emberFade    = 0.8
	emberRMin    = 200
	emberRRng    = 55
	emberGMin    = 80
	emberGRng    = 100
	emberB       = 20
)

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

type explosionParams struct {
	vx, vy float64
	ch     rune
	life   float64
	fade   float64
}

func randomExplosionParticle() explosionParams {
	angle := rand.Float64() * 2 * math.Pi
	speed := explosionSpeedMin + rand.Float64()*explosionSpeedRng
	ch := '*'
	if rand.Intn(2) == 0 {
		ch = '+'
	}
	return explosionParams{
		vx:   math.Cos(angle) * speed,
		vy:   math.Sin(angle)*speed - explosionUpBias,
		ch:   ch,
		life: explosionLifeMin + rand.Float64()*explosionLifeRng,
		fade: explosionFade,
	}
}

func SpawnExplosion(ps *ParticleSystem, x, y int, color tcell.Color, count int) {
	r, g, b := colorRGB(color)
	fg := tcell.NewRGBColor(int32(r), int32(g), int32(b))

	for i := 0; i < count; i++ {
		p := randomExplosionParticle()
		style := StyleDefault.Foreground(fg)
		ps.Spawn(float64(x), float64(y), p.vx, p.vy, p.ch, style, p.life, p.fade)
	}
}

func SpawnRain(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < rainCount; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + rainStartRow)
		style := StyleDefault.Foreground(tcell.NewRGBColor(100, 150, 255))
		ps.Spawn(rx, ry, rainVX, rainVY, '|', style, rainLife, rainFade)
	}
}

func SpawnSmoke(ps *ParticleSystem, x, y int, count int) {
	for i := 0; i < count; i++ {
		vx := (rand.Float64() - 0.5) * smokeVXRng
		vy := smokeVYMin - rand.Float64()*smokeVYRng
		ch := '~'
		if rand.Intn(3) == 0 {
			ch = ':'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(128, 128, 128))
		ps.Spawn(float64(x)+rand.Float64()*2-1, float64(y), vx, vy, ch, style, smokeLifeMin+rand.Float64()*smokeLifeRng, smokeFade)
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
	vx := float64(side) * (menuDriftVXMin + rand.Float64()*menuDriftVXRng)
	vy := -(menuDriftVYMin + rand.Float64()*menuDriftVYRng)
	life := menuDriftLifeMin + rand.Float64()*menuDriftLifeRng
	ps.Spawn(float64(x), float64(y), vx, vy, ch, style, life, menuDriftFade)
}

func SpawnMuzzleFlash(ps *ParticleSystem, x, y int) {
	flashRunes := []rune{'.', '*', '°', '+'}
	for i := 0; i < muzzleCount; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := muzzleSpeedMin + rand.Float64()*muzzleSpeedRng
		vx := math.Cos(angle) * speed
		vy := math.Sin(angle) * speed
		ch := flashRunes[rand.Intn(len(flashRunes))]
		brightness := muzzleBrightnessMin + rand.Intn(muzzleBrightnessRng)
		fg := tcell.NewRGBColor(int32(brightness), int32(brightness-30), int32(50+rand.Intn(80)))
		style := StyleDefault.Foreground(fg)
		ps.Spawn(float64(x), float64(y), vx, vy, ch, style, muzzleLifeMin+rand.Float64()*muzzleLifeRng, muzzleFade)
	}
}

func SpawnSnow(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < snowCount; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + snowStartRow)
		ch := '*'
		if rand.Intn(3) == 0 {
			ch = '.'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(200, 210, 255))
		ps.Spawn(rx, ry, (rand.Float64()-0.5)*snowVXRng, snowVYMin+rand.Float64()*snowVYRng, ch, style, snowLifeMin+rand.Float64()*snowLifeRng, snowFade)
	}
}

func SpawnDust(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < dustCount; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + rand.Intn(h))
		ch := '.'
		if rand.Intn(3) == 0 {
			ch = '~'
		}
		style := StyleDefault.Foreground(tcell.NewRGBColor(180, 160, 120))
		ps.Spawn(rx, ry, dustVXMin+rand.Float64()*dustVXRng, (rand.Float64()-0.5)*dustVYRng, ch, style, dustLifeMin+rand.Float64()*dustLifeRng, dustFade)
	}
}

func SpawnEmbers(ps *ParticleSystem, x, y int, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	for i := 0; i < emberCount; i++ {
		rx := float64(x + rand.Intn(w))
		ry := float64(y + rand.Intn(h))
		ch := '.'
		if rand.Intn(3) == 0 {
			ch = '*'
		}
		r := emberRMin + rand.Intn(emberRRng)
		g := emberGMin + rand.Intn(emberGRng)
		style := StyleDefault.Foreground(tcell.NewRGBColor(int32(r), int32(g), emberB))
		ps.Spawn(rx, ry, (rand.Float64()-0.5)*emberVXRng, -emberVYMin-rand.Float64()*emberVYRng, ch, style, emberLifeMin+rand.Float64()*emberLifeRng, emberFade)
	}
}
