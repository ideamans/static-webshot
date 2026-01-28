// Package record provides the record command logic.
package record

// Config holds configuration for the record command.
type Config struct {
	// URL is the target URL to capture.
	URL string

	// OutputPath is the path where the screenshot will be saved.
	OutputPath string

	// Preset is the device preset to use (desktop, mobile).
	Preset string

	// ViewportWidth is the viewport width in CSS pixels.
	ViewportWidth int

	// ViewportHeight is the viewport height in CSS pixels.
	ViewportHeight int

	// ResizeWidth is the output image width (0 = no resize).
	ResizeWidth int

	// ResizeHeight is the output image height (0 = use aspect ratio).
	ResizeHeight int

	// WaitAfter is the time to wait after page load in milliseconds.
	WaitAfter int

	// Headless specifies whether to run in headless mode.
	Headless bool

	// ProxyServer is the HTTP proxy URL.
	ProxyServer string

	// IgnoreHTTPSErrors ignores SSL certificate errors.
	IgnoreHTTPSErrors bool

	// Masks are CSS selectors for elements to hide.
	Masks []string

	// WaitSelectors are CSS selectors to wait for before capture.
	WaitSelectors []string

	// InjectCSS is custom CSS to inject into the page.
	InjectCSS string

	// MockTime is the fixed time to use for Date and related APIs.
	MockTime string

	// ChromePath is the path to the Chrome executable.
	ChromePath string

	// Timeout is the maximum time to wait for operations.
	Timeout int
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		OutputPath: "./capture.png",
		Preset:     "desktop",
		WaitAfter:  0,
		Headless:   true,
		Timeout:    30,
	}
}
