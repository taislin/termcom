package base

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/soldier"
)

// LoadoutScreen is a pre-mission loadout screen that lets the player select
// which healthy soldiers to deploy and adjust their equipment before launch.
type LoadoutScreen struct {
	Game    *engine.Game
	Base    *Base
	Mission string // mission type label for the header
	NodeID  string // city/node name for the header
	OnLaunch func(selectedSoldiers []*soldier.Soldier)

	Selected []bool // parallel to Base.Soldiers (only healthy are toggleable)
	Cursor   int    // soldier list cursor
	Slot     int    // 0=none, 1=weapon, 2=armor, 3=backpack
	CycleIdx int
	Message  string
}

// NewLoadoutScreen creates a loadout screen for the given base and mission.
func NewLoadoutScreen(g *engine.Game, b *Base, missionType, nodeID string, onLaunch func([]*soldier.Soldier)) *LoadoutScreen {
	if onLaunch == nil {
		onLaunch = func([]*soldier.Soldier) {}
	}
	selected := make([]bool, len(b.Soldiers))
	for i, s := range b.Soldiers {
		selected[i] = s.CanDeploy()
	}
	return &LoadoutScreen{
		Game:     g,
		Base:     b,
		Mission:  missionType,
		NodeID:   nodeID,
		OnLaunch: onLaunch,
		Selected: selected,
	}
}

func (ls *LoadoutScreen) Update() {}

