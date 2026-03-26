package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/domain"
	"github.com/jeckersberger/rentflow/services/transport-service/internal/ports"
)

type VehicleService struct {
	repo   ports.VehicleRepository
	logger logger.Logger
}

func NewVehicleService(repo ports.VehicleRepository, log logger.Logger) *VehicleService {
	return &VehicleService{repo: repo, logger: log}
}

func (s *VehicleService) CreateVehicle(ctx context.Context, cmd CreateVehicleCommand) (*VehicleDTO, error) {
	if cmd.TenantID == "" || cmd.Name == "" || cmd.LicensePlate == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now()
	vehicle := &domain.Vehicle{
		ID:            uuid.New().String(),
		TenantID:      cmd.TenantID,
		Name:          cmd.Name,
		LicensePlate:  cmd.LicensePlate,
		CapacityKg:    cmd.CapacityKg,
		CapacityM3:    cmd.CapacityM3,
		VehicleType:   domain.VehicleType(cmd.VehicleType),
		Status:        domain.VehicleStatusAvailable,
		DGUVLastCheck: cmd.DGUVLastCheck,
		DGUVNextCheck: cmd.DGUVNextCheck,
		CreatedAt:     now,
		UpdatedAt:     now,
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
	if cmd.DGUVLastCheck != nil {
		vehicle.DGUVLastCheck = cmd.DGUVLastCheck
	}
	if cmd.DGUVNextCheck != nil {
		vehicle.DGUVNextCheck = cmd.DGUVNextCheck
	}
	vehicle.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, vehicle); err != nil {
		s.logger.Error("Failed to update vehicle", err)
		return nil, err
	}

	return VehicleToDTO(vehicle), nil
}

type TourService struct {
	tourRepo           ports.TourRepository
	vehicleRepo        ports.VehicleRepository
	equipmentRepo      ports.TourEquipmentRepository
	driverLogRepo      ports.DriverLogRepository
	documentClient     ports.DocumentClient
	logger             logger.Logger
}

func NewTourService(tourRepo ports.TourRepository, vehicleRepo ports.VehicleRepository,
	equipmentRepo ports.TourEquipmentRepository, driverLogRepo ports.DriverLogRepository,
	documentClient ports.DocumentClient, log logger.Logger) *TourService {
	return &TourService{
		tourRepo:       tourRepo,
		vehicleRepo:    vehicleRepo,
		equipmentRepo:  equipmentRepo,
		driverLogRepo:  driverLogRepo,
		documentClient: documentClient,
		logger:         log,
	}
}

