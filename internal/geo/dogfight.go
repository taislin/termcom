package geo

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/battle"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type DogfightScreen struct {
	game        *engine.Game
	gs          *Geoscape
	interceptor *Interceptor
	ufo         *UFO
	ufoMaxHP    int

	state    string // "player_turn" or "done"
	result   string // "ufo_destroyed" / "inter_destroyed" / "disengaged" / "ufo_escaped"
	rangePct float64

	log1 string
	log2 string
	log3 string

	cityName string
}

func NewDogfightScreen(game *engine.Game, gs *Geoscape, inter *Interceptor, ufo *UFO) *DogfightScreen {
	ufoMaxHP := ufo.Type.MaxHP
	if ufoMaxHP <= 0 {
		ufoMaxHP = ufo.Type.Toughness
	}
	cityName := ""
	if city := gs.CityByID(ufo.CurrentNode()); city != nil {
		cityName = city.LangName()
	}
	startRange := rangeFractionCautious
	switch inter.Mode {
	case data.CombatAttack:
		startRange = rangeFractionAttack
	case data.CombatBreakoff:
		startRange = rangeFractionBreakoff
	}
	return &DogfightScreen{
		game:        game,
		gs:          gs,
		interceptor: inter,
		ufo:         ufo,
		ufoMaxHP:    ufoMaxHP,
		state:       "player_turn",
		rangePct:    startRange,
		cityName:    cityName,
	}
}

func (ds *DogfightScreen) Update() {}

func (ds *DogfightScreen) HandleKey(ev *tcell.EventKey) {
	if ds.state == "done" {
		ds.finish()
		return
	}
	if ds.state == "escaping" {
		ds.finishEscapeAttempt()
		return
	}
	if ds.state != "player_turn" {
		return
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		ds.breakOff()
	case tcell.KeyLeft:
		ds.handleOpen()
	case tcell.KeyRight:
		ds.handleClose()
	default:
		switch ev.Str() {
		case "f", "F":
			ds.fire()
		case "b", "B":
			ds.breakOff()
		case "[", "{":
			ds.handleOpen()
		case "]", "}":
			ds.handleClose()
		case "-", "_":
			ds.handleOpen()
		case "=", "+":
			ds.handleClose()
		}
	}
}

func (ds *DogfightScreen) HandleMouse(e *tcell.EventMouse) {
	if e.Buttons() == tcell.ButtonNone {
		return
	}

	// When done, any click dismisses
	if ds.state == "done" {
		ds.finish()
		return
	}
	if ds.state == "escaping" {
		ds.finishEscapeAttempt()
		return
	}
	if ds.state != "player_turn" {
		return
	}

	x, y := e.Position()
	_, h := ds.game.ScreenSize()

	// Action bar at y = h-4
	if y != h-4 {
		return
	}

	type seg struct {
		start int
		end   int
		key   string
	}
	sep := "  "
	xOff := 2
	actions := []struct {
		text string
		key  string
	}{
		{"[F] Fire", "F"},
		{"[←] Retreat", "←"},
		{"[→] Advance", "→"},
		{"[B] Break Off", "B"},
		{"[Esc] Back", "Esc"},
		{"[?] Help", "?"},
	}
	var segs []seg
	for _, a := range actions {
		start := xOff
		end := xOff + engine.StringWidth(a.text)
		segs = append(segs, seg{start, end, a.key})
		xOff = end + engine.StringWidth(sep)
	}
	for _, s := range segs {
		if x >= s.start && x < s.end {
			switch s.key {
			case "F":
				ds.fire()
			case "←":
				ds.handleOpen()
			case "→":
				ds.handleClose()
			case "B":
				ds.breakOff()
			case "Esc":
				ds.game.PopState()
			case "?":
				ds.game.SetScreen(engine.StateHelp, engine.NewHelpScreen(ds.game, engine.StateDogfight))
				ds.game.PushState(engine.StateHelp)
			}
			return
		}
	}
}

func (ds *DogfightScreen) handleClose() {
	ds.closeRange()
	ds.ufoTakeTurn()
}

func (ds *DogfightScreen) handleOpen() {
	ds.openRange()
	ds.ufoTakeTurn()
}

func (ds *DogfightScreen) clampRange() {
	if ds.rangePct < 0.05 {
		ds.rangePct = 0.05
	}
	if ds.rangePct > 1.0 {
		ds.rangePct = 1.0
	}
}

