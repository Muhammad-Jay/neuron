# Executor Registry & Resolution — Reference

**Date:** 2026-09-05
**Status:** current, mirrors the implementation at commit `1b71c14`
**Scope:** the external executor subsystem (`shared/types/executor`, `application/executor`, `application/internal/executorctl`, `application/internal/cli/executor`, `application/internal/cli/register`, `nore/internal/plugin`)

This document is the single reference for how Neuron discovers, resolves,
installs, freezes, and executes external executor packages. It is written for
future developers (including past-us) to answer three questions:

1. What is an executor, and what is a requirement?
2. How does resolution + installation actually work?
3. What happens when `neuron register` (or any executor command) runs?

---

## 1. Concepts

Core executors (`set`, `ai`, `log`, `http`, `delay`, `command`) are in-process
implementations hardcoded in `nore/internal/registry`. **External executors**
are the same idea, but shipped independently and executed out of process.

| Term | Meaning |
| --- | --- |
| **Requirement** | What a service *asks for*: a logical type + optional version + registries. A request, never a resolution. |
| **Executor type** | The logical name of an executor, e.g. `github:read`, `hashicorp:vault:auth`. |
| **Package** | An immutable, resolvable artifact from a registry (one exact version, pre-install). |
| **Provider / Registry** | Something that serves Packages (`github`, `local`). The catalog of providers is the `Registry`. |
| **SelectVersion** | Version selection above a provider: semver-aware, never trusts the provider's ordering. |
| **Installed** | A verified, atomically-renamed artifact in the store (`~/.neuron/executors/...`). |
| **ResolvedExecutor** | The *frozen* wire record persisted in a Deployment: pinned version, registry, digest, launch info. |
| **Deployment** | A registered System + its frozen executor set. Instances execute from the frozen set; they never resolve or install. |

The pipeline is layered so each stage is replaceable:

```
Requirement ──▶ Resolver ──▶ Provider(s) ──▶ Package ──▶ Installer ──▶ Store ──▶ Installed
                                    ▲             ▲                                  │
                              GitHub / local   version pick                        ▼
                                                                         frozen ResolvedExecutor
                                                                                    │
                                                                                    ▼
                                                                              runtime / nore
```

---

## 2. Requirements & logical names

Source: `application/executor/requirement.go`.

A `Requirement` is:

```go
Requirement{
    Type:       "github:read",     // logical executor name
    Version:    "^1.0.0",          // optional: "" | exact "1.2.0" | constraint "^1.0.0", "~1.5.0", ">=2.0.0"
    Registries: []string{"github"}, // optional: fall back to configured defaults
}
```

### 2.1 Logical names (`ParseType`)

A logical name is a `:`-separated path. **The first segment is always the
owner**; at least one functional segment must follow.

| Type | Owner | PathSegments |
| --- | --- | --- |
| `github:read` | `github` | `[read]` |
| `Muhammad-Jay:github:read` | `Muhammad-Jay` | `[github, read]` |
| `hashicorp:vault:auth` | `hashicorp` | `[vault, auth]` |

Two projections matter downstream:

- `NameSplit.ToGitHubRepo()` → `owner/repo`, trailing segments **hyphen-joined**
  (GitHub has no nested repos). `Muhammad-Jay:github:read` → `Muhammad-Jay/github-read`.
- `NameSplit.ToLocalPath()` → store-relative directory, segments **slash-joined**.
  `Muhammad-Jay:github:read` → `Muhammad-Jay/github/read`.
- `NormalizeType` / `TypePath` round-trip a name to/from these forms.

`Validate()` requires a non-empty type that can be parsed (at least one `:`).

---

## 3. The wire contract (`shared/types/executor`)

Registry/resolution/install machinery never lives here. This package only
declares the schema **both sides** agree on, so N.O.R.E. never imports registry
code. It must stay dependency-free.

### 3.1 `executor.json` manifest (`manifest.go`)

The manifest shipped with every package. Constants: `APIVersion = "neuron/v1"`,
`Kind = "Executor"`, `ManifestFile = "executor.json"`, `InstallFile = "install.json"`.

```json
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": { "name": "github:read", "version": "1.2.0" },
  "runtime":  { "type": "process", "entrypoint": "read", "protocol": "neuron/executor-v1" },
  "services": ["read"],
  "capabilities": ["io.read"],
  "platforms": { "linux-amd64": { "artifact": "github-read-linux-amd64", "sha256": "..." } }
}
```

