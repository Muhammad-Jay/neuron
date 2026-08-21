package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	customersystem "development/systems/customer-system"

	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	input   string
	detach  bool
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a Neuron System",
		Long:  "Run a Neuron System, and the internal N.O.R.E runtime will execute the System.",
		RunE:  runCmdHandler,
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().StringVar(&input, "input", "", "JSON input for the execution (e.g., '{\"key\":\"value\"}')")
	cmd.Flags().BoolVar(&detach, "detach", false, "return immediately after starting execution (default: wait for completion with live event stream)")

	return cmd
}

func runCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if verbose {
		fmt.Println("Running in verbose mode.")
	}

	c, cleanup, err := bootstrap.SetupClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	sys := customersystem.System.Build()

	var execInput map[string]any
	if input != "" {
		if err := json.Unmarshal([]byte(input), &execInput); err != nil {
			return fmt.Errorf("invalid --input JSON: %w", err)
		}
	} else {
		execInput = map[string]any{
			"target_url": "https://jsonplaceholder.typicode.com/todos/1",
		}
	}

	var key = protocol.InstanceKey{SystemID: string(sys.Metadata.ID)}
	mode := "wait"
	if detach {
		mode = "detach"
	}

	result, err := c.Execute(ctx, key, sys, execInput, mode)
	if err != nil {
		return err
	}

	if detach {
		fmt.Printf("Instance: %s\nExecution: %s\nStatus: %s\n", result.InstanceID, result.ExecutionID, result.Status)
		return nil
	}

	// Wait mode: stream events in real-time, then show final result
	return streamEventsAndWait(ctx, c, result.InstanceID, result.ExecutionID, result)
}

func streamEventsAndWait(ctx context.Context, c *client.Client, instanceID string, executionID core.ID, finalResult protocol.ExecutionResult) error {
	eventCh := make(chan protocol.StreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		errCh <- c.StreamExecutionEvents(ctx, instanceID, executionID, func(evt protocol.StreamEvent) error {
			eventCh <- evt
			return nil
		})
	}()

	var terminal bool
	for !terminal {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil && !strings.Contains(err.Error(), "context canceled") {
				return err
			}
			return nil
		case evt, ok := <-eventCh:
			if !ok {
				return nil
			}
			printEvent(evt)
			if evt.Type == "execution.completed" || evt.Type == "execution.failed" || evt.Type == "execution.cancelled" {
				terminal = true
			}
		}
	}

	printFinalResult(finalResult)
	return nil
}

func printEvent(evt protocol.StreamEvent) {
	ts := "n/a"
	if evt.OccurredAt > 0 {
		ts = formatTime(evt.OccurredAt)
	}

	eventType := evt.Type
	svc := ""
	if evt.ServiceID != "" {
		svc = fmt.Sprintf(" [%s]", evt.ServiceID)
	}

	payloadStr := ""
	if len(evt.Payload) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(evt.Payload, &payload); err == nil {
			if msg, ok := payload["message"].(string); ok {
				payloadStr = " " + msg
			} else if prompt, ok := payload["prompt"].(string); ok {
				payloadStr = " " + prompt
			}
		}
	}

	fmt.Printf("[%s] %s%s%s\n", ts, eventType, svc, payloadStr)
}

func printFinalResult(result protocol.ExecutionResult) {
	fmt.Printf("\nExecution completed: %s\n", result.ExecutionID)
	fmt.Printf("Status: %s\n", result.Status)
	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
		os.Exit(1)
	}
	if len(result.Outputs) > 0 {
		for svcID, out := range result.Outputs {
			fmt.Printf("Service %s output: %v\n", svcID, out)
		}
	}
}

func formatTime(unixNano int64) string {
	if unixNano <= 0 {
		return "n/a"
	}
	return time.Unix(0, unixNano).Local().Format("15:04:05.000")
}