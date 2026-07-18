package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/battle"
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
	{"forest", "Forest biome (AssembleMap)"},
	{"desert", "Desert biome (AssembleMap)"},
	{"polar", "Polar biome (AssembleMap)"},
	{"all", "All biomes (AssembleMap) — shows biome name"},
}

func main() {
	gen, w, h, seed := parseArgs()
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

	var m *battle.BattleMap
	label := gen

	switch gen {
	case "crash":
		m, _ = battle.GenerateCrashSite(w, h, seed)
		label = "Crash Site"
	case "terror":
		m = battle.GenerateTerrorSite(w, h)
		label = "Terror Site"
	case "abduction":
		m = battle.GenerateAbductionSite(w, h)
		label = "Abduction Site"
	case "ufo_interior":
		m = battle.GenerateUFOInterior(w, h)
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
	case "forest", "desert", "polar":
		m = battle.AssembleMap(gen, w, h, rng)
		label = gen + " (AssembleMap)"
	case "all":
		m = battle.AssembleMap("urban", w, h, rng)
		label = "urban (AssembleMap)"
	default:
		m, _ = battle.GenerateCrashSite(w, h, seed)
		label = "Crash Site (fallback)"
	}

	drawMap(m, label)
}

func parseArgs() (gen string, w, h int, seed int64) {
	gen = ""
	w, h = 50, 50
	seed = time.Now().UnixNano()

	args := os.Args[1:]
	if len(args) == 0 {
		return "", 0, 0, 0
	}

	gen = strings.ToLower(args[0])
	if gen == "--list" || gen == "-l" {
		return gen, 0, 0, 0
	}

	// Check if the gen is valid, else treat as seed
	valid := false
	for _, g := range generators {
		if gen == g.name {
			valid = true
			break
		}
	}
	if !valid {
		// Maybe it's a seed, treat as crash with that seed
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
	fmt.Println("Usage: go run ./cmd/test_map <map_type> [width] [height] [seed]")
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

func drawMap(m *battle.BattleMap, label string) {
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

	draw := func() {
		s.Clear()
		sw, sh := s.Size()

		// If map smaller than screen, centre it; else top-left
		offX := (sw - m.Width) / 2
		if offX < 0 {
			offX = -camX
		}
		offY := (sh - m.LevelHeight) / 3
		if offY < 0 {
			offY = -camY
		}

		for y := 0; y < m.LevelHeight && y+offY < sh; y++ {
			for x := 0; x < m.Width && x+offX < sw; x++ {
				sx, sy := x+offX, y+offY
				if sx < 0 || sy < 0 {
					continue
				}
				t := m.AtLevel(x, y, level)
				ctx := tileContext(m, x, y, level)
				ch, style := battle.RenderTile(t, ctx, true, true, 0, x, y)
				s.SetContent(sx, sy, ch, nil, style)
			}
		}

		// Info bar at top
		info := fmt.Sprintf(" %s | %dx%d | level %d/%d | seed used", label, m.Width, m.LevelHeight, level+1, m.NumLevels)
		if m.NumLevels <= 1 {
			info = fmt.Sprintf(" %s | %dx%d", label, m.Width, m.Height)
		}
		for i, ch := range info {
			if i < sw {
				s.SetContent(i, 0, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack))
			}
		}

		// Help bar at bottom
		help := " q:quit  arrows:scroll  n/p:level  c:current-level overlay"
		if m.NumLevels <= 1 {
			help = " q:quit  arrows:scroll"
		}
		for i, ch := range help {
			if i < sw {
				s.SetContent(i, sh-1, ch, nil, tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack))
			}
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
					case tcell.KeyRune:
						switch e.Str() {
						case "q", "Q":
							close(done)
							return
						case "n", "N":
							if level < m.NumLevels-1 {
								level++
								draw()
							}
						case "p", "P":
							if level > 0 {
								level--
								draw()
							}
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
