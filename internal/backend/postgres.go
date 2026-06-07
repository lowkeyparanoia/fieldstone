package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fieldstone/fieldstone/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresBackend implements Backend interface using PostgreSQL
type PostgresBackend struct {
	pool *pgxpool.Pool
}

// NewPostgresBackend creates a new PostgreSQL backend
func NewPostgresBackend(dsn string, maxConns, minConns int) (*PostgresBackend, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	if maxConns > 0 {
		config.MaxConns = int32(maxConns)
	}
	if minConns > 0 {
		config.MinConns = int32(minConns)
	}

	// Add AfterConnect hook for RLS
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set default search path
		_, err := conn.Exec(ctx, "SET search_path TO public")
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	b := &PostgresBackend{pool: pool}
	if err := b.initializeSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return b, nil
}

func (b *PostgresBackend) initializeSchema() error {
	ctx := context.Background()

	schema := `
		CREATE TABLE IF NOT EXISTS _collections (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name TEXT NOT NULL,
			fields JSONB NOT NULL DEFAULT '[]',
			system BOOLEAN DEFAULT FALSE,
			list_rule TEXT,
			view_rule TEXT,
			create_rule TEXT,
			update_rule TEXT,
			delete_rule TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, name)
		);

		CREATE TABLE IF NOT EXISTS _users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			email TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			verified BOOLEAN DEFAULT FALSE,
			token_key UUID DEFAULT gen_random_uuid(),
			last_reset_at TIMESTAMPTZ,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, email),
			UNIQUE(token_key)
		);

		CREATE TABLE IF NOT EXISTS _tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			domain TEXT,
			settings JSONB DEFAULT '{}',
			isolation TEXT DEFAULT 'row',
			active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Insert default tenant
		INSERT INTO _tenants (id, name, slug, active) 
		VALUES ('00000000-0000-0000-0000-000000000000', 'Default', 'default', TRUE)
		ON CONFLICT (id) DO NOTHING;

		-- Create RLS helper functions
		CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
		BEGIN
			RETURN NULLIF(current_setting('app.current_tenant', TRUE), '')::UUID;
		END;
		$$ LANGUAGE plpgsql;

		CREATE OR REPLACE FUNCTION auth.uid() RETURNS UUID AS $$
		BEGIN
			RETURN NULLIF(current_setting('request.jwt.claims', TRUE)::json->>'sub', '')::UUID;
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;

		CREATE OR REPLACE FUNCTION auth.role() RETURNS TEXT AS $$
		BEGIN
			RETURN COALESCE(current_setting('request.jwt.claims', TRUE)::json->>'role', 'authenticated');
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;

		-- Enable RLS on tables
		ALTER TABLE _collections ENABLE ROW LEVEL SECURITY;
		ALTER TABLE _users ENABLE ROW LEVEL SECURITY;

		-- Create RLS policies for row-level isolation
		CREATE POLICY tenant_isolation_collections ON _collections
			USING (tenant_id = current_tenant_id());

		CREATE POLICY tenant_isolation_users ON _users
			USING (tenant_id = current_tenant_id());
	`

	_, err := b.pool.Exec(ctx, schema)
	return err
}

func (b *PostgresBackend) Ping(ctx context.Context) error {
	return b.pool.Ping(ctx)
}

// Pool returns the underlying pgxpool.Pool.
func (b *PostgresBackend) Pool() *pgxpool.Pool {
	return b.pool
}

func (b *PostgresBackend) Close() error {
	b.pool.Close()
	return nil
}

