package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fieldstone/fieldstone/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteBackend implements Backend interface using SQLite
type SQLiteBackend struct {
	db *sql.DB
}

// NewSQLiteBackend creates a new SQLite backend
func NewSQLiteBackend(dsn string) (*SQLiteBackend, error) {
	if dsn == "" {
		dsn = "fieldstone.db"
	}

	db, err := sql.Open("sqlite3", dsn+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Enable WAL mode for better concurrency
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite: %w", err)
	}

	b := &SQLiteBackend{db: db}
	if err := b.initializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return b, nil
}

func (b *SQLiteBackend) initializeSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS _collections (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			fields TEXT NOT NULL,
			system INTEGER DEFAULT 0,
			list_rule TEXT,
			view_rule TEXT,
			create_rule TEXT,
			update_rule TEXT,
			delete_rule TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, name)
		);

		CREATE TABLE IF NOT EXISTS _users (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			email TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			verified INTEGER DEFAULT 0,
			token_key TEXT NOT NULL,
			last_reset_at DATETIME,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, email),
			UNIQUE(token_key)
		);

		CREATE TABLE IF NOT EXISTS _tenants (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			domain TEXT,
			settings TEXT,
			isolation TEXT DEFAULT 'row',
			active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		INSERT OR IGNORE INTO _tenants (id, name, slug, active) 
		VALUES ('default', 'Default', 'default', 1);
	`

	if _, err := b.db.Exec(schema); err != nil {
		return err
	}

	return nil
}

func (b *SQLiteBackend) Ping(ctx context.Context) error {
	return b.db.PingContext(ctx)
}

func (b *SQLiteBackend) Close() error {
	return b.db.Close()
}

func (b *SQLiteBackend) CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	fieldsJSON, err := json.Marshal(collection.Fields)
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, `
		INSERT INTO _collections (id, tenant_id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, collection.ID, tenantID, collection.Name, fieldsJSON, collection.System,
		collection.ListRule, collection.ViewRule, collection.CreateRule, collection.UpdateRule, collection.DeleteRule)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return NewAlreadyExistsError("collection", "name", collection.Name)
		}
		return err
	}

	// Create the actual table for this collection
	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	if err := b.createCollectionTable(ctx, tableName, collection.Fields); err != nil {
		return err
	}

	return nil
}

func (b *SQLiteBackend) createCollectionTable(ctx context.Context, tableName string, fields []models.Field) error {
	// Sanitize table name to prevent SQL injection
	tableName = sanitizeIdentifier(tableName)

	var columns []string
	columns = append(columns, "id TEXT PRIMARY KEY")
	columns = append(columns, "created_at DATETIME DEFAULT CURRENT_TIMESTAMP")
	columns = append(columns, "updated_at DATETIME DEFAULT CURRENT_TIMESTAMP")

	for _, field := range fields {
		colType := sqliteType(field.Type)
		columns = append(columns, fmt.Sprintf("%s %s", sanitizeIdentifier(field.Name), colType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columns, ", "))
	_, err := b.db.ExecContext(ctx, query)
	return err
}

func sqliteType(ft models.FieldType) string {
	switch ft {
	case models.FieldTypeNumber:
		return "REAL"
	case models.FieldTypeBool:
		return "INTEGER"
	case models.FieldTypeDate:
		return "DATETIME"
	default:
		return "TEXT"
	}
}

func sanitizeIdentifier(name string) string {
	// Remove any characters that aren't alphanumeric or underscore
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (b *SQLiteBackend) UpdateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	fieldsJSON, err := json.Marshal(collection.Fields)
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, `
		UPDATE _collections 
		SET name = ?, fields = ?, system = ?, list_rule = ?, view_rule = ?, 
		    create_rule = ?, update_rule = ?, delete_rule = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, collection.Name, fieldsJSON, collection.System, collection.ListRule, collection.ViewRule,
		collection.CreateRule, collection.UpdateRule, collection.DeleteRule, collection.ID, tenantID)

	if err != nil {
		return err
	}

	return nil
}

func (b *SQLiteBackend) DeleteCollection(ctx context.Context, tenantID string, collectionID string) error {
	// Get collection name first
	var name string
	err := b.db.QueryRowContext(ctx, "SELECT name FROM _collections WHERE id = ? AND tenant_id = ?", collectionID, tenantID).Scan(&name)
	if err == sql.ErrNoRows {
		return NewNotFoundError("collection", collectionID)
	}
	if err != nil {
		return err
	}

	// Delete from metadata table
	_, err = b.db.ExecContext(ctx, "DELETE FROM _collections WHERE id = ? AND tenant_id = ?", collectionID, tenantID)
	if err != nil {
		return err
	}

	// Drop the actual table
	tableName := fmt.Sprintf("c_%s_%s", tenantID, name)
	tableName = sanitizeIdentifier(tableName)
	_, err = b.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	return err
}

func (b *SQLiteBackend) GetCollection(ctx context.Context, tenantID string, collectionID string) (*models.Collection, error) {
	var c models.Collection
	var fieldsJSON string

	err := b.db.QueryRowContext(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE id = ? AND tenant_id = ?
	`, collectionID, tenantID).Scan(
		&c.ID, &c.Name, &fieldsJSON, &c.System,
		&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("collection", collectionID)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(fieldsJSON), &c.Fields); err != nil {
		return nil, err
	}

	return &c, nil
}