func (ds *DogfightScreen) closeRange() {
	// Closing speed depends on interceptor speed vs UFO speed
	ratio := float64(ds.interceptor.Speed) / float64(ds.ufo.Type.Speed)
	delta := 0.1 * ratio
	if delta < 0.05 {
		delta = 0.05
	}
	if delta > 0.3 {
		delta = 0.3
	}
	ds.rangePct -= delta
	ds.clampRange()
}

func (ds *DogfightScreen) openRange() {
	// Retreating speed depends on UFO speed vs interceptor speed
	ratio := float64(ds.ufo.Type.Speed) / float64(ds.interceptor.Speed)
	delta := 0.1 * ratio
	if delta < 0.05 {
		delta = 0.05
	}
	if delta > 0.3 {
		delta = 0.3
	}
	ds.rangePct += delta
	ds.clampRange()
}

func (ds *DogfightScreen) fireAtRange(rangePct float64) int {
	if ds.interceptor.Ammo <= 0 {
		return 0
	}

	// Range-based accuracy: close = better, far = worse
	accuracy := ds.interceptor.EffectiveAccuracy()
	if rangePct > effectiveRangeRatioThreshold {
		accuracy = int(float64(accuracy) * (1.0 - (rangePct-effectiveRangeRatioThreshold)*rangeFalloffMultiplier))
	}

	switch ds.interceptor.Mode {
	case data.CombatAttack:
		accuracy += modeAccuracyAttackBonus
	case data.CombatBreakoff:
		accuracy -= modeAccuracyBreakoffPenalty
	}

	if accuracy < accuracyMin {
		accuracy = accuracyMin
	}
	if accuracy > accuracyMax {
		accuracy = accuracyMax
	}

	ds.interceptor.Ammo--
	if rand.Intn(100) >= accuracy {
		return 0
	}

	damage := ds.interceptor.Weapon.Damage + rand.Intn(ds.interceptor.Weapon.Damage/damageVarianceDivisor+1)
	if rand.Intn(100) < critChancePct {
		damage = damage * critMultiplierNum / critMultiplierDen
	}

	ds.ufo.Type.Toughness -= damage
	if ds.ufo.Type.Toughness <= 0 {
		ds.ufo.Active = false
		return -1
	}
	return damage
}

func (ds *DogfightScreen) fire() {
	if ds.interceptor.Ammo <= 0 {
		ds.log3, ds.log2 = ds.log2, ds.log1
		ds.log1 = language.String("MSG_OUT_OF_AMMO")
		return
	}

	damage := ds.fireAtRange(ds.rangePct)
	audio.PlayShoot()

	ds.log3, ds.log2 = ds.log2, ds.log1
	if damage == -1 {
		ds.log1 = language.String("DOGFIGHT_UFO_DESTROYED")
		ds.state = "done"
		ds.result = "ufo_destroyed"
		return
	}
	if damage > 0 {
		ds.log1 = fmt.Sprintf(language.String("DOGFIGHT_HIT"), damage)
	} else {
		ds.log1 = language.String("DOGFIGHT_MISS")
	}

	if !ds.ufo.Active {
		ds.state = "done"
		ds.result = "ufo_destroyed"
		return
	}

	// UFO takes its turn
	ds.ufoTakeTurn()
	if ds.state != "player_turn" {
		return
	}

	// Auto-adjust range based on combat mode and relative speed
	switch ds.interceptor.Mode {
	case data.CombatAttack:
		ratio := float64(ds.interceptor.Speed) / float64(ds.ufo.Type.Speed)
		delta := 0.1 * ratio
		if delta < 0.05 {
			delta = 0.05
		}
		if delta > 0.3 {
			delta = 0.3
		}
		ds.rangePct -= delta
		ds.clampRange()
	case data.CombatBreakoff:
		ratio := float64(ds.ufo.Type.Speed) / float64(ds.interceptor.Speed)
		delta := 0.1 * ratio
		if delta < 0.05 {
			delta = 0.05
		}
		if delta > 0.3 {
			delta = 0.3
		}
		ds.rangePct += delta
		ds.clampRange()
	}
}

