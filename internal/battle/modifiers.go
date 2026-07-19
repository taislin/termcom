package battle

import (
	"math/rand"

	"github.com/taislin/termcom/internal/language"
)

// Probability denominators for mission-modifier rolls (1-in-N chance).
const (
	nightOpsChance      = 3
	reinforcementsChance = 4
	timeLimitChance     = 5
	boobyTrapChance     = 3
	alienAmbushChance   = 4
	vipRescueChance     = 3
	heavyFogChance      = 5
	lowVisChance        = 6
	highGroundChance    = 6
	snowChance          = 2
	windChance          = 3
	rainChance          = 4
	stormChance         = 5
	marshFogChance      = 4
	defaultFogChance    = 6
)

// Weather accuracy/sight/fire tuning.
const (
	rainAccPenalty   = 5
	fogAccPerDensity = 5
	coldAccPenalty   = 3
	rainSightRed     = 2
	fogSightPerDensity = 3
	coldWindSightRed = 2
	fireSpreadRain   = 5
	fireSpreadWind   = 30
	fireSpreadBase   = 20
	fogRangeMin      = 1
	fogRangeSpan     = 2
	movePenaltyRain  = 2
)

// MissionModifier defines mission-specific environment rules.
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

// roll reports whether a 1-in-chance event occurs this rng draw.
func roll(rng *rand.Rand, chance int) bool {
	return rng.Intn(chance) == 0
}

func RollModifiers(rng *rand.Rand, missionType string) []MissionModifier {
	var mods []MissionModifier

	if roll(rng, nightOpsChance) {
		mods = append(mods, ModNightOps)
	}

	switch missionType {
	case "Terror", "Abduction":
		if roll(rng, reinforcementsChance) {
			mods = append(mods, ModReinforcements)
		}
		if roll(rng, timeLimitChance) {
			mods = append(mods, ModTimeLimit)
		}
	case "Supply Raid", "Alien Research":
		if roll(rng, boobyTrapChance) {
			mods = append(mods, ModBoobyTrapped)
		}
		if roll(rng, alienAmbushChance) {
			mods = append(mods, ModAlienAmbush)
		}
	case "Council":
		if roll(rng, vipRescueChance) {
			mods = append(mods, ModVIPRescue)
		}
	}

	if roll(rng, heavyFogChance) {
		mods = append(mods, ModHeavyFog)
	}
	if roll(rng, lowVisChance) {
		mods = append(mods, ModLowVisibility)
	}
	if roll(rng, highGroundChance) {
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
		if roll(rng, snowChance) {
			w.Snow = true
		}
		if roll(rng, windChance) {
			w.Wind = true
		}
		w.TempCold = true
	case "desert":
		if roll(rng, windChance) {
			w.Wind = true
		}
	case "marsh", "swamp":
		if roll(rng, rainChance) {
			w.Rain = true
		}
		if roll(rng, marshFogChance) {
			w.FogDensity = fogRangeMin + rng.Intn(fogRangeSpan)
		}
	case "mountain":
		if roll(rng, snowChance) {
			w.Snow = true
		}
		if roll(rng, windChance) {
			w.Wind = true
		}
		w.TempCold = true
	case "jungle":
		if roll(rng, rainChance*2) {
			w.Rain = true
		}
		if roll(rng, marshFogChance) {
			w.FogDensity = fogRangeMin + rng.Intn(fogRangeSpan)
		}
	case "farm", "coastal":
		if roll(rng, windChance) {
			w.Wind = true
		}
	default:
		if roll(rng, rainChance) {
			w.Rain = true
		}
		if roll(rng, stormChance) {
			w.Wind = true
		}
		if roll(rng, defaultFogChance) {
			w.FogDensity = fogRangeMin + rng.Intn(fogRangeSpan)
		}
	}

	return w
}

func (w Weather) AccuracyPenalty() int {
	pen := 0
	if w.Rain {
		pen += rainAccPenalty
	}
	if w.FogDensity > 0 {
		pen += w.FogDensity * fogAccPerDensity
	}
	if w.TempCold {
		pen += coldAccPenalty
	}
	return pen
}

func (w Weather) SightReduction() int {
	red := 0
	if w.Rain {
		red += rainSightRed
	}
	if w.FogDensity > 0 {
		red += w.FogDensity * fogSightPerDensity
	}
	if w.TempCold && w.Wind {
		red += coldWindSightRed
	}
	return red
}

func (w Weather) FireSpreadChance() int {
	if w.Rain {
		return fireSpreadRain
	}
	if w.Wind {
		return fireSpreadWind
	}
	return fireSpreadBase
}

// MovePenalty returns extra TU added per tile stepped on while the ground
// is muddy. Rain turns grass/dirt to mud (slower); cold+wind does not.
func (w Weather) MovePenalty() int {
	if w.Rain {
		return movePenaltyRain
	}
	return 0
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
