package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jeckersberger/rentflow/pkg/common/database"
	"github.com/jeckersberger/rentflow/services/inventory-service/internal/domain"
	"github.com/lib/pq"
)

type EquipmentTypePostgres struct {
	db *database.PostgresPool
}

func NewEquipmentTypePostgres(db *database.PostgresPool) *EquipmentTypePostgres {
	return &EquipmentTypePostgres{db: db}
}

func (r *EquipmentTypePostgres) Create(ctx context.Context, et *domain.EquipmentType) error {
	query := `
		INSERT INTO inventory.equipment_types (
			id, tenant_id, name, description, category_id, manufacturer, model,
			sku_prefix, rental_price_day, rental_price_week, replacement_value,
			weight, dim_length, dim_width, dim_height, dim_unit,
			image_url, tags, custom_fields, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`

	tags := pq.StringArray(et.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	customFields := toJSONB(et.CustomFields)

	dimUnit := et.Dimensions.Unit
	if dimUnit == "" {
		dimUnit = "cm"
	}

	_, err := r.db.Exec(ctx, query,
		et.ID, et.TenantID, et.Name, et.Description, et.CategoryID,
		et.Manufacturer, et.Model, et.SKUPrefix,
		et.RentalPriceDay, et.RentalPriceWeek, et.ReplacementValue,
		et.Weight, et.Dimensions.Length, et.Dimensions.Width, et.Dimensions.Height,
		dimUnit, et.ImageURL, tags, customFields,
		et.CreatedAt, et.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("equipment type with this name already exists for this tenant")
		}
		return fmt.Errorf("failed to create equipment type: %w", err)
	}

	return nil
}

func (r *EquipmentTypePostgres) Update(ctx context.Context, et *domain.EquipmentType) error {
	query := `
		UPDATE inventory.equipment_types SET
			name = $3, description = $4, category_id = $5, manufacturer = $6,
			model = $7, sku_prefix = $8, rental_price_day = $9, rental_price_week = $10,
			replacement_value = $11, weight = $12, dim_length = $13, dim_width = $14,
			dim_height = $15, dim_unit = $16, image_url = $17, tags = $18,
			custom_fields = $19, updated_at = $20
		WHERE id = $1 AND tenant_id = $2
	`

	tags := pq.StringArray(et.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}

	customFields := toJSONB(et.CustomFields)

	dimUnit := et.Dimensions.Unit
	if dimUnit == "" {
		dimUnit = "cm"
	}

	result, err := r.db.Exec(ctx, query,
		et.ID, et.TenantID, et.Name, et.Description, et.CategoryID,
		et.Manufacturer, et.Model, et.SKUPrefix,
		et.RentalPriceDay, et.RentalPriceWeek, et.ReplacementValue,
		et.Weight, et.Dimensions.Length, et.Dimensions.Width, et.Dimensions.Height,
		dimUnit, et.ImageURL, tags, customFields,
		et.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update equipment type: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("equipment type not found: %s", et.ID)
	}

	return nil
}

