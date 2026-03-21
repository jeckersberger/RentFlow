package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/project-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/project-service/internal/ports"
)

type ReservationService struct {
	repo   ports.ReservationRepository
	logger logger.Logger
}

func NewReservationService(repo ports.ReservationRepository, logger logger.Logger) *ReservationService {
	return &ReservationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ReservationService) CreateReservation(ctx context.Context, cmd CreateReservationCommand) (*ReservationDTO, error) {
	if cmd.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if cmd.ProjectID == "" {
		return nil, domain.NewDomainError("PROJECT_REQUIRED", "project ID is required", nil)
	}
	if cmd.EquipmentID == "" {
		return nil, domain.NewDomainError("EQUIPMENT_REQUIRED", "equipment ID is required", nil)
	}

	// Check for conflicts
	conflicts, err := s.repo.ListByEquipmentID(ctx, cmd.TenantID, cmd.EquipmentID,
		cmd.StartDate.Format(time.RFC3339), cmd.EndDate.Format(time.RFC3339))
	if err == nil && len(conflicts) > 0 {
		return nil, domain.NewDomainError("CONFLICT", "equipment has conflicting reservations", nil)
	}

	reservationID := fmt.Sprintf("resv_%d", hashString(cmd.TenantID+cmd.EquipmentID+cmd.StartDate.String()))

	reservation := domain.NewReservation(reservationID, cmd.TenantID, cmd.ProjectID, cmd.EquipmentID, cmd.StartDate, cmd.EndDate)

	if err := reservation.Validate(); err != nil {
		return nil, domain.NewDomainError("VALIDATION_ERROR", err.Error(), nil)
	}

	if err := s.repo.Create(ctx, reservation); err != nil {
		return nil, domain.NewDomainError("CREATE_ERROR", "failed to create reservation", err)
	}

	s.logger.Info("Reservation created", "id", reservation.ID, "tenant_id", cmd.TenantID)
	return ReservationToDTO(reservation), nil
}

func (s *ReservationService) GetReservation(ctx context.Context, tenantID, reservationID string) (*ReservationDTO, error) {
	if tenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	reservation, err := s.repo.GetByID(ctx, tenantID, reservationID)
	if err != nil {
		return nil, domain.NewDomainError("NOT_FOUND", "reservation not found", err)
	}

	return ReservationToDTO(reservation), nil
}

func (s *ReservationService) ListByProject(ctx context.Context, query ListReservationsQuery) ([]*ReservationDTO, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if query.ProjectID == "" {
		return nil, domain.NewDomainError("PROJECT_REQUIRED", "project ID is required", nil)
	}

	reservations, err := s.repo.ListByProjectID(ctx, query.TenantID, query.ProjectID)
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to list reservations", err)
	}

	dtos := make([]*ReservationDTO, len(reservations))
	for i, reservation := range reservations {
		dtos[i] = ReservationToDTO(reservation)
	}

	return dtos, nil
}

func (s *ReservationService) CheckConflicts(ctx context.Context, query CheckReservationConflictQuery) ([]*ReservationDTO, error) {
	if query.TenantID == "" {
		return nil, domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}
	if query.EquipmentID == "" {
		return nil, domain.NewDomainError("EQUIPMENT_REQUIRED", "equipment ID is required", nil)
	}

	reservations, err := s.repo.ListByEquipmentID(ctx, query.TenantID, query.EquipmentID,
		query.StartDate.Format(time.RFC3339), query.EndDate.Format(time.RFC3339))
	if err != nil {
		return nil, domain.NewDomainError("QUERY_ERROR", "failed to check conflicts", err)
	}

	dtos := make([]*ReservationDTO, len(reservations))
	for i, reservation := range reservations {
		dtos[i] = ReservationToDTO(reservation)
	}

	return dtos, nil
}

func (s *ReservationService) ConfirmReservation(ctx context.Context, cmd ConfirmReservationCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	reservation, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "reservation not found", err)
	}

	if err := reservation.Confirm(); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, reservation); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to confirm reservation", err)
	}

	s.logger.Info("Reservation confirmed", "id", cmd.ID)
	return nil
}

func (s *ReservationService) CancelReservation(ctx context.Context, cmd CancelReservationCommand) error {
	if cmd.TenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	reservation, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.ID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "reservation not found", err)
	}

	if err := reservation.Cancel(); err != nil {
		return domain.NewDomainError("INVALID_STATUS", err.Error(), nil)
	}

	if err := s.repo.Update(ctx, reservation); err != nil {
		return domain.NewDomainError("UPDATE_ERROR", "failed to cancel reservation", err)
	}

	s.logger.Info("Reservation cancelled", "id", cmd.ID, "reason", cmd.Reason)
	return nil
}

func (s *ReservationService) DeleteReservation(ctx context.Context, tenantID, reservationID string) error {
	if tenantID == "" {
		return domain.NewDomainError("TENANT_REQUIRED", "tenant ID is required", nil)
	}

	_, err := s.repo.GetByID(ctx, tenantID, reservationID)
	if err != nil {
		return domain.NewDomainError("NOT_FOUND", "reservation not found", err)
	}

	if err := s.repo.Delete(ctx, tenantID, reservationID); err != nil {
		return domain.NewDomainError("DELETE_ERROR", "failed to delete reservation", err)
	}

	s.logger.Info("Reservation deleted", "id", reservationID, "tenant_id", tenantID)
	return nil
}
