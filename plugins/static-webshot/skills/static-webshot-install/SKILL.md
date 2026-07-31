---
name: static-webshot-install
description: Make the static-webshot command available, installing it only if it is missing. Use when another skill reports that `static-webshot` is not on PATH, or when the user asks to install, update or upgrade the ideamans deterministic screenshot CLI. Prefers an already-installed binary, then the latest GitHub release, then a build from source with go install.
license: MIT
compatibility: Requires curl (or wget) and tar to install from a release, or a Go toolchain for the source fallback. Standalone — does not need static-webshot to be present already. Installs from the public repository github.com/ideamans/static-webshot, so no GitHub authentication is needed. The tool additionally needs Chrome or Chromium at run time, which this skill does not install.
allowed-tools: Bash(curl:*) Bash(wget:*) Bash(tar:*) Bash(unzip:*) Bash(go:*) Bash(uname:*) Bash(command:*) Bash(which:*) Bash(mkdir:*) Bash(mv:*) Bash(cp:*) Bash(rm:*) Bash(chmod:*) Bash(ls:*) Bash(test:*) Bash(echo:*) Read
---

# static-webshot-install

Make the `static-webshot` command usable, doing the least work that achieves it.

## Route 1 — an existing installation on PATH

```bash
command -v static-webshot && static-webshot --version
```

If that resolves, **use it and stop here.** Do not check for a newer release —
it costs an API call and the user did not ask for an upgrade.

Two checks before trusting the hit:

- **It is the right tool.** `static-webshot llm | head -1` must read
  `# static-webshot — reference for AI agents`. If something else owns the
  name, say so and use an explicit path rather than shadowing theirs.
- **It is recent enough.** If `llm` is not a known command, the binary predates
  the embedded reference — continue to route 2 to upgrade it.

Continue past this section only when the command is missing, is the wrong tool,
is too old, or the user explicitly asked to update.

## Route 2 — the latest GitHub release

The repository is public, so no authentication is needed.

```bash
VERSION=$(curl -fsSL https://api.github.com/repos/ideamans/static-webshot/releases/latest \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)   # e.g. v0.3.0
```

Asset names keep the **`v` prefix** on the version, and use lowercase OS and
arch:

```
static-webshot_<version-with-v>_<os>_<arch>.tar.gz
```

`<os>` is `darwin`, `linux` or `windows`; `<arch>` is `amd64` or `arm64`, so
`uname -m` reporting `x86_64` maps to `amd64`. Windows ships a `.zip`.
**There is no `darwin_amd64` build** — Intel Macs fall through to route 3.

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')             # darwin | linux
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64  # amd64 | arm64
curl -fsSL -o /tmp/static-webshot.tar.gz \
  "https://github.com/ideamans/static-webshot/releases/download/${VERSION}/static-webshot_${VERSION}_${OS}_${ARCH}.tar.gz"
```

If the download 404s, list the actual assets on the release page rather than
retrying variations.

### Install onto PATH

```bash
tar -xzf /tmp/static-webshot.tar.gz -C /tmp
mkdir -p ~/.local/bin && mv /tmp/static-webshot ~/.local/bin/ && chmod +x ~/.local/bin/static-webshot
```

Prefer the first writable directory already on PATH — `~/.local/bin`, then
`/usr/local/bin`. Two things not to do on your own initiative:

- If nothing on PATH is writable, leave the binary in `/tmp`, print the exact
  `sudo mv` command and let the user run it. Do not run `sudo` yourself.
- If `~/.local/bin` is not on PATH, give the user the line to add to their shell
  profile. Do not edit the profile for them.

## Route 3 — build from source

Needs a Go toolchain. Note the binary is built from a subdirectory, and
`go install` names the command after **that directory**, not the module:

```bash
go install github.com/ideamans/static-webshot/cmd/staticwebshot@latest
```

That produces a command called **`staticwebshot`** (no hyphen), while the
release archive contains **`static-webshot`**. If you install this way, tell the
user which name they have, or rename it:

```bash
mv "$(go env GOPATH)/bin/staticwebshot" ~/.local/bin/static-webshot
```

## Verify

```bash
command -v static-webshot && static-webshot --version && static-webshot llm | head -1
```

Report the version and the path. Then say what is still needed: **Chrome or
Chromium must be installed** for captures to work. This skill does not install
a browser. If one is present somewhere unusual, captures can point at it with
`--chrome-path`.
