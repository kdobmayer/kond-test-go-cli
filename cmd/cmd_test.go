package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	dir := t.TempDir()

	rootCmd.SetArgs([]string{"init", "test-pipeline", "-d", dir})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command error = %v", err)
	}

	// Verify file was created
	path := filepath.Join(dir, "test-pipeline.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("pipeline file not created: %v", err)
	}
}

func TestInitCommand_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.yaml")
	os.WriteFile(path, []byte("name: existing"), 0644)

	rootCmd.SetArgs([]string{"init", "existing", "-d", dir})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when file already exists")
	}
}

func TestInitCommand_CustomSteps(t *testing.T) {
	dir := t.TempDir()

	rootCmd.SetArgs([]string{"init", "multi", "-d", dir, "--steps", "4"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command error = %v", err)
	}

	// Read and verify step count
	data, _ := os.ReadFile(filepath.Join(dir, "multi.yaml"))
	content := string(data)
	// Should have step-4
	if !bytes.Contains([]byte(content), []byte("step-4")) {
		t.Error("expected step-4 in pipeline")
	}
}

func TestValidateCommand_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.yaml")
	os.WriteFile(path, []byte(`
name: valid
steps:
  - name: a
    command: echo a
`), 0644)

	rootCmd.SetArgs([]string{"validate", path})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("validate command error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("valid")) {
		t.Error("expected 'valid' in output")
	}
}

func TestValidateCommand_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	os.WriteFile(path, []byte(`
name: ""
steps: []
`), 0644)

	rootCmd.SetArgs([]string{"validate", path})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid pipeline")
	}
}

func TestRunCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pipeline.yaml")
	os.WriteFile(path, []byte(`
name: dry-run-test
steps:
  - name: a
    command: echo hello
  - name: b
    command: echo world
    depends_on: [a]
`), 0644)

	rootCmd.SetArgs([]string{"run", path, "--dry-run"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("run --dry-run error = %v", err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("Level")) {
		t.Error("expected execution plan output")
	}
}

func TestGenerateTemplateSteps(t *testing.T) {
	steps := generateTemplateSteps(3)
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	if steps[0].Name != "step-1" {
		t.Errorf("first step name = %q, want %q", steps[0].Name, "step-1")
	}
	if len(steps[0].DependsOn) != 0 {
		t.Error("first step should have no dependencies")
	}
	if len(steps[2].DependsOn) != 1 || steps[2].DependsOn[0] != "step-2" {
		t.Error("step-3 should depend on step-2")
	}
}

func TestSearchCommand_Match(t *testing.T) {
	rootCmd.SetArgs([]string{"search", "--pattern", "foo"})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetIn(strings.NewReader("foo bar\nbaz\nfoo again\n"))

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search command error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "1: foo bar") {
		t.Errorf("expected line 1 match, got: %s", got)
	}
	if !strings.Contains(got, "3: foo again") {
		t.Errorf("expected line 3 match, got: %s", got)
	}
	if strings.Contains(got, "baz") {
		t.Errorf("non-matching line should not appear, got: %s", got)
	}
}

func TestSearchCommand_InvalidRegex(t *testing.T) {
	rootCmd.SetArgs([]string{"search", "--pattern", "["})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetIn(strings.NewReader("anything\n"))

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid regex pattern") {
		t.Errorf("error should mention 'invalid regex pattern', got: %v", err)
	}
}

func TestSearchCommand_MissingPattern(t *testing.T) {
	// Reset flag state that may be polluted by prior tests sharing rootCmd.
	searchPattern = ""
	if f := searchCmd.Flags().Lookup("pattern"); f != nil {
		f.Changed = false
	}

	rootCmd.SetArgs([]string{"search"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --pattern flag is missing")
	}
	if !strings.Contains(err.Error(), `required flag(s) "pattern" not set`) {
		t.Errorf("error should mention required flag, got: %v", err)
	}
}
