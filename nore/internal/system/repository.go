package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Muhammad-Jay/neuron/nore/internal/storage"
	"github.com/Muhammad-Jay/neuron/nore/internal/storage/sqlite"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

// Repository persists RegisteredSystem artifacts under
// systems/{systemID}/{version}/{hash}/{env}/manifest keys in the shared
// storage.Store. Each content hash is an immutable artifact: registering a
// system with a different hash produces a new key, never a mutation of an
// existing one.
type Repository struct {
	store storage.Store
}

func NewRepository(store storage.Store) *Repository {
	return &Repository{store: store}
}

// Register stores registration. It is idempotent: registering the same key
// again updates nothing and returns created=false.
func (r *Repository) Register(ctx context.Context, reg RegisteredSystem) (bool, error) {
	key := r.key(reg.Key)
	if _, err := r.store.Get(ctx, key); err == nil {
		return false, nil
	}

	data, err := json.Marshal(reg)
	if err != nil {
		return false, fmt.Errorf("marshal registered system: %w", err)
	}
	if err := r.store.Put(ctx, key, data); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repository) Get(ctx context.Context, key protocol.InstanceKey) (RegisteredSystem, error) {
	reg, err := r.resolve(ctx, key)
	if err != nil {
		return RegisteredSystem{}, err
	}
	return reg, nil
}

// resolve loads the artifact for key. When key.Env is empty the artifact is
// resolved regardless of environment: if exactly one environment registered the
// same system/version/hash it is returned, otherwise an ambiguity error.
func (r *Repository) resolve(ctx context.Context, key protocol.InstanceKey) (RegisteredSystem, error) {
	if key.Env != "" {
		reg, err := r.getByStorageKey(ctx, r.key(key))
		if err != nil {
			return RegisteredSystem{}, fmt.Errorf("registered system %s: %w", key.String(), err)
		}
		return reg, nil
	}

	prefix := sqlite.SanitizeKey("systems", key.SystemID, key.Version, key.Hash)
	keys, err := r.store.List(ctx, prefix)
	if err != nil {
		return RegisteredSystem{}, err
	}

	var found string
	var reg RegisteredSystem
	for _, storageKey := range keys {
		if !strings.HasSuffix(storageKey, "/manifest") {
			continue
		}
		candidate, err := r.getByStorageKey(ctx, storageKey)
		if err != nil {
			continue
		}
		env := envFromStorageKey(storageKey)
		if found != "" && env != envFromStorageKey(found) {
			return RegisteredSystem{}, fmt.Errorf(
				"system %s is registered in multiple environments (%s, %s); specify env in the key",
				key.String(), envFromStorageKey(found), env,
			)
		}
		found = storageKey
		reg = candidate
	}

	if found == "" {
		return RegisteredSystem{}, fmt.Errorf("%w: no registration for %s", storage.ErrNotFound, key.String())
	}
	return reg, nil
}

// envFromStorageKey extracts the environment bucket from a manifest storage key
// (systems/{id}/{version}/{hash}/{env}/manifest).
func envFromStorageKey(storageKey string) string {
	trimmed := strings.TrimSuffix(storageKey, "/manifest")
	segments := strings.Split(trimmed, "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

func (r *Repository) Exists(ctx context.Context, key protocol.InstanceKey) (bool, error) {
	_, err := r.resolve(ctx, key)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, storage.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (r *Repository) Delete(ctx context.Context, key protocol.InstanceKey) error {
	return r.store.Delete(ctx, r.key(key))
}

// List returns every registered artifact, one per indexing system/version/hash.
func (r *Repository) List(ctx context.Context) ([]RegisteredSystem, error) {
	keys, err := r.store.List(ctx, "systems/")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	records := make([]RegisteredSystem, 0, len(keys))
	for _, key := range keys {
		rec, err := r.getByStorageKey(ctx, key)
		if err != nil {
			continue
		}
		identity := rec.Key.SystemID + "\x00" + rec.Key.Version + "\x00" + rec.Key.Hash
		if seen[identity] {
			continue
		}
		seen[identity] = true
		records = append(records, rec)
	}
	return records, nil
}

func (r *Repository) key(key protocol.InstanceKey) string {
	return sqlite.SanitizeKey("systems", key.SystemID, key.Version, key.Hash, key.Env, "manifest")
}

func (r *Repository) getByStorageKey(ctx context.Context, storageKey string) (RegisteredSystem, error) {
	var reg RegisteredSystem
	data, err := r.store.Get(ctx, storageKey)
	if err != nil {
		return reg, err
	}
	if err := json.Unmarshal(data, &reg); err != nil {
		return reg, err
	}
	return reg, nil
}