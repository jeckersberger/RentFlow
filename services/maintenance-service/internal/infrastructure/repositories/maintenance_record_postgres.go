package repositories

import (
	"context"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/maintenance-service/internal/domain"
)

type MaintenanceRecordPostgres struct {
	db *database.PostgresPool
}

func NewMaintenanceRecordPostgres(db *database.PostgresPool) *MaintenanceRecordPostgres {
	return &MaintenanceRecordPostgres{db: db}
}

func (r *MaintenanceRecordPostgres) Create(ctx context.Context, record *domain.MaintenanceRecord) error {
	query := `
		INSERT INTO maintenance_records (id, tenant_id, equipment_id, type, status, scheduled_date, technician, cost, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		record.ID, record.TenantID, record.EquipmentID, record.Type, record.Status,
		record.ScheduledDate, record.Technician, record.Cost, record.Notes, record.CreatedAt, record.UpdatedAt,
	)
	return err
}

func (r *MaintenanceRecordPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.MaintenanceRecord, error) {
	query := `
		SELECT id, tenant_id, equipment_id, type, status, scheduled_date, completed_date, technician, cost, notes, certificate_ref, created_at, updated_at
		FROM maintenance_records WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	record := &domain.MaintenanceRecord{}
	err := row.Scan(&record.ID, &record.TenantID, &record.EquipmentID, &record.Type, &record.Status,
		&record.ScheduledDate, &record.CompletedDate, &record.Technician, &record.Cost, &record.Notes, &record.CertificateRef, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *MaintenanceRecordPostgres) ListByTenant(ctx context.Context, tenantID string) ([]*domain.MaintenanceRecord, error) {
	query := `
		SELECT id, tenant_id, equipment_id, type, status, scheduled_date, completed_date, technician, cost, notes, certificate_ref, created_at, updated_at
		FROM maintenance_records WHERE tenant_id = $1 ORDER BY scheduled_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.MaintenanceRecord
	for rows.Next() {
		record := &domain.MaintenanceRecord{}
		err := rows.Scan(&record.ID, &record.TenantID, &record.EquipmentID, &record.Type, &record.Status,
			&record.ScheduledDate, &record.CompletedDate, &record.Technician, &record.Cost, &record.Notes, &record.CertificateRef, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *MaintenanceRecordPostgres) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.MaintenanceRecord, error) {
	query := `
		SELECT id, tenant_id, equipment_id, type, status, scheduled_date, completed_date, technician, cost, notes, certificate_ref, created_at, updated_at
		FROM maintenance_records WHERE tenant_id = $1 AND equipment_id = $2 ORDER BY scheduled_date DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.MaintenanceRecord
	for rows.Next() {
		record := &domain.MaintenanceRecord{}
		err := rows.Scan(&record.ID, &record.TenantID, &record.EquipmentID, &record.Type, &record.Status,
			&record.ScheduledDate, &record.CompletedDate, &record.Technician, &record.Cost, &record.Notes, &record.CertificateRef, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *MaintenanceRecordPostgres) ListOverdue(ctx context.Context, tenantID string) ([]*domain.MaintenanceRecord, error) {
	query := `
		SELECT id, tenant_id, equipment_id, type, status, scheduled_date, completed_date, technician, cost, notes, certificate_ref, created_at, updated_at
		FROM maintenance_records WHERE tenant_id = $1 AND status != $2 AND scheduled_date < $3 ORDER BY scheduled_date ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID, domain.MaintenanceStatusCompleted, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*domain.MaintenanceRecord
	for rows.Next() {
		record := &domain.MaintenanceRecord{}
		err := rows.Scan(&record.ID, &record.TenantID, &record.EquipmentID, &record.Type, &record.Status,
			&record.ScheduledDate, &record.CompletedDate, &record.Technician, &record.Cost, &record.Notes, &record.CertificateRef, &record.CreatedAt, &record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *MaintenanceRecordPostgres) Update(ctx context.Context, record *domain.MaintenanceRecord) error {
	query := `
		UPDATE maintenance_records
		SET status = $1, completed_date = $2, notes = $3, certificate_ref = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7
	`
	_, err := r.db.Exec(ctx, query,
		record.Status, record.CompletedDate, record.Notes, record.CertificateRef, record.UpdatedAt, record.ID, record.TenantID,
	)
	return err
}

func (r *MaintenanceRecordPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM maintenance_records WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}
