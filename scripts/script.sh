#!/usr/bin/env bash

set -o pipefail

for mod in ./application ./shared ./nore; do
    echo "Processing $mod..."
    (
        cd "$mod" || exit 1
        go mod tidy
        go test ./... &&
        go build ./... &&
        echo "$mod built successfully" || echo "$mod build failed"
    ) || exit 1
done

echo "Processing packages/sdk..."
(
    pnpm typecheck:sdk &&
    pnpm test:sdk &&
    pnpm build:sdk &&
    echo "sdk built successfully"
) || exit 1

echo "All tests and builds passed successfully!"