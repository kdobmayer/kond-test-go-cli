package pipeline

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Pipeline represents a pipeline definition
type Pipeline struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Steps       []Step            `yaml:"steps"`
}

// Step represents a single step in a pipeline
type Step struct {
	Name      string            `yaml:"name"`
	Command   string            `yaml:"command"`
	Timeout   string            `yaml:"timeout,omitempty"`
	DependsOn []string          `yaml:"depends_on,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	WorkDir   string            `yaml:"workdir,omitempty"`
}

// StepStatus represents the execution status of a step
type StepStatus struct {
	Name      string        `yaml:"name" json:"name"`
	Status    string        `yaml:"status" json:"status"`
	ExitCode  int           `yaml:"exit_code" json:"exit_code"`
	StartTime time.Time     `yaml:"start_time" json:"start_time"`
	EndTime   time.Time     `yaml:"end_time,omitempty" json:"end_time,omitempty"`
	Duration  time.Duration `yaml:"duration,omitempty" json:"duration,omitempty"`
	Error     string        `yaml:"error,omitempty" json:"error,omitempty"`
}

// PipelineRun represents the state of a pipeline execution
type PipelineRun struct {
	PipelineName string       `yaml:"pipeline_name" json:"pipeline_name"`
	RunID        string       `yaml:"run_id" json:"run_id"`
	Status       string       `yaml:"status" json:"status"`
	StartTime    time.Time    `yaml:"start_time" json:"start_time"`
	EndTime      time.Time    `yaml:"end_time,omitempty" json:"end_time,omitempty"`
	Steps        []StepStatus `yaml:"steps" json:"steps"`
}

// StepLog holds captured output from a step execution
type StepLog struct {
	StepName string `yaml:"step_name" json:"step_name"`
	Stdout   string `yaml:"stdout" json:"stdout"`
	Stderr   string `yaml:"stderr" json:"stderr"`
}

// LoadPipeline reads and parses a pipeline YAML file
func LoadPipeline(path string) (*Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading pipeline file: %w", err)
	}
	return ParsePipeline(data)
}

// ParsePipeline parses pipeline YAML data
func ParsePipeline(data []byte) (*Pipeline, error) {
	var p Pipeline
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing pipeline YAML: %w", err)
	}
	return &p, nil
}

// SavePipeline writes a pipeline to a YAML file
func SavePipeline(path string, p *Pipeline) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshaling pipeline: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Validate checks a pipeline definition for errors
func (p *Pipeline) Validate() []ValidationError {
	var errs []ValidationError

	if p.Name == "" {
		errs = append(errs, ValidationError{
			Field:   "name",
			Message: "pipeline name is required",
		})
	}

	if len(p.Steps) == 0 {
		errs = append(errs, ValidationError{
			Field:   "steps",
			Message: "pipeline must have at least one step",
		})
	}

	stepNames := make(map[string]bool)
	for i, step := range p.Steps {
		if step.Name == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("steps[%d].name", i),
				Message: "step name is required",
			})
		} else if stepNames[step.Name] {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("steps[%d].name", i),
				Message: fmt.Sprintf("duplicate step name: %s", step.Name),
			})
		} else {
			stepNames[step.Name] = true
		}

		if step.Command == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("steps[%d].command", i),
				Message: "step command is required",
			})
		}

		for _, dep := range step.DependsOn {
			if !stepNames[dep] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("steps[%d].depends_on", i),
					Message: fmt.Sprintf("dependency %q not found or defined after this step", dep),
				})
			}
		}

		if step.Timeout != "" {
			duration, err := time.ParseDuration(step.Timeout)
			if err != nil {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("steps[%d].timeout", i),
					Message: fmt.Sprintf("invalid timeout duration: %s", step.Timeout),
				})
			} else if duration <= 0 {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("steps[%d].timeout", i),
					Message: fmt.Sprintf("timeout duration must be positive: %s", step.Timeout),
				})
			}
		}
	}

	// Check for circular dependencies
	if cyclicErrs := detectCycles(p.Steps); len(cyclicErrs) > 0 {
		errs = append(errs, cyclicErrs...)
	}

	return errs
}

// ValidationError represents a validation issue
type ValidationError struct {
	Field   string `yaml:"field" json:"field"`
	Message string `yaml:"message" json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// detectCycles checks for circular dependencies in steps
func detectCycles(steps []Step) []ValidationError {
	var errs []ValidationError
	graph := make(map[string][]string)
	for _, s := range steps {
		graph[s.Name] = s.DependsOn
	}

	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var visit func(node string) bool
	visit = func(node string) bool {
		visited[node] = true
		inStack[node] = true

		for _, dep := range graph[node] {
			if !visited[dep] {
				if visit(dep) {
					return true
				}
			} else if inStack[dep] {
				errs = append(errs, ValidationError{
					Field:   "steps",
					Message: fmt.Sprintf("circular dependency detected involving step %q", node),
				})
				return true
			}
		}

		inStack[node] = false
		return false
	}

	for _, s := range steps {
		if !visited[s.Name] {
			visit(s.Name)
		}
	}

	return errs
}

// TopologicalSort returns steps in execution order respecting dependencies
func TopologicalSort(steps []Step) ([][]Step, error) {
	inDegree := make(map[string]int)
	stepMap := make(map[string]Step)
	dependents := make(map[string][]string)

	for _, s := range steps {
		inDegree[s.Name] = len(s.DependsOn)
		stepMap[s.Name] = s
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], s.Name)
		}
	}

	var levels [][]Step
	remaining := len(steps)

	for remaining > 0 {
		var level []Step
		for name, deg := range inDegree {
			if deg == 0 {
				level = append(level, stepMap[name])
			}
		}
		if len(level) == 0 {
			return nil, fmt.Errorf("circular dependency detected")
		}
		for _, s := range level {
			delete(inDegree, s.Name)
			remaining--
			for _, dep := range dependents[s.Name] {
				inDegree[dep]--
			}
		}
		levels = append(levels, level)
	}

	return levels, nil
}
