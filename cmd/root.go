package cmd

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		foo, err := cmd.Flags().GetBool("foo")
		if err != nil {
			return fmt.Errorf("getting foo flag: %w", err)
		}

		bar, err := cmd.Flags().GetBool("bar")
		if err != nil {
			return fmt.Errorf("getting bar flag: %w", err)
		}

		if foo {
			fmt.Fprintln(cmd.OutOrStdout(), "foo")
		}
		if bar {
			fmt.Fprintln(cmd.OutOrStdout(), "bar")
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, yaml")
	rootCmd.Flags().Bool("foo", false, "Print foo and exit")
	rootCmd.Flags().Bool("bar", false, "Print bar and exit")
}

func Execute() error {
	return rootCmd.Execute()
}
