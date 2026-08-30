package config

import (
	"os"
	"path/filepath"
	"strings"
)

// GlobalDir returns the machine/user-level Neuron configuration directory.
func GlobalDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "neuron"), nil
}

// GlobalConfigPath returns the path of the global configuration file.
func GlobalConfigPath() (string, error) {
	dir, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// ProjectStateDir returns the project-local Neuron state directory.
func ProjectStateDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".neuron")
}

// DefaultLocalSocket returns the default Unix socket for local clients.
func DefaultLocalSocket() string {
	if value := os.Getenv("NEURON_SOCKET"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/neuron/nore.sock"
	}
	return filepath.Join(home, ".neuron", "nore.sock")
}

// DefaultDataDir returns the default persistent data directory.
func DefaultDataDir() string {
	if value := os.Getenv("NEURON_DATA_DIR"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/neuron/data"
	}
	return filepath.Join(home, ".local", "share", "neuron")
}

// DefaultPIDFile returns the default daemon PID file location.
func DefaultPIDFile() string {
	return filepath.Join(os.TempDir(), "nore.pid")
}

// Expand resolves a possibly-user-relative path into an absolute path.
//
// A leading "~" is expanded to the user's home directory; an otherwise
// relative path is resolved against baseDir. It returns the input unchanged
// when it is already absolute.
func Expand(path, baseDir string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}

	if !filepath.IsAbs(path) {
		base := baseDir
		if base == "" {
			base = "."
		}
		path = filepath.Join(base, path)
	}

	return filepath.Clean(path)
}
