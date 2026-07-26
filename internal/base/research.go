package base

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/language"
)

type topicStatus int

const (
	topicDone topicStatus = iota
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
	scrollOffset    int // tracks list scroll position for mouse click adjustment
	listStartY      int // row where the topic list begins (set during Render)
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

	// Dynamic layout: info lines start at row 2, topic list below them
	infoY := 2 // current info row
	ctx.DrawString(2, infoY, fmt.Sprintf(language.String("LABS_INFO"), rs.Base.TotalLabs(), rs.Base.Scientists), engine.StyleCyanBold)
	infoY++

	if rs.Base.ActiveResearch != nil && !rs.Base.ActiveResearch.Completed {
		topic := data.ResearchByID(rs.Base.ActiveResearch.TopicID)
		if topic != nil {
			cost := rs.Base.ActiveResearch.Cost
			if cost <= 0 {
				cost = 1
			}
			pct := rs.Base.ActiveResearch.Progress * 100 / cost
			ctx.DrawString(2, infoY, fmt.Sprintf(language.String("RESEARCH_IN_PROGRESS"),
				topic.DisplayName(), pct, rs.Base.ActiveResearch.Scientists), engine.StyleGreen)
			ctx.DrawString(2, infoY+1, fmt.Sprintf(language.String("RESEARCH_UNASSIGNED"), rs.Base.UnassignedScientists), engine.StyleYellow)
			infoY += 2
		}
	}
	if rs.Base.ActiveResearch == nil || rs.Base.ActiveResearch.Completed {
		ctx.DrawString(2, infoY, language.String("NO_ACTIVE_RESEARCH"), engine.StyleGray)
		infoY++
	}

	// Show captured aliens line
	if len(rs.Base.LiveAliens) > 0 {
		names := strings.Join(rs.Base.LiveAliens, ", ")
		ctx.DrawString(2, infoY, fmt.Sprintf(language.String("RESEARCH_CAPTURED_NAMES"), names), engine.StyleYellow)
		infoY++
	}
	corpses := rs.Base.AlienCorpseTypes()
	if len(corpses) > 0 {
		names := strings.Join(corpses, ", ")
		ctx.DrawString(2, infoY, fmt.Sprintf(language.String("RESEARCH_CORPSES"), names), engine.StyleYellow)
		infoY++
	}

	// Topic list header
	ctx.DrawString(2, infoY, language.String("ALL_TOPICS"), engine.StyleCyanBold)
	infoY++
	rs.listStartY = infoY
	listStartY := rs.listStartY

	entries := rs.getAllTopics()
	if len(entries) == 0 {
		ctx.DrawString(2, listStartY, language.String("NO_TOPICS"), engine.StyleGray)
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
	if rs.ShowTree && !engine.Layout.IsMobile() {
		listW = w/2 - 2
	}

	// Compute visible range with scrolling
	maxRows := h - listStartY - 2 // leave room for help bar + message
	if maxRows < 1 {
		maxRows = 1
	}
	scrollOffset := 0
	if rs.Selection >= maxRows {
		scrollOffset = rs.Selection - maxRows + 1
	}
	rs.scrollOffset = scrollOffset

	for i := scrollOffset; i < len(entries) && i-scrollOffset < maxRows; i++ {
		entry := entries[i]
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
			if entry.status == topicDone || entry.status == topicLocked {
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
					reqStr += rt.DisplayName()
				} else {
					reqStr += r
				}
			}
			req = fmt.Sprintf(language.String("RESEARCH_REQUIRES"), reqStr)
		}

		line := fmt.Sprintf(language.String("RESEARCH_COST"), entry.topic.Tier, entry.topic.DisplayName(), entry.topic.Cost, req)
		displayLine := marker + line
		if len([]rune(displayLine)) > listW {
			runes := []rune(displayLine)
			displayLine = string(runes[:listW])
		}
		ctx.DrawString(2, listStartY+i-scrollOffset, displayLine, style)
	}

	// Show scroll indicator if content is clipped
	if scrollOffset > 0 {
		ctx.DrawString(w-4, listStartY, "\u2191", engine.StyleGray)
	}
	if scrollOffset+maxRows < len(entries) {
		ctx.DrawString(w-4, h-3, "\u2193", engine.StyleGray)
	}

	if rs.ShowTree && !engine.Layout.IsMobile() {
		selEntry := &topicEntry{}
		if rs.Selection >= 0 && rs.Selection < len(entries) {
			selEntry = &entries[rs.Selection]
		}
		rs.renderTree(ctx, w/2+1, listStartY, w/2-2, h-listStartY-3, selEntry)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	help := language.String("HELP_RESEARCH")
	if rs.ShowTree && !engine.Layout.IsMobile() {
		help = language.String("HELP_RESEARCH_TREE")
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
		ctx.DrawString(x+2, y, language.String("SIDE_NONE"), engine.StyleGray)
		y++
	} else {
		for _, reqID := range t.Requires {
			if y-startY >= maxH {
				break
			}
			rt := data.ResearchByID(reqID)
			name := reqID
			if rt != nil {
				name = rt.DisplayName()
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
		ctx.DrawString(x+2, y, language.String("SIDE_NONE"), engine.StyleGray)
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
	ctx.DrawString(x, y, language.String("RESEARCH_UNLOCKS"), engine.StyleYellow)
	y++
	children := rs.getChildren(t)
	if len(children) == 0 {
		ctx.DrawString(x+2, y, language.String("SIDE_NONE"), engine.StyleGray)
	} else {
		for _, child := range children {
			if y-startY >= maxH {
				break
			}
			done := rs.Base.HasResearch(child.ID)
			prefix := "\u251C\u2500\u2500 "
			childLine := fmt.Sprintf(language.String("RESEARCH_TREE_CHILD_FMT"), child.Tier, child.DisplayName())
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
		unlocks = append(unlocks, language.String("RESEARCH_UNLOCK_ITEM")+" "+item)
	}
	for _, weap := range t.UnlockWeap {
		unlocks = append(unlocks, language.String("RESEARCH_UNLOCK_WEAPON")+" "+weap)
	}
	for _, arm := range t.UnlockArmor {
		unlocks = append(unlocks, language.String("RESEARCH_UNLOCK_ARMOR")+" "+arm)
	}
	if t.AlienLore {
		unlocks = append(unlocks, language.String("RESEARCH_UNLOCK_ALIEN_LORE"))
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
		// Hide locked autopsy topics for aliens the player hasn't encountered
		if status == topicLocked && strings.HasSuffix(topic.ID, "_autopsy") {
			name := strings.TrimSuffix(topic.ID, "_autopsy")
			corpseID := "corpse_" + strings.ToLower(name)
			if rs.Base.Stores[corpseID] <= 0 {
				continue
			}
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
		rs.Message = language.String("MSG_CANNOT_RESEARCH")
		return
	}
	if rs.Base.StartResearch(entry.topic.ID) {
		rs.Message = fmt.Sprintf(language.String("MSG_RESEARCH_STARTED"), entry.topic.DisplayName())
	} else {
		rs.Message = language.String("MSG_CANNOT_RESEARCH")
	}
}

func (rs *ResearchScreen) HandleKey(e *tcell.EventKey) {
	entries := rs.getAllTopics()
	switch e.Key() {
	case tcell.KeyUp:
		rs.Selection--
		if rs.Selection < 0 {
			rs.Selection = 0
		}
		if rs.Selection >= len(entries) {
			rs.Selection = len(entries) - 1
		}
	case tcell.KeyDown:
		rs.Selection++
		if rs.Selection >= len(entries) {
			rs.Selection = len(entries) - 1
		}
	case tcell.KeyEnter:
		if rs.InterrogateMode {
			rs.doInterrogate()
		} else {
			rs.startResearch()
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
		help := language.String("HELP_RESEARCH")
		if rs.ShowTree && !engine.Layout.IsMobile() {
			help = language.String("HELP_RESEARCH_TREE")
		}
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
			segEnd := col + engine.StringWidth(string(runes[i:end+1]))
			if x >= segStart && x <= segEnd {
				key := string(runes[i+1 : end])
				switch key {
				case "↑", "↓":
					entries := rs.getAllTopics()
					if rs.Selection < len(entries)-1 {
						rs.Selection++
					}
				case "Enter":
					rs.startResearch()
				case "T":
					if rs.ShowTree {
						rs.ShowTree = false
					} else if !engine.Layout.IsMobile() {
						rs.ShowTree = true
					}
				case "I":
					if len(rs.Base.LiveAliens) > 0 && rs.Base.TotalLabs() > 0 {
						rs.InterrogateMode = true
						rs.Message = language.String("RESEARCH_INTERROGATE_PROMPT")
					} else if len(rs.Base.LiveAliens) == 0 {
						rs.Message = language.String("MSG_INTERROGATE_NO_ALIEN")
					} else {
						rs.Message = language.String("MSG_INTERROGATE_NO_LABS")
					}
				case "Esc":
					rs.Game.PopState()
				}
				return
			}
			col = segEnd
			i = end + 1
		}
		return
	}

	if y >= rs.listStartY && y < h-2 {
		rs.Selection = y - rs.listStartY + rs.scrollOffset
		entries := rs.getAllTopics()
		if rs.Selection >= len(entries) {
			rs.Selection = len(entries) - 1
		}
		if rs.Selection < 0 {
			rs.Selection = 0
		}
	}

	if x > 0 && y >= 3 && y <= 4 {
		rs.startResearch()
	}
}
