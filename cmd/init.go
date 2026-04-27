package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize providers.yaml configuration",
	Long: `Creates the ccswap config directory and a providers.yaml template.

Non-interactive. Sets up the default structure for managing
Claude Code provider profiles.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit(cmd)
	},
}

func runInit(cmd *cobra.Command) error {
	configDir := config.ConfigDir()
	if cf, _ := cmd.Flags().GetString("config"); cf != "" {
		configDir = cf
	}

	providersPath := filepath.Join(configDir, "providers.yaml")

	if _, err := os.Stat(providersPath); err == nil {
		return fmt.Errorf("providers.yaml already exists. Use `ccswap add` to add providers")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	yamlContent := buildProvidersYAML()

	if err := os.WriteFile(providersPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("write providers.yaml: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	out := cmd.ErrOrStderr()
	fmt.Fprintf(out, "%s Created %s\n\n", green("✓"), providersPath)
	fmt.Fprintf(out, "Next steps:\n")
	fmt.Fprintf(out, "  1. Edit %s with your provider(s)\n", providersPath)
	fmt.Fprintf(out, "  2. Run %s to see available providers\n", green("ccswap list"))
	fmt.Fprintf(out, "  3. Run %s <provider> to switch\n", green("ccswap use"))

	return nil
}

func buildProvidersYAML() string {
	title := "# ccswap providers.yaml\n"
	comment := "# Uncomment and modify the anthropic entry below to get started.\n"
	comment += "# Then add additional providers with `ccswap add`.\n\n"

	yamlBody := "providers:\n"
	yamlBody += "  # anthropic:\n"
	yamlBody += "  #   auth_token: YOUR_API_KEY\n"
	yamlBody += "  #   base_url: https://api.anthropic.com\n"
	yamlBody += "  #   timeout_ms: 300000\n"
	yamlBody += "  #   models:\n"
	yamlBody += "  #     sonnet: claude-sonnet-4-20250514\n"
	yamlBody += "  #     opus: claude-opus-4-20250514\n"
	yamlBody += "  #     haiku: claude-haiku-3-5-20250101\n"

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	authToken := os.Getenv("ANTHROPIC_AUTH_TOKEN")

	token := apiKey
	if token == "" {
		token = authToken
	}

	if token != "" {
		yamlBody = "providers:\n"
		yamlBody += "  anthropic:\n"
		yamlBody += fmt.Sprintf("    auth_token: \"%s\"\n", token)
		yamlBody += "    base_url: https://api.anthropic.com\n"
		yamlBody += "    timeout_ms: 300000\n"
		yamlBody += "    models:\n"
		yamlBody += "      sonnet: claude-sonnet-4-20250514\n"
		yamlBody += "      opus: claude-opus-4-20250514\n"
		yamlBody += "      haiku: claude-haiku-3-5-20250101\n"
	}

	return title + comment + yamlBody
}
