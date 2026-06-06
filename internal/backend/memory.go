// Package backend provides an in-memory implementation for testing.
// This is useful for unit tests where you don't want to spin up a real database.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fieldstone/fieldstone/pkg/models"
)

// MemoryBackend is an in-memory implementation of Backend for testing
type MemoryBackend struct {
	mu           sync.RWMutex
	collections  map[string]models.Collection
	records      map[string]map[string]models.Record // collectionID -> recordID -> record
	users        map[string]models.User
	tenants      map[string]models.Tenant
	migrations   map[int]bool
}

// NewMemoryBackend creates a new in-memory backend
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		collections: make(map[string]models.Collection),
		records:     make(map[string]map[string]models.Record),
		users:       make(map[string]models.User),
		tenants:     make(map[string]models.Tenant),
		migrations:  make(map[int]bool),
	}
}

func (b *MemoryBackend) Ping(ctx context.Context) error {
	return nil
}

func (b *MemoryBackend) Close() error {
	return nil
}

func (b *MemoryBackend) CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := fmt.Sprintf("%s:%s", tenantID, collection.Name)
	if _, exists := b.collections[key]; exists {
		return NewAlreadyExistsError("collection", "name", collection.Name)
	}

	b.collections[key] = *collection
	b.records[collection.ID] = make(map[string]models.Record)
	return nil
}

func (b *MemoryBackend) UpdateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find and remove old key (name may have changed)
	for k, c := range b.collections {
		if c.ID == collection.ID {
			delete(b.collections, k)
			break
		}
	}
	key := fmt.Sprintf("%s:%s", tenantID, collection.Name)
	b.collections[key] = *collection
	return nil
}

func (b *MemoryBackend) DeleteCollection(ctx context.Context, tenantID string, collectionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find collection by ID
	var key string
	for k, c := range b.collections {
		if c.ID == collectionID {
			key = k
			break
		}
	}
	if key == "" {
		return NewNotFoundError("collection", collectionID)
	}

	delete(b.collections, key)
	delete(b.records, collectionID)
	return nil
}

func (b *MemoryBackend) GetCollection(ctx context.Context, tenantID string, collectionID string) (*models.Collection, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, c := range b.collections {
		if c.ID == collectionID {
			return &c, nil
		}
	}
	return nil, NewNotFoundError("collection", collectionID)
}

func (b *MemoryBackend) GetCollectionByName(ctx context.Context, tenantID string, name string) (*models.Collection, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", tenantID, name)
	if c, ok := b.collections[key]; ok {
		return &c, nil
	}
	return nil, NewNotFoundError("collection", name)
}

func (b *MemoryBackend) ListCollections(ctx context.Context, tenantID string) ([]models.Collection, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []models.Collection
	for key, c := range b.collections {
		if len(key) > len(tenantID)+1 && key[:len(tenantID)] == tenantID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (b *MemoryBackend) CreateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.records[collectionID] == nil {
		b.records[collectionID] = make(map[string]models.Record)
	}
	b.records[collectionID][record.ID] = *record
	return nil
}

func (b *MemoryBackend) UpdateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.records[collectionID] == nil {
		return NewNotFoundError("collection", collectionID)
	}
	b.records[collectionID][record.ID] = *record
	return nil
}

func (b *MemoryBackend) DeleteRecord(ctx context.Context, tenantID string, collectionID string, recordID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.records[collectionID] == nil {
		return NewNotFoundError("collection", collectionID)
	}
	delete(b.records[collectionID], recordID)
	return nil
}

func (b *MemoryBackend) GetRecord(ctx context.Context, tenantID string, collectionID string, recordID string) (*models.Record, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.records[collectionID] == nil {
		return nil, NewNotFoundError("collection", collectionID)
	}
	if r, ok := b.records[collectionID][recordID]; ok {
		return &r, nil
	}
	return nil, NewNotFoundError("record", recordID)
}

func (b *MemoryBackend) QueryRecords(ctx context.Context, tenantID string, collectionID string, opts models.QueryOptions) (*models.QueryResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.records[collectionID] == nil {
		return &models.QueryResult{
			Items:      []models.Record{},
			TotalItems: 0,
			Page:       opts.Page,
			PerPage:    opts.PerPage,
			TotalPages: 0,
		}, nil
	}

	var items []models.Record
	for _, r := range b.records[collectionID] {
		items = append(items, r)
	}

	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	total := len(items)
	totalPages := (total + opts.PerPage - 1) / opts.PerPage
	start := (opts.Page - 1) * opts.PerPage
	end := start + opts.PerPage
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	return &models.QueryResult{
		Items:      items[start:end],
		TotalItems: total,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (b *MemoryBackend) ListRecords(ctx context.Context, tenantID string, collectionID string, opts ListOptions) ([]models.Record, int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.records[collectionID] == nil {
		return []models.Record{}, 0, nil
	}

	var items []models.Record
	for _, r := range b.records[collectionID] {
		items = append(items, r)
	}

	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	total := len(items)
	start := (opts.Page - 1) * opts.PerPage
	end := start + opts.PerPage
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	return items[start:end], total, nil
}

func (b *MemoryBackend) CreateUser(ctx context.Context, tenantID string, user *models.User) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Duplicate ID check
	if _, exists := b.users[user.ID]; exists {
		return NewAlreadyExistsError("user", "id", user.ID)
	}
	// Duplicate email per tenant check (the critical one for register)
	for _, u := range b.users {
		if u.TenantID == tenantID && u.Email == user.Email {
			return NewAlreadyExistsError("user", "email", user.Email)
		}
	}
	b.users[user.ID] = *user
	return nil
}

func (b *MemoryBackend) UpdateUser(ctx context.Context, tenantID string, user *models.User) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.users[user.ID] = *user
	return nil
}

