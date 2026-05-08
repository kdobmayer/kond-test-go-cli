package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List all tasks with their ID, status, and title.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks := store.List()
		if len(tasks) == 0 {
			fmt.Println("No tasks.")
			return nil
		}
		for _, t := range tasks {
			status := "[ ]"
			if t.Done {
				status = "[x]"
			}
			fmt.Printf("%d %s %s\n", t.ID, status, t.Title)
		}
		return nil
	},
}
