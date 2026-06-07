package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// OTPRequest represents an OTP send request
type OTPRequest struct {
	Phone string `json:"phone"`
}

// OTPVerifyRequest represents an OTP verify request
type OTPVerifyRequest struct {
	Phone string `json:"phone"`
	Token string `json:"token"`
}

// MagicLinkRequest represents a magic link request
type MagicLinkRequest struct {
	Email string `json:"email"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation
type PasswordResetConfirmRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// handleSendOTP sends an OTP via SMS (pluggable hook).
func (s *Server) handleSendOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req OTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		s.sendError(w, http.StatusBadRequest, "phone required")
		return
	}

	// Find or create user by phone (stored in metadata for now)
	// In production: query _users by phone index or dedicated _auth_identities
	code, err := s.auth.GenerateOTP()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate otp")
		return
	}

	// Store OTP in memory or job queue for SMS delivery
	// For MVP, enqueue an SMS job if jobs queue supports it
	s.jobs.Enqueue(ctx, "sms:send", "default", map[string]interface{}{
		"phone": req.Phone,
		"code":  code,
	})

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"message": "OTP sent",
	})
}

// handleVerifyOTP verifies an OTP and returns a session.
func (s *Server) handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req OTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.Token == "" {
		s.sendError(w, http.StatusBadRequest, "phone and token required")
		return
	}

	// In production: verify code against stored hash and expiry
	// For MVP, we accept any 6-digit code as a stub
	if len(req.Token) != 6 {
		s.sendError(w, http.StatusUnauthorized, "invalid otp")
		return
	}

	// Generate a user id based on phone hash (stub)
	userID := uuid.New().String()
	tokens, err := s.auth.GenerateTokenPair(userID, "default", req.Phone, "otp-key", "authenticated", nil, nil)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokens.AccessToken,
		"user": map[string]interface{}{
			"id":    userID,
			"phone": req.Phone,
		},
	})
}

// handleMagicLink sends a magic link email.
func (s *Server) handleMagicLink(w http.ResponseWriter, r *http.Request) {
	var req MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		s.sendError(w, http.StatusBadRequest, "email required")
		return
	}

	// Generate magic link token
	token, _ := s.auth.GenerateTokenKey()
	// Store token with expiry (in production, persist to _auth_magic_links table)
	_ = token

	s.jobs.Enqueue(r.Context(), "email:send", "default", map[string]interface{}{
		"to":      req.Email,
		"subject": "Your magic link",
		"body":    fmt.Sprintf("Click to sign in: https://your-app.com/auth/callback?token=%s", token),
	})

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Magic link sent",
	})
}

// handlePasswordReset initiates a password reset.
func (s *Server) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		s.sendError(w, http.StatusBadRequest, "email required")
		return
	}

	token, _ := s.auth.GenerateTokenKey()
	_ = token

	s.jobs.Enqueue(r.Context(), "email:send", "default", map[string]interface{}{
		"to":      req.Email,
		"subject": "Password reset",
		"body":    fmt.Sprintf("Reset token: %s", token),
	})

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Password reset email sent",
	})
}

// handleOAuthRedirect redirects to the OAuth provider.
func (s *Server) handleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	// In production: lookup provider config, build auth URL, redirect
	http.Redirect(w, r, fmt.Sprintf("/api/auth/oauth/callback?provider=%s", provider), http.StatusTemporaryRedirect)
}

// handleOAuthCallback handles the OAuth callback.
func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	code := r.URL.Query().Get("code")

	// In production: exchange code for token, fetch user info, upsert user
	_ = code
	userID := uuid.New().String()
	tokens, err := s.auth.GenerateTokenPair(userID, "default", "oauth@example.com", "oauth-key", "authenticated", map[string]interface{}{"provider": provider}, nil)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokens.AccessToken,
		"user": map[string]interface{}{
			"id":    userID,
			"email": "oauth@example.com",
		},
	})
}
