package plugin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Muhammad-Jay/neuron/nore/internal/contracts"
	"github.com/Muhammad-Jay/neuron/nore/internal/registry"
	core "github.com/Muhammad-Jay/neuron/shared/types/core"
	shadexec "github.com/Muhammad-Jay/neuron/shared/types/executor"
)

const (
	echoSourceDir = "../../../examples/executors/echo"
	spinSourceDir = "testdata/spin"
)

// fixtures are compiled exactly once per test binary (see TestMain) so the
// suite spends most of its time executing, not rebuilding modules.
var fixtures struct {
	native, wasm, spin string
}

func buildGo(srcDir, out string, env []string) error {
	cmd := exec.Command("go", "build", "-C", srcDir, "-o", out, ".")
	cmd.Env = append(os.Environ(), append([]string{"GOWORK=off"}, env...)...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build (%s): %v\n%s", out, err, stderr.String())
	}
	return nil
}

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "neuron-plugin-fixtures-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	steps := []struct {
		src string
		out string
		env []string
	}{
		{echoSourceDir, filepath.Join(dir, "echo"), nil},
		{echoSourceDir, filepath.Join(dir, "echo.wasm"), []string{"GOOS=wasip1", "GOARCH=wasm"}},
		{spinSourceDir, filepath.Join(dir, "spin.wasm"), []string{"GOOS=wasip1", "GOARCH=wasm"}},
	}
	for _, s := range steps {
		if err := buildGo(s.src, s.out, s.env); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fixtures.native = filepath.Join(dir, "echo")
	fixtures.wasm = filepath.Join(dir, "echo.wasm")
	fixtures.spin = filepath.Join(dir, "spin.wasm")

	code := m.Run()

	if rt, rterr := sharedProcessRuntime(); rterr == nil {
		_ = rt.close(context.Background())
	}
	os.Exit(code)
}

func resolvedExecutor(t *testing.T, typ, entrypoint, rootDir, runtimeKind string) shadexec.ResolvedExecutor {
	t.Helper()
	return shadexec.ResolvedExecutor{
		Type:             typ,
		RequestedVersion: "1.0.0",
		ResolvedVersion:  "1.0.0",
		Registry:         "local",
		Runtime: shadexec.RuntimeInfo{
			Type:       runtimeKind,
			Protocol:   shadexec.ProtocolV1,
			Entrypoint: entrypoint,
		},
		RootDir: rootDir,
	}
}

func echoResolved(t *testing.T, typ, kind string) shadexec.ResolvedExecutor {
	t.Helper()
	entry, root := fixtures.wasm, filepath.Dir(fixtures.wasm)
	if kind == shadexec.RuntimeKindProcess {
		entry, root = fixtures.native, filepath.Dir(fixtures.native)
	}
	return resolvedExecutor(t, typ, filepath.Base(entry), root, kind)
}

func executionContext(input map[string]any) contracts.ExecutionContext {
	return contracts.ExecutionContext{
		ExecutionID:   "exec-1",
		CorrelationID: "corr-1",
		Service: core.Service{
			Metadata: core.Metadata{Name: "echo", Version: "1.0.0"},
			Type:     "example:echo",
		},
		Input: input,
	}
}

func assertEchoOutput(t *testing.T, got map[string]any, expectedType string) {
	t.Helper()

	if got["type"] != expectedType {
		t.Errorf("output.type = %v, want %s", got["type"], expectedType)
	}
	if got["protocol"] != shadexec.ProtocolV1 {
		t.Errorf("output.protocol = %v, want %s", got["protocol"], shadexec.ProtocolV1)
	}
	if got["version"] != "1.0.0" {
		t.Errorf("output.version = %v, want 1.0.0", got["version"])
	}
	if got["value"] != "hello" {
		t.Errorf("output.value = %v, want hello", got["value"])
	}
}

