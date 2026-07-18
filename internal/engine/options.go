package engine

import (
	"fmt"

	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type flagCell struct {
	Ch   rune
	Fg   color.Color
	Bg   color.Color
	BgOk bool
}

type flagGrid [3][6]flagCell

const (
	optionsCenterXOff = 15
	optionsCenterYOff = 10
	boolHitWidth      = 30
	langHitWidth      = 35
	langFlagOffset    = 7
	maxActionDelay    = 20
	maxSfxVolume      = 10
)

func fc(fg color.Color) flagCell {
	return flagCell{Ch: '█', Fg: fg}
}

func fcb(ch rune, fg, bg color.Color) flagCell {
	return flagCell{Ch: ch, Fg: fg, Bg: bg, BgOk: true}
}

var langFlags = map[string]flagGrid{
	"en": {
		{fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm15), fc(color.XTerm15)},
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm15), fc(color.XTerm15)},
	},
	"zh": {
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm11), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
	},
	"es": {
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm11), fc(color.XTerm11), fc(color.XTerm11), fc(color.XTerm11), fc(color.XTerm11), fc(color.XTerm11)},
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
	},
	"fr": {
		{fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm9), fc(color.XTerm9)},
	},
	"ru": {
		{fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm15), fc(color.XTerm15)},
		{fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm4), fc(color.XTerm4)},
		{fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
	},
	"pt": {
		{fc(color.XTerm2), fc(color.XTerm2), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm2), fc(color.XTerm2), fc(color.XTerm11), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
		{fc(color.XTerm2), fc(color.XTerm2), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9), fc(color.XTerm9)},
	},
	"ja": {
		{fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
		{fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('●', color.XTerm9, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
		{fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
	},
	"ko": {
		{fcb('█', color.XTerm15, color.XTerm15), fcb('/', color.XTerm0, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('\\', color.XTerm0, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
		{fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('▀', color.XTerm9, color.XTerm12), fcb('▄', color.XTerm9, color.XTerm12), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
		{fcb('█', color.XTerm15, color.XTerm15), fcb('\\', color.XTerm0, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15), fcb('/', color.XTerm0, color.XTerm15), fcb('█', color.XTerm15, color.XTerm15)},
	},
}

func drawFlag(ctx *ScreenCtx, x, y int, code string) {
	f, ok := langFlags[code]
	if !ok {
		return
	}
	for dy := 0; dy < 3; dy++ {
		for dx := 0; dx < 6; dx++ {
			cell := f[dy][dx]
			style := tcell.StyleDefault.Foreground(cell.Fg)
			if cell.BgOk {
				style = style.Background(cell.Bg)
			}
			ctx.SetCell(x+dx, y+dy, cell.Ch, style)
		}
	}
}

type OptionsScreen struct {
	Game      *Game
	Selection int
}

func NewOptionsScreen(g *Game) *OptionsScreen {
	return &OptionsScreen{Game: g, Selection: 0}
}

func (os *OptionsScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("OPTIONS_TITLE"), StyleDefault)

	startY := h/2 - optionsCenterYOff
	baseX := w/2 - optionsCenterXOff

	// Bool toggles
	boolOpts := []struct {
		Label string
		Value *bool
	}{
		{language.String("OPTIONS_BLOOM"), &Config.BloomEnabled},
		{language.String("OPTIONS_LIGHTING"), &Config.LightingEnabled},
		{language.String("OPTIONS_SOUND"), &Config.SoundEnabled},
		{language.String("OPTIONS_AUTOSAVE"), &Config.AutosaveEnabled},
		{language.String("OPTIONS_SHAKE"), &Config.ScreenShake},
		{language.String("OPTIONS_MOUSE"), &Config.MouseEnabled},
		{language.String("OPTIONS_GRID"), &Config.GridLines},
		{language.String("OPTIONS_CONFIRM"), &Config.ConfirmDialogs},
		{language.String("OPTIONS_PAUSE_ALIEN"), &Config.PauseOnAlienDetect},
	}
	for i, opt := range boolOpts {
		style := StyleDefault
		if i == os.Selection {
			style = StyleHighlight
		}
		status := language.String("OPTIONS_OFF")
		if *opt.Value {
			status = language.String("OPTIONS_ON")
		}
		ctx.DrawString(baseX, startY+i, fmt.Sprintf("[%s] %s", status, opt.Label), style)
	}

	// Theme cycler
	themeIdx := len(boolOpts)
	speedIdx := themeIdx + 1
	volIdx := speedIdx + 1
	langIdx := volIdx + 1
	tutorialIdx := langIdx + 1
	themeStyle := StyleDefault
	if os.Selection == themeIdx {
		themeStyle = StyleHighlight
	}
	themeName := Config.Theme
	if tn, ok := themeDisplayNames[Config.Theme]; ok {
		themeName = tn
	}
	ctx.DrawString(baseX, startY+themeIdx+1, fmt.Sprintf("%s: [%s]", language.String("OPTIONS_THEME"), themeName), themeStyle)

	// Speed slider
	speedStyle := StyleDefault
	if os.Selection == speedIdx {
		speedStyle = StyleHighlight
	}
	ctx.DrawString(baseX, startY+speedIdx+1, fmt.Sprintf("%s: %d", language.String("OPTIONS_RESOLUTION_SPEED"), Config.ActionDelay), speedStyle)

	// Volume slider
	volStyle := StyleDefault
	if os.Selection == volIdx {
		volStyle = StyleHighlight
	}
	ctx.DrawString(baseX, startY+volIdx+1, fmt.Sprintf("%s: %d", language.String("OPTIONS_VOLUME"), Config.SfxVolume), volStyle)

	// Language with flag
	langStyle := StyleDefault
	if os.Selection == langIdx {
		langStyle = StyleHighlight
	}
	langs := language.Available()
	li := 0
	for i, l := range langs {
		if l == language.Current() {
			li = i
			break
		}
	}
		flagY := startY + langIdx + 2
	drawFlag(ctx, baseX, flagY, language.Current())
	ctx.DrawString(baseX+langFlagOffset, flagY+1, fmt.Sprintf("  %s: [%s]", language.String("OPTIONS_LANGUAGE"), langs[li]), langStyle)

	// Replay Tutorial action
	tutorialStyle := StyleDefault
	if os.Selection == tutorialIdx {
		tutorialStyle = StyleHighlight
	}
	ctx.DrawString(baseX, startY+tutorialIdx+1, fmt.Sprintf("  %s", language.String("OPTIONS_REPLAY_TUTORIAL")), tutorialStyle)

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, language.String("OPTIONS_HELP"), StyleGray, StyleHotkey)
}

func (os *OptionsScreen) HandleKey(e *tcell.EventKey) {
	const totalOptions = 14
	switch e.Key() {
	case tcell.KeyUp:
		audio.PlayMenuNav()
		os.Selection--
		if os.Selection < 0 {
			os.Selection = totalOptions - 1
		}
	case tcell.KeyDown:
		audio.PlayMenuNav()
		os.Selection++
		if os.Selection >= totalOptions {
			os.Selection = 0
		}
	case tcell.KeyEnter:
		audio.PlaySelect()
		os.toggle()
	case tcell.KeyLeft:
		audio.PlayMenuNav()
		os.applyOptionDelta(os.Selection, -1)
	case tcell.KeyRight:
		audio.PlayMenuNav()
		os.applyOptionDelta(os.Selection, 1)
	case tcell.KeyEsc:
		os.Game.PopState()
		SaveConfig()
	}
}

func (os *OptionsScreen) toggle() {
	switch os.Selection {
	case 0:
		Config.BloomEnabled = !Config.BloomEnabled
	case 1:
		Config.LightingEnabled = !Config.LightingEnabled
	case 2:
		Config.SoundEnabled = !Config.SoundEnabled
		audio.SetAudioEnabled(Config.SoundEnabled)
	case 3:
		Config.AutosaveEnabled = !Config.AutosaveEnabled
	case 4:
		Config.ScreenShake = !Config.ScreenShake
	case 5:
		Config.MouseEnabled = !Config.MouseEnabled
	case 6:
		Config.GridLines = !Config.GridLines
	case 7:
		Config.ConfirmDialogs = !Config.ConfirmDialogs
	case 8:
		Config.PauseOnAlienDetect = !Config.PauseOnAlienDetect
	case 9:
		os.cycleTheme(1)
	case 13:
		Config.TutorialShown = false
		SaveConfig()
		os.Game.RegisterScreen(StateTutorial, NewTutorialScreen(os.Game, nil))
		os.Game.PushState(StateTutorial)
	}
}

func cycleOption(dir int, options []string, current string, setter func(string)) {
	idx := 0
	for i, o := range options {
		if o == current {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(options) - 1
	}
	if idx >= len(options) {
		idx = 0
	}
	setter(options[idx])
}

var themes = []string{"default", "high_contrast", "amber", "green", "paper"}

var themeDisplayNames = map[string]string{
	"default":       language.String("THEME_DEFAULT"),
	"high_contrast": language.String("THEME_HIGH_CONTRAST"),
	"amber":         language.String("THEME_AMBER"),
	"green":         language.String("THEME_GREEN"),
	"paper":         language.String("THEME_PAPER"),
}

func (os *OptionsScreen) cycleTheme(dir int) {
	cycleOption(dir, themes, Config.Theme, func(t string) {
		Config.Theme = t
		os.Game.screen.SetTheme(Config.Theme)
	})
}

func (os *OptionsScreen) cycleLang(dir int) {
	cycleOption(dir, language.Available(), language.Current(), func(l string) {
		language.SetLanguage(l)
		Config.Language = l
	})
}

func (os *OptionsScreen) HandleMouse(e *tcell.EventMouse) {
	if !Config.MouseEnabled {
		return
	}
	x, y := e.Position()
	w, h := os.Game.ScreenSize()
	baseX := w/2 - optionsCenterXOff
	startY := h/2 - optionsCenterYOff

	buttons := e.Buttons()
	if buttons&tcell.Button1 == 0 && buttons&tcell.Button2 == 0 {
		return
	}

	// Determine which option row was clicked, accounting for non-sequential spacing
	optIndex := -1
	for i := 0; i < 9; i++ {
		if y == startY+i && x >= baseX && x < baseX+boolHitWidth {
			optIndex = i
			break
		}
	}
	if optIndex < 0 {
		if y == startY+10 && x >= baseX && x < baseX+boolHitWidth {
			optIndex = 9 // theme
		} else if y == startY+11 && x >= baseX && x < baseX+boolHitWidth {
			optIndex = 10 // speed
		} else if y == startY+12 && x >= baseX && x < baseX+boolHitWidth {
			optIndex = 11 // volume
		} else if y == startY+15 && x >= baseX+langFlagOffset && x < baseX+langHitWidth {
			optIndex = 12 // language
		} else if y == startY+14 && x >= baseX && x < baseX+boolHitWidth {
			optIndex = 13 // tutorial
		}
	}
	if optIndex < 0 {
		return
	}
	os.Selection = optIndex
	audio.PlayMenuNav()

	if buttons&tcell.Button2 != 0 {
		if optIndex >= 0 && optIndex < 9 {
			os.toggle()
		} else {
			os.applyOptionDelta(optIndex, -1)
		}
		return
	}

	if optIndex >= 0 && optIndex < 9 {
		os.toggle()
	} else {
		os.applyOptionDelta(optIndex, 1)
	}
}

func (os *OptionsScreen) applyOptionDelta(idx, dir int) {
	switch idx {
	case 9:
		os.cycleTheme(dir)
	case 10:
		Config.ActionDelay += dir
		if Config.ActionDelay < 1 {
			Config.ActionDelay = 1
		}
		if Config.ActionDelay > maxActionDelay {
			Config.ActionDelay = maxActionDelay
		}
		os.Game.ActionDelay = Config.ActionDelay
	case 11:
		Config.SfxVolume += dir
		if Config.SfxVolume < 0 {
			Config.SfxVolume = 0
		}
		if Config.SfxVolume > maxSfxVolume {
			Config.SfxVolume = maxSfxVolume
		}
		audio.SetSfxVolume(Config.SfxVolume)
	case 12:
		os.cycleLang(dir)
	case 13:
		// No left/right adjustment for tutorial action
	}
}

func (os *OptionsScreen) Update() {
	// Nothing to update
}
