# ccswap — Claude Code Provider Switcher

`ccswap` is a command-line tool that allows you to easily switch between different model providers for [Claude Code](https://claude.ai/code). It manages a named list of provider profiles and swaps between them by updating your `~/.claude/settings.json` file safely and atomically.

## Features

- **Fast Switching**: Swap Claude Code providers in a single command.
- **Provider Management**: Add, list, edit, and remove provider profiles.
- **Secure**: API keys are stored in a separate configuration file and masked in command output.
- **Environment Variable Interpolation**: Supports `$ENV_VAR` in configuration for improved security and flexibility.
- **Non-Destructive**: Only updates the `env` block in `settings.json`, preserving all other settings (permissions, MCP servers, etc.).
- **Cross-Platform**: Works on Linux, macOS, and Windows.

## Installation

### From Source

Ensure you have [Go](https://go.dev/doc/install) installed (version 1.25+ recommended).

```bash
# Clone the repository
git clone https://github.com/ctru0009/ccswap
cd ccswap

# Build the binary
go build -o ccswap .

# (Optional) Move it to your PATH
mv ccswap ~/.local/bin/
```

Or install directly:
```bash
go install github.com/ctru0009/ccswap@latest
```

## Quick Start

1. Initialize the configuration:
   ```bash
   ccswap init
   ```
2. Add a new provider interactively:
   ```bash
   ccswap add
   ```
3. Switch to the provider:
   ```bash
   ccswap use my-provider
   ```

## Usage

### `ccswap init`
Initializes the `providers.yaml` configuration file. If `ANTHROPIC_API_KEY` is set in your environment, it will automatically create an `anthropic` profile using that key and the default Anthropic API settings.

### `ccswap add`
Interactively adds a new provider. You will be prompted for:
- Provider name (lowercase alphanumeric + hyphens)
- Auth token (ANTHROPIC_AUTH_TOKEN)
- Base URL (ANTHROPIC_BASE_URL)
- Model names for Sonnet, Opus, and Haiku
- API timeout (ms)

### `ccswap use <provider>`
Switches the active Claude Code provider.

Example output:
```
✓ Switched to zai
  Base URL:    https://api.z.ai/api/anthropic
  Sonnet:      glm-5-turbo
  Opus:        glm-5.1
  Haiku:       glm-4.5-air
  Auth:        zai-123...4567

Open a new Claude Code session for changes to take effect.
```

### `ccswap status`
Shows the currently active provider and its configuration by inspecting `~/.claude/settings.json`.

### `ccswap list`
Lists all configured providers in a table. The active provider is marked with an asterisk (`*`).

### `ccswap edit <provider>`
Opens the specified provider profile in your default editor (`$EDITOR`). If no editor is set, it falls back to `nano`, then `vi`. The configuration is validated before being saved.

### `ccswap remove <provider>`
Removes a provider profile. If the provider is currently active, it will ask for confirmation before removal.

## Configuration

### `providers.yaml` Schema

The configuration file is located at:
- Linux: `~/.config/ccswap/providers.yaml`
- macOS: `~/Library/Application Support/ccswap/providers.yaml`
- Windows: `%APPDATA%\ccswap\providers.yaml`

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
`ccswap` supports environment variable interpolation using `$VAR` or `${VAR}` syntax in `auth_token`, `base_url`, and model names. This is the recommended way to handle API keys to avoid storing secrets in plaintext.

If a variable is used but not set in the environment, `ccswap` will warn you and the operation will fail to prevent writing empty tokens to your Claude Code settings.

## Security

- **Masked Output**: API keys are masked in the terminal output (e.g., `sk-ant...1234`), ensuring they aren't accidentally exposed in logs or screen shares.
- **Separate Storage**: Your API keys are stored in `providers.yaml`, keeping them out of Claude Code's shared `settings.json`.
- **Atomic Writes**: `ccswap` writes to a temporary file and then renames it to ensure `settings.json` is never left in a corrupt state.
- **Automatic Backup**: Before any write, a backup of the current `settings.json` is created at `~/.claude/settings.json.ccswap.bak`.

## Cross-Platform

`ccswap` is designed to be cross-platform:
- On Linux and macOS, it follows standard home directory and config directory conventions.
- On Windows, it uses `%USERPROFILE%\.claude` for Claude Code settings and `%APPDATA%\ccswap` for its own configuration.

## Exit Codes

- `0`: Success
- `1`: User error (e.g., provider not found, invalid name, empty environment variable)
- `2`: System error (e.g., file permission issues, I/O errors)
