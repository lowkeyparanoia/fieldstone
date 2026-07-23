package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/pkg/models"
	"github.com/google/uuid"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token  string      `json:"token"`
	Record models.User `json:"record"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		s.sendError(w, http.StatusBadRequest, "email and password required")
		return
	}

	// Hash password
	hash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Generate token key
	tokenKey, err := s.auth.GenerateTokenKey()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate token key")
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: hash,
		Verified:     false,
		TokenKey:     tokenKey,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.backend.CreateUser(ctx, tenantID, user); err != nil {
		if backend.IsAlreadyExists(err) {
			s.sendError(w, http.StatusConflict, "user already exists")
			return
		}
		s.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.activity.record(Activity{
		Type:    "user_created",
		Message: "New user registered: " + user.Email,
		UserID:  user.ID,
	})

	// Generate tokens
	tokens, err := s.auth.GenerateTokenPair(user.ID, tenantID, user.Email, user.TokenKey)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	// Return user without password hash
	user.PasswordHash = ""
	s.sendJSON(w, http.StatusCreated, map[string]interface{}{
		"token":  tokens.AccessToken,
		"record": user, "user": user,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := s.getTenantID(r)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user by email
	user, err := s.backend.GetUserByEmail(ctx, tenantID, req.Email)
	if err != nil {
		s.sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Verify password
	if !s.auth.VerifyPassword(req.Password, user.PasswordHash) {
		s.sendError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Rotate token key for security
	newTokenKey, err := s.auth.GenerateTokenKey()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to rotate token")
		return
	}
	user.TokenKey = newTokenKey
	if err := s.backend.UpdateUserTokenKey(ctx, tenantID, user.ID, newTokenKey); err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to update token key")
		return
	}

	// Generate tokens
	tokens, err := s.auth.GenerateTokenPair(user.ID, tenantID, user.Email, user.TokenKey)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	// Return user without password hash
	user.PasswordHash = ""
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"token":  tokens.AccessToken,
		"record": user, "user": user,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
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
	tokenStr := parts[1]

	// Validate current token (allow expired — refresh grace period)
	claims, err := s.auth.ValidateToken(tokenStr)
	if err != nil {
		s.sendError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
		return
	}

	ctx := r.Context()
	tenantID := claims.TenantID

	// Fetch user to verify tokenKey (rotation invalidation check)
	user, err := s.backend.GetUser(ctx, tenantID, claims.UserID)
	if err != nil {
		s.sendError(w, http.StatusUnauthorized, "user not found")
		return
	}

	// TokenKey mismatch means token was already rotated (logout invalidates key)
	if user.TokenKey != claims.TokenKey {
		s.sendError(w, http.StatusUnauthorized, "token has been invalidated — please log in again")
		return
	}

	// Rotate the token key on every refresh, so the presented token is
	// invalidated the moment a new one is issued.
	//
	// Previously the same key was reused and rotation happened only on logout,
	// which meant a stolen token stayed valid for its full lifetime even after
	// the legitimate client had refreshed. Rotating on refresh is what OAuth
	// 2.0 security BCP recommends, and it turns token theft into something
	// detectable: if an old key is ever presented again, either the attacker or
	// the real client is using a token that was already superseded.
	//
	// Known trade-off: refresh is now single use, so two clients sharing an
	// account (two browser tabs, say) can race, and the loser is forced to log
	// in again. The usual mitigation is a short grace window that still accepts
	// the immediately previous key. That is deliberately not implemented here,
	// because it would weaken exactly the property the test asserts: that the
	// old token stops working straight away. Add it only alongside reuse
	// detection.
	newKey, err := s.auth.GenerateTokenKey()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to rotate token key")
		return
	}
	if err := s.backend.UpdateUserTokenKey(ctx, tenantID, user.ID, newKey); err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to rotate token key")
		return
	}
	user.TokenKey = newKey

	tokens, err := s.auth.GenerateTokenPair(user.ID, tenantID, user.Email, newKey)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"token":        tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresAt":    tokens.ExpiresAt,
		"record":       user,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Invalidate token by rotating tokenKey — old JWT becomes unusable
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			if claims, err := s.auth.ValidateToken(parts[1]); err == nil {
				ctx := r.Context()
				if newKey, err := s.auth.GenerateTokenKey(); err == nil {
					// Rotate key — any existing token with old key is now invalid
					_ = s.backend.UpdateUserTokenKey(ctx, claims.TenantID, claims.UserID, newKey)
				}
			}
		}
	}
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "logged out — token invalidated",
	})
}
