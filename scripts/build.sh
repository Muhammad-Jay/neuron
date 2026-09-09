#!/usr/bin/env bash
#
# build.sh — Build Neuron release archives for all supported platforms.
#
# Usage:
#   scripts/build.sh [version]
#
# If no version is provided, the script reads it from the git tag or defaults
# to "dev". The version is injected into both binaries via -ldflags.
#
# Output:
#   dist/
#     neuron_<version>_<os>_<arch>.tar.gz   (linux, darwin)
#     neuron_<version>_<os>_<arch>.zip      (windows)
#     SHA256SUMS

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="${REPO_ROOT}/dist"

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    # Try to derive from the current git tag, fall back to "dev".
    VERSION="$(git describe --tags --exact-match 2>/dev/null || echo "")"
    VERSION="${VERSION#v}"  # strip leading v if present
    if [ -z "$VERSION" ]; then
        VERSION="dev"
    fi
fi

echo "Building Neuron ${VERSION}"

# Module entry points.
NEURON_CMD="./application/cmd/neuron"
NORE_CMD="./nore/cmd/nore"

# Platforms to build for.
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

LDFLAGS="-s -w -X main.Version=${VERSION}"

# Clean previous output.
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

cd "${REPO_ROOT}"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r goos goarch <<< "$platform"

    suffix=""
    if [ "$goos" = "windows" ]; then
        suffix=".exe"
    fi

    archive_name="neuron_${VERSION}_${goos}_${goarch}"
    archive_dir="${DIST_DIR}/${archive_name}"
    mkdir -p "${archive_dir}"

    echo "  Building ${goos}/${goarch}..."

    # Build neuron CLI.
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -ldflags "${LDFLAGS}" -o "${archive_dir}/neuron${suffix}" "${NEURON_CMD}"

    # Build nore daemon.
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -ldflags "${LDFLAGS}" -o "${archive_dir}/nore${suffix}" "${NORE_CMD}"

    # Create archive.
    if [ "$goos" = "windows" ]; then
        (cd "${DIST_DIR}" && zip -qr "${archive_name}.zip" "${archive_name}")
    else
        (cd "${DIST_DIR}" && tar czf "${archive_name}.tar.gz" "${archive_name}")
    fi

    rm -rf "${archive_dir}"
done

# Generate checksums.
echo "  Generating SHA256SUMS..."
(cd "${DIST_DIR}" && sha256sum -- * > SHA256SUMS)

echo "Done. Output in ${DIST_DIR}/"
ls -lh "${DIST_DIR}/"
