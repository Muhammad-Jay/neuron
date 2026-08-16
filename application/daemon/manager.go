package daemon

import (
	"context"
	"fmt"
	"time"
)

type HealthChecker interface {
	Healthy(ctx context.Context) error
}

type Manager struct {
	config  Config
	process Process
	health  HealthChecker
}

func NewManager(config Config, health HealthChecker) *Manager {
	return &Manager{
		config: config,
		health: health,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	// 1. Check if N.O.R.E. is already healthy
	if m.health != nil && m.health.Healthy(ctx) == nil {
		return nil
	}

	// 2. Start the local process
	if err := m.process.Start(m.config); err != nil {
		return fmt.Errorf("start nore daemon: %w", err)
	}

	// 3. Wait until N.O.R.E. is healthy or timeout occurs
	if err := m.waitForHealth(ctx); err != nil {
		// If it fails to become healthy, shut down the misbehaving process
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = m.process.Stop(shutdownCtx)
		return err
	}

	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	return m.process.StopCrossProcess(ctx, m.config)
}

func (m *Manager) waitForHealth(ctx context.Context) error {
	timeout := m.config.StartTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	interval := m.config.HealthInterval
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := m.health.Healthy(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", ErrStartTimeout, ctx.Err())
		case <-ticker.C:
		}
	}
}