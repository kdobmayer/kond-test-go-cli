package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRunFixture creates a minimal run.json for testing status commands.
func writeRunFixture(t *testing.T, runDir, runID, status string) {
	t.Helper()
	runPath := filepath.Join(runDir, runID)
	if err := os.MkdirAll(runPath, 0755); err != nil {
		t.Fatalf("creating run dir: %v", err)
	}
	data := `{"run_id":"` + runID + `","pipeline_name":"test","status":"` + status + `","start_time":"2024-01-01T12:00:00Z","steps":[]}`
	if err := os.WriteFile(filepath.Join(runPath, "run.json"), []byte(data), 0644); err != nil {
		t.Fatalf("writing run.json: %v", err)
	}
}

func resetStatusFlags() {
	outputFormat = "table"
	statusAll = false
	statusJSONFlag = false
}

func TestStatusAll_JSONFlag(t *testing.T) {
	home := t.TempDir()
	runDir := filepath.Join(home, ".pipeline", "runs")
	writeRunFixture(t, runDir, "run-001", "success")
	writeRunFixture(t, runDir, "run-002", "failed")

	t.Setenv("HOME", home)
	t.Cleanup(resetStatusFlags)

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"status", "--all", "--json"})
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("status --all --json error = %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v\nout=%q", err, buf.String())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	// Sorted by run ID
	if got := result[0]["run_id"]; got != "run-001" {
		t.Errorf("result[0].run_id = %v, want run-001", got)
	}
	if got := result[0]["status"]; got != "success" {
		t.Errorf("result[0].status = %v, want success", got)
	}
	if got := result[1]["run_id"]; got != "run-002" {
		t.Errorf("result[1].run_id = %v, want run-002", got)
	}
	if got := result[1]["status"]; got != "failed" {
		t.Errorf("result[1].status = %v, want failed", got)
	}
}

func TestStatusAll_OutputFlagJSON(t *testing.T) {
	home := t.TempDir()
	runDir := filepath.Join(home, ".pipeline", "runs")
	writeRunFixture(t, runDir, "run-abc", "success")

	t.Setenv("HOME", home)
	t.Cleanup(resetStatusFlags)

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"status", "--all", "--output", "json"})
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("status --all --output json error = %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v\nout=%q", err, buf.String())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if got := result[0]["run_id"]; got != "run-abc" {
		t.Errorf("run_id = %v, want run-abc", got)
	}
}

func TestStatusAll_TableFormat(t *testing.T) {
	home := t.TempDir()
	runDir := filepath.Join(home, ".pipeline", "runs")
	writeRunFixture(t, runDir, "run-xyz", "success")

	t.Setenv("HOME", home)
	t.Cleanup(resetStatusFlags)

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"status", "--all"})
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("status --all error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "run-xyz") {
		t.Errorf("table output should contain run ID; got %q", out)
	}
	if !strings.Contains(out, "success") {
		t.Errorf("table output should contain status; got %q", out)
	}
	if !strings.Contains(out, "RUN ID") {
		t.Errorf("table output should contain header RUN ID; got %q", out)
	}
}

func TestStatusAll_NoRuns(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Cleanup(resetStatusFlags)

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"status", "--all", "--json"})
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No pipeline runs found") {
		t.Errorf("expected empty-state message; got %q", out)
	}
}

func TestStatusAll_JSONFlag_IsArray(t *testing.T) {
	home := t.TempDir()
	runDir := filepath.Join(home, ".pipeline", "runs")
	writeRunFixture(t, runDir, "single-run", "success")

	t.Setenv("HOME", home)
	t.Cleanup(resetStatusFlags)

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"status", "--all", "--json"})
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be a JSON array, not an object
	out := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(out, "[") || !strings.HasSuffix(out, "]") {
		t.Errorf("JSON output must be an array; got %q", out)
	}
}
