package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/taislin/termcom/internal/battle"
	_ "github.com/taislin/termcom/internal/engine"
	"github.com/taislin/termcom/internal/mapgen"
)

type genEntry struct {
	name string
	desc string
}

var generators = []genEntry{
	{"crash", "Random crash site (default)"},
	{"terror", "Terror site (urban)"},
	{"abduction", "Abduction site"},
	{"ufo_interior", "Hand-crafted UFO interior"},
	{"ufo_wfc", "WFC-generated UFO interior (2 levels)"},
	{"alien_base", "Hand-crafted alien base"},
	{"alien_base_wfc", "WFC-generated alien base (2 levels)"},
	{"building", "WFC urban building (1 level)"},
	{"building2", "WFC urban building (2 levels)"},
	{"cydonia", "Cydonia final mission"},
	{"urban", "Urban biome (AssembleMap)"},
	{"forest", "Forest biome (AssembleMap)"},
	{"desert", "Desert biome (AssembleMap)"},
	{"polar", "Polar biome (AssembleMap)"},
	{"rural", "Rural biome (AssembleMap)"},
	{"ufo", "UFO biome (AssembleMap)"},
	{"alien", "Alien biome (AssembleMap)"},
	{"farm", "Farm biome (AssembleMap)"},
	{"coastal", "Coastal biome (AssembleMap)"},
	{"mountain", "Mountain biome (AssembleMap)"},
	{"swamp", "Swamp biome (AssembleMap)"},
	{"jungle", "Jungle biome (AssembleMap)"},
	{"all", "All biomes (AssembleMap) — cycle with 'b' key"},
}

var tileTypeNames = map[battle.TileType]string{
	battle.TileFloor:       "Floor",
	battle.TileWall:        "Wall",
	battle.TileDoor:        "Door",
	battle.TileWindow:      "Window",
	battle.TileGrass:       "Grass",
	battle.TileTree:        "Tree",
	battle.TileRock:        "Rock",
	battle.TileWater:       "Water",
	battle.TileUFOFloor:    "UFO Floor",
	battle.TileUFOWall:     "UFO Wall",
	battle.TileStairs:      "Stairs Up",
	battle.TileStairsDown:  "Stairs Down",
	battle.TilePavement:    "Pavement",
	battle.TileSand:        "Sand",
	battle.TileSnow:        "Snow",
	battle.TileMarsh:       "Marsh",
	battle.TileBush:        "Bush",
	battle.TileFence:       "Fence",
	battle.TileRubble:      "Rubble",
	battle.TileObject:      "Object",
	battle.TileConsole:     "Console",
	battle.TileMachinery:   "Machinery",
	battle.TilePod:         "Pod",
	battle.TilePowerSource: "Power Source",
	battle.TileStorage:     "Storage",
	battle.TileAlienTech:   "Alien Tech",
	battle.TileDesk:        "Desk",
	battle.TileChair:       "Chair",
	battle.TileComputer:    "Computer",
	battle.TileBed:         "Bed",
	battle.TileLocker:      "Locker",
	battle.TileCabinet:     "Cabinet",
	battle.TileCar:         "Car (L)",
	battle.TileCarMid:      "Car (Mid)",
	battle.TileCarRight:    "Car (R)",
	battle.TileForklift:      "Forklift (L)",
	battle.TileForkliftRight: "Forklift (R)",
	battle.TileFuelPump:       "Fuel Pump",
	battle.TileContainerRed:    "Container (R)",
	battle.TileContainerBlue:   "Container (B)",
	battle.TileContainerYellow: "Container (Y)",
	battle.TileAdobe:         "Adobe Wall",
	battle.TileMetalWall:     "Metal Wall",
	battle.TileWreck:         "Aircraft Wreck",
	battle.TileTimber:        "Timber Stack",
	battle.TileDish:          "Satellite Dish",
	battle.TileTruck:         "Supply Truck",
	battle.TileIce:           "Frozen Ice",
	battle.TileStreetlamp:    "Streetlamp",
	battle.TileGlass:         "Broken Glass",
	battle.TileDebris:        "Debris",
	battle.TileCryoPipe:      "Cryo Pipe",
	battle.TileSkylight:      "Skylight",
	battle.TileWheat:         "Wheat",
	battle.TileHayBale:       "Hay Bale",
	battle.TilePier:          "Pier",
	battle.TileDockCrate:     "Dock Crate",
	battle.TileCliffFace:     "Cliff Face",
	battle.TileScree:         "Scree",
	battle.TileBoulder:       "Boulder",
	battle.TileSwampWater:    "Swamp Water",
	battle.TileCypressTree:   "Cypress Tree",
	battle.TileMud:           "Mud",
	battle.TileVine:          "Vine",
	battle.TileBamboo:        "Bamboo",
	battle.TileHeloBody:      "Heli Body",
	battle.TileHeloTail:      "Heli Tail",
	battle.TileHeloNose:      "Heli Nose",
	battle.TileHeloRotor:     "Heli Rotor",
}

