package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/inventory/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock repositories
// ---------------------------------------------------------------------------

type mockPriceRuleRepo struct {
	byEquipment map[uuid.UUID]*domain.PriceRule
	byCategory  map[uuid.UUID]*domain.PriceRule
}

func newMockPriceRuleRepo() *mockPriceRuleRepo {
	return &mockPriceRuleRepo{
		byEquipment: make(map[uuid.UUID]*domain.PriceRule),
		byCategory:  make(map[uuid.UUID]*domain.PriceRule),
	}
}

func (m *mockPriceRuleRepo) Create(_ context.Context, _ *domain.PriceRule) error { return nil }
func (m *mockPriceRuleRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.PriceRule, error) {
	return nil, domain.ErrPriceRuleNotFound
}
func (m *mockPriceRuleRepo) List(_ context.Context, _ uuid.UUID, _ int, _ int) ([]*domain.PriceRule, int64, error) {
	return nil, 0, nil
}
func (m *mockPriceRuleRepo) Update(_ context.Context, _ *domain.PriceRule) error { return nil }
func (m *mockPriceRuleRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockPriceRuleRepo) FindByEquipment(_ context.Context, equipmentID uuid.UUID, _ uuid.UUID) (*domain.PriceRule, error) {
	r, ok := m.byEquipment[equipmentID]
	if !ok {
		return nil, domain.ErrPriceRuleNotFound
	}
	return r, nil
}

func (m *mockPriceRuleRepo) FindByCategory(_ context.Context, categoryID uuid.UUID, _ uuid.UUID) (*domain.PriceRule, error) {
	r, ok := m.byCategory[categoryID]
	if !ok {
		return nil, domain.ErrPriceRuleNotFound
	}
	return r, nil
}

type mockEquipmentRepo struct {
	equipment map[uuid.UUID]*domain.Equipment
}

func newMockEquipmentRepo() *mockEquipmentRepo {
	return &mockEquipmentRepo{equipment: make(map[uuid.UUID]*domain.Equipment)}
}

func (m *mockEquipmentRepo) Create(_ context.Context, _ *domain.Equipment) error { return nil }
func (m *mockEquipmentRepo) GetByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (*domain.Equipment, error) {
	eq, ok := m.equipment[id]
	if !ok {
		return nil, domain.ErrEquipmentNotFound
	}
	return eq, nil
}
func (m *mockEquipmentRepo) GetByBarcode(_ context.Context, _ uuid.UUID, _ string) (*domain.Equipment, error) {
	return nil, domain.ErrEquipmentNotFound
}
func (m *mockEquipmentRepo) GetByRFID(_ context.Context, _ uuid.UUID, _ string) (*domain.Equipment, error) {
	return nil, domain.ErrEquipmentNotFound
}
func (m *mockEquipmentRepo) GetBySerialNumber(_ context.Context, _ uuid.UUID, _ string) (*domain.Equipment, error) {
	return nil, domain.ErrEquipmentNotFound
}
func (m *mockEquipmentRepo) List(_ context.Context, _ uuid.UUID, _ domain.EquipmentFilter) ([]*domain.Equipment, int64, error) {
	return nil, 0, nil
}
func (m *mockEquipmentRepo) Update(_ context.Context, _ *domain.Equipment) error { return nil }
func (m *mockEquipmentRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockEquipmentRepo) UpdateCondition(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockEquipmentRepo) AssignRFID(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockEquipmentRepo) Deactivate(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockEquipmentRepo) Search(_ context.Context, _ uuid.UUID, _ string, _ int, _ int) ([]*domain.Equipment, int64, error) {
	return nil, 0, nil
}

// ---------------------------------------------------------------------------
// Helper to build a PricingService with mocks
// ---------------------------------------------------------------------------

func newTestPricingService(priceRuleRepo *mockPriceRuleRepo, equipmentRepo *mockEquipmentRepo) *PricingService {
	logger := noopLogger()
	return NewPricingService(priceRuleRepo, equipmentRepo, logger)
}

func noopLogger() zerolog.Logger {
	return zerolog.Nop()
}

// intPtr returns a pointer to an int value.
func intPtr(v int) *int { return &v }

