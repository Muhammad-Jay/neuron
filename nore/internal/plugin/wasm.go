package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

// sharedProcessRuntime is created once per process and reused by every
// instance. It owns the wazero.Runtime (compilation engine and the WASI host
// function bridge) plus an in-memory cache of compiled modules keyed by the
// frozen entrypoint path.
//
// Why shared: a wazero.Runtime holds the compiled-code engine and the module
// registry. Creating one per instance wastes memory and, with the compiler
// engine, recompiles native code. Compiled modules are immutable and safe to
// instantiate concurrently, so executions never serialize on this structure:
// each execution instantiates its own sandboxed module from the cached
// CompiledModule against the shared runtime.
//
// WithCloseOnContextDone makes wazero insert periodic checks so that an
// in-flight function call is interrupted (and its module closed) when the
// context passed to Call is canceled or reaches its deadline. This is how a
// runaway pure-compute WASI module gets killed cleanly on timeout.
var sharedProcessRuntime = sync.OnceValues(func() (*wasmRuntime, error) {
	rt := wazero.NewRuntimeWithConfig(
		context.Background(),
		wazero.NewRuntimeConfig().WithCloseOnContextDone(true),
	)
	if _, err := wasi_snapshot_preview1.Instantiate(context.Background(), rt); err != nil {
		_ = rt.Close(context.Background())
		return nil, fmt.Errorf("instantiate wasi: %w", err)
	}
	return &wasmRuntime{rt: rt, compiled: map[string]wazero.CompiledModule{}}, nil
})

// wasmRuntime is the process-wide wazero runtime and compiled-module cache.
type wasmRuntime struct {
	rt wazero.Runtime

	mu       sync.RWMutex // guards compiled
	compiled map[string]wazero.CompiledModule
}

// compiledModule returns the compiled module for a frozen entrypoint path,
// compiling and caching it on first use. Compiling is done once per module;
// every adapter for the same frozen module shares the returned CompiledModule.
func (w *wasmRuntime) compiledModule(ctx context.Context, entrypoint string) (wazero.CompiledModule, error) {
	w.mu.RLock()
	mod, ok := w.compiled[entrypoint]
	w.mu.RUnlock()
	if ok {
		return mod, nil
	}

	bin, err := os.ReadFile(entrypoint)
	if err != nil {
		return nil, err
	}

	compiled, err := w.rt.CompileModule(ctx, bin)
	if err != nil {
		return nil, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	// A concurrent caller may have compiled the same module while we were
	// compiling; prefer the existing cached copy and release the duplicate.
	if existing, ok := w.compiled[entrypoint]; ok {
		_ = compiled.Close(ctx)
		return existing, nil
	}
	w.compiled[entrypoint] = compiled
	return compiled, nil
}

// close releases the compiled modules and the runtime. It is only safe to
// call once all executions have finished (e.g. process shutdown).
func (w *wasmRuntime) close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for key, mod := range w.compiled {
		if err := mod.Close(ctx); err != nil {
			return fmt.Errorf("close compiled module %s: %w", key, err)
		}
		delete(w.compiled, key)
	}
	return w.rt.Close(ctx)
}

// WasmAdapter runs a resolved executor as an embedded WASI module. It speaks
// the exact same protocol as ProcessAdapter (JSON request on stdin, JSON
// response on stdout, NEURON_EXECUTOR_* env vars): an executor author compiles
// once to a native binary and a .wasm module with no protocol changes.
//
// The wazero.Runtime and the compiled module are process-global and shared by
// every adapter (see sharedProcessRuntime); each execution instantiates a
// fresh sandboxed module with its own stdin/stdout buffers, runs _start, and
// the module is torn down when the call returns or times out.
type WasmAdapter struct {
	resolved    shadexec.ResolvedExecutor
	unitTimeout time.Duration

	runtime  *wasmRuntime
	compiled wazero.CompiledModule
}

// NewWasmAdapter builds an adapter for a single frozen executor. It resolves
// the shared runtime (creating it once for the process) and compiles the
// frozen module, so it fails fast when the module is missing or invalid.
func NewWasmAdapter(resolved shadexec.ResolvedExecutor) (*WasmAdapter, error) {
	entrypoint := resolved.EntrypointPath()
	if entrypoint == "" {
		return nil, fmt.Errorf("executor %s: no entrypoint", resolved.Type)
	}

	rt, err := sharedProcessRuntime()
	if err != nil {
		return nil, fmt.Errorf("executor %s: initialize wasm runtime: %w", resolved.Type, err)
	}

	compiled, err := rt.compiledModule(context.Background(), entrypoint)
	if err != nil {
		return nil, fmt.Errorf("executor %s: compile wasm module %s: %w", resolved.Type, entrypoint, err)
	}

	return &WasmAdapter{
		resolved:    resolved,
		unitTimeout: 10 * time.Minute,
		runtime:     rt,
		compiled:    compiled,
	}, nil
}

