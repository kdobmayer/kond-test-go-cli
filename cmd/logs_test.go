package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kdobmayer/kond-test-go-cli/pipeline"
	"gopkg.in/yaml.v3"
)

const testRunID = "run-20240101-000000"

// setupLogsFixture writes a run + step logs under a temp HOME directory and
// returns the temp dir. Call t.Setenv("HOME", ...) is handled internally.
func setupLogsFixture(t *testing.T, steps []pipeline.StepStatus, logs map[string]*pipeline.StepLog) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	runDir := filepath.Join(tmpDir, ".pipeline", "runs", testRunID)
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("creating dirs: %v", err)
	}

	run := &pipeline.PipelineRun{
		PipelineName: "test",
		RunID:        testRunID,
		Status:       "success",
		StartTime:    time.Now(),
		Steps:        steps,
	}
	data, _ := pipeline.MarshalYAML(run)
	if err := os.WriteFile(filepath.Join(runDir, "run.yaml"), data, 0644); err != nil {
		t.Fatalf("writing run.yaml: %v", err)
	}

	for name, log := range logs {
		logData, _ := pipeline.MarshalYAML(log)
		if err := os.WriteFile(filepath.Join(logDir, name+".yaml"), logData, 0644); err != nil {
			t.Fatalf("writing log %s: %v", name, err)
		}
	}
}

// execLogs runs the logs command with the given args and returns captured output.
func execLogs(t *testing.T, args ...string) string {
	t.Helper()
	rootCmd.SetArgs(append([]string{"logs"}, args...))
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("logs command error: %v", err)
	}
	return buf.String()
}

// TestLimitLines covers the helper in isolation.
func TestLimitLines(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		n     int
		want  string
	}{
		{"zero limit unchanged", "a\nb\nc\n", 0, "a\nb\nc\n"},
		{"empty unchanged", "", 3, ""},
		{"limit exceeds lines", "a\nb\n", 5, "a\nb\n"},
		{"limit equals lines", "a\nb\nc\n", 3, "a\nb\nc\n"},
		{"last 2 of 4 with trailing newline", "a\nb\nc\nd\n", 2, "c\nd\n"},
		{"last 2 of 3 no trailing newline", "a\nb\nc", 2, "b\nc"},
		{"limit 1", "x\ny\nz\n", 1, "z\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := limitLines(tc.s, tc.n)
			if got != tc.want {
				t.Errorf("limitLines(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
			}
		})
	}
}

// TestLogsCmd_SingleStep_Table checks table output line capping.
func TestLogsCmd_SingleStep_Table(t *testing.T) {
	stdout := "line1\nline2\nline3\nline4\nline5\n"
	stderr := "err1\nerr2\nerr3\n"
	setupLogsFixture(t,
		[]pipeline.StepStatus{{Name: "step-a", Status: "success"}},
		map[string]*pipeline.StepLog{"step-a": {StepName: "step-a", Stdout: stdout, Stderr: stderr}},
	)

	t.Run("no limit shows all lines", func(t *testing.T) {
		out := execLogs(t, testRunID, "step-a", "-o", "table", "--limit", "0")
		if !strings.Contains(out, "line1") {
			t.Error("expected line1 in unlimited output")
		}
		if !strings.Contains(out, "line5") {
			t.Error("expected line5 in unlimited output")
		}
	})

	t.Run("limit 2 shows last 2 lines of each stream", func(t *testing.T) {
		out := execLogs(t, testRunID, "step-a", "-o", "table", "--limit", "2")
		if strings.Contains(out, "line1") || strings.Contains(out, "line2") || strings.Contains(out, "line3") {
			t.Error("expected early lines to be absent with --limit 2")
		}
		if !strings.Contains(out, "line4") || !strings.Contains(out, "line5") {
			t.Error("expected last 2 stdout lines to be present")
		}
		if strings.Contains(out, "err1") {
			t.Error("expected err1 to be absent with --limit 2")
		}
		if !strings.Contains(out, "err2") || !strings.Contains(out, "err3") {
			t.Error("expected last 2 stderr lines to be present")
		}
	})

	t.Run("limit greater than line count shows all", func(t *testing.T) {
		out := execLogs(t, testRunID, "step-a", "-o", "table", "--limit", "100")
		if !strings.Contains(out, "line1") || !strings.Contains(out, "line5") {
			t.Error("expected all lines with limit > line count")
		}
	})
}

