package mapgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MapgenChunk represents a CDDA-style map definition loaded from JSON.
// Rows are ASCII art; terrain maps each character to a tile type name.
// The furniture field is deprecated — merged into terrain at load time.
type MapgenChunk struct {
	ID          string            `json:"id"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	Weight      int               `json:"weight,omitempty"`
	NoRotate    bool              `json:"no_rotate,omitempty"`
	Rows        []string          `json:"rows"`
	Terrain     map[string]string `json:"terrain"`
	Furniture   map[string]string `json:"furniture,omitempty"` // deprecated, merged into Terrain
}

func (c *MapgenChunk) EffectiveWeight() int {
	if c.Weight < 1 {
		return 1
	}
	return c.Weight
}

// registry holds all loaded chunks keyed by ID.
var registry = map[string]*MapgenChunk{}

// Reset clears the registry (used by tests).
func Reset() {
	registry = map[string]*MapgenChunk{}
}

// Get returns a chunk by ID, or nil if not found.
func Get(id string) *MapgenChunk {
	return registry[id]
}

// ByTag returns all chunks tagged with the given tag.
func ByTag(tag string) []*MapgenChunk {
	var out []*MapgenChunk
	for _, c := range registry {
		for _, t := range c.Tags {
			if t == tag {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// stripJSONCComments removes // and /* */ comments from JSONC data while
// preserving string contents. It returns clean JSON that can be parsed by
// encoding/json.
func stripJSONCComments(data []byte) []byte {
	out := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		// String literal — copy verbatim
		if data[i] == '"' {
			out = append(out, '"')
			i++
			for i < len(data) {
				c := data[i]
				out = append(out, c)
				i++
				if c == '\\' && i < len(data) {
					out = append(out, data[i])
					i++
				} else if c == '"' {
					break
				}
			}
			continue
		}
		// // comment
		if data[i] == '/' && i+1 < len(data) && data[i+1] == '/' {
			i += 2
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		}
		// /* */ comment
		if data[i] == '/' && i+1 < len(data) && data[i+1] == '*' {
			i += 2
			for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			if i+1 < len(data) {
				i += 2 // skip */
			}
			continue
		}
		out = append(out, data[i])
		i++
	}
	return out
}

// ReadFileJSONC reads a file and strips JSONC comments, returning clean JSON.
func ReadFileJSONC(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return stripJSONCComments(data), nil
}

// IsJSONFile reports whether name has a .json or .jsonc extension.
func IsJSONFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".jsonc")
}


// LoadDir parses every .json/.jsonc file in dir into the global registry.
func LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("mapgen: read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !IsJSONFile(e.Name()) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := LoadFile(path); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile parses a single JSON/JSONC file. The file may contain a single chunk
// object or an array of chunks. Comments (// and /* */) are stripped before
// parsing.
func LoadFile(path string) error {
	data, err := ReadFileJSONC(path)
	if err != nil {
		return fmt.Errorf("mapgen: read %s: %w", path, err)
	}

	finish := func(c *MapgenChunk) {
		mergeFurniture(c)
		registry[c.ID] = c
	}

	var chunks []MapgenChunk
	if err := json.Unmarshal(data, &chunks); err == nil {
		for i := range chunks {
			c := &chunks[i]
			if err := validate(c); err != nil {
				return fmt.Errorf("mapgen: %s chunk %d: %w", path, i, err)
			}
			finish(c)
		}
		return nil
	}

	var single MapgenChunk
	if err := json.Unmarshal(data, &single); err != nil {
		return fmt.Errorf("mapgen: parse %s: %w", path, err)
	}
	if err := validate(&single); err != nil {
		return fmt.Errorf("mapgen: %s: %w", path, err)
	}
	finish(&single)
	return nil
}

// mergeFurniture merges Furniture entries into Terrain so there's only one
// lookup. Furniture wins on key conflict.
func mergeFurniture(c *MapgenChunk) {
	if c.Furniture == nil {
		return
	}
	if c.Terrain == nil {
		c.Terrain = c.Furniture
		c.Furniture = nil
		return
	}
	for k, v := range c.Furniture {
		c.Terrain[k] = v
	}
	c.Furniture = nil
}

func validate(c *MapgenChunk) error {
	if c.ID == "" {
		return fmt.Errorf("chunk missing id")
	}
	if len(c.Rows) == 0 {
		return fmt.Errorf("chunk %s: rows is empty", c.ID)
	}
	if c.Height != len(c.Rows) {
		return fmt.Errorf("chunk %s: height %d but got %d rows", c.ID, c.Height, len(c.Rows))
	}
	return nil
}

// Init loads all mapgen and alien template data from the data directory.
// It also rebuilds the sprite registry from loaded JSON templates.
func Init() error {
	dirs := []string{"data/maps", "../data/maps", "../../data/maps"}
	loaded := false
	for _, d := range dirs {
		if err := LoadDir(d); err == nil {
			loaded = true
			break
		}
	}
	if !loaded {
		return fmt.Errorf("mapgen: could not find data/maps/ in any search path")
	}

	alienDirs := []string{"data/aliens", "../data/aliens", "../../data/aliens"}
	for _, d := range alienDirs {
		if err := LoadAlienTemplates(d); err == nil {
			return nil
		}
	}
	return fmt.Errorf("mapgen: could not find data/aliens/ in any search path")
}
