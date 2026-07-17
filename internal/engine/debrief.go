package engine

import (
	"fmt"
	"strings"

	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

const FundsDisplayK = 1000

// DebriefSoldier holds per-soldier info for the after-action report.
type DebriefSoldier struct {
	Name      string
	Rank      string
	Died      bool
	StatGains string
}

// DebriefData carries all information needed to render the post-mission screen.
type DebriefData struct {
	Won          bool
	MissionName  string
	BaseName     string
	Kills        int
	Casualties   []string
	LootItems    []string
	StunnedCount int
	FundsEarned  int64
	Soldiers     []DebriefSoldier

	// Cydonia special case
	CydoniaVictory bool
	BaseDestroyed  bool
}

// DebriefScreen shows a full-screen after-action report after a tactical battle.
type DebriefScreen struct {
	game *Game
	data *DebriefData
}

func NewDebriefScreen(g *Game, d *DebriefData) *DebriefScreen {
	return &DebriefScreen{game: g, data: d}
}

func (ds *DebriefScreen) Update() {}
func (ds *DebriefScreen) HandleMouse(e *tcell.EventMouse) {
	if e.Buttons()&tcell.Button1 != 0 {
		ds.dismiss()
	}
}

func (ds *DebriefScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyEnter, tcell.KeyEscape, tcell.KeyCtrlC:
		ds.dismiss()
	default:
		if e.Str() == " " {
			ds.dismiss()
		}
	}
}

func (ds *DebriefScreen) dismiss() {
	if ds.data.CydoniaVictory {
		ds.game.GameOver(true, language.String("DEBRIEF_CYDONIA_VICTORY"))
		return
	}
	ds.game.PopState()
}

func (ds *DebriefScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	d := ds.data

	// Dark overlay background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ctx.SetCell(x, y, ' ', StyleDefault.Background(tcell.NewRGBColor(5, 5, 20)))
		}
	}

	// Title
	title := language.String("DEBRIEF_TITLE")
	if d.Won {
		ctx.DrawPanel(2, 1, w-4, h-3, title, StyleGreen.Bold(true))
	} else {
		if d.BaseDestroyed {
			title = language.String("DEBRIEF_BASE_LOST")
		}
		ctx.DrawPanel(2, 1, w-4, h-3, title, StyleRed.Bold(true))
	}

	yOff := 3

	// Mission info
	ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_MISSION"), d.MissionName), StyleCyan)
	yOff++
	ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_BASE"), d.BaseName), StyleDefault)
	yOff += 2

	// Outcome
	outcome := language.String("DEBRIEF_VICTORY")
	outStyle := StyleGreen.Bold(true)
	if !d.Won {
		outcome = language.String("DEBRIEF_DEFEAT")
		outStyle = StyleRed.Bold(true)
	}
	ctx.DrawString(4, yOff, language.String("DEBRIEF_OUTCOME"), StyleDefault)
	ctx.DrawString(4+len(language.String("DEBRIEF_OUTCOME")), yOff, outcome, outStyle)
	yOff += 2

	// Kills & casualties
	ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_ALIENS_KILLED"), d.Kills), StyleDefault)
	yOff++
	if len(d.Casualties) > 0 {
		ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_SOLDIERS_LOST"), strings.Join(d.Casualties, ", ")), StyleRed)
	} else {
		ctx.DrawString(4, yOff, language.String("DEBRIEF_SOLDIERS_LOST_NONE"), StyleGreen)
	}
	yOff++
	ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_LOOT_RECOVERED"), len(d.LootItems)), StyleYellow)
	if d.StunnedCount > 0 {
		yOff++
		ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_ALIENS_CAPTURED"), d.StunnedCount), StyleMagenta)
	}
	if d.FundsEarned > 0 {
		yOff++
		ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_FUNDS_EARNED"), d.FundsEarned/FundsDisplayK), StyleGreen)
	}
	yOff += 2

	// Per-soldier report
	if len(d.Soldiers) > 0 {
		ctx.DrawString(4, yOff, language.String("DEBRIEF_SQUAD"), StyleCyanBold)
		yOff++
		header := fmt.Sprintf(language.String("DEBRIEF_HEADER_FORMAT"), language.String("DEBRIEF_HEADER_NAME"), language.String("DEBRIEF_HEADER_RANK"), language.String("DEBRIEF_HEADER_CHANGES"))
		ctx.DrawString(4, yOff, header, StyleGray)
		yOff++
		for _, s := range d.Soldiers {
			nameStyle := StyleDefault
			if s.Died {
				nameStyle = StyleRed
			}
			gains := s.StatGains
			if gains == "" {
				if s.Died {
					gains = language.String("DEBRIEF_KIA")
				} else {
					gains = language.String("DEBRIEF_NO_CHANGE")
				}
			}
			ctx.DrawString(4, yOff, fmt.Sprintf(language.String("DEBRIEF_SOLDIER_FORMAT"), s.Name, s.Rank, gains), nameStyle)
			yOff++
			if yOff >= h-4 {
				break
			}
		}
	}

	// Dismiss prompt
	prompt := language.String("DEBRIEF_PROMPT")
	if d.CydoniaVictory {
		prompt = language.String("DEBRIEF_VICTORY_PROMPT")
	}
	ctx.DrawString(w/2-len(prompt)/2, h-2, prompt, StyleGray)
}
