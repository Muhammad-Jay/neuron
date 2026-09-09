package executor

import "context"

// Runtime kinds identify how a frozen executor artifact is launched. The value
// stored in Manifest.Runtime.Type and RuntimeInfo.Type is one of these
// constants. The runtime registry dispatches on it to select the backend.
const (
	// RuntimeKindProcess runs the entrypoint as an OS child process
	// communicating over gRPC via Unix domain sockets.
	RuntimeKindProcess = "process"

	// RuntimeKindWasm runs the entrypoint inside an embedded WASI runtime,
	// speaking the stdin/stdout JSON protocol.
	RuntimeKindWasm = "wasm"

	// RuntimeKindContainer runs the entrypoint inside an OCI container for
	// strong isolation. Reserved for future implementation.
	RuntimeKindContainer = "container"

	// RuntimeKindRemote runs the entrypoint on a remote executor host.
	// Reserved for future implementation.
	RuntimeKindRemote = "remote"
)

// SupportedRuntimeKinds returns the runtime kinds the plugin layer can launch.
func SupportedRuntimeKinds() []string {
	return []string{RuntimeKindProcess, RuntimeKindWasm}
}

// Runtime starts executor instances and manages their lifecycle. Each runtime
// backend (process, wasm, container, remote) implements this interface. The
// runtime registry dispatches to the correct backend based on the executor
// manifest's runtime.type field.
//
// Implementations must be safe for concurrent use. A single Runtime instance
// may be shared across multiple executor types of the same kind.
type Runtime interface {
	// Start launches an executor instance for the given specification. The
	// returned Instance is ready to accept Execute calls. The runtime owns
	// the instance lifecycle; the caller must call Close when done.
	Start(ctx context.Context, spec StartSpec) (Instance, error)

	// RuntimeName returns the kind identifier (e.g. "process", "wasm").
	// This matches the runtime.type declared in executor manifests.
	RuntimeName() string
}

// StartSpec describes what to launch. It is derived from the frozen
// ResolvedExecutor record persisted with a registered system.
type StartSpec struct {
	// Type is the logical executor name (e.g. "github:read").
	Type string

	// Version is the exact resolved version.
	Version string

	// Protocol is the wire protocol the executor speaks
	// (e.g. "neuron/executor-v1").
	Protocol string

	// Entrypoint is the absolute path to the binary or wasm module.
	Entrypoint string

	// RootDir is the absolute path of the installed executor directory.
	RootDir string

	// MaxWorkers is the maximum number of concurrent workers for this
	// executor type. The runtime may start fewer workers based on demand.
	// A value of 0 means the runtime uses its own default.
	MaxWorkers int

	// Config carries executor-specific configuration that the runtime
	// backend may forward to the executor process or module.
	Config map[string]any
}

// Instance represents a running executor backend. It provides the uniform
// execution contract regardless of whether the underlying executor is a
// native process, WASM module, container, or remote service.
//
// Instances are not safe for concurrent Execute calls unless the runtime
// backend explicitly documents concurrent safety. The runtime backend is
// responsible for multiplexing requests across workers.
type Instance interface {
	// Execute runs one request through the executor and returns the result.
	// The context controls cancellation and deadline propagation.
	Execute(ctx context.Context, req *Request) (*Response, error)

	// Health reports whether the instance is ready to accept requests.
	// A healthy instance has at least one worker available and connected.
	Health(ctx context.Context) error

	// Close gracefully shuts down the instance, draining any in-flight
	// requests. After Close returns, the instance must not be reused.
	Close(ctx context.Context) error
}
