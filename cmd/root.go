package cmd

import (
	"github.com/spf13/cobra"
)

var (
	Version      = "dev"
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "A CLI tool for managing and executing pipelines",
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

func Execute() error {
	rootCmd.Version = Version
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	return rootCmd.Execute()
}
