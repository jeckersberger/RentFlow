package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/scanner-service/internal/ports"
)

type DevicePostgres struct {
	db *database.PostgresPool
}

func NewDevicePostgres(db *database.PostgresPool) ports.DeviceRepository {
	return &DevicePostgres{db: db}
}

func (r *DevicePostgres) Create(ctx context.Context, device *domain.Device) error {
	query := `
		INSERT INTO scanner.devices
		(id, tenant_id, name, type, serial, active, location, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		device.ID,
		device.TenantID,
		device.Name,
		string(device.Type),
		device.Serial,
		device.Active,
		device.Location,
		device.CreatedAt,
		device.UpdatedAt,
	)

	return err
}

func (r *DevicePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Device, error) {
	query := `
		SELECT id, tenant_id, name, type, serial, active, location, created_at, updated_at
		FROM scanner.devices
		WHERE tenant_id = $1 AND id = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, id)
	return deviceRowToDevice(row)
}

func (r *DevicePostgres) GetBySerial(ctx context.Context, tenantID, serial string) (*domain.Device, error) {
	query := `
		SELECT id, tenant_id, name, type, serial, active, location, created_at, updated_at
		FROM scanner.devices
		WHERE tenant_id = $1 AND serial = $2
	`

	row := r.db.QueryRow(ctx, query, tenantID, serial)
	return deviceRowToDevice(row)
}

func (r *DevicePostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Device, int, error) {
	query := `
		SELECT id, tenant_id, name, type, serial, active, location, created_at, updated_at
		FROM scanner.devices
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	devices := make([]*domain.Device, 0)
	for rows.Next() {
		device, err := deviceRowsToDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, device)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM scanner.devices WHERE tenant_id = $1 WHERE tenant_id = $1"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	return devices, total, rows.Err()
}

func (r *DevicePostgres) Update(ctx context.Context, device *domain.Device) error {
	query := `
		UPDATE scanner.devices
		SET name = $1, active = $2, location = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6
	`

	_, err := r.db.Exec(ctx, query,
		device.Name,
		device.Active,
		device.Location,
		device.UpdatedAt,
		device.ID,
		device.TenantID,
	)

	return err
}

func (r *DevicePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM scanner.devices WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func deviceRowToDevice(row *sql.Row) (*domain.Device, error) {
	device := &domain.Device{}
	err := row.Scan(
		&device.ID,
		&device.TenantID,
		&device.Name,
		(*string)(&device.Type),
		&device.Serial,
		&device.Active,
		&device.Location,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found")
		}
		return nil, err
	}
	return device, nil
}

func deviceRowsToDevice(rows *sql.Rows) (*domain.Device, error) {
	device := &domain.Device{}
	err := rows.Scan(
		&device.ID,
		&device.TenantID,
		&device.Name,
		(*string)(&device.Type),
		&device.Serial,
		&device.Active,
		&device.Location,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	return device, err
}
