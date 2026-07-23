package battle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/datafs"
	"github.com/taislin/termcom/internal/mapgen"
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
	tileRegistry        = map[TileType]*TileDef{}
	tileRegistryByName  = map[string]TileType{}
	nextCustomTileType  TileType = TileCustomBase
	tileLibLoaded       int32
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
	ensureTilesLoaded()
	return tileRegistry[t]
}

// TileDefGlyph returns the glyph for a tile type from the registry.
func TileDefGlyph(t TileType) rune {
	if d := tileRegistry[t]; d != nil {
		return d.Glyph
	}
	return '?'
}

// TileDefColor returns the color for a tile type from the registry.
func TileDefColor(t TileType) tcell.Color {
	if d := tileRegistry[t]; d != nil {
		return d.Color
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
	return tcell.GetColor(hex)
}

// LoadCustomTiles loads custom tile definitions from a JSON/JSONC file.
// The file format is: { "tiles": [ { "id": "...", ... }, ... ] }
func LoadCustomTiles(path string) error {
	data, err := mapgen.ReadFileJSONC(path)
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

// InitCustomTiles loads all .json/.jsonc files from data/tiles/ directory.
func InitCustomTiles() {
	// Try embedded/virtual filesystem first (WASM path)
	if entries, err := datafs.ReadDir("data/tiles"); err == nil {
		for _, e := range entries {
			if e.IsDir() || !mapgen.IsJSONFile(e.Name()) {
				continue
			}
			path := "data/tiles/" + e.Name()
			if err := LoadCustomTiles(path); err != nil {
				continue
			}
		}
		return
	}
	// Fallback: real filesystem
	dirs := []string{"data/tiles", "../data/tiles", "../../data/tiles"}
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil || len(entries) == 0 {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !mapgen.IsJSONFile(e.Name()) {
				continue
			}
			path := filepath.Join(d, e.Name())
			if err := LoadCustomTiles(path); err != nil {
				continue
			}
		}
		return
	}
}

// Strip comments from JSONC text (both // and /* */ comments).
func stripJSONCComments(raw []byte) []byte {
	// We do a simple single-pass stripper.
	// This is good enough for our controlled tile definition files.
	var out []byte
	inString := false
	escape := false
	for i := 0; i < len(raw); i++ {
		b := raw[i]
		if escape {
			out = append(out, b)
			escape = false
			continue
		}
		if b == '\\' && inString {
			escape = true
			out = append(out, b)
			continue
		}
		if b == '"' {
			inString = !inString
			out = append(out, b)
			continue
		}
		if inString {
			out = append(out, b)
			continue
		}
		// Check for // comment
		if b == '/' && i+1 < len(raw) && raw[i+1] == '/' {
			for i < len(raw) && raw[i] != '\n' {
				i++
			}
			out = append(out, '\n')
			continue
		}
		// Check for /* */ comment
		if b == '/' && i+1 < len(raw) && raw[i+1] == '*' {
			for i+1 < len(raw) && !(raw[i] == '*' && i+1 < len(raw) && raw[i+1] == '/') {
				i++
			}
			i += 2 // skip */
			continue
		}
		out = append(out, b)
	}
	return out
}

// jsoncTileDef matches the compact field format used in data/tiles/*.jsonc.
type jsoncTileDef struct {
	ID           string `json:"id"`
	GlyphStr     string `json:"g"`
	ColorHex     string `json:"c"`
	Cover        int    `json:"cv"`
	Passable     *bool  `json:"pa"`
	Opaque       *bool  `json:"op"`
	Destructible *bool  `json:"de"`
	Flammable    *bool  `json:"fl"`
	Explodes     int    `json:"ex"`
	Noisy        *bool  `json:"no"`
	MoveCost     int    `json:"mv"`
	NameKey      string `json:"nm"`
}

// loadBuiltinTileLibrary loads tile definitions from data/tiles/*.jsonc
// and registers them as the built-in tile library.
func loadBuiltinTileLibrary() {
	if entries, err := datafs.ReadDir("data/tiles"); err == nil {
		var paths []string
		for _, e := range entries {
			if e.IsDir() || !mapgen.IsJSONFile(e.Name()) {
				continue
			}
			paths = append(paths, "data/tiles/"+e.Name())
		}
		if len(paths) > 0 {
			loadTileFiles(paths, datafs.ReadFile)
			return
		}
	}

	// Fallback: search real filesystem by walking up from cwd
	cwd, _ := os.Getwd()
	var matches []string
	for _, root := range []string{cwd} {
		dir := root
		for depth := 0; depth < 8; depth++ {
			p := filepath.Join(dir, "data", "tiles", "*.jsonc")
			m, err := filepath.Glob(p)
			if err == nil && len(m) > 0 {
				matches = m
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if len(matches) > 0 {
			break
		}
	}

	// Fallback: use runtime.Caller to find the source file directory
	if len(matches) == 0 {
		_, srcFile, _, ok := runtime.Caller(0)
		if ok {
			dir := filepath.Dir(srcFile)
			for depth := 0; depth < 6; depth++ {
				p := filepath.Join(dir, "data", "tiles", "*.jsonc")
				m, err := filepath.Glob(p)
				if err == nil && len(m) > 0 {
					matches = m
					break
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}
	}

	if len(matches) > 0 {
		loadTileFiles(matches, os.ReadFile)
	}
}

func loadTileFiles(matches []string, readFile func(string) ([]byte, error)) {
	for _, path := range matches {
		raw, err := readFile(path)
		if err != nil {
			continue
		}
		cleaned := stripJSONCComments(raw)
		var tiles []jsoncTileDef
		if err := json.Unmarshal(cleaned, &tiles); err != nil {
			continue
		}
		for _, jt := range tiles {
			if jt.ID == "" || jt.GlyphStr == "" {
				continue
			}
			def := &TileDef{
				GlyphStr:     jt.GlyphStr,
				Glyph:        []rune(jt.GlyphStr)[0],
				ColorHex:     jt.ColorHex,
				Color:        parseHexColor(jt.ColorHex),
				Cover:        jt.Cover,
				Passable:     boolPtrVal(jt.Passable, true),
				Opaque:       boolPtrVal(jt.Opaque, false),
				Destructible: boolPtrVal(jt.Destructible, true),
				Flammable:    boolPtrVal(jt.Flammable, false),
				Explodes:     jt.Explodes,
				Noisy:        boolPtrVal(jt.Noisy, false),
				MoveCost:     jt.MoveCost,
				NameKey:      jt.NameKey,
			}
			if def.MoveCost == 0 {
				def.MoveCost = 4
			}
			if t, ok := tileRegistryByName[jt.ID]; ok {
				tileRegistry[t] = def
			} else {
				RegisterTile(jt.ID, def)
			}
		}
	}
}

// ensureTilesLoaded ensures the built-in tile library is loaded.
// This is called lazily on first access to avoid issues with working directory
// during package init (especially in tests).
func ensureTilesLoaded() {
	if atomic.LoadInt32(&tileLibLoaded) == 0 {
		atomic.StoreInt32(&tileLibLoaded, 1)
		registerBuiltinNameMappings()
		loadBuiltinTileLibrary()
	}
}

// registerBuiltinNameMappings populates tileRegistryByName with the known
// TileType constant values so that loadBuiltinTileLibrary can update the
// registry at the correct constant-based indices.
func registerBuiltinNameMappings() {
	for id, t := range tileTypeByName {
		tileRegistryByName[id] = t
	}
}