func (b *MemoryBackend) DeleteUser(ctx context.Context, tenantID string, userID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.users, userID)
	return nil
}

func (b *MemoryBackend) GetUser(ctx context.Context, tenantID string, userID string) (*models.User, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if u, ok := b.users[userID]; ok {
		return &u, nil
	}
	return nil, NewNotFoundError("user", userID)
}

func (b *MemoryBackend) GetUserByEmail(ctx context.Context, tenantID string, email string) (*models.User, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, u := range b.users {
		if u.Email == email && u.TenantID == tenantID {
			return &u, nil
		}
	}
	return nil, NewNotFoundError("user", email)
}

func (b *MemoryBackend) GetUserByTokenKey(ctx context.Context, tokenKey string) (*models.User, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, u := range b.users {
		if u.TokenKey == tokenKey {
			return &u, nil
		}
	}
	return nil, NewNotFoundError("user", tokenKey)
}

func (b *MemoryBackend) ListUsers(ctx context.Context, tenantID string, opts models.QueryOptions) (*models.QueryResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var items []models.Record
	for _, u := range b.users {
		if u.TenantID == tenantID {
			userJSON, _ := json.Marshal(u)
			items = append(items, models.Record{
				ID:   u.ID,
				Data: userJSON,
			})
		}
	}

	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	total := len(items)
	totalPages := (total + opts.PerPage - 1) / opts.PerPage

	return &models.QueryResult{
		Items:      items,
		TotalItems: total,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (b *MemoryBackend) UpdateUserPassword(ctx context.Context, tenantID string, userID string, hashedPassword string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if u, ok := b.users[userID]; ok {
		u.PasswordHash = hashedPassword
		b.users[userID] = u
		return nil
	}
	return NewNotFoundError("user", userID)
}

func (b *MemoryBackend) UpdateUserTokenKey(ctx context.Context, tenantID string, userID string, tokenKey string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if u, ok := b.users[userID]; ok {
		u.TokenKey = tokenKey
		b.users[userID] = u
		return nil
	}
	return NewNotFoundError("user", userID)
}

func (b *MemoryBackend) VerifyUser(ctx context.Context, tenantID string, userID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if u, ok := b.users[userID]; ok {
		u.Verified = true
		b.users[userID] = u
		return nil
	}
	return NewNotFoundError("user", userID)
}

func (b *MemoryBackend) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.tenants[tenant.ID]; exists {
		return NewAlreadyExistsError("tenant", "id", tenant.ID)
	}
	b.tenants[tenant.ID] = *tenant
	return nil
}

func (b *MemoryBackend) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tenants[tenant.ID] = *tenant
	return nil
}

func (b *MemoryBackend) DeleteTenant(ctx context.Context, tenantID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.tenants, tenantID)
	return nil
}

func (b *MemoryBackend) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if t, ok := b.tenants[tenantID]; ok {
		return &t, nil
	}
	return nil, NewNotFoundError("tenant", tenantID)
}

func (b *MemoryBackend) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, t := range b.tenants {
		if t.Slug == slug {
			return &t, nil
		}
	}
	return nil, NewNotFoundError("tenant", slug)
}

func (b *MemoryBackend) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []models.Tenant
	for _, t := range b.tenants {
		result = append(result, t)
	}
	return result, nil
}

func (b *MemoryBackend) ExecMigration(ctx context.Context, version int, up bool, sql string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if up {
		b.migrations[version] = true
	} else {
		delete(b.migrations, version)
	}
	return nil
}

func (b *MemoryBackend) GetMigrationVersion(ctx context.Context) (int, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	max := 0
	for v := range b.migrations {
		if v > max {
			max = v
		}
	}
	return max, nil
}

func (b *MemoryBackend) BeginTx(ctx context.Context) (Tx, error) {
	// In-memory backend doesn't support transactions
	return nil, fmt.Errorf("transactions not supported in memory backend")
}

func (b *MemoryBackend) Subscribe(ctx context.Context, collectionID string, callback func(event Event)) (Subscription, error) {
	// In-memory backend doesn't support subscriptions
	return nil, fmt.Errorf("subscriptions not supported in memory backend")
}
