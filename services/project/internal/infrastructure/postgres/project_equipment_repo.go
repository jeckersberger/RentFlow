package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// projectEquipmentColumns lists all columns of the project_equipment table.
const projectEquipmentColumns = `
	id, project_id, equipment_id, quantity,
	allocated_from, allocated_until, status, notes, created_at`

// ProjectEquipmentRepo implements domain.ProjectEquipmentRepository using PostgreSQL.
type ProjectEquipmentRepo struct {
	pool *pgxpool.Pool
}

// NewProjectEquipmentRepo creates a new ProjectEquipmentRepo.
func NewProjectEquipmentRepo(pool *pgxpool.Pool) *ProjectEquipmentRepo {
	return &ProjectEquipmentRepo{pool: pool}
}

// scanProjectEquipment scans a single project_equipment row into a domain.ProjectEquipment.
func scanProjectEquipment(row pgx.Row) (*domain.ProjectEquipment, error) {
	pe := &domain.ProjectEquipment{}
	var (
		allocatedFrom  *time.Time
		allocatedUntil *time.Time
		status         *string
		notes          *string
	)

	err := row.Scan(
		&pe.ID, &pe.ProjectID, &pe.EquipmentID, &pe.Quantity,
		&allocatedFrom, &allocatedUntil, &status, &notes, &pe.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	pe.AllocatedFrom = allocatedFrom
	pe.AllocatedUntil = allocatedUntil
	pe.Status = derefString(status)
	pe.Notes = derefString(notes)

	return pe, nil
}

// Add inserts a new project equipment allocation.
func (r *ProjectEquipmentRepo) Add(ctx context.Context, pe *domain.ProjectEquipment) error {
	query := `
		INSERT INTO project_equipment (
			id, project_id, equipment_id, quantity,
			allocated_from, allocated_until, status, notes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING created_at`

	err := r.pool.QueryRow(ctx, query,
		pe.ID, pe.ProjectID, pe.EquipmentID, pe.Quantity,
		pe.AllocatedFrom, pe.AllocatedUntil,
		nilIfEmpty(pe.Status), nilIfEmpty(pe.Notes),
	).Scan(&pe.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("project_equipment_repo: add: %w", err)
	}
	return nil
}

// Remove deletes a project equipment allocation by ID.
func (r *ProjectEquipmentRepo) Remove(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM project_equipment WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("project_equipment_repo: remove: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// ListByProject returns all equipment allocations for a given project, ordered by created_at.
func (r *ProjectEquipmentRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectEquipment, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM project_equipment WHERE project_id = $1 ORDER BY created_at ASC`,
		projectEquipmentColumns,
	)

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_equipment_repo: list_by_project query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ProjectEquipment
	for rows.Next() {
		pe, scanErr := scanProjectEquipment(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("project_equipment_repo: list_by_project scan: %w", scanErr)
		}
		items = append(items, pe)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("project_equipment_repo: list_by_project rows: %w", err)
	}
	return items, nil
}

// GetByID retrieves a single project equipment allocation by primary key.
func (r *ProjectEquipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectEquipment, error) {
	query := fmt.Sprintf(`SELECT %s FROM project_equipment WHERE id = $1`, projectEquipmentColumns)
	pe, err := scanProjectEquipment(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("project_equipment_repo: get_by_id: %w", err)
	}
	return pe, nil
}
