package battle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// TileCustomBase is the starting TileType ID for dynamically registered tiles.
const TileCustomBase TileType = 1000

// TileDef describes all gameplay and rendering properties of a tile type.
type TileDef struct {
	Glyph        rune        `json:"-"`
	GlyphStr     string      `json:"glyph"`
	Color        tcell.Color `json:"-"`
	ColorHex     string      `json:"color"`
	Cover        int         `json:"cover"`
	Passable     bool        `json:"passable"`
	Opaque       bool        `json:"opaque"`
	Destructible bool        `json:"destructible"`
	Flammable    bool        `json:"flammable"`
	Explodes     int         `json:"explodes"`
	Noisy        bool        `json:"noisy"`
	MoveCost     int         `json:"move_cost"`
	NameKey      string      `json:"name"`
}

var (
	tileRegistry      = map[TileType]*TileDef{}
	tileRegistryByName = map[string]TileType{}
	nextCustomTileType TileType = TileCustomBase
)

// RegisterTile adds a custom tile type and returns its assigned TileType.
func RegisterTile(id string, def *TileDef) TileType {
	if t, ok := tileRegistryByName[id]; ok {
		tileRegistry[t] = def
		return t
	}
	t := nextCustomTileType
	nextCustomTileType++
	tileRegistry[t] = def
	tileRegistryByName[id] = t
	return t
}

// LookupTileType returns the TileType for a given string ID, or -1 if not found.
func LookupTileType(id string) (TileType, bool) {
	if t, ok := tileRegistryByName[id]; ok {
		return t, true
	}
	return -1, false
}

// GetTileDef returns the TileDef for a TileType, or nil if not registered.
func GetTileDef(t TileType) *TileDef {
	return tileRegistry[t]
}

// TileDefGlyph returns the glyph for a tile type from the registry.
func TileDefGlyph(t TileType) rune {
	if d := tileRegistry[t]; d != nil {
		return d.Glyph
	}
	if r, ok := tileChars[t]; ok {
		return r
	}
	return '?'
}

// TileDefColor returns the color for a tile type from the registry.
func TileDefColor(t TileType) tcell.Color {
	if d := tileRegistry[t]; d != nil {
		return d.Color
	}
	if c, ok := tilePalette[t]; ok {
		return c
	}
	return tcell.ColorDefault
}

// TileDefName returns the display name for a tile type from the registry.
func TileDefName(t TileType) string {
	if d := tileRegistry[t]; d != nil && d.NameKey != "" {
		return d.NameKey
	}
	return ""
}

// TileDefCover returns the cover value for a tile type from the registry.
func TileDefCover(t TileType) int {
	if d := tileRegistry[t]; d != nil {
		return d.Cover
	}
	return 0
}

// TileDefPassable returns whether a tile type is passable from the registry.
func TileDefPassable(t TileType) bool {
	if d := tileRegistry[t]; d != nil {
		return d.Passable
	}
	return false
}

// TileDefOpaque returns whether a tile type blocks LOS from the registry.
func TileDefOpaque(t TileType) bool {
	if d := tileRegistry[t]; d != nil {
		return d.Opaque
	}
	return false
}

// TileDefDestructible returns whether a tile type is destructible from the registry.
func TileDefDestructible(t TileType) bool {
	if d := tileRegistry[t]; d != nil {
		return d.Destructible
	}
	return false
}

// TileDefFlammable returns whether a tile type is flammable from the registry.
func TileDefFlammable(t TileType) bool {
	if d := tileRegistry[t]; d != nil {
		return d.Flammable
	}
	return false
}

// TileDefExplodes returns the explosion radius on destruction (0 = none).
func TileDefExplodes(t TileType) int {
	if d := tileRegistry[t]; d != nil {
		return d.Explodes
	}
	return 0
}

// TileDefNoisy returns whether a tile alerts aliens when stepped on.
func TileDefNoisy(t TileType) bool {
	if d := tileRegistry[t]; d != nil {
		return d.Noisy
	}
	return false
}

// TileDefMoveCost returns the base movement cost for a tile type.
func TileDefMoveCost(t TileType) int {
	if d := tileRegistry[t]; d != nil && d.MoveCost > 0 {
		return d.MoveCost
	}
	return 4
}

