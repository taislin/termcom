//go:build !windows

package audio

import (
	"math"
	mrand "math/rand"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

const (
	sampleRate = 44100
	channels   = 1
	format     = oto.FormatSignedInt16LE
)

var (
	otoCtx   *oto.Context
	otoOnce  sync.Once
	otoReady chan struct{}
	mixer    *mixerStream
)

type mixerStream struct {
	mu      sync.Mutex
	buffers [][]float32
}

func (m *mixerStream) Read(buf []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	samples := len(buf) / 2 // 16-bit mono
	f32 := make([]float32, samples)

	for i := range f32 {
		var sum float32
		for b := len(m.buffers) - 1; b >= 0; b-- {
			if len(m.buffers[b]) > 0 {
				sum += m.buffers[b][0]
				m.buffers[b] = m.buffers[b][1:]
			}
			if len(m.buffers[b]) == 0 {
				m.buffers = append(m.buffers[:b], m.buffers[b+1:]...)
			}
		}
		if sum > 1.0 {
			sum = 1.0
		}
		if sum < -1.0 {
			sum = -1.0
		}
		f32[i] = sum
	}

	for i, s := range f32 {
		v := int16(s * 32767)
		buf[i*2] = byte(v)
		buf[i*2+1] = byte(v >> 8)
	}
	return len(buf), nil
}

func ensureOto() {
	if audioDisabled {
		return
	}
	otoOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				audioDisabled = true
			}
		}()
		otoReady = make(chan struct{})
		mixer = &mixerStream{}
		op := &oto.NewContextOptions{
			SampleRate:   sampleRate,
			ChannelCount: channels,
			Format:       format,
			BufferSize:   time.Duration(40) * time.Millisecond,
		}
		var err error
		otoCtx, _, err = oto.NewContext(op)
		if err != nil {
			audioDisabled = true
			close(otoReady)
			return
		}
		player := otoCtx.NewPlayer(mixer)
		player.Play()
		close(otoReady)
	})
	if otoReady != nil {
		<-otoReady
	}
}

func playPCM(samples []float32) {
	if audioDisabled {
		return
	}
	ensureOto()
	if mixer == nil {
		return
	}
	scaled := make([]float32, len(samples))
	for i, s := range samples {
		scaled[i] = s * float32(sfxVolume)
	}
	mixer.mu.Lock()
	mixer.buffers = append(mixer.buffers, scaled)
	mixer.mu.Unlock()
}

func midiToFreq(note byte) float64 {
	return 440.0 * math.Pow(2.0, float64(note-69)/12.0)
}

func samplesFor(dur float64) int {
	return int(float64(sampleRate) * dur)
}

func generateTone(freq float64, dur time.Duration, waveform func(float64) float64, vol float64) []float32 {
	samples := samplesFor(dur.Seconds())
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		out[i] = float32(waveform(t*freq*2*math.Pi) * vol)
	}
	return out
}

func sine(t float64) float64   { return math.Sin(t) }
func square(t float64) float64 { if math.Sin(t) >= 0 { return 1.0 }; return -1.0 }
func saw(t float64) float64   { return 2.0*(t/(2*math.Pi)-math.Floor(t/(2*math.Pi))) - 1.0 }
func noise() float64          { return mrand.Float64()*2 - 1 }

func envDecay(t, dur float64) float64 {
	return 1.0 - t/dur
}

func freqSweep(startFreq, endFreq float64, t, dur float64) float64 {
	progress := t / dur
	return startFreq + (endFreq-startFreq)*progress
}

type pcmBackend struct{}

var _ Backend = (*pcmBackend)(nil)

func init() { RegisterBackend(&pcmBackend{}) }

func (b *pcmBackend) Init() { ensureOto() }
func (b *pcmBackend) Close() {}

