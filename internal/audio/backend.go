package audio

// Sound identifies a discrete sound effect produced by the game. It is the
// shared vocabulary between all platform backends so they cannot drift apart.
type Sound int

const (
	SoundClick Sound = iota
	SoundSelect
	SoundMove
	SoundMenuNav
	SoundChime
	SoundReload
	SoundShoot
	SoundBallisticFire
	SoundLaserFire
	SoundPlasmaFire
	SoundMeleeFire
	SoundHit
	SoundMiss
	SoundExplosion
	SoundGrenade
	SoundDistantExplosion
	SoundAlert
	SoundAlienTurn
	SoundVictory
	SoundDefeat
	SoundMedikit
	SoundResearchComplete
	SoundManufactureComplete
	SoundUFODetected
	SoundMissionWarning
	SoundWind

	soundCount
)

// Backend is the cross-platform audio abstraction. Each platform provides an
// implementation: Windows uses MIDI (winmm.dll), other platforms use PCM
// synthesis via oto. Both backends consume the same Sound vocabulary.
type Backend interface {
	Init()
	Close()
	Play(s Sound)
}

var active Backend

// RegisterBackend installs the platform-specific backend. It is called from
// each platform file's init so the dispatcher always has a target.
func RegisterBackend(b Backend) {
	if active == nil {
		active = b
	}
}

func play(s Sound) {
	if active != nil {
		active.Play(s)
	}
}

// Init wires up the active backend. It is safe to call before any Play.
func Init() {
	if active != nil {
		active.Init()
	}
}

// Close releases backend resources.
func Close() {
	if active != nil {
		active.Close()
	}
}
