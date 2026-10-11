package record

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ideamans/static-webshot/pkg/adapters/chromebrowser"
	"github.com/ideamans/static-webshot/pkg/adapters/logger"
	"github.com/ideamans/static-webshot/pkg/adapters/osfilesystem"
)

// The test page is 800px wide in a 800x600 viewport:
//
//	   0- 600  hero (100vh), under a fixed 50px header
//	 600-4600  filler
//	4600-4700  lazy-loaded red image (loading="lazy")
//	4700-4800  green footer
//
// The image sits 4000px below the fold because Chrome fetches lazy images
// within roughly 1250-2500px of the viewport on its own; closer than that the
// test would pass even without the eager switch.
const fullPageHTML = `<!doctype html>
<html><head><style>
body { margin: 0 }
header { position: fixed; top: 0; left: 0; right: 0; height: 50px; background: #000 }
.hero { height: 100vh; background: #00f }
.filler { height: 4000px; background: #fff }
img { display: block; width: 800px; height: 100px }
footer { height: 100px; background: #0f0 }
</style></head><body>
<header></header>
<div class="hero"></div>
<div class="filler"></div>
<img src="/red.png" loading="lazy">
<footer></footer>
</body></html>`

const tallHTML = `<!doctype html>
<html><body style="margin:0"><div style="height:20000px;background:#ccc"></div></body></html>`

func newTestSite(t *testing.T) *httptest.Server {
	t.Helper()

	red := image.NewRGBA(image.Rect(0, 0, 800, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 800; x++ {
			red.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var redPNG bytes.Buffer
	if err := png.Encode(&redPNG, red); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fullPageHTML))
	})
	mux.HandleFunc("/tall", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(tallHTML))
	})
	mux.HandleFunc("/red.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(redPNG.Bytes())
	})
	// The same page with the image served late, or never.
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(strings.Replace(fullPageHTML, `src="/red.png"`, `src="/slow.png" decoding="async"`, 1)))
	})
	mux.HandleFunc("/slow.png", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "image/png")
		w.Write(redPNG.Bytes())
	})
	stalled := make(chan struct{})
	mux.HandleFunc("/stalled", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(strings.Replace(fullPageHTML, `src="/red.png"`, `src="/stalled.png"`, 1)))
	})
	mux.HandleFunc("/stalled.png", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-stalled:
		case <-r.Context().Done():
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(stalled) }) // runs first, so Close does not wait on the handler
	return srv
}

// capture runs the real capture pipeline against Chrome and decodes the PNG.
func capture(t *testing.T, chrome string, cfg Config) image.Image {
	t.Helper()

	cfg.Preset = "desktop"
	cfg.ViewportWidth = 800
	cfg.ViewportHeight = 600
	cfg.Headless = true
	cfg.ChromePath = chrome
	if cfg.Timeout == 0 {
		cfg.Timeout = 60
	}
	cfg.OutputPath = filepath.Join(t.TempDir(), "capture.png")

	executor := NewExecutor(chromebrowser.New(), osfilesystem.New(), logger.New())
	if err := executor.Execute(context.Background(), cfg); err != nil {
		t.Fatalf("capture %s: %v", cfg.URL, err)
	}

	f, err := os.Open(cfg.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func requireChrome(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping browser test in -short mode")
	}
	chrome := chromebrowser.FindSystemChrome()
	if chrome == "" {
		t.Skip("no Chrome or Chromium installed (set CHROME_PATH to run this test)")
	}
	return chrome
}

func assertColor(t *testing.T, img image.Image, x, y int, want color.RGBA, what string) {
	t.Helper()
	r, g, b, _ := img.At(x, y).RGBA()
	got := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
	near := func(a, b uint8) bool { d := int(a) - int(b); return d > -8 && d < 8 }
	if !near(got.R, want.R) || !near(got.G, want.G) || !near(got.B, want.B) {
		t.Errorf("%s at (%d,%d): got rgb(%d,%d,%d), want rgb(%d,%d,%d)",
			what, x, y, got.R, got.G, got.B, want.R, want.G, want.B)
	}
}

func TestCaptureViewportOnly(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	img := capture(t, chrome, Config{URL: srv.URL + "/"})
	if got := img.Bounds().Size(); got != image.Pt(800, 600) {
		t.Fatalf("viewport capture size = %v, want 800x600", got)
	}
}

func TestCaptureFullPage(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	img := capture(t, chrome, Config{URL: srv.URL + "/", FullPage: true})
	if got := img.Bounds().Size(); got != image.Pt(800, 4800) {
		t.Fatalf("full-page capture size = %v, want 800x4800 (a 100vh section must not stretch)", got)
	}

	assertColor(t, img, 400, 25, color.RGBA{0, 0, 0, 255}, "fixed header at the top")
	assertColor(t, img, 400, 300, color.RGBA{0, 0, 255, 255}, "hero")
	assertColor(t, img, 400, 625, color.RGBA{255, 255, 255, 255}, "filler (fixed header must not repeat)")
	assertColor(t, img, 400, 4650, color.RGBA{255, 0, 0, 255}, "lazy-loaded image below the fold")
	assertColor(t, img, 400, 4750, color.RGBA{0, 255, 0, 255}, "footer")

	again := capture(t, chrome, Config{URL: srv.URL + "/", FullPage: true})
	if again.Bounds() != img.Bounds() {
		t.Fatalf("second capture size = %v, first was %v", again.Bounds().Size(), img.Bounds().Size())
	}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if img.At(x, y) != again.At(x, y) {
				t.Fatalf("two full-page captures of the same page differ at (%d,%d)", x, y)
			}
		}
	}
}

func TestCaptureFullPageMaxHeight(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	img := capture(t, chrome, Config{URL: srv.URL + "/", FullPage: true, MaxHeight: 1000})
	if got := img.Bounds().Size(); got != image.Pt(800, 1000) {
		t.Fatalf("--max-height 1000 capture size = %v, want 800x1000", got)
	}
}

func TestCaptureFullPageChromeLimit(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	img := capture(t, chrome, Config{URL: srv.URL + "/tall", FullPage: true})
	if got := img.Bounds().Size(); got != image.Pt(800, MaxCaptureHeight) {
		t.Fatalf("20000px page capture size = %v, want 800x%d", got, MaxCaptureHeight)
	}
}

// Regression for v0.4.0: the waits did not await their Promise, so a lazy image
// switched to eager was captured before it arrived and came out white.
func TestCaptureFullPageWaitsForSlowLazyImage(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	img := capture(t, chrome, Config{URL: srv.URL + "/slow", FullPage: true})
	assertColor(t, img, 400, 4650, color.RGBA{255, 0, 0, 255}, "lazy image served after 2s")
}

// Now that the wait really blocks, an image that never arrives must not hang
// the capture: --timeout cuts the wait and the capture goes ahead without it.
func TestCaptureFullPageGivesUpOnStalledImage(t *testing.T) {
	chrome := requireChrome(t)
	srv := newTestSite(t)

	start := time.Now()
	img := capture(t, chrome, Config{URL: srv.URL + "/stalled", FullPage: true, Timeout: 3})
	if elapsed := time.Since(start); elapsed > 20*time.Second {
		t.Errorf("capture took %v with --timeout 3 and a stalled image", elapsed)
	}
	assertColor(t, img, 400, 4750, color.RGBA{0, 255, 0, 255}, "footer")
}
