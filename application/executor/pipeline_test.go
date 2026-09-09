package executor_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Muhammad-Jay/neuron/application/executor"
	"github.com/Muhammad-Jay/neuron/application/executor/source/local"
	execstore "github.com/Muhammad-Jay/neuron/application/executor/store"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

// writeExecutorPackage lays out a local registry package:
//
//	<root>/github/read/1.2.0/executor.json
//	<root>/github/read/1.2.0/executor.sh
func writeExecutorPackage(t *testing.T, root, typ, version string) string {
	t.Helper()

	split, err := executor.ParseType(typ)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, filepath.FromSlash(split.ToLocalPath()), version)
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}

	entrypoint := "executor.sh"
	executorPath := filepath.Join(base, entrypoint)
	if err := os.WriteFile(executorPath, []byte("#!/bin/sh\necho '{\"output\":{}}'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	m := &shadexec.Manifest{
		APIVersion: shadexec.APIVersion,
		Kind:       shadexec.Kind,
		Metadata: shadexec.ManifestMetadata{
			Name:        typ,
			Version:     version,
			Description: "test executor",
		},
		Runtime: shadexec.ManifestRuntime{
			Type:       "process",
			Entrypoint: entrypoint,
			Protocol:   shadexec.ProtocolJSONV1,
		},
		Services:     []string{"read"},
		Capabilities: []string{"io.read"},
		Platforms: map[string]shadexec.Platform{
			executor.HostPlatform(): {Artifact: entrypoint},
		},
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, shadexec.ManifestFile), data, 0o644); err != nil {
		t.Fatal(err)
	}

	return base
}

func TestLocalRegistryPipelineTest(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	writeExecutorPackage(t, root, "github:read", "1.2.0")
	writeExecutorPackage(t, root, "github:read", "2.0.0")
	writeExecutorPackage(t, root, "Muhammad-Jay:github:read", "1.0.0")

	reg, err := local.New(root)
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}

	types, err := reg.Types(ctx)
	if err != nil {
		t.Fatalf("Types: %v", err)
	}
	if len(types) != 2 {
		t.Fatalf("Types = %v, want 2", types)
	}

	versions, err := reg.Versions(ctx, "github:read")
	if err != nil {
		t.Fatalf("Versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("Versions = %v, want 2", versions)
	}

	pkg, err := reg.Package(ctx, "github:read", "2.0.0")
	if err != nil {
		t.Fatalf("Package: %v", err)
	}
	if pkg.Version != "2.0.0" {
		t.Errorf("Package.Version = %q", pkg.Version)
	}
	if pkg.Artifact.URL == "" {
		t.Error("Package.Artifact should be bound to the local file")
	}
}

