package base

import (
	"fmt"
	"sort"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/engine"
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type ResearchScreen struct {
	Game       *engine.Game
	Base       *Base
	Selection  int
	Message    string
}

func NewResearchScreen(g *engine.Game, b *Base) *ResearchScreen {
	return &ResearchScreen{
		Game: g,
		Base: b,
	}
}

func (rs *ResearchScreen) Update() {}

func (rs *ResearchScreen) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	ctx.DrawPanel(0, 0, w, h, language.String("RESEARCH_TITLE"), engine.StyleDefault)

	if rs.Base.TotalLabs() == 0 {
		ctx.DrawString(2, 3, language.String("NO_LABS_RESEARCH"), engine.StyleGray)
		ctx.DrawString(2, 5, language.String("PRESS_ESC"), engine.StyleGray)
		return
	}

	ctx.DrawString(2, 2, fmt.Sprintf(language.String("LABS_INFO"), rs.Base.TotalLabs(), rs.Base.Scientists), engine.StyleCyanBold)

	if rs.Base.ActiveResearch != nil && !rs.Base.ActiveResearch.Completed {
		topic := data.ResearchByID(rs.Base.ActiveResearch.TopicID)
		if topic != nil {
			pct := rs.Base.ActiveResearch.Progress * 100 / rs.Base.ActiveResearch.Cost
			ctx.DrawString(2, 3, fmt.Sprintf(language.String("RESEARCH_IN_PROGRESS"),
				topic.Name, pct, rs.Base.ActiveResearch.Scientists), engine.StyleGreen)
			ctx.DrawString(2, 4, fmt.Sprintf(language.String("RESEARCH_UNASSIGNED"), rs.Base.UnassignedScientists), engine.StyleYellow)
		}
	} else {
		ctx.DrawString(2, 3, language.String("NO_ACTIVE_RESEARCH"), engine.StyleGray)
	}

	ctx.DrawString(2, 5, language.String("AVAILABLE_TOPICS"), engine.StyleCyanBold)

	topics := rs.getAvailableTopics()
	if len(topics) == 0 {
		ctx.DrawString(2, 7, language.String("NO_TOPICS"), engine.StyleGray)
		return
	}
	if rs.Selection >= len(topics) {
		rs.Selection = len(topics) - 1
	}

	for i, topic := range topics {
		if 7+i >= h-3 {
			break
		}
		style := engine.StyleDefault
		if i == rs.Selection {
			style = engine.StyleHighlight
		}
		req := ""
		if len(topic.Requires) > 0 {
			reqStr := ""
			for j, r := range topic.Requires {
				if j > 0 {
					reqStr += ", "
				}
				rt := data.ResearchByID(r)
				if rt != nil {
					reqStr += rt.Name
				} else {
					reqStr += r
				}
			}
			req = fmt.Sprintf(language.String("RESEARCH_REQUIRES"), reqStr)
		}
		line := fmt.Sprintf(language.String("RESEARCH_COST"), topic.Name, topic.Cost, req)
		ctx.DrawString(2, 7+i, line, style)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_RESEARCH")
	ctx.DrawString(1, h-1, help, engine.StyleGray)

	if rs.Message != "" {
		ctx.DrawString(2, h-2, rs.Message, engine.StyleYellow)
	}
}

func (rs *ResearchScreen) getAvailableTopics() []*data.ResearchTopic {
	var topics []*data.ResearchTopic
	for i := range data.ResearchTree {
		topic := &data.ResearchTree[i]
		if rs.Base.CanResearch(topic) {
			topics = append(topics, topic)
		}
	}
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Cost < topics[j].Cost
	})
	return topics
}

func (rs *ResearchScreen) startResearch() {
	topics := rs.getAvailableTopics()
	if rs.Selection >= len(topics) {
		rs.Selection = 0
	}
	if len(topics) == 0 {
		return
	}
	topic := topics[rs.Selection]
	if rs.Base.StartResearch(topic.ID) {
		rs.Message = fmt.Sprintf(language.String("MSG_RESEARCH_STARTED"), topic.Name)
	} else {
		rs.Message = language.String("MSG_CANNOT_RESEARCH")
	}
}

func (rs *ResearchScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		rs.Selection--
		if rs.Selection < 0 {
			rs.Selection = 0
		}
	case tcell.KeyDown:
		topics := rs.getAvailableTopics()
		rs.Selection++
		if rs.Selection >= len(topics) {
			rs.Selection = len(topics) - 1
		}
	}
	switch e.Str() {
	case "\r":
		rs.startResearch()
	case "+":
		rs.Base.AssignScientists(1)
	case "-":
		rs.Base.AssignScientists(-1)
	}
}

func (rs *ResearchScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := rs.Game.ScreenSize()

	// Handle help bar clicks (bottom bar)
	if y == h-1 {
		// Help bar: "j/k=Select  Enter=Start  [Esc]=Back"
		switch {
		case x >= 1 && x <= 3: // j/k=Select
			// Scroll down
			topics := rs.getAvailableTopics()
			if rs.Selection < len(topics)-1 {
				rs.Selection++
			}
		case x >= 5 && x <= 12: // Enter=Start
			rs.startResearch()
		case x >= 14 && x <= 20: // [Esc]=Back
			rs.Game.PopState()
		}
		return
	}

	if y >= 7 && y < h-2 {
		rs.Selection = y - 7
	}

	if x > 0 && y >= 3 && y <= 4 {
		rs.startResearch()
	}
}
