package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- TestLoadProviders_ValidYAML ---

func TestLoadProviders_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	content := `providers:
  zai:
    auth_token: "secret-token"
    base_url: "https://api.example.com"
    timeout_ms: 5000
    models:
      sonnet: "claude-3-5-sonnet"
      opus: "claude-3-opus"
      haiku: "claude-3-haiku"
  ollama-cloud:
    auth_token: "another-token"
    base_url: "https://ollama.example.com"
    timeout_ms: 10000
    models:
      sonnet: "llama3-sonnet"
      opus: "llama3-opus"
      haiku: "llama3-haiku"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadProviders(path)
	if err != nil {
		t.Fatalf("LoadProviders returned error: %v", err)
	}

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}

	p, ok := cfg.Providers["zai"]
	if !ok {
		t.Fatal("expected provider 'zai' to exist")
	}
	if p.AuthToken != "secret-token" {
		t.Errorf("expected auth_token 'secret-token', got %q", p.AuthToken)
	}
	if p.BaseURL != "https://api.example.com" {
		t.Errorf("expected base_url 'https://api.example.com', got %q", p.BaseURL)
	}
	if p.TimeoutMs != 5000 {
		t.Errorf("expected timeout_ms 5000, got %d", p.TimeoutMs)
	}
	if p.Models.Sonnet != "claude-3-5-sonnet" {
		t.Errorf("expected sonnet 'claude-3-5-sonnet', got %q", p.Models.Sonnet)
	}
	if p.Models.Opus != "claude-3-opus" {
		t.Errorf("expected opus 'claude-3-opus', got %q", p.Models.Opus)
	}
	if p.Models.Haiku != "claude-3-haiku" {
		t.Errorf("expected haiku 'claude-3-haiku', got %q", p.Models.Haiku)
	}
}

// --- TestLoadProviders_MissingFile ---

func TestLoadProviders_MissingFile(t *testing.T) {
	_, err := LoadProviders("/nonexistent/path/providers.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- TestLoadProviders_InvalidYAML ---

func TestLoadProviders_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	content := `providers:
  zai:
    auth_token: [invalid yaml
    base_url: "https://api.example.com"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadProviders(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// --- TestSaveProviders (round-trip) ---

func TestSaveProviders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	cfg := &ProvidersConfig{
		Providers: map[string]Provider{
			"zai": {
				AuthToken: "round-trip-token",
				BaseURL:   "https://api.roundtrip.com",
				TimeoutMs: 3000,
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "model-opus",
					Haiku:  "model-haiku",
				},
			},
		},
	}

	// Save
	if err := SaveProviders(path, cfg); err != nil {
		t.Fatalf("SaveProviders returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after SaveProviders")
	}

	// Load it back (round-trip)
	loaded, err := LoadProviders(path)
	if err != nil {
		t.Fatalf("LoadProviders after save returned error: %v", err)
	}

	if len(loaded.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(loaded.Providers))
	}

	p := loaded.Providers["zai"]
	if p.AuthToken != "round-trip-token" {
		t.Errorf("expected auth_token 'round-trip-token', got %q", p.AuthToken)
	}
	if p.BaseURL != "https://api.roundtrip.com" {
		t.Errorf("expected base_url 'https://api.roundtrip.com', got %q", p.BaseURL)
	}
	if p.TimeoutMs != 3000 {
		t.Errorf("expected timeout_ms 3000, got %d", p.TimeoutMs)
	}
	if p.Models.Sonnet != "model-sonnet" {
		t.Errorf("expected sonnet 'model-sonnet', got %q", p.Models.Sonnet)
	}
	if p.Models.Opus != "model-opus" {
		t.Errorf("expected opus 'model-opus', got %q", p.Models.Opus)
	}
	if p.Models.Haiku != "model-haiku" {
		t.Errorf("expected haiku 'model-haiku', got %q", p.Models.Haiku)
	}
}

// --- TestValidateProviderName ---

func TestValidateProviderName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid short", "zai", false},
		{"valid with hyphen", "ollama-cloud", false},
		{"valid alphanumeric", "abc123", false},
		{"valid boundary 32 chars", "abcdefghijklmnopqrstuvwxyz123456", false},
		{"invalid uppercase", "ZAI", true},
		{"invalid space", "ollama cloud", true},
		{"invalid too long 33 chars", "abcdefghijklmnopqrstuvwxyz1234567", true},
		{"invalid starts with hyphen", "-start", true},
		{"invalid ends with hyphen", "end-", true},
		{"invalid empty", "", true},
		{"invalid special char", "pro@vider", true},
		{"valid single char", "a", false},
		{"valid all digits", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviderName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// --- TestValidateProvider_RequiredFields ---

func TestValidateProvider_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		provider Provider
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid provider",
			provider: Provider{
				AuthToken: "token",
				BaseURL:   "https://api.example.com",
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "model-opus",
					Haiku:  "model-haiku",
				},
			},
			wantErr: false,
		},
		{
			name: "empty auth_token",
			provider: Provider{
				AuthToken: "",
				BaseURL:   "https://api.example.com",
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "model-opus",
					Haiku:  "model-haiku",
				},
			},
			wantErr: true,
			errMsg:  "auth_token",
		},
		{
			name: "empty sonnet model",
			provider: Provider{
				AuthToken: "token",
				BaseURL:   "https://api.example.com",
				Models: Models{
					Sonnet: "",
					Opus:  "model-opus",
					Haiku: "model-haiku",
				},
			},
			wantErr: true,
			errMsg:  "sonnet",
		},
		{
			name: "empty opus model",
			provider: Provider{
				AuthToken: "token",
				BaseURL:   "https://api.example.com",
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "",
					Haiku:  "model-haiku",
				},
			},
			wantErr: true,
			errMsg:  "opus",
		},
		{
			name: "empty haiku model",
			provider: Provider{
				AuthToken: "token",
				BaseURL:   "https://api.example.com",
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "model-opus",
					Haiku:  "",
				},
			},
			wantErr: true,
			errMsg:  "haiku",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProvider(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// --- TestValidateProvider_URLValidation ---

func TestValidateProvider_URLValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{"valid https", "https://api.example.com", false},
		{"valid http", "http://localhost:8080", false},
		{"invalid ftp scheme", "ftp://files.example.com", true},
		{"invalid no scheme", "api.example.com", true},
		{"invalid empty url", "", true},
		{"invalid scheme only", "https://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Provider{
				AuthToken: "token",
				BaseURL:   tt.baseURL,
				Models: Models{
					Sonnet: "model-sonnet",
					Opus:   "model-opus",
					Haiku:  "model-haiku",
				},
			}
			err := ValidateProvider(p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProvider() with baseURL=%q error = %v, wantErr %v", tt.baseURL, err, tt.wantErr)
			}
		})
	}
}