// Helper to set tenant context for RLS
func (b *PostgresBackend) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000000"
	}
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		return ctx, err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "SET LOCAL app.current_tenant = $1", tenantID)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (b *PostgresBackend) CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	fieldsJSON, err := json.Marshal(collection.Fields)
	if err != nil {
		return err
	}

	_, err = b.pool.Exec(ctx, `
		INSERT INTO _collections (id, tenant_id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, collection.ID, tenantID, collection.Name, fieldsJSON, collection.System,
		collection.ListRule, collection.ViewRule, collection.CreateRule, collection.UpdateRule, collection.DeleteRule)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"_collections_tenant_id_name_key\" (SQLSTATE 23505)" {
			return NewAlreadyExistsError("collection", "name", collection.Name)
		}
		return err
	}

	// Create actual table
	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	if err := b.createCollectionTable(ctx, tableName, collection.Fields); err != nil {
		return err
	}

	return nil
}

func (b *PostgresBackend) createCollectionTable(ctx context.Context, tableName string, fields []models.Field) error {
	// Sanitize table name
	tableName = sanitizeIdentifier(tableName)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)
	`, tableName)

	_, err := b.pool.Exec(ctx, query)
	if err != nil {
		return err
	}

	// Add columns for each field
	for _, field := range fields {
		colType := postgresType(field.Type)
		alterQuery := fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s",
			tableName, sanitizeIdentifier(field.Name), colType,
		)
		_, err = b.pool.Exec(ctx, alterQuery)
		if err != nil {
			return err
		}
	}

	return nil
}

func postgresType(ft models.FieldType) string {
	switch ft {
	case models.FieldTypeNumber:
		return "NUMERIC"
	case models.FieldTypeBool:
		return "BOOLEAN"
	case models.FieldTypeDate:
		return "TIMESTAMPTZ"
	case models.FieldTypeJSON:
		return "JSONB"
	default:
		return "TEXT"
	}
}

func (b *PostgresBackend) UpdateCollection(ctx context.Context, tenantID string, collection *models.Collection) error {
	fieldsJSON, err := json.Marshal(collection.Fields)
	if err != nil {
		return err
	}

	_, err = b.pool.Exec(ctx, `
		UPDATE _collections 
		SET name = $1, fields = $2, system = $3, list_rule = $4, view_rule = $5, 
		    create_rule = $6, update_rule = $7, delete_rule = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND tenant_id = $10
	`, collection.Name, fieldsJSON, collection.System, collection.ListRule, collection.ViewRule,
		collection.CreateRule, collection.UpdateRule, collection.DeleteRule, collection.ID, tenantID)

	return err
}

func (b *PostgresBackend) DeleteCollection(ctx context.Context, tenantID string, collectionID string) error {
	var name string
	err := b.pool.QueryRow(ctx, 
		"SELECT name FROM _collections WHERE id = $1 AND tenant_id = $2", 
		collectionID, tenantID).Scan(&name)
	
	if err == pgx.ErrNoRows {
		return NewNotFoundError("collection", collectionID)
	}
	if err != nil {
		return err
	}

	_, err = b.pool.Exec(ctx, 
		"DELETE FROM _collections WHERE id = $1 AND tenant_id = $2", 
		collectionID, tenantID)
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, name)
	tableName = sanitizeIdentifier(tableName)
	_, err = b.pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	return err
}

