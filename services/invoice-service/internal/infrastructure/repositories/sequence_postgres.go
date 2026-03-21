package repositories

import (
	"context"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
)

type NumberSequencePostgres struct {
	db *database.PostgresPool
}

func NewNumberSequencePostgres(db *database.PostgresPool) *NumberSequencePostgres {
	return &NumberSequencePostgres{db: db}
}

// GetNextNumber returns the next sequential number for a given sequence type
// This ensures GoBD compliance - no gaps allowed in invoice/quote numbers
func (r *NumberSequencePostgres) GetNextNumber(ctx context.Context, tenantID, sequenceType string) (int, error) {
	var nextNum int

	// Use PostgreSQL's FOR UPDATE to ensure atomicity
	query := `
		SELECT next_value FROM invoice.number_sequences
		WHERE tenant_id = $1 AND sequence_type = $2
		FOR UPDATE
	`

	row := r.db.QueryRow(ctx, query, tenantID, sequenceType)
	err := row.Scan(&nextNum)

	if err != nil {
		// Sequence doesn't exist, create it
		createQuery := `
			INSERT INTO invoice.number_sequences (tenant_id, sequence_type, next_value, created_at)
			VALUES ($1, $2, 1, NOW())
			RETURNING next_value
		`
		err := r.db.QueryRow(ctx, createQuery, tenantID, sequenceType).Scan(&nextNum)
		if err != nil {
			return 0, fmt.Errorf("failed to initialize sequence: %w", err)
		}
		return nextNum, nil
	}

	// Increment the sequence
	updateQuery := `
		UPDATE invoice.number_sequences
		SET next_value = next_value + 1, updated_at = NOW()
		WHERE tenant_id = $1 AND sequence_type = $2
		RETURNING next_value - 1
	`

	var currentNum int
	err = r.db.QueryRow(ctx, updateQuery, tenantID, sequenceType).Scan(&currentNum)
	if err != nil {
		return 0, fmt.Errorf("failed to increment sequence: %w", err)
	}

	return currentNum, nil
}

// ResetSequence resets a sequence to 1
func (r *NumberSequencePostgres) ResetSequence(ctx context.Context, tenantID, sequenceType string) error {
	query := `
		UPDATE invoice.number_sequences
		SET next_value = 1, updated_at = NOW()
		WHERE tenant_id = $1 AND sequence_type = $2
	`

	_, err := r.db.Exec(ctx, query, tenantID, sequenceType)
	if err != nil {
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	return nil
}
