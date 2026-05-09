// Package cmd implements the CLI commands for the TODO task manager.
package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

var store *internal.Store
var cache internal.Cache

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

// invalidateListCache deletes the cached task list. Errors are silently ignored.
// If no cache is injected (cache == nil), it attempts a best-effort delete via a
// transient Redis client so that stale entries from prior --cache runs are cleared.
func invalidateListCache() {
	ctx := context.Background()
	if cache != nil {
		_ = cache.Delete(ctx, cacheKey)
		return
	}
	_ = internal.NewRedisCache("localhost:6379").Delete(ctx, cacheKey)
}