func (b *PostgresBackend) GetCollection(ctx context.Context, tenantID string, collectionID string) (*models.Collection, error) {
	var c models.Collection
	var fieldsJSON []byte

	err := b.pool.QueryRow(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE id = $1 AND tenant_id = $2
	`, collectionID, tenantID).Scan(
		&c.ID, &c.Name, &fieldsJSON, &c.System,
		&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("collection", collectionID)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fieldsJSON, &c.Fields); err != nil {
		return nil, err
	}

	return &c, nil
}

func (b *PostgresBackend) GetCollectionByName(ctx context.Context, tenantID string, name string) (*models.Collection, error) {
	var c models.Collection
	var fieldsJSON []byte

	err := b.pool.QueryRow(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE name = $1 AND tenant_id = $2
	`, name, tenantID).Scan(
		&c.ID, &c.Name, &fieldsJSON, &c.System,
		&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("collection", name)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(fieldsJSON, &c.Fields); err != nil {
		return nil, err
	}

	return &c, nil
}

func (b *PostgresBackend) ListCollections(ctx context.Context, tenantID string) ([]models.Collection, error) {
	rows, err := b.pool.Query(ctx, `
		SELECT id, name, fields, system, list_rule, view_rule, create_rule, update_rule, delete_rule, created_at, updated_at
		FROM _collections WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []models.Collection
	for rows.Next() {
		var c models.Collection
		var fieldsJSON []byte
		err := rows.Scan(
			&c.ID, &c.Name, &fieldsJSON, &c.System,
			&c.ListRule, &c.ViewRule, &c.CreateRule, &c.UpdateRule, &c.DeleteRule,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(fieldsJSON, &c.Fields); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}

	return collections, rows.Err()
}

func (b *PostgresBackend) CreateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
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
	placeholders := []string{"$1", "CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"}
	values := []interface{}{record.ID}

	i := 4
	for _, field := range collection.Fields {
		if val, ok := data[field.Name]; ok {
			columns = append(columns, sanitizeIdentifier(field.Name))
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			values = append(values, val)
			i++
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		joinStrings(columns, ", "),
		joinStrings(placeholders, ", "),
	)

	_, err = b.pool.Exec(ctx, query, values...)
	return err
}

func (b *PostgresBackend) UpdateRecord(ctx context.Context, tenantID string, collectionID string, record *models.Record) error {
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

	i := 2
	for _, field := range collection.Fields {
		if val, ok := data[field.Name]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", sanitizeIdentifier(field.Name), i))
			values = append(values, val)
			i++
		}
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	values = append(values, record.ID)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		tableName,
		joinStrings(setClauses, ", "),
		i,
	)

	_, err = b.pool.Exec(ctx, query, values...)
	return err
}

func (b *PostgresBackend) DeleteRecord(ctx context.Context, tenantID string, collectionID string, recordID string) error {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
	_, err = b.pool.Exec(ctx, query, recordID)
	return err
}

func (b *PostgresBackend) GetRecord(ctx context.Context, tenantID string, collectionID string, recordID string) (*models.Record, error) {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	var columns []string
	columns = append(columns, "id", "created_at", "updated_at")
	for _, field := range collection.Fields {
		columns = append(columns, sanitizeIdentifier(field.Name))
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", joinStrings(columns, ", "), tableName)
	row := b.pool.QueryRow(ctx, query, recordID)

	// Scan into map
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, NewNotFoundError("record", recordID)
		}
		return nil, err
	}

	data := make(map[string]interface{})
	var recordID2 string
	var createdAt, updatedAt time.Time

	for i, col := range columns {
		switch col {
		case "id":
			if v, ok := values[i].(string); ok {
				recordID2 = v
			}
		case "created_at":
			if v, ok := values[i].(time.Time); ok {
				createdAt = v
			}
		case "updated_at":
			if v, ok := values[i].(time.Time); ok {
				updatedAt = v
			}
		default:
			if values[i] != nil {
				data[col] = values[i]
			}
		}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &models.Record{
		ID:           recordID2,
		CollectionID: collectionID,
		TenantID:     tenantID,
		Data:         dataJSON,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

func (b *PostgresBackend) QueryRecords(ctx context.Context, tenantID string, collectionID string, opts models.QueryOptions) (*models.QueryResult, error) {
	collection, err := b.GetCollection(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	tableName := fmt.Sprintf("c_%s_%s", tenantID, collection.Name)
	tableName = sanitizeIdentifier(tableName)

	// Count total
	var total int
	if err := b.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&total); err != nil {
		return nil, err
	}

	// Build select query
	var columns []string
	columns = append(columns, "id", "created_at", "updated_at")
	for _, field := range collection.Fields {
		columns = append(columns, sanitizeIdentifier(field.Name))
	}

	selectQuery := fmt.Sprintf("SELECT %s FROM %s", joinStrings(columns, ", "), tableName)

	// Pagination
	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PerPage
	selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", opts.PerPage, offset)

	rows, err := b.pool.Query(ctx, selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		data := make(map[string]interface{})
		var recordID string
		var createdAt, updatedAt time.Time

		for i, col := range columns {
			switch col {
			case "id":
				if v, ok := values[i].(string); ok {
					recordID = v
				}
			case "created_at":
				if v, ok := values[i].(time.Time); ok {
					createdAt = v
				}
			case "updated_at":
				if v, ok := values[i].(time.Time); ok {
					updatedAt = v
				}
			default:
				if values[i] != nil {
					data[col] = values[i]
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

// User operations
func (b *PostgresBackend) CreateUser(ctx context.Context, tenantID string, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := b.pool.Exec(ctx, `
		INSERT INTO _users (id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, tenantID, user.Email, user.PasswordHash, user.Verified, user.TokenKey, user.LastResetAt, metadataJSON)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"_users_tenant_id_email_key\" (SQLSTATE 23505)" {
			return NewAlreadyExistsError("user", "email", user.Email)
		}
		return err
	}
	return nil
}

func (b *PostgresBackend) UpdateUser(ctx context.Context, tenantID string, user *models.User) error {
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := b.pool.Exec(ctx, `
		UPDATE _users 
		SET email = $1, verified = $2, metadata = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND tenant_id = $5
	`, user.Email, user.Verified, metadataJSON, user.ID, tenantID)
	return err
}

func (b *PostgresBackend) DeleteUser(ctx context.Context, tenantID string, userID string) error {
	_, err := b.pool.Exec(ctx, "DELETE FROM _users WHERE id = $1 AND tenant_id = $2", userID, tenantID)
	return err
}

func (b *PostgresBackend) GetUser(ctx context.Context, tenantID string, userID string) (*models.User, error) {
	var u models.User
	var metadataJSON []byte

	err := b.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE id = $1 AND tenant_id = $2
	`, userID, tenantID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("user", userID)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadataJSON, &u.Metadata)
	return &u, nil
}

func (b *PostgresBackend) GetUserByEmail(ctx context.Context, tenantID string, email string) (*models.User, error) {
	var u models.User
	var metadataJSON []byte

	err := b.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE email = $1 AND tenant_id = $2
	`, email, tenantID).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("user", email)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadataJSON, &u.Metadata)
	return &u, nil
}

