package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/civ13/termcom/internal/data"
)

const (
	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiDim       = "\033[2m"
	ansiUnderline = "\033[4m"

	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiWhite   = "\033[37m"
	ansiGray    = "\033[90m"

	ansiBgRed   = "\033[41m"
	ansiBgGreen = "\033[42m"
)

var damageColors = []string{
	ansiMagenta, // Plasma
	ansiCyan,    // Laser
	ansiYellow,  // Explosive
	ansiRed,     // Melee
	ansiWhite,   // Kinetic
	ansiBlue,    // Psionic
}

var rankNames = []string{
	"Rank 0 (Rookie)",
	"Rank 1 (Veteran)",
	"Rank 2 (Elite)",
	"Rank 3 (Commander)",
	"Rank 4 (Overlord)",
	"Rank 5 (Ancient)",
	"Rank 6 (Godlike)",
}

func rgbFg(r, g, b int32) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func rgbBg(r, g, b int32) string {
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

func renderPortraitHalfBlock(sp data.StyledPortrait) {
	if len(sp.Lines) == 0 {
		return
	}

	lines := len(sp.Lines)
	// Pad all lines to 7 runes
	content := make([][]rune, lines)
	for i, l := range sp.Lines {
		r := []rune(l.Content)
		for len(r) < 7 {
			r = append(r, ' ')
		}
		content[i] = r[:7]
	}

	// Render pairs of rows using half-block characters
	for row := 0; row < lines; row += 2 {
		// Top line uses content[row]'s color, bottom uses content[row+1]'s color (or same)
		topColor := sp.Lines[row].Color
		botColor := topColor
		if row+1 < lines {
			botColor = sp.Lines[row+1].Color
		}

		topChars := content[row]
		botChars := content[row]
		if row+1 < lines {
			botChars = content[row+1]
		}

		fmt.Print("    ")
		for col := 0; col < 7; col++ {
			topR := topChars[col]
			botR := botChars[col]
			topFill := runeDensity(topR)
			botFill := runeDensity(botR)

			// Top pixel color
			tr, tg, tb := topColor[0], topColor[1], topColor[2]
			// Bottom pixel color
			br, bg, bb := botColor[0], botColor[1], botColor[2]

			// Draw top half-block (▀) with top color as FG and bottom color as BG
			// Only show character if at least one pixel is filled
			if topFill == 0 && botFill == 0 {
				fmt.Print("  ")
			} else {
				// Set BG to bottom color, FG to top color
				fmt.Printf("%s%s%s%s", rgbBg(br, bg, bb), rgbFg(tr, tg, tb), "\u2580", ansiReset)
			}
		}
		fmt.Println()
	}
}

func runeDensity(r rune) int {
	switch r {
	case ' ', 0:
		return 0
	case '.', '·', '°', '*':
		return 1
	case '|', '-', '/', '\\', '+', '¤', '~', '†', 'o':
		return 2
	default:
		return 3
	}
}

func resistStr(val int) string {
	if val > 0 {
		return fmt.Sprintf("%s+%d%%%s", ansiGreen, val, ansiReset)
	} else if val < 0 {
		return fmt.Sprintf("%s%d%%%s", ansiRed, val, ansiReset)
	}
	return fmt.Sprintf("%s0%%%s", ansiGray, ansiReset)
}

func printDivider() {
	fmt.Printf("%s%s%s\n", ansiGray, strings.Repeat("─", 72), ansiReset)
}

func printStat(label, value string) {
	fmt.Printf("  %s%-14s%s %s\n", ansiDim, label, ansiReset, value)
}

func main() {
	seed := int64(42)
	if len(os.Args) > 1 {
		var s int64
		if _, err := fmt.Sscanf(os.Args[1], "%d", &s); err == nil {
			seed = s
		}
	}

	// Generate procedural species (same as normal game)
	_, allTypes := data.GenerateSpecies(seed)

	fmt.Println()
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%s%s  TERMCOM — Alien Roster Viewer (seed: %d)%s\n", ansiBold, ansiCyan, seed, ansiReset)
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("  %s%s aliens generated across %s procedural species%s\n",
		ansiBold, fmt.Sprintf("%d", len(allTypes)), ansiYellow, ansiReset)
	fmt.Println()

	// Group by species prefix
	speciesOrder := make([]string, 0)
	speciesMap := make(map[string][]*data.AlienType)
	for _, at := range allTypes {
		prefix := at.ShortName[:min(3, len(at.ShortName))]
		if _, exists := speciesMap[prefix]; !exists {
			speciesOrder = append(speciesOrder, prefix)
		}
		speciesMap[prefix] = append(speciesMap[prefix], at)
	}

	alienIdx := 0
	for _, prefix := range speciesOrder {
		types := speciesMap[prefix]
		first := types[0]

		dmgColor := damageColors[first.DamageType]

		fmt.Printf("%s%s╔══ %s %s (%s) %s══╗%s\n",
			ansiBold, dmgColor,
			ansiWhite, first.Name, data.DamageTypeStr(first.DamageType),
			dmgColor, ansiReset)
		fmt.Printf("%s%s║  Icon: %c   Species: %s%s\n",
			dmgColor, ansiReset,
			first.Icon, ansiBold, prefix)
		fmt.Printf("%s%s║%s\n", dmgColor, ansiReset, ansiReset)

		for _, at := range types {
			alienIdx++
			rankLabel := ""
			if at.Rank < len(rankNames) {
				rankLabel = rankNames[at.Rank]
			} else {
				rankLabel = fmt.Sprintf("Rank %d", at.Rank)
			}

			fmt.Printf("%s  ┌─ %s#%02d %s [%c] %s─┐%s\n",
				dmgColor, ansiBold, alienIdx, at.Name, at.Icon, dmgColor, ansiReset)

			// Portrait
			portrait := at.GetPortrait()
			fmt.Println()
			renderPortraitHalfBlock(portrait)
			fmt.Println()

			// Stats block
			fmt.Printf("  %s%sStats:%s\n", ansiBold, ansiWhite, ansiReset)
			printStat("Rank:", rankLabel)
			printStat("HP:", fmt.Sprintf("%s%d%s", ansiBold, at.HP, ansiReset))
			printStat("TU:", fmt.Sprintf("%d", at.TU))
			printStat("Accuracy:", fmt.Sprintf("%s%d%%%s", ansiCyan, at.Accuracy, ansiReset))
			printStat("Bravery:", fmt.Sprintf("%d", at.Bravery))
			printStat("Reactions:", fmt.Sprintf("%d", at.Reactions))
			printStat("Strength:", fmt.Sprintf("%d", at.Strength))
			printStat("Psi:", fmt.Sprintf("%s%d%s", ansiBlue, at.Psi, ansiReset))
			printStat("Armour:", fmt.Sprintf("%s%d%s", ansiWhite, at.Armour, ansiReset))
			printStat("Aggression:", fmt.Sprintf("%d", at.Aggression))
			printStat("Kill XP:", fmt.Sprintf("%s%d%s", ansiYellow, at.Points, ansiReset))
			printStat("Weapon:", fmt.Sprintf("%s%s%s", ansiGreen, at.Weapon, ansiReset))

			fmt.Println()

			// Resistances
			fmt.Printf("  %s%sResistances:%s ", ansiBold, ansiWhite, ansiReset)
			fmt.Printf("Pla:%s ", resistStr(at.ResistPlasma))
			fmt.Printf("Las:%s ", resistStr(at.ResistLaser))
			fmt.Printf("Exp:%s ", resistStr(at.ResistExplosive))
			fmt.Printf("Mle:%s ", resistStr(at.ResistMelee))
			fmt.Printf("Kin:%s ", resistStr(at.ResistKinetic))
			fmt.Printf("Psi:%s", resistStr(at.ResistPsionic))
			fmt.Println()

			// Morphology
			if at.Morphology != nil {
				m := at.Morphology
				fmt.Printf("  %s%sMorphology:%s %s%s%s | %d arms, %d legs",
					ansiBold, ansiWhite, ansiReset,
					ansiDim, m.BodyType, ansiReset,
					m.Arms, m.Legs)
				if m.IsFloating() {
					fmt.Printf(" %s(floating)%s", ansiCyan, ansiReset)
				}
				if m.MultiArmed() {
					fmt.Printf(" %s(multi-armed)%s", ansiMagenta, ansiReset)
				}
				if m.IsLarge() {
					fmt.Printf(" %s(large)%s", ansiYellow, ansiReset)
				}
				fmt.Println()
				fmt.Printf("  %sSubtype:%s %s | Eyes: %s | Hearing: %s | Therm: %s | Psi: %s | Chem: %s%s\n",
					ansiDim, ansiReset,
					m.BodySubtype,
					m.Eyesight, m.Hearing, m.ThermalSense, m.PsionicSense, m.ChemicalSense,
					ansiReset)
			} else {
				fmt.Printf("  %s%sMorphology:%s %s(default: organic, 2 arms, 2 legs)%s\n",
					ansiDim, ansiBold, ansiReset, ansiDim, ansiReset)
			}

			// Lore
			if at.Lore != "" {
				fmt.Printf("  %s%sLore:%s %s%s%s\n",
					ansiDim, ansiBold, ansiReset, ansiDim, at.Lore, ansiReset)
			}

			if alienIdx < len(allTypes) {
				fmt.Println()
				fmt.Printf("%s%s  └──────────────────────────────────────────────────────┘%s\n", dmgColor, ansiDim, ansiReset)
				fmt.Println()
			}
		}

		printDivider()
		fmt.Println()
	}

	// Summary table
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%s%s  SUMMARY%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("%s%s══════════════════════════════════════════════════════════════════════%s\n", ansiBold, ansiCyan, ansiReset)
	fmt.Printf("  %-24s %4s %3s %4s %3s %3s %4s %3s %5s  %s\n",
		"Name", "HP", "TU", "Acc", "Str", "Psi", "Arm", "Agr", "KillXP", "Weapon")
	printDivider()

	for _, at := range allTypes {
		dmgColor := damageColors[at.DamageType]
		fmt.Printf("  %s%c%s %-20s %3d %3d %3d%% %3d %3d %4d %3d %5d  %s%s%s\n",
			dmgColor, at.Icon, ansiReset,
			at.Name,
			at.HP, at.TU, at.Accuracy,
			at.Strength, at.Psi, at.Armour, at.Aggression, at.Points,
			ansiDim, at.Weapon, ansiReset)
	}

	printDivider()
	fmt.Printf("  %s%sTotal: %d aliens | Seed: %d%s\n", ansiBold, ansiWhite, len(allTypes), seed, ansiReset)
	fmt.Println()
}
