package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SequenceRepo implements domain.NumberSequenceRepository using PostgreSQL.
type SequenceRepo struct {
	pool *pgxpool.Pool
}

// NewSequenceRepo creates a new SequenceRepo.
func NewSequenceRepo(pool *pgxpool.Pool) *SequenceRepo {
	return &SequenceRepo{pool: pool}
}

// EnsureSequence creates a number sequence row if it does not already exist.
func (r *SequenceRepo) EnsureSequence(ctx context.Context, tenantID uuid.UUID, prefix string, year int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO number_sequences (tenant_id, prefix, year)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (tenant_id, prefix, year) DO NOTHING`,
		tenantID, prefix, year,
	)
	if err != nil {
		return fmt.Errorf("sequence_repo: ensure_sequence: %w", err)
	}
	return nil
}

// NextNumber atomically increments the sequence and returns the new number.
func (r *SequenceRepo) NextNumber(ctx context.Context, tenantID uuid.UUID, prefix string, year int) (int64, error) {
	if err := r.EnsureSequence(ctx, tenantID, prefix, year); err != nil {
		return 0, err
	}

	var lastNumber int64
	err := r.pool.QueryRow(ctx,
		`UPDATE number_sequences SET last_number = last_number + 1
		 WHERE tenant_id = $1 AND prefix = $2 AND year = $3
		 RETURNING last_number`,
		tenantID, prefix, year,
	).Scan(&lastNumber)
	if err != nil {
		return 0, fmt.Errorf("sequence_repo: next_number: %w", err)
	}
	return lastNumber, nil
}
