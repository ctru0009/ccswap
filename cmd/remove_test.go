package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/spf13/cobra"
)

func newRemoveRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(removeCmd)
	return root
}

func writeTestState(t *testing.T, dir string, activeProvider string) {
	t.Helper()
	if activeProvider == "" {
		return
	}
	stateYAML := "active_provider: " + activeProvider + "\nlast_switched: 2025-01-01T00:00:00Z\n"
	if err := os.WriteFile(filepath.Join(dir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatalf("failed to write state.yaml: %v", err)
	}
}

func TestRemove_Success(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "Removed") {
		t.Errorf("output should contain 'Removed', got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider name, got: %q", output)
	}

	// Verify provider was actually removed from file
	cfg, err := config.LoadProviders(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to reload providers: %v", err)
	}
	if _, exists := cfg.Providers["zai"]; exists {
		t.Error("provider 'zai' should have been removed")
	}
	if _, exists := cfg.Providers["anthropic"]; !exists {
		t.Error("provider 'anthropic' should still exist")
	}
}

func TestRemove_ProviderNotFound(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "nonexistent"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

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

	// Verify providers file was not modified
	cfg, err := config.LoadProviders(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to reload providers: %v", err)
	}
	if _, exists := cfg.Providers["zai"]; !exists {
		t.Error("provider 'zai' should still exist after failed removal")
	}
}

func TestRemove_ActiveProviderCancelled(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestState(t, configDir, "zai")

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "zai"})

	// Mock stdin: answer "n" (default behavior)
	r, w, _ := os.Pipe()
	root.SetIn(r)
	w.WriteString("n\n")
	w.Close()

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "cancelled") {
		t.Errorf("output should mention 'cancelled', got: %q", output)
	}
	if strings.Contains(output, "✓") {
		t.Errorf("output should NOT contain checkmark when cancelled, got: %q", output)
	}

	// Verify provider was NOT removed
	cfg, err := config.LoadProviders(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to reload providers: %v", err)
	}
	if _, exists := cfg.Providers["zai"]; !exists {
		t.Error("provider 'zai' should still exist after cancellation")
	}
}

func TestRemove_ActiveProviderConfirmed(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestState(t, configDir, "zai")

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "zai"})

	// Mock stdin: answer "y"
	r, w, _ := os.Pipe()
	root.SetIn(r)
	w.WriteString("y\n")
	w.Close()

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "Removed") {
		t.Errorf("output should contain 'Removed', got: %q", output)
	}
	if !strings.Contains(output, "zai") {
		t.Errorf("output should contain provider name, got: %q", output)
	}

	// Verify provider was actually removed
	cfg, err := config.LoadProviders(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to reload providers: %v", err)
	}
	if _, exists := cfg.Providers["zai"]; exists {
		t.Error("provider 'zai' should have been removed after confirmation")
	}
}

func TestRemove_NoArgs(t *testing.T) {
	root := newRemoveRootCmd()
	root.SetArgs([]string{"remove"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no provider arg given, got nil")
	}
}

func TestRemove_ProvidersFileMissing(t *testing.T) {
	configDir := t.TempDir()

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "zai"})

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing providers.yaml, got nil")
	}
}

func TestRemove_ActiveProviderCapitalY(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestState(t, configDir, "zai")

	root := newRemoveRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "remove", "zai"})

	// Mock stdin: answer capital "Y"
	r, w, _ := os.Pipe()
	root.SetIn(r)
	w.WriteString("Y\n")
	w.Close()

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark for 'Y', got: %q", output)
	}
}