func (ls *LoadoutScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	title := fmt.Sprintf(language.String("LOADOUT_TITLE"), ls.Mission, ls.NodeID)
	ctx.DrawPanel(0, 0, w, h, title, engine.StyleDefault)

	if len(ls.Base.Soldiers) == 0 {
		ctx.DrawString(2, 3, language.String("NO_SOLDIERS"), engine.StyleGray)
		ctx.DrawString(2, 5, language.String("PRESS_ESC"), engine.StyleGray)
		return
	}

	// Clamp cursor
	if ls.Cursor < 0 {
		ls.Cursor = 0
	}
	if ls.Cursor >= len(ls.Base.Soldiers) {
		ls.Cursor = len(ls.Base.Soldiers) - 1
	}

	// Left panel: soldier list with checkboxes
	rightX := engine.Layout.EquipSplitX(w)
	ctx.DrawString(2, 2, language.String("SECTION_SOLDIER"), engine.StyleCyanBold)

	maxRows := h - 6
	if maxRows < 1 {
		maxRows = 1
	}
	scrollOff := 0
	if len(ls.Base.Soldiers) > maxRows {
		scrollOff = ls.Cursor - maxRows/2
		if scrollOff < 0 {
			scrollOff = 0
		}
		if scrollOff > len(ls.Base.Soldiers)-maxRows {
			scrollOff = len(ls.Base.Soldiers) - maxRows
		}
	}

	for row := 0; row < maxRows && scrollOff+row < len(ls.Base.Soldiers); row++ {
		i := scrollOff + row
		s := ls.Base.Soldiers[i]
		mark := " "
		style := engine.StyleGray
		if s.CanDeploy() {
			if ls.Selected[i] {
				mark = "☑"
				style = engine.StyleDefault
			} else {
				mark = "☐"
				style = engine.StyleYellow
			}
		}
		if i == ls.Cursor {
			style = engine.StyleHighlight
		}
		line := fmt.Sprintf("%s %-12s %s  HP:%d/%d", mark, s.Name, s.Rank.String(), s.HP, s.MaxHP)
		ctx.DrawString(2, 3+row, line, style)
	}

	// Right panel: equipment of selected soldier
	ctx.DrawString(rightX, 2, language.String("SECTION_EQUIPMENT"), engine.StyleCyanBold)
	s := ls.Base.Soldiers[ls.Cursor]

	if s.CanDeploy() {
		// Weapon
		wName := "---"
		if s.Weapon != "" {
			if w, ok := data.RuleItems[s.Weapon]; ok {
				wName = fmt.Sprintf(language.String("EQUIP_WEAPON_INFO"), w.DisplayName(), w.Damage, w.Accuracy, w.TU)
			}
		}
		ctx.DrawString(rightX, 3, language.String("LABEL_WEAPON"), engine.StyleDefault)
		wStyle := engine.StyleDefault
		if ls.Slot == 1 {
			wStyle = engine.StyleHighlight
		}
		ctx.DrawString(rightX+8, 3, wName, wStyle)

		// Armor
		aName := "---"
		if s.Armor != "" {
			if a, ok := data.Armors[s.Armor]; ok {
				aName = fmt.Sprintf(language.String("EQUIP_ARMOR_INFO"), a.DisplayNameByKey(s.Armor), a.Undersuit)
			}
		}
		ctx.DrawString(rightX, 4, language.String("LABEL_ARMOR"), engine.StyleDefault)
		aStyle := engine.StyleDefault
		if ls.Slot == 2 {
			aStyle = engine.StyleHighlight
		}
		ctx.DrawString(rightX+8, 4, aName, aStyle)

		// Backpack
		ctx.DrawString(rightX, 5, language.String("LABEL_BACKPACK"), engine.StyleDefault)
		bpStyle := engine.StyleDefault
		if ls.Slot == 3 {
			bpStyle = engine.StyleHighlight
		}
		ctx.DrawString(rightX+8, 5, fmt.Sprintf("%d items", len(s.Inventory)), bpStyle)

		// Encumbrance
		enc := s.Encumbrance()
		limit := s.WeightLimit()
		pen := s.TotalTUPenalty()
		ctx.DrawString(rightX, 7, fmt.Sprintf("Weight: %d/%d  TU -%d", enc, limit, pen), engine.StyleYellow)
		if enc > limit {
			ctx.DrawString(rightX, 8, "OVER-ENCUMBERED!", engine.StyleRed)
		}

		// Available items for swapping
		available := ls.getAvailableItems()
		listY := 10
		if listY < h-6 {
			ctx.DrawString(rightX, listY-1, language.String("SECTION_AVAILABLE"), engine.StyleCyanBold)
			for i, item := range available {
				if listY >= h-4 {
					break
				}
				style := engine.StyleDefault
				if i == ls.CycleIdx {
					style = engine.StyleHighlight
				}
				qty := ls.Base.CountItem(item)
				var info string
				if w, ok := data.RuleItems[item]; ok {
					info = fmt.Sprintf(language.String("EQUIP_ITEM_WEAPON"), w.DisplayName(), qty, w.Damage, w.Accuracy)
				} else if a, ok := data.Armors[item]; ok {
					info = fmt.Sprintf(language.String("EQUIP_ITEM_ARMOR"), a.DisplayNameByKey(item), qty, a.Undersuit)
				} else {
					info = fmt.Sprintf(language.String("EQUIP_ITEM_GENERIC"), data.ItemDisplayName(item), qty)
				}
				ctx.DrawString(rightX, listY, info, style)
				listY++
			}
			if len(available) == 0 {
				ctx.DrawString(rightX, listY, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
			}
		}
	} else {
		ctx.DrawString(rightX, 3, "Wounded - cannot deploy", engine.StyleRed)
	}

	// Summary bar
	selCount := 0
	for _, v := range ls.Selected {
		if v {
			selCount++
		}
	}
	totalDeployable := 0
	for _, s := range ls.Base.Soldiers {
		if s.CanDeploy() {
			totalDeployable++
		}
	}
	summary := fmt.Sprintf(language.String("LOADOUT_SUMMARY"), selCount, totalDeployable)
	ctx.DrawString(2, h-3, summary, engine.StyleYellow)
	ctx.DrawString(2, h-2, language.String("LOADOUT_LAUNCH_HELP"), engine.StyleGray)

	// Bottom help bar
	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_LOADOUT")
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)

	if ls.Message != "" {
		ctx.DrawString(2, h-4, ls.Message, engine.StyleYellow)
	}
}

