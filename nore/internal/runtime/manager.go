package runtime

import (
	"time"
)

func (e *Execution) Status() Status {
	return e.status
}

func (e *Execution) CompletedAt() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.completedAt == nil {
		return time.Time{}
	}
	return e.completedAt.UTC()
}

func (e *Execution) StartTime() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.startedAt == nil {
		return time.Time{}
	}
	return e.startedAt.UTC()
}

func (e *Execution) StartedAt() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.startedAt == nil {
		return time.Time{}
	}
	return e.startedAt.UTC()
}

func (e *Execution) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.executionError
}
