package cmd

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Print the last N lines from stdin",
	Long:  `Read all lines from stdin and print only the last N lines (default 10).`,
	Args:  cobra.NoArgs,
	RunE:  runTail,
}

func init() {
	rootCmd.AddCommand(tailCmd)
	tailCmd.Flags().IntP("lines", "n", 10, "Number of lines to print from the end")
}

func runTail(cmd *cobra.Command, args []string) error {
	n, err := cmd.Flags().GetInt("lines")
	if err != nil {
		return fmt.Errorf("getting lines flag: %w", err)
	}

	if n < 0 {
		return fmt.Errorf("invalid value for --lines: must be >= 0")
	}

	scanner := bufio.NewScanner(cmd.InOrStdin())
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	if n <= 0 {
		return nil
	}

	start := len(lines) - n
	if start < 0 {
		start = 0
	}

	out := cmd.OutOrStdout()
	for _, line := range lines[start:] {
		fmt.Fprintln(out, line)
	}
	return nil
}
