// Package record provides the record command execution logic.
package record

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"time"

	"golang.org/x/image/draw"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/adapters/chromebrowser"
	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// Executor executes the record command.
type Executor struct {
	browser    ports.Browser
	filesystem ports.FileSystem
	logger     ports.Logger
}

// NewExecutor creates a new Executor with the given dependencies.
func NewExecutor(browser ports.Browser, filesystem ports.FileSystem, logger ports.Logger) *Executor {
	return &Executor{
		browser:    browser,
		filesystem: filesystem,
		logger:     logger,
	}
}

// Execute runs the record command with the given configuration.
func (e *Executor) Execute(ctx context.Context, cfg Config) error {
	// Apply preset if specified
	preset := GetPreset(cfg.Preset)

	// Override with explicit values if provided
	viewportWidth := cfg.ViewportWidth
	if viewportWidth == 0 {
		viewportWidth = preset.ViewportWidth
	}
	viewportHeight := cfg.ViewportHeight
	if viewportHeight == 0 {
		viewportHeight = preset.ViewportHeight
	}

	e.logger.Info("Launching browser...")

	// Launch browser
	launchOpts := ports.BrowserOptions{
		Headless:          cfg.Headless,
		ChromePath:        cfg.ChromePath,
		UserAgent:         preset.UserAgent,
		ViewportWidth:     viewportWidth,
		ViewportHeight:    viewportHeight,
		IsMobile:          preset.IsMobile,
		IgnoreHTTPSErrors: cfg.IgnoreHTTPSErrors,
		ProxyServer:       cfg.ProxyServer,
	}

	if err := e.browser.Launch(ctx, launchOpts); err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	defer e.browser.Close()

	// Inject deterministic scripts before navigation
	e.logger.Info("Injecting deterministic scripts...")
	deterministicScripts := chromebrowser.GetAllDeterministicScripts(cfg.MockTime)
	if err := e.browser.InjectScript(ctx, deterministicScripts); err != nil {
		return fmt.Errorf("inject deterministic scripts: %w", err)
	}

	// Navigate to URL
	e.logger.Info("Navigating to %s...", cfg.URL)
	navCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		navCtx, cancel = context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Second)
		defer cancel()
	}

	if err := e.browser.Navigate(navCtx, cfg.URL); err != nil {
		return fmt.Errorf("navigate: %w", err)
	}

	// Wait after load if specified
	if cfg.WaitAfter > 0 {
		e.logger.Debug("Waiting %dms after load...", cfg.WaitAfter)
		time.Sleep(time.Duration(cfg.WaitAfter) * time.Millisecond)
	}

	// Inject CSS to disable animations
	e.logger.Debug("Injecting animation disabling CSS...")
	if err := e.browser.InjectCSS(ctx, chromebrowser.DisableAnimationsCSS); err != nil {
		e.logger.Warn("Failed to inject animation CSS: %v", err)
	}

	// Inject custom CSS if provided
	if cfg.InjectCSS != "" {
		e.logger.Debug("Injecting custom CSS...")
		if err := e.browser.InjectCSS(ctx, cfg.InjectCSS); err != nil {
			e.logger.Warn("Failed to inject custom CSS: %v", err)
		}
	}

	// Wait for selectors if specified
	for _, selector := range cfg.WaitSelectors {
		e.logger.Debug("Waiting for selector: %s", selector)
		if err := e.browser.WaitForSelector(ctx, selector); err != nil {
			e.logger.Warn("Failed to wait for selector %s: %v", selector, err)
		}
	}

	// Wait for fonts and images
	e.logger.Debug("Waiting for fonts...")
	if err := e.browser.WaitForFonts(ctx); err != nil {
		e.logger.Warn("Failed to wait for fonts: %v", err)
	}

	e.logger.Debug("Waiting for images...")
	if err := e.browser.WaitForImages(ctx); err != nil {
		e.logger.Warn("Failed to wait for images: %v", err)
	}

	// Apply masks if specified
	if len(cfg.Masks) > 0 {
		e.logger.Debug("Applying masks...")
		if err := e.browser.ApplyMasks(ctx, cfg.Masks); err != nil {
			e.logger.Warn("Failed to apply masks: %v", err)
		}
	}

	// Small delay to ensure everything is rendered
	time.Sleep(100 * time.Millisecond)

	// Take screenshot
	e.logger.Info("Taking screenshot...")
	screenshot, err := e.browser.Screenshot(ctx)
	if err != nil {
		return fmt.Errorf("take screenshot: %w", err)
	}

	// Resize if specified
	if cfg.ResizeWidth > 0 {
		screenshot, err = resizeScreenshot(screenshot, cfg.ResizeWidth, cfg.ResizeHeight)
		if err != nil {
			return fmt.Errorf("resize screenshot: %w", err)
		}
		if cfg.ResizeHeight > 0 {
			e.logger.Info("Saving to %s (%dx%d)...", cfg.OutputPath, cfg.ResizeWidth, cfg.ResizeHeight)
		} else {
			e.logger.Info("Saving to %s (width=%d)...", cfg.OutputPath, cfg.ResizeWidth)
		}
	} else {
		e.logger.Info("Saving to %s...", cfg.OutputPath)
	}

	// Save screenshot
	if err := e.filesystem.WriteFile(cfg.OutputPath, screenshot, 0644); err != nil {
		return fmt.Errorf("save screenshot: %w", err)
	}

	e.logger.Info("Done! Screenshot saved to %s", cfg.OutputPath)
	return nil
}

// resizeScreenshot resizes the screenshot to the specified dimensions.
// If height is 0, it maintains aspect ratio based on width.
func resizeScreenshot(data []byte, width, height int) ([]byte, error) {
	// Decode PNG
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode png: %w", err)
	}

	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Calculate target height if not specified (maintain aspect ratio)
	if height == 0 {
		height = origHeight * width / origWidth
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use high-quality resampling
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Encode back to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}

	return buf.Bytes(), nil
}