func (b *SQLiteBackend) GetCollectionByName(ctx context.Context, tenantID string, name string) (*models.Collection, error) {
	var c models.Collection
	var fieldsJSON string

	err := b.db.QueryRowContext(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE name = ? AND tenant_id = ?
	`, name, tenantID).Scan(
		&c.ID, &c.Name, &fieldsJSON, &c.System,
		&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("collection", name)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(fieldsJSON), &c.Fields); err != nil {
		return nil, err
	}

	return &c, nil
}

func (b *SQLiteBackend) ListCollections(ctx context.Context, tenantID string) ([]models.Collection, error) {
	rows, err := b.db.QueryContext(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE tenant_id = ?
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []models.Collection
	for rows.Next() {
		var c models.Collection
		var fieldsJSON string
		err := rows.Scan(
			&c.ID, &c.Name, &fieldsJSON, &c.System,
			&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(fieldsJSON), &c.Fields); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}

	return collections, rows.Err()
}

func (b *SQLiteBackend) CreateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	var data map[string]interface{}
	if err := json.Unmarshal(record.Data, &data); err != nil {
		return err
	}

	columns := []string{"id", "created_at", "updated_at"}
	placeholders := []string{"?", "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"}
	values := []interface{}{record.ID}

	for _, field := range collection.Fields {
		if val, ok := data[field.Name]; ok {
			columns = append(columns, sanitizeIdentifier(field.Name))
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = b.db.ExecContext(ctx, query, values...)
	return err
}

func (b *SQLiteBackend) UpdateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	var data map[string]interface{}
	if err := json.Unmarshal(record.Data, &data); err != nil {
		return err
	}

	var setClauses []string
	var values []interface{}

	for _, field := range collection.Fields {
		if val, ok := data[field.Name]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", sanitizeIdentifier(field.Name)))
			values = append(values, val)
		}
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	values = append(values, record.ID)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		tableName,
		strings.Join(setClauses, ", "),
	)

	_, err = b.db.ExecContext(ctx, query, values...)
	return err
}

func (b *SQLiteBackend) DeleteRecord(ctx context.Context, tenantID string, collectionID string, recordID string) error {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	_, err = b.db.ExecContext(ctx, query, recordID)
	return err
}

func (b *SQLiteBackend) GetRecord(ctx context.Context, tenantID string, collectionID string, recordID string) (*models.Record, error) {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	// Build query dynamically based on fields
	var columns []string
	columns = append(columns, "id", "created_at", "updated_at")
	for _, field := range collection.Fields {
		columns = append(columns, sanitizeIdentifier(field.Name))
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(columns, ", "), tableName)
	row := b.db.QueryRowContext(ctx, query, recordID)

	// Scan into map
	dest := make([]interface{}, len(columns))
	for i := range dest {
		var s interface{}
		dest[i] = &s
	}

	if err := row.Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewNotFoundError("record", recordID)
		}
		return nil, err
	}

	// Convert to map then JSON
	data := make(map[string]interface{})
	for i, col := range columns {
		val := *(dest[i].(*interface{}))
		if val != nil {
			data[col] = val
		}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &models.Record{
		ID:           recordID,
		CollectionID: collectionID,
		TenantID:     tenantID,
		Data:         dataJSON,
		CreatedAt:    data["created_at"].(time.Time),
		UpdatedAt:    data["updated_at"].(time.Time),
	}, nil
}

func (b *SQLiteBackend) QueryRecords(ctx context.Context, tenantID string, collectionID string, opts models.QueryOptions) (*models.QueryResult, error) {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	// For MVP, return all records without complex filtering
	// In production, you'd parse opts.Filter and build WHERE clauses
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var total int
	if err := b.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return nil, err
	}

	// Build select query
	var columns []string
	columns = append(columns, "id", "created_at", "updated_at")
	for _, field := range collection.Fields {
		columns = append(columns, sanitizeIdentifier(field.Name))
	}

	selectQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)

	// Add pagination
	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PerPage
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", opts.PerPage, offset)

	rows, err := b.db.QueryContext(ctx, selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		dest := make([]interface{}, len(columns))
		for i := range dest {
			var s interface{}
			dest[i] = &s
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}

		data := make(map[string]interface{})
		var recordID string
		var createdAt, updatedAt time.Time

		for i, col := range columns {
			val := *(dest[i].(*interface{}))
			switch col {
			case "id":
				recordID = val.(string)
			case "created_at":
				createdAt = val.(time.Time)
			case "updated_at":
				updatedAt = val.(time.Time)
			default:
				if val != nil {
					data[col] = val
				}
			}
		}

		dataJSON, _ := json.Marshal(data)
		records = append(records, models.Record{
			ID:           recordID,
			CollectionID: collectionID,
			TenantID:     tenantID,
			Data:         dataJSON,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		})
	}

	totalPages := (total + opts.PerPage - 1) / opts.PerPage

	return &models.QueryResult{
		Items:      records,
		TotalItems: total,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: totalPages,
	}, rows.Err()
}

// Implement user operations
func (b *SQLiteBackend) CreateUser(ctx context.Context, tenantID string, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := b.db.ExecContext(ctx, `
		INSERT INTO _users (id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, tenantID, user.Email, user.PasswordHash, user.Verified, user.TokenKey, user.LastResetAt, metadataJSON)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return NewAlreadyExistsError("user", "email", user.Email)
		}
		return err
	}
	return nil
}

