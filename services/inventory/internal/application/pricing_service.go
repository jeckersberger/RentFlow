package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// ---------------------------------------------------------------------------
// Request / Response DTOs
// ---------------------------------------------------------------------------

// PriceCalculationRequest holds parameters for a price calculation.
type PriceCalculationRequest struct {
	EquipmentID uuid.UUID `json:"equipment_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Days        int       `json:"days"`
	Quantity    int       `json:"quantity"`
	Date        string    `json:"date"` // ISO format YYYY-MM-DD
}

// PriceCalculationResult returns a detailed price breakdown.
type PriceCalculationResult struct {
	BasePricePerDay int64  `json:"base_price_per_day"`
	Days            int    `json:"days"`
	Quantity        int    `json:"quantity"`
	SubtotalBase    int64  `json:"subtotal_base"`
	TierDiscount    int64  `json:"tier_discount"`
	QtyDiscount     int64  `json:"qty_discount"`
	SeasonSurcharge int64  `json:"season_surcharge"`
	TotalNet        int64  `json:"total_net"`
	AppliedRule     string `json:"applied_rule"`
}

// CreatePriceRuleRequest holds the data to create a new price rule.
type CreatePriceRuleRequest struct {
	EquipmentID          *uuid.UUID `json:"equipment_id,omitempty"`
	CategoryID           *uuid.UUID `json:"category_id,omitempty"`
	BasePriceDay         int64      `json:"base_price_day"`
	BasePriceWeek        *int64     `json:"base_price_week,omitempty"`
	Tier2FromDays        *int       `json:"tier2_from_days,omitempty"`
	Tier2PriceDay        *int64     `json:"tier2_price_day,omitempty"`
	Tier3FromDays        *int       `json:"tier3_from_days,omitempty"`
	Tier3PriceDay        *int64     `json:"tier3_price_day,omitempty"`
	QtyDiscountThreshold *int       `json:"qty_discount_threshold,omitempty"`
	QtyDiscountPct       int        `json:"qty_discount_pct"`
	SeasonStart          *string    `json:"season_start,omitempty"`
	SeasonEnd            *string    `json:"season_end,omitempty"`
	SeasonSurchargePct   int        `json:"season_surcharge_pct"`
}

// UpdatePriceRuleRequest holds optional fields for updating a price rule.
type UpdatePriceRuleRequest struct {
	EquipmentID          *uuid.UUID `json:"equipment_id,omitempty"`
	CategoryID           *uuid.UUID `json:"category_id,omitempty"`
	BasePriceDay         *int64     `json:"base_price_day,omitempty"`
	BasePriceWeek        *int64     `json:"base_price_week,omitempty"`
	Tier2FromDays        *int       `json:"tier2_from_days,omitempty"`
	Tier2PriceDay        *int64     `json:"tier2_price_day,omitempty"`
	Tier3FromDays        *int       `json:"tier3_from_days,omitempty"`
	Tier3PriceDay        *int64     `json:"tier3_price_day,omitempty"`
	QtyDiscountThreshold *int       `json:"qty_discount_threshold,omitempty"`
	QtyDiscountPct       *int       `json:"qty_discount_pct,omitempty"`
	SeasonStart          *string    `json:"season_start,omitempty"`
	SeasonEnd            *string    `json:"season_end,omitempty"`
	SeasonSurchargePct   *int       `json:"season_surcharge_pct,omitempty"`
	IsActive             *bool      `json:"is_active,omitempty"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// PricingService implements price calculation use cases.
type PricingService struct {
	priceRuleRepo domain.PriceRuleRepository
	equipmentRepo domain.EquipmentRepository
	logger        zerolog.Logger
}

// NewPricingService constructs a new PricingService.
func NewPricingService(
	priceRuleRepo domain.PriceRuleRepository,
	equipmentRepo domain.EquipmentRepository,
	logger zerolog.Logger,
) *PricingService {
	return &PricingService{
		priceRuleRepo: priceRuleRepo,
		equipmentRepo: equipmentRepo,
		logger:        logger.With().Str("service", "pricing").Logger(),
	}
}

// Calculate computes a price for the given parameters, applying tier pricing,
// quantity discounts, and seasonal surcharges.
func (s *PricingService) Calculate(ctx context.Context, tenantID uuid.UUID, req PriceCalculationRequest) (*PriceCalculationResult, error) {
	if req.Days < 1 {
		req.Days = 1
	}
	if req.Quantity < 1 {
		req.Quantity = 1
	}

	// Parse date for seasonal check
	calcDate := time.Now()
	if req.Date != "" {
		parsed, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			calcDate = parsed
		}
	}

	// 1. Find applicable price rule:
	//    equipment-specific first, then category-level, then equipment.rental_price_day
	var rule *domain.PriceRule
	var appliedRule string

	// Try equipment-specific rule
	if req.EquipmentID != uuid.Nil {
		r, err := s.priceRuleRepo.FindByEquipment(ctx, req.EquipmentID, tenantID)
		if err == nil {
			rule = r
			appliedRule = "equipment_rule"
		}
	}

	// Try category-level rule
	if rule == nil && req.CategoryID != uuid.Nil {
		r, err := s.priceRuleRepo.FindByCategory(ctx, req.CategoryID, tenantID)
		if err == nil {
			rule = r
			appliedRule = "category_rule"
		}
	}

	// Fallback: use equipment's rental_price_day
	if rule == nil && req.EquipmentID != uuid.Nil {
		eq, err := s.equipmentRepo.GetByID(ctx, req.EquipmentID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("calculate price - equipment: %w", err)
		}

		if eq.RentalPriceDay == 0 {
			return nil, domain.ErrNoPricingAvailable
		}

		// Build a virtual rule from equipment fields
		rule = &domain.PriceRule{
			BasePriceDay: eq.RentalPriceDay,
			IsActive:     true,
		}
		if eq.RentalPriceWeek > 0 {
			rule.BasePriceWeek = &eq.RentalPriceWeek
		}
		appliedRule = "equipment_default"
	}

	if rule == nil {
		return nil, domain.ErrNoPricingAvailable
	}

	// 2. Determine effective price per day (tier pricing)
	basePricePerDay := rule.BasePriceDay
	effectivePricePerDay := basePricePerDay

	if rule.Tier3FromDays != nil && rule.Tier3PriceDay != nil && req.Days >= *rule.Tier3FromDays {
		effectivePricePerDay = *rule.Tier3PriceDay
	} else if rule.Tier2FromDays != nil && rule.Tier2PriceDay != nil && req.Days >= *rule.Tier2FromDays {
		effectivePricePerDay = *rule.Tier2PriceDay
	}

	// 3. Calculate subtotal
	subtotalBase := basePricePerDay * int64(req.Days) * int64(req.Quantity)
	subtotalEffective := effectivePricePerDay * int64(req.Days) * int64(req.Quantity)
	tierDiscount := subtotalBase - subtotalEffective

	// 4. Apply quantity discount
	var qtyDiscount int64
	if rule.QtyDiscountThreshold != nil && rule.QtyDiscountPct > 0 && req.Quantity >= *rule.QtyDiscountThreshold {
		qtyDiscount = subtotalEffective * int64(rule.QtyDiscountPct) / 100
	}

	afterQty := subtotalEffective - qtyDiscount

	// 5. Apply seasonal surcharge
	var seasonSurcharge int64
	if rule.SeasonStart != nil && rule.SeasonEnd != nil && rule.SeasonSurchargePct > 0 {
		if isInSeason(calcDate, *rule.SeasonStart, *rule.SeasonEnd) {
			seasonSurcharge = afterQty * int64(rule.SeasonSurchargePct) / 100
		}
	}

	totalNet := afterQty + seasonSurcharge

	return &PriceCalculationResult{
		BasePricePerDay: basePricePerDay,
		Days:            req.Days,
		Quantity:        req.Quantity,
		SubtotalBase:    subtotalBase,
		TierDiscount:    tierDiscount,
		QtyDiscount:     qtyDiscount,
		SeasonSurcharge: seasonSurcharge,
		TotalNet:        totalNet,
		AppliedRule:     appliedRule,
	}, nil
}

