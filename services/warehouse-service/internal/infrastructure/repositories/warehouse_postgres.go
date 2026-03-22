package repositories

import (
	"context"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/warehouse-service/internal/domain"
)

type WarehousePostgres struct {
	db *database.PostgresPool
}

func NewWarehousePostgres(db *database.PostgresPool) *WarehousePostgres {
	return &WarehousePostgres{db: db}
}

func (r *WarehousePostgres) Create(ctx context.Context, warehouse *domain.Warehouse) error {
	query := `
		INSERT INTO warehouses (id, tenant_id, name, code, address, city, postal_code, country, latitude, longitude, total_capacity, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		warehouse.ID, warehouse.TenantID, warehouse.Name, warehouse.Code,
		warehouse.Address, warehouse.City, warehouse.PostalCode, warehouse.Country,
		warehouse.Latitude, warehouse.Longitude, warehouse.TotalCapacity, warehouse.Status,
	)
	return err
}

func (r *WarehousePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Warehouse, error) {
	query := `
		SELECT id, tenant_id, name, code, address, city, postal_code, country, latitude, longitude,
		       total_capacity, current_occupancy, status, created_at, updated_at
		FROM warehouses
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	var w domain.Warehouse
	err := row.Scan(
		&w.ID, &w.TenantID, &w.Name, &w.Code, &w.Address, &w.City, &w.PostalCode, &w.Country,
		&w.Latitude, &w.Longitude, &w.TotalCapacity, &w.CurrentOccupancy, &w.Status,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (r *WarehousePostgres) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Warehouse, int, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM warehouses WHERE tenant_id = $1"
	var total int
	err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, name, code, address, city, postal_code, country, latitude, longitude,
		       total_capacity, current_occupancy, status, created_at, updated_at
		FROM warehouses
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var warehouses []*domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		err := rows.Scan(
			&w.ID, &w.TenantID, &w.Name, &w.Code, &w.Address, &w.City, &w.PostalCode, &w.Country,
			&w.Latitude, &w.Longitude, &w.TotalCapacity, &w.CurrentOccupancy, &w.Status,
			&w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		warehouses = append(warehouses, &w)
	}

	return warehouses, total, rows.Err()
}

func (r *WarehousePostgres) Update(ctx context.Context, warehouse *domain.Warehouse) error {
	query := `
		UPDATE warehouses
		SET name = $1, address = $2, city = $3, postal_code = $4, country = $5,
		    latitude = $6, longitude = $7, total_capacity = $8, status = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`
	_, err := r.db.Exec(ctx, query,
		warehouse.Name, warehouse.Address, warehouse.City, warehouse.PostalCode, warehouse.Country,
		warehouse.Latitude, warehouse.Longitude, warehouse.TotalCapacity, warehouse.Status,
		warehouse.UpdatedAt, warehouse.ID, warehouse.TenantID,
	)
	return err
}

func (r *WarehousePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM warehouses WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

// ZonePostgres
type ZonePostgres struct {
	db *database.PostgresPool
}

func NewZonePostgres(db *database.PostgresPool) *ZonePostgres {
	return &ZonePostgres{db: db}
}

func (r *ZonePostgres) Create(ctx context.Context, zone *domain.Zone) error {
	query := `
		INSERT INTO zones (id, tenant_id, warehouse_id, name, code, zone_type, description,
		                   temperature_min, temperature_max, humidity_min, humidity_max, capacity, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		zone.ID, zone.TenantID, zone.WarehouseID, zone.Name, zone.Code, zone.ZoneType, zone.Description,
		zone.TemperatureMin, zone.TemperatureMax, zone.HumidityMin, zone.HumidityMax,
		zone.Capacity, zone.SortOrder,
	)
	return err
}

func (r *ZonePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Zone, error) {
	query := `
		SELECT id, tenant_id, warehouse_id, name, code, zone_type, description,
		       temperature_min, temperature_max, humidity_min, humidity_max,
		       capacity, current_occupancy, sort_order, created_at, updated_at
		FROM zones
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	var z domain.Zone
	err := row.Scan(
		&z.ID, &z.TenantID, &z.WarehouseID, &z.Name, &z.Code, &z.ZoneType, &z.Description,
		&z.TemperatureMin, &z.TemperatureMax, &z.HumidityMin, &z.HumidityMax,
		&z.Capacity, &z.Occupancy, &z.SortOrder, &z.CreatedAt, &z.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &z, nil
}

func (r *ZonePostgres) ListByWarehouse(ctx context.Context, tenantID, warehouseID string, limit, offset int) ([]*domain.Zone, int, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM zones WHERE tenant_id = $1 AND warehouse_id = $2"
	var total int
	err := r.db.QueryRow(ctx, countQuery, tenantID, warehouseID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, warehouse_id, name, code, zone_type, description,
		       temperature_min, temperature_max, humidity_min, humidity_max,
		       capacity, current_occupancy, sort_order, created_at, updated_at
		FROM zones
		WHERE tenant_id = $1 AND warehouse_id = $2
		ORDER BY sort_order ASC, created_at ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, warehouseID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var zones []*domain.Zone
	for rows.Next() {
		var z domain.Zone
		err := rows.Scan(
			&z.ID, &z.TenantID, &z.WarehouseID, &z.Name, &z.Code, &z.ZoneType, &z.Description,
			&z.TemperatureMin, &z.TemperatureMax, &z.HumidityMin, &z.HumidityMax,
			&z.Capacity, &z.Occupancy, &z.SortOrder, &z.CreatedAt, &z.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		zones = append(zones, &z)
	}

	return zones, total, rows.Err()
}

func (r *ZonePostgres) Update(ctx context.Context, zone *domain.Zone) error {
	query := `
		UPDATE zones
		SET name = $1, zone_type = $2, description = $3,
		    temperature_min = $4, temperature_max = $5, humidity_min = $6, humidity_max = $7,
		    capacity = $8, sort_order = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12
	`
	_, err := r.db.Exec(ctx, query,
		zone.Name, zone.ZoneType, zone.Description,
		zone.TemperatureMin, zone.TemperatureMax, zone.HumidityMin, zone.HumidityMax,
		zone.Capacity, zone.SortOrder, zone.UpdatedAt, zone.ID, zone.TenantID,
	)
	return err
}

func (r *ZonePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM zones WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

// RackPostgres
type RackPostgres struct {
	db *database.PostgresPool
}

func NewRackPostgres(db *database.PostgresPool) *RackPostgres {
	return &RackPostgres{db: db}
}

func (r *RackPostgres) Create(ctx context.Context, rack *domain.Rack) error {
	query := `
		INSERT INTO racks (id, tenant_id, zone_id, warehouse_id, name, code, rack_type,
		                   aisle, row_number, column_number, capacity, height, width, depth,
		                   weight_capacity, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.Exec(ctx, query,
		rack.ID, rack.TenantID, rack.ZoneID, rack.WarehouseID, rack.Name, rack.Code, rack.RackType,
		rack.Aisle, rack.RowNumber, rack.ColumnNumber, rack.Capacity, rack.Height, rack.Width,
		rack.Depth, rack.WeightCapacity, rack.SortOrder,
	)
	return err
}

func (r *RackPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Rack, error) {
	query := `
		SELECT id, tenant_id, zone_id, warehouse_id, name, code, rack_type, aisle, row_number,
		       column_number, capacity, current_occupancy, height, width, depth, weight_capacity,
		       sort_order, created_at, updated_at
		FROM racks
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	var rk domain.Rack
	err := row.Scan(
		&rk.ID, &rk.TenantID, &rk.ZoneID, &rk.WarehouseID, &rk.Name, &rk.Code, &rk.RackType,
		&rk.Aisle, &rk.RowNumber, &rk.ColumnNumber, &rk.Capacity, &rk.Occupancy,
		&rk.Height, &rk.Width, &rk.Depth, &rk.WeightCapacity,
		&rk.SortOrder, &rk.CreatedAt, &rk.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &rk, nil
}

func (r *RackPostgres) ListByZone(ctx context.Context, tenantID, zoneID string, limit, offset int) ([]*domain.Rack, int, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM racks WHERE tenant_id = $1 AND zone_id = $2"
	var total int
	err := r.db.QueryRow(ctx, countQuery, tenantID, zoneID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, zone_id, warehouse_id, name, code, rack_type, aisle, row_number,
		       column_number, capacity, current_occupancy, height, width, depth, weight_capacity,
		       sort_order, created_at, updated_at
		FROM racks
		WHERE tenant_id = $1 AND zone_id = $2
		ORDER BY sort_order ASC, created_at ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, zoneID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var racks []*domain.Rack
	for rows.Next() {
		var rk domain.Rack
		err := rows.Scan(
			&rk.ID, &rk.TenantID, &rk.ZoneID, &rk.WarehouseID, &rk.Name, &rk.Code, &rk.RackType,
			&rk.Aisle, &rk.RowNumber, &rk.ColumnNumber, &rk.Capacity, &rk.Occupancy,
			&rk.Height, &rk.Width, &rk.Depth, &rk.WeightCapacity,
			&rk.SortOrder, &rk.CreatedAt, &rk.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		racks = append(racks, &rk)
	}

	return racks, total, rows.Err()
}

func (r *RackPostgres) Update(ctx context.Context, rack *domain.Rack) error {
	query := `
		UPDATE racks
		SET name = $1, rack_type = $2, aisle = $3, row_number = $4, column_number = $5,
		    capacity = $6, height = $7, width = $8, depth = $9, weight_capacity = $10,
		    sort_order = $11, updated_at = $12
		WHERE id = $13 AND tenant_id = $14
	`
	_, err := r.db.Exec(ctx, query,
		rack.Name, rack.RackType, rack.Aisle, rack.RowNumber, rack.ColumnNumber,
		rack.Capacity, rack.Height, rack.Width, rack.Depth, rack.WeightCapacity,
		rack.SortOrder, rack.UpdatedAt, rack.ID, rack.TenantID,
	)
	return err
}

func (r *RackPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM racks WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}

// BayPostgres
type BayPostgres struct {
	db *database.PostgresPool
}

func NewBayPostgres(db *database.PostgresPool) *BayPostgres {
	return &BayPostgres{db: db}
}

func (r *BayPostgres) Create(ctx context.Context, bay *domain.Bay) error {
	query := `
		INSERT INTO bays (id, tenant_id, rack_id, zone_id, warehouse_id, name, code,
		                  bay_number, bay_level, capacity, weight_capacity, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		bay.ID, bay.TenantID, bay.RackID, bay.ZoneID, bay.WarehouseID, bay.Name, bay.Code,
		bay.BayNumber, bay.BayLevel, bay.Capacity, bay.WeightCapacity, bay.SortOrder,
	)
	return err
}

func (r *BayPostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.Bay, error) {
	query := `
		SELECT id, tenant_id, rack_id, zone_id, warehouse_id, name, code, bay_number, bay_level,
		       capacity, current_occupancy, weight_capacity, sort_order, created_at, updated_at
		FROM bays
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRow(ctx, query, id, tenantID)

	var b domain.Bay
	err := row.Scan(
		&b.ID, &b.TenantID, &b.RackID, &b.ZoneID, &b.WarehouseID, &b.Name, &b.Code,
		&b.BayNumber, &b.BayLevel, &b.Capacity, &b.Occupancy, &b.WeightCapacity,
		&b.SortOrder, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r *BayPostgres) ListByRack(ctx context.Context, tenantID, rackID string, limit, offset int) ([]*domain.Bay, int, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM bays WHERE tenant_id = $1 AND rack_id = $2"
	var total int
	err := r.db.QueryRow(ctx, countQuery, tenantID, rackID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, rack_id, zone_id, warehouse_id, name, code, bay_number, bay_level,
		       capacity, current_occupancy, weight_capacity, sort_order, created_at, updated_at
		FROM bays
		WHERE tenant_id = $1 AND rack_id = $2
		ORDER BY bay_level ASC, sort_order ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(ctx, query, tenantID, rackID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bays []*domain.Bay
	for rows.Next() {
		var b domain.Bay
		err := rows.Scan(
			&b.ID, &b.TenantID, &b.RackID, &b.ZoneID, &b.WarehouseID, &b.Name, &b.Code,
			&b.BayNumber, &b.BayLevel, &b.Capacity, &b.Occupancy, &b.WeightCapacity,
			&b.SortOrder, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		bays = append(bays, &b)
	}

	return bays, total, rows.Err()
}

func (r *BayPostgres) Update(ctx context.Context, bay *domain.Bay) error {
	query := `
		UPDATE bays
		SET name = $1, bay_number = $2, bay_level = $3, capacity = $4, weight_capacity = $5,
		    sort_order = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`
	_, err := r.db.Exec(ctx, query,
		bay.Name, bay.BayNumber, bay.BayLevel, bay.Capacity, bay.WeightCapacity,
		bay.SortOrder, bay.UpdatedAt, bay.ID, bay.TenantID,
	)
	return err
}

func (r *BayPostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM bays WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	return err
}
