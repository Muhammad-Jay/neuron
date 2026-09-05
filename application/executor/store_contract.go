package executor

import (
	"context"
)

// Store is the installed-executor catalog contract consumed by the Resolver
// and Installer. Concrete implementations (executor/store) live below this
// package so the import direction stays executor → store.
type Store interface {
	// Root returns the absolute store root directory.
	Root() string

	// Stage creates a fresh, empty staging directory inside the store.
	Stage() (string, error)

	// Commit atomically renames a fully-populated staging directory into the
	// final store location and returns the installed record it contains.
	Commit(ctx context.Context, stage, typ, version string) (*Installed, error)

	// Get returns the installed executor for an exact version. Not-found is
	// signaled with ErrNotFound.
	Get(ctx context.Context, typ, version string) (*Installed, error)

	// List returns every installed version of typ, newest first. An empty typ
	// lists all installed executors.
	List(ctx context.Context, typ string) ([]Installed, error)

	// Remove deletes an installed version.
	Remove(ctx context.Context, typ, version string) error
}
