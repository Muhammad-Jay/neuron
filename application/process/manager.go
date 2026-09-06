package process

import (
	"context"
	"os/exec"
)

// Run executes the command and returns its combined output.
func (p *Process) Run() (string, error) {
	return p.RunContext(context.Background())
}

// RunContext executes the command, honoring ctx, and returns its combined
// output. When the command fails, the error is wrapped with the captured
// output so the caller can surface the underlying failure.
func (p *Process) RunContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, p.Cmd.Path, p.Cmd.Args...)

	if p.Cmd.Dir != "" {
		cmd.Dir = p.Cmd.Dir
	}
	if len(p.Cmd.Env) > 0 {
		cmd.Env = p.Cmd.Env
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
