package mapgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AlienTemplate represents a single body-part pixel template loaded from JSON.
type AlienTemplate struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`       // "head", "eye", "torso", "leg", "weapon"
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Pixels     []string `json:"pixels"`
	// Head
	Sense string `json:"sense,omitempty"`
	// Eye
	Style string `json:"style,omitempty"`
	// Torso
	Manip     []string `json:"manip,omitempty"`
	BodyType  string   `json:"body_type,omitempty"`
	// Leg
	Locomotion string `json:"locomotion,omitempty"`
	// Weapon
	DamageType int `json:"damage_type,omitempty"`
}

// alienRegistry holds loaded alien templates grouped by type.
var alienRegistry = map[string][]*AlienTemplate{}

// ResetAliens clears the alien template registry.
func ResetAliens() {
	alienRegistry = map[string][]*AlienTemplate{}
}

// LoadAlienTemplates reads all .json files in dir and registers the alien
// templates they contain.
func LoadAlienTemplates(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("alien templates: read dir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := loadAlienFile(path); err != nil {
			return err
		}
	}
	return nil
}

func loadAlienFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("alien templates: read %s: %w", path, err)
	}
	var templates []AlienTemplate
	if err := json.Unmarshal(data, &templates); err != nil {
		return fmt.Errorf("alien templates: parse %s: %w", path, err)
	}
	for i := range templates {
		t := &templates[i]
		if t.ID == "" {
			return fmt.Errorf("alien template %s[%d]: missing id", path, i)
		}
		if len(t.Pixels) == 0 {
			return fmt.Errorf("alien template %s[%d] (%s): empty pixels", path, i, t.ID)
		}
		alienRegistry[t.Type] = append(alienRegistry[t.Type], t)
	}
	return nil
}

// GetAlienTemplates returns all templates of the given type ("head", "eye",
// "torso", "leg", "weapon").
func GetAlienTemplates(typ string) []*AlienTemplate {
	return alienRegistry[typ]
}

// ToTemplateData converts loaded AlienTemplates to AlienTemplateData slices
// suitable for passing to data.SpriteRegistry.RebuildFromTemplates.
func ToTemplateData(typ string) []AlienTemplateData {
	src := alienRegistry[typ]
	out := make([]AlienTemplateData, len(src))
	for i, t := range src {
		out[i] = AlienTemplateData{
			Pixels:     t.Pixels,
			Sense:      t.Sense,
			Style:      t.Style,
			Manip:      t.Manip,
			BodyType:   t.BodyType,
			Locomotion: t.Locomotion,
			DamageType: t.DamageType,
		}
	}
	return out
}

// AlienTemplateData is the bridge type that avoids circular imports.
type AlienTemplateData struct {
	Pixels     []string
	Sense      string
	Style      string
	Manip      []string
	BodyType   string
	Locomotion string
	DamageType int
}
