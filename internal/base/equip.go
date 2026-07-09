package base

import (
	"fmt"
	"sort"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
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
	ctx.DrawPanel(0, 0, w, h, language.String("EQUIP_TITLE"), engine.StyleDefault)

	if len(es.Base.Soldiers) == 0 {
		ctx.DrawString(2, 3, language.String("NO_SOLDIERS"), engine.StyleGray)
		ctx.DrawString(2, 5, language.String("PRESS_ESC"), engine.StyleGray)
		return
	}

	if es.SelectedSol >= len(es.Base.Soldiers) {
		es.SelectedSol = len(es.Base.Soldiers) - 1
	}

	rightX := w / 2

	ctx.DrawString(2, 2, language.String("SECTION_SOLDIER"), engine.StyleCyanBold)
	for i, s := range es.Base.Soldiers {
		style := engine.StyleDefault
		if i == es.SelectedSol {
			style = engine.StyleHighlight
		}
		line := fmt.Sprintf(language.String("EQUIP_SOLDIER_LINE"), s.Name, s.Rank, s.HP, s.MaxHP)
		ctx.DrawString(2, 3+i, line, style)
	}

	s := es.Base.Soldiers[es.SelectedSol]

	ctx.DrawString(rightX, 2, language.String("SECTION_EQUIPMENT"), engine.StyleCyanBold)

	weaponLabel := language.String("LABEL_WEAPON")
	armorLabel := language.String("LABEL_ARMOR")
	weaponStyle := engine.StyleDefault
	armorStyle := engine.StyleDefault
	if es.SelectedSlot == 0 {
		weaponStyle = engine.StyleHighlight
	} else {
		armorStyle = engine.StyleHighlight
	}

	wName := "---"
	if s.Weapon != "" {
		if w, ok := data.RuleItems[s.Weapon]; ok {
			wName = fmt.Sprintf(language.String("EQUIP_WEAPON_INFO"), w.Name, w.Damage, w.Accuracy, w.TU)
		}
	}
	aName := "---"
	if s.Armor != "" {
		if a, ok := data.Armors[s.Armor]; ok {
			aName = fmt.Sprintf(language.String("EQUIP_ARMOR_INFO"), a.Name, a.Undersuit)
		}
	}

	ctx.DrawString(rightX, 3, weaponLabel, weaponStyle)
	ctx.DrawString(rightX+8, 3, wName, weaponStyle)
	ctx.DrawString(rightX, 4, armorLabel, armorStyle)
	ctx.DrawString(rightX+8, 4, aName, armorStyle)

	ctx.DrawString(rightX, 6, language.String("SECTION_AVAILABLE"), engine.StyleCyanBold)

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
		if w, ok := data.RuleItems[item]; ok {
			info = fmt.Sprintf(language.String("EQUIP_ITEM_WEAPON"), w.Name, qty, w.Damage, w.Accuracy)
		} else if a, ok := data.Armors[item]; ok {
			info = fmt.Sprintf(language.String("EQUIP_ITEM_ARMOR"), a.Name, qty, a.Undersuit)
		} else {
			info = fmt.Sprintf(language.String("EQUIP_ITEM_GENERIC"), item, qty)
		}
		ctx.DrawString(rightX, y, info, style)
		y++
	}

	if len(available) == 0 {
		ctx.DrawString(rightX, 7, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_EQUIP")
	if len(available) > 0 {
		help = language.String("HELP_EQUIP_TAB")
	}
	ctx.DrawString(1, h-1, help, engine.StyleGray)

	if es.Message != "" {
		ctx.DrawString(2, h-2, es.Message, engine.StyleYellow)
	}
}

func (es *EquipScreen) getAvailableItems() []string {
	var items []string
	if es.SelectedSlot == 0 {
		for k := range data.RuleItems {
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
		es.Message = language.String("MSG_NO_ITEMS")
		return
	}
	if es.CycleIdx >= len(available) {
		es.CycleIdx = 0
	}
	item := available[es.CycleIdx]

	if es.SelectedSlot == 0 {
		if es.Base.EquipWeapon(es.SelectedSol, item) {
			if w, ok := data.RuleItems[item]; ok {
				es.Message = fmt.Sprintf(language.String("MSG_EQUIPPED"), w.Name)
			} else {
				es.Message = language.String("MSG_EQUIPPED_DONE")
			}
		} else {
			es.Message = language.String("MSG_CANNOT_EQUIP")
		}
	} else {
		if es.Base.EquipArmor(es.SelectedSol, item) {
			if a, ok := data.Armors[item]; ok {
				es.Message = fmt.Sprintf(language.String("MSG_EQUIPPED"), a.Name)
			} else {
				es.Message = language.String("MSG_EQUIPPED_DONE")
			}
		} else {
			es.Message = language.String("MSG_CANNOT_EQUIP")
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
	}
	switch e.Str() {
	case "1":
		es.SelectedSlot = 0
		es.CycleIdx = 0
	case "2":
		es.SelectedSlot = 1
		es.CycleIdx = 0
	case " ":
		es.equipSelected()
	}
}

func (es *EquipScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := es.Game.ScreenSize()

	// Handle help bar clicks (bottom bar)
	if y == h-1 {
		// Help bar: "j/k=Select  1=Weapon  2=Armor  Space=Equip  [Esc]=Back"
		switch {
		case x >= 1 && x <= 3: // j/k=Select
			// Scroll down
			if es.CycleIdx < len(es.getAvailableItems())-1 {
				es.CycleIdx++
			}
		case x >= 5 && x <= 11: // 1=Weapon
			es.SelectedSlot = 0
		case x >= 13 && x <= 19: // 2=Armor
			es.SelectedSlot = 1
		case x >= 21 && x <= 29: // Space=Equip
			es.equipSelected()
		case x >= 31 && x <= 37: // [Esc]=Back
			es.Game.PopState()
		}
		return
	}

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
