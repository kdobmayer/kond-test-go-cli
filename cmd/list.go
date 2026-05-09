package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

var listFormat string

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
		switch listFormat {
		case "table":
			printTable(tasks)
			return nil
		case "csv":
			return printCSV(tasks)
		default:
			return fmt.Errorf("unknown format %q: must be table or csv", listFormat)
		}
	},
}

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table or csv")
}

func printTable(tasks []internal.Task) {
	for _, t := range tasks {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d %s %s\n", t.ID, status, t.Title)
	}
}

func printCSV(tasks []internal.Task) error {
	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"id", "done", "title"})
	for _, t := range tasks {
		w.Write([]string{strconv.Itoa(t.ID), strconv.FormatBool(t.Done), t.Title})
	}
	w.Flush()
	return w.Error()
}
