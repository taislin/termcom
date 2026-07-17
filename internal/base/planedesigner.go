package base

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type PlaneDesignerScreen struct {
	Game     *engine.Game
	Base     *Base
	HangarID int
	Config   data.PlaneConfig
	Orig     data.PlaneConfig
	Param    int
	Message  string
}

func NewPlaneDesignerScreen(g *engine.Game, b *Base, hangarID int) *PlaneDesignerScreen {
	cfg := data.DefaultPlaneConfig()
	if hangarID >= 0 && hangarID < len(b.Hangars) {
		hg := b.Hangars[hangarID]
		if hg.PlaneConfig != nil {
			cfg = *hg.PlaneConfig
		}
	}
	return &PlaneDesignerScreen{
		Game:     g,
		Base:     b,
		HangarID: hangarID,
		Config:   cfg,
		Orig:     cfg,
	}
}

func (pd *PlaneDesignerScreen) Update() {}

func (pd *PlaneDesignerScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h-3, language.String("PLANE_DESIGNER_TITLE"), engine.StyleDefault)

	leftW := w * 45 / 100
	rightX := leftW + 4
	rightW := w - rightX - 2

	paramY := h - 8
	pd.renderPreview(ctx, 2, 3, leftW, paramY-4)
	pd.renderStats(ctx, rightX, 3, rightW, paramY-4)
	pd.renderParams(ctx, 2, paramY, rightX)

	fundsStr := fmt.Sprintf(language.String("GEOSCAPE_FUNDS"), pd.Game.Funds/engine.FundsDisplayK)
	ctx.DrawString(w/2, h-3, fundsStr, engine.StyleGreen)
	if pd.Message != "" {
		ctx.DrawString(w*3/4, h-3, pd.Message, engine.StyleYellow)
	}
	help := language.String("PLANE_DESIGNER_HELP")
	ctx.DrawMarkupString(2, h-1, help, engine.StyleGray, engine.StyleHotkey)
}

func (pd *PlaneDesignerScreen) renderPreview(ctx *engine.ScreenCtx, px, py, pw, ph int) {
	ctx.DrawString(px, py-1, language.String("PLANE_PREVIEW"), engine.StyleCyanBold)

	cells := data.RenderPlanePreview(pd.Config)
	if len(cells) == 0 {
		return
	}

	// Compute bounding box so we can centre both axes.
	minX, maxX := cells[0].X, cells[0].X
	minY, maxY := cells[0].Y, cells[0].Y
	for _, c := range cells {
		if c.X < minX {
			minX = c.X
		}
		if c.X > maxX {
			maxX = c.X
		}
		if c.Y < minY {
			minY = c.Y
		}
		if c.Y > maxY {
			maxY = c.Y
		}
	}
	previewW := maxX - minX + 1
	previewH := maxY - minY + 1

	// Reserve one row at the bottom for labels.
	drawH := ph - 1
	if drawH < 1 {
		drawH = 1
	}
	offsetX := px + (pw-previewW)/2
	offsetY := py + (drawH-previewH)/2 - minY

	for _, c := range cells {
		sx := offsetX + c.X - minX
		sy := offsetY + c.Y
		if sx >= px && sx < px+pw && sy >= py && sy < py+drawH {
			style := engine.StyleCyan
			switch c.Rune {
			case '\u258C', '\u25A0', '\u25A3':
				style = engine.StyleCyanBold
			case '\u2501':
				style = engine.StyleYellow
			}
			ctx.SetCell(sx, sy, c.Rune, style)
		}
	}

	stats := data.CalcPlaneStats(pd.Config)
	labelY := py + ph - 1
	if labelY < py {
		labelY = py
	}
	dimLabel := fmt.Sprintf(language.String("PLANE_LABEL_DIMS"), pd.Config.Length, pd.Config.Wingspan*2+1)
	ctx.DrawString(px, labelY, dimLabel, engine.StyleGray)
	speedLabel := fmt.Sprintf("%s %.1f", language.String("PLANE_LABEL_SPEED"), stats.Speed)
	if len(speedLabel) <= pw {
		ctx.DrawString(px+pw-len(speedLabel), labelY, speedLabel, engine.StyleGray)
	}
}