const infoPanelWidth = 22

func main() {
	gen, w, h, seed, dump := parseArgs()
	if gen == "" {
		printUsage()
		return
	}
	if gen == "--list" {
		printGenerators()
		return
	}

	if err := mapgen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "mapgen init: %v\n", err)
		os.Exit(1)
	}

	rng := rand.New(rand.NewSource(seed))

	allBiomes := []string{"urban", "forest", "desert", "polar", "rural", "ufo", "alien", "farm", "coastal", "mountain", "swamp", "jungle"}

	var m *battle.BattleMap
	label := gen
	biomeIdx := 0

	switch gen {
	case "crash":
		m, _ = battle.GenerateCrashSite(w, h, seed, -1, -1)
		label = "Crash Site"
	case "terror":
		m = battle.GenerateTerrorSite(w, h, seed)
		label = "Terror Site"
	case "abduction":
		m = battle.GenerateAbductionSite(w, h)
		label = "Abduction Site"
	case "ufo_interior":
		m = battle.GenerateUFOInterior(w, h, seed)
		label = "UFO Interior (hand-crafted)"
	case "ufo_wfc":
		m = battle.GenerateUFOInteriorWFC(w, h, rng)
		label = "UFO Interior (WFC)"
	case "alien_base":
		m = battle.GenerateAlienBase(w, h)
		label = "Alien Base (hand-crafted)"
	case "alien_base_wfc":
		m = battle.GenerateAlienBaseWFC(w, h, rng)
		label = "Alien Base (WFC)"
	case "building":
		m = battle.GenerateUrbanBuildingWFC(w, h, rng)
		label = "Urban Building (WFC)"
	case "building2":
		m = battle.GenerateUrbanBuildingWFCLevels(w, h, 2, rng)
		label = "Urban Building 2-Level (WFC)"
	case "cydonia":
		m = battle.GenerateCydonia(w, h)
		label = "Cydonia"
	case "urban", "forest", "desert", "polar", "rural", "ufo", "alien", "farm", "coastal", "mountain", "swamp", "jungle":
		m = battle.AssembleMap(gen, w, h, rng)
		label = gen + " (AssembleMap)"
	case "all":
		m = battle.AssembleMap(allBiomes[0], w, h, rng)
		biomeIdx = 0
		label = fmt.Sprintf("all biomes [%d/%d] %s (AssembleMap)", biomeIdx+1, len(allBiomes), allBiomes[0])
	default:
		m, _ = battle.GenerateCrashSite(w, h, seed, -1, -1)
		label = "Crash Site (fallback)"
	}

	if dump {
		dumpMap(m, label)
		return
	}

	allBiomeNames := allBiomes

	var cycleFn func() (string, *battle.BattleMap)
	if gen == "all" {
		cycleFn = func() (string, *battle.BattleMap) {
			biomeIdx = (biomeIdx + 1) % len(allBiomeNames)
			b := allBiomeNames[biomeIdx]
			rng2 := rand.New(rand.NewSource(seed))
			nm := battle.AssembleMap(b, w, h, rng2)
			nl := fmt.Sprintf("all biomes [%d/%d] %s (AssembleMap)", biomeIdx+1, len(allBiomeNames), b)
			return nl, nm
		}
	}

	drawMap(m, label, cycleFn)
}