func parseHexColor(hex string) tcell.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return tcell.ColorDefault
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// LoadCustomTiles loads custom tile definitions from a JSON file.
// The file format is: { "tiles": [ { "id": "...", ... }, ... ] }
func LoadCustomTiles(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var file struct {
		Tiles []struct {
			ID           string `json:"id"`
			GlyphStr     string `json:"glyph"`
			ColorHex     string `json:"color"`
			Cover        int    `json:"cover"`
			Passable     *bool  `json:"passable"`
			Opaque       *bool  `json:"opaque"`
			Destructible *bool  `json:"destructible"`
			Flammable    *bool  `json:"flammable"`
			Explodes     int    `json:"explodes"`
			Noisy        *bool  `json:"noisy"`
			MoveCost     int    `json:"move_cost"`
			NameKey      string `json:"name"`
		} `json:"tiles"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	for _, t := range file.Tiles {
		if t.ID == "" {
			continue
		}
		def := &TileDef{
			GlyphStr:     t.GlyphStr,
			Glyph:        []rune(t.GlyphStr)[0],
			ColorHex:     t.ColorHex,
			Color:        parseHexColor(t.ColorHex),
			Cover:        t.Cover,
			Passable:     boolPtrVal(t.Passable, true),
			Opaque:       boolPtrVal(t.Opaque, false),
			Destructible: boolPtrVal(t.Destructible, true),
			Flammable:    boolPtrVal(t.Flammable, false),
			Explodes:     t.Explodes,
			Noisy:        boolPtrVal(t.Noisy, false),
			MoveCost:     t.MoveCost,
			NameKey:      t.NameKey,
		}
		if def.MoveCost == 0 {
			def.MoveCost = 4
		}
		RegisterTile(t.ID, def)
	}
	return nil
}

func boolPtrVal(p *bool, def bool) bool {
	if p != nil {
		return *p
	}
	return def
}

// InitCustomTiles loads all .json files from data/tiles/ directory.
func InitCustomTiles() {
	dirs := []string{"data/tiles", "../data/tiles", "../../data/tiles"}
	for _, d := range dirs {
		matches, err := filepath.Glob(filepath.Join(d, "*.json"))
		if err != nil || len(matches) == 0 {
			continue
		}
		for _, path := range matches {
			if err := LoadCustomTiles(path); err != nil {
				// silently skip bad files
				continue
			}
		}
		return
	}
}

func init() {
 populateBuiltinRegistry()
}

// populateBuiltinRegistry fills the tile registry from the hardcoded tile constants.
func populateBuiltinRegistry() {
	type builtinDef struct {
		id           string
		t            TileType
		glyph        rune
		color        tcell.Color
		cover        int
		passable     bool
		opaque       bool
		destructible bool
		flammable    bool
		explodes     int
		noisy        bool
		moveCost     int
		nameKey      string
	}
	builtins := []builtinDef{
		{"TileFloor", TileFloor, '.', rgb(95, 90, 85), 0, true, false, false, false, 0, false, 4, "TILE_FLOOR"},
		{"TileWall", TileWall, '#', rgb(160, 155, 150), 80, false, true, true, false, 0, false, 4, "TILE_WALL"},
		{"TileDoor", TileDoor, '+', rgb(140, 100, 50), 0, true, false, true, true, 0, false, 4, "TILE_DOOR"},
		{"TileWindow", TileWindow, '⊞', rgb(120, 170, 220), 20, false, false, true, false, 0, false, 4, "TILE_WINDOW"},
		{"TileGrass", TileGrass, '·', rgb(50, 110, 40), 0, true, false, false, true, 0, false, 4, "TILE_GRASS"},
		{"TileTree", TileTree, '♣', rgb(35, 90, 25), 60, false, true, true, true, 0, false, 8, "TILE_TREE"},
		{"TileRock", TileRock, '∩', rgb(130, 125, 120), 70, false, true, true, false, 0, false, 8, "TILE_ROCK"},
		{"TileWater", TileWater, '≈', rgb(40, 80, 200), 0, false, false, false, false, 0, false, 8, "TILE_WATER"},
		{"TileUFOFloor", TileUFOFloor, '≡', rgb(50, 75, 110), 0, true, false, false, false, 0, false, 4, "TILE_UFO_FLOOR"},
		{"TileUFOWall", TileUFOWall, '█', rgb(70, 100, 150), 80, false, true, true, false, 0, false, 4, "TILE_UFO_WALL"},
		{"TileStairs", TileStairs, '▒', rgb(110, 105, 100), 0, true, false, false, false, 0, false, 4, "TILE_STAIRS"},
		{"TilePavement", TilePavement, '░', rgb(120, 120, 120), 0, true, false, false, false, 0, false, 3, "TILE_PAVEMENT"},
		{"TileSand", TileSand, '·', rgb(200, 180, 120), 0, true, false, false, false, 0, false, 6, "TILE_SAND"},
		{"TileSnow", TileSnow, '∗', rgb(230, 235, 245), 0, true, false, false, false, 0, false, 5, "TILE_SNOW"},
		{"TileMarsh", TileMarsh, '≋', rgb(60, 100, 70), 0, true, false, false, false, 0, false, 6, "TILE_MARSH"},
		{"TileBush", TileBush, '†', rgb(45, 100, 35), 40, true, false, false, true, 0, false, 5, "TILE_BUSH"},
		{"TileFence", TileFence, '│', rgb(145, 120, 80), 30, false, true, true, true, 0, false, 4, "TILE_FENCE"},
		{"TileRubble", TileRubble, '▒', rgb(120, 115, 110), 20, true, false, false, false, 0, false, 4, "TILE_RUBBLE"},
		{"TileObject", TileObject, '•', rgb(170, 170, 170), 50, false, false, false, false, 0, false, 4, "TILE_OBJECT"},
		{"TileConsole", TileConsole, '⌸', rgb(70, 210, 130), 10, true, false, true, false, 0, false, 4, "TILE_CONSOLE"},
		{"TileMachinery", TileMachinery, '⊛', rgb(180, 180, 180), 30, true, false, true, false, 0, false, 4, "TILE_MACHINERY"},
		{"TilePod", TilePod, '◈', rgb(130, 70, 190), 30, true, false, true, false, 0, false, 4, "TILE_POD"},
		{"TilePowerSource", TilePowerSource, '⌁', rgb(240, 200, 60), 20, true, false, true, false, 0, false, 4, "TILE_POWER_SOURCE"},
		{"TileStorage", TileStorage, '▤', rgb(180, 140, 90), 30, true, false, true, false, 0, false, 4, "TILE_STORAGE"},
		{"TileAlienTech", TileAlienTech, '⊕', rgb(230, 70, 70), 20, true, false, true, false, 0, false, 4, "TILE_ALIEN_TECH"},
		{"TileStairsDown", TileStairsDown, '▓', rgb(80, 75, 70), 0, true, false, false, false, 0, false, 4, "TILE_STAIRS_DOWN"},
		{"TileDesk", TileDesk, '◊', rgb(160, 120, 80), 30, true, false, true, false, 0, false, 4, "TILE_DESK"},
		{"TileChair", TileChair, '⊟', rgb(150, 100, 60), 10, true, false, true, false, 0, false, 4, "TILE_CHAIR"},
		{"TileChairLeft", TileChairLeft, '⅃', rgb(150, 100, 60), 10, true, false, true, false, 0, false, 4, "TILE_CHAIR"},
		{"TileChairRight", TileChairRight, 'L', rgb(150, 100, 60), 10, true, false, true, false, 0, false, 4, "TILE_CHAIR"},
		{"TileComputer", TileComputer, '⌸', rgb(70, 180, 210), 10, true, false, true, false, 0, false, 4, "TILE_COMPUTER"},
		{"TileBed", TileBed, '□', rgb(200, 200, 200), 20, true, false, true, false, 0, false, 4, "TILE_BED"},
		{"TileLocker", TileLocker, '◫', rgb(140, 160, 180), 30, true, false, true, false, 0, false, 4, "TILE_LOCKER"},
		{"TileCabinet", TileCabinet, '⊞', rgb(170, 130, 90), 30, true, false, true, false, 0, false, 4, "TILE_CABINET"},
		{"TileCar", TileCar, '▄', rgb(50, 100, 180), 50, false, false, true, false, 0, false, 4, "TILE_CAR"},
		{"TileCarMid", TileCarMid, '█', rgb(50, 100, 180), 50, false, false, true, false, 0, false, 4, "TILE_CAR"},
		{"TileCarRight", TileCarRight, '▄', rgb(50, 100, 180), 50, false, false, true, false, 0, false, 4, "TILE_CAR"},
		{"TileForklift", TileForklift, '█', rgb(200, 160, 40), 50, false, false, true, false, 0, false, 4, "TILE_FORKLIFT"},
		{"TileForkliftRight", TileForkliftRight, '⊏', rgb(200, 160, 40), 50, false, false, true, false, 0, false, 4, "TILE_FORKLIFT"},
		{"TileFuelPump", TileFuelPump, '8', rgb(200, 60, 40), 30, false, false, true, false, 5, false, 4, "TILE_FUEL_PUMP"},
		{"TileContainerRed", TileContainerRed, '█', rgb(180, 50, 40), 80, false, false, false, false, 0, false, 4, "TILE_CONTAINER"},
		{"TileContainerBlue", TileContainerBlue, '█', rgb(50, 80, 180), 80, false, false, false, false, 0, false, 4, "TILE_CONTAINER"},
		{"TileContainerYellow", TileContainerYellow, '█', rgb(200, 170, 40), 80, false, false, false, false, 0, false, 4, "TILE_CONTAINER"},
		{"TileAdobe", TileAdobe, '█', rgb(200, 130, 70), 80, false, true, false, false, 0, false, 4, "TILE_ADOBE"},
		{"TileMetalWall", TileMetalWall, '█', rgb(180, 185, 195), 80, false, true, false, false, 0, false, 4, "TILE_METAL_WALL"},
		{"TileWreck", TileWreck, '▤', rgb(150, 95, 60), 80, false, true, false, false, 0, false, 4, "TILE_WRECK"},
		{"TileTimber", TileTimber, '≡', rgb(150, 110, 60), 80, false, true, true, true, 0, false, 4, "TILE_TIMBER"},
		{"TileDish", TileDish, '◗', rgb(170, 175, 185), 50, false, true, false, false, 0, false, 4, "TILE_DISH"},
		{"TileTruck", TileTruck, '▄', rgb(90, 110, 70), 80, false, true, true, false, 0, false, 4, "TILE_TRUCK"},
		{"TileIce", TileIce, '≈', rgb(180, 220, 235), 0, true, false, false, false, 0, false, 5, "TILE_ICE"},
		{"TileStreetlamp", TileStreetlamp, '⌖', rgb(220, 210, 120), 10, false, false, true, false, 0, false, 4, "TILE_STREETLAMP"},
		{"TileGlass", TileGlass, ',', rgb(190, 200, 210), 0, true, false, true, false, 0, true, 4, "TILE_GLASS"},
		{"TileDebris", TileDebris, '`', rgb(150, 140, 130), 0, true, false, true, false, 0, true, 4, "TILE_DEBRIS"},
		{"TileCryoPipe", TileCryoPipe, '╪', rgb(140, 200, 230), 30, false, false, true, false, 0, false, 4, "TILE_CRYO_PIPE"},
		{"TileSkylight", TileSkylight, '⊙', rgb(180, 210, 240), 0, true, false, true, false, 0, false, 4, "TILE_SKYLIGHT"},
		{"TileWheat", TileWheat, 'ψ', rgb(200, 180, 60), 20, true, false, true, true, 0, false, 5, "TILE_WHEAT"},
		{"TileHayBale", TileHayBale, '█', rgb(160, 140, 60), 60, false, true, true, true, 0, false, 4, "TILE_HAY_BALE"},
		{"TilePier", TilePier, '═', rgb(140, 100, 60), 10, true, false, false, false, 0, false, 3, "TILE_PIER"},
		{"TileDockCrate", TileDockCrate, '▣', rgb(150, 120, 80), 50, false, true, true, false, 0, false, 4, "TILE_DOCK_CRATE"},
		{"TileCliffFace", TileCliffFace, '░', rgb(140, 120, 100), 80, false, true, false, false, 0, false, 4, "TILE_CLIFF_FACE"},
		{"TileScree", TileScree, '·', rgb(160, 150, 130), 10, true, false, true, false, 0, true, 6, "TILE_SCREE"},
		{"TileBoulder", TileBoulder, '∩', rgb(130, 125, 120), 70, false, true, false, false, 0, false, 8, "TILE_BOULDER"},
		{"TileSwampWater", TileSwampWater, '≋', rgb(50, 100, 80), 5, true, false, false, false, 0, false, 8, "TILE_SWAMP_WATER"},
		{"TileCypressTree", TileCypressTree, '♣', rgb(40, 85, 50), 80, false, true, true, true, 0, false, 8, "TILE_CYPRESS_TREE"},
		{"TileSnowTree", TileSnowTree, '♣', rgb(220, 235, 245), 80, false, true, true, true, 0, false, 8, "TILE_SNOW_TREE"},
		{"TileMud", TileMud, '≋', rgb(110, 80, 50), 5, true, false, false, false, 0, false, 6, "TILE_MUD"},
		{"TileVine", TileVine, '‡', rgb(50, 130, 50), 20, true, false, true, true, 0, false, 6, "TILE_VINE"},
		{"TileBamboo", TileBamboo, '♣', rgb(80, 150, 60), 60, false, true, true, true, 0, false, 8, "TILE_BAMBOO"},
		{"TileDryBush", TileDryBush, '*', rgb(170, 140, 60), 20, true, false, true, true, 0, false, 4, "TILE_DRY_BUSH"},
		{"TileBusEnd", TileBusEnd, '▄', rgb(200, 180, 60), 50, false, true, true, false, 0, false, 4, "TILE_BUS"},
		{"TileBusMid", TileBusMid, '█', rgb(200, 180, 60), 50, false, true, true, false, 0, false, 4, "TILE_BUS"},
		{"TileHeloBody", TileHeloBody, '█', rgb(60, 70, 85), 50, false, true, true, false, 0, false, 4, "TILE_HELICOPTER"},
		{"TileHeloTail", TileHeloTail, '▄', rgb(60, 70, 85), 30, false, true, true, false, 0, false, 4, "TILE_HELICOPTER"},
		{"TileHeloNose", TileHeloNose, '▷', rgb(130, 200, 230), 40, false, true, true, false, 0, false, 4, "TILE_HELICOPTER"},
		{"TileHeloRotor", TileHeloRotor, '+', rgb(180, 180, 180), 0, true, false, true, false, 0, false, 4, "TILE_HELO_ROTOR"},
		{"TileHeloRotorSides", TileHeloRotorSides, '-', rgb(180, 180, 180), 0, true, false, true, false, 0, false, 4, "TILE_HELO_ROTOR"},
		{"TileHeloBodyBack", TileHeloBodyBack, '█', rgb(60, 70, 85), 50, false, true, true, false, 0, false, 4, "TILE_HELICOPTER"},
		{"TileHeloRotorBack", TileHeloRotorBack, 'x', rgb(180, 180, 180), 0, true, false, true, false, 0, false, 4, "TILE_HELO_ROTOR"},
		{"TileHeloWindow", TileHeloWindow, '◣', rgb(130, 200, 230), 0, false, true, true, false, 0, false, 4, "TILE_WINDOW"},
		{"TileTractorCab", TileTractorCab, '◣', rgb(130, 200, 230), 30, false, true, true, false, 0, false, 4, "TILE_TRACTOR"},
		{"TileTractorBody", TileTractorBody, '█', rgb(180, 60, 40), 30, false, true, true, false, 0, false, 4, "TILE_TRACTOR"},
		{"TileCrawlerLeft", TileCrawlerLeft, '◢', rgb(130, 70, 190), 50, false, true, true, false, 0, false, 4, "TILE_CRAWLER"},
		{"TileCrawlerMid", TileCrawlerMid, '█', rgb(130, 70, 190), 50, false, true, true, false, 0, false, 4, "TILE_CRAWLER"},
		{"TileCrawlerRight", TileCrawlerRight, '◣', rgb(130, 70, 190), 50, false, true, true, false, 0, false, 4, "TILE_CRAWLER"},
		{"TileCrawlerLeg", TileCrawlerLeg, '^', rgb(100, 50, 160), 20, false, true, true, false, 0, false, 4, "TILE_CRAWLER_LEG"},
		{"TileWheel", TileWheel, 'O', rgb(60, 60, 60), 10, false, false, true, false, 0, false, 4, "TILE_WHEEL"},
		{"TileWheelSmall", TileWheelSmall, 'o', rgb(60, 60, 60), 10, false, false, true, false, 0, false, 4, "TILE_WHEEL"},
	}
	for _, b := range builtins {
		def := &TileDef{
			Glyph:        b.glyph,
			GlyphStr:     string(b.glyph),
			Color:        b.color,
			ColorHex:     colorToHex(b.color),
			Cover:        b.cover,
			Passable:     b.passable,
			Opaque:       b.opaque,
			Destructible: b.destructible,
			Flammable:    b.flammable,
			Explodes:     b.explodes,
			Noisy:        b.noisy,
			MoveCost:     b.moveCost,
			NameKey:      b.nameKey,
		}
		tileRegistry[b.t] = def
		tileRegistryByName[b.id] = b.t
	}
}

func rgb(r, g, b int32) tcell.Color {
	return tcell.NewRGBColor(r, g, b)
}

func colorToHex(c tcell.Color) string {
	r, g, b := c.RGB()
	return "#" + strconv.FormatInt(int64((r>>16)&0xFF), 16) +
		strconv.FormatInt(int64((g>>8)&0xFF), 16) +
		strconv.FormatInt(int64(b&0xFF), 16)
}
