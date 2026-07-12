package data

// AlienEquipTier defines weapon/armor upgrades aliens receive as the
// campaign progresses.  Each tier is gated by a minimum game month.
type AlienEquipTier struct {
	MinMonth   int      // first month this tier is active
	Weapons    []string // weapon pool (picked per alien by rank)
	ArmorBonus int      // extra armour applied to every alien
	HPBonus    int      // extra HP applied to every alien
}

// AlienEquipTiers defines the escalation schedule. Higher tiers give aliens
// access to more dangerous weapons and small stat bumps.
var AlienEquipTiers = []AlienEquipTier{
	{
		MinMonth: 0,
		Weapons:  []string{"plasma_pistol", "plasma_pistol", "plasma_rifle"},
	},
	{
		MinMonth:   3,
		Weapons:    []string{"plasma_rifle", "plasma_rifle", "plasma_pistol", "heavy_plasma"},
		ArmorBonus: 2,
		HPBonus:    2,
	},
	{
		MinMonth:   6,
		Weapons:    []string{"plasma_rifle", "heavy_plasma", "heavy_plasma", "alien_cannon", "alien_laser"},
		ArmorBonus: 4,
		HPBonus:    4,
	},
	{
		MinMonth:   9,
		Weapons:    []string{"heavy_plasma", "heavy_plasma", "alien_cannon", "alien_laser", "heavy_plasma"},
		ArmorBonus: 6,
		HPBonus:    6,
	},
}

// GetAlienEquipTier returns the tier index active for a given game month.
func GetAlienEquipTier(gameMonth int) int {
	tier := 0
	for i, t := range AlienEquipTiers {
		if gameMonth >= t.MinMonth {
			tier = i
		}
	}
	return tier
}

// GetTierWeapon picks a weapon from the current tier's pool.
// Higher-rank aliens favour heavier weapons from the pool.
func GetTierWeapon(tierIdx int, alienRank int) string {
	if tierIdx < 0 || tierIdx >= len(AlienEquipTiers) {
		return "plasma_pistol"
	}
	pool := AlienEquipTiers[tierIdx].Weapons
	if len(pool) == 0 {
		return "plasma_pistol"
	}
	// Higher rank = later index in the pool
	idx := alienRank
	if idx >= len(pool) {
		idx = len(pool) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return pool[idx]
}

// GetTierStatBonus returns the HP and armour bonus for a given tier.
func GetTierStatBonus(tierIdx int) (hpBonus, armorBonus int) {
	if tierIdx < 0 || tierIdx >= len(AlienEquipTiers) {
		return 0, 0
	}
	t := AlienEquipTiers[tierIdx]
	return t.HPBonus, t.ArmorBonus
}
