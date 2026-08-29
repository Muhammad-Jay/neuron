package register

import (
	"github.com/Muhammad-Jay/neuron/nore/internal/planner"
	"github.com/Muhammad-Jay/neuron/nore/internal/system"
)

type Handler struct {
	systems  *system.Repository
	compiler *planner.Compiler
}

func New(systems *system.Repository, compiler *planner.Compiler) *Handler {
	return &Handler{systems: systems, compiler: compiler}
}