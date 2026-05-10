package main

import (
	"fmt"
	"os"

	"github.com/kdobmayer/kond-test-go-cli/cmd"
)

const VERSION = "0.1.0"

func main() {
	if err := cmd.Execute(VERSION); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
