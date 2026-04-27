package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newRootCmd returns a fresh rootCmd for test isolation.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ccswap",
		Short: "ccswap is a CLI for switching Claude Code providers",
		Long: `ccswap switches Claude Code provider profiles by editing ` +
			`~/.claude/settings.json safely.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.PersistentFlags().String("config", "", "path to config directory")
	return cmd
}

func TestRootCommand_Help(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ccswap") {
		t.Errorf("help output should contain 'ccswap', got: %s", output)
	}
}

func TestRootCommand_UnknownCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %v", err)
	}
}

func TestExecute_ErrorExit(t *testing.T) {
	var exitCode int
	origExit := exitFunc
	exitFunc = func(code int) {
		exitCode = code
	}
	t.Cleanup(func() { exitFunc = origExit })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	rootCmd.SetArgs([]string{"--unknown-flag"})

	Execute()

	w.Close()
	stderrBytes, _ := io.ReadAll(r)
	r.Close()
	stderrOutput := string(stderrBytes)

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderrOutput, "Error:") {
		t.Errorf("expected error on stderr, got: %q", stderrOutput)
	}
}

func TestRootCommand_ConfigFlag(t *testing.T) {
	cmd := newRootCmd()
	err := cmd.PersistentFlags().Parse([]string{"--config", "/custom/path"})
	if err != nil {
		t.Fatalf("failed to parse --config flag: %v", err)
	}

	config, err := cmd.PersistentFlags().GetString("config")
	if err != nil {
		t.Fatalf("--config flag not found: %v", err)
	}
	if config != "/custom/path" {
		t.Errorf("expected '/custom/path', got %q", config)
	}
}
