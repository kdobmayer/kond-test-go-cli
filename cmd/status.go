package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/output"
	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status [run-id]",
	Short: "Show pipeline run status",
	Long:  `Display the status of a pipeline run. If no run-id is given, shows the latest run.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().Bool("all", false, "Show all runs")
	statusCmd.Flags().Bool("json", false, "Output as JSON (shorthand for --output=json)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	showAll, _ := cmd.Flags().GetBool("all")
	jsonFlag, _ := cmd.Flags().GetBool("json")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	effectiveFormat := outputFormat
	if jsonFlag {
		effectiveFormat = "json"
	}

	if showAll {
		return showAllRuns(cmd, cfg.RunDir, effectiveFormat)
	}

	var runID string
	if len(args) > 0 {
		runID = args[0]
	} else {
		// Find latest run
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

	run, err := pipeline.LoadRun(cfg.RunDir, runID)
	if err != nil {
		return fmt.Errorf("loading run: %w", err)
	}

	return renderRunStatus(cmd, run)
}

func showAllRuns(cmd *cobra.Command, runDir, format string) error {
	runs, err := pipeline.ListRuns(runDir)
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}
	if len(runs) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No pipeline runs found.")
		return nil
	}
	sort.Strings(runs)

	type runSummary struct {
		RunID  string `json:"run_id" yaml:"run_id"`
		Status string `json:"status" yaml:"status"`
	}
	var summaries []runSummary
	for _, id := range runs {
		run, err := pipeline.LoadRun(runDir, id)
		if err != nil {
			continue
		}
		summaries = append(summaries, runSummary{RunID: id, Status: run.Status})
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	headers := []string{"RUN ID", "STATUS"}
	var rows []output.TableRow
	for _, s := range summaries {
		rows = append(rows, output.TableRow{Columns: []string{s.RunID, s.Status}})
	}
	return formatter.Render(headers, rows, summaries)
}

func renderRunStatus(cmd *cobra.Command, run *pipeline.PipelineRun) error {
	// NOTE: duplicated output formatting (intentional rough edge)
	switch outputFormat {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(run)
	case "yaml":
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(run)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline: %s\n", run.PipelineName)
		fmt.Fprintf(cmd.OutOrStdout(), "Run ID:   %s\n", run.RunID)
		fmt.Fprintf(cmd.OutOrStdout(), "Status:   %s\n", run.Status)
		fmt.Fprintf(cmd.OutOrStdout(), "Started:  %s\n", run.StartTime.Format("2006-01-02 15:04:05"))
		if !run.EndTime.IsZero() {
			fmt.Fprintf(cmd.OutOrStdout(), "Ended:    %s\n", run.EndTime.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(cmd.OutOrStdout(), "Duration: %s\n", run.EndTime.Sub(run.StartTime))
		}
		fmt.Fprintln(cmd.OutOrStdout())

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STEP\tSTATUS\tDURATION\tEXIT CODE\tERROR")
		fmt.Fprintln(w, "----\t------\t--------\t---------\t-----")
		for _, s := range run.Steps {
			errStr := s.Error
			if len(errStr) > 50 {
				errStr = errStr[:50] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				s.Name, s.Status, s.Duration, s.ExitCode, errStr)
		}
		return w.Flush()
	}
}
