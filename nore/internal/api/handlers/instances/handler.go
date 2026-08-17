package instances

import (
	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
)

type Handler struct {
	instances *instance.Manager
}

func New(m *instance.Manager) *Handler {
	return &Handler{instances: m}
}
