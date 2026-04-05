package httphandler

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/services/invoice/internal/application"
)

// DatevHandler handles DATEV export HTTP endpoints.
type DatevHandler struct {
	datevService *application.DatevService
	logger       zerolog.Logger
}

// NewDatevHandler creates a new DatevHandler.
func NewDatevHandler(datevService *application.DatevService, logger zerolog.Logger) *DatevHandler {
	return &DatevHandler{
		datevService: datevService,
		logger:       logger.With().Str("handler", "datev").Logger(),
	}
}

// ExportCSV generates and returns a DATEV-compatible CSV file.
func (h *DatevHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	format := r.URL.Query().Get("format")

	if fromDate == "" || toDate == "" {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Parameter 'from' und 'to' sind erforderlich (YYYY-MM-DD)"))
		return
	}

	if format == "" {
		format = "skr03"
	}

	req := application.DatevExportRequest{
		FromDate: fromDate,
		ToDate:   toDate,
		Format:   format,
	}

	csvData, err := h.datevService.ExportCSV(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	filename := "DATEV_Export_" + fromDate + "_" + toDate + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=iso-8859-1")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	w.Write(csvData)
}