func (b *pcmBackend) Play(s Sound) {
	switch s {
	case SoundClick:
		playPCM(generateTone(midiToFreq(70), 50*time.Millisecond, sine, 0.3))
	case SoundSelect:
		playPCM(generateTone(midiToFreq(65), 30*time.Millisecond, sine, 0.25))
	case SoundMove:
		playPCM(generateTone(midiToFreq(60), 20*time.Millisecond, sine, 0.15))
	case SoundMenuNav:
		playPCM(generateTone(midiToFreq(72), 20*time.Millisecond, sine, 0.15))
	case SoundChime:
		playPCM(generateTone(midiToFreq(72), 200*time.Millisecond, sine, 0.3))
	case SoundReload:
		s1 := generateTone(midiToFreq(55), 30*time.Millisecond, sine, 0.25)
		s2 := generateTone(midiToFreq(60), 30*time.Millisecond, sine, 0.25)
		pad := make([]float32, samplesFor(0.01))
		s1 = append(s1, pad...)
		s1 = append(s1, s2...)
		playPCM(s1)
	case SoundShoot:
		b.noiseBurst(0.1, 0.4, 250)
	case SoundBallisticFire:
		b.noiseBurst(0.1, 0.4, 250)
	case SoundLaserFire:
		samples := samplesFor(0.1)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			freq := freqSweep(2000, 800, t, 0.1)
			vol := envDecay(t, 0.1) * 0.3
			out[i] = float32(sine(t*freq*2*math.Pi) * vol)
		}
		playPCM(out)
	case SoundPlasmaFire:
		samples := samplesFor(0.14)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			freq := freqSweep(150, 60, t, 0.14)
			vol := envDecay(t, 0.14) * 0.4
			out[i] = float32((noise()*0.5 + square(t*freq*2*math.Pi)*0.5) * vol)
		}
		playPCM(out)
	case SoundMeleeFire:
		samples := samplesFor(0.09)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			vol := envDecay(t, 0.09) * 0.35
			out[i] = float32((noise()*0.4 + saw(t*200*2*math.Pi)*0.6) * vol)
		}
		playPCM(out)
	case SoundHit:
		s1 := generateTone(midiToFreq(50), 50*time.Millisecond, square, 0.3)
		pad := make([]float32, samplesFor(0.01))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(55), 50*time.Millisecond, square, 0.3)
		s1 = append(s1, s2...)
		playPCM(s1)
	case SoundMiss:
		b.noiseBurst(0.08, 0.2, 0)
	case SoundExplosion:
		b.noiseBurst(0.35, 0.5, 40)
	case SoundGrenade:
		samples := samplesFor(0.35)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			freq := freqSweep(120, 30, t, 0.35)
			vol := envDecay(t, 0.35) * 0.5
			out[i] = float32((noise()*0.6 + square(t*freq*2*math.Pi)*0.4) * vol)
		}
		playPCM(out)
	case SoundDistantExplosion:
		samples := samplesFor(0.5)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			vol := envDecay(t, 0.5) * 0.2
			out[i] = float32((noise()*0.8 + sine(t*30*2*math.Pi)*0.2) * vol)
		}
		playPCM(out)
	case SoundAlert:
		b.sequence([]byte{72, 60, 72, 60}, 250*time.Millisecond, sine, 0.3)
	case SoundAlienTurn:
		s1 := generateTone(midiToFreq(45), 100*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(40), 100*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		playPCM(s1)
	case SoundVictory:
		b.sequence([]byte{60, 64, 67, 72}, 200*time.Millisecond, sine, 0.3)
	case SoundDefeat:
		b.sequence([]byte{60, 55, 50, 45}, 250*time.Millisecond, sine, 0.3)
	case SoundMedikit:
		s1 := generateTone(midiToFreq(67), 100*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(72), 100*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		playPCM(s1)
	case SoundResearchComplete:
		b.sequence([]byte{60, 64, 67, 72, 76}, 150*time.Millisecond, sine, 0.25)
	case SoundManufactureComplete:
		s1 := generateTone(midiToFreq(69), 80*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(76), 120*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		playPCM(s1)
	case SoundUFODetected:
		b.sequence([]byte{72, 69, 72, 69}, 200*time.Millisecond, sine, 0.35)
	case SoundMissionWarning:
		b.sequence([]byte{74, 72, 74, 72}, 150*time.Millisecond, square, 0.3)
	case SoundWind:
		samples := samplesFor(0.6)
		out := make([]float32, samples)
		for i := range out {
			t := float64(i) / float64(sampleRate)
			vol := 0.08 * (1.0 - t/0.6)
			out[i] = float32(noise() * vol)
		}
		playPCM(out)
	}
}

// noiseBurst synthesizes a short percussive noise hit with optional low tone.
func (b *pcmBackend) noiseBurst(dur float64, vol, lowHz float64) {
	samples := samplesFor(dur)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		v := envDecay(t, dur) * vol
		n := noise() * 0.7
		if lowHz > 0 {
			n += square(t*lowHz*2*math.Pi) * 0.3
		}
		out[i] = float32(n * v)
	}
	playPCM(out)
}

// sequence plays a series of tones spaced by the given gap.
func (b *pcmBackend) sequence(notes []byte, gap time.Duration, wave func(float64) float64, vol float64) {
	var combined []float32
	pad := make([]float32, samplesFor(gap.Seconds()*0.2))
	for _, n := range notes {
		tone := generateTone(midiToFreq(n), gap, wave, vol)
		combined = append(combined, tone...)
		combined = append(combined, pad...)
	}
	playPCM(combined)
}