func (ls *LoadoutScreen) getAvailableItems() []string {
	if ls.Slot == 3 {
		return ls.getAvailableConsumables()
	}
	items := ls.getAvailableWeapons()
	if ls.Slot == 2 {
		items = ls.getAvailableArmors()
	}
	return items
}

func (ls *LoadoutScreen) getAvailableWeapons() []string {
	var items []string
	for k := range data.RuleItems {
		if ls.Base.CountItem(k) > 0 {
			items = append(items, k)
		}
	}
	for k := range ls.Base.CustomWeapons {
		if ls.Base.CountItem(k) > 0 {
			items = append(items, k)
		}
	}
	sortStrings(items)
	return items
}

func (ls *LoadoutScreen) getAvailableArmors() []string {
	var items []string
	for k := range data.Armors {
		if k == "none" {
			continue
		}
		if ls.Base.CountItem(k) > 0 {
			items = append(items, k)
		}
	}
	sortStrings(items)
	return items
}

func (ls *LoadoutScreen) getAvailableConsumables() []string {
	var items []string
	for k, ri := range data.RuleItems {
		if ri.MaxCarry > 0 && ls.Base.CountItem(k) > 0 {
			items = append(items, k)
		}
	}
	sortStrings(items)
	return items
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func (ls *LoadoutScreen) equipCurrent() {
	if ls.Cursor < 0 || ls.Cursor >= len(ls.Base.Soldiers) {
		return
	}
	s := ls.Base.Soldiers[ls.Cursor]
	if !s.CanDeploy() {
		ls.Message = language.String("MSG_CANNOT_EQUIP")
		return
	}
	if ls.Slot == 0 {
		return
	}
	available := ls.getAvailableItems()
	if len(available) == 0 {
		ls.Message = language.String("MSG_NO_ITEMS")
		return
	}
	if ls.CycleIdx >= len(available) {
		ls.CycleIdx = 0
	}
	item := available[ls.CycleIdx]

	if ls.Slot == 3 {
		ri, ok := data.RuleItems[item]
		if !ok {
			return
		}
		maxCarry := ri.MaxCarry
		if maxCarry <= 0 {
			maxCarry = 99
		}
		if s.CountItem(item) >= maxCarry {
			ls.Message = fmt.Sprintf("Max %d %s per soldier", maxCarry, ri.Name)
			return
		}
		if ls.Base.CountItem(item) <= 0 {
			ls.Message = language.String("MSG_NO_ITEMS")
			return
		}
		ls.Base.RemoveItem(item, 1)
		s.AddItem(item)
		if ri.Name != "" {
			ls.Message = fmt.Sprintf("+1 %s", ri.Name)
		} else {
			ls.Message = language.String("MSG_EQUIPPED_DONE")
		}
		return
	}

	if ls.Slot == 1 {
		if ls.Base.EquipWeapon(ls.Cursor, item) {
			if w, ok := data.RuleItems[item]; ok {
				ls.Message = fmt.Sprintf(language.String("MSG_EQUIPPED"), w.Name)
			} else {
				ls.Message = language.String("MSG_EQUIPPED_DONE")
			}
			s.WeaponAmmo = data.RuleItems[item].AmmoMax
		} else {
			ls.Message = language.String("MSG_CANNOT_EQUIP")
		}
	} else if ls.Slot == 2 {
		if ls.Base.EquipArmor(ls.Cursor, item) {
			if a, ok := data.Armors[item]; ok {
				ls.Message = fmt.Sprintf(language.String("MSG_EQUIPPED"), a.Name)
			} else {
				ls.Message = language.String("MSG_EQUIPPED_DONE")
			}
		} else {
			ls.Message = language.String("MSG_CANNOT_EQUIP")
		}
	}
}

func (ls *LoadoutScreen) removeCurrentItem() {
	s := ls.Base.Soldiers[ls.Cursor]
	if !s.CanDeploy() {
		return
	}
	switch ls.Slot {
	case 1:
		if s.Weapon == "" {
			return
		}
		ls.Base.AddItem(s.Weapon, 1)
		ls.Message = fmt.Sprintf(language.String("MSG_REMOVED"), data.ItemDisplayName(s.Weapon))
		s.Weapon = ""
		s.WeaponAmmo = 0
	case 2:
		if s.Armor == "" {
			return
		}
		ls.Base.AddItem(s.Armor, 1)
		ls.Message = fmt.Sprintf(language.String("MSG_REMOVED"), data.ItemDisplayName(s.Armor))
		s.Armor = ""
	case 3:
		available := ls.getAvailableItems()
		if ls.CycleIdx < len(available) {
			item := available[ls.CycleIdx]
			if s.CountItem(item) <= 0 {
				return
			}
			s.RemoveItem(item)
			ls.Base.AddItem(item, 1)
			if ri, ok := data.RuleItems[item]; ok {
				ls.Message = fmt.Sprintf("-1 %s", ri.Name)
			} else {
				ls.Message = language.String("MSG_EQUIPPED_DONE")
			}
		}
	}
}

func (ls *LoadoutScreen) launch() {
	var squad []*soldier.Soldier
	for i, s := range ls.Base.Soldiers {
		if ls.Selected[i] && s.CanDeploy() {
			squad = append(squad, s)
		}
	}
	if len(squad) == 0 {
		ls.Message = language.String("MSG_NO_SOLDIERS_SELECTED")
		return
	}
	if ls.OnLaunch != nil {
		ls.OnLaunch(squad)
	}
}

func (ls *LoadoutScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		ls.Cursor--
		if ls.Cursor < 0 {
			ls.Cursor = 0
		}
		ls.CycleIdx = 0
		ls.Message = ""
	case tcell.KeyDown:
		ls.Cursor++
		if ls.Cursor >= len(ls.Base.Soldiers) {
			ls.Cursor = len(ls.Base.Soldiers) - 1
		}
		ls.CycleIdx = 0
		ls.Message = ""
	case tcell.KeyTab:
		available := ls.getAvailableItems()
		if len(available) > 0 {
			ls.CycleIdx++
			if ls.CycleIdx >= len(available) {
				ls.CycleIdx = 0
			}
		}
		ls.Message = ""
	case tcell.KeyBackspace2, tcell.KeyBackspace:
		if ls.Slot > 0 {
			ls.removeCurrentItem()
		}
	case tcell.KeyDelete:
		ls.removeCurrentItem()
	}
	switch e.Str() {
	case "1":
		ls.Slot = 1
		ls.CycleIdx = 0
		ls.Message = ""
	case "2":
		ls.Slot = 2
		ls.CycleIdx = 0
		ls.Message = ""
	case "3":
		ls.Slot = 3
		ls.CycleIdx = 0
		ls.Message = ""
	case " ":
		if ls.Cursor >= 0 && ls.Cursor < len(ls.Base.Soldiers) {
			s := ls.Base.Soldiers[ls.Cursor]
			if s.CanDeploy() {
				ls.Selected[ls.Cursor] = !ls.Selected[ls.Cursor]
				ls.Message = ""
			}
		}
	case "\x1b", "Escape":
		ls.Game.PopState()
	}
	if e.Key() == tcell.KeyEnter || e.Str() == "l" || e.Str() == "L" {
		ls.launch()
	}
	// Space for consumables +/-
	available := ls.getAvailableItems()
	if ls.CycleIdx < len(available) {
		switch e.Str() {
		case "+":
			if ls.Slot == 3 {
				ls.equipCurrent()
			}
		case "-":
			if ls.Slot == 3 {
				ls.removeCurrentItem()
			}
		}
	}
}

func (ls *LoadoutScreen) HandleMouse(e *tcell.EventMouse) {
	// Not implemented for now
}
