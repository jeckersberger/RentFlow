package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jeckersberger/rentflow/services/reporting-service/internal/domain"
)

type KPIRepository interface {
	CreateSnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error
	GetLatestSnapshot(ctx context.Context, tenantID uuid.UUID, kpiType domain.KPIType) (*domain.KPISnapshot, error)
	GetSnapshotsByDateRange(ctx context.Context, tenantID uuid.UUID, kpiType domain.KPIType, startDate, endDate time.Time) ([]domain.KPISnapshot, error)
}

type KPIService struct {
	kpiRepo KPIRepository
}

func NewKPIService(kpiRepo KPIRepository) *KPIService {
	return &KPIService{kpiRepo: kpiRepo}
}

func (s *KPIService) SnapshotKPIs(ctx context.Context, cmd *SnapshotKPIsCommand) error {
	kpiTypes := []domain.KPIType{
		domain.KPITypeRevenue,
		domain.KPITypeUtilization,
		domain.KPITypeEquipmentCount,
		domain.KPITypeActiveProjects,
		domain.KPITypeOverdueInvoices,
		domain.KPITypeAvgRentalDays,
	}

	now := time.Now()
	today := now.Truncate(24 * time.Hour)

	for _, kpiType := range kpiTypes {
		snapshot := &domain.KPISnapshot{
			ID:           uuid.New(),
			TenantID:     cmd.TenantID,
			SnapshotDate: today,
			KPIType:      kpiType,
			Value:        nil,
			PreviousValue: nil,
			CreatedAt:    now,
		}

		yesterdayStart := today.AddDate(0, 0, -1)
		yesterdayEnd := today.Add(-1 * time.Nanosecond)

		prevSnapshots, err := s.kpiRepo.GetSnapshotsByDateRange(ctx, cmd.TenantID, kpiType, yesterdayStart, yesterdayEnd)
		if err != nil && err != sql.ErrNoRows {
			continue
		}

		if len(prevSnapshots) > 0 {
			snapshot.PreviousValue = prevSnapshots[0].Value
		}

		if snapshot.Value != nil && snapshot.PreviousValue != nil {
			diff := *snapshot.Value - *snapshot.PreviousValue
			if *snapshot.PreviousValue != 0 {
				changePercent := (diff / *snapshot.PreviousValue) * 100
				snapshot.ChangePercentage = &changePercent
			}
		}

		if err := s.kpiRepo.CreateSnapshot(ctx, snapshot); err != nil {
			continue
		}
	}

	return nil
}

func (s *KPIService) GetDashboard(ctx context.Context, cmd *GetKPIDashboardCommand) (*domain.KPIDashboard, error) {
	period := domain.Period(cmd.Period)
	if period == "" {
		period = domain.PeriodMonth
	}

	dashboard := &domain.KPIDashboard{
		Period:     period,
		ReportedAt: time.Now(),
		Items:      []domain.KPIDashboardItem{},
	}

	kpiTypes := []domain.KPIType{
		domain.KPITypeRevenue,
		domain.KPITypeUtilization,
		domain.KPITypeEquipmentCount,
		domain.KPITypeActiveProjects,
		domain.KPITypeOverdueInvoices,
		domain.KPITypeAvgRentalDays,
	}

	for _, kpiType := range kpiTypes {
		latest, err := s.kpiRepo.GetLatestSnapshot(ctx, cmd.TenantID, kpiType)
		if err != nil && err != sql.ErrNoRows {
			continue
		}

		item := domain.KPIDashboardItem{
			Type:              kpiType,
			CurrentValue:      nil,
			PreviousValue:     nil,
			ChangePercentage:  nil,
			FormattedValue:    "-",
			FormattedPrevious: "-",
		}

		if latest != nil {
			item.CurrentValue = latest.Value
			item.PreviousValue = latest.PreviousValue
			item.ChangePercentage = latest.ChangePercentage

			if latest.Value != nil {
				item.FormattedValue = fmt.Sprintf("%.2f", *latest.Value)
			}
			if latest.PreviousValue != nil {
				item.FormattedPrevious = fmt.Sprintf("%.2f", *latest.PreviousValue)
			}
		}

		dashboard.Items = append(dashboard.Items, item)
	}

	return dashboard, nil
}

