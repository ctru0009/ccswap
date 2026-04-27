package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newUseRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(useCmd)
	return root
}

func writeTestProviders(t *testing.T, dir string, yamlContent string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "providers.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}
}

func writeTestSettings(t *testing.T, claudeDir string, data map[string]interface{}) {
	t.Helper()
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), out, 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}
}

const validProvidersYAML = `providers:
  zai:
    auth_token: "sk-ant-abc123key456"
    base_url: "https://example.com/v1"
    timeout_ms: 300000
    models:
      sonnet: "kimi-k2.6:cloud"
      opus: "deepseek-v4-flash:cloud"
      haiku: "glm-4.7:cloud"
  anthropic:
    auth_token: "sk-ant-xyz789"
    base_url: "https://api.anthropic.com"
    timeout_ms: 60000
    models:
      sonnet: "claude-sonnet-4-20250514"
      opus: "claude-opus-4-20250514"
      haiku: "claude-haiku-3-5-20250101"
`

func TestUse_Success(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "old-token",
		},
	})

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider name 'zai', got: %q", output)
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
	if !strings.Contains(output, "new Claude Code session") {
		t.Errorf("output should mention new session, got: %q", output)
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}
	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		t.Fatal("settings.json should have env object")
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-ant-abc123key456" {
		t.Errorf("expected auth_token to be set, got: %v", env["ANTHROPIC_AUTH_TOKEN"])
	}
	if env["ANTHROPIC_BASE_URL"] != "https://example.com/v1" {
		t.Errorf("expected base_url to be set, got: %v", env["ANTHROPIC_BASE_URL"])
	}

	if _, err := os.Stat(filepath.Join(claudeDir, "settings.json.ccswap.bak")); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}

	stateData, err := os.ReadFile(filepath.Join(configDir, "state.yaml"))
	if err != nil {
		t.Fatalf("failed to read state.yaml: %v", err)
	}
	if !strings.Contains(string(stateData), "zai") {
		t.Errorf("state.yaml should contain provider name 'zai', got: %s", string(stateData))
	}
}

func TestUse_ProviderNotFound(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "nonexistent"})

	origProvidersPath := providersPathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
	if !strings.Contains(err.Error(), "zai") && !strings.Contains(err.Error(), "anthropic") {
		t.Errorf("error should list available providers, got: %v", err)
	}
}

func TestUse_AlreadyActive(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{},
	})

	stateYAML := "active_provider: zai\nlast_switched: 2025-01-01T00:00:00Z\n"
	if err := os.WriteFile(filepath.Join(configDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already active") {
		t.Errorf("output should mention 'already active', got: %q", output)
	}
}

func TestUse_MalformedSettings(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{invalid json!!!}"), 0644); err != nil {
		t.Fatalf("failed to write malformed settings: %v", err)
	}

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for malformed settings.json, got nil")
	}
	if !strings.Contains(err.Error(), "settings.json") {
		t.Errorf("error should mention settings.json, got: %v", err)
	}
}

func TestUse_ProvidersFileMissing(t *testing.T) {
	configDir := t.TempDir()

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing providers.yaml, got nil")
	}
}

func TestUse_NoArgs(t *testing.T) {
	root := newUseRootCmd()
	root.SetArgs([]string{"use"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no provider arg given, got nil")
	}
}

func TestUse_CreatesClaudeDir(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := filepath.Join(t.TempDir(), "claude-nested")
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Error("claude directory should have been created")
	}
}

func TestUse_PreservesNonTargetEnvKeys(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]interface{}{
			"ANTHROPIC_AUTH_TOKEN": "old-token",
			"SOME_OTHER_KEY":      "preserved-value",
		},
		"permissions": map[string]interface{}{
			"allow": []string{"tool1"},
		},
	})

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "SOME_OTHER_KEY") {
		t.Error("non-target env keys should be preserved")
	}
	if !strings.Contains(content, "preserved-value") {
		t.Error("non-target env values should be preserved")
	}
	if !strings.Contains(content, "permissions") {
		t.Error("non-env top-level keys should be preserved")
	}
}

func TestUse_NullSettingsJSON_DoesNotOverwrite(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	originalContent := "null"
	if err := os.WriteFile(settingsPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write null settings: %v", err)
	}

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for null settings.json, got nil")
	}

	currentContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading current settings: %v", err)
	}
	if string(currentContent) != originalContent {
		t.Errorf("null settings.json should not be overwritten; got %q", string(currentContent))
	}
}

func TestUse_ProvidersFileMissing_ActionableMessage(t *testing.T) {
	configDir := t.TempDir()

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing providers.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "ccswap init") {
		t.Errorf("error should suggest running 'ccswap init', got: %v", err)
	}
}

func TestUse_StaleTempFileOverwritten(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "old-token",
		},
	})

	tempPath := filepath.Join(claudeDir, "settings.json.ccswap.tmp")
	if err := os.WriteFile(tempPath, []byte("stale temp content"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	root := newUseRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "use", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("stale temp file should be cleaned up after write")
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	if strings.Contains(string(data), "stale temp content") {
		t.Error("stale temp content should not appear in settings.json")
	}
}