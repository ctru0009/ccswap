package cmd

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteProviderNames(t *testing.T) {
	configDir := t.TempDir()
	writeTestProviders(t, configDir, validProvidersYAML)

	origProvidersPath := providersPathFunc
	providersPathFunc = func(string) string { return filepath.Join(configDir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	cmd := &cobra.Command{}
	got, directive := completeProviderNames(cmd, nil, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}

	want := []string{"anthropic", "zai"}
	if !equalStrings(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestCompleteProviderNames_ArgsAlreadyPresent(t *testing.T) {
	got, directive := completeProviderNames(nil, []string{"anthropic"}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	if len(got) != 0 {
		t.Errorf("expected no completions with args present, got %v", got)
	}
}

func TestCompleteProviderNames_NoProvidersFile(t *testing.T) {
	configDir := t.TempDir()

	origProvidersPath := providersPathFunc
	providersPathFunc = func(string) string { return filepath.Join(configDir, "providers.yaml") }
	t.Cleanup(func() { providersPathFunc = origProvidersPath })

	cmd := &cobra.Command{}
	got, directive := completeProviderNames(cmd, nil, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
	}
	if len(got) != 0 {
		t.Errorf("expected no completions for missing file, got %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sorted := make([]string, len(a))
	copy(sorted, a)
	sort.Strings(sorted)
	sortedB := make([]string, len(b))
	copy(sortedB, b)
	sort.Strings(sortedB)
	for i := range sorted {
		if sorted[i] != sortedB[i] {
			return false
		}
	}
	return true
}
