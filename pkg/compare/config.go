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

	// Threshold is the acceptable pixel difference ratio (0.0 - 1.0).
	Threshold float64

	// ColorThreshold is the per-pixel color difference threshold (0-255).
	ColorThreshold int

	// IgnoreAntialiasing ignores antialiased pixels when comparing.
	IgnoreAntialiasing bool

	// MaxHeight limits comparison to the top N pixels (0 = no limit).
	MaxHeight int

	// DiffOverlay overlays diff markers on the current image.
	DiffOverlay bool

	// JSONOutputPath is the path for JSON result output.
	JSONOutputPath string

	// DigestPath is the path for digest output (optional).
	DigestPath string
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		OutputPath:     "./diff.png",
		Threshold:      0.15,
		ColorThreshold: 10,
		DiffOverlay:    true, // Default to overlay mode
	}
}
