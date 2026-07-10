package audio

func PlayWeaponFire(weapon string) {
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
