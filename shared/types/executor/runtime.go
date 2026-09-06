package executor

// Runtime kinds identify how a frozen executor artifact is launched. The value
// stored in Manifest.Runtime.Type and RuntimeInfo.Type is one of these
// constants. The plugin layer dispatches on it to select the adapter.
const (
	// RuntimeKindProcess runs the entrypoint as an OS child process.
	RuntimeKindProcess = "process"

	// RuntimeKindWasm runs the entrypoint inside an embedded WASI runtime,
	// speaking the same stdin/stdout JSON protocol as process executors.
	RuntimeKindWasm = "wasm"

	// RuntimeKindContainer, RuntimeKindRemote, and RuntimeKindEmbed are
	// reserved for future launchers. The plugin layer rejects them today with
	// an explicit "unsupported runtime kind" error rather than mis-executing.
	RuntimeKindContainer = "container"
	RuntimeKindRemote    = "remote"
)

// SupportedRuntimeKinds returns the runtime kinds the plugin layer can launch.
func SupportedRuntimeKinds() []string {
	return []string{RuntimeKindProcess, RuntimeKindWasm}
}