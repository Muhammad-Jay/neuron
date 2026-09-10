# Neuron

Neuron is a runtime for building and operating complex software systems from composable, executable capabilities.

It is built around a simple idea:

> Software should be composed from things that can do something, connected by explicit relationships, and operated by a runtime that does not need to understand what those things are.

Neuron is not a workflow platform.

It is not tied to a particular kind of application, service, language, or execution model.

It is closer to a small operating environment — a micro-kernel-like foundation for systems whose capabilities can be composed, connected, and operated independently of the technology used to implement them.

---



## Status

**Version:** `0.1.0` — first public development release.

Neuron is in active, early development. The core ideas are implemented and usable, the architecture is deliberate, and the repository is structured for long-term growth — but everything is still subject to change as the platform matures.

Current state in one sentence:

- **Working today:** the `neuron` CLI, system definition in YAML and TypeScript, compilation and registration, the N.O.R.E. runtime engine, in-process built-in modules, external modules hosted as processes or WebAssembly, system instances, and execution with live event streaming.
- **Experimental:** the external module ecosystem, Git-based module resolution, and several runtime refinements.
- **Planned:** runtime hardening, broader module distribution, and additional execution models.

See [docs/STATUS.md](./docs/STATUS.md) for the current supported surface, and [TODO.md](./TODO.md) for what is known to need work.

---



## What Is Neuron?

Modern software is usually built as a collection of applications, services, workers, libraries, queues, databases, APIs, and infrastructure.

As systems grow, the difficulty is rarely writing one individual component. The difficulty is making all of those components work together as one coherent system.

Neuron approaches this differently.

Instead of making the runtime understand every kind of application or service, it defines a small set of primitives:

- **System** — what a complete software system is made of
- **Module** — an executable capability, packaged for Neuron
- **Connector** — how capabilities communicate
- **Instance** — a living realization of a System
- **N.O.R.E.** — the runtime engine in which Systems exist and operate

These concepts deliberately remain independent.

A module does not need to be a function. A connector does not need to be an HTTP request. A module does not need to be written in the same language as the system using it. And N.O.R.E. does not need to know what a module actually does — it only needs to know how to operate it.

**Neuron treats software as a composition of capabilities rather than a collection of applications.**

A capability can be almost anything: a function, a library, an API, a database operation, a model, a filesystem operation, a browser automation task, a GitHub operation, a native program, a WebAssembly module, a process running on another machine, a service written in another programming language, or another system exposed as a capability.

Neuron does not need to understand the implementation. It only needs a contract describing what the capability provides and how it can be reached. This makes the boundary between "application", "service", "worker", and "external system" much less important. They can all become **modules**.

---



# Core Concepts



## Module

A **module** is the unified name for any executable capability packaged for Neuron.

Neuron uses one word deliberately: whether a capability is a logical operation or the machinery that executes it, to every other part of Neuron it is simply a **module**, something with a name, a contract, and a way to be run.

When more precision is needed, a module is one of two things:

- **Service** —an executable capability. `github.read`, `customer.verify`, `payment.authorize`, `database.query`, `model.predict`, `filesystem.read`, `email.send`. The name does not determine how the capability is implemented: one Service could run locally, another remotely, another as a WebAssembly module, another implemented in Rust, Go, Python, or JavaScript.
- **Executor** — the mechanism that actually runs a Service. The Executor provides the machinery required to make that capability operate.

From Neuron's perspective, they are all capabilities that can participate in a System. The Service remains the logical capability; the Executor provides the machinery. This separation is what allows Neuron to support capabilities implemented using different technologies without turning the core runtime into a collection of special cases.

Modules are referenced by name (for example `example:echo`), either because the module is built into N.O.R.E. and runs in-process, or because it is resolved, installed, and hosted out-of-process by N.O.R.E. See [docs/MODULES.md](./docs/MODULES.md) for the full module model.

## System

A **System** is the definition of a software system. It describes the capabilities that belong together and the relationships between them. A System can be small:

> receive a request → validate it → return a result

Or it can become extremely large:

> authentication → billing → inventory → notifications → analytics → external APIs → background processing

The important distinction is that a System describes **what exists and how it is connected**, rather than forcing everything to be implemented as one application. A System is therefore a composition. Systems can themselves become building blocks for larger systems.

```text
                 System
                    │
        ┌───────────┼───────────┐
        │           │           │
     Service     Service     Service
        │           │           │
        └─────── Connectors ────┘
```



## Connector

