// Package main is the entry point for the agentctl CLI application.
package main

import (
	"fmt"
	"os"

	"github.com/ryantking/agentctl/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
