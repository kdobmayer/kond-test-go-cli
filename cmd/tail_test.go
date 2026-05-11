package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestTailCommand(t *testing.T) {
	seq20 := func() string {
		var sb strings.Builder
		for i := 1; i <= 20; i++ {
			fmt.Fprintf(&sb, "%d\n", i)
		}
		return sb.String()
	}()

	tests := []struct {
		name  string
		input string
		args  []string
		want  string
	}{
		{
			name:  "default 10 lines from 20",
			input: seq20,
			args:  []string{"tail"},
			want:  "11\n12\n13\n14\n15\n16\n17\n18\n19\n20\n",
		},
		{
			name:  "fewer lines than default",
			input: "a\nb\nc\n",
			args:  []string{"tail"},
			want:  "a\nb\nc\n",
		},
		{
			name:  "short flag -n 3",
			input: seq20,
			args:  []string{"tail", "-n", "3"},
			want:  "18\n19\n20\n",
		},
		{
			name:  "long flag --lines 3",
			input: seq20,
			args:  []string{"tail", "--lines", "3"},
			want:  "18\n19\n20\n",
		},
		{
			name:  "n zero prints nothing",
			input: "a\nb\nc\n",
			args:  []string{"tail", "-n", "0"},
			want:  "",
		},
		{
			name:  "empty stdin",
			input: "",
			args:  []string{"tail"},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				rootCmd.SetArgs(nil)
				rootCmd.SetOut(nil)
				rootCmd.SetIn(nil)
			})

			rootCmd.SetArgs(tt.args)
			var outBuf bytes.Buffer
			rootCmd.SetOut(&outBuf)
			rootCmd.SetIn(strings.NewReader(tt.input))

			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := outBuf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTailCommand_NegativeLines(t *testing.T) {
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetIn(nil)
	})

	rootCmd.SetArgs([]string{"tail", "-n", "-1"})
	rootCmd.SetIn(strings.NewReader("a\nb\nc\n"))

	var errBuf bytes.Buffer
	rootCmd.SetErr(&errBuf)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for negative line count")
	}

	if !strings.Contains(err.Error(), "must be >= 0") {
		t.Fatalf("unexpected error: %v", err)
	}
}
