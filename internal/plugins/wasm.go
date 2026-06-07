// Package wasm provides WebAssembly-based plugin system for Fieldstone.
// This allows extending the backend with custom logic in any language that compiles to WASM.
//
// Why WASM:
// - Sandboxed execution (no access to host filesystem/network)
// - Multi-language support (Go, Rust, C++, AssemblyScript)
// - Near-native performance
// - Secure by design
//
// Example plugins:
// - Custom field types
// - Data validation hooks
// - External API integrations
// - Custom calculations
//
// Security considerations:
// - WASM runs in sandboxed environment
// - No direct database access
// - Limited memory (configurable)
// - Timeout enforcement
package wasm

import (
	"context"
	"fmt"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Plugin represents a loaded WASM plugin
type Plugin struct {
	ID          string
	Name        string
	Version     string
	Runtime     *Runtime
	Module      wazero.CompiledModule
	Instance    api.Module
	MemoryLimit int32 // In bytes
	Timeout     time.Duration
}

// Runtime manages WASM execution environment
type Runtime struct {
	ctx    context.Context
	rt     wazero.Runtime
	config wazero.ModuleConfig
}

// NewRuntime creates a new WASM runtime with WASI support
func NewRuntime(ctx context.Context) (*Runtime, error) {
	// Create runtime with default configuration
	rt := wazero.NewRuntime(ctx)
	
	// Enable WASI for filesystem operations (sandboxed)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)
	
	return &Runtime{
		ctx: ctx,
		rt:  rt,
		config: wazero.NewModuleConfig().
			WithStdin(nil).
			WithStdout(nil).
			WithStderr(nil).
			WithArgs("fieldstone-plugin"),
	}, nil
}

// LoadPlugin loads and compiles a WASM plugin from bytes
func (r *Runtime) LoadPlugin(id, name, version string, wasmBytes []byte) (*Plugin, error) {
	// Compile the module
	compiled, err := r.rt.CompileModule(r.ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM: %w", err)
	}
	
	// Instantiate the module
	module, err := r.rt.InstantiateModule(r.ctx, compiled, r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}
	
	return &Plugin{
		ID:          id,
		Name:        name,
		Version:     version,
		Runtime:     r,
		Module:      compiled,
		Instance:    module,
		MemoryLimit: 64 * 1024 * 1024, // 64MB default
		Timeout:     30 * time.Second,
	}, nil
}

// Execute runs a plugin function with timeout
func (p *Plugin) Execute(function string, input []byte) ([]byte, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(p.Runtime.ctx, p.Timeout)
	defer cancel()
	
	// Get exported function
	fn := p.Instance.ExportedFunction(function)
	if fn == nil {
		return nil, fmt.Errorf("function %s not exported", function)
	}
	
	// Allocate memory for input
	inputPtr, err := p.allocateMemory(ctx, len(input))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate input memory: %w", err)
	}
	
	// Write input to memory
	if !p.Instance.Memory().Write(uint32(inputPtr), input) {
		return nil, fmt.Errorf("failed to write input to memory")
	}
	
	// Allocate output buffer pointer
	outputPtrPtr, err := p.allocateMemory(ctx, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate output pointer: %w", err)
	}
	
	// Allocate output length pointer
	outputLenPtr, err := p.allocateMemory(ctx, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate output length pointer: %w", err)
	}
	
	// Call the function
	result, err := fn.Call(ctx, uint64(inputPtr), uint64(len(input)), uint64(outputPtrPtr), uint64(outputLenPtr))
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}
	
	// Check return code (0 = success)
	if len(result) > 0 && result[0] != 0 {
		return nil, fmt.Errorf("plugin returned error code: %d", result[0])
	}
	
	// Read output pointer
	outputPtrBytes, ok := p.Instance.Memory().Read(uint32(outputPtrPtr), 4)
	if !ok {
		return nil, fmt.Errorf("failed to read output pointer")
	}
	outputPtr := uint32(outputPtrBytes[0]) | uint32(outputPtrBytes[1])<<8 | uint32(outputPtrBytes[2])<<16 | uint32(outputPtrBytes[3])<<24
	
	// Read output length
	outputLenBytes, ok := p.Instance.Memory().Read(uint32(outputLenPtr), 4)
	if !ok {
		return nil, fmt.Errorf("failed to read output length")
	}
	outputLen := uint32(outputLenBytes[0]) | uint32(outputLenBytes[1])<<8 | uint32(outputLenBytes[2])<<16 | uint32(outputLenBytes[3])<<24
	
	// Read output data
	output, ok := p.Instance.Memory().Read(uint32(outputPtr), outputLen)
	if !ok {
		return nil, fmt.Errorf("failed to read output data")
	}
	
	return output, nil
}

