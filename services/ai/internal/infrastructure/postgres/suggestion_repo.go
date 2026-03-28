package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

const suggestionColumns = `
	id, tenant_id, type, title, description, data,
	status, accepted_by, accepted_at, created_at`

type SuggestionRepo struct {
	pool *pgxpool.Pool
}

func NewSuggestionRepo(pool *pgxpool.Pool) *SuggestionRepo {
	return &SuggestionRepo{pool: pool}
}

func scanSuggestion(row pgx.Row) (*domain.AISuggestion, error) {
	sug := &domain.AISuggestion{}
	var (
		data        []byte
		description *string
		acceptedBy  *uuid.UUID
	)

	err := row.Scan(
		&sug.ID, &sug.TenantID, &sug.Type, &sug.Title, &description, &data,
		&sug.Status, &acceptedBy, &sug.AcceptedAt, &sug.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if data != nil {
		sug.Data = json.RawMessage(data)
	} else {
		sug.Data = json.RawMessage("{}")
	}
	if description != nil {
		sug.Description = *description
	}
	sug.AcceptedBy = acceptedBy

	return sug, nil
}

func (r *SuggestionRepo) Create(ctx context.Context, sug *domain.AISuggestion) error {
	query := `
		INSERT INTO ai_suggestions (
			id, tenant_id, type, title, description, data, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		sug.ID, sug.TenantID, sug.Type, sug.Title,
		nilIfEmpty(sug.Description), sug.Data, sug.Status,
	).Scan(&sug.CreatedAt)
	if err != nil {
		return fmt.Errorf("suggestion_repo: create: %w", err)
	}
	return nil
}

func (r *SuggestionRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.AISuggestion, error) {
	query := fmt.Sprintf(`SELECT %s FROM ai_suggestions WHERE id = $1 AND tenant_id = $2`, suggestionColumns)
	sug, err := scanSuggestion(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("suggestion_repo: get_by_id: %w", err)
	}
	return sug, nil
}

func (r *SuggestionRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.SuggestionFilter) ([]*domain.AISuggestion, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM ai_suggestions WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("suggestion_repo: list count: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM ai_suggestions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		suggestionColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("suggestion_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AISuggestion
	for rows.Next() {
		sug, scanErr := scanSuggestion(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("suggestion_repo: list scan: %w", scanErr)
		}
		items = append(items, sug)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("suggestion_repo: list rows: %w", err)
	}
	return items, total, nil
}

func (r *SuggestionRepo) Update(ctx context.Context, sug *domain.AISuggestion) error {
	query := `
		UPDATE ai_suggestions SET
			status = $3, accepted_by = $4, accepted_at = $5
		WHERE id = $1 AND tenant_id = $2
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		sug.ID, sug.TenantID,
		sug.Status, sug.AcceptedBy, sug.AcceptedAt,
	).Scan(&sug.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("suggestion_repo: update: %w", err)
	}
	return nil
}
