package main

import (
	"cu-sync/internal/cli/command"
	"cu-sync/internal/version"
)

var (
	ver    = "dev"
	commit = "unknown"
	date   = "unknown"
)

func main() {
	version.Set(ver, commit, date)

	if err := command.RootCmd.Execute(); err != nil {
		panic(err)
	}
}
