package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ctru0009/ccswap/internal/claude"
	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <provider>",
	Short: "Switch to a provider profile",
	Long: `Switch Claude Code to use the specified provider profile.
Edits ~/.claude/settings.json to set the provider's auth token,
base URL, and model configuration.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeProviderNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUse(cmd, args[0])
	},
}

var providersPathFunc = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
var statePathFunc = func(dir string) string { return filepath.Join(dir, "state.yaml") }
var settingsPathFunc = func(dir string) string { return config.ClaudeSettingsPath() }
var backupPathFunc = func(dir string) string { return config.ClaudeSettingsBackupPath() }
var tempPathFunc = func(dir string) string { return config.ClaudeSettingsTempPath() }
var claudeDirFunc = func() string { return config.ClaudeSettingsDir() }

func runUse(cmd *cobra.Command, providerName string) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	providersPath := providersPathFunc(configDir)
	statePath := statePathFunc(configDir)
	settingsPath := settingsPathFunc(configDir)
	backupPath := backupPathFunc(configDir)
	tempPath := tempPathFunc(configDir)
	claudeDir := claudeDirFunc()

	cfg, err := config.LoadProviders(providersPath)
	if err != nil {
		return err
	}

	provider, exists := cfg.Providers[providerName]
	if !exists {
		names := sortedProviderNames(cfg.Providers)
		return fmt.Errorf("provider %q not found\nAvailable providers: %s", providerName, strings.Join(names, ", "))
	}

	state, err := config.LoadState(statePath)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	if state.ActiveProvider == providerName {
		fmt.Fprintf(cmd.ErrOrStderr(), "Provider %s is already active.\n", providerName)
		return nil
	}

	if err := claude.CreateClaudeDir(claudeDir); err != nil {
		return fmt.Errorf("creating claude directory: %w", err)
	}

	settings, err := claude.ReadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("settings file %s: %w", settingsPath, err)
	}

	merged, err := claude.MergeEnv(settings, provider)
	if err != nil {
		return fmt.Errorf("merging env: %w", err)
	}

	if err := claude.WriteSettings(settingsPath, merged, backupPath, tempPath); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	newState := &config.State{
		ActiveProvider: providerName,
		LastSwitched:   time.Now(),
	}
	if err := config.SaveState(statePath, newState); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	expanded, err := config.ExpandProvider(provider)
	if err != nil {
		expanded = provider
	}

	printConfirmation(cmd, providerName, expanded)
	return nil
}

func resolveConfigDir(cmd *cobra.Command) (string, error) {
	if cf, _ := cmd.Flags().GetString("config"); cf != "" {
		return cf, nil
	}
	return config.ConfigDir(), nil
}

func completeProviderNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	cfg, err := config.LoadProviders(providersPathFunc(configDir))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return sortedProviderNames(cfg.Providers), cobra.ShellCompDirectiveNoFileComp
}

func sortedProviderNames(providers map[string]config.Provider) []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func printConfirmation(cmd *cobra.Command, name string, p config.Provider) {
	out := cmd.ErrOrStderr()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Fprintf(out, "%s Switched to %s\n", green("✓"), name)
	fmt.Fprintf(out, "  Base URL:    %s\n", p.BaseURL)
	fmt.Fprintf(out, "  Sonnet:      %s\n", p.Models.Sonnet)
	fmt.Fprintf(out, "  Opus:        %s\n", p.Models.Opus)
	fmt.Fprintf(out, "  Haiku:       %s\n", p.Models.Haiku)
	fmt.Fprintf(out, "  Auth:        %s\n", config.MaskKey(p.AuthToken))
	fmt.Fprintf(out, "\nOpen a new Claude Code session for changes to take effect.\n")
}