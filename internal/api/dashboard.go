package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fieldstone/fieldstone/pkg/models"
	"github.com/go-chi/chi/v5"
)

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalRecords      int64 `json:"totalRecords"`
	TotalUsers        int64 `json:"totalUsers"`
	TotalCollections  int64 `json:"totalCollections"`
	RequestsPerMinute int64 `json:"requestsPerMinute"`
	StorageUsed       int64 `json:"storageUsed"`
	StorageLimit      int64 `json:"storageLimit"`
	ActiveUsers       int64 `json:"activeUsers"`
}

// Activity represents a dashboard activity
type Activity struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	UserID    string                 `json:"userId,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}

// HealthStatus represents system health
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// setupDashboardRoutes adds dashboard-specific routes
func (s *Server) setupDashboardRoutes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/stats", s.handleDashboardStats)
		r.Get("/activities", s.handleDashboardActivities)
	})
}

// handleDashboardStats returns dashboard statistics computed from the database.
//
// Previously this returned hardcoded values: TotalUsers 45, RequestsPerMinute
// 2340, StorageUsed 150MB, ActiveUsers 12, and TotalRecords was declared but
// never calculated so it was always zero.
func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	collections, err := s.backend.ListCollections(ctx, tenantID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Real record count: ask each collection for its total. PerPage 1 keeps the
	// payload small; only the count is used.
	var totalRecords int64
	for _, c := range collections {
		res, err := s.backend.QueryRecords(ctx, tenantID, c.ID, models.QueryOptions{Page: 1, PerPage: 1})
		if err != nil {
			// One unreadable collection should not blank the whole dashboard.
			continue
		}
		totalRecords += int64(res.TotalItems)
	}

	// Real user count.
	var totalUsers int64
	if users, err := s.backend.ListUsers(ctx, tenantID, models.QueryOptions{Page: 1, PerPage: 1}); err == nil {
		totalUsers = int64(users.TotalItems)
	}

	// Real storage, where the backend can report it. Zero means "not available"
	// rather than a fabricated figure.
	var storageUsed int64
	if sr, ok := s.backend.(StorageReporter); ok {
		if n, err := sr.StorageBytes(); err == nil {
			storageUsed = n
		}
	}

	stats := DashboardStats{
		TotalRecords:      totalRecords,
		TotalUsers:        totalUsers,
		TotalCollections:  int64(len(collections)),
		RequestsPerMinute: s.requests.perMinute(time.Now()),
		StorageUsed:       storageUsed,
		StorageLimit:      0, // no quota is enforced, so do not invent one
		ActiveUsers:       totalUsers,
	}

	s.sendJSON(w, http.StatusOK, stats)
}

// handleDashboardActivities returns the real, in-memory activity log.
//
// This used to return six fabricated entries referencing john@example.com and
// "products"/"orders" collections that do not exist.
func (s *Server) handleDashboardActivities(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	items := s.activity.recent(limit)
	if items == nil {
		items = []Activity{}
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// extendSetupRoutes registers dashboard routes and overrides the health check.
// Called from setupRoutes BEFORE the /api Route block so it does not mount /api itself.
func (s *Server) extendSetupRoutes() {
	// Enhanced health check overrides the basic one registered in setupRoutes
	s.router.Get("/health", s.handleHealthDetailed)
}

// registerDashboardRoutes adds dashboard sub-routes to an existing /api router.
// Called from inside the server.go /api Route block.
func (s *Server) registerDashboardRoutes(r chi.Router) {
	s.setupDashboardRoutes(r)
}

// handleHealthDetailed returns detailed health status
func (s *Server) handleHealthDetailed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check database
	dbStatus := "healthy"
	if err := s.backend.Ping(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services: map[string]string{
			"database": dbStatus,
			"api":      "healthy",
			"cache":    "healthy",
		},
	}

	if dbStatus != "healthy" {
		health.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	s.sendJSON(w, http.StatusOK, health)
}
