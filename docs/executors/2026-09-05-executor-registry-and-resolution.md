# Executor Registry & Resolution

Status: current, tracks the runtime-backed executor architecture on the
`feature/executor-runtime` branch.

This document is the reference for how Neuron handles external executors: how
a service declares one, how it is found and installed, how it is frozen into a
Deployment, and how N.O.R.E. eventually runs it — either through the process
runtime (gRPC worker pool or legacy one-shot JSON) or as an embedded Wasm
module. It is written for whoever works on any layer of this stack, including
future us, and it answers three questions:

1. What is an executor, and what is a requirement?
2. How does resolution and installation actually work?
3. What happens when `neuron register` runs, and how does an Instance launch the
   frozen executors?

Every section links to the source that implements what it describes, so you can
follow along in the code instead of trusting prose.

## The big picture

Core executors (`set`, `ai`, `log`, `http`, `delay`, `command`) are in-process
implementations that live in [nore/internal/registry](../../nore/internal/registry). An external executor is
the same idea, but shipped independently: a native binary or a `wasm32-wasi`
module that speaks the same wire protocol. Because the artifact is installed
separately, everything about it is discoverable, versioned, and verifiable at
registration time.

The flow has four stages, and each stage is replaceable on its own:

A project declares a *requirement*: a logical type plus an optional version
constraint and an optional registry. The *resolver* turns that requirement into
a *package* from a *provider* (GitHub Releases or the local directory
registry). The *installer* verifies the artifact and puts it into the local
store atomically, producing an *installed* executor. At register time, the
exact resolutions are *frozen* into the Deployment as `ResolvedExecutor`
records. Instances only ever consume frozen records; they never resolve or
install on their own.

The contract that connects the two halves of the system is the wire schema in
[shared/types/executor](../../shared/types/executor). The application layer (`application/`), which resolves
and installs, and the runtime layer (`nore/`), which executes, both compile
against that package and neither imports the other's machinery.

## Project authoring and the canonical manifest

A project is written in one of two authoring languages: YAML or TypeScript.
The project-wide identifier for a language is a `language.Language`
(`yaml` or `typescript`), defined in [application/language/language.go](../../application/language/language.go). The
accepted tokens are `yaml`, `yml`, `typescript`, and `ts`, normalized
case-insensitively to the canonical form.

Every building block that wants to produce a manifest goes through the builder
registry in [application/build/build.go](../../application/build/build.go). A builder implements the `builder.Builder`
contract from [application/build/builder/builder.go](../../application/build/builder/builder.go): it declares the language
it handles and produces `.neuron/manifest.json` for a project root. The YAML
builder is [application/build/yaml/yaml.go](../../application/build/yaml/yaml.go); it resolves the project structure
(`systems/`, `services/`, `connectors/`) and materializes the manifest. The
TypeScript builder is [application/build/typescript/typescript.go](../../application/build/typescript/typescript.go); it compiles
a generated SDK program that declares the same structure. Both are registered
in the `init` of the build package, so a future language is just a new builder —
the CLI and the runtime never change.

The manifest file is the single artifact that connects authoring to the
platform. It is defined by [application/compiler/manifest](../../application/compiler/manifest) and written with
`manifest.SaveToProjectRoot`. Everything downstream reads this file, never the
original project source.

## The wire contract

The registry, resolver, and installer are consumer-side machinery; none of it
lives in [shared/types/executor](../../shared/types/executor). That package only declares the schema both
sides agree on, and it stays dependency-free.

### Runtime kinds

[shared/types/executor/runtime.go](../../shared/types/executor/runtime.go) defines how a frozen artifact is launched.
The value stored in a manifest and in `RuntimeInfo.Type` is one of these
constants:

- `process` launches the entrypoint as an OS child process. Executors built on
  the SDK ([packages/executor-go](../../packages/executor-go)) speak gRPC over a Unix domain socket via
  long-lived worker processes; see the worker pool section below.
- `wasm` launches the entrypoint inside the embedded WASI runtime, speaking the
  stdin/stdout JSON protocol.
- `container`, `remote`, and older reserved names are rejected by the runtime
  layer with an explicit "unsupported runtime kind" error, rather than being
  silently mis-executed.

`SupportedRuntimeKinds()` lists what the runtime layer can actually launch.

### The `executor.json` manifest

Every package ships a manifest ([shared/types/executor/manifest.go](../../shared/types/executor/manifest.go)) describing
its name, version, runtime, services, capabilities, and per-platform artifacts:

