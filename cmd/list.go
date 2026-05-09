package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var jsonOutput bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List all tasks with their ID, status, and title.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks := store.List()
		if jsonOutput {
			data, err := json.MarshalIndent(tasks, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling tasks: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}
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

func init() {
	listCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output tasks as a JSON array")
}
