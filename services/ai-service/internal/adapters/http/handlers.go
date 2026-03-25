package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/application"
)

// Handlers handles HTTP requests
type Handlers struct {
	aiService      *application.AIService
	visionService  *application.ReceiptVisionService
	log            logger.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(aiService *application.AIService, visionService *application.ReceiptVisionService, log logger.Logger) *Handlers {
	return &Handlers{
		aiService:     aiService,
		visionService: visionService,
		log:           log,
	}
}

// CreateAIRequest handles POST /api/v1/ai/complete
func (h *Handlers) CreateAIRequest(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateAIRequestCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dto, err := h.aiService.CreateAIRequest(r.Context(), cmd)
	if err != nil {
		h.log.Error("failed to create AI request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// GetAIRequest handles GET /api/v1/ai/requests/{id}
func (h *Handlers) GetAIRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	dto, err := h.aiService.GetAIRequest(r.Context(), id)
	if err != nil {
		h.log.Error("failed to get AI request", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if dto == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// ListAIRequests handles GET /api/v1/ai/requests
func (h *Handlers) ListAIRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	page := 1
	perPage := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil {
			perPage = v
		}
	}

	dtos, total, err := h.aiService.ListAIRequests(r.Context(), tenantID, page, perPage)
	if err != nil {
		h.log.Error("failed to list AI requests", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": dtos,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// SubmitFeedback handles POST /api/v1/ai/feedback
func (h *Handlers) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var cmd application.SubmitFeedbackCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dto, err := h.aiService.SubmitFeedback(r.Context(), cmd)
	if err != nil {
		h.log.Error("failed to submit feedback", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// Anonymize handles POST /api/v1/ai/anonymize
func (h *Handlers) Anonymize(w http.ResponseWriter, r *http.Request) {
	var cmd application.AnonymizeCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	anonymized := h.aiService.Anonymize(cmd.Text)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"original":    cmd.Text,
		"anonymized": anonymized,
	})
}

// PriceOptimization handles POST /api/v1/ai/price-optimize
func (h *Handlers) PriceOptimization(w http.ResponseWriter, r *http.Request) {
	var cmd application.PriceOptimizationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.aiService.OptimizePrice(r.Context(), cmd.EquipmentType, cmd.RentalDays, cmd.Season)
	if err != nil {
		h.log.Error("failed to optimize price", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"recommendation": result,
	})
}

// DemandForecast handles POST /api/v1/ai/demand-forecast
func (h *Handlers) DemandForecast(w http.ResponseWriter, r *http.Request) {
	var cmd application.DemandForecastCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.aiService.ForecastDemand(r.Context(), cmd.Category, cmd.Period)
	if err != nil {
		h.log.Error("failed to forecast demand", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"forecast": result,
	})
}

// SmartAssetCreator handles POST /api/v1/ai/asset-create
func (h *Handlers) SmartAssetCreator(w http.ResponseWriter, r *http.Request) {
	var cmd application.SmartAssetCreatorCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.aiService.CreateAsset(r.Context(), cmd.Description)
	if err != nil {
		h.log.Error("failed to create asset", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"metadata": result,
	})
}

// PredictiveMaintenance handles POST /api/v1/ai/predict-maintenance
func (h *Handlers) PredictiveMaintenance(w http.ResponseWriter, r *http.Request) {
	var cmd application.PredictiveMaintenanceCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var lastMaint *time.Time
	if cmd.LastMaintenanceAt != nil {
		t, err := time.Parse(time.RFC3339, *cmd.LastMaintenanceAt)
		if err == nil {
			lastMaint = &t
		}
	}

	result, err := h.aiService.PredictMaintenance(r.Context(), cmd.EquipmentID, cmd.UsageHours, lastMaint)
	if err != nil {
		h.log.Error("failed to predict maintenance", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"recommendation": result,
	})
}

// ListFewShots handles GET /api/v1/ai/few-shots
func (h *Handlers) ListFewShots(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	page := 1
	perPage := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}

	dtos, total, err := h.aiService.ListFewShotExamples(r.Context(), tenantID, page, perPage)
	if err != nil {
		h.log.Error("failed to list few-shot examples", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"examples": dtos,
		"total":    total,
		"page":     page,
	})
}

// CreateFewShot handles POST /api/v1/ai/few-shots
func (h *Handlers) CreateFewShot(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateFewShotExampleCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dto, err := h.aiService.CreateFewShotExample(r.Context(), cmd)
	if err != nil {
		h.log.Error("failed to create few-shot example", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// ListProviders handles GET /api/v1/ai/providers
func (h *Handlers) ListProviders(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	dtos, err := h.aiService.ListAIProviders(r.Context(), tenantID)
	if err != nil {
		h.log.Error("failed to list providers", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": dtos,
	})
}

// RegisterProvider handles POST /api/v1/ai/providers
func (h *Handlers) RegisterProvider(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateAIProviderCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dto, err := h.aiService.RegisterAIProvider(r.Context(), cmd)
	if err != nil {
		h.log.Error("failed to register provider", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// GetDashboard handles GET /api/v1/ai/dashboard
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	dashboard, err := h.aiService.GetDashboard(r.Context(), tenantID)
	if err != nil {
		h.log.Error("failed to get dashboard", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// AnalyzeReceipt handles POST /api/v1/ai/analyze-receipt
// Accepts multipart form with "file" field (image or PDF)
func (h *Handlers) AnalyzeReceipt(w http.ResponseWriter, r *http.Request) {
	if h.visionService == nil {
		http.Error(w, "Vision service not available", http.StatusServiceUnavailable)
		return
	}

	// Max 20 MB
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "File too large (max 20MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Determine MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		// Detect from content
		mimeType = http.DetectContentType(data)
	}

	// Validate MIME type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/webp":      true,
		"image/gif":       true,
		"application/pdf": true,
	}
	if !allowedTypes[mimeType] {
		http.Error(w, "Unsupported file type. Allowed: JPEG, PNG, WebP, GIF, PDF", http.StatusBadRequest)
		return
	}

	// Analyze with Claude Vision
	result, err := h.visionService.AnalyzeReceiptImage(r.Context(), data, mimeType)
	if err != nil {
		h.log.Error("failed to analyze receipt", err)
		http.Error(w, "Failed to analyze receipt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