func (b *PostgresBackend) GetUserByTokenKey(ctx context.Context, tokenKey string) (*models.User, error) {
	var u models.User
	var metadataJSON []byte

	err := b.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, verified, token_key, last_reset_at, metadata, created_at, updated_at
		FROM _users WHERE token_key = $1
	`, tokenKey).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Verified, &u.TokenKey, &u.LastResetAt, &metadataJSON, &u.CreatedAt, &u.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("user", tokenKey)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metadataJSON, &u.Metadata)
	return &u, nil
}

func (b *PostgresBackend) ListUsers(ctx context.Context, tenantID string, opts models.QueryOptions) (*models.QueryResult, error) {
	var total int
	if err := b.pool.QueryRow(ctx, "SELECT COUNT(*) FROM _users WHERE tenant_id = $1", tenantID).Scan(&total); err != nil {
		return nil, err
	}

	if opts.PerPage <= 0 {
		opts.PerPage = 30
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}
	offset := (opts.Page - 1) * opts.PerPage

	rows, err := b.pool.Query(ctx, `
		SELECT id, tenant_id, email, verified, token_key, metadata, created_at, updated_at
		FROM _users WHERE tenant_id = $1
		LIMIT $2 OFFSET $3
	`, tenantID, opts.PerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var u models.User
		var metadataJSON []byte
		err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Verified, &u.TokenKey, &metadataJSON, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(metadataJSON, &u.Metadata)

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

func (b *PostgresBackend) UpdateUserPassword(ctx context.Context, tenantID string, userID string, hashedPassword string) error {
	_, err := b.pool.Exec(ctx, `
		UPDATE _users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND tenant_id = $3
	`, hashedPassword, userID, tenantID)
	return err
}

func (b *PostgresBackend) UpdateUserTokenKey(ctx context.Context, tenantID string, userID string, tokenKey string) error {
	_, err := b.pool.Exec(ctx, `
		UPDATE _users SET token_key = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND tenant_id = $3
	`, tokenKey, userID, tenantID)
	return err
}

func (b *PostgresBackend) VerifyUser(ctx context.Context, tenantID string, userID string) error {
	_, err := b.pool.Exec(ctx, `
		UPDATE _users SET verified = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND tenant_id = $2
	`, userID, tenantID)
	return err
}

// Tenant operations
func (b *PostgresBackend) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	_, err := b.pool.Exec(ctx, `
		INSERT INTO _tenants (id, name, slug, domain, settings, isolation, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, tenant.ID, tenant.Name, tenant.Slug, tenant.Domain, settingsJSON, tenant.Isolation, tenant.Active)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"_tenants_slug_key\" (SQLSTATE 23505)" {
			return NewAlreadyExistsError("tenant", "slug", tenant.Slug)
		}
		return err
	}
	return nil
}

