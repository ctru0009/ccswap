# Product Requirements Document
## ccswap — Claude Code Provider Switcher CLI

**Version:** 1.0.0  
**Author:** Cong Truong  
**Date:** April 27, 2026  
**Status:** Draft

---

## 1. Overview

### 1.1 Problem Statement

Claude Code enforces a 5-hour usage window on Claude Pro/Max subscriptions. When this limit is hit, developers who subscribe to alternative model providers (Z.ai GLM Coding Plan, Ollama Cloud, OpenRouter) currently have no fast way to switch Claude Code to use a different backend. The manual process requires editing `~/.claude/settings.json` directly, knowing the exact env var names, and restarting the session — a slow, error-prone process that interrupts flow.

### 1.2 Solution

`ccswap` is a Go CLI that manages a named list of Claude Code provider profiles and swaps between them with a single command. It reads and writes the `env` block in `~/.claude/settings.json`, leaving all other Claude Code configuration untouched.

### 1.3 Goals

- Swap Claude Code provider in under 2 seconds
- Zero disruption to existing Claude Code settings (permissions, MCP servers, etc.)
- Store provider configs securely with API keys outside of Claude Code's config
- Work seamlessly on Linux (NUC/SSH), macOS, and Windows

### 1.4 Non-Goals

- Not a proxy or gateway — does not intercept API traffic
- Not a usage monitor — does not track when the 5-hour limit is hit
- Not a model benchmarker — does not evaluate provider quality
- Does not manage Claude Code installation or updates

---

## 2. Target Users

**Primary:** Cong — junior software engineer, active job seeker, running Claude Code on a Ubuntu NUC homelab via SSH + tmux, subscribed to Claude Pro + Z.ai Coding Plan + Ollama Cloud.

**Secondary:** Any Claude Code user with multiple provider subscriptions who hits usage limits regularly.

---

## 3. User Stories

| ID | As a... | I want to... | So that... |
|----|---------|--------------|------------|
| US-01 | CLI user | run `ccswap use zai` | Claude Code immediately routes to Z.ai GLM without manual config editing |
| US-02 | CLI user | run `ccswap status` | I can see which provider is currently active at a glance |
| US-03 | CLI user | run `ccswap list` | I can see all my configured providers and their model mappings |
| US-04 | CLI user | run `ccswap add` | I can add a new provider interactively without editing YAML manually |
| US-05 | CLI user | run `ccswap remove <name>` | I can clean up providers I no longer use |
| US-06 | CLI user | run `ccswap edit <name>` | I can update an existing provider's keys or model mappings |
| US-07 | CLI user | have my API keys stored separately from Claude Code's config | My keys aren't accidentally committed if someone shares their settings.json |
| US-08 | CLI user | have the swap be non-destructive | My Claude Code permissions, MCP servers, and other settings are never touched |
| US-09 | CLI user | run `ccswap use anthropic` after my limit resets | I can seamlessly return to native Claude without any manual steps |
| US-10 | CLI user | see a confirmation after swapping | I know the swap succeeded before opening a new Claude Code session |
| US-11 | CLI user | run `ccswap import` | I can save my existing Claude Code env config as a named provider without manual entry |

---

## 4. Functional Requirements

### 4.1 Commands

#### `ccswap use <provider>`

Switches the active Claude Code provider.

**Behaviour:**
1. Reads `~/.config/ccswap/providers.yaml`
2. Looks up the named provider
3. Reads `~/.claude/settings.json` (creates it if missing)
4. Deep-merges the provider's env vars into the `env` block
5. Writes `settings.json` back atomically (write to temp file, then rename)
6. Prints confirmation with the provider name and active models

**Output example:**
```
✓ Switched to zai
  Base URL  : https://api.z.ai/api/anthropic
  Sonnet    : glm-5-turbo
  Opus      : glm-5.1
  Haiku     : glm-4.5-air

Open a new Claude Code session for changes to take effect.
```

**Error cases:**
- Provider not found → list available providers, exit 1
- `settings.json` is malformed JSON → print error with file path, do not overwrite, exit 1
- `providers.yaml` missing → prompt user to run `ccswap init`, exit 1

---

#### `ccswap status`

Shows the currently active provider by reading `~/.claude/settings.json` and matching the `ANTHROPIC_BASE_URL` against known providers.

**Output example:**
```
Active provider : zai
Base URL        : https://api.z.ai/api/anthropic
Sonnet          : glm-5-turbo
Opus            : glm-5.1
Haiku           : glm-4.5-air

Last switched   : 2026-04-27 14:32:11 AEST
```