// int64Ptr returns a pointer to an int64 value.
func int64Ptr(v int64) *int64 { return &v }

// timePtr returns a pointer to a time.Time value.
func timePtr(v time.Time) *time.Time { return &v }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCalculate_BasePrice(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	priceRepo := newMockPriceRuleRepo()
	priceRepo.byEquipment[eqID] = &domain.PriceRule{
		BasePriceDay: 5000, // 50.00 EUR per day
		IsActive:     true,
	}

	svc := newTestPricingService(priceRepo, newMockEquipmentRepo())

	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        3,
		Quantity:    1,
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// 3 days * 5000 cents * 1 quantity = 15000 cents
	wantTotal := int64(15000)
	if result.TotalNet != wantTotal {
		t.Errorf("TotalNet = %d, want %d", result.TotalNet, wantTotal)
	}
	if result.SubtotalBase != wantTotal {
		t.Errorf("SubtotalBase = %d, want %d", result.SubtotalBase, wantTotal)
	}
	if result.TierDiscount != 0 {
		t.Errorf("TierDiscount = %d, want 0", result.TierDiscount)
	}
}

func TestCalculate_TierPricing(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	tier2Days := 7
	tier2Price := int64(4000) // 40.00 EUR per day from 7 days

	priceRepo := newMockPriceRuleRepo()
	priceRepo.byEquipment[eqID] = &domain.PriceRule{
		BasePriceDay:  5000,
		Tier2FromDays: &tier2Days,
		Tier2PriceDay: &tier2Price,
		IsActive:      true,
	}

	svc := newTestPricingService(priceRepo, newMockEquipmentRepo())

	// Request for 10 days -- should trigger tier 2
	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        10,
		Quantity:    1,
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// SubtotalBase: 5000 * 10 * 1 = 50000 (at base price)
	// Effective:    4000 * 10 * 1 = 40000 (at tier 2 price)
	// TierDiscount: 50000 - 40000 = 10000
	// TotalNet:     40000
	if result.SubtotalBase != 50000 {
		t.Errorf("SubtotalBase = %d, want 50000", result.SubtotalBase)
	}
	if result.TierDiscount != 10000 {
		t.Errorf("TierDiscount = %d, want 10000", result.TierDiscount)
	}
	if result.TotalNet != 40000 {
		t.Errorf("TotalNet = %d, want 40000", result.TotalNet)
	}
}

func TestCalculate_QuantityDiscount(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	threshold := 5
	discountPct := 10 // 10% off

	priceRepo := newMockPriceRuleRepo()
	priceRepo.byEquipment[eqID] = &domain.PriceRule{
		BasePriceDay:         2000, // 20.00 EUR per day
		QtyDiscountThreshold: &threshold,
		QtyDiscountPct:       discountPct,
		IsActive:             true,
	}

	svc := newTestPricingService(priceRepo, newMockEquipmentRepo())

	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        1,
		Quantity:    5,
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// SubtotalBase = 2000 * 1 * 5 = 10000
	// QtyDiscount  = 10000 * 10 / 100 = 1000
	// TotalNet     = 10000 - 1000 = 9000
	if result.QtyDiscount != 1000 {
		t.Errorf("QtyDiscount = %d, want 1000", result.QtyDiscount)
	}
	if result.TotalNet != 9000 {
		t.Errorf("TotalNet = %d, want 9000", result.TotalNet)
	}
}

func TestCalculate_SeasonSurcharge_InSeason(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	seasonStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	seasonEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	priceRepo := newMockPriceRuleRepo()
	priceRepo.byEquipment[eqID] = &domain.PriceRule{
		BasePriceDay:       10000, // 100.00 EUR per day
		SeasonStart:        &seasonStart,
		SeasonEnd:          &seasonEnd,
		SeasonSurchargePct: 20, // 20% surcharge
		IsActive:           true,
	}

	svc := newTestPricingService(priceRepo, newMockEquipmentRepo())

	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        1,
		Quantity:    1,
		Date:        "2026-07-15", // Within season
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	// SubtotalBase = 10000
	// SeasonSurcharge = 10000 * 20 / 100 = 2000
	// TotalNet = 10000 + 2000 = 12000
	if result.SeasonSurcharge != 2000 {
		t.Errorf("SeasonSurcharge = %d, want 2000", result.SeasonSurcharge)
	}
	if result.TotalNet != 12000 {
		t.Errorf("TotalNet = %d, want 12000", result.TotalNet)
	}
}

