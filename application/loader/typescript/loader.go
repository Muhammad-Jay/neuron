package typescript

import (
	"github.com/Muhammad-Jay/neuron/application/process"
)

type TSLoader struct {
	process *process.Process
}

func NewTSLoader(cmd process.Command) *TSLoader {
	return &TSLoader{
		process: process.NewProcess(cmd),
	}
}

func (l *TSLoader) Build() error {
	if _, err := l.process.Run(); err != nil {
		return err
	}
	return nil
}