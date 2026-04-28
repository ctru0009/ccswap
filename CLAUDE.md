# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
go build -o ccswap .                # build binary
go install .                        # install to $GOPATH/bin
go test ./...                       # all tests
go test -run TestUse_Success ./cmd/...  # single test by pattern
go test -run TestE2E ./cmd/...      # integration tests only
```

No Makefile, no CI, no Docker. Go 1.25+ required. All tests use `t.TempDir()` — no external services.

## Architecture

Go CLI (Cobra) that swaps Claude Code providers by editing `~/.claude/settings.json`. Pure local file I/O, no network calls.

- **`cmd/`** — One file per subcommand (`use.go`, `status.go`, `list.go`, `add.go`, `edit.go`, `remove.go`, `init.go`, `import.go`). Commands are thin; logic lives in `internal/`. `completeProviderNames` in `use.go` provides tab-completion for `use`, `edit`, and `remove`.
- **`internal/config/`** — YAML config I/O: `providers.yaml` (provider profiles), `state.yaml` (active provider), path resolution, env var expansion, key masking.
- **`internal/claude/`** — JSON settings I/O: read/write `settings.json`, merge the 6 target env keys.

### Data flow for `ccswap use <provider>`

Load providers.yaml → look up provider → read settings.json → `MergeEnv()` updates only the 6 env keys → atomic write (temp + rename) → save state.yaml

### Critical design decisions

- **`settings.json` is parsed as `map[string]json.RawMessage`** — never define a typed struct, it would silently drop unknown fields.
- **Atomic writes everywhere**: temp file → `os.Rename()`. Never write directly to the final path.
- **Env var interpolation**: `os.ExpandEnv` on `auth_token`, `base_url`, model names. Empty auth token after expansion = error.
- **Exit codes**: 0 success, 1 user error, 2 system/panic. Controlled via `exitFunc` var in `root.go`.
- **Shell completion**: `ccswap completion bash|zsh|fish|powershell` generates scripts. `use`, `edit`, `remove` tab-complete provider names via `completeProviderNames`.
- **Owner-only permissions**: `SaveProviders` writes `providers.yaml` with mode `0600` to protect API keys.
- **Env var expansion in list**: `ccswap list` expands env vars before matching against `settings.json` for active-provider detection.

### The 6 target env keys

`ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_HAIKU_MODEL`, `API_TIMEOUT_MS`

Any change to this set requires updating `claude.MergeEnv()` + its tests.

## Testing patterns

- **Path function overrides**: Commands use package-level `var pathFunc = ...`. Tests override these and restore via `t.Cleanup()`. See `use_test.go` for the canonical pattern.
- **`exitFunc` override**: `root_test.go` captures exit codes without killing the process.
- **Helper functions**: `writeTestProviders()`, `writeTestSettings()`, `newRootCmd()`, `execCmd()`, `overridePathFuncs()` — reuse these, don't reinvent.
- **Shell completion tests**: `cmd/completion_test.go` — tests `completeProviderNames` for correct output, arg-present edge case, and missing file.

## Style

- Error wrapping with `fmt.Errorf("context: %w", err)` — no bare errors.
- Confirmations to stderr, data (status, list) to stdout.
- API keys masked via `config.MaskKey()`: `sk-ant-...y456`.
- Provider names: `^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$`