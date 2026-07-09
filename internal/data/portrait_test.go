package data

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrintPortraits(t *testing.T) {
	species, _ := GenerateSpecies(42)

	fmt.Println("=== PROCEDURAL ALIEN PORTRAITS ===")
	fmt.Println()

	for _, sp := range species {
		fmt.Printf("Species: %s  (DMG: %s)\n", sp.Name, DamageTypeStr(sp.PrimaryDMG))
		fmt.Println(strings.Repeat("-", 40))
		for _, at := range sp.Types {
			portrait := at.GetPortrait()
			lines := strings.Split(portrait, "\n")
			fmt.Printf("  %-20s Rank %d  (HP:%d ACC:%d PSI:%d)\n", at.Name, at.Rank, at.HP, at.Accuracy, at.Psi)
			for _, line := range lines {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println()
		}
	}
}
