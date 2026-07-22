package engine

import (
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type HelpScreen struct {
	Game *Game
	Page int
}

func NewHelpScreen(g *Game, prev GameState) *HelpScreen {
	page := 0
	switch prev {
	case StateGeoscape:
		page = 0
	case StateBase, StateEquip, StateResearch, StateManufacture:
		page = 1
	case StateBattlescape:
		page = 2
	case StateDogfight:
		page = 5
	}
	return &HelpScreen{Game: g, Page: page}
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
				">" + language.String("HELP_GEO_MISSION"),
				">" + language.String("HELP_GEO_TRANSPORT"),
				">" + language.String("HELP_GEO_CYCLE"),
				">" + language.String("HELP_GEO_NEW"),
				">" + language.String("HELP_GEO_TRANSFER"),
				">" + language.String("HELP_GEO_ENCYCLOPEDIA"),
				">" + language.String("HELP_GEO_RADAR"),
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
		{
			title: language.String("HELP_DOG_TITLE"),
			lines: []string{
				"#" + language.String("HELP_DOG_CONTROLS"),
				">" + language.String("HELP_DOG_FIRE"),
				">" + language.String("HELP_DOG_CLOSE"),
				">" + language.String("HELP_DOG_FAR"),
				">" + language.String("HELP_DOG_BREAK"),
				">" + language.String("HELP_DOG_ESC"),
				"",
				"#" + language.String("HELP_DOG_TIPS"),
				">" + language.String("HELP_DOG_T1"),
				">" + language.String("HELP_DOG_T2"),
				">" + language.String("HELP_DOG_T3"),
				">" + language.String("HELP_DOG_T4"),
			},
		},
	}
}

func (hs *HelpScreen) HandleKey(e *tcell.EventKey) {
	pages := hs.getPages()
	switch e.Key() {
	case tcell.KeyRight, tcell.KeyTab:
		hs.Page++
		if hs.Page >= len(pages) {
			hs.Page = 0
		}
	case tcell.KeyLeft:
		hs.Page--
		if hs.Page < 0 {
			hs.Page = len(pages) - 1
		}
	}
	if e.Str() >= "1" && e.Str() <= "9" {
		if p := int(e.Str()[0] - '1'); p >= 0 && p < len(pages) {
			hs.Page = p
		}
	}
}

func (hs *HelpScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := hs.Game.ScreenSize()

	// Help bar clicks
	if y == h-1 {
		nav := language.String("HELP_NAV")
		col := 1
		runes := []rune(nav)
		for i := 0; i < len(runes); {
			if runes[i] != '[' {
				col += StringWidth(string(runes[i]))
				i++
				continue
			}
			segStart := col
			end := i + 1
			for end < len(runes) && runes[end] != ']' {
				end++
			}
			if end >= len(runes) {
				break
			}
			segEnd := col + StringWidth(string(runes[i:end+1]))
			if x >= segStart && x <= segEnd {
				key := string(runes[i+1 : end])
				if key == "Esc" {
					hs.Game.PopState()
					return
				}
				if key == "Tab" {
					hs.Page++
					if pages := hs.getPages(); hs.Page >= len(pages) {
						hs.Page = 0
					}
					return
				}
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	pages := hs.getPages()
	if x < w/2 {
		hs.Page--
		if hs.Page < 0 {
			hs.Page = len(pages) - 1
		}
	} else {
		hs.Page++
		if hs.Page >= len(pages) {
			hs.Page = 0
		}
	}
}
