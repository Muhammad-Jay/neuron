package execution

import "context"

// Done exposes a channel that is closed exactly once the execution reaches a
// terminal state (completed, failed, or cancelled). It lets callers integrate
// execution completion into select-based concurrency without polling.
func (e *Execution) Done() <-chan struct{} {
	if e == nil {
		// A nil execution is never valid; return a never-closed channel so
		// callers that guard with a nil check cannot deadlock on a nil map.
		return make(chan struct{})
	}
	return e.done
}

// Wait blocks until the execution reaches a terminal state. It returns nil
// when the execution completed, the execution's terminal error when it failed
// or was cancelled, or ctx.Err() if the context is cancelled first.
//
// The execution engine always runs asynchronously; Wait only observes the
// terminal signal and never drives the execution itself.
func (e *Execution) Wait(ctx context.Context) error {
	if e == nil {
		return context.Canceled
	}
	select {
	case <-e.done:
		if msg := e.Error(); msg != "" {
			return &ExecutionError{ExecutionID: string(e.ID), Message: msg}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ExecutionError is the terminal error surfaced by Wait for a failed or
// cancelled execution. It carries the execution identifier for attribution.
type ExecutionError struct {
	ExecutionID string
	Message     string
}

func (e *ExecutionError) Error() string {
	if e.Message == "" {
		return "execution " + e.ExecutionID + " did not complete"
	}
	return e.Message
}