Validation (`Manifest.Validate()`) requires: correct `apiVersion`/`kind`,
non-empty `metadata.name`, `metadata.version`, `runtime.type`, `runtime.entrypoint`,
and at least one service. `HasCapability` checks an opt-in capability.

### 3.2 Process protocol (`protocol.go`)

A process executor reads **one JSON Request from stdin**, writes **one JSON
Response to stdout**, and exits zero. Everything else on stderr is diagnostics.

```json
// stdin                           // stdout
{ "input": { "repo": "..." } }    { "output": { "data": "..." }, "error": "" }
```

Request: `{ "input": map[string]any }`.
Response: `{ "output": map[string]any, "error": string }` — non-empty `error`
is a controlled failure even with exit 0.

Env vars injected by the runtime: `NEURON_EXECUTOR_PROTOCOL`,
`NEURON_EXECUTOR_TYPE`, `NEURON_EXECUTOR_VERSION`.

### 3.3 Frozen record (`resolved.go`)

`ResolvedExecutor` is what a Deployment persists. It records the **requested
constraint** next to the **exact resolved version** so a Deployment is
reproducible ("whatever is latest tomorrow" is never silently picked). Includes
`RuntimeInfo{Type,Protocol,Entrypoint}`, `Digest`, `Registry`, `RootDir`.
`EntrypointPath()` resolves the absolute entrypoint from `RootDir`.

---

## 4. Resolution & installation pipeline (`application/executor`)

### 4.1 Provider & Store interfaces (consumer-side)

`Provider` (`provider.go`) and `Store` (`store_contract.go`) live **in** the
`executor` package, not in the implementation packages. This keeps imports
one-directional and avoids cycles:

```
executor ─────────────┐
  ▲                   │
source/github         │
source/local  ────────┴──▶ executor   (providers import executor.Package)
executor/store ─────────────────────▶ executor   (store imports executor types)
```

`Provider` contract: `Name()`, `Types(ctx)`, `Versions(ctx, typ)`,
`Package(ctx, typ, version)`.

`Store` contract: `Root()`, `Stage()`, `Commit(ctx, stage, typ, version)`,
`Get(ctx, typ, version)`, `List(ctx, typ)`, `Remove(ctx, typ, version)`.

### 4.2 Registry catalog (`registry.go`)

`Registry` is the set of configured `Provider`s keyed by name. `Get(name)`
returns a provider or `ErrRegistryNotConfigured` when the requirement names one
that is not configured. It is *not* the installed catalog; resolve vs. install
are separate.

### 4.3 Resolver (`resolver.go`)

Turns a `Requirement` into an `Installed`. Resolution order:

1. **Exact pin already installed** → `store.Get(type, version)`.
2. **Constrained requirement satisfied by an installed version** → best
   installed version matching the constraint (no network).
3. **Floating (`version == ""`) requirements always go to the registries** —
   "latest" is defined by the registry, not by whatever happens to be
   installed. (A floating check would stale-lock upgrades.)
4. **Registry loop**: for each registry in `req.Registries`, ask `Versions`,
   pick with `SelectVersion`, fetch `Package`, install.

`ResolveMany` dedupes by type and returns an `Environment` (the frozen set an
Instance consumes).

`installedSatisfying` implements steps 1–3. Exact pins short-circuit the
registry entirely (offline-safe).

### 4.4 Version selection (`selector.go`)

Always performed by the resolver, never trusted to a provider:

- Empty constraint → newest **valid semver** (invalid entries ignored).
- Constraint (`^`, `~`, ranges) → newest version satisfying it, via
  `github.com/Masterminds/semver/v3`.
- `IsExactVersion("1.2.0")` distinguishes a pin from a constraint/floating.

### 4.5 Installer (`installer.go`) — atomic install

```
Store.Stage()                → <root>/.tmp-<rand>/   (inside the store)
download artifact → dst      → materialize (extract/copy/place binary)
verify SHA-256                → mandatory when expected digest present for single files
write authoritative executor.json (from pkg.Manifest; never trust the payload copy)
resolve entrypoint            → from manifest, falling back to artifact file name
write install.json            → RecordFor(installed) with STAGING paths
Store.Commit(stage, type, v)  → atomic os.Rename(stage → final dir)
                               + rewrite install.json with FINAL paths
```

