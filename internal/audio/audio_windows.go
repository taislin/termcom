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
		// Open MIDI Mapper (Device ID -1)
		_, _, err := midiOutOpen.Call(uintptr(unsafe.Pointer(&handle)), 0xFFFFFFFF, 0, 0, 0)
		if err != nil && err.Error() != "The operation completed successfully." {
			audioDisabled = true
		}
	})
}

func Init() { ensureMIDI() }

func Close() {
	if audioDisabled {
		return
	}
	midiOutClose.Call(handle)
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
	// Note On
	msgOn := uint32(0x90|channel) | (uint32(note) << 8) | (uint32(velocity) << 16)
	sendMIDI(msgOn)

	// Note Off
	go func() {
		time.Sleep(duration)
		msgOff := uint32(0x80|channel) | (uint32(note) << 8) | (uint32(0) << 16)
		sendMIDI(msgOff)
	}()
}

func PlayClick() { playNote(70, 100, 0, 50*time.Millisecond) }
func PlayShoot() { playNote(38, 120, 9, 100*time.Millisecond) }
func PlayExplosion() { playNote(35, 127, 9, 300*time.Millisecond) }
func PlayChime() { playNote(72, 100, 0, 200*time.Millisecond) }

// PlayAlert - 4 note siren: high-low-high-low, 0.25s per note
func PlayAlert() {
	notes := []byte{72, 60, 72, 60} // C5, C4, C5, C4
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 250 * time.Millisecond)
			playNote(n, 100, 0, 250*time.Millisecond)
		}(note, i)
	}
}

func PlayHit() {
	playNote(50, 100, 9, 50*time.Millisecond)
	go func() {
		time.Sleep(60 * time.Millisecond)
		playNote(55, 100, 9, 50*time.Millisecond)
	}()
}

func PlayMiss() { playNote(40, 80, 9, 80*time.Millisecond) }

func PlayAlienTurn() {
	playNote(45, 100, 0, 100*time.Millisecond)
	go func() {
		time.Sleep(120 * time.Millisecond)
		playNote(40, 100, 0, 100*time.Millisecond)
	}()
}

func PlayVictory() {
	notes := []byte{60, 64, 67, 72}
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 150 * time.Millisecond)
			playNote(n, 100, 0, 200*time.Millisecond)
		}(note, i)
	}
}

func PlayDefeat() {
	notes := []byte{60, 55, 50, 45}
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 200 * time.Millisecond)
			playNote(n, 100, 0, 250*time.Millisecond)
		}(note, i)
	}
}

func PlayGrenade() {
	playNote(35, 127, 9, 150*time.Millisecond)
	go func() {
		time.Sleep(160 * time.Millisecond)
		playNote(30, 127, 9, 200*time.Millisecond)
	}()
}

func PlaySelect() { playNote(65, 80, 0, 30*time.Millisecond) }

func PlayMove() { playNote(60, 60, 0, 20*time.Millisecond) }

func PlayReload() {
	playNote(55, 80, 0, 30*time.Millisecond)
	go func() {
		time.Sleep(40 * time.Millisecond)
		playNote(60, 80, 0, 30*time.Millisecond)
	}()
}

func PlayMedikit() {
	playNote(67, 100, 0, 100*time.Millisecond)
	go func() {
		time.Sleep(120 * time.Millisecond)
		playNote(72, 100, 0, 100*time.Millisecond)
	}()
}

func PlayLaserFire() {
	playNote(80, 100, 0, 60*time.Millisecond)
	go func() {
		time.Sleep(20 * time.Millisecond)
		playNote(84, 80, 0, 40*time.Millisecond)
	}()
}

func PlayPlasmaFire() {
	playNote(38, 120, 9, 80*time.Millisecond)
	go func() {
		time.Sleep(40 * time.Millisecond)
		playNote(42, 100, 9, 60*time.Millisecond)
	}()
}

func PlayMeleeFire() {
	playNote(48, 110, 9, 40*time.Millisecond)
	go func() {
		time.Sleep(30 * time.Millisecond)
		playNote(44, 120, 9, 50*time.Millisecond)
	}()
}

func PlayBallisticFire() {
	playNote(42, 120, 9, 50*time.Millisecond)
}

func PlayMenuNav() {
	playNote(72, 60, 0, 20*time.Millisecond)
}

func PlayResearchComplete() {
	notes := []byte{60, 64, 67, 72, 76}
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 100 * time.Millisecond)
			playNote(n, 90, 0, 150*time.Millisecond)
		}(note, i)
	}
}

func PlayManufactureComplete() {
	playNote(69, 100, 0, 80*time.Millisecond)
	go func() {
		time.Sleep(100 * time.Millisecond)
		playNote(76, 100, 0, 120*time.Millisecond)
	}()
}

func PlayUFODetected() {
	notes := []byte{72, 69, 72, 69}
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 200 * time.Millisecond)
			playNote(n, 110, 0, 200*time.Millisecond)
		}(note, i)
	}
}

func PlayMissionWarning() {
	notes := []byte{74, 72, 74, 72}
	for i, note := range notes {
		go func(n byte, delay int) {
			time.Sleep(time.Duration(delay) * 150 * time.Millisecond)
			playNote(n, 110, 0, 150*time.Millisecond)
		}(note, i)
	}
}

func PlayDistantExplosion() {
	playNote(28, 80, 9, 400*time.Millisecond)
}

func PlayWind() {
	playNote(74, 30, 0, 500*time.Millisecond)
}
