package cmd

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat string
)

const defaultVersion = "dev"

var rootCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "A CLI tool for managing and executing pipelines",
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
	Version: defaultVersion,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
}

func SetVersion(v string) {
	if v == "" {
		rootCmd.Version = defaultVersion
		return
	}

	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}
