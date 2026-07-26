package base

import (
	"fmt"
	"sort"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type ManufacturePlan struct {
	Name      string
	ItemKey   string
	Days      int
	Materials map[string]int
	RequiresResearch string // optional: research topic ID needed to unlock this plan
}

func (p *ManufacturePlan) DisplayName() string {
	if r, ok := data.RuleItems[p.ItemKey]; ok {
		return r.DisplayName()
	}
	if a, ok := data.Armors[p.ItemKey]; ok {
		return a.DisplayNameByKey(p.ItemKey)
	}
	return p.Name
}

var ManufacturePlans = []ManufacturePlan{
	{Name: "Pistol", ItemKey: "pistol", Days: 3, Materials: map[string]int{"alloys": 1}},
	{Name: "Rifle", ItemKey: "rifle", Days: 5, Materials: map[string]int{"alloys": 2}},
	{Name: "Heavy Cannon", ItemKey: "heavy", Days: 7, Materials: map[string]int{"alloys": 3}},
	{Name: "Auto Cannon", ItemKey: "auto", Days: 6, Materials: map[string]int{"alloys": 3}},
	{Name: "Rocket Launcher", ItemKey: "rocket", Days: 8, Materials: map[string]int{"alloys": 4, "elerium": 1}},
	{Name: "Stun Rod", ItemKey: "stun_rod", Days: 2, Materials: map[string]int{"alloys": 1}},
	{Name: "Laser Pistol", ItemKey: "laser_pistol", Days: 6, Materials: map[string]int{"alloys": 3, "elerium": 1}, RequiresResearch: "laser_weapons"},
	{Name: "Laser Rifle", ItemKey: "laser_rifle", Days: 10, Materials: map[string]int{"alloys": 5, "elerium": 2}, RequiresResearch: "laser_weapons"},
	{Name: "Plasma Pistol", ItemKey: "plasma_pistol", Days: 8, Materials: map[string]int{"alloys": 4, "elerium": 3}, RequiresResearch: "plasma_weapons"},
	{Name: "Plasma Rifle", ItemKey: "plasma_rifle", Days: 14, Materials: map[string]int{"alloys": 8, "elerium": 4}, RequiresResearch: "plasma_weapons"},
	{Name: "Heavy Plasma", ItemKey: "heavy_plasma", Days: 18, Materials: map[string]int{"alloys": 10, "elerium": 6}, RequiresResearch: "heavy_plasma"},
	{Name: "Personal Armour", ItemKey: "personal", Days: 6, Materials: map[string]int{"alloys": 2}, RequiresResearch: "personal_armour"},
	{Name: "Light Suit", ItemKey: "light", Days: 10, Materials: map[string]int{"alloys": 4, "elerium": 1}, RequiresResearch: "light_suit"},
	{Name: "Medium Suit", ItemKey: "medium", Days: 14, Materials: map[string]int{"alloys": 6, "elerium": 2}, RequiresResearch: "medium_suit"},
	{Name: "Heavy Suit", ItemKey: "heavy", Days: 18, Materials: map[string]int{"alloys": 8, "elerium": 3}, RequiresResearch: "heavy_suit"},
	{Name: "Medi-Kit", ItemKey: "medi_kit", Days: 3, Materials: map[string]int{"alloys": 1}},
}

type ManufactureScreen struct {
	Game       *engine.Game
	Base       *Base
	Selection  int
	Message    string
}

func NewManufactureScreen(g *engine.Game, b *Base) *ManufactureScreen {
	return &ManufactureScreen{
		Game: g,
		Base: b,
	}
}

func (ms *ManufactureScreen) Update() {}

func (ms *ManufactureScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("MANUFACTURE_TITLE"), engine.StyleDefault)

	if ms.Base.TotalWorkshops() == 0 {
		ctx.DrawString(2, 3, language.String("NO_WORKSHOPS_MFG"), engine.StyleGray)
		ctx.DrawString(2, 5, language.String("PRESS_ESC"), engine.StyleGray)
		return
	}

	ctx.DrawString(2, 2, fmt.Sprintf(language.String("WORKSHOPS_INFO"), ms.Base.TotalWorkshops(), ms.Base.Engineers), engine.StyleCyanBold)

	plans := ms.getBuildablePlans()
	plansLen := len(plans)
	queueLen := len(ms.Base.ManufactureQueue)
	totalLen := plansLen + queueLen

	// Clamp selection
	if totalLen > 0 && ms.Selection >= totalLen {
		ms.Selection = totalLen - 1
	}
	if ms.Selection < 0 {
		ms.Selection = 0
	}

	// --- Queue section ---
	if queueLen > 0 {
		ctx.DrawString(2, 3, language.String("MFG_ACTIVE_QUEUE"), engine.StyleGreen)
		maxQueueY := h/2 - 3
		if maxQueueY > h-4 {
			maxQueueY = h - 4
		}
		if maxQueueY < 4 {
			maxQueueY = 4
		}
		y := 4
		for i, job := range ms.Base.ManufactureQueue {
			if y >= maxQueueY {
				break
			}
			pct := 0
			if job.CostDays > 0 {
				pct = job.Progress * 100 / job.CostDays
			}
			status := fmt.Sprintf(language.String("MFG_QUEUE_LINE"), job.DisplayName(), job.Count, pct, job.Engineers)
			if job.Completed {
				status += language.String("MFG_DONE")
			}
			style := engine.StyleDefault
			if ms.Selection >= plansLen && i == ms.Selection-plansLen {
				style = engine.StyleHighlight
			}
			ctx.DrawString(2, y, status, style)
			y++
		}
	}

	// --- Unassigned engineers ---
	unassignedY := 3
	if queueLen > 0 {
		unassignedY = 4 + queueLen
		if unassignedY > h/2-2 {
			unassignedY = h/2 - 2
		}
	}
	if unassignedY < 3 {
		unassignedY = 3
	}
	if unassignedY < h-2 {
		ctx.DrawString(2, unassignedY, fmt.Sprintf(language.String("MFG_UNASSIGNED"), ms.Base.UnassignedEngineers), engine.StyleYellow)
	}

	// --- Buildable plans section ---
	ctx.DrawString(2, h/2, language.String("MFG_BUILDABLE"), engine.StyleCyanBold)

	if len(plans) == 0 {
		ctx.DrawString(2, h/2+2, language.String("MFG_NO_ITEMS"), engine.StyleGray)
		return
	}

	startY := h/2 + 1
	for i, plan := range plans {
		if startY+i >= h-3 {
			break
		}
		style := engine.StyleDefault
		if i == ms.Selection {
			style = engine.StyleHighlight
		}
		matStr := ""
		for mat, qty := range plan.Materials {
			have := ms.Base.CountItem(mat)
			matStr += fmt.Sprintf(language.String("MFG_MATERIAL_COUNT"), mat, have, qty)
		}
		line := fmt.Sprintf(language.String("MFG_BUILDABLE_LINE"), plan.DisplayName(), plan.Days, matStr)
		ctx.DrawString(2, startY+i, line, style)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	ctx.DrawMarkupString(1, h-1, language.String("HELP_MANUFACTURE"), engine.StyleGray, engine.StyleHotkey)

	if ms.Message != "" {
		ctx.DrawString(2, h-2, ms.Message, engine.StyleYellow)
	}
}

func (ms *ManufactureScreen) getBuildablePlans() []ManufacturePlan {
	var plans []ManufacturePlan
	for _, plan := range ManufacturePlans {
		// Check research unlock requirement
		if plan.RequiresResearch != "" && !ms.Base.HasResearch(plan.RequiresResearch) {
			continue
		}
		canBuild := true
		for mat, qty := range plan.Materials {
			if ms.Base.CountItem(mat) < qty {
				canBuild = false
				break
			}
		}
		if canBuild {
			plans = append(plans, plan)
		}
	}
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Days < plans[j].Days
	})
	return plans
}

