package base

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type WeaponDesignerScreen struct {
	Game      *engine.Game
	Base      *Base
	Design    data.WeaponDesign
	Param     int
	Message   string
	nextID    int
}

func NewWeaponDesignerScreen(g *engine.Game, b *Base) *WeaponDesignerScreen {
	nextID := len(b.CustomWeapons)
	return &WeaponDesignerScreen{
		Game:   g,
		Base:   b,
		Design: data.WeaponDesign{BaseType: "rifle", Barrel: 1, Optics: 0, Auto: false, AmmoType: 0, Stock: 1},
		nextID: nextID,
	}
}

func (wd *WeaponDesignerScreen) Update() {}

func (wd *WeaponDesignerScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h-3, language.String("WEAPON_DESIGNER_TITLE"), engine.StyleDefault)

	leftW := w * 45 / 100
	rightX := leftW + 4
	rightW := w - rightX - 2

	wd.renderPreview(ctx, 2, 4, leftW, h-10)
	wd.renderStats(ctx, rightX, 4, rightW, h-10)
	wd.renderParams(ctx, 2, h-8, w-4)

	fundsStr := fmt.Sprintf(language.String("GEOSCAPE_FUNDS"), wd.Game.Funds/1000)
	ctx.DrawString(w/2, h-3, fundsStr, engine.StyleGreen)
	if wd.Message != "" {
		ctx.DrawString(w*3/4, h-3, wd.Message, engine.StyleYellow)
	}
	help := "[\u2191\u2193] Nav  [\u2190\u2192] Adjust  [Tab] Next  [Enter] Build  [Esc] Cancel"
	ctx.DrawMarkupString(2, h-1, help, engine.StyleGray, engine.StyleHotkey)
}

func (wd *WeaponDesignerScreen) renderPreview(ctx *engine.ScreenCtx, px, py, pw, ph int) {
	ctx.DrawString(px, py-1, language.String("WEAPON_PREVIEW"), engine.StyleCyanBold)

	name := data.WeaponDesignName(wd.Design)
	ctx.DrawString(px+1, py+1, name, engine.StyleYellow)

	// Draw ASCII weapon art
	cells := wd.renderWeaponArt()
	offsetX := px + 2
	offsetY := py + 4

	for _, c := range cells {
		sx := offsetX + c.X
		sy := offsetY + c.Y
		if sx >= px && sx < px+pw && sy >= py && sy < py+ph {
			ctx.SetCell(sx, sy, c.Rune, c.Style)
		}
	}

	// Show base type
	ctx.DrawString(px+1, py+ph-2, fmt.Sprintf("Base: %s", wd.Design.BaseType), engine.StyleGray)
	ctx.DrawString(px+1, py+ph-1, fmt.Sprintf("Cost: $%dk", wd.cost()/1000), engine.StyleGray)
}

type weaponCell struct {
	X, Y  int
	Rune  rune
	Style tcell.Style
}

