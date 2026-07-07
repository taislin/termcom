package base

import (
	"fmt"

	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
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
			bs.Message = fmt.Sprintf(language.String("MSG_BUILDING"), def.Name, def.Cost/1000)
		} else {
			bs.Message = language.String("MSG_INSUFFICIENT_FUNDS")
		}
	}
}

func (bs *BaseScreen) SellFacility() {
	if bs.Tab == 0 && bs.Selection >= 0 && bs.Selection < len(bs.Base.Facilities) {
		fac := bs.Base.Facilities[bs.Selection]
		if fac.Building {
			bs.Message = language.String("MSG_CANNOT_SELL_BUILDING")
			return
		}
		def := FacilityDefs[fac.Type]
		refund := int64(def.Cost) / 2
		bs.Game.Funds += refund
		bs.Base.Facilities = append(bs.Base.Facilities[:bs.Selection], bs.Base.Facilities[bs.Selection+1:]...)
		bs.Message = fmt.Sprintf(language.String("MSG_SOLD"), def.Name, refund/1000)
		if bs.Selection >= len(bs.Base.Facilities) {
			bs.Selection = len(bs.Base.Facilities) - 1
		}
		if bs.Selection < 0 {
			bs.Selection = 0
		}
	}
}

func (bs *BaseScreen) HireSoldier() {
	if bs.Game.Funds < int64(HireCost) {
		bs.Message = language.String("MSG_CANNOT_HIRE")
		return
	}
	ok, msg := bs.Base.HireSoldier()
	if ok {
		bs.Game.Funds -= int64(HireCost)
		bs.Message = msg + fmt.Sprintf(" ($%dK)", HireCost/1000)
	} else {
		bs.Message = msg
	}
}

func (bs *BaseScreen) DismissSoldier() {
	if bs.Tab == 1 && bs.Selection >= 0 && bs.Selection < len(bs.Base.Soldiers) {
		name := bs.Base.Soldiers[bs.Selection].Name
		bs.Base.DismissSoldier(bs.Selection)
		bs.Message = name + language.String("MSG_DISMISSED")
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

	ctx.DrawPanel(0, 0, w, h-2, language.String("BASE_MANAGEMENT"), engine.StyleDefault)

	tabs := []string{language.String("TAB_FACILITIES"), language.String("TAB_SOLDIERS"), language.String("TAB_RESEARCH"), language.String("TAB_MANUFACTURE"), language.String("TAB_TRANSFER")}
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
	soldStr := fmt.Sprintf(language.String("BASE_SOLDIERS"), len(bs.Base.Soldiers), cap)
	ctx.DrawString(2, h-2, fmt.Sprintf(language.String("BASE_PERSONNEL"), bs.Base.Scientists, bs.Base.Engineers, soldStr), engine.StyleDefault)
	if bs.Message != "" {
		ctx.DrawString(w/2, h-2, bs.Message, engine.StyleYellow)
	}
	help := language.String("HELP_BASE")
	if bs.Tab == 0 {
		help = language.String("HELP_FACILITIES")
	} else if bs.Tab == 1 {
		help = language.String("HELP_SOLDIERS")
	} else if bs.Tab == 2 {
		help = language.String("HELP_TAB_RESEARCH")
	} else if bs.Tab == 3 {
		help = language.String("HELP_TAB_MANUFACTURE")
	}
	ctx.DrawString(2, h-1, help, engine.StyleGray)
}

func (bs *BaseScreen) renderFacilities(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_FACILITIES"), engine.StyleCyanBold)
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
				if building == "" {
					building = fmt.Sprintf(language.String("INFO_BUILDING_DAYS"), f.DaysLeft)
				} else {
					building = language.String("INFO_BUILDING")
				}
				break
			}
		}
		line := fmt.Sprintf("%-20s x%d $%dK%s", def.Name, count, def.Cost/1000, building)
		ctx.DrawString(x, y+2+i, line, style)
	}
}

func (bs *BaseScreen) renderSoldiers(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_SOLDIERS"), engine.StyleCyanBold)

	squad := bs.Base.Soldiers
	if len(squad) == 0 {
		ctx.DrawString(x, y+2, language.String("SECTION_NO_SOLDIERS"), engine.StyleGray)
		return
	}

	header := fmt.Sprintf("%-12s %-10s %4s %4s %4s %4s %4s %4s %6s",
		language.String("COL_NAME"), language.String("COL_RANK"), language.String("COL_HP"),
		language.String("COL_TU"), language.String("COL_ACC"), language.String("COL_BRA"),
		language.String("COL_STR"), language.String("COL_KILLS"), language.String("COL_WOUNDS"))
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

	info := fmt.Sprintf(language.String("INFO_HIRE_COST"), HireCost/1000)
	ctx.DrawString(x, y+h-1, info, engine.StyleGray)
}

func (bs *BaseScreen) renderResearch(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_RESEARCH_LABS"), engine.StyleCyanBold)
	labs := bs.Base.TotalLabs()
	if labs == 0 {
		ctx.DrawString(x, y+2, language.String("SECTION_NO_LABS"), engine.StyleGray)
	} else {
		ctx.DrawString(x, y+2, fmt.Sprintf(language.String("SECTION_LABS_INFO"), labs), engine.StyleGray)
	}
	if bs.Base.ActiveResearch != nil && !bs.Base.ActiveResearch.Completed {
		ctx.DrawString(x, y+4, fmt.Sprintf(language.String("INFO_ACTIVE"), bs.Base.ActiveResearch.TopicID), engine.StyleGreen)
	}
}

func (bs *BaseScreen) renderManufacture(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_WORKSHOPS"), engine.StyleCyanBold)
	wrks := bs.Base.TotalWorkshops()
	if wrks == 0 {
		ctx.DrawString(x, y+2, language.String("SECTION_NO_WORKSHOPS"), engine.StyleGray)
	} else {
		ctx.DrawString(x, y+2, fmt.Sprintf(language.String("SECTION_WORKSHOPS_INFO"), wrks), engine.StyleGray)
	}
	if len(bs.Base.ManufactureQueue) > 0 {
		ctx.DrawString(x, y+4, fmt.Sprintf(language.String("INFO_ACTIVE_JOBS"), len(bs.Base.ManufactureQueue)), engine.StyleGreen)
	}
}

func (bs *BaseScreen) renderTransfer(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_STORES"), engine.StyleCyanBold)
	y += 2
	for item, qty := range bs.Base.Stores {
		if qty > 0 && y < h+2 {
			ctx.DrawString(x, y, fmt.Sprintf("%-15s x%d", item, qty), engine.StyleDefault)
			y++
		}
	}
	if y == 2 {
		ctx.DrawString(x, y, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
	}
}

func (bs *BaseScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		bs.Selection--
		if bs.Selection < 0 {
			if bs.Tab == 1 {
				bs.Selection = len(bs.Base.Soldiers) - 1
			} else {
				bs.Selection = 6
			}
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
			if bs.Tab == 1 {
				if bs.Selection >= len(bs.Base.Soldiers) {
					bs.Selection = 0
				}
			} else {
				if bs.Selection > 6 {
					bs.Selection = 0
				}
			}
		case 'k':
			bs.Selection--
			if bs.Selection < 0 {
				if bs.Tab == 1 {
					bs.Selection = len(bs.Base.Soldiers) - 1
				} else {
					bs.Selection = 6
				}
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
		case x >= 1 && x <= 9 && bs.Tab == 0:
			bs.BuildFacility()
		case x >= 11 && x <= 18 && bs.Tab == 1:
			bs.HireSoldier()
		}
	}
}
