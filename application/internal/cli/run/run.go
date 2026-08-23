package run

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	customersystem "development/systems/customer-system"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
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
	cmd.Flags().BoolVar(&detach, "detach", true, "return immediately after starting execution (default: wait for completion with live event stream)")

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
	mode := "detach"

	result, err := c.Execute(ctx, key, sys, execInput, mode)
	if err != nil {
		return err
	}

	if !detach {
		// Wait mode: stream events in real-time, then show final result
		return streamEventsAndWait(ctx, c, result.InstanceID, result.ExecutionID)
	}

	fmt.Printf("execution ID: %s\n instance ID: %s\n status: %s\n\n", result.InstanceID, result.InstanceID, result.Status)
	return nil
}

func streamEventsAndWait(ctx context.Context, c *client.Client, instanceID string, executionID core.ID) error {
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

	return nil
}

func printEvent(evt protocol.StreamEvent) {
	ts := "n/a"
	if evt.OccurredAt > 0 {
		ts = formatTime(evt.OccurredAt)
	}

	colorReset := "\033[94m"
	colorDim := "\033[36m"
	colorSvc := "\033[35m"

	colorEvt := "\033[36m"
	if strings.HasSuffix(evt.Type, ".completed") {
		colorEvt = "\033[32m"
	} else if strings.HasSuffix(evt.Type, ".failed") {
		colorEvt = "\033[31m"
	} else if strings.HasSuffix(evt.Type, ".log") {
		colorEvt = "\033[33m"
	}

	svc := ""
	if evt.ServiceID != "" {
		svc = fmt.Sprintf(" %s[%s]%s", colorSvc, evt.ServiceID, colorReset)
	}

	// Print the event header
	fmt.Printf("\n%s[%s]%s %s%s%s%s\n", colorDim, ts, colorReset, colorEvt, evt.Type, colorReset, svc)

	if len(evt.Payload) == 0 {
		return
	}

	isLog := evt.Type == "service.log"

	// Skip rendering non-log payloads unless verbose is enabled
	if !verbose && !isLog {
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(evt.Payload, &payload); err == nil && len(payload) > 0 {
		// Special formatting specifically for service.log
		if isLog {
			level, _ := payload["Level"].(string)
			msg, _ := payload["Message"].(string)
			if level == "" {
				level = "info"
			}

			// Print the core log message neatly
			fmt.Printf("  \033[36m[%s]\033[0m %s\n", level, msg)

			// Transform the Fields array (Key/Value objects) into clean key: value rendering
			if fields, ok := payload["Fields"].([]any); ok {
				for _, f := range fields {
					if fm, isMap := f.(map[string]any); isMap {
						if k, hasKey := fm["Key"].(string); hasKey {
							printNode("  ", k, fm["Value"])
						}
					}
				}
			}
			return
		}

		// Standard payload formatting (only hits this if verbose == true)
		keys := make([]string, 0, len(payload))
		for k := range payload {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			printNode("  ", k, payload[k])
		}
	}
}

func printNode(indent string, key string, val any) {
	keyColor := "\033[94m" // Light blue
	reset := "\033[0m"

	switch v := val.(type) {
	case map[string]any:
		if len(v) == 0 {
			fmt.Printf("%s%s%s%s: {}\n", indent, keyColor, key, reset)
			return
		}
		fmt.Printf("%s%s%s%s:\n", indent, keyColor, key, reset)
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			printNode(indent+"  ", k, v[k])
		}
	case []any:
		if len(v) == 0 {
			fmt.Printf("%s%s%s%s: []\n", indent, keyColor, key, reset)
			return
		}
		fmt.Printf("%s%s%s%s:\n", indent, keyColor, key, reset)
		for i, item := range v {
			printNode(indent+"  ", fmt.Sprintf("[%d]", i), item)
		}
	case string:
		if strings.Contains(v, "\n") {
			fmt.Printf("%s%s%s%s: |\n", indent, keyColor, key, reset)
			lines := strings.Split(strings.TrimSpace(v), "\n")
			for _, line := range lines {
				if line == "" {
					fmt.Println(indent + "  ") // Preserve blank lines cleanly
				} else {
					fmt.Printf("%s  %s\n", indent, line)
				}
			}
		} else {
			fmt.Printf("%s%s%s%s: %s\n", indent, keyColor, key, reset, v)
		}
	default:
		if v == nil {
			fmt.Printf("%s%s%s%s: null\n", indent, keyColor, key, reset)
		} else {
			fmt.Printf("%s%s%s%s: %v\n", indent, keyColor, key, reset, v)
		}
	}
}

func formatTime(unixNano int64) string {
	if unixNano <= 0 {
		return "n/a"
	}
	return time.Unix(0, unixNano).Local().Format("15:04:05.000")
}