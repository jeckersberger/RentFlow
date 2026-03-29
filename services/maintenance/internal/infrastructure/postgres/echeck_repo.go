package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

const echeckColumns = `
	id, tenant_id, equipment_id, check_date, next_check_date, result,
	performed_by, measuring_device, protection_conductor_resistance,
	insulation_resistance, leakage_current, notes, created_at`

// ECheckRepo implements domain.ECheckRepository using PostgreSQL.
type ECheckRepo struct {
	pool *pgxpool.Pool
}

// NewECheckRepo creates a new ECheckRepo.
func NewECheckRepo(pool *pgxpool.Pool) *ECheckRepo {
	return &ECheckRepo{pool: pool}
}

// scanECheck scans a single echeck_records row into a domain.ECheckRecord.
func scanECheck(row pgx.Row) (*domain.ECheckRecord, error) {
	r := &domain.ECheckRecord{}
	var (
		nextCheckDate *string
		performedBy   *string
		measuringDev  *string
		notes         *string
	)

	err := row.Scan(
		&r.ID, &r.TenantID, &r.EquipmentID, &r.CheckDate, &nextCheckDate, &r.Result,
		&performedBy, &measuringDev, &r.ProtectionConductorResistance,
		&r.InsulationResistance, &r.LeakageCurrent, &notes, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	r.NextCheckDate = nextCheckDate
	if performedBy != nil {
		r.PerformedBy = *performedBy
	}
	if measuringDev != nil {
		r.MeasuringDevice = *measuringDev
	}
	if notes != nil {
		r.Notes = *notes
	}

	return r, nil
}

// Create inserts a new E-Check record.
func (repo *ECheckRepo) Create(ctx context.Context, record *domain.ECheckRecord) error {
	query := `
		INSERT INTO echeck_records (
			id, tenant_id, equipment_id, check_date, next_check_date, result,
			performed_by, measuring_device, protection_conductor_resistance,
			insulation_resistance, leakage_current, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at`

	err := repo.pool.QueryRow(ctx, query,
		record.ID, record.TenantID, record.EquipmentID,
		record.CheckDate, nilIfEmpty(ptrToStr(record.NextCheckDate)), record.Result,
		nilIfEmpty(record.PerformedBy), nilIfEmpty(record.MeasuringDevice),
		record.ProtectionConductorResistance, record.InsulationResistance,
		record.LeakageCurrent, nilIfEmpty(record.Notes),
	).Scan(&record.CreatedAt)
	if err != nil {
		return fmt.Errorf("echeck_repo: create: %w", err)
	}
	return nil
}

// List returns a filtered, paginated list of E-Check records for a tenant.
func (repo *ECheckRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ECheckFilter) ([]*domain.ECheckRecord, int64, error) {
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EquipmentID != nil {
		conditions = append(conditions, fmt.Sprintf("equipment_id = $%d", argIdx))
		args = append(args, *filter.EquipmentID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM echeck_records WHERE %s`, where)
	err := repo.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("echeck_repo: list count: %w", err)
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
		`SELECT %s FROM echeck_records WHERE %s ORDER BY check_date DESC LIMIT $%d OFFSET $%d`,
		echeckColumns, where, argIdx, argIdx+1,
	)
	args = append(args, perPage, offset)

	rows, err := repo.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("echeck_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ECheckRecord
	for rows.Next() {
		record, scanErr := scanECheck(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("echeck_repo: list scan: %w", scanErr)
		}
		items = append(items, record)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("echeck_repo: list rows: %w", err)
	}
	return items, total, nil
}

// ListOverdue returns all E-Check records where next_check_date is in the past.
func (repo *ECheckRepo) ListOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.ECheckRecord, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM echeck_records WHERE tenant_id = $1 AND next_check_date < NOW() ORDER BY next_check_date ASC`,
		echeckColumns,
	)

	rows, err := repo.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("echeck_repo: list_overdue query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ECheckRecord
	for rows.Next() {
		record, scanErr := scanECheck(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("echeck_repo: list_overdue scan: %w", scanErr)
		}
		items = append(items, record)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("echeck_repo: list_overdue rows: %w", err)
	}
	return items, nil
}

// ListByEquipment returns all E-Check records for a specific equipment item.
func (repo *ECheckRepo) ListByEquipment(ctx context.Context, equipmentID uuid.UUID, tenantID uuid.UUID) ([]*domain.ECheckRecord, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM echeck_records WHERE equipment_id = $1 AND tenant_id = $2 ORDER BY check_date DESC`,
		echeckColumns,
	)

	rows, err := repo.pool.Query(ctx, query, equipmentID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("echeck_repo: list_by_equipment query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ECheckRecord
	for rows.Next() {
		record, scanErr := scanECheck(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("echeck_repo: list_by_equipment scan: %w", scanErr)
		}
		items = append(items, record)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("echeck_repo: list_by_equipment rows: %w", err)
	}
	return items, nil
}

// ptrToStr safely converts a *string to a string, returning "" if nil.
func ptrToStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Ensure interface compliance at compile time.
var _ domain.ECheckRepository = (*ECheckRepo)(nil)
