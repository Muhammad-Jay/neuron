package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    any    `json:"data,omitempty"`
}

type Server struct {
	mux       *http.ServeMux
	instances *instance.Manager
}

func NewServer(ctx context.Context, workers int) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		instances: instance.NewManager(ctx, workers),
	}

	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /v1/instances", s.handleInstances)
	s.mux.HandleFunc("POST /v1/instances", s.handleCreateInstance)
	s.mux.HandleFunc("GET /v1/instances/{id}", s.handleInstance)
	s.mux.HandleFunc("POST /v1/instances/{id}/executions", s.handleExecute)

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	httpServer := &http.Server{
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	return httpServer.Serve(listener)
}

func (s *Server) StopInstances() {
	for _, i := range s.instances.List() {
		_ = i.Stop()
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func pathID(value string) string {
	return strings.TrimSpace(value)
}

func errorJSON(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, Response{Message: err.Error(), Status: status})
}
