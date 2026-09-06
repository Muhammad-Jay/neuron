#!/usr/bin/env bash
# Builds the example echo executors into the local registry catalog layout
# consumed by `neuron executor install` and `neuron register`:
#
#   catalog/example/echo/1.0.0/         process-runtime executor
#   catalog/example/echo-wasm/1.0.0/    wasm-runtime executor
#
# Both are compiled from the same source in ./echo. Generated binaries are
# gitignored; re-run this script after changing the source or to rebuild.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$ROOT/echo"
CATALOG="$ROOT/catalog"
VERSION="1.0.0"

# The echo example is a standalone stdlib-only module; ignore the workspace so
# it builds offline and independently of the repo workspace.
export GOWORK=off

echo_dir="$CATALOG/example/echo/$VERSION"
wasm_dir="$CATALOG/example/echo-wasm/$VERSION"
mkdir -p "$echo_dir" "$wasm_dir"

GOOS="$(go env GOOS)"
GOARCH="$(go env GOARCH)"

# Native binary for the process runtime.
go build -C "$SRC" -o "$echo_dir/echo" .

# Portability-proven WASI module for the wasm runtime.
GOOS=wasip1 GOARCH=wasm go build -C "$SRC" -o "$wasm_dir/echo.wasm" .

cat > "$echo_dir/executor.json" <<EOF
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": {
    "name": "example:echo",
    "version": "$VERSION",
    "description": "Echoes the execution input and captures NEURON_EXECUTOR_* env vars."
  },
  "runtime": { "type": "process", "entrypoint": "echo", "protocol": "neuron/executor-v1" },
  "services": ["example:echo"],
  "capabilities": [],
  "platforms": { "$GOOS-$GOARCH": { "artifact": "echo" } }
}
EOF

cat > "$wasm_dir/executor.json" <<EOF
{
  "apiVersion": "neuron/v1",
  "kind": "Executor",
  "metadata": {
    "name": "example:echo-wasm",
    "version": "$VERSION",
    "description": "Echoes the execution input as a WASI module and captures NEURON_EXECUTOR_* env vars."
  },
  "runtime": { "type": "wasm", "entrypoint": "echo.wasm", "protocol": "neuron/executor-v1" },
  "services": ["example:echo-wasm"],
  "capabilities": [],
  "platforms": {
    "linux-amd64":   { "artifact": "echo.wasm" },
    "linux-arm64":   { "artifact": "echo.wasm" },
    "darwin-amd64":  { "artifact": "echo.wasm" },
    "darwin-arm64":  { "artifact": "echo.wasm" },
    "windows-amd64": { "artifact": "echo.wasm" }
  }
}
EOF

echo "Built example executors into $CATALOG"