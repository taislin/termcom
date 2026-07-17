package geo

import (
	"fmt"
	"sort"

	"github.com/taislin/termcom/internal/base"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

const (
	panelMarginX  = 2
	titleY        = 1
	fromLabelY    = 2
	messageY      = 3
	soldierListY  = 4
	listStartY   = 5
	panelBottomOffset = 2
	helpBarYOffset    = 1
)

type TransferScreen struct {
	Game       *engine.Game
	Geo        *Geoscape
	FromIdx    int
	ToIdx      int
	SelSoldier int
	SelItem    int
	Tab        int // 0 = soldiers, 1 = items
	Message    string
}

func (gs *Geoscape) NewTransferScreen() *TransferScreen {
	si := gs.ActiveBase
	di := gs.ActiveBase
	if len(gs.Bases) > 1 {
		di = (gs.ActiveBase + 1) % len(gs.Bases)
	}
	return &TransferScreen{
		Game:    gs.Game,
		Geo:     gs,
		FromIdx: si,
		ToIdx:   di,
	}
}

func (ts *TransferScreen) Update() {}

func (ts *TransferScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h-panelBottomOffset, language.String("TRANSFER_TITLE"), engine.StyleDefault)

	from := ts.Geo.Bases[ts.FromIdx]
	to := ts.Geo.Bases[ts.ToIdx]

	ctx.DrawString(panelMarginX, titleY, fmt.Sprintf(language.String("TRANSFER_FROM"), from.Name), engine.StyleCyanBold)
	ctx.DrawString(panelMarginX, fromLabelY, fmt.Sprintf(language.String("TRANSFER_TO"), to.Name), engine.StyleGreen)

	if ts.Tab == 0 {
		ts.drawSoldierList(ctx, from, h)
	} else {
		ts.drawItemList(ctx, from, h)
	}

	if ts.Message != "" {
		ctx.DrawString(panelMarginX, h-messageY, ts.Message, engine.StyleYellow)
	}
	help := language.String("HELP_TRANSFER")
	ctx.DrawMarkupString(panelMarginX, h-helpBarYOffset, help, engine.StyleGray, engine.StyleHotkey)
}

func (ts *TransferScreen) drawSoldierList(ctx *engine.ScreenCtx, from *base.Base, h int) {
	ctx.DrawString(panelMarginX, soldierListY, language.String("TRANSFER_SOLDIERS"), engine.StyleYellow)
	if len(from.Soldiers) == 0 {
		ctx.DrawString(listStartY+1, listStartY, language.String("SECTION_NO_SOLDIERS"), engine.StyleGray)
	}
	for i, s := range from.Soldiers {
		if listStartY+i >= h-panelBottomOffset {
			break
		}
		style := engine.StyleDefault
		if i == ts.SelSoldier {
			style = engine.StyleHighlight
		}
		line := fmt.Sprintf(language.String("TRANSFER_SOLDIER_LINE"), s.Name, s.Rank, s.HP)
		ctx.DrawString(listStartY+1, listStartY+i, line, style)
	}
}

func (ts *TransferScreen) drawItemList(ctx *engine.ScreenCtx, from *base.Base, h int) {
	ctx.DrawString(panelMarginX, soldierListY, language.String("TRANSFER_ITEMS"), engine.StyleYellow)
	items := sortedStoreItems(from)
	if len(items) == 0 {
		ctx.DrawString(listStartY+1, listStartY, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
	}
	for i, item := range items {
		if listStartY+i >= h-panelBottomOffset {
			break
		}
		style := engine.StyleDefault
		if i == ts.SelItem {
			style = engine.StyleHighlight
		}
		qty := from.CountItem(item)
		line := fmt.Sprintf(language.String("TRANSFER_ITEM_LINE"), item, qty)
		ctx.DrawString(listStartY+1, listStartY+i, line, style)
	}
}

func (ts *TransferScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyEscape:
		ts.Game.PopState()
		return
	case tcell.KeyTab:
		ts.cycleDest()
		return
	case tcell.KeyUp:
		ts.moveSel(-1)
	case tcell.KeyDown:
		ts.moveSel(1)
	}
	switch e.Str() {
	case "t", "T":
		ts.cycleDest()
	case " ":
		ts.transferSoldier()
	case "\n", "\r", "e", "E":
		ts.transferItem()
	}
}

