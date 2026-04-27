package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ctru0009/ccswap/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new provider interactively",
	Long: `Interactively add a new provider profile to providers.yaml.
Prompts for provider name, auth token, base URL, models, and timeout.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(cmd)
	},
}

func runAdd(cmd *cobra.Command) error {
	configDir, err := resolveConfigDir(cmd)
	if err != nil {
		return err
	}

	providersPath := filepath.Join(configDir, "providers.yaml")

	cfg, err := loadOrInitProviders(providersPath)
	if err != nil {
		return fmt.Errorf("loading providers: %w", err)
	}

	in := cmd.InOrStdin()
	out := cmd.ErrOrStderr()
	scanner := bufio.NewScanner(in)

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

	var authToken string
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprint(out, "Auth token: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(out)
		if err != nil {
			return fmt.Errorf("reading auth token: %w", err)
		}
		authToken = string(bytePassword)
	} else {
		authToken, err = promptField(scanner, out, "Auth token", "", true)
		if err != nil {
			return err
		}
	}

	baseURL, err := promptField(scanner, out, "Base URL", "", false)
	if err != nil {
		return err
	}

	sonnet, err := promptField(scanner, out, "Model (sonnet)", "", false)
	if err != nil {
		return err
	}

	opus, err := promptField(scanner, out, "Model (opus)", "", false)
	if err != nil {
		return err
	}

	haiku, err := promptField(scanner, out, "Model (haiku)", "", false)
	if err != nil {
		return err
	}

	timeoutStr, err := promptField(scanner, out, "Timeout (ms)", "3000000", true)
	if err != nil {
		return err
	}

	timeoutMs := 3000000
	if timeoutStr != "" {
		timeoutMs, err = strconv.Atoi(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout value %q: must be a number", timeoutStr)
		}
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

	if err := config.ValidateProvider(provider); err != nil {
		return err
	}

	cfg.Providers[name] = provider

	if err := config.SaveProviders(providersPath, cfg); err != nil {
		return fmt.Errorf("saving providers: %w", err)
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(out, "%s Provider %s added. Run: %s\n", green("✓"), name, green("ccswap use "+name))

	return nil
}

func promptField(scanner *bufio.Scanner, out io.Writer, label, defaultVal string, allowDefault bool) (string, error) {
	if defaultVal != "" && allowDefault {
		fmt.Fprintf(out, "%s [%s]: ", label, defaultVal)
	} else {
		fmt.Fprintf(out, "%s: ", label)
	}

	if !scanner.Scan() {
		return "", fmt.Errorf("reading %s: unexpected EOF", strings.ToLower(label))
	}
	value := strings.TrimSpace(scanner.Text())

	if value == "" && allowDefault {
		return defaultVal, nil
	}

	return value, nil
}

func loadOrInitProviders(path string) (*config.ProvidersConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := &config.ProvidersConfig{
			Providers: make(map[string]config.Provider),
		}
		return cfg, nil
	}
	return config.LoadProviders(path)
}