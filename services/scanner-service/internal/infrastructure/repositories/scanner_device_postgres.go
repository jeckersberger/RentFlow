package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type ScannerDevicePostgres struct {
	db *database.PostgresPool
}

func NewScannerDevicePostgres(db *database.PostgresPool) ports.ScannerDeviceRepository {
	return &ScannerDevicePostgres{db: db}
}

func (r *ScannerDevicePostgres) Upsert(ctx context.Context, device *ports.ScannerDevice) error {
	query := `
		INSERT INTO scanner_devices (tenant_id, device_id, device_name, device_type, fcm_token, last_seen, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, device_id) DO UPDATE
		SET device_name = EXCLUDED.device_name,
		    fcm_token = EXCLUDED.fcm_token,
		    last_seen = EXCLUDED.last_seen
		RETURNING id, created_at
	`

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	err := r.db.QueryRow(ctx, query,
		device.TenantID,
		device.DeviceID,
		device.DeviceName,
		device.DeviceType,
		device.FCMToken,
		now,
		now,
	).Scan(&device.ID, &device.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert scanner device: %w", err)
	}

	device.LastSeen = &nowStr
	return nil
}

func (r *ScannerDevicePostgres) List(ctx context.Context, tenantID string) ([]*ports.ScannerDevice, error) {
	query := `
		SELECT id, tenant_id, device_id, device_name, device_type, fcm_token,
		       ring_requested, last_seen, created_at
		FROM scanner_devices
		WHERE tenant_id = $1
		ORDER BY last_seen DESC NULLS LAST
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list scanner devices: %w", err)
	}
	defer rows.Close()

	var devices []*ports.ScannerDevice
	for rows.Next() {
		d, err := scanScannerDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}

	return devices, rows.Err()
}

func (r *ScannerDevicePostgres) GetByDeviceID(ctx context.Context, tenantID, deviceID string) (*ports.ScannerDevice, error) {
	query := `
		SELECT id, tenant_id, device_id, device_name, device_type, fcm_token,
		       ring_requested, last_seen, created_at
		FROM scanner_devices
		WHERE tenant_id = $1 AND device_id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, deviceID)
	return scanScannerDeviceRow(row)
}

func (r *ScannerDevicePostgres) GetByID(ctx context.Context, id string) (*ports.ScannerDevice, error) {
	query := `
		SELECT id, tenant_id, device_id, device_name, device_type, fcm_token,
		       ring_requested, last_seen, created_at
		FROM scanner_devices
		WHERE id::text = $1
	`

	row := r.db.QueryRow(ctx, query, id)
	return scanScannerDeviceRow(row)
}

func (r *ScannerDevicePostgres) SetRingRequested(ctx context.Context, id string, requested bool) error {
	query := `UPDATE scanner_devices SET ring_requested = $1 WHERE id::text = $2`
	_, err := r.db.Exec(ctx, query, requested, id)
	return err
}

func (r *ScannerDevicePostgres) UpdateLastSeen(ctx context.Context, tenantID, deviceID string) error {
	query := `UPDATE scanner_devices SET last_seen = $1 WHERE tenant_id = $2 AND device_id = $3`
	_, err := r.db.Exec(ctx, query, time.Now(), tenantID, deviceID)
	return err
}

func scanScannerDevice(rows *sql.Rows) (*ports.ScannerDevice, error) {
	d := &ports.ScannerDevice{}
	var lastSeen sql.NullTime
	var createdAt time.Time
	err := rows.Scan(
		&d.ID, &d.TenantID, &d.DeviceID, &d.DeviceName, &d.DeviceType,
		&d.FCMToken, &d.RingRequested, &lastSeen, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	d.CreatedAt = createdAt.Format(time.RFC3339)
	if lastSeen.Valid {
		ls := lastSeen.Time.Format(time.RFC3339)
		d.LastSeen = &ls
	}
	return d, nil
}

func scanScannerDeviceRow(row *sql.Row) (*ports.ScannerDevice, error) {
	d := &ports.ScannerDevice{}
	var lastSeen sql.NullTime
	var createdAt time.Time
	err := row.Scan(
		&d.ID, &d.TenantID, &d.DeviceID, &d.DeviceName, &d.DeviceType,
		&d.FCMToken, &d.RingRequested, &lastSeen, &createdAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scanner device not found")
		}
		return nil, err
	}
	d.CreatedAt = createdAt.Format(time.RFC3339)
	if lastSeen.Valid {
		ls := lastSeen.Time.Format(time.RFC3339)
		d.LastSeen = &ls
	}
	return d, nil
}
