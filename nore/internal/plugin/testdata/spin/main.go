// Command spin never returns. It exists to exercise the wasm runtime's
// per-execution timeout handling in tests.
package main

func main() {
	for i := uint64(0); ; i++ {
		// Keep the loop from being optimized into an empty body while
		// remaining pure CPU work that a context deadline can interrupt.
		_ = i * i
	}
}