func (b *SQLiteBackend) UpdateUser(ctx context.Context, tenantID string, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := b.db.ExecContext(ctx, `
		UPDATE _users 
		SET email = ?, verified = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, user.Email, user.Verified, metadataJSON, user.ID, tenantID)
	return err
}

func (b *SQLiteBackend) DeleteUser(ctx context.Context, tenantID string, userID string) error {
	_, err := b.db.ExecContext(ctx, "DELETE FROM _users WHERE id = ? AND tenant_id = ?", userID, tenantID)
	return err
}

func (b *SQLiteBackend) GetUser(ctx context.Context, tenantID string, userID string) (*models.User, error) {
	var u models.User
	var metadataJSON string

	err := b.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE id = ? AND tenant_id = ?
	`, userID, tenantID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("user", userID)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &u.Metadata)
	return &u, nil
}

func (b *SQLiteBackend) GetUserByEmail(ctx context.Context, tenantID string, email string) (*models.User, error) {
	var u models.User
	var metadataJSON string

	err := b.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE email = ? AND tenant_id = ?
	`, email, tenantID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("user", email)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &u.Metadata)
	return &u, nil
}

func (b *SQLiteBackend) GetUserByTokenKey(ctx context.Context, tokenKey string) (*models.User, error) {
	var u models.User
	var metadataJSON string

	err := b.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE token_key = ?
	`, tokenKey).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("user", tokenKey)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(metadataJSON), &u.Metadata)
	return &u, nil
}

func (b *SQLiteBackend) ListUsers(ctx context.Context, tenantID string, opts models.QueryOptions) (*models.QueryResult, error) {
	// Count total
	var total int
	if err := b.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM _users WHERE tenant_id = ?", tenantID).Scan(&total); err != nil {
		return nil, err
	}

	// Pagination
	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PerPage

	rows, err := b.db.QueryContext(ctx, `
		SELECT id, tenant_id, email, verified, token_key, metadata, created_at, updated_at
		FROM _users WHERE tenant_id = ?
		LIMIT ? OFFSET ?
	`, tenantID, opts.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var u models.User
		var metadataJSON string
		err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Verified, &u.TokenKey, &metadataJSON, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(metadataJSON), &u.Metadata)

		// Convert User to Record for generic result
		userJSON, _ := json.Marshal(u)
		records = append(records, models.Record{
			ID:   u.ID,
			Data: userJSON,
		})
	}

	totalPages := (total + opts.PerPage - 1) / opts.PerPage
	return &models.QueryResult{
		Items:      records,
		TotalItems: total,
		Page:       opts.Page,
		PerPage:    opts.PerPage,
		TotalPages: totalPages,
	}, rows.Err()
}

func (b *SQLiteBackend) UpdateUserPassword(ctx context.Context, tenantID string, userID string, hashedPassword string) error {
	_, err := b.db.ExecContext(ctx, `
		UPDATE _users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, hashedPassword, userID, tenantID)
	return err
}

func (b *SQLiteBackend) UpdateUserTokenKey(ctx context.Context, tenantID string, userID string, tokenKey string) error {
	_, err := b.db.ExecContext(ctx, `
		UPDATE _users SET token_key = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, tokenKey, userID, tenantID)
	return err
}

