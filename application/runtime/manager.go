package runtime

import (
	"context"

	"github.com/Muhammad-Jay/neuron/application/connection"
	"github.com/Muhammad-Jay/neuron/application/daemon"
)

// connectionHealthAdapter adapts your Connection interface to daemon.HealthChecker
type connectionHealthAdapter struct {
	conn connection.Connection
}

func (c connectionHealthAdapter) Healthy(ctx context.Context) error {
	return c.conn.Health(ctx)
}

type Manager struct {
	Daemon *daemon.Manager
	Conn   connection.Connection
}

func NewManager(cfg daemon.Config, conn connection.Connection) *Manager {
	return &Manager{
		Conn: conn,
		Daemon: daemon.NewManager(
			cfg,
			connectionHealthAdapter{conn: conn},
		),
	}
}

// Ensure guarantees that the N.O.R.E connection is ready.
// If it's a local connection and N.O.R.E is down, it starts it.
func (m *Manager) Ensure(ctx context.Context, isRemote bool) error {
	// If it's already healthy (local or remote), we are good to go.
	if err := m.Conn.Health(ctx); err == nil {
		return nil
	}

	// If the user specified a remote URL, we should not try to start a local daemon.
	if isRemote {
		return m.Conn.Health(ctx) // Return the actual connection error
	}

	// It's local and down. Let the daemon manager start it and wait for health.
	return m.Daemon.Start(ctx)
}