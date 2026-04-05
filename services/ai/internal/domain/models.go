package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AIPrediction struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	Type          string          `json:"type"`
	ReferenceID   *uuid.UUID      `json:"reference_id,omitempty"`
	ReferenceType string          `json:"reference_type,omitempty"`
	Prediction    json.RawMessage `json:"prediction"`
	Confidence    float64         `json:"confidence"`
	ModelVersion  string          `json:"model_version,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type AISuggestion struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Data        json.RawMessage `json:"data"`
	Status      string          `json:"status"`
	AcceptedBy  *uuid.UUID      `json:"accepted_by,omitempty"`
	AcceptedAt  *time.Time      `json:"accepted_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type AITrainingData struct {
	ID         uuid.UUID       `json:"id"`
	TenantID   uuid.UUID       `json:"tenant_id"`
	Type       string          `json:"type"`
	InputData  json.RawMessage `json:"input_data"`
	OutputData json.RawMessage `json:"output_data"`
	Feedback   string          `json:"feedback,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type PredictionFilter struct {
	Page    int
	PerPage int
	Type    string
}

type SuggestionFilter struct {
	Page    int
	PerPage int
	Type    string
	Status  string
}

type TrainingDataFilter struct {
	Page    int
	PerPage int
	Type    string
}

// ---------------------------------------------------------------------------
// Packlist Suggestion (Ollama / Gemma)
// ---------------------------------------------------------------------------

// PacklistSuggestRequest is the input for AI-based packlist generation.
type PacklistSuggestRequest struct {
	Project    ProjectSummary      `json:"project"`
	Candidates []EquipmentCandidate `json:"candidates"`
	MaxItems   int                  `json:"max_items,omitempty"`
}

// ProjectSummary holds the project metadata sent to the AI model.
type ProjectSummary struct {
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	DateFrom  string `json:"date_from"`
	DateTo    string `json:"date_to"`
	Notes     string `json:"notes,omitempty"`
}

// EquipmentCandidate describes a single equipment type available for packing.
type EquipmentCandidate struct {
	EquipmentTypeID string `json:"equipment_type_id"`
	Name            string `json:"name"`
	AvailableQty    int    `json:"available_qty"`
}

// PacklistSuggestion is one line-item in the AI-generated packlist.
type PacklistSuggestion struct {
	EquipmentTypeID string `json:"equipment_type_id"`
	Quantity        int    `json:"quantity"`
	Reason          string `json:"reason,omitempty"`
}

// PacklistSuggestResponse wraps the full AI packlist result.
type PacklistSuggestResponse struct {
	Suggestions []PacklistSuggestion `json:"suggestions"`
	Warnings    []string             `json:"warnings,omitempty"`
	ModelUsed   string               `json:"model_used"`
	Cached      bool                 `json:"cached"`
}
