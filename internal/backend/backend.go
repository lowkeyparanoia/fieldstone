// Package backend defines the storage backend interface and common utilities.
// This abstraction allows Fieldstone to work with SQLite, PostgreSQL, or other databases.
package backend

import (
	"context"
	"fmt"

	"github.com/fieldstone/fieldstone/pkg/models"
)

// ListOptions provides pagination and filtering for list queries
type ListOptions struct {
	Page    int    // Page number (1-based)
	PerPage int    // Items per page
	Filter  string // Filter query string
	Sort    string // Sort field and direction (e.g., "-created_at")
}

// Backend defines the interface that all storage implementations must satisfy.
// This abstraction enables switching between SQLite (local dev) and PostgreSQL (production)
// without changing application code.
type Backend interface {
	// Connection management
	Ping(ctx context.Context) error
	Close() error

	// Collection operations
	CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) error
	UpdateCollection(ctx context.Context, tenantID string, collection *models.Collection) error
	DeleteCollection(ctx context.Context, tenantID string, collectionID string) error
	GetCollection(ctx context.Context, tenantID string, collectionID string) (*models.Collection, error)
	GetCollectionByName(ctx context.Context, tenantID string, name string) (*models.Collection, error)
	ListCollections(ctx context.Context, tenantID string) ([]models.Collection, error)

	// Record operations (tenant-scoped)
	CreateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error
	UpdateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error
	DeleteRecord(ctx context.Context, tenantID string, collectionID string, recordID string) error
	GetRecord(ctx context.Context, tenantID string, collectionID string, recordID string) (*models.Record, error)
	ListRecords(ctx context.Context, tenantID string, collectionID string, opts ListOptions) ([]models.Record, int, error)
	QueryRecords(ctx context.Context, tenantID string, collectionID string, opts models.QueryOptions) (*models.QueryResult, error)

	// User operations
	CreateUser(ctx context.Context, tenantID string, user *models.User) error
	UpdateUser(ctx context.Context, tenantID string, user *models.User) error
	DeleteUser(ctx context.Context, tenantID string, userID string) error
	GetUser(ctx context.Context, tenantID string, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, tenantID string, email string) (*models.User, error)
	GetUserByTokenKey(ctx context.Context, tokenKey string) (*models.User, error)
	ListUsers(ctx context.Context, tenantID string, opts models.QueryOptions) (*models.QueryResult, error)

	// Authentication
	UpdateUserPassword(ctx context.Context, tenantID string, userID string, hashedPassword string) error
	UpdateUserTokenKey(ctx context.Context, tenantID string, userID string, tokenKey string) error
	VerifyUser(ctx context.Context, tenantID string, userID string) error

	// Tenant operations
	CreateTenant(ctx context.Context, tenant *models.Tenant) error
	UpdateTenant(ctx context.Context, tenant *models.Tenant) error
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	ListTenants(ctx context.Context) ([]models.Tenant, error)

	// Migration support
	ExecMigration(ctx context.Context, version int, up bool, sql string) error
	GetMigrationVersion(ctx context.Context) (int, error)

	// Transaction support
	BeginTx(ctx context.Context) (Tx, error)

	// Realtime support
	Subscribe(ctx context.Context, collectionID string, callback func(event Event)) (Subscription, error)
}

// Tx represents a database transaction
type Tx interface {
	Commit() error
	Rollback() error
	// Transaction-scoped operations would go here
}

// Event represents a database change event for realtime subscriptions
type Event struct {
	Action       string          // "create", "update", "delete"
	CollectionID string
	RecordID     string
	Data         interface{}
	Timestamp    int64
}

// Subscription represents an active realtime subscription
type Subscription interface {
	Unsubscribe() error
}

// Config holds backend configuration
type Config struct {
	Type     string // "sqlite", "postgres", "memory"
	DSN      string // Connection string
	MaxConns int
	MinConns int
}

// New creates a new backend based on configuration
func New(cfg Config) (Backend, error) {
	switch cfg.Type {
	case "sqlite":
		return NewSQLiteBackend(cfg.DSN)
	case "postgres":
		return NewPostgresBackend(cfg.DSN, cfg.MaxConns, cfg.MinConns)
	case "memory":
		return NewMemoryBackend(), nil
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", cfg.Type)
	}
}

// BackendError represents a backend-specific error
type BackendError struct {
	Code    string
	Message string
	Cause   error
}

func (e *BackendError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BackendError) Unwrap() error {
	return e.Cause
}

// Common error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeAlreadyExist = "ALREADY_EXISTS"
	ErrCodeInvalidData  = "INVALID_DATA"
	ErrCodeConstraint   = "CONSTRAINT_VIOLATION"
	ErrCodeAuth         = "AUTH_FAILED"
	ErrCodePermission   = "PERMISSION_DENIED"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

// IsNotFound checks if error is a not-found error
func IsNotFound(err error) bool {
	if be, ok := err.(*BackendError); ok {
		return be.Code == ErrCodeNotFound
	}
	return false
}

// IsAlreadyExists checks if error is an already-exists error
func IsAlreadyExists(err error) bool {
	if be, ok := err.(*BackendError); ok {
		return be.Code == ErrCodeAlreadyExist
	}
	return false
}

// NewNotFoundError creates a not-found error
func NewNotFoundError(resource string, id string) error {
	return &BackendError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s with id %s not found", resource, id),
	}
}

// NewAlreadyExistsError creates an already-exists error
func NewAlreadyExistsError(resource string, field string, value string) error {
	return &BackendError{
		Code:    ErrCodeAlreadyExist,
		Message: fmt.Sprintf("%s with %s '%s' already exists", resource, field, value),
	}
}
