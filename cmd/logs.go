package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
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
}

func runLogs(cmd *cobra.Command, args []string) error {
	showStderr, _ := cmd.Flags().GetBool("stderr")
	showStdout, _ := cmd.Flags().GetBool("stdout")

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
		return renderStepLog(cmd, log, showStdout, showStderr)
	}

	// Show all step logs
	return renderAllLogs(cmd, cfg.RunDir, runID, run, showStdout, showStderr)
}

func renderStepLog(cmd *cobra.Command, log *pipeline.StepLog, showStdout, showStderr bool) error {
	// NOTE: duplicated output formatting (intentional rough edge — same pattern in status cmd)
	switch currentOutputFormat(cmd) {
	case "json":
		data := buildLogOutput(log, showStdout, showStderr)
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "yaml":
		data := buildLogOutput(log, showStdout, showStderr)
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(data)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "=== Step: %s ===\n", log.StepName)
		if !showStderr {
			fmt.Fprintf(cmd.OutOrStdout(), "\n--- stdout ---\n%s", log.Stdout)
		}
		if !showStdout {
			fmt.Fprintf(cmd.OutOrStdout(), "\n--- stderr ---\n%s", log.Stderr)
		}
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}
}

func renderAllLogs(cmd *cobra.Command, runDir, runID string, run *pipeline.PipelineRun, showStdout, showStderr bool) error {
	// NOTE: duplicated output formatting (intentional rough edge)
	switch currentOutputFormat(cmd) {
	case "json":
		var logs []interface{}
		for _, s := range run.Steps {
			log, err := pipeline.LoadStepLog(runDir, runID, s.Name)
			if err != nil {
				continue
			}
			logs = append(logs, buildLogOutput(log, showStdout, showStderr))
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
			logs = append(logs, buildLogOutput(log, showStdout, showStderr))
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
				continue
			}
			fmt.Fprintf(w, "%s\t%d\t%d\n", s.Name, len(log.Stdout), len(log.Stderr))
		}
		return w.Flush()
	}
}

func buildLogOutput(log *pipeline.StepLog, showStdout, showStderr bool) map[string]string {
	data := map[string]string{"step_name": log.StepName}
	if !showStderr {
		data["stdout"] = log.Stdout
	}
	if !showStdout {
		data["stderr"] = log.Stderr
	}
	return data
}
