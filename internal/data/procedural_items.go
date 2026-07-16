package data

import (
	"fmt"
	"math/rand"
)

type ProceduralWeapon struct {
	ID         string
	Name       string
	ShortName  string
	Damage     int
	Accuracy   int
	TU         int
	Range      int
	AmmoMax    int
	BurstSize  int
	Weight     int
	CostBuy    int
	CostSell   int
	BattleType int
	DamageType int
}

type ProceduralArmor struct {
	ID        string
	Name      string
	ShortName string
	Undersuit int
	Health    int
	TUMod     int
	Value     int
}

var weaponPrefixes = map[int][]string{
	DMG_PLASMA:    {"Plasma", "Ion", "Fusion"},
	DMG_LASER:     {"Laser", "Photon", "Beam"},
	DMG_EXPLOSIVE: {"Rocket", "Missile", "Grenade"},
	DMG_MELEE:     {"Blade", "Claw", "Fang"},
	DMG_KINETIC:   {"Rail", "Gauss", "Slug"},
	DMG_PSIONIC:   {"Psi", "Mind", "Neural"},
}

var weaponSuffixes = []string{
	"Pistol", "Rifle", "Carbine", "Blaster", "Cannon", "Emitter",
}

var armorPrefixes = map[int][]string{
	DMG_PLASMA:    {"Plasma-Shielded", "Heat-Resistant", "Thermal"},
	DMG_LASER:     {"Reflective", "Light-Bending", "Mirror"},
	DMG_EXPLOSIVE: {"Blast-Resistant", "Impact", "Reinforced"},
	DMG_MELEE:     {"Puncture-Resistant", "Mesh", "Chain"},
	DMG_KINETIC:   {"Ballistic", "Composite", "Layered"},
	DMG_PSIONIC:   {"Psi-Shielded", "Mind-Warding", "Warded"},
}

var armorSuffixes = []string{
	"Vest", "Suit", "Plating", "Armour", "Guard",
}

func GenerateProceduralItems(seed int64, aliens []*AlienSpecies) ([]ProceduralWeapon, []ProceduralArmor) {
	rng := rand.New(rand.NewSource(seed))

	weaponCount := 2 + rng.Intn(2)
	armorCount := 1 + rng.Intn(2)

	weapons := make([]ProceduralWeapon, 0, weaponCount)
	armors := make([]ProceduralArmor, 0, armorCount)

	usedDmgTypes := make(map[int]bool)
	for _, sp := range aliens {
		usedDmgTypes[sp.PrimaryDMG] = true
	}

	dmgTypes := make([]int, 0, len(usedDmgTypes))
	for dt := range usedDmgTypes {
		dmgTypes = append(dmgTypes, dt)
	}
	rng.Shuffle(len(dmgTypes), func(i, j int) {
		dmgTypes[i], dmgTypes[j] = dmgTypes[j], dmgTypes[i]
	})

	if len(dmgTypes) == 0 {
		dmgTypes = []int{DMG_KINETIC}
	}

	for i := 0; i < weaponCount; i++ {
		dt := dmgTypes[i%len(dmgTypes)]
		w := generateProceduralWeapon(rng, i, dt)
		weapons = append(weapons, w)
	}

	for i := 0; i < armorCount; i++ {
		dt := dmgTypes[i%len(dmgTypes)]
		a := generateProceduralArmor(rng, i, dt)
		armors = append(armors, a)
	}

	return weapons, armors
}

func generateProceduralWeapon(rng *rand.Rand, idx int, dmgType int) ProceduralWeapon {
	prefixPool := weaponPrefixes[dmgType]
	prefix := prefixPool[rng.Intn(len(prefixPool))]
	suffix := weaponSuffixes[rng.Intn(len(weaponSuffixes))]
	name := prefix + " " + suffix

	damage := 20 + rng.Intn(40)
	accuracy := 55 + rng.Intn(30)
	tu := 15 + rng.Intn(15)
	ammoMax := 6 + rng.Intn(15)
	rangeVal := 10 + rng.Intn(20)
	burstSize := 1
	if rng.Intn(3) == 0 {
		burstSize = 3
		ammoMax *= 3
	}
	weight := 2 + rng.Intn(8)
	costBuy := 5000 + rng.Intn(10000)

	shortName := fmt.Sprintf("PW%d", idx+1)
	id := fmt.Sprintf("proc_weapon_%d", idx)

	return ProceduralWeapon{
		ID:         id,
		Name:       name,
		ShortName:  shortName,
		Damage:     damage,
		Accuracy:   accuracy,
		TU:         tu,
		Range:      rangeVal,
		AmmoMax:    ammoMax,
		BurstSize:  burstSize,
		Weight:     weight,
		CostBuy:    costBuy,
		CostSell:   costBuy * 3 / 4,
		BattleType: BT_FIREARM,
		DamageType: dmgType,
	}
}

func generateProceduralArmor(rng *rand.Rand, idx int, dmgType int) ProceduralArmor {
	prefixPool := armorPrefixes[dmgType]
	prefix := prefixPool[rng.Intn(len(prefixPool))]
	suffix := armorSuffixes[rng.Intn(len(armorSuffixes))]
	name := prefix + " " + suffix

	undersuit := 15 + rng.Intn(30)
	health := rng.Intn(15)
	tuMod := -(5 + rng.Intn(10))
	value := 20000 + rng.Intn(40000)

	shortName := fmt.Sprintf("PA%d", idx+1)
	id := fmt.Sprintf("proc_armor_%d", idx)

	return ProceduralArmor{
		ID:        id,
		Name:      name,
		ShortName: shortName,
		Undersuit: undersuit,
		Health:    health,
		TUMod:     tuMod,
		Value:     value,
	}
}

func RegisterProceduralItems(seed int64, aliens []*AlienSpecies) {
	weapons, armors := GenerateProceduralItems(seed, aliens)

	for _, w := range weapons {
		RuleItems[w.ID] = RuleItem{
			Type:       "STR_" + w.ShortName,
			Name:       w.Name,
			ShortName:  w.ShortName,
			Weight:     w.Weight,
			CostBuy:    w.CostBuy,
			CostSell:   w.CostSell,
			BattleType: w.BattleType,
			Damage:     w.Damage,
			Accuracy:   w.Accuracy,
			TU:         w.TU,
			Range:      w.Range,
			AmmoMax:    w.AmmoMax,
			AmmoCur:    w.AmmoMax,
			BurstSize:  w.BurstSize,
			Strength:   10,
		}
		Weapons[w.ID] = RuleItems[w.ID]
	}

	for _, a := range armors {
		Armors[a.ID] = Armor{
			Name:      a.Name,
			ShortName: a.ShortName,
			Undersuit: a.Undersuit,
			Health:    a.Health,
			TUMod:     a.TUMod,
			Value:     a.Value,
		}
	}
}
