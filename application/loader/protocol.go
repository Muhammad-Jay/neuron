package loader

import "context"

type Contract interface {
	Build() error

	// BuildContext builds honoring ctx. It supersedes Build for callers that
	// can thread a context through.
	BuildContext(ctx context.Context) error
}