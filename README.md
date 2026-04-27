# ccswap

> Switch Claude Code model providers in a single command

`ccswap` is a fast, safe CLI for managing and switching between multiple Claude Code provider profiles (Anthropic, Z.ai, Ollama Cloud, OpenRouter, etc.). It updates your `~/.claude/settings.json` atomically — never touching permissions, MCP servers, or any other Claude Code configuration.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue)](https://go.dev)

---

## Why ccswap?

Claude Code enforces a 5-hour usage limit on Claude Pro/Max subscriptions. When you hit that wall, switching to an alternative provider (Z.ai, Ollama Cloud, OpenRouter) should take seconds — not require manual JSON editing and a session restart.

`ccswap` solves this by letting you define named provider profiles and swap between them instantly:

```bash
ccswap use zai        # Switch to Z.ai GLM models
ccswap use anthropic  # Switch back to Anthropic
```

All swaps are **atomic** and **non-destructive** — your existing Claude Code settings are always preserved.

---

## Features

- ✨ **Single-command switching** — `ccswap use <provider>` in under 500ms
- 🔒 **Secure by default** — API keys stored separately, masked in all output
- 🛡️ **Non-destructive** — only touches the `env` block in `settings.json`
- 🔄 **Atomic writes** — temp file + rename; never leaves a corrupted config
- 💾 **Automatic backups** — every write creates a `.ccswap.bak` first
- 🔧 **Environment variable interpolation** — `$ANTHROPIC_API_KEY` in YAML config
- 📋 **Provider management** — add, edit, list, remove providers interactively
- 🖥️ **Cross-platform** — Linux, macOS, and Windows

---

## Installation

### From source (requires Go 1.25+)

```bash
go install github.com/ctru0009/ccswap@latest
```

Or clone and build manually:

```bash
git clone https://github.com/ctru0009/ccswap
cd ccswap
go build -o ccswap .
```

### Pre-built binaries (soon)

Binary releases for Linux, macOS, and Windows will be available via GitHub Releases.

```bash
# Coming soon:
curl -sSL https://raw.githubusercontent.com/ctru0009/ccswap/main/install.sh | sh
```

---

## Quick Start

```bash
# 1. Initialize config (creates ~/.config/ccswap/providers.yaml)
ccswap init

# 2. Add your first provider interactively
ccswap add

# 3. Switch to it
ccswap use my-provider

# 4. Check what's active
ccswap status
```

**Note:** Open a new Claude Code session after switching for changes to take effect.

---

## Commands

| Command | Description |
|---------|-------------|
| `ccswap init` | Create `providers.yaml` with a commented template |
| `ccswap add` | Interactively add a new provider profile |
| `ccswap use <provider>` | Switch to the specified provider |
| `ccswap status` | Show the currently active provider |
| `ccswap list` | List all providers in a table |
| `ccswap edit <provider>` | Open provider in `$EDITOR` for editing |
| `ccswap remove <provider>` | Remove a provider profile |

### `ccswap use`

Switch the active Claude Code provider:

```bash
$ ccswap use zai
✓ Switched to zai
  Base URL:    https://api.z.ai/api/anthropic
  Sonnet:      glm-5-turbo
  Opus:        glm-5.1
  Haiku:       glm-4.5-air
  Auth:        zai-123...4567

Open a new Claude Code session for changes to take effect.
```

### `ccswap status`

Show currently active provider:

```bash
$ ccswap status
Active provider: zai
Base URL:        https://api.z.ai/api/anthropic
Sonnet:          glm-5-turbo
Opus:            glm-5.1
Haiku:           glm-4.5-air
Last switched:   2026-04-27T14:32:11+10:00
```

### `ccswap list`

List all configured providers. The active provider is marked with `*`:

```bash
$ ccswap list
PROVIDER     BASE URL                           SONNET           OPUS               HAIKU
*zai          https://api.z.ai/api/anthropic   glm-5-turbo      glm-5.1            glm-4.5-air
anthropic    https://api.anthropic.com        claude-sonnet-4  claude-opus-4      claude-haiku-4
```

### `ccswap add`

Interactive prompt to add a new provider:

```bash
$ ccswap add
Provider name: ollama-cloud
Auth token:
Base URL: https://ollama.com/v1
Model (sonnet): kimi-k2.6:cloud
Model (opus): deepseek-v4-flash:cloud
Model (haiku): glm-4.7:cloud
Timeout (ms) [3000000]:

✓ Provider ollama-cloud added. Run: ccswap use ollama-cloud
```

*Auth token is hidden when input is a terminal (uses `term.ReadPassword`).*

---

## Configuration

Config files are stored at platform-appropriate locations:

| File | Linux/macOS | Windows |
|------|-------------|---------|
| `providers.yaml` | `~/.config/ccswap/providers.yaml` | `%APPDATA%\ccswap\providers.yaml` |
| `state.yaml` | `~/.config/ccswap/state.yaml` | `%APPDATA%\ccswap\state.yaml` |
| `settings.json` | `~/.claude/settings.json` | `%USERPROFILE%\.claude\settings.json` |

### `providers.yaml` Schema

```yaml
providers:
  anthropic:
    auth_token: $ANTHROPIC_API_KEY
    base_url: https://api.anthropic.com
    timeout_ms: 600000
    models:
      sonnet: claude-sonnet-4-20250514
      opus: claude-opus-4-20250514
      haiku: claude-haiku-3-5-20250101

  zai:
    auth_token: $ZAI_API_KEY
    base_url: https://api.z.ai/api/anthropic
    timeout_ms: 3000000
    models:
      sonnet: glm-5-turbo
      opus: glm-5.1
      haiku: glm-4.5-air
```

### Environment Variable Interpolation

All string fields (`auth_token`, `base_url`, and model names) support `$VAR` or `${VAR}` syntax. This is the recommended way to keep API keys out of plaintext files:

```yaml
auth_token: $ANTHROPIC_API_KEY
base_url: ${CUSTOM_BASE_URL:-https://api.anthropic.com}
```

If a variable is not set in the environment, `ccswap` will fail with a clear error rather than write an empty token to your Claude Code settings.

### What ccswap Modifies

`ccswap` only touches these 6 keys inside `settings.json`'s `env` block. Everything else is preserved byte-for-byte:

- `ANTHROPIC_AUTH_TOKEN`
- `ANTHROPIC_BASE_URL`
- `ANTHROPIC_DEFAULT_SONNET_MODEL`
- `ANTHROPIC_DEFAULT_OPUS_MODEL`
- `ANTHROPIC_DEFAULT_HAIKU_MODEL`
- `API_TIMEOUT_MS`

---

## Safety Guarantees

- **Atomic writes** — `settings.json` is written to a temp file, then renamed. No partial writes.
- **Automatic backups** — Before every write, the current `settings.json` is copied to `settings.json.ccswap.bak`.
- **Validation first** — `settings.json` is parsed and validated before any modification. If it's malformed or just `null`, the tool refuses to overwrite it.
- **Key masking** — API keys are never printed in full. Example: `sk-ant-...y456` (first 7 + last 4 chars).
- **Separate storage** — Keys live in `providers.yaml`, not in Claude Code's shared `settings.json`.

---

## Exit Codes

| Code | Meaning | Example |
|------|---------|---------|
| `0` | Success | — |
| `1` | User error | Provider not found, invalid name, missing config |
| `2` | System error | File permission issues, I/O errors, panic recovery |

All panics are caught at the top level and converted to exit code 2 with an error message.

---

## Development

```bash
# Build
go build -o ccswap .

# Run all tests
go test ./...

# Run integration tests only
go test -run TestE2E ./cmd/...

# Run a specific command's tests
go test -run TestUse_ ./cmd/...
```

Tests are fully isolated using `t.TempDir()` — no external services or shared state required.

---

## Roadmap

- [ ] **Binary releases** — GitHub Actions + goreleaser for automatic cross-platform builds
- [ ] **`ccswap watch`** — Detect Claude Code rate-limit errors and auto-switch providers
- [ ] **`ccswap rotate`** — Cycle through a configured provider rotation
- [ ] **Shell completions** — `bash`, `zsh`, `fish` tab completion for provider names
- [ ] **Keychain integration** — Store API keys in OS keychain instead of plaintext YAML
- [ ] **`ccswap import`** — Detect existing provider config in `settings.json` and save as a named profile

---

## Contributing

Contributions are welcome! Please open an issue or pull request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Ensure tests pass (`go test ./...`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

---

## License

MIT © [Cong Truong](https://github.com/ctru0009)

---

## Acknowledgements

- Built with [spf13/cobra](https://github.com/spf13/cobra) for CLI structure
- Table output via [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
- Colored output via [fatih/color](https://github.com/fatih/color)
