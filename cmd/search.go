package cmd

import (
	"bufio"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

var searchPattern string

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search stdin using a regex pattern",
	Long:  `Searches stdin line by line for a given regex pattern and prints matching lines with line numbers.`,
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&searchPattern, "pattern", "p", "", "Regex pattern to search for (required)")
	searchCmd.MarkFlagRequired("pattern")
}

func runSearch(cmd *cobra.Command, args []string) error {
	re, err := regexp.Compile(searchPattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern %q: %w", searchPattern, err)
	}

	scanner := bufio.NewScanner(cmd.InOrStdin())
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if re.MatchString(line) {
			fmt.Fprintf(cmd.OutOrStdout(), "%d: %s\n", lineNumber, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stdin: %w", err)
	}
	return nil
}
