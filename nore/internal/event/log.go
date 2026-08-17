package event

import "github.com/Muhammad-Jay/neuron/shared/types/core"

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

type LogPayload struct {
	Level   LogLevel
	Message string
	NodeID  core.ID
}