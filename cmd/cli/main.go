package main

import (
	"cu-sync/internal/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		panic(err)
	}
}