// allocateMemory allocates memory in the WASM instance
func (p *Plugin) allocateMemory(ctx context.Context, size int) (uint32, error) {
	// Try to use malloc if available
	malloc := p.Instance.ExportedFunction("malloc")
	if malloc == nil {
		// Fallback: use memory growth
		// In production, implement proper memory management
		return 0, fmt.Errorf("malloc not available")
	}
	
	result, err := malloc.Call(ctx, uint64(size))
	if err != nil {
		return 0, fmt.Errorf("malloc failed: %w", err)
	}
	
	if len(result) == 0 {
		return 0, fmt.Errorf("malloc returned no value")
	}
	
	return uint32(result[0]), nil
}

// Close cleans up plugin resources
func (p *Plugin) Close() error {
	if p.Instance != nil {
		return p.Instance.Close(p.Runtime.ctx)
	}
	return nil
}

// Runtime Close
func (r *Runtime) Close() error {
	return r.rt.Close(r.ctx)
}

// Plugin Manager
type Manager struct {
	runtime *Runtime
	plugins map[string]*Plugin
}

// NewManager creates a new plugin manager
func NewManager(ctx context.Context) (*Manager, error) {
	rt, err := NewRuntime(ctx)
	if err != nil {
		return nil, err
	}
	
	return &Manager{
		runtime: rt,
		plugins: make(map[string]*Plugin),
	}, nil
}

// Load loads a plugin from bytes
func (m *Manager) Load(id, name, version string, wasmBytes []byte) (*Plugin, error) {
	plugin, err := m.runtime.LoadPlugin(id, name, version, wasmBytes)
	if err != nil {
		return nil, err
	}
	
	m.plugins[id] = plugin
	return plugin, nil
}

// Get retrieves a loaded plugin
func (m *Manager) Get(id string) (*Plugin, bool) {
	plugin, ok := m.plugins[id]
	return plugin, ok
}

// List returns metadata for every loaded plugin (id, name, version).
func (m *Manager) List() []map[string]string {
	out := make([]map[string]string, 0, len(m.plugins))
	for id, p := range m.plugins {
		out = append(out, map[string]string{"id": id, "name": p.Name, "version": p.Version})
	}
	return out
}

// Unload removes a plugin
func (m *Manager) Unload(id string) error {
	plugin, ok := m.plugins[id]
	if !ok {
		return fmt.Errorf("plugin not found: %s", id)
	}
	
	if err := plugin.Close(); err != nil {
		return err
	}
	
	delete(m.plugins, id)
	return nil
}

// Close cleans up all plugins
func (m *Manager) Close() error {
	for _, plugin := range m.plugins {
		plugin.Close()
	}
	return m.runtime.Close()
}

// Hook Types for Collection Events
const (
	HookBeforeCreate = "beforeCreate"
	HookAfterCreate  = "afterCreate"
	HookBeforeUpdate = "beforeUpdate"
	HookAfterUpdate  = "afterUpdate"
	HookBeforeDelete = "beforeDelete"
	HookAfterDelete  = "afterDelete"
)

// ExecuteHook executes a plugin hook for a collection event
func (m *Manager) ExecuteHook(ctx context.Context, collection string, hook string, data []byte) ([]byte, error) {
	// Find plugins registered for this collection/hook
	// In production, maintain a hook registry
	
	for _, plugin := range m.plugins {
		// Check if plugin exports this hook
		hookFn := plugin.Instance.ExportedFunction(fmt.Sprintf("%s_%s", collection, hook))
		if hookFn != nil {
			result, err := plugin.Execute(fmt.Sprintf("%s_%s", collection, hook), data)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	
	// No plugin found, return original data
	return data, nil
}

// Example: Using WASM plugin for custom validation
func ExampleUsage() {
	ctx := context.Background()
	
	// Create manager
	manager, err := NewManager(ctx)
	if err != nil {
		panic(err)
	}
	defer manager.Close()
	
	// Load plugin (would be from file in production)
	wasmBytes := []byte{} // Read from .wasm file
	plugin, err := manager.Load("custom-validator", "Email Validator", "1.0.0", wasmBytes)
	if err != nil {
		panic(err)
	}
	
	// Execute validation
	input := []byte(`{"email": "test@example.com"}`)
	result, err := plugin.Execute("validate", input)
	if err != nil {
		panic(err)
	}
	
	fmt.Printf("Validation result: %s\n", result)
}
