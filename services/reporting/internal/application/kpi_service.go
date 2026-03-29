package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/services/reporting/internal/domain"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateKPISnapshotRequest holds the data for creating a daily KPI snapshot.
type CreateKPISnapshotRequest struct {
	ActiveProjects        int   `json:"active_projects"`
	EquipmentOutCount     int   `json:"equipment_out_count"`
	TotalEquipment        int   `json:"total_equipment"`
	OpenInvoicesAmount    int64 `json:"open_invoices_amount"`
	OverdueInvoicesAmount int64 `json:"overdue_invoices_amount"`
	MonthlyRevenue        int64 `json:"monthly_revenue"`
	CustomerCount         int   `json:"customer_count"`
	UtilizationPct        int   `json:"utilization_pct"`
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

// KPIService implements the application-level use cases for KPI snapshots.
type KPIService struct {
	repo   domain.KPISnapshotRepository
	logger zerolog.Logger
}

// NewKPIService constructs a new KPIService.
func NewKPIService(repo domain.KPISnapshotRepository, logger zerolog.Logger) *KPIService {
	return &KPIService{
		repo:   repo,
		logger: logger.With().Str("service", "kpi").Logger(),
	}
}

// GetLatest returns the most recent KPI snapshot for the given tenant.
func (s *KPIService) GetLatest(ctx context.Context, tenantID uuid.UUID) (*domain.KPISnapshot, error) {
	snap, err := s.repo.GetLatest(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get latest KPI: %w", err)
	}
	return snap, nil
}

// GetHistory returns KPI snapshots within the given date range for charts.
func (s *KPIService) GetHistory(ctx context.Context, tenantID uuid.UUID, from, to string) ([]*domain.KPISnapshot, error) {
	if from == "" {
		from = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}

	items, err := s.repo.GetHistory(ctx, tenantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get KPI history: %w", err)
	}
	return items, nil
}

// CreateSnapshot creates or updates today's KPI snapshot (for CRON jobs).
func (s *KPIService) CreateSnapshot(ctx context.Context, tenantID uuid.UUID, req CreateKPISnapshotRequest) (*domain.KPISnapshot, error) {
	snap := &domain.KPISnapshot{
		ID:                    uuid.New(),
		TenantID:              tenantID,
		SnapshotDate:          time.Now().Format("2006-01-02"),
		ActiveProjects:        req.ActiveProjects,
		EquipmentOutCount:     req.EquipmentOutCount,
		TotalEquipment:        req.TotalEquipment,
		OpenInvoicesAmount:    req.OpenInvoicesAmount,
		OverdueInvoicesAmount: req.OverdueInvoicesAmount,
		MonthlyRevenue:        req.MonthlyRevenue,
		CustomerCount:         req.CustomerCount,
		UtilizationPct:        req.UtilizationPct,
	}

	if err := s.repo.Upsert(ctx, snap); err != nil {
		return nil, fmt.Errorf("create KPI snapshot: %w", err)
	}

	s.logger.Info().
		Str("snapshot_id", snap.ID.String()).
		Str("date", snap.SnapshotDate).
		Msg("KPI snapshot created")

	return snap, nil
}
