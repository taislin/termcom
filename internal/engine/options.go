package engine

import (
	"fmt"

	"github.com/civ13/ycom/internal/audio"
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

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

	boolOpts := []struct {
		Label string
		Value *bool
	}{
		{language.String("OPTIONS_BLOOM"), &Config.BloomEnabled},
		{language.String("OPTIONS_DISTORTION"), &Config.DistortionEnabled},
		{language.String("OPTIONS_LIGHTING"), &Config.LightingEnabled},
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

		line := fmt.Sprintf("[%s] %s", status, opt.Label)
		ctx.DrawString(w/2-15, h/2-4+i, line, style)
	}

	// Resolution speed option (int slider)
	speedIdx := len(boolOpts)
	speedStyle := StyleDefault
	if os.Selection == speedIdx {
		speedStyle = StyleHighlight
	}
	speed := os.Game.ActionDelay
	line := fmt.Sprintf("%s: %d", language.String("OPTIONS_RESOLUTION_SPEED"), speed)
	ctx.DrawString(w/2-15, h/2-4+speedIdx, line, speedStyle)

	// Language option
	langY := h/2 - 4 + speedIdx + 1
	langStyle := StyleDefault
	if os.Selection == speedIdx+1 {
		langStyle = StyleHighlight
	}
	langs := language.Available()
	langIdx := 0
	for i, l := range langs {
		if l == language.Current() {
			langIdx = i
			break
		}
	}
	line = fmt.Sprintf("      %s: [%s]", language.String("OPTIONS_LANGUAGE"), langs[langIdx])
	ctx.DrawString(w/2-15, langY+1, line, langStyle)

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, "[\u2190]/[\u2192]=Adjust  [\u2191]/[\u2193]=Select  Enter=Toggle  [Esc]=Back", StyleGray, StyleHotkey)
}

func (os *OptionsScreen) HandleKey(e *tcell.EventKey) {
	speedIdx := 3 // 3 bool toggles precede the speed option
	totalOptions := 5
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
		if os.Selection == speedIdx {
			os.Game.ActionDelay--
			if os.Game.ActionDelay < 1 {
				os.Game.ActionDelay = 1
			}
		} else if os.Selection == speedIdx+1 {
			os.cycleLang(-1)
		}
	case tcell.KeyRight:
		audio.PlayMenuNav()
		if os.Selection == speedIdx {
			os.Game.ActionDelay++
			if os.Game.ActionDelay > 20 {
				os.Game.ActionDelay = 20
			}
		} else if os.Selection == speedIdx+1 {
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
		Config.DistortionEnabled = !Config.DistortionEnabled
	case 2:
		Config.LightingEnabled = !Config.LightingEnabled
	case 4:
		os.cycleLang(1)
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

func (os *OptionsScreen) HandleMouse(e *tcell.EventMouse) {
	// Not implemented for options yet, just return
}

func (os *OptionsScreen) Update() {
	// Nothing to update
}
