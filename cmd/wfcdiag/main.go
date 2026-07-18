package main

import (
	"fmt"
	"os"
	"strings"
)

type jtile struct {
	ID        int
	Name      string
	Rows      []string
	Neighbors map[string][]int
}

func main() {
	open := []int{0, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	structural := []int{1, 2, 3, 4, 5, 6, 7, 8}
	anyTile := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	nb := func(n, s, e, w []int) map[string][]int {
		return map[string][]int{"N": n, "E": e, "S": s, "W": w}
	}
	allAny := func() map[string][]int { return nb(anyTile, anyTile, anyTile, anyTile) }
	allOpen := func() map[string][]int { return nb(open, open, open, open) }

	tiles := []jtile{
		{ID: 0, Name: "Floor", Rows: []string{"...", "...", "..."}, Neighbors: allOpen()},
		{ID: 1, Name: "WallN", Rows: []string{"###", "...", "..."}, Neighbors: nb(structural, open, anyTile, anyTile)},
		{ID: 2, Name: "WallE", Rows: []string{"..#", "..#", "..#"}, Neighbors: nb(anyTile, anyTile, structural, open)},
		{ID: 3, Name: "WallS", Rows: []string{"...", "...", "###"}, Neighbors: nb(open, structural, anyTile, anyTile)},
		{ID: 4, Name: "WallW", Rows: []string{"#..", "#..", "#.."}, Neighbors: nb(anyTile, anyTile, open, structural)},
		{ID: 5, Name: "CornerNE", Rows: []string{"###", "#..", "..."}, Neighbors: allAny()},
		{ID: 6, Name: "CornerSE", Rows: []string{"...", "#..", "###"}, Neighbors: allAny()},
		{ID: 7, Name: "CornerSW", Rows: []string{"...", "..#", "###"}, Neighbors: allAny()},
		{ID: 8, Name: "CornerNW", Rows: []string{"###", "..#", "..."}, Neighbors: allAny()},
		{ID: 9, Name: "ConsoleRoom", Rows: []string{".C.", "CCC", "..."}, Neighbors: allOpen()},
		{ID: 10, Name: "ConsoleRoom90", Rows: []string{".C.", ".C.", ".C."}, Neighbors: allOpen()},
		{ID: 11, Name: "MachineryNW", Rows: []string{"M..", ".#.", "..."}, Neighbors: allOpen()},
		{ID: 12, Name: "MachinerySE", Rows: []string{"...", ".#.", "..M"}, Neighbors: allOpen()},
		{ID: 13, Name: "PodRoom", Rows: []string{"PPP", "...", "PPP"}, Neighbors: allOpen()},
		{ID: 14, Name: "PowerRoom", Rows: []string{".S.", "SSS", ".#."}, Neighbors: allOpen()},
		{ID: 15, Name: "AlienTechRoom", Rows: []string{"T.T", "...", "T.T"}, Neighbors: allOpen()},
		{ID: 16, Name: "StorageRoom", Rows: []string{"S..", "...", "..S"}, Neighbors: allOpen()},
		{ID: 17, Name: "CorridorT", Rows: []string{"###", ".#.", ".#."}, Neighbors: nb(structural, open, anyTile, anyTile)},
		{ID: 18, Name: "CorridorB", Rows: []string{".#.", ".#.", "###"}, Neighbors: nb(open, structural, anyTile, anyTile)},
		{ID: 19, Name: "CorridorL", Rows: []string{".#.", "###", ".#."}, Neighbors: nb(anyTile, anyTile, open, anyTile)},
		{ID: 20, Name: "CorridorR", Rows: []string{"#..", "#..", "#.."}, Neighbors: nb(anyTile, anyTile, anyTile, open)},
	}

	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString("  \"tiles\": [\n")
	for i, t := range tiles {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"id\": %d,\n", t.ID))
		sb.WriteString(fmt.Sprintf("      \"name\": %q,\n", t.Name))
		sb.WriteString("      \"rows\": [")
		for j, r := range t.Rows {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", r))
		}
		sb.WriteString("],\n")
		sb.WriteString("      \"neighbors\": {")
		dirs := []string{"N", "E", "S", "W"}
		for di, d := range dirs {
			if di > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q: [", d))
			arr := t.Neighbors[d]
			for j, v := range arr {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%d", v))
			}
			sb.WriteString("]")
		}
		sb.WriteString("}\n")
		end := "    }"
		if i < len(tiles)-1 {
			end += ","
		}
		sb.WriteString(end + "\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	os.WriteFile("data/wfc/alien_base.json", []byte(sb.String()), 0644)
	fmt.Println("wrote data/wfc/alien_base.json")
}
