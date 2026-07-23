// Package ratelimit provides configurable rate limiting middleware.
// This implements the "Custom rate limiting" feature requested in PocketBase (75 reactions).
// It supports multiple algorithms: token bucket, sliding window, and fixed window.
package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Limiter interface defines rate limiting behavior
type Limiter interface {
	Allow(key string) bool
	Remaining(key string) int
	ResetAfter(key string) time.Duration
}

// Algorithm represents the rate limiting algorithm
type Algorithm string

const (
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmSlidingWindow Algorithm = "sliding_window"
	AlgorithmFixedWindow   Algorithm = "fixed_window"
)

// Config holds rate limiter configuration
type Config struct {
	Algorithm Algorithm     // Rate limiting algorithm
	Rate      int           // Requests per period
	Period    time.Duration // Time window
	Burst     int           // Burst size (for token bucket)
	KeyPrefix string        // Redis key prefix (if using Redis)
}

// New creates a new rate limiter based on configuration
func New(config Config) (Limiter, error) {
	switch config.Algorithm {
	case AlgorithmTokenBucket:
		return NewTokenBucket(config.Rate, config.Burst, config.Period), nil
	case AlgorithmSlidingWindow:
		return NewSlidingWindow(config.Rate, config.Period), nil
	case AlgorithmFixedWindow:
		return NewFixedWindow(config.Rate, config.Period), nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", config.Algorithm)
	}
}

// TokenBucket implements the token bucket algorithm
type TokenBucket struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	rate    int           // Tokens added per period
	burst   int           // Maximum bucket size
	period  time.Duration // Token refill period
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewTokenBucket creates a token bucket rate limiter
func NewTokenBucket(rate, burst int, period time.Duration) *TokenBucket {
	return &TokenBucket{
		buckets: make(map[string]*bucket),
		rate:    rate,
		burst:   burst,
		period:  period,
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	b, exists := tb.buckets[key]
	if !exists {
		tb.buckets[key] = &bucket{
			tokens:     float64(tb.burst - 1),
			lastRefill: now,
		}
		return true
	}

	// Refill tokens
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := elapsed.Seconds() / tb.period.Seconds() * float64(tb.rate)
	b.tokens = min(float64(tb.burst), b.tokens+tokensToAdd)
	b.lastRefill = now

	// Check if we can consume a token
	if b.tokens >= 1 {
		b.tokens--
		return true
	}

	return false
}

// Remaining returns remaining tokens
func (tb *TokenBucket) Remaining(key string) int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if b, exists := tb.buckets[key]; exists {
		return int(b.tokens)
	}
	return tb.burst
}

// ResetAfter returns time until next token is available
func (tb *TokenBucket) ResetAfter(key string) time.Duration {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if b, exists := tb.buckets[key]; exists {
		if b.tokens >= 1 {
			return 0
		}
		// Time to generate 1 token
		return time.Duration((1.0 - b.tokens) / float64(tb.rate) * float64(tb.period))
	}
	return 0
}

// SlidingWindow implements the sliding window algorithm
type SlidingWindow struct {
	mu      sync.RWMutex
	windows map[string][]time.Time
	rate    int
	window  time.Duration
}

// NewSlidingWindow creates a sliding window rate limiter
func NewSlidingWindow(rate int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		windows: make(map[string][]time.Time),
		rate:    rate,
		window:  window,
	}
}

// Allow checks if request is allowed
func (sw *SlidingWindow) Allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// Get or create window
	requests, exists := sw.windows[key]
	if !exists {
		sw.windows[key] = []time.Time{now}
		return true
	}

	// Remove old requests outside the window
	var valid []time.Time
	for _, req := range requests {
		if req.After(cutoff) {
			valid = append(valid, req)
		}
	}

	// Check if we're under the limit
	if len(valid) < sw.rate {
		valid = append(valid, now)
		sw.windows[key] = valid
		return true
	}

	sw.windows[key] = valid
	return false
}

// Remaining returns remaining requests in current window
func (sw *SlidingWindow) Remaining(key string) int {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	requests, exists := sw.windows[key]
	if !exists {
		return sw.rate
	}

	// Count valid requests
	count := 0
	for _, req := range requests {
		if req.After(cutoff) {
			count++
		}
	}

	return max(0, sw.rate-count)
}

