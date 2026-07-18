package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

const SaveFile = "xcom_save.json"

const (
	titleLines    = 6
	titleGap      = 1
	subtitleOffset = 4
	menuStartY    = 2
)

var starRunes = [3]rune{'.', '+', '*'}

const (
	starVerticalSpread = 0.55
	starMinDist        = 0.15
	starColorScale     = 180.0
	starBlueBase       = 80.0
	starBlueScale      = 175.0
)

type menuStar struct {
	angle   float64
	dist    float64
	speed   float64
	baseBri float64
	size    int // 0='.', 1='+', 2='*'
}

type MenuScreen struct {
	Game          *Game
	Selection     int
	lastSelection int

	// Starfield
	stars        []menuStar
	starsSeeded  bool
	starW, starH int

	// Bracket animation
	bracketPhase float64

	// Drift particles
	menuParticles *ParticleSystem
	driftTick     int

	// Timing
	lastUpdate time.Time

	// Update check
	latestVersion     atomic.Value // stores string (empty = not fetched)
	updateCheckStarted bool
}

func NewMenuScreen(g *Game) *MenuScreen {
	ms := &MenuScreen{
		Game:          g,
		Selection:     0,
		lastSelection: -1,
		menuParticles: NewParticleSystem(80),
		lastUpdate:    time.Now(),
	}
	ms.latestVersion.Store("")
	ms.startVersionCheck()
	return ms
}

func gameVersionStr() string {
	return "v" + GameVersion
}

func HasSave() bool {
	if _, err := os.Stat(SaveFile); err == nil {
		return true
	}
	for slot := 1; slot <= 10; slot++ {
		if _, err := os.Stat(fmt.Sprintf("save_slot_%d.json", slot)); err == nil {
			return true
		}
	}
	if _, err := os.Stat("autosave.json"); err == nil {
		return true
	}
	return false
}

func (ms *MenuScreen) seedStars(w, h int) {
	const numStars = 150
	ms.stars = make([]menuStar, numStars)
	for i := range ms.stars {
		ms.stars[i] = menuStar{
			angle:   rand.Float64() * 2 * math.Pi,
			dist:    rand.Float64(),
			speed:   0.04 + rand.Float64()*0.12,
			baseBri: 0.4 + rand.Float64()*0.6,
			size:    rand.Intn(3),
		}
	}
	ms.starW = w
	ms.starH = h
	ms.starsSeeded = true
}

func (ms *MenuScreen) Update() {
	now := time.Now()
	dt := now.Sub(ms.lastUpdate).Seconds()
	if dt > 0.1 {
		dt = 0.1
	}
	ms.lastUpdate = now

	// Reset bracket phase and clear particles when selection changes
	if ms.Selection != ms.lastSelection {
		ms.bracketPhase = 0
		ms.menuParticles.Clear()
		ms.lastSelection = ms.Selection
	}

	ms.bracketPhase += dt

	for i := range ms.stars {
		ms.stars[i].dist += ms.stars[i].speed * dt
		if ms.stars[i].dist > 1.0 {
			ms.stars[i].dist = 0.0
			ms.stars[i].angle = rand.Float64() * 2 * math.Pi
		}
	}

	ms.menuParticles.Update(dt)

	// Spawn drift particles from both edges of selected option every 8 ticks (~130 ms)
	ms.driftTick++
	if ms.driftTick%8 == 0 {
		w, _ := ms.Game.ScreenSize()
		opts := ms.options()
		if ms.Selection >= 0 && ms.Selection < len(opts) {
			// menuY = menuStartY + titleLines + titleGap + subtitleOffset
			const menuY = menuStartY + titleLines + titleGap + subtitleOffset
			optY := menuY + ms.Selection*2
			textLen := StringWidth(opts[ms.Selection])
			textX := w/2 - textLen/2
			SpawnMenuDrift(ms.menuParticles, textX, optY, -1)
			SpawnMenuDrift(ms.menuParticles, textX+textLen-1, optY, 1)
		}
	}
}

