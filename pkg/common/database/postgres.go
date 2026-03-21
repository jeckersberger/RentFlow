package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// PostgresPool wraps a database connection pool
// It provides a high-level interface for database operations
type PostgresPool struct {
	pool *sql.DB
	mu   sync.RWMutex
}

// NewPostgresPool creates a new PostgreSQL connection pool
func NewPostgresPool(connString string) (*PostgresPool, error) {
	if connString == "" {
		return nil, fmt.Errorf("connection string cannot be empty")
	}

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresPool{pool: db}, nil
}

// Close closes the database connection pool
func (p *PostgresPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool == nil {
		return nil
	}

	return p.pool.Close()
}

// Query executes a query and returns the result rows
func (p *PostgresPool) Query(ctx context.Context, sql string, args ...interface{}) (*sql.Rows, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pool == nil {
		return nil, fmt.Errorf("database connection is closed")
	}

	return p.pool.QueryContext(ctx, sql, args...)
}

// QueryRow executes a query that returns a single row
func (p *PostgresPool) QueryRow(ctx context.Context, sql string, args ...interface{}) *sql.Row {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pool == nil {
		return nil
	}

	return p.pool.QueryRowContext(ctx, sql, args...)
}

// Exec executes a query that doesn't return rows
func (p *PostgresPool) Exec(ctx context.Context, sql string, args ...interface{}) (sql.Result, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pool == nil {
		return nil, fmt.Errorf("database connection is closed")
	}

	return p.pool.ExecContext(ctx, sql, args...)
}

// WithTx executes a function within a transaction
// If the function returns an error, the transaction is rolled back
func (p *PostgresPool) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	p.mu.RLock()
	pool := p.pool
	p.mu.RUnlock()

	if pool == nil {
		return fmt.Errorf("database connection is closed")
	}

	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Health checks the health of the database connection
func (p *PostgresPool) Health(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pool == nil {
		return fmt.Errorf("database connection is closed")
	}

	return p.pool.PingContext(ctx)
}

// Stats returns connection pool statistics
func (p *PostgresPool) Stats() sql.DBStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pool == nil {
		return sql.DBStats{}
	}

	return p.pool.Stats()
}

// QueryBuilder provides a simple interface for building SQL queries
// For production, consider using sqlc or similar tools
type QueryBuilder struct {
	query string
	args  []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(query string) *QueryBuilder {
	return &QueryBuilder{
		query: query,
		args:  []interface{}{},
	}
}

// AddArg adds an argument to the query
func (qb *QueryBuilder) AddArg(arg interface{}) *QueryBuilder {
	qb.args = append(qb.args, arg)
	return qb
}

// Build returns the query string and arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

// Scan helper for scanning a single row
type Row struct {
	row *sql.Row
}

// NewRow creates a new row wrapper
func NewRow(row *sql.Row) *Row {
	return &Row{row: row}
}

// Scan scans the row into the provided values
func (r *Row) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// Rows helper for scanning multiple rows
type Rows struct {
	rows *sql.Rows
}

// NewRows creates a new rows wrapper
func NewRows(rows *sql.Rows) *Rows {
	return &Rows{rows: rows}
}

// Scan scans the current row into the provided values
func (r *Rows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Next moves to the next row
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Close closes the rows
func (r *Rows) Close() error {
	return r.rows.Close()
}

// Err returns any error encountered during iteration
func (r *Rows) Err() error {
	return r.rows.Err()
}
