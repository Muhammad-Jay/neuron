// Package local implements a directory-backed executor registry. It exists so
// the full resolve → version-select → package → install pipeline works offline
// and in tests without a network, and it demonstrates that registries are
// pluggable: "local" is just another source.Registry.
//
// Layout (mirrors the installed store, minus install.json):
//
//	<root>/github/read/1.2.0/
//	    executor.json
//	    github-read-linux-amd64        <- optional platform binary/archive
package local

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/executor"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// Registry serves executor packages laid out under a local directory.
type Registry struct {
	root string
}

// New validates and returns a local registry rooted at root.
func New(root string) (*Registry, error) {
	if strings.TrimSpace(root) == "" {
		return nil, executor.ErrNotFound // malformed config: no packages available
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(abs); err != nil || !info.IsDir() {
		return nil, executor.ErrNotFound
	}
	return &Registry{root: abs}, nil
}

func (r *Registry) Name() string {
	return "local"
}

func (r *Registry) Types(ctx context.Context) ([]string, error) {
	var types []string
	for _, manifestPath := range r.findManifests(r.root) {
		rel, _ := filepath.Rel(r.root, manifestPath)
		dir := filepath.Dir(rel)
		segments := strings.Split(filepath.ToSlash(dir), "/")
		if len(segments) < 2 {
			continue
		}
		typ := strings.Join(segments[:len(segments)-1], ":")
		types = append(types, typ)
	}
	sort.Strings(types)
	return dedupe(types), nil
}

func (r *Registry) Versions(ctx context.Context, typ string) ([]string, error) {
	dir, err := r.typeDir(typ)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, executor.ErrNotFound
		}
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	return versions, nil
}

func (r *Registry) Package(ctx context.Context, typ, version string) (*executor.Package, error) {
	versionDir, err := r.versionDir(typ, version)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(versionDir, shadexec.ManifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, executor.ErrNotFound
		}
		return nil, err
	}

	m, err := executor.ParseManifest(data)
	if err != nil {
		return nil, err
	}

	pkg := executor.PackageFromManifest(m, r.Name(), version)
	pkg.Manifest = data

	// Bind the host-platform artifact if the package directory ships one.
	if platform, ok := m.Platforms[executor.HostPlatform()]; ok && platform.Artifact != "" {
		artifactPath := filepath.Join(versionDir, platform.Artifact)
		if info, err := os.Stat(artifactPath); err == nil && !info.IsDir() {
			pkg.Artifact = executor.Artifact{
				URL:    "file://" + filepath.ToSlash(artifactPath),
				Name:   platform.Artifact,
				SHA256: platform.SHA256,
			}
		}
	}

	return pkg, nil
}

func (r *Registry) typeDir(typ string) (string, error) {
	path, err := executor.TypePath(typ)
	if err != nil {
		return "", err
	}
	return filepath.Join(r.root, path), nil
}

func (r *Registry) versionDir(typ, version string) (string, error) {
	base, err := r.typeDir(typ)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, version), nil
}

// findManifests walks root and returns every executor.json path with at least
// two parent segments (owner/.../version/executor.json).
func (r *Registry) findManifests(root string) []string {
	var out []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() != shadexec.ManifestFile {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out
}

func dedupe(in []string) []string {
	var out []string
	for i, v := range in {
		if i > 0 && v == in[i-1] {
			continue
		}
		out = append(out, v)
	}
	return out
}
