package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/transport/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateOrderRequest holds the data needed to create a transport order.
type CreateOrderRequest struct {
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty"`
	DriverID        *uuid.UUID `json:"driver_id,omitempty"`
	Type            string     `json:"type,omitempty"`
	PickupAddress   string     `json:"pickup_address,omitempty"`
	DeliveryAddress string     `json:"delivery_address,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	Notes           string     `json:"notes,omitempty"`
}

// UpdateOrderRequest holds the data needed to update a transport order.
type UpdateOrderRequest struct {
	ProjectID       *uuid.UUID `json:"project_id,omitempty"`
	VehicleID       *uuid.UUID `json:"vehicle_id,omitempty"`
	DriverID        *uuid.UUID `json:"driver_id,omitempty"`
	Type            string     `json:"type,omitempty"`
	Status          string     `json:"status,omitempty"`
	PickupAddress   string     `json:"pickup_address,omitempty"`
	DeliveryAddress string     `json:"delivery_address,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	Notes           string     `json:"notes,omitempty"`
}

// AddItemRequest holds the data needed to add an item to a transport order.
type AddItemRequest struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	Quantity    int       `json:"quantity,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// OrderService implements the application-level use cases for transport orders.
type OrderService struct {
	orderRepo domain.TransportOrderRepository
	itemRepo  domain.TransportItemRepository
	logger    zerolog.Logger
}

// NewOrderService constructs a new OrderService.
func NewOrderService(
	orderRepo domain.TransportOrderRepository,
	itemRepo domain.TransportItemRepository,
	logger zerolog.Logger,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		itemRepo:  itemRepo,
		logger:    logger.With().Str("service", "order").Logger(),
	}
}

// Create creates a new transport order.
func (s *OrderService) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	req CreateOrderRequest,
) (*domain.TransportOrder, error) {
	oType := req.Type
	if oType == "" {
		oType = domain.OrderTypeDelivery
	}
	if !domain.ValidOrderType(oType) {
		return nil, domain.ErrInvalidOrderType
	}

	order := &domain.TransportOrder{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ProjectID:       req.ProjectID,
		VehicleID:       req.VehicleID,
		DriverID:        req.DriverID,
		Type:            oType,
		Status:          domain.StatusPlanned,
		PickupAddress:   req.PickupAddress,
		DeliveryAddress: req.DeliveryAddress,
		ScheduledAt:     req.ScheduledAt,
		Notes:           req.Notes,
		CreatedBy:       &userID,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to create transport order")
		return nil, fmt.Errorf("create order: %w", err)
	}

	s.logger.Info().
		Str("order_id", order.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("transport order created")

	return order, nil
}

// List returns a filtered, paginated list of transport orders.
func (s *OrderService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	filter domain.TransportOrderFilter,
) ([]*domain.TransportOrder, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	items, total, err := s.orderRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error().Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("failed to list transport orders")
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	return items, total, nil
}

// GetByID returns a transport order by its ID within a tenant scope.
func (s *OrderService) GetByID(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.TransportOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("order_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to get transport order")
		return nil, fmt.Errorf("get order: %w", err)
	}
	return order, nil
}

// Update updates an existing transport order.
func (s *OrderService) Update(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
	req UpdateOrderRequest,
) (*domain.TransportOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	if req.ProjectID != nil {
		order.ProjectID = req.ProjectID
	}
	if req.VehicleID != nil {
		order.VehicleID = req.VehicleID
	}
	if req.DriverID != nil {
		order.DriverID = req.DriverID
	}
	if req.Type != "" {
		if !domain.ValidOrderType(req.Type) {
			return nil, domain.ErrInvalidOrderType
		}
		order.Type = req.Type
	}
	if req.Status != "" {
		if !domain.ValidStatus(req.Status) {
			return nil, domain.ErrInvalidStatus
		}
		order.Status = req.Status
	}
	if req.PickupAddress != "" {
		order.PickupAddress = req.PickupAddress
	}
	if req.DeliveryAddress != "" {
		order.DeliveryAddress = req.DeliveryAddress
	}
	if req.ScheduledAt != nil {
		order.ScheduledAt = req.ScheduledAt
	}
	if req.Notes != "" {
		order.Notes = req.Notes
	}

	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.Error().Err(err).
			Str("order_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to update transport order")
		return nil, fmt.Errorf("update order: %w", err)
	}

	s.logger.Info().
		Str("order_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("transport order updated")

	return order, nil
}

// Complete marks a transport order as completed.
func (s *OrderService) Complete(
	ctx context.Context,
	id uuid.UUID,
	tenantID uuid.UUID,
) (*domain.TransportOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("complete order: %w", err)
	}

	if order.Status == domain.StatusCompleted {
		return nil, domain.ErrOrderAlreadyCompleted
	}

	now := time.Now()
	order.Status = domain.StatusCompleted
	order.CompletedAt = &now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		s.logger.Error().Err(err).
			Str("order_id", id.String()).
			Str("tenant_id", tenantID.String()).
			Msg("failed to complete transport order")
		return nil, fmt.Errorf("complete order: %w", err)
	}

	s.logger.Info().
		Str("order_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("transport order completed")

	return order, nil
}

// AddItem adds an equipment item to a transport order.
func (s *OrderService) AddItem(
	ctx context.Context,
	orderID uuid.UUID,
	tenantID uuid.UUID,
	req AddItemRequest,
) (*domain.TransportItem, error) {
	if req.EquipmentID == uuid.Nil {
		return nil, domain.ErrMissingEquipmentID
	}

	// Verify order exists and belongs to tenant.
	_, err := s.orderRepo.GetByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}

	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	item := &domain.TransportItem{
		ID:          uuid.New(),
		OrderID:     orderID,
		EquipmentID: req.EquipmentID,
		Quantity:    quantity,
		Notes:       req.Notes,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		s.logger.Error().Err(err).
			Str("order_id", orderID.String()).
			Str("equipment_id", req.EquipmentID.String()).
			Msg("failed to add transport item")
		return nil, fmt.Errorf("add item: %w", err)
	}

	s.logger.Info().
		Str("item_id", item.ID.String()).
		Str("order_id", orderID.String()).
		Msg("transport item added")

	return item, nil
}

// ListItems returns all items for a transport order.
func (s *OrderService) ListItems(
	ctx context.Context,
	orderID uuid.UUID,
	tenantID uuid.UUID,
) ([]*domain.TransportItem, error) {
	// Verify order exists and belongs to tenant.
	_, err := s.orderRepo.GetByID(ctx, orderID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	items, err := s.itemRepo.ListByOrder(ctx, orderID)
	if err != nil {
		s.logger.Error().Err(err).
			Str("order_id", orderID.String()).
			Msg("failed to list transport items")
		return nil, fmt.Errorf("list items: %w", err)
	}
	return items, nil
}
