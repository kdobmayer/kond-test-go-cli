package main

import (
	"fmt"
	"os"

	"github.com/kdobmayer/kond-test-go-cli/cmd"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, buildDate)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
