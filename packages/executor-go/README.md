# executor-go

`executor-go` is the official Go SDK for building Neuron executors. It lets you
implement a capability as a plain Go `Handler` and leave transport details to
the SDK: protocol negotiation, Unix domain socket setup, readiness signaling,
and lifecycle management are handled for you.

An executor is a program that exposes the Neuron execution contract. It receives
an input map and returns an output map (or a controlled error). The SDK makes
sure the same binary works in both execution modes N.O.R.E. supports:

- as a long-lived gRPC worker speaking `neuron/executor-v1` over a Unix domain
  socket, and
- as a one-shot command speaking `neuron/executor-v1-json` on stdin/stdout.

You never select the mode yourself. The SDK does it at startup based on the
environment the runtime launches you with.

Write your executor once, declare it once (`executor.json`), and the same source
can serve both the `process` and `wasm` runtime backends — see [One source, two
runtimes](#one-source-two-runtimes).

## Module

```text
github.com/Muhammad-Jay/neuron/packages/executor-go
```

Requirements:

- Go 1.26 or newer
- `google.golang.org/grpc`
- the shared Neuron contract module `github.com/Muhammad-Jay/neuron/shared`

## The Handler contract

An executor is implemented as an `executor.Handler`. It has four callbacks:

```go
type Handler struct {
    Initialize func(ctx context.Context, protocol string, metadata map[string]string) (*InitializeResult, error)
    Execute    func(ctx context.Context, input map[string]any) (map[string]any, error)
    Health     func(ctx context.Context) error
    Shutdown   func(ctx context.Context) error
}
```

- `Initialize` is called once per process, before anything else. It negotiates
  the protocol version and returns the executor's identity and capabilities.
  Return the protocol version you support in `InitializeResult.ProtocolVersion`; a
  known protocol is mirrored back, and when left empty the SDK fills in the
  canonical value (`neuron/executor-v1`).
- `Execute` runs one execution. The input is the resolved execution input for
  the service; the returned map is the output. Returning a Go error records a
  controlled failure — the runtime sees a failed execution, not a dead process.
- `Health` reports whether the executor can accept requests. If it is nil, the
  executor is assumed healthy.
- `Shutdown` is called before the process terminates, so you can release
  resources. If it is nil, the executor terminates immediately.

`InitializeResult` carries the executor's side of the negotiation:

```go
type InitializeResult struct {
    ProtocolVersion string
    Capabilities    []string
    Metadata        map[string]string
}
```

`Serve` requires both `Initialize` and `Execute`; a handler missing either is
rejected at startup.

## Starting the executor

```go
func main() {
    if err := executor.Serve(executor.Handler{
        Initialize: func(ctx context.Context, protocol string, metadata map[string]string) (*executor.InitializeResult, error) {
            return &executor.InitializeResult{
                ProtocolVersion: "neuron/executor-v1",
                Capabilities:    []string{},
                Metadata:        map[string]string{},
            }, nil
        },
        Execute: func(ctx context.Context, input map[string]any) (map[string]any, error) {
            return input, nil
        },
    }); err != nil {
        log.Fatal(err)
    }
}
```

`Serve` blocks until the executor is shut down. It returns nil after a graceful
stop (or in the one-shot JSON mode after one request) and an error otherwise.

## Transport selection

The SDK picks the transport at startup by inspecting the environment it was
launched with:

| Variable                 | When set / meaning                                                                                    |
| ------------------------ | ----------------------------------------------------------------------------------------------------- |
| `NEURON_EXECUTOR_SOCKET` | Path of the Unix domain socket. If set, the executor starts a gRPC server listening there.            |
| `NEURON_EXECUTOR_READY`  | Path of a file to create once the server is accepting connections (the launcher waits for this file). |
| `NEURON_EXECUTOR_PROTOCOL` | The execution protocol the runtime expects (informational; read via `os.Getenv`).                   |
| `NEURON_EXECUTOR_TYPE`   | The logical executor type, e.g. `example:echo` (informational).                                       |
| `NEURON_EXECUTOR_VERSION` | The resolved executor version (informational).                                                       |

Decisions made from this environment:

- If `NEURON_EXECUTOR_SOCKET` is set, the executor runs as a gRPC worker on the
  Unix domain socket at that path and signals readiness by creating the file at
  `NEURON_EXECUTOR_READY` when the server is accepting connections.
- If `NEURON_EXECUTOR_SOCKET` is not set, the executor runs as a one-shot
  stdin/stdout command: it reads a single JSON request from stdin, writes a
  single JSON response to stdout, and exits. This is the mode a WASI module (and
  any legacy command) always uses, since WASI has no socket interface.

The SDK cleans up a stale socket file from a crashed process before listening,
so a restart never fails because an old socket file is still present.

## Data conversion

Inputs and outputs are dynamic maps (`map[string]any`). The SDK converts them to
and from the protobuf `Value` type used on the gRPC transport. Supported Go
types are preserved exactly:

- `nil`, `bool`
- numbers: `float64`, `float32`, `int`, `int32`, `int64`, `uint`, `uint32`,
  `uint64`
- `string`, `[]byte` (as string), `time.Time` (as RFC 3339)
- `[]any` and `map[string]any`, recursively
- anything else falls back to a string rendering

On the way back, numeric protobuf values that are whole numbers are returned as
`int64`; fractional values return as `float64`.

## One source, two runtimes

Because the SDK selects its transport from the environment, the same Go source
compiles into both runtime backends N.O.R.E. supports:

```text
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o capability .
GOOS=wasip1 GOARCH=wasm go build -o capability.wasm .
```

Two manifests, one pair of binaries. The process binary declares the gRPC
protocol (`neuron/executor-v1`) under the host's platform key; the WASM module
declares the JSON protocol (`neuron/executor-v1-json`) under the portable
`wasm32-wasi` key:

```json
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": { "name": "my:capability", "version": "1.2.0" },
  "runtime": { "type": "process", "entrypoint": "capability", "protocol": "neuron/executor-v1" },
  "services": ["capability"],
  "capabilities": [],
  "platforms": { "linux-amd64": { "artifact": "capability", "sha256": "..." } }
}
```

```json
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": { "name": "my:capability", "version": "1.2.0" },
  "runtime": { "type": "wasm", "entrypoint": "capability.wasm", "protocol": "neuron/executor-v1-json" },
  "services": ["capability"],
  "capabilities": [],
  "platforms": { "wasm32-wasi": { "artifact": "capability.wasm" } }
}
```

To distribute both artifacts as one versioned unit, pack each pair into a
canonical executor package archive (`<name>-<version>-executor.neuron.tar.gz`:
`executor.json` plus the platform artifact at the archive root). Registries
prefer the archive over per-platform assets and reconcile the inner manifest at
install time.

## Smoke-testing locally

You do not need N.O.R.E. to exercise the JSON path. A one-shot executor reads a
single `{"input": {...}}` document on stdin and writes `{"output": {...}}` to
stdout:

```text
printf '{"input":{"who":"world"}}' | ./capability
{"output":{"echo":{"who":"world"}}}
```

## A minimal complete executor

```go
package main

import (
    "context"
    "log"

    executor "github.com/Muhammad-Jay/neuron/packages/executor-go"
)

func main() {
    err := executor.Serve(executor.Handler{
        Initialize: func(ctx context.Context, _ string, _ map[string]string) (*executor.InitializeResult, error) {
            return &executor.InitializeResult{ProtocolVersion: "neuron/executor-v1"}, nil
        },
        Execute: func(ctx context.Context, input map[string]any) (map[string]any, error) {
            return map[string]any{"echo": input}, nil
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Working without the SDK

The SDK is a convenience, not a requirement. Executors may also be written by
hand against either of the two declared protocols. You do not need Go at all:

- For the gRPC transport, implement `ExecutorService` from the proto schema in
  `shared/protocol/executor/v1/executor.proto` in any language with gRPC
  support.
- For the JSON transport, read one JSON request from stdin, write one JSON
  response to stdout, and exit.

See the runtime documentation for the full contract and lifecycle.

## Related documentation

- `docs/RUNTIME.md` — how N.O.R.E. executes services end to end
- `docs/RUNTIME_PROCESS.md` — the process runtime and its two transports
- `docs/RUNTIME_WASM.md` — the WASM runtime
- `docs/executors/2026-09-05-executor-registry-and-resolution.md` — packages, resolution, and the package archive
- `examples/executors/` — reference executor sources and its build script