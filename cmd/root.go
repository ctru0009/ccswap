package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// exitFunc allows overriding os.Exit in tests.
var exitFunc = os.Exit

var rootCmd = &cobra.Command{
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

func init() {
	rootCmd.PersistentFlags().String("config", "", "path to config directory")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(listCmd)
}

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Error: panic:", r)
			exitFunc(2)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		exitFunc(1)
	}
}
