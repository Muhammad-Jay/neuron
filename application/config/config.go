// Package config produces the effective Neuron runtime configuration.
//
// Configuration is assembled from several layers, each overriding the one
// below it:
//
//	1. built-in defaults
//	2. global user configuration  (~/.config/neuron/config.yaml)
//	3. project configuration       (./neuron.yaml)
//	4. environment variables       (NEURON_*)
//	5. command-line overrides      (Options.CLI)
//
// The rest of Neuron never touches Viper or the YAML representation. It only
// sees this typed Config, which is why the loader is the only file in this
// package that depends on Viper.
package config

// Config is the effective Neuron configuration.
type Config struct {
	Version   int             `yaml:"version,omitempty"   mapstructure:"version"`
	Runtime   RuntimeConfig   `yaml:"runtime"             mapstructure:"runtime"`
	Daemon    DaemonConfig    `yaml:"daemon"              mapstructure:"daemon"`
	Storage   StorageConfig   `yaml:"storage"             mapstructure:"storage"`
	Executors ExecutorsConfig `yaml:"executors"           mapstructure:"executors"`
	Inspector InspectorConfig `yaml:"inspector"           mapstructure:"inspector"`
}

// RuntimeConfig holds N.O.R.E.-related runtime defaults.
type RuntimeConfig struct {
	Execution ExecutionConfig `yaml:"execution" mapstructure:"execution"`
	Workers   WorkerConfig    `yaml:"workers"   mapstructure:"workers"`
}

// ExecutionConfig controls how executions behave by default.
type ExecutionConfig struct {
	// Mode is "wait" or "detach".
	Mode string `yaml:"mode" mapstructure:"mode"`

	// Timeout is a duration string such as "30m". Kept as a string so it
	// unmarshals from YAML without custom parsing.
	Timeout string `yaml:"timeout" mapstructure:"timeout"`
}

// WorkerConfig describes the executor worker pool.
type WorkerConfig struct {
	Min int `yaml:"min,omitempty" mapstructure:"min"`
	Max int `yaml:"max,omitempty" mapstructure:"max"`
}

// DaemonConfig controls the local N.O.R.E. daemon.
type DaemonConfig struct {
	// Endpoint, when non-empty, selects a remote N.O.R.E. over its
	// daemon/socket. An empty value means the local Unix socket daemon.
	Endpoint string `yaml:"endpoint,omitempty" mapstructure:"endpoint"`

	// Socket is the local Unix socket the daemon listens on.
	Socket string `yaml:"socket,omitempty" mapstructure:"socket"`

	// NorePath overrides the daemon binary path.
	NorePath string `yaml:"norePath,omitempty" mapstructure:"norePath"`

	// PIDFile is where the daemon records its process id.
	PIDFile string `yaml:"pidFile,omitempty" mapstructure:"pidFile"`
}

// StorageConfig selects the storage provider and its root directory.
type StorageConfig struct {
	Provider string `yaml:"provider" mapstructure:"provider"`

	// Directory is the provider's data directory. The storage implementation
	// decides the individual subdirectories (systems, instances, ...).
	Directory string `yaml:"directory" mapstructure:"directory"`
}

// ExecutorsConfig lists the executable registries Neuron can resolve services
// against. A service only declares `type`/`version`; the registry resolution
// system determines where the executor comes from.
type ExecutorsConfig struct {
	Registries []ExecutorRegistry `yaml:"registries,omitempty" mapstructure:"registries"`
}

// ExecutorRegistry identifies a source of executor definitions.
type ExecutorRegistry struct {
	Name string `yaml:"name,omitempty" mapstructure:"name"`
	URL  string `yaml:"url"            mapstructure:"url"`
}

// InspectorConfig controls the inspector endpoint.
type InspectorConfig struct {
	Enabled bool   `yaml:"enabled,omitempty" mapstructure:"enabled"`
	Address string `yaml:"address,omitempty" mapstructure:"address"`
}
