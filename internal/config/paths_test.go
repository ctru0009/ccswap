package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		cfgDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("os.UserConfigDir() failed: %v", err)
		}
		expected := filepath.Join(cfgDir, "ccswap")
		result := ConfigDir()
		if result != expected {
			t.Errorf("ConfigDir() = %q; want %q", result, expected)
		}
	})
}

func TestProvidersPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		cfgDir, _ := os.UserConfigDir()
		expected := filepath.Join(cfgDir, "ccswap", "providers.yaml")
		result := ProvidersPath()
		if result != expected {
			t.Errorf("ProvidersPath() = %q; want %q", result, expected)
		}
	})
}

func TestStatePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		cfgDir, _ := os.UserConfigDir()
		expected := filepath.Join(cfgDir, "ccswap", "state.yaml")
		result := StatePath()
		if result != expected {
			t.Errorf("StatePath() = %q; want %q", result, expected)
		}
	})
}

func TestClaudeSettingsDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude")
		result := ClaudeSettingsDir()
		if result != expected {
			t.Errorf("ClaudeSettingsDir() = %q; want %q", result, expected)
		}
	})
}

func TestClaudeSettingsPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json")
		result := ClaudeSettingsPath()
		if result != expected {
			t.Errorf("ClaudeSettingsPath() = %q; want %q", result, expected)
		}
	})
}

func TestClaudeSettingsBackupPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json.ccswap.bak")
		result := ClaudeSettingsBackupPath()
		if result != expected {
			t.Errorf("ClaudeSettingsBackupPath() = %q; want %q", result, expected)
		}
	})
}

func TestClaudeSettingsTempPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping linux test on windows")
	}
	t.Run("linux", func(t *testing.T) {
		t.Setenv("HOME", "/home/testuser")
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("APPDATA", "")

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json.ccswap.tmp")
		result := ClaudeSettingsTempPath()
		if result != expected {
			t.Errorf("ClaudeSettingsTempPath() = %q; want %q", result, expected)
		}
	})
}