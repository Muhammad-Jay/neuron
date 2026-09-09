# Status

**Current release: v0.1.0 — First Public Development Release**

---

## Available

These features are implemented and functional:

- **Neuron CLI** — command-line interface for project management, system registration, and execution.
- **N.O.R.E. runtime** — execution engine with TCP and Unix socket APIs.
- **System registration** — build, compile, and register systems with the runtime.
- **Instance management** — create, list, remove, and clear system instances.
- **YAML project support** — define systems, services, and connectors in YAML.
- **TypeScript SDK** — define systems programmatically with full type safety (`@neuron/sdk`).
- **Go SDK** — define systems programmatically in Go (`application/sdk`).
- **Manifest compilation** — transform source definitions into canonical manifests.
- **CEL expression engine** — evaluate mapping and validation expressions.
- **Core executors** — `set`, `ai`, `log`, `http`, `delay`, `command` (in-process).
- **External executor architecture** — resolve, install, and verify external executors.
- **Process runtime** — host executors as OS processes (gRPC worker pool or one-shot JSON).
- **WASM runtime** — host executors as WebAssembly modules via wazero.
- **SQLite persistence** — store system registrations, instances, and execution history.
- **Executor Go SDK** — write custom executors in Go (`packages/executor-go`).

---

## Experimental

These features work but may change significantly:

- **External executor distribution** — GitHub Releases and local directory registries.
- **Executor version resolution** — semver-based version selection with constraints.
- **CEL-based mappings and validations** — expression evaluation for connector mappings.
- **Streaming support** — partial implementation in `nore/internal/stream`.
- **Event system** — internal event bus for state transitions and diagnostics.

---

## Planned

These features do not yet exist:

- **Container runtime** — executor kind declared but not implemented.
- **Remote runtime** — executor kind declared but not implemented.
- **`neuron update`** — automatic self-update from GitHub Releases.
- **Package-manager distribution** — Homebrew, WinGet, Scoop, apt.
- **Release signing** — cryptographic signing of release artifacts.
- **Distributed execution** — executors running across multiple machines.
- **Additional executor providers** — beyond GitHub Releases and local directories.
- **Web UI** — browser-based system inspection and management.
- **Metrics and observability** — OpenTelemetry integration.
- **Multi-system composition** — systems referencing other systems as capabilities.
