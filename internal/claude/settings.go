package claude

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadSettings reads and parses a settings.json file.
// If the file is missing, it returns a minimal {"env":{}} map (not an error).
// If the JSON is malformed, the error message includes the file path.
func ReadSettings(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			result := make(map[string]json.RawMessage)
			result["env"] = json.RawMessage(`{}`)
			return result, nil
		}
		return nil, fmt.Errorf("reading settings file %s: %w", path, err)
	}

	result := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing settings file %s: %w", path, err)
	}

	// json.Unmarshal of `null` into a map sets it to nil rather than erroring.
	// We treat this as invalid: settings.json must be a JSON object.
	if result == nil {
		return nil, fmt.Errorf("settings file %s: must be a JSON object, got null", path)
	}

	return result, nil
}

// WriteSettings atomically writes settings data to the given path.
// It first creates a backup at backupPath, then writes to tempPath and renames
// to the final path. If the backup step fails, the entire operation is aborted
// and the original file is left unchanged.
func WriteSettings(path string, data map[string]json.RawMessage, backupPath string, tempPath string) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading existing settings for backup: %w", err)
	}
	if err == nil {
		if err := os.WriteFile(backupPath, existing, 0644); err != nil {
			return fmt.Errorf("creating backup at %s: %w", backupPath, err)
		}
	}

	if err := os.WriteFile(tempPath, out, 0644); err != nil {
		return fmt.Errorf("writing temp file %s: %w", tempPath, err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tempPath, path, err)
	}

	return nil
}

// CreateClaudeDir creates the ~/.claude/ directory if it doesn't exist.
func CreateClaudeDir(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dirPath, err)
	}
	return nil
}