func (s *KPIService) GetTrends(ctx context.Context, cmd *GetKPITrendsCommand) (*domain.KPITrendData, error) {
	kpiType := domain.KPIType(cmd.KPIType)
	periods := cmd.Periods
	if periods <= 0 {
		periods = 12
	}

	trendData := &domain.KPITrendData{
		KPIType:    kpiType,
		Period:     domain.PeriodMonth,
		Trends:     []domain.KPITrend{},
		ReportedAt: time.Now(),
	}

	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, -periods, 0)

	snapshots, err := s.kpiRepo.GetSnapshotsByDateRange(ctx, cmd.TenantID, kpiType, startDate, endDate)
	if err != nil && err != sql.ErrNoRows {
		return trendData, nil
	}

	for _, snapshot := range snapshots {
		trendData.Trends = append(trendData.Trends, domain.KPITrend{
			Date:  snapshot.SnapshotDate,
			Value: snapshot.Value,
		})
	}

	return trendData, nil
}

// GetDashboardByRole returns KPIs filtered by user role
func (s *KPIService) GetDashboardByRole(ctx context.Context, cmd *GetKPIDashboardCommand, role string) (*domain.KPIDashboard, error) {
	// Define KPIs available per role
	var kpiTypes []domain.KPIType

	switch role {
	case "gf": // Geschäftsführer (Business Director) - all KPIs
		kpiTypes = []domain.KPIType{
			domain.KPITypeRevenue,
			domain.KPITypeUtilization,
			domain.KPITypeEquipmentCount,
			domain.KPITypeActiveProjects,
			domain.KPITypeOverdueInvoices,
			domain.KPITypeAvgRentalDays,
		}
	case "lager": // Warehouse - equipment and utilization only
		kpiTypes = []domain.KPIType{
			domain.KPITypeEquipmentCount,
			domain.KPITypeUtilization,
		}
	case "buchhaltung": // Accounting - revenue and overdue invoices only
		kpiTypes = []domain.KPIType{
			domain.KPITypeRevenue,
			domain.KPITypeOverdueInvoices,
		}
	default:
		// Default to empty dashboard for unknown roles
		return &domain.KPIDashboard{
			Period:     domain.PeriodMonth,
			ReportedAt: time.Now(),
			Items:      []domain.KPIDashboardItem{},
		}, nil
	}

	period := domain.Period(cmd.Period)
	if period == "" {
		period = domain.PeriodMonth
	}

	dashboard := &domain.KPIDashboard{
		Period:     period,
		ReportedAt: time.Now(),
		Items:      []domain.KPIDashboardItem{},
	}

	for _, kpiType := range kpiTypes {
		latest, err := s.kpiRepo.GetLatestSnapshot(ctx, cmd.TenantID, kpiType)
		if err != nil && err != sql.ErrNoRows {
			continue
		}

		item := domain.KPIDashboardItem{
			Type:              kpiType,
			CurrentValue:      nil,
			PreviousValue:     nil,
			ChangePercentage:  nil,
			FormattedValue:    "-",
			FormattedPrevious: "-",
		}

		if latest != nil {
			item.CurrentValue = latest.Value
			item.PreviousValue = latest.PreviousValue
			item.ChangePercentage = latest.ChangePercentage

			if latest.Value != nil {
				item.FormattedValue = fmt.Sprintf("%.2f", *latest.Value)
			}
			if latest.PreviousValue != nil {
				item.FormattedPrevious = fmt.Sprintf("%.2f", *latest.PreviousValue)
			}
		}

		dashboard.Items = append(dashboard.Items, item)
	}

	return dashboard, nil
}
