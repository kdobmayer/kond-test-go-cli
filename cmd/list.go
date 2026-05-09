package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

var listOutputFormat string

type jsonListResponse struct {
	Result []internal.Task `json:"result"`
	Error  *string         `json:"error"`
}

func writeListJSON(w io.Writer, tasks []internal.Task, err error) error {
	resp := jsonListResponse{}
	if err != nil {
		msg := err.Error()
		resp.Error = &msg
	} else {
		resp.Result = tasks
	}
	data, marshalErr := json.MarshalIndent(resp, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("marshaling list response: %w", marshalErr)
	}
	_, writeErr := fmt.Fprintln(w, string(data))
	return writeErr
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List all tasks with their ID, status, and title.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tasks := store.List()
		if listOutputFormat == "json" {
			return writeListJSON(cmd.OutOrStdout(), tasks, nil)
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
	listCmd.Flags().StringVar(&listOutputFormat, "output", "", `output format; use "json" for machine-readable output`)
}
