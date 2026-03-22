package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/ports"
	"github.com/lib/pq"
)

type EquipmentPostgres struct {
	db *database.PostgresPool
}

func NewEquipmentPostgres(db *database.PostgresPool) *EquipmentPostgres {
	return &EquipmentPostgres{db: db}
}

func (r *EquipmentPostgres) Create(ctx context.Context, eq *domain.Equipment) error {
	query := `
		INSERT INTO inventory.equipment (
			id, tenant_id, name, description, category_id, sku, serial_number,
			barcode, status, condition, purchase_date, purchase_price,
			rental_price_day, rental_price_week, weight, dim_length, dim_width,
			dim_height, dim_unit, location_id, image_refs, tags, custom_fields,
			created_at, updated_at, created_by_user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26
		)
	`

	imageRefs := pq.StringArray(eq.ImageRefs)
	if imageRefs == nil {
		imageRefs = pq.StringArray{}
	}

	tags := pq.StringArray(eq.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	customFields := toJSONB(eq.CustomFields)

	_, err := r.db.Exec(ctx, query,
		eq.ID, eq.TenantID, eq.Name, eq.Description, eq.CategoryID, eq.SKU,
		eq.SerialNumber, eq.Barcode, string(eq.Status), string(eq.Condition),
		eq.PurchaseDate, eq.PurchasePrice, eq.RentalPriceDay, eq.RentalPriceWeek,
		eq.Weight, eq.Dimensions.Length, eq.Dimensions.Width, eq.Dimensions.Height,
		eq.Dimensions.Unit, eq.LocationID, imageRefs, tags, customFields,
		eq.CreatedAt, eq.UpdatedAt, eq.CreatedByUserID,
	)

	if err != nil {
		return fmt.Errorf("failed to create equipment: %w", err)
	}

	return nil
}

func (r *EquipmentPostgres) Update(ctx context.Context, eq *domain.Equipment) error {
	query := `
		UPDATE inventory.equipment SET
			name = $3, description = $4, category_id = $5, sku = $6,
			serial_number = $7, barcode = $8, status = $9, condition = $10,
			purchase_date = $11, purchase_price = $12, rental_price_day = $13,
			rental_price_week = $14, weight = $15, dim_length = $16, dim_width = $17,
			dim_height = $18, dim_unit = $19, location_id = $20, image_refs = $21,
			tags = $22, custom_fields = $23, updated_at = $24
		WHERE id = $1 AND tenant_id = $2
	`

	imageRefs := pq.StringArray(eq.ImageRefs)
	if imageRefs == nil {
		imageRefs = pq.StringArray{}
	}

	tags := pq.StringArray(eq.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	customFields := toJSONB(eq.CustomFields)

	result, err := r.db.Exec(ctx, query,
		eq.ID, eq.TenantID, eq.Name, eq.Description, eq.CategoryID, eq.SKU,
		eq.SerialNumber, eq.Barcode, string(eq.Status), string(eq.Condition),
		eq.PurchaseDate, eq.PurchasePrice, eq.RentalPriceDay, eq.RentalPriceWeek,
		eq.Weight, eq.Dimensions.Length, eq.Dimensions.Width, eq.Dimensions.Height,
		eq.Dimensions.Unit, eq.LocationID, imageRefs, tags, customFields,
		eq.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("equipment not found: %s", eq.ID)
	}

	return nil
}

func (r *EquipmentPostgres) GetByID(ctx context.Context, tenantID, equipmentID string) (*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, description, category_id, sku, serial_number,
		       barcode, status, condition, purchase_date, purchase_price,
		       rental_price_day, rental_price_week, weight, dim_length, dim_width,
		       dim_height, dim_unit, location_id, image_refs, tags, custom_fields,
		       created_at, updated_at, created_by_user_id
		FROM inventory.equipment
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, equipmentID, tenantID)
	return r.scanEquipment(row)
}

func (r *EquipmentPostgres) GetByBarcode(ctx context.Context, tenantID, barcode string) (*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, description, category_id, sku, serial_number,
		       barcode, status, condition, purchase_date, purchase_price,
		       rental_price_day, rental_price_week, weight, dim_length, dim_width,
		       dim_height, dim_unit, location_id, image_refs, tags, custom_fields,
		       created_at, updated_at, created_by_user_id
		FROM inventory.equipment
		WHERE barcode = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, barcode, tenantID)
	return r.scanEquipment(row)
}

func (r *EquipmentPostgres) List(ctx context.Context, query *ports.EquipmentListQuery) (*ports.EquipmentListResult, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, query.TenantID)
	argIndex++

	if query.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*query.Status))
		argIndex++
	}

	if query.CategoryID != nil && *query.CategoryID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *query.CategoryID)
		argIndex++
	}

	if query.LocationID != nil && *query.LocationID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("location_id = $%d", argIndex))
		args = append(args, *query.LocationID)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory.equipment WHERE %s", whereClause)
	row := r.db.QueryRow(ctx, countQuery, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count equipment: %w", err)
	}

	// Get paginated results
	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, category_id, sku, serial_number,
		       barcode, status, condition, purchase_date, purchase_price,
		       rental_price_day, rental_price_week, weight, dim_length, dim_width,
		       dim_height, dim_unit, location_id, image_refs, tags, custom_fields,
		       created_at, updated_at, created_by_user_id
		FROM inventory.equipment
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query equipment: %w", err)
	}
	defer rows.Close()

	var items []*domain.Equipment
	for rows.Next() {
		eq, err := r.scanEquipmentRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, eq)
	}

	return &ports.EquipmentListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

