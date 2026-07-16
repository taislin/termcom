package base

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type BaseScreen struct {
	Game        *engine.Game
	Base        *Base
	Tab         int
	Selection   int
	Message     string
	storesItems []string
}

func NewBaseScreen(g *engine.Game, b *Base) *BaseScreen {
	return &BaseScreen{
		Game: g,
		Base: b,
		Tab:  0,
	}
}

func (bs *BaseScreen) BuildFacility() {
	types := []FacilityType{FacLivingQuarters, FacLab, FacWorkshop, FacStorage, FacRadar, FacContainment, FacPsiLab, FacHangar}
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
		bs.Message = msg + fmt.Sprintf(language.String("MSG_HIRE_COST_SUFFIX"), HireCost/1000)
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

func (bs *BaseScreen) SellSelectedItem() {
	if bs.Tab == 4 && bs.Selection >= 0 && bs.Selection < len(bs.storesItems) {
		item := bs.storesItems[bs.Selection]
		value := bs.Base.SellItem(item)
		if value > 0 {
			bs.Game.Funds += value
			bs.Message = fmt.Sprintf(language.String("MSG_SOLD"), data.ItemDisplayName(item), value/1000)
		}
	}
}

func (bs *BaseScreen) BuyInterceptor() {
	if bs.Base.BuyInterceptor("avalanche", &bs.Game.Funds) {
		bs.Message = language.String("MSG_INTERCEPTOR_PURCHASED")
	} else {
		bs.Message = language.String("MSG_CANNOT_BUY_INTERCEPTOR")
	}
}

func (bs *BaseScreen) Update() {}

func (bs *BaseScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()

	ctx.DrawPanel(0, 0, w, h-3, language.String("BASE_MANAGEMENT"), engine.StyleDefault)

	tabs := []string{language.String("TAB_FACILITIES"), language.String("TAB_SOLDIERS"), language.String("TAB_RESEARCH"), language.String("TAB_MANUFACTURE"), language.String("TAB_TRANSFER"), language.String("TAB_HANGARS")}
	tabW := 0
	for _, t := range tabs {
		tw := engine.StringWidth(t) + 4 // brackets + space
		if tw > tabW {
			tabW = tw
		}
	}
	if tabW < 12 {
		tabW = 12
	}
	for i, t := range tabs {
		style := engine.StyleDefault
		if i == bs.Tab {
			style = engine.StyleHighlight
		}
		ctx.DrawString(2+i*tabW, 1, fmt.Sprintf("[%s]", t), style)
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
	case 5:
		bs.renderHangars(ctx, 2, contentY, w-4, h-6)
	}

	cap := bs.Base.LivingCapacity()
	soldStr := fmt.Sprintf(language.String("BASE_SOLDIERS"), len(bs.Base.Soldiers), cap)
	ctx.DrawString(2, h-3, fmt.Sprintf(language.String("BASE_PERSONNEL"), bs.Base.Scientists, bs.Base.Engineers, soldStr), engine.StyleDefault)
	fundsStr := fmt.Sprintf(language.String("GEOSCAPE_FUNDS"), bs.Game.Funds/1000)
	ctx.DrawString(w/2, h-3, fundsStr, engine.StyleGreen)
	if bs.Message != "" {
		ctx.DrawString(w*3/4, h-3, bs.Message, engine.StyleYellow)
	}
	help := language.String("HELP_BASE")
	if bs.Tab == 0 {
		help = language.String("HELP_FACILITIES")
	} else if bs.Tab == 1 {
		help = fmt.Sprintf(language.String("HELP_SOLDIERS"), HireCost/1000)
	} else if bs.Tab == 2 {
		help = language.String("HELP_TAB_RESEARCH")
	} else if bs.Tab == 3 {
		help = language.String("HELP_TAB_MANUFACTURE")
	} else if bs.Tab == 4 {
		help = language.String("HELP_TAB_TRANSFER")
	} else if bs.Tab == 5 {
		help = language.String("HELP_HANGARS")
	}
	ctx.DrawMarkupString(2, h-1, help, engine.StyleGray, engine.StyleHotkey)
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
		line := fmt.Sprintf(language.String("FACILITY_LINE_FORMAT"), FacilityDisplayName(ft), count, def.Cost/1000, building)
		ctx.DrawString(x, y+2+i, line, style)
	}

	// Adjacency bonus info
	yOff := y + 2 + len(facTypes) + 1
	adj := []struct {
		label string
		fac1  FacilityType
		fac2  FacilityType
		bonus string
	}{
		{language.String("FAC_ADJACENT_LABS"), FacLab, FacLab, language.String("FAC_BONUS_RESEARCH")},
		{language.String("FAC_ADJACENT_WORKSHOPS"), FacWorkshop, FacWorkshop, language.String("FAC_BONUS_MANUFACTURE")},
		{language.String("FAC_ADJACENT_LIVING"), FacLivingQuarters, FacLivingQuarters, language.String("FAC_BONUS_HP")},
	}
	for _, a := range adj {
		n := bs.Base.AdjacentCount(a.fac1, a.fac2)
		ctx.DrawString(x, yOff, fmt.Sprintf(language.String("ADJACENCY_LINE_FORMAT"), a.label, n, a.bonus), engine.StyleGray)
		yOff++
	}
}

