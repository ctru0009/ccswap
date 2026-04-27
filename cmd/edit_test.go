package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newEditRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(editCmd)
	return root
}

func TestEdit_ProviderNotFound(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "nonexistent"})

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
}

func TestEdit_SuccessWithCatEditor(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	t.Setenv("EDITOR", "cat")

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

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

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	if !strings.Contains(string(data), "zai") {
		t.Errorf("providers.yaml should still contain 'zai', got:\n%s", string(data))
	}
}

func TestEdit_InvalidYAMLAfterEdit(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		return os.WriteFile(tmpPath, []byte(":::invalid yaml:::"), 0644)
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "yaml") {
		t.Errorf("error should mention YAML parsing issue, got: %v", err)
	}
}

func TestEdit_NameChangeRejected(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		renamed := `zai-renamed:
  auth_token: "sk-ant-abc123key456"
  base_url: "https://example.com/v1"
  timeout_ms: 300000
  models:
    sonnet: "kimi-k2.6:cloud"
    opus: "deepseek-v4-flash:cloud"
    haiku: "glm-4.7:cloud"
`
		return os.WriteFile(tmpPath, []byte(renamed), 0644)
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for name change, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

func TestEdit_ValidationFailsAfterEdit(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		invalid := `zai:
  auth_token: ""
  base_url: "https://example.com/v1"
  timeout_ms: 300000
  models:
    sonnet: "kimi-k2.6:cloud"
    opus: "deepseek-v4-flash:cloud"
    haiku: "glm-4.7:cloud"
`
		return os.WriteFile(tmpPath, []byte(invalid), 0644)
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for validation failure, got nil")
	}
	if !strings.Contains(err.Error(), "auth_token") {
		t.Errorf("error should mention 'auth_token', got: %v", err)
	}
}

func TestEdit_SuccessWithModifiedValues(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		modified := `zai:
  auth_token: "sk-ant-abc123key456"
  base_url: "https://new-url.example.com/v2"
  timeout_ms: 60000
  models:
    sonnet: "new-sonnet-model"
    opus: "deepseek-v4-flash:cloud"
    haiku: "glm-4.7:cloud"
`
		return os.WriteFile(tmpPath, []byte(modified), 0644)
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "https://new-url.example.com/v2") {
		t.Errorf("providers.yaml should contain updated base_url, got:\n%s", content)
	}
	if !strings.Contains(content, "new-sonnet-model") {
		t.Errorf("providers.yaml should contain updated sonnet model, got:\n%s", content)
	}
	if !strings.Contains(content, "anthropic") {
		t.Errorf("providers.yaml should still contain anthropic, got:\n%s", content)
	}
}

func TestEdit_TempFileCleanedUpOnSuccess(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	var capturedTmpPath string
	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		capturedTmpPath = tmpPath
		return nil
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTmpPath == "" {
		t.Fatal("expected temp file path to be captured")
	}
	if _, err := os.Stat(capturedTmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file should be cleaned up, but exists at: %s", capturedTmpPath)
	}
}

func TestEdit_TempFileCleanedUpOnError(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	var capturedTmpPath string
	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	editorLaunchFunc = func(editor string, tmpPath string) error {
		capturedTmpPath = tmpPath
		return os.WriteFile(tmpPath, []byte(":::invalid:::"), 0644)
	}
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	_ = root.Execute()

	if capturedTmpPath == "" {
		t.Fatal("expected temp file path to be captured")
	}
	if _, err := os.Stat(capturedTmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file should be cleaned up even on error, but exists at: %s", capturedTmpPath)
	}
}

func TestEdit_ProvidersFileMissing(t *testing.T) {
	configDir := t.TempDir()

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	origProvidersPath := providersPathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing providers.yaml, got nil")
	}
}

func TestEdit_EditorFallbackToVi(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	t.Setenv("EDITOR", "")

	origProvidersPath := providersPathFunc
	origEditorFunc := editorLaunchFunc
	origLookupPath := editorLookupFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }

	var capturedEditor string
	editorLaunchFunc = func(editor string, tmpPath string) error {
		capturedEditor = editor
		return nil
	}
	editorLookupFunc = func(name string) (string, error) {
		return "", os.ErrNotExist
	}

	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		editorLaunchFunc = origEditorFunc
		editorLookupFunc = origLookupPath
	})

	root := newEditRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "edit", "zai"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedEditor, "vi") {
		t.Errorf("expected editor to fall back to vi, got: %q", capturedEditor)
	}
}

func TestEdit_NoArgs(t *testing.T) {
	root := newEditRootCmd()
	root.SetArgs([]string{"edit"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no provider arg given, got nil")
	}
}