package engine

import (
	"os"

	"github.com/gdamore/tcell/v2"
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
		" ____    __    ____  _______  __       __   __ ",
		"/\\  _`\\ /\\ \\  /\\  _\\/\\  ___\\/\\ \\     /\\ \\ / / ",
		"\\ \\ ,__\\ \\ \\ \\ \\ \\__\\ \\___/\\ \\ \\____\\ \\ '/ /  ",
		" \\ \\  _\\ \\ \\ \\ \\ \\__/\\ \\    \\ \\  __`\\ \\ , <   ",
		"  \\ \\ \\  \\ \\_/ /\\ \\  \\ \\    \\ \\ \\_/ \\ \\ \\`\\  ",
		"   \\ \\_/  \\___/  \\ \\  \\ \\    \\ \\____/\\ \\_\\ \\_",
		"   \\ /    \\___/    \\_/  \\_/    \\___/  \\/_/\\/_/",
	}

	startY := 2
	for i, line := range title {
		x := (w - len(line)) / 2
		if x < 0 {
			x = 0
		}
		style := StyleDefault
		switch {
		case i == 0 || i == 6:
			style = StyleMagenta
		case i == 1 || i == 5:
			style = StyleMagenta.Bold(true)
		case i == 2 || i == 3 || i == 4:
			style = StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true)
		}
		ctx.DrawString(x, startY+i, line, style)
	}

	subY := startY + len(title) + 1
	subtitle := "U F O   D E F E N S E"
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
	ctx.DrawString(decX, subY+2, deco, StyleGray)

	verX := (w - 17) / 2
	if verX < 0 {
		verX = 0
	}
	ctx.DrawString(verX, subY+3, "ASCII Demake v0.1", StyleGray)

	menuY := subY + 6
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

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawString(1, h-1, "j/k=Select  Enter=Confirm  Q=Quit", StyleGray)
}

func (ms *MenuScreen) options() []string {
	if HasSave() {
		return []string{"New Game", "Continue", "Quit"}
	}
	return []string{"New Game", "Quit"}
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
	case tcell.KeyRune:
		switch e.Rune() {
		case 'j':
			ms.Selection++
			if ms.Selection > maxSel {
				ms.Selection = 0
			}
		case 'k':
			ms.Selection--
			if ms.Selection < 0 {
				ms.Selection = maxSel
			}
		case 'q', 'Q':
			ms.Game.Quit()
		case '1':
			ms.Selection = 0
			ms.confirm()
		case '2':
			if HasSave() {
				ms.Selection = 1
				ms.confirm()
			}
		case '3':
			if HasSave() {
				ms.Selection = 2
				ms.confirm()
			}
		}
	}
}

func (ms *MenuScreen) confirm() {
	opts := ms.options()
	if ms.Selection < 0 || ms.Selection >= len(opts) {
		return
	}
	switch opts[ms.Selection] {
	case "New Game":
		if ms.Game.OnNewGame != nil {
			ms.Game.OnNewGame()
		}
	case "Continue":
		if ms.Game.OnContinue != nil {
			ms.Game.OnContinue()
		}
	case "Quit":
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

	menuY := 16
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
