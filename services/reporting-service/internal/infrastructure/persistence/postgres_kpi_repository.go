package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type PostgresKPIRepository struct {
	db *sql.DB
}

func NewPostgresKPIRepository(db *sql.DB) *PostgresKPIRepository {
	return &PostgresKPIRepository{db: db}
}

func (r *PostgresKPIRepository) CreateSnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error {
	query := `
		INSERT INTO kpi_snapshots (
			id, tenant_id, snapshot_date, kpi_type, value, previous_value,
			change_percentage, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id, kpi_type, snapshot_date) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		snapshot.ID, snapshot.TenantID, snapshot.SnapshotDate, snapshot.KPIType,
		snapshot.Value, snapshot.PreviousValue, snapshot.ChangePercentage,
		snapshot.Metadata, snapshot.CreatedAt,
	)

	return err
}

func (r *PostgresKPIRepository) GetLatestSnapshot(ctx context.Context, tenantID uuid.UUID, kpiType domain.KPIType) (*domain.KPISnapshot, error) {
	query := `
		SELECT id, tenant_id, snapshot_date, kpi_type, value, previous_value,
			   change_percentage, metadata, created_at
		FROM kpi_snapshots
		WHERE tenant_id = $1 AND kpi_type = $2
		ORDER BY snapshot_date DESC
		LIMIT 1
	`

	snapshot := &domain.KPISnapshot{}
	err := r.db.QueryRowContext(ctx, query, tenantID, kpiType).Scan(
		&snapshot.ID, &snapshot.TenantID, &snapshot.SnapshotDate, &snapshot.KPIType,
		&snapshot.Value, &snapshot.PreviousValue, &snapshot.ChangePercentage,
		&snapshot.Metadata, &snapshot.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (r *PostgresKPIRepository) GetSnapshotsByDateRange(ctx context.Context, tenantID uuid.UUID, kpiType domain.KPIType, startDate, endDate time.Time) ([]domain.KPISnapshot, error) {
	query := `
		SELECT id, tenant_id, snapshot_date, kpi_type, value, previous_value,
			   change_percentage, metadata, created_at
		FROM kpi_snapshots
		WHERE tenant_id = $1 AND kpi_type = $2 AND snapshot_date >= $3 AND snapshot_date <= $4
		ORDER BY snapshot_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, kpiType, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []domain.KPISnapshot
	for rows.Next() {
		snapshot := domain.KPISnapshot{}
		err := rows.Scan(
			&snapshot.ID, &snapshot.TenantID, &snapshot.SnapshotDate, &snapshot.KPIType,
			&snapshot.Value, &snapshot.PreviousValue, &snapshot.ChangePercentage,
			&snapshot.Metadata, &snapshot.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, rows.Err()
}
