# Installation

## Download

Download the latest release from [GitHub Releases](https://github.com/Muhammad-Jay/neuron/releases).

Each release provides archives for:

| Platform | Architecture | Archive |
|----------|-------------|---------|
| Linux | amd64 | `neuron_<version>_linux_amd64.tar.gz` |
| Linux | arm64 | `neuron_<version>_linux_arm64.tar.gz` |
| macOS | amd64 | `neuron_<version>_darwin_amd64.tar.gz` |
| macOS | arm64 | `neuron_<version>_darwin_arm64.tar.gz` |
| Windows | amd64 | `neuron_<version>_windows_amd64.zip` |

Each archive contains both `neuron` (the CLI) and `nore` (the runtime daemon).

Every release includes a `SHA256SUMS` file for verifying download integrity.

---

## Install

### Linux / macOS

```bash
# Extract the archive
tar xzf neuron_<version>_<os>_<arch>.tar.gz

# Move both binaries to a directory on your PATH
sudo mv neuron_<version>_<os>_<arch>/neuron /usr/local/bin/
sudo mv neuron_<version>_<os>_<arch>/nore /usr/local/bin/
```

### Windows

```powershell
# Extract the archive
Expand-Archive neuron_<version>_windows_amd64.zip -DestinationPath neuron

# Move both executables to a directory on your PATH
Move-Item neuron\neuron.exe neuron\nore.exe $env:LOCALAPPDATA\Microsoft\WinGet\Links\
```

---

## Verify

```bash
neuron --version
# neuron 0.1.0

nore --version
# nore 0.1.0
```

---

## Verify Checksums

Download the `SHA256SUMS` file from the same release and verify:

```bash
sha256sum -c SHA256SUMS --ignore-missing
```

All archives should report `OK`.

---

## Platform Notes

### Linux

Neuron is built as a static binary with `CGO_ENABLED=0`. No system libraries are required.

### macOS

On macOS, you may need to allow the binary to run:

```bash
xattr -d com.apple.quarantine /usr/local/bin/neuron
xattr -d com.apple.quarantine /usr/local/bin/nore
```

### Windows

Windows builds produce `.exe` executables. Both `neuron.exe` and `nore.exe` must be on your `PATH`.

---

## Data Directory

N.O.R.E. stores data in `~/.neuron/nore/` by default. This can be changed with the `--data-dir` flag.

The Unix socket for local CLI communication is placed at `~/.neuron/nore.sock` by default. This can be changed with the `--socket` flag or the `NEURON_SOCKET` environment variable.

---

## From Source

If you prefer to build from source:

```bash
git clone https://github.com/Muhammad-Jay/neuron.git
cd neuron

# Build both binaries
go build -o neuron ./application/cmd/neuron
go build -o nore ./nore/cmd/nore
```

Or use the build script to produce release archives:

```bash
bash scripts/build.sh
```

Output will be in the `dist/` directory.
