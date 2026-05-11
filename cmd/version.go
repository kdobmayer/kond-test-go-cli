package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

// SetVersion injects build-time version metadata from main.
func SetVersion(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "pipeline version %s (commit: %s, built: %s)\n", appVersion, appCommit, appDate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
