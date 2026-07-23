// internal/api/handlers.go - Complete API implementation
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/internal/cache"
	"github.com/fieldstone/fieldstone/pkg/models"
)

// CollectionRequest represents a collection creation/update request
type CollectionRequest struct {
	Name   string         `json:"name"`
	Fields []models.Field `json:"fields"`
}

// RecordRequest represents a record creation/update request
type RecordRequest struct {
	Data map[string]interface{} `json:"data"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"perPage"`
	TotalPages int         `json:"totalPages"`
}

// handleCreateCollection creates a new collection
func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	var req CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Validation
	if req.Name == "" {
		s.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Check for duplicate names
	existing, _ := s.backend.GetCollectionByName(ctx, tenantID, req.Name)
	if existing != nil {
		s.sendError(w, http.StatusConflict, "collection with this name already exists")
		return
	}

	collection := &models.Collection{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      req.Name,
		Fields:    req.Fields,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.backend.CreateCollection(ctx, tenantID, collection); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.activity.record(Activity{
		Type:    "collection_created",
		Message: "Collection created: " + collection.Name,
	})

	// Invalidate collections list cache
	if s.cache != nil {
		if err := s.cache.DeletePattern(ctx, cache.CollectionsListKey(tenantID, 0, 0)+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate collections cache")
		}
	}

	s.sendJSON(w, http.StatusCreated, collection)
}

// handleUpdateCollection updates an existing collection
func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	id := chi.URLParam(r, "id")

	// Check if collection exists
	existing, err := s.backend.GetCollection(ctx, tenantID, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Update fields
	if req.Name != "" && req.Name != existing.Name {
		// Check for name conflict
		conflict, _ := s.backend.GetCollectionByName(ctx, tenantID, req.Name)
		if conflict != nil && conflict.ID != id {
			s.sendError(w, http.StatusConflict, "collection with this name already exists")
			return
		}
		existing.Name = req.Name
	}
	if req.Fields != nil {
		existing.Fields = req.Fields
	}
	existing.UpdatedAt = time.Now()

	if err := s.backend.UpdateCollection(ctx, tenantID, existing); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate collection cache
	if s.cache != nil {
		if err := s.cache.Delete(ctx, cache.CollectionKey(id)); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate collection cache")
		}
		if err := s.cache.DeletePattern(ctx, cache.CollectionsListKey(tenantID, 0, 0)+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate collections list cache")
		}
	}

	s.sendJSON(w, http.StatusOK, existing)
}

// handleDeleteCollection deletes a collection and all its records
func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	id := chi.URLParam(r, "id")

	// Check if collection exists
	_, err := s.backend.GetCollection(ctx, tenantID, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.backend.DeleteCollection(ctx, tenantID, id); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate collection cache and all related records
	if s.cache != nil {
		if err := s.cache.Delete(ctx, cache.CollectionKey(id)); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate collection cache")
		}
		if err := s.cache.DeletePattern(ctx, cache.CollectionsListKey(tenantID, 0, 0)+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate collections list cache")
		}
		// Invalidate all records for this collection
		if err := s.cache.DeletePattern(ctx, "records:"+id+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate records cache")
		}
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "collection deleted",
	})
}

