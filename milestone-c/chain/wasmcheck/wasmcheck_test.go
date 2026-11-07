// Package wasmcheck is the S7 live wasm sandbox capability check
// (Milestone C): it proves the pinned wasm VM (wasmvm/v2 v2.3.5,
// wasmd v0.60.9 matrix) instantiates on this machine. It is a
// capability check only: it does not wire CosmWasm into the app and
// does not make M2 depend on CosmWasm.
package wasmcheck

import (
	"testing"

	wasmvm "github.com/CosmWasm/wasmvm/v2"
)

func TestWasmVMInstantiates(t *testing.T) {
	ver, err := wasmvm.LibwasmvmVersion()
	if err != nil {
		t.Fatalf("LibwasmvmVersion: %v", err)
	}
	if ver == "" {
		t.Fatal("LibwasmvmVersion returned an empty version")
	}
	t.Logf("libwasmvm version: %s", ver)

	vm, err := wasmvm.NewVM(t.TempDir(), []string{"iterator", "staking", "stargate"}, 512, false, 0)
	if err != nil {
		t.Fatalf("NewVM: %v", err)
	}
	vm.Cleanup()
}
