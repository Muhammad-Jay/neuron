package routes

import (
	"net/http"

	"github.com/Muhammad-Jay/neuron/nore/internal/api/handlers/health"
	"github.com/Muhammad-Jay/neuron/nore/internal/api/handlers/instances"
	"github.com/Muhammad-Jay/neuron/nore/internal/api/middleware"
	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
)

func BuildRoutes(mux *http.ServeMux, mgr *instance.Manager) http.Handler {
	// Health
	mux.HandleFunc("GET /health", health.Health)

	// Instances
	instHandler := instances.New(mgr)
	mux.HandleFunc("GET /v1/instances", instHandler.ListInstances)
	mux.HandleFunc("POST /v1/instances", instHandler.CreateInstance)
	mux.HandleFunc("GET /v1/instances/{id}", instHandler.GetInstanceByID)

	// Executions
	mux.HandleFunc("POST /v1/instances/{id}/executions", instHandler.Execute)
	mux.HandleFunc("GET /v1/instances/{id}/executions", instHandler.ListExecutions)
	mux.HandleFunc("GET /v1/instances/{id}/executions/{execID}", instHandler.GetExecutionState)
	mux.HandleFunc("GET /v1/instances/{id}/executions/{execID}/events", instHandler.GetExecutionEvents)
	mux.HandleFunc("GET /v1/instances/{id}/executions/{execID}/events/stream", instHandler.StreamExecutionEvents)

	handler := middleware.Recovery(mux)
	handler = middleware.Logging(handler)

	return handler
}
