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

func newE2ERootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(initCmd)
	root.AddCommand(useCmd)
	root.AddCommand(addCmd)
	root.AddCommand(statusCmd)
	root.AddCommand(listCmd)
	root.AddCommand(removeCmd)
	root.AddCommand(editCmd)
	root.AddCommand(importCmd)
	return root
}

func overridePathFuncs(t *testing.T, configDir, claudeDir string) {
	t.Helper()

	origProvidersPath := providersPathFunc
	origStatePath := statePathFunc
	origSettingsPath := settingsPathFunc
	origBackupPath := backupPathFunc
	origTempPath := tempPathFunc
	origClaudeDir := claudeDirFunc
	origStatusSettingsPath := statusSettingsPath
	origStatusProvidersPath := statusProvidersPath
	origStatusStatePath := statusStatePath
	origReadSettingsFile := readSettingsFile
	origImportSettingsPath := importSettingsPath

	providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
	settingsPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json") }
	backupPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.bak") }
	tempPathFunc = func(dir string) string { return filepath.Join(claudeDir, "settings.json.ccswap.tmp") }
	claudeDirFunc = func() string { return claudeDir }

	statusSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }
	statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
	statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }

	importSettingsPath = func() string { return filepath.Join(claudeDir, "settings.json") }

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

	t.Cleanup(func() {
		providersPathFunc = origProvidersPath
		statePathFunc = origStatePath
		settingsPathFunc = origSettingsPath
		backupPathFunc = origBackupPath
		tempPathFunc = origTempPath
		claudeDirFunc = origClaudeDir
		statusSettingsPath = origStatusSettingsPath
		statusProvidersPath = origStatusProvidersPath
		statusStatePath = origStatusStatePath
		readSettingsFile = origReadSettingsFile
		importSettingsPath = origImportSettingsPath
	})
}

func execCmd(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestE2E_InitThenAddThenUseThenStatus(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	root := newE2ERootCmd()
	output, err := execCmd(root, "--config", configDir, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("init should show checkmark, got: %q", output)
	}
	if _, err := os.Stat(filepath.Join(configDir, "providers.yaml")); os.IsNotExist(err) {
		t.Fatal("init should have created providers.yaml")
	}

	root = newE2ERootCmd()
	_, err = execCmd(root, "--config", configDir, "use", "zai")
	if err == nil {
		t.Fatal("use should fail when no providers are defined")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("use error should mention 'not found', got: %v", err)
	}

	input := buildAddInput("zai", "sk-ant-abc123key456", "https://example.com/v1", "kimi-k2.6:cloud", "deepseek-v4-flash:cloud", "glm-4.7:cloud", "3000000")
	r, w, _ := os.Pipe()
	root = newE2ERootCmd()
	root.SetIn(r)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	go func() { w.WriteString(input); w.Close() }()
	if err := root.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if !strings.Contains(buf.String(), "✓") {
		t.Errorf("add should show checkmark, got: %q", buf.String())
	}

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_BASE_URL": "https://example.com/v1",
		},
	})
	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "list"); err != nil {
		t.Fatalf("list failed: %v", err)
	} else if !strings.Contains(output, "zai") {
		t.Errorf("list should show 'zai', got: %q", output)
	}

	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{},
	})
	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "use", "zai"); err != nil {
		t.Fatalf("use zai failed: %v", err)
	} else if !strings.Contains(output, "Switched to zai") {
		t.Errorf("use should confirm switch, got: %q", output)
	}

	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if !strings.Contains(string(settingsData), "https://example.com/v1") {
		t.Errorf("settings.json should have zai's base URL, got: %s", string(settingsData))
	}

	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "status"); err != nil {
		t.Fatalf("status failed: %v", err)
	} else if !strings.Contains(output, "zai") || !strings.Contains(output, "Active provider") {
		t.Errorf("status should show zai as active, got: %q", output)
	}
}

