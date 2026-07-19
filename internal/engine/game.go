package engine

import (
	"os"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
)

type GameState int

const (
	controlBoxW     = 46
	controlBoxH     = 7
	controlBtnW     = 16
	controlBtnGap   = 4
	frameSleepMS    = 16
	keyChanBuffer   = 20
	startYear       = 1999
	startMonth      = time.March
	startDay        = 1
)

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
	StateTutorial
	StateLanguageSelect
	StatePlaneDesigner
	StateWeaponDesigner
	StateSeed
	StateLoadout
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

// Minimal view interfaces used by the touch control menu so engine does not
// import the geo/base/battle packages (would create an import cycle).
type geoView interface {
	UFOCount() int
	MissionCount() int
	HasSelectedBase() bool
	CanConfirm() bool
}
type battleView interface {
	HasSelectedUnit() bool
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
	screen      *ScreenRaw
	state       GameState
	stateStack  []GameState
	running     bool
	quitConfirm    bool
	confirmYesRect Rect
	confirmNoRect  Rect
	transition     float64 // 1.0 right after a state change, eases to 0 (fade-from-black)

	GameTime   time.Time
	TimeSpeed  int
	Paused     bool
	Funds      int64
	Difficulty int // 0=Beginner, 1=Experienced, 2=Veteran, 3=Genius, 4=Superhuman

	screens      map[GameState]Screen
	keyChan      chan tcell.Event
	eventDone    chan struct{}
	ActiveBattle *BattleResult
	Memorial     []*soldier.Soldier

	SpeciesSeed    int64
	AlienSpecies   []*data.AlienSpecies
	AlienTypes     []*data.AlienType
	AlienKnowledge map[string]int
	ActionDelay    int

	Tactics PlayerTactics

	FrameCount int

	controlMenuEval func() []ControlButton

	OnNewGame      func()
	OnContinue     func()
	OnLoadGame     func()
	OnCustomBattle func()

	// OnScreenChange is invoked by the main loop whenever the active game
	// state changes (e.g. menu -> geoscape -> battlescape). Frontends that use
	// a differential renderer (web, android) hook this to force a full repaint
	// so stale cells from the previous screen don't linger as artifacts.
	OnScreenChange func()

	// lastState tracks the previously rendered state so the loop can detect
	// transitions and fire OnScreenChange exactly once per switch.
	lastState GameState

	WebNotice string
}

func (g *Game) GameOver(won bool, stats string) {
	g.SetScreen(StateGameOver, NewGameOverScreen(g, won, stats))
	g.state = StateGameOver
	g.stateStack = nil
}

func NewGame() (*Game, error) {
	LoadConfig()
	scr, err := NewScreenRaw()
	if err != nil {
		return nil, err
	}
	audio.Init()

	initialState := StateMenu
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		initialState = StateLanguageSelect
	}

	return newGameWithScreen(scr, initialState), nil
}

// NewGameWithScreen creates a Game with a pre-built ScreenRaw.
// Used by the Android port where the screen is an androidScreen.
func NewGameWithScreen(scr *ScreenRaw, initialState GameState) *Game {
	audio.Init()
	return newGameWithScreen(scr, initialState)
}

func newGameWithScreen(scr *ScreenRaw, initialState GameState) *Game {
	g := &Game{
		screen:         scr,
		state:          initialState,
		running:        true,
		GameTime:       time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC),
		TimeSpeed:      0,
		Paused:         true,
		Funds:          startingFunds,
		screens:        make(map[GameState]Screen),
		keyChan:        make(chan tcell.Event, keyChanBuffer),
		eventDone:      make(chan struct{}),
		AlienKnowledge: make(map[string]int),
		ActionDelay:    Config.ActionDelay,
		lastState:      -1,
	}
	g.initSpecies()
	Menu.SetGame(g)
	return g
}

