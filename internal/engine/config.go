package engine

import (
	"encoding/json"
	"log"
	"os"

	"github.com/civ13/termcom/internal/audio"
	"github.com/civ13/termcom/internal/language"
)

const GameVersion = "0.34"

type GlobalConfig struct {
	BloomEnabled       bool   `json:"bloom_enabled"`
	LightingEnabled    bool   `json:"lighting_enabled"`
	SoundEnabled       bool   `json:"sound_enabled"`
	AutosaveEnabled    bool   `json:"autosave_enabled"`
	ScreenShake        bool   `json:"screen_shake"`
	MouseEnabled       bool   `json:"mouse_enabled"`
	GridLines          bool   `json:"grid_lines"`
	Theme              string `json:"theme"`
	ConfirmDialogs     bool   `json:"confirm_dialogs"`
	PauseOnAlienDetect bool   `json:"pause_on_alien_detect"`
	ActionDelay        int    `json:"action_delay"`
	SfxVolume          int    `json:"sfx_volume"`
	Language           string `json:"language"`
}

var Config = GlobalConfig{
	BloomEnabled:    true,
	LightingEnabled: true,
	SoundEnabled:    true,
	AutosaveEnabled: true,
	ScreenShake:     true,
	MouseEnabled:    true,
	GridLines:       false,
	Theme:           "default",
	ConfirmDialogs:     true,
	PauseOnAlienDetect: true,
	ActionDelay:        8,
	SfxVolume:       10,
	Language:        "en",
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
	audio.SetAudioEnabled(Config.SoundEnabled)
	audio.SetSfxVolume(Config.SfxVolume)
	ApplyTheme(Config.Theme)
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
