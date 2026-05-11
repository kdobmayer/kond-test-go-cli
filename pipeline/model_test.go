package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePipeline(t *testing.T) {
	yamlData := []byte(`
name: test-pipeline
description: A test pipeline
env:
  FOO: bar
steps:
  - name: step-1
    command: echo hello
    timeout: 30s
  - name: step-2
    command: echo world
    depends_on: [step-1]
    env:
      BAZ: qux
`)
	p, err := ParsePipeline(yamlData)
	if err != nil {
		t.Fatalf("ParsePipeline() error = %v", err)
	}
	if p.Name != "test-pipeline" {
		t.Errorf("Name = %q, want %q", p.Name, "test-pipeline")
	}
	if len(p.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(p.Steps))
	}
	if p.Steps[1].DependsOn[0] != "step-1" {
		t.Errorf("Steps[1].DependsOn[0] = %q, want %q", p.Steps[1].DependsOn[0], "step-1")
	}
	if p.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO] = %q, want %q", p.Env["FOO"], "bar")
	}
}

func TestParsePipelineInvalid(t *testing.T) {
	_, err := ParsePipeline([]byte("not: [valid: yaml: {{"))
	if err == nil {
		t.Error("ParsePipeline() expected error for invalid YAML")
	}
}

func TestValidate_Valid(t *testing.T) {
	p := &Pipeline{
		Name: "valid",
		Steps: []Step{
			{Name: "a", Command: "echo a"},
			{Name: "b", Command: "echo b", DependsOn: []string{"a"}},
		},
	}
	errs := p.Validate()
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, want 0: %v", len(errs), errs)
	}
}

func TestValidate_MissingName(t *testing.T) {
	p := &Pipeline{Steps: []Step{{Name: "a", Command: "echo a"}}}
	errs := p.Validate()
	if len(errs) != 1 || errs[0].Field != "name" {
		t.Errorf("expected 1 error on field 'name', got %v", errs)
	}
}

func TestValidate_NoSteps(t *testing.T) {
	p := &Pipeline{Name: "empty"}
	errs := p.Validate()
	if len(errs) != 1 || errs[0].Field != "steps" {
		t.Errorf("expected 1 error on field 'steps', got %v", errs)
	}
}

func TestValidate_DuplicateStepName(t *testing.T) {
	p := &Pipeline{
		Name:  "dup",
		Steps: []Step{{Name: "a", Command: "echo a"}, {Name: "a", Command: "echo b"}},
	}
	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Message == "duplicate step name: a" {
			found = true
		}
	}
	if !found {
		t.Error("expected duplicate step name error")
	}
}

func TestValidate_MissingCommand(t *testing.T) {
	p := &Pipeline{Name: "no-cmd", Steps: []Step{{Name: "a"}}}
	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "steps[0].command" {
			found = true
		}
	}
	if !found {
		t.Error("expected missing command error")
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	p := &Pipeline{
		Name:  "bad-timeout",
		Steps: []Step{{Name: "a", Command: "echo a", Timeout: "not-a-duration"}},
	}
	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "steps[0].timeout" {
			found = true
		}
	}
	if !found {
		t.Error("expected invalid timeout error")
	}
}

func TestValidate_NonPositiveTimeout(t *testing.T) {
	p := &Pipeline{
		Name:  "bad-timeout-non-positive",
		Steps: []Step{{Name: "a", Command: "echo a", Timeout: "0s"}},
	}
	errs := p.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "steps[0].timeout" {
			found = true
		}
	}
	if !found {
		t.Error("expected non-positive timeout error")
	}
}

func TestValidate_MissingDependency(t *testing.T) {
	p := &Pipeline{
		Name:  "bad-dep",
		Steps: []Step{{Name: "a", Command: "echo a", DependsOn: []string{"nonexistent"}}},
	}
	errs := p.Validate()
	if len(errs) == 0 {
		t.Error("expected dependency error")
	}
}