// CreatePriceRule creates a new price rule.
func (s *PricingService) CreatePriceRule(ctx context.Context, tenantID uuid.UUID, req CreatePriceRuleRequest) (*domain.PriceRule, error) {
	if req.BasePriceDay < 0 {
		return nil, fmt.Errorf("base_price_day must be non-negative")
	}

	rule := &domain.PriceRule{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		EquipmentID:          req.EquipmentID,
		CategoryID:           req.CategoryID,
		BasePriceDay:         req.BasePriceDay,
		BasePriceWeek:        req.BasePriceWeek,
		Tier2FromDays:        req.Tier2FromDays,
		Tier2PriceDay:        req.Tier2PriceDay,
		Tier3FromDays:        req.Tier3FromDays,
		Tier3PriceDay:        req.Tier3PriceDay,
		QtyDiscountThreshold: req.QtyDiscountThreshold,
		QtyDiscountPct:       req.QtyDiscountPct,
		SeasonSurchargePct:   req.SeasonSurchargePct,
		IsActive:             true,
	}

	if req.SeasonStart != nil {
		t, err := time.Parse("2006-01-02", *req.SeasonStart)
		if err != nil {
			return nil, fmt.Errorf("invalid season_start date: %w", err)
		}
		rule.SeasonStart = &t
	}
	if req.SeasonEnd != nil {
		t, err := time.Parse("2006-01-02", *req.SeasonEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid season_end date: %w", err)
		}
		rule.SeasonEnd = &t
	}

	if err := s.priceRuleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("create price rule: %w", err)
	}

	s.logger.Info().
		Str("rule_id", rule.ID.String()).
		Str("tenant_id", tenantID.String()).
		Msg("price rule created")

	return rule, nil
}

