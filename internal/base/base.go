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

func NewBaseScreen(g *engine.Game) *BaseScreen {
	b := NewBase("Base 1")
	b.Facilities = append(b.Facilities, &Facility{Type: FacLivingQuarters, Row: 0, Col: 0})
	b.Facilities = append(b.Facilities, &Facility{Type: FacLab, Row: 0, Col: 1})
	b.Facilities = append(b.Facilities, &Facility{Type: FacWorkshop, Row: 0, Col: 2})
	b.Facilities = append(b.Facilities, &Facility{Type: FacStorage, Row: 0, Col: 3})
	b.Facilities = append(b.Facilities, &Facility{Type: FacRadar, Row: 0, Col: 4})
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

func (bs *BaseScreen) HireSoldier() {
	cap := bs.Base.LivingCapacity()
	if cap > 0 {
		bs.Message = fmt.Sprintf("Hiring soldiers... Capacity: %d", cap)
	} else {
		bs.Message = "Build Living Quarters first!"
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
	ctx.DrawString(2, h-2, fmt.Sprintf("Scientists: %d  Engineers: %d", bs.Base.Scientists, bs.Base.Engineers), engine.StyleDefault)
	if bs.Message != "" {
		ctx.DrawString(w/2, h-2, bs.Message, engine.StyleYellow)
	}
	ctx.DrawString(2, h-1, "[B]uild  [H]ire  1-5=Tab  j/k=Navigate  Esc=Back", engine.StyleGray)
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
	ctx.DrawString(x, y+2, "No soldiers yet.", engine.StyleGray)
}

func (bs *BaseScreen) renderResearch(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "RESEARCH LABS:", engine.StyleCyanBold)
	ctx.DrawString(x, y+2, "No active research.", engine.StyleGray)
}

func (bs *BaseScreen) renderManufacture(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "WORKSHOPS:", engine.StyleCyanBold)
	ctx.DrawString(x, y+2, "No production.", engine.StyleGray)
}

func (bs *BaseScreen) renderTransfer(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, "TRANSFERS:", engine.StyleCyanBold)
	ctx.DrawString(x, y+2, "Single base. No transfers.", engine.StyleGray)
}

func (bs *BaseScreen) HandleKey(e *tcell.EventKey) {
	if e.Key() != tcell.KeyRune {
		return
	}
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
		if bs.Selection > 6 {
			bs.Selection = 0
		}
	case 'k':
		bs.Selection--
		if bs.Selection < 0 {
			bs.Selection = 6
		}
	case 'b', 'B':
		bs.BuildFacility()
	case 'h', 'H':
		bs.HireSoldier()
	}
}