func (wd *WeaponDesignerScreen) renderWeaponArt() []weaponCell {
	var cells []weaponCell
	barrel := wd.Design.Barrel
	optics := wd.Design.Optics
	isAuto := wd.Design.Auto
	stock := wd.Design.Stock

	// Muzzle
	muzzleLen := 1 + barrel
	for i := 0; i < muzzleLen; i++ {
		cells = append(cells, weaponCell{X: i, Y: 1, Rune: '\u2500', Style: engine.StyleCyan})
	}
	// Muzzle tip
	cells = append(cells, weaponCell{X: muzzleLen, Y: 1, Rune: '\u25B8', Style: engine.StyleCyanBold})

	// Barrel body
	barrelStart := muzzleLen + 1
	barrelEnd := barrelStart + 2
	for x := barrelStart; x <= barrelEnd; x++ {
		cells = append(cells, weaponCell{X: x, Y: 0, Rune: '\u2580', Style: engine.StyleCyan})
		cells = append(cells, weaponCell{X: x, Y: 1, Rune: '\u2588', Style: engine.StyleCyan})
		cells = append(cells, weaponCell{X: x, Y: 2, Rune: '\u2584', Style: engine.StyleCyan})
	}

	// Optics (on top of receiver)
	receiverStart := barrelEnd + 1
	if optics > 0 {
		opticChar := '\u25CE'
		if optics >= 2 {
			opticChar = '\u25C9'
		}
		if optics >= 3 {
			opticChar = '\u25DF'
		}
		cells = append(cells, weaponCell{X: receiverStart + 1, Y: 0, Rune: opticChar, Style: engine.StyleYellow})
	}

	// Receiver
	receiverEnd := receiverStart + 2
	for x := receiverStart; x <= receiverEnd; x++ {
		cells = append(cells, weaponCell{X: x, Y: 1, Rune: '\u25A0', Style: engine.StyleCyanBold})
	}

	// Grip
	cells = append(cells, weaponCell{X: receiverEnd - 1, Y: 2, Rune: '\u2514', Style: engine.StyleCyan})
	cells = append(cells, weaponCell{X: receiverEnd - 1, Y: 3, Rune: '\u2502', Style: engine.StyleCyan})

	// Stock
	stockStart := receiverEnd + 1
	stockLen := 2 + stock
	if stock > 0 {
		for i := 0; i < stockLen; i++ {
			cells = append(cells, weaponCell{X: stockStart + i, Y: 1, Rune: '\u2588', Style: engine.StyleCyan})
		}
	}

	// Magazine (if ammo type > 0)
	if wd.Design.AmmoType > 0 {
		magChar := '\u25BC'
		cells = append(cells, weaponCell{X: receiverStart + 1, Y: 2, Rune: magChar, Style: engine.StyleRed})
		cells = append(cells, weaponCell{X: receiverStart + 1, Y: 3, Rune: '\u2502', Style: engine.StyleRed})
	}

	// Auto indicator
	if isAuto {
		cells = append(cells, weaponCell{X: receiverEnd + 1, Y: 0, Rune: '\u25B6', Style: engine.StyleGreen})
		cells = append(cells, weaponCell{X: receiverEnd + 2, Y: 0, Rune: '\u25B6', Style: engine.StyleGreen})
	}

	return cells
}

func (wd *WeaponDesignerScreen) renderStats(ctx *engine.ScreenCtx, sx, sy, sw, sh int) {
	ctx.DrawString(sx, sy-1, language.String("WEAPON_STATS"), engine.StyleCyanBold)

	damage, accuracy, tu, rng, ammoMax, strength, weight, _ := data.CalcDesignStats(wd.Design)

	rows := []struct {
		label string
		val   string
		style tcell.Style
	}{
		{language.String("WEAPON_STAT_DAMAGE"), fmt.Sprintf("%d", damage), engine.StyleRed},
		{language.String("WEAPON_STAT_ACCURACY"), fmt.Sprintf("%d%%", accuracy), engine.StyleGreen},
		{language.String("WEAPON_STAT_TU"), fmt.Sprintf("%d TU", tu), engine.StyleYellow},
		{language.String("WEAPON_STAT_RANGE"), fmt.Sprintf("%d", rng), engine.StyleCyan},
		{language.String("WEAPON_STAT_AMMO"), fmt.Sprintf("%d", ammoMax), engine.StyleCyan},
		{language.String("WEAPON_STAT_WEIGHT"), fmt.Sprintf("%.1f kg", weight), engine.StyleGray},
		{language.String("WEAPON_STAT_STR"), fmt.Sprintf("%d", strength), engine.StyleGray},
	}

	for i, r := range rows {
		if sy+i >= sy+sh {
			break
		}
		ctx.DrawString(sx, sy+i, r.label+": ", engine.StyleGray)
		ctx.DrawString(sx+18, sy+i, r.val, r.style)
	}
}

func (wd *WeaponDesignerScreen) renderParams(ctx *engine.ScreenCtx, px, py, pw int) {
	ctx.DrawString(px, py-1, language.String("WEAPON_PARAMETERS"), engine.StyleCyanBold)

	labels := data.WeaponDesignBarLabels(wd.Design)
	paramNames := []string{
		language.String("WEAPON_PARAM_BARREL"),
		language.String("WEAPON_PARAM_OPTICS"),
		language.String("WEAPON_PARAM_FIREMODE"),
		language.String("WEAPON_PARAM_AMMO"),
		language.String("WEAPON_PARAM_STOCK"),
	}
	for i, name := range paramNames {
		style := engine.StyleDefault
		if i == wd.Param {
			style = engine.StyleHighlight
		}
		bar := ""
		if i < len(labels) {
			bar = labels[i]
		}
		ctx.DrawString(px, py+1+i, fmt.Sprintf("%-12s  < %s >", name, bar), style)
	}
}

