package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type ScanEventPostgres struct {
	db *database.PostgresPool
}

func NewScanEventPostgres(db *database.PostgresPool) *ScanEventPostgres {
	return &ScanEventPostgres{db: db}
}

func (r *ScanEventPostgres) Create(ctx context.Context, event *domain.ScanEvent) error {
	query := `
		INSERT INTO scan_events
		(id, tenant_id, barcode, scan_type, equipment_id, project_id, location_id,
		 user_id, device_id, device_type, timestamp, latitude, longitude, notes, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(ctx, query,
		event.ID,
		event.TenantID,
		event.Barcode,
		string(event.ScanType),
		event.EquipmentID,
		event.ProjectID,
		event.LocationID,
		event.UserID,
		event.DeviceID,
		string(event.DeviceType),
		event.Timestamp,
		event.Latitude,
		event.Longitude,
		event.Notes,
		string(event.Status),
		event.CreatedAt,
	)

	return err
}

func (r *ScanEventPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.ScanEvent, error) {
	query := `
		SELECT id, tenant_id, barcode, scan_type, equipment_id, project_id, location_id,
		       user_id, device_id, device_type, timestamp, latitude, longitude, notes, status, created_at
		FROM scan_events
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return scanRowToEvent(row)
}

func (r *ScanEventPostgres) GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.ScanEvent, error) {
	query := `
		SELECT id, tenant_id, barcode, scan_type, equipment_id, project_id, location_id,
		       user_id, device_id, device_type, timestamp, latitude, longitude, notes, status, created_at
		FROM scan_events
		WHERE tenant_id = $1 AND barcode = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRow(ctx, query, tenantID, barcode)
	return scanRowToEvent(row)
}

func (r *ScanEventPostgres) List(ctx context.Context, tenantID string, query *ports.ScanListQuery) (*ports.ScanListResult, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIndex := 2

	if query.EquipmentID != nil {
		where += fmt.Sprintf(" AND equipment_id = $%d", argIndex)
		args = append(args, *query.EquipmentID)
		argIndex++
	}
	if query.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIndex)
		args = append(args, *query.ProjectID)
		argIndex++
	}
	if query.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *query.UserID)
		argIndex++
	}
	if query.DeviceID != nil {
		where += fmt.Sprintf(" AND device_id = $%d", argIndex)
		args = append(args, *query.DeviceID)
		argIndex++
	}
	if query.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *query.Status)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM scan_events %s", where)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, barcode, scan_type, equipment_id, project_id, location_id,
		       user_id, device_id, device_type, timestamp, latitude, longitude, notes, status, created_at
		FROM scan_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIndex, argIndex+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.ScanEvent, 0)
	for rows.Next() {
		event, err := scanRowsToEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, event)
	}

	return &ports.ScanListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, rows.Err()
}

func (r *ScanEventPostgres) Update(ctx context.Context, event *domain.ScanEvent) error {
	query := `
		UPDATE scan_events
		SET equipment_id = $1, project_id = $2, location_id = $3,
		    notes = $4, status = $5
		WHERE id = $6 AND tenant_id = $7
	`

	_, err := r.db.Exec(ctx, query,
		event.EquipmentID,
		event.ProjectID,
		event.LocationID,
		event.Notes,
		string(event.Status),
		event.ID,
		event.TenantID,
	)

	return err
}

func scanRowToEvent(row *sql.Row) (*domain.ScanEvent, error) {
	event := &domain.ScanEvent{}
	err := row.Scan(
		&event.ID,
		&event.TenantID,
		&event.Barcode,
		(*string)(&event.ScanType),
		&event.EquipmentID,
		&event.ProjectID,
		&event.LocationID,
		&event.UserID,
		&event.DeviceID,
		(*string)(&event.DeviceType),
		&event.Timestamp,
		&event.Latitude,
		&event.Longitude,
		&event.Notes,
		(*string)(&event.Status),
		&event.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan event not found")
		}
		return nil, err
	}
	return event, nil
}

func scanRowsToEvent(rows *sql.Rows) (*domain.ScanEvent, error) {
	event := &domain.ScanEvent{}
	err := rows.Scan(
		&event.ID,
		&event.TenantID,
		&event.Barcode,
		(*string)(&event.ScanType),
		&event.EquipmentID,
		&event.ProjectID,
		&event.LocationID,
		&event.UserID,
		&event.DeviceID,
		(*string)(&event.DeviceType),
		&event.Timestamp,
		&event.Latitude,
		&event.Longitude,
		&event.Notes,
		(*string)(&event.Status),
		&event.CreatedAt,
	)
	return event, err
}