A **Connector** describes how one module communicates with another.

This distinction is important. A Service describes what can be done. A Connector describes how it can be reached or connected.

Depending on the environment, a Connector could represent communication through an in-process interface, a local process, a Unix socket, a network connection, an RPC protocol, a message channel, an external API, or another supported transport. This allows the architecture of a System to remain independent from the transport used underneath it.

```text
Service A
   │
   │ Connector
   ▼
Service B
```

The same logical relationship can exist whether both services are on the same machine or separated across a network.

## Instance

A System is a definition. An **Instance** is a living realization of that definition.

A System might describe "Customer Verification". An Instance represents an actual running or available realization of that system with its own state, resources, configuration, and activity.

This distinction allows the same System definition to have many independent instances. The same definition can be reused without forcing every realization to share the same state.

```text
                 System
          Customer Verification
                   │
          ┌────────┼────────┐
          │        │        │
       Instance  Instance  Instance
          A        B        C
```

This is one of the foundations that allows Neuron to move beyond simple task execution.

## N.O.R.E.

**Neuron Operational Runtime Engine**

N.O.R.E. is the runtime engine of Neuron.

It is where Systems are registered, instantiated, operated, and connected to the capabilities they require. N.O.R.E. provides the environment in which a System can become something that actually exists.

```text
                    Neuron
                      │
                      ▼
                   N.O.R.E.
                      │
          ┌───────────┼───────────┐
          │           │           │
       Systems     Instances    Services
          │           │           │
          └───────────┼───────────┘
                      │
                  Executors
                      │
          ┌───────────┼───────────┐
          │           │           │
       Local       WASM        Remote
```

N.O.R.E. is intentionally not responsible for understanding the business meaning of the capabilities it operates. It provides the runtime primitives; the capabilities provide the behavior.

You do not interact with N.O.R.E. directly. The `neuron` CLI manages it for you — starting it, stopping it, and communicating with it over a local connection as needed. From a user's perspective there is a single product: **Neuron**.

---



## How Neuron Works

The conceptual lifecycle of a Neuron System is:

```text
Definition
    │
    ▼
Resolution
    │
    ▼
System
    │
    ▼
Registration
    │
    ▼
N.O.R.E.
    │
    ▼
Instance
    │
    ▼
Execution
```

The important part is that a System definition is not the same thing as a running Instance. A definition describes what should exist. N.O.R.E. provides the environment. An Instance makes that definition operational. Modules provide the actual capabilities.

With the `neuron` CLI this becomes a short workflow:

1. **Define** a System in YAML or TypeScript.
2. **Register** it with N.O.R.E. — the CLI builds the project, resolves any external modules it needs, and hands a compiled System to the runtime.
3. **Run** it — the CLI asks N.O.R.E. to create an Instance and execute the System, streaming live execution events back to you.

```text
Neuron CLI
    │
    ├── project → build → compiler
    ├── module resolution
    └── N.O.R.E. client
             │
             ▼
         N.O.R.E.
             │
             ▼
         Instance
             │
             ▼
        Service Executor
```



## The Runtime as a Small Operating Environment

The operating-system analogy is useful, but with an important distinction. Neuron is not trying to become an operating system for hardware. It is an operating environment for software capabilities.

An operating system provides primitives such as processes, memory, resources, communication, isolation, scheduling, identity, and persistence. Neuron applies a similar philosophy at a higher level, providing a foundation around Systems, Instances, Services, Executors, Connectors, resources, state, communication, and execution.

This is why N.O.R.E. can be thought of as a micro-kernel-like runtime for Neuron. The kernel should remain small. Capabilities should live outside it.

## Everything Is a Capability

One of the most important ideas in Neuron is that a Service does not have to correspond to a traditional "microservice". A Service can represent a capability at any level.

```text
                 System
                    │
        ┌───────────┼───────────┐
        │           │           │
    Database      Model       GitHub
     Service     Service      Service
        │           │           │
        └───────────┼───────────┘
                    │
                Application
```

Or:

```text
System
  │
  └── Another System
          │
          ├── Service
          ├── Service
          └── Service
```

This creates a recursive model of software composition. Complex systems can be constructed from smaller systems without requiring the runtime to treat them as fundamentally different things.

## Local and Distributed

Neuron is designed so that location does not have to define the architecture.

```text
Same process
     ↓
Same machine
     ↓
Another local process
     ↓
Another machine
     ↓
Another runtime
     ↓
Another environment
```

