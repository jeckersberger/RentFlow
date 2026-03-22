package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type KPIType string

const (
	KPITypeRevenue          KPIType = "revenue"
	KPITypeUtilization      KPIType = "utilization"
	KPITypeEquipmentCount   KPIType = "equipment_count"
	KPITypeActiveProjects   KPIType = "active_projects"
	KPITypeOverdueInvoices  KPIType = "overdue_invoices"
	KPITypeAvgRentalDays    KPIType = "avg_rental_days"
)

type KPISnapshot struct {
	ID              uuid.UUID       `db:"id"`
	TenantID        uuid.UUID       `db:"tenant_id"`
	SnapshotDate    time.Time       `db:"snapshot_date"`
	KPIType         KPIType         `db:"kpi_type"`
	Value           *float64        `db:"value"`
	PreviousValue   *float64        `db:"previous_value"`
	ChangePercentage *float64       `db:"change_percentage"`
	Metadata        json.RawMessage `db:"metadata"`
	CreatedAt       time.Time       `db:"created_at"`
}

type KPIDashboardItem struct {
	Type              KPIType  `json:"type"`
	CurrentValue      *float64 `json:"current_value"`
	PreviousValue     *float64 `json:"previous_value"`
	ChangePercentage  *float64 `json:"change_percentage"`
	FormattedValue    string   `json:"formatted_value"`
	FormattedPrevious string   `json:"formatted_previous"`
}

type KPIDashboard struct {
	Period   Period                    `json:"period"`
	ReportedAt time.Time               `json:"reported_at"`
	Items    []KPIDashboardItem        `json:"items"`
}

type KPITrend struct {
	Date  time.Time `json:"date"`
	Value *float64  `json:"value"`
}

type KPITrendData struct {
	KPIType  KPIType    `json:"kpi_type"`
	Period   Period     `json:"period"`
	Trends   []KPITrend `json:"trends"`
	ReportedAt time.Time `json:"reported_at"`
}
