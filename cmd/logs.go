package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var logsCmd = &cobra.Command{
	Use:   "logs [run-id] [step-name]",
	Short: "Show step logs from a pipeline run",
	Long:  `Display stdout/stderr output from pipeline steps.`,
	Args:  cobra.RangeArgs(0, 2),
	RunE:  runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().Bool("stderr", false, "Show only stderr output")
	logsCmd.Flags().Bool("stdout", false, "Show only stdout output")
	logsCmd.Flags().Int("limit", 0, "Limit output to last N lines per stream (0 = unlimited)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	showStderr, _ := cmd.Flags().GetBool("stderr")
	showStdout, _ := cmd.Flags().GetBool("stdout")
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return fmt.Errorf("invalid --limit %d: must be >= 0", limit)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Determine run ID
	var runID string
	if len(args) > 0 {
		runID = args[0]
	} else {
		runs, err := pipeline.ListRuns(cfg.RunDir)
		if err != nil {
			return fmt.Errorf("listing runs: %w", err)
		}
		if len(runs) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No pipeline runs found.")
			return nil
		}
		sort.Strings(runs)
		runID = runs[len(runs)-1]
	}

	// Load run to get step names
	run, err := pipeline.LoadRun(cfg.RunDir, runID)
	if err != nil {
		return fmt.Errorf("loading run: %w", err)
	}

	// If step name specified, show just that step
	if len(args) > 1 {
		stepName := args[1]
		log, err := pipeline.LoadStepLog(cfg.RunDir, runID, stepName)
		if err != nil {
			return fmt.Errorf("loading step log: %w", err)
		}
		return renderStepLog(cmd, log, showStdout, showStderr, limit)
	}

	// Show all step logs
	return renderAllLogs(cmd, cfg.RunDir, runID, run, showStdout, showStderr, limit)
}

func renderStepLog(cmd *cobra.Command, log *pipeline.StepLog, showStdout, showStderr bool, limit int) error {
	// NOTE: duplicated output formatting (intentional rough edge — same pattern in status cmd)
	switch outputFormat {
	case "json":
		data := buildLogOutput(log, showStdout, showStderr, limit)
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml":
		data := buildLogOutput(log, showStdout, showStderr, limit)
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(data)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "=== Step: %s ===\n", log.StepName)
		if !showStderr {
			fmt.Fprintf(cmd.OutOrStdout(), "\n--- stdout ---\n%s", limitLines(log.Stdout, limit))
		}
		if !showStdout {
			fmt.Fprintf(cmd.OutOrStdout(), "\n--- stderr ---\n%s", limitLines(log.Stderr, limit))
		}
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}
}

func renderAllLogs(cmd *cobra.Command, runDir, runID string, run *pipeline.PipelineRun, showStdout, showStderr bool, limit int) error {
	// NOTE: duplicated output formatting (intentional rough edge)
	switch outputFormat {
	case "json":
		var logs []interface{}
		for _, s := range run.Steps {
			log, err := pipeline.LoadStepLog(runDir, runID, s.Name)
			if err != nil {
				continue
			}
			logs = append(logs, buildLogOutput(log, showStdout, showStderr, limit))
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(logs)
	case "yaml":
		var logs []interface{}
		for _, s := range run.Steps {
			log, err := pipeline.LoadStepLog(runDir, runID, s.Name)
			if err != nil {
				continue
			}
			logs = append(logs, buildLogOutput(log, showStdout, showStderr, limit))
		}
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(logs)
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STEP\tSTDOUT (bytes)\tSTDERR (bytes)")
		fmt.Fprintln(w, "----\t--------------\t--------------")
		for _, s := range run.Steps {
			log, err := pipeline.LoadStepLog(runDir, runID, s.Name)
			if err != nil {
				fmt.Fprintf(w, "%s\t-\t-\n", s.Name)
			} else {
				fmt.Fprintf(w, "%s\t%d\t%d\n", s.Name, len(log.Stdout), len(log.Stderr))
			}
		}
		return w.Flush()
	}
}

func buildLogOutput(log *pipeline.StepLog, showStdout, showStderr bool, limit int) map[string]string {
	data := map[string]string{"step_name": log.StepName}
	if !showStderr {
		data["stdout"] = limitLines(log.Stdout, limit)
	}
	if !showStdout {
		data["stderr"] = limitLines(log.Stderr, limit)
	}
	return data
}

// limitLines returns the last n lines of s. Returns s unchanged when n <= 0.
func limitLines(s string, n int) string {
	if n <= 0 || s == "" {
		return s
	}
	lines := strings.SplitAfter(s, "\n")
	// SplitAfter on a trailing newline produces a trailing empty element; exclude it from the count.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "")
}
