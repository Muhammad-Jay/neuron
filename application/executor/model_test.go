package executor

import (
	"runtime"
	"testing"

	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

func TestPlatformForRuntime(t *testing.T) {
	cases := []struct {
		runtimeType string
		want        string
	}{
		{shadexec.RuntimeKindWasm, shadexec.ExecutorPlatformWasm},
		{shadexec.RuntimeKindProcess, HostPlatform()},
		{"", HostPlatform()},
		{shadexec.RuntimeKindContainer, HostPlatform()},
		{shadexec.RuntimeKindRemote, HostPlatform()},
	}
	for _, tc := range cases {
		if got := PlatformForRuntime(tc.runtimeType); got != tc.want {
			t.Errorf("PlatformForRuntime(%q) = %q, want %q", tc.runtimeType, got, tc.want)
		}
	}
	if HostPlatform() == shadexec.ExecutorPlatformWasm {
		t.Error("HostPlatform must not collide with the wasm platform key")
	}
}

func TestHostPlatformShape(t *testing.T) {
	got := HostPlatform()
	if got != runtime.GOOS+"-"+runtime.GOARCH {
		t.Errorf("HostPlatform() = %q, want %q", got, runtime.GOOS+"-"+runtime.GOARCH)
	}
}
