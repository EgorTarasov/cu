package main

import (
	"fmt"
	"os"

	"github.com/EgorTarasov/cu/internal/cli/command"
	"github.com/EgorTarasov/cu/internal/version"
)

var (
	ver    = "dev"
	commit = "unknown"
	date   = "unknown"
)

func main() {
	version.Set(ver, commit, date)
	command.RootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", ver, commit, date)

	if err := command.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
