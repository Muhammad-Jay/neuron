package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Options controls project resolution.
//
// These options are intentionally project-level options.
// Runtime/N.O.R.E. options belong elsewhere.
type Options struct {
	// ProjectRoot is the project directory.
	//
	// If empty, the current working directory is used.
	ProjectRoot string

	// WriteArtifact controls whether the resolved project is
	// persisted into .neuron/resolved/.
	WriteArtifact bool

	// CleanArtifact controls whether the previous resolved
	// project is removed before resolving.
	CleanArtifact bool

	// Validate controls basic source validation.
	Validate bool

	// Verbose enables diagnostic information for callers that
	// want to expose resolution details.
	//
	// The project package does not print anything itself.
	Verbose bool
}

// DefaultOptions returns sensible MVP defaults.
func DefaultOptions() Options {
	return Options{
		WriteArtifact: true,
		CleanArtifact: true,
		Validate:      true,
	}
}

// Result contains the result of source resolution.
type Result struct {
	Project *ResolvedProject
}

// Resolve reads and resolves the project source.
//
// It does not load .neuron/resolved.
// It does not parse into core.System.
// It does not start N.O.R.E.
// It does not install executors.
func Resolve(
	ctx context.Context,
	opts Options,
) (*Result, error) {

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	root := opts.ProjectRoot

	if root == "" {
		var err error

		root, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf(
				"get current directory: %w",
				err,
			)
		}
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve project root: %w",
			err,
		)
	}

	root = filepath.Clean(root)

	if opts.CleanArtifact {
		if err := removeResolvedArtifact(root); err != nil {
			return nil, err
		}
	}

	resolver, err := NewResolver(root)
	if err != nil {
		return nil, err
	}

	resolved, err := resolver.ResolveProject()
	if err != nil {
		return nil, err
	}

	if opts.WriteArtifact {
		if err := SaveResolved(
			root,
			resolved,
		); err != nil {
			return nil, err
		}
	}

	return &Result{
		Project: resolved,
	}, nil
}

// LoadResolvedProject only loads the generated artifact.
//
// It does not read source YAML files.
func LoadResolvedProject(
	projectRoot string,
) (*ResolvedProject, error) {

	return LoadResolved(projectRoot)
}

func removeResolvedArtifact(
	projectRoot string,
) error {

	dir := ArtifactPath(projectRoot)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf(
			"remove previous resolved project: %w",
			err,
		)
	}

	return nil
}