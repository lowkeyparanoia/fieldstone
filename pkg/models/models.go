// Package models defines the core data structures for Fieldstone.
// These types represent collections, records, users, and tenant configurations.
package models

import (
	"encoding/json"
	"time"
)

// FieldType represents the data type of a collection field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeBool     FieldType = "bool"
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
	FieldTypeDate     FieldType = "date"
	FieldTypeJSON     FieldType = "json"
	FieldTypeRelation FieldType = "relation"
	FieldTypeFile     FieldType = "file"
)

// FieldOption represents validation options for a field
type FieldOption struct {
	MinLength    *int              `json:"minLength,omitempty"`
	MaxLength    *int              `json:"maxLength,omitempty"`
	Min          *float64          `json:"min,omitempty"`
	Max          *float64          `json:"max,omitempty"`
	Pattern      *string           `json:"pattern,omitempty"`
	Values       []string          `json:"values,omitempty"`
	Required     bool              `json:"required"`
	Unique       bool              `json:"unique"`
	DefaultValue interface{}       `json:"defaultValue,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
}

// Field defines a single field in a collection schema
type Field struct {
	Name        string       `json:"name"`
	Type        FieldType    `json:"type"`
	Options     FieldOption  `json:"options,omitempty"`
	Description string       `json:"description,omitempty"`
}

// Collection represents a database table/collection definition
type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TenantID    string    `json:"tenantId,omitempty"`
	Fields      []Field   `json:"fields"`
	System      bool      `json:"system"`      // System collections cannot be deleted
	ListRule    *string   `json:"listRule,omitempty"`   // Filter rule for listing
	ViewRule    *string   `json:"viewRule,omitempty"`   // Filter rule for viewing single
	CreateRule  *string   `json:"createRule,omitempty"` // Filter rule for creating
	UpdateRule  *string   `json:"updateRule,omitempty"` // Filter rule for updating
	DeleteRule  *string   `json:"deleteRule,omitempty"` // Filter rule for deleting
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Record represents a single record/row in a collection
type Record struct {
	ID           string          `json:"id"`
	CollectionID string          `json:"collectionId"`
	TenantID     string          `json:"tenantId,omitempty"`
	Data         json.RawMessage `json:"data"` // Dynamic data based on collection schema
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

// User represents an authenticated user
type User struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId,omitempty"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"` // Never serialized
	Verified       bool      `json:"verified"`
	TokenKey       string    `json:"tokenKey"` // For JWT rotation
	LastResetAt    time.Time `json:"lastResetAt,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Tenant represents an isolated organization/workspace
type Tenant struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"` // URL-friendly identifier
	Domain      *string         `json:"domain,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
	Isolation   IsolationLevel  `json:"isolation"`
	Active      bool            `json:"active"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// IsolationLevel defines how tenant data is separated
type IsolationLevel string

const (
	IsolationRowLevel   IsolationLevel = "row"     // tenant_id column
	IsolationSchema     IsolationLevel = "schema"  // separate schemas
	IsolationDatabase   IsolationLevel = "database" // separate databases
)

// QueryOptions represents filtering and pagination options
type QueryOptions struct {
	Filter   string   // Filter DSL expression
	Sort     string   // Sort field and direction
	Page     int      // Page number (1-based)
	PerPage  int      // Items per page
	Expand   []string // Relations to expand
	Fields   []string // Fields to select
}

// QueryResult wraps a paginated query result
type QueryResult struct {
	Items      []Record `json:"items"`
	TotalItems int      `json:"totalItems"`
	Page       int      `json:"page"`
	PerPage    int      `json:"perPage"`
	TotalPages int      `json:"totalPages"`
}

// AuthMethod represents supported authentication methods
type AuthMethod string

const (
	AuthMethodEmail    AuthMethod = "email"
	AuthMethodOAuth    AuthMethod = "oauth"
	AuthMethodOTP      AuthMethod = "otp"
	AuthMethodLDAP     AuthMethod = "ldap"
)

// OAuthProvider configuration
type OAuthProvider struct {
	Name         string `json:"name"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AuthURL      string `json:"authUrl"`
	TokenURL     string `json:"tokenUrl"`
	UserInfoURL  string `json:"userInfoUrl"`
	Scopes       []string `json:"scopes"`
}

// FilterOperator represents valid filter operators
type FilterOperator string

const (
	OpEqual        FilterOperator = "="
	OpNotEqual     FilterOperator = "!="
	OpGreater      FilterOperator = ">"
	OpGreaterEqual FilterOperator = ">="
	OpLess         FilterOperator = "<"
	OpLessEqual    FilterOperator = "<="
	OpLike         FilterOperator = "~"
	OpNotLike      FilterOperator = "!~"
	OpContains     FilterOperator = "?="
	OpIn           FilterOperator = "?="
	OpAny          FilterOperator = "?="
)

// FilterNode represents a parsed filter expression node
type FilterNode struct {
	Operator FilterOperator `json:"op,omitempty"`
	Field    string         `json:"field,omitempty"`
	Value    interface{}    `json:"value,omitempty"`
	And      []*FilterNode  `json:"and,omitempty"`
	Or       []*FilterNode  `json:"or,omitempty"`
}
