package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// Installer turns a resolved Package into an Installed executor using the
// store. Installation is atomic: everything happens in a staging directory
// inside the store, and the executor only becomes visible after a final
// rename.
//
//	stage        → download → verify → extract → validate executor.json → write install.json
//	rename       → <root>/<owner>/<...segments>/<version>/   (only now installed)
type Installer struct {
	Store      Store
	Downloader Downloader
}

// InstallResult reports what happened.
type InstallResult struct {
	Installed *Installed
	// AlreadyPresent is true when the exact version was already installed and
	// the call was idempotent (no download occurred).
	AlreadyPresent bool
}

// Install materializes pkg into the store, returning the installed record.
// The exact version already being installed is not an error.
func (i *Installer) Install(ctx context.Context, pkg *Package) (*InstallResult, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package is nil")
	}

	if existing, err := i.Store.Get(ctx, pkg.Type, pkg.Version); err == nil && existing != nil {
		return &InstallResult{Installed: existing, AlreadyPresent: true}, nil
	}

	stage, err := i.Store.Stage()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)

	// 1. Materialize the artifact into the staging directory.
	if err := i.materialize(ctx, pkg, stage); err != nil {
		return nil, fmt.Errorf("install executor %s@%s: %w", pkg.Type, pkg.Version, err)
	}

	// 2. executor.json must be present and authoritative.
	manifestDir := stage
	manifestPath := filepath.Join(stage, shadexec.ManifestFile)

	if pkg.Manifest != nil {
		if err := os.WriteFile(manifestPath, pkg.Manifest, 0o644); err != nil {
			return nil, err
		}
	} else if _, err := os.Stat(manifestPath); err != nil {
		return nil, fmt.Errorf("%w: executor.json not found in package payload", ErrManifestInvalid)
	}

	m, err := ReadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	// 3. Resolve the concrete entrypoint (may differ from the package manifest
	// when the payload is a single binary rather than a pre-extracted dir).
	entrypoint, err := i.resolveEntrypoint(pkg, stage, m.Runtime.Entrypoint)
	if err != nil {
		return nil, err
	}

	// 4. Persist the install record.
	rec := RecordFor(&Installed{
		Type:         pkg.Type,
		Version:      pkg.Version,
		Digest:       pkg.Digest,
		Registry:     pkg.Registry,
		Platform:     HostPlatform(),
		RootDir:      stage,
		ManifestPath: manifestPath,
		ArtifactPath: stage,
		Runtime: RuntimeSpec{
			Type:       m.Runtime.Type,
			Entrypoint: entrypoint,
			Protocol:   m.Runtime.Protocol,
		},
		Capabilities: m.Capabilities,
		Services:     m.Services,
	})
	if err := WriteInstallRecord(filepath.Join(manifestDir, InstallFile), rec); err != nil {
		return nil, err
	}

	// 5. Atomic commit.
	installed, err := i.Store.Commit(ctx, stage, pkg.Type, pkg.Version)
	if err != nil {
		return nil, err
	}
	installed.Runtime.Entrypoint = entrypoint

	return &InstallResult{Installed: installed}, nil
}

// materialize downloads (or copies) the artifact and lays it into stage.
func (i *Installer) materialize(ctx context.Context, pkg *Package, stage string) error {
	if pkg.Artifact.URL == "" {
		// A package with no artifact (e.g. a source-only or local layout)
		// must carry the executor.json itself via pkg.Manifest.
		return nil
	}

	dst := filepath.Join(stage, ".artifact")

	if i.Downloader == nil {
		return fmt.Errorf("no downloader configured")
	}

	if err := i.Downloader.Download(ctx, pkg.Artifact.URL, dst); err != nil {
		return err
	}

	info, err := os.Stat(dst)
	if err != nil {
		return err
	}

	// Checksum verification: mandatory for single-file artifacts.
	if !info.IsDir() {
		if pkg.Artifact.SHA256 != "" {
			if err := VerifySHA256(dst, pkg.Artifact.SHA256); err != nil {
				return err
			}
		}
		if pkg.Digest == "" {
			digest, err := DigestFile(dst)
			if err != nil {
				return err
			}
			pkg.Digest = digest
		}

		if IsArchive(dst) {
			if err := ExtractTarGz(dst, stage); err != nil {
				return fmt.Errorf("extract artifact archive: %w", err)
			}
			return nil
		}

		// Single binary: place it next to the manifest.
		return placeBinary(dst, stage, fileNameOf(pkg))
	}

	// Directory payload: copy its contents into the staging root.
	return copyDirStage(dst, stage)
}

func (i *Installer) resolveEntrypoint(pkg *Package, stage, manifestEntrypoint string) (string, error) {
	if manifestEntrypoint != "" {
		p := filepath.Join(stage, filepath.FromSlash(manifestEntrypoint))
		if _, err := os.Stat(p); err == nil {
			return filepath.ToSlash(manifestEntrypoint), nil
		}
	}

	// Fall back to the artifact's own binary name (covers single-binary
	// packages whose manifest entrypoint did not match the file name).
	name := fileNameOf(pkg)
	if name != "" {
		if _, err := os.Stat(filepath.Join(stage, name)); err == nil {
			return name, nil
		}
	}

	return "", fmt.Errorf("%w: entrypoint %q not materialized", ErrManifestInvalid, manifestEntrypoint)
}

func placeBinary(src, stage, name string) error {
	if name == "" {
		return fmt.Errorf("cannot place binary artifact: no file name")
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stage, name), data, 0o755)
}

func copyDirStage(src, stage string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(stage, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		mode := 0o644
		if entry.Type()&0o111 != 0 {
			mode = 0o755
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstPath, data, os.FileMode(mode)); err != nil {
			return err
		}
	}
	return nil
}

func copyDirRecursive(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		mode := 0o644
		if entry.Type()&0o111 != 0 {
			mode = 0o755
		}
		if err := os.WriteFile(dstPath, data, os.FileMode(mode)); err != nil {
			return err
		}
	}
	return nil
}

func fileNameOf(pkg *Package) string {
	if name := strings.TrimSpace(pkg.Artifact.Name); name != "" {
		return name
	}
	return FileName(pkg.Artifact.URL)
}
