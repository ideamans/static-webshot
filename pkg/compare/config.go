// Package compare provides the compare command logic.
package compare

// Config holds configuration for the compare command.
type Config struct {
	// BaselinePath is the path to the baseline image.
	BaselinePath string

	// CurrentPath is the path to the current image.
	CurrentPath string

	// OutputPath is the path where the diff image will be saved.
	OutputPath string

	// ColorThreshold is the per-pixel color difference threshold (0-255).
	ColorThreshold int

	// IgnoreAntialiasing ignores antialiased pixels when comparing.
	IgnoreAntialiasing bool

	// MaxHeight limits comparison to the top N pixels (0 = no limit).
	MaxHeight int

	// DiffOverlay overlays diff markers on the current image.
	DiffOverlay bool

	// DigestTxtPath is the path for text digest output (optional).
	DigestTxtPath string

	// DigestJSONPath is the path for JSON digest output (optional).
	DigestJSONPath string

	// LabelFontPath is the path to a TrueType font file for labels (optional).
	LabelFontPath string

	// LabelFontSize is the font size for labels in points (default: 14).
	LabelFontSize float64

	// BaselineLabel is the label text for the baseline panel (default: "baseline").
	BaselineLabel string

	// DiffLabel is the label text for the diff panel (default: "diff").
	DiffLabel string

	// CurrentLabel is the label text for the current panel (default: "current").
	CurrentLabel string
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		OutputPath:     "./diff.png",
		ColorThreshold: 10,
		DiffOverlay:    true, // Default to overlay mode
		BaselineLabel:  "baseline",
		DiffLabel:      "diff",
		CurrentLabel:   "current",
	}
}
