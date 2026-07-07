package engine

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type GameState int

const (
	StateGeoscape GameState = iota
	StateBase
	StateBattlescape
	StateResearch
	StateManufacture
	StateEquip
	StateQuit
)

type Screen interface {
	Update()
	Render(*ScreenCtx)
	HandleKey(*tcell.EventKey)
}

type ScreenCtx struct {
	*ScreenRaw
}

type Game struct {
	screen   *ScreenRaw
	state    GameState
	stateStack []GameState
	running  bool

	GameTime     time.Time
	TimeSpeed    int
	Paused       bool
	Funds        int64

	screens map[GameState]Screen
}

func NewGame() (*Game, error) {
	scr, err := NewScreenRaw()
	if err != nil {
		return nil, err
	}

	g := &Game{
		screen:    scr,
		state:     StateGeoscape,
		running:   true,
		GameTime:  time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC),
		TimeSpeed: 0,
		Paused:    true,
		Funds:     500000,
		screens:   make(map[GameState]Screen),
	}
	return g, nil
}

func (g *Game) RegisterScreen(s GameState, sc Screen) {
	g.screens[s] = sc
}

func (g *Game) Run() {
	defer g.screen.Close()
	for g.running {
		g.screen.Clear()
		g.HandleInput()
		if sc, ok := g.screens[g.state]; ok {
			sc.Update()
		}
		ctx := &ScreenCtx{g.screen}
		if sc, ok := g.screens[g.state]; ok {
			sc.Render(ctx)
		}
		g.screen.Flush()
		time.Sleep(16 * time.Millisecond)
	}
}

func (g *Game) PushState(s GameState) {
	g.stateStack = append(g.stateStack, g.state)
	g.state = s
}

func (g *Game) PopState() {
	if len(g.stateStack) > 0 {
		g.state = g.stateStack[len(g.stateStack)-1]
		g.stateStack = g.stateStack[:len(g.stateStack)-1]
	}
}

func (g *Game) HandleInput() {
	for {
		ev := g.screen.screen.PollEvent()
		if ev == nil {
			return
		}
		switch e := ev.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyEscape && e.Rune() == 0 {
				switch g.state {
				case StateGeoscape:
					g.running = false
				default:
					g.PopState()
				}
				return
			}
			if sc, ok := g.screens[g.state]; ok {
				sc.HandleKey(e)
			}
		}
	}
}

func (g *Game) ScreenSize() (int, int) {
	return g.screen.Size()
}

func (g *Game) Quit() {
	g.running = false
}
