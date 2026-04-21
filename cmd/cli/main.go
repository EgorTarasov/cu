package main

import (
	"cu-sync/internal/cli/command"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		panic(err)
	}
}
