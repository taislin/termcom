package engine

import (
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type HelpScreen struct {
	Game *Game
	Page int
}

func NewHelpScreen(g *Game) *HelpScreen {
	return &HelpScreen{Game: g, Page: 0}
}

func (hs *HelpScreen) Update() {}

func (hs *HelpScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("HELP_TITLE"), StyleDefault)

	pages := hs.getPages()
	if hs.Page >= len(pages) {
		hs.Page = 0
	}

	title := pages[hs.Page].title
	lines := pages[hs.Page].lines

	ctx.DrawString(2, 2, title, StyleCyanBold)
	for i, line := range lines {
		if 4+i >= h-2 {
			break
		}
		style := StyleDefault
		if len(line) > 0 && line[0] == '>' {
			line = line[1:]
			style = StyleGreen
		} else if len(line) > 0 && line[0] == '#' {
			line = line[1:]
			style = StyleYellow
		}
		ctx.DrawString(2, 4+i, line, style)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawString(1, h-1, language.String("HELP_NAV"), StyleGray)
}

type helpPage struct {
	title string
	lines []string
}

func (hs *HelpScreen) getPages() []helpPage {
	return []helpPage{
		{
			title: language.String("HELP_GEO_TITLE"),
			lines: []string{
				"#" + language.String("HELP_GEO_CONTROLS"),
				">" + language.String("HELP_GEO_MOVE"),
				">" + language.String("HELP_GEO_PAUSE"),
				">" + language.String("HELP_GEO_SPEED"),
				">" + language.String("HELP_GEO_BASE"),
				">" + language.String("HELP_GEO_LAUNCH"),
				">" + language.String("HELP_GEO_AUTO"),
				">" + language.String("HELP_GEO_SAVE"),
				">" + language.String("HELP_GEO_LOAD"),
				">" + language.String("HELP_GEO_QUIT"),
				"",
				"#" + language.String("HELP_GEO_GAMEPLAY"),
				">" + language.String("HELP_GEO_G1"),
				">" + language.String("HELP_GEO_G2"),
				">" + language.String("HELP_GEO_G3"),
				">" + language.String("HELP_GEO_G4"),
				">" + language.String("HELP_GEO_G5"),
				">" + language.String("HELP_GEO_G6"),
				">" + language.String("HELP_GEO_G7"),
			},
		},
		{
			title: language.String("HELP_BASE_TITLE"),
			lines: []string{
				"#" + language.String("HELP_BASE_CONTROLS"),
				">" + language.String("HELP_BASE_TABS"),
				">" + language.String("HELP_BASE_NAV"),
				">" + language.String("HELP_BASE_BACK"),
				"",
				"#" + language.String("HELP_BASE_NAMES"),
				">" + language.String("HELP_BASE_FAC"),
				">" + language.String("HELP_BASE_SOLD"),
				">" + language.String("HELP_BASE_RES"),
				">" + language.String("HELP_BASE_MFG"),
				">" + language.String("HELP_BASE_STORE"),
				"",
				"#" + language.String("HELP_BASE_KEYS"),
				">" + language.String("HELP_BASE_K1"),
				">" + language.String("HELP_BASE_K2"),
				">" + language.String("HELP_BASE_K3"),
				">" + language.String("HELP_BASE_K4"),
				">" + language.String("HELP_BASE_K5"),
				">" + language.String("HELP_BASE_K6"),
				">" + language.String("HELP_BASE_K7"),
			},
		},
		{
			title: language.String("HELP_BAT_TITLE"),
			lines: []string{
				"#" + language.String("HELP_BAT_CONTROLS"),
				">" + language.String("HELP_BAT_MOVE"),
				">" + language.String("HELP_BAT_SELECT"),
				">" + language.String("HELP_BAT_CYCLE"),
				">" + language.String("HELP_BAT_FIRE"),
				">" + language.String("HELP_BAT_RELOAD"),
				">" + language.String("HELP_BAT_END"),
				">" + language.String("HELP_BAT_CROUCH"),
				">" + language.String("HELP_BAT_GRENADE"),
				">" + language.String("HELP_BAT_MEDIKIT"),
				">" + language.String("HELP_BAT_WHEEL"),
				">" + language.String("HELP_BAT_LCLICK"),
				">" + language.String("HELP_BAT_RCLICK"),
				"",
				"#" + language.String("HELP_BAT_MECHANICS"),
				">" + language.String("HELP_BAT_M1"),
				">" + language.String("HELP_BAT_M2"),
				">" + language.String("HELP_BAT_M3"),
				">" + language.String("HELP_BAT_M4"),
				">" + language.String("HELP_BAT_M5"),
			},
		},
		{
			title: language.String("HELP_RES_TITLE"),
			lines: []string{
				"#" + language.String("HELP_RES_RESEARCH"),
				">" + language.String("HELP_RES_R1"),
				">" + language.String("HELP_RES_R2"),
				">" + language.String("HELP_RES_R3"),
				">" + language.String("HELP_RES_R4"),
				"#" + language.String("HELP_RES_MFG"),
				">" + language.String("HELP_RES_M1"),
				">" + language.String("HELP_RES_M2"),
				">" + language.String("HELP_RES_M3"),
				">" + language.String("HELP_RES_M4"),
				"",
				"#" + language.String("HELP_RES_ORDER"),
				">" + language.String("HELP_RES_O1"),
				">" + language.String("HELP_RES_O2"),
				">" + language.String("HELP_RES_O3"),
			},
		},
		{
			title: language.String("HELP_STRAT_TITLE"),
			lines: []string{
				">" + language.String("HELP_STRAT_1"),
				">" + language.String("HELP_STRAT_2"),
				">" + language.String("HELP_STRAT_3"),
				">" + language.String("HELP_STRAT_4"),
				">" + language.String("HELP_STRAT_5"),
				">" + language.String("HELP_STRAT_6"),
				">" + language.String("HELP_STRAT_7"),
				">" + language.String("HELP_STRAT_8"),
				">" + language.String("HELP_STRAT_9"),
				">" + language.String("HELP_STRAT_10"),
				"",
				"#" + language.String("HELP_STRAT_VICTORY"),
				">" + language.String("HELP_STRAT_V1"),
				">" + language.String("HELP_STRAT_V2"),
				">" + language.String("HELP_STRAT_V3"),
			},
		},
	}
}

func (hs *HelpScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyRight, tcell.KeyTab:
		pages := hs.getPages()
		hs.Page++
		if hs.Page >= len(pages) {
			hs.Page = 0
		}
	case tcell.KeyLeft:
		hs.Page--
		if hs.Page < 0 {
			pages := hs.getPages()
			hs.Page = len(pages) - 1
		}
	}
	switch e.Str() {
	case "1":
		hs.Page = 0
	case "2":
		hs.Page = 1
	case "3":
		hs.Page = 2
	case "4":
		hs.Page = 3
	case "5":
		hs.Page = 4
	}
}

func (hs *HelpScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, _ := e.Position()
	w, _ := hs.Game.ScreenSize()
	if x < w/2 {
		hs.Page--
		if hs.Page < 0 {
			pages := hs.getPages()
			hs.Page = len(pages) - 1
		}
	} else {
		pages := hs.getPages()
		hs.Page++
		if hs.Page >= len(pages) {
			hs.Page = 0
		}
	}
}
