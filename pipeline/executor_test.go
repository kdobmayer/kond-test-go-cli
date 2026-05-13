package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecutor_SimpleSuccess(t *testing.T) {
	p := &Pipeline{
		Name: "test",
		Steps: []Step{
			{Name: "hello", Command: "echo hello"},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if executor.Run.Status != "completed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "completed")
	}

	if len(executor.Run.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(executor.Run.Steps))
	}

	if executor.Run.Steps[0].Status != "completed" {
		t.Errorf("Step status = %q, want %q", executor.Run.Steps[0].Status, "completed")
	}

	// Check logs were captured
	log, ok := executor.Logs["hello"]
	if !ok {
		t.Fatal("expected log for step 'hello'")
	}
	if log.Stdout != "hello\n" {
		t.Errorf("stdout = %q, want %q", log.Stdout, "hello\n")
	}
}

func TestExecutor_StepFailure(t *testing.T) {
	p := &Pipeline{
		Name: "fail-test",
		Steps: []Step{
			{Name: "fail", Command: "exit 1"},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	err := executor.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() expected error")
	}

	if executor.Run.Status != "failed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "failed")
	}

	if executor.Run.Steps[0].ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", executor.Run.Steps[0].ExitCode)
	}
}

func TestExecutor_MultipleSteps(t *testing.T) {
	p := &Pipeline{
		Name: "multi",
		Steps: []Step{
			{Name: "first", Command: "echo first"},
			{Name: "second", Command: "echo second", DependsOn: []string{"first"}},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, s := range executor.Run.Steps {
		if s.Status != "completed" {
			t.Errorf("step %q status = %q, want %q", s.Name, s.Status, "completed")
		}
	}
}

func TestExecutor_ParallelSteps(t *testing.T) {
	p := &Pipeline{
		Name: "parallel",
		Steps: []Step{
			{Name: "a", Command: "echo a"},
			{Name: "b", Command: "echo b"},
			{Name: "c", Command: "echo c"},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if executor.Run.Status != "completed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "completed")
	}
}

func TestExecutor_Environment(t *testing.T) {
	p := &Pipeline{
		Name: "env-test",
		Env:  map[string]string{"PIPELINE_VAR": "pipeline_value"},
		Steps: []Step{
			{
				Name:    "env",
				Command: "echo $PIPELINE_VAR $STEP_VAR",
				Env:     map[string]string{"STEP_VAR": "step_value"},
			},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	log := executor.Logs["env"]
	if log.Stdout != "pipeline_value step_value\n" {
		t.Errorf("stdout = %q, want %q", log.Stdout, "pipeline_value step_value\n")
	}
}

func TestExecutor_SaveRun(t *testing.T) {
	p := &Pipeline{
		Name:  "save-test",
		Steps: []Step{{Name: "a", Command: "echo saved"}},
	}
	dir := t.TempDir()
	executor := NewExecutor(p, dir)
	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify run was saved and loadable
	run, err := LoadRun(dir, executor.Run.RunID)
	if err != nil {
		t.Fatalf("LoadRun() error = %v", err)
	}
	if run.Status != "completed" {
		t.Errorf("loaded run status = %q, want %q", run.Status, "completed")
	}
	log, err := LoadStepLog(dir, executor.Run.RunID, "a")
	if err != nil {
		t.Fatalf("LoadStepLog() error = %v", err)
	}
	if log.Stdout != "saved\n" {
		t.Errorf("loaded log stdout = %q, want %q", log.Stdout, "saved\n")
	}
}

func TestExecutor_EmptyRunDir(t *testing.T) {
	p := &Pipeline{
		Name: "no-save",
		Steps: []Step{
			{Name: "a", Command: "echo ok"},
		},
	}

	executor := NewExecutor(p, "")
	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	// Should not panic with empty RunDir
}

func TestExecutor_StepWithStderr(t *testing.T) {
	p := &Pipeline{
		Name: "stderr-test",
		Steps: []Step{
			{Name: "err", Command: "echo error >&2"},
		},
	}

	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	if err := executor.Execute(context.Background()); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	log := executor.Logs["err"]
	if log.Stderr != "error\n" {
		t.Errorf("stderr = %q, want %q", log.Stderr, "error\n")
	}
}

func TestNewExecutor_RunID(t *testing.T) {
	p := &Pipeline{Name: "test-id"}
	e := NewExecutor(p, "")
	if e.Run.RunID == "" {
		t.Error("expected non-empty RunID")
	}
	if e.Run.Status != "pending" {
		t.Errorf("initial status = %q, want %q", e.Run.Status, "pending")
	}
}

func TestExecutor_Timeout(t *testing.T) {
	p := &Pipeline{
		Name: "timeout-test",
		Steps: []Step{
			{Name: "slow", Command: "sleep 60"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	executor := NewExecutor(p, "")
	err := executor.Execute(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Execute() expected error from timeout")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error should wrap context.DeadlineExceeded, got: %v", err)
	}
	if executor.Run.Status != "failed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "failed")
	}
	// Should terminate well before the 60s sleep completes
	if elapsed > 10*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

func TestExecutor_TimeoutFastPipeline(t *testing.T) {
	p := &Pipeline{
		Name: "fast-with-timeout",
		Steps: []Step{
			{Name: "fast", Command: "echo done"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	executor := NewExecutor(p, "")
	if err := executor.Execute(ctx); err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if executor.Run.Status != "completed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "completed")
	}
}

func TestExecutor_NilContext(t *testing.T) {
	p := &Pipeline{
		Name: "nil-context",
		Steps: []Step{
			{Name: "fast", Command: "echo done"},
		},
	}

	executor := NewExecutor(p, "")
	if err := executor.Execute(nil); err != nil {
		t.Fatalf("Execute(nil) unexpected error = %v", err)
	}
	if executor.Run.Status != "completed" {
		t.Errorf("Run.Status = %q, want %q", executor.Run.Status, "completed")
	}
}
