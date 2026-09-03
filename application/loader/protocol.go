package loader

import "context"

type Contract interface {
	load(ctx context.Context, path string) (string, error)
}
