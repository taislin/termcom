package base

import (
	"fmt"

	"github.com/civ13/ycom/internal/engine"
	"github.com/gdamore/tcell/v2"
)

type BaseScreen struct {
	Game      *engine.Game
	Base      *Base
	Tab       int
	Selection int
	Message   string
}

func NewBaseScreen(g *engine.Game, b *Base) *BaseScreen {
	return &BaseScreen{
		Game: g,
		Base: b,
		Tab:  0,
	}
}

func (bs *BaseScreen) BuildFacility() {
	types := []FacilityType{FacLivingQuarters, FacLab, FacWorkshop, FacStorage, FacRadar, FacContainment, FacHangar}
	if bs.Selection >= 0 && bs.Selection < len(types) {
		ft := types[bs.Selection]
		def := FacilityDefs[ft]
		if bs.Game.Funds >= int64(def.Cost) {
			bs.Game.Funds -= int64(def.Cost)
			bs.Base.BuildFacility(ft)
			bs.Message = fmt.Sprintf("Building %s ($%dK)", def.Name, def.Cost/1000)
		} else {
			bs.Message = "Insufficient funds!"
		}
	}
}

func (bs *BaseScreen) SellFacility() {
	if bs.Tab == 0 && bs.Selection >= 0 && bs.Selection < len(bs.Base.Facilities) {
		fac := bs.Base.Facilities[bs.Selection]
		if fac.Building {
			bs.Message = "Cannot sell under construction!"
			return
		}
		def := FacilityDefs[fac.Type]
		refund := int64(def.Cost) / 2
		bs.Game.Funds += refund
		bs.Base.Facilities = append(bs.Base.Facilities[:bs.Selection], bs.Base.Facilities[bs.Selection+1:]...)
		bs.Message = fmt.Sprintf("Sold %s for $%dK", def.Name, refund/1000)
		if bs.Selection >= len(bs.Base.Facilities) {
			bs.Selection = len(bs.Base.Facilities) - 1
		}
		if bs.Selection < 0 {
			bs.Selection = 0
		}
	}
}

func (bs *BaseScreen) HireSoldier() {
	ok, msg := bs.Base.HireSoldier()
	if ok {
		if bs.Game.Funds >= int64(HireCost) {
			bs.Game.Funds -= int64(HireCost)
			bs.Message = msg + fmt.Sprintf(" ($%dK)", HireCost/1000)
		} else {
			bs.Base.DismissSoldier(len(bs.Base.Soldiers) - 1)
			bs.Message = "Insufficient funds to hire!"
		}
	} else {
		bs.Message = msg
	}
}

func (bs *BaseScreen) DismissSoldier() {
	if bs.Tab == 1 && bs.Selection >= 0 && bs.Selection < len(bs.Base.Soldiers) {
		name := bs.Base.Soldiers[bs.Selection].Name
		bs.Base.DismissSoldier(bs.Selection)
		bs.Message = name + " dismissed."
		if bs.Selection >= len(bs.Base.Soldiers) {
			bs.Selection = len(bs.Base.Soldiers) - 1
		}
		if bs.Selection < 0 {
			bs.Selection = 0
		}
	}
}

func (bs *BaseScreen) Update() {}

