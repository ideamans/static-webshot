// Package record provides device presets for the record command.
package record

// Preset defines viewport and device settings.
type Preset struct {
	ViewportWidth  int
	ViewportHeight int
	IsMobile       bool
	UserAgent      string
}

// Presets defines available device presets.
var Presets = map[string]Preset{
	"desktop": {
		ViewportWidth:  1920,
		ViewportHeight: 1080,
		IsMobile:       false,
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	},
	"mobile": {
		ViewportWidth:  390,
		ViewportHeight: 844,
		IsMobile:       true,
		UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	},
}

// GetPreset returns the preset by name, defaulting to desktop.
func GetPreset(name string) Preset {
	if preset, ok := Presets[name]; ok {
		return preset
	}
	return Presets["desktop"]
}
