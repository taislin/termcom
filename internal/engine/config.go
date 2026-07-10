package engine

import (
	"encoding/json"
	"os"

	"github.com/civ13/ycom/internal/language"
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
	_ = json.Unmarshal(data, &Config)
	if Config.Language != "" {
		language.SetLanguage(Config.Language)
	}
}

func SaveConfig() {
	data, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(ConfigFile, data, 0644)
}
