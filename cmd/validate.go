package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateCmd = &cobra.Command{
	Use:   "validate [pipeline-file]",
	Short: "Validate a pipeline YAML file",
	Long:  `Check a pipeline definition for errors without executing it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().Bool("strict", false, "Enable strict validation (check for best practices)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	pipelineFile := args[0]
	strict, _ := cmd.Flags().GetBool("strict")

	p, err := pipeline.LoadPipeline(pipelineFile)
	if err != nil {
		return fmt.Errorf("loading pipeline: %w", err)
	}

	errs := p.Validate()

	// Strict mode adds additional checks
	if strict {
		strictErrs := validateStrict(p)
		errs = append(errs, strictErrs...)
	}

	if len(errs) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline %q is valid.\n", p.Name)
		return nil
	}

	return renderValidationErrors(cmd, errs)
}

// validateStrict performs additional best-practice checks
func validateStrict(p *pipeline.Pipeline) []pipeline.ValidationError {
	var errs []pipeline.ValidationError

	for i, step := range p.Steps {
		// Check for missing timeout
		if step.Timeout == "" {
			errs = append(errs, pipeline.ValidationError{
				Field:   fmt.Sprintf("steps[%d].timeout", i),
				Message: fmt.Sprintf("step %q has no timeout configured", step.Name),
			})
		}

		// Check for missing description (workdir as proxy for organization)
		if step.WorkDir == "" && len(p.Steps) > 5 {
			errs = append(errs, pipeline.ValidationError{
				Field:   fmt.Sprintf("steps[%d].workdir", i),
				Message: fmt.Sprintf("step %q has no workdir (recommended for large pipelines)", step.Name),
			})
		}
	}

	// Check for pipeline-level env
	if len(p.Env) == 0 {
		errs = append(errs, pipeline.ValidationError{
			Field:   "env",
			Message: "no pipeline-level environment variables defined",
		})
	}

	return errs
}

func renderValidationErrors(cmd *cobra.Command, errs []pipeline.ValidationError) error {
	// NOTE: duplicated output formatting (intentional rough edge)
	switch outputFormat {
	case "json":
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(errs)
	case "yaml":
		enc := yaml.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(errs)
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tMESSAGE")
		fmt.Fprintln(w, "-----\t-------")
		for _, e := range errs {
			fmt.Fprintf(w, "%s\t%s\n", e.Field, e.Message)
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}

	return fmt.Errorf("pipeline has %d validation error(s)", len(errs))
}
