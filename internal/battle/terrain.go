package battle

import (
	"github.com/taislin/termcom/internal/engine"
	"github.com/gdamore/tcell/v3"
)

// UFO hull geometry glyphs
const (
	GlyphUFOCornerNW rune = '◤'
	GlyphUFOCornerNE rune = '◥'
	GlyphUFOCornerSW rune = '◣'
	GlyphUFOCornerSE rune = '◢'
	GlyphUFOHullH    rune = '▬'
	GlyphUFOHullV    rune = '▐'
	GlyphUFOHatch    rune = '⊠'
)

// Human building box-drawing glyphs
const (
	GlyphBuildingTL   rune = '╔'
	GlyphBuildingTR   rune = '╗'
	GlyphBuildingBL   rune = '╚'
	GlyphBuildingBR   rune = '╝'
	GlyphBuildingH    rune = '═'
	GlyphBuildingV    rune = '║'
	GlyphBuildingDoor rune = '▒'
	GlyphBuildingWin  rune = '┼'
)

// TilePalette maps TileType to a curated true-color RGB value.
var TilePalette = map[TileType]tcell.Color{
	TileFloor:       tcell.NewRGBColor(60, 55, 50),
	TileWall:        tcell.NewRGBColor(120, 115, 110),
	TileDoor:        tcell.NewRGBColor(100, 60, 30),
	TileWindow:      tcell.NewRGBColor(100, 150, 200),
	TileGrass:       tcell.NewRGBColor(35, 75, 25),
	TileTree:        tcell.NewRGBColor(20, 60, 15),
	TileRock:        tcell.NewRGBColor(90, 85, 80),
	TileWater:       tcell.NewRGBColor(25, 50, 140),
	TileUFOFloor:    tcell.NewRGBColor(30, 50, 75),
	TileUFOWall:     tcell.NewRGBColor(45, 70, 110),
	TileStairs:      tcell.NewRGBColor(70, 65, 60),
	TileStairsDown:  tcell.NewRGBColor(50, 45, 40),
	TilePavement:    tcell.NewRGBColor(75, 75, 75),
	TileSand:        tcell.NewRGBColor(170, 150, 95),
	TileSnow:        tcell.NewRGBColor(220, 225, 235),
	TileMarsh:       tcell.NewRGBColor(45, 75, 55),
	TileBush:        tcell.NewRGBColor(30, 65, 20),
	TileFence:       tcell.NewRGBColor(110, 85, 60),
	TileRubble:      tcell.NewRGBColor(85, 80, 75),
	TileObject:      tcell.NewRGBColor(130, 130, 130),
	TileConsole:     tcell.NewRGBColor(50, 180, 100),
	TileMachinery:   tcell.NewRGBColor(140, 140, 140),
	TilePod:         tcell.NewRGBColor(100, 50, 150),
	TilePowerSource: tcell.NewRGBColor(220, 180, 40),
	TileStorage:     tcell.NewRGBColor(150, 110, 70),
	TileAlienTech:   tcell.NewRGBColor(200, 50, 50),
}

// TileBaseColor returns the resolved color for a tile.
func TileBaseColor(t Tile) tcell.Color {
	if t.BaseColor != tcell.ColorDefault {
		return t.BaseColor
	}
	if col, ok := TilePalette[t.Type]; ok {
		return col
	}
	return tcell.NewRGBColor(128, 128, 128) // neutral grey fallback
}