// ListPriceRules returns paginated price rules for a tenant.
func (s *PricingService) ListPriceRules(ctx context.Context, tenantID uuid.UUID, page int, perPage int) ([]*domain.PriceRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return s.priceRuleRepo.List(ctx, tenantID, page, perPage)
}

// UpdatePriceRule updates an existing price rule.
func (s *PricingService) UpdatePriceRule(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, req UpdatePriceRuleRequest) (*domain.PriceRule, error) {
	existing, err := s.priceRuleRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update price rule - fetch: %w", err)
	}

	if req.EquipmentID != nil {
		existing.EquipmentID = req.EquipmentID
	}
	if req.CategoryID != nil {
		existing.CategoryID = req.CategoryID
	}
	if req.BasePriceDay != nil {
		existing.BasePriceDay = *req.BasePriceDay
	}
	if req.BasePriceWeek != nil {
		existing.BasePriceWeek = req.BasePriceWeek
	}
	if req.Tier2FromDays != nil {
		existing.Tier2FromDays = req.Tier2FromDays
	}
	if req.Tier2PriceDay != nil {
		existing.Tier2PriceDay = req.Tier2PriceDay
	}
	if req.Tier3FromDays != nil {
		existing.Tier3FromDays = req.Tier3FromDays
	}
	if req.Tier3PriceDay != nil {
		existing.Tier3PriceDay = req.Tier3PriceDay
	}
	if req.QtyDiscountThreshold != nil {
		existing.QtyDiscountThreshold = req.QtyDiscountThreshold
	}
	if req.QtyDiscountPct != nil {
		existing.QtyDiscountPct = *req.QtyDiscountPct
	}
	if req.SeasonSurchargePct != nil {
		existing.SeasonSurchargePct = *req.SeasonSurchargePct
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.SeasonStart != nil {
		t, parseErr := time.Parse("2006-01-02", *req.SeasonStart)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid season_start date: %w", parseErr)
		}
		existing.SeasonStart = &t
	}
	if req.SeasonEnd != nil {
		t, parseErr := time.Parse("2006-01-02", *req.SeasonEnd)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid season_end date: %w", parseErr)
		}
		existing.SeasonEnd = &t
	}

	if err := s.priceRuleRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("update price rule: %w", err)
	}

	s.logger.Info().
		Str("rule_id", id.String()).
		Str("tenant_id", tenantID.String()).
		Msg("price rule updated")

	return existing, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// isInSeason checks whether calcDate falls within the season range.
// It compares month and day only so the season works across years.
func isInSeason(calcDate time.Time, seasonStart time.Time, seasonEnd time.Time) bool {
	calcMD := calcDate.Month()*100 + time.Month(calcDate.Day())
	startMD := seasonStart.Month()*100 + time.Month(seasonStart.Day())
	endMD := seasonEnd.Month()*100 + time.Month(seasonEnd.Day())

	if startMD <= endMD {
		// Normal range, e.g., Jun 1 – Sep 30
		return calcMD >= startMD && calcMD <= endMD
	}
	// Wrapping range, e.g., Nov 1 – Feb 28
	return calcMD >= startMD || calcMD <= endMD
}
