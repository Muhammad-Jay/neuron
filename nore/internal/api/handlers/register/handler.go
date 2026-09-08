package register

import (
	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
	"github.com/Muhammad-Jay/neuron/nore/internal/planner"
	"github.com/Muhammad-Jay/neuron/nore/internal/system"
)

type Handler struct {
	instances *instance.Manager
	systems   *system.Repository
	compiler  *planner.Compiler
}

func New(instances *instance.Manager, systems *system.Repository, compiler *planner.Compiler) *Handler {
	return &Handler{instances: instances, systems: systems, compiler: compiler}
}
