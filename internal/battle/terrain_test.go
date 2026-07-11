package battle

import (
	"testing"

	"github.com/gdamore/tcell/v3/color"
)

func TestTileGeomRune_UFOCorners(t *testing.T) {
	tile := Tile{Type: TileUFOWall}

	// NW corner context
	var ctxNW [3][3]TileType
	ctxNW[0][1] = TileGrass
	ctxNW[1][0] = TileGrass
	ctxNW[1][2] = TileUFOWall
	ctxNW[2][1] = TileUFOWall

	rNW := TileGeomRune(tile, ctxNW)
	if rNW != GlyphUFOCornerNW {
		t.Errorf("Expected NW corner ◤, got %c", rNW)
	}

	// NE corner context
	var ctxNE [3][3]TileType
	ctxNE[0][1] = TileGrass
	ctxNE[1][2] = TileGrass
	ctxNE[1][0] = TileUFOWall
	ctxNE[2][1] = TileUFOWall

	rNE := TileGeomRune(tile, ctxNE)
	if rNE != GlyphUFOCornerNE {
		t.Errorf("Expected NE corner ◥, got %c", rNE)
	}

	// SW corner context
	var ctxSW [3][3]TileType
	ctxSW[2][1] = TileGrass
	ctxSW[1][0] = TileGrass
	ctxSW[1][2] = TileUFOWall
	ctxSW[0][1] = TileUFOWall

	rSW := TileGeomRune(tile, ctxSW)
	if rSW != GlyphUFOCornerSW {
		t.Errorf("Expected SW corner ◣, got %c", rSW)
	}

	// SE corner context
	var ctxSE [3][3]TileType
	ctxSE[2][1] = TileGrass
	ctxSE[1][2] = TileGrass
	ctxSE[1][0] = TileUFOWall
	ctxSE[0][1] = TileUFOWall

	rSE := TileGeomRune(tile, ctxSE)
	if rSE != GlyphUFOCornerSE {
		t.Errorf("Expected SE corner ◢, got %c", rSE)
	}
}

func TestTileGeomRune_UFOHull(t *testing.T) {
	tile := Tile{Type: TileUFOWall}

	// Horizontal hull: west=UFO, east=UFO, north=grass, south=grass
	var ctxH [3][3]TileType
	ctxH[1][0] = TileUFOWall
	ctxH[1][2] = TileUFOWall
	ctxH[0][1] = TileGrass
	ctxH[2][1] = TileGrass

	rH := TileGeomRune(tile, ctxH)
	if rH != GlyphUFOHullH {
		t.Errorf("Expected horizontal hull ▬, got %c", rH)
	}

	// Vertical hull: north=UFO, south=UFO, west=grass, east=grass
	var ctxV [3][3]TileType
	ctxV[0][1] = TileUFOWall
	ctxV[2][1] = TileUFOWall
	ctxV[1][0] = TileGrass
	ctxV[1][2] = TileGrass

	rV := TileGeomRune(tile, ctxV)
	if rV != GlyphUFOHullV {
		t.Errorf("Expected vertical hull ▐, got %c", rV)
	}
}

func TestTileGeomRune_BuildingCorners(t *testing.T) {
	tile := Tile{Type: TileWall}

	// TL corner
	var ctxTL [3][3]TileType
	ctxTL[0][1] = TileGrass
	ctxTL[1][0] = TileGrass
	ctxTL[1][2] = TileWall
	ctxTL[2][1] = TileWall
	if r := TileGeomRune(tile, ctxTL); r != GlyphBuildingTL {
		t.Errorf("Expected ╔, got %c", r)
	}

	// TR corner
	var ctxTR [3][3]TileType
	ctxTR[0][1] = TileGrass
	ctxTR[1][2] = TileGrass
	ctxTR[1][0] = TileWall
	ctxTR[2][1] = TileWall
	if r := TileGeomRune(tile, ctxTR); r != GlyphBuildingTR {
		t.Errorf("Expected ╗, got %c", r)
	}

	// BL corner
	var ctxBL [3][3]TileType
	ctxBL[2][1] = TileGrass
	ctxBL[1][0] = TileGrass
	ctxBL[1][2] = TileWall
	ctxBL[0][1] = TileWall
	if r := TileGeomRune(tile, ctxBL); r != GlyphBuildingBL {
		t.Errorf("Expected ╚, got %c", r)
	}

	// BR corner
	var ctxBR [3][3]TileType
	ctxBR[2][1] = TileGrass
	ctxBR[1][2] = TileGrass
	ctxBR[1][0] = TileWall
	ctxBR[0][1] = TileWall
	if r := TileGeomRune(tile, ctxBR); r != GlyphBuildingBR {
		t.Errorf("Expected ╝, got %c", r)
	}
}