// NewGameWeb creates a Game backed by an in-memory virtual screen (no real TTY).
// cols and rows specify the initial terminal dimensions for the browser client.
// The returned *nullScreen can be used to inject events and read back the frame.
func NewGameWeb(cols, rows int) (*Game, *nullScreen, error) {
	LoadConfig()
	scr, ns, err := NewScreenRawWeb(cols, rows)
	if err != nil {
		return nil, nil, err
	}
	audio.Init()

	initialState := StateMenu
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		initialState = StateLanguageSelect
	}

	g := newGameWithScreen(scr, initialState)
	g.initSpecies()
	return g, ns, nil
}

// InjectKey posts a synthetic key event into the game loop.
func (g *Game) InjectKey(ev *tcell.EventKey) {
	select {
	case g.keyChan <- ev:
	default:
	}
}

// InjectMouse posts a synthetic mouse event into the game loop.
func (g *Game) InjectMouse(ev *tcell.EventMouse) {
	select {
	case g.keyChan <- ev:
	default:
	}
}

// InjectResize posts a synthetic resize event and updates the screen dimensions.
func (g *Game) InjectResize(cols, rows int) {
	if ns, ok := g.screen.screen.(*nullScreen); ok {
		ns.SetSize(cols, rows)
	}
	g.screen.UpdateSize()
	ev := tcell.NewEventResize(cols, rows)
	select {
	case g.keyChan <- ev:
	default:
	}
}

// ScreenRaw returns the underlying ScreenRaw so the webserver can render it.
func (g *Game) WebScreen() *ScreenRaw {
	return g.screen
}

// RandomSeed returns a fresh pseudo-random seed for a new run.
func RandomSeed() int64 {
	return time.Now().UnixNano()
}

// initSpecies generates the procedural alien roster from a random seed.
func (g *Game) initSpecies() {
	g.initSpeciesWithSeed(RandomSeed())
}

// initSpeciesWithSeed regenerates the procedural alien roster and research
// tree using the provided seed. Passing the same seed reproduces the same
// aliens, so players can share or replay specific runs.
func (g *Game) initSpeciesWithSeed(seed int64) {
	g.SpeciesSeed = seed
	g.AlienSpecies, g.AlienTypes = data.GenerateSpecies(g.SpeciesSeed)
	g.AlienKnowledge = make(map[string]int)
	data.InitResearchTree(g.SpeciesSeed, g.AlienSpecies)
	data.RegisterProceduralItems(g.SpeciesSeed, g.AlienSpecies)
}