func (ts *TransferScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := ts.Game.ScreenSize()

	if y == h-1 {
		ts.clickHelpBar(x)
		return
	}

	// Click on item/soldier list area
	if y >= soldierListY && buttons&tcell.Button1 != 0 {
		ts.moveSel(y - soldierListY)
	}
}

func (ts *TransferScreen) clickHelpBar(x int) {
	help := language.String("HELP_TRANSFER")
	col := 2
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
			key := string(runes[i+1 : end])
			ts.dispatchHelpKey(key)
			return
		}
		col = segEnd
		i = end + 1
	}
}

func (ts *TransferScreen) dispatchHelpKey(key string) {
	switch key {
	case "\u2191":
		ts.moveSel(-1)
	case "\u2193":
		ts.moveSel(1)
	case "Tab":
		ts.cycleDest()
	case "Space":
		ts.transferSoldier()
	case "Enter":
		ts.transferItem()
	case "Esc":
		ts.Game.PopState()
	}
}

func (ts *TransferScreen) cycleDest() {
	if len(ts.Geo.Bases) < 2 {
		return
	}
	ts.ToIdx = (ts.ToIdx + 1) % len(ts.Geo.Bases)
	if ts.ToIdx == ts.FromIdx {
		ts.ToIdx = (ts.ToIdx + 1) % len(ts.Geo.Bases)
	}
}

func (ts *TransferScreen) moveSel(d int) {
	from := ts.Geo.Bases[ts.FromIdx]
	if ts.Tab == 0 {
		ts.SelSoldier += d
		if ts.SelSoldier < 0 {
			ts.SelSoldier = 0
		}
		if ts.SelSoldier >= len(from.Soldiers) {
			ts.SelSoldier = len(from.Soldiers) - 1
		}
	} else {
		items := sortedStoreItems(from)
		ts.SelItem += d
		if ts.SelItem < 0 {
			ts.SelItem = 0
		}
		if ts.SelItem >= len(items) {
			ts.SelItem = len(items) - 1
		}
	}
}

func (ts *TransferScreen) transferSoldier() {
	from := ts.Geo.Bases[ts.FromIdx]
	to := ts.Geo.Bases[ts.ToIdx]
	if ts.SelSoldier < 0 || ts.SelSoldier >= len(from.Soldiers) {
		return
	}
	s := from.Soldiers[ts.SelSoldier]
	from.Soldiers = append(from.Soldiers[:ts.SelSoldier], from.Soldiers[ts.SelSoldier+1:]...)
	to.Soldiers = append(to.Soldiers, s)
	if ts.SelSoldier >= len(from.Soldiers) {
		ts.SelSoldier = len(from.Soldiers) - 1
	}
	if ts.SelSoldier < 0 {
		ts.SelSoldier = 0
	}
	ts.Message = fmt.Sprintf(language.String("MSG_TRANSFER_SOLDIER"), s.Name, to.Name)
}

func (ts *TransferScreen) transferItem() {
	from := ts.Geo.Bases[ts.FromIdx]
	to := ts.Geo.Bases[ts.ToIdx]
	items := sortedStoreItems(from)
	if ts.SelItem < 0 || ts.SelItem >= len(items) {
		return
	}
	item := items[ts.SelItem]
	if from.CountItem(item) <= 0 {
		return
	}
	from.RemoveItem(item, 1)
	to.AddItem(item, 1)
	ts.Message = fmt.Sprintf(language.String("MSG_TRANSFER_ITEM"), 1, item, to.Name)
}

func sortedStoreItems(b *base.Base) []string {
	items := make([]string, 0, len(b.Stores))
	for k, v := range b.Stores {
		if v > 0 {
			items = append(items, k)
		}
	}
	sort.Strings(items)
	return items
}