func (bs *BaseScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	ctx.DrawPanel(0, 0, w, h-2, "BASE MANAGEMENT", engine.StyleDefault)

	tabs := []string{"Facilities", "Soldiers", "Research", "Manufacture", "Transfer"}
	for i, t := range tabs {
		style := engine.StyleDefault
		if i == bs.Tab {
			style = engine.StyleHighlight
		}
		ctx.DrawString(2+i*14, 1, fmt.Sprintf("[%s]", t), style)
	}

	contentY := 3
	switch bs.Tab {
	case 0:
		bs.renderFacilities(ctx, 2, contentY, w-4, h-6)
	case 1:
		bs.renderSoldiers(ctx, 2, contentY, w-4, h-6)
	case 2:
		bs.renderResearch(ctx, 2, contentY, w-4, h-6)
	case 3:
		bs.renderManufacture(ctx, 2, contentY, w-4, h-6)
	case 4:
		bs.renderTransfer(ctx, 2, contentY, w-4, h-6)
	}

	ctx.DrawPanel(0, h-2, w, 2, "", engine.StyleDefault)
	cap := bs.Base.LivingCapacity()
	soldStr := fmt.Sprintf("Soldiers: %d/%d", len(bs.Base.Soldiers), cap)
	ctx.DrawString(2, h-2, fmt.Sprintf("Scientists: %d  Engineers: %d  %s", bs.Base.Scientists, bs.Base.Engineers, soldStr), engine.StyleDefault)
	if bs.Message != "" {
		ctx.DrawString(w/2, h-2, bs.Message, engine.StyleYellow)
	}
	help := "[B]uild  [H]ire  1-5=Tab  j/k=Navigate  Esc=Back"
	if bs.Tab == 0 {
		help = "[B]uild  [S]ell  j/k=Navigate  Esc=Back"
	} else if bs.Tab == 1 {
		help = "[H]ire  [E]quip  [D]ismiss  j/k=Navigate  Esc=Back"
	} else if bs.Tab == 2 {
		help = "[R]esearch  j/k=Navigate  Esc=Back"
	} else if bs.Tab == 3 {
		help = "[M]anufacture  j/k=Navigate  Esc=Back"
	}
	ctx.DrawString(2, h-1, help, engine.StyleGray)
}

func (bs *BaseScreen) renderFacilities(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "FACILITIES:", engine.StyleCyanBold)
	facTypes := []FacilityType{FacLivingQuarters, FacLab, FacWorkshop, FacStorage, FacRadar, FacContainment, FacHangar}
	for i, ft := range facTypes {
		count := bs.Base.CountFacility(ft)
		def := FacilityDefs[ft]
		style := engine.StyleDefault
		if i == bs.Selection {
			style = engine.StyleHighlight
		}
		building := ""
		for _, f := range bs.Base.Facilities {
			if f.Type == ft && f.Building {
				building = fmt.Sprintf(" (Building: %d days)", f.DaysLeft)
			}
		}
		line := fmt.Sprintf("%-20s x%d $%dK%s", def.Name, count, def.Cost/1000, building)
		ctx.DrawString(x, y+2+i, line, style)
	}
}

func (bs *BaseScreen) renderSoldiers(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "SOLDIERS:", engine.StyleCyanBold)

	squad := bs.Base.Soldiers
	if len(squad) == 0 {
		ctx.DrawString(x, y+2, "No soldiers. Press [H] to hire.", engine.StyleGray)
		return
	}

	header := fmt.Sprintf("%-12s %-10s %4s %4s %4s %4s %4s %4s %6s",
		"Name", "Rank", "HP", "TU", "ACC", "BRA", "STR", "Kills", "Wounds")
	ctx.DrawString(x, y+1, header, engine.StyleGray)

	for i, s := range squad {
		if y+3+i >= y+h {
			break
		}
		style := engine.StyleDefault
		if i == bs.Selection {
			style = engine.StyleHighlight
		}
		woundsStr := ""
		if s.Wounds > 0 {
			woundsStr = fmt.Sprintf("%dd", s.Wounds)
		}
		line := fmt.Sprintf("%-12s %-10s %4d %4d %4d %4d %4d %4d %6s",
			s.Name, s.Rank, s.HP, s.MaxTU, s.Accuracy, s.Bravery, s.Strength, s.Kills, woundsStr)
		if s.Wounds > 0 {
			ctx.DrawString(x, y+3+i, line, engine.StyleRed)
		} else {
			ctx.DrawString(x, y+3+i, line, style)
		}
	}

	info := fmt.Sprintf("Hire cost: $%dK  |  [H]ire  [D]ismiss", HireCost/1000)
	ctx.DrawString(x, y+h-1, info, engine.StyleGray)
}

