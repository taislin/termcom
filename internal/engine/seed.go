package engine

import (
	"strconv"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
)

// SeedScreen is shown right before a new game starts. It displays the map
// seed that will be used to generate the procedural alien roster and lets the
// player reroll a random seed or type a specific one before confirming.
type SeedScreen struct {
	Game      *Game
	Seed      int64
	Edit      []rune // editable seed text while the user types a custom value
	Editing   bool
	OnConfirm func(seed int64)
}

// NewSeedScreen creates the seed dialog. The initial seed is taken from the
// game's current SpeciesSeed (already randomized at construction); the player
// may reroll or override it.
func NewSeedScreen(g *Game, onConfirm func(seed int64)) *SeedScreen {
	return &SeedScreen{
		Game:      g,
		Seed:      g.SpeciesSeed,
		Edit:      []rune(strconv.FormatInt(g.SpeciesSeed, 10)),
		OnConfirm: onConfirm,
	}
}

func (ss *SeedScreen) Update() {}

func (ss *SeedScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("SEED_TITLE"), StyleDefault)

	title := language.String("SEED_PROMPT")
	ctx.DrawString(2, 2, title, StyleCyanBold)

	// Current seed
	seedLabel := language.String("SEED_VALUE")
	seedStr := strconv.FormatInt(ss.Seed, 10)
	ctx.DrawString(2, 5, seedLabel, StyleGray)
	ctx.DrawString(2+StringWidth(seedLabel)+1, 5, seedStr, StyleHighlight)

	// Editing field
	editLabel := language.String("SEED_EDIT_LABEL")
	editStr := string(ss.Edit)
	estyle := StyleDefault
	if ss.Editing {
		editStr += "_"
		estyle = StyleHotkey
	}
	ctx.DrawString(2, 7, editLabel, StyleGray)
	ctx.DrawString(2+StringWidth(editLabel)+1, 7, editStr, estyle)

	// Controls
	help := language.String("SEED_HELP")
	ctx.DrawMarkupString(2, h-1, help, StyleGray, StyleHotkey)
}

func (ss *SeedScreen) confirm() {
	// Apply the chosen seed and regenerate the procedural content.
	ss.Game.initSpeciesWithSeed(ss.Seed)
	if ss.OnConfirm != nil {
		ss.OnConfirm(ss.Seed)
	}
}

func (ss *SeedScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyEnter:
		ss.confirm()
	case tcell.KeyEscape:
		ss.Game.PopState()
	case tcell.KeyCtrlR:
		ss.reroll()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if ss.Editing && len(ss.Edit) > 0 {
			ss.Edit = ss.Edit[:len(ss.Edit)-1]
			ss.applyEdit()
		}
	case tcell.KeyRune:
		switch e.Str() {
		case "r", "R":
			ss.reroll()
		case "e", "E":
			ss.Editing = !ss.Editing
		case "/":
			ss.Editing = true
		default:
			if ss.Editing {
				r := []rune(e.Str())
				if len(r) == 1 && r[0] >= '0' && r[0] <= '9' {
					ss.Edit = append(ss.Edit, r[0])
					ss.applyEdit()
				}
			}
		}
	}
}

func (ss *SeedScreen) applyEdit() {
	if len(ss.Edit) == 0 {
		ss.Seed = 0
		return
	}
	if v, err := strconv.ParseInt(string(ss.Edit), 10, 64); err == nil {
		ss.Seed = v
	}
}

func (ss *SeedScreen) reroll() {
	ss.Seed = RandomSeed()
	ss.Edit = []rune(strconv.FormatInt(ss.Seed, 10))
	ss.Editing = false
}

func (ss *SeedScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	if buttons&tcell.Button1 != 0 {
		// Left click confirms (mirrors Enter).
		ss.confirm()
	}
}
