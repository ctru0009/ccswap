package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	Long:  `Display all configured providers in a table with masked auth tokens.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(cmd)
	},
}

func runList(cmd *cobra.Command) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	providersPath := filepath.Join(configDir, "providers.yaml")

	cfg, err := config.LoadProviders(providersPath)
	if err != nil {
		return err
	}

	if len(cfg.Providers) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "No providers configured. Run `ccswap add` to add one.")
		return nil
	}

	settingsPath := config.ClaudeSettingsPath()
	activeBaseURL := ""
	if settings, err := readSettingsFile(settingsPath); err == nil {
		if envRaw, ok := settings["env"]; ok {
			var env map[string]string
			if err := json.Unmarshal(envRaw, &env); err == nil {
				activeBaseURL = env["ANTHROPIC_BASE_URL"]
			}
		}
	}

	names := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		names = append(names, name)
	}
	sort.Strings(names)

	out := cmd.OutOrStdout()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	table := tablewriter.NewTable(out,
		tablewriter.WithHeader([]string{"PROVIDER", "BASE URL", "SONNET", "OPUS", "HAIKU", "AUTH"}),
	)

	for _, name := range names {
		provider := cfg.Providers[name]

		displayName := name
		if provider.BaseURL == activeBaseURL {
			displayName = cyan("*" + name)
		}

		table.Append(displayName, provider.BaseURL, provider.Models.Sonnet, provider.Models.Opus, provider.Models.Haiku, config.MaskKey(provider.AuthToken))
	}

	return table.Render()
}

var readSettingsFile = func(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
