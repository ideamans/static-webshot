// Package chromebrowser provides a browser implementation using chromedp.
package chromebrowser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// Browser implements ports.Browser using chromedp.
type Browser struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

// New creates a new Browser.
func New() *Browser {
	return &Browser{}
}

// Launch starts the browser with the given options.
func (b *Browser) Launch(ctx context.Context, opts ports.BrowserOptions) error {
	// Start with chromedp defaults + minimal additions (like Playwright)
	chromedpOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("autoplay-policy", "user-gesture-required"),
		chromedp.Flag("disable-background-media-suspend", true),
	)

	// Handle headless mode
	if !opts.Headless {
		chromedpOpts = append(chromedpOpts, chromedp.Flag("headless", false))
	}

	// Chrome path resolution (explicit path > env > system > Playwright auto-install)
	chromePath := ResolveChromePath(opts.ChromePath)
	if chromePath != "" {
		chromedpOpts = append(chromedpOpts, chromedp.ExecPath(chromePath))
	}

	if opts.UserAgent != "" {
		chromedpOpts = append(chromedpOpts, chromedp.UserAgent(opts.UserAgent))
	}

	// Set window size
	width := opts.ViewportWidth
	height := opts.ViewportHeight
	if width == 0 {
		width = 1920
	}
	if height == 0 {
		height = 1080
	}
	chromedpOpts = append(chromedpOpts, chromedp.WindowSize(width, height))

	// Ignore HTTPS certificate errors
	if opts.IgnoreHTTPSErrors {
		chromedpOpts = append(chromedpOpts, chromedp.Flag("ignore-certificate-errors", true))
	}

	// HTTP proxy server
	if opts.ProxyServer != "" {
		chromedpOpts = append(chromedpOpts, chromedp.Flag("proxy-server", opts.ProxyServer))
	}

	b.allocCtx, b.allocCancel = chromedp.NewExecAllocator(ctx, chromedpOpts...)
	b.ctx, b.cancel = chromedp.NewContext(b.allocCtx)

	// Set custom headers if provided
	if len(opts.Headers) > 0 {
		headers := make(map[string]any)
		for k, v := range opts.Headers {
			headers[k] = v
		}
		if err := chromedp.Run(b.ctx, network.SetExtraHTTPHeaders(network.Headers(headers))); err != nil {
			return fmt.Errorf("set headers: %w", err)
		}
	}

	// Set viewport with device emulation
	if err := chromedp.Run(b.ctx,
		emulation.SetDeviceMetricsOverride(int64(width), int64(height), 1, opts.IsMobile),
	); err != nil {
		return fmt.Errorf("set viewport: %w", err)
	}

	return nil
}

// Navigate loads the specified URL and waits for the load event.
func (b *Browser) Navigate(ctx context.Context, url string) error {
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, chromedp.Navigate(url))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// InjectScript executes JavaScript in the page context before any other scripts.
func (b *Browser) InjectScript(ctx context.Context, script string) error {
	done := make(chan error, 1)
	go func() {
		err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
			return err
		}))
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// InjectCSS injects CSS styles into the page.
func (b *Browser) InjectCSS(ctx context.Context, css string) error {
	escapedCSS := strings.ReplaceAll(css, "\\", "\\\\")
	escapedCSS = strings.ReplaceAll(escapedCSS, "`", "\\`")

	script := fmt.Sprintf(`
(() => {
  const style = document.createElement('style');
  style.type = 'text/css';
  style.textContent = %s;
  (document.head || document.documentElement).appendChild(style);
})();
`, "`"+escapedCSS+"`")

	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, chromedp.Evaluate(script, nil))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// WaitForSelector waits for an element matching the CSS selector to appear.
func (b *Browser) WaitForSelector(ctx context.Context, selector string) error {
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, chromedp.WaitVisible(selector, chromedp.ByQuery))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// WaitForFonts waits for all fonts to be loaded.
func (b *Browser) WaitForFonts(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, chromedp.Evaluate(`document.fonts.ready`, nil))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// WaitForImages waits for all images to be loaded.
func (b *Browser) WaitForImages(ctx context.Context) error {
	script := `
(() => {
  const images = Array.from(document.querySelectorAll('img'));
  return Promise.all(images.map(img => {
    if (img.complete) return Promise.resolve();
    return new Promise((resolve) => {
      img.addEventListener('load', resolve);
      img.addEventListener('error', resolve);
    });
  }));
})()
`
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, chromedp.Evaluate(script, nil))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// ApplyMasks hides elements matching the given CSS selectors.
func (b *Browser) ApplyMasks(ctx context.Context, selectors []string) error {
	if len(selectors) == 0 {
		return nil
	}

	var cssRules []string
	for _, selector := range selectors {
		cssRules = append(cssRules, fmt.Sprintf("%s { visibility: hidden !important; }", selector))
	}
	return b.InjectCSS(ctx, strings.Join(cssRules, "\n"))
}

// Screenshot captures a viewport screenshot (not full page).
func (b *Browser) Screenshot(ctx context.Context) ([]byte, error) {
	var buf []byte

	done := make(chan error, 1)
	go func() {
		// CaptureScreenshot captures the visible viewport only
		done <- chromedp.Run(b.ctx, chromedp.CaptureScreenshot(&buf))
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
}

// Close shuts down the browser.
func (b *Browser) Close() error {
	if b.cancel != nil {
		b.cancel()
	}
	time.Sleep(100 * time.Millisecond)
	if b.allocCancel != nil {
		b.allocCancel()
	}
	return nil
}

var _ ports.Browser = (*Browser)(nil)
