package main

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/process"
)

func main() {
	cmd := process.Command{
		Path: "podman",
		Args: []string{"--version"},
		Env: append(os.Environ()),
	}

	proc := process.NewProcess(cmd)

	output, err := proc.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Output:", output)
}