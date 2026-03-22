package repositories

import (
	"context"
	"database/sql"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/domain"
)

// PostgresFewShotExampleRepository implements FewShotExampleRepository
type PostgresFewShotExampleRepository struct {
	db     *sql.DB
	logger logger.Logger
}

// NewPostgresFewShotExampleRepository creates a new few-shot example repository
func NewPostgresFewShotExampleRepository(db *sql.DB, log logger.Logger) *PostgresFewShotExampleRepository {
	return &PostgresFewShotExampleRepository{
		db:     db,
		logger: log,
	}
}

// FindByID retrieves an example by ID
func (r *PostgresFewShotExampleRepository) FindByID(ctx context.Context, id string) (*domain.FewShotExample, error) {
	query := `
		SELECT id, tenant_id, request_type, input_example, output_example,
		       is_active, usage_count, avg_rating, created_at
		FROM few_shot_examples
		WHERE id = $1
	`

	var example domain.FewShotExample
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&example.ID, &example.TenantID, &example.RequestType, &example.InputExample,
		&example.OutputExample, &example.IsActive, &example.UsageCount,
		&example.AvgRating, &example.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query few-shot example", err)
		return nil, err
	}

	return &example, nil
}

// ListByType retrieves examples for a request type
func (r *PostgresFewShotExampleRepository) ListByType(ctx context.Context, requestType domain.RequestType) ([]*domain.FewShotExample, error) {
	query := `
		SELECT id, tenant_id, request_type, input_example, output_example,
		       is_active, usage_count, avg_rating, created_at
		FROM few_shot_examples
		WHERE request_type = $1 AND is_active = true
		ORDER BY usage_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query, requestType)
	if err != nil {
		r.logger.Error("failed to query examples by type", err)
		return nil, err
	}
	defer rows.Close()

	var examples []*domain.FewShotExample
	for rows.Next() {
		var ex domain.FewShotExample
		if err := rows.Scan(
			&ex.ID, &ex.TenantID, &ex.RequestType, &ex.InputExample,
			&ex.OutputExample, &ex.IsActive, &ex.UsageCount,
			&ex.AvgRating, &ex.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan example", err)
			return nil, err
		}
		examples = append(examples, &ex)
	}

	return examples, rows.Err()
}

// List retrieves all examples with pagination
func (r *PostgresFewShotExampleRepository) List(ctx context.Context, tenantID string, page, perPage int) ([]*domain.FewShotExample, int, error) {
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM few_shot_examples WHERE tenant_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		r.logger.Error("failed to count examples", err)
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, request_type, input_example, output_example,
		       is_active, usage_count, avg_rating, created_at
		FROM few_shot_examples
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, perPage, offset)
	if err != nil {
		r.logger.Error("failed to query examples", err)
		return nil, 0, err
	}
	defer rows.Close()

	var examples []*domain.FewShotExample
	for rows.Next() {
		var ex domain.FewShotExample
		if err := rows.Scan(
			&ex.ID, &ex.TenantID, &ex.RequestType, &ex.InputExample,
			&ex.OutputExample, &ex.IsActive, &ex.UsageCount,
			&ex.AvgRating, &ex.CreatedAt,
		); err != nil {
			r.logger.Error("failed to scan example", err)
			return nil, 0, err
		}
		examples = append(examples, &ex)
	}

	return examples, total, rows.Err()
}

// Save persists an example
func (r *PostgresFewShotExampleRepository) Save(ctx context.Context, example *domain.FewShotExample) error {
	query := `
		INSERT INTO few_shot_examples 
		(id, tenant_id, request_type, input_example, output_example, is_active,
		 usage_count, avg_rating, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
		input_example = $4, output_example = $5, is_active = $6,
		usage_count = $7, avg_rating = $8
	`

	_, err := r.db.ExecContext(ctx, query,
		example.ID, example.TenantID, example.RequestType, example.InputExample,
		example.OutputExample, example.IsActive, example.UsageCount,
		example.AvgRating, example.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save few-shot example", err)
		return err
	}

	return nil
}

// Delete deletes an example
func (r *PostgresFewShotExampleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM few_shot_examples WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete example", err)
		return err
	}
	return nil
}

// IncrementUsageCount increments usage count
func (r *PostgresFewShotExampleRepository) IncrementUsageCount(ctx context.Context, id string) error {
	query := `UPDATE few_shot_examples SET usage_count = usage_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to increment usage count", err)
		return err
	}
	return nil
}
