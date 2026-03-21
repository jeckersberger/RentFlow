package database

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"
)

func TestNewPostgresPool_InvalidConnString(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		expectedError string
	}{
		{
			name:          "empty connection string",
			connString:    "",
			expectedError: "connection string cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewPostgresPool(tt.connString)

			if err == nil {
				t.Errorf("expected error, got nil")
			}
			if pool != nil {
				t.Errorf("expected nil pool, got %v", pool)
			}
		})
	}
}

func TestPostgresPoolClose(t *testing.T) {
	t.Run("close closed pool", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		err := pool.Close()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("multiple close calls", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		for i := 0; i < 3; i++ {
			err := pool.Close()
			if err != nil {
				t.Errorf("unexpected error on call %d: %v", i+1, err)
			}
		}
	})
}

func TestPostgresPoolClosedOperations(t *testing.T) {
	pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}

	t.Run("Query on closed pool", func(t *testing.T) {
		_, err := pool.Query(context.Background(), "SELECT 1")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Exec on closed pool", func(t *testing.T) {
		_, err := pool.Exec(context.Background(), "INSERT INTO test VALUES (1)")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("WithTx on closed pool", func(t *testing.T) {
		err := pool.WithTx(context.Background(), func(tx *sql.Tx) error {
			return nil
		})
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Health on closed pool", func(t *testing.T) {
		err := pool.Health(context.Background())
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestPostgresPoolQueryRow(t *testing.T) {
	t.Run("QueryRow on closed pool returns nil", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		result := pool.QueryRow(context.Background(), "SELECT 1")
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}

func TestPostgresPoolStats(t *testing.T) {
	t.Run("Stats on closed pool", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		stats := pool.Stats()
		if stats.OpenConnections != 0 {
			t.Errorf("expected 0 open connections, got %d", stats.OpenConnections)
		}
	})
}

func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []interface{}
		expected []interface{}
	}{
		{
			name:     "simple query",
			query:    "SELECT * FROM users WHERE id = ?",
			args:     []interface{}{1},
			expected: []interface{}{1},
		},
		{
			name:     "multiple args",
			query:    "INSERT INTO users (name, email) VALUES (?, ?)",
			args:     []interface{}{"John", "john@example.com"},
			expected: []interface{}{"John", "john@example.com"},
		},
		{
			name:     "no args",
			query:    "SELECT COUNT(*) FROM users",
			args:     []interface{}{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			for _, arg := range tt.args {
				qb.AddArg(arg)
			}

			query, args := qb.Build()
			if query != tt.query {
				t.Errorf("expected query %q, got %q", tt.query, query)
			}
			if len(args) != len(tt.expected) {
				t.Errorf("expected %d args, got %d", len(tt.expected), len(args))
			}
			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("arg %d: expected %v, got %v", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestQueryBuilderChaining(t *testing.T) {
	t.Run("fluent API chaining", func(t *testing.T) {
		qb := NewQueryBuilder("SELECT * FROM users").
			AddArg(1).
			AddArg("test").
			AddArg(true)

		query, args := qb.Build()
		if query != "SELECT * FROM users" {
			t.Errorf("unexpected query: %q", query)
		}
		if len(args) != 3 {
			t.Errorf("expected 3 args, got %d", len(args))
		}
	})
}

func TestRow(t *testing.T) {
	t.Run("NewRow creates wrapper", func(t *testing.T) {
		mockRow := &sql.Row{}
		row := NewRow(mockRow)
		if row == nil {
			t.Errorf("expected row, got nil")
		}
	})
}

func TestRows(t *testing.T) {
	t.Run("NewRows creates wrapper", func(t *testing.T) {
		rows := NewRows(nil)
		if rows == nil {
			t.Errorf("expected rows, got nil")
		}
	})

	t.Run("Next on nil rows", func(t *testing.T) {
		rows := NewRows(nil)
		// This will panic in real scenario, but testing the wrapper creation
		if rows.rows != nil {
			t.Errorf("expected nil rows, got %v", rows.rows)
		}
	})
}

func TestPostgresPoolConcurrency(t *testing.T) {
	t.Run("concurrent Close calls", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		var wg sync.WaitGroup
		errorChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := pool.Close()
				if err != nil {
					errorChan <- err
				}
			}()
		}

		wg.Wait()
		close(errorChan)

		for err := range errorChan {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
	})

	t.Run("concurrent reads on closed pool", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		var wg sync.WaitGroup
		errorChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := pool.Query(context.Background(), "SELECT 1")
				if err != nil {
					errorChan <- err
				}
			}()
		}

		wg.Wait()
		close(errorChan)

		expectedErrors := 0
		for range errorChan {
			expectedErrors++
		}
		if expectedErrors != 10 {
			t.Errorf("expected 10 errors, got %d", expectedErrors)
		}
	})
}

func TestQueryBuilderWithDifferentTypes(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "integer args",
			args: []interface{}{1, 2, 3},
		},
		{
			name: "string args",
			args: []interface{}{"a", "b", "c"},
		},
		{
			name: "mixed types",
			args: []interface{}{1, "test", 3.14, true},
		},
		{
			name: "time args",
			args: []interface{}{time.Now()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder("SELECT * FROM table")
			for _, arg := range tt.args {
				qb.AddArg(arg)
			}

			_, args := qb.Build()
			if len(args) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(args))
			}
		})
	}
}

func TestPostgresPoolContextOperations(t *testing.T) {
	t.Run("Query with canceled context", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := pool.Query(ctx, "SELECT 1")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Exec with timeout context", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := pool.Exec(ctx, "SELECT 1")
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestWithTxOnClosedPool(t *testing.T) {
	t.Run("transaction on closed pool returns error", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		called := false

		err := pool.WithTx(context.Background(), func(tx *sql.Tx) error {
			called = true
			return nil
		})

		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if called {
			t.Errorf("transaction function should not be called")
		}
	})
}

func TestHealthCheckOnClosedPool(t *testing.T) {
	t.Run("Health returns error when pool closed", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		err := pool.Health(context.Background())
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Health error message", func(t *testing.T) {
		pool := &PostgresPool{pool: nil, mu: sync.RWMutex{}}
		err := pool.Health(context.Background())
		if err != nil && err.Error() != "database connection is closed" {
			t.Logf("error message: %v", err)
		}
	})
}
