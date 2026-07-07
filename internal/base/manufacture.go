package base

import (
	"fmt"
	"sort"

	"github.com/civ13/ycom/internal/engine"
	"github.com/gdamore/tcell/v2"
)

type ManufacturePlan struct {
	Name      string
	ItemKey   string
	Days      int
	Materials map[string]int
}

var ManufacturePlans = []ManufacturePlan{
	{Name: "Pistol", ItemKey: "pistol", Days: 3, Materials: map[string]int{"alloys": 1}},
	{Name: "Rifle", ItemKey: "rifle", Days: 5, Materials: map[string]int{"alloys": 2}},
	{Name: "Heavy Cannon", ItemKey: "heavy", Days: 7, Materials: map[string]int{"alloys": 3}},
	{Name: "Auto Cannon", ItemKey: "auto", Days: 6, Materials: map[string]int{"alloys": 3}},
	{Name: "Rocket Launcher", ItemKey: "rocket", Days: 8, Materials: map[string]int{"alloys": 4, "elerium": 1}},
	{Name: "Stun Rod", ItemKey: "stun_rod", Days: 2, Materials: map[string]int{"alloys": 1}},
	{Name: "Personal Armour", ItemKey: "personal", Days: 6, Materials: map[string]int{"alloys": 2}},
	{Name: "Light Suit", ItemKey: "light", Days: 10, Materials: map[string]int{"alloys": 4, "elerium": 1}},
	{Name: "Medium Suit", ItemKey: "medium", Days: 14, Materials: map[string]int{"alloys": 6, "elerium": 2}},
	{Name: "Heavy Suit", ItemKey: "heavy", Days: 18, Materials: map[string]int{"alloys": 8, "elerium": 3}},
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
	ctx.DrawPanel(0, 0, w, h, "MANUFACTURING", engine.StyleDefault)

	if ms.Base.TotalWorkshops() == 0 {
		ctx.DrawString(2, 3, "No workshops. Build a Workshop first.", engine.StyleGray)
		ctx.DrawString(2, 5, "Press Esc to return.", engine.StyleGray)
		return
	}

	ctx.DrawString(2, 2, fmt.Sprintf("Workshops: %d  Engineers: %d", ms.Base.TotalWorkshops(), ms.Base.Engineers), engine.StyleCyanBold)

	if len(ms.Base.ManufactureQueue) > 0 {
		ctx.DrawString(2, 3, "ACTIVE QUEUE:", engine.StyleGreen)
		y := 4
		for _, job := range ms.Base.ManufactureQueue {
			if y >= h-4 {
				break
			}
			pct := 0
			if job.CostDays > 0 {
				pct = job.Progress * 100 / job.CostDays
			}
			status := fmt.Sprintf("%s x%d (%d%%)", job.ItemKey, job.Count, pct)
			if job.Completed {
				status += " [DONE]"
			}
			ctx.DrawString(2, y, status, engine.StyleDefault)
			y++
		}
	}

	ctx.DrawString(2, h/2, "BUILDABLE ITEMS:", engine.StyleCyanBold)

	plans := ms.getBuildablePlans()
	if len(plans) == 0 {
		ctx.DrawString(2, h/2+2, "No items available. Collect more alloys/elerium.", engine.StyleGray)
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
			matStr += fmt.Sprintf(" %s:%d/%d", mat, have, qty)
		}
		line := fmt.Sprintf("%-18s Days:%d%s", plan.Name, plan.Days, matStr)
		ctx.DrawString(2, startY+i, line, style)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := "j/k=Select  Enter=Build  Esc=Back"
	ctx.DrawString(1, h-1, help, engine.StyleGray)

	if ms.Message != "" {
		ctx.DrawString(2, h-2, ms.Message, engine.StyleYellow)
	}
}

func (ms *ManufactureScreen) getBuildablePlans() []ManufacturePlan {
	var plans []ManufacturePlan
	for _, plan := range ManufacturePlans {
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
		ms.Message = fmt.Sprintf("Manufacturing started: %s", plan.Name)
	} else {
		ms.Message = "Cannot manufacture!"
	}
}

func (ms *ManufactureScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		ms.Selection--
		if ms.Selection < 0 {
			ms.Selection = 0
		}
	case tcell.KeyDown:
		ms.Selection++
	case tcell.KeyRune:
		switch e.Rune() {
		case 'j':
			ms.Selection++
		case 'k':
			ms.Selection--
			if ms.Selection < 0 {
				ms.Selection = 0
			}
		case '\r':
			ms.startManufacture()
		}
	case tcell.KeyEnter:
		ms.startManufacture()
	}
}

func (ms *ManufactureScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := ms.Game.ScreenSize()

	startY := h/2 + 1
	if y >= startY && y < h-2 {
		ms.Selection = y - startY
	}

	if x > 0 && y == h-2 {
		ms.startManufacture()
	}
}
