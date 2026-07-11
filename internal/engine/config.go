package engine

import (
	"encoding/json"
	"log"
	"os"

	"github.com/civ13/termcom/internal/language"
)

type GlobalConfig struct {
	BloomEnabled      bool   `json:"bloom_enabled"`
	DistortionEnabled bool   `json:"distortion_enabled"`
	LightingEnabled   bool   `json:"lighting_enabled"`
	Language          string `json:"language"`
}

var Config = GlobalConfig{
	BloomEnabled:      true,
	DistortionEnabled: false,
	LightingEnabled:   true,
	Language:          "en",
}

const ConfigFile = "config.json"

func LoadConfig() {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &Config); err != nil {
		log.Printf("config: ignoring corrupt %s: %v", ConfigFile, err)
		return
	}
	if Config.Language != "" {
		language.SetLanguage(Config.Language)
	}
}

func SaveConfig() {
	data, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(ConfigFile, data, 0644); err != nil {
		log.Printf("config: failed to save %s: %v", ConfigFile, err)
	}
}
