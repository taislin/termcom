//go:build windows

package audio

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	winmm = syscall.NewLazyDLL("winmm.dll")
	midiOutOpen = winmm.NewProc("midiOutOpen")
	midiOutClose = winmm.NewProc("midiOutClose")
	midiOutShortMsg = winmm.NewProc("midiOutShortMsg")
)

var handle uintptr

func Init() {
	// Open MIDI Mapper (Device ID -1)
	midiOutOpen.Call(uintptr(unsafe.Pointer(&handle)), 0xFFFFFFFF, 0, 0, 0)
}

func Close() {
	midiOutClose.Call(handle)
}

func sendMIDI(msg uint32) {
	midiOutShortMsg.Call(handle, uintptr(msg))
}

func playNote(note byte, velocity byte, channel byte, duration time.Duration) {
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
func PlayAlert() { playNote(60, 100, 0, 500*time.Millisecond) }