func (wd *WeaponDesignerScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		audio.PlayMenuNav()
		wd.Param--
		if wd.Param < 0 {
			wd.Param = 4
		}
	case tcell.KeyDown:
		audio.PlayMenuNav()
		wd.Param++
		if wd.Param > 4 {
			wd.Param = 0
		}
	case tcell.KeyLeft:
		audio.PlayMenuNav()
		wd.adjustParam(-1)
	case tcell.KeyRight:
		audio.PlayMenuNav()
		wd.adjustParam(1)
	case tcell.KeyTab:
		audio.PlayMenuNav()
		wd.Param++
		if wd.Param > 4 {
			wd.Param = 0
		}
	case tcell.KeyEnter:
		wd.build()
	case tcell.KeyEscape:
		wd.Game.PopState()
	}
	switch e.Str() {
	case "1":
		wd.adjustParam(-1)
	case "2":
		wd.adjustParam(1)
	case "q", "Q":
		wd.Game.PopState()
	}
}

func (wd *WeaponDesignerScreen) adjustParam(delta int) {
	switch wd.Param {
	case 0: // barrel
		wd.Design.Barrel += delta
		if wd.Design.Barrel < 0 {
			wd.Design.Barrel = len(data.Barrels) - 1
		}
		if wd.Design.Barrel >= len(data.Barrels) {
			wd.Design.Barrel = 0
		}
	case 1: // optics
		wd.Design.Optics += delta
		if wd.Design.Optics < 0 {
			wd.Design.Optics = len(data.OpticsList) - 1
		}
		if wd.Design.Optics >= len(data.OpticsList) {
			wd.Design.Optics = 0
		}
	case 2: // fire mode
		wd.Design.Auto = !wd.Design.Auto
	case 3: // ammo
		wd.Design.AmmoType += delta
		if wd.Design.AmmoType < 0 {
			wd.Design.AmmoType = len(data.AmmoTypes) - 1
		}
		if wd.Design.AmmoType >= len(data.AmmoTypes) {
			wd.Design.AmmoType = 0
		}
	case 4: // stock
		wd.Design.Stock += delta
		if wd.Design.Stock < 0 {
			wd.Design.Stock = len(data.Stocks) - 1
		}
		if wd.Design.Stock >= len(data.Stocks) {
			wd.Design.Stock = 0
		}
	}
}

func (wd *WeaponDesignerScreen) cost() int {
	_, _, _, _, _, _, _, cost := data.CalcDesignStats(wd.Design)
	return cost
}

func (wd *WeaponDesignerScreen) build() {
	cost := int64(wd.cost())
	if wd.Game.Funds < cost {
		wd.Message = language.String("WEAPON_MSG_INSUFFICIENT_FUNDS")
		return
	}

	wd.Game.Funds -= cost
	item := data.MakeDesignItem(wd.Design)
	item.Type = fmt.Sprintf("custom_%d", wd.nextID)
	wd.Design.ID = item.Type

	// Register in RuleItems so battlescape can look up stats
	data.RuleItems[item.Type] = item

	wd.Base.CustomWeapons[item.Type] = &wd.Design
	wd.Base.Stores[item.Type] = 1

	wd.Message = fmt.Sprintf(language.String("WEAPON_MSG_BUILT"), data.WeaponDesignName(wd.Design))
	wd.Game.PopState()
}

func (wd *WeaponDesignerScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := wd.Game.ScreenSize()

	if y >= h-8 && y <= h-4 {
		idx := y - (h - 8) - 1
		if idx >= 0 && idx < 5 {
			wd.Param = idx
			if x > 20 {
				wd.adjustParam(1)
			} else {
				wd.adjustParam(-1)
			}
		}
	}
}