// TestLogsCmd_SingleStep_JSON checks JSON output line capping.
func TestLogsCmd_SingleStep_JSON(t *testing.T) {
	stdout := "a\nb\nc\nd\n"
	stderr := "x\ny\n"
	setupLogsFixture(t,
		[]pipeline.StepStatus{{Name: "step-a", Status: "success"}},
		map[string]*pipeline.StepLog{"step-a": {StepName: "step-a", Stdout: stdout, Stderr: stderr}},
	)

	t.Run("no limit", func(t *testing.T) {
		out := execLogs(t, testRunID, "step-a", "-o", "json", "--limit", "0")
		var result map[string]string
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if result["stdout"] != stdout {
			t.Errorf("stdout = %q, want %q", result["stdout"], stdout)
		}
	})

	t.Run("limit 2", func(t *testing.T) {
		out := execLogs(t, testRunID, "step-a", "-o", "json", "--limit", "2")
		var result map[string]string
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		wantStdout := "c\nd\n"
		if result["stdout"] != wantStdout {
			t.Errorf("stdout = %q, want %q", result["stdout"], wantStdout)
		}
		wantStderr := "x\ny\n"
		if result["stderr"] != wantStderr {
			t.Errorf("stderr = %q, want %q", result["stderr"], wantStderr)
		}
	})
}

// TestLogsCmd_SingleStep_YAML checks YAML output line capping.
func TestLogsCmd_SingleStep_YAML(t *testing.T) {
	stdout := "a\nb\nc\n"
	stderr := "p\nq\nr\n"
	setupLogsFixture(t,
		[]pipeline.StepStatus{{Name: "step-a", Status: "success"}},
		map[string]*pipeline.StepLog{"step-a": {StepName: "step-a", Stdout: stdout, Stderr: stderr}},
	)

	out := execLogs(t, testRunID, "step-a", "-o", "yaml", "--limit", "2")
	var result map[string]string
	if err := yaml.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("yaml unmarshal: %v", err)
	}
	wantStdout := "b\nc\n"
	if result["stdout"] != wantStdout {
		t.Errorf("stdout = %q, want %q", result["stdout"], wantStdout)
	}
	wantStderr := "q\nr\n"
	if result["stderr"] != wantStderr {
		t.Errorf("stderr = %q, want %q", result["stderr"], wantStderr)
	}
}

// TestLogsCmd_AllSteps_Table checks that --limit does not change aggregate row count.
func TestLogsCmd_AllSteps_Table(t *testing.T) {
	steps := []pipeline.StepStatus{
		{Name: "alpha", Status: "success"},
		{Name: "beta", Status: "success"},
		{Name: "gamma", Status: "success"},
	}
	logs := map[string]*pipeline.StepLog{
		"alpha": {StepName: "alpha", Stdout: "out\n", Stderr: ""},
		"beta":  {StepName: "beta", Stdout: "out\n", Stderr: ""},
		"gamma": {StepName: "gamma", Stdout: "out\n", Stderr: ""},
	}
	setupLogsFixture(t, steps, logs)

	t.Run("no limit shows all 3 steps", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "table", "--limit", "0")
		for _, name := range []string{"alpha", "beta", "gamma"} {
			if !strings.Contains(out, name) {
				t.Errorf("expected %q in unlimited output", name)
			}
		}
	})

	t.Run("limit 2 still shows all steps", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "table", "--limit", "2")
		for _, name := range []string{"alpha", "beta", "gamma"} {
			if !strings.Contains(out, name) {
				t.Errorf("expected %q in output", name)
			}
		}
	})

	t.Run("limit greater than step count shows all", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "table", "--limit", "100")
		for _, name := range []string{"alpha", "beta", "gamma"} {
			if !strings.Contains(out, name) {
				t.Errorf("expected %q with limit > step count", name)
			}
		}
	})
}

