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

	// The standalone executor.json is optional when the version directory
	// ships an executor package archive; the archive's inner manifest is then
	// authoritative and reconciled at install time.
	manifestPath := filepath.Join(versionDir, shadexec.ManifestFile)
	data, err := os.ReadFile(manifestPath)
	var m *shadexec.Manifest
	if err == nil {
		m, err = executor.ParseManifest(data)
		if err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	pkg := &executor.Package{Type: typ, Version: version, Registry: r.Name()}
	if m != nil {
		pkg.Manifest = data
		pkg.Description = m.Metadata.Description
		pkg.Runtime = executor.RuntimeSpec{
			Type:       m.Runtime.Type,
			Entrypoint: m.Runtime.Entrypoint,
			Protocol:   m.Runtime.Protocol,
			MaxWorkers: m.Runtime.MaxWorkers,
		}
		pkg.Capabilities = m.Capabilities
		pkg.Services = m.Services
		pkg.Platforms = m.Platforms
	}

	// The executor package archive is the preferred payload: one file that
	// installs the whole executor. Prefer the exact <type>-<version> name,
	// then any canonical archive in the version directory.
	if archive, ok := archivePayload(versionDir, typ, version); ok {
		pkg.Artifact = executor.Artifact{
			URL:  "file://" + filepath.ToSlash(archive),
			Name: filepath.Base(archive),
		}
		return pkg, nil
	}

	if m == nil {
		return nil, executor.ErrNotFound
	}

	// No archive: bind the platform artifact for the executor's runtime kind.
	if platform, ok := m.Platforms[executor.PlatformForRuntime(m.Runtime.Type)]; ok && platform.Artifact != "" {
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

// archivePayload finds the canonical executor package archive in dir,
// preferring <type>-<version>-executor.neuron.tar.gz then any asset carrying
// the canonical package archive suffix.
func archivePayload(dir, typ, version string) (string, bool) {
	exact := filepath.Join(dir, shadexec.PackageArchiveName(typ, version))
	if info, err := os.Stat(exact); err == nil && !info.IsDir() {
		return exact, true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), shadexec.PackageArchiveSuffix) {
			return filepath.Join(dir, e.Name()), true
		}
	}
	return "", false
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
