package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Muhammad-Jay/neuron/application/executor"
)

// manifestAssetName is the preferred executor.json location inside a release.
const manifestAssetName = "executor.json"

// buildPackage assembles an executor.Package for one release: it reads the
// executor.json manifest (release asset, falling back to the raw repository
// file at the release tag), then binds the host-platform binary asset.
func (r *Registry) buildPackage(
	ctx context.Context,
	typ, version string,
	ref RepoRef,
	release *Release,
) (*executor.Package, error) {

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

	// Bind the host-platform artifact when the manifest describes one.
	if platform, ok := m.Platforms[executor.HostPlatform()]; ok && platform.Artifact != "" {
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
