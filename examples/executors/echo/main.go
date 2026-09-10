// Command echo is an example Neuron executor. It is compiled twice from this
// single source:
//
//   - a native binary (runtime type "process")
//   - a WebAssembly module (GOOS=wasip1 GOARCH=wasm, runtime type "wasm")
//
// Both speak the exact same wire protocol: one JSON Request on stdin, one JSON
// Response on stdout, NEURON_EXECUTOR_* environment variables describing the
// execution. It deliberately imports nothing but the standard library so it
// builds offline and cross-compiles to wasip1 without any toolchain besides Go.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Request mirrors the protocol Request. It is declared locally (instead of
// importing the shared module) to keep this example dependency-free.
type Request struct {
	Input map[string]any `json:"input"`
}

// Response mirrors the protocol Response.
type Response struct {
	Output map[string]any `json:"output"`
	Error  string         `json:"error,omitempty"`
}

func main() {
	req, err := readRequest()
	if err != nil {
		writeError(fmt.Sprintf("invalid request: %v", err))
		os.Exit(1)
	}

	// A caller can trigger a controlled failure through the input contract:
	// {"input": {"error": "message"}}.
	if msg, ok := req.Input["error"].(string); ok && msg != "" {
		if err := writeResponse(Response{Output: map[string]any{}, Error: msg}); err != nil {
			os.Exit(1)
		}
		return
	}

	out := map[string]any{
		"input":    req.Input,
		"protocol": os.Getenv("NEURON_EXECUTOR_PROTOCOL"),
		"type":     os.Getenv("NEURON_EXECUTOR_TYPE"),
		"version":  os.Getenv("NEURON_EXECUTOR_VERSION"),
	}
	if value, ok := req.Input["value"]; ok {
		out["value"] = value
	}

	if err := writeResponse(Response{Output: out}); err != nil {
		os.Exit(1)
	}
}

func readRequest() (Request, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return Request{}, err
	}
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return Request{}, err
	}
	return req, nil
}

func writeResponse(resp Response) error {
	return json.NewEncoder(os.Stdout).Encode(resp)
}

func writeError(message string) {
	_ = writeResponse(Response{Output: map[string]any{}, Error: message})
}
