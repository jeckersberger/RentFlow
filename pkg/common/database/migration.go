package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Migrator handles database migrations
type Migrator struct {
	db  *PostgresPool
	dir string
}

// NewMigrator creates a new migrator
func NewMigrator(db *PostgresPool, migrationsDir string) *Migrator {
	return &Migrator{
		db:  db,
		dir: migrationsDir,
	}
}

// Migration represents a single migration
type Migration struct {
	Name      string
	Content   string
	Applied   bool
	AppliedAt *time.Time
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return err
	}

	// Get list of migrations
	migrations, err := m.getMigrations(ctx)
	if err != nil {
		return err
	}

	// Read migration files
	files, err := ioutil.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".sql") && !strings.HasSuffix(file.Name(), ".down.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Run pending migrations
	for _, filename := range migrationFiles {
		// Check if migration already applied
		if m.isMigrationApplied(migrations, filename) {
			continue
		}

		// Read migration file
		filePath := filepath.Join(m.dir, filename)
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration in transaction
		if err := m.db.WithTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", filename, err)
			}

			// Record migration
			_, err := tx.ExecContext(ctx,
				"INSERT INTO migrations (name, applied_at) VALUES ($1, $2)",
				filename, time.Now(),
			)
			return err
		}); err != nil {
			return err
		}

		fmt.Printf("Applied migration: %s\n", filename)
	}

	return nil
}

// Down rolls back migrations
// Note: This requires down_*.sql migration files
func (m *Migrator) Down(ctx context.Context, steps int) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return err
	}

	// Get applied migrations in reverse order
	rows, err := m.db.Query(ctx, `
		SELECT name FROM migrations
		WHERE applied_at IS NOT NULL
		ORDER BY applied_at DESC
		LIMIT $1
	`, steps)
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var appliedMigrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		appliedMigrations = append(appliedMigrations, name)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Rollback migrations
	for _, filename := range appliedMigrations {
		// Convert up_*.sql to down_*.sql
		downFilename := strings.Replace(filename, "up_", "down_", 1)

		// Read migration file
		filePath := filepath.Join(m.dir, downFilename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("down migration not found: %s", downFilename)
		}

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", downFilename, err)
		}

		// Execute migration in transaction
		if err := m.db.WithTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", downFilename, err)
			}

			// Remove migration record
			_, err := tx.ExecContext(ctx,
				"DELETE FROM migrations WHERE name = $1",
				filename,
			)
			return err
		}); err != nil {
			return err
		}

		fmt.Printf("Rolled back migration: %s\n", filename)
	}

	return nil
}

// Status returns the status of all migrations
func (m *Migrator) Status(ctx context.Context) ([]Migration, error) {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return nil, err
	}

	// Get applied migrations
	appliedMigrations := make(map[string]Migration)
	rows, err := m.db.Query(ctx, "SELECT name, applied_at FROM migrations WHERE applied_at IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var appliedAt time.Time
		if err := rows.Scan(&name, &appliedAt); err != nil {
			return nil, err
		}
		appliedMigrations[name] = Migration{
			Name:      name,
			Applied:   true,
			AppliedAt: &appliedAt,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Read migration files
	var migrations []Migration
	files, err := ioutil.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".sql") && !strings.HasSuffix(file.Name(), ".down.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Populate migration status
	for _, filename := range migrationFiles {
		if applied, exists := appliedMigrations[filename]; exists {
			migrations = append(migrations, applied)
		} else {
			migrations = append(migrations, Migration{
				Name:    filename,
				Applied: false,
			})
		}
	}

	return migrations, nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	_, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getMigrations retrieves all applied migrations
func (m *Migrator) getMigrations(ctx context.Context) (map[string]bool, error) {
	migrations := make(map[string]bool)

	rows, err := m.db.Query(ctx, "SELECT name FROM migrations WHERE applied_at IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations[name] = true
	}

	return migrations, rows.Err()
}

// isMigrationApplied checks if a migration has been applied
func (m *Migrator) isMigrationApplied(migrations map[string]bool, name string) bool {
	return migrations[name]
}