func TestSharedWasmRuntimeIsProcessGlobal(t *testing.T) {
	a, err := sharedProcessRuntime()
	if err != nil {
		t.Fatal(err)
	}
	b, err := sharedProcessRuntime()
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("wazero runtime is recreated per call; want one per process")
	}

	cm1, err := a.compiledModule(context.Background(), fixtures.wasm)
	if err != nil {
		t.Fatal(err)
	}
	cm2, err := b.compiledModule(context.Background(), fixtures.wasm)
	if err != nil {
		t.Fatal(err)
	}
	if cm1 != cm2 {
		t.Fatal("compiled module is not cached; want one per frozen module")
	}
}

func TestRegisterResolvedExecutorsDispatch(t *testing.T) {
	cases := []struct {
		name string
		kind string
		res  shadexec.ResolvedExecutor
	}{
		{"process", shadexec.RuntimeKindProcess, echoResolved(t, "example:echo", shadexec.RuntimeKindProcess)},
		{"process_implicit_default", "", echoResolved(t, "example:echo", "")},
		{"wasm", shadexec.RuntimeKindWasm, echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Dispatch through the public registry path.
			reg := registry.New()
			if err := RegisterResolvedExecutors(reg, []shadexec.ResolvedExecutor{tc.res}); err != nil {
				t.Fatalf("RegisterResolvedExecutors: %v", err)
			}
			ex, err := reg.Resolve(core.ServiceType(tc.res.Type))
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if ex == nil {
				t.Fatal("resolved executor is nil")
			}
			defer closeIfCloser(t, ex)
		})
	}
}

func TestNewAdapterRejectsUnknownRuntime(t *testing.T) {
	res := shadexec.ResolvedExecutor{
		Type: "example:echo",
		Runtime: shadexec.RuntimeInfo{Type: "container", Entrypoint: "echo"},
	}
	_, err := NewAdapter(res)
	if err == nil {
		t.Fatal("expected error for unsupported runtime kind")
	}
	if !strings.Contains(err.Error(), "unsupported runtime kind") {
		t.Errorf("error = %q, want unsupported runtime kind", err)
	}
}

func TestProcessAdapterRoundTripWithEnv(t *testing.T) {
	adapter, err := NewProcessAdapter(echoResolved(t, "example:echo", shadexec.RuntimeKindProcess))
	if err != nil {
		t.Fatal(err)
	}
	defer closeIfCloser(t, adapter)

	got, err := adapter.Execute(context.Background(), executionContext(map[string]any{"value": "hello"}))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	assertEchoOutput(t, got, "example:echo")
}

func TestWasmAdapterRoundTripWithEnv(t *testing.T) {
	adapter, err := NewWasmAdapter(echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm))
	if err != nil {
		t.Fatal(err)
	}
	defer closeIfCloser(t, adapter)

	got, err := adapter.Execute(context.Background(), executionContext(map[string]any{"value": "hello"}))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	assertEchoOutput(t, got, "example:echo-wasm")
}

func TestAdaptersSurfaceControlledError(t *testing.T) {
	for _, tc := range []struct {
		name string
		res  shadexec.ResolvedExecutor
	}{
		{"process", echoResolved(t, "example:echo", shadexec.RuntimeKindProcess)},
		{"wasm", echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := NewAdapter(tc.res)
			if err != nil {
				t.Fatal(err)
			}
			defer closeIfCloser(t, adapter)

			_, err = adapter.Execute(context.Background(), executionContext(map[string]any{"error": "boom"}))
			if err == nil {
				t.Fatal("expected controlled error from executor")
			}
			if !strings.Contains(err.Error(), "boom") {
				t.Errorf("error = %q, want boom", err)
			}
		})
	}
}

func TestProcessAdapterMissingEntrypoint(t *testing.T) {
	res := resolvedExecutor(t, "example:echo", "does-not-exist", t.TempDir(), shadexec.RuntimeKindProcess)
	if _, err := NewAdapter(res); err == nil {
		t.Fatal("expected error for missing entrypoint")
	}
}

