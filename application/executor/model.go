package executor

import (
	"runtime"

	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// Package is the immutable package obtained from a registry. It describes a
// single resolvable version of one executor, before installation.
type Package struct {
	// Type is the logical executor name (e.g. "github:read").
	Type string

	// Version is the exact package version.
	Version string

	// Digest is the content digest of the artifact (sha256:...).
	Digest string

	// Registry is the registry that supplied this package.
	Registry string

	// Description from the package manifest.
	Description string

	// Runtime describes how the installed artifact is launched.
	Runtime RuntimeSpec

	// Capabilities declared by the package manifest.
	Capabilities []string

	// Services the package can execute.
	Services []string

	// Platforms declared by the package manifest (GOOS-GOARCH -> artifact).
	Platforms map[string]shadexec.Platform

	// Artifact is the platform-matching download, when the package is
	// distributed as a remote binary.
	Artifact Artifact

	// Manifest is the validated executor.json bytes for this package. The
	// installer writes it into the installed directory so an installed
	// executor is fully self-describing.
	Manifest []byte
}

// RuntimeSpec describes how an executor artifact is launched.
type RuntimeSpec struct {
	// Type is the runtime adapter kind: "process" today; "wasm", "container",
	// "remote" in the future.
	Type string

	// Entrypoint is the executable path relative to the installed root.
	Entrypoint string

	// Protocol is the wire protocol the executor speaks
	// (neuron/executor-v1 for gRPC process executors,
	// neuron/executor-v1-json for stdin/stdout JSON executors).
	Protocol string

	// MaxWorkers bounds the number of concurrent worker processes the runtime
	// may spawn for this executor. A value of 0 means the runtime default.
	// Only meaningful for runtime types with long-lived workers (process).
	MaxWorkers int
}

// Artifact is a downloadable binary for one platform.
type Artifact struct {
	// URL is the download location (release asset URL, local path, ...).
	URL string

	// Name is the artifact's base file name (e.g. "github-read-linux-amd64").
	// When empty the installer derives it from URL.
	Name string

	// SHA256 is the expected checksum of the artifact bytes.
	SHA256 string
}

// HostPlatform returns the GOOS-GOARCH key used to select a platform artifact
// from an executor manifest (e.g. "linux-amd64").
func HostPlatform() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

// Installed is a locally installed, immutable executor artifact living in the
// executor store (~/.neuron/executors/...). It is the materialized result of
// installing a Package: a verified, atomically renamed directory.
type Installed struct {
	// Type is the logical executor name.
	Type string

	// Version is the exact installed version.
	Version string

	// Digest is the content digest recorded at install time.
	Digest string

	// Registry is the registry that supplied the package.
	Registry string

	// Platform is the GOOS-GOARCH the artifact was installed for.
	Platform string

	// RootDir is the absolute installed directory.
	RootDir string

	// ManifestPath is the absolute path to executor.json.
	ManifestPath string

	// ArtifactPath is the absolute path to the extracted artifact root (the
	// directory containing the entrypoint).
	ArtifactPath string

	// Runtime mirrors the manifest launch spec.
	Runtime RuntimeSpec

	// Capabilities declared by the manifest.
	Capabilities []string

	// Services the installed executor can execute.
	Services []string
}

// Manifest loads and validates the installed executor.json.
func (i *Installed) Manifest() (*shadexec.Manifest, error) {
	return ReadManifestFile(i.ManifestPath)
}

// Frozen converts an installed executor into the wire record a Deployment
// persists. It records the requested constraint alongside the exact resolved
// version, so restarting a Deployment reuses the pinned artifact rather than
// whatever is latest tomorrow.
func (i *Installed) Frozen(requestedVersion string) *shadexec.ResolvedExecutor {
	return &shadexec.ResolvedExecutor{
		Type:             i.Type,
		RequestedVersion: requestedVersion,
		ResolvedVersion:  i.Version,
		Registry:         i.Registry,
		Digest:           i.Digest,
		Runtime: shadexec.RuntimeInfo{
			Type:       i.Runtime.Type,
			Protocol:   i.Runtime.Protocol,
			Entrypoint: i.Runtime.Entrypoint,
			MaxWorkers: i.Runtime.MaxWorkers,
		},
		Capabilities: i.Capabilities,
		Services:     i.Services,
		RootDir:      i.RootDir,
	}
}

// Environment is the resolved execution environment handed to an Instance: the
// compiled system plus the exact executor artifacts its services require. An
// Instance never performs dependency resolution or installation.
type Environment struct {
	// Executors is the frozen, installed executor set.
	Executors []*Installed
}

// Resolved returns the frozen wire records for the environment.
func (e *Environment) Resolved(requested map[string]string) []*shadexec.ResolvedExecutor {
	out := make([]*shadexec.ResolvedExecutor, 0, len(e.Executors))
	for _, installed := range e.Executors {
		out = append(out, installed.Frozen(requested[installed.Type]))
	}
	return out
}
