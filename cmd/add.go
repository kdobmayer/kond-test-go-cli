package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Long:  "Add a new task with the given title. The title is all remaining arguments joined by spaces.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		task := store.Add(title)
		if err := store.Save(); err != nil {
			return fmt.Errorf("saving tasks: %w", err)
		}
		invalidateListCache()
		fmt.Printf("Added task %d: %s\n", task.ID, task.Title)
		return nil
	},
}
