package runtime

import (
	"time"
)

func (e *Execution) Status() Status  {
	return e.status
}

func (e *Execution) CompletedAt() time.Time {
	return e.completedAt.UTC()
}

func (e *Execution) StartTime() time.Time {
	return e.startedAt.UTC()
}

func (e *Execution) StartedAt() time.Time {
	return e.startedAt.UTC()
}

func (e *Execution) Error() string {
	return e.executionError
}

