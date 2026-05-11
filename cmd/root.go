package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	fooFlag      bool
)

var rootCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "A CLI tool for managing and executing pipelines",
	Long: `Pipeline is a CLI tool that allows you to define, execute, and monitor
multi-step pipelines with dependency management and parallel execution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fooFlag {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "foo"); err != nil {
				return fmt.Errorf("writing foo output: %w", err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&fooFlag, "foo", false, "Print foo")
}

func Execute() error {
	return rootCmd.Execute()
}
