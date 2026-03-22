package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/insurance-service/internal/domain"
)

// ClaimItemRepository implements the ports.ClaimItemRepository interface
type ClaimItemRepository struct {
	db *sql.DB
}

// NewClaimItemRepository creates a new claim item repository
func NewClaimItemRepository(db *sql.DB) *ClaimItemRepository {
	return &ClaimItemRepository{db: db}
}

// StringArray handles JSONB array of strings
type StringArray []string

// Scan implements sql.Scanner
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = []string{}
		return nil
	}

	data, ok := value.([]byte)
	if !ok {
		return errors.New("invalid string array type")
	}

	return json.Unmarshal(data, (*[]string)(sa))
}

// Value implements driver.Valuer
func (sa StringArray) Value() (driver.Value, error) {
	return json.Marshal(sa)
}

// Create inserts a new claim item
func (cr *ClaimItemRepository) Create(ctx context.Context, item *domain.ClaimItem) error {
	query := `
		INSERT INTO claim_items 
		(id, claim_id, equipment_id, description, replacement_value, repair_cost, photo_urls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	photoURLs := StringArray(item.PhotoURLs)
	_, err := cr.db.ExecContext(ctx, query,
		item.ID, item.ClaimID, item.EquipmentID, item.Description,
		item.ReplacementValue, item.RepairCost, photoURLs, item.CreatedAt,
	)

	return err
}

// GetByID retrieves a claim item by ID
func (cr *ClaimItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClaimItem, error) {
	query := `
		SELECT id, claim_id, equipment_id, description, replacement_value, repair_cost, photo_urls, created_at
		FROM claim_items WHERE id = $1
	`

	var item domain.ClaimItem
	var photoURLs StringArray

	err := cr.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.ClaimID, &item.EquipmentID, &item.Description,
		&item.ReplacementValue, &item.RepairCost, &photoURLs, &item.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.ErrClaimItemNotFound{ID: id.String()}
		}
		return nil, err
	}

	item.PhotoURLs = photoURLs
	return &item, nil
}

// GetByClaimID retrieves all items for a claim
func (cr *ClaimItemRepository) GetByClaimID(ctx context.Context, claimID uuid.UUID) ([]*domain.ClaimItem, error) {
	query := `
		SELECT id, claim_id, equipment_id, description, replacement_value, repair_cost, photo_urls, created_at
		FROM claim_items WHERE claim_id = $1 ORDER BY created_at DESC
	`

	rows, err := cr.db.QueryContext(ctx, query, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.ClaimItem
	for rows.Next() {
		var item domain.ClaimItem
		var photoURLs StringArray

		if err := rows.Scan(
			&item.ID, &item.ClaimID, &item.EquipmentID, &item.Description,
			&item.ReplacementValue, &item.RepairCost, &photoURLs, &item.CreatedAt,
		); err != nil {
			return nil, err
		}

		item.PhotoURLs = photoURLs
		items = append(items, &item)
	}

	return items, rows.Err()
}

// Update updates an existing claim item
func (cr *ClaimItemRepository) Update(ctx context.Context, item *domain.ClaimItem) error {
	query := `
		UPDATE claim_items 
		SET description = $1, replacement_value = $2, repair_cost = $3, photo_urls = $4
		WHERE id = $5
	`

	photoURLs := StringArray(item.PhotoURLs)
	_, err := cr.db.ExecContext(ctx, query,
		item.Description, item.ReplacementValue, item.RepairCost, photoURLs, item.ID,
	)

	return err
}

// Delete deletes a claim item
func (cr *ClaimItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := cr.db.ExecContext(ctx, "DELETE FROM claim_items WHERE id = $1", id)
	return err
}
