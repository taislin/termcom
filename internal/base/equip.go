package base

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type EquipScreen struct {
	Game         *engine.Game
	Base         *Base
	SelectedSol  int
	SelectedSlot int // 0=weapon, 1=armor, 2=backpack
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
	if es.SelectedSol < 0 {
		es.SelectedSol = 0
	}

	rightX := engine.Layout.EquipSplitX(w)

	ctx.DrawString(2, 2, language.String("SECTION_SOLDIER"), engine.StyleCyanBold)

	const (
		portraitW = 20
		portraitH = 24
	)
	framedH := portraitH/2 + 2 // rows occupied by the framed portrait

	// Portrait sits at the bottom-left. Reserve the last two rows for the
	// message + help bars and clamp so it never spills off a short screen.
	portY := h - 2 - framedH
	if portY < 3 {
		portY = 3
	}

	// The soldier roster fills the rows between the section header and the
	// portrait. Scroll it (keeping the selection centred) so a long roster no
	// longer draws over — and visually clips — the portrait below it.
	maxListRows := portY - 3
	if maxListRows < 1 {
		maxListRows = 1
	}
	start := 0
	if len(es.Base.Soldiers) > maxListRows {
		start = es.SelectedSol - maxListRows/2
		if start < 0 {
			start = 0
		}
		if start > len(es.Base.Soldiers)-maxListRows {
			start = len(es.Base.Soldiers) - maxListRows
		}
	}
	for row := 0; row < maxListRows && start+row < len(es.Base.Soldiers); row++ {
		i := start + row
		sd := es.Base.Soldiers[i]
		style := engine.StyleDefault
		if i == es.SelectedSol {
			style = engine.StyleHighlight
		}
		line := fmt.Sprintf(language.String("EQUIP_SOLDIER_LINE"), sd.Name, sd.Rank, sd.HP, sd.MaxHP)
		ctx.DrawString(2, 3+row, line, style)
	}

	s := es.Base.Soldiers[es.SelectedSol]

	soldierImg := engine.MakeSoldierPortrait(s.Name, portraitW, portraitH)
	ctx.DrawPixelImageFramed(2, portY, soldierImg, engine.StyleCyan)

	ctx.DrawString(rightX, 2, language.String("SECTION_EQUIPMENT"), engine.StyleCyanBold)

	weaponLabel := language.String("LABEL_WEAPON")
	armorLabel := language.String("LABEL_ARMOR")
	backpackLabel := language.String("LABEL_BACKPACK")
	weaponStyle := engine.StyleDefault
	armorStyle := engine.StyleDefault
	backpackStyle := engine.StyleDefault
	switch es.SelectedSlot {
	case 0:
		weaponStyle = engine.StyleHighlight
	case 1:
		armorStyle = engine.StyleHighlight
	case 2:
		backpackStyle = engine.StyleHighlight
	}

	wName := "---"
	if s.Weapon != "" {
		if w, ok := data.RuleItems[s.Weapon]; ok {
			wName = fmt.Sprintf(language.String("EQUIP_WEAPON_INFO"), w.DisplayName(), w.Damage, w.Accuracy, w.TU)
		}
	}
	aName := "---"
	if s.Armor != "" {
		if a, ok := data.Armors[s.Armor]; ok {
			aName = fmt.Sprintf(language.String("EQUIP_ARMOR_INFO"), a.DisplayNameByKey(s.Armor), a.Undersuit)
		}
	}

	ctx.DrawString(rightX, 3, weaponLabel, weaponStyle)
	ctx.DrawString(rightX+8, 3, wName, weaponStyle)
	ctx.DrawString(rightX, 4, armorLabel, armorStyle)
	ctx.DrawString(rightX+8, 4, aName, armorStyle)
	ctx.DrawString(rightX, 5, backpackLabel, backpackStyle)
	ctx.DrawString(rightX+8, 5, fmt.Sprintf("%d items", len(s.Inventory)), backpackStyle)

	ctx.DrawString(rightX, 7, language.String("SECTION_AVAILABLE"), engine.StyleCyanBold)

	available := es.getAvailableItems()
	y := 8
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
			info = fmt.Sprintf(language.String("EQUIP_ITEM_WEAPON"), w.DisplayName(), qty, w.Damage, w.Accuracy)
		} else if a, ok := data.Armors[item]; ok {
			info = fmt.Sprintf(language.String("EQUIP_ITEM_ARMOR"), a.DisplayNameByKey(item), qty, a.Undersuit)
		} else {
			info = fmt.Sprintf(language.String("EQUIP_ITEM_GENERIC"), data.ItemDisplayName(item), qty)
		}
		ctx.DrawString(rightX, y, info, style)
		y++
	}

	if len(available) == 0 {
		ctx.DrawString(rightX, 8, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
	}

	// Encumbrance and backpack contents
	encY := y + 2
	enc := s.Encumbrance()
	limit := s.WeightLimit()
	pen := s.TotalTUPenalty()
	ctx.DrawString(rightX, encY, fmt.Sprintf("Weight: %d/%d  TU -%d", enc, limit, pen), engine.StyleYellow)
	if enc > limit {
		ctx.DrawString(rightX, encY+1, "OVER-ENCUMBERED!", engine.StyleRed)
	}

	if len(s.Inventory) > 0 {
		ctx.DrawString(rightX, encY+3, language.String("SECTION_BACKPACK"), engine.StyleCyanBold)
		bpy := encY + 4
		for key, qty := range s.Inventory {
			if bpy >= h-3 {
				break
			}
			ctx.DrawString(rightX, bpy, fmt.Sprintf("%s x%d", data.ItemDisplayName(key), qty), engine.StyleDefault)
			bpy++
		}
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_EQUIP")
	if len(available) > 0 {
		help = language.String("HELP_EQUIP_TAB")
	}
	if es.SelectedSlot == 2 {
		help = language.String("HELP_EQUIP_BACKPACK")
	}
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)

	if es.Message != "" {
		ctx.DrawString(2, h-2, es.Message, engine.StyleYellow)
	}
}