func (ms *MenuScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	if !ms.starsSeeded || ms.starW != w || ms.starH != h {
		ms.seedStars(w, h)
	}

	// ── 1. Starfield — polar coords, origin behind title, rushing outward ──────
	// Y origin sits at the vertical midpoint of the title block (rows 2..7 → mid = 5)
	const starOriginY = 5
	halfW := float64(w) / 2.0
	halfH := float64(h) / 2.0
	for _, st := range ms.stars {
		bri := st.dist * st.baseBri
		sx := w/2 + int(math.Cos(st.angle)*st.dist*halfW)
		// 0.55 compresses vertical spread to account for taller terminal cells
		sy := starOriginY + int(math.Sin(st.angle)*st.dist*halfH*starVerticalSpread)
		if sx < 0 || sx >= w || sy < 0 || sy >= h {
			continue
		}
		rv := int32(bri * starColorScale)
		gv := int32(bri * starColorScale)
		bv := int32(starBlueBase + bri*starBlueScale)
		ch := starRunes[st.size]
		if st.dist < starMinDist {
			ch = '.'
		}
		ctx.SetCell(sx, sy, ch, StyleDefault.Foreground(tcell.NewRGBColor(rv, gv, bv)))
	}

	// ── 2. Drift particles (render behind title) ──────────────────────────────
	ms.menuParticles.DrawScreen(ctx.ScreenRaw)

	// ── 3. Title (existing per-character glow wave, unchanged) ────────────────
	title := []string{
		"████████╗███████╗██████╗ ███╗   ███╗       ██████╗ ██████╗ ███╗   ███╗",
		"╚══██╔══╝██╔════╝██╔══██╗████╗ ████║      ██╔════╝██╔═══██╗████╗ ████║",
		"   ██║   █████╗  ██████╔╝██╔████╔██║█████╗██║     ██║   ██║██╔████╔██║",
		"   ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║╚════╝██║     ██║   ██║██║╚██╔╝██║",
		"   ██║   ███████╗██║  ██║██║ ╚═╝ ██║      ╚██████╗╚██████╔╝██║ ╚═╝ ██║",
		"   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝       ╚═════╝ ╚═════╝ ╚═╝     ╚═╝",
	}

	nowSec := float64(time.Now().UnixNano()) / 1e9
	startY := menuStartY
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

	// ── 4. Version / Update box ───────────────────────────────────────────────
	versionStr := gameVersionStr()
	latest := ms.latestVersion.Load()
	if latest != nil {
		if v := latest.(string); v != "" {
			updateText := language.String("MENU_UPDATE_AVAILABLE")
			vLen := len([]rune(versionStr))
			uLen := len([]rune(updateText))
			contentW := vLen
			if uLen > contentW {
				contentW = uLen
			}
			contentW += 2
			boxW := contentW + 2
			boxX := w - boxW - 2
			boxY := 0

			// top border
			ctx.SetCell(boxX, boxY, '┌', StyleGray)
			for dx := 1; dx <= contentW; dx++ {
				ctx.SetCell(boxX+dx, boxY, '─', StyleGray)
			}
			ctx.SetCell(boxX+contentW+1, boxY, '┐', StyleGray)

			// row 1: version (centered)
			vOff := 1 + (contentW-vLen)/2
			ctx.SetCell(boxX, boxY+1, '│', StyleGray)
			for dx := 1; dx <= contentW; dx++ {
				ctx.SetCell(boxX+dx, boxY+1, ' ', StyleDefault)
			}
			ctx.DrawString(boxX+vOff, boxY+1, versionStr, StyleGray)
			ctx.SetCell(boxX+contentW+1, boxY+1, '│', StyleGray)

			// row 2: update text (centered)
			uOff := 1 + (contentW-uLen)/2
			ctx.SetCell(boxX, boxY+2, '│', StyleGray)
			for dx := 1; dx <= contentW; dx++ {
				ctx.SetCell(boxX+dx, boxY+2, ' ', StyleDefault)
			}
			ctx.DrawString(boxX+uOff, boxY+2, updateText, StyleYellow)
			ctx.SetCell(boxX+contentW+1, boxY+2, '│', StyleGray)

			// bottom border
			ctx.SetCell(boxX, boxY+3, '└', StyleGray)
			for dx := 1; dx <= contentW; dx++ {
				ctx.SetCell(boxX+dx, boxY+3, '─', StyleGray)
			}
			ctx.SetCell(boxX+contentW+1, boxY+3, '┘', StyleGray)
		} else {
			ctx.DrawString(w-len([]rune(versionStr))-2, 0, versionStr, StyleGray)
		}
	} else {
		ctx.DrawString(w-len([]rune(versionStr))-2, 0, versionStr, StyleGray)
	}

	// ── 5. Subtitle + decorations ─────────────────────────────────────────────
	subY := startY + len(title) + 1
	subtitle := language.String("MENU_TITLE")
	subX := (w - StringWidth(subtitle)) / 2
	if subX < 0 {
		subX = 0
	}
	ctx.DrawString(subX, subY, subtitle, StyleCyanBold)

	deco := "\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550"
	decX := (w - len([]rune(deco))) / 2
	if decX < 0 {
		decX = 0
	}
	ctx.DrawString(decX, subY-1, deco, StyleGray)
	ctx.DrawString(decX, subY+1, deco, StyleGray)

	// ── 5. Menu items ─────────────────────────────────────────────────────────
	menuY := subY + 4
	options := ms.options()

	// Bracket width: 0..2 extra spaces per side, driven by a 3 Hz sine
	expansion := int(math.Round((math.Sin(ms.bracketPhase*3.0)+1.0)/2.0*2.0))

	// Bracket color: neon cyan→white at 2.7 Hz (out-of-phase with width)
	bSin := math.Sin(ms.bracketPhase * 2.7)
	bracketStyle := StyleDefault.
		Foreground(tcell.NewRGBColor(
			int32(160.0+bSin*95.0),
			int32(220.0+bSin*35.0),
			255,
		)).Bold(true)

	// Selected text color: violet (#c040ff) → neon magenta (#ff40c0) at 2 Hz
	tPhase := (math.Sin(ms.bracketPhase*2.0) + 1.0) / 2.0
	selStyle := StyleDefault.
		Foreground(tcell.NewRGBColor(
			int32(192.0+tPhase*63.0),
			64,
			int32(255.0-tPhase*63.0),
		)).Bold(true)

	// Unselected: dim gray-purple so selected item pops
	dimStyle := StyleDefault.Foreground(tcell.NewRGBColor(0x58, 0x58, 0x68))

	for i, opt := range options {
		y := menuY + i*2
		textLen := StringWidth(opt)
		textX := w/2 - textLen/2

		// Fix: Draw background for option row
		for dx := -1-expansion; dx <= textLen+expansion; dx++ {
			if textX+dx >= 0 && textX+dx < w {
				ctx.SetCell(textX+dx, y, ' ', StyleDefault)
			}
		}

		if i == ms.Selection {
			// Brackets expand/contract symmetrically around the text
			ctx.SetCell(textX-1-expansion, y, '[', bracketStyle)
			ctx.SetCell(textX+textLen+expansion, y, ']', bracketStyle)
			ctx.DrawString(textX, y, opt, selStyle)
		} else {
			ctx.DrawString(textX, y, opt, dimStyle)
		}
	}

	// ── 6. Status bar ─────────────────────────────────────────────────────────
	ctx.DrawPanel(0, h-3, w, 3, "", StyleGray)
	if ms.Game.WebNotice != "" {
		ctx.DrawMarkupString(1, h-2, ms.Game.WebNotice, StyleCyanBold, StyleHotkey)
	} else {
		ctx.DrawMarkupString(1, h-2, language.String("MENU_HELP"), StyleGray, StyleHotkey)
	}
}

