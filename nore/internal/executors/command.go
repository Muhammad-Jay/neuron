package executors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"os/exec"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

type CommandExecutor struct{}

func (CommandExecutor) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	params := make(map[string]any)
	maps.Copy(params, execution.ServiceConfigurations)
	maps.Copy(params, execution.Input)

	cmdString, _ := params["command"].(string)
	if cmdString == "" {
		return nil, fmt.Errorf("command executor requires a 'command' string")
	}

	// Parse arguments safely
	var cmdArgs []string
	if argsRaw, ok := params["args"].([]any); ok {
		for _, arg := range argsRaw {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", arg))
		}
	}

	// Create command with context (this auto-kills the process if N.O.R.E times out)
	cmd := exec.CommandContext(ctx, cmdString, cmdArgs...)

	// Capture outputs
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
	}

	return map[string]any{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": float64(exitCode), // JSON numbers are float64 by default in Go
		"success":   exitCode == 0,
	}, nil
}