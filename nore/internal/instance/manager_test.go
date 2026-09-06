package instance

import (
	"testing"

	"github.com/Muhammad-Jay/neuron/nore/internal/system"
	"github.com/Muhammad-Jay/neuron/shared/types/protocol"
)

func TestWithExecutorsInvalidPayloadReturnsError(t *testing.T) {
	_, err := withExecutors(system.RegisteredSystem{
		Key:                     protocol.InstanceKey{SystemID: "sys_test"},
		ExecutionConfigurations: make(chan int),
	})
	if err == nil {
		t.Fatal("withExecutors() error = nil, want decode failure")
	}
}

func TestWithExecutorsValidPayloadReturnsOption(t *testing.T) {
	opt, err := withExecutors(system.RegisteredSystem{
		Key: protocol.InstanceKey{SystemID: "sys_test"},
		ExecutionConfigurations: map[string]any{
			"resolved_executors": []any{},
		},
	})
	if err != nil {
		t.Fatalf("withExecutors() error = %v", err)
	}
	if opt == nil {
		t.Fatal("withExecutors() option = nil, want non-nil Option")
	}
}
