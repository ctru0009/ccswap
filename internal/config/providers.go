package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Provider struct {
	AuthToken string `yaml:"auth_token"`
	BaseURL   string `yaml:"base_url"`
	TimeoutMs int    `yaml:"timeout_ms"`
	Models    Models `yaml:"models"`
}

type Models struct {
	Sonnet string `yaml:"sonnet"`
	Opus   string `yaml:"opus"`
	Haiku  string `yaml:"haiku"`
}

type ProvidersConfig struct {
	Providers map[string]Provider `yaml:"providers"`
}

var providerNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$`)

func LoadProviders(path string) (*ProvidersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("providers file %s not found. Run `ccswap init` to create one", path)
		}
		return nil, fmt.Errorf("reading providers file %s: %w", path, err)
	}

	var cfg ProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing providers file %s: %w", path, err)
	}

	if cfg.Providers == nil {
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing providers file %s: %w", path, err)
		}
		if _, ok := raw["providers"]; !ok {
			return nil, fmt.Errorf("providers file %s has no 'providers' key. Add a 'providers:' block or run `ccswap init` to regenerate", path)
		}
		cfg.Providers = make(map[string]Provider)
	}

	return &cfg, nil
}

func SaveProviders(path string, config *ProvidersConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal providers: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "providers-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

func ValidateProviderName(name string) error {
	if len(name) > 32 {
		return fmt.Errorf("provider name %q exceeds 32 characters (len=%d)", name, len(name))
	}

	if !providerNameRe.MatchString(name) {
		return fmt.Errorf("provider name %q must match ^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$", name)
	}

	return nil
}

func ValidateProvider(p Provider) error {
	if p.AuthToken == "" {
		return fmt.Errorf("auth_token is required")
	}

	u, err := url.Parse(p.BaseURL)
	if err != nil {
		return fmt.Errorf("base_url parse error: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("base_url scheme must be http or https, got %q", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("base_url must have a host")
	}

	if p.Models.Sonnet == "" {
		return fmt.Errorf("models.sonnet is required")
	}

	if p.Models.Opus == "" {
		return fmt.Errorf("models.opus is required")
	}

	if p.Models.Haiku == "" {
		return fmt.Errorf("models.haiku is required")
	}

	return nil
}

func ExpandProvider(p Provider) (Provider, error) {
	p.AuthToken = os.ExpandEnv(p.AuthToken)
	p.BaseURL = os.ExpandEnv(p.BaseURL)
	p.Models.Sonnet = os.ExpandEnv(p.Models.Sonnet)
	p.Models.Opus = os.ExpandEnv(p.Models.Opus)
	p.Models.Haiku = os.ExpandEnv(p.Models.Haiku)

	if p.AuthToken == "" {
		return Provider{}, fmt.Errorf("auth_token is empty after env expansion")
	}

	return p, nil
}