package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/ai/internal/domain"
)

const trainingDataColumns = `id, tenant_id, type, input_data, output_data, feedback, created_at`

type TrainingDataRepo struct {
	pool *pgxpool.Pool
}

func NewTrainingDataRepo(pool *pgxpool.Pool) *TrainingDataRepo {
	return &TrainingDataRepo{pool: pool}
}

func scanTrainingData(row pgx.Row) (*domain.AITrainingData, error) {
	td := &domain.AITrainingData{}
	var (
		inputData  []byte
		outputData []byte
		feedback   *string
	)
	err := row.Scan(&td.ID, &td.TenantID, &td.Type, &inputData, &outputData, &feedback, &td.CreatedAt)
	if err != nil {
		return nil, err
	}
	if inputData != nil {
		td.InputData = json.RawMessage(inputData)
	} else {
		td.InputData = json.RawMessage("{}")
	}
	if outputData != nil {
		td.OutputData = json.RawMessage(outputData)
	} else {
		td.OutputData = json.RawMessage("{}")
	}
	if feedback != nil {
		td.Feedback = *feedback
	}
	return td, nil
}

func (r *TrainingDataRepo) Create(ctx context.Context, td *domain.AITrainingData) error {
	query := `
		INSERT INTO ai_training_data (
			id, tenant_id, type, input_data, output_data, feedback
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		td.ID, td.TenantID, td.Type, td.InputData,
		td.OutputData, nilIfEmpty(td.Feedback),
	).Scan(&td.CreatedAt)
	if err != nil {
		return fmt.Errorf("training_data_repo: create: %w", err)
	}
	return nil
}

func (r *TrainingDataRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.TrainingDataFilter) ([]*domain.AITrainingData, int64, error) {
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
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM ai_training_data WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("training_data_repo: list count: %w", err)
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
		`SELECT %s FROM ai_training_data WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		trainingDataColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("training_data_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.AITrainingData
	for rows.Next() {
		td, scanErr := scanTrainingData(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("training_data_repo: list scan: %w", scanErr)
		}
		items = append(items, td)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("training_data_repo: list rows: %w", err)
	}
	return items, total, nil
}
