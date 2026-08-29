# Neuron CLI

The `neuron` command-line interface is the entry point for interacting with the Neuron workflow engine and its local N.O.R.E. runtime.

## Usage

```
neuron [command] [flags]
```

Run `neuron --help` for the full command list, or `neuron <command> --help` for details on a specific command.

## Global Flags

These flags apply to every command.

| Flag                     | Description                                                       |
| ------------------------ | ----------------------------------------------------------------- |
| `--config <file>`        | Path to a config file (default: `./neuron.yaml`)                  |
| `--log-level <level>`    | System logging level (default: `info`)                            |
| `--remote <endpoint>`    | Remote N.O.R.E. endpoint, e.g. `https://api.nore.example.com`     |
| `-v, --verbose`          | Enable verbose output                                             |
| `-h, --help`             | Show help for the command                                         |

## Commands

### `build`

Build the neuron system.

```
neuron build [flags]
```

| Flag          | Description                           |
| ------------- | ------------------------------------- |
| `-w, --watch` | Watch for file changes and rebuild    |
| `-r, --root`  | Set the project root directory        |

### `daemon`

Manage the local N.O.R.E. daemon.

```
neuron daemon [command]
```

Running `neuron daemon` without a subcommand prints the command help.

#### `daemon stop`

Stop the background N.O.R.E. daemon.

```
neuron daemon stop
```

### `execution`

List all execution instances.

```
neuron execution
```

### `init`

Initialize a new Neuron workspace.

```
neuron init [Target] [flags]
```

| Argument   | Description                       |
| ---------- | --------------------------------- |
| `Target`   | Optional target directory to init |

### `instance`

Manage N.O.R.E. system instances.

```
neuron instance [command]
```

Running `neuron instance` without a subcommand prints the command help.

#### `instance list`

List instances, or list the executions of a specific instance.

With no argument, this lists instances. Provide an instance ID either as a positional argument or via `--target` to list that instance's executions.

```
neuron instance list [instance-id] [flags]
```

Examples:

```
# List running instances
neuron instance list

# List all instances (including inactive ones)
neuron instance list --all

# List instances filtered by status
neuron instance list --status running

# List executions of a specific instance (positional argument)
neuron instance list <instance-id>

# List executions of a specific instance (flag)
neuron instance list --target <instance-id>
neuron instance list -t <instance-id>
```

> `--all`, `--status`, and `--target` are mutually exclusive — only one may be used at a time. An instance ID given as both a positional argument and `--target` is an error.

| Flag                  | Description                                            |
| --------------------- | ------------------------------------------------------ |
| `-a, --all`           | List all instances including inactive ones             |
| `-s, --status <state>`| Filter instances by status (e.g. `running`, `stopped`) |
| `-t, --target <id>`   | List executions of the instance with the given ID      |

### `run`

Run a registered Neuron System. The internal N.O.R.E. runtime executes the system.

`run` executes only: it does **not** build, resolve, or parse the project. The system must be registered first with `neuron register`; the registration key is read from `.neuron/register.json` in the current directory. Running without a prior registration fails with a message pointing to `neuron register`.

By default `run` streams live execution events; pass `--detach` to just print the execution handles.

```
neuron run [flags]
```

| Flag             | Description                                             |
| ---------------- | ------------------------------------------------------- |
| `-v, --verbose`  | Verbose output, includes event payloads in the stream   |
| `--input <json>` | JSON input payload for execution                        |
| `--detach`       | Print execution handles immediately, no event streaming |

### `register`

Build, resolve, and register the current project with N.O.R.E. Registration stores the system durably (keyed by `systemID:version:hash:env`) without creating an instance; instances are created lazily on first execution.

On success the server-confirmed registration key is written to `.neuron/register.json`, which `neuron run` reads.

```
neuron register [flags]
```

There are no `register`-specific flags; use `--remote` to target a N.O.R.E. endpoint and `--config` to select a config file.

### `completion`

Generate an autocompletion script for the specified shell.

```
neuron completion [bash|zsh|fish|powershell]
```