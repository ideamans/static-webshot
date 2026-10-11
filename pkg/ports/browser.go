// Package ports defines interfaces for external dependencies.
package ports

import (
	"context"
)

// BrowserOptions configures browser launch settings.
type BrowserOptions struct {
	Headless          bool              // Run in headless mode (default: true)
	ChromePath        string            // Path to Chrome executable
	UserAgent         string            // Custom user agent
	Headers           map[string]string // Custom HTTP headers
	ViewportWidth     int               // Viewport width in CSS pixels
	ViewportHeight    int               // Viewport height in CSS pixels
	IsMobile          bool              // Enable mobile emulation
	IgnoreHTTPSErrors bool              // Ignore HTTPS certificate errors
	ProxyServer       string            // HTTP proxy server URL
}

// ScreenshotOptions configures what area a screenshot covers.
type ScreenshotOptions struct {
	FullPage bool // Capture the whole page instead of the viewport
	Height   int  // Height to capture in CSS pixels when FullPage is set (0 = whole page)
}

// Browser abstracts browser automation for page screenshot capture.
type Browser interface {
	// Launch starts the browser with the given options.
	Launch(ctx context.Context, opts BrowserOptions) error

	// Navigate loads the specified URL and waits for the load event.
	Navigate(ctx context.Context, url string) error

	// InjectScript executes JavaScript in the page context before any other scripts.
	InjectScript(ctx context.Context, script string) error

	// InjectCSS injects CSS styles into the page.
	InjectCSS(ctx context.Context, css string) error

	// WaitForSelector waits for an element matching the CSS selector to appear.
	WaitForSelector(ctx context.Context, selector string) error

	// WaitForFonts waits for all fonts to be loaded.
	WaitForFonts(ctx context.Context) error

	// WaitForImages waits for all images to be loaded.
	WaitForImages(ctx context.Context) error

	// ApplyMasks hides elements matching the given CSS selectors.
	ApplyMasks(ctx context.Context, selectors []string) error

	// LoadLazyContent switches native lazy loading (loading="lazy") to eager
	// so that images and iframes below the fold are fetched before capture.
	LoadLazyContent(ctx context.Context) error

	// PageHeight returns the height of the page content in CSS pixels.
	PageHeight(ctx context.Context) (int, error)

	// Screenshot captures the viewport, or the whole page when opts.FullPage is set.
	Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)

	// Close shuts down the browser.
	Close() error
}