func (ds *DogfightScreen) ufoTakeTurn() {
	ufoHPPct := float64(ds.ufo.Type.Toughness) / float64(ds.ufoMaxHP)
	interHPPct := float64(ds.interceptor.HP) / float64(ds.interceptor.MaxHP)

	// Transports always try to flee
	isTransport := ds.ufo.Type.Name == "Transport"
	// Badly damaged
	lowHP := ufoHPPct < 0.3
	// Significantly outmatched
	outmatched := ufoHPPct < 0.5 && interHPPct > 0.5

	if isTransport || lowHP || outmatched {
		ds.ufoRetreat()
		return
	}

	// Otherwise, fight
	ds.ufoAttack()
	if ds.state != "player_turn" {
		return
	}

	// UFO adjusts range for optimal fighting position
	ds.ufoAdjustRange()
}

func (ds *DogfightScreen) ufoRetreat() {
	// UFO tries to open range
	ratio := float64(ds.ufo.Type.Speed) / float64(ds.interceptor.Speed)
	delta := 0.15 * ratio
	if delta < 0.05 {
		delta = 0.05
	}
	if delta > 0.3 {
		delta = 0.3
	}
	ds.rangePct += delta
	ds.clampRange()

	ds.log3, ds.log2 = ds.log2, ds.log1

	if ds.rangePct >= 0.9 {
		ds.state = "done"
		ds.result = "ufo_escaped"
		ds.log1 = fmt.Sprintf(language.String("MSG_UFO_ESCAPED"), ds.ufo.Type.DisplayName())
		return
	}

	ds.log1 = language.String("DOGFIGHT_UFO_RETREATING")
}

func (ds *DogfightScreen) ufoAttack() {
	ufoDmg := ds.ufo.FireAtInterceptor(ds.interceptor)
	audio.PlayPlasmaFire()
	if ufoDmg > 0 {
		ds.log3, ds.log2 = ds.log2, ds.log1
		ds.log1 = fmt.Sprintf(language.String("MSG_UFO_HIT_INTERCEPTOR"), ufoDmg, ds.interceptor.HP, ds.interceptor.MaxHP)
	}

	if ds.interceptor.HP <= 0 {
		ds.state = "done"
		ds.result = "inter_destroyed"
	}
}

func (ds *DogfightScreen) ufoAdjustRange() {
	ufoHPPct := float64(ds.ufo.Type.Toughness) / float64(ds.ufoMaxHP)
	ratio := float64(ds.ufo.Type.Speed) / float64(ds.interceptor.Speed)
	delta := 0.08 * ratio
	if delta < 0.03 {
		delta = 0.03
	}
	if delta > 0.2 {
		delta = 0.2
	}

	if ufoHPPct > 0.6 {
		if ds.rangePct > 0.5 {
			ds.rangePct -= delta
		}
	} else {
		if ds.rangePct < 0.8 {
			ds.rangePct += delta
		}
	}
	ds.clampRange()
}

func (ds *DogfightScreen) breakOff() {
	ds.state = "escaping"
	ds.log3, ds.log2 = ds.log2, ds.log1
	ds.log1 = language.String("DOGFIGHT_BREAKING_OFF")
}

func (ds *DogfightScreen) finishEscapeAttempt() {
	if rand.Intn(100) < 70 {
		ds.state = "done"
		ds.result = "disengaged"
		return
	}
	ds.log3, ds.log2 = ds.log2, ds.log1
	ds.log1 = language.String("DOGFIGHT_ESCAPE_FAILED")
	ufoDmg := ds.ufo.FireAtInterceptor(ds.interceptor)
	audio.PlayPlasmaFire()
	if ufoDmg > 0 {
		ds.log3, ds.log2 = ds.log2, ds.log1
		ds.log1 = fmt.Sprintf(language.String("MSG_UFO_HIT_INTERCEPTOR"), ufoDmg, ds.interceptor.HP, ds.interceptor.MaxHP)
	}
	if ds.interceptor.HP <= 0 {
		ds.state = "done"
		ds.result = "inter_destroyed"
		return
	}
	ds.state = "player_turn"
}