`AlreadyPresent` is idempotent: if the exact version is already in the store,
no download occurs. `InstallResult{Installed, AlreadyPresent}`. A nil
Downloader, missing artifact, or checksum mismatch fails the whole install;
a partially materialized staging dir is removed by `defer`.

### 4.6 Downloader (`download.go`)

`Downloader` interface + default `HTTPDownloader` supporting `http(s)://` and
`file:` (or bare path) schemes. Directories are copied recursively; remote
resources streamed to disk (5-minute client timeout).

### 4.7 Archive extraction (`archive.go`)

`IsArchive` sniffs the gzip magic `1f 8b`. `ExtractTarGz` strips a single shared
top-level directory (the common `<name>-v1.2.0/` release layout) and rejects
paths escaping the destination (`inside` check).

### 4.8 Verification (`verifier.go`)

`VerifySHA256(path, expected)` accepts `sha256:<hex>` or bare hex; empty
expected digest is a hard error (**never silently skip**). `DigestFile(path)`
returns `sha256:<hex>`.

### 4.9 Store layout (`executor/store/filesystem.go`)

```
~/.neuron/executors/
└── <owner>/<...segments>/<version>/
    ├── executor.json      # authoritative manifest
    ├── install.json       # InstallRecord (audit trail)
    └── <artifact…>        # binary, extracted payload, or pre-extracted dir
```

`InstallRecord` (`installed_record.go`): `type/version/digest/registry/platform/
rootDir/manifestPath/artifactPath/runtime/capabilities/services/installedAt`.
`RecordFor` preserves the runtime contract needed to launch later.
`Commit` renames the staging dir into the final version dir and *rewrites*
`install.json` with the final paths (the staging copy would otherwise record a
stale `.tmp-*` path). `List` supports an empty type (walk-all) and a specific
type; newest-first ordering.

---

## 5. Registries (providers)

### 5.1 `local` (`source/local/registry.go`)

A directory-backed registry used for offline testing and demos. Layout mirrors
the store minus `install.json`:

```
<root>/github/read/1.2.0/
    executor.json
    github-read-linux-amd64   # optional artifact
```

`New(root)` errors with `ErrNotFound` if the root is missing. `Package` binds
the host-platform (`GOOS-GOARCH`) artifact as a `file://` URL. This is the
registry exercised by the integration tests.

### 5.2 `github` (`source/github/`)

Backed by the GitHub Releases REST API (minimal client, no SDK).

- **Repository derivation:** `repoFor(type)` uses the catalog override if
  present (`WithCatalog(map[type]RepoRef)`), else the convention
  `owner/segments-hyphen-joined` from `ParseType`.
- **Versions:** `ListReleases(owner, repo)` per_page=100, drafts excluded.
  `versionsFromReleases` prefers stable tags, falls back to prereleases only
  when no stable release exists (floating requirements never silently grab
  alphas); leading `v` stripped.
- **Manifest fetch precedence** (`fetchManifest`): a release *asset* named
  `executor.json`, else the raw repository file at the release tag
  (`raw.githubusercontent.com/<owner>/<repo>/<tag>/executor.json`).
- **Version consistency:** the manifest's `metadata.version` must equal the
  release version — otherwise the package is rejected.
- **Artifact binding:** the release asset whose name matches the manifest's
  host-platform entry (with `BrowserDownloadURL`), falling back to the raw
  download URL (`github.com/.../releases/download/<tag>/<asset>`). The manifest
  is authoritative for the artifact name.
- **Auth:** optional token (`WithToken`, or `NEURON_GITHUB_TOKEN` via the
  wiring layer) for higher rate limits / private repos.
- **Types:** catalog-driven (only explicitly configured types `GitHub` reports;
  never guesses at conventions, to avoid surprising 404s).

---

## 6. Runtime (`application/executor/runtime`)

The boundary N.O.R.E. sees: `Executor.Execute(req) → resp`. It never knows
about registries or downloads.

- `Executor` interface: `Type()`, `Version()`, `Execute(ctx, *shadexec.Request)`.
- `Factory(ctx, installed) → (Executor, error)` — one per runtime **kind**.
- `Runtime` dispatches `Materialize` by `installed.Runtime.Type`.
  `DefaultRuntime()` registers the built-in `process` kind.
