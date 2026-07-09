package engine

import (
	"github.com/gdamore/tcell/v3"
)

type GameOverScreen struct {
	Game *Game
	Won  bool
	Stats string
}

func NewGameOverScreen(g *Game, won bool, stats string) *GameOverScreen {
	return &GameOverScreen{
		Game:  g,
		Won:   won,
		Stats: stats,
	}
}

func (gos *GameOverScreen) Update() {}

func (gos *GameOverScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	title := "GAME OVER"
	if gos.Won {
		title = "VICTORY"
	}
	ctx.DrawString(w/2-len(title)/2, h/2-2, title, StyleRedBold)
	ctx.DrawString(w/2-len(gos.Stats)/2, h/2, gos.Stats, StyleDefault)
	ctx.DrawString(w/2-10, h/2+2, "Press ESC to Quit", StyleGray)
}

func (gos *GameOverScreen) HandleKey(e *tcell.EventKey) {
	if e.Key() == tcell.KeyEscape {
		gos.Game.Quit()
	}
}

func (gos *GameOverScreen) HandleMouse(e *tcell.EventMouse) {}
