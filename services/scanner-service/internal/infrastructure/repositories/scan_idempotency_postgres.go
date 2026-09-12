package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
)

// CreateIfAbsent atomically creates a scan event by primary-key ID. This is
// used for retryable offline work where the same logical queue item must never
// materialize more than one scan event.
func (r *ScanEventPostgres) CreateIfAbsent(ctx context.Context, event *domain.ScanEvent) (bool, error) {
	query := `
		INSERT INTO scanner.scan_events
		(id, tenant_id, barcode, scan_type, equipment_id, project_id, location_id,
		 session_id, user_id, device_id, device_type, timestamp, latitude, longitude, notes, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (id) DO NOTHING
	`

	result, err := r.db.Exec(ctx, query,
		event.ID,
		event.TenantID,
		event.Barcode,
		string(event.ScanType),
		event.EquipmentID,
		event.ProjectID,
		event.LocationID,
		event.SessionID,
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
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}