func (ms *MenuScreen) startVersionCheck() {
	if ms.updateCheckStarted {
		return
	}
	ms.updateCheckStarted = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/taislin/termcom/releases/latest", nil)
		if err != nil {
			return
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return
		}
		if release.TagName != "" && isNewerVersion(release.TagName, GameVersion) {
			ms.latestVersion.Store(release.TagName)
		}
	}()
}

func isNewerVersion(latest, current string) bool {
	la := strings.Split(strings.TrimPrefix(latest, "v"), ".")
	ca := strings.Split(strings.TrimPrefix(current, "v"), ".")
	maxLen := len(la)
	if len(ca) > maxLen {
		maxLen = len(ca)
	}
	for i := 0; i < maxLen; i++ {
		var lv, cv int
		if i < len(la) {
			lv, _ = strconv.Atoi(la[i])
		}
		if i < len(ca) {
			cv, _ = strconv.Atoi(ca[i])
		}
		if lv != cv {
			return lv > cv
		}
	}
	return false
}

func (ms *MenuScreen) options() []string {
	if HasSave() {
		return []string{language.String("MENU_NEW_GAME"), language.String("MENU_CONTINUE"), language.String("MENU_LOAD_GAME"), language.String("MENU_CUSTOM_BATTLE"), language.String("MENU_OPTIONS"), language.String("MENU_WEBSITE"), language.String("MENU_QUIT")}
	}
	return []string{language.String("MENU_NEW_GAME"), language.String("MENU_CUSTOM_BATTLE"), language.String("MENU_OPTIONS"), language.String("MENU_WEBSITE"), language.String("MENU_QUIT")}
}