func (ms *ManufactureScreen) startManufacture() {
	plans := ms.getBuildablePlans()
	if ms.Selection >= len(plans) {
		ms.Selection = 0
	}
	if len(plans) == 0 {
		return
	}
	plan := plans[ms.Selection]
	if ms.Base.StartManufacture(plan.ItemKey, 1, plan.Materials) {
		ms.Message = fmt.Sprintf(language.String("MSG_MFG_STARTED"), plan.Name)
	} else {
		ms.Message = language.String("MSG_CANNOT_MFG")
	}
}

func (ms *ManufactureScreen) HandleKey(e *tcell.EventKey) {
	plans := ms.getBuildablePlans()
	plansLen := len(plans)
	queueLen := len(ms.Base.ManufactureQueue)
	totalLen := plansLen + queueLen
	if totalLen == 0 {
		return
	}

	switch e.Key() {
	case tcell.KeyUp:
		if ms.Selection > 0 {
			ms.Selection--
		}
	case tcell.KeyDown:
		if ms.Selection < totalLen-1 {
			ms.Selection++
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		if ms.Selection >= plansLen {
			idx := ms.Selection - plansLen
			if idx < queueLen {
				ms.Base.CancelManufacture(idx)
				queueLen = len(ms.Base.ManufactureQueue)
				totalLen = plansLen + queueLen
				if ms.Selection >= totalLen && totalLen > 0 {
					ms.Selection = totalLen - 1
				}
				if totalLen == 0 {
					ms.Selection = 0
				}
			}
		}
	}

	switch e.Str() {
	case "\r":
		if ms.Selection < plansLen {
			ms.startManufacture()
		}
	case "+":
		if ms.Selection >= plansLen {
			idx := ms.Selection - plansLen
			if idx < len(ms.Base.ManufactureQueue) {
				ms.Base.AssignEngineers(idx, 1)
			}
		}
	case "-":
		if ms.Selection >= plansLen {
			idx := ms.Selection - plansLen
			if idx < len(ms.Base.ManufactureQueue) {
				ms.Base.AssignEngineers(idx, -1)
			}
		}
	}
}

func (ms *ManufactureScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := ms.Game.ScreenSize()

	// Handle help bar clicks (bottom bar) by parsing [key] tokens
	if y == h-1 {
		help := language.String("HELP_MANUFACTURE")
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
			segEnd := col + engine.StringWidth(string(runes[i+1 : end]))
			if x >= segStart && x <= segEnd {
				key := string(runes[i+1 : end])
				switch key {
				case "↑", "↓":
					plans := ms.getBuildablePlans()
					totalLen := len(plans) + len(ms.Base.ManufactureQueue)
					if ms.Selection < totalLen-1 {
						ms.Selection++
					}
				case "Enter":
					ms.startManufacture()
				case "+":
					plansLen := len(ms.getBuildablePlans())
					if ms.Selection >= plansLen {
						idx := ms.Selection - plansLen
						if idx < len(ms.Base.ManufactureQueue) {
							ms.Base.AssignEngineers(idx, 1)
						}
					}
				case "-":
					plansLen := len(ms.getBuildablePlans())
					if ms.Selection >= plansLen {
						idx := ms.Selection - plansLen
						if idx < len(ms.Base.ManufactureQueue) {
							ms.Base.AssignEngineers(idx, -1)
						}
					}
				case "Del":
					plansLen := len(ms.getBuildablePlans())
					if ms.Selection >= plansLen {
						idx := ms.Selection - plansLen
						if idx < len(ms.Base.ManufactureQueue) {
							ms.Base.CancelManufacture(idx)
							totalLen := plansLen + len(ms.Base.ManufactureQueue)
							if ms.Selection >= totalLen && totalLen > 0 {
								ms.Selection = totalLen - 1
							}
						}
					}
				case "Esc":
					ms.Game.PopState()
				}
				return
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	plans := ms.getBuildablePlans()
	plansLen := len(plans)
	queueLen := len(ms.Base.ManufactureQueue)

	// Queue section click (top area)
	if queueLen > 0 {
		queueStartY := 4
		if y >= queueStartY && y < h/2-1 {
			idx := y - queueStartY
			if idx < queueLen {
				ms.Selection = plansLen + idx
				return
			}
		}
	}

	// Buildable plans section click (bottom area)
	startY := h/2 + 1
	if y >= startY && y < h-2 {
		idx := y - startY
		if idx < plansLen {
			ms.Selection = idx
		}
		return
	}

	// Click on message line to start manufacture
	if x > 0 && y == h-2 && ms.Selection < plansLen {
		ms.startManufacture()
	}
}