If the current `ANTHROPIC_BASE_URL` doesn't match any configured provider:
```
Active provider : unknown (manually configured or not set)
Base URL        : https://custom.endpoint.example.com
```

---

#### `ccswap list`

Lists all configured providers.

**Output example:**
```
PROVIDER        BASE URL                              SONNET              OPUS                  HAIKU
anthropic  *    https://api.anthropic.com             claude-sonnet-4-6   claude-opus-4-6       claude-haiku-4-5
zai             https://api.z.ai/api/anthropic        glm-5-turbo         glm-5.1               glm-4.5-air
ollama-cloud    https://ollama.com/v1                 kimi-k2.6:cloud     deepseek-v4-flash:cl  glm-4.7:cloud
openrouter      https://openrouter.ai/api/v1          anthropic/claude-…  deepseek/deepseek-v4  google/gemini-fla

* = currently active
```

---

#### `ccswap add`

Interactive prompt to add a new provider.

**Prompt flow:**
```
Provider name: ollama-cloud
Auth token (ANTHROPIC_AUTH_TOKEN): ****
Base URL (ANTHROPIC_BASE_URL): https://ollama.com/v1
Sonnet model: kimi-k2.6:cloud
Opus model: deepseek-v4-flash:cloud
Haiku model: glm-4.7:cloud
API timeout ms [3000000]: 

✓ Provider 'ollama-cloud' added.
Run: ccswap use ollama-cloud
```

---

#### `ccswap remove <provider>`

Removes a provider from `providers.yaml`. Prompts for confirmation if the provider is currently active.

**Output example:**
```
Remove provider 'zai'? [y/N]: y
✓ Removed 'zai'.
```

If active:
```
'zai' is currently active. Remove anyway? [y/N]:
```

---

#### `ccswap edit <provider>`

Opens the provider entry in `$EDITOR` (falls back to `nano`, then `vi`) with the current values pre-filled as YAML. On save, validates and writes back.

---

#### `ccswap init`

First-run setup. Creates `~/.config/ccswap/providers.yaml` with a commented template including the Anthropic default profile pre-filled if `ANTHROPIC_API_KEY` is set in the environment.

```
✓ Created ~/.config/ccswap/providers.yaml
  Edit this file to add your providers, then run:
  ccswap list
  ccswap use <provider>
```

---

#### `ccswap import`

Detects existing provider configuration in `~/.claude/settings.json` and saves it as a named provider profile in `providers.yaml`.

**Behaviour:**
1. Reads `~/.claude/settings.json` and extracts the 6 target env keys
2. If `ANTHROPIC_BASE_URL` is empty, prints a message and exits (no provider detected)
3. Displays the detected configuration (base URL, models, masked auth token)
4. If the detected `ANTHROPIC_BASE_URL` matches an existing provider's expanded base URL, suggests running `ccswap use <provider>` instead
5. Prompts for a provider name interactively
6. Validates the name and checks for duplicates
7. Saves the provider to `providers.yaml`

**Output example (provider detected):**
```
→ Detected provider configuration:
  Base URL:  https://api.anthropic.com
  Sonnet:    claude-sonnet-4-20250514
  Opus:      claude-opus-4-20250514
  Haiku:     claude-haiku-3-5-20250101
  Auth:      sk-ant-...y456
  Timeout:   600000 ms

Provider name: anthropic
✓ Provider anthropic imported. Run: ccswap use anthropic
```

**Output example (matches existing provider):**
```
This configuration matches provider anthropic.
Run ccswap use anthropic to switch to it.
```

**Output example (no provider configured):**
```
No provider configuration found in settings.json.
Run `ccswap add` to add a provider manually.
```

---

### 4.2 Settings.json Merge Behaviour

`ccswap` only touches the following keys inside the `env` block:

```
ANTHROPIC_AUTH_TOKEN
ANTHROPIC_BASE_URL
ANTHROPIC_DEFAULT_SONNET_MODEL
ANTHROPIC_DEFAULT_OPUS_MODEL
ANTHROPIC_DEFAULT_HAIKU_MODEL
API_TIMEOUT_MS
```

All other keys at any level of `settings.json` are preserved exactly as-is. The merge is a targeted key-by-key update, not a full file replacement.

**Atomic write pattern:**
1. Write new JSON to `~/.claude/settings.json.ccswap.tmp`
2. `os.Rename()` to `~/.claude/settings.json`

