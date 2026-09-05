package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Muhammad-Jay/neuron/application/executor"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// ProcessConfig tunes the process adapter.
type ProcessConfig struct {
	// Timeout bounds every execution when the caller context has no deadline.
	Timeout time.Duration
}

// ProcessExecutor spawns an installed executor binary per execution, speaking
// the neuron/executor-v1 wire protocol over stdin/stdout. Spawn-per-execution
// keeps instances isolated and makes resource management trivial; persistent
// worker processes can be added later without changing the contract.
type ProcessExecutor struct {
	typ        string
	version    string
	entrypoint string
	protocol   string
	timeout    time.Duration
}

func (p *ProcessExecutor) Type() string    { return p.typ }
func (p *ProcessExecutor) Version() string { return p.version }

func (p *ProcessExecutor) Execute(ctx context.Context, req *shadexec.Request) (*shadexec.Response, error) {
	if req == nil {
		req = &shadexec.Request{}
	}

	if p.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode executor request: %w", err)
	}

	cmd := exec.CommandContext(ctx, p.entrypoint)
	cmd.Stdin = bytes.NewReader(payload)

	cmd.Env = append(os.Environ(),
		shadexec.EnvProtocol+"="+p.protocol,
		shadexec.EnvType+"="+p.typ,
		shadexec.EnvVersion+"="+p.version,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	// A response may accompany a non-zero exit; prefer parsing it so a
	// controlled failure reported via Response.Error is surfaced properly.
	var resp shadexec.Response
	if out := bytes.TrimSpace(stdout.Bytes()); len(out) > 0 {
		if err := json.Unmarshal(out, &resp); err != nil {
			if runErr != nil {
				return nil, fmt.Errorf("executor %s@%s: %w: %s", p.typ, p.version, runErr, stderr.String())
			}
			return nil, fmt.Errorf("executor %s@%s: decode response: %w", p.typ, p.version, err)
		}
	}

	if runErr != nil {
		// A valid response with an Error message is a controlled failure;
		// surface it as the error rather than the raw exit code.
		if resp.Error != "" {
			return &resp, nil
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("executor %s@%s: timed out", p.typ, p.version)
		}
		return nil, fmt.Errorf("executor %s@%s: %w: %s", p.typ, p.version, runErr, strings.TrimSpace(stderr.String()))
	}

	if resp.Error != "" {
		return &resp, nil
	}

	return &resp, nil
}

// NewProcessFactory returns a Factory that materializes installed executors
// whose runtime kind is process.
func NewProcessFactory(cfg *ProcessConfig) Factory {
	timeout := time.Duration(0)
	if cfg != nil {
		timeout = cfg.Timeout
	}

	return func(ctx context.Context, installed *executor.Installed) (Executor, error) {
		entrypoint := installedEntrypoint(installed)
		if entrypoint == "" {
			return nil, fmt.Errorf("executor %s@%s: no entrypoint", installed.Type, installed.Version)
		}
		info, err := os.Stat(entrypoint)
		if err != nil {
			return nil, fmt.Errorf("executor %s@%s: entrypoint %s is not materialized", installed.Type, installed.Version, entrypoint)
		}
		if info.IsDir() || info.Mode()&0o111 == 0 {
			return nil, fmt.Errorf("executor %s@%s: entrypoint %s is not executable", installed.Type, installed.Version, entrypoint)
		}

		protocol := installed.Runtime.Protocol
		if protocol == "" {
			protocol = shadexec.ProtocolV1
		}

		return &ProcessExecutor{
			typ:        installed.Type,
			version:    installed.Version,
			entrypoint: entrypoint,
			protocol:   protocol,
			timeout:    timeout,
		}, nil
	}
}

// installedEntrypoint resolves the absolute entrypoint for an installed
// executor, handling both manifest-relative and store-relative paths.
func installedEntrypoint(i *executor.Installed) string {
	if i.Runtime.Entrypoint != "" {
		p := i.Runtime.Entrypoint
		if strings.HasPrefix(p, "/") {
			return p
		}
		if i.RootDir != "" {
			return i.RootDir + "/" + p
		}
		return p
	}
	return i.ArtifactPath
}
