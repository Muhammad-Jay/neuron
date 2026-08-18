package event

type Type uint16

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
	ServiceLog
)

const All Type = 0xffff
