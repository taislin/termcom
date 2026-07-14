package engine

import (
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type tutorialStep struct {
	titleKey string
	msgKey   string
}

var tutorialSteps = []tutorialStep{
	{"TUTORIAL_TITLE", "TUTORIAL_WELCOME"},
	{"TUTORIAL_TITLE", "TUTORIAL_TIME"},
	{"TUTORIAL_TITLE", "TUTORIAL_UFO"},
	{"TUTORIAL_TITLE", "TUTORIAL_INTERCEPT"},
	{"TUTORIAL_TITLE", "TUTORIAL_MISSION"},
	{"TUTORIAL_TITLE", "TUTORIAL_BASE"},
	{"TUTORIAL_TITLE", "TUTORIAL_BATTLE"},
	{"TUTORIAL_TITLE", "TUTORIAL_COMPLETE"},
}

type TutorialScreen struct {
	Game      *Game
	step      int
	OnDismiss func()
}

func NewTutorialScreen(g *Game, onDismiss func()) *TutorialScreen {
	return &TutorialScreen{
		Game:      g,
		step:      0,
		OnDismiss: onDismiss,
	}
}

func (ts *TutorialScreen) Update() {}

func (ts *TutorialScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	boxW := 62
	boxH := 14
	x := (w - boxW) / 2
	y := (h - boxH) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	for fy := y; fy < y+boxH; fy++ {
		for fx := x; fx < x+boxW; fx++ {
			ctx.SetCell(fx, fy, ' ', StyleDefault)
		}
	}

	title := language.String(tutorialSteps[ts.step].titleKey)
	ctx.DrawPanel(x, y, boxW, boxH, title, StyleCyanBold)

	msg := language.String(tutorialSteps[ts.step].msgKey)
	wrapDrawString(ctx, x+2, y+2, boxW-4, msg, StyleDefault)

	progress := ""
	for i := 0; i < len(tutorialSteps); i++ {
		if i <= ts.step {
			progress += "\u2588 "
		} else {
			progress += "\u2591 "
		}
	}
	progW := len(progress)
	ctx.DrawString(x+(boxW-progW)/2, y+boxH-4, progress, StyleGray)

	if ts.step < len(tutorialSteps)-1 {
		hint := language.String("TUTORIAL_NEXT")
		ctx.DrawMarkupString(x+2, y+boxH-2, hint, StyleGray, StyleHotkey)
	} else {
		hint := language.String("TUTORIAL_DISMISS")
		ctx.DrawMarkupString(x+2, y+boxH-2, hint, StyleGray, StyleHotkey)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, language.String("TUTORIAL_SKIP"), StyleGray, StyleHotkey)
}

func (ts *TutorialScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyEnter:
		if ts.step < len(tutorialSteps)-1 {
			ts.step++
		} else {
			ts.dismiss()
		}
	case tcell.KeyEscape:
		ts.dismiss()
	}
	switch e.Str() {
	case "s", "S":
		ts.dismiss()
	}
}

func (ts *TutorialScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	if buttons&tcell.Button1 != 0 {
		if ts.step < len(tutorialSteps)-1 {
			ts.step++
		} else {
			ts.dismiss()
		}
	}
}

func (ts *TutorialScreen) dismiss() {
	Config.TutorialShown = true
	SaveConfig()
	ts.Game.PopState()
	if ts.OnDismiss != nil {
		ts.OnDismiss()
	}
}

func wrapDrawString(ctx *ScreenCtx, x, y, maxWidth int, s string, style tcell.Style) {
	runes := []rune(s)
	lineStart := 0
	curY := y
	lastSpace := -1
	col := 0

	for i, ch := range runes {
		if ch == ' ' {
			lastSpace = i
		}
		col++
		if col >= maxWidth {
			if lastSpace > lineStart {
				ctx.DrawString(x, curY, string(runes[lineStart:lastSpace]), style)
				lineStart = lastSpace + 1
				lastSpace = -1
			} else {
				ctx.DrawString(x, curY, string(runes[lineStart:i]), style)
				lineStart = i
			}
			curY++
			col = i - lineStart
			if col < 0 {
				col = 0
			}
		}
	}
	if lineStart < len(runes) {
		ctx.DrawString(x, curY, string(runes[lineStart:]), style)
	}
}
