package engine

import (
	"time"

	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
	"github.com/gdamore/tcell/v3"
)

type GameState int

const (
	StateMenu GameState = iota
	StateGeoscape
	StateBase
	StateBattlescape
	StateResearch
	StateManufacture
	StateEquip
	StateHelp
	StateEncyclopedia
	StateOptions
	StateSlotPicker
	StateDifficulty
	StateGameOver
	StateDebrief
	StateQuit
)

type Screen interface {
	Update()
	Render(*ScreenCtx)
	HandleKey(*tcell.EventKey)
	HandleMouse(*tcell.EventMouse)
}

type ScreenCtx struct {
	*ScreenRaw
}

type BattleResult struct {
	Won           bool
	Kills         int
	Soldiers      []*soldier.Soldier
	LootItems     []string
	StunnedAliens []string // Added
}

type PlayerTactics struct {
	BattleCount        int
	TotalAlienKills    int
	TotalSoldierLosses int
	AverageRange       float64
	GrenadeUsage       int
	FlankingObserved   int
}

type Game struct {
	screen     *ScreenRaw
	state      GameState
	stateStack []GameState
	running    bool
	quitConfirm bool
	transition  float64 // 1.0 right after a state change, eases to 0 (fade-from-black)

	GameTime  time.Time
	TimeSpeed int
	Paused    bool
	Funds     int64
	Difficulty int // 0=Beginner, 1=Experienced, 2=Veteran, 3=Genius, 4=Superhuman

	screens      map[GameState]Screen
	keyChan      chan tcell.Event
	eventDone    chan struct{}
	ActiveBattle *BattleResult

	SpeciesSeed    int64
	AlienSpecies   []*data.AlienSpecies
	AlienTypes     []*data.AlienType
	AlienKnowledge map[string]int
	ActionDelay    int

	Tactics PlayerTactics

	FrameCount int

	OnNewGame      func()
	OnContinue     func()
	OnLoadGame     func()
	OnCustomBattle func()
}

func (g *Game) GameOver(won bool, stats string) {
	g.SetScreen(StateGameOver, NewGameOverScreen(g, won, stats))
	g.PushState(StateGameOver)
}

func NewGame() (*Game, error) {
	LoadConfig()
	scr, err := NewScreenRaw()
	if err != nil {
		return nil, err
	}
	audio.Init()

	g := &Game{
		screen:         scr,
		state:          StateMenu,
		running:        true,
		GameTime:       time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC),
		TimeSpeed:      0,
		Paused:         true,
		Funds:          500000,
		screens:        make(map[GameState]Screen),
		keyChan:        make(chan tcell.Event, 20),
		eventDone:      make(chan struct{}),
		AlienKnowledge: make(map[string]int),
		ActionDelay:    Config.ActionDelay,
	}
	g.initSpecies()
	return g, nil
}

func (g *Game) initSpecies() {
	g.SpeciesSeed = time.Now().UnixNano()
	g.AlienSpecies, g.AlienTypes = data.GenerateSpecies(g.SpeciesSeed)
	g.AlienKnowledge = make(map[string]int)
	data.InitResearchTree(g.SpeciesSeed, g.AlienSpecies)
	data.RegisterProceduralItems(g.SpeciesSeed, g.AlienSpecies)
}

// LearnAlien increases knowledge level for an alien type.
// Levels: 0=unknown, 1=sighted, 2=killed, 3=autopsied
func (g *Game) LearnAlien(name string, level int) {
	if g.AlienKnowledge[name] < level {
		g.AlienKnowledge[name] = level
	}
}

// GetAlienTypes returns the procedural alien types for the current run.
func (g *Game) GetAlienTypes() []*data.AlienType {
	if len(g.AlienTypes) > 0 {
		return g.AlienTypes
	}
	result := make([]*data.AlienType, len(data.AlienTypes))
	for i := range data.AlienTypes {
		result[i] = &data.AlienTypes[i]
	}
	return result
}

// GetHardcodedAliens returns the hardcoded alien roster (reserved for scripted missions).
func (g *Game) GetHardcodedAliens() []*data.AlienType {
	result := make([]*data.AlienType, len(data.AlienTypes))
	for i := range data.AlienTypes {
		result[i] = &data.AlienTypes[i]
	}
	return result
}

func (g *Game) RegisterScreen(s GameState, sc Screen) {
	if g.screens == nil {
		g.screens = make(map[GameState]Screen)
	}
	g.screens[s] = sc
}

func (g *Game) OpenEncyclopedia(completed []string, weapons []string, armor []string) {
	enc := NewEncyclopediaScreen(g, completed, weapons, armor)
	if g.screens == nil {
		g.screens = make(map[GameState]Screen)
	}
	g.screens[StateEncyclopedia] = enc
	g.PushState(StateEncyclopedia)
}

