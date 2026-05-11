package cmd

import (
	"fmt"
	"os"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/output"
	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [pipeline-file]",
	Short: "Execute a pipeline",
	Long:  `Run a pipeline defined in a YAML file, executing steps in dependency order.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPipeline,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Bool("dry-run", false, "Show execution plan without running")
	runCmd.Flags().Bool("verbose", false, "Show step output in real-time")
}

func runPipeline(cmd *cobra.Command, args []string) error {
	pipelineFile := args[0]
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Load pipeline
	p, err := pipeline.LoadPipeline(pipelineFile)
	if err != nil {
		return fmt.Errorf("loading pipeline: %w", err)
	}

	// Validate first
	if errs := p.Validate(); len(errs) > 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Pipeline validation failed:")
		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e.Error())
		}
		return fmt.Errorf("pipeline has %d validation error(s)", len(errs))
	}

	// Show execution plan
	levels, err := pipeline.TopologicalSort(p.Steps)
	if err != nil {
		return fmt.Errorf("computing execution order: %w", err)
	}

	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Execution plan for %q:\n\n", p.Name)
		for i, level := range levels {
			fmt.Fprintf(cmd.OutOrStdout(), "Level %d (parallel):\n", i+1)
			for _, step := range level {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s: %s\n", step.Name, step.Command)
			}
		}
		return nil
	}

	// Get run directory from config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Ensure run directory exists
	if err := os.MkdirAll(cfg.RunDir, 0755); err != nil {
		return fmt.Errorf("creating run directory: %w", err)
	}

	// Execute
	executor := pipeline.NewExecutor(p, cfg.RunDir)
	fmt.Fprintf(cmd.OutOrStdout(), "Running pipeline %q (run: %s)...\n", p.Name, executor.Run.RunID)

	execErr := executor.Execute(cmd.Context())

	// Display results using output formatter
	// NOTE: duplicated formatting logic (intentional rough edge — same pattern in status cmd)
	formatter := output.NewFormatter(outputFormat, cmd.OutOrStdout())
	headers := []string{"STEP", "STATUS", "DURATION", "EXIT CODE"}
	var rows []output.TableRow
	for _, s := range executor.Run.Steps {
		rows = append(rows, output.TableRow{
			Columns: []string{
				s.Name,
				s.Status,
				s.Duration.String(),
				fmt.Sprintf("%d", s.ExitCode),
			},
		})
	}

	fmt.Fprintln(cmd.OutOrStdout())
	if err := formatter.Render(headers, rows, executor.Run); err != nil {
		return fmt.Errorf("rendering output: %w", err)
	}

	if execErr != nil {
		return fmt.Errorf("pipeline failed: %w", execErr)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nPipeline %q completed successfully.\n", p.Name)
	return nil
}