func TestWasmAdapterMissingModule(t *testing.T) {
	res := resolvedExecutor(t, "example:echo-wasm", "missing.wasm", t.TempDir(), shadexec.RuntimeKindWasm)
	if _, err := NewAdapter(res); err == nil {
		t.Fatal("expected error for missing wasm module")
	}
}

func TestWasmAdapterTimesOut(t *testing.T) {
	res := resolvedExecutor(t, "example:spin", filepath.Base(fixtures.spin), filepath.Dir(fixtures.spin), shadexec.RuntimeKindWasm)

	adapter, err := NewWasmAdapter(res)
	if err != nil {
		t.Fatal(err)
	}
	defer closeIfCloser(t, adapter)
	adapter.SetUnitTimeout(300 * time.Millisecond)

	start := time.Now()
	_, err = adapter.Execute(context.Background(), executionContext(map[string]any{"value": "x"}))
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %q, want timed out", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("timeout took too long: %v", elapsed)
	}
}

func TestWasmAdapterReusesRuntimeAcrossExecutions(t *testing.T) {
	adapter, err := NewWasmAdapter(echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm))
	if err != nil {
		t.Fatal(err)
	}
	defer closeIfCloser(t, adapter)

	for i := 0; i < 3; i++ {
		got, err := adapter.Execute(context.Background(), executionContext(map[string]any{"value": "hi"}))
		if err != nil {
			t.Fatalf("Execute #%d: %v", i, err)
		}
		if got["value"] != "hi" {
			t.Errorf("Execute #%d: value = %v, want hi", i, got["value"])
		}
	}
}

// TestWasmAdapterConcurrentExecutions proves executions do not serialize: many
// adapters (instances) hot-sharing one runtime and one compiled module must
// serve concurrent executions from their own sandboxed modules.
func TestWasmAdapterConcurrentExecutions(t *testing.T) {
	res := echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm)

	adapters := make([]*WasmAdapter, 4)
	for i := range adapters {
		adapter, err := NewWasmAdapter(res)
		if err != nil {
			t.Fatal(err)
		}
		adapters[i] = adapter
		if i > 0 && adapter.runtime != adapters[0].runtime {
			t.Fatal("adapters do not share the process runtime")
		}
		if i > 0 && adapter.compiled != adapters[0].compiled {
			t.Fatal("adapters do not share the compiled module")
		}
	}
	defer closeIfCloser(t, adapters[0])

	var wg sync.WaitGroup
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(adapter *WasmAdapter, g int) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				got, err := adapter.Execute(context.Background(), executionContext(map[string]any{"value": "hi"}))
				if err != nil {
					t.Errorf("goroutine %d Execute #%d: %v", g, i, err)
					return
				}
				if got["value"] != "hi" {
					t.Errorf("goroutine %d Execute #%d: value = %v, want hi", g, i, got["value"])
				}
			}
		}(adapters[g], g)
	}
	wg.Wait()
}

// TestWasmRuntimeSurvivesAdapterClose verifies Close is a no-op: the shared
// runtime keeps serving other adapters after one is closed.
func TestWasmRuntimeSurvivesAdapterClose(t *testing.T) {
	res := echoResolved(t, "example:echo-wasm", shadexec.RuntimeKindWasm)

	adapter, err := NewWasmAdapter(res)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Execute(context.Background(), executionContext(map[string]any{"value": "hi"})); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if err := adapter.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	again, err := NewWasmAdapter(res)
	if err != nil {
		t.Fatal(err)
	}
	defer closeIfCloser(t, again)
	got, err := again.Execute(context.Background(), executionContext(map[string]any{"value": "hi"}))
	if err != nil {
		t.Fatalf("Execute after Close: %v", err)
	}
	if got["value"] != "hi" {
		t.Errorf("value = %v, want hi", got["value"])
	}
}

func closeIfCloser(t *testing.T, ex any) {
	t.Helper()
	closer, ok := ex.(contracts.ExecutorCloser)
	if !ok {
		return
	}
	if err := closer.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}