This prevents partial writes from corrupting the config.

### 4.3 State Tracking

`ccswap` writes a small state file at `~/.config/ccswap/state.yaml` to track:

```yaml
active_provider: zai
last_switched: "2026-04-27T14:32:11+10:00"
```

This is used by `ccswap status` to show last-switched time without re-parsing `settings.json`.

---

## 5. Non-Functional Requirements

### 5.1 Performance
- `ccswap use` must complete in under 500ms on any reasonable machine
- No network calls during any operation — entirely local file I/O

### 5.2 Safety
- API keys never printed to stdout in plaintext — masked as `****` in all output
- `ccswap list` shows masked keys: `sk-ant-...abc1` (first 7 chars + last 4)
- Never overwrites `settings.json` if it cannot be parsed as valid JSON first
- Automatic backup: before any write, copy current `settings.json` to `~/.claude/settings.json.ccswap.bak`

### 5.3 Portability
- Supports Linux, macOS, Windows
- Config path resolves via `os.UserHomeDir()` and `os.UserConfigDir()` — no hardcoded paths
- On Windows, `~/.claude/` resolves to `%USERPROFILE%\.claude\`

### 5.4 Error Handling
- All errors print to stderr with a clear message and actionable suggestion
- Exit codes: 0 = success, 1 = user error (bad provider name, missing config), 2 = system error (file permission, disk full)
- Never panic in production — all panics caught and converted to exit 2

---

## 6. Configuration Schema

### `~/.config/ccswap/providers.yaml`

```yaml
providers:
  anthropic:
    auth_token: "sk-ant-..."           # maps to ANTHROPIC_AUTH_TOKEN
    base_url: "https://api.anthropic.com"  # maps to ANTHROPIC_BASE_URL
    timeout_ms: 600000                 # maps to API_TIMEOUT_MS
    models:
      sonnet: "claude-sonnet-4-6"     # ANTHROPIC_DEFAULT_SONNET_MODEL
      opus: "claude-opus-4-6"         # ANTHROPIC_DEFAULT_OPUS_MODEL
      haiku: "claude-haiku-4-5"       # ANTHROPIC_DEFAULT_HAIKU_MODEL

  zai:
    auth_token: "zai-..."
    base_url: "https://api.z.ai/api/anthropic"
    timeout_ms: 3000000
    models:
      sonnet: "glm-5-turbo"
      opus: "glm-5.1"
      haiku: "glm-4.5-air"

  ollama-cloud:
    auth_token: "ollama-..."
    base_url: "https://ollama.com/v1"
    timeout_ms: 3000000
    models:
      sonnet: "kimi-k2.6:cloud"
      opus: "deepseek-v4-flash:cloud"
      haiku: "glm-4.7:cloud"

  openrouter:
    auth_token: "sk-or-..."
    base_url: "https://openrouter.ai/api/v1"
    timeout_ms: 3000000
    models:
      sonnet: "anthropic/claude-sonnet-4-6"
      opus: "deepseek/deepseek-v4-flash"
      haiku: "google/gemini-flash-2.5"
```

**Validation rules:**
- `auth_token` — required, non-empty string
- `base_url` — required, must parse as valid URL with http/https scheme
- `models.sonnet`, `models.opus`, `models.haiku` — all required, non-empty strings
- `timeout_ms` — optional, integer, defaults to 600000 if omitted
- Provider name — lowercase alphanumeric + hyphens only, max 32 chars

---

## 7. Technical Design

### 7.1 Tech Stack

| Concern | Choice | Reason |
|---------|--------|--------|
| Language | Go 1.22+ | Fast binary, great stdlib for file I/O, fits CV goals |
| CLI framework | `github.com/spf13/cobra` | Industry standard, familiar to hiring teams |
| YAML parsing | `gopkg.in/yaml.v3` | Canonical Go YAML library |
| JSON handling | `encoding/json` (stdlib) | No external dep needed |
| Coloured output | `github.com/fatih/color` | Lightweight, cross-platform |
| Table output | `github.com/olekukonko/tablewriter` | Clean list output |

### 7.2 Project Structure

```
ccswap/
├── main.go                  # entry point, cobra root command
├── cmd/
│   ├── root.go              # root command + global flags
│   ├── use.go               # ccswap use <provider>
│   ├── status.go            # ccswap status
│   ├── list.go              # ccswap list
│   ├── add.go               # ccswap add (interactive)
│   ├── remove.go            # ccswap remove <provider>
│   ├── edit.go              # ccswap edit <provider>
│   ├── init.go              # ccswap init
│   └── import.go            # ccswap import (detects env config from settings.json)
├── internal/
│   ├── config/
│   │   ├── providers.go     # load/save providers.yaml
│   │   ├── state.go         # load/save state.yaml
│   │   └── paths.go         # resolve config/claude paths cross-platform
│   └── claude/
│       ├── settings.go      # read/write ~/.claude/settings.json
│       └── merge.go         # targeted env key merge logic
├── go.mod
├── go.sum
└── README.md
```

### 7.3 Key Data Types

```go
// internal/config/providers.go
type Provider struct {
    AuthToken string `yaml:"auth_token"`
    BaseURL   string `yaml:"base_url"`
    TimeoutMs int    `yaml:"timeout_ms"`
    Models    Models `yaml:"models"`
}

