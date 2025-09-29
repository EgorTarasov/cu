package main

import (
	"log"

	"cu-sync/internal/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