```json
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": { "name": "github:read", "version": "1.2.0" },
  "runtime":  { "type": "process", "entrypoint": "read", "protocol": "neuron/executor-v1", "maxWorkers": 4 },
  "services": ["read"],
  "capabilities": ["io.read"],
  "platforms": { "linux-amd64": { "artifact": "github-read-linux-amd64", "sha256": "..." } }
}
```

The `protocol` and `maxWorkers` fields of the runtime select the transport and
worker-pool bound:

- `neuron/executor-v1` — canonical gRPC protocol over Unix domain sockets for
  process executors. `maxWorkers` bounds the pool of long-lived workers (0 or
  unset uses the runtime default).
- `neuron/executor-v1-json` — legacy stdin/stdout JSON protocol, used by WASI
  executors and by process executors that predate gRPC. `maxWorkers` is ignored.

Validation requires a correct `apiVersion` and `kind`, a non-empty name,
version, runtime type, entrypoint, and at least one service.

### The execution protocol

[shared/types/executor/protocol.go](../../shared/types/executor/protocol.go) fixes the wire data model for an
execution (`Request` with `Input`, `Response` with `Output`/`Error`). The
transport that carries it depends on the declared runtime protocol:

For `neuron/executor-v1-json`, an executor reads one JSON request from stdin,
writes one JSON response to stdout, and exits. Everything on stderr is
diagnostics.

```json
{ "input": { "repo": "..." } }
{ "output": { "data": "..." }, "error": "" }
```

For `neuron/executor-v1`, the same `Request`/`Response` travel as protobuf
messages over gRPC ([shared/protocol/executor/v1](../../shared/protocol/executor/v1)), on top of a
handshake (`Initialize`: protocol version negotiation, executor identity,
capabilities), `Execute`, `Health`, and `Shutdown` calls.

A non-empty `error` in the response is a controlled failure even when the exit
code is zero. The runtime injects three environment variables:
`NEURON_EXECUTOR_PROTOCOL`, `NEURON_EXECUTOR_TYPE`, and
`NEURON_EXECUTOR_VERSION`.

### The resolved (frozen) record

[shared/types/executor/resolved.go](../../shared/types/executor/resolved.go) is what a Deployment persists instead of a
requirement. It records the requested constraint next to the exact resolved
version so a Deployment is reproducible — "whatever is latest tomorrow" is
never silently picked. It includes the runtime info, the digest, the registry,
and the install root. `EntrypointPath()` joins the install root with the
manifest-relative entrypoint, normalizing forward slashes to the host
separator.

## Requirements and logical naming

A requirement is a request, never a resolution. It is defined in
[application/executor/requirement.go](../../application/executor/requirement.go):

```go
Requirement{
    Type:       "github:read",      // logical executor name
    Version:    "^1.0.0",           // "" | "1.2.0" | "^1.0.0" | "~1.5.0" | ">=2.0.0"
    Registries: []string{"github"}, // optional; falls back to configured defaults
}
```

A logical name is a `:`-separated path. The first segment is always the owner and
at least one functional segment must follow, so `github:read` is valid and
`Muhammad-Jay:github:read` is valid. Two projections matter downstream:
`ToGitHubRepo()` hyphen-joins the trailing segments for the `owner/repo` GitHub
form (`Muhammad-Jay/github-read`), and `ToLocalPath()` slash-joins them for the
store layout (`Muhammad-Jay/github/read`). `NormalizeType` and `TypePath`
round-trip a name to and from these forms.

## Resolution and installation

[application/executor](../../application/executor) owns the pipeline. Its public contracts — the `Provider`
interface and the `Store` interface — are declared in `provider.go` and
`store_contract.go` inside this package, so providers and stores import the
executor package and never the reverse.

A `Registry` (`registry.go`) is the set of configured providers keyed by name.
It answers "which provider serves this registry name"; it is not the catalog of
installed versions.

The `Resolver` (`resolver.go`) turns a requirement into an installed executor.
Resolution order matters:

1. An exact pin that is already installed short-circuits everything.
2. A constrained requirement satisfied by an installed version uses that
   version without touching the network.
3. A floating requirement (empty version) always goes to the registries, so
   "latest" is defined by the registry and never stale-locked by whatever
   happens to be installed.
4. Otherwise the resolver walks the requirement's registries, asks each for its
   versions, picks one with `SelectVersion` (`selector.go`), and installs it.

