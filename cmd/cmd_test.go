package cmd

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestRunCommand_Verbose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pipeline.yaml")
	if err := os.WriteFile(path, []byte(`name: verbose-test
steps:
  - name: step-1
    command: echo ok
`), 0644); err != nil {
		t.Fatalf("write pipeline file: %v", err)
	}

	t.Run("with verbose", func(t *testing.T) {
		// Reset flags that may carry over from prior test runs.
		if err := runCmd.Flags().Set("dry-run", "false"); err != nil {
			t.Fatalf("reset dry-run flag: %v", err)
		}
		if err := runCmd.Flags().Set("verbose", "false"); err != nil {
			t.Fatalf("reset verbose flag: %v", err)
		}
		rootCmd.SetArgs([]string{"run", path, "--verbose"})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("run --verbose error = %v", err)
		}

		if !bytes.Contains(buf.Bytes(), []byte("Running step: step-1")) {
			t.Error("expected step name line in verbose output")
		}
	})

	t.Run("without verbose", func(t *testing.T) {
		// Reset flags that may carry over from prior test runs.
		if err := runCmd.Flags().Set("dry-run", "false"); err != nil {
			t.Fatalf("reset dry-run flag: %v", err)
		}
		if err := runCmd.Flags().Set("verbose", "false"); err != nil {
			t.Fatalf("reset verbose flag: %v", err)
		}
		rootCmd.SetArgs([]string{"run", path})
		var buf bytes.Buffer
		rootCmd.SetOut(&buf)

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("run error = %v", err)
		}

		if bytes.Contains(buf.Bytes(), []byte("Running step:")) {
			t.Error("expected no step-name line without --verbose")
		}
	})
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
