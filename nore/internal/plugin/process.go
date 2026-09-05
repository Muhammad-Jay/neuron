// Package plugin adapts frozen executor artifacts into N.O.R.E.'s in-process
// Executor contract. A ResolvedExecutor is launched as a child process speaking
// the Neuron executor wire protocol (JSON request on stdin, JSON response on
// stdout). N.O.R.E. stays decoupled from the registry source and only knows the
// wire types in shared/types/executor.
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	core "github.com/Muhammad-Jay/neuron/shared/types/core"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// config stores frozen executor specifications keyed by the logical type.
type config struct {
	ResolvedExecutors []shadexec.ResolvedExecutor `json:"resolved_executors"`
}

// DecodeResolvedExecutors extracts the frozen executor set from the opaque
// ExecutionConfigurations payload stored on a RegisteredSystem. The payload is
// JSON-round-tripped so it works for both typed values and values re-read from
// disk as map[string]any.
func DecodeResolvedExecutors(payload any) ([]shadexec.ResolvedExecutor, error) {
	if payload == nil {
		return nil, nil
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal execution configurations: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(buf, &cfg); err != nil {
		return nil, fmt.Errorf("decode resolved executors: %w", err)
	}
	return cfg.ResolvedExecutors, nil
}

// RegisterProcessExecutors registers a process executor for every frozen type
// that does not already have an in-process executor (core executors win).
func RegisterProcessExecutors(reg contracts.ExecutorRegistry, resolved []shadexec.ResolvedExecutor) error {
	for _, r := range resolved {
		if _, err := reg.Resolve(core.ServiceType(r.Type)); err == nil {
			// Core in-process executor already registered; prefer it.
			continue
		}
		adapter, err := NewProcessAdapter(r)
		if err != nil {
			return fmt.Errorf("create process executor for %s: %w", r.Type, err)
		}
		if err := reg.Register(core.ServiceType(r.Type), adapter); err != nil {
			return fmt.Errorf("register process executor for %s: %w", r.Type, err)
		}
	}
	return nil
}

// ProcessAdapter runs a resolved executor as a subprocess per execution.
type ProcessAdapter struct {
	resolved shadexec.ResolvedExecutor

	unitTimeout time.Duration
}

// NewProcessAdapter builds an adapter for a single frozen executor.
func NewProcessAdapter(resolved shadexec.ResolvedExecutor) (*ProcessAdapter, error) {
	entrypoint := resolved.EntrypointPath()
	if entrypoint == "" {
		return nil, fmt.Errorf("executor %s: no entrypoint", resolved.Type)
	}
	return &ProcessAdapter{
		resolved:    resolved,
		unitTimeout: 10 * time.Minute,
	}, nil
}

// Execute spawns the process, feeds it the protocol request, and returns the
// parsed response output.
func (p *ProcessAdapter) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	entrypoint := p.resolved.EntrypointPath()

	timeout := p.unitTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, entrypoint)
	cmd.Env = append(os.Environ(),
		shadexec.EnvProtocol+"="+p.protocol(),
		shadexec.EnvType+"="+p.resolved.Type,
		shadexec.EnvVersion+"="+p.resolved.ResolvedVersion,
	)

	req := shadexec.Request{Input: execution.Input}

	stdin, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("executor %s: encode request: %w", p.resolved.Type, err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("executor %s: timed out after %s", p.resolved.Type, timeout)
		}
		return nil, fmt.Errorf("executor %s: %w: %s", p.resolved.Type, err, stderr.String())
	}

	resp, err := decodeResponse(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("executor %s: decode response: %w (stderr: %s)", p.resolved.Type, err, stderr.String())
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("executor %s: %s", p.resolved.Type, resp.Error)
	}

	return resp.Output, nil
}

func (p *ProcessAdapter) protocol() string {
	if p.resolved.Runtime.Protocol == "" {
		return shadexec.ProtocolV1
	}
	return p.resolved.Runtime.Protocol
}

func decodeResponse(r io.Reader) (*shadexec.Response, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var resp shadexec.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetUnitTimeout overrides the per-execution process timeout (for tests).
func (p *ProcessAdapter) SetUnitTimeout(d time.Duration) {
	p.unitTimeout = d
}
