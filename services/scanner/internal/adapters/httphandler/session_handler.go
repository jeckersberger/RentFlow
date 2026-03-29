package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/scanner/internal/application"
)

// SessionHandler handles scan session and resolve HTTP endpoints.
type SessionHandler struct {
	sessionService *application.SessionService
	logger         zerolog.Logger
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(
	sessionService *application.SessionService,
	logger zerolog.Logger,
) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		logger:         logger.With().Str("handler", "session").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// Resolve handles GET /api/v1/scanner/resolve/{identifier} — equipment lookup.
func (h *SessionHandler) Resolve(w http.ResponseWriter, r *http.Request) {
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

	authHeader := r.Header.Get("Authorization")

	result, err := h.sessionService.Resolve(r.Context(), claims.TenantID, identifier, authHeader)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// CreateSession handles POST /api/v1/scanner/sessions — create a scan session.
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	session, err := h.sessionService.CreateSession(r.Context(), claims.TenantID, claims.UserID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Created(w, session)
}

// EndSession handles PUT /api/v1/scanner/sessions/{id}/end — end a scan session.
func (h *SessionHandler) EndSession(w http.ResponseWriter, r *http.Request) {
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

	session, err := h.sessionService.EndSession(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, session)
}

// GetProtocol handles GET /api/v1/scanner/sessions/{id}/protocol — session protocol.
func (h *SessionHandler) GetProtocol(w http.ResponseWriter, r *http.Request) {
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

	protocol, err := h.sessionService.GetProtocol(r.Context(), id, claims.TenantID)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, protocol)
}

// SaveSignature handles POST /api/v1/scanner/sessions/{id}/signature — save signature.
func (h *SessionHandler) SaveSignature(w http.ResponseWriter, r *http.Request) {
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

	var req application.SignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	if err := h.sessionService.SaveSignature(r.Context(), id, claims.TenantID, req); err != nil {
		errors.HandleError(w, err)
		return
	}

	response.NoContent(w)
}
