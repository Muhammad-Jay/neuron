package daemon

import (
	"errors"
	"time"
)

var (
	ErrAlreadyRunning = errors.New("nore daemon is already running")
	ErrNotRunning     = errors.New("nore daemon is not running")
	ErrStartTimeout   = errors.New("timed out waiting for nore daemon to become healthy")
	ErrInvalidConfig  = errors.New("invalid daemon configuration")
)

type Config struct {
	BinaryPath     string
	Args           []string
	WorkingDir     string
	Env            []string
	StartTimeout   time.Duration
	HealthInterval time.Duration
	AttachOutput   bool
}

func DefaultConfig() Config {
	return Config{
		BinaryPath:     "nore",
		Args:           []string{"serve"},
		StartTimeout:   10 * time.Second,
		HealthInterval: 100 * time.Millisecond,
		AttachOutput:   true,
	}
}