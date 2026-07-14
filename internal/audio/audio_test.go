package audio

import "testing"

// TestAllSoundsPlayable ensures the active backend handles every Sound value
// without panicking. Both platform backends share the same Sound vocabulary
// (see backend.go) and must keep their Play switches in sync.
func TestAllSoundsPlayable(t *testing.T) {
	if active == nil {
		t.Fatal("no audio backend registered")
	}
	for s := Sound(0); s < soundCount; s++ {
		active.Play(s)
	}
}

// TestSoundEnumStable guards the Sound vocabulary: adding a new sound must
// update this count so backend switches and tests stay aligned.
func TestSoundEnumStable(t *testing.T) {
	const expectedSounds = 26
	if soundCount != expectedSounds {
		t.Errorf("soundCount = %d, expected %d (update backends + this test when adding sounds)", soundCount, expectedSounds)
	}
}


