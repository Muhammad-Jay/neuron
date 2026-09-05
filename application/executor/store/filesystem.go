package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/executor"
)

// FilesystemStore is a Store rooted at a local directory, by default
// ~/.neuron/executors.
type FilesystemStore struct {
	root string
}

// NewFilesystemStore validates and returns a filesystem store rooted at root.
func NewFilesystemStore(root string) (*FilesystemStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("executor store root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve executor store root: %w", err)
	}
	return &FilesystemStore{root: filepath.Clean(abs)}, nil
}

func (s *FilesystemStore) Root() string {
	return s.root
}

// DefaultStoreDir returns the default executor store directory
// (~/.neuron/executors).
func DefaultStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".neuron/executors"
	}
	return filepath.Join(home, ".neuron", "executors")
}

func (s *FilesystemStore) Stage() (string, error) {
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return "", fmt.Errorf("create executor store root: %w", err)
	}
	return os.MkdirTemp(s.root, ".tmp-*")
}

func (s *FilesystemStore) Commit(ctx context.Context, stage, typ, version string) (*executor.Installed, error) {
	rec, err := executor.ReadInstallRecord(filepath.Join(stage, executor.InstallFile))
	if err != nil {
		return nil, fmt.Errorf("installed record missing in staging dir: %w", err)
	}

	final, err := s.installedDir(typ, version)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(stage); err != nil {
		return nil, fmt.Errorf("staging dir missing: %w", err)
	}
	if _, err := os.Stat(final); err == nil {
		return nil, fmt.Errorf("%w: %s@%s", executor.ErrAlreadyInstalled, typ, version)
	}

	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		return nil, fmt.Errorf("create executor parent directory: %w", err)
	}

	if err := os.Rename(stage, final); err != nil {
		return nil, fmt.Errorf("commit executor %s@%s: %w", typ, version, err)
	}

	// The install record was written against the staging directory. Rewrite it
	// in place with the final paths so later reads return a correct record.
	installed := rec.Installed()
	installed.RootDir = final
	installed.ManifestPath = filepath.Join(final, executor.ManifestFile)
	installed.ArtifactPath = final

	if err := executor.WriteInstallRecord(filepath.Join(final, executor.InstallFile), executor.RecordFor(installed)); err != nil {
		return nil, fmt.Errorf("rewrite install record: %w", err)
	}

	return installed, nil
}

func (s *FilesystemStore) Get(ctx context.Context, typ, version string) (*executor.Installed, error) {
	dir, err := s.installedDir(typ, version)
	if err != nil {
		return nil, err
	}
	recPath := filepath.Join(dir, executor.InstallFile)
	if _, err := os.Stat(recPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s@%s", executor.ErrNotFound, typ, version)
		}
		return nil, err
	}
	rec, err := executor.ReadInstallRecord(recPath)
	if err != nil {
		return nil, err
	}
	return rec.Installed(), nil
}

func (s *FilesystemStore) List(ctx context.Context, typ string) ([]executor.Installed, error) {
	var out []executor.Installed

	if strings.TrimSpace(typ) == "" {
		// List everything: find every install.json under the store root.
		for _, dir := range s.walkInstallDirs(s.root) {
			rec, err := executor.ReadInstallRecord(filepath.Join(dir, executor.InstallFile))
			if err != nil {
				continue
			}
			rec.Installed().RootDir = dir
			out = append(out, *rec.Installed())
		}
	} else {
		base, err := s.typeDir(typ)
		if err != nil {
			return nil, err
		}
		for _, versionDir := range s.listVersionDirs(base) {
			recPath := filepath.Join(versionDir, executor.InstallFile)
			rec, err := executor.ReadInstallRecord(recPath)
			if err != nil {
				continue
			}
			out = append(out, *rec.Installed())
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Version > out[j].Version
	})
	return out, nil
}

// walkInstallDirs returns every directory below root that directly contains an
// install.json record, flattening the <root>/owner/.../version tree.
func (s *FilesystemStore) walkInstallDirs(root string) []string {
	var out []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || info.Name() != executor.InstallFile {
			return nil
		}
		out = append(out, filepath.Dir(path))
		return nil
	})
	return out
}

func (s *FilesystemStore) Remove(ctx context.Context, typ, version string) error {
	dir, err := s.installedDir(typ, version)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s@%s", executor.ErrNotFound, typ, version)
		}
		return err
	}
	return os.RemoveAll(dir)
}

// installedDir returns the final directory for an exact (type, version).
func (s *FilesystemStore) installedDir(typ, version string) (string, error) {
	base, err := s.typeDir(typ)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(version) == "" {
		return "", fmt.Errorf("executor version is required")
	}
	return filepath.Join(base, strings.TrimSpace(version)), nil
}

// typeDir maps a logical executor name to its store branch:
// "github:read" -> <root>/github/read.
func (s *FilesystemStore) typeDir(typ string) (string, error) {
	path, err := executor.TypePath(typ)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.root, path), nil
}

func (s *FilesystemStore) listVersionDirs(base string) []string {
	entries, err := os.ReadDir(base)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirs = append(dirs, filepath.Join(base, e.Name()))
	}
	return dirs
}
