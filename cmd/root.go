// KOND reconcile test
// Package cmd implements the CLI commands for the TODO task manager.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

var store *internal.Store

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "A simple TODO task manager",
	Long:  "Manage your TODO tasks from the command line. Tasks are persisted to ~/.tasks.json.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		store, err = internal.NewStore()
		return err
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)
}
