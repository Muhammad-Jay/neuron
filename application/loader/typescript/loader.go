package typescript

import (
	"context"
	"os/exec"

	"github.com/Muhammad-Jay/neuron/application/process"
)

type TSLoader struct {
	process *process.Process
}

func NewTSLoader(process *process.Process) *TSLoader {
	return &TSLoader{
		process: process,
	}
}

func (l *TSLoader) Load(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "tslint", "-s", path)

	cmd.Dir = l.process.Cmd.Dir
}