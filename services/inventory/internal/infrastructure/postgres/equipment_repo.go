package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"

	apperrors "github.com/jeckersberger/EquipFlow/pkg/common/errors"
)

// equipmentColumns lists all columns of the equipment table for consistent scanning.
const equipmentColumns = `
	id, tenant_id, category_id, equipment_type_id, name, description, sku, barcode, qr_code,
	serial_number, rfid_tag, status, condition, quantity_total, quantity_available,
	rental_price_day, rental_price_week, replacement_value,
	weight_grams, width_mm, height_mm, depth_mm,
	location_id, purchase_date, purchase_price, manufacturer, model, image_url,
	custom_fields, notes, is_active, created_at, updated_at`

// EquipmentRepo implements domain.EquipmentRepository using PostgreSQL.
type EquipmentRepo struct {
	pool *pgxpool.Pool
}

// NewEquipmentRepo creates a new EquipmentRepo.
func NewEquipmentRepo(pool *pgxpool.Pool) *EquipmentRepo {
	return &EquipmentRepo{pool: pool}
}

// scanEquipment scans a single equipment row into a domain.Equipment, handling nullable columns.
func scanEquipment(row pgx.Row) (*domain.Equipment, error) {
	e := &domain.Equipment{}
	var (
		description  *string
		sku          *string
		barcode      *string
		qrCode       *string
		serialNumber *string
		rfidTag      *string
		manufacturer *string
		model        *string
		imageURL     *string
		notes        *string
		customFields []byte
	)

	err := row.Scan(
		&e.ID, &e.TenantID, &e.CategoryID, &e.EquipmentTypeID,
		&e.Name, &description, &sku, &barcode, &qrCode,
		&serialNumber, &rfidTag, &e.Status, &e.Condition,
		&e.QuantityTotal, &e.QuantityAvailable,
		&e.RentalPriceDay, &e.RentalPriceWeek, &e.ReplacementValue,
		&e.WeightGrams, &e.WidthMM, &e.HeightMM, &e.DepthMM,
		&e.LocationID, &e.PurchaseDate, &e.PurchasePrice,
		&manufacturer, &model, &imageURL,
		&customFields, &notes,
		&e.IsActive, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if description != nil {
		e.Description = *description
	}
	if sku != nil {
		e.SKU = *sku
	}
	if barcode != nil {
		e.Barcode = *barcode
	}
	if qrCode != nil {
		e.QRCode = *qrCode
	}
	if serialNumber != nil {
		e.SerialNumber = *serialNumber
	}
	if rfidTag != nil {
		e.RFIDTag = *rfidTag
	}
	if manufacturer != nil {
		e.Manufacturer = *manufacturer
	}
	if model != nil {
		e.Model = *model
	}
	if imageURL != nil {
		e.ImageURL = *imageURL
	}
	if notes != nil {
		e.Notes = *notes
	}
	if customFields != nil {
		e.CustomFields = json.RawMessage(customFields)
	}

	return e, nil
}

// Create inserts a new equipment record and scans back the generated fields.
func (r *EquipmentRepo) Create(ctx context.Context, equipment *domain.Equipment) error {
	query := `
		INSERT INTO equipment (
			id, tenant_id, category_id, equipment_type_id, name, description, sku, barcode, qr_code,
			serial_number, rfid_tag, status, condition, quantity_total, quantity_available,
			rental_price_day, rental_price_week, replacement_value,
			weight_grams, width_mm, height_mm, depth_mm,
			location_id, purchase_date, purchase_price, manufacturer, model, image_url,
			custom_fields, notes, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26, $27, $28,
			$29, $30, $31
		) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		equipment.ID, equipment.TenantID, equipment.CategoryID, equipment.EquipmentTypeID,
		equipment.Name, nilIfEmpty(equipment.Description), nilIfEmpty(equipment.SKU),
		nilIfEmpty(equipment.Barcode), nilIfEmpty(equipment.QRCode),
		nilIfEmpty(equipment.SerialNumber), nilIfEmpty(equipment.RFIDTag),
		equipment.Status, equipment.Condition,
		equipment.QuantityTotal, equipment.QuantityAvailable,
		equipment.RentalPriceDay, equipment.RentalPriceWeek, equipment.ReplacementValue,
		equipment.WeightGrams, equipment.WidthMM, equipment.HeightMM, equipment.DepthMM,
		equipment.LocationID, equipment.PurchaseDate, equipment.PurchasePrice,
		nilIfEmpty(equipment.Manufacturer), nilIfEmpty(equipment.Model), nilIfEmpty(equipment.ImageURL),
		nilIfEmptyJSON(equipment.CustomFields), nilIfEmpty(equipment.Notes),
		equipment.IsActive,
	).Scan(&equipment.ID, &equipment.CreatedAt, &equipment.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "barcode") {
				return domain.ErrDuplicateBarcode
			}
			if strings.Contains(pgErr.ConstraintName, "rfid") {
				return domain.ErrDuplicateRFID
			}
		}
		return fmt.Errorf("equipment_repo: create: %w", err)
	}
	return nil
}

// GetByID retrieves a single equipment record by primary key scoped to a tenant.
func (r *EquipmentRepo) GetByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*domain.Equipment, error) {
	query := fmt.Sprintf(`SELECT %s FROM equipment WHERE id = $1 AND tenant_id = $2`, equipmentColumns)
	e, err := scanEquipment(r.pool.QueryRow(ctx, query, id, tenantID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("equipment_repo: get_by_id: %w", err)
	}
	return e, nil
}

// GetByBarcode retrieves equipment by its unique barcode within a tenant.
func (r *EquipmentRepo) GetByBarcode(ctx context.Context, tenantID uuid.UUID, barcode string) (*domain.Equipment, error) {
	query := fmt.Sprintf(`SELECT %s FROM equipment WHERE tenant_id = $1 AND barcode = $2`, equipmentColumns)
	e, err := scanEquipment(r.pool.QueryRow(ctx, query, tenantID, barcode))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("equipment_repo: get_by_barcode: %w", err)
	}
	return e, nil
}

// GetByRFID retrieves equipment by its unique RFID tag within a tenant.
func (r *EquipmentRepo) GetByRFID(ctx context.Context, tenantID uuid.UUID, rfidTag string) (*domain.Equipment, error) {
	query := fmt.Sprintf(`SELECT %s FROM equipment WHERE tenant_id = $1 AND rfid_tag = $2`, equipmentColumns)
	e, err := scanEquipment(r.pool.QueryRow(ctx, query, tenantID, rfidTag))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("equipment_repo: get_by_rfid: %w", err)
	}
	return e, nil
}

// GetBySerialNumber retrieves equipment by its serial number within a tenant.
func (r *EquipmentRepo) GetBySerialNumber(ctx context.Context, tenantID uuid.UUID, serialNumber string) (*domain.Equipment, error) {
	query := fmt.Sprintf(`SELECT %s FROM equipment WHERE tenant_id = $1 AND serial_number = $2`, equipmentColumns)
	e, err := scanEquipment(r.pool.QueryRow(ctx, query, tenantID, serialNumber))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("equipment_repo: get_by_serial_number: %w", err)
	}
	return e, nil
}

// List returns a filtered, paginated list of equipment for a tenant plus total count.
func (r *EquipmentRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.EquipmentFilter) ([]*domain.Equipment, int64, error) {
	// Build dynamic WHERE clause.
	conditions := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Condition != nil {
		conditions = append(conditions, fmt.Sprintf("condition = $%d", argIdx))
		args = append(args, *filter.Condition)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR description ILIKE $%d OR barcode ILIKE $%d OR serial_number ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM equipment WHERE %s`, where)
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list count: %w", err)
	}

	// Data query with pagination.
	offset := (filter.Page - 1) * filter.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM equipment WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		equipmentColumns, where, argIdx, argIdx+1,
	)
	args = append(args, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Equipment
	for rows.Next() {
		e, scanErr := scanEquipment(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("equipment_repo: list scan: %w", scanErr)
		}
		items = append(items, e)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: list rows: %w", err)
	}
	return items, total, nil
}

