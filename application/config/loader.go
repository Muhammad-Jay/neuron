package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Options controls how Load assembles the effective configuration.
type Options struct {
	// ProjectDir is the project root used to locate ./neuron.yaml and to
	// resolve relative paths. It may be empty.
	ProjectDir string

	// GlobalPath overrides the global configuration file. When empty the
	// default ~/.config/neuron/config.yaml is used.
	GlobalPath string

	// ProjectPath overrides the project configuration file. When empty it is
	// derived from ProjectDir (neuron.yaml / neuron.yml).
	ProjectPath string

	// Environment enables NEURON_* environment-variable overrides.
	Environment bool

	// CLI holds the final command-level overrides keyed by config path, e.g.
	// {"daemon.endpoint": "...", "runtime.execution.mode": "detach"}.
	CLI map[string]any
}

// Load assembles the effective configuration in precedence order:
//
//	Defaults() → global config → project config → environment → CLI
func Load(opts Options) (Config, error) {
	cfg := Defaults()

	v := viper.New()
	registerDefaults(v, cfg)

	// Global user configuration.
	globalPath := opts.GlobalPath
	if globalPath == "" {
		var err error
		globalPath, err = GlobalConfigPath()
		if err != nil {
			return Config{}, err
		}
	}
	if fileExists(globalPath) {
		v.SetConfigFile(globalPath)
		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read global config: %w", err)
		}
	}

	// Project configuration.
	projectPath := opts.ProjectPath
	if projectPath == "" && opts.ProjectDir != "" {
		projectPath = findProjectConfig(opts.ProjectDir)
	}
	if projectPath != "" {
		if !fileExists(projectPath) {
			return Config{}, fmt.Errorf("project config not found: %s", projectPath)
		}
		v.SetConfigFile(projectPath)
		if err := v.MergeInConfig(); err != nil {
			return Config{}, fmt.Errorf("read project config: %w", err)
		}
	}

	// Environment variables.
	if opts.Environment {
		v.SetEnvPrefix("NEURON")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
		v.AutomaticEnv()
	}

	// Command-line overrides.
	for key, value := range opts.CLI {
		v.Set(key, value)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}

	resolvePaths(&cfg, opts.ProjectDir)

	return cfg, nil
}

// registerDefaults seeds every config key with its default so that partial
// overrides from lower layers fall through to the compiled-in value.
func registerDefaults(v *viper.Viper, cfg Config) {
	v.SetDefault("version", cfg.Version)

	v.SetDefault("runtime.execution.mode", cfg.Runtime.Execution.Mode)
	v.SetDefault("runtime.execution.timeout", cfg.Runtime.Execution.Timeout)
	v.SetDefault("runtime.workers.min", cfg.Runtime.Workers.Min)
	v.SetDefault("runtime.workers.max", cfg.Runtime.Workers.Max)

	v.SetDefault("daemon.endpoint", cfg.Daemon.Endpoint)
	v.SetDefault("daemon.socket", cfg.Daemon.Socket)
	v.SetDefault("daemon.norePath", cfg.Daemon.NorePath)
	v.SetDefault("daemon.pidFile", cfg.Daemon.PIDFile)

	v.SetDefault("storage.provider", cfg.Storage.Provider)
	v.SetDefault("storage.directory", cfg.Storage.Directory)

	v.SetDefault("inspector.enabled", cfg.Inspector.Enabled)
	v.SetDefault("inspector.address", cfg.Inspector.Address)

	if len(cfg.Executors.Registries) > 0 {
		items := make([]map[string]any, 0, len(cfg.Executors.Registries))
		for _, reg := range cfg.Executors.Registries {
			items = append(items, map[string]any{
				"name": reg.Name,
				"url":  reg.URL,
			})
		}
		v.SetDefault("executors.registries", items)
	}
}

// resolvePaths expands any relative or "~"-prefixed path fields against the
// project root so downstream consumers always receive absolute paths.
func resolvePaths(cfg *Config, projectDir string) {
	cfg.Storage.Directory = Expand(cfg.Storage.Directory, projectDir)
	cfg.Daemon.Socket = Expand(cfg.Daemon.Socket, projectDir)
	cfg.Daemon.PIDFile = Expand(cfg.Daemon.PIDFile, projectDir)

	if cfg.Daemon.NorePath != "" {
		cfg.Daemon.NorePath = Expand(cfg.Daemon.NorePath, projectDir)
	}
}

func findProjectConfig(projectDir string) string {
	for _, name := range []string{"neuron.yaml", "neuron.yml"} {
		p := filepath.Join(projectDir, name)
		if fileExists(p) {
			return p
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
