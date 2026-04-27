package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadState(t *testing.T) {
	// Create a temp dir for the test file
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.yaml")

	// Write valid YAML
	validYAML := `active_provider: test-provider
last_switched: "2024-01-15T10:30:00Z"
`
	if err := os.WriteFile(stateFile, []byte(validYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}

	if state.ActiveProvider != "test-provider" {
		t.Errorf("expected ActiveProvider 'test-provider', got '%s'", state.ActiveProvider)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	if !state.LastSwitched.Equal(expectedTime) {
		t.Errorf("expected LastSwitched %v, got %v", expectedTime, state.LastSwitched)
	}
}

func TestSaveState_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.yaml")

	state := &State{
		ActiveProvider: "atomic-provider",
		LastSwitched:   time.Date(2024, 3, 20, 15, 45, 0, 0, time.UTC),
	}

	if err := SaveState(stateFile, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatal("state file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}

	if !strings.Contains(string(data), "atomic-provider") {
		t.Error("state file does not contain expected provider name")
	}
	if !strings.Contains(string(data), "2024-03-20") {
		t.Error("state file does not contain expected date")
	}

	// Verify atomic write: ensure no temp files left behind
	files, _ := os.ReadDir(tmpDir)
	if len(files) != 1 {
		t.Errorf("expected 1 file in temp dir, found %d: %v", len(files), files)
	}
}

func TestLoadState_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent.yaml")

	state, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState should not return error for missing file, got: %v", err)
	}

	if state.ActiveProvider != "" {
		t.Errorf("expected empty ActiveProvider, got '%s'", state.ActiveProvider)
	}
	if !state.LastSwitched.IsZero() {
		t.Errorf("expected zero LastSwitched, got %v", state.LastSwitched)
	}
}

func TestLoadState_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	if err := os.WriteFile(stateFile, []byte("invalid: [yaml: content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadState(stateFile)
	if err == nil {
		t.Fatal("LoadState should return error for invalid YAML")
	}
}

func TestSaveState_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nested", "deep", "state.yaml")

	state := &State{
		ActiveProvider: "nested-provider",
		LastSwitched:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}

	if err := SaveState(stateFile, state); err != nil {
		t.Fatalf("SaveState failed with nested directory: %v", err)
	}

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatal("state file should have been created")
	}

	loaded, err := LoadState(stateFile)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if loaded.ActiveProvider != "nested-provider" {
		t.Errorf("expected ActiveProvider 'nested-provider', got '%s'", loaded.ActiveProvider)
	}
}

func TestSaveState_AtomicWrite_NestedDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "config")
	stateFile := filepath.Join(nestedDir, "state.yaml")

	state := &State{
		ActiveProvider: "test-provider",
		LastSwitched:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := SaveState(stateFile, state); err != nil {
		t.Fatalf("SaveState returned error: %v", err)
	}

	files, _ := os.ReadDir(nestedDir)
	tmpCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			tmpCount++
		}
	}
	if tmpCount != 0 {
		t.Errorf("expected 0 temp files, found %d", tmpCount)
	}
}

func TestLoadState_InvalidYAML_IncludesPath(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(stateFile, []byte("invalid: [yaml: content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadState(stateFile)
	if err == nil {
		t.Fatal("LoadState should return error for invalid YAML")
	}
	if !strings.Contains(err.Error(), stateFile) {
		t.Errorf("error should contain file path %q, got: %v", stateFile, err)
	}
}