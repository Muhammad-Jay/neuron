package instance

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Muhammad-Jay/neuron/nore/internal/event"
	"github.com/Muhammad-Jay/neuron/nore/internal/execution"
	"github.com/Muhammad-Jay/neuron/nore/internal/storage"
	"github.com/Muhammad-Jay/neuron/nore/internal/system"
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
	metadata *metadataStore
	systems *system.Repository
}

func NewManager(parent context.Context, workers int, store storage.Store, systems *system.Repository) *Manager {
	if parent == nil {
		parent = context.Background()
	}
	if workers <= 0 {
		workers = 8
	}
	m := &Manager{
		instancesByKey: make(map[protocol.InstanceKey]*Instance),
		instancesByID:  make(map[string]*Instance),
		parent:         parent,
		workers:        workers,
		store:          store,
		metadata:       newMetadataStore(store),
		systems:        systems,
	}
	m.reconcile()
	return m
}

// reconcile registers instances whose metadata was persisted by a previous
// process so their executions and events remain queryable after a restart.
// Instances are restored as metadata-only records and are not started.
func (m *Manager) reconcile() {
	records, err := m.metadata.List(context.Background())
	if err != nil {
		log.Printf("load instance metadata: %v", err)
		return
	}
	for _, rec := range records {
		i := restoreInstance(
			m.parent,
			rec,
			execution.NewExecutionStore(m.store),
			event.NewStore(m.store),
		)
		m.instancesByID[rec.ID] = i
		m.instancesByKey[i.Key] = i
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

// GetOrCreate returns the live runtime for key, lazily constructing it from
// the durable RegisteredSystem when it is not already running. Registration
// itself never creates an instance; this is the only entry point that does.
func (m *Manager) GetOrCreate(ctx context.Context, key protocol.InstanceKey) (*Instance, bool, error) {
	m.mu.RLock()
	previous, ok := m.instancesByKey[key]
	m.mu.RUnlock()
	if ok && (previous.Status() == StatusRunning || previous.Status() == StatusStarting) {
		return previous, false, nil
	}

	reg, err := m.systems.Get(ctx, key)
	if err != nil {
		return nil, false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := core2.NewID("inst_")
	if previous, ok := m.instancesByKey[key]; ok {
		id = core2.ID(previous.ID)
		delete(m.instancesByKey, key)
		delete(m.instancesByID, previous.ID)
	}

	i, err := New(m.parent, string(id), key, &reg.System, m.workers, m.store)
	if err != nil {
		return nil, false, err
	}
	if err := i.Start(); err != nil {
		return nil, false, err
	}

	m.instancesByKey[key] = i
	m.instancesByID[string(id)] = i
	if err := m.metadata.Save(context.Background(), recordFor(i, i.Status())); err != nil {
		log.Printf("persist instance metadata for %s: %v", i.ID, err)
	}
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
	m.mu.RLock()
	i, ok := m.instancesByID[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("instance %s not found", id)
	}

	if err := i.Stop(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.metadata.Save(context.Background(), recordFor(i, StatusStopped)); err != nil {
		log.Printf("persist instance metadata for %s: %v", i.ID, err)
	}
	return nil
}
