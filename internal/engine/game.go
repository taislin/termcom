package engine

import (
	"time"

	"github.com/civ13/ycom/internal/audio"
	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/soldier"
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
	StateGameOver
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

	SpeciesSeed  int64
	AlienSpecies []*data.AlienSpecies
	AlienTypes   []*data.AlienType
	AlienKnowledge map[string]int // alien name -> knowledge level (0=unknown, 1=sighted, 2=killed, 3=autopsied)
	
	FrameCount int

	OnNewGame  func()
	OnContinue func()
	OnLoadGame func()
}

func (g *Game) GameOver(won bool, stats string) {
	g.SetScreen(StateGameOver, NewGameOverScreen(g, won, stats))
	g.PushState(StateGameOver)
}

func NewGame() (*Game, error) {
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
	}
	g.initSpecies()
	return g, nil
}

func (g *Game) initSpecies() {
	g.SpeciesSeed = time.Now().UnixNano()
	g.AlienSpecies, g.AlienTypes = data.GenerateSpecies(g.SpeciesSeed)
	g.AlienKnowledge = make(map[string]int)
	data.InitResearchTree(g.SpeciesSeed, g.AlienSpecies)
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

func (g *Game) RegisterScreen(s GameState, sc Screen) {
	g.screens[s] = sc
}

func (g *Game) OpenEncyclopedia(completed []string, weapons []string, armor []string) {
	enc := NewEncyclopediaScreen(g, completed, weapons, armor)
	g.screens[StateEncyclopedia] = enc
	g.PushState(StateEncyclopedia)
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

		if Config.DistortionEnabled {
			ApplyDistortion(g.screen, g.screen.FrameBuffer(), float64(g.FrameCount))
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
			if e.Key() == tcell.KeyEscape || e.Str() == "\x1b" {
				switch g.state {
				case StateGeoscape, StateMenu:
					g.running = false
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

func (g *Game) PushScreen(sc Screen) {
	g.screens[StateSlotPicker] = sc
	g.PushState(StateSlotPicker)
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
