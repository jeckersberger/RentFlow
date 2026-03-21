package http

import (
	"net/http"

	"github.com/jeckersberger/rentflow/pkg/common/logger"
	"github.com/jeckersberger/rentflow/services/ai-service/internal/application"
)

func NewRouter(aiSvc *application.AIService, log logger.Logger) *http.ServeMux {
	router := http.NewServeMux()
	handler := NewHandler(aiSvc, log)

	router.HandleFunc("GET /health", healthHandler)
	router.HandleFunc("GET /ready", readyHandler)

	// AI endpoints
	router.HandleFunc("POST /api/v1/ai/predict", handler.Predict)
	router.HandleFunc("POST /api/v1/ai/classify", handler.Classify)
	router.HandleFunc("POST /api/v1/ai/anonymize", handler.Anonymize)
	router.HandleFunc("POST /api/v1/ai/deanonymize", handler.Deanonymize)
	router.HandleFunc("POST /api/v1/ai/suggest", handler.Suggest)
	router.HandleFunc("GET /api/v1/ai/providers", handler.GetProviders)
	router.HandleFunc("GET /api/v1/ai/usage", handler.GetUsage)

	// Anonymization rules
	router.HandleFunc("GET /api/v1/ai/anonymization-rules", handler.GetAnonymizationRules)
	router.HandleFunc("POST /api/v1/ai/anonymization-rules", handler.CreateAnonymizationRule)
	router.HandleFunc("DELETE /api/v1/ai/anonymization-rules/{id}", handler.DeleteAnonymizationRule)

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"ai-service"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"ai-service"}`))
}
