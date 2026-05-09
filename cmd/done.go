package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done [id]",
	Short: "Mark a task as done",
	Long:  "Mark the task with the given ID as completed.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID %q: %w", args[0], err)
		}
		if err := store.MarkDone(id); err != nil {
			return err
		}
		if err := store.Save(); err != nil {
			return fmt.Errorf("saving tasks: %w", err)
		}
		invalidateListCache()
		fmt.Printf("Task %d marked as done.\n", id)
		return nil
	},
}