- `ProcessExecutor` (`process.go`): spawns the installed entrypoint **per
  execution**, writes the JSON Request to stdin, reads the JSON Response from
  stdout, injects the three env vars, honors the context deadline. Checks the
  entrypoint exists and is executable before launch.

Future kinds (`wasm`, `container`, `remote`) register the same Factory contract.

---

## 7. Configuration

`application/config` `ExecutorsConfig` block:

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

Build defaults register `github` + `local`. `storeDir` is path-expanded by the
loader. The `github` URL is the API base; the `local` URL is the local registry
root when it is not `local://`.

---

## 8. Wiring (`application/internal/executorctl`)

`executorctl.BuildCatalog(CatalogConfig)` assembles the pipeline from config:

- `store.NewFilesystemStore(storeDir)` (default `~/.neuron/executors`).
- A fresh `executor.Registry`; for each configured registry:
  - `github` → `github.New(WithToken(token))` (+ optional catalog override).
    Token falls back to `NEURON_GITHUB_TOKEN`.
  - `local` → `local.New(url)`.
  - Unknown names are tolerated at build time and fail at resolve time.
- `Installer{Store, Downloader: NewHTTPDownloader()}`,
  `Resolver(reg, store, installer)`.

Exposed operations: `Require(type, version, registries)` (fills
`DefaultRegistries`), `Resolve`, `Install`, `List`, `Inspect`, `Remove`.

---

## 9. CLI commands (`application/internal/cli/executor`)

`neuron executor <subcommand>` registers into the root CLI. References are
`name@version` — version is optional.

| Command | Behavior |
| --- | --- |
| `neuron executor install github:read@^1.0.0` | Resolve (registry, installed-first for pins/constraints) and install; prints `installed <type>@<version> (from <registry>)`. |
| `neuron executor list [-t <type>]` | Lists installed executors (all or filtered) newest-first. |
| `neuron executor inspect github:read` | `name@version` or newest installed for the type; prints record + runtime + capabilities/services. |
| `neuron executor remove github:read@1.2.0` | Removes one installed version (**version required** — immutable artifacts are deleted whole). |

Each command loads the effective config from the command context
(`config.FromContext`) and builds the catalog on demand. No daemon required.

---

## 10. `neuron register` — end-to-end

`application/internal/cli/register/register.go`. Sequence:

```
load config ──▶ bootstrap.SetupClient ──▶ manifest.LoadFromProjectRoot (neuron build output)
     │                                          │
     │                                          ▼
     │                                   compiler.Compile(m) → core.System
     │                                          ▼
     │                                   compiler.InstanceKey(m) → protocol.InstanceKey
     │                                          ▼
     │                                   configs = compiler.BuildExecutionConfigurations(m)
     │                                          ▼
     └─▶ resolveFrozenExecutors(ctx, cfg, configs.ExecutorRequirements)
                │
                ├─ executorctl.BuildCatalog
                ├─ catalog.Require(name, version, [registry]) per requirement
                ├─ catalog.Resolve → executor.Environment   (installs anything missing)
                └─ installed.Frozen(requested[type]) per executor → []ResolvedExecutor
                    (requested version keyed by type, not by list index — ResolveMany
                     dedupes; Environment.Resolved() is the equivalent helper in model.go)
                          │
                          ▼
              configs.ResolvedExecutors = frozen  (JSON: "resolved_executors")
                          │
                          ▼
        protocol.RegisterRequest{Key, System, ExecutionConfigurations: configs}
                          │
                          ▼
        client.Register ──▶ project.SaveRegistrationKey ──▶ print
```

What register **requires**: `neuron build` must have produced
`.neuron/manifest.json`. What register **does**: compiles the manifest to a
System, resolves every executor requirement through the wired catalog (so it
is *installed locally*), freezes the exact resolutions into the payload, and
sends it to N.O.R.E.

### 10.1 N.O.R.E. side (`nore/internal/plugin`, `nore/internal/instance`)

`RegisteredSystem.ExecutionConfigurations` is stored **opaquely** (`any`) so
N.O.R.E. never depends on `application/compiler`. When an Instance is created
(`Manager.GetOrCreate`):

1. `plugin.DecodeResolvedExecutors(payload)` JSON-round-trips the opaque payload
   into `[]shadexec.ResolvedExecutor` (works for both typed values in-process
   and `map[string]any` re-read from disk).
2. `registry.RegisterCoreServiceExecutors()` registers the 6 in-process
   executors first.
