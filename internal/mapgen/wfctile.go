package mapgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Direction keys used in WFC tile JSON. They map to the battle package's
// internal direction indices (N=0, E=1, S=2, W=3).
var wfcDirKeys = []string{"N", "E", "S", "W"}

// WFCTileDef is the JSON representation of a single modular WFC piece.
// Rows is an array of equal-length strings; each character is a tile glyph
// ('.' floor, '#' wall, 'D' door, letters = furniture). Neighbors lists the
// tile IDs allowed to sit in each cardinal direction relative to this tile.
// Weight controls selection frequency (higher = more common, default 1).
type WFCTileDef struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Rows      []string          `json:"rows"`
	Neighbors map[string][]int  `json:"neighbors"`
	Weight    float64           `json:"weight,omitempty"`
}

// WFCLibrary is a parsed collection of WFC tiles loaded from JSON.
type WFCLibrary struct {
	Tiles []WFCTileDef `json:"tiles"`
}

// LoadWFCLibrary parses a single WFC library JSON file.
// Comments (// and /* */) are stripped before parsing.
func LoadWFCLibrary(path string) (*WFCLibrary, error) {
	data, err := ReadFileJSONC(path)
	if err != nil {
		return nil, fmt.Errorf("mapgen: read %s: %w", path, err)
	}
	var lib WFCLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("mapgen: parse %s: %w", path, err)
	}
	if err := validateWFCLibrary(&lib, path); err != nil {
		return nil, err
	}
	return &lib, nil
}

func validateWFCLibrary(lib *WFCLibrary, path string) error {
	if len(lib.Tiles) == 0 {
		return fmt.Errorf("mapgen: %s: no tiles", path)
	}
	ids := map[int]bool{}
	for _, t := range lib.Tiles {
		if ids[t.ID] {
			return fmt.Errorf("mapgen: %s: duplicate tile id %d", path, t.ID)
		}
		ids[t.ID] = true
	}
	for _, t := range lib.Tiles {
		if len(t.Rows) == 0 {
			return fmt.Errorf("mapgen: %s: tile %q has no rows", path, t.Name)
		}
		w := len([]rune(t.Rows[0]))
		for _, r := range t.Rows {
			if len([]rune(r)) != w {
				return fmt.Errorf("mapgen: %s: tile %q rows not uniform width", path, t.Name)
			}
		}
		for key, ns := range t.Neighbors {
			if !contains(wfcDirKeys, key) {
				return fmt.Errorf("mapgen: %s: tile %q invalid direction %q", path, t.Name, key)
			}
			for _, n := range ns {
				if !ids[n] {
					return fmt.Errorf("mapgen: %s: tile %q references unknown neighbor id %d", path, t.Name, n)
				}
			}
		}
	}
	return nil
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// LoadWFCLibrariesDir loads every *.json/*.jsonc file in dir as a WFC library
// and returns them keyed by filename stem (e.g. "ufo", "urban").
func LoadWFCLibrariesDir(dir string) (map[string]*WFCLibrary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("mapgen: read dir %s: %w", dir, err)
	}
	out := map[string]*WFCLibrary{}
	for _, e := range entries {
		if e.IsDir() || !IsJSONFile(e.Name()) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		lib, err := LoadWFCLibrary(path)
		if err != nil {
			return nil, err
		}
		stem := strings.TrimSuffix(e.Name(), ".json")
		stem = strings.TrimSuffix(stem, ".jsonc")
		out[stem] = lib
	}
	return out, nil
}
