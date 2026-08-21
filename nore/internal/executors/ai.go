package executors

import (
	"context"
	"fmt"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
)

// AIMockExecutor simulates an AI service: it resolves a "prompt" from the
// service configuration and returns a canned response. In a real deployment
// this would be replaced by an executor that calls an LLM provider.
type AIMockExecutor struct{}

// Execute resolves the configured prompt and emits it as a ServiceLog before
// returning a mock response.
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
	execution.Logger.Info(ctx, "resolved AI prompt", contracts.F("prompt", prompt))

	// Simulate AI processing delay
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	return map[string]any{"content": "[mock AI response] " + prompt}, nil
}
