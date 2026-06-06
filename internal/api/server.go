// Package api provides HTTP handlers and middleware for the Fieldstone API.
// It implements a RESTful API following PocketBase conventions.
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/internal/cache"
	"github.com/fieldstone/fieldstone/internal/jobs"
	"github.com/fieldstone/fieldstone/pkg/models"
)

// Server holds all API dependencies
type Server struct {
	router  *chi.Mux
	backend backend.Backend
	auth    *auth.Service
	jobs    *jobs.Queue
	cache   cache.Cache
	logger  zerolog.Logger
}

// NewServer creates a new API server
func NewServer(be backend.Backend, authService *auth.Service, jobQueue *jobs.Queue, cache cache.Cache) *Server {
	s := &Server{
		router:  chi.NewRouter(),
		backend: be,
		auth:    authService,
		jobs:    jobQueue,
		cache:   cache,
		logger:  log.Logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Request ID
	s.router.Use(middleware.RequestID)

	// Real IP
	s.router.Use(middleware.RealIP)

	// Logger
	s.router.Use(middleware.Logger)

	// Recoverer
	s.router.Use(middleware.Recoverer)

	// Timeout
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Cache middleware for API routes
	if s.cache != nil {
		cacheMiddleware := cache.Middleware(cache.MiddlewareConfig{
			Cache:         s.cache,
			DefaultTTL:    5 * time.Minute,
			KeyPrefix:     "api",
			ExcludePaths:  []string{"/health", "/api/auth", "/ws"},
			ExcludeMethods: []string{"POST", "PUT", "DELETE", "PATCH"},
		})
		s.router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/api/") {
					cacheMiddleware(next).ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		})
		log.Info().Msg("Cache middleware enabled")
	}

	// Content-Type JSON for API routes
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	})
}

