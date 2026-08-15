package executors

import (
	"context"
	"fmt"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

type AIMockExecutor struct{}

func (AIMockExecutor) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	prompt, ok := execution.ServiceConfigurations["prompt"].(string)
	if !ok || prompt == "" {
		return nil, fmt.Errorf("AI service requires a resolved prompt string")
	}
	fmt.Println("Resolved AI prompt:", prompt)
	return map[string]any{"content": "[mock AI response] " + prompt}, nil
}