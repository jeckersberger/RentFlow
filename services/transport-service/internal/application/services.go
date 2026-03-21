package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/ports"
)

type VehicleService struct {
	repo   ports.VehicleRepository
	logger *logger.Logger
}

func NewVehicleService(repo ports.VehicleRepository, log *logger.Logger) *VehicleService {
	return &VehicleService{repo: repo, logger: log}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, cmd CreateVehicleCommand) (*VehicleDTO, error) {
	if cmd.TenantID == "" || cmd.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	vehicle := &domain.Vehicle{
		ID:           fmt.Sprintf("veh_%d", time.Now().UnixNano()),
		TenantID:     cmd.TenantID,
		Name:         cmd.Name,
		LicensePlate: cmd.LicensePlate,
		Type:         domain.VehicleType(cmd.Type),
		Capacity:     cmd.Capacity,
		Status:       domain.VehicleStatusAvailable,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, vehicle); err != nil {
		s.logger.Error("Failed to create vehicle", err)
		return nil, err
	}

	return VehicleToDTO(vehicle), nil
}

func (s *VehicleService) GetVehicle(ctx context.Context, tenantID, id string) (*VehicleDTO, error) {
	vehicle, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, domain.ErrVehicleNotFound
	}
	return VehicleToDTO(vehicle), nil
}

func (s *VehicleService) ListVehicles(ctx context.Context, tenantID string) ([]*VehicleDTO, error) {
	vehicles, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*VehicleDTO, len(vehicles))
	for i, v := range vehicles {
		dtos[i] = VehicleToDTO(v)
	}
	return dtos, nil
}

func (s *VehicleService) UpdateVehicle(ctx context.Context, cmd UpdateVehicleCommand) (*VehicleDTO, error) {
	vehicle, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, domain.ErrVehicleNotFound
	}

	vehicle.Status = domain.VehicleStatus(cmd.Status)
	vehicle.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, vehicle); err != nil {
		s.logger.Error("Failed to update vehicle", err)
		return nil, err
	}

	return VehicleToDTO(vehicle), nil
}

type TourService struct {
	tourRepo    ports.TourRepository
	vehicleRepo ports.VehicleRepository
	logger      *logger.Logger
}

func NewTourService(tourRepo ports.TourRepository, vehicleRepo ports.VehicleRepository, log *logger.Logger) *TourService {
	return &TourService{tourRepo: tourRepo, vehicleRepo: vehicleRepo, logger: log}
}

func (s *TourService) CreateTour(ctx context.Context, cmd CreateTourCommand) (*TourDTO, error) {
	if cmd.TenantID == "" || cmd.VehicleID == "" {
		return nil, domain.ErrInvalidInput
	}

	stops := make([]domain.TourStop, len(cmd.Stops))
	for i, stop := range cmd.Stops {
		stops[i] = domain.TourStop{
			Address:       stop["address"].(string),
			ArrivalTime:   stop["arrival_time"].(time.Time),
			DepartureTime: stop["departure_time"].(time.Time),
			Type:          stop["type"].(string),
		}
	}

	tour := &domain.Tour{
		ID:        fmt.Sprintf("tour_%d", time.Now().UnixNano()),
		TenantID:  cmd.TenantID,
		ProjectID: cmd.ProjectID,
		VehicleID: cmd.VehicleID,
		DriverID:  cmd.DriverID,
		Date:      cmd.Date,
		Stops:     stops,
		Status:    domain.TourStatusPlanned,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.tourRepo.Create(ctx, tour); err != nil {
		s.logger.Error("Failed to create tour", err)
		return nil, err
	}

	return TourToDTO(tour), nil
}

func (s *TourService) GetTour(ctx context.Context, tenantID, id string) (*TourDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}
	return TourToDTO(tour), nil
}

func (s *TourService) ListTours(ctx context.Context, tenantID string) ([]*TourDTO, error) {
	tours, err := s.tourRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*TourDTO, len(tours))
	for i, t := range tours {
		dtos[i] = TourToDTO(t)
	}
	return dtos, nil
}

func (s *TourService) ListToursByDate(ctx context.Context, tenantID string, date time.Time) ([]*TourDTO, error) {
	tours, err := s.tourRepo.ListByDate(ctx, tenantID, date)
	if err != nil {
		return nil, err
	}
	dtos := make([]*TourDTO, len(tours))
	for i, t := range tours {
		dtos[i] = TourToDTO(t)
	}
	return dtos, nil
}

func (s *TourService) UpdateTourStatus(ctx context.Context, cmd UpdateTourStatusCommand) (*TourDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	tour.Status = domain.TourStatus(cmd.Status)
	tour.UpdatedAt = time.Now()

	if err := s.tourRepo.Update(ctx, tour); err != nil {
		s.logger.Error("Failed to update tour", err)
		return nil, err
	}

	return TourToDTO(tour), nil
}
