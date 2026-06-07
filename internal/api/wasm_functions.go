package api

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	wasm "github.com/fieldstone/fieldstone/internal/plugins"
)

// SetPluginManager wires the in-process WASM functions runtime. When set, the
// /api/functions/wasm/* routes execute sandboxed WASM plugins (the same wazero
// engine as internal/plugins) as fast, network-isolated transforms — a
// complement to the Deno/Node edge-functions sidecar at /functions/v1/*.
func (s *Server) SetPluginManager(m *wasm.Manager) { s.pluginManager = m }

// handleWasmList lists the loaded WASM function plugins.
// GET /api/functions/wasm
func (s *Server) handleWasmList(w http.ResponseWriter, _ *http.Request) {
	if s.pluginManager == nil {
		s.sendJSON(w, http.StatusOK, map[string]interface{}{"enabled": false, "plugins": []interface{}{}})
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"enabled": true, "plugins": s.pluginManager.List()})
}

// handleWasmTransform runs a function exported by a loaded WASM plugin.
// POST /api/functions/wasm/{plugin}/{fn} — the request body is passed to the
// plugin verbatim and the plugin's output is returned verbatim.
func (s *Server) handleWasmTransform(w http.ResponseWriter, r *http.Request) {
	if s.pluginManager == nil {
		s.sendError(w, http.StatusServiceUnavailable, "wasm functions runtime not enabled")
		return
	}
	pluginID := chi.URLParam(r, "plugin")
	fn := chi.URLParam(r, "fn")

	p, ok := s.pluginManager.Get(pluginID)
	if !ok {
		s.sendError(w, http.StatusNotFound, "plugin not loaded: "+pluginID)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<20)) // 4 MiB cap
	if err != nil {
		s.sendError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	out, err := p.Execute(fn, body)
	if err != nil {
		s.sendError(w, http.StatusBadGateway, "wasm execution failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}