type Models struct {
    Sonnet string `yaml:"sonnet"`
    Opus   string `yaml:"opus"`
    Haiku  string `yaml:"haiku"`
}

type ProvidersConfig struct {
    Providers map[string]Provider `yaml:"providers"`
}

// internal/claude/settings.go — settings.json is parsed as a map, not a struct.
// This ensures all unknown top-level keys are preserved byte-for-byte.
// The "env" key is unmarshaled into map[string]string when needed.
// We never define a typed ClaudeSettings struct because json:"-" on Extra
// would silently discard keys we don't know about.
settings := make(map[string]json.RawMessage)
```

### 7.4 Merge Algorithm

```
1. Read settings.json → unmarshal into map[string]json.RawMessage
2. Extract "env" key → unmarshal into map[string]string
3. Set/overwrite only the 6 target keys
4. Re-marshal env map → set back into outer map
5. Marshal outer map → write atomically
```

This approach preserves all unknown keys at all levels without needing to know the full settings.json schema.

---

## 8. Installation

### From source (primary during development)
```bash
git clone https://github.com/ctru0009/ccswap
cd ccswap
go build -o ccswap .
mv ccswap ~/.local/bin/
```

### Future: binary releases via GitHub Actions
- Build matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Goreleaser for release automation
- Install script: `curl -sSL https://raw.githubusercontent.com/ctru0009/ccswap/main/install.sh | sh`

---

## 9. Milestones

| Milestone | Scope | Target |
|-----------|-------|--------|
| M1 — Core | `use`, `status`, `list`, `init` commands + merge logic | Week 1 |
| M2 — Management | `add`, `remove`, `edit`, `import` commands + validation | Week 2 |
| M3 — Polish | Coloured output, table formatting, backup logic, cross-platform testing | Week 2-3 |
| M4 — Release | README, install script, GitHub Actions release pipeline | Week 3 |

---

## 10. Success Metrics

- `ccswap use <provider>` completes in under 500ms ✓
- Zero fields lost from `settings.json` after a swap ✓
- Works on Ubuntu 24.04 (NUC/SSH), Windows 11 ✓
- All 6 target env vars correctly written and readable by a new Claude Code session ✓
- API keys never appear in stdout during normal operation ✓

---

## 11. Open Questions

| # | Question | Impact |
|---|----------|--------|
| 1 | Should `ccswap` support a `--global` vs `--project` flag to target project-level `.claude/settings.json` instead of user-level? | Low for now, useful later for per-project provider locking |
| 2 | Should there be a `ccswap watch` mode that auto-swaps on detecting a rate limit error in Claude Code logs? | High value but complex — v2 feature |
| 3 | ~~Is it worth supporting env var interpolation in providers.yaml?~~ ✅ DECISION: Yes — implemented in v1. `os.ExpandEnv` is applied to all string fields (`auth_token`, `base_url`, model names). Empty expansion after interpolation returns a clear error and prevents writing an empty token to settings.json. |
| 4 | Should `ccswap` support a default provider that it falls back to when none is specified? | Low — `ccswap status` + tab completion handles this well enough |

---

## 12. Future Scope (v2)

- **`ccswap watch`** — tail Claude Code logs, detect 5-hour limit error, auto-swap to next provider in a configured rotation
- **`ccswap rotate`** — define a provider rotation order, cycle through on demand
- **Keychain integration** — store API keys in OS keychain (libsecret on Linux, Keychain on macOS, Credential Manager on Windows) instead of plaintext YAML
- **Shell completions** — `ccswap completion bash/zsh/fish` for provider name tab completion (v2 only; v1 guardrail disables the default `completion` subcommand to avoid scope creep)
