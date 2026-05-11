package main

import (
	"fmt"
	"os"

	"github.com/kdobmayer/kond-test-go-cli/cmd"
)

func main() {
	cmd.SetVersion(formatVersion(Version, Commit, BuildDate))
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
