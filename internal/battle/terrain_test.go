package battle

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestTileGeomRune_UFOCorners(t *testing.T) {
	tile := Tile{Type: TileUFOWall}

	// NW corner context: north=Grass, west=Grass, east=UFOWall, south=UFOWall
	var ctxNW [3][3]TileType
	ctxNW[0][1] = TileGrass   // North
	ctxNW[1][0] = TileGrass   // West
	ctxNW[1][2] = TileUFOWall // East
	ctxNW[2][1] = TileUFOWall // South

	rNW := TileGeomRune(tile, ctxNW)
	if rNW != GlyphUFOCornerNW {
		t.Errorf("Expected NW corner U+25E4 (◤), got %c", rNW)
	}

	// NE corner context: north=Grass, east=Grass, west=UFOWall, south=UFOWall
	var ctxNE [3][3]TileType
	ctxNE[0][1] = TileGrass   // North
	ctxNE[1][2] = TileGrass   // East
	ctxNE[1][0] = TileUFOWall // West
	ctxNE[2][1] = TileUFOWall // South

	rNE := TileGeomRune(tile, ctxNE)
	if rNE != GlyphUFOCornerNE {
		t.Errorf("Expected NE corner U+25E5 (◥), got %c", rNE)
	}
}

func TestTileGeomRune_BuildingCorners(t *testing.T) {
	tile := Tile{Type: TileWall}

	// TL corner: north=Grass, west=Grass, east=Wall, south=Wall
	var ctxTL [3][3]TileType
	ctxTL[0][1] = TileGrass // North
	ctxTL[1][0] = TileGrass // West
	ctxTL[1][2] = TileWall  // East
	ctxTL[2][1] = TileWall  // South

	rTL := TileGeomRune(tile, ctxTL)
	if rTL != GlyphBuildingTL {
		t.Errorf("Expected building TL corner (╔), got %c", rTL)
	}
}

func TestRenderTile_Visibility(t *testing.T) {
	tile := Tile{Type: TileGrass}
	var ctx [3][3]TileType

	// 1. Unseen tile
	rUnseen, styleUnseen := RenderTile(tile, ctx, false, false, 0)
	if rUnseen != ' ' {
		t.Errorf("Expected blank rune for unseen tile, got %c", rUnseen)
	}
	bgUnseen := styleUnseen.GetBackground()
	if bgUnseen != tcell.ColorBlack {
		t.Errorf("Expected black BG for unseen tile, got %v", bgUnseen)
	}

	// 2. Visible tile
	rVisible, styleVisible := RenderTile(tile, ctx, true, true, 0)
	if rVisible != '·' {
		t.Errorf("Expected '·' for Grass, got %c", rVisible)
	}
	fgVisible := styleVisible.GetForeground()
	expectedColor := TilePalette[TileGrass]
	if fgVisible != expectedColor {
		t.Errorf("Expected color %v, got %v", expectedColor, fgVisible)
	}

	// 3. Seen but not currently visible (Fog of War)
	_, styleFog := RenderTile(tile, ctx, false, true, 0)
	fgFog := styleFog.GetForeground()
	// Should be darker than visible
	rV, gV, bV := fgVisible.RGB()
	rF, gF, bF := fgFog.RGB()
	if rF >= rV && gF >= gV && bF >= bV && (rV > 0 || gV > 0 || bV > 0) {
		t.Errorf("Expected Fog of War color to be darker than visible color")
	}
}

func TestRenderTile_RuneOverride(t *testing.T) {
	tile := Tile{Type: TileGrass, Rune: 'X'}
	var ctx [3][3]TileType

	r, _ := RenderTile(tile, ctx, true, true, 0)
	if r != 'X' {
		t.Errorf("Expected rune override 'X', got %c", r)
	}
}