func parseArgs() (gen string, w, h int, seed int64, dump bool) {
	gen = ""
	w, h = 50, 50
	seed = time.Now().UnixNano()

	args := os.Args[1:]
	if len(args) == 0 {
		return "", 0, 0, 0, false
	}

	if args[0] == "--dump" || args[0] == "-d" {
		dump = true
		args = args[1:]
		if len(args) == 0 {
			return "", 0, 0, 0, false
		}
	}

	gen = strings.ToLower(args[0])
	if gen == "--list" || gen == "-l" {
		return gen, 0, 0, 0, false
	}

	valid := false
	for _, g := range generators {
		if gen == g.name {
			valid = true
			break
		}
	}
	if !valid {
		if s, err := strconv.ParseInt(gen, 10, 64); err == nil {
			seed = s
			gen = "crash"
		}
	}

	if len(args) > 1 {
		if v, err := strconv.Atoi(args[1]); err == nil && v > 0 {
			w = v
		}
	}
	if len(args) > 2 {
		if v, err := strconv.Atoi(args[2]); err == nil && v > 0 {
			h = v
		}
	}
	if len(args) > 3 {
		if v, err := strconv.ParseInt(args[3], 10, 64); err == nil {
			seed = v
		}
	}
	return
}

func printUsage() {
	fmt.Println("Usage: go run ./cmd/test_map [--dump] <map_type> [width] [height] [seed]")
	fmt.Println("  --dump      Print the generated map as ASCII to stdout (non-interactive)")
	fmt.Println()
	printGenerators()
}

func printGenerators() {
	fmt.Println("Available map types:")
	for _, g := range generators {
		fmt.Printf("  %-16s  %s\n", g.name, g.desc)
	}
}

func tileContext(m *battle.BattleMap, x, y, level int) [3][3]battle.TileType {
	var ctx [3][3]battle.TileType
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= m.Width || ny < 0 || ny >= m.LevelHeight {
				ctx[dy+1][dx+1] = battle.TileWall
			} else {
				ctx[dy+1][dx+1] = m.AtLevel(nx, ny, level).Type
			}
		}
	}
	return ctx
}

