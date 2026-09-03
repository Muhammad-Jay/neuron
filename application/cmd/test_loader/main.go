package main

import (
	"fmt"
	"os"

	"github.com/Muhammad-Jay/neuron/application/loader"
	"github.com/Muhammad-Jay/neuron/application/loader/typescript"
	"github.com/Muhammad-Jay/neuron/application/process"
)

func main() {
	cmd := process.Command{
		Path: "neuron-sdk",
		Args: []string{"build", "--path", "../../../examples/ecommerce_order_ts"},
		Env: append(os.Environ()),
	}

	tsLoader := typescript.NewTSLoader(cmd)

	ld := loader.New(tsLoader)

	err := ld.Handler.Build()
	if err != nil {
		fmt.Println(err)
	}
}
