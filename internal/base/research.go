package base

import (
	"fmt"
	"sort"
	"strings"

	"github.com/civ13/termcom/internal/data"
	"github.com/civ13/termcom/internal/engine"
	"github.com/civ13/termcom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type topicStatus int

const (
	topicDone     topicStatus = iota
	topicAvailable
	topicLocked
)

type topicEntry struct {
	topic  *data.ResearchTopic
	status topicStatus
}

type ResearchScreen struct {
	Game            *engine.Game
	Base            *Base
	Selection       int
	Message         string
	ShowTree        bool
	InterrogateMode bool
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

	ctx.DrawString(2, 4, language.String("ALL_TOPICS"), engine.StyleCyanBold)

	// Show captured aliens line
	if len(rs.Base.LiveAliens) > 0 {
		ctx.DrawString(2, 5, fmt.Sprintf(language.String("RESEARCH_CAPTURED"), len(rs.Base.LiveAliens)), engine.StyleYellow)
	}

	entries := rs.getAllTopics()
	if len(entries) == 0 {
		ctx.DrawString(2, 7, language.String("NO_TOPICS"), engine.StyleGray)
		return
	}
	if rs.Selection >= len(entries) {
		if len(entries) > 0 {
			rs.Selection = len(entries) - 1
		} else {
			rs.Selection = 0
		}
	}

	listW := w - 2
	if rs.ShowTree {
		listW = w/2 - 2
	}

	for i, entry := range entries {
		if 7+i >= h-3 {
			break
		}
		style := engine.StyleDefault
		marker := "  "

		switch entry.status {
		case topicDone:
			style = engine.StyleGray
			marker = language.String("RESEARCH_DONE") + " "
		case topicLocked:
			style = engine.StyleGray
			marker = language.String("RESEARCH_LOCKED") + " "
		case topicAvailable:
			style = engine.StyleDefault
			marker = "  "
		}

		if i == rs.Selection {
			if entry.status == topicDone {
				style = engine.StyleGray.Bold(true)
			} else if entry.status == topicLocked {
				style = engine.StyleGray.Bold(true)
			} else {
				style = engine.StyleHighlight
			}
		}

		req := ""
		if len(entry.topic.Requires) > 0 {
			reqStr := ""
			for j, r := range entry.topic.Requires {
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

		line := fmt.Sprintf(language.String("RESEARCH_COST"), entry.topic.Tier, entry.topic.Name, entry.topic.Cost, req)
		displayLine := marker + line
		if len(displayLine) > listW {
			displayLine = displayLine[:listW]
		}
		ctx.DrawString(2, 7+i, displayLine, style)
	}

	if rs.ShowTree {
		selEntry := &topicEntry{}
		if rs.Selection >= 0 && rs.Selection < len(entries) {
			selEntry = &entries[rs.Selection]
		}
		rs.renderTree(ctx, w/2+1, 7, w/2-2, h-10, selEntry)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_RESEARCH")
	if rs.ShowTree {
		help = "\u2191/\u2193=Select  Enter=Start  [Esc]=Back  [T]ree=Hide"
	}
	ctx.DrawMarkupString(1, h-1, help, engine.StyleGray, engine.StyleHotkey)

	if rs.Message != "" {
		ctx.DrawString(2, h-2, rs.Message, engine.StyleYellow)
	}
}

func (rs *ResearchScreen) renderTree(ctx *engine.ScreenCtx, x, y, maxW, maxH int, entry *topicEntry) {
	ctx.DrawString(x, y, language.String("RESEARCH_TREE_TITLE"), engine.StyleCyanBold)
	y++
	ctx.DrawString(x, y, strings.Repeat("\u2500", maxW), engine.StyleGray)
	y++
	startY := y

	if entry == nil || entry.topic == nil {
		return
	}

	t := entry.topic

	// Show prerequisites
	ctx.DrawString(x, y, language.String("RESEARCH_PREREQS"), engine.StyleYellow)
	y++
	if len(t.Requires) == 0 {
		ctx.DrawString(x+2, y, "(none)", engine.StyleGray)
		y++
	} else {
		for _, reqID := range t.Requires {
			if y-startY >= maxH {
				break
			}
			rt := data.ResearchByID(reqID)
			name := reqID
			if rt != nil {
				name = rt.Name
			}
			done := rs.Base.HasResearch(reqID)
			prefix := "\u251C\u2500\u2500 "
			if done {
				prefix = "\u251C\u2500\u2500 "
				ctx.DrawString(x+2, y, prefix+language.String("RESEARCH_DONE")+" "+name, engine.StyleGreen)
			} else {
				ctx.DrawString(x+2, y, prefix+language.String("RESEARCH_LOCKED")+" "+name, engine.StyleRed)
			}
			y++
		}
	}

	y++
	ctx.DrawString(x, y, language.String("RESEARCH_UNLOCKS"), engine.StyleYellow)
	y++

	unlocks := rs.getUnlocks(t)
	if len(unlocks) == 0 {
		ctx.DrawString(x+2, y, "(none)", engine.StyleGray)
		y++
	} else {
		for _, u := range unlocks {
			if y-startY >= maxH {
				break
			}
			prefix := "\u251C\u2500\u2500 "
			ctx.DrawString(x+2, y, prefix+u, engine.StyleCyan)
			y++
		}
	}

	// Show children (topics that require this one)
	y++
	ctx.DrawString(x, y, "Unlocks topics:", engine.StyleYellow)
	y++
	children := rs.getChildren(t)
	if len(children) == 0 {
		ctx.DrawString(x+2, y, "(none)", engine.StyleGray)
	} else {
		for _, child := range children {
			if y >= y+maxH {
				break
			}
			done := rs.Base.HasResearch(child.ID)
			prefix := "\u251C\u2500\u2500 "
			childLine := fmt.Sprintf("[T%d] %s", child.Tier, child.Name)
			if done {
				ctx.DrawString(x+2, y, prefix+language.String("RESEARCH_DONE")+" "+childLine, engine.StyleGreen)
			} else {
				ctx.DrawString(x+2, y, prefix+language.String("RESEARCH_LOCKED")+" "+childLine, engine.StyleCyan)
			}
			y++
		}
	}
}

func (rs *ResearchScreen) getUnlocks(t *data.ResearchTopic) []string {
	var unlocks []string
	for _, item := range t.UnlockItems {
		unlocks = append(unlocks, "Item: "+item)
	}
	for _, weap := range t.UnlockWeap {
		unlocks = append(unlocks, "Weapon: "+weap)
	}
	for _, arm := range t.UnlockArmor {
		unlocks = append(unlocks, "Armor: "+arm)
	}
	if t.AlienLore {
		unlocks = append(unlocks, "Alien Lore")
	}
	return unlocks
}

func (rs *ResearchScreen) getChildren(t *data.ResearchTopic) []*data.ResearchTopic {
	var children []*data.ResearchTopic
	for i := range data.ResearchTree {
		topic := &data.ResearchTree[i]
		for _, req := range topic.Requires {
			if req == t.ID {
				children = append(children, topic)
				break
			}
		}
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].Tier < children[j].Tier
	})
	return children
}

func (rs *ResearchScreen) getAllTopics() []topicEntry {
	var entries []topicEntry
	for i := range data.ResearchTree {
		topic := &data.ResearchTree[i]
		status := topicLocked
		if rs.Base.HasResearch(topic.ID) {
			status = topicDone
		} else if rs.Base.CanResearch(topic) {
			status = topicAvailable
		}
		entries = append(entries, topicEntry{topic: topic, status: status})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].status != entries[j].status {
			return entries[i].status < entries[j].status
		}
		if entries[i].topic.Tier != entries[j].topic.Tier {
			return entries[i].topic.Tier < entries[j].topic.Tier
		}
		return entries[i].topic.Cost < entries[j].topic.Cost
	})
	return entries
}

