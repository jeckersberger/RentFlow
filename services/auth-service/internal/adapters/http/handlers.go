package http

import (
	"encoding/json"
	"net/http"

	"github.com/jeckersberger/rentflow/services/auth-service/internal/application"
	"github.com/jeckersberger/rentflow/pkg/common/logger"
)

// AuthHandler handles authentication HTTP endpoints
type AuthHandler struct {
	registerUC *application.RegisterUserUseCase
	log        logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(registerUC *application.RegisterUserUseCase, log logger.Logger) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		log:        log,
	}
}

// RegisterRequest handles user registration requests
func (h *AuthHandler) RegisterRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req application.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	resp, err := h.registerUC.Execute(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
