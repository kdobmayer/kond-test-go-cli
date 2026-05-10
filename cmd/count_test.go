package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func resetCountState() {
	countCmd.Flags().Set("unique", "false")
	outputFormat = "table"
}

func TestCountCommand_TotalLines(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\nc\n"))
	rootCmd.SetArgs([]string{"count"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count command error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "3") {
		t.Errorf("output %q does not contain count 3", got)
	}
}

func TestCountCommand_TotalWithDuplicates(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\na\n"))
	rootCmd.SetArgs([]string{"count"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count command error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "3") {
		t.Errorf("output %q does not contain total count 3", got)
	}
}

func TestCountCommand_UniqueLong(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\na\n"))
	rootCmd.SetArgs([]string{"count", "--unique"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count --unique error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "2") {
		t.Errorf("output %q does not contain unique count 2", got)
	}
}

func TestCountCommand_UniqueShort(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\na\n"))
	rootCmd.SetArgs([]string{"count", "-u"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count -u error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "2") {
		t.Errorf("output %q does not contain unique count 2", got)
	}
}

func TestCountCommand_Empty(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader(""))
	rootCmd.SetArgs([]string{"count"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count empty error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "0") {
		t.Errorf("output %q does not contain count 0", got)
	}
}

func TestCountCommand_SingleLine(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("hello\n"))
	rootCmd.SetArgs([]string{"count"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count single-line error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "1") {
		t.Errorf("output %q does not contain count 1", got)
	}
}

func TestCountCommand_JSONOutput(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\nc\n"))
	rootCmd.SetArgs([]string{"count", "--output", "json"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count --output json error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"count": 3`) {
		t.Errorf("output %q does not contain JSON count field", got)
	}
}

func TestCountCommand_YAMLOutput(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader("a\nb\nc\n"))
	rootCmd.SetArgs([]string{"count", "--output", "yaml"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("count --output yaml error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "count: 3") {
		t.Errorf("output %q does not contain YAML count field", got)
	}
}

func TestCountCommand_UnexpectedArg(t *testing.T) {
	t.Cleanup(resetCountState)

	rootCmd.SetIn(strings.NewReader(""))
	rootCmd.SetArgs([]string{"count", "somearg"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	if err := rootCmd.Execute(); err == nil {
		t.Error("expected error for unexpected positional argument")
	}
}
