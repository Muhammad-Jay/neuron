package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Muhammad-Jay/neuron/application/client"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/bootstrap"
	"github.com/Muhammad-Jay/neuron/application/internal/cli/command"
	"github.com/Muhammad-Jay/neuron/application/project"
	"github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
	"github.com/spf13/cobra"
)

const (
	colorReset     = "\033[0m"
	colorTimestamp = "\033[36m" // Cyan
	colorKey       = "\033[94m" // Light blue
)

var (
	verbose bool
	input   string
	detach  bool
)

// New constructs and configures the Cobra command for executing Neuron systems.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Run,
		Short: "Run a Neuron System",
		Long:  "Run a Neuron System using the internal N.O.R.E runtime execution engine.",
		RunE:  runCmdHandler,
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output to display event payloads")
	cmd.Flags().StringVar(&input, "input", "", "JSON input payload for execution (e.g., '{\"key\":\"value\"}')")
	cmd.Flags().BoolVar(&detach, "detach", false, "Return execution handles immediately without streaming live events")

	return cmd
}

// runCmdHandler loads the registered system key and triggers execution. It
// deliberately does not build or parse the project: that is the responsibility
// of `neuron register`. Running requires a prior registration.
func runCmdHandler(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if verbose {
		fmt.Println("Running in verbose mode.")
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	var key protocol.InstanceKey
	if err := project.LoadRegistrationKey(root, &key); err != nil {
		if errors.Is(err, project.ErrNotRegistered) {
			return fmt.Errorf("project is not registered; run `neuron register` first")
		}
		return fmt.Errorf("load registration: %w", err)
	}

	c, cleanup, err := bootstrap.SetupClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	var execInput map[string]any
	if input != "" {
		if err := json.Unmarshal([]byte(input), &execInput); err != nil {
			return fmt.Errorf("invalid --input JSON: %w", err)
		}
	} else {
		execInput = map[string]any{}
	}

	// `neuron run` command run in detach mode by default to render event logs,
	//  and hide logs if `--detach`
	mode := core.ExecutionModeDetach

	execResult, err := c.ExecuteByKey(ctx, key, execInput, mode)
	if err != nil {
		return err
	}

	if !detach {
		return streamEventsAndWait(ctx, c, execResult.InstanceID, execResult.ExecutionID)
	}

	ts := formatTime(time.Now().UnixNano())
	fmt.Printf("%s[%s]%s execution started\n", colorTimestamp, ts, colorReset)

	printNode("  ", "execution_id", string(execResult.ExecutionID))
	printNode("  ", "instance_id", execResult.InstanceID)
	printNode("  ", "status", execResult.Status)

	return nil
}

// streamEventsAndWait initiates a stream channel to capture execution events in real time until terminal status.
func streamEventsAndWait(ctx context.Context, c *client.Client, instanceID string, executionID core.ID) error {
	eventCh := make(chan protocol.StreamEvent, 64)
	errCh := make(chan error, 1)

	go func() {
		errCh <- c.StreamExecutionEvents(ctx, instanceID, executionID, func(evt protocol.StreamEvent) error {
			eventCh <- evt
			return nil
		})
	}()

	for {
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
				return nil
			}
		}
	}
}

// printEvent handles formatting and output for live stream events.
func printEvent(evt protocol.StreamEvent) {
	ts := formatTime(evt.OccurredAt)

	svc := ""
	if evt.ServiceID != "" {
		svc = fmt.Sprintf(" [%s]", evt.ServiceID)
	}

	if evt.Type == "service.log" {
		level := "info"
		msg := ""

		if len(evt.Payload) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(evt.Payload, &payload); err == nil {
				if l, ok := payload["Level"].(string); ok && l != "" {
					level = l
				}
				if m, ok := payload["Message"].(string); ok {
					msg = m
				}
			}
		}

		fmt.Printf("%s[%s]%s %s%s %s[%s]%s %s\n", colorTimestamp, ts, colorReset, evt.Type, svc, colorTimestamp, level, colorReset, msg)
		return
	}

	// Print standardized standard lifecycle event
	fmt.Printf("%s[%s]%s %s%s\n", colorTimestamp, ts, colorReset, evt.Type, svc)

	// Render non-log payloads only when verbose mode is enabled
	if !verbose || len(evt.Payload) == 0 {
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(evt.Payload, &payload); err == nil && len(payload) > 0 {
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

// printNode recursively renders key-value data structures into formatted tree branches.
func printNode(indent string, key string, val any) {
	switch v := val.(type) {
	case map[string]any:
		if len(v) == 0 {
			fmt.Printf("%s%s%s%s: {}\n", indent, colorKey, key, colorReset)
			return
		}
		fmt.Printf("%s%s%s%s:\n", indent, colorKey, key, colorReset)
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
			fmt.Printf("%s%s%s%s: []\n", indent, colorKey, key, colorReset)
			return
		}
		fmt.Printf("%s%s%s%s:\n", indent, colorKey, key, colorReset)
		for i, item := range v {
			printNode(indent+"  ", fmt.Sprintf("[%d]", i), item)
		}
	case string:
		if strings.Contains(v, "\n") {
			fmt.Printf("%s%s%s%s: |\n", indent, colorKey, key, colorReset)
			lines := strings.Split(strings.TrimSpace(v), "\n")
			for _, line := range lines {
				if line == "" {
					fmt.Println(indent + "  ")
				} else {
					fmt.Printf("%s  %s\n", indent, line)
				}
			}
		} else {
			fmt.Printf("%s%s%s%s: %s\n", indent, colorKey, key, colorReset, v)
		}
	default:
		if v == nil {
			fmt.Printf("%s%s%s%s: null\n", indent, colorKey, key, colorReset)
		} else {
			fmt.Printf("%s%s%s%s: %v\n", indent, colorKey, key, colorReset, v)
		}
	}
}

// formatTime converts a unix nanosecond timestamp into a readable localized format.
func formatTime(unixNano int64) string {
	if unixNano <= 0 {
		return "n/a"
	}
	return time.Unix(0, unixNano).Local().Format("15:04:05.000")
}

// ExecuteResponse represents the result of triggering a system execution.
type ExecuteResponse struct {
	ExecutionID core.ID   `json:"execution_id"`
	InstanceID  string    `json:"instance_id"`
	Status      string    `json:"status"`
	Time        time.Time `json:"time"`
}