func (s *Server) setupRoutes() {
	// Health check (detailed version will override this)
	s.router.Get("/health", s.handleHealth)
	
	// Setup dashboard routes
	s.extendSetupRoutes()

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/refresh", s.handleRefresh)
			r.Post("/logout", s.handleLogout)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware)

			// Collections
			r.Route("/collections", func(r chi.Router) {
				r.Get("/", s.handleListCollections)
				r.Post("/", s.handleCreateCollection)
				r.Get("/{id}", s.handleGetCollection)
				r.Put("/{id}", s.handleUpdateCollection)
				r.Delete("/{id}", s.handleDeleteCollection)

				// Records
				r.Get("/{id}/records", s.handleListRecords)
				r.Post("/{id}/records", s.handleCreateRecord)
				r.Get("/{id}/records/{recordId}", s.handleGetRecord)
				r.Put("/{id}/records/{recordId}", s.handleUpdateRecord)
				r.Delete("/{id}/records/{recordId}", s.handleDeleteRecord)
			})

			// Users
			r.Route("/users", func(r chi.Router) {
				r.Get("/", s.handleListUsers)
				r.Get("/{id}", s.handleGetUser)
				r.Put("/{id}", s.handleUpdateUser)
				r.Delete("/{id}", s.handleDeleteUser)
			})

			// Tenants (admin only)
			r.Route("/tenants", func(r chi.Router) {
				r.Get("/", s.handleListTenants)
				r.Post("/", s.handleCreateTenant)
				r.Get("/{id}", s.handleGetTenant)
				r.Put("/{id}", s.handleUpdateTenant)
				r.Delete("/{id}", s.handleDeleteTenant)
			})

			// Dashboard stats
			s.registerDashboardRoutes(r)
		})
	})

	// Admin UI — serve React SPA from web/admin/dist/
	// Resolves path relative to binary location or working directory.
	adminDist := findAdminDist()
	if adminDist != "" {
		fileServer := http.FileServer(http.Dir(adminDist))
		// Serve /admin/* — strip prefix, fall back to index.html for SPA routing
		s.router.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
		})
		s.router.Get("/admin/*", func(w http.ResponseWriter, r *http.Request) {
			stripped := strings.TrimPrefix(r.URL.Path, "/admin")
			if stripped == "" || stripped == "/" {
				http.ServeFile(w, r, filepath.Join(adminDist, "index.html"))
				return
			}
			// Check if file exists; if not, serve index.html (SPA fallback)
			fpath := filepath.Join(adminDist, stripped)
			if _, err := os.Stat(fpath); os.IsNotExist(err) {
				http.ServeFile(w, r, filepath.Join(adminDist, "index.html"))
				return
			}
			http.StripPrefix("/admin", fileServer).ServeHTTP(w, r)
		})
	}

	// Serve admin assets at /assets/* and /favicon.svg (React build uses absolute paths)
	if adminDist != "" {
		s.router.Get("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(adminDist, "favicon.svg"))
		})
		s.router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(http.Dir(adminDist)).ServeHTTP(w, r)
		})
	}

	// Root landing page
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Fieldstone</title></head><body>
<h1>Fieldstone BaaS</h1>
<p><a href="/admin/">Admin UI</a> &nbsp;|&nbsp; <a href="/health">Health</a> &nbsp;|&nbsp; <a href="/api/collections">Collections API</a></p>
</body></html>`))
	})
}

// findAdminDist locates the admin dist folder relative to cwd or binary.
func findAdminDist() string {
	candidates := []string{
		"web/admin/dist",
		"../web/admin/dist",
		filepath.Join(filepath.Dir(os.Args[0]), "web/admin/dist"),
		filepath.Join(filepath.Dir(os.Args[0]), "../web/admin/dist"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

// Router returns the chi router for testing
func (s *Server) Router() *chi.Mux {
	return s.router
}

// authMiddleware validates JWT tokens and adds auth context
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.sendError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			s.sendError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		claims, err := s.auth.ValidateToken(parts[1])
		if err != nil {
			s.sendError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Add auth context to request
		authCtx := &auth.Context{
			UserID:   claims.UserID,
			TenantID: claims.TenantID,
			Email:    claims.Email,
		}
		ctx := auth.WithContext(r.Context(), authCtx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.backend.Ping(ctx); err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Helper methods
func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}

func (s *Server) getTenantID(r *http.Request) string {
	// Auth context (JWT) takes priority — set by authMiddleware after token validation
	if authCtx, ok := auth.FromContext(r.Context()); ok {
		return authCtx.TenantID
	}
	// Public endpoints (register, login) — honour explicit X-Tenant-ID header
	if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
		return tid
	}
	return "default"
}

func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	collections, err := s.backend.ListCollections(ctx, tenantID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if collections == nil {
		collections = []models.Collection{}
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"items": collections,
	})
}

func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	id := chi.URLParam(r, "id")

	collection, err := s.backend.GetCollection(ctx, tenantID, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, collection)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Mock users for now
	users := []map[string]interface{}{
		{
			"id":        "1",
			"email":     "admin@fieldstone.io",
			"verified":  true,
			"createdAt": "2024-01-01T00:00:00Z",
			"updatedAt": "2024-01-01T00:00:00Z",
		},
		{
			"id":          "2",
			"email":       "john@example.com",
			"verified":    true,
			"createdAt":   "2024-01-10T10:00:00Z",
			"updatedAt":   "2024-01-15T08:30:00Z",
			"lastLoginAt": "2024-01-15T08:30:00Z",
		},
		{
			"id":        "3",
			"email":     "jane@example.com",
			"verified":  false,
			"createdAt": "2024-01-12T14:00:00Z",
			"updatedAt": "2024-01-12T14:00:00Z",
		},
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"items": users,
	})
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// Mock user
	user := map[string]interface{}{
		"id":        id,
		"email":     "user@example.com",
		"verified":  true,
		"createdAt": "2024-01-01T00:00:00Z",
		"updatedAt": "2024-01-15T00:00:00Z",
	}

	s.sendJSON(w, http.StatusOK, user)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	// id := chi.URLParam(r, "id")

	// In production, actually delete the user
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}
