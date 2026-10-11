# Command catalog

Generated from the cobra command tree by `go generate ./...`.
Do not edit by hand — edit the command definitions instead.

## `static-webshot capture`

Capture a deterministic screenshot of a web page

Capture a deterministic screenshot of a web page.

The capture command navigates to the specified URL and captures a screenshot
with deterministic behavior (disabled animations, fixed time, etc.).

Examples:
  static-webshot capture https://example.com
  static-webshot capture https://example.com -o screenshot.png
  static-webshot capture https://example.com --preset mobile
  static-webshot capture https://example.com --viewport 1280x720
  static-webshot capture https://example.com --resize 800x600
  static-webshot capture https://example.com --resize 800
  static-webshot capture https://example.com --mask ".ad-banner" --mask ".cookie-notice"
  static-webshot capture https://example.com --full-page -o before.png
  static-webshot capture https://example.com --full-page --max-height 8000

Full-page captures:
  --full-page captures the whole page at the viewport width instead of only
  the first screen. Images with loading="lazy" are switched to eager first so
  the content below the fold is not blank. Pages taller than 16384px (Chrome's
  limit) are cut there with a warning; --max-height cuts earlier. Pages of
  different lengths can still be compared: compare pads the shorter image and
  reports the extra length as a difference.

```
static-webshot capture <url>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--chrome-path` | string | — | Path to Chrome executable |
| `--full-page` | bool | `false` | Capture the whole page, not just the viewport |
| `--headful` | bool | `false` | Run in headful mode (opposite of headless) |
| `--headless` | bool | `true` | Run in headless mode |
| `--ignore-tls-errors` | bool | `false` | Ignore TLS certificate errors |
| `--inject-css` | string | — | Custom CSS to inject |
| `--mask` | stringArray | `[]` | CSS selector for elements to hide (can be repeated) |
| `--max-height` | int | `0` | With --full-page, cut the capture at this height in CSS pixels (0 = Chrome's limit, 16384) |
| `--mock-time` | string | — | Fixed time for Date API (ISO 8601 format) |
| `-o`, `--output` | string | `./capture.png` | Output file path |
| `--preset` | string | `desktop` | Device preset (desktop, mobile) |
| `--proxy` | string | — | HTTP proxy URL |
| `--resize` | string | — | Output image size (WIDTH or WIDTHxHEIGHT) |
| `--timeout` | int | `30` | Timeout in seconds for navigation, and separately for waiting on fonts and images |
| `--user-agent` | string | — | Custom User-Agent string (overrides preset) |
| `-v`, `--verbose` | bool | `false` | Enable verbose output |
| `--viewport` | string | — | Viewport size (WIDTH or WIDTHxHEIGHT) |
| `--wait-after` | int | `0` | Wait time after page load in milliseconds |
| `--wait-selector` | stringArray | `[]` | CSS selector to wait for (can be repeated) |

## `static-webshot compare`

Compare two images and generate a diff image

Compare two images pixel by pixel and generate a diff image.

The compare command loads two images, compares them, and outputs a composite
image showing: baseline | diff | current (left to right).

The diff panel shows the baseline image at 50% brightness with red overlay
on pixels that differ between the two images.

Comparison results including diff percent are output to stdout.
Use --digest-txt or --digest-json to save results to a file.

Examples:
  static-webshot compare baseline.png current.png
  static-webshot compare baseline.png current.png -o diff.png
  static-webshot compare baseline.png current.png --digest-txt result.txt
  static-webshot compare baseline.png current.png --digest-json result.json

```
static-webshot compare <baseline> <current>
```

| flag | type | default | description |
| --- | --- | --- | --- |
| `--baseline-label` | string | `baseline` | Label text for the baseline panel |
| `--color-threshold` | int | `10` | Per-pixel color difference threshold (0-255) |
| `--current-label` | string | `current` | Label text for the current panel |
| `--diff-label` | string | `diff` | Label text for the diff panel |
| `--digest-json` | string | — | Path to save comparison digest as JSON (optional) |
| `--digest-txt` | string | — | Path to save comparison digest as text (optional) |
| `--ignore-antialiasing` | bool | `false` | Ignore antialiased pixels |
| `--label-font` | string | — | Path to TrueType font file for labels (optional) |
| `--label-font-size` | float64 | `14` | Font size for labels in points |
| `-o`, `--output` | string | `./diff.png` | Diff image output path |
| `-v`, `--verbose` | bool | `false` | Enable verbose output |
