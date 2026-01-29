# static-webshot - Static Web Screenshot Tool

A CLI tool for pixel-based visual regression testing of web pages. It captures deterministic screenshots of web pages before and after changes, then compares them at the pixel level to detect visual differences.

Many real-world web pages contain animations, carousel sliders, and other time-dependent dynamic elements that produce different screenshots on every capture, making naive pixel comparison unreliable. static-webshot was developed to solve this problem by automatically suppressing these non-deterministic factors, enabling efficient and accurate visual regression testing even on highly dynamic pages.

## Features

- **Deterministic Screenshots**: Captures consistent screenshots by disabling CSS/JS animations, carousel sliders, fixing random values, and freezing time — eliminating noise from dynamic elements
- **Pixel-Based Visual Regression**: Compares baseline and current screenshots at the pixel level, reporting the exact number and percentage of changed pixels
- **Device Presets**: Built-in presets for desktop and mobile viewports
- **Diff Overlay Output**: Generates a side-by-side diff image highlighting the changed regions
- **Flexible Options**: Customizable viewport, resize, masking, and more

## Installation

```bash
go install github.com/ideamans/go-page-visual-regression-tester/cmd/staticwebshot@latest
```

Or build from source:

```bash
git clone https://github.com/ideamans/go-page-visual-regression-tester.git
cd go-page-visual-regression-tester
go build -o static-webshot ./cmd/staticwebshot
```

## Usage

### Capture Screenshots

Capture a screenshot of a web page:

```bash
# Basic usage (desktop preset, 1920x1080)
static-webshot capture https://example.com -o screenshot.png

# Mobile preset (390x844, iPhone User-Agent)
static-webshot capture https://example.com -o mobile.png --preset mobile

# Custom viewport
static-webshot capture https://example.com -o custom.png --viewport 1280x720

# Resize output
static-webshot capture https://example.com -o small.png --resize 800
static-webshot capture https://example.com -o thumb.png --resize 400x300

# Wait after page load
static-webshot capture https://example.com -o loaded.png --wait-after 2000

# Hide specific elements
static-webshot capture https://example.com -o clean.png --mask ".ad-banner" --mask ".cookie-notice"
```

### Compare Images

Compare two screenshots and generate a diff image:

```bash
# Basic comparison
static-webshot compare baseline.png current.png -o diff.png

# Save results as text digest
static-webshot compare baseline.png current.png -o diff.png --digest-txt result.txt

# Save results as JSON
static-webshot compare baseline.png current.png -o diff.png --digest-json result.json

# Both text and JSON
static-webshot compare baseline.png current.png -o diff.png --digest-txt result.txt --digest-json result.json
```

Comparison results including diff percent are always output to stdout:

```
[Compare Result]
Baseline: baseline.png
Current: current.png
Output: ./diff.png
Diff Pixels: 100 / 100000
Diff Percent: 0.1000%
```

JSON digest output (`--digest-json`):

```json
{
  "pixelDiffCount": 100,
  "pixelDiffRatio": 0.001,
  "diffPercent": 0.1,
  "totalPixels": 100000,
  "baselinePath": "baseline.png",
  "currentPath": "current.png",
  "diffPath": "./diff.png"
}
```

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
| `--user-agent` | Custom User-Agent string (overrides preset) | Preset value |
| `--headful` | Run browser in headful mode | `false` |
| `--chrome-path` | Path to Chrome executable | Auto-detect |
| `-v, --verbose` | Enable verbose output | `false` |

## Compare Options

| Option | Description | Default |
|--------|-------------|---------|
| `-o, --output` | Diff image output path | `./diff.png` |
| `--digest-txt` | Path to save comparison digest as text | None |
| `--digest-json` | Path to save comparison digest as JSON | None |
| `--color-threshold` | Per-pixel color difference (0-255) | `10` |
| `--ignore-antialiasing` | Ignore antialiased pixels | `false` |
| `--label-font` | Path to TrueType font file for labels | Built-in |
| `--label-font-size` | Font size for labels in points | `14` |
| `--baseline-label` | Label text for the baseline panel | `baseline` |
| `--diff-label` | Label text for the diff panel | `diff` |
| `--current-label` | Label text for the current panel | `current` |
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

This means you can run static-webshot on a clean system without Chrome - it will automatically download Chromium using Playwright's browser management.

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
