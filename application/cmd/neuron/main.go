package main

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/internal/cli"
)

// Version is set at build time via -ldflags "-X main.Version=...".
// A development build defaults to "dev".
var Version = "dev"

func main() {
	cli.Version = Version

	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}