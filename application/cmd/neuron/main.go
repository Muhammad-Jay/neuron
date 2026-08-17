package main

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/internal/cli"
)

func main() {
	// Execute the root command. If it fails, exit with a non-zero status code.
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}