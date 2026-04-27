package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newAddRootCmd() *cobra.Command {
	root := newRootCmd()
	root.AddCommand(addCmd)
	return root
}

func buildAddInput(name, token, baseURL, sonnet, opus, haiku, timeout string) string {
	return strings.Join([]string{name, token, baseURL, sonnet, opus, haiku, timeout}, "\n") + "\n"
}

func TestAdd_NewProvider(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("ollama-cloud", "ollama-token", "https://ollama.com/v1", "kimi-k2.6", "deepseek-v4", "glm-4.7", "3000000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain checkmark, got: %q", output)
	}
	if !strings.Contains(output, "ollama-cloud") {
		t.Errorf("output should contain provider name, got: %q", output)
	}
	if !strings.Contains(output, "ccswap use ollama-cloud") {
		t.Errorf("output should contain next-step instruction, got: %q", output)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "ollama-cloud") {
		t.Errorf("providers.yaml should contain provider name, got: %s", content)
	}
	if !strings.Contains(content, "ollama-token") {
		t.Errorf("providers.yaml should contain auth_token, got: %s", content)
	}
	if !strings.Contains(content, "https://ollama.com/v1") {
		t.Errorf("providers.yaml should contain base_url, got: %s", content)
	}
	if !strings.Contains(content, "kimi-k2.6") {
		t.Errorf("providers.yaml should contain sonnet model, got: %s", content)
	}
}

func TestAdd_DuplicateNameRejected(t *testing.T) {
	configDir := t.TempDir()
	existingYAML := `providers:
  ollama-cloud:
    auth_token: "existing-token"
    base_url: "https://existing.com"
    timeout_ms: 300000
    models:
      sonnet: "existing-sonnet"
      opus: "existing-opus"
      haiku: "existing-haiku"
`
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte(existingYAML), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("ollama-cloud", "new-token", "https://new.com/v1", "new-sonnet", "new-opus", "new-haiku", "3000000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for duplicate provider name, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
	if !strings.Contains(err.Error(), "ccswap edit") {
		t.Errorf("error should mention 'ccswap edit', got: %v", err)
	}
}

func TestAdd_InvalidNameRejected(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("INVALID-NAME", "some-token", "https://example.com/v1", "sonnet", "opus", "haiku", "3000000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid provider name, got nil")
	}
	if !strings.Contains(err.Error(), "must match") {
		t.Errorf("error should mention validation pattern, got: %v", err)
	}
}

func TestAdd_EmptyModelRejected(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("myprovider", "token", "https://example.com/v1", "", "opus-model", "haiku-model", "3000000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for empty model, got nil")
	}
	if !strings.Contains(err.Error(), "sonnet") {
		t.Errorf("error should mention sonnet, got: %v", err)
	}
}

func TestAdd_InvalidBaseURL(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("myprovider", "token", "ftp://bad.com", "sonnet-model", "opus-model", "haiku-model", "3000000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid base_url scheme, got nil")
	}
	if !strings.Contains(err.Error(), "http") || !strings.Contains(err.Error(), "https") {
		t.Errorf("error should mention http/https scheme, got: %v", err)
	}
}

func TestAdd_TimeoutDefault(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := buildAddInput("myprovider", "token", "https://example.com/v1", "sonnet-model", "opus-model", "haiku-model", "")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "3000000") {
		t.Errorf("providers.yaml should have timeout_ms 3000000 (default), got: %s", content)
	}
}

func TestAdd_EOFHandling(t *testing.T) {
	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "providers.yaml"), []byte("providers: {}\n"), 0644); err != nil {
		t.Fatalf("failed to write providers.yaml: %v", err)
	}

	input := "myprovider\n"
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err == nil {
		t.Fatal("expected error for EOF during input, got nil")
	}
}

func TestAdd_NoProvidersFile(t *testing.T) {
	configDir := t.TempDir()
	input := buildAddInput("newprov", "token123", "https://api.test.com/v1", "sonnet-x", "opus-y", "haiku-z", "5000")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	root := newAddRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--config", configDir, "add"})
	root.SetIn(r)

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	err = root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configDir, "providers.yaml"))
	if err != nil {
		t.Fatalf("failed to read providers.yaml: %v", err)
	}
	if !strings.Contains(string(data), "newprov") {
		t.Errorf("providers.yaml should contain new provider, got: %s", string(data))
	}
}