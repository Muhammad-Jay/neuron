package config

import "context"

type contextKey struct{}

// NewContext returns a context carrying the effective configuration.
func NewContext(ctx context.Context, c Config) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

// FromContext returns the effective configuration previously stored by
// NewContext, if present.
func FromContext(ctx context.Context) (Config, bool) {
	c, ok := ctx.Value(contextKey{}).(Config)
	return c, ok
}
