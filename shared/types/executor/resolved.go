package executor

// RuntimeInfo is the runtime portion of a frozen resolved executor. It is
// duplicated from the package manifest so the deployment carries everything
// required to launch the artifact without re-reading remote metadata.
type RuntimeInfo struct {
	Type       string `json:"type"`
	Protocol   string `json:"protocol,omitempty"`
	Entrypoint string `json:"entrypoint"`
}

// ResolvedExecutor is the frozen dependency record produced when a System is
// compiled into a Deployment. Authoring a System only declares requirements
// (type + constraint); this record pins the exact resolved version, registry,
// checksum, and installed artifact so a Deployment is reproducible.
//
// N.O.R.E. consumes this record to execute services without ever performing
// dependency resolution or installation.
type ResolvedExecutor struct {
	// Type is the logical executor name (e.g. "github:read").
	Type string `json:"type"`

	// RequestedVersion is the original constraint (e.g. "^1.0.0"). Empty when
	// the requirement was floating.
	RequestedVersion string `json:"requestedVersion,omitempty"`

	// ResolvedVersion is the exact version frozen into the deployment
	// (e.g. "1.2.0").
	ResolvedVersion string `json:"resolvedVersion"`

	// Registry is the registry that supplied the package.
	Registry string `json:"registry,omitempty"`

	// Digest is the content digest of the installed artifact (sha256:...).
	Digest string `json:"digest,omitempty"`

	// Runtime describes how to launch the artifact.
	Runtime RuntimeInfo `json:"runtime"`

	// Capabilities declared by the executor manifest.
	Capabilities []string `json:"capabilities,omitempty"`

	// Services the executor can execute.
	Services []string `json:"services,omitempty"`

	// RootDir is the absolute path of the installed executor directory. The
	// runtime resolves the entrypoint relative to it. It points into the
	// local executor store (~/.neuron/executors/...) and is only meaningful
	// on the host that created the deployment.
	RootDir string `json:"rootDir,omitempty"`
}

// EntrypointPath returns the absolute path to the executor's entrypoint.
func (r *ResolvedExecutor) EntrypointPath() string {
	if r.RootDir == "" || r.Runtime.Entrypoint == "" {
		return r.Runtime.Entrypoint
	}
	return r.RootDir + "/" + r.Runtime.Entrypoint
}