func (ms *MenuScreen) HandleKey(e *tcell.EventKey) {
	ms.Game.WebNotice = ""
	opts := ms.options()
	maxSel := len(opts) - 1

	switch e.Key() {
	case tcell.KeyUp:
		ms.Selection--
		if ms.Selection < 0 {
			ms.Selection = maxSel
		}
	case tcell.KeyDown:
		ms.Selection++
		if ms.Selection > maxSel {
			ms.Selection = 0
		}
	case tcell.KeyEnter:
		ms.confirm()
	}
	switch e.Str() {
	case "q", "Q":
		ms.Game.Quit()
	case "j", "J":
		ms.Selection++
		if ms.Selection > maxSel {
			ms.Selection = 0
		}
	case "k", "K":
		ms.Selection--
		if ms.Selection < 0 {
			ms.Selection = maxSel
		}
	case "1":
		ms.handleNumericShortcut(1, opts)
	case "2":
		ms.handleNumericShortcut(2, opts)
	case "3":
		ms.handleNumericShortcut(3, opts)
	case "4":
		ms.handleNumericShortcut(4, opts)
	case "5":
		ms.handleNumericShortcut(5, opts)
	case "6":
		ms.handleNumericShortcut(6, opts)
	}
}

func (ms *MenuScreen) handleNumericShortcut(num int, opts []string) {
	idx := num - 1
	if idx < len(opts) {
		ms.Selection = idx
		ms.confirm()
	}
}

func (ms *MenuScreen) confirm() {
	opts := ms.options()
	if ms.Selection < 0 || ms.Selection >= len(opts) {
		return
	}
	switch opts[ms.Selection] {
	case language.String("MENU_NEW_GAME"):
		if ms.Game.OnNewGame != nil {
			ms.Game.OnNewGame()
		}
	case language.String("MENU_CONTINUE"):
		if ms.Game.OnContinue != nil {
			ms.Game.OnContinue()
		}
	case language.String("MENU_LOAD_GAME"):
		if ms.Game.OnLoadGame != nil {
			ms.Game.OnLoadGame()
		}
	case language.String("MENU_OPTIONS"):
		if _, ok := ms.Game.screens[StateOptions]; !ok {
			ms.Game.SetScreen(StateOptions, NewOptionsScreen(ms.Game))
		}
		ms.Game.PushState(StateOptions)
	case language.String("MENU_CUSTOM_BATTLE"):
		if ms.Game.OnCustomBattle != nil {
			ms.Game.OnCustomBattle()
		}
	case language.String("MENU_QUIT"):
		ms.Game.Quit()
	case language.String("MENU_WEBSITE"):
		if ms.Game.IsWeb() {
			ms.Game.WebNotice = WebsiteURL
		} else {
			if err := openBrowser(WebsiteURL); err != nil {
			// Silently ignore; most terminals can't open URLs anyway
		}
		}
	}
}

func (ms *MenuScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := ms.Game.ScreenSize()

	// Help bar clicks (y = h-2)
	if y == h-2 {
		help := language.String("MENU_HELP")
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
				case "Q", "q":
					ms.Game.Quit()
				case "F5":
					ms.Game.PushState(StateSlotPicker)
				case "F9":
					ms.Game.PushState(StateSlotPicker)
				case "Enter":
					ms.confirm()
				}
				return
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	// Update notification box click
	if latest := ms.latestVersion.Load(); latest != nil && latest.(string) != "" {
		vLen := len([]rune(gameVersionStr()))
		uLen := len([]rune(language.String("MENU_UPDATE_AVAILABLE")))
		contentW := vLen
		if uLen > contentW {
			contentW = uLen
		}
		contentW += 2
		boxW := contentW + 2
		boxX := w - boxW - 2
		if y >= 0 && y <= 3 && x >= boxX && x < boxX+boxW {
			if buttons&tcell.Button1 != 0 {
				openBrowser("https://github.com/taislin/termcom/releases/latest")
			}
			return
		}
	}

	subY := 9
	menuY := subY + 4
	opts := ms.options()

	for i := range opts {
		if y == menuY+i*2 && x >= w/2-10 && x <= w/2+10 {
			ms.Selection = i
			if buttons&tcell.Button1 != 0 {
				ms.confirm()
			}
			return
		}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("cmd", "/c", "start", "", url)
	}
	return cmd.Start()
}
