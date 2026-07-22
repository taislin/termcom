package engine

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"unsafe"

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
	ActionDelay        int    `json:"action_delay"`        // ms between frames; 8 = ~120fps cap
	SfxVolume          int    `json:"sfx_volume"`          // 0..10
	Language           string `json:"language"`
	TutorialShown      bool   `json:"tutorial_shown"`
	TouchMode          bool   `json:"touch_mode"`
	TouchButtonSize    int    `json:"touch_button_size"`   // rows per touch button; 4 = compact
	DefaultCombatMode  int    `json:"default_combat_mode"` // 0=cautious, 1=attack, 2=breakoff
}

// Config is the live configuration. It is a pointer so that LoadConfig can
// publish a freshly unmarshalled struct via atomic pointer swap, never mutating
// a struct that other goroutines are concurrently reading. All reads use the
// Config.X field syntax, which remains valid because Config is a *GlobalConfig.
var Config = &GlobalConfig{
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
	DefaultCombatMode:  0,
}

func storeConfig(c *GlobalConfig) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Config)), unsafe.Pointer(c))
}

// applyConfig applies the side effects of a configuration (language, audio, theme).
func applyConfig(c *GlobalConfig) {
	if c.Language != "" {
		language.SetLanguage(c.Language)
	}
	audio.SetAudioEnabled(c.SoundEnabled)
	audio.SetSfxVolume(c.SfxVolume)
	ApplyTheme(c.Theme)
}

const ConfigFile = "config.json"

func LoadConfig() {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return
	}
	var c GlobalConfig
	if err := json.Unmarshal(data, &c); err != nil {
		log.Printf("config: ignoring corrupt %s: %v", ConfigFile, err)
		return
	}
	applyConfig(&c)
	storeConfig(&c)
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