func (ds *DogfightScreen) finish() {
	switch ds.result {
	case "ufo_destroyed":
		ds.gs.Game.Funds += int64(ds.ufo.Type.Points * 1000)
		city := ds.gs.CityByID(ds.ufo.CurrentNode())
		if city != nil && GetTile(city.X, city.Y) != 0 {
			biome := battle.CrashBiomeFromCoords(city.X, city.Y)
			ds.gs.CrashSites = append(ds.gs.CrashSites, &CrashSite{
				UFOName: ds.ufo.Type.Name,
				NodeID:  ds.ufo.CurrentNode(),
				Seed:    rand.Int63(),
				Biome:   biome,
			})
			ds.gs.Message = fmt.Sprintf(language.String("MSG_UFO_CRASHED"), ds.ufo.Type.DisplayName())
		} else {
			ds.gs.Message = fmt.Sprintf(language.String("MSG_UFO_LOST_AT_SEA"), ds.ufo.Type.DisplayName())
		}
		ds.interceptor.Disengage()
	case "inter_destroyed":
		ds.interceptor.Disengage()
		ds.gs.Message = language.String("MSG_INTERCEPTOR_DESTROYED")
	case "disengaged":
		ds.interceptor.Disengage()
		ds.gs.Message = fmt.Sprintf("%s recalled to base", ds.interceptor.Name)
	case "ufo_escaped":
		ds.ufo.Active = false
		ds.interceptor.Disengage()
		ds.gs.Message = fmt.Sprintf(language.String("MSG_UFO_ESCAPED"), ds.ufo.Type.DisplayName())
	}
	ds.gs.MessageTimer = ds.gs.Game.GameTime
	ds.game.PopState()
}