func (s *TourService) CreateTour(ctx context.Context, cmd CreateTourCommand) (*TourDTO, error) {
	if cmd.TenantID == "" || cmd.VehicleID == "" {
		return nil, domain.ErrInvalidInput
	}

	// Verify vehicle exists
	vehicle, err := s.vehicleRepo.GetByID(ctx, cmd.TenantID, cmd.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, domain.ErrVehicleNotFound
	}

	now := time.Now()
	tour := &domain.Tour{
		ID:          uuid.New().String(),
		TenantID:    cmd.TenantID,
		ProjectID:   cmd.ProjectID,
		VehicleID:   cmd.VehicleID,
		DriverID:    cmd.DriverID,
		Status:      domain.TourStatusPlanned,
		DepartureAt: cmd.DepartureAt,
		Notes:       cmd.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
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

func (s *TourService) StartTour(ctx context.Context, cmd StartTourCommand) (*TourDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	tour.Status = domain.TourStatusInTransit
	tour.KmStart = cmd.KmStart
	now := time.Now()
	tour.DepartureAt = &now
	tour.UpdatedAt = now

	if err := s.tourRepo.Update(ctx, tour); err != nil {
		s.logger.Error("Failed to start tour", err)
		return nil, err
	}

	return TourToDTO(tour), nil
}

func (s *TourService) CompleteTour(ctx context.Context, cmd CompleteTourCommand) (*TourDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	tour.Status = domain.TourStatusCompleted
	tour.KmEnd = cmd.KmEnd
	tour.TotalCost = cmd.TotalCost
	now := time.Now()
	tour.ArrivalAt = &now
	tour.UpdatedAt = now

	if err := s.tourRepo.Update(ctx, tour); err != nil {
		s.logger.Error("Failed to complete tour", err)
		return nil, err
	}

	return TourToDTO(tour), nil
}

func (s *TourService) AddEquipmentToTour(ctx context.Context, cmd AddEquipmentToTourCommand) (*TourEquipmentDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	// Get vehicle to check capacity
	vehicle, err := s.vehicleRepo.GetByID(ctx, cmd.TenantID, tour.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, domain.ErrVehicleNotFound
	}

	// Check capacity
	totalWeight, totalVolume, err := s.equipmentRepo.GetCapacityByTour(ctx, cmd.TourID)
	if err != nil {
		return nil, err
	}

	if totalWeight+cmd.WeightKg > vehicle.CapacityKg || totalVolume+cmd.VolumeM3 > vehicle.CapacityM3 {
		return nil, domain.ErrCapacityExceeded
	}

	now := time.Now()
	equipment := &domain.TourEquipment{
		ID:          uuid.New().String(),
		TourID:      cmd.TourID,
		EquipmentID: cmd.EquipmentID,
		WeightKg:    cmd.WeightKg,
		VolumeM3:    cmd.VolumeM3,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.equipmentRepo.Create(ctx, equipment); err != nil {
		s.logger.Error("Failed to add equipment to tour", err)
		return nil, err
	}

	return TourEquipmentToDTO(equipment), nil
}

func (s *TourService) RemoveEquipmentFromTour(ctx context.Context, cmd RemoveEquipmentFromTourCommand) error {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return err
	}
	if tour == nil {
		return domain.ErrTourNotFound
	}

	if err := s.equipmentRepo.DeleteByTourAndEquipment(ctx, cmd.TourID, cmd.EquipmentID); err != nil {
		s.logger.Error("Failed to remove equipment from tour", err)
		return err
	}

	return nil
}

func (s *TourService) GetTourCapacity(ctx context.Context, tenantID, tourID string) (*TourCapacityDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, tenantID, tourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	vehicle, err := s.vehicleRepo.GetByID(ctx, tenantID, tour.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, domain.ErrVehicleNotFound
	}

	totalWeight, totalVolume, err := s.equipmentRepo.GetCapacityByTour(ctx, tourID)
	if err != nil {
		return nil, err
	}

	exceeded := totalWeight > vehicle.CapacityKg || totalVolume > vehicle.CapacityM3

	return &TourCapacityDTO{
		TourID:           tourID,
		TotalCapacityKg:  vehicle.CapacityKg,
		TotalCapacityM3:  vehicle.CapacityM3,
		UsedCapacityKg:   totalWeight,
		UsedCapacityM3:   totalVolume,
		RemainingKg:      vehicle.CapacityKg - totalWeight,
		RemainingM3:      vehicle.CapacityM3 - totalVolume,
		CapacityExceeded: exceeded,
	}, nil
}

func (s *TourService) LogDriverActivity(ctx context.Context, cmd LogDriverActivityCommand) (*DriverLogDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, cmd.TenantID, cmd.TourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	now := time.Now()
	log := &domain.DriverLog{
		ID:            uuid.New().String(),
		TourID:        cmd.TourID,
		DriverID:      cmd.DriverID,
		StartTime:     cmd.StartTime,
		EndTime:       cmd.EndTime,
		BreakMinutes:  cmd.BreakMinutes,
		KmDriven:      cmd.KmDriven,
		Notes:         cmd.Notes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.driverLogRepo.Create(ctx, log); err != nil {
		s.logger.Error("Failed to log driver activity", err)
		return nil, err
	}

	return DriverLogToDTO(log), nil
}

func (s *TourService) GetDriverLogs(ctx context.Context, tenantID, tourID string) ([]*DriverLogDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, tenantID, tourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	logs, err := s.driverLogRepo.ListByTour(ctx, tourID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*DriverLogDTO, len(logs))
	for i, l := range logs {
		dtos[i] = DriverLogToDTO(l)
	}
	return dtos, nil
}

func (s *TourService) GenerateDeliveryNote(ctx context.Context, tenantID, tourID string) (*TourDTO, error) {
	tour, err := s.tourRepo.GetByID(ctx, tenantID, tourID)
	if err != nil {
		return nil, err
	}
	if tour == nil {
		return nil, domain.ErrTourNotFound
	}

	// Get all equipment for the tour
	equipmentList, err := s.equipmentRepo.ListByTour(ctx, tourID)
	if err != nil {
		s.logger.Error("Failed to get tour equipment", err, "tourID", tourID)
		return nil, err
	}

	// Convert equipment to delivery note items
	items := make([]ports.DeliveryNoteItem, len(equipmentList))
	for i, eq := range equipmentList {
		items[i] = ports.DeliveryNoteItem{
			EquipmentID: eq.EquipmentID,
			WeightKg:    eq.WeightKg,
			VolumeM3:    eq.VolumeM3,
		}
	}

	// Call document-service to generate delivery note
	result, err := s.documentClient.GenerateDeliveryNote(ctx, tenantID, tourID, items)
	if err != nil {
		s.logger.Error("Failed to generate delivery note", err, "tourID", tourID)
		return nil, err
	}

	// Update tour with delivery note information
	now := time.Now()
	tour.DeliveryNoteNumber = &result.DeliveryNoteNumber
	tour.DeliveryNoteGeneratedAt = &now
	tour.UpdatedAt = now

	if err := s.tourRepo.Update(ctx, tour); err != nil {
		s.logger.Error("Failed to update tour with delivery note", err, "tourID", tourID)
		return nil, err
	}

	return TourToDTO(tour), nil
}
