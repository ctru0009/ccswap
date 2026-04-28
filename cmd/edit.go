package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var editCmd = &cobra.Command{
	Use:   "edit <provider>",
	Short: "Edit a provider profile in $EDITOR",
	Long: `Open the specified provider's configuration in your editor ($EDITOR).
After editing, validates the result and updates providers.yaml.
Provider name changes are not allowed — use remove + add instead.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runEdit(cmd, args[0])
	},
}

var editorLaunchFunc = launchEditor
var editorLookupFunc = exec.LookPath

func runEdit(cmd *cobra.Command, providerName string) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	providersPath := providersPathFunc(configDir)

	cfg, err := config.LoadProviders(providersPath)
	if err != nil {
		return fmt.Errorf("loading providers: %w", err)
	}

	provider, exists := cfg.Providers[providerName]
	if !exists {
		names := sortedProviderNames(cfg.Providers)
		return fmt.Errorf("provider %q not found\nAvailable providers: %s", providerName, strings.Join(names, ", "))
	}

	tmpFile, err := os.CreateTemp("", "ccswap-edit-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	tmpData := map[string]config.Provider{providerName: provider}
	tmpBytes, err := yaml.Marshal(tmpData)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("marshal provider: %w", err)
	}

	if _, err := tmpFile.Write(tmpBytes); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	editor := resolveEditor()
	if err := editorLaunchFunc(editor, tmpPath); err != nil {
		return fmt.Errorf("editor: %w", err)
	}

	editedData, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("read temp file: %w", err)
	}

	var edited map[string]config.Provider
	if err := yaml.Unmarshal(editedData, &edited); err != nil {
		return fmt.Errorf("parse edited YAML: %w", err)
	}

	if len(edited) != 1 {
		return fmt.Errorf("edited file must contain exactly one provider, found %d", len(edited))
	}

	var editedName string
	var editedProvider config.Provider
	for k, v := range edited {
		editedName = k
		editedProvider = v
	}

	if editedName != providerName {
		return fmt.Errorf("provider name change not allowed: %q → %q; use remove + add to rename", providerName, editedName)
	}

	if err := config.ValidateProvider(editedProvider); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	cfg.Providers[providerName] = editedProvider

	if err := config.SaveProviders(providersPath, cfg); err != nil {
		return fmt.Errorf("save providers: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(cmd.ErrOrStderr(), "%s Updated provider %s\n", green("✓"), providerName)

	return nil
}

func resolveEditor() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	if _, err := editorLookupFunc("nano"); err == nil {
		return "nano"
	}
	return "vi"
}

func launchEditor(editor string, tmpPath string) error {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return fmt.Errorf("editor command is empty")
	}
	args := append(parts[1:], tmpPath)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}