func (r *EquipmentPostgres) Search(ctx context.Context, tenantID, term string, limit, offset int) (*ports.EquipmentListResult, error) {
	searchTerm := "%" + term + "%"

	countQuery := `
		SELECT COUNT(*) FROM inventory.equipment
		WHERE tenant_id = $1 AND (
			name ILIKE $2 OR description ILIKE $2 OR sku ILIKE $2 OR serial_number ILIKE $2 OR barcode ILIKE $2
		)
	`

	row := r.db.QueryRow(ctx, countQuery, tenantID, searchTerm)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	listQuery := `
		SELECT id, tenant_id, name, description, category_id, sku, serial_number,
		       barcode, status, condition, purchase_date, purchase_price,
		       rental_price_day, rental_price_week, weight, dim_length, dim_width,
		       dim_height, dim_unit, location_id, image_refs, tags, custom_fields,
		       created_at, updated_at, created_by_user_id
		FROM inventory.equipment
		WHERE tenant_id = $1 AND (
			name ILIKE $2 OR description ILIKE $2 OR sku ILIKE $2 OR serial_number ILIKE $2 OR barcode ILIKE $2
		)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, listQuery, tenantID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search equipment: %w", err)
	}
	defer rows.Close()

	var items []*domain.Equipment
	for rows.Next() {
		eq, err := r.scanEquipmentRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, eq)
	}

	return &ports.EquipmentListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (r *EquipmentPostgres) Delete(ctx context.Context, tenantID, equipmentID string) error {
	query := "DELETE FROM inventory.equipment WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, equipmentID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}
	return nil
}

// Helper methods

func (r *EquipmentPostgres) scanEquipmentFields(scanner interface {
	Scan(...interface{}) error
}) (*domain.Equipment, error) {
	eq := &domain.Equipment{}
	var imageRefs pq.StringArray
	var tags pq.StringArray
	var customFieldsJSON []byte

	// Use sql.Null* types for nullable columns
	var description, sku, serialNumber, locationID, dimUnit sql.NullString
	var purchasePrice, rentalPriceDay, rentalPriceWeek, weight sql.NullFloat64
	var dimLength, dimWidth, dimHeight sql.NullFloat64
	var status, condition sql.NullString

	err := scanner.Scan(
		&eq.ID, &eq.TenantID, &eq.Name, &description, &eq.CategoryID, &sku,
		&serialNumber, &eq.Barcode, &status, &condition,
		&eq.PurchaseDate, &purchasePrice, &rentalPriceDay, &rentalPriceWeek,
		&weight, &dimLength, &dimWidth, &dimHeight,
		&dimUnit, &locationID, &imageRefs, &tags, &customFieldsJSON,
		&eq.CreatedAt, &eq.UpdatedAt, &eq.CreatedByUserID,
	)

	if err != nil {
		return nil, err
	}

	eq.Description = description.String
	eq.SKU = sku.String
	eq.SerialNumber = serialNumber.String
	eq.LocationID = locationID.String
	eq.PurchasePrice = purchasePrice.Float64
	eq.RentalPriceDay = rentalPriceDay.Float64
	eq.RentalPriceWeek = rentalPriceWeek.Float64
	eq.Weight = weight.Float64
	eq.Dimensions.Length = dimLength.Float64
	eq.Dimensions.Width = dimWidth.Float64
	eq.Dimensions.Height = dimHeight.Float64
	eq.Dimensions.Unit = dimUnit.String
	if eq.Dimensions.Unit == "" {
		eq.Dimensions.Unit = "cm"
	}
	eq.Status = domain.EquipmentStatus(status.String)
	if eq.Status == "" {
		eq.Status = domain.StatusAvailable
	}
	eq.Condition = domain.EquipmentCondition(condition.String)
	if eq.Condition == "" {
		eq.Condition = domain.ConditionGood
	}
	eq.ImageRefs = []string(imageRefs)
	eq.Tags = []string(tags)
	eq.CustomFields = fromJSONB(customFieldsJSON)

	return eq, nil
}

func (r *EquipmentPostgres) scanEquipment(row *sql.Row) (*domain.Equipment, error) {
	eq, err := r.scanEquipmentFields(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrEquipmentNotFound
		}
		return nil, fmt.Errorf("failed to scan equipment: %w", err)
	}
	return eq, nil
}

func (r *EquipmentPostgres) scanEquipmentRow(rows interface {
	Scan(...interface{}) error
}) (*domain.Equipment, error) {
	eq, err := r.scanEquipmentFields(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan equipment row: %w", err)
	}
	return eq, nil
}
