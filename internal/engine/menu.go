package engine

import (
	"os"

	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

const SaveFile = "xcom_save.json"

type MenuScreen struct {
	Game      *Game
	Selection int
}

func NewMenuScreen(g *Game) *MenuScreen {
	return &MenuScreen{Game: g, Selection: 0}
}

func HasSave() bool {
	_, err := os.Stat(SaveFile)
	return err == nil
}

func (ms *MenuScreen) Update() {}

func (ms *MenuScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	title := []string{
		" _________ _______  _______  _______         _______  _______  _______ ",
		" \\__   __/(  ____ \\(  ____ )(       )       (  ____ \\(  ___  )(       )",
		"    ) (   | (    \\/| (    )|| () () |       | (    \\/| (   ) || () () |",
		"    | |   | (__    | (____)|| || || | _____ | |      | |   | || || || |",
		"    | |   |  __)   |     __)| |(_)| |(_____)| |      | |   | || |(_)| |",
		"    | |   | (      | (\\ (   | |   | |       | |      | |   | || |   | |",
		"    | |   | (____/\\| ) \\ \\__| )   ( |       | (____/\\| (___) || )   ( |",
		"    )_(   (_______/|/   \\__/|/     \\|       (_______/(_______)|/     \\|",
		"                                                                         ",
	}

	// Purple gradient from light (top) to dark (bottom)
	purpleGradient := []tcell.Color{
		tcell.GetColor("#E0FF"), // lightest
		tcell.GetColor("#D0EF"),
		tcell.GetColor("#C0DF"),
		tcell.GetColor("#B0CF"),
		tcell.GetColor("#A0BF"),
		tcell.GetColor("#90AF"),
		tcell.GetColor("#809F"),
		tcell.GetColor("#708F"),
		tcell.GetColor("#607F"), // darkest
	}

	startY := 2
	for i, line := range title {
		x := (w - len(line)) / 2
		if x < 0 {
			x = 0
		}
		style := StyleDefault.Foreground(purpleGradient[i]).Bold(true)
		ctx.DrawString(x, startY+i, line, style)
	}

	subY := startY + len(title) + 1
	subtitle := language.String("MENU_TITLE")
	subX := (w - len(subtitle)) / 2
	if subX < 0 {
		subX = 0
	}
	ctx.DrawString(subX, subY, subtitle, StyleCyanBold)

	deco := "==================================================="
	decX := (w - len(deco)) / 2
	if decX < 0 {
		decX = 0
	}
	ctx.DrawString(decX, subY-1, deco, StyleGray)
	ctx.DrawString(decX, subY+1, deco, StyleGray)

	verX := (w - 9) / 2
	if verX < 0 {
		verX = 0
	}
	ctx.DrawString(verX, subY+3, language.String("MENU_SUBTITLE"), StyleGray)

	menuY := subY + 8
	options := ms.options()

	for i, opt := range options {
		style := StyleDefault
		if i == ms.Selection {
			style = StyleHighlight
			prefix := "> "
			ctx.DrawString(w/2-10, menuY+i*2, prefix, StyleMagenta)
		}
		ctx.DrawString(w/2-8, menuY+i*2, opt, style)
	}

	ctx.DrawPanel(0, h-3, w, 3, "", StyleGray)
	ctx.DrawString(1, h-2, language.String("MENU_HELP"), StyleGray)
}

func (ms *MenuScreen) options() []string {
	if HasSave() {
		return []string{language.String("MENU_NEW_GAME"), language.String("MENU_CONTINUE"), language.String("MENU_QUIT")}
	}
	return []string{language.String("MENU_NEW_GAME"), language.String("MENU_QUIT")}
}

func (ms *MenuScreen) HandleKey(e *tcell.EventKey) {
	opts := ms.options()
	maxSel := len(opts) - 1

	switch e.Key() {
	case tcell.KeyUp:
		ms.Selection--
		if ms.Selection < 0 {
			ms.Selection = maxSel
		}
	case tcell.KeyDown:
		ms.Selection++
		if ms.Selection > maxSel {
			ms.Selection = 0
		}
	case tcell.KeyEnter:
		ms.confirm()
	}
	switch e.Str() {
	case "j":
		ms.Selection++
		if ms.Selection > maxSel {
			ms.Selection = 0
		}
	case "k":
		ms.Selection--
		if ms.Selection < 0 {
			ms.Selection = maxSel
		}
	case "q", "Q":
		ms.Game.Quit()
	case "1":
		ms.Selection = 0
		ms.confirm()
	case "2":
		if HasSave() {
			ms.Selection = 1
			ms.confirm()
		}
	case "3":
		if HasSave() {
			ms.Selection = 2
			ms.confirm()
		}
	}
}

func (ms *MenuScreen) confirm() {
	opts := ms.options()
	if ms.Selection < 0 || ms.Selection >= len(opts) {
		return
	}
	switch opts[ms.Selection] {
	case language.String("MENU_NEW_GAME"):
		if ms.Game.OnNewGame != nil {
			ms.Game.OnNewGame()
		}
	case language.String("MENU_CONTINUE"):
		if ms.Game.OnContinue != nil {
			ms.Game.OnContinue()
		}
	case language.String("MENU_QUIT"):
		ms.Game.Quit()
	}
}

func (ms *MenuScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, _ := ms.Game.ScreenSize()

	subY := 10
	menuY := subY + 8
	opts := ms.options()

	for i := range opts {
		if y == menuY+i*2 && x >= w/2-10 && x <= w/2+10 {
			ms.Selection = i
			if buttons&tcell.Button1 != 0 {
				ms.confirm()
			}
			return
		}
	}
}
