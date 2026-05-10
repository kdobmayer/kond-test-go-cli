package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestConfigList_TableOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	orig := outputFormat
	outputFormat = "table"
	t.Cleanup(func() { outputFormat = orig })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"config", "list"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config list: %v", err)
	}

	out := buf.Bytes()
	if !bytes.Contains(out, []byte("KEY")) {
		t.Errorf("expected KEY header in table output:\n%s", buf.String())
	}
	if !bytes.Contains(out, []byte("pipeline_dir")) {
		t.Errorf("expected pipeline_dir in table output:\n%s", buf.String())
	}
}

func TestConfigList_JSONOutput_OutputFlag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"config", "list", "--output", "json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config list --output json: %v", err)
	}

	var pairs []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &pairs); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if len(pairs) == 0 {
		t.Fatal("expected non-empty JSON array")
	}
	if _, ok := pairs[0]["key"]; !ok {
		t.Error("expected 'key' field in JSON objects")
	}
	if _, ok := pairs[0]["value"]; !ok {
		t.Error("expected 'value' field in JSON objects")
	}
}

func TestConfigList_JSONOutput_JSONFlag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	orig := outputFormat
	outputFormat = "table"
	t.Cleanup(func() { outputFormat = orig })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"config", "list", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config list --json: %v", err)
	}

	var pairs []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &pairs); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if len(pairs) == 0 {
		t.Fatal("expected non-empty JSON array")
	}
	if _, ok := pairs[0]["key"]; !ok {
		t.Error("expected 'key' field in JSON objects")
	}
	if _, ok := pairs[0]["value"]; !ok {
		t.Error("expected 'value' field in JSON objects")
	}
}

func TestConfigList_EmptyConfig_JSONFlag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	orig := outputFormat
	outputFormat = "table"
	t.Cleanup(func() { outputFormat = orig })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"config", "list", "--json"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config list --json (no config file): %v", err)
	}

	var pairs []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &pairs); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	// DefaultConfig always has at least the 4 built-in keys.
	if len(pairs) < 4 {
		t.Errorf("expected at least 4 config keys, got %d", len(pairs))
	}
}
