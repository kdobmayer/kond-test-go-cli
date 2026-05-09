package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kdobmayer/kond-test-go-cli/internal"
)

const (
	cacheKey = "tasks:list"
	cacheTTL = 60 * time.Second
)

var useCache bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  "List all tasks with their ID, status, and title.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if useCache {
			if cache == nil {
				cache = internal.NewRedisCache("localhost:6379")
			}
			if data, err := cache.Get(ctx, cacheKey); err == nil {
				var tasks []internal.Task
				if json.Unmarshal(data, &tasks) == nil {
					printTasks(tasks)
					return nil
				}
			}
		}

		tasks := store.List()

		if useCache && cache != nil {
			if data, err := json.Marshal(tasks); err == nil {
				_ = cache.Set(ctx, cacheKey, data, cacheTTL)
			}
		}

		printTasks(tasks)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&useCache, "cache", false, "serve task list from Redis cache (localhost:6379, 60s TTL)")
}

func printTasks(tasks []internal.Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks.")
		return
	}
	for _, t := range tasks {
		status := "[ ]"
		if t.Done {
			status = "[x]"
		}
		fmt.Printf("%d %s %s\n", t.ID, status, t.Title)
	}
}
