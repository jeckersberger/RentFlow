package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/application"
)

// SetupRoutes sets up all HTTP routes
func SetupRoutes(
	mux *http.ServeMux,
	aiSvc *application.AIService,
	visionSvc *application.ReceiptVisionService,
	log logger.Logger,
) {
	handlers := NewHandlers(aiSvc, visionSvc, log)

	// AI request routes
	mux.HandleFunc("POST /api/v1/ai/complete", handlers.CreateAIRequest)
	mux.HandleFunc("GET /api/v1/ai/requests", handlers.ListAIRequests)
	mux.HandleFunc("GET /api/v1/ai/requests/{id}", handlers.GetAIRequest)

	// Feedback routes
	mux.HandleFunc("POST /api/v1/ai/feedback", handlers.SubmitFeedback)

	// Anonymization route
	mux.HandleFunc("POST /api/v1/ai/anonymize", handlers.Anonymize)

	// Prediction routes
	mux.HandleFunc("POST /api/v1/ai/price-optimize", handlers.PriceOptimization)
	mux.HandleFunc("POST /api/v1/ai/demand-forecast", handlers.DemandForecast)
	mux.HandleFunc("POST /api/v1/ai/asset-create", handlers.SmartAssetCreator)
	mux.HandleFunc("POST /api/v1/ai/predict-maintenance", handlers.PredictiveMaintenance)

	// Few-shot example routes
	mux.HandleFunc("GET /api/v1/ai/few-shots", handlers.ListFewShots)
	mux.HandleFunc("POST /api/v1/ai/few-shots", handlers.CreateFewShot)

	// Provider routes
	mux.HandleFunc("GET /api/v1/ai/providers", handlers.ListProviders)
	mux.HandleFunc("POST /api/v1/ai/providers", handlers.RegisterProvider)

	// Dashboard route
	mux.HandleFunc("GET /api/v1/ai/dashboard", handlers.GetDashboard)

	// Receipt/Invoice Vision Analysis
	mux.HandleFunc("POST /api/v1/ai/analyze-receipt", handlers.AnalyzeReceipt)
}