func (b *PostgresBackend) UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	settingsJSON, _ := json.Marshal(tenant.Settings)

	_, err := b.pool.Exec(ctx, `
		UPDATE _tenants 
		SET name = $1, slug = $2, domain = $3, settings = $4, isolation = $5, active = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`, tenant.Name, tenant.Slug, tenant.Domain, settingsJSON, tenant.Isolation, tenant.Active, tenant.ID)
	return err
}

func (b *PostgresBackend) DeleteTenant(ctx context.Context, tenantID string) error {
	_, err := b.pool.Exec(ctx, "DELETE FROM _tenants WHERE id = $1", tenantID)
	return err
}

func (b *PostgresBackend) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	var t models.Tenant
	var settingsJSON []byte
	var domain *string

	err := b.pool.QueryRow(ctx, `
		SELECT id, name, slug, domain, settings, isolation, active, created_at, updated_at
		FROM _tenants WHERE id = $1
	`, tenantID).Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("tenant", tenantID)
	}
	if err != nil {
		return nil, err
	}

	t.Domain = domain
	json.Unmarshal(settingsJSON, &t.Settings)
	return &t, nil
}

func (b *PostgresBackend) GetTenantBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var t models.Tenant
	var settingsJSON []byte
	var domain *string

	err := b.pool.QueryRow(ctx, `
		SELECT id, name, slug, domain, settings, isolation, active, created_at, updated_at
		FROM _tenants WHERE slug = $1
	`, slug).Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, NewNotFoundError("tenant", slug)
	}
	if err != nil {
		return nil, err
	}

	t.Domain = domain
	json.Unmarshal(settingsJSON, &t.Settings)
	return &t, nil
}

func (b *PostgresBackend) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	rows, err := b.pool.Query(ctx, `
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
		var settingsJSON []byte
		var domain *string
		err := rows.Scan(&t.ID, &t.Name, &t.Slug, &domain, &settingsJSON, &t.Isolation, &t.Active, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		t.Domain = domain
		json.Unmarshal(settingsJSON, &t.Settings)
		tenants = append(tenants, t)
	}

	return tenants, rows.Err()
}

func (b *PostgresBackend) ExecMigration(ctx context.Context, version int, up bool, sql string) error {
	if up {
		_, err := b.pool.Exec(ctx, sql)
		if err != nil {
			return err
		}
		_, err = b.pool.Exec(ctx, "INSERT INTO _migrations (version) VALUES ($1)", version)
		return err
	}
	return nil
}

func (b *PostgresBackend) GetMigrationVersion(ctx context.Context) (int, error) {
	var version int
	err := b.pool.QueryRow(ctx, "SELECT COALESCE(MAX(version), 0) FROM _migrations").Scan(&version)
	return version, err
}

func (b *PostgresBackend) BeginTx(ctx context.Context) (Tx, error) {
	return nil, fmt.Errorf("transactions not yet implemented")
}

func (b *PostgresBackend) Subscribe(ctx context.Context, collectionID string, callback func(event Event)) (Subscription, error) {
	// PostgreSQL LISTEN/NOTIFY could be used here
	return nil, fmt.Errorf("subscriptions not yet implemented for PostgreSQL")
}

// Helper function
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ListRecords delegates to QueryRecords to satisfy the Backend interface.
func (b *PostgresBackend) ListRecords(ctx context.Context, tenantID string, collectionID string, opts ListOptions) ([]models.Record, int, error) {
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
