package engine

import (
	"fmt"

	"github.com/civ13/termcom/internal/audio"
	"github.com/civ13/termcom/internal/language"
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

	const (
		themeIdx = 9
		speedIdx = 10
		volIdx   = 11
		langIdx  = 12
	)
	startY := h/2 - 5
	baseX := w/2 - 15

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
	themeStyle := StyleDefault
	if os.Selection == themeIdx {
		themeStyle = StyleHighlight
	}
	themeName := Config.Theme
	if tn, ok := themeDisplayNames[Config.Theme]; ok {
		themeName = tn
	}
	ctx.DrawString(baseX, startY+themeIdx, fmt.Sprintf("%s: [%s]", language.String("OPTIONS_THEME"), themeName), themeStyle)

	// Speed slider
	speedStyle := StyleDefault
	if os.Selection == speedIdx {
		speedStyle = StyleHighlight
	}
	ctx.DrawString(baseX, startY+speedIdx, fmt.Sprintf("%s: %d", language.String("OPTIONS_RESOLUTION_SPEED"), Config.ActionDelay), speedStyle)

	// Volume slider
	volStyle := StyleDefault
	if os.Selection == volIdx {
		volStyle = StyleHighlight
	}
	ctx.DrawString(baseX, startY+volIdx, fmt.Sprintf("%s: %d", language.String("OPTIONS_VOLUME"), Config.SfxVolume), volStyle)

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
	flagY := startY + langIdx
	drawFlag(ctx, baseX, flagY, language.Current())
	ctx.DrawString(baseX+7, flagY+1, fmt.Sprintf("  %s: [%s]", language.String("OPTIONS_LANGUAGE"), langs[li]), langStyle)

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, "[\u2190]/[\u2192]=Adjust  [\u2191]/[\u2193]=Select  Enter=Toggle  [Esc]=Back", StyleGray, StyleHotkey)
}

func (os *OptionsScreen) HandleKey(e *tcell.EventKey) {
	const totalOptions = 13
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
		switch os.Selection {
		case 9:
			os.cycleTheme(-1)
		case 10:
			Config.ActionDelay--
			if Config.ActionDelay < 1 {
				Config.ActionDelay = 1
			}
			os.Game.ActionDelay = Config.ActionDelay
		case 11:
			Config.SfxVolume--
			if Config.SfxVolume < 0 {
				Config.SfxVolume = 0
			}
			audio.SetSfxVolume(Config.SfxVolume)
		case 12:
			os.cycleLang(-1)
		}
	case tcell.KeyRight:
		audio.PlayMenuNav()
		switch os.Selection {
		case 9:
			os.cycleTheme(1)
		case 10:
			Config.ActionDelay++
			if Config.ActionDelay > 20 {
				Config.ActionDelay = 20
			}
			os.Game.ActionDelay = Config.ActionDelay
		case 11:
			Config.SfxVolume++
			if Config.SfxVolume > 10 {
				Config.SfxVolume = 10
			}
			audio.SetSfxVolume(Config.SfxVolume)
		case 12:
			os.cycleLang(1)
		}
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
	}
}

func (os *OptionsScreen) cycleLang(dir int) {
	langs := language.Available()
	idx := 0
	for i, l := range langs {
		if l == language.Current() {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(langs) - 1
	}
	if idx >= len(langs) {
		idx = 0
	}
	language.SetLanguage(langs[idx])
	Config.Language = langs[idx]
}

var themes = []string{"default", "high_contrast", "amber", "green", "paper"}

var themeDisplayNames = map[string]string{
	"default":       "Default",
	"high_contrast": "HighCon",
	"amber":         "Amber",
	"green":         "Green",
	"paper":         "Paper",
}

func (os *OptionsScreen) cycleTheme(dir int) {
	idx := 0
	for i, t := range themes {
		if t == Config.Theme {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(themes) - 1
	}
	if idx >= len(themes) {
		idx = 0
	}
	Config.Theme = themes[idx]
	os.Game.screen.SetTheme(Config.Theme)
}

func (os *OptionsScreen) HandleMouse(e *tcell.EventMouse) {
	// Not implemented for options yet, just return
}

func (os *OptionsScreen) Update() {
	// Nothing to update
}
