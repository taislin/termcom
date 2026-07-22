package geo

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
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

	mode   data.CombatMode
	state  string // "player_turn" or "done"
	result string // "ufo_destroyed" / "inter_destroyed" / "disengaged"

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
	return &DogfightScreen{
		game:        game,
		gs:          gs,
		interceptor: inter,
		ufo:         ufo,
		ufoMaxHP:    ufoMaxHP,
		mode:        inter.Mode,
		state:       "player_turn",
		cityName:    cityName,
	}
}

func (ds *DogfightScreen) Update() {}

func (ds *DogfightScreen) HandleKey(ev *tcell.EventKey) {
	if ds.state == "done" {
		ds.finish()
		return
	}
	if ds.state != "player_turn" {
		return
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		ds.breakOff()
	default:
		switch ev.Str() {
		case "f", "F":
			ds.fire()
		case "m", "M":
			ds.cycleMode()
		case "b", "B":
			ds.breakOff()
		}
	}
}

func (ds *DogfightScreen) HandleMouse(*tcell.EventMouse) {}

func (ds *DogfightScreen) fire() {
	if ds.interceptor.Ammo <= 0 {
		ds.log3, ds.log2 = ds.log2, ds.log1
		ds.log1 = language.String("MSG_OUT_OF_AMMO")
		return
	}

	damage := ds.interceptor.FireAt(ds.ufo)
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

func (ds *DogfightScreen) cycleMode() {
	switch ds.mode {
	case data.CombatAttack:
		ds.mode = data.CombatCautious
	case data.CombatCautious:
		ds.mode = data.CombatBreakoff
	case data.CombatBreakoff:
		ds.mode = data.CombatAttack
	}
	ds.interceptor.SetMode(ds.mode)
}

func (ds *DogfightScreen) breakOff() {
	ds.state = "done"
	ds.result = "disengaged"
}

func (ds *DogfightScreen) finish() {
	switch ds.result {
	case "ufo_destroyed":
		ds.gs.Game.Funds += int64(ds.ufo.Type.Points * 1000)
		city := ds.gs.CityByID(ds.ufo.CurrentNode())
		if city != nil && GetTile(city.X, city.Y) != 0 {
			ds.gs.CrashSites = append(ds.gs.CrashSites, &CrashSite{
				UFOName: ds.ufo.Type.Name,
				NodeID:  ds.ufo.CurrentNode(),
				Seed:    rand.Int63(),
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
	}
	ds.gs.MessageTimer = ds.gs.Game.GameTime
	ds.game.PopState()
}

func (ds *DogfightScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	// Title
	title := "  D O G F I G H T  "
	ctx.DrawString((w-engine.StringWidth(title))/2, 0, title, engine.StyleDefault)
	ctx.DrawString(2, 2, ds.ufo.Type.Name, engine.StyleRed)

	// Interceptor panel
	panelY := 4
	panelH := 6
	panelW := (w / 2) - 2
	ix := 1
	ctx.DrawPanel(ix, panelY, panelW, panelH, ds.interceptor.Name, engine.StyleDefault)

	// HP bar
	barLen := panelW - 8
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
	ctx.DrawString(ix+2, panelY+1, fmt.Sprintf(language.String("DOGFIGHT_INTER_BAR"), bar, ds.interceptor.HP, ds.interceptor.MaxHP), hpStyle)

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
	ctx.DrawString(ix+2, panelY+2, fmt.Sprintf("Ammo: %s %d/%d", ammoBar, ds.interceptor.Ammo, ammoMax), ammoStyle)

	// Weapon and mode
	ctx.DrawString(ix+2, panelY+3, fmt.Sprintf("W: %s", ds.interceptor.Weapon.Name), engine.StyleDefault)
	ctx.DrawString(ix+2, panelY+4, fmt.Sprintf("%s", ds.mode.String()), engine.StyleGray)

	// UFO panel
	ux := w/2 + 1
	ctx.DrawPanel(ux, panelY, panelW, panelH, ds.ufo.Type.Short, engine.StyleDefault)

	// UFO HP bar
	ufoPct := float64(ds.ufo.Type.Toughness) / float64(ds.ufoMaxHP)
	if ufoPct < 0 {
		ufoPct = 0
	}
	ufoBar := makeHpBar(barLen, ufoPct)
	ufoStyle := engine.StyleRed
	ctx.DrawString(ux+2, panelY+1, fmt.Sprintf(language.String("DOGFIGHT_UFO_BAR"), ufoBar, ds.ufo.Type.Toughness, ds.ufoMaxHP), ufoStyle)

	// UFO weapon
	ctx.DrawString(ux+2, panelY+2, fmt.Sprintf("WPN: %s", ds.ufo.Type.Weapon), engine.StyleDefault)

	// Log messages
	logY := panelY + panelH + 2
	if ds.log1 != "" {
		ctx.DrawString(2, logY, ds.log1, engine.StyleYellow)
	}
	if ds.log2 != "" {
		ctx.DrawString(2, logY+1, ds.log2, engine.StyleYellow)
	}
	if ds.log3 != "" {
		ctx.DrawString(2, logY+2, ds.log3, engine.StyleYellow)
	}

	// Action bar
	actionY := h - 5
	sep := "  "
	xOff := 2
	ctx.DrawString(xOff, actionY, "[F] Fire", engine.StyleOrange)
	xOff += engine.StringWidth("[F] Fire") + engine.StringWidth(sep)
	ctx.DrawString(xOff, actionY, "[M] Mode", engine.StyleOrange)
	xOff += engine.StringWidth("[M] Mode") + engine.StringWidth(sep)
	ctx.DrawString(xOff, actionY, "[B] Break Off", engine.StyleRed)
	xOff += engine.StringWidth("[B] Break Off") + engine.StringWidth(sep)
	ctx.DrawString(xOff, actionY, "[Esc] Back", engine.StyleGray)

	// Status line
	statusY := h - 2
	if ds.state == "player_turn" {
		ctx.DrawString(2, statusY, language.String("PHASE_YOUR_TURN"), engine.StyleGreen)
		ctx.DrawString(2, statusY+1, language.String("HELP_BAT_SELECT"), engine.StyleGray)
	} else if ds.state == "done" {
		switch ds.result {
		case "ufo_destroyed":
			ctx.DrawString(2, statusY, language.String("DOGFIGHT_UFO_DESTROYED"), engine.StyleGreen)
		case "inter_destroyed":
			ctx.DrawString(2, statusY, language.String("MSG_INTERCEPTOR_DESTROYED"), engine.StyleRed)
		case "disengaged":
			ctx.DrawString(2, statusY, "Disengaged", engine.StyleYellow)
		}
		ctx.DrawString(2, statusY+1, "Press any key to continue", engine.StyleGray)
	}

	// Dogfight ASCII visual
	visY := panelY + panelH + 5
	if visY < statusY-4 {
		centerX := w / 2
		ctx.DrawString(centerX-6, visY, "  ◄══════►  ◉", engine.StyleCyan)
		ctx.DrawString(centerX-7, visY+1, ds.interceptor.Name, engine.StyleDefault)
		ctx.DrawString(centerX+8, visY+1, ds.ufo.Type.Short, engine.StyleRed)
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
