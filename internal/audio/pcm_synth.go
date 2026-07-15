//go:build !windows

package audio

import (
	"math"
	mrand "math/rand"
	"time"
)

const sampleRate = 44100

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

// noiseBurst synthesizes a short percussive noise hit with optional low tone.
func noiseBurst(dur float64, vol, lowHz float64) []float32 {
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
	return out
}

// sequence plays a series of tones spaced by the given gap.
func sequence(notes []byte, gap time.Duration, wave func(float64) float64, vol float64) []float32 {
	var combined []float32
	pad := make([]float32, samplesFor(gap.Seconds()*0.2))
	for _, n := range notes {
		tone := generateTone(midiToFreq(n), gap, wave, vol)
		combined = append(combined, tone...)
		combined = append(combined, pad...)
	}
	return combined
}

// soundSamples renders a Sound identifier to PCM samples.
func soundSamples(s Sound) []float32 {
	switch s {
	case SoundClick:
		return generateTone(midiToFreq(70), 50*time.Millisecond, sine, 0.3)
	case SoundSelect:
		return generateTone(midiToFreq(65), 30*time.Millisecond, sine, 0.25)
	case SoundMove:
		return generateTone(midiToFreq(60), 20*time.Millisecond, sine, 0.15)
	case SoundMenuNav:
		return generateTone(midiToFreq(72), 20*time.Millisecond, sine, 0.15)
	case SoundChime:
		return generateTone(midiToFreq(72), 200*time.Millisecond, sine, 0.3)
	case SoundReload:
		s1 := generateTone(midiToFreq(55), 30*time.Millisecond, sine, 0.25)
		s2 := generateTone(midiToFreq(60), 30*time.Millisecond, sine, 0.25)
		pad := make([]float32, samplesFor(0.01))
		s1 = append(s1, pad...)
		s1 = append(s1, s2...)
		return s1
	case SoundShoot:
		return noiseBurst(0.1, 0.4, 250)
	case SoundBallisticFire:
		return noiseBurst(0.1, 0.4, 250)
	case SoundLaserFire:
		return laserSamples()
	case SoundPlasmaFire:
		return plasmaSamples()
	case SoundMeleeFire:
		return meleeSamples()
	case SoundHit:
		s1 := generateTone(midiToFreq(50), 50*time.Millisecond, square, 0.3)
		pad := make([]float32, samplesFor(0.01))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(55), 50*time.Millisecond, square, 0.3)
		s1 = append(s1, s2...)
		return s1
	case SoundMiss:
		return noiseBurst(0.08, 0.2, 0)
	case SoundExplosion:
		return noiseBurst(0.35, 0.5, 40)
	case SoundGrenade:
		return grenadeSamples()
	case SoundDistantExplosion:
		return distantExplosionSamples()
	case SoundAlert:
		return sequence([]byte{72, 60, 72, 60}, 250*time.Millisecond, sine, 0.3)
	case SoundAlienTurn:
		s1 := generateTone(midiToFreq(45), 100*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(40), 100*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		return s1
	case SoundVictory:
		return sequence([]byte{60, 64, 67, 72}, 200*time.Millisecond, sine, 0.3)
	case SoundDefeat:
		return sequence([]byte{60, 55, 50, 45}, 250*time.Millisecond, sine, 0.3)
	case SoundMedikit:
		s1 := generateTone(midiToFreq(67), 100*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(72), 100*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		return s1
	case SoundResearchComplete:
		return sequence([]byte{60, 64, 67, 72, 76}, 150*time.Millisecond, sine, 0.25)
	case SoundManufactureComplete:
		s1 := generateTone(midiToFreq(69), 80*time.Millisecond, sine, 0.3)
		pad := make([]float32, samplesFor(0.02))
		s1 = append(s1, pad...)
		s2 := generateTone(midiToFreq(76), 120*time.Millisecond, sine, 0.3)
		s1 = append(s1, s2...)
		return s1
	case SoundUFODetected:
		return sequence([]byte{72, 69, 72, 69}, 200*time.Millisecond, sine, 0.35)
	case SoundMissionWarning:
		return sequence([]byte{74, 72, 74, 72}, 150*time.Millisecond, square, 0.3)
	case SoundWind:
		return windSamples()
	}
	return nil
}

func laserSamples() []float32 {
	samples := samplesFor(0.1)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := freqSweep(2000, 800, t, 0.1)
		vol := envDecay(t, 0.1) * 0.3
		out[i] = float32(sine(t*freq*2*math.Pi) * vol)
	}
	return out
}

func plasmaSamples() []float32 {
	samples := samplesFor(0.14)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := freqSweep(150, 60, t, 0.14)
		vol := envDecay(t, 0.14) * 0.4
		out[i] = float32((noise()*0.5 + square(t*freq*2*math.Pi)*0.5) * vol)
	}
	return out
}

func meleeSamples() []float32 {
	samples := samplesFor(0.09)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		vol := envDecay(t, 0.09) * 0.35
		out[i] = float32((noise()*0.4 + saw(t*200*2*math.Pi)*0.6) * vol)
	}
	return out
}

func grenadeSamples() []float32 {
	samples := samplesFor(0.35)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := freqSweep(120, 30, t, 0.35)
		vol := envDecay(t, 0.35) * 0.5
		out[i] = float32((noise()*0.6 + square(t*freq*2*math.Pi)*0.4) * vol)
	}
	return out
}

func distantExplosionSamples() []float32 {
	samples := samplesFor(0.5)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		vol := envDecay(t, 0.5) * 0.2
		out[i] = float32((noise()*0.8 + sine(t*30*2*math.Pi)*0.2) * vol)
	}
	return out
}

func windSamples() []float32 {
	samples := samplesFor(0.6)
	out := make([]float32, samples)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		vol := 0.08 * (1.0 - t/0.6)
		out[i] = float32(noise() * vol)
	}
	return out
}