// Update modifies an existing equipment record.
func (r *EquipmentRepo) Update(ctx context.Context, equipment *domain.Equipment) error {
	query := `
		UPDATE equipment SET
			category_id = $3, equipment_type_id = $4, name = $5, description = $6,
			sku = $7, barcode = $8, qr_code = $9, serial_number = $10, rfid_tag = $11,
			status = $12, condition = $13, quantity_total = $14, quantity_available = $15,
			rental_price_day = $16, rental_price_week = $17, replacement_value = $18,
			weight_grams = $19, width_mm = $20, height_mm = $21, depth_mm = $22,
			location_id = $23, purchase_date = $24, purchase_price = $25,
			manufacturer = $26, model = $27, image_url = $28,
			custom_fields = $29, notes = $30, is_active = $31,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		equipment.ID, equipment.TenantID,
		equipment.CategoryID, equipment.EquipmentTypeID,
		equipment.Name, nilIfEmpty(equipment.Description),
		nilIfEmpty(equipment.SKU), nilIfEmpty(equipment.Barcode), nilIfEmpty(equipment.QRCode),
		nilIfEmpty(equipment.SerialNumber), nilIfEmpty(equipment.RFIDTag),
		equipment.Status, equipment.Condition,
		equipment.QuantityTotal, equipment.QuantityAvailable,
		equipment.RentalPriceDay, equipment.RentalPriceWeek, equipment.ReplacementValue,
		equipment.WeightGrams, equipment.WidthMM, equipment.HeightMM, equipment.DepthMM,
		equipment.LocationID, equipment.PurchaseDate, equipment.PurchasePrice,
		nilIfEmpty(equipment.Manufacturer), nilIfEmpty(equipment.Model), nilIfEmpty(equipment.ImageURL),
		nilIfEmptyJSON(equipment.CustomFields), nilIfEmpty(equipment.Notes),
		equipment.IsActive,
	).Scan(&equipment.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "barcode") {
				return domain.ErrDuplicateBarcode
			}
			if strings.Contains(pgErr.ConstraintName, "rfid") {
				return domain.ErrDuplicateRFID
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("equipment_repo: update: %w", err)
	}
	return nil
}

// UpdateStatus changes the status of an equipment item.
func (r *EquipmentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE equipment SET status = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, status,
	)
	if err != nil {
		return fmt.Errorf("equipment_repo: update_status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateCondition changes the condition of an equipment item.
func (r *EquipmentRepo) UpdateCondition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, condition string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE equipment SET condition = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, condition,
	)
	if err != nil {
		return fmt.Errorf("equipment_repo: update_condition: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AssignRFID sets the RFID tag for an equipment item.
func (r *EquipmentRepo) AssignRFID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, rfidTag string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE equipment SET rfid_tag = $3, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID, rfidTag,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateRFID
		}
		return fmt.Errorf("equipment_repo: assign_rfid: %w", err)
	}
	return nil
}

// Deactivate soft-deletes an equipment item by setting is_active to false.
func (r *EquipmentRepo) Deactivate(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE equipment SET is_active = false, updated_at = NOW() WHERE id = $1 AND tenant_id = $2`,
		id, tenantID,
	)
	if err != nil {
		return fmt.Errorf("equipment_repo: deactivate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// Search performs a full-text search on equipment using German text search configuration.
func (r *EquipmentRepo) Search(ctx context.Context, tenantID uuid.UUID, query string, page int, perPage int) ([]*domain.Equipment, int64, error) {
	offset := (page - 1) * perPage

	ftsCondition := `tenant_id = $1 AND to_tsvector('german', name || ' ' || COALESCE(description,'')) @@ plainto_tsquery('german', $2)`

	// Count query.
	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM equipment WHERE %s`, ftsCondition)
	err := r.pool.QueryRow(ctx, countQuery, tenantID, query).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: search count: %w", err)
	}

	// Data query.
	dataQuery := fmt.Sprintf(
		`SELECT %s FROM equipment WHERE %s ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		equipmentColumns, ftsCondition,
	)

	rows, err := r.pool.Query(ctx, dataQuery, tenantID, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: search query: %w", err)
	}
	defer rows.Close()

	var items []*domain.Equipment
	for rows.Next() {
		e, scanErr := scanEquipment(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("equipment_repo: search scan: %w", scanErr)
		}
		items = append(items, e)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("equipment_repo: search rows: %w", err)
	}
	return items, total, nil
}

// nilIfEmpty returns nil if the string is empty, otherwise a pointer to it.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nilIfEmptyJSON returns nil if the JSON is empty or null, otherwise the raw bytes.
func nilIfEmptyJSON(j json.RawMessage) []byte {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	return []byte(j)
}
