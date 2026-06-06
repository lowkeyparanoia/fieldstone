// cmd/migrate/main.go
package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"

	"github.com/fieldstone/fieldstone/internal/migrate"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

func main() {
	var (
		driver   = flag.String("driver", "sqlite3", "Database driver (sqlite3 or postgres)")
		dsn      = flag.String("dsn", "", "Database connection string")
		direction = flag.String("direction", "up", "Migration direction (up or down)")
		steps    = flag.Int("steps", 0, "Number of migrations to apply (0 = all)")
		version  = flag.Bool("version", false, "Show current migration version")
		status   = flag.Bool("status", false, "Show migration status")
	)
	flag.Parse()

	if *dsn == "" {
		// Default DSN based on driver
		if *driver == "sqlite3" {
			*dsn = "data/fieldstone.db"
		} else {
			*dsn = "postgres://postgres:postgres@localhost:5432/fieldstone?sslmode=disable"
		}
	}

	// Open database
	db, err := sql.Open(*driver, *dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migrator
	migrator := migrate.NewMigrator(db, *driver)

	// Load migrations
	var migrations []*migrate.Migration
	if *driver == "sqlite3" {
		migrations, err = migrate.LoadMigrationsFromFS(sqliteMigrations, "migrations/sqlite")
	} else {
		migrations, err = migrate.LoadMigrationsFromFS(postgresMigrations, "migrations/postgres")
	}
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// Show version
	if *version {
		applied, err := migrator.GetAppliedMigrations()
		if err != nil {
			log.Fatalf("Failed to get applied migrations: %v", err)
		}
		if len(applied) == 0 {
			fmt.Println("No migrations applied")
		} else {
			fmt.Printf("Current version: %d\n", applied[len(applied)-1])
		}
		return
	}

	// Show status
	if *status {
		status, err := migrator.Status(migrations)
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		fmt.Println("Migration Status:")
		fmt.Println("-----------------")
		for _, s := range status {
			status := "PENDING"
			if s.Applied {
				status = "APPLIED"
			}
			fmt.Printf("%03d %s: %s\n", s.Version, s.Name, status)
		}
		return
	}

	// Apply migrations
	switch *direction {
	case "up":
		fmt.Printf("Applying migrations (driver: %s)...\n", *driver)
		if err := migrator.MigrateUp(migrations); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully!")

	case "down":
		if *steps == 0 {
			*steps = 1
		}
		fmt.Printf("Rolling back %d migration(s)...\n", *steps)
		if err := migrator.MigrateDown(migrations, *steps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed!")

	default:
		fmt.Fprintf(os.Stderr, "Invalid direction: %s (use 'up' or 'down')\n", *direction)
		os.Exit(1)
	}
}
