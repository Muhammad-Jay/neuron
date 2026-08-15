package client

import (
	"os"
	"path/filepath"
)

func DefaultSocketPath() string {
	if value := os.Getenv("NEURON_SOCKET"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/neuron/nore.sock"
	}
	return filepath.Join(home, ".neuron", "nore.sock")
}

func NewDefaultLocal() *Client {
	return NewLocal(DefaultSocketPath())
}