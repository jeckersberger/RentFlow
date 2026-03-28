package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/project/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

type CreateReservationRequest struct {
	ProjectID   uuid.UUID `json:"project_id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

type UpdateReservationStatusRequest struct {
	Status string `json:"status"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

type ReservationService struct {
	reservationRepo domain.ReservationRepository
	projectRepo     domain.ProjectRepository
	logger          zerolog.Logger
}

func NewReservationService(
	reservationRepo domain.ReservationRepository,
	projectRepo domain.ProjectRepository,
	logger zerolog.Logger,
) *ReservationService {
	return &ReservationService{
		reservationRepo: reservationRepo,
		projectRepo:     projectRepo,
		logger:          logger.With().Str("service", "reservation").Logger(),
	}
}

func (s *ReservationService) Create(ctx context.Context, tenantID uuid.UUID, req CreateReservationRequest) (*domain.Reservation, error) {
	// Verify project exists and belongs to tenant
	if _, err := s.projectRepo.GetByID(ctx, req.ProjectID, tenantID); err != nil {
		return nil, fmt.Errorf("create reservation – verify project: %w", err)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format (YYYY-MM-DD): %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format (YYYY-MM-DD): %w", err)
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	if req.Quantity < 1 {
		req.Quantity = 1
	}

	// Check for conflicts
	existing, err := s.reservationRepo.ListByEquipmentAndDateRange(ctx, req.EquipmentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("check reservation conflicts: %w", err)
	}

	totalReserved := 0
	for _, r := range existing {
		if r.Status != domain.ReservationStatusCancelled {
			totalReserved += r.Quantity
		}
	}
	// Note: We don't know the equipment's total quantity here (cross-service).
	// For now, just record the reservation. Conflict detection with actual
	// availability requires the inventory service and will be refined later.

	reservation := &domain.Reservation{
		ID:          uuid.New(),
		TenantID:    tenantID,
		ProjectID:   req.ProjectID,
		EquipmentID: req.EquipmentID,
		Quantity:    req.Quantity,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      domain.ReservationStatusPending,
		CreatedAt:   time.Now(),
	}

	_ = totalReserved // will be used for cross-service conflict check

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, fmt.Errorf("create reservation: %w", err)
	}

	s.logger.Info().Str("reservation_id", reservation.ID.String()).Msg("reservation created")
	return reservation, nil
}

func (s *ReservationService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Reservation, error) {
	r, err := s.reservationRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	return r, nil
}

func (s *ReservationService) List(ctx context.Context, tenantID uuid.UUID) ([]*domain.Reservation, error) {
	items, err := s.reservationRepo.List(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	return items, nil
}

func (s *ReservationService) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, req UpdateReservationStatusRequest) error {
	if err := s.reservationRepo.UpdateStatus(ctx, id, tenantID, req.Status); err != nil {
		return fmt.Errorf("update reservation status: %w", err)
	}

	s.logger.Info().Str("reservation_id", id.String()).Str("status", req.Status).Msg("reservation status updated")
	return nil
}
