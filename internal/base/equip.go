package base

import (
	"fmt"
	"sort"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/gdamore/tcell/v2"
)

type EquipScreen struct {
	Game         *engine.Game
	Base         *Base
	SelectedSol  int
	SelectedSlot int // 0=weapon, 1=armor
	CycleIdx     int
	Message      string
}

func NewEquipScreen(g *engine.Game, b *Base) *EquipScreen {
	return &EquipScreen{
		Game: g,
		Base: b,
	}
}

func (es *EquipScreen) Update() {}

func (es *EquipScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, "EQUIP SOLDIERS", engine.StyleDefault)

	if len(es.Base.Soldiers) == 0 {
		ctx.DrawString(2, 3, "No soldiers in roster.", engine.StyleGray)
		ctx.DrawString(2, 5, "Press Esc to return.", engine.StyleGray)
		return
	}

	if es.SelectedSol >= len(es.Base.Soldiers) {
		es.SelectedSol = len(es.Base.Soldiers) - 1
	}

	rightX := w / 2

	ctx.DrawString(2, 2, "SOLDIER:", engine.StyleCyanBold)
	for i, s := range es.Base.Soldiers {
		style := engine.StyleDefault
		if i == es.SelectedSol {
			style = engine.StyleHighlight
		}
		line := fmt.Sprintf("%-12s %s  HP:%d/%d", s.Name, s.Rank, s.HP, s.MaxHP)
		ctx.DrawString(2, 3+i, line, style)
	}

	s := es.Base.Soldiers[es.SelectedSol]

	ctx.DrawString(rightX, 2, "EQUIPMENT:", engine.StyleCyanBold)

	weaponLabel := "Weapon:"
	armorLabel := "Armor:"
	weaponStyle := engine.StyleDefault
	armorStyle := engine.StyleDefault
	if es.SelectedSlot == 0 {
		weaponStyle = engine.StyleHighlight
	} else {
		armorStyle = engine.StyleHighlight
	}

	wName := "---"
	if s.Weapon != "" {
		if w, ok := data.Weapons[s.Weapon]; ok {
			wName = fmt.Sprintf("%s (DMG:%d ACC:%d TU:%d)", w.Name, w.Damage, w.Accuracy, w.TU)
		}
	}
	aName := "---"
	if s.Armor != "" {
		if a, ok := data.Armors[s.Armor]; ok {
			aName = fmt.Sprintf("%s (DEF:%d)", a.Name, a.Undersuit)
		}
	}

	ctx.DrawString(rightX, 3, weaponLabel, weaponStyle)
	ctx.DrawString(rightX+8, 3, wName, weaponStyle)
	ctx.DrawString(rightX, 4, armorLabel, armorStyle)
	ctx.DrawString(rightX+8, 4, aName, armorStyle)

	ctx.DrawString(rightX, 6, "AVAILABLE IN STORES:", engine.StyleCyanBold)

	available := es.getAvailableItems()
	y := 7
	for i, item := range available {
		if y >= h-4 {
			break
		}
		style := engine.StyleDefault
		if i == es.CycleIdx {
			style = engine.StyleHighlight
		}
		qty := es.Base.CountItem(item)
		var info string
		if w, ok := data.Weapons[item]; ok {
			info = fmt.Sprintf("%-14s x%d  DMG:%d ACC:%d", w.Name, qty, w.Damage, w.Accuracy)
		} else if a, ok := data.Armors[item]; ok {
			info = fmt.Sprintf("%-14s x%d  DEF:%d", a.Name, qty, a.Undersuit)
		} else {
			info = fmt.Sprintf("%-14s x%d", item, qty)
		}
		ctx.DrawString(rightX, y, info, style)
		y++
	}

	if len(available) == 0 {
		ctx.DrawString(rightX, 7, "No items in stores.", engine.StyleGray)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := "j/k=Select  1=Weapon  2=Armor  Space=Equip  Esc=Back"
	if len(available) > 0 {
		help = "j/k=Soldier  Tab=Cycle  1=Wpn  2=Arm  Space=Equip  Esc=Back"
	}
	ctx.DrawString(1, h-1, help, engine.StyleGray)

	if es.Message != "" {
		ctx.DrawString(2, h-2, es.Message, engine.StyleYellow)
	}
}

func (es *EquipScreen) getAvailableItems() []string {
	var items []string
	if es.SelectedSlot == 0 {
		for k := range data.Weapons {
			if es.Base.CountItem(k) > 0 {
				items = append(items, k)
			}
		}
	} else {
		for k := range data.Armors {
			if k == "none" {
				continue
			}
			if es.Base.CountItem(k) > 0 {
				items = append(items, k)
			}
		}
	}
	sort.Strings(items)
	return items
}

func (es *EquipScreen) equipSelected() {
	available := es.getAvailableItems()
	if len(available) == 0 {
		es.Message = "No items available!"
		return
	}
	if es.CycleIdx >= len(available) {
		es.CycleIdx = 0
	}
	item := available[es.CycleIdx]

	if es.SelectedSlot == 0 {
		if es.Base.EquipWeapon(es.SelectedSol, item) {
			if w, ok := data.Weapons[item]; ok {
				es.Message = fmt.Sprintf("Equipped %s.", w.Name)
			} else {
				es.Message = "Equipped."
			}
		} else {
			es.Message = "Cannot equip!"
		}
	} else {
		if es.Base.EquipArmor(es.SelectedSol, item) {
			if a, ok := data.Armors[item]; ok {
				es.Message = fmt.Sprintf("Equipped %s.", a.Name)
			} else {
				es.Message = "Equipped."
			}
		} else {
			es.Message = "Cannot equip!"
		}
	}
}

func (es *EquipScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		es.SelectedSol--
		if es.SelectedSol < 0 {
			es.SelectedSol = len(es.Base.Soldiers) - 1
		}
		es.CycleIdx = 0
	case tcell.KeyDown:
		es.SelectedSol++
		if es.SelectedSol >= len(es.Base.Soldiers) {
			es.SelectedSol = 0
		}
		es.CycleIdx = 0
	case tcell.KeyTab:
		available := es.getAvailableItems()
		if len(available) > 0 {
			es.CycleIdx++
			if es.CycleIdx >= len(available) {
				es.CycleIdx = 0
			}
		}
	case tcell.KeyRune:
		switch e.Rune() {
		case 'j':
			es.SelectedSol++
			if es.SelectedSol >= len(es.Base.Soldiers) {
				es.SelectedSol = 0
			}
			es.CycleIdx = 0
		case 'k':
			es.SelectedSol--
			if es.SelectedSol < 0 {
				es.SelectedSol = len(es.Base.Soldiers) - 1
			}
			es.CycleIdx = 0
		case '1':
			es.SelectedSlot = 0
			es.CycleIdx = 0
		case '2':
			es.SelectedSlot = 1
			es.CycleIdx = 0
		case ' ':
			es.equipSelected()
		}
	}
}

func (es *EquipScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := es.Game.ScreenSize()

	if y >= 3 && y < 3+len(es.Base.Soldiers) {
		es.SelectedSol = y - 3
		es.CycleIdx = 0
	}

	if y >= 7 && y < h-2 {
		available := es.getAvailableItems()
		idx := y - 7
		if idx < len(available) {
			es.CycleIdx = idx
		}
	}

	if x > w/2 && y == 3 {
		es.SelectedSlot = 0
	}
	if x > w/2 && y == 4 {
		es.SelectedSlot = 1
	}
}
