// Package executor defines the wire contract shared between N.O.R.E. and the
// Neuron CLI for distributing and executing external executor packages.
//
// The contract is deliberately minimal. It answers three questions:
//
//  1. What is this artifact? (executor.json manifest)
//  2. How do I talk to the running executor?  (process/stdin protocol)
//  3. What was resolved/frozen for a Deployment? (ResolvedExecutor)
//
// Registry, resolution, and installation machinery never lives here. Those
// belong to application/executor. This package only declares the schema both
// sides agree on so the runtime never needs to import registry code.
package executor

import "fmt"

// Protocol identifiers for the executor runtime handshake.
const (
	// APIVersion is the schema version of every executor.json manifest.
	APIVersion = "neuron/v1"

	// Kind is the discriminant of every executor.json manifest.
	Kind = "Executor"

	// ProtocolV1 is the canonical wire protocol spoken by executor processes
	// over gRPC (Unix domain sockets). Executors must declare this (or a
	// newer compatible protocol) in their manifest runtime.protocol.
	ProtocolV1 = "neuron/executor-v1"

	// ProtocolJSONV1 is the legacy stdin/stdout JSON protocol spoken by
	// one-shot executor processes and WASI modules (WASI has no socket
	// interface, so gRPC is unavailable there). Executors that cannot host a
	// gRPC server MUST declare this protocol in their manifest.
	ProtocolJSONV1 = "neuron/executor-v1-json"

	// ManifestFile is the mandatory manifest file inside every executor
	// package and every installed executor directory.
	ManifestFile = "executor.json"

	// InstallFile is the installation record written into an installed
	// executor directory after a verified, atomic install.
	InstallFile = "install.json"
)

// Manifest is the executor.json package manifest. It describes one immutable
// executor artifact: what it is, what runtime launches it, which services it
// can execute, and which platform binaries back it.
type Manifest struct {
	APIVersion   string              `json:"apiVersion"`
	Kind         string              `json:"kind"`
	Metadata     ManifestMetadata    `json:"metadata"`
	Runtime      ManifestRuntime     `json:"runtime"`
	Services     []string            `json:"services"`
	Capabilities []string            `json:"capabilities,omitempty"`
	Platforms    map[string]Platform `json:"platforms"`
}

// ManifestMetadata identifies the executor artifact.
type ManifestMetadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// ManifestRuntime describes how the executor is launched.
type ManifestRuntime struct {
	// Type is the runtime adapter kind: "process" today, "wasm", "container",
	// "remote", ... in the future. The runtime dispatches on this value.
	Type string `json:"type"`

	// Entrypoint is the location of the executable relative to the installed
	// package root. For platform-specific artifacts the filesystem store
	// resolves the artifact name before writing it.
	Entrypoint string `json:"entrypoint"`

	// Protocol is the wire protocol spoken by the executor
	// (neuron/executor-v1 for gRPC process executors,
	// neuron/executor-v1-json for stdin/stdout JSON executors).
	Protocol string `json:"protocol,omitempty"`

	// MaxWorkers bounds the number of concurrent worker processes the runtime
	// may spawn for this executor. A value of 0 means the runtime default.
	// Only meaningful for runtime types with long-lived workers (process).
	MaxWorkers int `json:"maxWorkers,omitempty"`
}

// Platform maps one host platform (GOOS-GOARCH) to its artifact.
type Platform struct {
	// Artifact is the binary/archive file name for this platform.
	Artifact string `json:"artifact"`

	// SHA256 is the checksum of the artifact. Required when the artifact is
	// distributed as a release asset.
	SHA256 string `json:"sha256,omitempty"`
}

// Validate verifies the manifest conforms to the executor.json schema.
func (m *Manifest) Validate() error {
	if m == nil {
		return fmt.Errorf("executor manifest is nil")
	}
	if m.APIVersion != APIVersion {
		return fmt.Errorf("executor manifest: unsupported apiVersion %q (want %q)", m.APIVersion, APIVersion)
	}
	if m.Kind != Kind {
		return fmt.Errorf("executor manifest: unsupported kind %q (want %q)", m.Kind, Kind)
	}
	if m.Metadata.Name == "" {
		return fmt.Errorf("executor manifest: metadata.name is required")
	}
	if m.Metadata.Version == "" {
		return fmt.Errorf("executor manifest: metadata.version is required")
	}
	if m.Runtime.Type == "" {
		return fmt.Errorf("executor manifest %s: runtime.type is required", m.Metadata.Name)
	}
	if m.Runtime.Entrypoint == "" {
		return fmt.Errorf("executor manifest %s: runtime.entrypoint is required", m.Metadata.Name)
	}
	if len(m.Services) == 0 {
		return fmt.Errorf("executor manifest %s: at least one service is required", m.Metadata.Name)
	}
	return nil
}

// HasCapability reports whether the manifest declares cap and an empty
// capability list answers false (capabilities are opt-in declarations).
func (m *Manifest) HasCapability(cap string) bool {
	for _, c := range m.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}