func (rs *ResearchScreen) doInterrogate() {
	rs.InterrogateMode = false
	if len(rs.Base.LiveAliens) == 0 {
		rs.Message = language.String("MSG_INTERROGATE_NO_ALIEN")
		return
	}
	if rs.Base.TotalLabs() == 0 {
		rs.Message = language.String("MSG_INTERROGATE_NO_LABS")
		return
	}
	// Interrogate the first available captured alien
	alienName := rs.Base.LiveAliens[0]
	topicName, ok := rs.Base.InterrogateAlien(alienName)
	if ok {
		rs.Message = fmt.Sprintf(language.String("MSG_INTERROGATE_SUCCESS"), topicName)
	} else {
		rs.Message = language.String("MSG_INTERROGATE_NO_ALIEN")
	}
}

func (rs *ResearchScreen) startResearch() {
	entries := rs.getAllTopics()
	if rs.Selection >= len(entries) {
		rs.Selection = 0
	}
	if len(entries) == 0 {
		return
	}
	entry := entries[rs.Selection]
	if entry.status != topicAvailable {
		if entry.status == topicDone {
			rs.Message = language.String("MSG_CANNOT_RESEARCH")
		} else {
			rs.Message = language.String("MSG_CANNOT_RESEARCH")
		}
		return
	}
	if rs.Base.StartResearch(entry.topic.ID) {
		rs.Message = fmt.Sprintf(language.String("MSG_RESEARCH_STARTED"), entry.topic.Name)
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
		entries := rs.getAllTopics()
		rs.Selection++
		if rs.Selection >= len(entries) {
			rs.Selection = len(entries) - 1
		}
	}
	switch e.Str() {
	case "\r":
		if rs.InterrogateMode {
			rs.doInterrogate()
		} else {
			rs.startResearch()
		}
	case "i", "I":
		if len(rs.Base.LiveAliens) > 0 && rs.Base.TotalLabs() > 0 {
			rs.InterrogateMode = true
			rs.Message = language.String("RESEARCH_INTERROGATE_PROMPT")
		} else if len(rs.Base.LiveAliens) == 0 {
			rs.Message = language.String("MSG_INTERROGATE_NO_ALIEN")
		} else {
			rs.Message = language.String("MSG_INTERROGATE_NO_LABS")
		}
	case "+":
		rs.Base.AssignScientists(1)
	case "-":
		rs.Base.AssignScientists(-1)
	case "t", "T":
		rs.ShowTree = !rs.ShowTree
	}
}

func (rs *ResearchScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := rs.Game.ScreenSize()

	if y == h-1 {
		switch {
		case x >= 1 && x <= 3:
			entries := rs.getAllTopics()
			if rs.Selection < len(entries)-1 {
				rs.Selection++
			}
		case x >= 5 && x <= 12:
			rs.startResearch()
		case x >= 14 && x <= 20:
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
