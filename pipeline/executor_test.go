package pipeline

import (
	"bytes"
	"testing"
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

	if err := executor.Execute(); err != nil {
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

	err := executor.Execute()
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

	if err := executor.Execute(); err != nil {
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

	if err := executor.Execute(); err != nil {
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

	if err := executor.Execute(); err != nil {
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
	if err := executor.Execute(); err != nil {
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
	if err := executor.Execute(); err != nil {
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

	if err := executor.Execute(); err != nil {
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

func TestExecutor_VerboseOutput(t *testing.T) {
	p := &Pipeline{
		Name: "verbose-test",
		Steps: []Step{
			{Name: "greet", Command: "echo hello"},
			{Name: "warn", Command: "echo oops >&2"},
		},
	}
	dir := t.TempDir()
	executor := NewExecutor(p, dir)

	var verboseBuf bytes.Buffer
	executor.Verbose = &verboseBuf

	if err := executor.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Logs are still fully captured regardless of verbose
	if executor.Logs["greet"].Stdout != "hello\n" {
		t.Errorf("log stdout = %q, want %q", executor.Logs["greet"].Stdout, "hello\n")
	}
	if executor.Logs["warn"].Stderr != "oops\n" {
		t.Errorf("log stderr = %q, want %q", executor.Logs["warn"].Stderr, "oops\n")
	}

	out := verboseBuf.String()
	if !bytes.Contains([]byte(out), []byte("[greet] hello")) {
		t.Errorf("verbose output %q should contain %q", out, "[greet] hello")
	}
	if !bytes.Contains([]byte(out), []byte("[warn err] oops")) {
		t.Errorf("verbose output %q should contain %q", out, "[warn err] oops")
	}
}
