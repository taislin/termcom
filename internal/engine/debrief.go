package engine

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
)

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
func (ds *DebriefScreen) HandleMouse(_ *tcell.EventMouse) {}

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
		ds.game.GameOver(true, "Cydonia assault successful!")
		return
	}
	if ds.data.BaseDestroyed {
		// Already handled by geoscape — just pop back
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
	title := " AFTER ACTION REPORT "
	if d.Won {
		ctx.DrawPanel(2, 1, w-4, h-3, title, StyleGreen.Bold(true))
	} else {
		if d.BaseDestroyed {
			title = " BASE LOST "
		}
		ctx.DrawPanel(2, 1, w-4, h-3, title, StyleRed.Bold(true))
	}

	yOff := 3

	// Mission info
	ctx.DrawString(4, yOff, fmt.Sprintf("Mission: %s", d.MissionName), StyleCyan)
	yOff++
	ctx.DrawString(4, yOff, fmt.Sprintf("Base:    %s", d.BaseName), StyleDefault)
	yOff += 2

	// Outcome
	outcome := "VICTORY"
	outStyle := StyleGreen.Bold(true)
	if !d.Won {
		outcome = "DEFEAT"
		outStyle = StyleRed.Bold(true)
	}
	ctx.DrawString(4, yOff, "Outcome: ", StyleDefault)
	ctx.DrawString(13, yOff, outcome, outStyle)
	yOff += 2

	// Kills & casualties
	ctx.DrawString(4, yOff, fmt.Sprintf("Aliens eliminated: %d", d.Kills), StyleDefault)
	yOff++
	if len(d.Casualties) > 0 {
		ctx.DrawString(4, yOff, fmt.Sprintf("Soldiers lost: %s", strings.Join(d.Casualties, ", ")), StyleRed)
	} else {
		ctx.DrawString(4, yOff, "Soldiers lost: none", StyleGreen)
	}
	yOff++
	ctx.DrawString(4, yOff, fmt.Sprintf("Loot recovered: %d items", len(d.LootItems)), StyleYellow)
	if d.StunnedCount > 0 {
		yOff++
		ctx.DrawString(4, yOff, fmt.Sprintf("Aliens captured: %d", d.StunnedCount), StyleMagenta)
	}
	if d.FundsEarned > 0 {
		yOff++
		ctx.DrawString(4, yOff, fmt.Sprintf("Funds earned: $%dK", d.FundsEarned/1000), StyleGreen)
	}
	yOff += 2

	// Per-soldier report
	if len(d.Soldiers) > 0 {
		ctx.DrawString(4, yOff, "Squad:", StyleCyanBold)
		yOff++
		header := fmt.Sprintf("  %-16s %-10s %s", "Name", "Rank", "Changes")
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
					gains = "KIA"
				} else {
					gains = "no change"
				}
			}
			ctx.DrawString(4, yOff, fmt.Sprintf("  %-16s %-10s %s", s.Name, s.Rank, gains), nameStyle)
			yOff++
			if yOff >= h-4 {
				break
			}
		}
	}

	// Dismiss prompt
	prompt := "Press Enter, Space, or Esc to continue"
	if d.CydoniaVictory {
		prompt = "Press Enter, Space, or Esc to see victory screen"
	}
	ctx.DrawString(w/2-len(prompt)/2, h-2, prompt, StyleGray)
}
