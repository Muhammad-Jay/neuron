package server

import (
	"context"
	"encoding/json"
	maps0 "maps"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
	"github.com/Muhammad-Jay/neuron/nore/internal/storage"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
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

func NewServer(ctx context.Context, workers int, store storage.Store) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		instances: instance.NewManager(ctx, workers, store),
	}

	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /v1/instances", s.handleInstances)
	s.mux.HandleFunc("POST /v1/instances", s.handleCreateInstance)
	s.mux.HandleFunc("GET /v1/instances/{id}", s.handleInstance)
	s.mux.HandleFunc("GET /v1/instances/{id}/executions", s.handleInstanceExecutions)
	s.mux.HandleFunc("GET /v1/instances/{id}/executions/{execID}", s.handleInstanceExecution)
	s.mux.HandleFunc("GET /v1/instances/{id}/executions/{execID}/events", s.handleExecutionEvents)
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
	for _, i := range s.instances.List(protocol.ListOptions{}) {
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

func parseBool(q url.Values, key string, defaultVal bool) bool {
	val := q.Get(key)
	if strings.EqualFold(val, "true") {
		return true
	}
	if strings.EqualFold(val, "false") {
		return false
	}
	return defaultVal
}

func GetQueryParams(r *http.Request) map[string]any {
	params := make(map[string]any)
	query := r.URL.Query()

	for key, values := range query {
		if len(values) == 1 {
			params[key] = parseBoolVal(values[0])
		} else if len(values) > 1 {
			parsed := make([]any, len(values))
			for i, v := range values {
				parsed[i] = parseBoolVal(v)
			}
			params[key] = parsed
		}
	}

	return params
}

func parseBoolVal(v string) any {
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return v
}

func MergeMaps(maps ...map[string]any) map[string]any {
	merged := make(map[string]any)
	for _, m := range maps {
		maps0.Copy(merged, m)
	}
	return merged
}