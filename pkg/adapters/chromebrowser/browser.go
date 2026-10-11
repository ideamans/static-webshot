// Package chromebrowser provides a browser implementation using chromedp.
package chromebrowser

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/ideamans/static-webshot/pkg/ports"
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
		// chromedp gives Chrome 20s to start. A freshly installed Chrome on a
		// cold machine (first run, CI, containers) can take longer.
		chromedp.WSURLReadTimeout(60*time.Second),
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
	b.ctx, b.cancel = chromedp.NewContext(b.allocCtx, chromedp.WithErrorf(logChromedpError))

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

// logChromedpError is chromedp's default error log minus events it cannot
// decode. A Chrome newer than the pinned cdproto sends enum values it does not
// know (IPAddressSpace "Loopback" on every localhost request), which printed
// a hundred-odd ERROR lines on a capture that succeeded. Nothing here listens
// for those events; chromedp itself drops events it does not know at all.
func logChromedpError(format string, v ...any) {
	if strings.HasPrefix(format, "could not unmarshal event") {
		return
	}
	log.Printf("ERROR: "+format, v...)
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
	return b.run(ctx, evaluateAwait(`document.fonts.ready.then(() => true)`))
}

// waitForImagesScript resolves once every image that is going to load has
// loaded and been decoded. decode() also covers decoding="async", whose pixels
// can still be missing after the load event.
//
// A lazy image that has not started loading and is outside the viewport is
// skipped: it never loads on its own, so waiting would only run into the
// timeout. With --full-page, LoadLazyContent has already switched those to
// eager, so they are waited for.
const waitForImagesScript = `
(() => {
  const inViewport = (img) => {
    const r = img.getBoundingClientRect();
    return r.width > 0 && r.height > 0 && r.bottom > 0 && r.right > 0 &&
      r.top < window.innerHeight && r.left < window.innerWidth;
  };
  const images = Array.from(document.querySelectorAll('img')).filter(img =>
    img.complete || img.loading !== 'lazy' || inViewport(img));
  return Promise.all(images.map(img => {
    const loaded = img.complete ? Promise.resolve() : new Promise((resolve) => {
      img.addEventListener('load', resolve, { once: true });
      img.addEventListener('error', resolve, { once: true });
    });
    return loaded.then(() => img.decode()).catch(() => {});
  })).then(() => true);
})()
`

// WaitForImages waits for all images to be loaded and decoded.
func (b *Browser) WaitForImages(ctx context.Context) error {
	return b.run(ctx, evaluateAwait(waitForImagesScript))
}

// evaluateAwait evaluates an expression that returns a Promise and waits for
// it to settle. chromedp.Evaluate does not await by default, so without this
// the waits above return immediately.
func evaluateAwait(expression string) chromedp.Action {
	var settled bool
	return chromedp.Evaluate(expression, &settled, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
		return p.WithAwaitPromise(true)
	})
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

// LoadLazyContent switches native lazy loading to eager. Scrolling is not an
// option: the deterministic scripts turn window.scrollTo into a no-op, and
// lazy loading driven by IntersectionObserver is already forced visible there.
func (b *Browser) LoadLazyContent(ctx context.Context) error {
	script := `
(() => {
  document.querySelectorAll('img[loading="lazy"], iframe[loading="lazy"]').forEach(el => {
    el.loading = 'eager';
  });
})()
`
	return b.run(ctx, chromedp.Evaluate(script, nil))
}

// PageHeight returns the height of the page content in CSS pixels.
func (b *Browser) PageHeight(ctx context.Context) (int, error) {
	var height int
	err := b.run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
		if err != nil {
			return err
		}
		height = int(math.Ceil(contentSize.Height))
		return nil
	}))
	return height, err
}

// Screenshot captures the visible viewport, or the whole page when
// opts.FullPage is set. A full-page capture keeps the viewport width so that
// captures of the same page always have the same width.
func (b *Browser) Screenshot(ctx context.Context, opts ports.ScreenshotOptions) ([]byte, error) {
	var buf []byte
	var action chromedp.Action = chromedp.CaptureScreenshot(&buf)
	if opts.FullPage {
		action = chromedp.ActionFunc(func(ctx context.Context) error {
			layoutViewport, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}
			height := float64(opts.Height)
			if height <= 0 {
				height = math.Ceil(contentSize.Height)
			}
			buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithCaptureBeyondViewport(true).
				WithFromSurface(true).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  float64(layoutViewport.ClientWidth),
					Height: height,
					Scale:  1,
				}).
				Do(ctx)
			return err
		})
	}
	if err := b.run(ctx, action); err != nil {
		return nil, err
	}
	return buf, nil
}

// run executes actions on the page, giving up when ctx is cancelled.
func (b *Browser) run(ctx context.Context, actions ...chromedp.Action) error {
	done := make(chan error, 1)
	go func() {
		done <- chromedp.Run(b.ctx, actions...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
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
