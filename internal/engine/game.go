package engine

import (
	"time"

	"github.com/civ13/ycom/internal/audio"
	"github.com/civ13/ycom/internal/soldier"
	"github.com/gdamore/tcell/v2"
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
	Won       bool
	Kills     int
	Soldiers  []*soldier.Soldier
	LootItems []string
}

type Game struct {
	screen     *ScreenRaw
	state      GameState
	stateStack []GameState
	running    bool

	GameTime  time.Time
	TimeSpeed int
	Paused    bool
	Funds     int64

	screens      map[GameState]Screen
	keyChan      chan tcell.Event
	eventDone    chan struct{}
	ActiveBattle *BattleResult

	OnNewGame  func()
	OnContinue func()
}

func NewGame() (*Game, error) {
	scr, err := NewScreenRaw()
	if err != nil {
		return nil, err
	}
	audio.Init()

	g := &Game{
		screen:    scr,
		state:     StateMenu,
		running:   true,
		GameTime:  time.Date(1999, time.March, 1, 0, 0, 0, 0, time.UTC),
		TimeSpeed: 0,
		Paused:    true,
		Funds:     500000,
		screens:   make(map[GameState]Screen),
		keyChan:   make(chan tcell.Event, 20),
		eventDone: make(chan struct{}),
	}
	return g, nil
}

func (g *Game) RegisterScreen(s GameState, sc Screen) {
	g.screens[s] = sc
}

func (g *Game) SetScreen(s GameState, sc Screen) {
	g.screens[s] = sc
}

func (g *Game) Run() {
	defer g.screen.Close()
	defer audio.Close()
	defer close(g.eventDone)

	go func() {
		for {
			ev := g.screen.screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case g.keyChan <- ev:
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

		g.screen.Flush()
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
			if e.Key() == tcell.KeyEscape || (e.Key() == tcell.KeyRune && e.Rune() == 27) {
				switch g.state {
				case StateGeoscape, StateMenu:
					g.running = false
				default:
						g.PopState()
					}
				} else if e.Rune() == '?' {
					g.PushState(StateHelp)
				} else if sc, ok := g.screens[g.state]; ok {
					sc.HandleKey(e)
				}
			case *tcell.EventMouse:
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
}

func (g *Game) SetState(s GameState) {
	g.state = s
}

func (g *Game) PopState() {
	if len(g.stateStack) > 0 {
		g.state = g.stateStack[len(g.stateStack)-1]
		g.stateStack = g.stateStack[:len(g.stateStack)-1]
	}
}

func (g *Game) ScreenSize() (int, int) {
	return g.screen.Size()
}

func (g *Game) Quit() {
	g.running = false
}

func (g *Game) Bell() {
	g.screen.screen.Beep()
}
