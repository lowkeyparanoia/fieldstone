// Package auth provides authentication and authorization for Fieldstone.
// It supports multiple auth methods: email/password, OAuth2, and OTP.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	secret      []byte
	tokenExpiry time.Duration
	issuer      string
}

// NewService creates a new auth service
func NewService(secret string, tokenExpiry time.Duration, issuer string) *Service {
	if tokenExpiry == 0 {
		tokenExpiry = 24 * time.Hour
	}
	if issuer == "" {
		issuer = "fieldstone"
	}
	return &Service{
		secret:      []byte(secret),
		tokenExpiry: tokenExpiry,
		issuer:      issuer,
	}
}

// Claims represents JWT claims (Supabase-compatible)
type Claims struct {
	UserID      string                 `json:"userId"`
	TenantID    string                 `json:"tenantId"`
	Email       string                 `json:"email"`
	TokenKey    string                 `json:"tokenKey"`
	Role        string                 `json:"role"`
	AppMetadata map[string]interface{} `json:"app_metadata,omitempty"`
	UserMetadata map[string]interface{} `json:"user_metadata,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// GenerateTokenPair creates new access and refresh tokens (Supabase-compatible claims)
func (s *Service) GenerateTokenPair(userID, tenantID, email, tokenKey, role string, appMeta, userMeta map[string]interface{}) (*TokenPair, error) {
	now := time.Now()
	expiry := now.Add(s.tokenExpiry)

	if appMeta == nil {
		appMeta = make(map[string]interface{})
	}
	if userMeta == nil {
		userMeta = make(map[string]interface{})
	}
	appMeta["provider"] = "email"
	appMeta["providers"] = []string{"email"}

	accessClaims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Email:        email,
		TokenKey:     tokenKey,
		Role:         role,
		AppMetadata:  appMeta,
		UserMetadata: userMeta,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
		ExpiresAt:    expiry,
	}, nil
}

// ValidateToken validates and parses a JWT token
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *Service) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateTokenKey generates a new token key for JWT rotation
func (s *Service) GenerateTokenKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateOTP generates a 6-digit OTP code
func (s *Service) GenerateOTP() (string, error) {
	// Generate a 6-digit code
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	num := int(bytes[0])<<24 | int(bytes[1])<<16 | int(bytes[2])<<8 | int(bytes[3])
	code := fmt.Sprintf("%06d", num%1000000)
	return code, nil
}

// AuthContextKey is the key for storing auth info in context
type AuthContextKey struct{}

// Context holds authentication context
type Context struct {
	UserID    string
	TenantID  string
	Email     string
	Role      string
	RawClaims map[string]interface{}
}

// WithContext adds auth context to a context
func WithContext(ctx context.Context, authCtx *Context) context.Context {
	return context.WithValue(ctx, AuthContextKey{}, authCtx)
}

// FromContext extracts auth context from a context
func FromContext(ctx context.Context) (*Context, bool) {
	authCtx, ok := ctx.Value(AuthContextKey{}).(*Context)
	return authCtx, ok
}
