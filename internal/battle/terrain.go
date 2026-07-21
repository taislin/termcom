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

// Tile rendering tuning factors.
const (
	bgDarkenFactor      = 0.25 // base background darkening relative to tile color
	aoPerNeighbor       = 0.08 // ambient-occlusion darkening per adjacent opaque tile
	aoMinFactor         = 0.6  // clamp for accumulated AO darkening
	ditherFactor        = 0.92 // checkerboard dither darkening on even-parity tiles
	fogOfWarDim         = 0.45 // foreground/background dim for remembered (seen) tiles
	noiseAlertRadius    = 15   // tiles within which broken glass/debris alerts aliens
)

// Cryo-coolant freeze gas tuning.
const (
	freezeGasCoreDensity     = 3 // density at the vent tile
	freezeGasEdgeDensity     = 2 // density on the 8 surrounding tiles
	freezeTUDrainPerDensity  = 6  // TU lost per gas density level while chilled
)

const skylightFallDamage = 15 // HP damage when unit falls through a skylight

// Blood palette by blood type (1=human red, 2/3=alien green/purple).
var bloodPalette = map[int]tcell.Color{
	1: tcell.NewRGBColor(200, 0, 0),
	2: tcell.NewRGBColor(0, 180, 0),
	3: tcell.NewRGBColor(140, 0, 140),
}

// firePalette cycles by frame phase for animated flames.
var firePalette = []tcell.Color{
	tcell.NewRGBColor(255, 120, 0),
	tcell.NewRGBColor(255, 60, 0),
	tcell.NewRGBColor(255, 200, 0),
}

// Human building box-drawing glyphs
const (
	// Building wall glyphs — all solid blocks for uniform rendering.
	GlyphBuildingTL   rune = '╔'
	GlyphBuildingTR   rune = '╗'
	GlyphBuildingBL   rune = '╚'
	GlyphBuildingBR   rune = '╝'
	GlyphBuildingH    rune = '═'
	GlyphBuildingV    rune = '║'
	GlyphBuildingTJL  rune = '╠'
	GlyphBuildingTJR  rune = '╣'
	GlyphBuildingTJT  rune = '╦'
	GlyphBuildingTJB  rune = '╩'
	GlyphBuildingX    rune = '╬'
	GlyphBuildingDoor rune = '▒'
	GlyphBuildingWin  rune = '□'
)

// TileBaseColor returns the resolved color for a tile.
func TileBaseColor(t Tile) tcell.Color {
	if t.BaseColor != tcell.ColorDefault {
		return t.BaseColor
	}
	if d := GetTileDef(t.Type); d != nil {
		return d.Color
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
		return '#' // Default hash for isolated walls

	case TileWall, TileMetalWall:
		// Human building walls — full box-drawing lookup.
		// Doors and windows sit IN the wall, so count them as wall-like.
		// TileMetalWall also connects to TileWall for continuity.
		isWall := func(tt TileType) bool {
			return tt == TileWall || tt == TileMetalWall || tt == TileDoor || tt == TileWindow
		}
		nW := isWall(n)
		sW := isWall(s)
		wW := isWall(w)
		eW := isWall(e)

		// Cross
		if nW && sW && wW && eW {
			return GlyphBuildingX
		}
		// T-junctions (3 walls)
		if nW && sW && wW && !eW {
			return GlyphBuildingTJR // ╣
		}
		if nW && sW && !wW && eW {
			return GlyphBuildingTJL // ╠
		}
		if nW && !sW && wW && eW {
			return GlyphBuildingTJB // ╩
		}
		if !nW && sW && wW && eW {
			return GlyphBuildingTJT // ╦
		}
		// Corners (2 perpendicular, no opposite)
		if eW && sW && !nW && !wW {
			return GlyphBuildingTL
		}
		if wW && sW && !nW && !eW {
			return GlyphBuildingTR
		}
		if eW && nW && !sW && !wW {
			return GlyphBuildingBL
		}
		if wW && nW && !sW && !eW {
			return GlyphBuildingBR
		}
		// Straight lines
		if nW || sW {
			return GlyphBuildingV
		}
		if wW || eW {
			return GlyphBuildingH
		}
		// Isolated — solid block
		return '█'

	case TileDoor:
		return GlyphBuildingDoor
	case TileWindow:
		return GlyphBuildingWin
	}

	return TileChar(t.Type)
}

func bloodColor(bloodType int) tcell.Color {
	if col, ok := bloodPalette[bloodType]; ok {
		return col
	}
	return bloodPalette[1] // default human red
}

func fireColor(frame int) tcell.Color {
	return firePalette[frame%len(firePalette)]
}

func isOpaqueTile(t TileType) bool {
	if d := GetTileDef(t); d != nil {
		return d.Opaque
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

	// Make background a dark version of the base color for depth
	bg := engine.DarkenColor(baseCol, bgDarkenFactor)

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
			aoFactor := 1.0 - float64(aoCount)*aoPerNeighbor
			if aoFactor < aoMinFactor {
				aoFactor = aoMinFactor
			}
			bg = engine.DarkenColor(bg, aoFactor)
		}
	}

	// Subtle per-tile dither based on checkerboard parity
	if (tileX+tileY)%2 == 0 {
		bg = engine.DarkenColor(bg, ditherFactor)
	}

	if !visible && seen {
		// Fog of War: dim both foreground and background
		fg = engine.DarkenColor(fg, fogOfWarDim)
		bg = engine.DarkenColor(bg, fogOfWarDim)
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

	// Pavement: match foreground to background so the · glyph is
	// invisible and the tile renders as a solid coloured block.
	if t.Type == TilePavement {
		fg = bg
	}

	style := tcell.StyleDefault.Foreground(fg).Background(bg)
	return r, style
}
