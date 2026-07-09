package engine

import (
	"fmt"

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

	options := []struct {
		Label string
		Value *bool
	}{
		{language.String("OPTIONS_BLOOM"), &Config.BloomEnabled},
		{language.String("OPTIONS_DISTORTION"), &Config.DistortionEnabled},
		{language.String("OPTIONS_LIGHTING"), &Config.LightingEnabled},
	}

	for i, opt := range options {
		style := StyleDefault
		if i == os.Selection {
			style = StyleHighlight
		}

		status := language.String("OPTIONS_OFF")
		if *opt.Value {
			status = language.String("OPTIONS_ON")
		}

		line := fmt.Sprintf("[%s] %s", status, opt.Label)
		ctx.DrawString(w/2-15, h/2-3+i, line, style)
	}

	// Language option
	langY := h/2 - 3 + len(options)
	langStyle := StyleDefault
	if os.Selection == len(options) {
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
	line := fmt.Sprintf("      %s: [%s]", language.String("OPTIONS_LANGUAGE"), langs[langIdx])
	ctx.DrawString(w/2-15, langY+1, line, langStyle)

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, "[\u2190]/[\u2192]=Language  [\u2191]/[\u2193]=Select  Enter=Toggle  [Esc]=Back", StyleGray, StyleHotkey)
}

func (os *OptionsScreen) HandleKey(e *tcell.EventKey) {
	totalOptions := 4 // 3 toggles + 1 language
	switch e.Key() {
	case tcell.KeyUp:
		os.Selection--
		if os.Selection < 0 {
			os.Selection = totalOptions - 1
		}
	case tcell.KeyDown:
		os.Selection++
		if os.Selection >= totalOptions {
			os.Selection = 0
		}
	case tcell.KeyEnter:
		os.toggle()
	case tcell.KeyLeft:
		if os.Selection == 3 {
			os.cycleLang(-1)
		}
	case tcell.KeyRight:
		if os.Selection == 3 {
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
	case 3:
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
