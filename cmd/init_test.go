package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newInitRootCmd returns a root command with init subcommand for test isolation.
func newInitRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(initCmd)
	return root
}

func TestInit_CreatesProvidersYaml(t *testing.T) {
	dir := t.TempDir()
	root := newInitRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(dir, "providers.yaml")
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		t.Errorf("providers.yaml should exist at %s", providersPath)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "providers.yaml") {
		t.Errorf("output should mention providers.yaml, got: %q", output)
	}
}

func TestInit_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	deepDir := filepath.Join(dir, "nested", "deep")
	root := newInitRootCmd()
	root.SetArgs([]string{"--config", deepDir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(deepDir, "providers.yaml")
	if _, err := os.Stat(providersPath); os.IsNotExist(err) {
		t.Errorf("providers.yaml should exist at %s", providersPath)
	}
}

func TestInit_ExistingFileErrors(t *testing.T) {
	dir := t.TempDir()
	providersPath := filepath.Join(dir, "providers.yaml")
	if err := os.WriteFile(providersPath, []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	root := newInitRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for existing providers.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected error about already exists, got: %v", err)
	}
}

func TestInit_EnvVarPrefillApiKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-key-12345")

	root := newInitRootCmd()
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(dir, "providers.yaml")
	data, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "sk-ant-test-key-12345") {
		t.Errorf("providers.yaml should contain the API key, got:\n%s", content)
	}
	if !strings.Contains(content, "  anthropic:") {
		t.Errorf("providers.yaml should have uncommented anthropic entry, got:\n%s", content)
	}
	if strings.Contains(content, "# anthropic:") {
		t.Errorf("providers.yaml should not have commented anthropic entry when key is set, got:\n%s", content)
	}
}

func TestInit_EnvVarPrefillAuthToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "sk-ant-auth-token-value")

	root := newInitRootCmd()
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(dir, "providers.yaml")
	data, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "sk-ant-auth-token-value") {
		t.Errorf("providers.yaml should contain the auth token, got:\n%s", content)
	}
}

func TestInit_EnvVarPrefillApiKeyPreferred(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "api-key-value")
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "auth-token-value")

	root := newInitRootCmd()
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(dir, "providers.yaml")
	data, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "api-key-value") {
		t.Errorf("ANTHROPIC_API_KEY should take precedence, got:\n%s", content)
	}
}

func TestInit_ContainsTemplateDefaults(t *testing.T) {
	dir := t.TempDir()

	root := newInitRootCmd()
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	providersPath := filepath.Join(dir, "providers.yaml")
	data, err := os.ReadFile(providersPath)
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)

	expectedParts := []string{
		"ccswap",
		"providers:",
		"# anthropic:",
		"claude-sonnet-4-20250514",
		"claude-opus-4-20250514",
		"claude-haiku-3-5-20250101",
		"https://api.anthropic.com",
	}
	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("providers.yaml should contain %q, got:\n%s", part, content)
		}
	}
}

func TestInit_NextStepsOutput(t *testing.T) {
	dir := t.TempDir()
	root := newInitRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", dir, "init"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedLines := []string{
		"Next steps",
		"ccswap list",
		"ccswap use",
	}
	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("output should contain %q, got: %q", line, output)
		}
	}
}
