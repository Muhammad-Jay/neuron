package connection

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func NewLocal(socketPath string) Connection {
	socketPath = filepath.Clean(socketPath)

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
	}

	// The host is intentionally fake. The Unix socket determines the destination.
	return New(NewHTTPTransport(client, "http://nore.local"))
}

func LocalSocketExists(socketPath string) bool {
	_, err := os.Stat(socketPath)
	return err == nil
}