func (bs *BaseScreen) renderResearch(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "RESEARCH LABS:", engine.StyleCyanBold)
	labs := bs.Base.TotalLabs()
	if labs == 0 {
		ctx.DrawString(x, y+2, "Build a Laboratory first.", engine.StyleGray)
	} else {
		ctx.DrawString(x, y+2, fmt.Sprintf("%d lab(s) operational. [R] to open.", labs), engine.StyleGray)
	}
	if bs.Base.ActiveResearch != nil && !bs.Base.ActiveResearch.Completed {
		ctx.DrawString(x, y+4, fmt.Sprintf("Active: %s", bs.Base.ActiveResearch.TopicID), engine.StyleGreen)
	}
}

func (bs *BaseScreen) renderManufacture(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "WORKSHOPS:", engine.StyleCyanBold)
	wrks := bs.Base.TotalWorkshops()
	if wrks == 0 {
		ctx.DrawString(x, y+2, "Build a Workshop first.", engine.StyleGray)
	} else {
		ctx.DrawString(x, y+2, fmt.Sprintf("%d workshop(s) operational. [M] to open.", wrks), engine.StyleGray)
	}
	if len(bs.Base.ManufactureQueue) > 0 {
		ctx.DrawString(x, y+4, fmt.Sprintf("Active jobs: %d", len(bs.Base.ManufactureQueue)), engine.StyleGreen)
	}
}

func (bs *BaseScreen) renderTransfer(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "STORES:", engine.StyleCyanBold)
	y += 2
	for item, qty := range bs.Base.Stores {
		if qty > 0 && y < h+2 {
			ctx.DrawString(x, y, fmt.Sprintf("%-15s x%d", item, qty), engine.StyleDefault)
			y++
		}
	}
	if y == 2 {
		ctx.DrawString(x, y, "No items in stores.", engine.StyleGray)
	}
}

func (bs *BaseScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		bs.Selection--
		if bs.Selection < 0 {
			bs.Selection = 6
		}
	case tcell.KeyDown:
		bs.Selection++
		if bs.Tab == 1 {
			if bs.Selection >= len(bs.Base.Soldiers) {
				bs.Selection = 0
			}
		} else {
			if bs.Selection > 6 {
				bs.Selection = 0
			}
		}
	case tcell.KeyLeft:
		bs.Tab--
		if bs.Tab < 0 {
			bs.Tab = 4
		}
		bs.Selection = 0
	case tcell.KeyRight:
		bs.Tab++
		if bs.Tab > 4 {
			bs.Tab = 0
		}
		bs.Selection = 0
	case tcell.KeyRune:
		switch e.Rune() {
		case '1':
			bs.Tab = 0
		case '2':
			bs.Tab = 1
		case '3':
			bs.Tab = 2
		case '4':
			bs.Tab = 3
		case '5':
			bs.Tab = 4
		case 'j':
			bs.Selection++
		case 'k':
			bs.Selection--
			if bs.Selection < 0 {
				bs.Selection = 0
			}
		case 'b', 'B':
			bs.BuildFacility()
		case 's', 'S':
			bs.SellFacility()
		case 'h', 'H':
			bs.HireSoldier()
		case 'd', 'D':
			bs.DismissSoldier()
		case 'e', 'E':
			if bs.Tab == 1 && len(bs.Base.Soldiers) > 0 {
				bs.Game.PushState(engine.StateEquip)
			}
		case 'r', 'R':
			if bs.Tab == 2 {
				bs.Game.PushState(engine.StateResearch)
			}
		case 'm', 'M':
			if bs.Tab == 3 {
				bs.Game.PushState(engine.StateManufacture)
			}
		}
	}
}

func (bs *BaseScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := bs.Game.ScreenSize()

	if y == 1 {
		for i := 0; i < 5; i++ {
			tx := 2 + i*14
			if x >= tx && x <= tx+12 {
				bs.Tab = i
				bs.Selection = 0
				return
			}
		}
	}

	if y >= 5 && y <= 11 && bs.Tab == 0 {
		bs.Selection = y - 5
		return
	}

	if y == h-2 {
		switch {
		case x >= 1 && x <= 9:
			bs.BuildFacility()
		case x >= 11 && x <= 18:
			bs.HireSoldier()
		}
	}
}
