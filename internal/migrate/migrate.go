// Package migrate provides database migration functionality for Fieldstone.
// It supports both SQLite and PostgreSQL backends with versioned migrations.
package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

// Migration represents a single database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// Migrator handles database migrations
type Migrator struct {
	db     *sql.DB
	driver string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, driver string) *Migrator {
	return &Migrator{
		db:     db,
		driver: driver,
	}
}

// CreateMigrationsTable creates the migrations tracking table
func (m *Migrator) CreateMigrationsTable() error {
	var query string
	
	switch m.driver {
	case "sqlite3":
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`
	case "postgres":
		query = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	default:
		return fmt.Errorf("unsupported driver: %s", m.driver)
	}
	
	_, err := m.db.Exec(query)
	return err
}

// GetAppliedMigrations returns list of applied migration versions
func (m *Migrator) GetAppliedMigrations() ([]int, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	
	return versions, rows.Err()
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(migration *Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Apply migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("migration %d failed: %w", migration.Version, err)
	}
	
	// Record migration
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		migration.Version,
	); err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}
	
	return tx.Commit()
}

// RollbackMigration rolls back a single migration
func (m *Migrator) RollbackMigration(migration *Migration) error {
	if migration.Down == "" {
		return fmt.Errorf("migration %d has no rollback", migration.Version)
	}
	
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Apply rollback
	if _, err := tx.Exec(migration.Down); err != nil {
		return fmt.Errorf("rollback %d failed: %w", migration.Version, err)
	}
	
	// Remove migration record
	if _, err := tx.Exec(
		"DELETE FROM schema_migrations WHERE version = $1",
		migration.Version,
	); err != nil {
		return fmt.Errorf("failed to remove migration %d: %w", migration.Version, err)
	}
	
	return tx.Commit()
}

// MigrateUp applies all pending migrations
func (m *Migrator) MigrateUp(migrations []*Migration) error {
	if err := m.CreateMigrationsTable(); err != nil {
		return err
	}
	
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}
	
	appliedMap := make(map[int]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}
	
	for _, migration := range migrations {
		if appliedMap[migration.Version] {
			log.Debug().Int("version", migration.Version).Msg("Migration already applied")
			continue
		}
		
		log.Info().
			Int("version", migration.Version).
			Str("name", migration.Name).
			Msg("Applying migration")
		
		if err := m.ApplyMigration(migration); err != nil {
			return err
		}
	}
	
	return nil
}

// MigrateDown rolls back the last n migrations
func (m *Migrator) MigrateDown(migrations []*Migration, n int) error {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}
	
	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}
	
	// Sort applied versions in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(applied)))
	
	migrationMap := make(map[int]*Migration)
	for _, m := range migrations {
		migrationMap[m.Version] = m
	}
	
	for i := 0; i < n && i < len(applied); i++ {
		version := applied[i]
		migration, ok := migrationMap[version]
		if !ok {
			return fmt.Errorf("migration %d not found", version)
		}
		
		log.Info().
			Int("version", migration.Version).
			Str("name", migration.Name).
			Msg("Rolling back migration")
		
		if err := m.RollbackMigration(migration); err != nil {
			return err
		}
	}
	
	return nil
}

// LoadMigrationsFromFS loads migrations from embedded filesystem
func LoadMigrationsFromFS(fs embed.FS, dir string) ([]*Migration, error) {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	
	var migrations []*Migration
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		
		// Parse version from filename (e.g., "001_init.up.sql")
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}
		
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		
		// Read up migration
		upSQL, err := fs.ReadFile(path.Join(dir, name))
		if err != nil {
			return nil, err
		}
		
		// Try to read down migration
		downName := strings.Replace(name, ".up.sql", ".down.sql", 1)
		var downSQL []byte
		if downData, err := fs.ReadFile(path.Join(dir, downName)); err == nil {
			downSQL = downData
		}
		
		migrations = append(migrations, &Migration{
			Version: version,
			Name:    strings.TrimSuffix(parts[1], ".up.sql"),
			Up:      string(upSQL),
			Down:    string(downSQL),
		})
	}
	
	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	
	return migrations, nil
}

// MigrationStatus represents the current migration state
type MigrationStatus struct {
	Version     int
	Name        string
	Applied     bool
	AppliedAt   *string
}

// Status returns the status of all migrations
func (m *Migrator) Status(migrations []*Migration) ([]MigrationStatus, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}
	
	appliedMap := make(map[int]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}
	
	var status []MigrationStatus
	for _, m := range migrations {
		status = append(status, MigrationStatus{
			Version: m.Version,
			Name:    m.Name,
			Applied: appliedMap[m.Version],
		})
	}
	
	return status, nil
}
