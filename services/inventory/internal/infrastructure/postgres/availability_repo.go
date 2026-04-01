package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// AvailabilityRepo implements domain.AvailabilityRepository using PostgreSQL.
type AvailabilityRepo struct {
	pool *pgxpool.Pool
}

// NewAvailabilityRepo creates a new AvailabilityRepo.
func NewAvailabilityRepo(pool *pgxpool.Pool) *AvailabilityRepo {
	return &AvailabilityRepo{pool: pool}
}

// GetTypeAvailability returns per-type availability summaries.
//
// For each equipment type (within the tenant, excluding retired items) it counts:
//   - total_quantity: items with status != 'retired'
//   - reserved: items with status 'reserved' or 'checked_out'
//   - in_maintenance: items with status 'in_maintenance'
//   - available: total - reserved - in_maintenance
//
// When a category_id filter is provided, only equipment types within that
// category are returned. The date range is accepted for future use when
// project-based reservation queries become available.
func (r *AvailabilityRepo) GetTypeAvailability(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.AvailabilityFilter,
) ([]domain.TypeAvailabilitySummary, error) {
	// Build the query dynamically to support optional category filter.
	query := `
		SELECT
			COALESCE(e.equipment_type_id, '00000000-0000-0000-0000-000000000000') AS equipment_type_id,
			COALESCE(et.name, e.name) AS type_name,
			COUNT(*) FILTER (WHERE e.status != 'retired') AS total_quantity,
			COUNT(*) FILTER (WHERE e.status IN ('reserved', 'checked_out')) AS reserved,
			COUNT(*) FILTER (WHERE e.status = 'in_maintenance') AS in_maintenance
		FROM equipment e
		LEFT JOIN equipment_types et ON et.id = e.equipment_type_id AND et.tenant_id = e.tenant_id
		WHERE e.tenant_id = $1
		  AND e.is_active = true
		  AND e.status != 'retired'`

	args := []interface{}{tenantID}
	argIdx := 2

	if filter.CategoryID != nil {
		query += fmt.Sprintf(`
		  AND (e.category_id = $%d OR et.category_id = $%d)`, argIdx, argIdx)
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	query += `
		GROUP BY COALESCE(e.equipment_type_id, '00000000-0000-0000-0000-000000000000'), COALESCE(et.name, e.name)
		ORDER BY type_name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("availability_repo: get_type_availability: %w", err)
	}
	defer rows.Close()

	var results []domain.TypeAvailabilitySummary
	for rows.Next() {
		var s domain.TypeAvailabilitySummary
		if err := rows.Scan(
			&s.EquipmentTypeID,
			&s.Name,
			&s.TotalQuantity,
			&s.Reserved,
			&s.InMaintenance,
		); err != nil {
			return nil, fmt.Errorf("availability_repo: get_type_availability scan: %w", err)
		}

		s.Available = s.TotalQuantity - s.Reserved - s.InMaintenance
		if s.Available < 0 {
			s.Available = 0
		}

		if s.TotalQuantity > 0 {
			s.AvailabilityPct = (s.Available * 100) / s.TotalQuantity
		}

		results = append(results, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("availability_repo: get_type_availability rows: %w", err)
	}

	if results == nil {
		results = []domain.TypeAvailabilitySummary{}
	}

	return results, nil
}

// GetEquipmentBookings returns booking/reservation periods for a single
// equipment item overlapping the given date range.
//
// Currently this derives "bookings" from the equipment_history table by looking
// at status_changed events to 'reserved', 'checked_out', and back to 'available'.
// When a dedicated booking/reservation table exists, this can be swapped out.
//
// As a pragmatic first pass, if the item is currently reserved or checked_out,
// we return a single booking covering the requested range. History-based entries
// will be added when cross-service project data is available.
func (r *AvailabilityRepo) GetEquipmentBookings(
	ctx context.Context,
	tenantID uuid.UUID,
	equipmentID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.EquipmentBooking, error) {
	// Check current status first.
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT status FROM equipment WHERE id = $1 AND tenant_id = $2 AND is_active = true`,
		equipmentID, tenantID,
	).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("availability_repo: get_equipment_bookings status: %w", err)
	}

	var bookings []domain.EquipmentBooking

	// If the equipment is currently in a non-available state, return that as a booking.
	if status == domain.StatusReserved || status == domain.StatusCheckedOut || status == domain.StatusInMaintenance {
		bookings = append(bookings, domain.EquipmentBooking{
			EquipmentID: equipmentID,
			Status:      status,
			StartDate:   from,
			EndDate:     to,
		})
	}

	// Supplement with history-based booking records where we can find
	// checked_out / checked_in pairs within the requested range.
	historyQuery := `
		SELECT
			h.equipment_id,
			h.action,
			h.created_at
		FROM equipment_history h
		WHERE h.tenant_id = $1
		  AND h.equipment_id = $2
		  AND h.action IN ('checked_out', 'checked_in', 'status_changed')
		  AND h.created_at >= $3
		  AND h.created_at <= $4
		ORDER BY h.created_at ASC`

	rows, err := r.pool.Query(ctx, historyQuery, tenantID, equipmentID, from, to)
	if err != nil {
		// History query is supplemental; log but still return current-status booking.
		return bookings, nil
	}
	defer rows.Close()

	var pendingStart *time.Time
	for rows.Next() {
		var eqID uuid.UUID
		var action string
		var createdAt time.Time
		if scanErr := rows.Scan(&eqID, &action, &createdAt); scanErr != nil {
			continue
		}

		if action == domain.ActionCheckedOut && pendingStart == nil {
			t := createdAt
			pendingStart = &t
		} else if action == domain.ActionCheckedIn && pendingStart != nil {
			bookings = append(bookings, domain.EquipmentBooking{
				EquipmentID: equipmentID,
				Status:      domain.StatusCheckedOut,
				StartDate:   *pendingStart,
				EndDate:     createdAt,
			})
			pendingStart = nil
		}
	}

	// If there is an unclosed checkout, extend to the end of the range.
	if pendingStart != nil {
		bookings = append(bookings, domain.EquipmentBooking{
			EquipmentID: equipmentID,
			Status:      domain.StatusCheckedOut,
			StartDate:   *pendingStart,
			EndDate:     to,
		})
	}

	if bookings == nil {
		bookings = []domain.EquipmentBooking{}
	}

	return bookings, nil
}
