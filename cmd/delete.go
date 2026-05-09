package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a task",
	Long:  "Delete the task with the given ID permanently.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID %q: %w", args[0], err)
		}
		if err := store.Delete(id); err != nil {
			return err
		}
		if err := store.Save(); err != nil {
			return fmt.Errorf("saving tasks: %w", err)
		}
		invalidateListCache()
		fmt.Printf("Task %d deleted.\n", id)
		return nil
	},
}