func (g *Game) SetScreen(s GameState, sc Screen) {
	if g.screens == nil {
		g.screens = make(map[GameState]Screen)
	}
	g.screens[s] = sc
}

func (g *Game) Run() {
	defer g.screen.Close()
	defer audio.Close()
	defer close(g.eventDone)

	go func() {
		for {
			select {
			case ev := <-g.screen.screen.EventQ():
				select {
				case g.keyChan <- ev:
				case <-g.eventDone:
					return
				}
			case <-g.eventDone:
				return
			}
		}
	}()

	for g.running {
		g.screen.Clear()
		g.drainEvents()

		if sc, ok := g.screens[g.state]; ok {
			sc.Update()
		}
		ctx := &ScreenCtx{g.screen}
		if sc, ok := g.screens[g.state]; ok {
			sc.Render(ctx)
		}

		if g.quitConfirm {
			g.renderQuitConfirm(ctx)
		} else if g.transition > 0 {
			w, h := ctx.Size()
			DrawTransparentRect(ctx.ScreenRaw, ctx.FrameBuffer(), 0, 0, w, h, ColorBlack, g.transition)
			g.transition *= 0.85
			if g.transition < 0.03 {
				g.transition = 0
			}
		}

		g.screen.Flush()
		g.FrameCount++
		time.Sleep(16 * time.Millisecond)
	}
}

func (g *Game) drainEvents() {
	for {
		select {
		case ev := <-g.keyChan:
			switch e := ev.(type) {
			case *tcell.EventResize:
				g.screen.UpdateSize()
			case *tcell.EventKey:
			if g.quitConfirm {
				switch {
				case e.Str() == "y" || e.Str() == "Y" || e.Key() == tcell.KeyEnter:
					g.running = false
					return
				case e.Str() == "n" || e.Str() == "N" || e.Key() == tcell.KeyEscape || e.Str() == "\x1b":
					g.quitConfirm = false
				}
				continue
			}
				if e.Key() == tcell.KeyEscape || e.Str() == "\x1b" {
					switch g.state {
					case StateGeoscape, StateMenu:
						g.Quit()
					case StateBattlescape, StateDebrief:
						if sc, ok := g.screens[g.state]; ok {
							sc.HandleKey(e)
						}
					default:
						g.PopState()
					}
				} else if e.Str() == "?" {
					g.PushState(StateHelp)
				} else if e.Str() == "o" || e.Str() == "O" {
					if _, ok := g.screens[StateOptions]; !ok {
						g.SetScreen(StateOptions, NewOptionsScreen(g))
					}
					g.PushState(StateOptions)
				} else if sc, ok := g.screens[g.state]; ok {
					sc.HandleKey(e)
				}
			case *tcell.EventMouse:
				if g.quitConfirm {
					continue
				}
				if sc, ok := g.screens[g.state]; ok {
					sc.HandleMouse(e)
				}
			}
		default:
			return
		}
	}
}

func (g *Game) PushState(s GameState) {
	g.stateStack = append(g.stateStack, g.state)
	g.state = s
	g.transition = 1.0
}

func (g *Game) InState(s GameState) bool {
	return g.state == s
}

func (g *Game) PushScreen(sc Screen) {
	g.screens[StateSlotPicker] = sc
	g.PushState(StateSlotPicker)
}

func (g *Game) SetState(s GameState) {
	g.state = s
	g.transition = 1.0
}

func (g *Game) PopState() {
	if len(g.stateStack) > 0 {
		g.state = g.stateStack[len(g.stateStack)-1]
		g.stateStack = g.stateStack[:len(g.stateStack)-1]
		g.transition = 1.0
	}
}

func (g *Game) ScreenSize() (int, int) {
	return g.screen.Size()
}

func (g *Game) Quit() {
	if !Config.ConfirmDialogs {
		g.running = false
		return
	}
	g.quitConfirm = true
}

func (g *Game) renderQuitConfirm(ctx *ScreenCtx) {
	w, h := ctx.Size()
	boxW := 46
	boxH := 7
	x := (w - boxW) / 2
	y := (h - boxH) / 2
	// Fill the box with an opaque background so the screen underneath doesn't show through.
	for fy := y; fy < y+boxH; fy++ {
		for fx := x; fx < x+boxW; fx++ {
			ctx.SetCell(fx, fy, ' ', StyleGray)
		}
	}
	ctx.DrawPanel(x, y, boxW, boxH, "", StyleGray)
	msg := language.String("CONFIRM_QUIT")
	ctx.DrawString(x+(boxW-StringWidth(msg))/2, y+2, msg, StyleDefault)
	hint := language.String("CONFIRM_QUIT_HINT")
	ctx.DrawString(x+(boxW-StringWidth(hint))/2, y+4, hint, StyleGray)
}

func (g *Game) Bell() {
	if g.screen != nil && g.screen.screen != nil {
		g.screen.screen.Beep()
	}
}
