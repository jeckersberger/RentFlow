package domain

import (
	"testing"
)

func TestKPIUtilizationPercentage(t *testing.T) {
	kpi := &KPISnapshot{
		TotalEquipment:    200,
		EquipmentOutCount: 150,
		UtilizationPct:    75,
	}

	calculated := kpi.EquipmentOutCount * 100 / kpi.TotalEquipment
	if calculated != kpi.UtilizationPct {
		t.Errorf("calculated utilization %d != stored %d", calculated, kpi.UtilizationPct)
	}
}

func TestKPIRevenueInCents(t *testing.T) {
	kpi := &KPISnapshot{
		MonthlyRevenue:        1250000,  // 12500 EUR
		OpenInvoicesAmount:    350000,   // 3500 EUR
		OverdueInvoicesAmount: 75000,    // 750 EUR
	}

	if kpi.OverdueInvoicesAmount > kpi.OpenInvoicesAmount {
		t.Error("overdue should not exceed open invoices")
	}
	if kpi.MonthlyRevenue <= 0 {
		t.Error("monthly revenue should be positive")
	}
}

func TestReportFormats(t *testing.T) {
	validFormats := []string{"pdf", "csv", "xlsx", "json"}
	for _, f := range validFormats {
		def := &ReportDefinition{
			Name:     "Monatsbericht",
			Type:     "monthly",
			Format:   f,
			IsActive: true,
		}
		if def.Format == "" {
			t.Errorf("format should not be empty for %s", f)
		}
	}
}

func TestDashboardWidgetPosition(t *testing.T) {
	widgets := []*DashboardWidget{
		{Name: "Revenue", Position: 0, IsActive: true},
		{Name: "Equipment", Position: 1, IsActive: true},
		{Name: "Projects", Position: 2, IsActive: true},
	}

	for i, w := range widgets {
		if w.Position != i {
			t.Errorf("widget %q position = %d, want %d", w.Name, w.Position, i)
		}
	}
}

func TestReportSnapshotPeriod(t *testing.T) {
	snap := &ReportSnapshot{
		Title:       "Q1 2026",
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-03-31",
		Format:      "pdf",
	}

	if snap.PeriodStart >= snap.PeriodEnd {
		t.Error("period start should be before end")
	}
}