func TestTileGeomRune_BuildingHorizontal(t *testing.T) {
	tile := Tile{Type: TileWall}
	var ctx [3][3]TileType
	ctx[1][0] = TileWall
	ctx[1][2] = TileWall
	ctx[0][1] = TileGrass
	ctx[2][1] = TileGrass

	if r := TileGeomRune(tile, ctx); r != GlyphBuildingH {
		t.Errorf("Expected ═, got %c", r)
	}
}

func TestTileGeomRune_DoorAndWindow(t *testing.T) {
	var ctx [3][3]TileType
	if r := TileGeomRune(Tile{Type: TileDoor}, ctx); r != GlyphBuildingDoor {
		t.Errorf("Expected door glyph ▒, got %c", r)
	}
	if r := TileGeomRune(Tile{Type: TileWindow}, ctx); r != GlyphBuildingWin {
		t.Errorf("Expected window glyph ┼, got %c", r)
	}
}

func TestTileBaseColor_AllTileTypes(t *testing.T) {
	// Ensure every tile type in the palette resolves without panic
	for tileType, expectedColor := range TilePalette {
		tt := Tile{Type: tileType}
		got := TileBaseColor(tt)
		if got != expectedColor {
			t.Errorf("TileBaseColor(%v) = %v, want %v", tileType, got, expectedColor)
		}
	}
}

func TestTileBaseColor_Override(t *testing.T) {
	custom := color.NewRGBColor(100, 200, 50)
	tile := Tile{Type: TileGrass, BaseColor: custom}
	got := TileBaseColor(tile)
	if got != custom {
		t.Errorf("Expected custom color override, got %v", got)
	}
}

func TestTileBaseColor_Unknown(t *testing.T) {
	tile := Tile{Type: TileType(9999)}
	got := TileBaseColor(tile)
	r, g, b := got.RGB()
	if r != 128 || g != 128 || b != 128 {
		t.Errorf("Expected grey fallback for unknown tile, got (%d,%d,%d)", r, g, b)
	}
}

func TestTileGeomRune_Override(t *testing.T) {
	tile := Tile{Type: TileGrass, Rune: 'Z'}
	var ctx [3][3]TileType
	if r := TileGeomRune(tile, ctx); r != 'Z' {
		t.Errorf("Expected override rune Z, got %c", r)
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
	if bgUnseen != color.Black {
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
	rV, gV, bV := fgVisible.RGB()
	rF, gF, bF := fgFog.RGB()
	if rF >= rV && gF >= gV && bF >= bV && (rV > 0 || gV > 0 || bV > 0) {
		t.Errorf("Expected Fog of War color to be darker than visible color")
	}
}

func TestRenderTile_BloodOverlay(t *testing.T) {
	tile := Tile{Type: TileGrass, Blood: 1}
	var ctx [3][3]TileType
	_, style := RenderTile(tile, ctx, true, true, 0)
	fg := style.GetForeground()
	r, g, b := fg.RGB()
	if r < 150 || g > 50 || b > 50 {
		t.Errorf("Expected red blood color, got (%d,%d,%d)", r, g, b)
	}
}

func TestRenderTile_FireOverlay(t *testing.T) {
	tile := Tile{Type: TileGrass, Fire: 3}
	var ctx [3][3]TileType
	_, style := RenderTile(tile, ctx, true, true, 0)
	fg := style.GetForeground()
	r, g, b := fg.RGB()
	if r < 200 || g < 30 {
		t.Errorf("Expected fire color (orange/red), got (%d,%d,%d)", r, g, b)
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
