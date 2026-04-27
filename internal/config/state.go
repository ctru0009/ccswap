package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type State struct {
	ActiveProvider string    `yaml:"active_provider"`
	LastSwitched   time.Time `yaml:"last_switched"`
}

func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("reading state file %s: %w", path, err)
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state file %s: %w", path, err)
	}

	return &state, nil
}

func SaveState(path string, state *State) error {
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state directory %s: %w", dir, err)
	}

	tmpFile := filepath.Join(dir, "state.tmp")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("writing state temp file %s: %w", tmpFile, err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmpFile, path, err)
	}

	return nil
}