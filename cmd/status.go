package cmd

import (
	"fmt"
	"sort"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/output"
	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
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
	statusCmd.Flags().Bool("json", false, "Output in JSON format")
}

func runStatus(cmd *cobra.Command, args []string) error {
	showAll, _ := cmd.Flags().GetBool("all")
	useJSON, _ := cmd.Flags().GetBool("json")

	format := outputFormat
	if useJSON {
		format = "json"
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if showAll {
		return showAllRuns(cmd, cfg.RunDir, format)
	}

	var runID string
	if len(args) > 0 {
		runID = args[0]
	} else {
		runs, err := pipeline.ListRuns(cfg.RunDir)
		if err != nil {
			return fmt.Errorf("listing runs: %w", err)
		}
		if len(runs) == 0 {
			return renderNoRuns(cmd, format, false)
		}
		sort.Strings(runs)
		runID = runs[len(runs)-1]
	}

	run, err := pipeline.LoadRun(cfg.RunDir, runID)
	if err != nil {
		return fmt.Errorf("loading run: %w", err)
	}

	return renderRunStatus(cmd, run, format)
}

func showAllRuns(cmd *cobra.Command, runDir string, format string) error {
	runs, err := pipeline.ListRuns(runDir)
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}
	if len(runs) == 0 {
		return renderNoRuns(cmd, format, true)
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

	headers := []string{"RUN ID", "STATUS"}
	var rows []output.TableRow
	for _, s := range summaries {
		rows = append(rows, output.TableRow{Columns: []string{s.RunID, s.Status}})
	}

	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.Render(headers, rows, summaries)
}

func renderNoRuns(cmd *cobra.Command, format string, all bool) error {
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	switch format {
	case "json", "yaml":
		if all {
			return formatter.Render(nil, nil, []any{})
		}
		return formatter.Render(nil, nil, map[string]string{"message": "No pipeline runs found."})
	default:
		formatter.RenderMessage("No pipeline runs found.")
		return nil
	}
}

func renderRunStatus(cmd *cobra.Command, run *pipeline.PipelineRun, format string) error {
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	headers := []string{"STEP", "STATUS", "DURATION", "EXIT CODE", "ERROR"}
	var rows []output.TableRow
	for _, s := range run.Steps {
		errStr := s.Error
		if len(errStr) > 50 {
			errStr = errStr[:50] + "..."
		}
		rows = append(rows, output.TableRow{Columns: []string{
			s.Name, s.Status, s.Duration.String(), fmt.Sprintf("%d", s.ExitCode), errStr,
		}})
	}

	switch format {
	case "json", "yaml":
		return formatter.Render(headers, rows, run)
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
		return formatter.Table(headers, rows)
	}
}
