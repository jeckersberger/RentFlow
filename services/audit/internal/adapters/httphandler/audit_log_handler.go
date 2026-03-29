package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/pagination"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/audit/internal/application"
	"github.com/jeckersberger/EquipFlow/services/audit/internal/domain"
)

type AuditLogHandler struct {
	service      *application.AuditLogService
	hashChainSvc *application.HashChainService
	logger       zerolog.Logger
}

func NewAuditLogHandler(service *application.AuditLogService, hashChainSvc *application.HashChainService, logger zerolog.Logger) *AuditLogHandler {
	return &AuditLogHandler{
		service:      service,
		hashChainSvc: hashChainSvc,
		logger:       logger.With().Str("handler", "audit_log").Logger(),
	}
}

func (h *AuditLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.service.Create(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	p := pagination.Parse(r)
	entityType := r.URL.Query().Get("entity_type")

	filter := domain.AuditLogFilter{
		Page:       p.Page,
		PerPage:    p.PerPage,
		EntityType: entityType,
	}

	if raw := r.URL.Query().Get("entity_id"); raw != "" {
		id, err := parseUUID(raw)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		filter.EntityID = &id
	}

	if raw := r.URL.Query().Get("user_id"); raw != "" {
		id, err := parseUUID(raw)
		if err != nil {
			errors.HandleError(w, err)
			return
		}
		filter.UserID = &id
	}

	items, total, err := h.service.List(r.Context(), claims.TenantID, filter)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.SuccessWithMeta(w, items, response.Meta{
		Page:    p.Page,
		PerPage: p.PerPage,
		Total:   total,
	})
}

func (h *AuditLogHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	entry, err := h.service.GetByID(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, entry)
}

// Append creates a hash-chained audit log entry (for internal service calls).
func (h *AuditLogHandler) Append(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.AppendWithHashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	created, err := h.hashChainSvc.AppendWithHash(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, created)
}

// Verify checks the integrity of the audit hash chain.
func (h *AuditLogHandler) Verify(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var fromSeq, toSeq int64
	var err error

	if fromStr != "" {
		fromSeq, err = strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger 'from' Parameter"))
			return
		}
	} else {
		fromSeq = 1
	}

	if toStr != "" {
		toSeq, err = strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger 'to' Parameter"))
			return
		}
	} else {
		toSeq = 9223372036854775807 // max int64
	}

	result, err := h.hashChainSvc.VerifyChain(r.Context(), claims.TenantID, fromSeq, toSeq)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// Export exports audit logs as CSV for the given date range.
func (h *AuditLogHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate time.Time

	if fromDateStr != "" {
		var err error
		fromDate, err = time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiges Datum fuer 'from_date' (Format: YYYY-MM-DD)"))
			return
		}
	} else {
		fromDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	if toDateStr != "" {
		var err error
		toDate, err = time.Parse("2006-01-02", toDateStr)
		if err != nil {
			errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiges Datum fuer 'to_date' (Format: YYYY-MM-DD)"))
			return
		}
		// Include the entire day.
		toDate = toDate.Add(24*time.Hour - time.Nanosecond)
	} else {
		toDate = time.Now().Add(24 * time.Hour)
	}

	filter := application.ExportFilter{
		FromDate: fromDate,
		ToDate:   toDate,
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=audit_export.csv")
	w.WriteHeader(http.StatusOK)

	if err := h.hashChainSvc.ExportCSV(r.Context(), claims.TenantID, filter, w); err != nil {
		h.logger.Error().Err(err).Msg("failed to export audit logs as CSV")
	}
}
