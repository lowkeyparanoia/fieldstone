// Package backend_test provides unit tests for backend implementations.
// This demonstrates table-driven tests, test fixtures, and mocking patterns in Go.
package backend_test

import (
	"context"
	"testing"
	"time"

	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryBackend_CollectionCRUD tests collection create, read, update, delete operations
func TestMemoryBackend_CollectionCRUD(t *testing.T) {
	// Create a fresh in-memory backend for each test
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	tenantID := "test-tenant"

	t.Run("CreateCollection", func(t *testing.T) {
		collection := &models.Collection{
			ID:   uuid.New().String(),
			Name: "posts",
			Fields: []models.Field{
				{Name: "title", Type: models.FieldTypeText, Options: models.FieldOption{Required: true}},
				{Name: "content", Type: models.FieldTypeText},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := be.CreateCollection(ctx, tenantID, collection)
		require.NoError(t, err)

		// Verify we can retrieve it
		retrieved, err := be.GetCollection(ctx, tenantID, collection.ID)
		require.NoError(t, err)
		assert.Equal(t, collection.Name, retrieved.Name)
		assert.Len(t, retrieved.Fields, 2)
	})

	t.Run("CreateDuplicateCollection", func(t *testing.T) {
		collection := &models.Collection{
			ID:   uuid.New().String(),
			Name: "posts", // Same name as before
			Fields: []models.Field{
				{Name: "title", Type: models.FieldTypeText},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := be.CreateCollection(ctx, tenantID, collection)
		assert.Error(t, err)
		assert.True(t, backend.IsAlreadyExists(err))
	})

	t.Run("GetCollection_NotFound", func(t *testing.T) {
		_, err := be.GetCollection(ctx, tenantID, "non-existent-id")
		assert.Error(t, err)
		assert.True(t, backend.IsNotFound(err))
	})

	t.Run("UpdateCollection", func(t *testing.T) {
		// First create a collection
		collection := &models.Collection{
			ID:   uuid.New().String(),
			Name: "articles",
			Fields: []models.Field{
				{Name: "title", Type: models.FieldTypeText},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := be.CreateCollection(ctx, tenantID, collection)
		require.NoError(t, err)

		// Update it
		collection.Name = "updated-articles"
		err = be.UpdateCollection(ctx, tenantID, collection)
		require.NoError(t, err)

		// Verify update
		retrieved, err := be.GetCollection(ctx, tenantID, collection.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated-articles", retrieved.Name)
	})

	t.Run("DeleteCollection", func(t *testing.T) {
		// Create a collection
		collection := &models.Collection{
			ID:   uuid.New().String(),
			Name: "temp-collection",
			Fields: []models.Field{
				{Name: "field", Type: models.FieldTypeText},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := be.CreateCollection(ctx, tenantID, collection)
		require.NoError(t, err)

		// Delete it
		err = be.DeleteCollection(ctx, tenantID, collection.ID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = be.GetCollection(ctx, tenantID, collection.ID)
		assert.True(t, backend.IsNotFound(err))
	})

	t.Run("ListCollections", func(t *testing.T) {
		// Create multiple collections
		for i := 0; i < 3; i++ {
			collection := &models.Collection{
				ID:   uuid.New().String(),
				Name: "list-test-" + string(rune('a'+i)),
				Fields: []models.Field{
					{Name: "field", Type: models.FieldTypeText},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := be.CreateCollection(ctx, tenantID, collection)
			require.NoError(t, err)
		}

		// List them
		collections, err := be.ListCollections(ctx, tenantID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(collections), 3)
	})
}

// TestMemoryBackend_RecordCRUD tests record operations
func TestMemoryBackend_RecordCRUD(t *testing.T) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	tenantID := "test-tenant"

	// Create a test collection first
	collection := &models.Collection{
		ID:   uuid.New().String(),
		Name: "test-records",
		Fields: []models.Field{
			{Name: "title", Type: models.FieldTypeText},
			{Name: "count", Type: models.FieldTypeNumber},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := be.CreateCollection(ctx, tenantID, collection)
	require.NoError(t, err)

	t.Run("CreateRecord", func(t *testing.T) {
		record := &models.Record{
			ID:           uuid.New().String(),
			CollectionID: collection.ID,
			TenantID:     tenantID,
			Data:         []byte(`{"title": "Test Record", "count": 42}`),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := be.CreateRecord(ctx, tenantID, collection.ID, record)
		require.NoError(t, err)

		// Verify retrieval
		retrieved, err := be.GetRecord(ctx, tenantID, collection.ID, record.ID)
		require.NoError(t, err)
		assert.Equal(t, record.ID, retrieved.ID)
	})

	t.Run("UpdateRecord", func(t *testing.T) {
		// Create a record
		record := &models.Record{
			ID:           uuid.New().String(),
			CollectionID: collection.ID,
			TenantID:     tenantID,
			Data:         []byte(`{"title": "Original", "count": 1}`),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := be.CreateRecord(ctx, tenantID, collection.ID, record)
		require.NoError(t, err)

		// Update it
		record.Data = []byte(`{"title": "Updated", "count": 2}`)
		err = be.UpdateRecord(ctx, tenantID, collection.ID, record)
		require.NoError(t, err)

		// Verify
		retrieved, err := be.GetRecord(ctx, tenantID, collection.ID, record.ID)
		require.NoError(t, err)
		assert.Contains(t, string(retrieved.Data), "Updated")
	})

	t.Run("DeleteRecord", func(t *testing.T) {
		// Create a record
		record := &models.Record{
			ID:           uuid.New().String(),
			CollectionID: collection.ID,
			TenantID:     tenantID,
			Data:         []byte(`{"title": "To Delete"}`),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := be.CreateRecord(ctx, tenantID, collection.ID, record)
		require.NoError(t, err)

		// Delete it
		err = be.DeleteRecord(ctx, tenantID, collection.ID, record.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = be.GetRecord(ctx, tenantID, collection.ID, record.ID)
		assert.True(t, backend.IsNotFound(err))
	})

	t.Run("QueryRecords_Pagination", func(t *testing.T) {
		// Create multiple records
		for i := 0; i < 25; i++ {
			record := &models.Record{
				ID:           uuid.New().String(),
				CollectionID: collection.ID,
				TenantID:     tenantID,
				Data:         []byte(`{"title": "Record ` + string(rune('A'+i)) + `"}`),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			err := be.CreateRecord(ctx, tenantID, collection.ID, record)
			require.NoError(t, err)
		}

		// Query with pagination
		opts := models.QueryOptions{
			Page:    1,
			PerPage: 10,
		}
		result, err := be.QueryRecords(ctx, tenantID, collection.ID, opts)
		require.NoError(t, err)
		assert.Len(t, result.Items, 10)
		assert.GreaterOrEqual(t, result.TotalItems, 25)
		assert.GreaterOrEqual(t, result.TotalPages, 3)

		// Get second page
		opts.Page = 2
		result2, err := be.QueryRecords(ctx, tenantID, collection.ID, opts)
		require.NoError(t, err)
		assert.Len(t, result2.Items, 10)
	})
}

// TestMemoryBackend_UserCRUD tests user operations
func TestMemoryBackend_UserCRUD(t *testing.T) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	tenantID := "test-tenant"

	t.Run("CreateUser", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			TokenKey:     "token_key_123",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := be.CreateUser(ctx, tenantID, user)
		require.NoError(t, err)

		// Verify retrieval
		retrieved, err := be.GetUser(ctx, tenantID, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("CreateDuplicateUser", func(t *testing.T) {
		// First user
		user1 := &models.User{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			Email:        "duplicate@example.com",
			PasswordHash: "hash1",
			TokenKey:     "key1",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := be.CreateUser(ctx, tenantID, user1)
		require.NoError(t, err)

		// Second user with same ID
		user2 := &models.User{
			ID:           user1.ID, // Same ID
			TenantID:     tenantID,
			Email:        "other@example.com",
			PasswordHash: "hash2",
			TokenKey:     "key2",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err = be.CreateUser(ctx, tenantID, user2)
		assert.True(t, backend.IsAlreadyExists(err))
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			Email:        "byemail@example.com",
			PasswordHash: "hash",
			TokenKey:     "key",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := be.CreateUser(ctx, tenantID, user)
		require.NoError(t, err)

		retrieved, err := be.GetUserByEmail(ctx, tenantID, "byemail@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("VerifyUser", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			Email:        "unverified@example.com",
			PasswordHash: "hash",
			TokenKey:     "key",
			Verified:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := be.CreateUser(ctx, tenantID, user)
		require.NoError(t, err)

		// Verify user
		err = be.VerifyUser(ctx, tenantID, user.ID)
		require.NoError(t, err)

		// Check verification status
		retrieved, err := be.GetUser(ctx, tenantID, user.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.Verified)
	})
}

// TestMemoryBackend_TenantCRUD tests tenant operations
func TestMemoryBackend_TenantCRUD(t *testing.T) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()

	t.Run("CreateTenant", func(t *testing.T) {
		tenant := &models.Tenant{
			ID:        uuid.New().String(),
			Name:      "Test Tenant",
			Slug:      "test-tenant",
			Isolation: models.IsolationRowLevel,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := be.CreateTenant(ctx, tenant)
		require.NoError(t, err)

		// Verify
		retrieved, err := be.GetTenant(ctx, tenant.ID)
		require.NoError(t, err)
		assert.Equal(t, tenant.Name, retrieved.Name)
		assert.Equal(t, tenant.Slug, retrieved.Slug)
	})

	t.Run("GetTenantBySlug", func(t *testing.T) {
		tenant := &models.Tenant{
			ID:        uuid.New().String(),
			Name:      "Another Tenant",
			Slug:      "another-tenant",
			Isolation: models.IsolationRowLevel,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := be.CreateTenant(ctx, tenant)
		require.NoError(t, err)

		retrieved, err := be.GetTenantBySlug(ctx, "another-tenant")
		require.NoError(t, err)
		assert.Equal(t, tenant.ID, retrieved.ID)
	})

	t.Run("ListTenants", func(t *testing.T) {
		// Create a few tenants
		for i := 0; i < 3; i++ {
			tenant := &models.Tenant{
				ID:        uuid.New().String(),
				Name:      "List Tenant " + string(rune('A'+i)),
				Slug:      "list-tenant-" + string(rune('a'+i)),
				Isolation: models.IsolationRowLevel,
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := be.CreateTenant(ctx, tenant)
			require.NoError(t, err)
		}

		tenants, err := be.ListTenants(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(tenants), 3)
	})
}

// BenchmarkMemoryBackend_QueryRecords benchmarks record queries
func BenchmarkMemoryBackend_QueryRecords(b *testing.B) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	tenantID := "bench-tenant"

	// Setup: Create collection and records
	collection := &models.Collection{
		ID:   uuid.New().String(),
		Name: "bench-collection",
		Fields: []models.Field{
			{Name: "title", Type: models.FieldTypeText},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := be.CreateCollection(ctx, tenantID, collection)
	if err != nil {
		b.Fatal(err)
	}

	// Create 1000 records
	for i := 0; i < 1000; i++ {
		record := &models.Record{
			ID:           uuid.New().String(),
			CollectionID: collection.ID,
			TenantID:     tenantID,
			Data:         []byte(`{"title": "Record"}`),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := be.CreateRecord(ctx, tenantID, collection.ID, record); err != nil {
			b.Fatal(err)
		}
	}

	// Benchmark querying
	opts := models.QueryOptions{
		Page:    1,
		PerPage: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := be.QueryRecords(ctx, tenantID, collection.ID, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestMemoryBackend_Concurrency tests thread safety
func TestMemoryBackend_Concurrency(t *testing.T) {
	be := backend.NewMemoryBackend()
	ctx := context.Background()
	tenantID := "concurrent-tenant"

	// Create collection
	collection := &models.Collection{
		ID:   uuid.New().String(),
		Name: "concurrent-collection",
		Fields: []models.Field{
			{Name: "value", Type: models.FieldTypeNumber},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := be.CreateCollection(ctx, tenantID, collection)
	require.NoError(t, err)

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				record := &models.Record{
					ID:           uuid.New().String(),
					CollectionID: collection.ID,
					TenantID:     tenantID,
					Data:         []byte(`{"value": 1}`),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				if err := be.CreateRecord(ctx, tenantID, collection.ID, record); err != nil {
					t.Errorf("CreateRecord failed: %v", err)
				}
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify records were created
	result, err := be.QueryRecords(ctx, tenantID, collection.ID, models.QueryOptions{})
	require.NoError(t, err)
	assert.Equal(t, 1000, result.TotalItems)
}
