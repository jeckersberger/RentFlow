package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// scanEventColumns lists all columns of the scan_events table for consistent scanning.
const scanEventColumns = `
	id, tenant_id, user_id, device_id, barcode, rfid_tag, equipment_id,
	action, project_id, location_id, condition_rating, condition_notes,
	gps_lat, gps_lng, timestamp, synced_at, created_at`

// ScanEventRepo implements domain.ScanEventRepository using PostgreSQL.
type ScanEventRepo struct {
	pool *pgxpool.Pool
}

// NewScanEventRepo creates a new ScanEventRepo.
func NewScanEventRepo(pool *pgxpool.Pool) *ScanEventRepo {
	return &ScanEventRepo{pool: pool}
}

// scanScanEvent scans a single scan_events row into a domain.ScanEvent.
func scanScanEvent(row pgx.Row) (*domain.ScanEvent, error) {
	e := &domain.ScanEvent{}
	var (
		deviceID       *string
		barcode        *string
		rfidTag        *string
		conditionNotes *string
	)

	err := row.Scan(
		&e.ID, &e.TenantID, &e.UserID, &deviceID, &barcode, &rfidTag, &e.EquipmentID,
		&e.Action, &e.ProjectID, &e.LocationID, &e.ConditionRating, &conditionNotes,
		&e.GPSLat, &e.GPSLng, &e.Timestamp, &e.SyncedAt, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if deviceID != nil {
		e.DeviceID = *deviceID
	}
	if barcode != nil {
		e.Barcode = *barcode
	}
	if rfidTag != nil {
		e.RFIDTag = *rfidTag
	}
	if conditionNotes != nil {
		e.ConditionNotes = *conditionNotes
	}

	return e, nil
}

// Create inserts a new scan event record.
func (r *ScanEventRepo) Create(ctx context.Context, event *domain.ScanEvent) error {
	query := `
		INSERT INTO scan_events (
			id, tenant_id, user_id, device_id, barcode, rfid_tag, equipment_id,
			action, project_id, location_id, condition_rating, condition_notes,
			gps_lat, gps_lng, timestamp, synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16
		) RETURNING id, synced_at, created_at`

	err := r.pool.QueryRow(ctx, query,
		event.ID, event.TenantID, event.UserID,
		nilIfEmpty(event.DeviceID), nilIfEmpty(event.Barcode), nilIfEmpty(event.RFIDTag),
		event.EquipmentID,
		event.Action, event.ProjectID, event.LocationID,
		event.ConditionRating, nilIfEmpty(event.ConditionNotes),
		event.GPSLat, event.GPSLng, event.Timestamp, event.SyncedAt,
	).Scan(&event.ID, &event.SyncedAt, &event.CreatedAt)
	if err != nil {
		return fmt.Errorf("scan_event_repo: create: %w", err)
	}
	return nil
}

// List returns a filtered, paginated list of scan events for a tenant plus total count.
func (r *ScanEventRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.ScanEventFilter) ([]*domain.ScanEvent, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.EquipmentID != nil {
		conditions = append(conditions, fmt.Sprintf("equipment_id = $%d", argIdx))
		args = append(args, *filter.EquipmentID)
		argIdx++
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, *filter.ProjectID)
		argIdx++
	}
	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, *filter.Action)
		argIdx++
	}
	if filter.DeviceID != nil {
		conditions = append(conditions, fmt.Sprintf("device_id = $%d", argIdx))
		args = append(args, *filter.DeviceID)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM scan_events WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("scan_event_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM scan_events WHERE %s ORDER BY timestamp DESC LIMIT $%d OFFSET $%d`,
		scanEventColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("scan_event_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ScanEvent
	for rows.Next() {
		e, scanErr := scanScanEvent(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan_event_repo: list scan: %w", scanErr)
		}
		items = append(items, e)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("scan_event_repo: list rows: %w", err)
	}
	return items, total, nil
}

// ExistsByDedup checks if a scan event with the same tenant_id, device_id, and timestamp already exists.
func (r *ScanEventRepo) ExistsByDedup(ctx context.Context, tenantID uuid.UUID, deviceID string, timestamp time.Time) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM scan_events WHERE tenant_id = $1 AND device_id = $2 AND timestamp = $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, tenantID, deviceID, timestamp).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("scan_event_repo: exists_by_dedup: %w", err)
	}
	return exists, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Ensure interface compliance at compile time.
var _ domain.ScanEventRepository = (*ScanEventRepo)(nil)

// Ensure apperrors is used (silence import).
var _ = apperrors.ErrNotFound

// Ensure errors and pgx are used.
var _ = errors.Is
var _ = pgx.ErrNoRows
