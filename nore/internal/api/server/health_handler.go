package server

import (
	"net/http"

	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, protocol.Response{
		Message: "N.O.R.E. is healthy",
		Status:  http.StatusOK,
		Data:    map[string]string{"service": "nore"},
	})
}