// TestLogsCmd_AllSteps_JSON checks that --limit caps each stream, not the array length.
func TestLogsCmd_AllSteps_JSON(t *testing.T) {
	steps := []pipeline.StepStatus{
		{Name: "s1", Status: "success"},
		{Name: "s2", Status: "success"},
		{Name: "s3", Status: "success"},
	}
	logs := map[string]*pipeline.StepLog{
		"s1": {StepName: "s1", Stdout: "a\nb\nc\n", Stderr: "x\ny\n"},
		"s2": {StepName: "s2", Stdout: "d\ne\nf\n", Stderr: "z\nzz\n"},
		"s3": {StepName: "s3", Stdout: "g\nh\ni\n", Stderr: "q\nr\n"},
	}
	setupLogsFixture(t, steps, logs)

	t.Run("no limit returns all 3 objects", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "json", "--limit", "0")
		var result []map[string]string
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
	})

	t.Run("limit 2 returns all 3 objects with capped streams", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "json", "--limit", "2")
		var result []map[string]string
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("json unmarshal: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
		if result[0]["stdout"] != "b\nc\n" {
			t.Errorf("stdout = %q, want %q", result[0]["stdout"], "b\nc\n")
		}
		if result[1]["stdout"] != "e\nf\n" {
			t.Errorf("stdout = %q, want %q", result[1]["stdout"], "e\nf\n")
		}
	})
}

// TestLogsCmd_AllSteps_YAML checks that --limit caps each stream, not the array length.
func TestLogsCmd_AllSteps_YAML(t *testing.T) {
	steps := []pipeline.StepStatus{
		{Name: "s1", Status: "success"},
		{Name: "s2", Status: "success"},
		{Name: "s3", Status: "success"},
	}
	logs := map[string]*pipeline.StepLog{
		"s1": {StepName: "s1", Stdout: "a\nb\nc\n", Stderr: "x\ny\nz\n"},
		"s2": {StepName: "s2", Stdout: "d\ne\nf\n", Stderr: "p\nq\nr\n"},
		"s3": {StepName: "s3", Stdout: "g\nh\ni\n", Stderr: "l\nm\nn\n"},
	}
	setupLogsFixture(t, steps, logs)

	t.Run("limit 1 returns all 3 objects with capped streams", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "yaml", "--limit", "1")
		var result []map[string]string
		if err := yaml.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("yaml unmarshal: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
		if result[0]["stdout"] != "c\n" {
			t.Errorf("stdout = %q, want %q", result[0]["stdout"], "c\n")
		}
		if result[2]["stderr"] != "n\n" {
			t.Errorf("stderr = %q, want %q", result[2]["stderr"], "n\n")
		}
	})

	t.Run("no limit returns all 3", func(t *testing.T) {
		out := execLogs(t, testRunID, "-o", "yaml", "--limit", "0")
		var result []map[string]string
		if err := yaml.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("yaml unmarshal: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
	})
}

func TestLogsCmd_InvalidLimit(t *testing.T) {
	setupLogsFixture(t,
		[]pipeline.StepStatus{{Name: "step-a", Status: "success"}},
		map[string]*pipeline.StepLog{"step-a": {StepName: "step-a", Stdout: "line\n", Stderr: ""}},
	)

	rootCmd.SetArgs([]string{"logs", testRunID, "step-a", "--limit", "-1"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for negative --limit")
	}
	if !strings.Contains(err.Error(), "invalid --limit -1") {
		t.Fatalf("unexpected error: %v", err)
	}
}
