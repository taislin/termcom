package audio

var audioDisabled bool

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
