package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	Status    string              `json:"status"`
	Timestamp string              `json:"timestamp"`
	Services  map[string]string   `json:"services"`
}

// setupDashboardRoutes adds dashboard-specific routes
func (s *Server) setupDashboardRoutes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/stats", s.handleDashboardStats)
		r.Get("/activities", s.handleDashboardActivities)
	})
}

// handleDashboardStats returns dashboard statistics
func (s *Server) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	// Get collections
	collections, err := s.backend.ListCollections(ctx, tenantID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate total records
	var totalRecords int64

	// Mock stats for now (in production, these would come from metrics)
	stats := DashboardStats{
		TotalRecords:      totalRecords,
		TotalUsers:        45,
		TotalCollections:  int64(len(collections)),
		RequestsPerMinute: 2340,
		StorageUsed:       157286400, // 150MB
		StorageLimit:      10737418240, // 10GB
		ActiveUsers:       12,
	}

	s.sendJSON(w, http.StatusOK, stats)
}

// handleDashboardActivities returns recent activities
func (s *Server) handleDashboardActivities(w http.ResponseWriter, r *http.Request) {
	// Mock activities (in production, these would come from an activity log)
	activities := []Activity{
		{
			ID:        uuid.New().String(),
			Type:      "user_created",
			Message:   "New user registered: john@example.com",
			UserID:    "system",
			CreatedAt: time.Now().Add(-2 * time.Minute),
		},
		{
			ID:        uuid.New().String(),
			Type:      "collection_modified",
			Message:   "Collection schema updated: products",
			UserID:    "admin",
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        uuid.New().String(),
			Type:      "record_created",
			Message:   "New record created in orders collection",
			UserID:    "john@example.com",
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        uuid.New().String(),
			Type:      "backup_completed",
			Message:   "Automatic backup completed successfully",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			Type:      "webhook_triggered",
			Message:   "Webhook delivered: User Created Notification",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			Type:      "record_updated",
			Message:   "Record updated in users collection",
			UserID:    "admin",
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"items": activities,
	})
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