// handleListRecords lists records in a collection with pagination
func (s *Server) handleListRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	collectionID := chi.URLParam(r, "id")

	// Check if collection exists
	_, err := s.backend.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}

	// Parse filters
	filter := r.URL.Query().Get("filter")
	sort := r.URL.Query().Get("sort")

	records, total, err := s.backend.ListRecords(ctx, tenantID, collectionID, backend.ListOptions{
		Page:    page,
		PerPage: perPage,
		Filter:  filter,
		Sort:    sort,
	})
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (total + perPage - 1) / perPage

	s.sendJSON(w, http.StatusOK, ListResponse{
		Items:      records,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// handleCreateRecord creates a new record in a collection
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	collectionID := chi.URLParam(r, "id")

	// Check if collection exists
	collection, err := s.backend.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req RecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Validate data against schema
	if err := validateRecordData(req.Data, collection.Fields); err != nil {
		s.sendError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to encode data: "+err.Error())
		return
	}

	record := &models.Record{
		ID:           uuid.New().String(),
		CollectionID: collectionID,
		TenantID:     tenantID,
		Data:         dataBytes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.backend.CreateRecord(ctx, tenantID, collectionID, record); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.activity.record(Activity{
		Type:    "record_created",
		Message: "Record created in collection " + collectionID,
	})

	// Invalidate records list cache
	if s.cache != nil {
		if err := s.cache.DeletePattern(ctx, cache.RecordsListKey(collectionID, 0, 0, "", "")+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate records list cache")
		}
	}

	s.sendJSON(w, http.StatusCreated, record)
}

// handleGetRecord retrieves a single record
func (s *Server) handleGetRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	collectionID := chi.URLParam(r, "id")
	recordID := chi.URLParam(r, "recordId")

	record, err := s.backend.GetRecord(ctx, tenantID, collectionID, recordID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "record not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, record)
}

// handleUpdateRecord updates an existing record
func (s *Server) handleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	collectionID := chi.URLParam(r, "id")
	recordID := chi.URLParam(r, "recordId")

	// Get collection for schema validation
	collection, err := s.backend.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "collection not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get existing record
	existing, err := s.backend.GetRecord(ctx, tenantID, collectionID, recordID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "record not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req RecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	// Validate data against schema
	if err := validateRecordData(req.Data, collection.Fields); err != nil {
		s.sendError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	dataBytes, err := json.Marshal(req.Data)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to encode data: "+err.Error())
		return
	}

	existing.Data = dataBytes
	existing.UpdatedAt = time.Now()

	if err := s.backend.UpdateRecord(ctx, tenantID, collectionID, existing); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate record cache
	if s.cache != nil {
		if err := s.cache.Delete(ctx, cache.RecordKey(collectionID, recordID)); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate record cache")
		}
		if err := s.cache.DeletePattern(ctx, cache.RecordsListKey(collectionID, 0, 0, "", "")+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate records list cache")
		}
	}

	s.sendJSON(w, http.StatusOK, existing)
}

// handleDeleteRecord deletes a record
func (s *Server) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	collectionID := chi.URLParam(r, "id")
	recordID := chi.URLParam(r, "recordId")

	// Check if record exists
	_, err := s.backend.GetRecord(ctx, tenantID, collectionID, recordID)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "record not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.backend.DeleteRecord(ctx, tenantID, collectionID, recordID); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate record cache
	if s.cache != nil {
		if err := s.cache.Delete(ctx, cache.RecordKey(collectionID, recordID)); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate record cache")
		}
		if err := s.cache.DeletePattern(ctx, cache.RecordsListKey(collectionID, 0, 0, "", "")+"*"); err != nil {
			log.Error().Err(err).Msg("Failed to invalidate records list cache")
		}
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "record deleted",
	})
}

// handleUpdateUser updates a user
func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)
	id := chi.URLParam(r, "id")

	existing, err := s.backend.GetUser(ctx, tenantID, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "user not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Apply updates
	if email, ok := updates["email"].(string); ok && email != "" {
		existing.Email = email
	}
	if verified, ok := updates["verified"].(bool); ok {
		existing.Verified = verified
	}
	existing.UpdatedAt = time.Now()

	if err := s.backend.UpdateUser(ctx, tenantID, existing); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, existing)
}

// handleListTenants lists all tenants
func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenants, err := s.backend.ListTenants(ctx)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"items": tenants,
	})
}

// handleCreateTenant creates a new tenant
func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name     string                 `json:"name"`
		Domain   string                 `json:"domain,omitempty"`
		Settings map[string]interface{} `json:"settings,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Name == "" {
		s.sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	settingsJSON, _ := json.Marshal(req.Settings)
	tenant := &models.Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Domain:    &req.Domain,
		Settings:  settingsJSON,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.backend.CreateTenant(ctx, tenant); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusCreated, tenant)
}

// handleGetTenant retrieves a tenant
func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	tenant, err := s.backend.GetTenant(ctx, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "tenant not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, tenant)
}

// handleUpdateTenant updates a tenant
func (s *Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	existing, err := s.backend.GetTenant(ctx, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "tenant not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		existing.Name = name
	}
	if domain, ok := updates["domain"].(string); ok {
		existing.Domain = &domain
	}
	if settings, ok := updates["settings"].(map[string]interface{}); ok {
		settingsJSON, _ := json.Marshal(settings)
		existing.Settings = settingsJSON
	}

	existing.UpdatedAt = time.Now()

	if err := s.backend.UpdateTenant(ctx, existing); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, existing)
}

// handleDeleteTenant deletes a tenant
func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	_, err := s.backend.GetTenant(ctx, id)
	if err != nil {
		if backend.IsNotFound(err) {
			s.sendError(w, http.StatusNotFound, "tenant not found")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.backend.DeleteTenant(ctx, id); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "tenant deleted",
	})
}

// validateRecordData validates record data against collection fields
func validateRecordData(data map[string]interface{}, fields []models.Field) error {
	for _, field := range fields {
		if field.Options.Required {
			if _, ok := data[field.Name]; !ok {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}
	return nil
}