// LearnAlien increases knowledge level for an alien type.
// Levels: 0=unknown, 1=sighted, 2=killed, 3=autopsied
func (g *Game) LearnAlien(name string, level int) {
	if g.AlienKnowledge == nil {
		g.AlienKnowledge = make(map[string]int)
	}
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
		cp := data.AlienTypes[i]
		result[i] = &cp
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
	defer SaveConfig()

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
		// Guarantee the background is always themed and refreshes
		// instantly when the theme changes (some screens rely on the
		// cleared background rather than painting it explicitly).
		w, h := g.screen.Size()
		g.screen.DrawRect(0, 0, w, h, ' ', StyleDefault)
		g.drainEvents()

		// On a screen/state transition, ask the frontend to do a full
		// repaint so differential renderers (web, android) don't leave
		// cells from the previous screen behind as artifacts.
		if g.state != g.lastState {
			g.lastState = g.state
			if g.OnScreenChange != nil {
				g.OnScreenChange()
			}
		}

		if sc, ok := g.screens[g.state]; ok {
			sc.Update()
		}
		ctx := &ScreenCtx{g.screen}
		if sc, ok := g.screens[g.state]; ok {
			sc.Render(ctx)
		}

		// Render control menu overlay (always pinned to bottom in touch mode)
		if Config.TouchMode && !HideTouchOverlay {
			w, h := g.screen.Size()
			Menu.SetScreenSize(w, h)
			if Menu.AlwaysShow && g.controlMenuEval != nil {
				Menu.SetButtons(g.controlMenuEval())
			}
			Menu.Render(g.screen)
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
		time.Sleep(frameSleepMS * time.Millisecond)
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
					g.SetScreen(StateHelp, NewHelpScreen(g, g.state))
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
					if e.Buttons() != tcell.ButtonNone {
						x, y := e.Position()
						if inRect(x, y, g.confirmYesRect) {
							g.running = false
							return
						}
						if inRect(x, y, g.confirmNoRect) {
							g.quitConfirm = false
						}
					}
					continue
				}
				// Let control menu consume the event first
				if Config.TouchMode && !HideTouchOverlay {
					w, h := g.ScreenSize()
					Menu.SetScreenSize(w, h)
					x, y := e.Position()
					if Menu.HamburgerHit(x, y) {
						Menu.Toggle()
						continue
					}
					if Menu.HandleMouse(e) {
						continue
					}
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
	g.setupControlMenu()
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
	g.setupControlMenu()
}

func (g *Game) PopState() {
	if len(g.stateStack) > 0 {
		g.state = g.stateStack[len(g.stateStack)-1]
		g.stateStack = g.stateStack[:len(g.stateStack)-1]
		g.transition = 1.0
		g.setupControlMenu()
	}
}

func (g *Game) ScreenSize() (int, int) {
	return g.screen.Size()
}

// Stop signals the game loop to exit on its next iteration.
// Used by the Android port to cleanly shut down the background goroutine.
func (g *Game) Stop() {
	g.running = false
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
	boxW := controlBoxW
	boxH := controlBoxH
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
	ctx.DrawMarkupString(x+(boxW-StringWidth(hint))/2, y+4, hint, StyleGray, StyleHotkey)

	yesLabel := language.String("CTRL_YES")
	noLabel := language.String("CTRL_NO")
	btnW := controlBtnW
	gap := controlBtnGap
	totalW := btnW*2 + gap
	by := y + boxH - 2
	bx := x + (boxW-totalW)/2
	yesRect := Rect{bx, by, btnW, 1}
	noRect := Rect{bx + btnW + gap, by, btnW, 1}
	g.confirmYesRect = yesRect
	g.confirmNoRect = noRect
	if Config.TouchMode {
		for _, r := range []Rect{yesRect, noRect} {
			for dy := 0; dy < r.H; dy++ {
				for dx := 0; dx < r.W; dx++ {
					ctx.SetCell(r.X+dx, r.Y+dy, ' ', StyleDefault)
				}
			}
			ctx.DrawBorder(r.X, r.Y, r.W, r.H, StyleDefault)
		}
		ctx.DrawString(yesRect.X+(yesRect.W-StringWidth(yesLabel))/2, yesRect.Y, yesLabel, StyleDefault)
		ctx.DrawString(noRect.X+(noRect.W-StringWidth(noLabel))/2, noRect.Y, noLabel, StyleDefault)
	}
}

func (g *Game) Bell() {
	if g.screen != nil && g.screen.screen != nil {
		g.screen.screen.Beep()
	}
}

func (g *Game) IsWeb() bool {
	_, ok := g.screen.screen.(*nullScreen)
	return ok
}

// keyBtn builds a touch-menu ControlButton that injects a key event into the
// game loop when tapped. label is the displayed text, hotkey the keyboard
// equivalent shown to the user, key the tcell key, and str an optional rune
// string (used for KeyRune events; pass "" for named keys like Enter/F5).
func (g *Game) keyBtn(label, hotkey string, key tcell.Key, str string) ControlButton {
	return ControlButton{
		Label:  label,
		Hotkey: hotkey,
		Action: func() { g.InjectKey(tcell.NewEventKey(key, str, tcell.ModNone)) },
	}
}

func (g *Game) setupControlMenu() {
	Menu.TouchFirst = false
	Menu.AlwaysShow = true
	switch g.state {
	case StateGeoscape:
		gs, _ := g.screens[StateGeoscape].(geoView)
		menu := func() []ControlButton {
			btns := []ControlButton{
				g.keyBtn(language.String("CTRL_CONFIRM"), "Enter", tcell.KeyEnter, ""),
				g.keyBtn(language.String("CTRL_PAUSE"), "Space", tcell.KeyRune, " "),
				g.keyBtn(language.String("CTRL_SPEED_1"), "1", tcell.KeyRune, "1"),
				g.keyBtn(language.String("CTRL_SPEED_2"), "2", tcell.KeyRune, "2"),
				g.keyBtn(language.String("CTRL_SPEED_3"), "3", tcell.KeyRune, "3"),
				g.keyBtn(language.String("CTRL_SPEED_4"), "4", tcell.KeyRune, "4"),
				g.keyBtn(language.String("CTRL_BASE"), "B", tcell.KeyRune, "b"),
				g.keyBtn(language.String("CTRL_LAUNCH"), "L", tcell.KeyRune, "l"),
				g.keyBtn(language.String("CTRL_AUTORESOLVE"), "A", tcell.KeyRune, "a"),
				g.keyBtn(language.String("CTRL_RESPOND"), "M", tcell.KeyRune, "m"),
				g.keyBtn(language.String("CTRL_DISPATCH"), "R", tcell.KeyRune, "r"),
				g.keyBtn(language.String("CTRL_ENCYCLOPEDIA"), "E", tcell.KeyRune, "e"),
				g.keyBtn(language.String("CTRL_SAVE"), "F5", tcell.KeyF5, ""),
				g.keyBtn(language.String("CTRL_LOAD"), "F9", tcell.KeyF9, ""),
				g.keyBtn(language.String("CTRL_QUIT"), "Q", tcell.KeyRune, "q"),
				g.keyBtn(language.String("CTRL_HELP"), "?", tcell.KeyRune, "?"),
			}
			const (
				btnConfirm      = 0  // Confirm launch at target
				btnAutoresolve  = 8  // Autoresolve needs a UFO
				btnDispatch     = 10 // Dispatch needs a mission
				btnEncyclopedia = 11 // Encyclopedia needs a base
				btnBase         = 6  // Base needs a base
			)
			for i := range btns {
				btns[i].Enabled = true
			}
			if gs != nil {
				btns[btnConfirm].Enabled = gs.CanConfirm()
				btns[btnAutoresolve].Enabled = gs.UFOCount() > 0
				btns[btnDispatch].Enabled = gs.MissionCount() > 0
				btns[btnEncyclopedia].Enabled = gs.HasSelectedBase()
				btns[btnBase].Enabled = gs.HasSelectedBase()
			}
			return btns
		}
		Menu.SetButtons(menu())
		g.controlMenuEval = menu
	case StateBattlescape:
		bs, _ := g.screens[StateBattlescape].(battleView)
		menu := func() []ControlButton {
			btns := []ControlButton{
				g.keyBtn("↑", "Up", tcell.KeyUp, ""),
				g.keyBtn("↓", "Down", tcell.KeyDown, ""),
				g.keyBtn("←", "Left", tcell.KeyLeft, ""),
				g.keyBtn("→", "Right", tcell.KeyRight, ""),
				g.keyBtn(language.String("CTRL_SELECT"), "Enter", tcell.KeyEnter, ""),
				g.keyBtn(language.String("CTRL_DESELECT"), "Space", tcell.KeyRune, " "),
				g.keyBtn(language.String("CTRL_MOVE"), "M", tcell.KeyRune, "m"),
				g.keyBtn(language.String("CTRL_FIRE"), "F", tcell.KeyRune, "f"),
				g.keyBtn(language.String("CTRL_RELOAD"), "R", tcell.KeyRune, "r"),
				g.keyBtn(language.String("CTRL_END_TURN"), "E", tcell.KeyRune, "e"),
				g.keyBtn(language.String("CTRL_GRENADE"), "G", tcell.KeyRune, "g"),
				g.keyBtn(language.String("CTRL_MEDIKIT"), "H", tcell.KeyRune, "h"),
				g.keyBtn(language.String("CTRL_CROUCH"), "C", tcell.KeyRune, "c"),
				g.keyBtn(language.String("CTRL_PSI"), "P", tcell.KeyRune, "p"),
				g.keyBtn(language.String("CTRL_CYCLE"), "Q", tcell.KeyRune, "q"),
				g.keyBtn(language.String("CTRL_SCANNER"), "Y", tcell.KeyRune, "y"),
				g.keyBtn(language.String("CTRL_MINES"), "T", tcell.KeyRune, "t"),
				g.keyBtn(language.String("CTRL_PAN_UP"), "W", tcell.KeyRune, "w"),
				g.keyBtn(language.String("CTRL_PAN_DOWN"), "S", tcell.KeyRune, "s"),
				g.keyBtn(language.String("CTRL_PAN_LEFT"), "A", tcell.KeyRune, "a"),
				g.keyBtn(language.String("CTRL_PAN_RIGHT"), "D", tcell.KeyRune, "d"),
				g.keyBtn(language.String("CTRL_VISION"), "V", tcell.KeyRune, "v"),
				g.keyBtn(language.String("CTRL_OPTIONS"), "O", tcell.KeyRune, "o"),
				g.keyBtn(language.String("CTRL_HELP"), "?", tcell.KeyRune, "?"),
			}
			const (
				btnFire     = 7
				btnReload   = 8
				btnGrenade  = 10
				btnMedikit  = 11
				btnPsi      = 13
				btnScanner  = 15
				btnMines    = 16
				btnMove     = 6
			)
			for i := range btns {
				btns[i].Enabled = true
			}
			if bs != nil {
				hasSel := bs.HasSelectedUnit()
				btns[btnFire].Enabled = hasSel
				btns[btnReload].Enabled = hasSel
				btns[btnGrenade].Enabled = hasSel
				btns[btnMedikit].Enabled = hasSel
				btns[btnPsi].Enabled = hasSel
				btns[btnScanner].Enabled = hasSel
				btns[btnMines].Enabled = hasSel
				btns[btnMove].Enabled = true
			}
			return btns
		}
		Menu.SetButtons(menu())
		g.controlMenuEval = menu
	case StateBase:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_FACILITIES"), Hotkey: "1", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "1", tcell.ModNone)) }},
			{Label: language.String("CTRL_SOLDIERS"), Hotkey: "2", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "2", tcell.ModNone)) }},
			{Label: language.String("CTRL_RESEARCH"), Hotkey: "3", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "3", tcell.ModNone)) }},
			{Label: language.String("CTRL_MANUFACTURE"), Hotkey: "4", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "4", tcell.ModNone)) }},
			{Label: language.String("CTRL_TRANSFER"), Hotkey: "5", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "5", tcell.ModNone)) }},
			{Label: language.String("CTRL_HANGARS"), Hotkey: "6", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "6", tcell.ModNone)) }},
			{Label: language.String("CTRL_BACK"), Hotkey: "Esc", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	case StateEquip:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_PAN_UP"), Hotkey: "↑", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyUp, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_PAN_DOWN"), Hotkey: "↓", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_SLOT_WEAPON"), Hotkey: "1", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "1", tcell.ModNone)) }},
			{Label: language.String("CTRL_SLOT_ARMOR"), Hotkey: "2", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "2", tcell.ModNone)) }},
			{Label: language.String("CTRL_CYCLE_ITEM"), Hotkey: "Tab", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_EQUIP"), Hotkey: "Space", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone)) }},
			{Label: language.String("CTRL_AUTO_EQUIP"), Hotkey: "A", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone)) }},
			{Label: language.String("CTRL_BACK"), Hotkey: "Esc", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	case StateResearch, StateManufacture:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_PAN_UP"), Hotkey: "↑", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyUp, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_PAN_DOWN"), Hotkey: "↓", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyDown, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_BACK"), Hotkey: "Esc", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	case StateDebrief:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_DISMISS"), Hotkey: "Enter", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	case StateMenu:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	case StateGameOver:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_QUIT"), Hotkey: "Q", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "q", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	default:
		Menu.SetButtons([]ControlButton{
			{Label: language.String("CTRL_BACK"), Hotkey: "Esc", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)) }},
			{Label: language.String("CTRL_HELP"), Hotkey: "?", Action: func() { g.InjectKey(tcell.NewEventKey(tcell.KeyRune, "?", tcell.ModNone)) }},
		})
		g.controlMenuEval = nil
	}
}

func inRect(x, y int, r Rect) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}
