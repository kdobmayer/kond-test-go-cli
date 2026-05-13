package cmd

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	Version      = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "pipeline",
	Short:   "A CLI tool for managing and executing pipelines",
	Version: Version,
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
}

func Execute() error {
	// Keep Cobra's version output aligned with the current package variable so
	// runtime overrides and tests don't use the stale init-time value.
	rootCmd.Version = Version
	return rootCmd.Execute()
}
