
## Updating the AI-facing layer

**Added a flag, changed a default, changed what the browser scripts suppress —
update all three before finishing.**

| Update | Where | How |
| --- | --- | --- |
| ① Documentation | `README.md` | only when usage changes |
| ② Help text | cobra `Short` / `Long` / flag descriptions in `cmd/staticwebshot/` | both humans and agents read this first |
| ③ **LLM knowledge** | `internal/llmdocs/00-guide.md` | ground rules, the determinism list, failure modes |
| | `internal/llmdocs/90-commands.md` | **generated — never hand-edit** → `go generate ./...` |
| | `plugins/static-webshot/skills/*/SKILL.md` | when the workflow or its prerequisites change |
| | `context7.json` `rules` | when a new trap appears |

③ is the one that rots. Stale docs and stale `--help` get noticed by a human
reading them; **stale LLM knowledge is noticed by nobody** — the agent just
quietly gets it wrong.

Special case for this repository: `pkg/adapters/chromebrowser/scripts.go`
decides what a capture suppresses. The list of carousel libraries handled by
name is documented in three places (`00-guide.md`, the usage skill,
`context7.json`). Change the code and they all have to move, or an agent will
believe a slider is frozen when it is not and report a false regression.

### Verify

```bash
go generate ./...
git diff --exit-code -- internal/llmdocs
go test ./cmd/... ./pkg/...
go run ./cmd/staticwebshot llm | head
```

`pkg/record/fullpage_integration_test.go` drives a real Chrome through the
whole capture pipeline. It uses an installed Chrome (or `CHROME_PATH`) and
skips when there is none or under `-short` — it never triggers the Playwright
download. CI runs it on linux-amd64, macOS and Windows (linux-arm64 has no Chrome
in `test.yml` and skips). On Linux the workflows lift Ubuntu's AppArmor
user-namespace restriction first, or Chrome refuses to start.

`PluginVersion` (`cmd/staticwebshot/main.go`) and
`plugins/static-webshot/.claude-plugin/plugin.json` must always agree, and the
release tag must match both — the release workflow checks it.

Standard: <https://github.com/ideamans/go-llm-cli-kit/blob/main/LLM.md>