func (pd *PlaneDesignerScreen) renderStats(ctx *engine.ScreenCtx, sx, sy, sw, sh int) {
	ctx.DrawString(sx, sy-1, language.String("PLANE_STATS"), engine.StyleCyanBold)
	stats := data.CalcPlaneStats(pd.Config)

	rows := []struct {
		label string
		val   string
		style tcell.Style
	}{
		{language.String("PLANE_STAT_SPEED"), fmt.Sprintf("%.1f", stats.Speed), engine.StyleGreen},
		{language.String("PLANE_STAT_FIREPOWER"), fmt.Sprintf("%.0f", stats.Firepower), engine.StyleRed},
		{language.String("PLANE_STAT_HULL"), fmt.Sprintf("%d", stats.Hull), engine.StyleYellow},
		{language.String("PLANE_STAT_MASS"), fmt.Sprintf("%.1f%s", stats.Mass, language.String("UNIT_TONNES")), engine.StyleGray},
		{language.String("PLANE_STAT_THRUST"), fmt.Sprintf("%.0f%s", stats.Thrust, language.String("UNIT_KILONEWTON")), engine.StyleGreen},
		{language.String("PLANE_STAT_RANGE"), fmt.Sprintf("%d", stats.Range), engine.StyleCyan},
	}
	for i, r := range rows {
		if sy+i >= sy+sh {
			break
		}
		ctx.DrawString(sx, sy+i, r.label+":", engine.StyleGray)
	}
	// Find longest label for alignment.
	maxLabel := 0
	for _, r := range rows {
		if len(r.label) > maxLabel {
			maxLabel = len(r.label)
		}
	}
	valCol := sx + maxLabel + 2
	for i, r := range rows {
		if sy+i >= sy+sh {
			break
		}
		ctx.DrawString(valCol, sy+i, r.val, r.style)
	}

	yOff := sy + len(rows) + 1
	if pd.Config.Weapon >= 0 && pd.Config.Weapon < len(data.PlaneWeapons) {
		w := data.PlaneWeapons[pd.Config.Weapon]
		ctx.DrawString(sx, yOff, language.String("PLANE_LABEL_WEAPON"), engine.StyleGray)
		ctx.DrawString(sx+2, yOff+1, language.Sprintf("PLANE_WEAPON_LINE", w.Name, w.Damage, w.Accuracy), engine.StyleRed)
	}
	yOff += 3
	if pd.Config.Armor >= 0 && pd.Config.Armor < len(data.PlaneArmors) {
		a := data.PlaneArmors[pd.Config.Armor]
		ctx.DrawString(sx, yOff, language.String("PLANE_LABEL_ARMOR"), engine.StyleGray)
		ctx.DrawString(sx+2, yOff+1, language.Sprintf("PLANE_ARMOR_LINE", a.Name, a.HP, a.DR), engine.StyleYellow)
	}
}

func (pd *PlaneDesignerScreen) renderParams(ctx *engine.ScreenCtx, px, py, cx int) {
	ctx.DrawString(px, py-1, language.String("PLANE_PARAMETERS"), engine.StyleCyanBold)
	params := pd.paramList()
	for i, p := range params {
		colX := px
		colY := py + 1 + i
		if i >= 3 {
			colX = cx
			colY = py + 1 + (i - 3)
		}
		style := engine.StyleDefault
		if i == pd.Param {
			style = engine.StyleHighlight
		}
		ctx.DrawString(colX, colY, fmt.Sprintf("%-14s %s", p.label, p.value), style)
	}
}

type paramInfo struct {
	label string
	value string
}

func (pd *PlaneDesignerScreen) paramList() []paramInfo {
	cfg := pd.Config
	wpn := "---"
	if cfg.Weapon >= 0 && cfg.Weapon < len(data.PlaneWeapons) {
		wpn = data.PlaneWeapons[cfg.Weapon].Name
	}
	arm := "---"
	if cfg.Armor >= 0 && cfg.Armor < len(data.PlaneArmors) {
		arm = data.PlaneArmors[cfg.Armor].Name
	}
	return []paramInfo{
		{language.String("PLANE_PARAM_LENGTH"), fmt.Sprintf("[%d] %s", cfg.Length, pd.bar(cfg.Length, 3, 7))},
		{language.String("PLANE_PARAM_WINGSPAN"), fmt.Sprintf("[%d] %s (%d %s)", cfg.Wingspan, pd.bar(cfg.Wingspan, 1, 4), cfg.Wingspan*2+1, language.String("UNIT_CELLS"))},
		{language.String("PLANE_PARAM_ENGINES"), fmt.Sprintf("[%d] %s", cfg.Engines, pd.bar(cfg.Engines, 1, 3))},
		{language.String("PLANE_PARAM_FUEL"), fmt.Sprintf("[%d] %s", cfg.Fuel, pd.bar(cfg.Fuel/10, 2, 10))},
		{language.String("PLANE_PARAM_WEAPON"), fmt.Sprintf("< %s >", wpn)},
		{language.String("PLANE_PARAM_ARMOR"), fmt.Sprintf("< %s >", arm)},
	}
}

