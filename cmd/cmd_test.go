package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
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

func writeRunFile(t *testing.T, dir, id, status string) {
	t.Helper()
	data, err := json.Marshal(map[string]interface{}{"run_id": id, "status": status})
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, id+".json"), data, 0644); err != nil {
		t.Fatalf("write run file: %v", err)
	}
}

func TestShowAllRunsJSON(t *testing.T) {
	dir := t.TempDir()
	writeRunFile(t, dir, "run-001", "success")
	writeRunFile(t, dir, "run-002", "failed")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := showAllRuns(cmd, dir, "json"); err != nil {
		t.Fatalf("showAllRuns error = %v", err)
	}

	var got []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(got))
	}
	if got[0]["run_id"] != "run-001" || got[0]["status"] != "success" {
		t.Errorf("first entry = %v, want run-001/success", got[0])
	}
	if got[1]["run_id"] != "run-002" || got[1]["status"] != "failed" {
		t.Errorf("second entry = %v, want run-002/failed", got[1])
	}
}

func TestShowAllRunsTable(t *testing.T) {
	dir := t.TempDir()
	writeRunFile(t, dir, "run-001", "success")

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := showAllRuns(cmd, dir, "table"); err != nil {
		t.Fatalf("showAllRuns error = %v", err)
	}

	out := buf.Bytes()
	if !bytes.Contains(out, []byte("RUN ID")) {
		t.Error("expected table header 'RUN ID' in output")
	}
	if !bytes.Contains(out, []byte("run-001")) {
		t.Error("expected 'run-001' in output")
	}
}

func TestShowAllRunsEmpty(t *testing.T) {
	dir := t.TempDir()

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := showAllRuns(cmd, dir, "json"); err != nil {
		t.Fatalf("showAllRuns error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("No pipeline runs found")) {
		t.Error("expected 'No pipeline runs found' message")
	}
}

func TestStatusCmdHasJSONFlag(t *testing.T) {
	f := statusCmd.Flags().Lookup("json")
	if f == nil {
		t.Fatal("statusCmd missing --json flag")
	}
	if f.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", f.DefValue, "false")
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
