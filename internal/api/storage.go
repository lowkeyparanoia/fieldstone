package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleCreateBucket(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	var req struct {
		Name   string `json:"name"`
		Public bool   `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := s.storage.CreateBucket(r.Context(), req.Name, req.Public); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusCreated, map[string]interface{}{"name": req.Name})
}

func (s *Server) handleListBuckets(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	buckets, err := s.storage.ListBuckets(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"buckets": buckets})
}

func (s *Server) handleDeleteBucket(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	if err := s.storage.DeleteBucket(r.Context(), bucket); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *Server) handleUploadObject(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	path := r.URL.Query().Get("path")
	if path == "" {
		s.sendError(w, http.StatusBadRequest, "path query param required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	obj, err := s.storage.Upload(r.Context(), bucket, path, r.Body, r.ContentLength, contentType)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusCreated, obj)
}

func (s *Server) handleDownloadObject(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	path := chi.URLParam(r, "path")
	path = strings.TrimPrefix(path, "/")

	reader, err := s.storage.Download(r.Context(), bucket, path)
	if err != nil {
		s.sendError(w, http.StatusNotFound, err.Error())
		return
	}
	defer reader.Close()
	io.Copy(w, reader)
}

func (s *Server) handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	path := chi.URLParam(r, "path")
	path = strings.TrimPrefix(path, "/")

	if err := s.storage.Delete(r.Context(), bucket, path); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *Server) handleListObjects(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	prefix := r.URL.Query().Get("prefix")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
		limit = n
	}
	objects, err := s.storage.List(r.Context(), bucket, prefix, limit)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"objects": objects})
}

func (s *Server) handleStoragePublicDownload(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.sendError(w, http.StatusNotImplemented, "storage not configured")
		return
	}
	bucket := chi.URLParam(r, "bucket")
	path := chi.URLParam(r, "path")
	path = strings.TrimPrefix(path, "/")

	// Validate signed URL params if present
	// In production: verify HMAC signature
	reader, err := s.storage.Download(r.Context(), bucket, path)
	if err != nil {
		s.sendError(w, http.StatusNotFound, err.Error())
		return
	}
	defer reader.Close()
	io.Copy(w, reader)
}
