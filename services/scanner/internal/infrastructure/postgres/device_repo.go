package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/scanner/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// deviceColumns lists all columns of the scanner_devices table for consistent scanning.
const deviceColumns = `
	id, tenant_id, device_id, device_name, device_type, fcm_token,
	ring_requested, last_seen, created_at`

// DeviceRepo implements domain.ScannerDeviceRepository using PostgreSQL.
type DeviceRepo struct {
	pool *pgxpool.Pool
}

// NewDeviceRepo creates a new DeviceRepo.
func NewDeviceRepo(pool *pgxpool.Pool) *DeviceRepo {
	return &DeviceRepo{pool: pool}
}

// scanDevice scans a single scanner_devices row into a domain.ScannerDevice.
func scanDevice(row pgx.Row) (*domain.ScannerDevice, error) {
	d := &domain.ScannerDevice{}
	var (
		deviceName *string
		deviceType *string
		fcmToken   *string
	)

	err := row.Scan(
		&d.ID, &d.TenantID, &d.DeviceID, &deviceName, &deviceType, &fcmToken,
		&d.RingRequested, &d.LastSeen, &d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if deviceName != nil {
		d.DeviceName = *deviceName
	}
	if deviceType != nil {
		d.DeviceType = *deviceType
	}
	if fcmToken != nil {
		d.FCMToken = *fcmToken
	}

	return d, nil
}

// Upsert inserts or updates a scanner device (on conflict by device_id + tenant_id).
func (r *DeviceRepo) Upsert(ctx context.Context, device *domain.ScannerDevice) error {
	query := `
		INSERT INTO scanner_devices (
			id, tenant_id, device_id, device_name, device_type, fcm_token, last_seen
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (device_id, tenant_id) DO UPDATE SET
			device_name = EXCLUDED.device_name,
			device_type = EXCLUDED.device_type,
			fcm_token = EXCLUDED.fcm_token,
			last_seen = EXCLUDED.last_seen
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		device.ID, device.TenantID, device.DeviceID,
		nilIfEmpty(device.DeviceName), nilIfEmpty(device.DeviceType),
		nilIfEmpty(device.FCMToken), device.LastSeen,
	).Scan(&device.ID, &device.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateDevice
		}
		return fmt.Errorf("device_repo: upsert: %w", err)
	}
	return nil
}

// List returns all scanner devices for a tenant.
func (r *DeviceRepo) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.ScannerDevice, error) {
	query := fmt.Sprintf(`SELECT %s FROM scanner_devices WHERE tenant_id = $1 ORDER BY last_seen DESC NULLS LAST`, deviceColumns)

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("device_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.ScannerDevice
	for rows.Next() {
		d, scanErr := scanDevice(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("device_repo: list scan: %w", scanErr)
		}
		items = append(items, d)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("device_repo: list rows: %w", err)
	}
	return items, nil
}

// GetByDeviceID retrieves a scanner device by its device_id within a tenant scope.
func (r *DeviceRepo) GetByDeviceID(ctx context.Context, tenantID uuid.UUID, deviceID string) (*domain.ScannerDevice, error) {
	query := fmt.Sprintf(`SELECT %s FROM scanner_devices WHERE tenant_id = $1 AND device_id = $2`, deviceColumns)
	d, err := scanDevice(r.pool.QueryRow(ctx, query, tenantID, deviceID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("device_repo: get_by_device_id: %w", err)
	}
	return d, nil
}

// GetByID retrieves a scanner device by its primary key within a tenant scope.
func (r *DeviceRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.ScannerDevice, error) {
	query := fmt.Sprintf(`SELECT %s FROM scanner_devices WHERE id = $1 AND tenant_id = $2`, deviceColumns)
	d, err := scanDevice(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("device_repo: get_by_id: %w", err)
	}
	return d, nil
}

// UpdateLastSeen updates the last_seen timestamp for a device.
func (r *DeviceRepo) UpdateLastSeen(ctx context.Context, tenantID uuid.UUID, deviceID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scanner_devices SET last_seen = NOW() WHERE tenant_id = $1 AND device_id = $2`,
		tenantID, deviceID,
	)
	if err != nil {
		return fmt.Errorf("device_repo: update_last_seen: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// SetRingRequested sets the ring_requested flag for a device.
func (r *DeviceRepo) SetRingRequested(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, ring bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE scanner_devices SET ring_requested = $3 WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, ring,
	)
	if err != nil {
		return fmt.Errorf("device_repo: set_ring_requested: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ domain.ScannerDeviceRepository = (*DeviceRepo)(nil)