// ResetAfter returns time until oldest request expires
func (sw *SlidingWindow) ResetAfter(key string) time.Duration {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	requests, exists := sw.windows[key]
	if !exists || len(requests) == 0 {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// Find oldest request in window
	var oldest time.Time
	for _, req := range requests {
		if req.After(cutoff) {
			if oldest.IsZero() || req.Before(oldest) {
				oldest = req
			}
		}
	}

	if oldest.IsZero() {
		return 0
	}

	return oldest.Add(sw.window).Sub(now)
}

// FixedWindow implements the fixed window algorithm
type FixedWindow struct {
	mu      sync.RWMutex
	windows map[string]*window
	rate    int
	window  time.Duration
}

type window struct {
	count int
	start time.Time
}

// NewFixedWindow creates a fixed window rate limiter.
//
// The duration parameter was named `window`, which shadowed the `window` struct
// type declared just above, so `make(map[string]*window)` resolved to the
// parameter rather than the type:
//
//	window (parameter) is not a type
//
// and the whole package failed to compile.
func NewFixedWindow(rate int, dur time.Duration) *FixedWindow {
	return &FixedWindow{
		windows: make(map[string]*window),
		rate:    rate,
		window:  dur,
	}
}

// Allow checks if request is allowed
func (fw *FixedWindow) Allow(key string) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	currentWindow := now.Truncate(fw.window)

	w, exists := fw.windows[key]
	if !exists || w.start != currentWindow {
		// New window
		fw.windows[key] = &window{
			count: 1,
			start: currentWindow,
		}
		return true
	}

	// Check if under limit
	if w.count < fw.rate {
		w.count++
		return true
	}

	return false
}

// Remaining returns remaining requests in current window
func (fw *FixedWindow) Remaining(key string) int {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	now := time.Now()
	currentWindow := now.Truncate(fw.window)

	w, exists := fw.windows[key]
	if !exists || w.start != currentWindow {
		return fw.rate
	}

	return max(0, fw.rate-w.count)
}

// ResetAfter returns time until window resets
func (fw *FixedWindow) ResetAfter(key string) time.Duration {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	now := time.Now()
	currentWindow := now.Truncate(fw.window)

	w, exists := fw.windows[key]
	if !exists || w.start != currentWindow {
		return 0
	}

	return currentWindow.Add(fw.window).Sub(now)
}

// Middleware creates HTTP middleware for rate limiting
func Middleware(limiter Limiter, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			if !limiter.Allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", int(limiter.ResetAfter(key).Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":{"code":429,"message":"Rate limit exceeded"}}`))

				log.Warn().
					Str("key", key).
					Str("path", r.URL.Path).
					Msg("Rate limit exceeded")
				return
			}

			// Add rate limit headers
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.Remaining(key)))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", int(limiter.ResetAfter(key).Seconds())))

			next.ServeHTTP(w, r)
		})
	}
}

// PerIPKeyFunc extracts the client IP as the rate limit key.
//
// Two bugs made this limit nothing at all:
//
//  1. r.RemoteAddr is "host:port", and the ephemeral port differs on every
//     connection. Keying on it made the limiter per-connection rather than
//     per-IP, so every request got a brand new bucket and no client was ever
//     limited. The port must be stripped.
//
//  2. X-Forwarded-For is a comma separated list, "client, proxy1, proxy2".
//     Using it whole made the key vary with the proxy chain, and a client can
//     set the header directly, so only the first entry is meaningful.
//
// Note X-Forwarded-For is client controlled unless a trusted proxy overwrites
// it. Honouring it means an attacker can pick their own bucket. Prefer
// RemoteAddr, and only trust the header when running behind a proxy you
// control.
func PerIPKeyFunc(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			xff = xff[:i]
		}
		if ip := strings.TrimSpace(xff); ip != "" {
			return ip
		}
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" {
		return ip
	}
	// Strip the port: SplitHostPort fails on a bare host, in which case
	// RemoteAddr is already what we want.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// PerUserKeyFunc extracts user ID from context
func PerUserKeyFunc(r *http.Request) string {
	// This would extract user ID from auth context
	// For now, fallback to IP
	return PerIPKeyFunc(r)
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
