package application

// CreateAIRequestCommand represents a command to create an AI request
type CreateAIRequestCommand struct {
	TenantID    string `json:"tenant_id"`
	ProviderID  string `json:"provider_id"`
	RequestType string `json:"request_type"`
	InputText   string `json:"input_text"`
	Anonymize   bool `json:"anonymize,omitempty"`
}

// SubmitFeedbackCommand represents a command to submit feedback
type SubmitFeedbackCommand struct {
	TenantID  string `json:"tenant_id"`
	RequestID string `json:"request_id"`
	Rating    int `json:"rating"`
	Comment   *string `json:"comment,omitempty"`
	IsCorrect *bool `json:"is_correct,omitempty"`
}

// CreateFewShotExampleCommand represents a command to create a few-shot example
type CreateFewShotExampleCommand struct {
	TenantID      string `json:"tenant_id"`
	RequestType   string `json:"request_type"`
	InputExample  string `json:"input_example"`
	OutputExample string `json:"output_example"`
}

// CreateAIProviderCommand represents a command to register an AI provider
type CreateAIProviderCommand struct {
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	APIEndpoint *string                `json:"api_endpoint,omitempty"`
	ModelName   string                 `json:"model_name"`
	IsActive    bool                   `json:"is_active"`
	Priority    int                    `json:"priority"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// AnonymizeCommand represents a command to anonymize text
type AnonymizeCommand struct {
	TenantID string `json:"tenant_id"`
	Text     string `json:"text"`
}

// PriceOptimizationCommand represents a command for price optimization
type PriceOptimizationCommand struct {
	TenantID      string `json:"tenant_id"`
	EquipmentType string `json:"equipment_type"`
	RentalDays    int `json:"rental_days"`
	Season        string `json:"season"`
}

// DemandForecastCommand represents a command for demand forecasting
type DemandForecastCommand struct {
	TenantID string `json:"tenant_id"`
	Category string `json:"category"`
	Period   string `json:"period"`
}

// SmartAssetCreatorCommand represents a command for smart asset creation
type SmartAssetCreatorCommand struct {
	TenantID    string `json:"tenant_id"`
	Description string `json:"description"`
}

// PredictiveMaintenanceCommand represents a command for predictive maintenance
type PredictiveMaintenanceCommand struct {
	TenantID          string `json:"tenant_id"`
	EquipmentID       string `json:"equipment_id"`
	UsageHours        int `json:"usage_hours"`
	LastMaintenanceAt *string `json:"last_maintenance_at,omitempty"`
}
