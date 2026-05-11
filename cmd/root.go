package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	jsonOutput   bool
)

var rootCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "A CLI tool for managing and executing pipelines",
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (shorthand for --output json)")
}

func Execute() error {
	return rootCmd.Execute()
}

func currentOutputFormat(cmd *cobra.Command) string {
	if cmd != nil {
		jsonFlag := cmd.Flags().Lookup("json")
		outputFlag := cmd.Flags().Lookup("output")
		if jsonFlag != nil && strings.EqualFold(jsonFlag.Value.String(), "true") &&
			(outputFlag == nil || !outputFlag.Changed) {
			return "json"
		}
	}

	return outputFormat
}
