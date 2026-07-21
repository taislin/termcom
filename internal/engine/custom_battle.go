package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/language"
	"github.com/taislin/termcom/internal/mapgen"
)

const (
	nameTruncPad         = 7
	scrollThresholdOff   = 6
	statusBarH           = 3
	listStartRow         = 3
	dividerEndOffset     = 3
	leftTextPad          = 2
)

type CustomBattleEntry struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Night       bool   `json:"night"`
	FilePath    string
}

type CustomBattleScreen struct {
	Game      *Game
	Entries   []CustomBattleEntry
	Selection int
	OnPick    func(entry CustomBattleEntry)
}

func NewCustomBattleScreen(g *Game, onSelect func(CustomBattleEntry)) *CustomBattleScreen {
	entries := scanCustomMaps()
	return &CustomBattleScreen{
		Game:    g,
		Entries: entries,
		OnPick:  onSelect,
	}
}

func scanCustomMaps() []CustomBattleEntry {
	mapsDir := "maps"
	if _, err := os.Stat(mapsDir); os.IsNotExist(err) {
		return nil
	}
	dirEntries, err := os.ReadDir(mapsDir)
	if err != nil {
		return nil
	}
	var result []CustomBattleEntry
	for _, e := range dirEntries {
		if e.IsDir() || !mapgen.IsJSONFile(e.Name()) {
			continue
		}
		path := filepath.Join(mapsDir, e.Name())
		data, err := mapgen.ReadFileJSONC(path)
		if err != nil {
			continue
		}
		var entry CustomBattleEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		entry.FilePath = path
		if entry.Name == "" {
			entry.Name = e.Name()
		}
		result = append(result, entry)
	}
	return result
}

func (cs *CustomBattleScreen) Update() {}

func (cs *CustomBattleScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	ctx.DrawPanel(0, 0, w, h, language.String("MENU_CUSTOM_BATTLE"), StyleDefault)

	if len(cs.Entries) == 0 {
		ctx.DrawString(2, 3, language.String("CUSTOM_NO_BATTLES"), StyleGray)
		ctx.DrawString(2, 5, language.String("CUSTOM_PLACE_JSON"), StyleGray)
		ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
		ctx.DrawMarkupString(1, h-1, language.String("CUSTOM_HELP"), StyleGray, StyleHotkey)
		return
	}

	// Split panel: left list, right details
	leftW := Layout.CustomBattleLeftWidth(w)
	rightX := Layout.CustomBattleRightX(w)

	ctx.DrawString(2, 2, language.String("CUSTOM_MISSION_SELECT"), StyleCyanBold)
	// Vertical divider
	for y := listStartRow; y < h-dividerEndOffset; y++ {
		ctx.SetCell(leftW, y, '│', StyleGray)
	}

	// Left panel: list
	maxVisible := h - scrollThresholdOff
	scrollOffset := cs.scrollOffset(maxVisible)

	for i := 0; i < maxVisible && i+scrollOffset < len(cs.Entries); i++ {
		idx := i + scrollOffset
		entry := cs.Entries[idx]
		y := listStartRow + i
		style := StyleDefault
		if idx == cs.Selection {
			style = StyleHighlight
		}
		truncW := leftW - nameTruncPad
		label := entry.Name
		if StringWidth(label) > leftW-4 {
			runes := []rune(label)
			for len(runes) > 0 && StringWidth(string(runes)) > truncW {
				runes = runes[:len(runes)-1]
			}
			label = string(runes) + "..."
		}
		ctx.DrawString(leftTextPad, y, label, style)
	}

	// Right panel: details of selected
	if cs.Selection >= 0 && cs.Selection < len(cs.Entries) {
		entry := cs.Entries[cs.Selection]
		ry := listStartRow

		ctx.DrawString(rightX, ry, entry.Name, StyleCyanBold)
		ry++

		if entry.Author != "" {
			ctx.DrawString(rightX, ry, fmt.Sprintf(language.String("CUSTOM_AUTHOR"), entry.Author), StyleGray)
			ry++
		}
		if entry.Date != "" {
			ctx.DrawString(rightX, ry, fmt.Sprintf(language.String("CUSTOM_DATE"), entry.Date), StyleGray)
			ry++
		}
		ry++

		// Word-wrap description
		if entry.Description != "" {
			ctx.DrawString(rightX, ry, language.String("CUSTOM_DESC"), StyleGray)
			ry++
			words := strings.Fields(entry.Description)
			line := ""
			for _, word := range words {
				if StringWidth(line)+StringWidth(word)+1 > w-rightX-2 {
					ctx.DrawString(rightX+2, ry, line, StyleDefault)
					ry++
					line = word
				} else {
					if line != "" {
						line += " "
					}
					line += word
				}
			}
			if line != "" {
				ctx.DrawString(rightX+2, ry, line, StyleDefault)
				ry++
			}
		}
		ry++

		if entry.Night {
			ctx.DrawString(rightX, ry, language.String("CUSTOM_TIME_NIGHT"), StyleBlue)
		} else {
			ctx.DrawString(rightX, ry, language.String("CUSTOM_TIME_DAY"), StyleDefault)
		}
		ry++

		ctx.DrawString(rightX, ry, fmt.Sprintf(language.String("CUSTOM_FILE"), filepath.Base(entry.FilePath)), StyleGray)
	}

	// Status bar
	ctx.DrawPanel(0, h-statusBarH, w, statusBarH, "", StyleGray)
	ctx.DrawMarkupString(1, h-2, language.String("CUSTOM_HELP"), StyleGray, StyleHotkey)
}

func (cs *CustomBattleScreen) scrollOffset(maxVisible int) int {
	if cs.Selection >= maxVisible {
		return cs.Selection - maxVisible + 1
	}
	return 0
}

func (cs *CustomBattleScreen) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		cs.Selection--
		if cs.Selection < 0 {
			cs.Selection = 0
		}
	case tcell.KeyDown:
		cs.Selection++
		if cs.Selection >= len(cs.Entries) {
			cs.Selection = len(cs.Entries) - 1
		}
	case tcell.KeyEnter:
		cs.confirm()
	case tcell.KeyEscape:
		cs.Game.PopState()
	}
	switch e.Str() {
	case "q", "Q":
		cs.Game.PopState()
	case "j", "J":
		cs.Selection++
		if cs.Selection >= len(cs.Entries) {
			cs.Selection = len(cs.Entries) - 1
		}
	case "k", "K":
		cs.Selection--
		if cs.Selection < 0 {
			cs.Selection = 0
		}
	}
}

func (cs *CustomBattleScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := cs.Game.ScreenSize()

	if y == h-1 {
		switch {
		case x >= 1 && x <= 3:
			cs.Game.PopState()
		case x >= 5 && x <= 10:
			cs.confirm()
		}
		return
	}

	leftW := Layout.CustomBattleLeftWidth(w)
	if x < leftW {
		// Click in left panel
		clickIdx := y - listStartRow
		scrollOffset := 0
		if cs.Selection >= h-scrollThresholdOff {
			scrollOffset = cs.Selection - (h - scrollThresholdOff) + 1
		}
		clickIdx += scrollOffset
		if clickIdx >= 0 && clickIdx < len(cs.Entries) {
			cs.Selection = clickIdx
			if buttons&tcell.Button1 != 0 {
				cs.confirm()
			}
		}
	}
}

func (cs *CustomBattleScreen) confirm() {
	if cs.Selection >= 0 && cs.Selection < len(cs.Entries) && cs.OnPick != nil {
		cs.OnPick(cs.Entries[cs.Selection])
	}
}
