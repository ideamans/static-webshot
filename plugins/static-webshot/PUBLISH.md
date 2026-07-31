# Publishing the static-webshot plugin

## Before every release

1. `go generate ./...` — regenerates `internal/llmdocs/90-commands.md`; commit
   any diff.
2. `go test ./...` — includes `TestPluginSkills`, which enforces that
   `plugin.json.version` equals `PluginVersion` in `cmd/staticwebshot/main.go`
   and that the SKILL.md frontmatter stays within the Agent Skills standard.
3. `claude plugin validate plugins/static-webshot`.
4. Bump `PluginVersion` and `plugin.json.version` together, in the same commit
   as the release tag. The release workflow refuses a mismatched tag.

## Registering in the marketplace (first release only)

The repository is public, so the entry goes in
`ideamans/claude-public-plugins`:

```json
{
  "name": "static-webshot",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/static-webshot.git",
    "path": "plugins/static-webshot"
  }
}
```

Use the `marketplace-register` skill rather than hand-editing. Register the
repository at <https://context7.com/add-package> as well — `context7.json` is
committed but registration is a manual step.

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install static-webshot@ideamans-plugins
/static-webshot-usage
```

```bash
gh skill install ideamans/static-webshot/plugins/static-webshot/skills/static-webshot-usage --agent copilot
```