func (bs *BaseScreen) renderSoldiers(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_SOLDIERS"), engine.StyleCyanBold)

	squad := bs.Base.Soldiers
	if len(squad) == 0 {
		ctx.DrawString(x, y+2, language.String("SECTION_NO_SOLDIERS"), engine.StyleGray)
		return
	}

	header := fmt.Sprintf(language.String("SOLDIER_HEADER_FORMAT"),
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
			woundsStr = fmt.Sprintf(language.String("WOUNDS_FORMAT"), s.Wounds)
		}
		line := fmt.Sprintf(language.String("SOLDIER_ROW_FORMAT"),
			s.Name, s.Rank, s.HP, s.MaxTU, s.Accuracy, s.Bravery, s.Strength, s.Kills, woundsStr)
		if s.Wounds > 0 {
			ctx.DrawString(x, y+3+i, line, engine.StyleRed)
		} else {
			ctx.DrawString(x, y+3+i, line, style)
		}
	}

	if bs.Selection >= 0 && bs.Selection < len(squad) {
		s := squad[bs.Selection]
		portraitImg := engine.MakeSoldierPortrait(s.Name, s.Armor, 20, 24)
		portX := x + w - portraitImg.Width - 4
		ctx.DrawPixelImageFramed(portX, y+2, portraitImg, engine.StyleCyan)
	}
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

	// Sort items alphabetically to prevent flickering
	var items []string
	for item, qty := range bs.Base.Stores {
		if qty > 0 {
			items = append(items, item)
		}
	}
	sort.Strings(items)
	bs.storesItems = items

	for i, item := range items {
		if y+i >= h+2 {
			break
		}
		qty := bs.Base.Stores[item]
		style := engine.StyleDefault
		if i == bs.Selection {
			style = engine.StyleHighlight
		}
		ctx.DrawString(x, y+i, fmt.Sprintf(language.String("STORES_LINE_FORMAT"), data.ItemDisplayName(item), qty), style)
	}
	if len(items) == 0 {
		ctx.DrawString(x, y, language.String("SECTION_NO_ITEMS"), engine.StyleGray)
	}
}

func (bs *BaseScreen) renderHangars(ctx *engine.ScreenCtx, x, y, w, h int) {
	ctx.DrawString(x, y, language.String("SECTION_HANGARS"), engine.StyleCyanBold)
	y += 2
	idx := 0
	for _, hg := range bs.Base.Hangars {
		statusKey := "INTERCEPTOR_STATUS_" + strings.ToUpper(hg.Status)
		if hg.Status == language.String("INTERCEPTOR_STATUS_DESTROYED") {
			continue
		}
		style := engine.StyleDefault
		if idx == bs.Selection {
			style = engine.StyleHighlight
		}
		wpn := data.InterceptorWeapons[hg.WeaponKey]
		line := fmt.Sprintf(language.String("LINE_HANGAR_INFO"), idx+1, language.String(statusKey), hg.HP, hg.MaxHP, wpn.DisplayName(hg.WeaponKey), hg.Ammo)
		ctx.DrawString(x, y+idx, line, style)
		idx++
	}
	if idx == 0 {
		ctx.DrawString(x, y, language.String("MSG_NO_INTERCEPTORS"), engine.StyleGray)
	}
}

