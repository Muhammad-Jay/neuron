package instances

import (
	"net/http"
	"strings"

	"github.com/Muhammad-Jay/neuron/nore/internal/instance"
	"github.com/Muhammad-Jay/neuron/nore/internal/planner"
	"github.com/Muhammad-Jay/neuron/nore/internal/system"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

type Handler struct {
	instances *instance.Manager
	systems   *system.Repository
	compiler  *planner.Compiler
}

func New(m *instance.Manager, systems *system.Repository, compiler *planner.Compiler) *Handler {
	return &Handler{instances: m, systems: systems, compiler: compiler}
}

// resolveInstance maps a URL segment to a live instance. The segment may be an
// instance ID (inst_*) or a colon-encoded system key
// (systemID:version:hash[:env]); keys resolve to an existing runtime without
// creating one.
func (h *Handler) resolveInstance(r *http.Request, id string) (*instance.Instance, bool) {
	if strings.HasPrefix(id, "inst_") {
		return h.instances.GetByID(id)
	}
	key, err := protocol.ParseKey(id)
	if err != nil {
		return nil, false
	}
	return h.instances.Resolve(r.Context(), key)
}
