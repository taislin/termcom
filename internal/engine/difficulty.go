package engine

import (
	"fmt"

	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

const startingFunds = 500000

type DifficultyEntry struct {
	Name        string
	Description string
	NameKey     string
	DescKey     string
	AlienScale  float64 // multiplier for alien stats
	UFOScale    float64 // multiplier for UFO spawn rate & count
	FundsScale  float64 // starting funds multiplier
}

func (d *DifficultyEntry) LangName() string {
	if d.NameKey != "" {
		return language.String(d.NameKey)
	}
	return d.Name
}

func (d *DifficultyEntry) LangDesc() string {
	if d.DescKey != "" {
		return language.String(d.DescKey)
	}
	return d.Description
}

var Difficulties = []DifficultyEntry{
	{Name: "Beginner", Description: "Weaker aliens, slower UFOs, more funds", NameKey: "DIFF_BEGINNER", DescKey: "DIFF_BEGINNER_DESC", AlienScale: 0.7, UFOScale: 0.7, FundsScale: 1.5},
	{Name: "Experienced", Description: "Standard challenge", NameKey: "DIFF_EXPERIENCED", DescKey: "DIFF_EXPERIENCED_DESC", AlienScale: 1.0, UFOScale: 1.0, FundsScale: 1.0},
	{Name: "Veteran", Description: "Stronger aliens, faster UFOs", NameKey: "DIFF_VETERAN", DescKey: "DIFF_VETERAN_DESC", AlienScale: 1.2, UFOScale: 1.3, FundsScale: 0.8},
	{Name: "Genius", Description: "Much harder combat and economy", NameKey: "DIFF_GENIUS", DescKey: "DIFF_GENIUS_DESC", AlienScale: 1.5, UFOScale: 1.6, FundsScale: 0.6},
	{Name: "Superhuman", Description: "Maximum alien threat", NameKey: "DIFF_SUPERHUMAN", DescKey: "DIFF_SUPERHUMAN_DESC", AlienScale: 2.0, UFOScale: 2.0, FundsScale: 0.5},
}

type DifficultyScreen struct {
	Game       *Game
	Selection  int
	OnConfirm  func(difficulty int)
}

func NewDifficultyScreen(g *Game, onConfirm func(int)) *DifficultyScreen {
	return &DifficultyScreen{
		Game:      g,
		Selection: 1, // default to Experienced
		OnConfirm: onConfirm,
	}
}

func (ds *DifficultyScreen) Update() {}

func (ds *DifficultyScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("DIFFICULTY_TITLE"), StyleDefault)

	title := language.String("DIFFICULTY_PROMPT")
	ctx.DrawString(2, 2, title, StyleCyanBold)

	for i, d := range Difficulties {
		if 4+i >= h-3 {
			break
		}
		style := StyleDefault
		if i == ds.Selection {
			style = StyleHighlight
		}
		line := fmt.Sprintf(language.String("DIFFICULTY_LINE_FORMAT"), d.LangName(), d.LangDesc())
		ctx.DrawString(2, 4+i, line, style)
	}

	help := language.String("HELP_DIFFICULTY")
	ctx.DrawMarkupString(2, h-1, help, StyleGray, StyleHotkey)
}

func (ds *DifficultyScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		ds.Selection--
		if ds.Selection < 0 {
			ds.Selection = len(Difficulties) - 1
		}
	case tcell.KeyDown:
		ds.Selection++
		if ds.Selection >= len(Difficulties) {
			ds.Selection = 0
		}
	case tcell.KeyEnter:
		ds.Game.Difficulty = ds.Selection
		ds.Game.Funds = int64(float64(startingFunds) * Difficulties[ds.Selection].FundsScale)
		if ds.OnConfirm != nil {
			ds.OnConfirm(ds.Selection)
		}
	}
	switch e.Str() {
	case "\r":
		ds.Game.Difficulty = ds.Selection
		ds.Game.Funds = int64(float64(startingFunds) * Difficulties[ds.Selection].FundsScale)
		if ds.OnConfirm != nil {
			ds.OnConfirm(ds.Selection)
		}
	}
}

func (ds *DifficultyScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, _ := ds.Game.ScreenSize()

	if y >= 4 && y < 4+len(Difficulties) && x < w {
		ds.Selection = y - 4
		ds.Game.Difficulty = ds.Selection
		ds.Game.Funds = int64(float64(startingFunds) * Difficulties[ds.Selection].FundsScale)
		if ds.OnConfirm != nil {
			ds.OnConfirm(ds.Selection)
		}
	}
}
