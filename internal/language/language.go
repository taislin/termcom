package language

import "fmt"

var (
	current    = "en"
	strings    = map[string]string{}
	available  = []string{"en"}
	registered = map[string]bool{"en": true}
)

// SetLanguage switches the active language.
func SetLanguage(lang string) {
	if registered[lang] {
		current = lang
	}
}

// Current returns the active language code.
func Current() string {
	return current
}

// Available returns the list of registered language codes.
func Available() []string {
	return available
}

// String returns the translated string for the given key.
// Falls back to the key itself if not found.
func String(key string) string {
	if s, ok := strings[key]; ok {
		return s
	}
	return key
}

// Sprintf returns a formatted translated string.
// Usage: language.Sprintf("BATTLE_HIT", damage, name)
func Sprintf(key string, args ...interface{}) string {
	return fmt.Sprintf(String(key), args...)
}

// register adds a language's string map to the registry.
func register(lang string, strs map[string]string) {
	for k, v := range strs {
		strings[k] = v
	}
	if !registered[lang] {
		registered[lang] = true
		available = append(available, lang)
	}
}
