// Standalone harness that runs the slugify.wasm plugin against mock data using
// the SAME host ABI as internal/plugins/wasm.go. It has its own go.mod so it can
// run on an older Go toolchain than the main module (which targets go 1.24).
//
//	go run .            # uses ../slugify.wasm
//	go run . path.wasm  # custom path
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// execute mirrors internal/plugins/wasm.go Plugin.Execute exactly.
func execute(ctx context.Context, mod api.Module, fnName string, input []byte) ([]byte, error) {
	fn := mod.ExportedFunction(fnName)
	if fn == nil {
		return nil, fmt.Errorf("function %s not exported", fnName)
	}
	malloc := mod.ExportedFunction("malloc")
	if malloc == nil {
		return nil, fmt.Errorf("malloc not exported")
	}
	mem := mod.Memory()

	alloc := func(n int) (uint32, error) {
		res, err := malloc.Call(ctx, uint64(n))
		if err != nil {
			return 0, err
		}
		if len(res) == 0 || res[0] == 0 {
			return 0, fmt.Errorf("malloc(%d) failed", n)
		}
		return uint32(res[0]), nil
	}

	inPtr, err := alloc(len(input))
	if err != nil {
		return nil, err
	}
	if !mem.Write(inPtr, input) {
		return nil, fmt.Errorf("failed to write input")
	}
	outPtrPtr, err := alloc(4)
	if err != nil {
		return nil, err
	}
	outLenPtr, err := alloc(4)
	if err != nil {
		return nil, err
	}

	res, err := fn.Call(ctx, uint64(inPtr), uint64(len(input)), uint64(outPtrPtr), uint64(outLenPtr))
	if err != nil {
		return nil, err
	}
	if len(res) > 0 && res[0] != 0 {
		return nil, fmt.Errorf("plugin returned error code %d", res[0])
	}

	pb, ok := mem.Read(outPtrPtr, 4)
	if !ok {
		return nil, fmt.Errorf("read out ptr failed")
	}
	lb, ok := mem.Read(outLenPtr, 4)
	if !ok {
		return nil, fmt.Errorf("read out len failed")
	}
	outPtr := binary.LittleEndian.Uint32(pb)
	outLen := binary.LittleEndian.Uint32(lb)
	out, ok := mem.Read(outPtr, outLen)
	if !ok {
		return nil, fmt.Errorf("read output failed")
	}
	cp := make([]byte, len(out))
	copy(cp, out)
	return cp, nil
}

func main() {
	path := "../slugify.wasm"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read wasm:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	mod, err := rt.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithArgs("fieldstone-plugin"))
	if err != nil {
		fmt.Println("instantiate:", err)
		os.Exit(1)
	}

	// Mock data: titles a BaaS would receive in a beforeCreate hook.
	cases := []struct{ in, want string }{
		{"Shipping Realtime to Production!", "shipping-realtime-to-production"},
		{"  Hello, World  ", "hello-world"},
		{"Multi---tenancy   patterns", "multi-tenancy-patterns"},
		{"Café del Mar 2026", "caf-del-mar-2026"}, // non-ASCII dropped (ASCII-only slug)
		{"___edge__case___", "edge-case"},
		{"ALLCAPS", "allcaps"},
	}

	fmt.Printf("Loaded %s (%d bytes)\n\n", path, len(wasmBytes))
	fmt.Printf("%-36s -> %s\n", "INPUT", "SLUG")
	fmt.Println("---------------------------------------------------------------")
	pass, t0 := 0, time.Now()
	for _, c := range cases {
		out, err := execute(ctx, mod, "slugify", []byte(c.in))
		if err != nil {
			fmt.Printf("ERROR %q: %v\n", c.in, err)
			continue
		}
		got := string(out)
		ok := got == c.want
		mark := "ok"
		if !ok {
			mark = "FAIL want=" + c.want
		} else {
			pass++
		}
		fmt.Printf("%-36q -> %-32q [%s]\n", c.in, got, mark)
	}
	fmt.Printf("\n%d/%d passed in %s\n", pass, len(cases), time.Since(t0).Round(time.Microsecond))
	if pass != len(cases) {
		os.Exit(1)
	}
}
