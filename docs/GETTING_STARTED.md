# Getting Started

This guide walks you through installing Neuron and running your first system.

---

## Prerequisites

- **Go 1.26+** (for building from source) or a release binary
- **Node.js 22+** and **pnpm** (only if using the TypeScript SDK)

---

## Installation

### From a Release

Download the archive for your platform from the [GitHub Releases](https://github.com/Muhammad-Jay/neuron/releases) page.

Each archive contains both `neuron` and `nore`:

```bash
# Linux amd64
tar xzf neuron_0.1.0_linux_amd64.tar.gz
cd neuron_0.1.0_linux_amd64
sudo mv neuron nore /usr/local/bin/
```

```bash
# macOS arm64
tar xzf neuron_0.1.0_darwin_arm64.tar.gz
cd neuron_0.1.0_darwin_arm64
sudo mv neuron nore /usr/local/bin/
```

```powershell
# Windows amd64
Expand-Archive neuron_0.1.0_windows_amd64.zip -DestinationPath neuron
Move-Item neuron\neuron.exe neuron\nore.exe $env:LOCALAPPDATA\Microsoft\WinGet\Links\
```

Verify:

```bash
neuron --version
nore --version
```

### From Source

```bash
git clone https://github.com/Muhammad-Jay/neuron.git
cd neuron
go build ./application/cmd/neuron
go build ./nore/cmd/nore
```

---

## Verify Installation

```bash
neuron --version
# neuron 0.1.0

nore --version
# nore 0.1.0
```

---

## Create a Project

```bash
mkdir my-first-system
cd my-first-system
neuron init
```

This creates a `neuron.yaml` project file.

---

## Define a System

Create a system definition. For YAML projects, the structure is:

```
my-first-system/
├── neuron.yaml                          Project configuration
├── systems/
│   └── my-system/
│       └── system.yaml                  System definition
└── services/
    └── hello.yaml                       Service definition
```

A minimal service (`services/hello.yaml`):

```yaml
apiVersion: neuron/v1
kind: Service
metadata:
  name: hello
  version: 1.0.0
spec:
  executor:
    type: neuron:core:set
  config:
    message: "Hello from Neuron"
```

A minimal system (`systems/my-system/system.yaml`):

```yaml
apiVersion: neuron/v1
kind: System
metadata:
  name: my-system
  version: 1.0.0
services:
  - ref: hello
    entry: ../../services/hello.yaml
```

---

## Start N.O.R.E.

In a separate terminal:

```bash
nore
```

N.O.R.E. starts listening on TCP `:7432` and a Unix socket at `~/.neuron/nore.sock`.

---

## Register the System

```bash
neuron register --root .
```

This builds the project, compiles the manifest, resolves executors, and registers the system with N.O.R.E.

---

## Run the System

```bash
neuron run my-system
```

---

## What Happened

1. The CLI loaded your project definition.
2. The compiler transformed it into a canonical system manifest.
3. Executor requirements were resolved (core `set` executor was found).
4. The system was registered with N.O.R.E.
5. N.O.R.E. created an instance and executed the service.
6. The result was returned.

---

## Next Steps

- Read the [Architecture](./ARCHITECTURE.md) document to understand how Neuron works internally.
- Explore the [examples](../examples/) directory for YAML, TypeScript, and Go projects.
- See the [TypeScript SDK](../packages/sdk/README.md) for defining systems programmatically.
- Read the [Executor documentation](./executors/) for writing custom executors.
