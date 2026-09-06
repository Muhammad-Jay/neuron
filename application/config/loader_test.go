package config

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(Options{GlobalPath: "/nonexistent/global.yaml", ProjectDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Runtime.Execution.Mode != "wait" {
		t.Errorf("default mode = %q, want wait", cfg.Runtime.Execution.Mode)
	}
	if cfg.Runtime.Execution.Timeout != "30m" {
		t.Errorf("default timeout = %q, want 30m", cfg.Runtime.Execution.Timeout)
	}
	if cfg.Storage.Provider != "file" {
		t.Errorf("default provider = %q, want file", cfg.Storage.Provider)
	}
	if len(cfg.Executors.Registries) != 2 {
		t.Errorf("default registries = %d, want 2", len(cfg.Executors.Registries))
	}
}

func TestProjectPartialOverridePreservesGlobal(t *testing.T) {
	dir := t.TempDir()

	global := write(t, dir, "global.yaml", `
storage:
  provider: file
  directory: /data/global
`)
	write(t, dir, "neuron.yaml", `
storage:
  directory: /data/project
`)

	cfg, err := Load(Options{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "neuron.yaml"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// The project only overrode directory; provider must survive from global.
	if cfg.Storage.Provider != "file" {
		t.Errorf("provider = %q, want file (preserved from global)", cfg.Storage.Provider)
	}
	if cfg.Storage.Directory != "/data/project" {
		t.Errorf("directory = %q, want /data/project", cfg.Storage.Directory)
	}
}

func TestCLIOverridesProject(t *testing.T) {
	dir := t.TempDir()

	write(t, dir, "neuron.yaml", `
runtime:
  execution:
    mode: wait
    timeout: 10m
`)

	cfg, err := Load(Options{
		GlobalPath:  "/nonexistent/global.yaml",
		ProjectPath: filepath.Join(dir, "neuron.yaml"),
		CLI: map[string]any{
			"runtime.execution.mode": "detach",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Runtime.Execution.Mode != "detach" {
		t.Errorf("mode = %q, want detach (CLI wins)", cfg.Runtime.Execution.Mode)
	}
	if cfg.Runtime.Execution.Timeout != "10m" {
		t.Errorf("timeout = %q, want 10m (project preserved)", cfg.Runtime.Execution.Timeout)
	}
}

func TestDefaultTimeoutSurvivesProjectModeOverride(t *testing.T) {
	dir := t.TempDir()

	// Project only sets mode; timeout should fall back to the 30m default.
	write(t, dir, "neuron.yaml", `
runtime:
  execution:
    mode: detach
`)

	cfg, err := Load(Options{
		GlobalPath:  "/nonexistent/global.yaml",
		ProjectPath: filepath.Join(dir, "neuron.yaml"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Runtime.Execution.Mode != "detach" {
		t.Errorf("mode = %q, want detach", cfg.Runtime.Execution.Mode)
	}
	if cfg.Runtime.Execution.Timeout != "30m" {
		t.Errorf("timeout = %q, want 30m default", cfg.Runtime.Execution.Timeout)
	}
}
