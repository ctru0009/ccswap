package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newStatusRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(statusCmd)
	return root
}

func TestStatus_ActiveProviderMatched(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":              "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN":            "sk-ant-abc123key456",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":  "kimi-k2.6:cloud",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":    "deepseek-v4-flash:cloud",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "glm-4.7:cloud",
		},
	})

	stateYAML := "active_provider: zai\nlast_switched: 2026-04-26T21:06:57+10:00\n"
	if err := os.WriteFile(filepath.Join(configDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain provider name highlighted
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider name 'zai', got: %q", output)
	}
	if !strings.Contains(output, "Active provider:") {
		t.Errorf("output should contain 'Active provider:', got: %q", output)
	}
	if !strings.Contains(output, "https://example.com/v1") {
		t.Errorf("output should contain base URL, got: %q", output)
	}
	if !strings.Contains(output, "kimi-k2.6:cloud") {
		t.Errorf("output should contain sonnet model, got: %q", output)
	}
	if !strings.Contains(output, "deepseek-v4-flash:cloud") {
		t.Errorf("output should contain opus model, got: %q", output)
	}
	if !strings.Contains(output, "glm-4.7:cloud") {
		t.Errorf("output should contain haiku model, got: %q", output)
	}
	if !strings.Contains(output, "sk-ant-...y456") {
		t.Errorf("auth_token should be masked, got: %q", output)
	}
	if strings.Contains(output, "sk-ant-abc123key456") {
		t.Errorf("full auth_token should NOT appear in output, got: %q", output)
	}
	if !strings.Contains(output, "2026-04-26T21:06:57+10:00") {
		t.Errorf("output should contain last_switched timestamp, got: %q", output)
	}
	if !strings.Contains(output, "Last switched:") {
		t.Errorf("output should contain 'Last switched:', got: %q", output)
	}
}

func TestStatus_UnmatchedBaseURL(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":   "https://custom.example.com/api",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-custom-token",
		},
	})

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "unknown") {
		t.Errorf("output should contain 'unknown', got: %q", output)
	}
	if !strings.Contains(output, "https://custom.example.com/api") {
		t.Errorf("output should contain the base URL, got: %q", output)
	}
}

func TestStatus_NoSettingsJSON(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	// No settings.json written — it doesn't exist

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { statusSettingsPath = origSettingsPath })

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No provider configured") {
		t.Errorf("output should contain 'No provider configured', got: %q", output)
	}
	if !strings.Contains(output, "ccswap init") {
		t.Errorf("output should suggest 'ccswap init', got: %q", output)
	}
}

func TestStatus_NoBaseURL(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-something",
		},
	})

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { statusSettingsPath = origSettingsPath })

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No provider configured") {
		t.Errorf("output should contain 'No provider configured', got: %q", output)
	}
}

func TestStatus_NoProvidersFile(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	// No providers.yaml — just settings.json with a base URL
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":   "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-abc123key456",
		},
	})

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "unknown") {
		t.Errorf("output should contain 'unknown' when providers.yaml missing, got: %q", output)
	}
	if !strings.Contains(output, "https://example.com/v1") {
		t.Errorf("output should contain the base URL, got: %q", output)
	}
}

func TestStatus_NoLastSwitched(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":   "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-abc123key456",
		},
	})
	// No state.yaml — no last_switched

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should still show provider info
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider name, got: %q", output)
	}
	// Should NOT contain "Last switched:"
	if strings.Contains(output, "Last switched:") {
		t.Errorf("output should NOT contain 'Last switched:' when state is missing, got: %q", output)
	}
}

func TestStatus_ExitCodeZero(t *testing.T) {
	// Status should always exit with code 0 (informational only)
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	// No settings.json — still exit 0

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { statusSettingsPath = origSettingsPath })

	err := root.Execute()
	if err != nil {
		t.Fatalf("status should not return error (exit code 0), got: %v", err)
	}
}

func TestStatus_MalformedStateYAML(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":   "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-abc123key456",
		},
	})

	// Write malformed state.yaml (unclosed bracket is invalid YAML)
	if err := os.WriteFile(filepath.Join(configDir, "state.yaml"), []byte("active_provider: [\n"), 0644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for malformed state.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "state") {
		t.Errorf("error should mention 'state', got: %v", err)
	}
}

func TestStatus_EmptyLastSwitched(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":   "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-abc123key456",
		},
	})

	// Write state with zero time
	stateYAML := "active_provider: zai\nlast_switched: 0001-01-01T00:00:00Z\n"
	if err := os.WriteFile(filepath.Join(configDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	origProvidersPath := statusProvidersPath
	origStatePath := statusStatePath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		statusSettingsPath = origSettingsPath
		statusProvidersPath = origProvidersPath
		statusStatePath = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Last switched:") {
		t.Errorf("output should NOT contain 'Last switched:' when it's zero, got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("output should still show provider info, got: %q", output)
	}
}

func TestStatus_CorruptSettingsJSON(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	// Write malformed JSON to settings.json
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	root := newStatusRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "status"})

	origSettingsPath := statusSettingsPath
	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { statusSettingsPath = origSettingsPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for corrupt settings.json, got nil")
	}
	if !strings.Contains(err.Error(), "settings") {
		t.Errorf("error should mention 'settings', got: %v", err)
	}
}
