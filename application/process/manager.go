package process

import (
	"os/exec"
)

func (p *Process) Run() (string, error) {
	process := exec.Command(p.Cmd.Path, p.Cmd.Args...)

	output, err := process.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