Version selection is always owned by the resolver, never trusted to a provider:
empty constraints pick the newest valid semantic version, and `^`, `~`, and
range constraints are evaluated with `Masterminds/semver`.

The `Installer` (`installer.go`) performs atomic installs. Everything happens in
a staging directory inside the store: the artifact is downloaded
(`download.go`), materialized (directories or tarballs, see `archive.go` for the
gzip sniffing and path-escape checks), verified against its SHA-256
(`verifier.go`, mandatory whenever a digest is declared), the authoritative
manifest is written from the package rather than trusted from the payload, and
finally the staging directory is renamed into the final version directory. An
`install.json` audit record is rewritten with the final paths after the commit.

The store layout ([application/executor/store/filesystem.go](../../application/executor/store/filesystem.go)) is:

```
~/.neuron/executors/<owner>/<...segments>/<version>/
    executor.json      # authoritative manifest
    install.json       # install audit record
    <artifact...>      # binary, extracted payload, or directory
```

## Registries

The `local` registry ([application/executor/source/local/registry.go](../../application/executor/source/local/registry.go)) is a
directory-backed registry used for offline testing and demos. Its layout mirrors
the store minus the install record. It binds the host-platform artifact from the
manifest as a `file://` URL and is the registry exercised by the integration
tests and the example executors.

The `github` registry ([application/executor/source/github/](../../application/executor/source/github/)) talks to the
GitHub Releases REST API with a minimal client in `client.go` — no SDK. The
repository for a type comes from the convention in `registry.go`
(`owner/hyphen-joined`) or from an explicit catalog override. Version discovery
prefers stable tags and falls back to prereleases only when no stable release
exists. The manifest is fetched from a release asset named `executor.json`, or
from the raw file at the release tag, and its `metadata.version` must equal the
release version. Artifacts are bound from release assets, falling back to the
`download` URL. Optionally a token (`NEURON_GITHUB_TOKEN` or `WithToken`) raises
the rate limits.

## Configuration

The `executors` block in [application/config](../../application/config) controls everything:

```yaml
executors:
  storeDir: "~/.neuron/executors"        # defaults to ~/.neuron/executors
  registries:
    - name: github
      url: https://api.github.com
    - name: local
      url: "local://"
  defaultRegistries: []                  # optional fallback list
```

Build defaults register `github` and `local`. A requirement that names no
registry (or a blank one) falls back to `defaultRegistries` instead of failing,
thanks to the normalization in [application/internal/executorctl/executorctl.go](../../application/internal/executorctl/executorctl.go).
`executorctl.BuildCatalog` assembles the store, the providers, the installer,
and the resolver from this configuration and exposes `Require`, `Resolve`,
`Install`, `List`, `Inspect`, and `Remove`.

## Register: from project to Deployment

`neuron register` is the single entry point for shipping a project to the
platform, defined in [application/internal/cli/register/register.go](../../application/internal/cli/register/register.go). It no
longer assumes a prior `neuron build`; building and registering are one step.

The command takes `--lang` (`yaml`, `yml`, `typescript`, `ts`) and `--root` for
the project directory, defaulting to the current directory. The effective
language comes from the flag or from the `lang` field in the project
configuration, resolved by `language.Resolve`. The root flag also guides
configuration discovery in [application/internal/cli/cli.go](../../application/internal/cli/cli.go) (`loadConfig`),
which searches for `neuron.yaml`, `neuron.yml`, `neuron.config.yaml`,
`neuron.config.yml`, and `neuron.config.json` beneath it.

With the language and root resolved, the handler:

1. Builds the project into `.neuron/manifest.json` through
   `build.Build`, dispatching to the registered builder for the language.
2. Compiles the manifest to a `core.System` and computes the instance key with
   [application/compiler](../../application/compiler).
3. Resolves every executor requirement the manifest declares through the wired
   catalog, installing anything missing, and freezes the exact results into
   `ExecutionConfigurations.ResolvedExecutors`.
4. Sends a `RegisterRequest` to N.O.R.E., saves the returned registration key
   into the project, and prints the `system@version#hash:env` line.

`resolveFrozenExecutors` is where a requirement becomes the frozen wire record:
each `manifest.ExecutorRequirement` becomes an `executor.Requirement`, the
catalog resolves them all into an `Environment`, and each installed executor is
turned into a `ResolvedExecutor` with the requested version pinned next to the
resolved one.

The CLI also ships dedicated `neuron executor` subcommands
([application/internal/cli/executor/](../../application/internal/cli/executor/)) for working with the local store
directly: `install`, `list`, `inspect`, and `remove`.

