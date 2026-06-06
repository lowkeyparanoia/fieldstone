package wasm

import (
	"context"
	"os"
	"testing"
)

// TestSlugifyPlugin loads the example slugify WASM plugin through the real
// plugin Manager/Runtime and exercises it with mock record data — the same
// path Fieldstone would use to run a plugin in production.
//
// Build the artifact first:  internal/plugins/examples/slugify/build.sh
// Then:                      go test ./internal/plugins -run TestSlugify -v
func TestSlugifyPlugin(t *testing.T) {
	const wasmPath = "examples/slugify/slugify.wasm"
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Skipf("plugin not built (%s): run examples/slugify/build.sh — %v", wasmPath, err)
	}

	ctx := context.Background()
	mgr, err := NewManager(ctx)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Close()

	plugin, err := mgr.Load("slugify", "Slugify", "1.0.0", wasmBytes)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Mock data: titles a beforeCreate hook would receive.
	cases := []struct{ in, want string }{
		{"Shipping Realtime to Production!", "shipping-realtime-to-production"},
		{"  Hello, World  ", "hello-world"},
		{"Multi---tenancy   patterns", "multi-tenancy-patterns"},
		{"___edge__case___", "edge-case"},
		{"ALLCAPS", "allcaps"},
		{"", ""},
	}

	for _, c := range cases {
		out, err := plugin.Execute("slugify", []byte(c.in))
		if err != nil {
			t.Errorf("Execute(%q): %v", c.in, err)
			continue
		}
		if got := string(out); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
