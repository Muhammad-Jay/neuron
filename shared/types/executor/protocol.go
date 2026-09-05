package executor

// Request is the stdin payload handed to a process executor. Executors read a
// single JSON document from stdin, execute, and write a single JSON Response
// document to stdout.
type Request struct {
	// Input carries the resolved execution input for the service. Keys map to
	// the executor's declared inputs.
	Input map[string]any `json:"input"`
}

// Response is the stdout payload produced by a process executor. Executors
// must write exactly one JSON Response document to stdout and exit zero on
// success.
type Response struct {
	// Output carries the results of the execution. Keys map to the executor's
	// declared outputs.
	Output map[string]any `json:"output"`

	// Error, when present, records a controlled failure. The executor may
	// still exit zero after reporting an error this way; N.O.R.E. treats a
	// non-empty Error as an execution failure and surfaces it downstream.
	Error string `json:"error,omitempty"`
}

// Environment variables injected into a process executor by the runtime.
const (
	// EnvProtocol declares the expected protocol version.
	EnvProtocol = "NEURON_EXECUTOR_PROTOCOL"

	// EnvType is the executor type (logical name) being executed.
	EnvType = "NEURON_EXECUTOR_TYPE"

	// EnvVersion is the exact resolved version of the executor.
	EnvVersion = "NEURON_EXECUTOR_VERSION"
)
