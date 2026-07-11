package battle

import (
	"math/rand"
)

type MissionModifier int

const (
	ModNone MissionModifier = iota
	ModNightOps
	ModReinforcements
	ModTimeLimit
	ModVIPRescue
	ModBoobyTrapped
	ModHeavyFog
	ModAlienAmbush
	ModLowVisibility
	ModHighGround
)

var modifierNames = map[MissionModifier]string{
	ModNone:           "None",
	ModNightOps:       "Night Ops",
	ModReinforcements: "Reinforcements",
	ModTimeLimit:      "Time Limit",
	ModVIPRescue:      "VIP Rescue",
	ModBoobyTrapped:   "Booby Trapped",
	ModHeavyFog:       "Heavy Fog",
	ModAlienAmbush:    "Alien Ambush",
	ModLowVisibility:  "Low Visibility",
	ModHighGround:     "High Ground",
}

var modifierDescriptions = map[MissionModifier]string{
	ModNightOps:       "Forced night battle. +20% loot reward.",
	ModReinforcements: "Extra alien wave arrives on turn 4.",
	ModTimeLimit:      "15 turns to eliminate all aliens or reach objective.",
	ModVIPRescue:      "Protect the VIP. Bonus $50K if VIP survives.",
	ModBoobyTrapped:   "More grenades and proximity mines on the map.",
	ModHeavyFog:       "Sight range reduced by 40%. Smoke lingers longer.",
	ModAlienAmbush:    "Aliens start in overwatch positions.",
	ModLowVisibility:  "Reduced accuracy (-10%) for all units.",
	ModHighGround:     "Map has elevated positions with accuracy bonus.",
}

func (m MissionModifier) String() string {
	if s, ok := modifierNames[m]; ok {
		return s
	}
	return "Unknown"
}

func (m MissionModifier) Description() string {
	if s, ok := modifierDescriptions[m]; ok {
		return s
	}
	return ""
}

func RollModifiers(rng *rand.Rand, missionType string) []MissionModifier {
	var mods []MissionModifier

	if rng.Intn(3) == 0 {
		mods = append(mods, ModNightOps)
	}

	switch missionType {
	case "Terror", "Abduction":
		if rng.Intn(4) == 0 {
			mods = append(mods, ModReinforcements)
		}
		if rng.Intn(5) == 0 {
			mods = append(mods, ModTimeLimit)
		}
	case "Supply Raid", "Alien Research":
		if rng.Intn(3) == 0 {
			mods = append(mods, ModBoobyTrapped)
		}
		if rng.Intn(4) == 0 {
			mods = append(mods, ModAlienAmbush)
		}
	case "Council":
		if rng.Intn(3) == 0 {
			mods = append(mods, ModVIPRescue)
		}
	}

	if rng.Intn(5) == 0 {
		mods = append(mods, ModHeavyFog)
	}
	if rng.Intn(6) == 0 {
		mods = append(mods, ModLowVisibility)
	}
	if rng.Intn(6) == 0 {
		mods = append(mods, ModHighGround)
	}

	return mods
}

func HasModifier(mods []MissionModifier, target MissionModifier) bool {
	for _, m := range mods {
		if m == target {
			return true
		}
	}
	return false
}

type Weather struct {
	Rain       bool
	Wind       bool
	Snow       bool
	FogDensity int
	TempCold   bool
}

func RollWeather(rng *rand.Rand, biome string) Weather {
	w := Weather{}

	switch biome {
	case "snow", "polar":
		if rng.Intn(2) == 0 {
			w.Snow = true
		}
		if rng.Intn(3) == 0 {
			w.Wind = true
		}
		w.TempCold = true
	case "desert":
		if rng.Intn(4) == 0 {
			w.Wind = true
		}
	case "marsh":
		if rng.Intn(3) == 0 {
			w.Rain = true
		}
		if rng.Intn(4) == 0 {
			w.FogDensity = 1 + rng.Intn(2)
		}
	default:
		if rng.Intn(4) == 0 {
			w.Rain = true
		}
		if rng.Intn(5) == 0 {
			w.Wind = true
		}
		if rng.Intn(6) == 0 {
			w.FogDensity = 1 + rng.Intn(2)
		}
	}

	return w
}

func (w Weather) AccuracyPenalty() int {
	pen := 0
	if w.Rain {
		pen += 5
	}
	if w.FogDensity > 0 {
		pen += w.FogDensity * 5
	}
	if w.TempCold {
		pen += 3
	}
	return pen
}

func (w Weather) SightReduction() int {
	red := 0
	if w.Rain {
		red += 2
	}
	if w.FogDensity > 0 {
		red += w.FogDensity * 3
	}
	if w.TempCold && w.Wind {
		red += 2
	}
	return red
}

func (w Weather) FireSpreadChance() int {
	if w.Rain {
		return 5
	}
	if w.Wind {
		return 30
	}
	return 20
}

func (w Weather) Name() string {
	if w.Snow {
		return "Snow"
	}
	if w.Rain && w.Wind {
		return "Storm"
	}
	if w.Rain {
		return "Rain"
	}
	if w.Wind {
		return "Wind"
	}
	if w.FogDensity > 0 {
		return "Fog"
	}
	if w.TempCold {
		return "Cold"
	}
	return "Clear"
}
