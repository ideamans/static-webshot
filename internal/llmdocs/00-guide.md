# static-webshot — reference for AI agents

`static-webshot` captures screenshots of web pages **deterministically** — the
same page rendered twice produces the same pixels — and compares two captures to
find what visually changed. It exists for visual regression testing, where a
carousel or a clock ticking over would otherwise report a false difference.

It drives a real Chrome, so a Chrome or Chromium install is required. Nothing
prompts; it reads flags only. Errors go to stderr and the exit code is non-zero
on failure. This reference is embedded in the binary, so `static-webshot llm`
always describes the exact version you are running.

## Ground rules

1. **Capture twice before you trust a diff.** Run `capture` on the same URL two
   times and `compare` the results. If that comes out non-empty, the page has a
   source of nondeterminism the built-in suppression does not cover, and any
   later diff is noise. Fix that first (see *When a page still moves*).
2. **`compare` takes images, not URLs.** `capture` first, then compare the two
   PNG files.
3. **`--mock-time` is what freezes clocks.** Without it, `Date`, `Math.random`
   and `performance.now` run normally and anything driven by them differs
   between runs. Pass the same ISO 8601 value to both captures.
4. **A non-zero pixel count is not automatically a regression.** Antialiasing
   and font rasterisation differ across machines. Compare captures taken on the
   same platform, and reach for `--ignore-antialiasing` and `--color-threshold`
   before concluding the page changed.

## What "deterministic" actually does

Every capture injects the following before the page runs. Knowing the list tells
you what is already handled and what is not:

- **Animations and transitions off** — `animation: none`, `transition: none`,
  zero durations and delays, applied under
  `@media (prefers-reduced-motion: no-preference)`.
- **Caret hidden** (`caret-color: transparent`) so a focused input does not
  blink between shots.
- **Instant scrolling** (`scroll-behavior: auto`).
- **Autoplay disabled**, video and audio elements frozen.
- **IntersectionObserver forced** so lazy-loaded content is present rather than
  appearing mid-capture.
- **Script-driven scrolling disabled** — `window.scrollTo`, `scrollIntoView`
  and friends are no-ops, so a page cannot scroll itself between shots.
- **Web Animations API disabled.**
- **Carousels stopped and reset** — Swiper, Slick, Owl Carousel, Flickity and
  Bootstrap 5 carousel specifically. **This list is not universal**: a custom or
  less common slider keeps moving.
- **`--mock-time <ISO8601>`** additionally pins `Date`, `Math.random` and
  `performance.now`.

## Commands

| Task | Command |
| --- | --- |
| Screenshot a page | `static-webshot capture <url>` |
| Diff two screenshots | `static-webshot compare <baseline> <current>` |

### capture

```bash
static-webshot capture https://example.com -o baseline.png \
  --preset desktop --mock-time 2026-01-01T00:00:00Z
```

`--preset` is `desktop` (default) or `mobile`; `--viewport WIDTH[xHEIGHT]`
overrides it and `--resize` scales the output. `--wait-selector` (repeatable)
waits for an element, `--wait-after` waits a fixed number of milliseconds.
`--mask` (repeatable) hides elements by CSS selector, and `--inject-css` adds
arbitrary CSS. `--headful` opens a visible browser for debugging.

**By default only the viewport (the first screen) is captured.** Add
`--full-page` to capture the whole page:

```bash
static-webshot capture https://example.com --full-page -o before.png
# ...refactor...
static-webshot capture https://example.com --full-page -o after.png
static-webshot compare before.png after.png --digest-json result.json
```

This is the check for a change that must not alter the look — a speed
optimisation (lazy loading, CSS cleanup, deferred scripts) or a code-quality
refactor. Breakage in such changes tends to sit *below* the fold, where a
viewport capture never looks.

- The width stays the viewport width; the height is the page's content height.
  `100vh` elements keep the viewport height — do **not** fake a full page with
  `--viewport 1280x9000`, which stretches them and still needs a guessed height.
- Images and iframes with `loading="lazy"` are switched to eager before the
  wait for images, so content below the fold is not captured blank.
  Script-driven lazy loading is covered by the forced IntersectionObserver.
- `position: fixed` / `sticky` elements appear once, at the top.
- Chrome cannot render taller than **16384px**; a longer page is cut there with
  a warning. `--max-height N` cuts earlier (it requires `--full-page`).
- Pages of different lengths still compare: `compare` sizes both to the larger
  one, so a change in page length shows up as a difference — usually what you
  want when checking a refactor. `compare --max-height` limits the comparison
  to the top N pixels.
- `--resize` with a full-page capture: give the width only (`--resize 800`);
  a fixed `WIDTHxHEIGHT` distorts pages of different lengths differently.

### compare

```bash
static-webshot compare baseline.png current.png -o diff.png \
  --ignore-antialiasing --digest-json result.json
```

The output is a three-panel image: baseline, diff, current. `--digest-json`
writes a machine-readable summary — **prefer it over parsing the console
output**. `--color-threshold` (0–255) sets how different a pixel must be to
count.

## When a page still moves

In this order:

1. `--mock-time` — the usual cause.
2. `--wait-selector` for content that arrives late, rather than a longer
   `--wait-after`.
3. `--mask '.ad-slot'` for regions that are genuinely unstable (ads, live
   counters, user avatars). Masked regions cannot report a regression, so mask
   as little as possible.
4. `--inject-css` for a custom slider the built-in list does not know.

## Failure modes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Chrome not found | no Chrome/Chromium on the machine | install one, or point at it with `--chrome-path` |
| `chrome failed to start` … `No usable sandbox!` | Ubuntu 23.10+ blocks the unprivileged user namespaces Chrome's sandbox needs | `sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0`, or install Chrome with its AppArmor profile; ask the user before changing the system |
| Navigation timeout | slow page or wrong URL | raise `--timeout` (seconds), or wait on an element with `--wait-selector` |
| TLS certificate error | staging host with a self-signed certificate | `--ignore-tls-errors` |
| Two captures of the same page differ | remaining nondeterminism | work through *When a page still moves* |
| Diff is large but the page looks identical | antialiasing or a different machine | `--ignore-antialiasing`, raise `--color-threshold`, compare on one platform |
| Screenshot is short or missing content | lazy content had not arrived | `--wait-selector`, then `--wait-after` |
| Only the first screen is captured; a break further down is missed | the default captures the viewport only | `--full-page` |
| Full-page capture ends before the footer | page taller than 16384px (warning on stderr) or `--max-height` | split the check with `--max-height`, or capture narrower sections of the site |

## What this CLI will not do

- It does not crawl. One URL per `capture` invocation; loop externally.
- It does not store baselines or track history — that is the caller's job.
- It does not log in. Use `--proxy`, or capture a page that is public.
- It cannot make every page deterministic. Custom sliders and third-party
  embeds may still need `--mask`.
