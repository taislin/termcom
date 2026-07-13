package audio

var audioDisabled bool
var sfxVolume float64 = 1.0

func SetAudioEnabled(enabled bool) {
	audioDisabled = !enabled
}

func SetSfxVolume(vol int) {
	if vol < 0 {
		vol = 0
	}
	if vol > 10 {
		vol = 10
	}
	sfxVolume = float64(vol) / 10.0
}

func PlayWeaponFire(weapon string) {
	if audioDisabled {
		return
	}
	switch weapon {
	case "laser_pistol", "laser_rifle":
		PlayLaserFire()
	case "plasma_pistol", "plasma_rifle", "heavy_plasma", "alien_grenade":
		PlayPlasmaFire()
	case "rocket":
		PlayExplosion()
	case "chryssalid_claw", "reaper_claw", "stun_rod":
		PlayMeleeFire()
	default:
		PlayShoot()
	}
}
