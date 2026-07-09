package engine

import (
	"fmt"
	"strings"

	"github.com/civ13/ycom/internal/data"
	"github.com/civ13/ycom/internal/language"
	"github.com/gdamore/tcell/v3"
)

type EncycloEntry struct {
	Category   string
	ID         string
	Name       string
	Desc       string
	Discovered bool
	AlienType  *data.AlienType
}

type EncyclopediaScreen struct {
	Game      *Game
	Entries   []EncycloEntry
	Selection int
	Page      int
	Tab       int
}

func NewEncyclopediaScreen(g *Game, completedResearch []string, unlockedWeapons []string, unlockedArmor []string) *EncyclopediaScreen {
	es := &EncyclopediaScreen{
		Game: g,
	}
	es.buildEntries(g, completedResearch, unlockedWeapons, unlockedArmor)
	return es
}

func (es *EncyclopediaScreen) buildEntries(g *Game, completed []string, weapons []string, armor []string) {
	completedMap := make(map[string]bool)
	for _, r := range completed {
		completedMap[r] = true
	}
	weaponMap := make(map[string]bool)
	for _, w := range weapons {
		weaponMap[w] = true
	}
	armorMap := make(map[string]bool)
	for _, a := range armor {
		armorMap[a] = true
	}

	knowledgeMap := g.AlienKnowledge

	for _, at := range g.GetAlienTypes() {
		level := knowledgeMap[at.Name]
		discovered := level >= 2
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   "Aliens",
			ID:         at.ShortName,
			Name:       at.Name,
			Desc:       at.Lore,
			Discovered: discovered,
			AlienType:  at,
		})
	}

	for key, item := range data.RuleItems {
		if item.IsAlien {
			es.Entries = append(es.Entries, EncycloEntry{
				Category:   "Weapons",
				ID:         key,
				Name:       item.Name,
				Desc:       fmt.Sprintf("DMG:%d ACC:%d%% TU:%d RNG:%d", item.Damage, item.Accuracy, item.TU, item.Range),
				Discovered: weaponMap[key],
			})
		}
	}

	for key, arm := range data.Armors {
		if key == "none" {
			continue
		}
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   "Armor",
			ID:         key,
			Name:       arm.Name,
			Desc:       fmt.Sprintf("Protection: %d", arm.Undersuit),
			Discovered: armorMap[key],
		})
	}

	for _, item := range data.Items {
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   "Items",
			ID:         item.ShortName,
			Name:       item.Name,
			Desc:       fmt.Sprintf("Weight:%d Value:$%d", item.Weight, item.Value),
			Discovered: item.Alien,
		})
	}

	for _, topic := range data.ResearchTree {
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   "Research",
			ID:         topic.ID,
			Name:       topic.Name,
			Desc:       fmt.Sprintf("Cost: %d man-days", topic.Cost),
			Discovered: completedMap[topic.ID],
		})
	}
}

var encTabs = []string{"Aliens", "Weapons", "Armor", "Items", "Research"}

func (es *EncyclopediaScreen) Update() {}

