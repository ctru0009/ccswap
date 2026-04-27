package config

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "ccswap")
	}
	return filepath.Join(cfgDir, "ccswap")
}

func ProvidersPath() string {
	return filepath.Join(ConfigDir(), "providers.yaml")
}

func StatePath() string {
	return filepath.Join(ConfigDir(), "state.yaml")
}

func ClaudeSettingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

func ClaudeSettingsPath() string {
	return filepath.Join(ClaudeSettingsDir(), "settings.json")
}

func ClaudeSettingsBackupPath() string {
	return filepath.Join(ClaudeSettingsDir(), "settings.json.ccswap.bak")
}

func ClaudeSettingsTempPath() string {
	return filepath.Join(ClaudeSettingsDir(), "settings.json.ccswap.tmp")
}