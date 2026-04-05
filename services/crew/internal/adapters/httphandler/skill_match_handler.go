package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/jeckersberger/EquipFlow/pkg/common/errors"
	"github.com/jeckersberger/EquipFlow/pkg/common/middleware"
	"github.com/jeckersberger/EquipFlow/pkg/common/response"
	"github.com/jeckersberger/EquipFlow/services/crew/internal/application"
)

// SkillMatchHandler handles the crew skill-matching HTTP endpoint.
type SkillMatchHandler struct {
	skillSvc *application.SkillMatchService
	logger   zerolog.Logger
}

// NewSkillMatchHandler creates a new SkillMatchHandler.
func NewSkillMatchHandler(
	skillSvc *application.SkillMatchService,
	logger zerolog.Logger,
) *SkillMatchHandler {
	return &SkillMatchHandler{
		skillSvc: skillSvc,
		logger:   logger.With().Str("handler", "skill_match").Logger(),
	}
}

// ---------------------------------------------------------------------------
// Endpoints
// ---------------------------------------------------------------------------

// MatchSkills handles POST /api/v1/crew/match-skills — find crew by skills + availability.
func (h *SkillMatchHandler) MatchSkills(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		errors.HandleError(w, errors.ErrUnauthorized)
		return
	}

	var req application.MatchSkillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.HandleError(w, errors.Wrap(errors.ErrBadRequest, "Ungueltiger Request-Body"))
		return
	}

	result, err := h.skillSvc.MatchSkills(r.Context(), claims.TenantID, req)
	if err != nil {
		errors.HandleError(w, err)
		return
	}

	response.Success(w, result)
}
