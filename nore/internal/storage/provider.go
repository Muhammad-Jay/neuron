package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("key not found")
)

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, data []byte) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	Close() error
}

type Config struct {
	DataDir string
}