func TestE2E_SwitchProvidersBackAndForth(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{"env": map[string]string{}})

	root := newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "use", "zai"); err != nil {
		t.Fatalf("use zai failed: %v", err)
	} else if !strings.Contains(output, "Switched to zai") {
		t.Errorf("should switch to zai, got: %q", output)
	}

	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if !strings.Contains(string(settingsData), "https://example.com/v1") {
		t.Errorf("settings.json should have zai's base URL, got: %s", string(settingsData))
	}

	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "use", "anthropic"); err != nil {
		t.Fatalf("use anthropic failed: %v", err)
	} else if !strings.Contains(output, "Switched to anthropic") {
		t.Errorf("should switch to anthropic, got: %q", output)
	}

	settingsData, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if !strings.Contains(string(settingsData), "https://api.anthropic.com") {
		t.Errorf("settings.json should have anthropic's base URL, got: %s", string(settingsData))
	}

	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "status"); err != nil {
		t.Fatalf("status failed: %v", err)
	} else if !strings.Contains(output, "anthropic") {
		t.Errorf("status should show anthropic as active, got: %q", output)
	}
}

func TestE2E_AddRemoveList(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, validProvidersYAML)

	input := buildAddInput("mock", "mock-token", "https://mock.example.com/v1", "mock-sonnet", "mock-opus", "mock-haiku", "3000000")
	r, w, _ := os.Pipe()
	root := newE2ERootCmd()
	root.SetIn(r)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	go func() { w.WriteString(input); w.Close() }()
	if err := root.Execute(); err != nil {
		t.Fatalf("add mock failed: %v", err)
	}

	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "list"); err != nil {
		t.Fatalf("list after add failed: %v", err)
	} else {
		if !strings.Contains(output, "zai") || !strings.Contains(output, "mock") {
			t.Errorf("list should show both providers, got: %q", output)
		}
	}

	r2, w2, _ := os.Pipe()
	root = newE2ERootCmd()
	root.SetIn(r2)
	buf2 := new(bytes.Buffer)
	root.SetOut(buf2)
	root.SetErr(buf2)
	root.SetArgs([]string{"--config", configDir, "remove", "mock"})
	go func() { w2.WriteString("y\n"); w2.Close() }()
	if err := root.Execute(); err != nil {
		t.Fatalf("remove mock failed: %v", err)
	}
	if !strings.Contains(buf2.String(), "Removed") {
		t.Errorf("remove should confirm removal, got: %q", buf2.String())
	}

	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "list"); err != nil {
		t.Fatalf("list after remove failed: %v", err)
	} else {
		if !strings.Contains(output, "zai") {
			t.Errorf("list should still show 'zai', got: %q", output)
		}
		if strings.Contains(output, "mock") {
			t.Errorf("list should NOT show 'mock' after removal, got: %q", output)
		}
	}
}

func TestE2E_SettingsPreservationAcrossSwitches(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{},
		"permissions": map[string]interface{}{
			"allow": []string{"Read", "Write"},
		},
		"hooks": map[string]interface{}{
			"pre_commit": "run-linter",
		},
	})

	root := newE2ERootCmd()
	if _, err := execCmd(root, "--config", configDir, "use", "zai"); err != nil {
		t.Fatalf("use zai failed: %v", err)
	}

	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	for _, key := range []string{"permissions", "hooks", "pre_commit"} {
		if !strings.Contains(string(settingsData), key) {
			t.Errorf("settings.json should still contain %q after use zai", key)
		}
	}

	root = newE2ERootCmd()
	if _, err := execCmd(root, "--config", configDir, "use", "anthropic"); err != nil {
		t.Fatalf("use anthropic failed: %v", err)
	}

	settingsData, _ = os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	for _, key := range []string{"permissions", "hooks", "pre_commit"} {
		if !strings.Contains(string(settingsData), key) {
			t.Errorf("settings.json should still contain %q after use anthropic", key)
		}
	}
}

func TestE2E_MissingConfigDir(t *testing.T) {
	emptyDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, emptyDir, claudeDir)

	root := newE2ERootCmd()
	_, err := execCmd(root, "--config", emptyDir, "list")
	if err == nil {
		t.Fatal("list should error when providers.yaml is missing")
	}
	if !strings.Contains(err.Error(), "ccswap init") {
		t.Errorf("error should suggest 'ccswap init', got: %v", err)
	}
}

