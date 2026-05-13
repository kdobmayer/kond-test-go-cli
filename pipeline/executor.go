package pipeline

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Executor runs pipeline steps
type Executor struct {
	Pipeline *Pipeline
	RunDir   string
	Run      *PipelineRun
	Logs     map[string]*StepLog
	mu       sync.Mutex
	Verbose  bool
	Out      io.Writer
}

// NewExecutor creates a new pipeline executor
func NewExecutor(p *Pipeline, runDir string) *Executor {
	runID := fmt.Sprintf("%s-%d", p.Name, time.Now().Unix())
	return &Executor{
		Pipeline: p,
		RunDir:   runDir,
		Run: &PipelineRun{
			PipelineName: p.Name,
			RunID:        runID,
			Status:       "pending",
			Steps:        make([]StepStatus, 0, len(p.Steps)),
		},
		Logs: make(map[string]*StepLog),
	}
}

// Execute runs the pipeline
func (e *Executor) Execute() error {
	e.Run.Status = "running"
	e.Run.StartTime = time.Now()

	// Initialize step statuses
	for _, step := range e.Pipeline.Steps {
		e.Run.Steps = append(e.Run.Steps, StepStatus{
			Name:   step.Name,
			Status: "pending",
		})
	}

	// Get execution levels (parallel groups)
	levels, err := TopologicalSort(e.Pipeline.Steps)
	if err != nil {
		e.Run.Status = "failed"
		e.Run.EndTime = time.Now()
		return fmt.Errorf("sorting steps: %w", err)
	}

	// Execute each level in parallel
	for _, level := range levels {
		if err := e.executeLevel(level); err != nil {
			e.Run.Status = "failed"
			e.Run.EndTime = time.Now()
			e.saveRun()
			return err
		}
	}

	e.Run.Status = "completed"
	e.Run.EndTime = time.Now()
	e.saveRun()
	return nil
}

// executeLevel runs all steps in a level concurrently
func (e *Executor) executeLevel(steps []Step) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(steps))

	for _, step := range steps {
		wg.Add(1)
		go func(s Step) {
			defer wg.Done()
			if err := e.executeStep(s); err != nil {
				errCh <- err
			}
		}(step)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any
	for err := range errCh {
		return err
	}
	return nil
}

// executeStep runs a single step
// NOTE: timeout field is parsed but NOT enforced (intentional rough edge)
func (e *Executor) executeStep(step Step) error {
	if err := e.writeVerboseStepStart(step.Name); err != nil {
		return fmt.Errorf("writing verbose output for step %q: %w", step.Name, err)
	}
	e.updateStepStatus(step.Name, "running", 0, "")

	startTime := time.Now()

	// Build environment
	env := os.Environ()
	for k, v := range e.Pipeline.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range step.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("sh", "-c", step.Command)
	cmd.Env = env
	if step.WorkDir != "" {
		cmd.Dir = step.WorkDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Store logs
	e.mu.Lock()
	e.Logs[step.Name] = &StepLog{
		StepName: step.Name,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
	e.mu.Unlock()

	exitCode := 0
	errMsg := ""
	status := "completed"

	if err != nil {
		status = "failed"
		errMsg = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Update step status with timing
	e.mu.Lock()
	for i := range e.Run.Steps {
		if e.Run.Steps[i].Name == step.Name {
			e.Run.Steps[i].Status = status
			e.Run.Steps[i].ExitCode = exitCode
			e.Run.Steps[i].StartTime = startTime
			e.Run.Steps[i].EndTime = endTime
			e.Run.Steps[i].Duration = duration
			e.Run.Steps[i].Error = errMsg
			break
		}
	}
	e.mu.Unlock()

	if err != nil {
		return fmt.Errorf("step %q failed: %w", step.Name, err)
	}
	return nil
}

func (e *Executor) writeVerboseStepStart(stepName string) error {
	if !e.Verbose || e.Out == nil {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, err := fmt.Fprintf(e.Out, "Running step: %s\n", stepName); err != nil {
		return fmt.Errorf("writing step start line: %w", err)
	}
	return nil
}

// updateStepStatus updates the status of a step
func (e *Executor) updateStepStatus(name, status string, exitCode int, errMsg string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.Run.Steps {
		if e.Run.Steps[i].Name == name {
			e.Run.Steps[i].Status = status
			e.Run.Steps[i].ExitCode = exitCode
			e.Run.Steps[i].Error = errMsg
			if status == "running" {
				e.Run.Steps[i].StartTime = time.Now()
			}
			break
		}
	}
}

// saveRun persists the run state to disk
func (e *Executor) saveRun() error {
	if e.RunDir == "" {
		return nil
	}

	runDir := filepath.Join(e.RunDir, e.Run.RunID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("creating run directory: %w", err)
	}

	// Save run state
	runPath := filepath.Join(runDir, "run.yaml")
	data, err := MarshalYAML(e.Run)
	if err != nil {
		return fmt.Errorf("marshaling run state: %w", err)
	}
	if err := os.WriteFile(runPath, data, 0644); err != nil {
		return fmt.Errorf("writing run state: %w", err)
	}

	// Save logs
	for name, log := range e.Logs {
		logDir := filepath.Join(runDir, "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("creating logs directory: %w", err)
		}
		logPath := filepath.Join(logDir, name+".yaml")
		logData, err := MarshalYAML(log)
		if err != nil {
			return fmt.Errorf("marshaling log: %w", err)
		}
		if err := os.WriteFile(logPath, logData, 0644); err != nil {
			return fmt.Errorf("writing log: %w", err)
		}
	}

	return nil
}

// MarshalYAML marshals a value to YAML bytes
func MarshalYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// LoadRun reads a pipeline run from disk
func LoadRun(runDir, runID string) (*PipelineRun, error) {
	runPath := filepath.Join(runDir, runID, "run.yaml")
	data, err := os.ReadFile(runPath)
	if err != nil {
		return nil, fmt.Errorf("reading run file: %w", err)
	}
	var run PipelineRun
	if err := yaml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parsing run file: %w", err)
	}
	return &run, nil
}

// LoadStepLog reads a step log from disk
func LoadStepLog(runDir, runID, stepName string) (*StepLog, error) {
	logPath := filepath.Join(runDir, runID, "logs", stepName+".yaml")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("reading log file: %w", err)
	}
	var log StepLog
	if err := yaml.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("parsing log file: %w", err)
	}
	return &log, nil
}

// ListRuns returns all run IDs in the run directory
func ListRuns(runDir string) ([]string, error) {
	entries, err := os.ReadDir(runDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading run directory: %w", err)
	}
	var runs []string
	for _, entry := range entries {
		if entry.IsDir() {
			runs = append(runs, entry.Name())
		}
	}
	return runs, nil
}
