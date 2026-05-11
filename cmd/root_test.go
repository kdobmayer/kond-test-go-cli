package cmd

import (
	"bytes"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	originalArgs := rootCmd.Flags().Args()
	originalOut := rootCmd.OutOrStdout()
	originalErr := rootCmd.ErrOrStderr()
	originalVersion := rootCmd.Version

	SetVersion("0.1.0")
	t.Cleanup(func() {
		rootCmd.SetArgs(originalArgs)
		rootCmd.SetOut(originalOut)
		rootCmd.SetErr(originalErr)
		rootCmd.Version = originalVersion
	})

	rootCmd.SetArgs([]string{"--version"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Cobra handles --version by printing and returning nil
	rootCmd.Execute() //nolint:errcheck

	if !bytes.Contains(buf.Bytes(), []byte("0.1.0")) {
		t.Errorf("expected version string in output, got: %q", buf.String())
	}
}

func TestSetVersion_EmptyFallsBackToDefault(t *testing.T) {
	originalVersion := rootCmd.Version
	t.Cleanup(func() {
		rootCmd.Version = originalVersion
	})

	SetVersion("")

	if rootCmd.Version != defaultVersion {
		t.Fatalf("rootCmd.Version = %q, want %q", rootCmd.Version, defaultVersion)
	}
}
