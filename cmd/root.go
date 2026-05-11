package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	showVersion  bool

	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "A CLI tool for managing and executing pipelines",
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVersion {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "version: %s\ncommit: %s\nbuilt: %s\n", version, commit, buildDate); err != nil {
				return fmt.Errorf("writing version output: %w", err)
			}
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Print version information")
}

func Execute() error {
	return rootCmd.Execute()
}
