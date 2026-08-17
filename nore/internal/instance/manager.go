package instance

import (
	"context"
	"fmt"
	"sync"

	"github.com/Muhammad-Jay/neuron/nore/internal/storage"
	core2 "github.com/Muhammad-Jay/neuron/shared/types/core"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

type Manager struct {
	mu sync.RWMutex

	instancesByKey map[protocol.InstanceKey]*Instance
	instancesByID  map[string]*Instance

	parent  context.Context
	workers int
	store   storage.Store
}

func NewManager(parent context.Context, workers int, store storage.Store) *Manager {
	if parent == nil {
		parent = context.Background()
	}
	if workers <= 0 {
		workers = 8
	}
	return &Manager{
		instancesByKey: make(map[protocol.InstanceKey]*Instance),
		instancesByID:  make(map[string]*Instance),
		parent:         parent,
		workers:        workers,
		store:          store,
	}
}

func (m *Manager) Get(key protocol.InstanceKey) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.instancesByKey[key]
	return i, ok
}

func (m *Manager) GetByID(id string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	i, ok := m.instancesByID[id]
	return i, ok
}

func (m *Manager) GetOrCreate(key protocol.InstanceKey, system *core2.System) (*Instance, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if i, ok := m.instancesByKey[key]; ok {
		if i.Status() == StatusRunning || i.Status() == StatusStarting {
			return i, false, nil
		}
		delete(m.instancesByKey, key)
		delete(m.instancesByID, i.ID)
	}

	id := core2.NewID("inst_")
	i, err := New(m.parent, string(id), key, system, m.workers, m.store)
	if err != nil {
		return nil, false, err
	}
	if err := i.Start(); err != nil {
		return nil, false, err
	}

	m.instancesByKey[key] = i
	m.instancesByID[string(id)] = i
	return i, true, nil
}

func (m *Manager) List(opts protocol.ListOptions) []*Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Instance, 0, len(m.instancesByID))

	for _, i := range m.instancesByID {
		if opts.All {
			result = append(result, i)
		} else if opts.Status != "" {
			if string(i.Status()) == opts.Status {
				result = append(result, i)
			}
		} else {
			if i.Status() != StatusStopped && i.Status() != StatusFailed {
				result = append(result, i)
			}
		}
	}

	return result
}

func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	i, ok := m.instancesByID[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("instance %s not found", id)
	}
	delete(m.instancesByID, id)
	delete(m.instancesByKey, i.Key)
	m.mu.Unlock()

	return i.Stop()
}