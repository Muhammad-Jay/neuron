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

func (t Type) String() string {
	switch t {
	case ExecutionStarted:
		return "execution.started"
	case ExecutionCompleted:
		return "execution.completed"
	case ExecutionFailed:
		return "execution.failed"
	case ExecutionCancelled:
		return "execution.cancelled"
	case ServiceReady:
		return "service.ready"
	case ServiceStarted:
		return "service.started"
	case ServiceCompleted:
		return "service.completed"
	case ServiceFailed:
		return "service.failed"
	case ServiceLog:
		return "service.log"
	default:
		return "unknown"
	}
}