func (r *EquipmentTypePostgres) GetByID(ctx context.Context, tenantID, id string) (*domain.EquipmentType, error) {
	query := `
		SELECT id, tenant_id, name, description, category_id, manufacturer, model,
		       sku_prefix, rental_price_day, rental_price_week, replacement_value,
		       weight, dim_length, dim_width, dim_height, dim_unit,
		       image_url, tags, custom_fields, created_at, updated_at
		FROM inventory.equipment_types
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRow(ctx, query, id, tenantID)
	return r.scanEquipmentType(row)
}

func (r *EquipmentTypePostgres) List(ctx context.Context, tenantID, categoryID string, limit, offset int) ([]*domain.EquipmentType, int64, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	whereConditions = append(whereConditions, fmt.Sprintf("tenant_id = $%d", argIndex))
	args = append(args, tenantID)
	argIndex++

	if categoryID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, categoryID)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory.equipment_types WHERE %s", whereClause)
	row := r.db.QueryRow(ctx, countQuery, args...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count equipment types: %w", err)
	}

	// List
	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, category_id, manufacturer, model,
		       sku_prefix, rental_price_day, rental_price_week, replacement_value,
		       weight, dim_length, dim_width, dim_height, dim_unit,
		       image_url, tags, custom_fields, created_at, updated_at
		FROM inventory.equipment_types
		WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query equipment types: %w", err)
	}
	defer rows.Close()

	var items []*domain.EquipmentType
	for rows.Next() {
		et, err := r.scanEquipmentTypeRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, et)
	}

	return items, total, nil
}

func (r *EquipmentTypePostgres) Delete(ctx context.Context, tenantID, id string) error {
	query := "DELETE FROM inventory.equipment_types WHERE id = $1 AND tenant_id = $2"
	_, err := r.db.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete equipment type: %w", err)
	}
	return nil
}

func (r *EquipmentTypePostgres) CountItems(ctx context.Context, tenantID, typeID string) (int64, error) {
	query := "SELECT COUNT(*) FROM inventory.equipment WHERE tenant_id = $1 AND equipment_type_id = $2"
	row := r.db.QueryRow(ctx, query, tenantID, typeID)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count items for type: %w", err)
	}
	return count, nil
}

func (r *EquipmentTypePostgres) NextItemNumber(ctx context.Context, tenantID, typeID string) (int, error) {
	query := "SELECT COALESCE(MAX(item_number), 0) + 1 FROM inventory.equipment WHERE tenant_id = $1 AND equipment_type_id = $2"
	row := r.db.QueryRow(ctx, query, tenantID, typeID)
	var next int
	if err := row.Scan(&next); err != nil {
		return 0, fmt.Errorf("failed to get next item number: %w", err)
	}
	return next, nil
}

// scan helpers

func (r *EquipmentTypePostgres) scanEquipmentTypeFields(scanner interface {
	Scan(...interface{}) error
}) (*domain.EquipmentType, error) {
	et := &domain.EquipmentType{}
	var tags pq.StringArray
	var customFieldsJSON []byte
	var description, manufacturer, model, skuPrefix, dimUnit, imageURL sql.NullString
	var rentalPriceDay, rentalPriceWeek, replacementValue, weight sql.NullFloat64
	var dimLength, dimWidth, dimHeight sql.NullFloat64

	err := scanner.Scan(
		&et.ID, &et.TenantID, &et.Name, &description, &et.CategoryID,
		&manufacturer, &model, &skuPrefix,
		&rentalPriceDay, &rentalPriceWeek, &replacementValue,
		&weight, &dimLength, &dimWidth, &dimHeight, &dimUnit,
		&imageURL, &tags, &customFieldsJSON,
		&et.CreatedAt, &et.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	et.Description = description.String
	et.Manufacturer = manufacturer.String
	et.Model = model.String
	et.SKUPrefix = skuPrefix.String
	et.RentalPriceDay = rentalPriceDay.Float64
	et.RentalPriceWeek = rentalPriceWeek.Float64
	et.ReplacementValue = replacementValue.Float64
	et.Weight = weight.Float64
	et.Dimensions.Length = dimLength.Float64
	et.Dimensions.Width = dimWidth.Float64
	et.Dimensions.Height = dimHeight.Float64
	et.Dimensions.Unit = dimUnit.String
	if et.Dimensions.Unit == "" {
		et.Dimensions.Unit = "cm"
	}
	et.ImageURL = imageURL.String
	et.Tags = []string(tags)
	et.CustomFields = fromJSONB(customFieldsJSON)

	return et, nil
}

func (r *EquipmentTypePostgres) scanEquipmentType(row *sql.Row) (*domain.EquipmentType, error) {
	et, err := r.scanEquipmentTypeFields(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewDomainError("NOT_FOUND", "equipment type not found", nil)
		}
		return nil, fmt.Errorf("failed to scan equipment type: %w", err)
	}
	return et, nil
}

func (r *EquipmentTypePostgres) scanEquipmentTypeRow(rows interface {
	Scan(...interface{}) error
}) (*domain.EquipmentType, error) {
	et, err := r.scanEquipmentTypeFields(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan equipment type row: %w", err)
	}
	return et, nil
}

// SetEquipmentTypeFields sets equipment_type_id and item_number on an equipment item
func (r *EquipmentTypePostgres) SetEquipmentTypeFields(ctx context.Context, tenantID, equipmentID, typeID string, itemNumber int) error {
	query := `UPDATE inventory.equipment SET equipment_type_id = $3, item_number = $4 WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.Exec(ctx, query, equipmentID, tenantID, typeID, itemNumber)
	if err != nil {
		return fmt.Errorf("failed to set equipment type fields: %w", err)
	}
	return nil
}