func TestValidate_CircularDependency(t *testing.T) {
	p := &Pipeline{
		Name: "circular",
		Steps: []Step{
			{Name: "a", Command: "echo a", DependsOn: []string{"b"}},
			{Name: "b", Command: "echo b", DependsOn: []string{"a"}},
		},
	}
	errs := p.Validate()
	if len(errs) == 0 {
		t.Error("expected circular dependency error")
	}
}

func TestTopologicalSort(t *testing.T) {
	steps := []Step{
		{Name: "a", Command: "echo a"},
		{Name: "b", Command: "echo b", DependsOn: []string{"a"}},
		{Name: "c", Command: "echo c", DependsOn: []string{"a"}},
		{Name: "d", Command: "echo d", DependsOn: []string{"b", "c"}},
	}
	levels, err := TopologicalSort(steps)
	if err != nil {
		t.Fatalf("TopologicalSort() error = %v", err)
	}
	if len(levels) < 2 {
		t.Fatalf("expected at least 2 levels, got %d", len(levels))
	}
	if len(levels[0]) != 1 || levels[0][0].Name != "a" {
		t.Errorf("first level should be [a], got %v", levels[0])
	}
}

func TestTopologicalSort_Parallel(t *testing.T) {
	steps := []Step{
		{Name: "a", Command: "echo a"},
		{Name: "b", Command: "echo b"},
		{Name: "c", Command: "echo c"},
	}
	levels, err := TopologicalSort(steps)
	if err != nil {
		t.Fatalf("TopologicalSort() error = %v", err)
	}
	if len(levels) != 1 || len(levels[0]) != 3 {
		t.Errorf("expected 1 level with 3 steps, got %d levels", len(levels))
	}
}

func TestSaveAndLoadPipeline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	p := &Pipeline{
		Name:  "save-test",
		Steps: []Step{{Name: "a", Command: "echo a", Timeout: "10s"}},
	}

	if err := SavePipeline(path, p); err != nil {
		t.Fatalf("SavePipeline() error = %v", err)
	}

	loaded, err := LoadPipeline(path)
	if err != nil {
		t.Fatalf("LoadPipeline() error = %v", err)
	}

	if loaded.Name != p.Name {
		t.Errorf("loaded Name = %q, want %q", loaded.Name, p.Name)
	}
	if len(loaded.Steps) != 1 {
		t.Fatalf("loaded Steps count = %d, want 1", len(loaded.Steps))
	}
	if loaded.Steps[0].Timeout != "10s" {
		t.Errorf("loaded timeout = %q, want %q", loaded.Steps[0].Timeout, "10s")
	}
}

func TestLoadPipeline_NotFound(t *testing.T) {
	_, err := LoadPipeline("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidationError_Error(t *testing.T) {
	e := ValidationError{Field: "name", Message: "is required"}
	got := e.Error()
	want := "name: is required"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestListRuns_Empty(t *testing.T) {
	dir := t.TempDir()
	runs, err := ListRuns(dir)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListRuns_NonExistent(t *testing.T) {
	runs, err := ListRuns("/nonexistent/dir")
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if runs != nil {
		t.Errorf("expected nil, got %v", runs)
	}
}

func TestLoadRun_NotFound(t *testing.T) {
	_, err := LoadRun("/nonexistent", "run-1")
	if err == nil {
		t.Error("expected error for nonexistent run")
	}
}

func TestLoadStepLog_NotFound(t *testing.T) {
	_, err := LoadStepLog("/nonexistent", "run-1", "step-1")
	if err == nil {
		t.Error("expected error for nonexistent log")
	}
}

func TestMarshalYAML(t *testing.T) {
	data := map[string]string{"key": "value"}
	out, err := MarshalYAML(data)
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestListRuns_WithEntries(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "run-1"), 0755)
	os.MkdirAll(filepath.Join(dir, "run-2"), 0755)
	os.WriteFile(filepath.Join(dir, "not-a-run.txt"), []byte("x"), 0644)

	runs, err := ListRuns(dir)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}
