package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/kdobmayer/kond-test-go-cli/output"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count lines from stdin",
	Long:  `Reads lines from standard input and prints the total line count. Use --unique to count distinct lines.`,
	Args:  cobra.NoArgs,
	RunE:  runCount,
}

func init() {
	rootCmd.AddCommand(countCmd)
	countCmd.Flags().BoolP("unique", "u", false, "Count only unique lines")
}

type countResult struct {
	Count int `json:"count" yaml:"count"`
}

func runCount(cmd *cobra.Command, args []string) error {
	unique, _ := cmd.Flags().GetBool("unique")

	// bufio.Reader.ReadLine handles arbitrarily long lines without a token-size limit.
	reader := bufio.NewReader(cmd.InOrStdin())
	count := 0
	var seen map[string]struct{}
	var lineBuilder strings.Builder
	if unique {
		seen = make(map[string]struct{})
	}

	for {
		chunk, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading stdin: %w", err)
		}
		if unique {
			lineBuilder.Write(chunk)
		}
		if !isPrefix {
			if unique {
				line := lineBuilder.String()
				lineBuilder.Reset()
				if _, ok := seen[line]; !ok {
					seen[line] = struct{}{}
					count++
				}
			} else {
				count++
			}
		}
	}

	f := output.NewFormatter(outputFormat, cmd.OutOrStdout())
	return f.Render(
		[]string{"COUNT"},
		[]output.TableRow{{Columns: []string{fmt.Sprintf("%d", count)}}},
		countResult{Count: count},
	)
}