3. `plugin.RegisterProcessExecutors(reg, resolved)` registers a
   `ProcessAdapter` for every frozen type that has **no** existing executor —
   core executors win; external types become subprocesses.
4. `Instance.New(..., WithResolvedExecutors(resolved))` builds the runtime;
   executions of non-core services dispatch to the `ProcessAdapter`, which
   spawns the frozen entrypoint, feeds `execution.Input` as the protocol
   Request, and surfaces `Response.Output` (or `Response.Error`).

Instances therefore never resolve or install anything — resolution happens at
**register** time and is frozen in the Deployment.

---

## 11. Error taxonomy (`errors.go`)

| Sentinel | Meaning | Retry? |
| --- | --- | --- |
| `ErrNotFound` | Executor cannot be located | try next registry |
| `ErrRegistryNotConfigured` | Requirement names an unconfigured registry | no |
| `ErrChecksumMismatch` | SHA-256 verification failed | pointless / dangerous |
| `ErrNoVersionSatisfies` | No available version meets the constraint | no |
| `ErrManifestInvalid` | `executor.json` failed to parse/validate | no |
| `ErrAlreadyInstalled` | Exact version already in the store | idempotent success |

`NotFoundError{Type,Version}` wraps not-founds with context.

---

## 12. Design decisions (why it is this way)

1. **Provider & Store contracts live consumer-side** in `executor` to keep
   imports acyclic. Providers and stores import `executor`; `executor` imports
   neither.
2. **Instances never resolve/install.** Resolution is frozen at register time;
   restarting a Deployment reuses the pinned artifacts.
3. **Floating requirements always consult the registry.** Installed-first only
   applies to pins and explicit constraints, so "latest" is never stale-locked.
4. **Version selection is resolver-owned**, not provider-owned — a provider's
   ordering is never trusted.
5. **Checksum verification is mandatory** when a digest is declared; an
   expected digest is never bypassed.
6. **Atomic installs**: everything happens in a staging dir inside the store;
   a version only becomes visible via a final rename, and `install.json` is
   rewritten with final paths.
7. **Core executors win** over frozen external artifacts for the same service
   type (built-in first, subprocess fallback).
8. **N.O.R.E. is decoupled from the registry code** — it only knows the shared
   wire types.

---

## 13. Extending

**Add a registry provider**: implement `executor.Provider` (`Name`, `Types`,
`Versions`, `Package`), then register it in `executorctl.BuildCatalog`
(or the config → provider switch). A `.tar.gz`/binary distribution convention
should follow the installer's materialize rules.

**Add a runtime kind**: implement a `runtime.Factory` and register it on the
`Runtime` (e.g. `wasm`, `container`). The manifest's `runtime.type` selects it.

**Add a core (in-process) executor**: register it in
`nore/internal/registry.RegisterCoreServiceExecutors` *before* the plugin pass;
it will shadow any frozen external executor for that service type.

---

## 14. Testing

- Unit: `requirement_test.go` (naming/parse/validate), `selector_test.go`
  (constraints, exact, latest, non-semver safety), `verifier_test.go`
  (SHA-256 formats).
- Integration: `pipeline_test.go` builds local-registry packages in a temp dir
  and runs the full resolve → install → store → frozen-record round trip.
- Config: `loader_test.go` asserts the default registries.
- Run: `go test ./application/...` (and `./nore/...`, `./shared/...`).

---

## 15. File map

| Area | Files |
| --- | --- |
| Wire contract | `shared/types/executor/{manifest,protocol,resolved}.go` |
| Core machinery | `application/executor/{requirement,model,errors,selector,registry,resolver,installer,verifier,manifest,archive,download,installed_record,provider,store_contract}.go` |
| Store | `application/executor/store/{store,filesystem}.go` |
| Registries | `application/executor/source/local/registry.go`, `application/executor/source/github/{client,registry,releases,package}.go` |
| Runtime | `application/executor/runtime/{executor,process}.go` |
| Wiring | `application/internal/executorctl/executorctl.go` |
| CLI | `application/internal/cli/executor/{executor,print}.go` |
| Register flow | `application/internal/cli/register/register.go` |
| Config | `application/config/{config,defaults,loader}.go`, `application/compiler/config.go` |
| N.O.R.E. adapter | `nore/internal/plugin/process.go`, `nore/internal/instance/{instance,manager}.go` |
| Tests | `application/executor/*_test.go` |