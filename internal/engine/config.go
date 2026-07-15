package engine

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/taislin/termcom/internal/audio"
	"github.com/taislin/termcom/internal/language"
)

const WebsiteURL = "https://taislin.github.io/termcom/"

var GameVersion = "dev"

func init() {
	if GameVersion != "dev" {
		return
	}
	data, err := os.ReadFile("VERSION")
	if err != nil {
		return
	}
	GameVersion = strings.TrimSpace(string(data))
}

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
	TutorialShown      bool   `json:"tutorial_shown"`
	TouchMode          bool   `json:"touch_mode"`
	TouchButtonSize    int    `json:"touch_button_size"`
}

var Config = GlobalConfig{
	BloomEnabled:       true,
	LightingEnabled:    true,
	SoundEnabled:       true,
	AutosaveEnabled:    true,
	ScreenShake:        true,
	MouseEnabled:       true,
	GridLines:          false,
	Theme:              "default",
	ConfirmDialogs:     true,
	PauseOnAlienDetect: true,
	ActionDelay:        8,
	SfxVolume:          10,
	Language:           "en",
	TouchButtonSize:    4,
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
