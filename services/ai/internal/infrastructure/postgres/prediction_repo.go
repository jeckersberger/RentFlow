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

const predictionColumns = `
	id, tenant_id, type, reference_id, reference_type,
	prediction, confidence, model_version, created_at`

type PredictionRepo struct {
	pool *pgxpool.Pool
}

func NewPredictionRepo(pool *pgxpool.Pool) *PredictionRepo {
	return &PredictionRepo{pool: pool}
}

func scanPrediction(row pgx.Row) (*domain.AIPrediction, error) {
	pred := &domain.AIPrediction{}
	var (
		prediction    []byte
		referenceID   *uuid.UUID
		referenceType *string
		modelVersion  *string
	)

	err := row.Scan(
		&pred.ID, &pred.TenantID, &pred.Type, &referenceID, &referenceType,
		&prediction, &pred.Confidence, &modelVersion, &pred.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if prediction != nil {
		pred.Prediction = json.RawMessage(prediction)
	} else {
		pred.Prediction = json.RawMessage("{}")
	}
	pred.ReferenceID = referenceID
	if referenceType != nil {
		pred.ReferenceType = *referenceType
	}
	if modelVersion != nil {
		pred.ModelVersion = *modelVersion
	}

	return pred, nil
}

func (r *PredictionRepo) Create(ctx context.Context, pred *domain.AIPrediction) error {
	query := `
		INSERT INTO ai_predictions (
			id, tenant_id, type, reference_id, reference_type,
			prediction, confidence, model_version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		pred.ID, pred.TenantID, pred.Type, pred.ReferenceID,
		nilIfEmpty(pred.ReferenceType), pred.Prediction,
		pred.Confidence, nilIfEmpty(pred.ModelVersion),
	).Scan(&pred.CreatedAt)
	if err != nil {
		return fmt.Errorf("prediction_repo: create: %w", err)
	}
	return nil
}

func (r *PredictionRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.AIPrediction, error) {
	query := fmt.Sprintf(`SELECT %s FROM ai_predictions WHERE id = $1 AND tenant_id = $2`, predictionColumns)
	pred, err := scanPrediction(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("prediction_repo: get_by_id: %w", err)
	}
	return pred, nil
}

func (r *PredictionRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.PredictionFilter) ([]*domain.AIPrediction, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM ai_predictions WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("prediction_repo: list count: %w", err)
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
		`SELECT %s FROM ai_predictions WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		predictionColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("prediction_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AIPrediction
	for rows.Next() {
		pred, scanErr := scanPrediction(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("prediction_repo: list scan: %w", scanErr)
		}
		items = append(items, pred)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("prediction_repo: list rows: %w", err)
	}
	return items, total, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
