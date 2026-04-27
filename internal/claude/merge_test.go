package claude

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ctru0009/ccswap/internal/config"
)

func TestMerge_SetsTargetKeys(t *testing.T) {
	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"old"}`),
	}
	provider := config.Provider{
		AuthToken: "sk-test-123",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 60000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-test-123" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "sk-test-123")
	}
	if env["ANTHROPIC_BASE_URL"] != "https://api.example.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q; want %q", env["ANTHROPIC_BASE_URL"], "https://api.example.com")
	}
	if env["ANTHROPIC_DEFAULT_SONNET_MODEL"] != "claude-sonnet-4" {
		t.Errorf("ANTHROPIC_DEFAULT_SONNET_MODEL = %q; want %q", env["ANTHROPIC_DEFAULT_SONNET_MODEL"], "claude-sonnet-4")
	}
	if env["ANTHROPIC_DEFAULT_OPUS_MODEL"] != "claude-opus-4" {
		t.Errorf("ANTHROPIC_DEFAULT_OPUS_MODEL = %q; want %q", env["ANTHROPIC_DEFAULT_OPUS_MODEL"], "claude-opus-4")
	}
	if env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] != "claude-haiku-4" {
		t.Errorf("ANTHROPIC_DEFAULT_HAIKU_MODEL = %q; want %q", env["ANTHROPIC_DEFAULT_HAIKU_MODEL"], "claude-haiku-4")
	}
	if env["API_TIMEOUT_MS"] != "60000" {
		t.Errorf("API_TIMEOUT_MS = %q; want %q", env["API_TIMEOUT_MS"], "60000")
	}
}

func TestMerge_PreservesNonTargetEnvKeys(t *testing.T) {
	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_API_KEY":"keep-this","ENABLE_LSP_TOOL":"true"}`),
	}
	provider := config.Provider{
		AuthToken: "sk-new",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["ANTHROPIC_API_KEY"] != "keep-this" {
		t.Errorf("ANTHROPIC_API_KEY = %q; want %q", env["ANTHROPIC_API_KEY"], "keep-this")
	}
	if env["ENABLE_LSP_TOOL"] != "true" {
		t.Errorf("ENABLE_LSP_TOOL = %q; want %q", env["ENABLE_LSP_TOOL"], "true")
	}
}

func TestMerge_PreservesTopLevelKeys(t *testing.T) {
	settings := map[string]json.RawMessage{
		"env":         json.RawMessage(`{}`),
		"permissions": json.RawMessage(`{"allow":["Bash"]}`),
		"hooks":       json.RawMessage(`{"pre_tool_use":"script.sh"}`),
	}
	provider := config.Provider{
		AuthToken: "sk-test",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	if string(result["permissions"]) != `{"allow":["Bash"]}` {
		t.Errorf("permissions = %q; want %q", string(result["permissions"]), `{"allow":["Bash"]}`)
	}
	if string(result["hooks"]) != `{"pre_tool_use":"script.sh"}` {
		t.Errorf("hooks = %q; want %q", string(result["hooks"]), `{"pre_tool_use":"script.sh"}`)
	}
}

func TestMerge_CreatesEnvBlockIfMissing(t *testing.T) {
	settings := map[string]json.RawMessage{
		"permissions": json.RawMessage(`{"allow":["Bash"]}`),
	}
	provider := config.Provider{
		AuthToken: "sk-test",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	envRaw, ok := result["env"]
	if !ok {
		t.Fatal("expected 'env' key in result")
	}

	var env map[string]string
	if err := json.Unmarshal(envRaw, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-test" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "sk-test")
	}
}

func TestMerge_OverwritesExistingTargetKeys(t *testing.T) {
	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_BASE_URL":"https://old-url.com","ANTHROPIC_AUTH_TOKEN":"old-token"}`),
	}
	provider := config.Provider{
		AuthToken: "sk-new",
		BaseURL:   "https://new-url.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["ANTHROPIC_BASE_URL"] != "https://new-url.com" {
		t.Errorf("ANTHROPIC_BASE_URL = %q; want %q", env["ANTHROPIC_BASE_URL"], "https://new-url.com")
	}
}

func TestMerge_StringConversionOfTimeoutMs(t *testing.T) {
	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{}`),
	}
	provider := config.Provider{
		AuthToken: "sk-test",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 3000000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["API_TIMEOUT_MS"] != "3000000" {
		t.Errorf("API_TIMEOUT_MS = %q; want %q", env["API_TIMEOUT_MS"], "3000000")
	}
}

func TestMerge_EnvVarExpansion(t *testing.T) {
	os.Setenv("ZAI_API_KEY", "expanded-key-value")
	defer os.Unsetenv("ZAI_API_KEY")

	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{}`),
	}
	provider := config.Provider{
		AuthToken: "$ZAI_API_KEY",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	result, err := MergeEnv(settings, provider)
	if err != nil {
		t.Fatalf("MergeEnv() error: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	if env["ANTHROPIC_AUTH_TOKEN"] != "expanded-key-value" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "expanded-key-value")
	}
}

func TestMerge_EmptyAuthTokenAfterExpansion_Error(t *testing.T) {
	os.Unsetenv("MISSING_VAR")

	settings := map[string]json.RawMessage{
		"env": json.RawMessage(`{}`),
	}
	provider := config.Provider{
		AuthToken: "$MISSING_VAR",
		BaseURL:   "https://api.example.com",
		TimeoutMs: 30000,
		Models: config.Models{
			Sonnet: "claude-sonnet-4",
			Opus:   "claude-opus-4",
			Haiku:  "claude-haiku-4",
		},
	}

	_, err := MergeEnv(settings, provider)
	if err == nil {
		t.Fatal("expected error when auth_token is empty after expansion")
	}
}