package event

type Type uint16

// TODO(FIX): make sure ExecutionCompleted Event is being published and used
const (
	Unknown Type = iota
	ExecutionStarted
	ExecutionCompleted
	ExecutionFailed
	ExecutionCancelled
	ServiceReady
	ServiceStarted
	ServiceCompleted
	ServiceFailed
)