# Getting Started

This guide walks through Neuron in about fifteen minutes: installing the CLI, registering your first system with the runtime engine, running it, and streaming live execution events.

No prior knowledge is assumed. Everything underneath — building, module resolution, compilation, starting the runtime engine — is automatic.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Install Neuron](#install-neuron)
- [The Big Picture](#the-big-picture)
- [Part 1 — Run a shipped example](#part-1--run-a-shipped-example)
- [Part 2 — Author your own system](#part-2--author-your-own-system)
- [Part 3 — Use an external module](#part-3--use-an-external-module)
- [Next steps](#next-steps)

---

## Prerequisites

- **Neuron** installed (see [Install Neuron](#install-neuron)); the `neuron` binary must be on your `PATH`.
- The repository checked out, to access the shipped examples:

  ```bash
  git clone https://github.com/Muhammad-Jay/neuron.git
  cd neuron
  ```

  (If you installed Neuron from a release archive, clone the repository anywhere to follow Part 1.)

No daemon setup is required. The runtime engine (N.O.R.E.) is started and stopped for you by the CLI.

---

## Install Neuron

If you have not installed Neuron yet:

```bash
neuron version
```

This prints the CLI version. If it fails, follow [docs/INSTALLATION.md](./INSTALLATION.md).

---

## The Big Picture

Three moving parts, one user interface:

```text
You                             The CLI                      The runtime
─────                           ───────                      ─────────
write a system  ──(register)──► build + resolve modules ──►  N.O.R.E. stores it
                                compile to a core System
                                hand it to the daemon
                                                                 │
──────────────────────────────────────────────────────────────── ┘
run it          ──(run)──────► ask N.O.R.E. for an instance ──► creates an Instance
                                stream events back                    │
                                                                      ▼
                                                              executes the System
```

- **System** — a definition: what capabilities exist and how they are connected.
- **Module** — an executable capability packaged for Neuron (a Service or the Executor that runs it).
- **Instance** — a living realization of a System, with its own state and activity.
- **N.O.R.E.** — the runtime engine (daemon) that hosts registered systems, creates instances, and executes them.

You only ever talk to the `neuron` CLI.

---

## Part 1 — Run a shipped example

The repository ships an order-processing pipeline defined in YAML (`examples/ecommerce_order`) and the identical pipeline in TypeScript (`examples/ecommerce_order_ts`). They use only built-in modules — capabilities that run in-process inside N.O.R.E. — so they need zero setup.

```bash
cd examples/ecommerce_order
neuron register
```

`neuron register` runs the whole authoring pipeline in one step:

1. The project is built into a canonical manifest (`.neuron/manifest.json`).
2. The manifest is validated and compiled into a core system representation.
3. All module references are resolved. Built-in modules are skipped — they run inside the runtime engine, no installation needed.
4. The compiled system is handed to N.O.R.E., which persists it and returns a system key.

Output ends with the registration key:

```text
order-processing@1.0.0#<key>:development
```

> N.O.R.E. was started automatically. You never start or stop it yourself.

Now run it:

```bash
neuron run
```

The CLI asks N.O.R.E. to create an instance and execute the system, streaming live execution events:

```text
execution.events        instance started
service.evaluating      validate-order
service.completed       validate-order
service.evaluating      parse-order
service.completed       parse-order
...
execution.completed     status: completed
```

The command returns when execution reaches a terminal state (`execution.completed`, `execution.failed`, or `execution.cancelled`).

Look at what you defined — `examples/ecommerce_order/systems/order-processing/system.yaml` describes the whole pipeline: services (validation, parsing, customer enrichment, totals, payment, shipment, confirmation) wired together by connectors, each connector optionally carrying mappings and validations between the two modules.

The TypeScript version works the same way:

```bash
cd ../ecommerce_order_ts
pnpm install
neuron register
neuron run
```

---

## Part 2 — Author your own system

### Create a project

```bash
neuron init my-first-system
cd my-first-system
```

`neuron init` creates the directory and a starter configuration file:

```bash
ls -la
# neuron.yaml
```

### Define a system

Create the standard layout:

```bash
mkdir -p systems/my-system services
```

A Neuron system is a composition of capabilities (modules) connected by explicit relationships. Define a service — a logical capability provided by a module:

```yaml
# services/echo.yaml
apiVersion: neuron/v1
kind: Service

metadata:
  name: echo
  version: 1.0.0
  description: Echoes the execution input back through an external module

spec:
  executor:
    type: example:echo@^1.0.0
```

`spec.executor.type` names the module that provides the capability — here the reference `example:echo` module. The `^1.0.0` says "any compatible 1.x version"; Neuron selects the best match with semantic versioning and **freezes** the exact resolved version into your registration.

Then define the system itself — a single-service system needs no connectors:

```yaml
# systems/my-system/system.yaml
apiVersion: neuron/v1
kind: System

metadata:
  name: my-system
  version: 1.0.0
  description: An echo system

services:
  - ref: echo
    entry: '../../services/echo.yaml'
```

### Register and run

```bash
neuron register
neuron run --input '{"message": "hello neuron"}'
neuron run -v
```

The `-v` flag renders full event payloads so you can see the echoed input travel through the system.

> **Note:** the service above references the external reference module. To resolve it you need a registry that serves it. Follow [Part 3](#part-3--use-an-external-module) to point a local registry at the built reference module; in production you would configure a published registry instead (see [docs/MODULES.md](./MODULES.md)).

### Inspect what is running

Your instance has real state in the runtime:

```bash
neuron instance list            # list instances
neuron instance list --all      # include inactive ones
neuron instance remove <id>     # remove one instance
neuron instance clear           # remove everything
```

---

## Part 3 — Use an external module

This part makes the Part 2 walkthrough resolvable offline using the included **reference module**, and demonstrates the full module lifecycle — author → build → serve → resolve → install → run.

### Build the reference module

The repository ships an `echo` module compiled from one Go source (`examples/executors/echo`) into both a native process runtime and a WebAssembly runtime. Build it into the local registry catalog:

```bash
cd examples/executors
./build.sh
```

This produces `examples/executors/catalog/`, containing for each module version:

```text
catalog/example/echo/1.0.0/
    executor.json                              module manifest
    echo                                       native binary (process runtime)
    example-echo-1.0.0-executor.neuron.tar.gz  canonical package archive
```

`executor.json` declares identity, runtime type, protocol, and platform artifacts:

```json
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": { "name": "example:echo", "version": "1.0.0" },
  "runtime": { "type": "process", "entrypoint": "echo", "protocol": "neuron/executor-v1-json" },
  "services": ["example:echo"],
  "capabilities": [],
  "platforms": { "linux-amd64": { "artifact": "echo" } }
}
```

### Register the catalog as a local registry

Add a `local` registry pointing at the catalog in your project's `neuron.yaml`:

```yaml
executors:
  registries:
    - name: local
      url: /absolute/path/to/neuron/examples/executors/catalog
```

The `local` registry is a directory-backed registry served offline. The default `github` registry serves the same kind of packages from GitHub Releases over the network.

### Resolve, install, run

```bash
cd my-first-system
neuron register
```

During registration Neuron resolves `example:echo@^1.0.0`:

1. the registry is queried for available versions (here: `1.0.0`);
2. the best satisfying version is selected;
3. the canonical package archive is downloaded, verified, and installed immutably into the executor store (`~/.neuron/executors`);
4. the exact version is frozen into your registration, so running an instance needs no further resolution.

You can also manage the module directly:

```bash
neuron add example:echo@^1.0.0      # resolve + install into the store
neuron executor list                # installed modules
neuron executor inspect example:echo@1.0.0
neuron remove example:echo@1.0.0
```

Now run:

```bash
cd my-first-system
neuron run --input '{"message": "echo me"}'
```

The execution launches the installed module out-of-process, passes your input through the executor protocol, and streams the events back.

---

## Next steps

- [docs/MODULES.md](./MODULES.md) — the unified module model, the `executor.json` contract, and authoring your own module
- [docs/ARCHITECTURE.md](./ARCHITECTURE.md) — how the platform is put together
- [application/README.md](../application/README.md) — every `neuron` command and flag
- [docs/STATUS.md](./STATUS.md) — what is available, experimental, and planned
- [TODO.md](../TODO.md) — known issues and future work