func (es *EquipScreen) getAvailableItems() []string {
	if es.SelectedSlot == 2 {
		return es.getAvailableConsumables()
	}
	var items []string
	if es.SelectedSlot == 0 {
		for k := range data.RuleItems {
			if es.Base.CountItem(k) > 0 {
				items = append(items, k)
			}
		}
		// Also include custom-designed weapons
		for k := range es.Base.CustomWeapons {
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

// getAvailableConsumables returns items that can go in the backpack slot.
// These are items with a MaxCarry > 0 that exist in base stores.
func (es *EquipScreen) getAvailableConsumables() []string {
	var items []string
	for k, ri := range data.RuleItems {
		if ri.MaxCarry > 0 && es.Base.CountItem(k) > 0 {
			items = append(items, k)
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

	if es.SelectedSlot == 2 {
		es.adjustBackpackQty(item, 1)
		return
	}

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

// adjustBackpackQty adds (+1) or removes (-1) a quantity of the given item
// from the selected soldier's backpack, transferring to/from base stores.
func (es *EquipScreen) adjustBackpackQty(item string, delta int) {
	s := es.Base.Soldiers[es.SelectedSol]
	ri, ok := data.RuleItems[item]
	if !ok {
		return
	}
	if delta > 0 {
		maxCarry := ri.MaxCarry
		if maxCarry <= 0 {
			maxCarry = 99
		}
		if s.CountItem(item) >= maxCarry {
			es.Message = fmt.Sprintf("Max %d %s per soldier", maxCarry, ri.Name)
			return
		}
		if es.Base.CountItem(item) <= 0 {
			es.Message = language.String("MSG_NO_ITEMS")
			return
		}
		es.Base.RemoveItem(item, 1)
		s.AddItem(item)
		es.Message = fmt.Sprintf("+1 %s", ri.Name)
	} else {
		if s.CountItem(item) <= 0 {
			es.Message = fmt.Sprintf("No %s to remove", ri.Name)
			return
		}
		s.RemoveItem(item)
		es.Base.AddItem(item, 1)
		es.Message = fmt.Sprintf("-1 %s", ri.Name)
	}
}

// autoEquip scans base stores and equips every soldier with the best available
// weapon (highest damage they have strength for) and best available armor.
func (es *EquipScreen) autoEquip() {
	equipped := 0
	// Collect eligible weapons from stores (firearms/melee with damage > 0)
	type wpnScore struct {
		key    string
		damage int
	}
	var weapons []wpnScore
	for key, item := range data.RuleItems {
		if item.Damage <= 0 || item.IsAmmo {
			continue
		}
		if es.Base.CountItem(key) <= 0 {
			continue
		}
		weapons = append(weapons, wpnScore{key, item.Damage})
	}
	// Sort descending by damage
	sort.Slice(weapons, func(i, j int) bool {
		return weapons[i].damage > weapons[j].damage
	})

	// Collect eligible armors from stores
	type armScore struct {
		key     string
		defense int
	}
	var armors []armScore
	for key, a := range data.Armors {
		if key == "none" {
			continue
		}
		if es.Base.CountItem(key) <= 0 {
			continue
		}
		armors = append(armors, armScore{key, a.Undersuit})
	}
	sort.Slice(armors, func(i, j int) bool {
		return armors[i].defense > armors[j].defense
	})

	for idx, s := range es.Base.Soldiers {
		// Best weapon soldier can use
		bestWpn := ""
		for _, w := range weapons {
			if s.Strength >= data.RuleItems[w.key].Strength {
				bestWpn = w.key
				break
			}
		}
		if bestWpn != "" {
			if es.Base.EquipWeapon(idx, bestWpn) {
				s.WeaponAmmo = data.RuleItems[bestWpn].AmmoMax
			}
		}

		// Best armor (try each until one is available)
		for _, a := range armors {
			if es.Base.EquipArmor(idx, a.key) {
				break
			}
		}
		equipped++
	}

	es.Message = fmt.Sprintf(language.String("MSG_AUTO_EQUIPPED"), equipped)
}

func (es *EquipScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		es.SelectedSol--
		if es.SelectedSol < 0 {
			es.SelectedSol = len(es.Base.Soldiers) - 1
		}
		es.CycleIdx = 0
		es.Message = ""
	case tcell.KeyDown:
		es.SelectedSol++
		if es.SelectedSol >= len(es.Base.Soldiers) {
			es.SelectedSol = 0
		}
		es.CycleIdx = 0
		es.Message = ""
	case tcell.KeyTab:
		available := es.getAvailableItems()
		if len(available) > 0 {
			es.CycleIdx++
			if es.CycleIdx >= len(available) {
				es.CycleIdx = 0
			}
		}
		es.Message = ""
	case tcell.KeyRune:
		switch e.Str() {
		case "+":
			if es.SelectedSlot == 2 {
				available := es.getAvailableItems()
				if es.CycleIdx < len(available) {
					es.adjustBackpackQty(available[es.CycleIdx], 1)
				}
			}
		case "-":
			if es.SelectedSlot == 2 {
				available := es.getAvailableItems()
				if es.CycleIdx < len(available) {
					es.adjustBackpackQty(available[es.CycleIdx], -1)
				}
			}
		}
	}
	switch e.Str() {
	case "1":
		es.SelectedSlot = 0
		es.CycleIdx = 0
		es.Message = ""
	case "2":
		es.SelectedSlot = 1
		es.CycleIdx = 0
		es.Message = ""
	case "3":
		es.SelectedSlot = 2
		es.CycleIdx = 0
		es.Message = ""
	case " ":
		es.equipSelected()
	case "a", "A":
		es.autoEquip()
	}
}

func (es *EquipScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := es.Game.ScreenSize()

	// Handle help bar clicks (bottom bar) by parsing the rendered markup
	// segments so the hit zones stay correct regardless of locale/text width.
	if y == h-1 {
		es.clickEquipHelpBar(x)
		return
	}

	if y >= 3 && y < 3+len(es.Base.Soldiers) {
		es.SelectedSol = y - 3
		es.CycleIdx = 0
	}

	if y >= 8 && y < h-2 {
		available := es.getAvailableItems()
		idx := y - 8
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
	if x > w/2 && y == 5 {
		es.SelectedSlot = 2
	}
}

// clickEquipHelpBar dispatches a click on the bottom help bar by matching the
// x coordinate against the rendered [key] markup segments, so the click zones
// remain aligned even when the help text width varies by language.
func (es *EquipScreen) clickEquipHelpBar(x int) {
	help := language.String("HELP_EQUIP")
	if len(es.getAvailableItems()) > 0 {
		help = language.String("HELP_EQUIP_TAB")
	}
	col := 1
	runes := []rune(help)
	for i := 0; i < len(runes); {
		if runes[i] != '[' {
			col += engine.StringWidth(string(runes[i]))
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
		segEnd := col + engine.StringWidth(string(runes[i:end+1]))
		if x >= segStart && x <= segEnd {
			es.dispatchEquipHelpKey(string(runes[i+1 : end]))
			return
		}
		col = segEnd
		i = end + 1
	}
}

func (es *EquipScreen) dispatchEquipHelpKey(key string) {
	switch key {
	case "↑", "↓":
		if es.CycleIdx < len(es.getAvailableItems())-1 {
			es.CycleIdx++
		}
	case "1":
		es.SelectedSlot = 0
	case "2":
		es.SelectedSlot = 1
	case "Tab":
		if es.CycleIdx < len(es.getAvailableItems())-1 {
			es.CycleIdx++
		}
	case "Space":
		es.equipSelected()
	case "A":
		es.autoEquip()
	case "Esc":
		es.Game.PopState()
	}
}
