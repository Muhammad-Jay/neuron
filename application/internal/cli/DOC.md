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

Run a Neuron System. The internal N.O.R.E. runtime executes the system.

```
neuron run [flags]
```

| Flag          | Description      |
| ------------- | ---------------- |
| `-v, --verbose` | Verbose output |

### `completion`

Generate an autocompletion script for the specified shell.

```
neuron completion [bash|zsh|fish|powershell]
```