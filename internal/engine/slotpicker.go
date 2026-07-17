package engine

import (
	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
)

type SlotPickerMode int

const (
	SlotPickerLoad SlotPickerMode = iota
	SlotPickerSave
)

type SlotInfo struct {
	Slot  int
	Label string
}

type SlotPickerScreen struct {
	Game      *Game
	Mode      SlotPickerMode
	Slots     []SlotInfo
	Selection int
	Message   string
	OnPick    func(slot int)
}

func NewSlotPickerScreen(g *Game, mode SlotPickerMode, slots []SlotInfo, onSelect func(int)) *SlotPickerScreen {
	return &SlotPickerScreen{
		Game:   g,
		Mode:   mode,
		Slots:  slots,
		OnPick: onSelect,
	}
}

const MaxSaveSlots = 10

func (sp *SlotPickerScreen) Update() {}

func (sp *SlotPickerScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()
	title := language.String("SLOT_PICKER_SAVE")
	if sp.Mode == SlotPickerLoad {
		title = language.String("SLOT_PICKER_LOAD")
	}
	ctx.DrawPanel(0, 0, w, h, title, StyleDefault)

	if len(sp.Slots) == 0 && sp.Mode == SlotPickerLoad {
		ctx.DrawString(2, 3, language.String("SLOT_PICKER_NO_SAVES"), StyleGray)
		ctx.DrawMarkupString(2, 5, language.String("SLOT_PICKER_HELP"), StyleGray, StyleHotkey)
		return
	}

	ctx.DrawMarkupString(2, 2, language.String("SLOT_PICKER_HELP"), StyleCyanBold, StyleHotkey)

	startY := 4
	for i, si := range sp.Slots {
		if startY+i >= h-3 {
			break
		}
		style := StyleDefault
		if i == sp.Selection {
			style = StyleHighlight
		}
		ctx.DrawString(2, startY+i, si.Label, style)
	}

	if sp.Mode == SlotPickerSave {
		if sp.Selection >= len(sp.Slots) {
			ctx.DrawString(2, startY+len(sp.Slots), language.String("SLOT_PICKER_NEW"), StyleHighlight)
		}
	}

	if sp.Message != "" {
		ctx.DrawString(2, h-3, sp.Message, StyleYellow)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawMarkupString(1, h-1, language.String("SLOT_PICKER_BAR"), StyleGray, StyleHotkey)
}

func (sp *SlotPickerScreen) maxSelection() int {
	maxSel := len(sp.Slots)
	if sp.Mode == SlotPickerSave {
		maxSel++
	} else if maxSel > 0 {
		maxSel--
	} else {
		maxSel = 0
	}
	return maxSel
}

func (sp *SlotPickerScreen) moveSelection(delta int) {
	sp.Selection += delta
	if sp.Selection < 0 {
		sp.Selection = 0
	}
	if maxSel := sp.maxSelection(); sp.Selection > maxSel {
		sp.Selection = maxSel
	}
}

func (sp *SlotPickerScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		sp.moveSelection(-1)
	case tcell.KeyDown:
		sp.moveSelection(1)
	case tcell.KeyEnter:
		sp.confirm()
	case tcell.KeyEscape:
		sp.Game.PopState()
	}
	switch e.Str() {
	case "q", "Q":
		sp.Game.PopState()
	case "j", "J":
		sp.moveSelection(1)
	case "k", "K":
		sp.moveSelection(-1)
	}
}

func (sp *SlotPickerScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := sp.Game.ScreenSize()

	if y == h-1 {
		switch {
		case x >= 1 && x <= 3:
			sp.Selection--
			if sp.Selection < 0 {
				sp.Selection = 0
			}
		case x >= 5 && x <= 10:
			sp.confirm()
		case x >= 12 && x <= 18:
			sp.Game.PopState()
		}
		return
	}

	startY := 4
	maxY := startY + len(sp.Slots)
	if sp.Mode == SlotPickerSave {
		maxY++ // allow selecting the "new slot" row
	}
	if y >= startY && y < maxY {
		sp.Selection = y - startY
		if buttons&tcell.Button1 != 0 {
			sp.confirm()
		}
	}
}

func (sp *SlotPickerScreen) confirm() {
	if sp.Mode == SlotPickerSave {
		if sp.Selection < len(sp.Slots) {
			slot := sp.Slots[sp.Selection].Slot
			if sp.OnPick != nil {
				sp.OnPick(slot)
			}
		} else {
			newSlot := 1
			for _, s := range sp.Slots {
				if s.Slot >= newSlot {
					newSlot = s.Slot + 1
				}
			}
			if newSlot > MaxSaveSlots {
				sp.Message = language.String("SLOT_PICKER_FULL")
				return
			}
			if sp.OnPick != nil {
				sp.OnPick(newSlot)
			}
		}
	} else {
		if sp.Selection >= 0 && sp.Selection < len(sp.Slots) {
			slot := sp.Slots[sp.Selection].Slot
			if sp.OnPick != nil {
				sp.OnPick(slot)
			}
		}
	}
	sp.Game.PopState()
}

func FormatSlotLabel(slot int, gameTime string, funds int64) string {
	return language.Sprintf("SLOT_FORMAT", slot, gameTime, funds/1000)
}
