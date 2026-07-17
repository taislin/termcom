//go:build windows

package audio

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	winmm           = syscall.NewLazyDLL("winmm.dll")
	midiOutOpen     = winmm.NewProc("midiOutOpen")
	midiOutClose    = winmm.NewProc("midiOutClose")
	midiOutShortMsg = winmm.NewProc("midiOutShortMsg")
)

var (
	handle   uintptr
	midiOnce sync.Once
)

// MIDI message status bytes and constants.
const (
	midiNoteOn   = 0x90 // status byte: note on (channel in low nibble)
	midiNoteOff  = 0x80 // status byte: note off (channel in low nibble)
	midiPercCh   = 9    // General MIDI percussion channel
	midiMinVol   = 10   // minimum note velocity so quiet sounds remain audible
	midiMapperID = 0xFFFFFFFF
)

func ensureMIDI() {
	if audioDisabled {
		return
	}
	midiOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				audioDisabled = true
			}
		}()
		// Open MIDI Mapper (Device ID -1).
		// Call returns (r1, r2, err); on Windows r1 is the HRESULT. A zero
		// result (MMSYSERR_NOERROR) indicates success, regardless of err text.
		r1, _, _ := midiOutOpen.Call(uintptr(unsafe.Pointer(&handle)), midiMapperID, 0, 0, 0)
		if r1 != 0 {
			audioDisabled = true
		}
	})
}

type midiBackend struct{}

var _ Backend = (*midiBackend)(nil)

func init() { RegisterBackend(&midiBackend{}) }

func (b *midiBackend) Init() { ensureMIDI() }

func (b *midiBackend) Close() {
	if audioDisabled {
		return
	}
	midiOutClose.Call(handle)
	audioDisabled = true
}

func sendMIDI(msg uint32) {
	if audioDisabled {
		return
	}
	midiOutShortMsg.Call(handle, uintptr(msg))
}

func playNote(note byte, velocity byte, channel byte, duration time.Duration) {
	if audioDisabled {
		return
	}
	ensureMIDI()
	vol := byte(float64(velocity) * sfxVolume)
	if vol < midiMinVol {
		vol = midiMinVol
	}
	// Note On
	msgOn := uint32(midiNoteOn|channel) | (uint32(note) << 8) | (uint32(vol) << 16)
	sendMIDI(msgOn)

	// Note Off
	go func() {
		time.Sleep(duration)
		msgOff := uint32(midiNoteOff|channel) | (uint32(note) << 8) | (uint32(0) << 16)
		sendMIDI(msgOff)
	}()
}

func (b *midiBackend) Play(s Sound) {
	switch s {
	case SoundClick:
		playNote(70, 100, 0, 50*time.Millisecond)
	case SoundSelect:
		playNote(65, 80, 0, 30*time.Millisecond)
	case SoundMove:
		playNote(60, 60, 0, 20*time.Millisecond)
	case SoundMenuNav:
		playNote(72, 60, 0, 20*time.Millisecond)
	case SoundChime:
		playNote(72, 100, 0, 200*time.Millisecond)
	case SoundReload:
		playNote(55, 80, 0, 30*time.Millisecond)
		go func() {
			time.Sleep(40 * time.Millisecond)
			playNote(60, 80, 0, 30*time.Millisecond)
		}()
	case SoundShoot:
		playNote(38, 120, midiPercCh, 100*time.Millisecond)
	case SoundBallisticFire:
		playNote(42, 120, midiPercCh, 50*time.Millisecond)
	case SoundLaserFire:
		playNote(80, 100, 0, 60*time.Millisecond)
		go func() {
			time.Sleep(20 * time.Millisecond)
			playNote(84, 80, 0, 40*time.Millisecond)
		}()
	case SoundPlasmaFire:
		playNote(38, 120, midiPercCh, 80*time.Millisecond)
		go func() {
			time.Sleep(40 * time.Millisecond)
			playNote(42, 100, midiPercCh, 60*time.Millisecond)
		}()
	case SoundMeleeFire:
		playNote(48, 110, midiPercCh, 40*time.Millisecond)
		go func() {
			time.Sleep(30 * time.Millisecond)
			playNote(44, 120, midiPercCh, 50*time.Millisecond)
		}()
	case SoundHit:
		playNote(50, 100, midiPercCh, 50*time.Millisecond)
		go func() {
			time.Sleep(60 * time.Millisecond)
			playNote(55, 100, midiPercCh, 50*time.Millisecond)
		}()
	case SoundMiss:
		playNote(40, 80, midiPercCh, 80*time.Millisecond)
	case SoundExplosion:
		playNote(35, 127, midiPercCh, 300*time.Millisecond)
	case SoundGrenade:
		playNote(35, 127, midiPercCh, 150*time.Millisecond)
		go func() {
			time.Sleep(160 * time.Millisecond)
			playNote(30, 127, midiPercCh, 200*time.Millisecond)
		}()
	case SoundDistantExplosion:
		playNote(28, 80, midiPercCh, 400*time.Millisecond)
	case SoundAlert:
		notes := []byte{72, 60, 72, 60}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 250 * time.Millisecond)
				playNote(n, 100, 0, 250*time.Millisecond)
			}(note, i)
		}
	case SoundAlienTurn:
		playNote(45, 100, 0, 100*time.Millisecond)
		go func() {
			time.Sleep(120 * time.Millisecond)
			playNote(40, 100, 0, 100*time.Millisecond)
		}()
	case SoundVictory:
		notes := []byte{60, 64, 67, 72}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 150 * time.Millisecond)
				playNote(n, 100, 0, 200*time.Millisecond)
			}(note, i)
		}
	case SoundDefeat:
		notes := []byte{60, 55, 50, 45}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 200 * time.Millisecond)
				playNote(n, 100, 0, 250*time.Millisecond)
			}(note, i)
		}
	case SoundMedikit:
		playNote(67, 100, 0, 100*time.Millisecond)
		go func() {
			time.Sleep(120 * time.Millisecond)
			playNote(72, 100, 0, 100*time.Millisecond)
		}()
	case SoundResearchComplete:
		notes := []byte{60, 64, 67, 72, 76}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 100 * time.Millisecond)
				playNote(n, 90, 0, 150*time.Millisecond)
			}(note, i)
		}
	case SoundManufactureComplete:
		playNote(69, 100, 0, 80*time.Millisecond)
		go func() {
			time.Sleep(100 * time.Millisecond)
			playNote(76, 100, 0, 120*time.Millisecond)
		}()
	case SoundUFODetected:
		notes := []byte{72, 69, 72, 69}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 200 * time.Millisecond)
				playNote(n, 110, 0, 200*time.Millisecond)
			}(note, i)
		}
	case SoundMissionWarning:
		notes := []byte{74, 72, 74, 72}
		for i, note := range notes {
			go func(n byte, delay int) {
				time.Sleep(time.Duration(delay) * 150 * time.Millisecond)
				playNote(n, 110, 0, 150*time.Millisecond)
			}(note, i)
		}
	case SoundWind:
		playNote(74, 30, 0, 500*time.Millisecond)
	}
}
