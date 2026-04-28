# ccswap — Agent Instructions

Go CLI tool that switches Claude Code model providers by editing `~/.claude/settings.json`. One binary, no network calls, purely local file I/O.

## Build & Run

```bash
go build -o ccswap .       # build binary
go install .               # install to $GOPATH/bin
```

CI via GitHub Actions (`.github/workflows/ci.yaml`): runs `go test ./...` and `go vet ./...` on every push/PR.
Releases via goreleaser (`.goreleaser.yaml`): push a `v*` tag to trigger cross-platform builds (5 targets) and automatic GitHub Release creation.
Install script at `install.sh`: curl-pipe-sh installer that detects OS/arch and downloads the latest release.
Go 1.25+ required (per `go.mod`).

## Test

```bash
go test ./...                       # all tests
go test ./cmd/...                    # cmd package only (unit + integration)
go test ./internal/config/...        # config package only
go test ./internal/claude/...         # claude package only
go test -run TestE2E ./cmd/...       # integration tests only
go test -run TestUse_ ./cmd/...      # specific command tests
```

Tests use `t.TempDir()` for isolation — no shared fixtures, no external services needed. All tests are safe to run in parallel.

## Project Structure

```
main.go              → entry point (just calls cmd.Execute())
cmd/                 → cobra commands (one file per subcommand)
  root.go            → root command, panic→exit2 recovery, SilenceErrors
  use.go             → ccswap use <provider>, shared completeProviderNames
  status.go          → ccswap status
  list.go            → ccswap list (expands env vars for active-provider matching)
  add.go             → ccswap add (interactive)
  edit.go            → ccswap edit <provider> (opens $EDITOR)
  remove.go          → ccswap remove <provider>
  init.go            → ccswap init (prefills API key from env)
  import.go          → ccswap import (detects env config from settings.json)
  integration_test.go→ end-to-end tests across commands
  completion_test.go → shell completion tests
.github/
  workflows/
    ci.yaml          → test + vet on push/PR
    release.yaml     → goreleaser release on v* tags
.goreleaser.yaml     → cross-platform build config (5 targets)
install.sh           → curl-pipe-sh installer
internal/
  config/            → providers.yaml, state.yaml, paths, validation, env expansion
  claude/            → settings.json read/write/merge
```

`internal/` is not importable outside this module — keep all non-CLI logic there.

## Key Architecture Decisions

- **settings.json parsed as `map[string]json.RawMessage`**, not a typed struct. This preserves all unknown keys byte-for-byte. Never define a `ClaudeSettings` struct — it would silently drop unknown fields via `json:"-"`.
- **Atomic writes everywhere**: write to temp file → `os.Rename()` to final path. Applies to `settings.json`, `providers.yaml`, and `state.yaml`.
- **Env var interpolation**: `os.ExpandEnv` applied to `auth_token`, `base_url`, and model names in `providers.yaml` via `$VAR` or `${VAR}` syntax. If `auth_token` is empty after expansion, the operation fails (prevents writing empty tokens).
- **Exit codes**: 0 = success, 1 = user error, 2 = system error/panic. Implemented via `exitFunc` var in root.go (overridable in tests).
- **CI & Releases**: push/PR triggers test+vet via `.github/workflows/ci.yaml`. Pushing a `v*` tag triggers `.github/workflows/release.yaml` → goreleaser builds 5 platform binaries → creates GitHub Release. Users install via `install.sh` or download from Releases page.
- **Shell completion**: `ccswap completion bash|zsh|fish|powershell` generates shell completion scripts. `use`, `edit`, and `remove` commands support tab-completion of provider names via `completeProviderNames` (defined in `use.go`).
- **Owner-only permissions**: `SaveProviders` writes `providers.yaml` with mode `0600` (owner read/write only) to protect API keys in the file.
- **Env var expansion in list**: `ccswap list` expands env vars in provider config before comparing to `settings.json` for active-provider detection, so providers using `$VAR` auth tokens still match correctly.

## Testing Patterns

- **Path function overrides**: Commands use package-level `var pathFunc = ...` for all file paths. Tests override these + restore via `t.Cleanup()`. See `use_test.go` for canonical example.
- **exitFunc override**: `root_test.go` overrides `exitFunc` to capture exit codes without killing the test process.
- **Integration tests** in `cmd/integration_test.go` chain multiple commands (init → add → use → status) with fully overridden paths.
- **Helper functions**: `writeTestProviders()`, `writeTestSettings()`, `newRootCmd()`, `newE2ERootCmd()`, `execCmd()`, `overridePathFuncs()` — reuse these, don't reinvent.
- **Shell completion tests** in `cmd/completion_test.go` verify `completeProviderNames` returns correct provider list, handles already-argmented invocations, and handles missing providers file gracefully.

## The 6 Target Env Keys

`ccswap` modifies **only** these keys in `settings.json`'s `env` block — nothing else:

```
ANTHROPIC_AUTH_TOKEN
ANTHROPIC_BASE_URL
ANTHROPIC_DEFAULT_SONNET_MODEL
ANTHROPIC_DEFAULT_OPUS_MODEL
ANTHROPIC_DEFAULT_HAIKU_MODEL
API_TIMEOUT_MS
```

Any change to this set requires updating `claude.MergeEnv()` + its tests.

## Config File Locations (cross-platform)

| File | Linux/macOS | Windows |
|------|-------------|---------|
| providers.yaml | `~/.config/ccswap/providers.yaml` | `%APPDATA%\ccswap\providers.yaml` |
| state.yaml | `~/.config/ccswap/state.yaml` | `%APPDATA%\ccswap\state.yaml` |
| settings.json | `~/.claude/settings.json` | `%USERPROFILE%\.claude\settings.json` |

Paths resolve via `os.UserConfigDir()` and `os.UserHomeDir()` — no hardcoded paths. Override with `--config <dir>` flag.

## Style Notes

- Errors use `fmt.Errorf("context: %w", err)` wrapping — never bare errors.
- Output goes to `cmd.ErrOrStderr()` for confirmations, `cmd.OutOrStdout()` for data (status, list). Status/list use `color` and `tablewriter`.
- API keys are masked via `config.MaskKey()`: `sk-ant-...y456` (first 7 + last 4 chars). Short keys (<11 chars) become `****`.
- Provider name regex: `^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$` (lowercase, alphanumeric + hyphens, max 32 chars, no leading/trailing hyphens).

## Common Pitfalls

- **Do not** define a typed struct for the full `settings.json` schema — it will silently drop unknown fields.
- **Do not** use `json.Number` or typed unmarshaling for the env block — it's `map[string]string`.
- `state.yaml` may not exist on first run — `LoadState` returns zero-value `State{}`, not an error.
- `settings.json` may not exist — `ReadSettings` returns `{"env":{}}`, not an error.
- `settings.json` containing just `null` is treated as invalid — the tool refuses to overwrite it.