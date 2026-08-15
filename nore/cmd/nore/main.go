package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/server"
)

func main() {
	var (
		port    string
		socket  string
		workers int
	)

	flag.StringVar(&port, "port", ":7432", "TCP address for the N.O.R.E. API; empty disables TCP")
	flag.StringVar(&socket, "socket", defaultSocket(), "Unix socket for local CLI clients; empty disables Unix socket")
	flag.IntVar(&workers, "workers", 8, "executor worker count")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.NewServer(ctx, workers)

	type listenerEntry struct {
		name string
		l    net.Listener
	}

	var listeners []listenerEntry

	if port != "" {
		l, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("listen on %s: %v", port, err)
		}
		listeners = append(listeners, listenerEntry{"tcp", l})
		fmt.Printf("N.O.R.E. listening on %s\n", port)
	}

	if socket != "" {
		if err := os.MkdirAll(filepath.Dir(socket), 0o755); err != nil {
			log.Fatalf("create socket directory: %v", err)
		}
		_ = os.Remove(socket)

		l, err := net.Listen("unix", socket)
		if err != nil {
			log.Fatalf("listen on unix socket %s: %v", socket, err)
		}
		_ = os.Chmod(socket, 0o600)
		listeners = append(listeners, listenerEntry{"unix", l})
		fmt.Printf("N.O.R.E. local socket: %s\n", socket)
	}

	if len(listeners) == 0 {
		log.Fatal("at least one of --port or --socket must be configured")
	}

	errCh := make(chan error, len(listeners))
	for _, entry := range listeners {
		go func() {
			if err := srv.Serve(ctx, entry.l); err != nil && ctx.Err() == nil {
				errCh <- fmt.Errorf("%s server: %w", entry.name, err)
			}
		}()
	}

	select {
	case <-ctx.Done():
		for _, entry := range listeners {
			_ = entry.l.Close()
		}
	case err := <-errCh:
		for _, entry := range listeners {
			_ = entry.l.Close()
		}
		log.Fatal(err)
	}
}

func defaultSocket() string {
	if value := os.Getenv("NEURON_SOCKET"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/neuron/nore.sock"
	}
	return filepath.Join(home, ".neuron", "nore.sock")
}
