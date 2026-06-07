package api

import (
	"encoding/json"
	"net/http"

	"github.com/fieldstone/fieldstone/internal/webhooks"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	if s.webhookManager == nil {
		s.sendError(w, http.StatusNotImplemented, "webhooks not configured")
		return
	}
	var req webhooks.Webhook
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.ID = uuid.New().String()
	if err := s.webhookManager.Store().Create(r.Context(), &req); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusCreated, req)
}

func (s *Server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	if s.webhookManager == nil {
		s.sendError(w, http.StatusNotImplemented, "webhooks not configured")
		return
	}
	items, err := s.webhookManager.Store().List(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (s *Server) handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	if s.webhookManager == nil {
		s.sendError(w, http.StatusNotImplemented, "webhooks not configured")
		return
	}
	id := chi.URLParam(r, "id")
	item, err := s.webhookManager.Store().Get(r.Context(), id)
	if err != nil {
		s.sendError(w, http.StatusNotFound, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	if s.webhookManager == nil {
		s.sendError(w, http.StatusNotImplemented, "webhooks not configured")
		return
	}
	id := chi.URLParam(r, "id")
	var req webhooks.Webhook
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid body")
		return
	}
	req.ID = id
	if err := s.webhookManager.Store().Update(r.Context(), &req); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, req)
}

func (s *Server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if s.webhookManager == nil {
		s.sendError(w, http.StatusNotImplemented, "webhooks not configured")
		return
	}
	id := chi.URLParam(r, "id")
	if err := s.webhookManager.Store().Delete(r.Context(), id); err != nil {
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
