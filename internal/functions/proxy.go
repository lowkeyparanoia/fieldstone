// Package functions proxies requests to a Deno/Node sidecar for edge functions.
package functions

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Proxy is a reverse proxy to the functions host.
type Proxy struct {
	target *url.URL
	rp     *httputil.ReverseProxy
}

// NewProxy creates a new functions proxy.
func NewProxy(functionsHost string) (*Proxy, error) {
	if functionsHost == "" {
		functionsHost = "http://localhost:8000"
	}
	target, err := url.Parse(functionsHost)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		target: target,
		rp:     httputil.NewSingleHostReverseProxy(target),
	}, nil
}

// Handler proxies requests to /functions/v1/*.
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	// Strip /functions/v1 prefix
	path := strings.TrimPrefix(r.URL.Path, "/functions/v1")
	if path == r.URL.Path {
		http.Error(w, `{"error":"invalid functions path"}`, http.StatusBadRequest)
		return
	}
	r.URL.Path = path
	p.rp.ServeHTTP(w, r)
}