func TestCalculate_SeasonSurcharge_OutOfSeason(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	seasonStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	seasonEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	priceRepo := newMockPriceRuleRepo()
	priceRepo.byEquipment[eqID] = &domain.PriceRule{
		BasePriceDay:       10000,
		SeasonStart:        &seasonStart,
		SeasonEnd:          &seasonEnd,
		SeasonSurchargePct: 20,
		IsActive:           true,
	}

	svc := newTestPricingService(priceRepo, newMockEquipmentRepo())

	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        1,
		Quantity:    1,
		Date:        "2026-02-15", // Outside season
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	if result.SeasonSurcharge != 0 {
		t.Errorf("SeasonSurcharge = %d, want 0 (out of season)", result.SeasonSurcharge)
	}
	if result.TotalNet != 10000 {
		t.Errorf("TotalNet = %d, want 10000", result.TotalNet)
	}
}

func TestCalculate_NoPricing_FallbackToEquipment(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	eqRepo := newMockEquipmentRepo()
	eqRepo.equipment[eqID] = &domain.Equipment{
		ID:             eqID,
		RentalPriceDay: 3000,
	}

	svc := newTestPricingService(newMockPriceRuleRepo(), eqRepo)

	result, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        2,
		Quantity:    1,
	})
	if err != nil {
		t.Fatalf("Calculate() error: %v", err)
	}

	if result.AppliedRule != "equipment_default" {
		t.Errorf("AppliedRule = %q, want \"equipment_default\"", result.AppliedRule)
	}
	if result.TotalNet != 6000 {
		t.Errorf("TotalNet = %d, want 6000", result.TotalNet)
	}
}

func TestCalculate_NoPricingAvailable(t *testing.T) {
	eqID := uuid.New()
	tenantID := uuid.New()

	eqRepo := newMockEquipmentRepo()
	eqRepo.equipment[eqID] = &domain.Equipment{
		ID:             eqID,
		RentalPriceDay: 0, // No price set
	}

	svc := newTestPricingService(newMockPriceRuleRepo(), eqRepo)

	_, err := svc.Calculate(context.Background(), tenantID, PriceCalculationRequest{
		EquipmentID: eqID,
		Days:        1,
		Quantity:    1,
	})
	if err == nil {
		t.Fatal("Calculate() should return error when no pricing available")
	}
	if err != domain.ErrNoPricingAvailable {
		t.Errorf("expected ErrNoPricingAvailable, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// isInSeason helper
// ---------------------------------------------------------------------------

func TestIsInSeason(t *testing.T) {
	tests := []struct {
		name        string
		calcDate    time.Time
		seasonStart time.Time
		seasonEnd   time.Time
		want        bool
	}{
		{
			name:        "within normal season",
			calcDate:    time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
			seasonStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			seasonEnd:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
			want:        true,
		},
		{
			name:        "before normal season",
			calcDate:    time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			seasonStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			seasonEnd:   time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
			want:        false,
		},
		{
			name:        "wrapping season - in December",
			calcDate:    time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC),
			seasonStart: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			seasonEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			want:        true,
		},
		{
			name:        "wrapping season - in January",
			calcDate:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			seasonStart: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			seasonEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			want:        true,
		},
		{
			name:        "wrapping season - outside in June",
			calcDate:    time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			seasonStart: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			seasonEnd:   time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInSeason(tt.calcDate, tt.seasonStart, tt.seasonEnd)
			if got != tt.want {
				t.Errorf("isInSeason(%v, %v, %v) = %v, want %v",
					tt.calcDate.Format("2006-01-02"),
					tt.seasonStart.Format("2006-01-02"),
					tt.seasonEnd.Format("2006-01-02"),
					got, tt.want)
			}
		})
	}
}

// Prevent unused import error for fmt.
var _ = fmt.Sprintf