func drawMap(m *battle.BattleMap, label string, cycleFn func() (string, *battle.BattleMap)) {
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tcell.NewScreen: %v\n", err)
		os.Exit(1)
	}
	if err := s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "tcell.Init: %v\n", err)
		os.Exit(1)
	}
	defer s.Fini()

	s.EnableMouse()
	s.SetStyle(tcell.StyleDefault)

	level := 0
	camX, camY := 0, 0
	cursorX, cursorY := -1, -1

	draw := func() {
		s.Clear()
		sw, sh := s.Size()

		infoW := 0
		infoX := sw
		if sw >= m.Width+infoPanelWidth+2 {
			infoW = infoPanelWidth
			infoX = sw - infoW
		}

		mapAreaW := sw
		if infoW > 0 {
			mapAreaW = infoX
		}

		offX := (mapAreaW - m.Width) / 2
		if offX < 0 {
			offX = -camX
		}
		offY := (sh - m.LevelHeight) / 3
		if offY < 0 {
			offY = -camY
		}

		// Draw tiles
		for y := 0; y < m.LevelHeight && y+offY < sh; y++ {
			for x := 0; x < m.Width && x+offX < mapAreaW; x++ {
				sx, sy := x+offX, y+offY
				if sx < 0 || sy < 0 {
					continue
				}
				t := m.AtLevel(x, y, level)
				ctx := tileContext(m, x, y, level)
				ch, style := battle.RenderTile(t, ctx, true, true, 0, x, y)
				s.SetContent(sx, sy, ch, nil, style)

				// Cursor highlight
				if x == cursorX && y == cursorY && cursorX >= 0 {
					s.SetContent(sx, sy, ch, nil, style.Reverse(true))
				}
			}
		}

		// Info bar at top (map area portion)
		info := fmt.Sprintf(" %s | %dx%d | level %d/%d", label, m.Width, m.LevelHeight, level+1, m.NumLevels)
		if m.NumLevels <= 1 {
			info = fmt.Sprintf(" %s | %dx%d", label, m.Width, m.Height)
		}
		for i, ch := range info {
			if i < sw {
				s.SetContent(i, 0, ch, nil, tcell.StyleDefault.Foreground(color.White).Background(color.Black))
			}
		}

		// Help bar at bottom
		help := " q:quit | click:select tile | arrows/wasd:scroll | tab:cycle floor"
		if cycleFn != nil {
			help += " | b:cycle biome"
		}
		for i, ch := range help {
			if i < sw {
				s.SetContent(i, sh-1, ch, nil, tcell.StyleDefault.Foreground(color.Gray).Background(color.Black))
			}
		}

		// Info panel on the right
		if infoW > 0 && cursorX >= 0 && cursorY >= 0 && cursorX < m.Width && cursorY < m.LevelHeight {
			t := m.AtLevel(cursorX, cursorY, level)
			drawInfoPanel(s, infoX, 2, infoW, sh-3, m, t, cursorX, cursorY, level)
		}

		s.Show()
	}

	draw()

	eventQ := s.EventQ()
	done := make(chan struct{})
	go func() {
		for {
			select {
			case ev := <-eventQ:
				switch e := ev.(type) {
				case *tcell.EventKey:
					switch e.Key() {
					case tcell.KeyEscape, tcell.KeyCtrlC:
						close(done)
						return
					case tcell.KeyLeft:
						camX -= 5
						if camX < 0 {
							camX = 0
						}
						draw()
					case tcell.KeyRight:
						camX += 5
						draw()
					case tcell.KeyUp:
						camY -= 5
						if camY < 0 {
							camY = 0
						}
						draw()
					case tcell.KeyDown:
						camY += 5
						draw()
					case tcell.KeyTab:
						level = (level + 1) % m.NumLevels
						draw()
					case tcell.KeyBacktab:
						level = (level + m.NumLevels - 1) % m.NumLevels
						draw()
					case tcell.KeyRune:
						switch e.Str() {
						case "q", "Q":
							close(done)
							return
						case "w", "W":
							camY -= 5
							if camY < 0 {
								camY = 0
							}
							draw()
						case "a", "A":
							camX -= 5
							if camX < 0 {
								camX = 0
							}
							draw()
						case "s", "S":
							camY += 5
							draw()
						case "d", "D":
							camX += 5
							draw()
						case "b", "B":
							if cycleFn != nil {
								newLabel, newMap := cycleFn()
								if newMap != nil {
									m = newMap
									label = newLabel
									level = 0
									camX, camY = 0, 0
									cursorX, cursorY = -1, -1
									draw()
								}
							}
						}
					}
				case *tcell.EventMouse:
					mx, my := e.Position()
					sw, sh := s.Size()

					infoW := 0
					infoX := sw
					if sw >= m.Width+infoPanelWidth+2 {
						infoW = infoPanelWidth
						infoX = sw - infoW
					}
					mapAreaW := sw
					if infoW > 0 {
						mapAreaW = infoX
					}

					offX := (mapAreaW - m.Width) / 2
					if offX < 0 {
						offX = -camX
					}
					offY := (sh - m.LevelHeight) / 3
					if offY < 0 {
						offY = -camY
					}

					tx := mx - offX
					ty := my - offY
					if tx >= 0 && tx < m.Width && ty >= 0 && ty < m.LevelHeight {
						if cursorX != tx || cursorY != ty {
							cursorX, cursorY = tx, ty
							draw()
						}
					}
				case *tcell.EventResize:
					draw()
				}
			case <-done:
				return
			}
		}
	}()
	<-done
}

func dumpMap(m *battle.BattleMap, label string) {
	fmt.Printf("%s | %dx%d | levels %d\n", label, m.Width, m.Height, m.NumLevels)
	for level := 0; level < m.NumLevels; level++ {
		if m.NumLevels > 1 {
			fmt.Printf("\n=== Level %d/%d ===\n", level+1, m.NumLevels)
		}
		for y := 0; y < m.LevelHeight; y++ {
			var sb strings.Builder
			for x := 0; x < m.Width; x++ {
				t := m.AtLevel(x, y, level)
				ctx := tileContext(m, x, y, level)
				ch, _ := battle.RenderTile(t, ctx, true, true, 0, x, y)
				sb.WriteRune(ch)
			}
			fmt.Println(sb.String())
		}
	}
}

