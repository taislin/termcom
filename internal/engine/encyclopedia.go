package engine

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/language"
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
			Category:   language.String("ENCYCLO_CAT_ALIENS"),
			ID:         at.ShortName,
			Name:       at.LangName(),
			Desc:       at.Lore,
			Discovered: discovered,
			AlienType:  at,
		})
	}

	for key, item := range data.RuleItems {
		if item.IsAlien {
			es.Entries = append(es.Entries, EncycloEntry{
				Category:   language.String("ENCYCLO_CAT_WEAPONS"),
				ID:         key,
				Name:       item.DisplayName(),
				Desc:       fmt.Sprintf(language.String("ENCYCLO_WEAPON_STATS"), item.Damage, item.Accuracy, item.TU, item.Range),
				Discovered: weaponMap[key],
			})
		}
	}

	for key, arm := range data.Armors {
		if key == "none" {
			continue
		}
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   language.String("ENCYCLO_CAT_ARMOR"),
			ID:         key,
			Name:       arm.DisplayNameByKey(key),
			Desc:       fmt.Sprintf(language.String("ENCYCLO_ARMOR_PROTECTION"), arm.Undersuit),
			Discovered: armorMap[key],
		})
	}

	for key, item := range data.Items {
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   language.String("ENCYCLO_CAT_ITEMS"),
			ID:         item.ShortName,
			Name:       data.ItemDisplayName(key),
			Desc:       fmt.Sprintf(language.String("ENCYCLO_ITEM_STATS"), item.Weight, item.Value),
			Discovered: item.Alien,
		})
	}

	for _, topic := range data.ResearchTree {
		es.Entries = append(es.Entries, EncycloEntry{
			Category:   language.String("ENCYCLO_CAT_RESEARCH"),
			ID:         topic.ID,
			Name:       topic.DisplayName(),
			Desc:       fmt.Sprintf(language.String("ENCYCLO_RESEARCH_COST"), topic.Cost),
			Discovered: completedMap[topic.ID],
		})
	}
}

func encTabs() []string {
	return []string{
		language.String("ENCYCLO_CAT_ALIENS"),
		language.String("ENCYCLO_CAT_WEAPONS"),
		language.String("ENCYCLO_CAT_ARMOR"),
		language.String("ENCYCLO_CAT_ITEMS"),
		language.String("ENCYCLO_CAT_RESEARCH"),
	}
}

func (es *EncyclopediaScreen) Update() {}

func (es *EncyclopediaScreen) Render(ctx *ScreenCtx) {
	w, h := ctx.Size()

	ctx.DrawPanel(0, 0, w, 3, language.String("ENCYCLOPEDIA"), StyleCyanBold)

	tabY := 3
	tx := 2
	tabs := encTabs()
	for i, tab := range tabs {
		style := StyleDefault
		if i == es.Tab {
			style = StyleHighlight
		}
		ctx.DrawString(tx, tabY, tab, style)
		tx += StringWidth(tab) + 3
	}

	listX := 1
	listY := 5
	listW := Layout.EncyclopediaListWidth(w)
	listH := h - 8

	infoX := Layout.EncyclopediaInfoX(w)
	infoW := Layout.EncyclopediaInfoWidth(w)

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
			e.Name = language.String("ENCYCLO_UNDISCOVERED")
		}
		if idx == es.Selection {
			style = StyleHighlight
		}
		name := e.Name
		if StringWidth(name) > listW-2 {
			runes := []rune(name)
			for len(runes) > 0 && StringWidth(string(runes)) > listW-2 {
				runes = runes[:len(runes)-1]
			}
			name = string(runes)
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

			if e.Category == language.String("ENCYCLO_CAT_ALIENS") && e.AlienType != nil {
				at := e.AlienType
				bgColor := tcell.NewRGBColor(20, 20, 28)
				alienImg := GenerateAlienSpriteFromSeed(int64(at.Icon), at.Morphology, bgColor)
				pX := infoX + infoW - alienImg.Width - 4
				pY := listY + 2

				ctx.DrawPixelImageFramed(pX, pY, alienImg, StyleRed)

				statY := pY + alienImg.Height/2 + 1
				ctx.DrawString(infoX+1, pY-1, at.LangName(), StyleRedBold)
				ctx.DrawString(infoX+1, statY, fmt.Sprintf(language.String("ENCYCLO_ALIEN_STATS_1"), at.HP, at.TU, at.Accuracy), StyleGray)
				ctx.DrawString(infoX+1, statY+1, fmt.Sprintf(language.String("ENCYCLO_ALIEN_STATS_2"), at.Strength, at.Psi, at.Bravery), StyleGray)
				weapName := "---"
			if at.Weapon != "" {
				if w, ok := data.RuleItems[at.Weapon]; ok {
					weapName = w.DisplayName()
				}
			}
			ctx.DrawString(infoX+1, statY+2, fmt.Sprintf(language.String("ENCYCLO_ALIEN_STATS_3"), data.DamageTypeStr(at.DamageType), weapName), StyleGray)

				if m := at.Morphology; m != nil {
					ctx.DrawString(infoX+1, statY+4, language.Sprintf("ENCYCLO_MORPH_BODY", m.BodyType, m.BodySubtype, m.Arms, m.Legs), StyleGray)
					ctx.DrawString(infoX+1, statY+5, language.Sprintf("ENCYCLO_MORPH_SENSES", m.Eyesight, m.Hearing, m.ThermalSense), StyleGray)
					ctx.DrawString(infoX+1, statY+6, language.Sprintf("ENCYCLO_MORPH_PSI", m.PsionicSense, m.ChemicalSense), StyleGray)
				}
			}
		} else {
			ctx.DrawString(infoX+1, listY+2, language.String("ENCYCLO_NOT_DISCOVERED"), StyleGray)
		}
	}

	ctx.DrawPanel(0, h-1, w, 1, "", StyleGray)
	ctx.DrawString(1, h-1, language.String("HELP_ENCYCLOPEDIA"), StyleGray)
}

func (es *EncyclopediaScreen) filteredEntries() []EncycloEntry {
	tab := encTabs()[es.Tab]
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
		if es.Selection < es.Page {
			es.Page--
		}
	case tcell.KeyDown:
		if es.Selection < len(entries)-1 {
			es.Selection++
		}
		_, h := es.Game.ScreenSize()
		listH := h - 8
		if es.Selection >= es.Page+listH {
			es.Page++
		}
	case tcell.KeyLeft:
		if es.Tab > 0 {
			es.Tab--
			es.Selection = 0
			es.Page = 0
		}
	case tcell.KeyRight:
		if es.Tab < len(encTabs())-1 {
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
		// Help bar: "[←]/[→]=Tab  [↑]/[↓]=Navigate  [Esc]=Back"
		switch {
		case x >= 1 && x <= 3: // [←]/[→]=Tab
			// Previous tab
			if es.Tab > 0 {
				es.Tab--
				es.Selection = 0
			}
		case x >= 5 && x <= 10: // [↑]/[↓]=Navigate
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
