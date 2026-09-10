package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/executor"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// manifestAssetName is the preferred executor.json location inside a release.
const manifestAssetName = "executor.json"

// packageArchiveAsset finds the executor package archive asset in a release.
// The exact <type>-<version>-executor.neuron.tar.gz name wins; otherwise any
// asset carrying the canonical package archive suffix is accepted, trusting
// the manifest inside the archive to reconcile identity at install time.
func packageArchiveAsset(release *Release, typ, version string) (Asset, bool) {
	if asset, ok := release.assetByName(shadexec.PackageArchiveName(typ, version)); ok {
		return asset, true
	}
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, shadexec.PackageArchiveSuffix) {
			return a, true
		}
	}
	return Asset{}, false
}

// buildPackage assembles an executor.Package for one release. The executor
// package archive (<type>-<version>-executor.neuron.tar.gz) is the preferred
// distribution: one immutable asset containing executor.json plus every
// platform artifact. Its inner manifest is authoritative, so no separate
// manifest fetch happens here and identity/version are reconciled at install
// time. When no package archive exists, the manifest is read from the release
// (asset, then raw repository file) and the platform artifact for the
// executor's runtime kind is bound.
func (r *Registry) buildPackage(
	ctx context.Context,
	typ, version string,
	ref RepoRef,
	release *Release,
) (*executor.Package, error) {

	// Package archive distribution.
	if asset, ok := packageArchiveAsset(release, typ, version); ok {
		return &executor.Package{
			Type:     typ,
			Version:  version,
			Registry: r.Name(),
			Artifact: executor.Artifact{
				URL:  asset.BrowserDownloadURL,
				Name: asset.Name,
			},
		}, nil
	}

	// Legacy per-platform distribution.
	manifestBytes, source, err := r.fetchManifest(ctx, ref, release)
	if err != nil {
		return nil, err
	}
	if len(manifestBytes) == 0 {
		return nil, fmt.Errorf(
			"executor %s@%s: executor.json not found in release %s or repository",
			typ, version, ref.Owner+"/"+ref.Repo,
		)
	}

	m, err := executor.ParseManifest(manifestBytes)
	if err != nil {
		return nil, err
	}

	// The release must actually serve this executor's version.
	if v := strings.TrimPrefix(m.Metadata.Version, "v"); v != "" && v != version {
		return nil, fmt.Errorf(
			"executor %s: manifest version %s does not match release version %s (release %s)",
			typ, m.Metadata.Version, version, source,
		)
	}

	pkg := executor.PackageFromManifest(m, r.Name(), version)
	pkg.Manifest = manifestBytes

	// Bind the platform artifact for the executor's runtime kind. WASM modules
	// select the "wasm32-wasi" key; process executors select the host key.
	platformKey := executor.PlatformForRuntime(m.Runtime.Type)
	if platform, ok := m.Platforms[platformKey]; ok && platform.Artifact != "" {
		if asset, found := release.assetByName(platform.Artifact); found {
			pkg.Artifact = executor.Artifact{
				URL:    asset.BrowserDownloadURL,
				Name:   asset.Name,
				SHA256: platform.SHA256,
			}
		} else {
			// Fall back to raw URL construction; the manifest is authoritative
			// for the artifact name even when the asset listing lags.
			pkg.Artifact = executor.Artifact{
				URL:    rawDownloadURL(ref.Owner, ref.Repo, release.TagName, platform.Artifact),
				Name:   platform.Artifact,
				SHA256: platform.SHA256,
			}
		}
	}

	return pkg, nil
}

// fetchManifest returns the validated executor.json bytes. Precedence: release
// asset named executor.json, then the raw repository file at the release tag.
func (r *Registry) fetchManifest(ctx context.Context, ref RepoRef, release *Release) ([]byte, string, error) {
	if asset, ok := release.assetByName(manifestAssetName); ok {
		data, err := r.client.Get(ctx, asset.BrowserDownloadURL)
		if err != nil {
			return nil, "", err
		}
		return data, "asset " + asset.Name, nil
	}

	raw := rawFileURL(ref.Owner, ref.Repo, release.TagName, manifestAssetName)
	data, err := r.client.Get(ctx, raw)
	if err != nil {
		if errors.Is(err, notFoundSentinel) {
			return nil, "", nil
		}
		return nil, "", err
	}
	return data, "raw " + raw, nil
}

func rawFileURL(owner, repo, tag, file string) string {
	return "https://raw.githubusercontent.com/" + escapePath(owner) + "/" + escapePath(repo) + "/" + escapePath(tag) + "/" + escapePath(file)
}

func rawDownloadURL(owner, repo, tag, asset string) string {
	return "https://github.com/" + escapePath(owner) + "/" + escapePath(repo) + "/releases/download/" + escapePath(tag) + "/" + escapePath(asset)
}
