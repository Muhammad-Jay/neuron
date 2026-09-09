# Neuron Architecture

This document describes the actual architecture of Neuron as implemented in the current codebase.

---

## 1. Overview

Neuron is a runtime for composing executable capabilities into coherent software systems. It separates the description of what a system does from the mechanisms that run it.

The two primary binaries are:

- **`neuron`** — the user-facing CLI for defining, building, and managing systems.
- **`nore`** — the N.O.R.E. (Neuron Operating Runtime Environment) daemon that hosts and executes systems.

---

## 2. Core Model

Neuron is built on six primitives:

| Primitive | Role |
|-----------|------|
| **System** | A composition of services and their relationships. |
| **Service** | An executable capability — what can be done. |
| **Connector** | How services communicate — field mappings, guard conditions. |
| **Executor** | The mechanism that actually runs a service (process, WASM, etc.). |
| **Instance** | A living realization of a System with its own state and activity. |
| **N.O.R.E.** | The runtime environment where instances exist and operate. |

These are deliberately independent. A service does not need to be a function. A connector does not need to be an HTTP request. An executor does not need to be written in the same language as the system using it.

---

## 3. Repository Architecture

```
neuron/
├── application/        Go module — CLI, compiler, loader, executor resolution
│   ├── cmd/neuron/     CLI entry point
│   ├── compiler/       Manifest compilation
│   ├── loader/         YAML and TypeScript loaders
│   ├── build/          Language-specific builders
│   ├── executor/       Executor resolution, installation, verification
│   ├── config/         Configuration loading
│   ├── connection/     Local and remote client connections
│   ├── daemon/         Daemon lifecycle management
│   ├── internal/cli/   Cobra CLI commands
│   ├── process/        OS process management
│   ├── project/        Project model and validation
│   ├── runtime/        Runtime management
│   └── sdk/            Go SDK for programmatic system definition
│
├── nore/               Go module — N.O.R.E. runtime engine
│   ├── cmd/nore/       Daemon entry point
│   └── internal/
│       ├── api/        HTTP API server (TCP + Unix socket)
│       ├── execution/  Execution engine
│       ├── executors/  Core in-process executors
│       ├── instance/   Instance lifecycle management
│       ├── planner/    Execution plan compilation
│       ├── plugin/     Executor adapter boundary
│       ├── registry/   Core executor registry
│       ├── resolver/   CEL expression compiler
│       ├── runtime/    Runtime backends (process, WASM)
│       ├── storage/    Persistence (SQLite)
│       ├── system/     System repository
│       └── stream/     Streaming support
│
├── shared/             Go module — dependency-free types and protocol
│   ├── types/core/     System, Service, Connector types
│   ├── types/executor/ Executor manifest, protocol, resolved types
│   ├── types/protocol/ Instance, registration, execution protocol types
│   └── protocol/executor/v1/  gRPC protobuf definitions
│
├── packages/
│   ├── sdk/            TypeScript SDK (@neuron/sdk) — definition layer
│   └── executor-go/    Go SDK for writing executors
│
├── examples/           Example projects (YAML, TypeScript, Go)
└── docs/               Architecture and runtime documentation
```

---

## 4. Canonical Data Flow

The intended data flow through the system:

```
Source Language (YAML / TypeScript / Go)
        ↓
Language-Specific Loader / Builder
        ↓
Canonical System Manifest (JSON)
        ↓
Validation
        ↓
Compiler (application/compiler)
        ↓
Core System (shared/types/core)
        ↓
Execution Plan (nore/internal/planner)
        ↓
N.O.R.E. Runtime
        ↓
Instance
        ↓
Service Executor
        ↓
Result
```

YAML, TypeScript, and Go source languages all converge on the canonical manifest representation before entering source-language-independent compiler logic.

---

## 5. Application Layer

The `application` module is the user-facing half of Neuron. It owns:

### CLI (`application/internal/cli/`)

A Cobra-based CLI with subcommands:

| Command | Purpose |
|---------|---------|
| `neuron init` | Initialize a new project |
| `neuron register` | Build, compile, resolve executors, and register a system with N.O.R.E. |
| `neuron run` | Execute a system |
| `neuron instance` | List, remove, or clear instances |
| `neuron execution` | Execution management |
| `neuron daemon` | Start/stop the N.O.R.E. daemon |
| `neuron executor` | List or inspect executors |
| `neuron add` | Add an executor |
| `neuron remove` | Remove an executor |

### Compiler (`application/compiler/`)

Transforms the canonical manifest into core system structures. The compiler is source-language agnostic — it operates on the canonical representation, not on YAML or TypeScript directly.

### Loader (`application/loader/`)

Loads project definitions from YAML or TypeScript source files and produces the canonical manifest. YAML and TypeScript have separate loader implementations.

### Executor Resolution (`application/executor/`)

A four-stage pipeline:

1. **Requirement** — the project declares what executor it needs (logical name, optional version constraint, optional registry).
2. **Resolution** — the resolver finds a matching package from a registry (GitHub Releases or local directory).
3. **Installation** — the installer downloads, verifies (SHA-256), and stores the artifact.
4. **Freezing** — at registration time, exact resolutions are frozen into the deployment configuration.