func (ds *DogfightScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	// Title
	title := "  D O G F I G H T  "
	ctx.DrawString((w-engine.StringWidth(title))/2, 0, title, engine.StyleDefault)
	ctx.DrawString(2, 2, ds.ufo.Type.DisplayName(), engine.StyleRed)

	// Interceptor panel
	panelY := 4
	panelH := 5
	panelW := (w / 2) - 2
	ix := 1
	ctx.DrawPanel(ix, panelY, panelW, panelH, ds.interceptor.Name, engine.StyleDefault)

	// HP bar — small fixed width
	barLen := 16
	interPct := float64(ds.interceptor.HP) / float64(ds.interceptor.MaxHP)
	if interPct < 0 {
		interPct = 0
	}
	bar := makeHpBar(barLen, interPct)
	hpStyle := engine.StyleGreen
	if interPct < 0.3 {
		hpStyle = engine.StyleRed
	} else if interPct < 0.6 {
		hpStyle = engine.StyleYellow
	}
	ctx.DrawString(ix+2, panelY+1, fmt.Sprintf("▲ %s %d/%d", bar, ds.interceptor.HP, ds.interceptor.MaxHP), hpStyle)

	// Ammo bar
	ammoMax := ds.interceptor.Weapon.FireRate * ammoPerFireRate
	ammoPct := float64(ds.interceptor.Ammo) / float64(ammoMax)
	if ammoPct > 1 {
		ammoPct = 1
	}
	ammoBar := makeHpBar(barLen, ammoPct)
	ammoStyle := engine.StyleCyan
	if ammoPct < 0.3 {
		ammoStyle = engine.StyleRed
	} else if ammoPct < 0.6 {
		ammoStyle = engine.StyleYellow
	}
	ctx.DrawString(ix+2, panelY+2, fmt.Sprintf("Ammo %s %d/%d", ammoBar, ds.interceptor.Ammo, ammoMax), ammoStyle)

	// Weapon and mode
	ctx.DrawString(ix+2, panelY+3, fmt.Sprintf("W: %s", ds.interceptor.Weapon.Name), engine.StyleDefault)
	ctx.DrawString(ix+2, panelY+4, ds.interceptor.Mode.String(), engine.StyleGray)

	// UFO panel
	ux := w/2 + 1
	ctx.DrawPanel(ux, panelY, panelW, panelH, ds.ufo.Type.DisplayName(), engine.StyleDefault)

	// UFO HP bar
	ufoPct := float64(ds.ufo.Type.Toughness) / float64(ds.ufoMaxHP)
	if ufoPct < 0 {
		ufoPct = 0
	}
	ufoBar := makeHpBar(barLen, ufoPct)
	ufoStyle := engine.StyleRed
	ctx.DrawString(ux+2, panelY+1, fmt.Sprintf("◉ %s %d/%d", ufoBar, ds.ufo.Type.Toughness, ds.ufoMaxHP), ufoStyle)

	// UFO weapon
	ctx.DrawString(ux+2, panelY+2, fmt.Sprintf("Weapon: %s", ds.ufo.Type.WeaponDisplayName()), engine.StyleDefault)

	// Range proximity bar — uses ▲ interceptor / ◉ alien on a track
	rangeBarY := panelY + panelH + 1
	rangeBarX := 2
	rangeBarLen := w - 4
	if rangeBarLen > 40 {
		rangeBarLen = 40
	}
	if rangeBarLen < 14 {
		rangeBarLen = 14
	}
	// Draw track
	for i := 0; i < rangeBarLen; i++ {
		ctx.SetCell(rangeBarX+i, rangeBarY, '─', engine.StyleGray)
	}
	// Position glyphs: interceptor at left (fixed), alien moves with range
	interGlyphX := rangeBarX
	alienGlyphX := rangeBarX + 2 + int(ds.rangePct*float64(rangeBarLen-4))
	if alienGlyphX >= rangeBarX+rangeBarLen-1 {
		alienGlyphX = rangeBarX + rangeBarLen - 2
	}
	if alienGlyphX <= interGlyphX+1 {
		alienGlyphX = interGlyphX + 2
	}
	ctx.SetCell(interGlyphX, rangeBarY, '▲', engine.StyleCyanBold)
	ctx.SetCell(alienGlyphX, rangeBarY, '◉', engine.StyleRedBold)
	// Range label
	rangeLabel := fmt.Sprintf(" %d%%", int(ds.rangePct*100))
	ctx.DrawString(rangeBarX+rangeBarLen+1, rangeBarY, rangeLabel, engine.StyleGray)

	// Log messages
	logY := rangeBarY + 2
	if ds.log1 != "" {
		ctx.DrawString(2, logY, ds.log1, engine.StyleYellow)
	}
	if ds.log2 != "" {
		ctx.DrawString(2, logY+1, ds.log2, engine.StyleYellow)
	}
	if ds.log3 != "" {
		ctx.DrawString(2, logY+2, ds.log3, engine.StyleYellow)
	}

	// Action bar — orange key in brackets, grey for the rest
	actionY := h - 4
	sep := "  "
	xOff := 2
	actionStyle := engine.StyleGray
	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "F", engine.StyleOrange)
	ctx.DrawString(xOff+2, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "Fire", actionStyle)
	xOff += engine.StringWidth("[F] Fire") + engine.StringWidth(sep)

	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "←", engine.StyleOrange)
	ctx.DrawString(xOff+2, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "Retreat", actionStyle)
	xOff += engine.StringWidth("[←] Retreat") + engine.StringWidth(sep)

	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "→", engine.StyleOrange)
	ctx.DrawString(xOff+2, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "Advance", actionStyle)
	xOff += engine.StringWidth("[→] Advance") + engine.StringWidth(sep)

	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "B", engine.StyleOrange)
	ctx.DrawString(xOff+2, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "Break Off", actionStyle)
	xOff += engine.StringWidth("[B] Break Off") + engine.StringWidth(sep)

	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "Esc", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+6, actionY, "Back", actionStyle)
	xOff += engine.StringWidth("[Esc] Back") + engine.StringWidth(sep)

	ctx.DrawString(xOff, actionY, "[", engine.StyleOrange)
	ctx.DrawString(xOff+1, actionY, "?", engine.StyleOrange)
	ctx.DrawString(xOff+2, actionY, "] ", engine.StyleOrange)
	ctx.DrawString(xOff+4, actionY, "Help", actionStyle)

	// Status line
	statusY := h - 2
	if ds.state == "player_turn" {
		ctx.DrawString(2, statusY, language.String("PHASE_YOUR_TURN"), engine.StyleGreen)
	} else if ds.state == "escaping" {
		ctx.DrawString(2, statusY, language.String("DOGFIGHT_BREAKING_OFF"), engine.StyleYellow)
		ctx.DrawString(2, statusY+1, "Press any key", engine.StyleGray)
	} else if ds.state == "done" {
		switch ds.result {
		case "ufo_destroyed":
			ctx.DrawString(2, statusY, language.String("DOGFIGHT_UFO_DESTROYED"), engine.StyleGreen)
		case "inter_destroyed":
			ctx.DrawString(2, statusY, language.String("MSG_INTERCEPTOR_DESTROYED"), engine.StyleRed)
		case "disengaged":
			ctx.DrawString(2, statusY, "Disengaged", engine.StyleYellow)
		case "ufo_escaped":
			ctx.DrawString(2, statusY, fmt.Sprintf(language.String("MSG_UFO_ESCAPED"), ds.ufo.Type.DisplayName()), engine.StyleYellow)
		}
		ctx.DrawString(2, statusY+1, "Press any key to continue", engine.StyleGray)
	}
}

func makeHpBar(length int, pct float64) string {
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(length))
	bar := ""
	for i := 0; i < length; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
