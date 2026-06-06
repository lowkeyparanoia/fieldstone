// Package cache provides HTTP caching middleware and helpers for Fieldstone.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// MiddlewareConfig holds configuration for cache middleware
type MiddlewareConfig struct {
	Cache         Cache
	DefaultTTL    time.Duration
	KeyPrefix     string
	ExcludePaths  []string
	ExcludeMethods []string
}

// Middleware creates HTTP caching middleware
func Middleware(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	if cfg.DefaultTTL == 0 {
		cfg.DefaultTTL = 5 * time.Minute
	}
	if cfg.ExcludeMethods == nil {
		cfg.ExcludeMethods = []string{"POST", "PUT", "DELETE", "PATCH"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip excluded methods
			for _, method := range cfg.ExcludeMethods {
				if r.Method == method {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Skip excluded paths
			for _, path := range cfg.ExcludePaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Generate cache key
			cacheKey := generateCacheKey(r, cfg.KeyPrefix)

			// Try to get from cache
			ctx := r.Context()
			cached, err := cfg.Cache.Get(ctx, cacheKey)
			if err == nil {
				// Cache HIT
				log.Debug().
					Str("key", cacheKey).
					Str("path", r.URL.Path).
					Msg("Cache HIT")
				
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(cached))
				return
			}

			// Cache MISS - wrap response writer
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           []byte{},
			}

			next.ServeHTTP(wrapped, r)

			// Cache successful GET responses
			if wrapped.statusCode == http.StatusOK && len(wrapped.body) > 0 {
				ttl := cfg.DefaultTTL
				// Check for custom TTL in header
				if customTTL := r.Header.Get("X-Cache-TTL"); customTTL != "" {
					if d, err := time.ParseDuration(customTTL); err == nil {
						ttl = d
					}
				}

				if err := cfg.Cache.Set(ctx, cacheKey, string(wrapped.body), ttl); err != nil {
					log.Error().Err(err).Str("key", cacheKey).Msg("Failed to cache response")
				} else {
					log.Debug().
						Str("key", cacheKey).
						Str("path", r.URL.Path).
						Dur("ttl", ttl).
						Msg("Cache MISS - stored")
				}
			}

			w.Header().Set("X-Cache", "MISS")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// generateCacheKey creates a unique cache key for a request
func generateCacheKey(r *http.Request, prefix string) string {
	// Include tenant ID if available
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	// Include user ID if authenticated
	userID := r.Header.Get("X-User-ID")

	// Build key components
	// Hash Authorization header so unauthenticated requests never hit
	// cache entries that were stored by authenticated requests.
	authHash := ""
	if auth := r.Header.Get("Authorization"); auth != "" {
		h := sha256.Sum256([]byte(auth))
		authHash = hex.EncodeToString(h[:8]) // 16-char prefix, collision-safe
	}

	components := []string{
		r.Method,
		r.URL.Path,
		r.URL.RawQuery,
		tenantID,
		authHash, // empty string for unauthenticated → different key
	}
	if userID != "" {
		components = append(components, userID)
	}

	key := strings.Join(components, "|")
	
	// Hash if too long
	if len(key) > 200 {
		hash := sha256.Sum256([]byte(key))
		key = hex.EncodeToString(hash[:])
	}

	if prefix != "" {
		return fmt.Sprintf("%s:%s", prefix, key)
	}
	return key
}

// InvalidateCache removes cache entries for a collection
func InvalidateCache(ctx context.Context, cache Cache, collectionID string) error {
	pattern := fmt.Sprintf("*%s*", collectionID)
	return cache.DeletePattern(ctx, pattern)
}

// InvalidateUserCache removes cache entries for a user
func InvalidateUserCache(ctx context.Context, cache Cache, userID string) error {
	pattern := fmt.Sprintf("*user*%s*", userID)
	return cache.DeletePattern(ctx, pattern)
}

// CacheKey helpers for specific resources
func CollectionKey(collectionID string) string {
	return fmt.Sprintf("collection:%s", collectionID)
}

func CollectionsListKey(tenantID string, page, perPage int) string {
	return fmt.Sprintf("collections:%s:%d:%d", tenantID, page, perPage)
}

func RecordKey(collectionID, recordID string) string {
	return fmt.Sprintf("record:%s:%s", collectionID, recordID)
}

func RecordsListKey(collectionID string, page, perPage int, filter, sort string) string {
	return fmt.Sprintf("records:%s:%d:%d:%s:%s", collectionID, page, perPage, filter, sort)
}

func UserKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

func UsersListKey(tenantID string, page, perPage int) string {
	return fmt.Sprintf("users:%s:%d:%d", tenantID, page, perPage)
}

// CachedHandler wraps a handler with caching logic
type CachedHandler struct {
	Cache      Cache
	TTL        time.Duration
	KeyFunc    func(r *http.Request) string
	Invalidate func(r *http.Request) []string
}

// Handler wraps an http.Handler with caching
func (ch *CachedHandler) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		key := ch.KeyFunc(r)

		// Try cache
		cached, err := ch.Cache.Get(ctx, key)
		if err == nil {
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cached))
			return
		}

		// Execute handler
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           []byte{},
		}
		next.ServeHTTP(wrapped, r)

		// Store in cache
		if wrapped.statusCode == http.StatusOK {
			ch.Cache.Set(ctx, key, string(wrapped.body), ch.TTL)
			w.Header().Set("X-Cache", "MISS")
		}
	})
}

// SetupCacheMiddleware configures caching for API routes
func SetupCacheMiddleware(r chi.Router, cache Cache) {
	config := MiddlewareConfig{
		Cache:      cache,
		DefaultTTL: 5 * time.Minute,
		KeyPrefix:  "api",
		ExcludePaths: []string{
			"/health",
			"/api/auth",
			"/ws",
		},
	}

	r.Use(Middleware(config))
}

// CacheControl sets cache control headers
func CacheControl(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds())))
			}
			next.ServeHTTP(w, r)
		})
	}
}
