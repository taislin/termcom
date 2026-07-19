package base

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type LoadoutScreen struct {
	Game        *engine.Game
	Base        *Base
	SelectedSol int
	Message     string
}

func NewLoadoutScreen(g *engine.Game, b *Base) *LoadoutScreen {
	return &LoadoutScreen{
		Game: g,
		Base: b,
	}
}

func (ls *LoadoutScreen) Update() {}

func (ls *LoadoutScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, "Armory", engine.StyleDefault)

	if len(ls.Base.Soldiers) == 0 {
		ctx.DrawString(2, 3, language.String("NO_SOLDIERS"), engine.StyleGray)
		ctx.DrawString(2, 5, language.String("PRESS_ESC"), engine.StyleGray)
		return
	}

	if ls.SelectedSol >= len(ls.Base.Soldiers) {
		ls.SelectedSol = len(ls.Base.Soldiers) - 1
	}
	if ls.SelectedSol < 0 {
		ls.SelectedSol = 0
	}

	// Header
	cols := []string{
		"#", "Name", "Rank", "HP", "TU", "Weapon", "Ammo", "Armor", "Enc",
	}
	colWidths := []int{3, 16, 6, 5, 5, 18, 7, 14, 5}
	x := 2
	headerY := 2
	ctx.DrawString(x, headerY, language.String("SECTION_SOLDIER"), engine.StyleCyanBold)
	x2 := 2
	for i, col := range cols {
		w2 := colWidths[i]
		ctx.DrawString(x2, headerY+1, col, engine.StyleGray)
		x2 += w2
	}

	// Row data
	tableX := 2
	for i, s := range ls.Base.Soldiers {
		style := engine.StyleDefault
		if i == ls.SelectedSol {
			style = engine.StyleHighlight
		}
		y2 := headerY + 2 + i
		if y2 >= h-2 {
			break
		}

		rankStr := s.Rank.String()
		tu := s.TU
		maxTU := s.MaxTU
		pen := s.TUPenalty()
		tuStr := fmt.Sprintf("%d/%d", tu, maxTU)
		if pen > 0 {
			tuStr = fmt.Sprintf("%d/%d(-%d)", tu, maxTU, pen)
		}

		wpnStr := s.Weapon
		if wpnStr == "" {
			wpnStr = "none"
		} else if d, ok := data.RuleItems[s.Weapon]; ok {
			wpnStr = d.ShortName
			if wpnStr == "" {
				wpnStr = d.Name
			}
		}

		ammoStr := ""
		if s.Weapon != "" {
			if d, ok := data.RuleItems[s.Weapon]; ok && d.AmmoMax < 99 {
				ammoStr = fmt.Sprintf("%d/%d", s.WeaponAmmo, d.AmmoMax)
			} else {
				ammoStr = "--"
			}
		} else {
			ammoStr = "--"
		}

		armStr := s.Armor
		if armStr == "" {
			armStr = "none"
		}

		enc := s.Encumbrance()

		rowData := []string{
			fmt.Sprintf("%2d", i+1),
			truncStr(s.Name, colWidths[1]-1),
			rankStr,
			fmt.Sprintf("%d", s.HP),
			tuStr,
			truncStr(wpnStr, colWidths[5]-1),
			ammoStr,
			truncStr(armStr, colWidths[7]-1),
			fmt.Sprintf("%d", enc),
		}
		cx := tableX
		for j, val := range rowData {
			ctx.DrawString(cx, y2, val, style)
			cx += colWidths[j]
		}
	}

	// Selected soldier detailed info
	sel := ls.Base.Soldiers[ls.SelectedSol]
	detailY := headerY + 2 + len(ls.Base.Soldiers) + 1
	if detailY < h-3 {
		ctx.DrawString(x, detailY, "Selected: "+sel.Name, engine.StyleCyan)
		detailY++
		ctx.DrawString(x+2, detailY, fmt.Sprintf("Rank: %s  HP: %d/%d  TU: %d/%d  Str: %d",
			sel.Rank.String(), sel.HP, sel.MaxHP, sel.TU, sel.MaxTU, sel.Strength), engine.StyleDefault)
		detailY++
		ctx.DrawString(x+2, detailY, fmt.Sprintf("Weapon: %s  Armor: %s",
			sel.Weapon, sel.Armor), engine.StyleDefault)
		detailY++
		ctx.DrawString(x+2, detailY, fmt.Sprintf("Accuracy: %d  Reactions: %d  Bravery: %d",
			sel.Accuracy, sel.Reactions, sel.Bravery), engine.StyleDefault)
		detailY++
		ctx.DrawString(x+2, detailY, fmt.Sprintf("Encumbrance: %d  TU Penalty: -%d",
			sel.Encumbrance(), sel.TUPenalty()), engine.StyleDefault)
	}

	// Bottom help bar
	helpY := h - 1
	ctx.DrawPanel(0, helpY, w, 1, "", engine.StyleGray)
	help := "[↑/↓] Select  [A]uto-equip all  [E]quip  [Esc] Back"
	ctx.DrawMarkupString(1, helpY, help, engine.StyleGray, engine.StyleHotkey)

	if ls.Message != "" {
		msgY := h - 3
		ctx.DrawString(2, msgY, ls.Message, engine.StyleYellow)
	}
}

func (ls *LoadoutScreen) HandleMouse(e *tcell.EventMouse) {}

func (ls *LoadoutScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyEscape:
		ls.Game.PopState()
	case tcell.KeyUp:
		if ls.SelectedSol > 0 {
			ls.SelectedSol--
		}
	case tcell.KeyDown:
		if ls.SelectedSol < len(ls.Base.Soldiers)-1 {
			ls.SelectedSol++
		}
	case tcell.KeyEnter:
		ls.openEquip()
	}
	switch e.Str() {
	case "a", "A":
		ls.autoEquipAll()
	case "e", "E":
		ls.openEquip()
	}
}

func (ls *LoadoutScreen) openEquip() {
	ls.Game.SetScreen(engine.StateEquip, NewEquipScreen(ls.Game, ls.Base))
	ls.Game.PushState(engine.StateEquip)
}

func (ls *LoadoutScreen) autoEquipAll() {
	screen := NewEquipScreen(ls.Game, ls.Base)
	screen.autoEquip()
	count := len(ls.Base.Soldiers)
	ls.Message = fmt.Sprintf("Auto-equipped %d soldiers.", count)
}

func truncStr(s string, maxLen int) string {
	if engine.StringWidth(s) <= maxLen {
		return s
	}
	rs := []rune(s)
	for len(rs) > 0 && engine.StringWidth(string(rs)) > maxLen {
		rs = rs[:len(rs)-1]
	}
	return string(rs)
}