func (bs *BaseScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		audio.PlayMenuNav()
		bs.Selection--
		if bs.Selection < 0 {
			if bs.Tab == 1 {
				bs.Selection = len(bs.Base.Soldiers) - 1
			} else if bs.Tab == 4 {
				bs.Selection = len(bs.storesItems) - 1
			} else if bs.Tab == 5 {
				bs.Selection = len(bs.Base.Hangars) - 1
			} else {
				bs.Selection = 6
			}
		}
	case tcell.KeyDown:
		audio.PlayMenuNav()
		bs.Selection++
		if bs.Tab == 1 {
			if bs.Selection >= len(bs.Base.Soldiers) {
				bs.Selection = 0
			}
		} else if bs.Tab == 4 {
			if bs.Selection >= len(bs.storesItems) {
				bs.Selection = 0
			}
		} else if bs.Tab == 5 {
			if bs.Selection >= len(bs.Base.Hangars) {
				bs.Selection = 0
			}
		} else {
			if bs.Selection > 6 {
				bs.Selection = 0
			}
		}
	case tcell.KeyLeft:
		audio.PlayMenuNav()
		bs.Tab--
		if bs.Tab < 0 {
			bs.Tab = 5
		}
		bs.Selection = 0
	case tcell.KeyRight:
		audio.PlayMenuNav()
		bs.Tab++
		if bs.Tab > 5 {
			bs.Tab = 0
		}
		bs.Selection = 0
	}
	switch e.Str() {
	case "1":
		bs.Tab = 0
	case "2":
		bs.Tab = 1
	case "3":
		bs.Tab = 2
	case "4":
		bs.Tab = 3
	case "5":
		bs.Tab = 4
	case "6":
		bs.Tab = 5
	case "b", "B":
		if bs.Tab == 5 {
			bs.BuyInterceptor()
		} else {
			bs.BuildFacility()
		}
	case "w", "W":
		if bs.Tab == 5 && len(bs.Base.Hangars) > 0 {
			name := bs.Base.ChangeInterceptorWeapon(bs.Selection)
			if name != "" {
				bs.Message = fmt.Sprintf(language.String("MSG_WEAPON_CHANGED"), name)
			}
		}
	case "s", "S":
		if bs.Tab == 4 {
			bs.SellSelectedItem()
		} else {
			bs.SellFacility()
		}
	case "h", "H":
		bs.HireSoldier()
	case "d", "D":
		if bs.Tab == 5 && len(bs.Base.Hangars) > 0 {
			bs.Game.SetScreen(engine.StatePlaneDesigner, NewPlaneDesignerScreen(bs.Game, bs.Base, bs.Selection))
			bs.Game.PushState(engine.StatePlaneDesigner)
		} else {
			bs.DismissSoldier()
		}
	case "e", "E":
		if bs.Tab == 1 && len(bs.Base.Soldiers) > 0 {
			bs.Game.PushState(engine.StateEquip)
		}
	case "r", "R":
		if bs.Tab == 2 {
			bs.Game.PushState(engine.StateResearch)
		}
	case "m", "M":
		if bs.Tab == 3 {
			bs.Game.PushState(engine.StateManufacture)
		}
	case "g", "G":
		bs.Game.SetScreen(engine.StateWeaponDesigner, NewWeaponDesignerScreen(bs.Game, bs.Base))
		bs.Game.PushState(engine.StateWeaponDesigner)
	}
}

func (bs *BaseScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := bs.Game.ScreenSize()

	// Handle help bar clicks (bottom bar)
	if y == h-1 {
		help := language.String("HELP_BASE")
		if bs.Tab == 0 {
			help = language.String("HELP_FACILITIES")
		} else if bs.Tab == 1 {
			help = language.String("HELP_SOLDIERS")
		} else if bs.Tab == 2 {
			help = language.String("HELP_TAB_RESEARCH")
		} else if bs.Tab == 3 {
			help = language.String("HELP_TAB_MANUFACTURE")
		} else if bs.Tab == 4 {
			help = language.String("HELP_TAB_TRANSFER")
		}
		helpActions := []string{"=Build", "=Sell", "=Hire", "=Equip", "=Dismiss", "=Research", "=Manufacture", "=Tab", "=Navigate", "=Back"}
		helpFuncs := []func(){
			func() { bs.Tab = 0; bs.BuildFacility() },
			func() {
				if bs.Tab == 4 {
					bs.SellSelectedItem()
				} else {
					bs.SellFacility()
				}
			},
			func() { bs.HireSoldier() },
			func() {
				if bs.Tab == 1 && len(bs.Base.Soldiers) > 0 {
					bs.Game.PushState(engine.StateEquip)
				}
			},
			func() { bs.DismissSoldier() },
			func() {
				if bs.Tab == 2 {
					bs.Game.PushState(engine.StateResearch)
				}
			},
			func() {
				if bs.Tab == 3 {
					bs.Game.PushState(engine.StateManufacture)
				}
			},
			func() { bs.Tab = (bs.Tab + 1) % 6; bs.Selection = 0 },
			nil,
			func() { bs.Game.PopState() },
		}
		off := 2
		for i, action := range helpActions {
			pos := strings.Index(help, action)
			if pos < 0 {
				continue
			}
			start := off + pos
			end := off + pos + len(action)
			if x >= start && x <= end && helpFuncs[i] != nil {
				helpFuncs[i]()
				return
			}
		}
		return
	}

	if y == 1 {
		tabs := []string{language.String("TAB_FACILITIES"), language.String("TAB_SOLDIERS"), language.String("TAB_RESEARCH"), language.String("TAB_MANUFACTURE"), language.String("TAB_TRANSFER"), language.String("TAB_HANGARS")}
		tabW := 0
		for _, t := range tabs {
			tw := engine.StringWidth(t) + 4
			if tw > tabW {
				tabW = tw
			}
		}
		if tabW < 12 {
			tabW = 12
		}
		for i := 0; i < len(tabs); i++ {
			tx := 2 + i*tabW
			if x >= tx && x <= tx+engine.StringWidth(tabs[i])+2 {
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