The logical System can remain the same while its physical deployment changes. This makes it possible to begin with a completely local system and gradually distribute individual capabilities as the system grows.

## WebAssembly

WebAssembly provides one possible execution environment for Services. Its value in Neuron is not simply that it is "fast". It provides a portable and controlled execution boundary.

That makes it useful when a Service needs to be portable, isolated, distributed as a single artifact, executable across different environments, implemented independently from the main runtime, or loaded dynamically. WebAssembly therefore fits naturally at the Executor boundary. It is one execution mechanism among many, rather than something every Service must become.

## Why Neuron Is Not a Workflow Engine

Workflow engines generally begin with a predefined abstraction: a workflow consists of a sequence of tasks. Neuron starts somewhere else: a System consists of capabilities and relationships.

That difference matters. A workflow is one possible thing that can be represented using Neuron — it is not the boundary of the platform. A System may represent an API backend, an automation system, an AI application, a data-processing environment, an interactive application, a distributed service, an internal platform, a long-running process, or something that does not fit neatly into traditional application categories.

Neuron is intended to provide the underlying runtime model rather than dictate the application model.

---



## Architecture

The repository is a monorepo containing the whole platform:

```text
nuron/
├── application/    The neuron CLI, compiler, loader, and project toolchain
├── nore/           N.O.R.E. — the Neuron Operational Runtime Engine
├── shared/         Canonical types and protocol contracts shared by both modules
├── packages/
│   ├── sdk/        TypeScript SDK for defining Neuron systems (@neuron/sdk)
│   └── executor-go/ Go SDK for authoring Neuron modules (executors)
├── examples/       Runnable example systems and reference executors
├── docs/           Architecture, getting started, installation, and module docs
└── scripts/        Workspace development and release helpers
```

`application` and `nore` are separate Go modules that only agree through the canonical types in `shared`. The CLI authoring toolchain never speaks to the runtime's internals, and the runtime never parses YAML or TypeScript. The source language converges on one canonical manifest before anything runtime-specific happens.

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for the full architecture.

---



## Installation

Neuron ships as one product: the `neuron` CLI plus the N.O.R.E. runtime engine, distributed together as a single archive for your operating system and architecture.

Supported release targets:


| Platform | Architectures    |
| -------- | ---------------- |
| Linux    | `amd64`, `arm64` |
| macOS    | `amd64`, `arm64` |
| Windows  | `amd64`          |


The quick path:

1. Download the latest release archive for your platform from the [releases page](https://github.com/Muhammad-Jay/neuron/releases).
2. Extract it.
3. Place the `neuron` binary on your `PATH`.

Verify the installation:

```bash
neuron version
```

The CLI runs the N.O.R.E. runtime engine for you. You do not install, start, or manage N.O.R.E. separately.

See [docs/INSTALLATION.md](./docs/INSTALLATION.md) for the complete, step-by-step installation guide, including installing from source.

---



## Quick Start

The fastest way to experience Neuron is with the [official example project](./examples/ecommerce_order_ts), which defines an order-processing pipeline in TypeScript.

```bash
# Clone the repository
git clone https://github.com/Muhammad-Jay/neuron.git
cd neuron

# From the example project
cd examples/ecommerce_order_ts
pnpm install

# Register the system with the runtime engine
neuron register

# Run it
neuron run
```

The same flow works with YAML — see the [YAML example](./examples/ecommerce_order).

To author your own project:

```bash
neuron init my-system
cd my-system
# ... define your system in YAML or TypeScript ...
neuron register
neuron run
```

`neuron init` scaffolds a project, `neuron register` builds and registers your system (starting the local runtime automatically), and `neuron run` executes it while streaming events.

Walk through it in detail in [docs/GETTING_STARTED.md](./docs/GETTING_STARTED.md).

---



## TypeScript SDK

The TypeScript SDK (`[@neuron/sdk](./packages/sdk)`) is Neuron's system-definition language: a typed, always-autocompleted way to describe services, connectors, and system composition in TypeScript, producing the canonical manifest the compiler consumes.

```typescript
import { Service, System } from "@neuron/sdk";

const validate = new Service("validate")
  .inputSchema({ order: "object" })
  .outputSchema({ valid: "boolean" });

const enrich = validate.next(new Service("enrich"));

export default System({
  name: "order-processing",
  version: "1.0.0",
  description: "An order-processing pipeline",
}).run(enrich);
```

The SDK is a definition tool — it describes systems; it does not execute them. The Go side remains responsible for parsing, validating, compiling, and running the canonical representations.

See [packages/sdk/README.md](./packages/sdk/README.md) for the SDK documentation.

---



## Modules & Executors

Neuron distinguishes two kinds of modules:

- **Built-in modules** — N.O.R.E. ships with a small set of in-process modules for common operations. Referencing one requires no installation; it runs inside the runtime engine.
- **External modules (executors)** — authored, packaged, and distributed independently. N.O.R.E. resolves them, verifies them, installs them, and hosts them out-of-process as native processes or WebAssembly modules.

External modules have a stable, unified contract:

- an `executor.json` manifest declaring identity, runtime type, protocol, and supported platforms;
- a canonical package archive (`<name>-<version>-executor.neuron.tar.gz`);
- a declared wire protocol so the runtime knows how to talk to them.

Neuron resolves external modules by name and version — for example `example:echo@1.0.0` — against configured registries (Git-releases based by default), selects the best matching version with semantic versioning, verifies the artifact, installs it immutably, and freezes the exact resolved set into every registered system.

A reference module is included in this repository: `[examples/executors/echo](./examples/executors/echo)`, built as both a native process and a WebAssembly module from the same Go source.

See [docs/MODULES.md](./docs/MODULES.md) for the complete module model, how resolution works, and how to author a module.

---



## Examples

The repository ships runnable examples for every authoring surface:


| Example                                                        | Surface          | What it demonstrates                                                     |
| -------------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------ |
| `[examples/ecommerce_order_ts](./examples/ecommerce_order_ts)` | TypeScript SDK   | Typed system definition of a full order-processing pipeline              |
| `[examples/ecommerce_order](./examples/ecommerce_order)`       | YAML             | The same pipeline expressed with YAML project, system, and service files |
| `[examples/executors](./examples/executors)`                   | Module authoring | Reference `echo` module compiled for both process and WASM runtimes      |
| `[examples/simple_response](./examples/simple_response)`       | Go               | A minimal Go-defined system using the SDK builders                       |


---



## Documentation

- [Architecture](./docs/ARCHITECTURE.md) — how the platform is put together
- [Getting Started](./docs/GETTING_STARTED.md) — run your first Neuron system
- [Installation](./docs/INSTALLATION.md) — install Neuron from an official release
- [Modules & Executors](./docs/MODULES.md) — the unified module model
- [Status](./docs/STATUS.md) — what is available, experimental, and planned
- [TODO](./TODO.md) — known issues and future work
- [Runtime deep-dives](./docs/RUNTIME.md) — the N.O.R.E. runtime in depth (maintainer-focused)
- [CLI reference](./application/README.md) — every `neuron` command and flag
- [N.O.R.E. reference](./nore/README.md) — the runtime engine in detail (maintainer-focused)

---



## Development

The repository is a Go workspace plus a pnpm monorepo.

Prerequisites:

- Go `1.26.5` or newer
- pnpm `10.33.0` (or newer, in the `10.x` series)
- Node.js (for the TypeScript SDK)

Verify everything locally:

```bash
# Go workspace
go test ./nore/... ./application/... ./shared/... ./packages/executor-go/... ./examples/simple_response/...
go vet  ./nore/... ./application/... ./shared/... ./packages/executor-go/... ./examples/simple_response/...
go build ./nore/... ./application/... ./shared/... ./packages/executor-go/... ./examples/simple_response/...

# TypeScript SDK
pnpm install
pnpm build:sdk
pnpm test:sdk
pnpm typecheck:sdk
```

Or run the convenience helper from the repository root:

```bash
npm run script
```

which validates every Go module and every npm package in the workspace.

A local smoke test of the full product flow:

```bash
go build -o /tmp/neuron ./application/cmd/neuron
/tmp/neuron version
cd examples/ecommerce_order_ts && /tmp/neuron register && /tmp/neuron run
```

Contributions should follow the engineering contract in [AGENTS.md](./AGENTS.md): preserve architectural boundaries, avoid duplication, remove dead code, and test at the correct boundary.

---



## Roadmap

- **0.1.x** — stabilization of the 0.1.0 foundation: bug fixes, documentation corrections, installation refinements, and runtime correctness.
- **0.2.x** — new runtime capabilities, a broader external module ecosystem, and new SDK capabilities as they mature.
- **1.0** — a stable public model and compatibility guarantees.

The granular, always-current list of intended work lives in [TODO.md](./TODO.md).

---



## License

Neuron is released under the MIT License. See [LICENSE](./LICENSE).