func TestE2E_EmptyProvidersYAML(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, "providers: {}\n")

	root := newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "list"); err != nil {
		t.Fatalf("list with empty providers should not error, got: %v", err)
	} else {
		if !strings.Contains(output, "No providers configured") {
			t.Errorf("should say no providers configured, got: %q", output)
		}
		if !strings.Contains(output, "ccswap add") {
			t.Errorf("should suggest 'ccswap add', got: %q", output)
		}
	}
}

func TestE2E_MalformedSettingsJSON(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, validProvidersYAML)
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}
	malformed := "{invalid}"
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(malformed), 0644); err != nil {
		t.Fatalf("failed to write malformed settings: %v", err)
	}

	root := newE2ERootCmd()
	_, err := execCmd(root, "--config", configDir, "use", "zai")
	if err == nil {
		t.Fatal("use should error with malformed settings.json")
	}
	if !strings.Contains(err.Error(), "settings.json") {
		t.Errorf("error should mention settings.json, got: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if string(data) != malformed {
		t.Errorf("settings.json should be unchanged after error, got: %q", string(data))
	}
}

func TestE2E_ConcurrentUseSimulated(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	writeTestProviders(t, configDir, validProvidersYAML)
	writeTestSettings(t, claudeDir, map[string]interface{}{"env": map[string]string{}})

	root := newE2ERootCmd()
	if _, err := execCmd(root, "--config", configDir, "use", "zai"); err != nil {
		t.Fatalf("use zai failed: %v", err)
	}

	root = newE2ERootCmd()
	if _, err := execCmd(root, "--config", configDir, "use", "anthropic"); err != nil {
		t.Fatalf("use anthropic failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("settings.json should be valid JSON, got parse error: %v\ncontent: %s", err, string(data))
	}

	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		t.Fatal("settings.json should have env object")
	}
	if env["ANTHROPIC_BASE_URL"] != "https://api.anthropic.com" {
		t.Errorf("last use (anthropic) should be active, got ANTHROPIC_BASE_URL: %v", env["ANTHROPIC_BASE_URL"])
	}
}

func TestE2E_ImportThenUse(t *testing.T) {
	configDir := t.TempDir()
	claudeDir := t.TempDir()
	overridePathFuncs(t, configDir, claudeDir)

	// Set up settings.json with a provider config not yet in providers.yaml
	writeTestSettings(t, claudeDir, map[string]interface{}{
		"env": map[string]string{
			"ANTHROPIC_AUTH_TOKEN":            "sk-ant-imported-key12345",
			"ANTHROPIC_BASE_URL":              "https://imported.example.com/v1",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":  "imported-sonnet",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":    "imported-opus",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":   "imported-haiku",
			"API_TIMEOUT_MS":                  "500000",
		},
	})

	// Create empty providers.yaml
	writeTestProviders(t, configDir, "providers: {}\n")

	// Import the provider
	root := newE2ERootCmd()
	input := "imported-prov\n"
	r, w, _ := os.Pipe()
	root.SetIn(r)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "import"})
	go func() { w.WriteString(input); w.Close() }()

	if err := root.Execute(); err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !strings.Contains(buf.String(), "imported") {
		t.Errorf("import should confirm success, got: %q", buf.String())
	}

	// Verify it was saved to providers.yaml
	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "imported-prov") {
		t.Errorf("providers.yaml should contain 'imported-prov', got: %s", content)
	}
	if !strings.Contains(content, "https://imported.example.com/v1") {
		t.Errorf("providers.yaml should contain base_url, got: %s", content)
	}

	// Now use the imported provider
	root = newE2ERootCmd()
	if output, err := execCmd(root, "--config", configDir, "use", "imported-prov"); err != nil {
		t.Fatalf("use imported-prov failed: %v", err)
	} else if !strings.Contains(output, "Switched to imported-prov") {
		t.Errorf("use should confirm switch, got: %q", output)
	}

	// Verify settings.json has the imported provider's values
	settingsData, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	if !strings.Contains(string(settingsData), "https://imported.example.com/v1") {
		t.Errorf("settings.json should have imported base URL, got: %s", string(settingsData))
	}
	if !strings.Contains(string(settingsData), "imported-sonnet") {
		t.Errorf("settings.json should have imported sonnet model, got: %s", string(settingsData))
	}
}