## Execution inside N.O.R.E.

The full runtime lifecycle is documented in [RUNTIME.md](../RUNTIME.md), with the
process and WASM backends detailed in [RUNTIME_PROCESS.md](../RUNTIME_PROCESS.md)
and [RUNTIME_WASM.md](../RUNTIME_WASM.md). This section summarizes the adapter
flow.

On the N.O.R.E. side, [nore/internal/plugin](../../nore/internal/plugin) adapts frozen executors to the
in-process `contracts.Executor` contract. `RegisterResolvedExecutors` registers
an adapter for every frozen type that does not already have a core executor —
built-ins win, external types become adapters. The dispatch in `NewAdapter`
reads the runtime kind and hands the frozen record to the matching runtime
backend through the runtime registry
([nore/internal/runtime](../../nore/internal/runtime)).

Instances never resolve or install. When `Manager.GetOrCreate` builds an
Instance ([nore/internal/instance/](../../nore/internal/instance/)), it decodes the opaque
`ExecutionConfigurations` payload into `[]ResolvedExecutor`, registers the core
executors, and then the frozen adapters. Execution of a non-core service
dispatches to its adapter, which proxies to the runtime-backed `Instance`.

### Runtime backends

`NewAdapter` picks a backend by the frozen record's runtime kind. The shared
[shared/types/executor/runtime.go](../../shared/types/executor/runtime.go) contract (`Runtime`, `StartSpec`,
`Instance`) keeps the plugin layer free of launch machinery; each backend owns
its own process, socket, or WASM state:

- The process runtime ([nore/internal/runtime/process](../../nore/internal/runtime/process)) hosts executors as
  OS child processes. Two transports are supported, selected by the declared
  protocol:
  - `neuron/executor-v1` (canonical): the executor speaks gRPC over a Unix
    domain socket. The runtime maintains a pool of long-lived worker processes
    per executor type, leasing requests to available workers, bounded by
    `maxWorkers`. Workers signal readiness through the
    `NEURON_EXECUTOR_SOCKET`/`NEURON_EXECUTOR_READY` handshake and negotiate
    the protocol version during `Initialize`.
  - `neuron/executor-v1-json` (legacy): the executor speaks the one-shot
    stdin/stdout JSON protocol. Each execution spawns a fresh process, feeds it
    the JSON request on stdin, and reads the single JSON response from stdout.
    This is the transport used by WASI executors and by process executors that
    predate the gRPC protocol.
- The WASM runtime ([nore/internal/runtime/wasm](../../nore/internal/runtime/wasm)) runs a frozen `.wasm`
  module inside the embedded wazero runtime. It speaks the exact same
  stdin/stdout JSON protocol and injects the three `NEURON_EXECUTOR_*`
  environment variables through the module configuration.

An executor author therefore compiles once to a native binary and once to a
`wasm32-wasi` module, and the source can remain identical — only the manifest
protocol field differs, and only process executors that want gRPC must adopt
the SDK ([packages/executor-go](../../packages/executor-go)).

Container and remote kinds are rejected with an explicit unsupported-runtime
error rather than silently mis-executed. `SupportedRuntimeKinds()` lists what
the runtime layer can actually launch.

### The gRPC worker pool

For `neuron/executor-v1`, the process runtime keeps a pool of long-lived
workers per executor type and version. Work is leased to idle workers; when the
pool is under `maxWorkers`, a new worker is started. Workers are reused across
executions (no per-request spawn), which matters because process startup is the
dominant cost. The pool handles:

- bounded concurrency (`maxWorkers`)
- health checks
- worker restart on failure
- graceful shutdown (`Shutdown` RPC, then process reaping)
- request cancellation and deadlines through context propagation

The protocol is versioned (`neuron/executor-v1`) and transport-neutral: the
gRPC schema lives in [shared/protocol/executor/v1](../../shared/protocol/executor/v1). New executor
languages implement `ExecutorService` over gRPC; they do not need to be Go or
WASM.

### The Wasm runtime

Two properties matter for how Wasm executors behave at scale.

The wazero runtime is created once per process and shared by every instance.
The same is true of compiled modules: each distinct frozen module file is
compiled into a `wazero.CompiledModule` exactly once, cached in memory keyed by
its entrypoint path, and reused by every subsequent execution and every other
instance that references the same file. If an Instance declares several `.wasm`
modules, they are all compiled and all cached independently — the cache holds
one compiled module per distinct module file, not one module for the whole
process. This is verified by the WASM runtime tests, which register multiple
distinct modules against the same runtime.

