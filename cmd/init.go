package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kdobmayer/kond-test-go-cli/config"
	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new pipeline configuration",
	Long:  `Initialize a new pipeline YAML file with a basic template.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("dir", "d", "", "Directory to create pipeline in (default: current directory)")
	initCmd.Flags().Int("steps", 2, "Number of template steps to include")
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir, _ := cmd.Flags().GetString("dir")
	numSteps, _ := cmd.Flags().GetInt("steps")

	if dir == "" {
		cfg, err := config.Load()
		if err == nil && cfg.PipelineDir != "" {
			dir = cfg.PipelineDir
		} else {
			dir = "."
		}
	}

	// Create directory if needed
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Build template pipeline
	p := &pipeline.Pipeline{
		Name:        name,
		Description: fmt.Sprintf("Pipeline: %s", name),
		Env: map[string]string{
			"PIPELINE_NAME": name,
		},
		Steps: generateTemplateSteps(numSteps),
	}

	// Write pipeline file
	filePath := filepath.Join(dir, name+".yaml")
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("pipeline file already exists: %s", filePath)
	}

	if err := pipeline.SavePipeline(filePath, p); err != nil {
		return fmt.Errorf("saving pipeline: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created pipeline %q at %s\n", name, filePath)
	return nil
}

func generateTemplateSteps(count int) []pipeline.Step {
	steps := make([]pipeline.Step, 0, count)
	for i := 1; i <= count; i++ {
		step := pipeline.Step{
			Name:    fmt.Sprintf("step-%d", i),
			Command: fmt.Sprintf("echo 'Running step %d'", i),
			Timeout: "30s",
		}
		if i > 1 {
			step.DependsOn = []string{fmt.Sprintf("step-%d", i-1)}
		}
		steps = append(steps, step)
	}
	return steps
}