// TileGeomRune selects the display rune for a tile, taking context into account.
func TileGeomRune(t Tile, ctx [3][3]TileType) rune {
	if t.Rune != 0 {
		return t.Rune
	}

	// Context coordinates are:
	// ctx[y][x] where:
	// y=0: north, y=1: center, y=2: south
	// x=0: west,  x=1: center, x=2: east
	n := ctx[0][1]
	s := ctx[2][1]
	w := ctx[1][0]
	e := ctx[1][2]

	switch t.Type {
	case TileUFOWall:
		// Corner calculations
		nIsUFO := n == TileUFOWall
		sIsUFO := s == TileUFOWall
		wIsUFO := w == TileUFOWall
		eIsUFO := e == TileUFOWall

		if !nIsUFO && !wIsUFO && eIsUFO && sIsUFO {
			return GlyphUFOCornerNW // North & West are external, East & South are UFO walls
		}
		if !nIsUFO && !eIsUFO && wIsUFO && sIsUFO {
			return GlyphUFOCornerNE // North & East are external, West & South are UFO walls
		}
		if !sIsUFO && !wIsUFO && eIsUFO && nIsUFO {
			return GlyphUFOCornerSW // South & West are external, East & North are UFO walls
		}
		if !sIsUFO && !eIsUFO && wIsUFO && nIsUFO {
			return GlyphUFOCornerSE // South & East are external, West & North are UFO walls
		}
		if wIsUFO && eIsUFO {
			return GlyphUFOHullH
		}
		if nIsUFO && sIsUFO {
			return GlyphUFOHullV
		}
		return '█' // Default block

	case TileWall:
		// Human building corners and lines
		nIsWall := n == TileWall
		sIsWall := s == TileWall
		wIsWall := w == TileWall
		eIsWall := e == TileWall

		if !nIsWall && !wIsWall && eIsWall && sIsWall {
			return GlyphBuildingTL
		}
		if !nIsWall && !eIsWall && wIsWall && sIsWall {
			return GlyphBuildingTR
		}
		if !sIsWall && !wIsWall && eIsWall && nIsWall {
			return GlyphBuildingBL
		}
		if !sIsWall && !eIsWall && wIsWall && nIsWall {
			return GlyphBuildingBR
		}
		if wIsWall || eIsWall {
			return GlyphBuildingH
		}
		if nIsWall || sIsWall {
			return GlyphBuildingV
		}
		return '#' // Default hash

	case TileDoor:
		return GlyphBuildingDoor
	case TileWindow:
		return GlyphBuildingWin
	}

	return TileChar(t.Type)
}

func bloodColor(bloodType int) tcell.Color {
	switch bloodType {
	case 1:
		return tcell.NewRGBColor(200, 0, 0) // red (human)
	case 2:
		return tcell.NewRGBColor(0, 180, 0) // green (alien)
	case 3:
		return tcell.NewRGBColor(140, 0, 140) // purple (alien)
	default:
		return tcell.NewRGBColor(200, 0, 0)
	}
}

func fireColor(frame int) tcell.Color {
	switch frame % 3 {
	case 0:
		return tcell.NewRGBColor(255, 120, 0)
	case 1:
		return tcell.NewRGBColor(255, 60, 0)
	default:
		return tcell.NewRGBColor(255, 200, 0)
	}
}

func isOpaqueTile(t TileType) bool {
	switch t {
	case TileWall, TileTree, TileRock, TileUFOWall, TileFence:
		return true
	}
	return false
}

// RenderTile produces the character and style for drawing a tile.
func RenderTile(t Tile, ctx [3][3]TileType, visible, seen bool, frame int, tileX, tileY int) (rune, tcell.Style) {
	if !visible && !seen {
		return ' ', engine.StyleDefault
	}

	baseCol := TileBaseColor(t)
	fg := baseCol
	bg := engine.StyleDefault.GetBackground()

	// Make background a very dark version of the base color to add rich depth
	bg = engine.DarkenColor(baseCol, 0.15)

	// Ambient occlusion: darken floor tiles adjacent to opaque walls
	if !isOpaqueTile(t.Type) {
		n := ctx[0][1]
		s := ctx[2][1]
		w := ctx[1][0]
		e := ctx[1][2]
		aoCount := 0
		if isOpaqueTile(n) {
			aoCount++
		}
		if isOpaqueTile(s) {
			aoCount++
		}
		if isOpaqueTile(w) {
			aoCount++
		}
		if isOpaqueTile(e) {
			aoCount++
		}
		if aoCount > 0 {
			aoFactor := 1.0 - float64(aoCount)*0.12
			if aoFactor < 0.5 {
				aoFactor = 0.5
			}
			bg = engine.DarkenColor(bg, aoFactor)
		}
	}

	// Subtle per-tile dither based on checkerboard parity
	if (tileX+tileY)%2 == 0 {
		bg = engine.DarkenColor(bg, 0.92)
	}

	if !visible && seen {
		// Fog of War: dim both foreground and background
		fg = engine.DarkenColor(fg, 0.45)
		bg = engine.DarkenColor(bg, 0.45)
	} else {
		// Overlay effects (blood, fire) only visible when tile is currently in line of sight
		if t.Blood > 0 {
			fg = bloodColor(t.Blood)
		}
		if t.Fire > 0 {
			fg = fireColor(frame)
		}
	}

	r := TileGeomRune(t, ctx)
	style := tcell.StyleDefault.Foreground(fg).Background(bg)
	return r, style
}
