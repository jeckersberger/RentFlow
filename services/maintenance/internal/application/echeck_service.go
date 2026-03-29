package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/maintenance/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateECheckRequest holds the data needed to record an E-Check result.
type CreateECheckRequest struct {
	EquipmentID                   uuid.UUID `json:"equipment_id"`
	CheckDate                     string    `json:"check_date"`
	NextCheckDate                 *string   `json:"next_check_date,omitempty"`
	PerformedBy                   string    `json:"performed_by,omitempty"`
	MeasuringDevice               string    `json:"measuring_device,omitempty"`
	ProtectionConductorResistance *float64  `json:"protection_conductor_resistance,omitempty"`
	InsulationResistance          *float64  `json:"insulation_resistance,omitempty"`
	LeakageCurrent                *float64  `json:"leakage_current,omitempty"`
	Notes                         string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// ECheckService implements the application-level use cases for E-Check records.
type ECheckService struct {
	repo   domain.ECheckRepository
	logger zerolog.Logger
}

// NewECheckService constructs a new ECheckService.
func NewECheckService(repo domain.ECheckRepository, logger zerolog.Logger) *ECheckService {
	return &ECheckService{
		repo:   repo,
		logger: logger.With().Str("service", "echeck").Logger(),
	}
}

// evaluateResult determines the E-Check result based on measured values.
// Thresholds per DGUV V3:
//   - Protection conductor resistance < 0.3 Ohm
//   - Insulation resistance > 1.0 MOhm
//   - Leakage current < 3.5 mA
func evaluateResult(pcr, ir, lc *float64) string {
	if pcr == nil && ir == nil && lc == nil {
		return "pending"
	}

	if pcr != nil && *pcr >= 0.3 {
		return "failed"
	}
	if ir != nil && *ir <= 1.0 {
		return "failed"
	}
	if lc != nil && *lc >= 3.5 {
		return "failed"
	}

	return "passed"
}

// Create records a new E-Check result with automatic pass/fail evaluation.
func (s *ECheckService) Create(ctx context.Context, tenantID uuid.UUID, req CreateECheckRequest) (*domain.ECheckRecord, error) {
	if req.EquipmentID == uuid.Nil {
		return nil, domain.ErrEquipmentRequired
	}
	if req.CheckDate == "" {
		return nil, domain.ErrCheckDateRequired
	}

	result := evaluateResult(
		req.ProtectionConductorResistance,
		req.InsulationResistance,
		req.LeakageCurrent,
	)

	record := &domain.ECheckRecord{
		ID:                            uuid.New(),
		TenantID:                      tenantID,
		EquipmentID:                   req.EquipmentID,
		CheckDate:                     req.CheckDate,
		NextCheckDate:                 req.NextCheckDate,
		Result:                        result,
		PerformedBy:                   req.PerformedBy,
		MeasuringDevice:               req.MeasuringDevice,
		ProtectionConductorResistance: req.ProtectionConductorResistance,
		InsulationResistance:          req.InsulationResistance,
		LeakageCurrent:                req.LeakageCurrent,
		Notes:                         req.Notes,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		s.logger.Error().Err(err).
			Str("equipment_id", req.EquipmentID.String()).
			Msg("failed to create echeck record")
		return nil, fmt.Errorf("create echeck: %w", err)
	}

	s.logger.Info().
		Str("echeck_id", record.ID.String()).
		Str("result", result).
		Msg("echeck record created")

	return record, nil
}

// List returns a paginated list of E-Check records for a tenant.
func (s *ECheckService) List(ctx context.Context, tenantID uuid.UUID, filter domain.ECheckFilter) ([]*domain.ECheckRecord, int64, error) {
	return s.repo.List(ctx, tenantID, filter)
}

// ListOverdue returns all E-Check records where the next check date is in the past.
func (s *ECheckService) ListOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.ECheckRecord, error) {
	return s.repo.ListOverdue(ctx, tenantID)
}

// ListByEquipment returns the E-Check history for a specific equipment item.
func (s *ECheckService) ListByEquipment(ctx context.Context, equipmentID uuid.UUID, tenantID uuid.UUID) ([]*domain.ECheckRecord, error) {
	if equipmentID == uuid.Nil {
		return nil, domain.ErrEquipmentRequired
	}
	return s.repo.ListByEquipment(ctx, equipmentID, tenantID)
}
