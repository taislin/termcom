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
		play(SoundLaserFire)
	case "plasma_pistol", "plasma_rifle", "heavy_plasma", "alien_grenade":
		play(SoundPlasmaFire)
	case "rocket":
		play(SoundExplosion)
	case "chryssalid_claw", "reaper_claw", "stun_rod":
		play(SoundMeleeFire)
	default:
		play(SoundShoot)
	}
}

func PlayClick()               { play(SoundClick) }
func PlaySelect()              { play(SoundSelect) }
func PlayMove()                { play(SoundMove) }
func PlayMenuNav()             { play(SoundMenuNav) }
func PlayChime()               { play(SoundChime) }
func PlayReload()              { play(SoundReload) }
func PlayShoot()               { play(SoundShoot) }
func PlayBallisticFire()       { play(SoundBallisticFire) }
func PlayLaserFire()           { play(SoundLaserFire) }
func PlayPlasmaFire()          { play(SoundPlasmaFire) }
func PlayMeleeFire()           { play(SoundMeleeFire) }
func PlayHit()                 { play(SoundHit) }
func PlayMiss()                { play(SoundMiss) }
func PlayExplosion()           { play(SoundExplosion) }
func PlayGrenade()             { play(SoundGrenade) }
func PlayDistantExplosion()    { play(SoundDistantExplosion) }
func PlayAlert()               { play(SoundAlert) }
func PlayAlienTurn()           { play(SoundAlienTurn) }
func PlayVictory()             { play(SoundVictory) }
func PlayDefeat()              { play(SoundDefeat) }
func PlayMedikit()             { play(SoundMedikit) }
func PlayResearchComplete()    { play(SoundResearchComplete) }
func PlayManufactureComplete() { play(SoundManufactureComplete) }
func PlayUFODetected()         { play(SoundUFODetected) }
func PlayMissionWarning()      { play(SoundMissionWarning) }
func PlayWind()                { play(SoundWind) }
