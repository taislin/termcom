package engine

import (
	"math"
	"time"

	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type langEntry struct {
	Code string
	Name string
}

var langList = []langEntry{
	{"en", "English"},
	{"es", "Español"},
	{"pt", "Português"},
	{"fr", "Français"},
	{"ru", "Русский"},
	{"zh", "中文"},
	{"ja", "日本語"},
	{"ko", "한국어"},
}

const (
	langTitleY   = 9
	langListY    = 13
	langCols     = 4
	langRowH     = 4
	langFlagW    = 6
	langLeftCol  = -26
	langRightCol = 3
)

type LanguageSelectScreen struct {
	Game      *Game
	Selection int
}

func NewLanguageSelectScreen(g *Game) *LanguageSelectScreen {
	return &LanguageSelectScreen{Game: g}
}

func (ls *LanguageSelectScreen) Update() {}

func (ls *LanguageSelectScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	// ── 1. TERM COM title ─────────────────────────────────────────────────
	title := []string{
		"████████╗███████╗██████╗ ███╗   ███╗       ██████╗ ██████╗ ███╗   ███╗",
		"╚══██╔══╝██╔════╝██╔══██╗████╗ ████║      ██╔════╝██╔═══██╗████╗ ████║",
		"   ██║   █████╗  ██████╔╝██╔████╔██║█████╗██║     ██║   ██║██╔████╔██║",
		"   ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║╚════╝██║     ██║   ██║██║╚██╔╝██║",
		"   ██║   ███████╗██║  ██║██║ ╚═╝ ██║      ╚██████╗╚██████╔╝██║ ╚═╝ ██║",
		"   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝       ╚═════╝ ╚═════╝ ╚═╝     ╚═╝",
	}

	nowSec := float64(time.Now().UnixNano()) / 1e9
	startY := 2
	for i, line := range title {
		x := (w - len([]rune(line))) / 2
		if x < 0 {
			x = 0
		}
		col := 0
		for _, ch := range line {
			if ch == ' ' {
				col++
				continue
			}
			phase := float64(col)*0.3 + float64(i)*0.2 + nowSec*2.0
			glow := (math.Sin(phase) + 1) / 2
			r := int32(128.0 + glow*127.0)
			g := int32(40.0 + glow*60.0)
			b := int32(180.0 + glow*75.0)
			ctx.SetCell(x+col, startY+i, ch, StyleDefault.Foreground(tcell.NewRGBColor(r, g, b)).Bold(true))
			col++
		}
	}

	// ── 2. Version ───────────────────────────────────────────────────────
	verStr := "v" + GameVersion
	ctx.DrawString(w-len([]rune(verStr))-2, 0, verStr, StyleGray)

	// ── 3. Prompt ────────────────────────────────────────────────────────
	prompt := language.String("LANGUAGE_SELECT_TITLE")
	px := (w - StringWidth(prompt)) / 2
	ctx.DrawString(px, langTitleY, prompt, StyleCyanBold)

	// ── 4. Two-column language list with flags ────────────────────────────
	colX := []int{w/2 + langLeftCol, w/2 + langRightCol}

	for i, lang := range langList {
		col := 0
		row := i
		if i >= langCols {
			col = 1
			row = i - langCols
		}
		y := langListY + row*langRowH
		label := "[" + lang.Code + "] " + lang.Name
		labelLen := StringWidth(label)
		entryW := 1 + langFlagW + 1 + labelLen
		xx := colX[col]

		if i == ls.Selection {
			pad := 3
			for dx := -pad; dx < entryW+pad; dx++ {
				if xx+dx >= 0 && xx+dx < w {
					ctx.SetCell(xx+dx, y, ' ', StyleDefault)
				}
			}
			bPhase := nowSec * 3.0
			expansion := int(math.Round((math.Sin(bPhase) + 1.0) / 2.0 * 2.0))
			bSin := math.Sin(nowSec * 2.7)
			bracketStyle := StyleDefault.
				Foreground(tcell.NewRGBColor(
					int32(160.0+bSin*95.0),
					int32(220.0+bSin*35.0),
					255,
				)).Bold(true)
			tPhase := (math.Sin(nowSec*2.0) + 1.0) / 2.0
			selStyle := StyleDefault.
				Foreground(tcell.NewRGBColor(
					int32(192.0+tPhase*63.0),
					64,
					int32(255.0-tPhase*63.0),
				)).Bold(true)

			ctx.SetCell(xx-1-expansion, y, '[', bracketStyle)
			ctx.SetCell(xx+entryW+expansion, y, ']', bracketStyle)
			drawFlag(ctx, xx+1, y-1, lang.Code)
			ctx.DrawString(xx+1+langFlagW+1, y, label, selStyle)
		} else {
			dimStyle := StyleDefault.Foreground(tcell.NewRGBColor(0x88, 0x88, 0x98))
			drawFlag(ctx, xx+1, y-1, lang.Code)
			ctx.DrawString(xx+1+langFlagW+1, y, label, dimStyle)
		}
	}

	// ── 5. Help bar ──────────────────────────────────────────────────────
	ctx.DrawPanel(0, h-3, w, 3, "", StyleGray)
	ctx.DrawMarkupString(1, h-2, language.String("LANGUAGE_SELECT_HELP"), StyleGray, StyleHotkey)
}

func (ls *LanguageSelectScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		ls.Selection--
		if ls.Selection < 0 {
			ls.Selection = len(langList) - 1
		}
	case tcell.KeyDown:
		ls.Selection++
		if ls.Selection >= len(langList) {
			ls.Selection = 0
		}
	case tcell.KeyEnter:
		ls.confirm()
	case tcell.KeyEscape:
		ls.Game.SetState(StateMenu)
	}
}

func (ls *LanguageSelectScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := ls.Game.ScreenSize()

	// Help bar clicks at y = h-2
	if y == h-2 {
		help := language.String("LANGUAGE_SELECT_HELP")
		col := 1
		runes := []rune(help)
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
				switch key {
				case "↑", "↓":
					ls.Selection++
					if ls.Selection >= len(langList) {
						ls.Selection = 0
					}
				case "Enter":
					ls.confirm()
				case "Esc":
					ls.Game.SetState(StateMenu)
				}
				return
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	colX := []int{w/2 + langLeftCol, w/2 + langRightCol}

	for i, lang := range langList {
		col := 0
		row := i
		if i >= langCols {
			col = 1
			row = i - langCols
		}
		yy := langListY + row*langRowH
		label := "[" + lang.Code + "] " + lang.Name
		labelLen := StringWidth(label)
		entryW := 1 + langFlagW + 1 + labelLen
		xx := colX[col]

		if y >= yy-1 && y <= yy+1 && x >= xx+1 && x < xx+1+entryW-1 {
			ls.Selection = i
			if buttons&tcell.Button1 != 0 {
				ls.confirm()
			}
			return
		}
	}
}

func (ls *LanguageSelectScreen) confirm() {
	lang := langList[ls.Selection]
	Config.Language = lang.Code
	language.SetLanguage(lang.Code)
	SaveConfig()
	ls.Game.SetState(StateMenu)
}
