package mapgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MapgenChunk represents a CDDA-style map definition loaded from JSON.
// Rows are ASCII art; terrain/furniture map each character to a tile type name.
// Weight controls how often the chunk appears relative to others of the same
// tag (higher = more frequent, default 1).
type MapgenChunk struct {
	ID        string            `json:"id"`
	Tags      []string          `json:"tags"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Weight    int               `json:"weight,omitempty"`
	Rows      []string          `json:"rows"`
	Terrain   map[string]string `json:"terrain"`
	Furniture map[string]string `json:"furniture"`
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

// LoadDir parses every .json file in dir into the global registry.
func LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("mapgen: read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := LoadFile(path); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile parses a single JSON file. The file may contain a single chunk
// object or an array of chunks.
func LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("mapgen: read %s: %w", path, err)
	}

	var chunks []MapgenChunk
	if err := json.Unmarshal(data, &chunks); err == nil {
		for i := range chunks {
			c := &chunks[i]
			if err := validate(c); err != nil {
				return fmt.Errorf("mapgen: %s chunk %d: %w", path, i, err)
			}
			registry[c.ID] = c
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
	registry[single.ID] = &single
	return nil
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
