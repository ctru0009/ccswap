package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ctru0009/ccswap/internal/claude"
	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active provider status",
	Long: `Show the currently active provider and its configuration.
Displays the provider name, base URL, models, auth token (masked),
and last switched time.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(cmd)
	},
}

var statusSettingsPath = func() string { return config.ClaudeSettingsPath() }
var statusProvidersPath = func(dir string) string { return filepath.Join(dir, "providers.yaml") }
var statusStatePath = func(dir string) string { return filepath.Join(dir, "state.yaml") }

func runStatus(cmd *cobra.Command) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	settings, err := claude.ReadSettings(statusSettingsPath())
	if err != nil {
		out := cmd.ErrOrStderr()
		fmt.Fprintln(out, "No provider configured. Run `ccswap init`")
		return nil
	}

	var envMap map[string]string
	if envRaw, ok := settings["env"]; ok {
		if err := json.Unmarshal(envRaw, &envMap); err != nil {
			envMap = make(map[string]string)
		}
	} else {
		envMap = make(map[string]string)
	}

	baseURL := envMap["ANTHROPIC_BASE_URL"]
	authToken := envMap["ANTHROPIC_AUTH_TOKEN"]

	if baseURL == "" {
		out := cmd.ErrOrStderr()
		fmt.Fprintln(out, "No provider configured. Run `ccswap init`")
		return nil
	}

	providersPath := statusProvidersPath(configDir)
	cfg, err := config.LoadProviders(providersPath)
	var matchedProvider *config.Provider
	var matchedName string

	if err == nil {
		for name, p := range cfg.Providers {
			expanded, expandErr := config.ExpandProvider(p)
			if expandErr != nil {
				continue
			}
			if expanded.BaseURL == baseURL {
				cp := p
				matchedProvider = &cp
				matchedName = name
				break
			}
		}
	}

	statePath := statusStatePath(configDir)
	state, _ := config.LoadState(statePath)

	out := cmd.OutOrStdout()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

	if matchedName != "" {
		fmt.Fprintf(out, "Active provider: %s\n", cyan(matchedName))
		fmt.Fprintf(out, "Base URL:        %s\n", baseURL)
		fmt.Fprintf(out, "Sonnet:          %s\n", matchedProvider.Models.Sonnet)
		fmt.Fprintf(out, "Opus:            %s\n", matchedProvider.Models.Opus)
		fmt.Fprintf(out, "Haiku:           %s\n", matchedProvider.Models.Haiku)
		fmt.Fprintf(out, "Auth:            %s\n", config.MaskKey(authToken))
		if !state.LastSwitched.IsZero() {
			fmt.Fprintf(out, "Last switched:   %s\n", state.LastSwitched.Format(time.RFC3339))
		}
	} else {
		fmt.Fprintln(out, "Active provider: unknown (manually configured or not set)")
		fmt.Fprintf(out, "Base URL:        %s\n", baseURL)
		if authToken != "" {
			fmt.Fprintf(out, "Auth:            %s\n", config.MaskKey(authToken))
		}
		if !state.LastSwitched.IsZero() {
			fmt.Fprintf(out, "Last switched:   %s\n", state.LastSwitched.Format(time.RFC3339))
		}
	}

	return nil
}