func (es *EncyclopediaScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	ctx.DrawPanel(0, 0, w, 3, language.String("ENCYCLOPEDIA"), StyleCyanBold)

	tabY := 3
	tx := 2
	for i, tab := range encTabs {
		style := StyleDefault
		if i == es.Tab {
			style = StyleHighlight
		}
		ctx.DrawString(tx, tabY, tab, style)
		tx += len(tab) + 3
	}

	listX := 1
	listY := 5
	listW := w / 3
	if listW < 20 {
		listW = 20
	}
	listH := h - 8

	infoX := listW + 3
	infoW := w - infoX - 2

	tabEntries := es.filteredEntries()

	for y := listY; y < listY+listH && y-listY < len(tabEntries); y++ {
		idx := y - listY + es.Page
		if idx >= len(tabEntries) {
			break
		}
		e := tabEntries[idx]
		style := StyleDefault
		if !e.Discovered {
			style = StyleGray
			e.Name = "???"
		}
		if idx == es.Selection {
			style = StyleHighlight
		}
		name := e.Name
		if len(name) > listW-2 {
			name = name[:listW-2]
		}
		ctx.DrawString(listX, y, name, style)
	}

	if es.Selection >= 0 && es.Selection < len(tabEntries) {
		e := tabEntries[es.Selection]
		if e.Discovered {
			ctx.DrawPanel(infoX, listY, infoW, 4, e.Name, StyleCyanBold)
			desc := e.Desc
			line := listY + 2
			for len(desc) > 0 {
				end := infoW - 2
				if end > len(desc) {
					end = len(desc)
				}
				ctx.DrawString(infoX+1, line, desc[:end], StyleDefault)
				desc = desc[end:]
				line++
			}

			if e.Category == "Aliens" && e.AlienType != nil {
				at := e.AlienType
				portrait := at.GetPortrait()
				pLines := strings.Split(portrait, "\n")
				pX := infoX + infoW - len(pLines[0]) - 2
				pY := listY + 5

				ctx.DrawString(pX, pY-1, at.Name, StyleRedBold)
				pStyle := StyleYellow
				for i, pl := range pLines {
					ctx.DrawString(pX, pY+i, pl, pStyle)
				}

				statY := pY + len(pLines) + 1
				ctx.DrawString(infoX+1, statY, fmt.Sprintf("HP:%d TU:%d ACC:%d", at.HP, at.TU, at.Accuracy), StyleGray)
				ctx.DrawString(infoX+1, statY+1, fmt.Sprintf("STR:%d PSI:%d BRA:%d", at.Strength, at.Psi, at.Bravery), StyleGray)
				ctx.DrawString(infoX+1, statY+2, fmt.Sprintf("DMG:%s WPN:%s", data.DamageTypeStr(at.DamageType), data.RuleItems[at.Weapon].Name), StyleGray)
			}
		} else {
			ctx.DrawString(infoX+1, listY+2, "Not yet discovered.", StyleGray)
		}
	}

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawString(1, h-1, language.String("HELP_ENCYCLOPEDIA"), StyleGray)
}

func (es *EncyclopediaScreen) filteredEntries() []EncycloEntry {
	tab := encTabs[es.Tab]
	var result []EncycloEntry
	for _, e := range es.Entries {
		if e.Category == tab {
			result = append(result, e)
		}
	}
	return result
}

func (es *EncyclopediaScreen) HandleKey(e *tcell.EventKey) {
	entries := es.filteredEntries()
	switch e.Key() {
	case tcell.KeyUp:
		if es.Selection > 0 {
			es.Selection--
		}
	case tcell.KeyDown:
		if es.Selection < len(entries)-1 {
			es.Selection++
		}
	case tcell.KeyLeft:
		if es.Tab > 0 {
			es.Tab--
			es.Selection = 0
			es.Page = 0
		}
	case tcell.KeyRight:
		if es.Tab < len(encTabs)-1 {
			es.Tab++
			es.Selection = 0
			es.Page = 0
		}
	}
}

func (es *EncyclopediaScreen) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	_, h := es.Game.ScreenSize()

	// Handle help bar clicks (bottom bar)
	if y == h-1 {
		// Help bar: "h/l=Tab  [j]/[k]=Navigate  [Esc]=Back"
		switch {
		case x >= 1 && x <= 3: // h/l=Tab
			// Previous tab
			if es.Tab > 0 {
				es.Tab--
				es.Selection = 0
			}
		case x >= 5 && x <= 10: // [j]/[k]=Navigate
			// Scroll down
			if es.Selection < len(es.Entries)-1 {
				es.Selection++
			}
		case x >= 12 && x <= 18: // [Esc]=Back
			es.Game.PopState()
		}
		return
	}
}
