package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

const kpiColumns = `
	id, tenant_id, snapshot_date, active_projects, equipment_out_count,
	total_equipment, open_invoices_amount, overdue_invoices_amount,
	monthly_revenue, customer_count, utilization_pct, created_at`

// KPIRepo implements domain.KPISnapshotRepository using PostgreSQL.
type KPIRepo struct {
	pool *pgxpool.Pool
}

// NewKPIRepo creates a new KPIRepo.
func NewKPIRepo(pool *pgxpool.Pool) *KPIRepo {
	return &KPIRepo{pool: pool}
}

func scanKPI(row pgx.Row) (*domain.KPISnapshot, error) {
	snap := &domain.KPISnapshot{}
	var snapshotDate time.Time

	err := row.Scan(
		&snap.ID, &snap.TenantID, &snapshotDate,
		&snap.ActiveProjects, &snap.EquipmentOutCount,
		&snap.TotalEquipment, &snap.OpenInvoicesAmount, &snap.OverdueInvoicesAmount,
		&snap.MonthlyRevenue, &snap.CustomerCount, &snap.UtilizationPct,
		&snap.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	snap.SnapshotDate = snapshotDate.Format("2006-01-02")
	return snap, nil
}

// Upsert creates or updates a KPI snapshot for the given tenant and date.
func (r *KPIRepo) Upsert(ctx context.Context, snap *domain.KPISnapshot) error {
	query := `
		INSERT INTO kpi_snapshots (
			id, tenant_id, snapshot_date, active_projects, equipment_out_count,
			total_equipment, open_invoices_amount, overdue_invoices_amount,
			monthly_revenue, customer_count, utilization_pct
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, snapshot_date)
		DO UPDATE SET
			active_projects = EXCLUDED.active_projects,
			equipment_out_count = EXCLUDED.equipment_out_count,
			total_equipment = EXCLUDED.total_equipment,
			open_invoices_amount = EXCLUDED.open_invoices_amount,
			overdue_invoices_amount = EXCLUDED.overdue_invoices_amount,
			monthly_revenue = EXCLUDED.monthly_revenue,
			customer_count = EXCLUDED.customer_count,
			utilization_pct = EXCLUDED.utilization_pct
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		snap.ID, snap.TenantID, snap.SnapshotDate,
		snap.ActiveProjects, snap.EquipmentOutCount,
		snap.TotalEquipment, snap.OpenInvoicesAmount, snap.OverdueInvoicesAmount,
		snap.MonthlyRevenue, snap.CustomerCount, snap.UtilizationPct,
	).Scan(&snap.ID, &snap.CreatedAt)
	if err != nil {
		return fmt.Errorf("kpi_repo: upsert: %w", err)
	}
	return nil
}

// GetLatest returns the most recent KPI snapshot for a tenant.
func (r *KPIRepo) GetLatest(ctx context.Context, tenantID uuid.UUID) (*domain.KPISnapshot, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM kpi_snapshots WHERE tenant_id = $1 ORDER BY snapshot_date DESC LIMIT 1`,
		kpiColumns,
	)
	snap, err := scanKPI(r.pool.QueryRow(ctx, query, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("kpi_repo: get_latest: %w", err)
	}
	return snap, nil
}

// GetHistory returns all KPI snapshots for a tenant within the given date range.
func (r *KPIRepo) GetHistory(ctx context.Context, tenantID uuid.UUID, from, to string) ([]*domain.KPISnapshot, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM kpi_snapshots
		 WHERE tenant_id = $1 AND snapshot_date >= $2 AND snapshot_date <= $3
		 ORDER BY snapshot_date ASC`,
		kpiColumns,
	)

	rows, err := r.pool.Query(ctx, query, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("kpi_repo: get_history query: %w", err)
	}
	defer rows.Close()

	var items []*domain.KPISnapshot
	for rows.Next() {
		snap, scanErr := scanKPI(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("kpi_repo: get_history scan: %w", scanErr)
		}
		items = append(items, snap)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("kpi_repo: get_history rows: %w", err)
	}
	return items, nil
}

// Ensure interface compliance at compile time.
var _ domain.KPISnapshotRepository = (*KPIRepo)(nil)
