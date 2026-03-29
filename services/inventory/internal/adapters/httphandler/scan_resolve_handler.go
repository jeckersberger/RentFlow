package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/inventory/internal/application"
)

// ScanResolveHandler handles scan resolve HTTP endpoints.
type ScanResolveHandler struct {
	equipmentService *application.EquipmentService
	logger           zerolog.Logger
}

// NewScanResolveHandler creates a new ScanResolveHandler.
func NewScanResolveHandler(
	equipmentService *application.EquipmentService,
	logger zerolog.Logger,
) *ScanResolveHandler {
	return &ScanResolveHandler{
		equipmentService: equipmentService,
		logger:           logger.With().Str("handler", "scan_resolve").Logger(),
	}
}

// Resolve handles GET /api/v1/scan/resolve/{identifier} — lookup by barcode, serial_number, or rfid_tag.
func (h *ScanResolveHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	identifier := chi.URLParam(r, "identifier")
	if identifier == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Identifier ist erforderlich"))
		return
	}

	equipment, err := h.equipmentService.ResolveByIdentifier(r.Context(), claims.TenantID, identifier)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, equipment)
}
