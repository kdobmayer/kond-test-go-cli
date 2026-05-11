package cmd

import (
	"bytes"
	"testing"
)

func TestBarFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--bar"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	t.Cleanup(func() { rootCmd.SetOut(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "bar\n" {
		t.Errorf("--bar: got %q, want %q", got, "bar\n")
	}
}

func TestBarFlag_NotPrintedWithoutFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--foo"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	t.Cleanup(func() { rootCmd.SetOut(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got == "bar\n" {
		t.Error("--bar not set but 'bar' was printed")
	}
}
