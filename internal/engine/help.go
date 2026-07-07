package engine

import "github.com/gdamore/tcell/v2"

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
	ctx.DrawPanel(0, 0, w, h, "HELP", StyleDefault)

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
	ctx.DrawString(1, h-1, "Left/Right or Tab=Page  Esc=Back", StyleGray)
}

type helpPage struct {
	title string
	lines []string
}

func (hs *HelpScreen) getPages() []helpPage {
	return []helpPage{
		{
			title: "GEOSCAPE",
			lines: []string{
				"#Controls:",
				">Arrow keys   Move camera",
				">Space        Pause/unpause time",
				">1-4          Time compression (1x/5x/20x/60x)",
				">B            Open base management",
				">L            Launch interceptor at nearest UFO",
				">A            Autoresolve nearest UFO",
				">F5           Save game",
				">F9           Load game",
				">Q            Quit",
				"",
				"#Gameplay:",
				">UFOs appear and fly toward cities",
				">Launch interceptors to shoot them down",
				">Downed UFOs trigger tactical battles",
				">Win battles to earn XP, loot, and funds",
				">Alien missions increase Alien Activity",
				">Reach 100% Alien Activity to lose",
				">Win 10 battles to save Earth",
			},
		},
		{
			title: "BASE MANAGEMENT",
			lines: []string{
				"#Controls:",
				">1-5 or Left/Right   Switch tabs",
				">j/k or Up/Down      Navigate",
				">Esc                 Back to geoscape",
				"",
				"#Tabs:",
				">Facilities  Build/sell base facilities",
				">Soldiers    Hire/dismiss, view roster",
				">Research    Assign scientists to topics",
				">Manufacture Queue item production",
				">Stores      View inventory",
				"",
				"#Key Bindings:",
				">B   Build selected facility",
				">S   Sell selected facility (50% refund)",
				">H   Hire new soldier ($50K)",
				">E   Open equip screen (Soldiers tab)",
				">D   Dismiss selected soldier",
				">R   Open research screen (Research tab)",
				">M   Open manufacture screen (Manufacture tab)",
			},
		},
		{
			title: "BATTLESCAPE",
			lines: []string{
				"#Controls:",
				">hjkl / Arrow keys   Move cursor",
				">Space / Enter       Select unit / Confirm move",
				">s                   Cycle soldiers",
				">f                   Fire at target",
				">r                   Reload weapon",
				">e / n               End turn",
				">c                   Crouch (cover bonus)",
				">g                   Throw grenade",
				">m                   Use medikit",
				">Mouse wheel          Cycle soldiers",
				">Left click           Select/move",
				">Right click          Fire",
				"",
				"#Mechanics:",
				">Actions cost Time Units (TU)",
				">Line of sight blocked by walls/trees",
				">Crouching gives +10% accuracy, -30% damage",
				">Killing aliens earns XP for soldiers",
				">Surviving soldiers keep HP and gains",
			},
		},
		{
			title: "RESEARCH & MANUFACTURING",
			lines: []string{
				"#Research:",
				">Select topics to research with scientists",
				">Prerequisites must be completed first",
				">Unlocks new weapons, armor, and items",
				">Progress = scientists per day",
				"",
				"#Manufacturing:",
				">Queue items to build with engineers",
				">Costs materials (alloys, elerium) from stores",
				">Completed items go to stores",
				">Engineers speed up production",
				"",
				"#Research Order (suggested):",
				">1. Alien Alloys → 2. Elerium → 3. Laser Weapons",
				">4. Sectoid Autopsy → 5. Plasma Weapons",
				">6. Personal Armour → 7. Light/Medium Suits",
			},
		},
		{
			title: "STRATEGY TIPS",
			lines: []string{
				">Hire soldiers early — you need a squad",
				">Build Living Quarters for more soldiers",
				">Build Labs before Workshops",
				">Research alien alloys first for better gear",
				">Equip soldiers before sending to battle",
				">Wounded soldiers heal 2 HP per day",
				">Use autoresolve for easy interceptions",
				">Radar facilities increase monthly funding",
				">Sell excess alien artifacts for cash",
				">Alien missions increase activity — intercept!",
				"",
				"#Victory Condition:",
				">Victory: Win 10 battles to save Earth",
				">Defeat: Alien Activity reaches 100%",
				">Unresolved missions raise activity +10%",
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
	case tcell.KeyRune:
		if e.Rune() == '1' {
			hs.Page = 0
		} else if e.Rune() == '2' {
			hs.Page = 1
		} else if e.Rune() == '3' {
			hs.Page = 2
		} else if e.Rune() == '4' {
			hs.Page = 3
		} else if e.Rune() == '5' {
			hs.Page = 4
		}
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