func (pd *PlaneDesignerScreen) bar(val, min, max int) string {
	n := 10
	filled := 0
	if max > min {
		filled = (val - min) * n / (max - min)
	}
	if filled < 0 {
		filled = 0
	}
	if filled > n {
		filled = n
	}
	return strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", n-filled)
}

func (pd *PlaneDesignerScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		audio.PlayMenuNav()
		pd.Param--
		if pd.Param < 0 {
			pd.Param = 5
		}
	case tcell.KeyDown:
		audio.PlayMenuNav()
		pd.Param++
		if pd.Param > 5 {
			pd.Param = 0
		}
	case tcell.KeyLeft:
		audio.PlayMenuNav()
		pd.adjustParam(-1)
	case tcell.KeyRight:
		audio.PlayMenuNav()
		pd.adjustParam(1)
	case tcell.KeyTab:
		audio.PlayMenuNav()
		pd.Param++
		if pd.Param > 5 {
			pd.Param = 0
		}
	case tcell.KeyEnter:
		pd.save()
	case tcell.KeyEscape:
		pd.cancel()
	}
	switch e.Str() {
	case "r", "R":
		pd.Config = pd.Orig
		pd.Message = language.String("PLANE_MSG_RESET")
	case "1":
		pd.adjustParam(-1)
	case "2":
		pd.adjustParam(1)
	case "q", "Q":
		pd.cancel()
	}
}

func (pd *PlaneDesignerScreen) adjustParam(delta int) {
	switch pd.Param {
	case 0:
		pd.Config.Length = clamp(pd.Config.Length+delta, 3, 7)
	case 1:
		pd.Config.Wingspan = clamp(pd.Config.Wingspan+delta, 1, 4)
	case 2:
		pd.Config.Engines = clamp(pd.Config.Engines+delta, 1, 3)
	case 3:
		pd.Config.Fuel = clamp(pd.Config.Fuel+delta*10, 20, 100)
	case 4:
		pd.Config.Weapon += delta
		if pd.Config.Weapon < 0 {
			pd.Config.Weapon = len(data.PlaneWeapons) - 1
		}
		if pd.Config.Weapon >= len(data.PlaneWeapons) {
			pd.Config.Weapon = 0
		}
	case 5:
		pd.Config.Armor += delta
		if pd.Config.Armor < 0 {
			pd.Config.Armor = len(data.PlaneArmors) - 1
		}
		if pd.Config.Armor >= len(data.PlaneArmors) {
			pd.Config.Armor = 0
		}
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (pd *PlaneDesignerScreen) save() {
	cfg := pd.Config
	if pd.HangarID >= 0 && pd.HangarID < len(pd.Base.Hangars) {
		pd.Base.Hangars[pd.HangarID].PlaneConfig = &cfg
		pd.Base.Hangars[pd.HangarID].Name = pd.planeName()
	}
	pd.Game.PopState()
}

func (pd *PlaneDesignerScreen) cancel() {
	pd.Game.PopState()
}

func (pd *PlaneDesignerScreen) planeName() string {
	cfg := pd.Config
	var parts []string
	switch cfg.Engines {
	case 1:
		parts = append(parts, language.String("PLANE_CLASS_LIGHT"))
	case 2:
		parts = append(parts, language.String("PLANE_CLASS_MEDIUM"))
	case 3:
		parts = append(parts, language.String("PLANE_CLASS_HEAVY"))
	}
	if cfg.Weapon >= 0 && cfg.Weapon < len(data.PlaneWeapons) {
		parts = append(parts, data.PlaneWeapons[cfg.Weapon].Name)
	}
	parts = append(parts, language.String("PLANE_NAME_FIGHTER"))
	return strings.Join(parts, " ")
}

func (pd *PlaneDesignerScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := pd.Game.ScreenSize()

	const (
		leftColPct   = 45
		colGap       = 4
		paramOffsetY = 8
	)
	leftW := w * leftColPct / 100
	rightX := leftW + colGap
	paramY := h - paramOffsetY

	if y >= paramY+1 && y <= paramY+3 {
		var idx int
		if x < rightX {
			idx = y - (paramY + 1)
		} else {
			idx = y - (paramY + 1) + 3
		}
		if idx >= 0 && idx < 6 {
			pd.Param = idx
			if x > paramOffsetY*2+4 {
				pd.adjustParam(1)
			} else {
				pd.adjustParam(-1)
			}
		}
	}
}