Each execution instantiates a fresh, sandboxed module from the shared compiled
module with its own stdin and stdout buffers, runs `_start`, and tears the
module down. Executions do not serialize: instantiating from a compiled module
is safe to do concurrently, so parallel requests run in parallel sandboxes.

Runaway modules are handled by construction. The runtime is configured with
`WithCloseOnContextDone`, which makes wazero insert periodic checks, so an
in-flight `_start` is interrupted when its execution context reaches its
deadline and the module is closed automatically. A module that loops forever
cannot leak a goroutine or block the process.

### Lifecycle

`contracts.ExecutorCloser` in [nore/internal/contracts/executor.go](../../nore/internal/contracts/executor.go) lets an
adapter release resources. Adapters are registered into a `Registry`
([nore/internal/registry/executor-registry.go](../../nore/internal/registry/executor-registry.go)), and `Instance.Stop` closes the
registry after in-flight work drains. Closing a plugin adapter closes the
runtime-backed `Instance`: the process pool shuts down workers, while the Wasm
instance's `Close` is a no-op by design — it does not own the runtime or its
compiled modules, so closing one adapter never tears down resources other
instances still need. The Wasm runtime and compiled modules live for the life
of the process.

## The example executor

[examples/executors/echo](../../examples/executors/echo) is a stdlib-only Go module that implements the
JSON protocol (so it also builds as `wasm32-wasi`): it reads the request,
echoes the input back, and reflects the three `NEURON_EXECUTOR_*`
environment variables into the output. A controlled error is produced when the
input contains an `error` field, which exercises the non-zero/`error` response
path without a crash. Because it speaks JSON, its manifests declare
`neuron/executor-v1-json`.

[examples/executors/build.sh](../../examples/executors/build.sh) compiles that single source twice into the catalog
layout a local registry expects: a native binary for the `process` runtime and a
`wasm32-wasi` module for the `wasm` runtime, each with its own `executor.json`.
The generated catalog is gitignored; rerun the script after changing the source.

## Extending

Add a registry provider by implementing `Provider` (`Name`, `Types`, `Versions`,
`Package`) and registering it in `executorctl.BuildCatalog`. Distribution
conventions are the installer's materialize rules.

Add an authoring language by implementing `builder.Builder` and registering it
in the build package. The CLI, the compiler, and the runtime are untouched.

Add a runtime kind by implementing `shadexec.Runtime` in
[nore/internal/runtime](../../nore/internal/runtime) and registering the backend in the runtime
registry used by [nore/internal/plugin](../../nore/internal/plugin). An empty `protocol` on a frozen
record resolves to `neuron/executor-v1-json` (the legacy JSON transport) so old
executors keep working unmodified.

Add a core in-process executor by registering it in
[`nore/internal/registry.RegisterCoreServiceExecutors`](../../nore/internal/registry/executor-registry.go) before the plugin pass; it
will shadow any frozen external executor for the same service type.

## Testing

The unit tests cover the stable inner contracts: requirement naming and
validation, version selection, SHA-256 verification, and the local-registry
pipeline that walks resolve, install, store, and freeze in a temp directory.

The plugin integration tests ([nore/internal/plugin/plugin_integration_test.go](../../nore/internal/plugin/plugin_integration_test.go))
build the echo and spin fixtures once per test binary and verify the runtime
dispatch surface: process and Wasm round-trips with environment capture,
controlled errors, missing entrypoint failures, frozen-record decode, and
unknown-runtime rejection.

The runtime tests cover each backend at its own boundary. The process runtime
tests ([nore/internal/runtime/process/runtime_test.go](../../nore/internal/runtime/process/runtime_test.go)) exercise both
transports — legacy JSON round-trip and the gRPC worker pool against a
gRPC test executor, including worker reuse, bounded concurrency, pool close,
and protocol rejection. The WASM runtime tests
([nore/internal/runtime/wasm/runtime_test.go](../../nore/internal/runtime/wasm/runtime_test.go)) cover round-trip, the
timeout path against a module that never returns, concurrent executions against
the shared wazero runtime, the compiled-module cache across distinct modules,
and that an instance close does not hurt the shared runtime.

Run them with `go test ./application/... ./nore/... ./shared/...` from the
repository root, or use [scripts/script.sh](../../scripts/script.sh) for the full workspace build.