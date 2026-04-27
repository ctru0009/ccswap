package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSettings_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-test-123",
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com"
  },
  "permissions": {
    "allow": ["Bash"]
  }
}`
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := ReadSettings(path)
	if err != nil {
		t.Fatalf("ReadSettings() error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 top-level keys, got %d", len(result))
	}

	envRaw, ok := result["env"]
	if !ok {
		t.Fatal("expected 'env' key in result")
	}

	var env map[string]string
	if err := json.Unmarshal(envRaw, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-test-123" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "sk-test-123")
	}
}

func TestReadSettings_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	result, err := ReadSettings(path)
	if err != nil {
		t.Fatalf("ReadSettings() error for missing file: %v", err)
	}

	envRaw, ok := result["env"]
	if !ok {
		t.Fatal("expected 'env' key in default result")
	}

	var env map[string]interface{}
	if err := json.Unmarshal(envRaw, &env); err != nil {
		t.Fatalf("unmarshal default env: %v", err)
	}
	if len(env) != 0 {
		t.Errorf("expected empty env map, got %d keys", len(env))
	}
}

func TestReadSettings_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{invalid json`), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := ReadSettings(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), path) {
		t.Errorf("error message should contain file path %q, got: %v", path, err)
	}
}

func TestReadSettings_NoEnvKey(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "permissions": {
    "allow": ["Bash"]
  }
}`
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := ReadSettings(path)
	if err != nil {
		t.Fatalf("ReadSettings() error: %v", err)
	}

	if _, ok := result["env"]; ok {
		t.Error("should not have 'env' key when file doesn't contain one")
	}

	if _, ok := result["permissions"]; !ok {
		t.Error("expected 'permissions' key in result")
	}
}

func TestReadSettings_PreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	originalPermissions := `{"allow":["Bash","Read"]}`
	originalModels := `{"sonnet":"claude-3-5-sonnet"}`
	originalHooks := `{"pre_tool_use":"script.sh"}`
	content := `{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "sk-test"
  },
  "permissions": {"allow":["Bash","Read"]},
  "models": {"sonnet":"claude-3-5-sonnet"},
  "hooks": {"pre_tool_use":"script.sh"}
}`
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	result, err := ReadSettings(path)
	if err != nil {
		t.Fatalf("ReadSettings() error: %v", err)
	}

	if string(result["permissions"]) != originalPermissions {
		t.Errorf("permissions = %q; want %q", string(result["permissions"]), originalPermissions)
	}
	if string(result["models"]) != originalModels {
		t.Errorf("models = %q; want %q", string(result["models"]), originalModels)
	}
	if string(result["hooks"]) != originalHooks {
		t.Errorf("hooks = %q; want %q", string(result["hooks"]), originalHooks)
	}
}

func TestWriteSettings_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	backupPath := filepath.Join(dir, "settings.json.bak")
	tempPath := filepath.Join(dir, "settings.json.tmp")

	data := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-new"}`),
	}

	if err := WriteSettings(path, data, backupPath, tempPath); err != nil {
		t.Fatalf("WriteSettings() error: %v", err)
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back settings: %v", err)
	}

	result := make(map[string]json.RawMessage)
	if err := json.Unmarshal(readBack, &result); err != nil {
		t.Fatalf("unmarshal written settings: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-new" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "sk-new")
	}

	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after atomic write")
	}
}

func TestWriteSettings_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	backupPath := filepath.Join(dir, "settings.json.bak")
	tempPath := filepath.Join(dir, "settings.json.tmp")

	originalContent := `{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-old"}}`
	if err := os.WriteFile(path, []byte(originalContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	data := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-new"}`),
	}

	if err := WriteSettings(path, data, backupPath, tempPath); err != nil {
		t.Fatalf("WriteSettings() error: %v", err)
	}

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}

	if string(backupData) != originalContent {
		t.Errorf("backup = %q; want %q", string(backupData), originalContent)
	}
}

func TestWriteSettings_BackupFails_AbortsWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	backupPath := filepath.Join(dir, "readonly", "settings.json.bak")
	tempPath := filepath.Join(dir, "settings.json.tmp")

	originalContent := `{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-original"}}`
	if err := os.WriteFile(path, []byte(originalContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	data := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-new"}`),
	}

	err := WriteSettings(path, data, backupPath, tempPath)
	if err == nil {
		t.Fatal("expected error when backup fails")
	}

	currentContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading current file: %v", err)
	}

	if string(currentContent) != originalContent {
		t.Errorf("original file should be unchanged; got %q", string(currentContent))
	}
}

func TestCreateClaudeDir(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")

	if err := CreateClaudeDir(claudeDir); err != nil {
		t.Fatalf("CreateClaudeDir() error: %v", err)
	}

	info, err := os.Stat(claudeDir)
	if err != nil {
		t.Fatalf("stat claude dir: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}

	if err := CreateClaudeDir(claudeDir); err != nil {
		t.Fatalf("CreateClaudeDir() on existing dir error: %v", err)
	}
}

func TestReadSettings_NullJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("null"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := ReadSettings(path)
	if err == nil {
		t.Fatal("expected error for null JSON, got nil")
	}
	if !strings.Contains(err.Error(), "must be a JSON object") {
		t.Errorf("error should mention 'must be a JSON object', got: %v", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error should contain file path %q, got: %v", path, err)
	}
}

func TestReadSettings_ArrayJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := ReadSettings(path)
	if err == nil {
		t.Fatal("expected error for array JSON, got nil")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error should contain file path %q, got: %v", path, err)
	}
}

func TestWriteSettings_OverwritesStaleTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	backupPath := filepath.Join(dir, "settings.json.ccswap.bak")
	tempPath := filepath.Join(dir, "settings.json.ccswap.tmp")

	staleContent := "stale temp file content"
	if err := os.WriteFile(tempPath, []byte(staleContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	data := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-test"}`),
	}

	if err := WriteSettings(path, data, backupPath, tempPath); err != nil {
		t.Fatalf("WriteSettings() error: %v", err)
	}

	readBack, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back settings: %v", err)
	}

	result := make(map[string]json.RawMessage)
	if err := json.Unmarshal(readBack, &result); err != nil {
		t.Fatalf("unmarshal written settings: %v", err)
	}

	var env map[string]string
	if err := json.Unmarshal(result["env"], &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "sk-test" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q; want %q", env["ANTHROPIC_AUTH_TOKEN"], "sk-test")
	}

	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("temp file should not exist after atomic write")
	}
}

func TestWriteSettings_OverwritesExistingBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	backupPath := filepath.Join(dir, "settings.json.ccswap.bak")
	tempPath := filepath.Join(dir, "settings.json.ccswap.tmp")

	originalContent := `{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-original"}}`
	if err := os.WriteFile(path, []byte(originalContent), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	firstData := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-first"}`),
	}
	if err := WriteSettings(path, firstData, backupPath, tempPath); err != nil {
		t.Fatalf("first WriteSettings() error: %v", err)
	}

	firstBackup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading first backup: %v", err)
	}
	if string(firstBackup) != originalContent {
		t.Errorf("first backup should contain original content, got: %q", string(firstBackup))
	}

	secondData := map[string]json.RawMessage{
		"env": json.RawMessage(`{"ANTHROPIC_AUTH_TOKEN":"sk-second"}`),
	}
	if err := WriteSettings(path, secondData, backupPath, tempPath); err != nil {
		t.Fatalf("second WriteSettings() error: %v", err)
	}

	secondBackup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("reading second backup: %v", err)
	}
	if string(secondBackup) == originalContent {
		t.Error("second backup should have been overwritten, but still contains original content")
	}
}