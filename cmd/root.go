package cmd

import (
	"github.com/spf13/cobra"
)

var (
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Stream step output to stderr in real-time")
}

func Execute() error {
	return rootCmd.Execute()
}
