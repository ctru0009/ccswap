package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newImportRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(importCmd)
	return root
}

func TestImport_Success(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":            "sk-ant-abc123key456",
			"ANTHROPIC_BASE_URL":              "https://api.anthropic.com",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":  "claude-sonnet-4-20250514",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":    "claude-opus-4-20250514",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "claude-haiku-3-5-20250101",
			"API_TIMEOUT_MS":                  "600000",
		},
	})

	input := "anthropic\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Detected provider configuration") {
		t.Errorf("output should show detected config, got: %q", output)
	}
	if !strings.Contains(output, "sk-ant-...y456") {
		t.Errorf("output should show masked auth token, got: %q", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "imported") {
		t.Errorf("output should contain 'imported', got: %q", output)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "anthropic") {
		t.Errorf("providers.yaml should contain 'anthropic', got: %s", content)
	}
	if !strings.Contains(content, "https://api.anthropic.com") {
		t.Errorf("providers.yaml should contain base_url, got: %s", content)
	}
	if !strings.Contains(content, "sk-ant-abc123key456") {
		t.Errorf("providers.yaml should contain auth_token, got: %s", content)
	}
	if !strings.Contains(content, "claude-sonnet-4-20250514") {
		t.Errorf("providers.yaml should contain sonnet model, got: %s", content)
	}
	if !strings.Contains(content, "600000") {
		t.Errorf("providers.yaml should contain timeout_ms, got: %s", content)
	}
}

func TestImport_PartialConfig(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	// Only base_url and auth_token set, no models
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "my-token",
			"ANTHROPIC_BASE_URL":  "https://custom.example.com/v1",
		},
	})

	input := "myprovider\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "myprovider") {
		t.Errorf("providers.yaml should contain 'myprovider', got: %s", content)
	}
	if !strings.Contains(content, "https://custom.example.com/v1") {
		t.Errorf("providers.yaml should contain base_url, got: %s", content)
	}
}

func TestImport_DuplicateName(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	writeTestProviders(t, configDir, validProvidersYAML)

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":  "https://new-provider.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "new-token",
		},
	})

	// Try to import with name "zai" which already exists
	input := "zai\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for duplicate provider name, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestImport_MatchesExistingProvider(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	writeTestProviders(t, configDir, validProvidersYAML)

	// Settings match the "zai" provider's config
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL": "https://example.com/v1",
		},
	})

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "matches provider") {
		t.Errorf("output should mention matching provider, got: %q", output)
	}
	if !strings.Contains(output, "ccswap use") {
		t.Errorf("output should suggest 'ccswap use', got: %q", output)
	}
}

func TestImport_NoBaseURL(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "some-token",
		},
	})

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No provider configuration found") {
		t.Errorf("output should mention no provider found, got: %q", output)
	}
	if !strings.Contains(output, "ccswap add") {
		t.Errorf("output should suggest 'ccswap add', got: %q", output)
	}
}

func TestImport_NoSettingsJSON(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	// No settings.json exists

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No provider configuration found") {
		t.Errorf("output should say no provider found when settings.json missing, got: %q", output)
	}
}

func TestImport_InvalidName(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":  "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "some-token",
		},
	})

	input := "INVALID-NAME\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid provider name, got nil")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("error should mention validation pattern, got: %v", err)
	}
}

func TestImport_MalformedSettings(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{invalid}"), 0644); err != nil {
		t.Fatalf("failed to write malformed settings: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for malformed settings.json, got nil")
	}
	if !strings.Contains(err.Error(), "settings") {
		t.Errorf("error should mention settings, got: %v", err)
	}
}

func TestImport_PreservesExistingProviders(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	writeTestProviders(t, configDir, validProvidersYAML)

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":              "https://new-provider.example.com/v1",
			"ANTHROPIC_AUTH_TOKEN":             "new-token-12345",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":  "new-sonnet",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":    "new-opus",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "new-haiku",
		},
	})

	input := "newprov\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	// Existing providers should still be present
	if !strings.Contains(content, "zai") {
		t.Errorf("providers.yaml should still contain 'zai', got: %s", content)
	}
	if !strings.Contains(content, "anthropic") {
		t.Errorf("providers.yaml should still contain 'anthropic', got: %s", content)
	}
	// New provider should also be present
	if !strings.Contains(content, "newprov") {
		t.Errorf("providers.yaml should contain 'newprov', got: %s", content)
	}
}

func TestImport_NoProvidersFile(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	// No providers.yaml — should be auto-created

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":  "https://api.anthropic.com",
			"ANTHROPIC_AUTH_TOKEN": "sk-ant-test",
		},
	})

	input := "imported\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	if !strings.Contains(string(data), "imported") {
		t.Errorf("providers.yaml should contain 'imported', got: %s", string(data))
	}
}

func TestImport_WarnsEmptyAuthToken(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	// base_url but no auth_token
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL": "https://example.com/v1",
		},
	})

	input := "notoken\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "no auth token detected") {
		t.Errorf("output should warn about missing auth token, got: %q", output)
	}
}

func TestImport_WarnsMissingModels(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	// base_url and auth_token but no models
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL":  "https://example.com/v1",
			"ANTHROPIC_AUTH_TOKEN": "some-token",
		},
	})

	input := "nomodels\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newImportRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	root.SetIn(r)

	origSettingsPath := importSettingsPath
	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	t.Cleanup(func() { importSettingsPath = origSettingsPath })

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "models are missing") {
		t.Errorf("output should warn about missing models, got: %q", output)
	}
}