package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type StockLocationPostgres struct {
	db *database.PostgresPool
}

func NewStockLocationPostgres(db *database.PostgresPool) *StockLocationPostgres {
	return &StockLocationPostgres{db: db}
}

func (r *StockLocationPostgres) Create(ctx context.Context, location *domain.StockLocation) error {
	query := `
		INSERT INTO stock_locations (id, tenant_id, warehouse_id, zone_id, rack_id, bay_id,
		                             location_code, barcode, qr_code_label, location_type,
		                             capacity, weight_capacity, equipment_id, is_available,
		                             access_level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.Exec(ctx, query,
		location.ID, location.TenantID, location.WarehouseID, location.ZoneID, location.RackID,
		location.BayID, location.LocationCode, location.Barcode, location.QRCodeLabel,
		location.LocationType, location.Capacity, location.WeightCapacity, location.EquipmentID,
		location.IsAvailable, location.AccessLevel,
	)
	return err
}

func (r *StockLocationPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.StockLocation, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, zone_id, rack_id, bay_id, location_code, barcode,
		       qr_code_label, location_type, capacity, current_occupancy, weight_capacity,
		       current_weight, equipment_id, is_available, access_level, created_at, updated_at
		FROM stock_locations
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	var sl domain.StockLocation
	err := row.Scan(
		&sl.ID, &sl.TenantID, &sl.WarehouseID, &sl.ZoneID, &sl.RackID, &sl.BayID,
		&sl.LocationCode, &sl.Barcode, &sl.QRCodeLabel, &sl.LocationType,
		&sl.Capacity, &sl.Occupancy, &sl.WeightCapacity, &sl.CurrentWeight,
		&sl.EquipmentID, &sl.IsAvailable, &sl.AccessLevel, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &sl, nil
}

func (r *StockLocationPostgres) GetByLocationCode(ctx context.Context, tenantID, code string) (*domain.StockLocation, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, zone_id, rack_id, bay_id, location_code, barcode,
		       qr_code_label, location_type, capacity, current_occupancy, weight_capacity,
		       current_weight, equipment_id, is_available, access_level, created_at, updated_at
		FROM stock_locations
		WHERE tenant_id = $1 AND location_code = $2
	`
	row := r.db.QueryRow(ctx, query, tenantID, code)

	var sl domain.StockLocation
	err := row.Scan(
		&sl.ID, &sl.TenantID, &sl.WarehouseID, &sl.ZoneID, &sl.RackID, &sl.BayID,
		&sl.LocationCode, &sl.Barcode, &sl.QRCodeLabel, &sl.LocationType,
		&sl.Capacity, &sl.Occupancy, &sl.WeightCapacity, &sl.CurrentWeight,
		&sl.EquipmentID, &sl.IsAvailable, &sl.AccessLevel, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &sl, nil
}

func (r *StockLocationPostgres) ListByBay(ctx context.Context, tenantID, bayID string, limit, offset int) ([]*domain.StockLocation, int, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM stock_locations WHERE tenant_id = $1 AND bay_id = $2"
	var total int
	err := r.db.QueryRow(ctx, countQuery, tenantID, bayID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, warehouse_id, zone_id, rack_id, bay_id, location_code, barcode,
		       qr_code_label, location_type, capacity, current_occupancy, weight_capacity,
		       current_weight, equipment_id, is_available, access_level, created_at, updated_at
		FROM stock_locations
		WHERE tenant_id = $1 AND bay_id = $2
		ORDER BY location_code ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, bayID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var locations []*domain.StockLocation
	for rows.Next() {
		var sl domain.StockLocation
		err := rows.Scan(
			&sl.ID, &sl.TenantID, &sl.WarehouseID, &sl.ZoneID, &sl.RackID, &sl.BayID,
			&sl.LocationCode, &sl.Barcode, &sl.QRCodeLabel, &sl.LocationType,
			&sl.Capacity, &sl.Occupancy, &sl.WeightCapacity, &sl.CurrentWeight,
			&sl.EquipmentID, &sl.IsAvailable, &sl.AccessLevel, &sl.CreatedAt, &sl.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		locations = append(locations, &sl)
	}

	return locations, total, rows.Err()
}

func (r *StockLocationPostgres) ListByEquipment(ctx context.Context, tenantID, equipmentID string) ([]*domain.StockLocation, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, zone_id, rack_id, bay_id, location_code, barcode,
		       qr_code_label, location_type, capacity, current_occupancy, weight_capacity,
		       current_weight, equipment_id, is_available, access_level, created_at, updated_at
		FROM stock_locations
		WHERE tenant_id = $1 AND equipment_id = $2
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, tenantID, equipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*domain.StockLocation
	for rows.Next() {
		var sl domain.StockLocation
		err := rows.Scan(
			&sl.ID, &sl.TenantID, &sl.WarehouseID, &sl.ZoneID, &sl.RackID, &sl.BayID,
			&sl.LocationCode, &sl.Barcode, &sl.QRCodeLabel, &sl.LocationType,
			&sl.Capacity, &sl.Occupancy, &sl.WeightCapacity, &sl.CurrentWeight,
			&sl.EquipmentID, &sl.IsAvailable, &sl.AccessLevel, &sl.CreatedAt, &sl.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, &sl)
	}

	return locations, rows.Err()
}

func (r *StockLocationPostgres) Update(ctx context.Context, location *domain.StockLocation) error {
	query := `
		UPDATE stock_locations
		SET barcode = $1, qr_code_label = $2, capacity = $3, current_occupancy = $4,
		    weight_capacity = $5, current_weight = $6, equipment_id = $7, is_available = $8,
		    access_level = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`
	_, err := r.db.Exec(ctx, query,
		location.Barcode, location.QRCodeLabel, location.Capacity, location.Occupancy,
		location.WeightCapacity, location.CurrentWeight, location.EquipmentID, location.IsAvailable,
		location.AccessLevel, location.UpdatedAt, location.ID, location.TenantID,
	)
	return err
}

func (r *StockLocationPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM stock_locations WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}
