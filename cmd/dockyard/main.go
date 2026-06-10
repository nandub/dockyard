package main

import (
	"os"

	"github.com/nandub/dockyard/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
