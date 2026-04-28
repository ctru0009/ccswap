package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <provider>",
	Short: "Remove a provider profile",
	Long: `Remove a provider profile from providers.yaml.

If the provider is currently active, you will be prompted to confirm.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRemove(cmd, args[0])
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		configDir, err := resolveConfigDir(cmd)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		providersPath := providersPathFunc(configDir)
		cfg, err := config.LoadProviders(providersPath)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, 0, len(cfg.Providers))
		for name := range cfg.Providers {
			names = append(names, name)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
}

func runRemove(cmd *cobra.Command, providerName string) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	providersPath := providersPathFunc(configDir)
	statePath := statePathFunc(configDir)

	cfg, err := config.LoadProviders(providersPath)
	if err != nil {
		return fmt.Errorf("loading providers: %w", err)
	}

	if _, exists := cfg.Providers[providerName]; !exists {
		names := sortedProviderNames(cfg.Providers)
		return fmt.Errorf("provider %q not found\nAvailable providers: %s", providerName, strings.Join(names, ", "))
	}

	state, err := config.LoadState(statePath)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}
	if state.ActiveProvider == providerName {
		fmt.Fprintf(cmd.ErrOrStderr(), "Provider %s is currently active. Remove anyway? [y/N]: ", providerName)
		scanner := bufio.NewScanner(cmd.InOrStdin())
		if scanner.Scan() {
			answer := strings.TrimSpace(scanner.Text())
			if answer != "y" && answer != "Y" {
				fmt.Fprintln(cmd.ErrOrStderr(), "Removal cancelled.")
				return nil
			}
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), "Removal cancelled.")
			return nil
		}
	}

	delete(cfg.Providers, providerName)

	if err := config.SaveProviders(providersPath, cfg); err != nil {
		return fmt.Errorf("saving providers: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(cmd.ErrOrStderr(), "%s Removed %s.\n", green("✓"), providerName)

	return nil
}


