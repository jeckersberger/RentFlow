package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/auth/internal/domain"
)

// SetupRepo implements domain.SetupRepository using PostgreSQL.
type SetupRepo struct {
	pool *pgxpool.Pool
}

// NewSetupRepo creates a new SetupRepo.
func NewSetupRepo(pool *pgxpool.Pool) *SetupRepo {
	return &SetupRepo{pool: pool}
}

// GetState retrieves the current setup state. Returns (nil, nil) if no setup
// row exists yet (i.e., setup has not been started).
func (r *SetupRepo) GetState(ctx context.Context) (*domain.SetupState, error) {
	query := `
		SELECT id, setup_token_hash, is_complete, completed_at, created_at
		FROM setup_state
		LIMIT 1`

	s := &domain.SetupState{}
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.ID, &s.TokenHash, &s.IsComplete, &s.CompletedAt, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("setup_repo: get_state: %w", err)
	}
	return s, nil
}

// CreateState inserts a new setup state row and scans back the generated fields.
func (r *SetupRepo) CreateState(ctx context.Context, state *domain.SetupState) error {
	query := `
		INSERT INTO setup_state (id, setup_token_hash, is_complete)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		state.ID, state.TokenHash, state.IsComplete,
	).Scan(&state.ID, &state.CreatedAt)
	if err != nil {
		return fmt.Errorf("setup_repo: create_state: %w", err)
	}
	return nil
}

// MarkComplete marks the setup as completed by setting is_complete and completed_at.
func (r *SetupRepo) MarkComplete(ctx context.Context, state *domain.SetupState) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE setup_state SET is_complete = true, completed_at = NOW() WHERE id = $1`,
		state.ID,
	)
	if err != nil {
		return fmt.Errorf("setup_repo: mark_complete: %w", err)
	}
	return nil
}
