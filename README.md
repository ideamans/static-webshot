# static-web-shot - Static Web Screenshot Tool

A CLI tool for capturing deterministic screenshots and comparing them for visual regression testing.

## Features

- **Deterministic Screenshots**: Captures consistent screenshots by disabling animations, fixing random values, and freezing time
- **Device Presets**: Built-in presets for desktop and mobile viewports
- **Visual Comparison**: Pixel-by-pixel image comparison with diff overlay output
- **Flexible Options**: Customizable viewport, resize, masking, and more

## Installation

```bash
go install github.com/ideamans/go-page-visual-regression-tester/cmd/static-web-shot@latest
```

Or build from source:

```bash
git clone https://github.com/ideamans/go-page-visual-regression-tester.git
cd go-page-visual-regression-tester
go build -o static-web-shot ./cmd/static-web-shot
```

## Usage

### Capture Screenshots

Capture a screenshot of a web page:

```bash
# Basic usage (desktop preset, 1920x1080)
static-web-shot capture https://example.com -o screenshot.png

# Mobile preset (390x844, iPhone User-Agent)
static-web-shot capture https://example.com -o mobile.png --preset mobile

# Custom viewport
static-web-shot capture https://example.com -o custom.png --viewport 1280x720

# Resize output
static-web-shot capture https://example.com -o small.png --resize 800
static-web-shot capture https://example.com -o thumb.png --resize 400x300

# Wait after page load
static-web-shot capture https://example.com -o loaded.png --wait-after 2000

# Hide specific elements
static-web-shot capture https://example.com -o clean.png --mask ".ad-banner" --mask ".cookie-notice"
```

### Compare Images

Compare two screenshots and generate a diff image:

```bash
# Basic comparison
static-web-shot compare baseline.png current.png -o diff.png

# With custom threshold (0.0-1.0)
static-web-shot compare baseline.png current.png -o diff.png --threshold 0.1
```

Exit codes:
- `0`: Images match within threshold (PASS)
- `1`: Images differ beyond threshold (FAIL)

## Capture Options

| Option | Description | Default |
|--------|-------------|---------|
| `-o, --output` | Output file path | `./capture.png` |
| `--preset` | Device preset (`desktop`, `mobile`) | `desktop` |
| `--viewport` | Viewport size (`WIDTHxHEIGHT` or `WIDTH`) | Preset value |
| `--resize` | Output image size (`WIDTHxHEIGHT` or `WIDTH`) | No resize |
| `--wait-after` | Wait time after page load (ms) | `0` |
| `--mask` | CSS selector for elements to hide (repeatable) | None |
| `--wait-selector` | CSS selector to wait for (repeatable) | None |
| `--inject-css` | Custom CSS to inject | None |
| `--mock-time` | Fixed time for Date API (ISO 8601) | None |
| `--proxy` | HTTP proxy URL | None |
| `--ignore-tls-errors` | Ignore TLS certificate errors | `false` |
| `--timeout` | Navigation timeout (seconds) | `30` |
| `--headful` | Run browser in headful mode | `false` |
| `--chrome-path` | Path to Chrome executable | Auto-detect |
| `-v, --verbose` | Enable verbose output | `false` |

## Compare Options

| Option | Description | Default |
|--------|-------------|---------|
| `-o, --output` | Diff image output path | `./diff.png` |
| `--threshold` | Acceptable pixel difference ratio (0.0-1.0) | `0.15` |
| `--color-threshold` | Per-pixel color difference (0-255) | `10` |
| `--ignore-antialiasing` | Ignore antialiased pixels | `false` |
| `-v, --verbose` | Enable verbose output | `false` |

## Device Presets

| Preset | Viewport | User-Agent |
|--------|----------|------------|
| `desktop` | 1920x1080 | Windows Chrome |
| `mobile` | 390x844 | iPhone Safari |

## Chrome Auto-Detection

The tool automatically finds Chrome in the following order:
1. `--chrome-path` option (explicit path)
2. `CHROME_PATH` environment variable
3. System-installed Chrome/Chromium
4. **Auto-install via Playwright** (if no Chrome found)

This means you can run static-web-shot on a clean system without Chrome - it will automatically download Chromium using Playwright's browser management.

## Deterministic Features

The tool automatically applies the following to ensure consistent screenshots:

**JavaScript Modifications:**
- Fixes `Date.now()` and `new Date()` to a constant value
- Fixes `Math.random()` to always return 0.5
- Fixes `Performance.now()` to return constant value
- Disables video/audio autoplay
- Makes all elements visible to IntersectionObserver (for lazy loading)
- Disables scroll-related behaviors
- Disables Web Animations API

**CSS Modifications:**
- Disables all CSS animations and transitions
- Hides text cursor (caret)
- Disables smooth scrolling

## License

MIT License
