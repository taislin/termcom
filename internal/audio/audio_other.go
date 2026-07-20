//go:build !windows

package audio

import (
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

const (
	channels = 1
	format   = oto.FormatSignedInt16LE

	// otoBufferMS is the output buffer size for the Oto PCM stream.
	otoBufferMS = 40
	// int16Max is the largest magnitude of a signed 16-bit PCM sample.
	int16Max = 32767
)

var (
	otoCtx   *oto.Context
	otoOnce  sync.Once
	otoReady chan struct{}
	mixer    *mixerStream
	otoPlayer *oto.Player
)

type mixerStream struct {
	mu      sync.Mutex
	buffers [][]float32
}

func (m *mixerStream) Read(buf []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	samples := len(buf) / 2
	f32 := make([]float32, samples)

	for i := range f32 {
		var sum float32
		for b := len(m.buffers) - 1; b >= 0; b-- {
			if len(m.buffers[b]) > 0 {
				sum += m.buffers[b][0]
				m.buffers[b] = m.buffers[b][1:]
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

	// Drop buffers that have been fully consumed. Compacting after the sample
	// loop (rather than while iterating backwards) avoids skipping a neighbour
	// when an element is removed.
	remaining := m.buffers[:0]
	for _, buf := range m.buffers {
		if len(buf) > 0 {
			remaining = append(remaining, buf)
		}
	}
	m.buffers = remaining

	for i, s := range f32 {
		v := int16(s * int16Max)
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
				if otoReady != nil {
					close(otoReady)
				}
			}
		}()
		otoReady = make(chan struct{})
		mixer = &mixerStream{}
		op := &oto.NewContextOptions{
			SampleRate:   sampleRate,
			ChannelCount: channels,
			Format:       format,
			BufferSize:   time.Duration(otoBufferMS) * time.Millisecond,
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
		otoPlayer = player
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

type pcmBackend struct{}

var _ Backend = (*pcmBackend)(nil)

func init() { RegisterBackend(&pcmBackend{}) }

func (b *pcmBackend) Init() { ensureOto() }
func (b *pcmBackend) Close() {
	if otoPlayer != nil {
		otoPlayer.Close()
	}
	if otoCtx != nil {
		otoCtx.Close()
	}
}

func (b *pcmBackend) Play(s Sound) {
	pcm := soundSamples(s)
	if pcm != nil {
		playPCM(pcm)
	}
}
