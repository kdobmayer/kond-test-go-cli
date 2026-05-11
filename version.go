package main

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func formatVersion(version, commit, buildDate string) string {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if buildDate == "" {
		buildDate = "unknown"
	}

	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate)
}