func TestResolveInstallPipeline(t *testing.T) {
	ctx := context.Background()

	regRoot := t.TempDir()
	storeRoot := filepath.Join(t.TempDir(), "store")
	writeExecutorPackage(t, regRoot, "github:read", "1.2.0")
	writeExecutorPackage(t, regRoot, "github:read", "2.0.0")

	reg, err := local.New(regRoot)
	if err != nil {
		t.Fatal(err)
	}

	fsStore, err := execstore.NewFilesystemStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}

	catalog := executor.NewRegistry()
	if err := catalog.Add(reg); err != nil {
		t.Fatal(err)
	}

	installer := &executor.Installer{Store: fsStore, Downloader: executor.NewHTTPDownloader()}
	resolver := executor.NewResolver(catalog, fsStore, installer)

	// Constrained: ^1 resolves to 1.2.0 only.
	req := executor.Requirement{Type: "github:read", Version: "^1.0.0", Registries: []string{"local"}}
	installed, err := resolver.Resolve(ctx, req)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if installed.Version != "1.2.0" {
		t.Errorf("resolved version = %q, want 1.2.0", installed.Version)
	}
	if installed.Registry != "local" {
		t.Errorf("registry = %q, want local", installed.Registry)
	}

	// Installed-first: resolving again returns the same artifact.
	again, err := resolver.Resolve(ctx, req)
	if err != nil {
		t.Fatalf("Resolve again: %v", err)
	}
	if again.Version != installed.Version || again.RootDir != installed.RootDir {
		t.Errorf("installed-first violated: %+v vs %+v", installed, again)
	}

	// Latest floating resolves to 2.0.0 and is installed.
	if _, err := resolver.Resolve(ctx, executor.Requirement{Type: "github:read", Registries: []string{"local"}}); err != nil {
		t.Fatalf("Resolve latest: %v", err)
	}

	list, err := fsStore.List(ctx, "github:read")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("store list = %d, want 2", len(list))
	}

	// Get + remove.
	got, err := fsStore.Get(ctx, "github:read", "1.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if got.RootDir == "" {
		t.Error("installed RootDir should be set")
	}
	if err := fsStore.Remove(ctx, "github:read", "1.2.0"); err != nil {
		t.Fatal(err)
	}
	if _, err := fsStore.Get(ctx, "github:read", "1.2.0"); err == nil {
		t.Error("expected ErrNotFound after remove")
	} else if !errors.Is(err, executor.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestResolveUnconfiguredRegistry(t *testing.T) {
	ctx := context.Background()

	fsStore, err := execstore.NewFilesystemStore(filepath.Join(t.TempDir(), "store"))
	if err != nil {
		t.Fatal(err)
	}

	catalog := executor.NewRegistry()
	installer := &executor.Installer{Store: fsStore, Downloader: executor.NewHTTPDownloader()}
	resolver := executor.NewResolver(catalog, fsStore, installer)

	req := executor.Requirement{Type: "github:read", Registries: []string{"nope"}}
	_, err = resolver.Resolve(ctx, req)
	if err == nil {
		t.Fatal("expected error for unconfigured registry")
	}
	if !errors.Is(err, executor.ErrRegistryNotConfigured) {
		t.Errorf("want ErrRegistryNotConfigured, got %v", err)
	}
}

func TestFrozenJSONRoundTrip(t *testing.T) {
	ctx := context.Background()

	regRoot := t.TempDir()
	storeRoot := filepath.Join(t.TempDir(), "store")
	writeExecutorPackage(t, regRoot, "Muhammad-Jay:github:read", "1.2.0")

	reg, err := local.New(regRoot)
	if err != nil {
		t.Fatal(err)
	}
	fsStore, err := execstore.NewFilesystemStore(storeRoot)
	if err != nil {
		t.Fatal(err)
	}
	catalog := executor.NewRegistry()
	if err := catalog.Add(reg); err != nil {
		t.Fatal(err)
	}
	installer := &executor.Installer{Store: fsStore, Downloader: executor.NewHTTPDownloader()}
	resolver := executor.NewResolver(catalog, fsStore, installer)

	installed, err := resolver.Resolve(ctx, executor.Requirement{Type: "Muhammad-Jay:github:read", Version: "^1.0.0", Registries: []string{"local"}})
	if err != nil {
		t.Fatal(err)
	}

	frozen := installed.Frozen("^1.0.0")
	if frozen.Type != "Muhammad-Jay:github:read" {
		t.Errorf("Frozen.Type = %q", frozen.Type)
	}
	if frozen.ResolvedVersion != "1.2.0" {
		t.Errorf("Frozen.ResolvedVersion = %q", frozen.ResolvedVersion)
	}
	if frozen.RequestedVersion != "^1.0.0" {
		t.Errorf("Frozen.RequestedVersion = %q", frozen.RequestedVersion)
	}
	if frozen.Registry != "local" {
		t.Errorf("Frozen.Registry = %q", frozen.Registry)
	}

	// Wire round-trip through JSON keeps all fields.
	data, err := json.Marshal(frozen)
	if err != nil {
		t.Fatal(err)
	}
	var back shadexec.ResolvedExecutor
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.ResolvedVersion != frozen.ResolvedVersion || back.Type != frozen.Type {
		t.Errorf("JSON round-trip mismatch: %+v", back)
	}
}
