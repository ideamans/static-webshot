---
name: static-webshot-usage
description: Capture deterministic screenshots of web pages and diff them for visual regression, using the static-webshot CLI. Animations, transitions, carousels, autoplay and clocks are suppressed so the same page renders identically twice. Use when the user asks to screenshot a page, check whether a page changed visually, set up or run visual regression testing, compare a staging page against production, or take a screenshot without animations moving.
license: MIT
compatibility: Requires the `static-webshot` binary on PATH — run the static-webshot-install skill if it is missing — and a Chrome or Chromium install for it to drive. No credentials or network service are involved.
allowed-tools: Bash(static-webshot:*) Bash(jq:*) Bash(command:*) Bash(ls:*) Read Write
---

# static-webshot-usage

Deterministic page screenshots, and diffs you can believe.

## 1. Confirm the tool and the browser

```bash
command -v static-webshot && static-webshot --version
```

Missing binary? Run the `static-webshot-install` skill. It also needs Chrome or
Chromium present; if the capture fails with a browser-not-found error, point at
one with `--chrome-path` rather than installing anything yourself.

## 2. Prove the page is stable before you diff anything

This is the step that separates a real regression from noise, and it is the one
most often skipped:

```bash
static-webshot capture "$URL" -o /tmp/a.png --mock-time 2026-01-01T00:00:00Z
static-webshot capture "$URL" -o /tmp/b.png --mock-time 2026-01-01T00:00:00Z
static-webshot compare /tmp/a.png /tmp/b.png --digest-json /tmp/self.json
```

Two captures of the *same* page must come out identical. If they do not, the
page still has something moving, and every later diff is noise. Work through
§4 before going any further. Say this to the user — do not quietly report a
diff you have not sanity-checked.

## 3. Pick the command by the question

| The user asks | Command |
| --- | --- |
| "screenshot this page" | `static-webshot capture <url> -o shot.png` |
| "…on mobile" | `capture <url> --preset mobile` |
| "…at this width" | `capture <url> --viewport 1280x720` |
| "did this page change?" | `capture` both, then `compare <baseline> <current>` |
| "show me what changed" | `compare a.png b.png -o diff.png` (three panels: baseline, diff, current) |
| "give me the numbers" | add `--digest-json result.json` |

`compare` takes **image paths, not URLs**.

## 4. Making a page hold still

Everything below is already automatic: animations and transitions off, caret
hidden, instant scrolling, autoplay off, video and audio frozen,
IntersectionObserver forced so lazy content is present, Web Animations API
disabled, and Swiper / Slick / Owl Carousel / Flickity / Bootstrap 5 carousels
stopped.

What is **not** automatic, in the order to try it:

1. `--mock-time 2026-01-01T00:00:00Z` — pins `Date`, `Math.random` and
   `performance.now`. The usual culprit. Use the same value for both captures.
2. `--wait-selector '.results'` — wait for late content. Prefer this over a
   longer `--wait-after`.
3. `--mask '.ad-slot'` — hide genuinely unstable regions (ads, live counters,
   avatars). Masked regions can no longer report a regression, so mask as
   little as possible and tell the user what you masked.
4. `--inject-css` — for a custom slider the built-in list does not know about.

## 5. Read the reference for anything else

```bash
static-webshot llm | sed -n '1,/^# Command catalog/p'
static-webshot capture --help
```

## 6. Interpret the diff honestly

Parse `--digest-json`, not the console output.

- **A non-zero pixel count is not automatically a regression.** Antialiasing and
  font rasterisation differ between machines. Compare captures taken on the same
  platform; try `--ignore-antialiasing` and raise `--color-threshold` before
  concluding anything.
- Report *what* changed and *where*, using the diff panel — not just a number.
- If you masked anything, say so in the same breath as the result.

## Failure modes

| Symptom | Fix |
| --- | --- |
| `command not found: static-webshot` | run the `static-webshot-install` skill |
| Chrome not found | install Chrome/Chromium, or pass `--chrome-path` |
| navigation timeout | raise `--timeout` (seconds), or wait with `--wait-selector` |
| TLS certificate error on staging | `--ignore-tls-errors` |
| two captures of the same page differ | remaining nondeterminism — work through §4 |
| huge diff, page looks identical | antialiasing or a different machine — `--ignore-antialiasing`, raise `--color-threshold` |
| screenshot short or missing content | lazy content had not arrived — `--wait-selector`, then `--wait-after` |
