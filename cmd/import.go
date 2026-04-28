package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/ctru0009/ccswap/internal/claude"
	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import provider from Claude Code settings",
	Long: `Detect the provider configuration in ~/.claude/settings.json
and save it as a named provider profile in providers.yaml.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImport(cmd)
	},
}

var importSettingsPath = func() string { return config.ClaudeSettingsPath() }

func runImport(cmd *cobra.Command) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	settingsPath := importSettingsPath()

	settings, err := claude.ReadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("reading settings: %w", err)
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
	if baseURL == "" {
		out := cmd.ErrOrStderr()
		fmt.Fprintln(out, "No provider configuration found in settings.json.")
		fmt.Fprintln(out, "Run `ccswap add` to add a provider manually.")
		return nil
	}

	authToken := envMap["ANTHROPIC_AUTH_TOKEN"]
	sonnet := envMap["ANTHROPIC_DEFAULT_SONNET_MODEL"]
	opus := envMap["ANTHROPIC_DEFAULT_OPUS_MODEL"]
	haiku := envMap["ANTHROPIC_DEFAULT_HAIKU_MODEL"]
	timeoutStr := envMap["API_TIMEOUT_MS"]

	timeoutMs := 0
	if timeoutStr != "" {
		timeoutMs, err = strconv.Atoi(timeoutStr)
		if err != nil {
			timeoutMs = 0
		}
	}

	providersPath := filepath.Join(configDir, "providers.yaml")
	cfg, err := loadOrInitProviders(providersPath)
	if err != nil {
		return fmt.Errorf("loading providers: %w", err)
	}

	// Check if an existing provider's expanded base_url matches
	for name, p := range cfg.Providers {
		expanded, expandErr := config.ExpandProvider(p)
		if expandErr != nil {
			continue
		}
		if expanded.BaseURL == baseURL {
			out := cmd.ErrOrStderr()
			cyan := color.New(color.FgCyan).SprintFunc()
			fmt.Fprintf(out, "This configuration matches provider %s.\n", cyan(name))
			fmt.Fprintf(out, "Run %s to switch to it.\n", cyan("ccswap use "+name))
			return nil
		}
	}

	out := cmd.ErrOrStderr()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Fprintf(out, "%s Detected provider configuration:\n", yellow("→"))
	fmt.Fprintf(out, "  Base URL:  %s\n", baseURL)
	if sonnet != "" {
		fmt.Fprintf(out, "  Sonnet:    %s\n", sonnet)
	}
	if opus != "" {
		fmt.Fprintf(out, "  Opus:      %s\n", opus)
	}
	if haiku != "" {
		fmt.Fprintf(out, "  Haiku:     %s\n", haiku)
	}
	if authToken != "" {
		fmt.Fprintf(out, "  Auth:      %s\n", config.MaskKey(authToken))
	}
	if timeoutMs != 0 {
		fmt.Fprintf(out, "  Timeout:   %d ms\n", timeoutMs)
	}
	fmt.Fprintln(out)

	scanner := bufio.NewScanner(cmd.InOrStdin())
	name, err := promptField(scanner, out, "Provider name", "", false)
	if err != nil {
		return err
	}

	if err := config.ValidateProviderName(name); err != nil {
		return err
	}

	if _, exists := cfg.Providers[name]; exists {
		return fmt.Errorf("Provider %s already exists. Use `ccswap edit %s` to modify.", name, name)
	}

	provider := config.Provider{
		AuthToken: authToken,
		BaseURL:   baseURL,
		TimeoutMs: timeoutMs,
		Models: config.Models{
			Sonnet: sonnet,
			Opus:   opus,
			Haiku:  haiku,
		},
	}

	// Import does not call ValidateProvider because settings.json may legitimately
	// have partial config (e.g. base_url with no model overrides). The base_url is
	// guaranteed non-empty since we check for it above. Warn about missing fields.
	if authToken == "" {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(out, "%s Warning: no auth token detected. `ccswap use` will fail without one.\n", yellow("⚠"))
	}
	if sonnet == "" || opus == "" || haiku == "" {
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Fprintf(out, "%s Warning: some models are missing. `ccswap use` will not override them.\n", yellow("⚠"))
	}

	cfg.Providers[name] = provider

	if err := config.SaveProviders(providersPath, cfg); err != nil {
		return fmt.Errorf("saving providers: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(out, "%s Provider %s imported. Run: %s\n", green("✓"), name, green("ccswap use "+name))

	return nil
}