func drawInfoPanel(s tcell.Screen, x, y, w, maxH int, m *battle.BattleMap, t battle.Tile, tileX, tileY, level int) {
	if w < 10 || maxH < 3 {
		return
	}

	white := tcell.StyleDefault.Foreground(color.White).Background(color.Black)
	gray := tcell.StyleDefault.Foreground(color.Gray).Background(color.Black)
	cyan := tcell.StyleDefault.Foreground(color.LightCyan).Background(color.Black)
	yellow := tcell.StyleDefault.Foreground(color.Yellow).Background(color.Black)
	green := tcell.StyleDefault.Foreground(color.LightGreen).Background(color.Black)
	red := tcell.StyleDefault.Foreground(color.Red).Background(color.Black)

	// Draw panel border
	s.SetContent(x, y-1, '┌', nil, gray)
	for i := 1; i < w-1; i++ {
		s.SetContent(x+i, y-1, '─', nil, gray)
	}
	s.SetContent(x+w-1, y-1, '┐', nil, gray)

	row := y

	// Title
	title := " Tile Info"
	for i, ch := range title {
		if i < w {
			s.SetContent(x+i, row, ch, nil, cyan)
		}
	}
	row++

	// Separator
	for i := 0; i < w; i++ {
		s.SetContent(x+i, row, '─', nil, gray)
	}
	row++

	// Coordinates
	coordStr := fmt.Sprintf(" (%d, %d, L%d)", tileX, tileY, level)
	for i, ch := range coordStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, ch, nil, white)
		}
	}
	row++

	// Tile type name
	name := tileTypeNames[t.Type]
	if name == "" {
		name = fmt.Sprintf("TileType(%d)", t.Type)
	}
	nameStr := fmt.Sprintf(" %s", name)
	for i, ch := range nameStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, ch, nil, yellow)
		}
	}
	row++

	// Symbol
	ctx := tileContext(m, tileX, tileY, level)
	ch, style := battle.RenderTile(t, ctx, true, true, 0, tileX, tileY)
	symStr := fmt.Sprintf(" Char: %c  (0x%04x)", ch, ch)
	for i, c := range symStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, c, nil, style)
		}
	}
	row++

	// Cover %
	covStr := fmt.Sprintf(" Cover: %d%%", t.Cover)
	sty := white
	if t.Cover >= 80 {
		sty = green
	} else if t.Cover >= 50 {
		sty = yellow
	} else if t.Cover > 0 {
		sty = red
	}
	for i, c := range covStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, c, nil, sty)
		}
	}
	row++

	// Flammable
	flamStr := " Flammable: no"
	if t.IsFlammable() {
		flamStr = " Flammable: yes"
	}
	fSty := gray
	if t.IsFlammable() {
		fSty = yellow
	}
	for i, c := range flamStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, c, nil, fSty)
		}
	}
	row++

	// Destroyed
	if t.Destroyed {
		destStr := " Destroyed"
		for i, c := range destStr {
			if row < maxH && i < w {
				s.SetContent(x+i, row, c, nil, red)
			}
		}
		row++
	}

	// Blood
	if t.Blood > 0 {
		bloodLabels := map[int]string{1: "Red", 2: "Alien", 3: "Alien"}
		bl := bloodLabels[t.Blood]
		bloodStr := fmt.Sprintf(" Blood: %s", bl)
		for i, c := range bloodStr {
			if row < maxH && i < w {
				s.SetContent(x+i, row, c, nil, red)
			}
		}
		row++
	}

	// Fire
	if t.Fire > 0 {
		fireStr := fmt.Sprintf(" Fire: %d turns", t.Fire)
		for i, c := range fireStr {
			if row < maxH && i < w {
				s.SetContent(x+i, row, c, nil, yellow)
			}
		}
		row++
	}

	// Opaque (blocks LOS)
	opStr := " Opaque: yes"
	if !m.Opaque(tileX, tileY) {
		opStr = " Opaque: no"
	}
	for i, c := range opStr {
		if row < maxH && i < w {
			s.SetContent(x+i, row, c, nil, white)
		}
	}
	row++

	// Base color preview
	if row < maxH {
		colorStr := " Color: "
		for i, c := range colorStr {
			if i < w {
				s.SetContent(x+i, row, c, nil, white)
			}
		}
		// Draw a colored sample
		bc := battle.TileBaseColor(t)
		if len(colorStr) < w {
			s.SetContent(x+len(colorStr), row, '█', nil, tcell.StyleDefault.Foreground(bc).Background(bc))
		}
		if len(colorStr)+1 < w {
			s.SetContent(x+len(colorStr)+1, row, '█', nil, tcell.StyleDefault.Foreground(bc).Background(bc))
		}
	}
}
