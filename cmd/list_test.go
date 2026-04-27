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

func newListRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(listCmd)
	return root
}

func TestList_ShowsProviders(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "PROVIDER") {
		t.Errorf("output should contain table header, got: %q", output)
	}
	if !strings.Contains(output, "BASE URL") {
		t.Errorf("output should contain BASE URL header, got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider 'zai', got: %q", output)
	}
	if !strings.Contains(output, "anthropic") {
		t.Errorf("output should contain provider 'anthropic', got: %q", output)
	}
	if !strings.Contains(output, "sk-ant-...y456") {
		t.Errorf("auth token should be masked, got: %q", output)
	}
	if !strings.Contains(output, "sk-ant-...z789") {
		t.Errorf("auth token should be masked, got: %q", output)
	}
	if strings.Contains(output, "sk-ant-abc123key456") {
		t.Errorf("full auth_token should NOT appear in output, got: %q", output)
	}
	if strings.Contains(output, "sk-ant-xyz789") {
		t.Errorf("full auth_token should NOT appear in output, got: %q", output)
	}
	if !strings.Contains(output, "kimi-k2.6:cloud") {
		t.Errorf("output should contain sonnet model, got: %q", output)
	}
}

func TestList_EmptyProviders(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, "providers: {}\n")

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No providers configured") {
		t.Errorf("output should suggest adding a provider, got: %q", output)
	}
	if !strings.Contains(output, "ccswap add") {
		t.Errorf("output should mention 'ccswap add', got: %q", output)
	}
}

func TestList_MissingProvidersFile(t *testing.T) {
	configDir := t.TempDir()

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing providers.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "ccswap init") {
		t.Errorf("error should suggest 'ccswap init', got: %v", err)
	}
}

func TestList_ShowsActiveMarker(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	settingsContent := `{
		"env": {
			"ANTHROPIC_BASE_URL": "https://api.anthropic.com"
		}
	}`
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settingsContent), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	origReadSettings := readSettingsFile
	readSettingsFile = func(path string) (map[string]json.RawMessage, error) {
		data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
		if err != nil {
			return nil, err
		}
		result := make(map[string]json.RawMessage)
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	t.Cleanup(func() { readSettingsFile = origReadSettings })

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "*anthropic") {
		t.Errorf("active provider should be marked with *, got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("non-active provider should still appear, got: %q", output)
	}
}

func TestList_NoActiveMatch(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	settingsContent := `{
		"env": {
			"ANTHROPIC_BASE_URL": "https://nonexistent.example.com"
		}
	}`
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settingsContent), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	origReadSettings := readSettingsFile
	readSettingsFile = func(path string) (map[string]json.RawMessage, error) {
		data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
		if err != nil {
			return nil, err
		}
		result := make(map[string]json.RawMessage)
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
	t.Cleanup(func() { readSettingsFile = origReadSettings })

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "*") {
		t.Errorf("no provider should be marked active when base URLs don't match, got: %q", output)
	}
}

func TestList_NoEnvSettings(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origReadSettings := readSettingsFile
	readSettingsFile = func(path string) (map[string]json.RawMessage, error) {
		return nil, os.ErrNotExist
	}
	t.Cleanup(func() { readSettingsFile = origReadSettings })

	root := newListRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "*") {
		t.Errorf("no active marker should appear when settings.json missing, got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("providers should still be listed, got: %q", output)
	}
}
