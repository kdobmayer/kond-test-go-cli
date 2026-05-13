package cmd

import (
	"fmt"
	"time"

	"github.com/kdobmayer/kond-test-go-cli/pipeline"
)

func findLatestRunID(runDir string) (string, error) {
	runs, err := pipeline.ListRuns(runDir)
	if err != nil {
		return "", fmt.Errorf("listing runs: %w", err)
	}
	if len(runs) == 0 {
		return "", nil
	}

	var latestID string
	var latestTime time.Time
	for _, runID := range runs {
		run, err := pipeline.LoadRun(runDir, runID)
		if err != nil {
			return "", fmt.Errorf("loading run %q: %w", runID, err)
		}
		if latestID == "" || run.StartTime.After(latestTime) {
			latestID = runID
			latestTime = run.StartTime
		}
	}

	return latestID, nil
}