// Execute instantiates a fresh module from the shared compiled module, feeds
// it the protocol request, and returns the parsed response output.
//
// The call is synchronous: wazero's WithCloseOnContextDone interrupts a
// still-running _start when execCtx reaches its deadline and closes the module
// automatically, so a runaway executor can never leak a goroutine or block
// forever. Concurrent executions are safe: instantiating from a compiled
// module is guarded by the runtime and each execution isolates its own module.
func (w *WasmAdapter) Execute(ctx context.Context, execution contracts.ExecutionContext) (map[string]any, error) {
	timeout := w.unitTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < timeout {
			timeout = remaining
		}
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req := shadexec.Request{Input: execution.Input}
	stdin, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("executor %s: encode request: %w", w.resolved.Type, err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	moduleConfig := wazero.NewModuleConfig().
		WithName("").
		WithStdin(bytes.NewReader(stdin)).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithArgs("neuron-executor", w.resolved.Type).
		WithEnv(shadexec.EnvProtocol, w.protocol()).
		WithEnv(shadexec.EnvType, w.resolved.Type).
		WithEnv(shadexec.EnvVersion, w.resolved.ResolvedVersion).
		WithSysWalltime().
		WithSysNanotime().
		WithSysNanosleep()

	// Instantiate without invoking start functions so we run _start ourselves
	// against execCtx: a deadline then interrupts the call and auto-closes the
	// module. Instantiation from the shared compiled module is concurrent-safe.
	mod, err := w.runtime.rt.InstantiateModule(execCtx, w.compiled, moduleConfig.WithStartFunctions())
	if err != nil {
		return nil, fmt.Errorf("executor %s: instantiate wasm module: %w", w.resolved.Type, err)
	}

	_start := mod.ExportedFunction("_start")
	if _start == nil {
		_ = mod.Close(context.Background())
		return nil, fmt.Errorf("executor %s: wasm module has no _start export", w.resolved.Type)
	}

	_, callErr := _start.Call(execCtx)

	if ctxErr := execCtx.Err(); ctxErr != nil {
		if errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, fmt.Errorf("executor %s: timed out after %s", w.resolved.Type, timeout)
		}
		return nil, fmt.Errorf("executor %s: %w", w.resolved.Type, ctxErr)
	}

	// A WASI command exits by calling proc_exit: exit code zero means success
	// and closes the module; a non-zero code is a controlled failure.
	var exitErr *sys.ExitError
	if callErr != nil && !errors.As(callErr, &exitErr) {
		return nil, fmt.Errorf("executor %s: run wasm: %w: %s", w.resolved.Type, callErr, stderr.String())
	}
	if exitErr != nil && exitErr.ExitCode() != 0 {
		return nil, fmt.Errorf("executor %s: run wasm: %w (stderr: %s)", w.resolved.Type, callErr, stderr.String())
	}

	if !mod.IsClosed() {
		if err := mod.Close(context.Background()); err != nil {
			return nil, fmt.Errorf("executor %s: close wasm module: %w", w.resolved.Type, err)
		}
	}

	resp, err := decodeResponse(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("executor %s: decode response: %w (stderr: %s)", w.resolved.Type, err, stderr.String())
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("executor %s: %s", w.resolved.Type, resp.Error)
	}

	return resp.Output, nil
}

// SetUnitTimeout overrides the per-execution timeout (for tests).
func (w *WasmAdapter) SetUnitTimeout(d time.Duration) {
	w.unitTimeout = d
}

// Close is a no-op: the wazero runtime and compiled modules are owned by the
// process and shared by every instance. They are released when the process
// exits (or shutdown closes the shared runtime explicitly). Closing an adapter
// never tears down runtimes other instances still need.
func (w *WasmAdapter) Close() error {
	return nil
}

func (w *WasmAdapter) protocol() string {
	if w.resolved.Runtime.Protocol == "" {
		return shadexec.ProtocolV1
	}
	return w.resolved.Runtime.Protocol
}