func (b *SQLiteBackend) VerifyUser(ctx context.Context, tenantID string, userID string) error {
	_, err := b.db.ExecContext(ctx, `
		UPDATE _users SET verified = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND tenant_id = ?
	`, userID, tenantID)
	return err
}

// Tenant operations
func (b *SQLiteBackend) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	_, err := b.db.ExecContext(ctx, `
		INSERT INTO _tenants (id, name, slug, domain, settings, isolation, active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, tenant.ID, tenant.Name, tenant.Slug, tenant.Domain, settingsJSON, tenant.Isolation, tenant.Active)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return NewAlreadyExistsError("tenant", "slug", tenant.Slug)
		}
		return err
	}
	return nil
}

func (b *SQLiteBackend) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	_, err := b.db.ExecContext(ctx, `
		UPDATE _tenants 
		SET name = ?, slug = ?, domain = ?, settings = ?, isolation = ?, active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, tenant.Name, tenant.Slug, tenant.Domain, settingsJSON, tenant.Isolation, tenant.Active, tenant.ID)
	return err
}

func (b *SQLiteBackend) DeleteTenant(ctx context.Context, tenantID string) error {
	_, err := b.db.ExecContext(ctx, "DELETE FROM _tenants WHERE id = ?", tenantID)
	return err
}

func (b *SQLiteBackend) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	var t models.Tenant
	var settingsJSON string
	var domain sql.NullString

	err := b.db.QueryRowContext(ctx, `
		SELECT id, name, slug, domain, settings, isolation, active, created_at, updated_at
		FROM _tenants WHERE id = ?
	`, tenantID).Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("tenant", tenantID)
	}
	if err != nil {
		return nil, err
	}

	if domain.Valid {
		t.Domain = &domain.String
	}
	json.Unmarshal([]byte(settingsJSON), &t.Settings)
	return &t, nil
}

func (b *SQLiteBackend) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var t models.Tenant
	var settingsJSON string
	var domain sql.NullString

	err := b.db.QueryRowContext(ctx, `
		SELECT id, name, slug, domain, settings, isolation, active, created_at, updated_at
		FROM _tenants WHERE slug = ?
	`, slug).Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, NewNotFoundError("tenant", slug)
	}
	if err != nil {
		return nil, err
	}

	if domain.Valid {
		t.Domain = &domain.String
	}
	json.Unmarshal([]byte(settingsJSON), &t.Settings)
	return &t, nil
}

func (b *SQLiteBackend) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	rows, err := b.db.QueryContext(ctx, `
		SELECT id, name, slug, domain, settings, isolation, active, created_at, updated_at
		FROM _tenants
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var t models.Tenant
		var settingsJSON string
		var domain sql.NullString
		err := rows.Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if domain.Valid {
			t.Domain = &domain.String
		}
		json.Unmarshal([]byte(settingsJSON), &t.Settings)
		tenants = append(tenants, t)
	}

	return tenants, rows.Err()
}

func (b *SQLiteBackend) ExecMigration(ctx context.Context, version int, up bool, sql string) error {
	if up {
		_, err := b.db.ExecContext(ctx, sql)
		if err != nil {
			return err
		}
		_, err = b.db.ExecContext(ctx, "INSERT INTO _migrations (version) VALUES (?)", version)
		return err
	}
	// Down migrations would go here
	return nil
}

func (b *SQLiteBackend) GetMigrationVersion(ctx context.Context) (int, error) {
	var version int
	err := b.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM _migrations").Scan(&version)
	return version, err
}

func (b *SQLiteBackend) BeginTx(ctx context.Context) (Tx, error) {
	// Implementation would use sql.Tx
	return nil, fmt.Errorf("transactions not yet implemented")
}

func (b *SQLiteBackend) Subscribe(ctx context.Context, collectionID string, callback func(event Event)) (Subscription, error) {
	// SQLite doesn't have built-in pub/sub
	return nil, fmt.Errorf("subscriptions not supported in SQLite backend")
}

// ListRecords delegates to QueryRecords to satisfy the Backend interface.
func (b *SQLiteBackend) ListRecords(ctx context.Context, tenantID string, collectionID string, opts ListOptions) ([]models.Record, int, error) {
	result, err := b.QueryRecords(ctx, tenantID, collectionID, models.QueryOptions{
		Page:    opts.Page,
		PerPage: opts.PerPage,
		Filter:  opts.Filter,
		Sort:    opts.Sort,
	})
	if err != nil {
		return nil, 0, err
	}
	return result.Items, result.TotalItems, nil
}