// --- TestEnvVarExpansion ---

func TestEnvVarExpansion(t *testing.T) {
	os.Setenv("CCSWAP_TEST_TOKEN", "expanded-token")
	os.Setenv("CCSWAP_TEST_URL", "https://expanded.example.com")
	os.Setenv("CCSWAP_TEST_SONNET", "expanded-sonnet")
	os.Setenv("CCSWAP_TEST_OPUS", "expanded-opus")
	os.Setenv("CCSWAP_TEST_HAIKU", "expanded-haiku")
	defer func() {
		os.Unsetenv("CCSWAP_TEST_TOKEN")
		os.Unsetenv("CCSWAP_TEST_URL")
		os.Unsetenv("CCSWAP_TEST_SONNET")
		os.Unsetenv("CCSWAP_TEST_OPUS")
		os.Unsetenv("CCSWAP_TEST_HAIKU")
	}()

	p := Provider{
		AuthToken: "$CCSWAP_TEST_TOKEN",
		BaseURL:   "$CCSWAP_TEST_URL",
		Models: Models{
			Sonnet: "$CCSWAP_TEST_SONNET",
			Opus:   "$CCSWAP_TEST_OPUS",
			Haiku:  "$CCSWAP_TEST_HAIKU",
		},
	}

	expanded, err := ExpandProvider(p)
	if err != nil {
		t.Fatalf("ExpandProvider returned error: %v", err)
	}

	if expanded.AuthToken != "expanded-token" {
		t.Errorf("expected auth_token 'expanded-token', got %q", expanded.AuthToken)
	}
	if expanded.BaseURL != "https://expanded.example.com" {
		t.Errorf("expected base_url 'https://expanded.example.com', got %q", expanded.BaseURL)
	}
	if expanded.Models.Sonnet != "expanded-sonnet" {
		t.Errorf("expected sonnet 'expanded-sonnet', got %q", expanded.Models.Sonnet)
	}
	if expanded.Models.Opus != "expanded-opus" {
		t.Errorf("expected opus 'expanded-opus', got %q", expanded.Models.Opus)
	}
	if expanded.Models.Haiku != "expanded-haiku" {
		t.Errorf("expected haiku 'expanded-haiku', got %q", expanded.Models.Haiku)
	}
}

// --- TestEnvVarExpansion_MissingVar_WarnsAndFails ---

func TestEnvVarExpansion_MissingVar_WarnsAndFails(t *testing.T) {
	// Ensure the env var is NOT set
	os.Unsetenv("CCSWAP_NONEXISTENT_TOKEN")
	defer os.Unsetenv("CCSWAP_NONEXISTENT_TOKEN")

	p := Provider{
		AuthToken: "$CCSWAP_NONEXISTENT_TOKEN",
		BaseURL:   "https://api.example.com",
		Models: Models{
			Sonnet: "model-sonnet",
			Opus:   "model-opus",
			Haiku:  "model-haiku",
		},
	}

	_, err := ExpandProvider(p)
	if err == nil {
		t.Fatal("expected error when auth_token env var is missing, got nil")
	}
	if !contains(err.Error(), "auth_token") {
		t.Errorf("expected error to mention 'auth_token', got %q", err.Error())
	}
}

// --- helper ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLoadProviders_NoProvidersKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	content := "some_other_key: some_value\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadProviders(path)
	if err == nil {
		t.Fatal("expected error for YAML with no providers key, got nil")
	}
	if !contains(err.Error(), "no 'providers' key") {
		t.Errorf("error should mention 'no providers key', got: %v", err)
	}
	if !contains(err.Error(), "ccswap init") {
		t.Errorf("error should mention 'ccswap init', got: %v", err)
	}
}

func TestLoadProviders_EmptyProvidersKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	content := "providers: {}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadProviders(path)
	if err != nil {
		t.Fatalf("expected no error for empty providers map, got: %v", err)
	}
	if len(cfg.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(cfg.Providers))
	}
}

func TestLoadProviders_ProvidersWithOnlyComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "providers.yaml")

	content := "providers:\n  # anthropic:\n  #   auth_token: test\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadProviders(path)
	if err != nil {
		t.Fatalf("expected no error for providers with only comments, got: %v", err)
	}
	if len(cfg.Providers) != 0 {
		t.Errorf("expected 0 providers, got %d", len(cfg.Providers))
	}
}

func TestLoadProviders_MissingFile_ActionableMessage(t *testing.T) {
	_, err := LoadProviders("/nonexistent/path/providers.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !contains(err.Error(), "ccswap init") {
		t.Errorf("error should mention 'ccswap init', got: %v", err)
	}
}