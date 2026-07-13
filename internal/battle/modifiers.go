package battle

import (
	"math/rand"

	"github.com/civ13/termcom/internal/language"
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
	switch m {
	case ModNightOps:
		return language.String("MODIFIER_NIGHT_OPS")
	case ModReinforcements:
		return language.String("MODIFIER_REINFORCEMENTS")
	case ModTimeLimit:
		return language.String("MODIFIER_TIME_LIMIT")
	case ModVIPRescue:
		return language.String("MODIFIER_VIP_RESCUE")
	case ModBoobyTrapped:
		return language.String("MODIFIER_BOOBY_TRAPPED")
	case ModHeavyFog:
		return language.String("MODIFIER_HEAVY_FOG")
	case ModAlienAmbush:
		return language.String("MODIFIER_ALIEN_AMBUSH")
	case ModLowVisibility:
		return language.String("MODIFIER_LOW_VISIBILITY")
	case ModHighGround:
		return language.String("MODIFIER_HIGH_GROUND")
	default:
		return language.String("NONE")
	}
}

func (m MissionModifier) Description() string {
	switch m {
	case ModNightOps:
		return language.String("MODDESC_NIGHT_OPS")
	case ModReinforcements:
		return language.String("MODDESC_REINFORCEMENTS")
	case ModTimeLimit:
		return language.String("MODDESC_TIME_LIMIT")
	case ModVIPRescue:
		return language.String("MODDESC_VIP_RESCUE")
	case ModBoobyTrapped:
		return language.String("MODDESC_BOOBY_TRAPPED")
	case ModHeavyFog:
		return language.String("MODDESC_HEAVY_FOG")
	case ModAlienAmbush:
		return language.String("MODDESC_ALIEN_AMBUSH")
	case ModLowVisibility:
		return language.String("MODDESC_LOW_VISIBILITY")
	case ModHighGround:
		return language.String("MODDESC_HIGH_GROUND")
	default:
		return ""
	}
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

func (w Weather) IsClear() bool {
	return !w.Snow && !w.Rain && !w.Wind && w.FogDensity == 0 && !w.TempCold
}

func (w Weather) Name() string {
	if w.Snow {
		return language.String("WEATHER_SNOW")
	}
	if w.Rain && w.Wind {
		return language.String("WEATHER_STORM")
	}
	if w.Rain {
		return language.String("WEATHER_RAIN")
	}
	if w.Wind {
		return language.String("WEATHER_WIND")
	}
	if w.FogDensity > 0 {
		return language.String("WEATHER_FOG")
	}
	if w.TempCold {
		return language.String("WEATHER_COLD")
	}
	return language.String("WEATHER_CLEAR")
}