---

## 6. N.O.R.E. Runtime

N.O.R.E. is the runtime engine. It is responsible for:

- Hosting the HTTP API (TCP on `:7432`, Unix socket at `~/.neuron/nore.sock`).
- Managing system registrations.
- Creating and managing instances.
- Executing services through executor backends.
- Persisting state to SQLite.

### Entry Point (`nore/cmd/nore/main.go`)

The daemon initializes storage, system repository, CEL compiler, planner, instance manager, and API server, then listens on configured addresses.

### API (`nore/internal/api/`)

An HTTP server with routes for system registration, instance management, execution, and health checks.

### Instance Management (`nore/internal/instance/`)

Instances are living realizations of systems. The instance manager handles creation, lifecycle, and teardown. Each instance runs five concurrent loops: analytics, scheduler, executor engine, event persistence, and execution persistence.

---

## 7. Executor Architecture

### Core Executors

Built into N.O.R.E. as in-process implementations in `nore/internal/executors/`:

| Executor | Purpose |
|----------|---------|
| `set` | Static value setting |
| `ai` | AI/LLM operations |
| `log` | Logging |
| `http` | HTTP requests |
| `delay` | Time delays |
| `command` | Shell command execution |

Core executors always take priority over external executors of the same type.

### External Executors

External executors are native binaries or WASM modules that speak the Neuron executor wire protocol. They are resolved, installed, and verified by the application layer, then hosted by N.O.R.E. at runtime.

### Executor Contract (`shared/types/executor/`)

The dependency-free wire contract:

```go
type Runtime interface {
    Start(ctx context.Context, spec StartSpec) (Instance, error)
    RuntimeName() string
}

type Instance interface {
    Execute(ctx context.Context, req *Request) (*Response, error)
    Health(ctx context.Context) error
    Close(ctx context.Context) error
}
```

Request/Response types:

```go
type Request struct {
    Input map[string]any
}

type Response struct {
    Output map[string]any
    Error  string
}
```

### Runtime Backends (`nore/internal/runtime/`)

| Backend | Transport | Use Case |
|---------|-----------|----------|
| **Process** (gRPC) | `neuron/executor-v1` | Long-lived worker pool over Unix sockets |
| **Process** (JSON) | `neuron/executor-v1-json` | One-shot stdin/stdout JSON |
| **WASM** | `neuron/executor-v1-json` | WebAssembly modules via wazero |

The `container` and `remote` runtime kinds are declared but not yet implemented.

### Adapter Boundary (`nore/internal/plugin/`)

The plugin adapter reads frozen `ResolvedExecutor` records and dispatches to the appropriate runtime backend. It bridges the executor resolution pipeline (application side) with the runtime execution (N.O.R.E. side).

---

## 8. gRPC Protocol (`shared/protocol/executor/v1/`)

The canonical executor protocol is defined in protobuf:

```protobuf
service ExecutorService {
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Execute(ExecuteRequest) returns (ExecuteResponse);
    rpc Health(HealthRequest) returns (HealthResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
}
```

This protocol is language-independent. Custom executors can be implemented in any language that supports gRPC.

---

## 9. TypeScript SDK (`packages/sdk/`)

The TypeScript SDK is a **definition layer**, not a runtime. It provides a typed API for describing systems:

```ts
const manifest = System({ name: "my-system", version: "1.0.0" })
  .run(pipeline)
  .toManifest();
```

The SDK produces a JSON-serializable `SystemManifest` that the Go-side compiler and N.O.R.E. runtime consume. It is published as `@neuron/sdk` on npm.

---

## 10. CLI to N.O.R.E. Relationship

```
neuron CLI
    │
    ├── project/compiler     Build and compile system definitions
    ├── executor resolution  Resolve, download, verify executors
    └── N.O.R.E. client      Register systems, manage instances
             │
             ▼
         N.O.R.E. daemon
             │
             ├── API server (TCP + Unix socket)
             ├── Instance manager
             ├── Executor runtime backends
             └── Storage (SQLite)
```

The CLI and daemon are separate processes. The CLI communicates with the daemon over TCP or a local Unix socket. The CLI can also start/stop the daemon via `neuron daemon`.

---

## 11. Persistence

N.O.R.E. uses SQLite for persistence:

- System registrations
- Instance state
- Execution history
- Event storage

The data directory defaults to `~/.neuron/nore/` and is configurable via `--data-dir`.

---

## 12. Current Limitations

The following are known limitations as of v0.1.0:

- **No container runtime** — `container` executor kind is declared but not implemented.
- **No remote runtime** — `remote` executor kind is declared but not implemented.
- **No self-update** — the `neuron update` command does not exist yet.
- **No package-manager distribution** — Homebrew, WinGet, and similar are not yet configured.
- **No release signing** — release artifacts are checksummed but not cryptographically signed.
- **No distributed execution** — all executors run locally.
- **Limited executor ecosystem** — only core executors and the example echo executor are available.
