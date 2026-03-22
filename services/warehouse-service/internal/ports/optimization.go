package ports

import "context"

// OptimizationRecommendation represents a suggested warehouse reorganization
type OptimizationRecommendation struct {
	EquipmentID       string  `json:"equipment_id"`
	CurrentLocation   string  `json:"current_location"`
	SuggestedLocation string  `json:"suggested_location"`
	Reason            string  `json:"reason"`
	Score             float64 `json:"score"` // 0.0-1.0 confidence
}

// WarehouseMetrics provides data for optimization algorithms
type WarehouseMetrics struct {
	TenantID          string   `json:"tenant_id"`
	WarehouseID       string   `json:"warehouse_id"`
	TotalLocations    int      `json:"total_locations"`
	OccupiedLocations int      `json:"occupied_locations"`
	OccupancyRate     float64  `json:"occupancy_rate"`
	AvgPickTime       float64  `json:"avg_pick_time_seconds"`
	MovementsPerDay   float64  `json:"movements_per_day"`
	HotZones          []string `json:"hot_zones"`   // frequently accessed zones
	ColdZones         []string `json:"cold_zones"` // rarely accessed zones
}

// WarehouseOptimizer defines the interface for AI-driven warehouse optimization (Phase 4)
// Implementations may use rule-based heuristics, ML models, or external AI services
type WarehouseOptimizer interface {
	// AnalyzeLayout computes metrics and identifies optimization opportunities
	AnalyzeLayout(ctx context.Context, tenantID, warehouseID string) (*WarehouseMetrics, error)

	// GetRecommendations returns suggested equipment relocations to optimize pick paths
	GetRecommendations(ctx context.Context, tenantID, warehouseID string, limit int) ([]OptimizationRecommendation, error)

	// ApplyRecommendation executes a single recommendation (creates movement)
	ApplyRecommendation(ctx context.Context, tenantID string, rec OptimizationRecommendation) error

	// TrainModel feeds historical movement data into the optimization model
	TrainModel(ctx context.Context, tenantID, warehouseID string) error
}
